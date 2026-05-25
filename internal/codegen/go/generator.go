package gocodegen

import (
	"bytes"
	"fmt"
	"strings"

	goformat "go/format"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
)

func Generate(file *ast.File, info *checker.Info) (string, error) {
	return GenerateIR(ir.LowerFile(file, info))
}

func GenerateIR(file *ir.File) (string, error) {
	g := &generator{file: file, imports: map[string]bool{}}
	for _, imp := range file.GoImports {
		g.imports[imp.Path] = true
	}
	if fileUsesType(file, checker.BigInt) {
		g.imports["math/big"] = true
	}
	if fileUsesType(file, checker.Regex) {
		g.imports["regexp"] = true
	}
	if fileUsesBinaryRuntime(file) {
		g.imports["encoding/binary"] = true
		g.imports["math"] = true
	}
	if fileUsesFSRuntime(file) {
		g.imports["os"] = true
	}
	if fileUsesTaskRuntime(file) {
		g.imports["sync"] = true
	}
	for _, fn := range file.Functions {
		ir.WalkExpr(fn.Body, func(expr ir.Expr) {
			g.collectExprImports(expr)
		})
	}
	for _, typ := range file.Types {
		for _, method := range typ.Methods {
			ir.WalkExpr(method.Body, func(expr ir.Expr) {
				g.collectExprImports(expr)
			})
		}
	}
	g.line("package main")
	g.line("")
	if len(g.imports) > 0 {
		g.line("import (")
		for name := range g.imports {
			g.linef("\t%q", name)
		}
		g.line(")")
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
	if fileUsesType(file, checker.BigInt) {
		g.bigIntRuntime()
		if len(file.Functions) > 0 || len(file.Types) > 0 {
			g.line("")
		}
	}
	if fileUsesType(file, checker.Regex) {
		g.regexRuntime()
		if len(file.Functions) > 0 || len(file.Types) > 0 {
			g.line("")
		}
	}
	if fileUsesBinaryRuntime(file) {
		g.binaryRuntime()
		if len(file.Functions) > 0 || len(file.Types) > 0 {
			g.line("")
		}
	}
	if fileUsesTaskRuntime(file) {
		g.taskRuntime()
		if len(file.Functions) > 0 || len(file.Types) > 0 {
			g.line("")
		}
	}
	if fileUsesResultRuntime(file) {
		g.resultRuntime()
		if len(file.Functions) > 0 || len(file.Types) > 0 {
			g.line("")
		}
	}
	if fileUsesErrorRuntime(file) {
		g.errorRuntime()
		if len(file.Functions) > 0 || len(file.Types) > 0 {
			g.line("")
		}
	}
	if fileUsesFSRuntime(file) {
		g.fsRuntime()
		if len(file.Functions) > 0 || len(file.Types) > 0 {
			g.line("")
		}
	}
	if fileUsesSignals(file) {
		g.signalRuntime()
		if len(file.Functions) > 0 {
			g.line("")
		}
	}
	for i, fn := range file.Functions {
		if i > 0 {
			g.line("")
		}
		if err := g.function(fn); err != nil {
			return "", err
		}
	}
	if hasMain(file) {
		if len(file.Functions) > 0 || len(file.Types) > 0 {
			g.line("")
		}
		g.line("func main() {")
		g.indent++
		if mainFn := mainFunction(file); mainFn != nil && mainFn.Routine {
			g.linef("runeAwait(%s())", mangleIdent("main"))
		} else {
			g.linef("%s()", mangleIdent("main"))
		}
		if fileUsesTaskRuntime(file) {
			g.line("runeWaitAll()")
		}
		g.indent--
		g.line("}")
	}
	if err := g.codegenError(); err != nil {
		return g.buf.String(), err
	}
	formatted, err := goformat.Source(g.buf.Bytes())
	if err != nil {
		return g.buf.String(), err
	}
	return string(formatted), nil
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
	for _, enum := range file.Enums {
		check(checker.Type(enum.Name))
	}
	return found
}

func typeContains(candidate checker.Type, typ checker.Type) bool {
	if candidate == typ {
		return true
	}
	return strings.Contains(string(candidate), string(typ))
}

func fileUsesBinaryRuntime(file *ir.File) bool {
	return fileUsesType(file, checker.Binary) ||
		fileUsesType(file, checker.Buffer) ||
		fileUsesType(file, checker.Reader) ||
		fileUsesType(file, checker.Writer)
}

func fileUsesTaskRuntime(file *ir.File) bool {
	if fileUsesGenericType(file, "Task") || fileUsesFSRuntime(file) {
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
	return false
}

func fileUsesResultRuntime(file *ir.File) bool {
	return fileUsesGenericType(file, "Result") || fileUsesFSRuntime(file)
}

func fileUsesErrorRuntime(file *ir.File) bool {
	return fileUsesType(file, checker.Error) || fileUsesFSRuntime(file)
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

func (g *generator) bigIntRuntime() {
	g.line("func runeBigInt(src string) *big.Int {")
	g.indent++
	g.line("value, ok := new(big.Int).SetString(src, 10)")
	g.line("if !ok {")
	g.indent++
	g.line("panic(\"invalid BigInt literal\")")
	g.indent--
	g.line("}")
	g.line("return value")
	g.indent--
	g.line("}")
}

func (g *generator) taskRuntime() {
	g.line("type runeUnit struct{}")
	g.line("")
	g.line("type runeTask[T any] <-chan T")
	g.line("")
	g.line("var runeTasks sync.WaitGroup")
	g.line("")
	g.line("func runeGo[T any](work func() T) runeTask[T] {")
	g.indent++
	g.line("runeTasks.Add(1)")
	g.line("ch := make(chan T, 1)")
	g.line("go func() {")
	g.indent++
	g.line("defer runeTasks.Done()")
	g.line("ch <- work()")
	g.indent--
	g.line("}()")
	g.line("return ch")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeWaitAll() {")
	g.indent++
	g.line("runeTasks.Wait()")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeAwait[T any](task runeTask[T]) T {")
	g.indent++
	g.line("return <-task")
	g.indent--
	g.line("}")
}

func (g *generator) resultRuntime() {
	g.line("type runeResult[T any, E any] struct {")
	g.indent++
	g.line("ok bool")
	g.line("value T")
	g.line("err E")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeOk[T any, E any](value T) runeResult[T, E] {")
	g.indent++
	g.line("return runeResult[T, E]{ok: true, value: value}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeErr[T any, E any](err E) runeResult[T, E] {")
	g.indent++
	g.line("return runeResult[T, E]{err: err}")
	g.indent--
	g.line("}")
}

func (g *generator) errorRuntime() {
	g.line("type runeError struct {")
	g.indent++
	g.line("__code int")
	g.line("__message string")
	g.line("__cause *runeError")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeErrorFrom(err error) *runeError {")
	g.indent++
	g.line("if err == nil {")
	g.indent++
	g.line("return nil")
	g.indent--
	g.line("}")
	g.line("return &runeError{__code: 1, __message: err.Error()}")
	g.indent--
	g.line("}")
}

func (g *generator) fsRuntime() {
	g.line("func runeReadFile(path string) runeTask[runeResult[[]byte, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[[]byte, *runeError] {")
	g.indent++
	g.line("data, err := os.ReadFile(path)")
	g.line("if err != nil {")
	g.indent++
	g.line("return runeErr[[]byte, *runeError](runeErrorFrom(err))")
	g.indent--
	g.line("}")
	g.line("return runeOk[[]byte, *runeError](data)")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
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

func blockUsesSignals(expr ir.Expr) bool {
	found := false
	ir.WalkExpr(expr, func(expr ir.Expr) {
		if _, ok := expr.(*ir.WatchExpr); ok {
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

type generator struct {
	buf       bytes.Buffer
	file      *ir.File
	imports   map[string]bool
	indent    int
	temp      int
	errors    []error
	thisNames []string
	signals   []map[string]checker.Type
}

func (g *generator) nextTemp(prefix string) string {
	g.temp++
	return fmt.Sprintf("%s%d", mangleIdent(prefix), g.temp)
}
