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
	if !ok || (sel.Name != "each" && sel.Name != "forEach") || len(call.Args) != 1 {
		return "", false
	}
	if _, ok := checker.ArrayElement(sel.Receiver.ResultType()); !ok {
		return "", false
	}
	fn, ok := g.file.Stdlib.Function("array", sel.Name)
	if !ok {
		return "", false
	}
	if sel.Name == "each" && fn.Body == nil && fn.Intrinsic != "array.each" {
		return "", false
	}
	if sel.Name == "forEach" && fn.Intrinsic != "array.each" {
		return "", false
	}
	lambda, ok := call.Args[0].(*ir.LambdaExpr)
	if !ok || len(lambda.Params) == 0 || len(lambda.Params) > 3 {
		return "", false
	}
	valueParam := mangleIdent(lambda.Params[0])
	indexParam := "_"
	if len(lambda.Params) >= 2 {
		indexParam = mangleIdent(lambda.Params[1])
	}
	var b strings.Builder
	receiver := g.expr(sel.Receiver)
	b.WriteString(fmt.Sprintf("for %s, %s := range %s {\n", indexParam, valueParam, receiver))
	b.WriteString(fmt.Sprintf("\t_ = %s\n", valueParam))
	if indexParam != "_" {
		b.WriteString(fmt.Sprintf("\t_ = %s\n", indexParam))
	}
	if len(lambda.Params) >= 3 {
		arrayParam := mangleIdent(lambda.Params[2])
		b.WriteString(fmt.Sprintf("\t%s := %s\n", arrayParam, receiver))
		b.WriteString(fmt.Sprintf("\t_ = %s\n", arrayParam))
	}
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
	if fn.Body != nil {
		return g.arrayBodyExpr(fn, g.expr(sel.Receiver), call), true
	}
	if fn.Intrinsic == "array.each" {
		return g.arrayForEachExpr(call, g.expr(sel.Receiver)), true
	}
	if fn.Intrinsic == "array.map" {
		return g.arrayMapExpr(call, g.expr(sel.Receiver)), true
	}
	args := make([]string, 0, len(call.Args))
	for _, arg := range call.Args {
		args = append(args, g.expr(arg))
	}
	return g.arrayFunctionExpr(fn, g.expr(sel.Receiver), args, call.ResultType(), sel.Receiver.ResultType()), true
}

func (g *generator) arrayFunctionExpr(fn *stdlib.Function, receiver string, args []string, resultType checker.Type, receiverType checker.Type) string {
	switch fn.Intrinsic {
	case "array.len":
		return fmt.Sprintf("len(%s)", receiver)
	case "array.push":
		if len(args) != 1 {
			return "/* invalid array.push */"
		}
		return fmt.Sprintf("func() int { %s = append(%s, %s); return len(%s) }()", receiver, receiver, args[0], receiver)
	case "array.set":
		if len(args) != 2 {
			return "/* invalid array.set */"
		}
		return fmt.Sprintf("func() %s { value := %s; %s[%s] = value; return value }()", goType(resultType), args[1], receiver, args[0])
	case "array.pop":
		return fmt.Sprintf("func() %s { value := %s[len(%s)-1]; %s = %s[:len(%s)-1]; return value }()", goType(resultType), receiver, receiver, receiver, receiver, receiver)
	case "array.first":
		return fmt.Sprintf("%s[0]", receiver)
	case "array.last":
		return fmt.Sprintf("%s[len(%s)-1]", receiver, receiver)
	case "array.slice":
		if len(args) != 2 {
			return "/* invalid array.slice */"
		}
		return fmt.Sprintf("append([]%s{}, %s[%s:%s]...)", goType(arrayElemOrUnknown(resultType, receiverType)), receiver, args[0], args[1])
	case "array.clone":
		return fmt.Sprintf("append([]%s{}, %s...)", goType(arrayElemOrUnknown(resultType, receiverType)), receiver)
	case "array.reverse":
		elemType := goType(arrayElemOrUnknown(resultType, receiverType))
		return fmt.Sprintf("func() []%s { out := append([]%s{}, %s...); for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 { out[i], out[j] = out[j], out[i] }; return out }()", elemType, elemType, receiver)
	case "array.contains":
		if len(args) != 1 {
			return "/* invalid array.contains */"
		}
		return fmt.Sprintf("func() bool { for _, item := range %s { if item == %s { return true } }; return false }()", receiver, args[0])
	case "array.get", "array.at":
		if len(args) != 1 {
			return "/* invalid array.at */"
		}
		return fmt.Sprintf("%s[%s]", receiver, args[0])
	case "array.each":
		return "/* array.forEach is only valid as a statement */"
	}
	if fn.Body != nil {
		return g.stdlibBodyExpr(ir.LowerExpr(fn.Body, nil), receiver)
	}
	return g.unsupportedIntrinsic(fn, resultType)
}

