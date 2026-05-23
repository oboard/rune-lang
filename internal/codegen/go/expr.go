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
		if g.isSignal(e.Name) {
			return mangleIdent(e.Name) + ".Get()"
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
		if e.Op == lexer.Minus && e.Expr.ResultType() == checker.BigInt {
			return fmt.Sprintf("new(big.Int).Neg(%s)", g.exprPrec(e.Expr, 5))
		}
		s := fmt.Sprintf("%s%s", e.Op, g.exprPrec(e.Expr, 5))
		if 5 < parentPrec {
			return "(" + s + ")"
		}
		return s
	case *ir.PostfixExpr:
		return g.postfixExpr(e)
	case *ir.BinaryExpr:
		if expr := g.bigIntBinaryExpr(e); expr != "" {
			return expr
		}
		prec := goPrecedence(e.Op)
		s := fmt.Sprintf("%s %s %s", g.exprPrec(e.Left, prec), e.Op, g.exprPrec(e.Right, prec+1))
		if prec < parentPrec {
			return "(" + s + ")"
		}
		return s
	case *ir.AssignExpr:
		if g.isSignal(e.Name) {
			return fmt.Sprintf("%s.Set(%s)", mangleIdent(e.Name), g.expr(e.Value))
		}
		return fmt.Sprintf("%s = %s", mangleIdent(e.Name), g.expr(e.Value))
	case *ir.CallExpr:
		if ffi, ok := g.goFFICall(e); ok {
			return ffi
		}
		if arrayCall, ok := g.arrayMethodCall(e); ok {
			return arrayCall
		}
		args := make([]string, 0, len(e.Args))
		params, _, hasFuncType := parseGoFuncType(string(e.Callee.ResultType()))
		for i, arg := range e.Args {
			if hasFuncType && i < len(params) {
				args = append(args, g.exprAs(arg, checker.Type(params[i])))
			} else {
				args = append(args, g.expr(arg))
			}
		}
		return fmt.Sprintf("%s(%s)", g.expr(e.Callee), strings.Join(args, ", "))
	case *ir.LambdaExpr:
		return g.lambda(e)
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
		elemType := checker.Unknown
		if elem, ok := checker.ArrayElement(e.ResultType()); ok {
			elemType = elem
		}
		if arrayLiteralHasSpread(e) {
			resultType := goType(elemType)
			var b strings.Builder
			b.WriteString(fmt.Sprintf("func() []%s { ", resultType))
			b.WriteString(fmt.Sprintf("out := []%s{}; ", resultType))
			for _, elem := range e.Elements {
				if spread, ok := elem.(*ir.SpreadExpr); ok {
					b.WriteString(fmt.Sprintf("out = append(out, %s...); ", g.expr(spread.Expr)))
					continue
				}
				b.WriteString(fmt.Sprintf("out = append(out, %s); ", g.expr(elem)))
			}
			b.WriteString("return out }()")
			return b.String()
		}
		elems := make([]string, 0, len(e.Elements))
		for _, elem := range e.Elements {
			elems = append(elems, g.expr(elem))
		}
		return fmt.Sprintf("[]%s{%s}", goType(elemType), strings.Join(elems, ", "))
	case *ir.SpreadExpr:
		return "/* spread is only supported inside array literals */"
	case *ir.ReactiveLiteral:
		return g.expr(e.Value)
	case *ir.StructLiteral:
		fields := make([]string, 0, len(e.Fields))
		for _, field := range e.Fields {
			fields = append(fields, fmt.Sprintf("%s: %s", mangleIdent(field.Name), g.expr(field.Value)))
		}
		return fmt.Sprintf("%s{%s}", mangleIdent(e.TypeName), strings.Join(fields, ", "))
	case *ir.AnonymousObjectLiteral:
		return fmt.Sprintf("%s{%s}", anonymousObjectType(e), anonymousObjectFields(g, e))
	case *ir.XMLElement:
		return "/* XML is only supported by the TypeScript backend */"
	case *ir.MatchExpr:
		return g.matchExpr(e)
	case *ir.WatchExpr:
		return g.watchExpr(e)
	case *ir.AtExpr:
		return e.Name
	case *ir.ThisExpr:
		if name := g.currentThisName(); name != "" {
			return name
		}
		return mangleIdent("this")
	default:
		return "/* unsupported */"
	}
}

