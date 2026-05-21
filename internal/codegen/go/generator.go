package gocodegen

import (
	"bytes"
	"fmt"
	goformat "go/format"
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/lexer"
)

func Generate(file *ast.File, info *checker.Info) (string, error) {
	g := &generator{info: info, imports: map[string]bool{}}
	for _, fn := range file.Functions {
		ast.WalkExpr(fn.Body, func(expr ast.Expr) {
			if usesFmt(expr) {
				g.imports["fmt"] = true
			}
		})
	}
	for _, typ := range file.Types {
		for _, method := range typ.Methods {
			ast.WalkExpr(method.Body, func(expr ast.Expr) {
				if usesFmt(expr) {
					g.imports["fmt"] = true
				}
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
	buf     bytes.Buffer
	info    *checker.Info
	imports map[string]bool
	indent  int
}

func (g *generator) structType(typ *ast.StructType) {
	g.linef("type %s struct {", mangleIdent(typ.Name))
	g.indent++
	for _, field := range typ.Fields {
		g.linef("%s %s", mangleIdent(field.Name), goType(checker.Type(field.Type)))
	}
	g.indent--
	g.line("}")
}

func (g *generator) function(fn *ast.Function) error {
	info := g.info.Functions[fn.Name]
	if info == nil {
		return fmt.Errorf("missing function info for %s", fn.Name)
	}
	params := make([]string, 0, len(info.Params))
	for _, param := range info.Params {
		params = append(params, fmt.Sprintf("%s %s", mangleIdent(param.Name), goType(param.Type)))
	}
	ret := ""
	if info.Return != checker.Void {
		ret = " " + goType(info.Return)
	}
	g.linef("func %s(%s)%s {", mangleIdent(fn.Name), strings.Join(params, ", "), ret)
	g.indent++
	if err := g.body(fn, fn.Body, info.Return); err != nil {
		return err
	}
	g.indent--
	g.line("}")
	return nil
}

func (g *generator) method(typ *ast.StructType, fn *ast.Function) error {
	structInfo := g.info.Types[typ.Name]
	if structInfo == nil {
		return fmt.Errorf("missing type info for %s", typ.Name)
	}
	info := structInfo.Methods[fn.Name]
	if info == nil {
		return fmt.Errorf("missing method info for %s.%s", typ.Name, fn.Name)
	}
	params := make([]string, 0, len(info.Params))
	for _, param := range info.Params {
		params = append(params, fmt.Sprintf("%s %s", mangleIdent(param.Name), goType(param.Type)))
	}
	ret := ""
	if info.Return != checker.Void {
		ret = " " + goType(info.Return)
	}
	g.linef("func (%s %s) %s(%s)%s {", mangleIdent("this"), mangleIdent(typ.Name), mangleIdent(fn.Name), strings.Join(params, ", "), ret)
	g.indent++
	if err := g.body(fn, fn.Body, info.Return); err != nil {
		return err
	}
	g.indent--
	g.line("}")
	return nil
}

func (g *generator) body(fn *ast.Function, expr ast.Expr, ret checker.Type) error {
	switch e := expr.(type) {
	case *ast.PatternBlock:
		return g.patternBlock(fn, e, ret)
	case *ast.BlockExpr:
		return g.block(e, ret)
	default:
		if ret == checker.Void {
			g.line(g.expr(expr))
		} else {
			g.linef("return %s", g.expr(expr))
		}
	}
	return nil
}

func (g *generator) block(block *ast.BlockExpr, ret checker.Type) error {
	for i, stmt := range block.Statements {
		last := i == len(block.Statements)-1
		switch s := stmt.(type) {
		case *ast.LetStmt:
			g.linef("%s := %s", mangleIdent(s.Name), g.expr(s.Value))
		case *ast.AssignStmt:
			g.linef("%s = %s", mangleIdent(s.Name), g.expr(s.Value))
		case *ast.ExprStmt:
			if last && ret != checker.Void {
				g.linef("return %s", g.expr(s.Expr))
			} else {
				g.line(g.expr(s.Expr))
			}
		}
	}
	if ret != checker.Void && len(block.Statements) == 0 {
		g.linef("return %s", zeroValue(ret))
	}
	return nil
}

func (g *generator) patternBlock(fn *ast.Function, block *ast.PatternBlock, ret checker.Type) error {
	if len(fn.Params) != 1 {
		return fmt.Errorf("%s: pattern blocks currently require exactly one parameter", block.Pos)
	}
	subject := mangleIdent(fn.Params[0].Name)
	g.line("switch {")
	g.indent++
	hasDefault := false
	for _, branch := range block.Branches {
		if _, ok := branch.Pattern.(*ast.WildcardPattern); ok {
			hasDefault = true
			g.line("default:")
		} else {
			g.linef("case %s:", g.patternCondition(subject, branch.Pattern))
		}
		g.indent++
		if ret == checker.Void {
			g.line(g.expr(branch.Expr))
		} else {
			g.linef("return %s", g.expr(branch.Expr))
		}
		g.indent--
	}
	g.indent--
	g.line("}")
	if ret != checker.Void && !hasDefault {
		g.linef("return %s", zeroValue(ret))
	}
	return nil
}

func (g *generator) patternCondition(subject string, pattern ast.Pattern) string {
	switch p := pattern.(type) {
	case *ast.LiteralPattern:
		return fmt.Sprintf("%s == %s", subject, g.expr(p.Value))
	case *ast.ComparePattern:
		return fmt.Sprintf("%s %s %s", subject, p.Op, g.expr(p.Value))
	case *ast.TuplePattern:
		parts := make([]string, 0, len(p.Elements))
		for i, elem := range p.Elements {
			parts = append(parts, g.patternCondition(fmt.Sprintf("%s[%d]", subject, i), elem))
		}
		return strings.Join(parts, " && ")
	default:
		return "true"
	}
}

func (g *generator) expr(expr ast.Expr) string {
	return g.exprPrec(expr, 0)
}

func (g *generator) exprPrec(expr ast.Expr, parentPrec int) string {
	switch e := expr.(type) {
	case *ast.Identifier:
		return mangleIdent(e.Name)
	case *ast.IntegerLiteral:
		return strconv.Itoa(e.Value)
	case *ast.StringLiteral:
		return strconv.Quote(e.Value)
	case *ast.BoolLiteral:
		if e.Value {
			return "true"
		}
		return "false"
	case *ast.UnaryExpr:
		s := fmt.Sprintf("%s%s", e.Op, g.exprPrec(e.Expr, 5))
		if 5 < parentPrec {
			return "(" + s + ")"
		}
		return s
	case *ast.BinaryExpr:
		prec := goPrecedence(e.Op)
		s := fmt.Sprintf("%s %s %s", g.exprPrec(e.Left, prec), e.Op, g.exprPrec(e.Right, prec+1))
		if prec < parentPrec {
			return "(" + s + ")"
		}
		return s
	case *ast.CallExpr:
		args := make([]string, 0, len(e.Args))
		for _, arg := range e.Args {
			args = append(args, g.expr(arg))
		}
		return fmt.Sprintf("%s(%s)", g.expr(e.Callee), strings.Join(args, ", "))
	case *ast.SelectorExpr:
		if at, ok := e.Receiver.(*ast.AtExpr); ok {
			if at.Name == "fmt" {
				return "fmt." + exportedFmtName(e.Name)
			}
		}
		return g.expr(e.Receiver) + "." + mangleIdent(e.Name)
	case *ast.StructLiteral:
		fields := make([]string, 0, len(e.Fields))
		for _, field := range e.Fields {
			fields = append(fields, fmt.Sprintf("%s: %s", mangleIdent(field.Name), g.expr(field.Value)))
		}
		return fmt.Sprintf("%s{%s}", mangleIdent(e.TypeName), strings.Join(fields, ", "))
	case *ast.AtExpr:
		return e.Name
	case *ast.ThisExpr:
		return mangleIdent("this")
	default:
		return "/* unsupported */"
	}
}

func goPrecedence(op lexer.Kind) int {
	switch op {
	case lexer.EqualEqual, lexer.BangEqual, lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
		return 1
	case lexer.Plus, lexer.Minus:
		return 2
	case lexer.Star, lexer.Slash, lexer.Percent:
		return 3
	default:
		return 0
	}
}

func (g *generator) line(s string) {
	for i := 0; i < g.indent; i++ {
		g.buf.WriteByte('\t')
	}
	g.buf.WriteString(s)
	g.buf.WriteByte('\n')
}

func (g *generator) linef(format string, args ...any) {
	g.line(fmt.Sprintf(format, args...))
}

func goType(typ checker.Type) string {
	switch typ {
	case checker.Int:
		return "int"
	case checker.String:
		return "string"
	case checker.Bool:
		return "bool"
	default:
		return mangleIdent(string(typ))
	}
}

func zeroValue(typ checker.Type) string {
	switch typ {
	case checker.Int:
		return "0"
	case checker.String:
		return `""`
	case checker.Bool:
		return "false"
	default:
		return fmt.Sprintf("%s{}", goType(typ))
	}
}

func hasMain(file *ast.File) bool {
	for _, fn := range file.Functions {
		if fn.Name == "main" {
			return true
		}
	}
	return false
}

func mangleIdent(name string) string {
	return "__" + name
}

func usesFmt(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	at, ok := sel.Receiver.(*ast.AtExpr)
	return ok && at.Name == "fmt"
}

func exportedFmtName(name string) string {
	switch name {
	case "println":
		return "Println"
	case "print":
		return "Print"
	case "printf":
		return "Printf"
	default:
		return name
	}
}
