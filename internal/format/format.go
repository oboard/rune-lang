package format

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
)

func File(file *ast.File) string {
	f := formatter{}
	for _, imp := range file.GoImports {
		f.linef("@go.import(%s)", strconv.Quote(imp.Path))
	}
	if len(file.GoImports) > 0 && (len(file.Types) > 0 || len(file.Functions) > 0) {
		f.line("")
	}
	for i, typ := range file.Types {
		if i > 0 {
			f.line("")
		}
		f.structType(typ)
	}
	if len(file.Types) > 0 && len(file.Functions) > 0 {
		f.line("")
	}
	for i, fn := range file.Functions {
		if i > 0 {
			f.line("")
		}
		f.function(fn)
	}
	return f.b.String()
}

type formatter struct {
	b      strings.Builder
	indent int
}

func (f *formatter) structType(typ *ast.StructType) {
	f.linef("%s: {", typ.Name)
	f.indent++
	for _, field := range typ.Fields {
		f.linef("%s: %s", field.Name, field.Type)
	}
	if len(typ.Fields) > 0 && len(typ.Methods) > 0 {
		f.line("")
	}
	for i, method := range typ.Methods {
		if i > 0 {
			f.line("")
		}
		f.function(method)
	}
	f.indent--
	f.line("}")
}

func (f *formatter) function(fn *ast.Function) {
	var params []string
	for _, param := range fn.Params {
		params = append(params, fmt.Sprintf("%s: %s", param.Name, param.Type))
	}
	ret := ""
	if fn.ReturnType != "" {
		ret = " -> " + fn.ReturnType
	}
	switch body := fn.Body.(type) {
	case *ast.PatternBlock:
		f.linef("%s(%s)%s => {", fn.Name, strings.Join(params, ", "), ret)
		f.indent++
		for _, branch := range body.Branches {
			f.linef("%s => %s", f.pattern(branch.Pattern), f.expr(branch.Expr))
		}
		f.indent--
		f.line("}")
	case *ast.BlockExpr:
		f.linef("%s(%s)%s => {", fn.Name, strings.Join(params, ", "), ret)
		f.indent++
		for _, stmt := range body.Statements {
			f.line(f.stmt(stmt))
		}
		f.indent--
		f.line("}")
	default:
		f.linef("%s(%s)%s => %s", fn.Name, strings.Join(params, ", "), ret, f.expr(fn.Body))
	}
}

func (f *formatter) stmt(stmt ast.Stmt) string {
	switch s := stmt.(type) {
	case *ast.LetStmt:
		op := ":="
		if s.Mutable {
			op = "~="
		}
		return fmt.Sprintf("%s %s %s", s.Name, op, f.expr(s.Value))
	case *ast.AssignStmt:
		return fmt.Sprintf("%s = %s", s.Name, f.expr(s.Value))
	case *ast.ExprStmt:
		return f.expr(s.Expr)
	default:
		return ""
	}
}

func (f *formatter) expr(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Identifier:
		return e.Name
	case *ast.AtExpr:
		return "@" + e.Name
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
		return e.Op.String() + f.expr(e.Expr)
	case *ast.BinaryExpr:
		return fmt.Sprintf("%s %s %s", f.exprWithParens(e.Left), e.Op, f.exprWithParens(e.Right))
	case *ast.CallExpr:
		args := make([]string, 0, len(e.Args))
		for _, arg := range e.Args {
			args = append(args, f.expr(arg))
		}
		return fmt.Sprintf("%s(%s)", f.expr(e.Callee), strings.Join(args, ", "))
	case *ast.SelectorExpr:
		if _, ok := e.Receiver.(*ast.ThisExpr); ok {
			return "." + e.Name
		}
		return f.expr(e.Receiver) + "." + e.Name
	case *ast.StructLiteral:
		var b strings.Builder
		b.WriteString(e.TypeName)
		b.WriteString(" {")
		for i, field := range e.Fields {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(" ")
			b.WriteString(field.Name)
			b.WriteString(": ")
			b.WriteString(f.expr(field.Value))
		}
		b.WriteString(" }")
		return b.String()
	case *ast.BlockExpr:
		return "{ ... }"
	case *ast.PatternBlock:
		return "{ ... }"
	case *ast.ThisExpr:
		return "this"
	default:
		return ""
	}
}

func (f *formatter) exprWithParens(expr ast.Expr) string {
	if _, ok := expr.(*ast.BinaryExpr); ok {
		return "(" + f.expr(expr) + ")"
	}
	return f.expr(expr)
}

func (f *formatter) pattern(pattern ast.Pattern) string {
	switch p := pattern.(type) {
	case *ast.WildcardPattern:
		return "_"
	case *ast.LiteralPattern:
		return f.expr(p.Value)
	case *ast.ComparePattern:
		return p.Op.String() + f.expr(p.Value)
	case *ast.TuplePattern:
		parts := make([]string, 0, len(p.Elements))
		for _, elem := range p.Elements {
			parts = append(parts, f.pattern(elem))
		}
		return "(" + strings.Join(parts, ", ") + ")"
	default:
		return "_"
	}
}

func (f *formatter) line(s string) {
	if s != "" {
		for i := 0; i < f.indent; i++ {
			f.b.WriteString("  ")
		}
	}
	f.b.WriteString(s)
	f.b.WriteByte('\n')
}

func (f *formatter) linef(format string, args ...any) {
	f.line(fmt.Sprintf(format, args...))
}