func arrayLiteralHasSpread(lit *ir.ArrayLiteral) bool {
	for _, elem := range lit.Elements {
		if _, ok := elem.(*ir.SpreadExpr); ok {
			return true
		}
	}
	return false
}

func (g *generator) bigIntBinaryExpr(e *ir.BinaryExpr) string {
	if e.Left.ResultType() != checker.BigInt || e.Right.ResultType() != checker.BigInt {
		return ""
	}
	left := g.expr(e.Left)
	right := g.expr(e.Right)
	switch e.Op {
	case lexer.Plus:
		return fmt.Sprintf("new(big.Int).Add(%s, %s)", left, right)
	case lexer.Minus:
		return fmt.Sprintf("new(big.Int).Sub(%s, %s)", left, right)
	case lexer.Star:
		return fmt.Sprintf("new(big.Int).Mul(%s, %s)", left, right)
	case lexer.Slash:
		return fmt.Sprintf("new(big.Int).Quo(%s, %s)", left, right)
	case lexer.Percent:
		return fmt.Sprintf("new(big.Int).Rem(%s, %s)", left, right)
	case lexer.EqualEqual:
		return fmt.Sprintf("%s.Cmp(%s) == 0", left, right)
	case lexer.BangEqual:
		return fmt.Sprintf("%s.Cmp(%s) != 0", left, right)
	case lexer.Less:
		return fmt.Sprintf("%s.Cmp(%s) < 0", left, right)
	case lexer.LessEqual:
		return fmt.Sprintf("%s.Cmp(%s) <= 0", left, right)
	case lexer.Greater:
		return fmt.Sprintf("%s.Cmp(%s) > 0", left, right)
	case lexer.GreaterEqual:
		return fmt.Sprintf("%s.Cmp(%s) >= 0", left, right)
	default:
		return ""
	}
}

func (g *generator) postfixExpr(expr *ir.PostfixExpr) string {
	target, ok := expr.Expr.(*ir.Identifier)
	if !ok || expr.Op != lexer.PlusPlus {
		return "/* unsupported postfix expression */"
	}
	if g.isSignal(target.Name) {
		name := mangleIdent(target.Name)
		return fmt.Sprintf("func() int { old := %s.Get(); %s.Set(old + 1); return old }()", name, name)
	}
	name := mangleIdent(target.Name)
	return fmt.Sprintf("func() int { old := %s; %s++; return old }()", name, name)
}

func (g *generator) exprRaw(expr ir.Expr) string {
	prev := g.signals
	g.signals = nil
	defer func() { g.signals = prev }()
	return g.expr(expr)
}

func (g *generator) watchExpr(watch *ir.WatchExpr) string {
	target, ok := watch.Target.(*ir.Identifier)
	if !ok || !g.isSignal(target.Name) {
		return "/* watch target is not a signal */"
	}
	handler, ok := watch.Handler.(*ir.LambdaExpr)
	if !ok {
		return "/* watch handler is not a lambda */"
	}
	params := handler.Params
	targetType := goType(g.signalType(target.Name))
	goParams := make([]string, 0, len(params))
	if len(params) == 0 {
		goParams = append(goParams, "_ "+targetType, "_ "+targetType)
	} else {
		for _, name := range params {
			goParams = append(goParams, fmt.Sprintf("%s %s", mangleIdent(name), targetType))
		}
	}
	body := g.lambdaBody(handler)
	return fmt.Sprintf("%s.Watch(func(%s) { %s })", mangleIdent(target.Name), strings.Join(goParams, ", "), body)
}

