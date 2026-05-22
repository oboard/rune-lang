package gocodegen

import (
	"bytes"

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

type generator struct {
	buf       bytes.Buffer
	file      *ir.File
	imports   map[string]bool
	indent    int
	thisNames []string
}
