package tscodegen

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
		if g.isSignal(e.Name) {
			return mangleIdent(e.Name) + ".get()"
		}
		return mangleIdent(e.Name)
	case *ir.AtExpr:
		return "@" + e.Name
	case *ir.ThisExpr:
		if name := g.currentThisName(); name != "" {
			return name
		}
		return mangleIdent("this")
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
	case *ir.PostfixExpr:
		return g.postfixExpr(e)
	case *ir.BinaryExpr:
		prec := tsPrecedence(e.Op)
		op := tsBinaryOp(e.Op)
		s := fmt.Sprintf("%s %s %s", g.exprPrec(e.Left, prec), op, g.exprPrec(e.Right, prec+1))
		if prec < parentPrec {
			return "(" + s + ")"
		}
		return s
	case *ir.CallExpr:
		if call, ok := g.stdlibCall(e); ok {
			return call
		}
		if call, ok := g.arrayMethodCall(e); ok {
			return call
		}
		if call, ok := g.methodCall(e); ok {
			return call
		}
		args := make([]string, 0, len(e.Args))
		for _, arg := range e.Args {
			args = append(args, g.expr(arg))
		}
		return fmt.Sprintf("%s(%s)", g.expr(e.Callee), strings.Join(args, ", "))
	case *ir.LambdaExpr:
		return g.lambda(e)
	case *ir.SelectorExpr:
		if at, ok := e.Receiver.(*ir.AtExpr); ok {
			return "@" + at.Name + "." + e.Name
		}
		return g.expr(e.Receiver) + "." + mangleIdent(e.Name)
	case *ir.IndexExpr:
		return fmt.Sprintf("%s[%s]", g.expr(e.Receiver), g.expr(e.Index))
	case *ir.ArrayLiteral:
		elems := make([]string, 0, len(e.Elements))
		for _, elem := range e.Elements {
			elems = append(elems, g.expr(elem))
		}
		return "[" + strings.Join(elems, ", ") + "]"
	case *ir.ReactiveLiteral:
		return g.reactiveLiteral(e)
	case *ir.StructLiteral:
		fields := make([]string, 0, len(e.Fields))
		for _, field := range e.Fields {
			fields = append(fields, fmt.Sprintf("%s: %s", mangleIdent(field.Name), g.expr(field.Value)))
		}
		return "{" + strings.Join(fields, ", ") + "}"
	case *ir.AnonymousObjectLiteral:
		fields := make([]string, 0, len(e.Fields))
		for _, field := range e.Fields {
			fields = append(fields, fmt.Sprintf("%s: %s", mangleIdent(field.Name), g.expr(field.Value)))
		}
		return "{" + strings.Join(fields, ", ") + "}"
	case *ir.MatchExpr:
		return g.matchExpr(e)
	case *ir.WatchExpr:
		return g.watchExpr(e)
	case *ir.XMLElement:
		return g.xmlExpr(e)
	default:
		return "undefined"
	}
}

func (g *generator) stmtExpr(expr ir.Expr) string {
	switch e := expr.(type) {
	case *ir.PostfixExpr:
		target, ok := e.Expr.(*ir.Identifier)
		if !ok || e.Op != lexer.PlusPlus {
			return g.expr(e)
		}
		name := mangleIdent(target.Name)
		if g.isSignal(target.Name) {
			return fmt.Sprintf("%s.set(%s.get() + 1)", name, name)
		}
		return name + "++"
	default:
		return g.expr(expr)
	}
}

func (g *generator) reactiveLiteral(lit *ir.ReactiveLiteral) string {
	switch value := lit.Value.(type) {
	case *ir.ArrayLiteral:
		return "runeReactiveArray(" + g.expr(value) + ")"
	case *ir.AnonymousObjectLiteral:
		return "runeReactiveObject(" + g.expr(value) + ")"
	default:
		return g.expr(lit.Value)
	}
}