func (g *generator) exprAs(expr ir.Expr, expected checker.Type) string {
	if obj, ok := expr.(*ir.AnonymousObjectLiteral); ok {
		if _, ok := parseGoObjectType(string(expected)); ok {
			return fmt.Sprintf("%s{%s}", goType(expected), anonymousObjectFieldsForType(g, obj, expected))
		}
		if g.hasStructType(expected) {
			return fmt.Sprintf("%s{%s}", goType(expected), anonymousObjectFieldsForType(g, obj, expected))
		}
	}
	return g.expr(expr)
}

func (g *generator) hasStructType(typ checker.Type) bool {
	for _, candidate := range g.file.Types {
		if candidate.Name == string(typ) {
			return true
		}
	}
	return false
}

func (g *generator) matchExpr(match *ir.MatchExpr) string {
	isVoid := match.ResultType() == checker.Void
	ret := ""
	if !isVoid {
		ret = " " + goType(match.ResultType())
	}
	subject := g.expr(match.Subject)
	var b strings.Builder
	b.WriteString("func()")
	b.WriteString(ret)
	b.WriteString(" { switch { ")
	hasDefault := false
	for _, branch := range match.Branches {
		if _, ok := branch.Pattern.(*ir.WildcardPattern); ok {
			hasDefault = true
			b.WriteString("default: ")
		} else {
			b.WriteString("case ")
			b.WriteString(g.patternCondition(subject, branch.Pattern))
			b.WriteString(": ")
		}
		if isVoid {
			b.WriteString(g.expr(branch.Expr))
			b.WriteString("; return; ")
		} else {
			b.WriteString("return ")
			b.WriteString(g.expr(branch.Expr))
			b.WriteString("; ")
		}
	}
	b.WriteString("}; ")
	if !hasDefault && !isVoid {
		b.WriteString("return ")
		b.WriteString(zeroValue(match.ResultType()))
		b.WriteString("; ")
	}
	b.WriteString("}()")
	return b.String()
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

func (g *generator) lambda(lambda *ir.LambdaExpr) string {
	params, ret, ok := parseGoFuncType(string(lambda.ResultType()))
	if !ok {
		params = make([]string, len(lambda.Params))
		for i := range params {
			params[i] = string(checker.Unknown)
		}
		ret = string(checker.Void)
	}
	goParams := make([]string, 0, len(lambda.Params))
	for i, name := range lambda.Params {
		typ := checker.Unknown
		if i < len(params) {
			typ = checker.Type(params[i])
		}
		goParams = append(goParams, fmt.Sprintf("%s %s", mangleIdent(name), goType(typ)))
	}
	result := "func(" + strings.Join(goParams, ", ") + ")"
	if ret != string(checker.Void) {
		result += " " + goType(checker.Type(ret))
	}
	result += " { "
	result += g.lambdaBodyWithReturn(lambda, checker.Type(ret))
	result += " }"
	return result
}

func (g *generator) lambdaBody(lambda *ir.LambdaExpr) string {
	_, ret, ok := parseGoFuncType(string(lambda.ResultType()))
	if !ok {
		ret = string(checker.Void)
	}
	return g.lambdaBodyWithReturn(lambda, checker.Type(ret))
}

func (g *generator) lambdaBodyWithReturn(lambda *ir.LambdaExpr, ret checker.Type) string {
	if block, ok := lambda.Body.(*ir.BlockExpr); ok {
		return g.blockInline(block, ret)
	}
	if ret == checker.Void {
		if expr := g.expr(lambda.Body); expr != "" {
			return expr
		}
		return ""
	}
	return "return " + g.expr(lambda.Body)
}

func (g *generator) blockInline(block *ir.BlockExpr, ret checker.Type) string {
	parts := make([]string, 0, len(block.Statements))
	for i, stmt := range block.Statements {
		last := i == len(block.Statements)-1
		switch s := stmt.(type) {
		case *ir.LetStmt:
			parts = append(parts, fmt.Sprintf("%s := %s; _ = %s", mangleIdent(s.Name), g.expr(s.Value), mangleIdent(s.Name)))
		case *ir.AssignStmt:
			if g.isSignal(s.Name) {
				parts = append(parts, fmt.Sprintf("%s.Set(%s)", mangleIdent(s.Name), g.expr(s.Value)))
			} else {
				parts = append(parts, fmt.Sprintf("%s = %s", mangleIdent(s.Name), g.expr(s.Value)))
			}
		case *ir.ExprStmt:
			expr := g.expr(s.Expr)
			if expr == "" {
				continue
			}
			if last && ret != checker.Void {
				parts = append(parts, "return "+expr)
			} else {
				parts = append(parts, expr)
			}
		}
	}
	return strings.Join(parts, "; ")
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

func (g *generator) isSignal(name string) bool {
	_, ok := g.lookupSignal(name)
	return ok
}

func (g *generator) signalType(name string) checker.Type {
	typ, ok := g.lookupSignal(name)
	if !ok {
		return checker.Unknown
	}
	return typ
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
		if ident, ok := e.(*ir.Identifier); ok && g.isSignal(ident.Name) {
			used = true
		}
	})
	return used
}