func (g *generator) arrayMapExpr(call *ir.CallExpr, receiver string) string {
	if len(call.Args) != 1 {
		return "/* invalid array.map */"
	}
	lambda, ok := call.Args[0].(*ir.LambdaExpr)
	if !ok || len(lambda.Params) == 0 || len(lambda.Params) > 3 {
		return "/* invalid array.map */"
	}
	elemType := checker.Unknown
	if elem, ok := checker.ArrayElement(call.ResultType()); ok {
		elemType = elem
	}
	valueParam := mangleIdent(lambda.Params[0])
	indexParam := "_"
	if len(lambda.Params) >= 2 {
		indexParam = mangleIdent(lambda.Params[1])
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("func() []%s {\n", goType(elemType)))
	b.WriteString(fmt.Sprintf("\t%s := make([]%s, 0, len(%s))\n", mangleIdent("result"), goType(elemType), receiver))
	b.WriteString(fmt.Sprintf("\tfor %s, %s := range %s {\n", indexParam, valueParam, receiver))
	b.WriteString(fmt.Sprintf("\t\t_ = %s\n", valueParam))
	if indexParam != "_" {
		b.WriteString(fmt.Sprintf("\t\t_ = %s\n", indexParam))
	}
	if len(lambda.Params) >= 3 {
		arrayParam := mangleIdent(lambda.Params[2])
		b.WriteString(fmt.Sprintf("\t\t%s := %s\n", arrayParam, receiver))
		b.WriteString(fmt.Sprintf("\t\t_ = %s\n", arrayParam))
	}
	b.WriteString(fmt.Sprintf("\t\t%s = append(%s, %s)\n", mangleIdent("result"), mangleIdent("result"), g.expr(lambda.Body)))
	b.WriteString("\t}\n")
	b.WriteString(fmt.Sprintf("\treturn %s\n", mangleIdent("result")))
	b.WriteString("}()")
	return b.String()
}

func (g *generator) arrayForEachExpr(call *ir.CallExpr, receiver string) string {
	if len(call.Args) != 1 {
		return "/* invalid array.forEach */"
	}
	lambda, ok := call.Args[0].(*ir.LambdaExpr)
	if !ok || len(lambda.Params) == 0 || len(lambda.Params) > 3 {
		return "/* invalid array.forEach */"
	}
	valueParam := mangleIdent(lambda.Params[0])
	indexParam := "_"
	if len(lambda.Params) >= 2 {
		indexParam = mangleIdent(lambda.Params[1])
	}
	var b strings.Builder
	b.WriteString("func() {\n")
	b.WriteString(fmt.Sprintf("\tfor %s, %s := range %s {\n", indexParam, valueParam, receiver))
	b.WriteString(fmt.Sprintf("\t\t_ = %s\n", valueParam))
	if indexParam != "_" {
		b.WriteString(fmt.Sprintf("\t\t_ = %s\n", indexParam))
	}
	if len(lambda.Params) >= 3 {
		arrayParam := mangleIdent(lambda.Params[2])
		b.WriteString(fmt.Sprintf("\t\t%s := %s\n", arrayParam, receiver))
		b.WriteString(fmt.Sprintf("\t\t_ = %s\n", arrayParam))
	}
	b.WriteString("\t\t")
	b.WriteString(g.expr(lambda.Body))
	b.WriteByte('\n')
	b.WriteString("\t}\n")
	b.WriteString("}()")
	return b.String()
}

func arrayElemOrUnknown(resultType checker.Type, receiverType checker.Type) checker.Type {
	if elem, ok := checker.ArrayElement(resultType); ok {
		return elem
	}
	if elem, ok := checker.ArrayElement(receiverType); ok {
		return elem
	}
	return checker.Unknown
}

func (g *generator) arrayBodyExpr(fn *stdlib.Function, receiver string, call *ir.CallExpr) string {
	if len(call.Args) != len(fn.ParamNames) {
		return fmt.Sprintf("/* invalid array.%s */", fn.Name)
	}
	ctx := &stdlibContext{
		g:       g,
		this:    receiver,
		vars:    map[string]string{},
		lambdas: map[string]*ir.LambdaExpr{},
	}
	for idx, name := range fn.ParamNames {
		if lambda, ok := call.Args[idx].(*ir.LambdaExpr); ok {
			ctx.lambdas[name] = lambda
			continue
		}
		ctx.vars[name] = g.expr(call.Args[idx])
	}
	return ctx.expr(ir.LowerExpr(fn.Body, nil), call.ResultType())
}

type stdlibContext struct {
	g       *generator
	this    string
	vars    map[string]string
	lambdas map[string]*ir.LambdaExpr
}