func (g *generator) postfixExpr(expr *ir.PostfixExpr) string {
	target, ok := expr.Expr.(*ir.Identifier)
	if !ok || expr.Op != lexer.PlusPlus {
		return "undefined"
	}
	name := mangleIdent(target.Name)
	if g.isSignal(target.Name) {
		return fmt.Sprintf("(() => { const old = %s.get(); %s.set(old + 1); return old; })()", name, name)
	}
	return fmt.Sprintf("(() => { const old = %s; %s++; return old; })()", name, name)
}

func (g *generator) stdlibCall(call *ir.CallExpr) (string, bool) {
	sel, ok := call.Callee.(*ir.SelectorExpr)
	if !ok {
		return "", false
	}
	at, ok := sel.Receiver.(*ir.AtExpr)
	if !ok {
		return "", false
	}
	if at.Name == "go" {
		return "undefined /* TypeScript backend does not support @go */", true
	}
	if at.Name != "io" {
		return "", false
	}
	args := make([]string, 0, len(call.Args))
	for _, arg := range call.Args {
		args = append(args, g.expr(arg))
	}
	switch sel.Name {
	case "print", "println":
		return "console.log(" + strings.Join(args, ", ") + ")", true
	case "printf":
		return "console.log(" + strings.Join(args, ", ") + ")", true
	default:
		return "", false
	}
}

func (g *generator) arrayMethodCall(call *ir.CallExpr) (string, bool) {
	sel, ok := call.Callee.(*ir.SelectorExpr)
	if !ok {
		return "", false
	}
	if _, ok := checker.ArrayElement(sel.Receiver.ResultType()); !ok {
		return "", false
	}
	receiver := g.expr(sel.Receiver)
	args := make([]string, 0, len(call.Args))
	for _, arg := range call.Args {
		args = append(args, g.expr(arg))
	}
	switch sel.Name {
	case "length":
		return receiver + ".length", true
	case "isEmpty":
		return receiver + ".length === 0", true
	case "push":
		return fmt.Sprintf("%s.push(%s)", receiver, strings.Join(args, ", ")), true
	case "each", "forEach":
		if len(args) != 1 {
			return "undefined", true
		}
		return fmt.Sprintf("%s.forEach(%s)", receiver, args[0]), true
	case "map":
		if len(args) != 1 {
			return "undefined", true
		}
		return fmt.Sprintf("%s.map(%s)", receiver, args[0]), true
	case "at":
		if len(args) != 1 {
			return "undefined", true
		}
		return fmt.Sprintf("%s.at(%s)", receiver, args[0]), true
	default:
		return "", false
	}
}

func (g *generator) methodCall(call *ir.CallExpr) (string, bool) {
	sel, ok := call.Callee.(*ir.SelectorExpr)
	if !ok {
		return "", false
	}
	typeName := baseTypeName(sel.Receiver.ResultType())
	if typeName == "" || strings.HasPrefix(typeName, "{") {
		return "", false
	}
	var typ *ir.StructType
	for _, candidate := range g.file.Types {
		if candidate.Name == typeName {
			typ = candidate
			break
		}
	}
	if typ == nil {
		return "", false
	}
	for _, method := range typ.Methods {
		if method.Name != sel.Name {
			continue
		}
		args := []string{g.expr(sel.Receiver)}
		for _, arg := range call.Args {
			args = append(args, g.expr(arg))
		}
		return fmt.Sprintf("%s(%s)", mangleMethod(typeName, sel.Name), strings.Join(args, ", ")), true
	}
	return "", false
}

