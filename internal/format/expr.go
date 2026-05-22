package format

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
)

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
	case *ast.LambdaExpr:
		params := strings.Join(e.Params, ", ")
		if len(e.Params) != 1 {
			params = "(" + params + ")"
		}
		return params + " => " + f.expr(e.Body)
	case *ast.IndexExpr:
		return fmt.Sprintf("%s[%s]", f.expr(e.Receiver), f.expr(e.Index))
	case *ast.SelectorExpr:
		if _, ok := e.Receiver.(*ast.ThisExpr); ok {
			return "." + e.Name
		}
		return f.expr(e.Receiver) + "." + e.Name
	case *ast.ArrayLiteral:
		elems := make([]string, 0, len(e.Elements))
		for _, elem := range e.Elements {
			elems = append(elems, f.expr(elem))
		}
		return "[" + strings.Join(elems, ", ") + "]"
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
	case *ast.AnonymousObjectLiteral:
		var b strings.Builder
		b.WriteString("{")
		for i, field := range e.Fields {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(" ")
			if lambda, ok := field.Value.(*ast.LambdaExpr); ok {
				b.WriteString(field.Name)
				b.WriteString("(")
				b.WriteString(strings.Join(lambda.Params, ", "))
				b.WriteString(") => ")
				b.WriteString(f.expr(lambda.Body))
			} else {
				b.WriteString(field.Name)
				b.WriteString(": ")
				b.WriteString(f.expr(field.Value))
			}
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
