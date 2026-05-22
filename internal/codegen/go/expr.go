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
	case *ir.AnonymousObjectLiteral:
		return fmt.Sprintf("%s{%s}", anonymousObjectType(e), anonymousObjectFields(g, e))
	case *ir.MatchExpr:
		return g.matchExpr(e)
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
	ret := goType(match.ResultType())
	subject := g.expr(match.Subject)
	var b strings.Builder
	b.WriteString("func() ")
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
		b.WriteString("return ")
		b.WriteString(g.expr(branch.Expr))
		b.WriteString("; ")
	}
	b.WriteString("}; ")
	if !hasDefault {
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
	if ret == string(checker.Void) {
		if expr := g.expr(lambda.Body); expr != "" {
			result += expr
		}
	} else {
		result += "return " + g.expr(lambda.Body)
	}
	result += " }"
	return result
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