func (g *generator) exprSignalDeps(expr ir.Expr) []string {
	seen := map[string]bool{}
	var deps []string
	ir.WalkExpr(expr, func(e ir.Expr) {
		if ident, ok := e.(*ir.Identifier); ok && g.isSignal(ident.Name) && !seen[ident.Name] {
			seen[ident.Name] = true
			deps = append(deps, ident.Name)
		}
	})
	return deps
}

func anonymousObjectType(obj *ir.AnonymousObjectLiteral) string {
	if fields, ok := parseGoObjectType(string(obj.ResultType())); ok && len(fields) > 0 {
		parts := make([]string, 0, len(fields))
		for _, field := range fields {
			parts = append(parts, fmt.Sprintf("%s %s", mangleIdent(field.name), goType(checker.Type(field.typ))))
		}
		return "struct{" + strings.Join(parts, "; ") + "}"
	}
	fields := make([]string, 0, len(obj.Fields))
	for _, field := range obj.Fields {
		fields = append(fields, fmt.Sprintf("%s %s", mangleIdent(field.Name), goType(field.Value.ResultType())))
	}
	return "struct{" + strings.Join(fields, "; ") + "}"
}

func anonymousObjectFields(g *generator, obj *ir.AnonymousObjectLiteral) string {
	return anonymousObjectFieldsForType(g, obj, obj.ResultType())
}

func anonymousObjectFieldsForType(g *generator, obj *ir.AnonymousObjectLiteral, typ checker.Type) string {
	fields := make([]string, 0, len(obj.Fields))
	if resultFields, ok := parseGoObjectType(string(typ)); ok && len(resultFields) > 0 {
		byName := map[string]ir.Expr{}
		for _, field := range obj.Fields {
			byName[field.Name] = field.Value
		}
		for _, field := range resultFields {
			if value := byName[field.name]; value != nil {
				fields = append(fields, fmt.Sprintf("%s: %s", mangleIdent(field.name), g.expr(value)))
			}
		}
		return strings.Join(fields, ", ")
	}
	for _, field := range obj.Fields {
		fields = append(fields, fmt.Sprintf("%s: %s", mangleIdent(field.Name), g.expr(field.Value)))
	}
	return strings.Join(fields, ", ")
}

func goPrecedence(op lexer.Kind) int {
	switch op {
	case lexer.OrOr:
		return 1
	case lexer.AndAnd:
		return 2
	case lexer.EqualEqual, lexer.BangEqual:
		return 3
	case lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
		return 4
	case lexer.Plus, lexer.Minus:
		return 5
	case lexer.Star, lexer.Slash, lexer.Percent:
		return 6
	default:
		return 0
	}
}