func (g *generator) matchExpr(match *ir.MatchExpr) string {
	ret := tsType(match.ResultType())
	subject := g.expr(match.Subject)
	var b strings.Builder
	b.WriteString("((): ")
	b.WriteString(ret)
	b.WriteString(" => { ")
	hasDefault := false
	for i, branch := range match.Branches {
		if _, ok := branch.Pattern.(*ir.WildcardPattern); ok {
			hasDefault = true
			if i == 0 {
				b.WriteString("{ ")
			} else {
				b.WriteString("else { ")
			}
		} else {
			if i == 0 {
				b.WriteString("if (")
			} else {
				b.WriteString("else if (")
			}
			b.WriteString(g.patternCondition(subject, branch.Pattern))
			b.WriteString(") { ")
		}
		if match.ResultType() == checker.Void {
			b.WriteString(g.stmtExpr(branch.Expr))
			b.WriteString("; return; } ")
		} else {
			b.WriteString("return ")
			b.WriteString(g.expr(branch.Expr))
			b.WriteString("; } ")
		}
	}
	if !hasDefault && match.ResultType() != checker.Void {
		b.WriteString("return ")
		b.WriteString(zeroValue(match.ResultType()))
		b.WriteString("; ")
	}
	b.WriteString("})()")
	return b.String()
}

func (g *generator) watchExpr(watch *ir.WatchExpr) string {
	target, ok := watch.Target.(*ir.Identifier)
	if !ok || !g.isSignal(target.Name) {
		return "undefined"
	}
	handler, ok := watch.Handler.(*ir.LambdaExpr)
	if !ok {
		return "undefined"
	}
	params := handler.Params
	tsParams := make([]string, 0, len(params))
	if len(params) == 0 {
		tsParams = append(tsParams, "_old", "_new")
	} else {
		for _, name := range params {
			tsParams = append(tsParams, mangleIdent(name))
		}
	}
	return fmt.Sprintf("%s.watch((%s) => { %s; })", mangleIdent(target.Name), strings.Join(tsParams, ", "), g.lambdaBody(handler, checker.Void))
}

func (g *generator) lambda(lambda *ir.LambdaExpr) string {
	params, ret, ok := parseTSFuncType(string(lambda.ResultType()))
	if !ok {
		params = make([]string, len(lambda.Params))
		for i := range params {
			params[i] = string(checker.Unknown)
		}
		ret = string(checker.Void)
	}
	tsParams := make([]string, 0, len(lambda.Params))
	for i, name := range lambda.Params {
		typ := checker.Unknown
		if i < len(params) {
			typ = checker.Type(params[i])
		}
		tsParams = append(tsParams, fmt.Sprintf("%s: %s", mangleIdent(name), tsType(typ)))
	}
	if block, ok := lambda.Body.(*ir.BlockExpr); ok {
		return "(" + strings.Join(tsParams, ", ") + "): " + tsType(checker.Type(ret)) + " => { " + g.blockInline(block, checker.Type(ret)) + " }"
	}
	if checker.Type(ret) == checker.Void {
		return "(" + strings.Join(tsParams, ", ") + ") => { " + g.stmtExpr(lambda.Body) + "; }"
	}
	return "(" + strings.Join(tsParams, ", ") + "): " + tsType(checker.Type(ret)) + " => " + g.arrowExprBody(lambda.Body)
}

func (g *generator) arrowExprBody(expr ir.Expr) string {
	body := g.expr(expr)
	switch expr.(type) {
	case *ir.StructLiteral, *ir.AnonymousObjectLiteral:
		return "(" + body + ")"
	default:
		return body
	}
}

func (g *generator) lambdaBody(lambda *ir.LambdaExpr, ret checker.Type) string {
	if block, ok := lambda.Body.(*ir.BlockExpr); ok {
		return g.blockInline(block, ret)
	}
	if ret == checker.Void {
		return g.stmtExpr(lambda.Body)
	}
	return "return " + g.expr(lambda.Body)
}

