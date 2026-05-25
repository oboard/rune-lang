package tscodegen

import (
	"fmt"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
)

func Generate(file *ast.File, info *checker.Info) (string, error) {
	return GenerateIR(ir.LowerFile(file, info))
}

func GenerateIR(file *ir.File) (string, error) {
	if len(file.GoImports) > 0 {
		return "", fmt.Errorf("TypeScript backend does not support @go.import")
	}
	if usesGoFFI(file) {
		return "", fmt.Errorf("TypeScript backend does not support @go FFI")
	}
	g := &generator{file: file}
	if fileUsesAsyncRuntime(file) {
		g.asyncRuntime()
		g.line("")
	}
	if fileUsesBinaryRuntime(file) {
		g.binaryRuntime()
		g.line("")
	}
	if fileUsesSignals(file) {
		g.signalRuntime()
		g.line("")
	}
	for i, enum := range file.Enums {
		if i > 0 {
			g.line("")
		}
		g.enumType(enum)
	}
	if len(file.Enums) > 0 && len(file.Types) > 0 {
		g.line("")
	}
	for i, typ := range file.Types {
		if i > 0 {
			g.line("")
		}
		g.structType(typ)
		for _, method := range typ.Methods {
			g.line("")
			if err := g.method(typ, method); err != nil {
				return "", err
			}
		}
	}
	if (len(file.Types) > 0 || len(file.Enums) > 0) && len(file.Functions) > 0 {
		g.line("")
	}
	for i, fn := range file.Functions {
		if i > 0 {
			g.line("")
		}
		if err := g.function(fn); err != nil {
			return "", err
		}
	}
	if len(file.Functions) > 0 {
		g.line("")
		exports := make([]string, 0, len(file.Functions))
		for _, fn := range file.Functions {
			exports = append(exports, fmt.Sprintf("%s as %s", mangleIdent(fn.Name), fn.Name))
		}
		g.linef("export { %s };", join(exports, ", "))
	}
	if err := g.codegenError(); err != nil {
		return g.buf.String(), err
	}
	return g.buf.String(), nil
}

func usesGoFFI(file *ir.File) bool {
	found := false
	for _, fn := range file.Functions {
		ir.WalkExpr(fn.Body, func(expr ir.Expr) {
			if selectorUsesGo(expr) {
				found = true
			}
		})
	}
	for _, typ := range file.Types {
		for _, method := range typ.Methods {
			ir.WalkExpr(method.Body, func(expr ir.Expr) {
				if selectorUsesGo(expr) {
					found = true
				}
			})
		}
	}
	return found
}

func selectorUsesGo(expr ir.Expr) bool {
	sel, ok := expr.(*ir.SelectorExpr)
	if !ok {
		return false
	}
	at, ok := sel.Receiver.(*ir.AtExpr)
	return ok && at.Name == "go"
}

func (g *generator) asyncRuntime() {
	g.line("type RuneResult<T, E> = { ok: true; value: T } | { ok: false; error: E };")
	g.line("type RuneError = { code: number; message: string; cause: RuneError | null };")
	g.line("")
	g.line("const runeTasks: Promise<unknown>[] = [];")
	g.line("")
	g.line("function runeGo<T>(work: () => T | Promise<T>): Promise<T> {")
	g.indent++
	g.line("const task = Promise.resolve().then(work);")
	g.line("runeTasks.push(task);")
	g.line("task.finally(() => {")
	g.indent++
	g.line("const index = runeTasks.indexOf(task);")
	g.line("if (index >= 0) {")
	g.indent++
	g.line("runeTasks.splice(index, 1);")
	g.indent--
	g.line("}")
	g.indent--
	g.line("});")
	g.line("return task;")
	g.indent--
	g.line("}")
	g.line("")
	g.line("async function runeWaitAll(): Promise<void> {")
	g.indent++
	g.line("while (runeTasks.length > 0) {")
	g.indent++
	g.line("await Promise.allSettled([...runeTasks]);")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("function runeOk<T, E>(value: T): RuneResult<T, E> {")
	g.indent++
	g.line("return { ok: true, value };")
	g.indent--
	g.line("}")
	g.line("")
	g.line("function runeErr<T, E>(error: E): RuneResult<T, E> {")
	g.indent++
	g.line("return { ok: false, error };")
	g.indent--
	g.line("}")
	g.line("")
	g.line("function runeErrorFrom(error: unknown): RuneError {")
	g.indent++
	g.line("return { code: 1, message: error instanceof Error ? error.message : String(error), cause: null };")
	g.indent--
	g.line("}")
	if fileUsesFSRuntime(g.file) {
		g.line("")
		g.line("async function runeReadFile(path: string): Promise<RuneResult<Uint8Array, RuneError>> {")
		g.indent++
		g.line("try {")
		g.indent++
		g.line("const fs = await import(\"node:fs/promises\");")
		g.line("return runeOk<Uint8Array, RuneError>(await fs.readFile(path));")
		g.indent--
		g.line("} catch (error) {")
		g.indent++
		g.line("return runeErr<Uint8Array, RuneError>(runeErrorFrom(error));")
		g.indent--
		g.line("}")
		g.indent--
		g.line("}")
	}
}

