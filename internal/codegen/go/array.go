package gocodegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func (g *generator) arrayPushStmt(expr ir.Expr) (string, bool) {
	call, ok := expr.(*ir.CallExpr)
	if !ok {
		return "", false
	}
	sel, ok := call.Callee.(*ir.SelectorExpr)
	if !ok || sel.Name != "push" {
		return "", false
	}
	if _, ok := checker.ArrayElement(sel.Receiver.ResultType()); !ok {
		return "", false
	}
	fn, ok := g.file.Stdlib.Function("array", "push")
	if !ok || fn.Intrinsic != "array.push" || len(call.Args) != 1 {
		return "", false
	}
	receiver := g.expr(sel.Receiver)
	return fmt.Sprintf("%s = append(%s, %s)", receiver, receiver, g.expr(call.Args[0])), true
}

func (g *generator) arrayEachStmt(expr ir.Expr) (string, bool) {
	call, ok := expr.(*ir.CallExpr)
	if !ok {
		return "", false
	}
	sel, ok := call.Callee.(*ir.SelectorExpr)
	if !ok || sel.Name != "each" || len(call.Args) != 1 {
		return "", false
	}
	if _, ok := checker.ArrayElement(sel.Receiver.ResultType()); !ok {
		return "", false
	}
	fn, ok := g.file.Stdlib.Function("array", "each")
	if !ok || fn.Intrinsic != "array.each" {
		return "", false
	}
	lambda, ok := call.Args[0].(*ir.LambdaExpr)
	if !ok || len(lambda.Params) != 1 {
		return "", false
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("for _, %s := range %s {\n", mangleIdent(lambda.Params[0]), g.expr(sel.Receiver)))
	b.WriteByte('\t')
	b.WriteString(g.expr(lambda.Body))
	b.WriteByte('\n')
	b.WriteByte('}')
	return b.String(), true
}

func (g *generator) arrayMethodCall(call *ir.CallExpr) (string, bool) {
	sel, ok := call.Callee.(*ir.SelectorExpr)
	if !ok {
		return "", false
	}
	if _, ok := checker.ArrayElement(sel.Receiver.ResultType()); !ok {
		return "", false
	}
	fn, ok := g.file.Stdlib.Function("array", sel.Name)
	if !ok {
		return "", false
	}
	if fn.Intrinsic == "array.map" {
		lambda, ok := singleLambdaArg(call)
		if !ok {
			return "/* invalid array.map */", true
		}
		return g.arrayMapExpr(g.expr(sel.Receiver), lambda, call), true
	}
	if fn.Intrinsic == "array.each" {
		return "/* array.each is only valid as a statement */", true
	}
	args := make([]string, 0, len(call.Args))
	for _, arg := range call.Args {
		args = append(args, g.expr(arg))
	}
	return g.arrayFunctionExpr(fn, g.expr(sel.Receiver), args), true
}

func (g *generator) arrayFunctionExpr(fn *stdlib.Function, receiver string, args []string) string {
	switch fn.Intrinsic {
	case "array.len":
		return fmt.Sprintf("len(%s)", receiver)
	case "array.push":
		if len(args) != 1 {
			return "/* invalid array.push */"
		}
		return fmt.Sprintf("append(%s, %s)", receiver, args[0])
	case "array.get":
		if len(args) != 1 {
			return "/* invalid array.get */"
		}
		return fmt.Sprintf("%s[%s]", receiver, args[0])
	case "array.each":
		return "/* array.each is only valid as a statement */"
	case "array.map":
		return "/* invalid array.map */"
	}
	if fn.Body != nil {
		return g.stdlibBodyExpr(ir.LowerExpr(fn.Body, nil), receiver)
	}
	return "/* unsupported array method */"
}

func (g *generator) arrayMapExpr(receiver string, lambda *ir.LambdaExpr, call *ir.CallExpr) string {
	if len(lambda.Params) != 1 {
		return "/* invalid array.map */"
	}
	resultType := checker.Unknown
	if elem, ok := checker.ArrayElement(call.ResultType()); ok {
		resultType = elem
	}
	out := "__rune_map_out"
	param := mangleIdent(lambda.Params[0])
	return fmt.Sprintf(`func() []%s {
%s := make([]%s, 0, len(%s))
for _, %s := range %s {
%s = append(%s, %s)
}
return %s
}()`, goType(resultType), out, goType(resultType), receiver, param, receiver, out, out, g.expr(lambda.Body), out)
}

func singleLambdaArg(call *ir.CallExpr) (*ir.LambdaExpr, bool) {
	if len(call.Args) != 1 {
		return nil, false
	}
	lambda, ok := call.Args[0].(*ir.LambdaExpr)
	return lambda, ok
}

func (g *generator) stdlibBodyExpr(expr ir.Expr, this string) string {
	switch e := expr.(type) {
	case *ir.ThisExpr:
		return this
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
		return e.Op.String() + g.stdlibBodyExpr(e.Expr, this)
	case *ir.BinaryExpr:
		prec := goPrecedence(e.Op)
		left := g.stdlibBodyExpr(e.Left, this)
		right := g.stdlibBodyExpr(e.Right, this)
		s := fmt.Sprintf("%s %s %s", left, e.Op, right)
		if prec < 1 {
			return "(" + s + ")"
		}
		return s
	case *ir.CallExpr:
		if sel, ok := e.Callee.(*ir.SelectorExpr); ok {
			if _, ok := sel.Receiver.(*ir.ThisExpr); ok {
				if fn, ok := g.file.Stdlib.Function("array", sel.Name); ok {
					args := make([]string, 0, len(e.Args))
					for _, arg := range e.Args {
						args = append(args, g.stdlibBodyExpr(arg, this))
					}
					return g.arrayFunctionExpr(fn, this, args)
				}
			}
		}
		args := make([]string, 0, len(e.Args))
		for _, arg := range e.Args {
			args = append(args, g.stdlibBodyExpr(arg, this))
		}
		return fmt.Sprintf("%s(%s)", g.stdlibBodyExpr(e.Callee, this), strings.Join(args, ", "))
	case *ir.SelectorExpr:
		return g.stdlibBodyExpr(e.Receiver, this) + "." + mangleIdent(e.Name)
	case *ir.IndexExpr:
		return fmt.Sprintf("%s[%s]", g.stdlibBodyExpr(e.Receiver, this), g.stdlibBodyExpr(e.Index, this))
	default:
		return g.expr(expr)
	}
}