func (g *generator) blockInline(block *ir.BlockExpr, ret checker.Type) string {
	parts := make([]string, 0, len(block.Statements))
	for i, stmt := range block.Statements {
		last := i == len(block.Statements)-1
		switch s := stmt.(type) {
		case *ir.LetStmt:
			kind := "const"
			if s.Mutable {
				kind = "let"
			}
			value := g.expr(s.Value)
			if _, ok := s.Value.(*ir.ReactiveLiteral); ok {
				parts = append(parts, fmt.Sprintf("%s %s = %s", kind, mangleIdent(s.Name), value))
				continue
			}
			if _, ok := s.Value.(*ir.AnonymousObjectLiteral); ok {
				value = g.withThisName(mangleIdent(s.Name), func() string {
					return g.expr(s.Value)
				})
			}
			parts = append(parts, fmt.Sprintf("%s %s = %s", kind, mangleIdent(s.Name), value))
		case *ir.AssignStmt:
			if g.isSignal(s.Name) {
				parts = append(parts, fmt.Sprintf("%s.set(%s)", mangleIdent(s.Name), g.expr(s.Value)))
			} else {
				parts = append(parts, fmt.Sprintf("%s = %s", mangleIdent(s.Name), g.expr(s.Value)))
			}
		case *ir.ExprStmt:
			if last && ret != checker.Void {
				parts = append(parts, "return "+g.expr(s.Expr))
			} else {
				parts = append(parts, g.stmtExpr(s.Expr))
			}
		}
	}
	return strings.Join(parts, "; ")
}

func (g *generator) withThisName(name string, render func() string) string {
	g.thisNames = append(g.thisNames, name)
	defer func() {
		g.thisNames = g.thisNames[:len(g.thisNames)-1]
	}()
	return render()
}

func (g *generator) currentThisName() string {
	if len(g.thisNames) == 0 {
		return ""
	}
	return g.thisNames[len(g.thisNames)-1]
}

func (g *generator) pushSignalScope() {
	g.signals = append(g.signals, map[string]checker.Type{})
}

func (g *generator) popSignalScope() {
	g.signals = g.signals[:len(g.signals)-1]
}

func (g *generator) addSignal(name string, typ checker.Type) {
	if len(g.signals) == 0 {
		g.pushSignalScope()
	}
	g.signals[len(g.signals)-1][name] = typ
}

func (g *generator) pushReactiveScope() {
	g.reactives = append(g.reactives, map[string]checker.Type{})
}

func (g *generator) popReactiveScope() {
	g.reactives = g.reactives[:len(g.reactives)-1]
}

func (g *generator) addReactive(name string, typ checker.Type) {
	if len(g.reactives) == 0 {
		g.pushReactiveScope()
	}
	g.reactives[len(g.reactives)-1][name] = typ
}

func (g *generator) isReactive(name string) bool {
	_, ok := g.lookupReactive(name)
	return ok
}

func (g *generator) lookupReactive(name string) (checker.Type, bool) {
	for i := len(g.reactives) - 1; i >= 0; i-- {
		if typ, ok := g.reactives[i][name]; ok {
			return typ, true
		}
	}
	return checker.Unknown, false
}

func (g *generator) isSignal(name string) bool {
	_, ok := g.lookupSignal(name)
	return ok
}

func (g *generator) lookupSignal(name string) (checker.Type, bool) {
	for i := len(g.signals) - 1; i >= 0; i-- {
		if typ, ok := g.signals[i][name]; ok {
			return typ, true
		}
	}
	return checker.Unknown, false
}

func (g *generator) exprUsesSignal(expr ir.Expr) bool {
	used := false
	ir.WalkExpr(expr, func(e ir.Expr) {
		if ident, ok := e.(*ir.Identifier); ok && (g.isSignal(ident.Name) || g.isReactive(ident.Name)) {
			used = true
		}
	})
	return used
}

func (g *generator) exprSignalDeps(expr ir.Expr) []string {
	seen := map[string]bool{}
	var deps []string
	ir.WalkExpr(expr, func(e ir.Expr) {
		if ident, ok := e.(*ir.Identifier); ok && (g.isSignal(ident.Name) || g.isReactive(ident.Name)) && !seen[ident.Name] {
			seen[ident.Name] = true
			deps = append(deps, ident.Name)
		}
	})
	return deps
}

func tsBinaryOp(op lexer.Kind) string {
	switch op {
	case lexer.EqualEqual:
		return "==="
	case lexer.BangEqual:
		return "!=="
	default:
		return op.String()
	}
}

func tsPrecedence(op lexer.Kind) int {
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