func (c *stdlibContext) child() *stdlibContext {
	vars := make(map[string]string, len(c.vars))
	for name, value := range c.vars {
		vars[name] = value
	}
	return &stdlibContext{
		g:       c.g,
		this:    c.this,
		vars:    vars,
		lambdas: cloneLambdaMap(c.lambdas),
	}
}

func (c *stdlibContext) expr(expr ir.Expr, expected checker.Type) string {
	switch e := expr.(type) {
	case *ir.ThisExpr:
		return c.this
	case *ir.Identifier:
		if value, ok := c.vars[e.Name]; ok {
			return value
		}
		return mangleIdent(e.Name)
	case *ir.IntegerLiteral:
		return strconv.Itoa(e.Value)
	case *ir.DoubleLiteral:
		if e.Raw != "" {
			return e.Raw
		}
		return strconv.FormatFloat(e.Value, 'f', -1, 64)
	case *ir.BigIntLiteral:
		return fmt.Sprintf("runeBigInt(%q)", e.Value)
	case *ir.StringLiteral:
		return strconv.Quote(e.Value)
	case *ir.BoolLiteral:
		if e.Value {
			return "true"
		}
		return "false"
	case *ir.NullLiteral:
		return "any(nil)"
	case *ir.UnaryExpr:
		return e.Op.String() + c.expr(e.Expr, checker.Unknown)
	case *ir.BinaryExpr:
		return fmt.Sprintf("%s %s %s", c.expr(e.Left, checker.Unknown), e.Op, c.expr(e.Right, checker.Unknown))
	case *ir.ArrayLiteral:
		elemType := checker.Unknown
		if elem, ok := checker.ArrayElement(expected); ok {
			elemType = elem
		} else if elem, ok := checker.ArrayElement(e.ResultType()); ok {
			elemType = elem
		}
		elems := make([]string, 0, len(e.Elements))
		for _, elem := range e.Elements {
			elems = append(elems, c.expr(elem, elem.ResultType()))
		}
		return fmt.Sprintf("[]%s{%s}", goType(elemType), strings.Join(elems, ", "))
	case *ir.CallExpr:
		if expr, ok := c.arrayCallExpr(e, expected); ok {
			return expr
		}
		if ident, ok := e.Callee.(*ir.Identifier); ok {
			if lambda, ok := c.lambdas[ident.Name]; ok {
				child := c.child()
				for idx, param := range lambda.Params {
					if idx < len(e.Args) {
						child.vars[param] = c.expr(e.Args[idx], e.Args[idx].ResultType())
					}
				}
				return child.expr(lambda.Body, e.ResultType())
			}
		}
		args := make([]string, 0, len(e.Args))
		for _, arg := range e.Args {
			args = append(args, c.expr(arg, arg.ResultType()))
		}
		return fmt.Sprintf("%s(%s)", c.expr(e.Callee, checker.Unknown), strings.Join(args, ", "))
	case *ir.SelectorExpr:
		return c.expr(e.Receiver, e.Receiver.ResultType()) + "." + mangleIdent(e.Name)
	case *ir.IndexExpr:
		return fmt.Sprintf("%s[%s]", c.expr(e.Receiver, e.Receiver.ResultType()), c.expr(e.Index, checker.Int))
	case *ir.BlockExpr:
		return c.blockExpr(e, expected)
	default:
		return c.g.expr(expr)
	}
}

func (c *stdlibContext) arrayCallExpr(call *ir.CallExpr, expected checker.Type) (string, bool) {
	sel, ok := call.Callee.(*ir.SelectorExpr)
	if !ok {
		return "", false
	}
	fn, ok := c.g.file.Stdlib.Function("array", sel.Name)
	if !ok {
		return "", false
	}
	receiver := c.expr(sel.Receiver, sel.Receiver.ResultType())
	args := make([]string, 0, len(call.Args))
	for _, arg := range call.Args {
		args = append(args, c.expr(arg, arg.ResultType()))
	}
	if fn.Body != nil {
		return c.arrayBodyExpr(fn, receiver, call.Args, expected), true
	}
	return c.g.arrayFunctionExpr(fn, receiver, args, call.ResultType(), sel.Receiver.ResultType()), true
}

func (c *stdlibContext) arrayBodyExpr(fn *stdlib.Function, receiver string, args []ir.Expr, expected checker.Type) string {
	if len(args) != len(fn.ParamNames) {
		return fmt.Sprintf("/* invalid array.%s */", fn.Name)
	}
	child := &stdlibContext{
		g:       c.g,
		this:    receiver,
		vars:    map[string]string{},
		lambdas: cloneLambdaMap(c.lambdas),
	}
	for idx, name := range fn.ParamNames {
		if lambda, ok := args[idx].(*ir.LambdaExpr); ok {
			child.lambdas[name] = lambda
			continue
		}
		child.vars[name] = c.expr(args[idx], args[idx].ResultType())
	}
	return child.expr(ir.LowerExpr(fn.Body, nil), expected)
}