func join(parts []string, sep string) string {
	out := ""
	for i, part := range parts {
		if i > 0 {
			out += sep
		}
		out += part
	}
	return out
}

func fileUsesBinaryRuntime(file *ir.File) bool {
	return fileUsesType(file, checker.Buffer) ||
		fileUsesType(file, checker.Reader) ||
		fileUsesType(file, checker.Writer)
}

func fileUsesType(file *ir.File, typ checker.Type) bool {
	found := false
	check := func(candidate checker.Type) {
		if found || typeContains(candidate, typ) {
			found = true
		}
	}
	for _, fn := range file.Functions {
		check(fn.Return)
		for _, param := range fn.Params {
			check(param.Type)
		}
		ir.WalkExpr(fn.Body, func(expr ir.Expr) {
			check(expr.ResultType())
		})
	}
	for _, test := range file.Tests {
		ir.WalkExpr(test.Body, func(expr ir.Expr) {
			check(expr.ResultType())
		})
	}
	for _, typDecl := range file.Types {
		for _, field := range typDecl.Fields {
			check(field.Type)
		}
		for _, method := range typDecl.Methods {
			check(method.Return)
			for _, param := range method.Params {
				check(param.Type)
			}
			ir.WalkExpr(method.Body, func(expr ir.Expr) {
				check(expr.ResultType())
			})
		}
	}
	return found
}

func typeContains(candidate checker.Type, typ checker.Type) bool {
	if candidate == typ {
		return true
	}
	return strings.Contains(string(candidate), string(typ))
}

func fileUsesSignals(file *ir.File) bool {
	for _, fn := range file.Functions {
		if blockUsesSignals(fn.Body) {
			return true
		}
	}
	for _, typ := range file.Types {
		for _, method := range typ.Methods {
			if blockUsesSignals(method.Body) {
				return true
			}
		}
	}
	return false
}

func fileUsesAsyncRuntime(file *ir.File) bool {
	if fileUsesType(file, checker.Data) || fileUsesType(file, checker.Error) || fileUsesGenericType(file, "Result") || fileUsesGenericType(file, "Task") {
		return true
	}
	for _, fn := range file.Functions {
		if fn.Routine || exprUsesAsync(fn.Body) {
			return true
		}
	}
	for _, typ := range file.Types {
		for _, method := range typ.Methods {
			if method.Routine || exprUsesAsync(method.Body) {
				return true
			}
		}
	}
	return fileUsesFSRuntime(file)
}

func fileUsesGenericType(file *ir.File, base string) bool {
	found := false
	check := func(candidate checker.Type) {
		if found {
			return
		}
		found = typeUsesGeneric(candidate, base)
	}
	for _, fn := range file.Functions {
		check(fn.Return)
		for _, param := range fn.Params {
			check(param.Type)
		}
		ir.WalkExpr(fn.Body, func(expr ir.Expr) {
			check(expr.ResultType())
		})
	}
	for _, typ := range file.Types {
		for _, field := range typ.Fields {
			check(field.Type)
		}
		for _, method := range typ.Methods {
			check(method.Return)
			for _, param := range method.Params {
				check(param.Type)
			}
			ir.WalkExpr(method.Body, func(expr ir.Expr) {
				check(expr.ResultType())
			})
		}
	}
	return found
}

func typeUsesGeneric(candidate checker.Type, base string) bool {
	name := string(candidate)
	return strings.HasPrefix(name, base+"[") || strings.Contains(name, ","+base+"[") || strings.Contains(name, "["+base+"[")
}

func exprUsesAsync(expr ir.Expr) bool {
	found := false
	ir.WalkExpr(expr, func(expr ir.Expr) {
		if call, ok := expr.(*ir.CallExpr); ok && call.Async {
			found = true
		}
		if _, ok := expr.(*ir.ResultUnwrapExpr); ok {
			found = true
		}
	})
	return found
}

func fileUsesFSRuntime(file *ir.File) bool {
	found := false
	check := func(expr ir.Expr) {
		call, ok := expr.(*ir.CallExpr)
		if !ok || file.Stdlib == nil {
			return
		}
		sel, ok := call.Callee.(*ir.SelectorExpr)
		if !ok {
			return
		}
		at, ok := sel.Receiver.(*ir.AtExpr)
		if !ok {
			return
		}
		fn, ok := file.Stdlib.Function(at.Name, sel.Name)
		if ok && fn.Intrinsic == "fs.readFile" {
			found = true
		}
	}
	for _, fn := range file.Functions {
		ir.WalkExpr(fn.Body, check)
	}
	for _, typ := range file.Types {
		for _, method := range typ.Methods {
			ir.WalkExpr(method.Body, check)
		}
	}
	return found
}

func blockUsesSignals(expr ir.Expr) bool {
	found := false
	ir.WalkExpr(expr, func(expr ir.Expr) {
		if _, ok := expr.(*ir.WatchExpr); ok {
			found = true
		}
		if _, ok := expr.(*ir.ReactiveLiteral); ok {
			found = true
		}
	})
	if found {
		return true
	}
	if block, ok := expr.(*ir.BlockExpr); ok {
		for _, stmt := range block.Statements {
			if let, ok := stmt.(*ir.LetStmt); ok && let.Signal {
				return true
			}
		}
	}
	return false
}
