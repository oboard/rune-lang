package gocodegen

import (
	"bytes"
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
	if len(file.Types) > 0 && len(file.Functions) > 0 {
		g.line("")
	}
	if fileUsesType(file, checker.BigInt) {
		g.bigIntRuntime()
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
		g.linef("%s()", mangleIdent("main"))
		g.indent--
		g.line("}")
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
	return found
}

func typeContains(candidate checker.Type, typ checker.Type) bool {
	if candidate == typ {
		return true
	}
	return strings.Contains(string(candidate), string(typ))
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
	thisNames []string
	signals   []map[string]checker.Type
}