func (c *stdlibContext) blockExpr(block *ir.BlockExpr, resultType checker.Type) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("func() %s {\n", goType(resultType)))
	resultName := returnedIdentifier(block)
	for idx, stmt := range block.Statements {
		last := idx == len(block.Statements)-1
		for _, line := range c.stmt(stmt, last, resultType, resultName) {
			b.WriteByte('\t')
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	b.WriteString("}()")
	return b.String()
}

func (c *stdlibContext) stmt(stmt ir.Stmt, last bool, resultType checker.Type, resultName string) []string {
	switch s := stmt.(type) {
	case *ir.LetStmt:
		c.vars[s.Name] = mangleIdent(s.Name)
		expected := s.Value.ResultType()
		if s.Name == resultName {
			expected = resultType
		}
		return []string{fmt.Sprintf("%s := %s", mangleIdent(s.Name), c.expr(s.Value, expected))}
	case *ir.AssignStmt:
		return []string{fmt.Sprintf("%s = %s", mangleIdent(s.Name), c.expr(s.Value, s.Value.ResultType()))}
	case *ir.ExprStmt:
		if last && resultType != checker.Void {
			return []string{"return " + c.expr(s.Expr, resultType)}
		}
		if stmt, ok := c.callStmt(s.Expr); ok {
			return stmt
		}
		return []string{c.expr(s.Expr, s.Expr.ResultType())}
	default:
		return nil
	}
}

func (c *stdlibContext) callStmt(expr ir.Expr) ([]string, bool) {
	call, ok := expr.(*ir.CallExpr)
	if !ok {
		return nil, false
	}
	sel, ok := call.Callee.(*ir.SelectorExpr)
	if !ok {
		return nil, false
	}
	if sel.Name == "push" && len(call.Args) == 1 {
		if fn, ok := c.g.file.Stdlib.Function("array", "push"); ok && fn.Intrinsic == "array.push" {
			receiver := c.expr(sel.Receiver, sel.Receiver.ResultType())
			return []string{fmt.Sprintf("%s = append(%s, %s)", receiver, receiver, c.expr(call.Args[0], call.Args[0].ResultType()))}, true
		}
	}
	if (sel.Name == "each" || sel.Name == "forEach") && len(call.Args) == 1 {
		if fn, ok := c.g.file.Stdlib.Function("array", sel.Name); ok && (fn.Intrinsic == "array.each") {
			lambda, ok := call.Args[0].(*ir.LambdaExpr)
			if !ok || len(lambda.Params) != 1 {
				return []string{"/* invalid array.each */"}, true
			}
			child := c.child()
			param := mangleIdent(lambda.Params[0])
			child.vars[lambda.Params[0]] = param
			lines := []string{fmt.Sprintf("for _, %s := range %s {", param, c.expr(sel.Receiver, sel.Receiver.ResultType()))}
			if stmt, ok := child.callStmt(lambda.Body); ok {
				for _, line := range stmt {
					lines = append(lines, "\t"+line)
				}
			} else {
				lines = append(lines, "\t"+child.expr(lambda.Body, lambda.Body.ResultType()))
			}
			lines = append(lines, "}")
			return lines, true
		}
	}
	return nil, false
}

func returnedIdentifier(block *ir.BlockExpr) string {
	if len(block.Statements) == 0 {
		return ""
	}
	stmt, ok := block.Statements[len(block.Statements)-1].(*ir.ExprStmt)
	if !ok {
		return ""
	}
	ident, ok := stmt.Expr.(*ir.Identifier)
	if !ok {
		return ""
	}
	return ident.Name
}

func cloneLambdaMap(in map[string]*ir.LambdaExpr) map[string]*ir.LambdaExpr {
	out := make(map[string]*ir.LambdaExpr, len(in))
	for name, lambda := range in {
		out[name] = lambda
	}
	return out
}

func (g *generator) stdlibBodyExpr(expr ir.Expr, this string) string {
	switch e := expr.(type) {
	case *ir.ThisExpr:
		return this
	case *ir.Identifier:
		return mangleIdent(e.Name)
	case *ir.IntegerLiteral:
		return strconv.Itoa(e.Value)
	case *ir.DoubleLiteral:
		if e.Raw != "" {
			return e.Raw
		}
		return strconv.FormatFloat(e.Value, 'f', -1, 64)
	case *ir.BigIntLiteral:
		return fmt.Sprintf("runeBigInt(%q)", e.Value)
	case *ir.StringLiteral:
		return strconv.Quote(e.Value)
	case *ir.BoolLiteral:
		if e.Value {
			return "true"
		}
		return "false"
	case *ir.NullLiteral:
		return "any(nil)"
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
					return g.arrayFunctionExpr(fn, this, args, e.ResultType(), sel.Receiver.ResultType())
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
