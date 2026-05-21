package gocodegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
	"github.com/oboard/rune-lang/internal/lexer"
)

func (g *generator) expr(expr ir.Expr) string {
	return g.exprPrec(expr, 0)
}

func (g *generator) exprPrec(expr ir.Expr, parentPrec int) string {
	switch e := expr.(type) {
	case *ir.Identifier:
		return mangleIdent(e.Name)
	case *ir.IntegerLiteral:
		return strconv.Itoa(e.Value)
	case *ir.StringLiteral:
		return strconv.Quote(e.Value)
	case *ir.BoolLiteral:
		if e.Value {
			return "true"
		}
		return "false"
	case *ir.UnaryExpr:
		s := fmt.Sprintf("%s%s", e.Op, g.exprPrec(e.Expr, 5))
		if 5 < parentPrec {
			return "(" + s + ")"
		}
		return s
	case *ir.BinaryExpr:
		prec := goPrecedence(e.Op)
		s := fmt.Sprintf("%s %s %s", g.exprPrec(e.Left, prec), e.Op, g.exprPrec(e.Right, prec+1))
		if prec < parentPrec {
			return "(" + s + ")"
		}
		return s
	case *ir.CallExpr:
		if ffi, ok := g.goFFICall(e); ok {
			return ffi
		}
		if arrayCall, ok := g.arrayMethodCall(e); ok {
			return arrayCall
		}
		args := make([]string, 0, len(e.Args))
		for _, arg := range e.Args {
			args = append(args, g.expr(arg))
		}
		return fmt.Sprintf("%s(%s)", g.expr(e.Callee), strings.Join(args, ", "))
	case *ir.LambdaExpr:
		return "/* lambda */"
	case *ir.IndexExpr:
		return fmt.Sprintf("%s[%s]", g.expr(e.Receiver), g.expr(e.Index))
	case *ir.SelectorExpr:
		if at, ok := e.Receiver.(*ir.AtExpr); ok {
			if fn, ok := g.file.Stdlib.Function(at.Name, e.Name); ok && fn.Go != nil && fn.Go.Symbol != "" {
				return fn.Go.Symbol
			}
		}
		return g.expr(e.Receiver) + "." + mangleIdent(e.Name)
	case *ir.ArrayLiteral:
		elems := make([]string, 0, len(e.Elements))
		for _, elem := range e.Elements {
			elems = append(elems, g.expr(elem))
		}
		elemType := checker.Unknown
		if elem, ok := checker.ArrayElement(e.ResultType()); ok {
			elemType = elem
		}
		return fmt.Sprintf("[]%s{%s}", goType(elemType), strings.Join(elems, ", "))
	case *ir.StructLiteral:
		fields := make([]string, 0, len(e.Fields))
		for _, field := range e.Fields {
			fields = append(fields, fmt.Sprintf("%s: %s", mangleIdent(field.Name), g.expr(field.Value)))
		}
		return fmt.Sprintf("%s{%s}", mangleIdent(e.TypeName), strings.Join(fields, ", "))
	case *ir.AtExpr:
		return e.Name
	case *ir.ThisExpr:
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
