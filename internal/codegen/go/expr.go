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
	case *ir.TemplateLiteral:
		return g.templateLiteral(e)
	case *ir.CharLiteral:
		return strconv.QuoteRune(e.Value)
	case *ir.RegexLiteral:
		return fmt.Sprintf("newRuneRegex(%q, %q)", e.Pattern, e.Flags)
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
		if e.Op == lexer.Tilde && e.Expr.ResultType() == checker.BigInt {
			return fmt.Sprintf("new(big.Int).Not(%s)", g.exprPrec(e.Expr, 5))
		}
		s := fmt.Sprintf("%s%s", e.Op, g.exprPrec(e.Expr, 5))
		if 5 < parentPrec {
			return "(" + s + ")"
		}
		return s
	case *ir.PostfixExpr:
		return g.postfixExpr(e)
	case *ir.ResultUnwrapExpr:
		return "/* result unwrap is only supported in statement position */"
	case *ir.BinaryExpr:
		if e.Op == lexer.QuestionQuestion {
			return g.nullCoalesceExpr(e)
		}
		if expr := g.bigIntBinaryExpr(e); expr != "" {
			return expr
		}
		prec := goPrecedence(e.Op)
		s := fmt.Sprintf("%s %s %s", g.exprPrec(e.Left, prec), goBinaryOp(e.Op), g.exprPrec(e.Right, prec+1))
		if prec < parentPrec {
			return "(" + s + ")"
		}
		return s
	case *ir.TernaryExpr:
		return g.ternaryExpr(e)
	case *ir.BlockExpr:
		ret := e.ResultType()
		if ret == checker.Void {
			return fmt.Sprintf("func() { %s }()", g.blockInline(e, ret))
		}
		return fmt.Sprintf("func() %s { %s }()", goType(ret), g.blockInline(e, ret))
	case *ir.PatternBlock:
		return "/* pattern block is only supported in function body */"
	case *ir.AssignExpr:
		if target, ok := e.Target.(*ir.IndexExpr); ok {
			if expr, ok := g.indexAssignExpr(target, e.Value); ok {
				return expr
			}
		}
		if g.isSignal(e.Name) {
			return fmt.Sprintf("%s.Set(%s)", mangleIdent(e.Name), g.expr(e.Value))
		}
		if e.Target != nil && e.Name == "" {
			return fmt.Sprintf("%s = %s", g.expr(e.Target), g.expr(e.Value))
		}
		return fmt.Sprintf("%s = %s", mangleIdent(e.Name), g.expr(e.Value))
	case *ir.CallExpr:
		return g.callExpr(e)
	case *ir.LambdaExpr:
		return g.lambda(e)
	case *ir.IndexExpr:
		if _, ok := checker.TupleElements(e.Receiver.ResultType()); ok {
			if index, ok := e.Index.(*ir.IntegerLiteral); ok {
				return fmt.Sprintf("%s.F%d", g.expr(e.Receiver), index.Value)
			}
		}
		if _, _, ok := checker.MapKeyValue(e.Receiver.ResultType()); ok {
			value := g.nextTemp("value")
			okName := g.nextTemp("ok")
			return fmt.Sprintf("func() any { %s, %s := %s[%s]; if !%s { return nil }; return %s }()", value, okName, g.expr(e.Receiver), g.expr(e.Index), okName, value)
		}
		return fmt.Sprintf("%s[%s]", g.expr(e.Receiver), g.expr(e.Index))
	case *ir.SelectorExpr:
		if e.Static {
			if ident, ok := e.Receiver.(*ir.Identifier); ok {
				return mangleIdent(ident.Name + "_" + e.Name)
			}
		}
		if member, ok := g.enumMemberSelector(e); ok {
			return member
		}
		if at, ok := e.Receiver.(*ir.AtExpr); ok {
			if fn, ok := g.file.Stdlib.Function(at.Name, e.Name); ok && fn.Go != nil && fn.Go.Symbol != "" {
				return fn.Go.Symbol
			}
		}
		if moduleName, ok := checker.ModuleNamespaceName(e.Receiver.ResultType()); ok {
			if fn, ok := g.file.Stdlib.Function(moduleName, e.Name); ok && fn.Go != nil && fn.Go.Symbol != "" {
				return fn.Go.Symbol
			}
		}
		if _, ok := checker.ImportNamespacePath(e.Receiver.ResultType()); ok {
			return mangleIdent(selectorResolvedName(e))
		}
		return g.expr(e.Receiver) + "." + mangleIdent(e.Name)
	case *ir.ArrayLiteral:
		elemType := checker.Unknown
		if elem, ok := checker.ArrayElement(e.ResultType()); ok {
			elemType = elem
		}
		return goArrayLiteral(elemType, e.Elements, func(expr ir.Expr) string {
			return g.expr(expr)
		})
	case *ir.TupleLiteral:
		elems := make([]string, 0, len(e.Elements))
		for idx, elem := range e.Elements {
			elems = append(elems, fmt.Sprintf("F%d: %s", idx, g.expr(elem)))
		}
		return fmt.Sprintf("%s{%s}", goType(e.ResultType()), strings.Join(elems, ", "))
	case *ir.MapLiteral:
		return g.mapLiteral(e)
	case *ir.SpreadExpr:
		return "/* spread is only supported inside array literals */"
	case *ir.ReactiveLiteral:
		return g.expr(e.Value)
	case *ir.StructLiteral:
		if e.TypeName == string(checker.Error) {
			fields := make([]string, 0, len(e.Fields))
			for _, field := range e.Fields {
				fields = append(fields, fmt.Sprintf("%s: %s", mangleIdent(field.Name), g.structFieldExpr(e.TypeName, field)))
			}
			return fmt.Sprintf("&runeError{%s}", strings.Join(fields, ", "))
		}
		fields := make([]string, 0, len(e.Fields))
		for _, field := range e.Fields {
			fields = append(fields, fmt.Sprintf("%s: %s", mangleIdent(field.Name), g.structFieldExpr(e.TypeName, field)))
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
		return "struct{}{}"
	case *ir.ThisExpr:
		if name := g.currentThisName(); name != "" {
			return name
		}
		return mangleIdent("this")
	default:
		return "/* unsupported */"
	}
}

func selectorResolvedName(sel *ir.SelectorExpr) string {
	if sel.ResolvedName != "" {
		return sel.ResolvedName
	}
	return sel.Name
}

func (g *generator) templateLiteral(lit *ir.TemplateLiteral) string {
	parts := make([]string, 0, len(lit.Parts))
	for _, part := range lit.Parts {
		if part.Text != "" {
			parts = append(parts, strconv.Quote(part.Text))
		}
		if part.Expr == nil {
			continue
		}
		expr := g.expr(part.Expr)
		if part.Expr.ResultType() == checker.String {
			parts = append(parts, expr)
		} else {
			parts = append(parts, fmt.Sprintf("runeTemplateString(%s)", expr))
		}
	}
	if len(parts) == 0 {
		return `""`
	}
	return strings.Join(parts, " + ")
}

func (g *generator) indexAssignExpr(target *ir.IndexExpr, value ir.Expr) (string, bool) {
	if _, _, ok := checker.MapKeyValue(target.Receiver.ResultType()); !ok {
		return "", false
	}
	return fmt.Sprintf("%s[%s] = %s", g.expr(target.Receiver), g.expr(target.Index), g.expr(value)), true
}

func (g *generator) nullCoalesceExpr(expr *ir.BinaryExpr) string {
	if expr.Left.ResultType() == checker.Null {
		return g.expr(expr.Right)
	}
	inner, ok := parseGoNullableType(string(expr.Left.ResultType()))
	if !ok {
		return g.expr(expr.Left)
	}
	value := g.nextTemp("coalesce")
	resultType := expr.ResultType()
	if _, nullable := parseGoNullableType(string(resultType)); nullable || resultType == checker.Null || resultType == checker.Unknown {
		return fmt.Sprintf("func() %s { %s := %s; if %s != nil { return %s }; return %s }()", goType(resultType), value, g.expr(expr.Left), value, value, g.expr(expr.Right))
	}
	return fmt.Sprintf("func() %s { %s := %s; if %s != nil { return %s.(%s) }; return %s }()", goType(resultType), value, g.expr(expr.Left), value, value, goType(checker.Type(inner)), g.expr(expr.Right))
}

func (g *generator) mapLiteral(lit *ir.MapLiteral) string {
	mapType := goType(lit.ResultType())
	entries := make([]string, 0, len(lit.Entries))
	for _, entry := range lit.Entries {
		entries = append(entries, fmt.Sprintf("%s: %s", g.expr(entry.Key), g.expr(entry.Value)))
	}
	return fmt.Sprintf("%s{%s}", mapType, strings.Join(entries, ", "))
}

func (g *generator) callExpr(e *ir.CallExpr) string {
	raw := g.callExprRaw(e)
	if e.Await {
		return "runeAwait(" + raw + ")"
	}
	return raw
}

func (g *generator) callExprRaw(e *ir.CallExpr) string {
	if ident, ok := e.Callee.(*ir.Identifier); ok {
		switch ident.Name {
		case "Ok":
			if len(e.Args) != 1 {
				return g.zeroValue(e.ResultType())
			}
			okType, errType := resultTypeArgs(e.ResultType())
			return fmt.Sprintf("runeOk[%s, %s](%s)", goType(okType), goType(errType), g.expr(e.Args[0]))
		case "Err":
			if len(e.Args) != 1 {
				return g.zeroValue(e.ResultType())
			}
			okType, errType := resultTypeArgs(e.ResultType())
			return fmt.Sprintf("runeErr[%s, %s](%s)", goType(okType), goType(errType), g.expr(e.Args[0]))
		}
	}
	if constructor, ok := g.enumConstructorCall(e); ok {
		return constructor
	}
	if intrinsicCall, ok := g.moduleIntrinsicCall(e); ok {
		return intrinsicCall
	}
	if intrinsicCall, ok := g.receiverIntrinsicCall(e); ok {
		return intrinsicCall
	}
	if arrayCall, ok := g.arrayMethodCall(e); ok {
		return arrayCall
	}
	if mapCall, ok := g.mapMethodCall(e); ok {
		return mapCall
	}
	if iterCall, ok := g.iterMethodCall(e); ok {
		return iterCall
	}
	if primitiveCall, ok := g.primitiveMethodCall(e); ok {
		return primitiveCall
	}
	if methodCall, ok := g.userMethodCall(e); ok {
		return methodCall
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
}

func (g *generator) enumConstructorCall(call *ir.CallExpr) (string, bool) {
	ident, ok := call.Callee.(*ir.Identifier)
	if !ok {
		return "", false
	}
	enum, member, ok := g.enumMemberForConstructor(call.ResultType(), ident.Name)
	if !ok || !enumHasPayload(enum) {
		return "", false
	}
	args := make([]string, 0, len(call.Args))
	for _, arg := range call.Args {
		args = append(args, g.expr(arg))
	}
	if len(args) == 0 {
		return fmt.Sprintf("%s{__tag: %s}", goType(call.ResultType()), mangleEnumMember(enum.Name, member.Name)), true
	}
	return fmt.Sprintf("%s{__tag: %s, __payload: []any{%s}}", goType(call.ResultType()), mangleEnumMember(enum.Name, member.Name), strings.Join(args, ", ")), true
}

func (g *generator) userMethodCall(call *ir.CallExpr) (string, bool) {
	sel, ok := call.Callee.(*ir.SelectorExpr)
	if !ok || !sel.Static {
		return "", false
	}
	typeName := string(sel.Receiver.ResultType())
	if base, _, ok := parseGoGenericType(typeName); ok {
		typeName = base
	}
	for _, typ := range g.file.Types {
		if typ.Name != typeName {
			continue
		}
		for _, method := range typ.Methods {
			if method.Name != sel.Name {
				continue
			}
			args := make([]string, 0, len(call.Args)+1)
			if !method.Static {
				continue
			}
			for _, arg := range call.Args {
				args = append(args, g.expr(arg))
			}
			return fmt.Sprintf("%s(%s)", mangleIdent(typeName+"_"+sel.Name), strings.Join(args, ", ")), true
		}
	}
	return "", false
}

func (g *generator) iterMethodCall(call *ir.CallExpr) (string, bool) {
	sel, ok := call.Callee.(*ir.SelectorExpr)
	if !ok {
		return "", false
	}
	elem, ok := checker.IterValue(sel.Receiver.ResultType())
	if !ok {
		return "", false
	}
	receiver := g.expr(sel.Receiver)
	elemType := goType(elem)
	switch sel.Name {
	case "toArray":
		return fmt.Sprintf("func() []%s { iter := %s; out := []%s{}; for { item := iter.__next(); if !item.F1 { return out }; out = append(out, item.F0) } }()", elemType, receiver, elemType), true
	case "each":
		if len(call.Args) != 1 {
			return g.zeroValue(call.ResultType()), true
		}
		callback, arity := g.iterCallback(call.Args[0], receiver)
		return fmt.Sprintf("func() any { iter := %s; index := 0; for { item := iter.__next(); if !item.F1 { return any(nil) }; %s; index++ } }()", receiver, g.iterCallbackCall(callback, arity, "item.F0", "index", "iter")), true
	case "map":
		if len(call.Args) != 1 {
			return g.zeroValue(call.ResultType()), true
		}
		outElem := checker.Unknown
		if arrayElem, ok := checker.ArrayElement(call.ResultType()); ok {
			outElem = arrayElem
		}
		callback, arity := g.iterCallback(call.Args[0], receiver)
		return fmt.Sprintf("func() []%s { iter := %s; out := []%s{}; index := 0; for { item := iter.__next(); if !item.F1 { return out }; out = append(out, %s); index++ } }()", goType(outElem), receiver, goType(outElem), g.iterCallbackCall(callback, arity, "item.F0", "index", "iter")), true
	default:
		return "", false
	}
}

func (g *generator) iterCallback(expr ir.Expr, receiver string) (string, int) {
	if lambda, ok := expr.(*ir.LambdaExpr); ok {
		return g.expr(lambda), len(lambda.Params)
	}
	params, _, ok := parseGoFuncType(string(expr.ResultType()))
	if !ok {
		return g.expr(expr), 3
	}
	return g.expr(expr), len(params)
}

func (g *generator) iterCallbackCall(callback string, arity int, value string, index string, iter string) string {
	args := []string{value, index, iter}
	if arity < len(args) {
		args = args[:arity]
	}
	return fmt.Sprintf("%s(%s)", callback, strings.Join(args, ", "))
}

func resultTypeArgs(typ checker.Type) (checker.Type, checker.Type) {
	base, args, ok := parseGoGenericType(string(typ))
	if !ok || base != "Result" || len(args) != 2 {
		return checker.Unknown, checker.Unknown
	}
	return checker.Type(args[0]), checker.Type(args[1])
}

func (g *generator) enumMemberSelector(sel *ir.SelectorExpr) (string, bool) {
	ident, ok := sel.Receiver.(*ir.Identifier)
	if !ok {
		return "", false
	}
	for _, enum := range g.file.Enums {
		enumType := checker.Type(enum.Name)
		if enum.Name != ident.Name || sel.ResultType() != enumType || ident.ResultType() != enumType {
			continue
		}
		for _, member := range enum.Members {
			if member.Name == sel.Name {
				if enumHasPayload(enum) {
					return fmt.Sprintf("%s{__tag: %s}", goType(enumType), mangleEnumMember(enum.Name, member.Name)), true
				}
				return mangleEnumMember(enum.Name, member.Name), true
			}
		}
	}
	return "", false
}

func enumHasPayload(enum *ir.EnumType) bool {
	if enum == nil {
		return false
	}
	for _, member := range enum.Members {
		if len(member.Params) > 0 {
			return true
		}
	}
	return false
}

func (g *generator) enumForType(typ checker.Type) *ir.EnumType {
	name := goBaseTypeName(typ)
	for _, enum := range g.file.Enums {
		if enum.Name == name {
			return enum
		}
	}
	return nil
}

func goBaseTypeName(typ checker.Type) string {
	if base, _, ok := parseGoGenericType(string(typ)); ok {
		return base
	}
	return string(typ)
}

func (g *generator) enumMemberForConstructor(typ checker.Type, name string) (*ir.EnumType, *ir.EnumMember, bool) {
	enum := g.enumForType(typ)
	if enum == nil {
		return nil, nil, false
	}
	for idx := range enum.Members {
		if enum.Members[idx].Name == name {
			return enum, &enum.Members[idx], true
		}
	}
	return nil, nil, false
}

func (g *generator) hasEnumType(typ checker.Type) bool {
	for _, enum := range g.file.Enums {
		if checker.Type(enum.Name) == typ {
			return true
		}
	}
	return false
}

func goArrayLiteral(elemType checker.Type, elements []ir.Expr, emit func(ir.Expr) string) string {
	if arrayElementsHaveSpread(elements) {
		resultType := goType(elemType)
		var b strings.Builder
		b.WriteString(fmt.Sprintf("func() []%s { ", resultType))
		b.WriteString(fmt.Sprintf("out := []%s{}; ", resultType))
		for _, elem := range elements {
			if spread, ok := elem.(*ir.SpreadExpr); ok {
				b.WriteString(fmt.Sprintf("out = append(out, %s...); ", emit(spread.Expr)))
				continue
			}
			b.WriteString(fmt.Sprintf("out = append(out, %s); ", emit(elem)))
		}
		b.WriteString("return out }()")
		return b.String()
	}
	elems := make([]string, 0, len(elements))
	for _, elem := range elements {
		elems = append(elems, emit(elem))
	}
	return fmt.Sprintf("[]%s{%s}", goType(elemType), strings.Join(elems, ", "))
}

func arrayElementsHaveSpread(elements []ir.Expr) bool {
	for _, elem := range elements {
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
	case lexer.BitAnd:
		return fmt.Sprintf("new(big.Int).And(%s, %s)", left, right)
	case lexer.BitOr:
		return fmt.Sprintf("new(big.Int).Or(%s, %s)", left, right)
	case lexer.BitXor:
		return fmt.Sprintf("new(big.Int).Xor(%s, %s)", left, right)
	case lexer.ShiftLeft:
		return fmt.Sprintf("new(big.Int).Lsh(%s, uint(%s.Int64()))", left, right)
	case lexer.ShiftRight:
		return fmt.Sprintf("new(big.Int).Rsh(%s, uint(%s.Int64()))", left, right)
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

func (g *generator) primitiveMethodCall(call *ir.CallExpr) (string, bool) {
	sel, ok := call.Callee.(*ir.SelectorExpr)
	if !ok {
		return "", false
	}
	receiver := g.expr(sel.Receiver)
	args := make([]string, 0, len(call.Args))
	for _, arg := range call.Args {
		args = append(args, g.expr(arg))
	}
	switch sel.Receiver.ResultType() {
	case checker.String:
		switch sel.Name {
		case "length":
			return fmt.Sprintf("len([]rune(%s))", receiver), true
		case "isEmpty":
			return fmt.Sprintf("len(%s) == 0", receiver), true
		case "toString":
			return receiver, true
		case "at", "charAt":
			if len(args) != 1 {
				return "/* invalid string.at */", true
			}
			return fmt.Sprintf("[]rune(%s)[%s]", receiver, args[0]), true
		case "slice":
			if len(args) != 2 {
				return "/* invalid string.slice */", true
			}
			return fmt.Sprintf("func() string { runes := []rune(%s); return string(runes[%s:%s]) }()", receiver, args[0], args[1]), true
		case "concat":
			if len(args) != 1 {
				return "/* invalid string.concat */", true
			}
			return fmt.Sprintf("%s + %s", receiver, args[0]), true
		case "includes":
			return fmt.Sprintf("strings.Contains(%s, %s)", receiver, args[0]), true
		case "startsWith":
			return fmt.Sprintf("strings.HasPrefix(%s, %s)", receiver, args[0]), true
		case "endsWith":
			return fmt.Sprintf("strings.HasSuffix(%s, %s)", receiver, args[0]), true
		case "indexOf":
			return fmt.Sprintf("strings.Index(%s, %s)", receiver, args[0]), true
		case "lastIndexOf":
			return fmt.Sprintf("strings.LastIndex(%s, %s)", receiver, args[0]), true
		case "toLowerCase":
			return fmt.Sprintf("strings.ToLower(%s)", receiver), true
		case "toUpperCase":
			return fmt.Sprintf("strings.ToUpper(%s)", receiver), true
		case "trim":
			return fmt.Sprintf("strings.TrimSpace(%s)", receiver), true
		case "trimStart":
			return fmt.Sprintf("strings.TrimLeftFunc(%s, unicode.IsSpace)", receiver), true
		case "trimEnd":
			return fmt.Sprintf("strings.TrimRightFunc(%s, unicode.IsSpace)", receiver), true
		case "repeat":
			return fmt.Sprintf("strings.Repeat(%s, %s)", receiver, args[0]), true
		case "replace":
			return fmt.Sprintf("strings.Replace(%s, %s, %s, 1)", receiver, args[0], args[1]), true
		case "replaceAll":
			return fmt.Sprintf("strings.ReplaceAll(%s, %s, %s)", receiver, args[0], args[1]), true
		case "split":
			return fmt.Sprintf("func() []string { parts := strings.Split(%s, %s); return parts }()", receiver, args[0]), true
		}
	case checker.Char:
		switch sel.Name {
		case "toString":
			return fmt.Sprintf("string(%s)", receiver), true
		}
	case checker.Bool:
		switch sel.Name {
		case "not":
			return "!" + receiver, true
		case "xor":
			if len(args) != 1 {
				return "/* invalid bool.xor */", true
			}
			return fmt.Sprintf("%s != %s", receiver, args[0]), true
		case "toString":
			return fmt.Sprintf("strconv.FormatBool(%s)", receiver), true
		}
	case checker.Regex:
		return g.regexMethodCall(receiver, sel.Name, args)
	}
	return "", false
}

func (g *generator) regexModuleCall(call *ir.CallExpr) (string, bool) {
	fn, ok := g.stdlibFunctionFromCall(call)
	if !ok {
		return "", false
	}
	switch fn.Intrinsic {
	case "regex.new":
		if len(call.Args) != 2 {
			return "/* invalid @regex.new */", true
		}
		return fmt.Sprintf("newRuneRegex(%s, %s)", g.expr(call.Args[0]), g.expr(call.Args[1])), true
	case "regex.escape":
		if len(call.Args) != 1 {
			return "/* invalid @regex.escape */", true
		}
		return fmt.Sprintf("regexp.QuoteMeta(%s)", g.expr(call.Args[0])), true
	default:
		return "", false
	}
}

func (g *generator) regexMethodCall(receiver string, name string, args []string) (string, bool) {
	switch name {
	case "exec", "match", "matchAll", "test", "replace", "replaceAll", "search", "split":
		return fmt.Sprintf("%s.%s(%s)", receiver, name, strings.Join(args, ", ")), true
	case "source":
		return receiver + ".source", true
	case "flags":
		return receiver + ".flags", true
	case "global":
		return fmt.Sprintf("regexHasFlag(%s.flags, 'g')", receiver), true
	case "ignoreCase":
		return fmt.Sprintf("regexHasFlag(%s.flags, 'i')", receiver), true
	case "multiline":
		return fmt.Sprintf("regexHasFlag(%s.flags, 'm')", receiver), true
	case "dotAll":
		return fmt.Sprintf("regexHasFlag(%s.flags, 's')", receiver), true
	case "unicode":
		return fmt.Sprintf("regexHasFlag(%s.flags, 'u')", receiver), true
	case "unicodeSets":
		return fmt.Sprintf("regexHasFlag(%s.flags, 'v')", receiver), true
	case "sticky":
		return fmt.Sprintf("regexHasFlag(%s.flags, 'y')", receiver), true
	case "hasIndices":
		return fmt.Sprintf("regexHasFlag(%s.flags, 'd')", receiver), true
	case "lastIndex":
		return receiver + ".lastIndex", true
	case "setLastIndex":
		if len(args) != 1 {
			return "/* invalid regex.setLastIndex */", true
		}
		return fmt.Sprintf("func() int { %s.lastIndex = %s; return %s.lastIndex }()", receiver, args[0], receiver), true
	default:
		return "", false
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
	if arr, ok := expr.(*ir.ArrayLiteral); ok {
		if elemType, ok := checker.ArrayElement(expected); ok {
			return goArrayLiteral(elemType, arr.Elements, func(expr ir.Expr) string {
				return g.expr(expr)
			})
		}
	}
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

func (g *generator) structFieldExpr(typeName string, field ir.FieldValue) string {
	if typ, ok := g.structFieldType(typeName, field.Name); ok {
		return g.exprAs(field.Value, typ)
	}
	return g.expr(field.Value)
}

func (g *generator) structFieldType(typeName string, fieldName string) (checker.Type, bool) {
	if base, _, ok := parseGoGenericType(typeName); ok {
		typeName = base
	}
	for _, typ := range g.file.Types {
		if typ.Name != typeName {
			continue
		}
		for _, field := range typ.Fields {
			if field.Name == fieldName {
				return field.Type, true
			}
		}
	}
	return checker.Unknown, false
}

func (g *generator) ternaryExpr(expr *ir.TernaryExpr) string {
	condition := g.expr(expr.Condition)
	consequence := g.expr(expr.Consequence)
	if expr.Alternative == nil {
		if expr.ResultType() == checker.Void {
			return fmt.Sprintf("func() { if %s { %s; return } }()", condition, consequence)
		}
		return fmt.Sprintf("func() %s { if %s { return %s }; return %s }()", goType(expr.ResultType()), condition, consequence, g.zeroValue(expr.ResultType()))
	}
	alternative := g.expr(expr.Alternative)
	if expr.ResultType() == checker.Void {
		return fmt.Sprintf("func() { if %s { %s; return }; %s }()", condition, consequence, alternative)
	}
	return fmt.Sprintf("func() %s { if %s { return %s }; return %s }()", goType(expr.ResultType()), condition, consequence, alternative)
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
	if matchNeedsSubjectTemp(match) {
		subject = g.nextTemp("match")
		b.WriteString(" { ")
		b.WriteString(subject)
		b.WriteString(" := ")
		b.WriteString(g.expr(match.Subject))
		b.WriteString("; ")
	} else {
		b.WriteString(" { ")
	}
	restoreMapGetters := g.pushMapPatternGetters(subject, match.Branches)
	defer restoreMapGetters()
	for _, line := range g.mapPatternGetterPrelude(subject, match.Branches) {
		b.WriteString(line)
		b.WriteString("; ")
	}
	b.WriteString("switch { ")
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
		b.WriteString(g.patternBinding(subject, branch.Pattern))
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
		b.WriteString(g.zeroValue(match.ResultType()))
		b.WriteString("; ")
	}
	b.WriteString("}()")
	return b.String()
}

func matchNeedsSubjectTemp(match *ir.MatchExpr) bool {
	for _, branch := range match.Branches {
		if patternNeedsSubjectTemp(branch.Pattern) {
			return true
		}
	}
	return false
}

func (g *generator) pushMapPatternGetters(subject string, branches []ir.PatternBranch) func() {
	getter := ""
	for _, branch := range branches {
		if patternHasMapLikeGet(branch.Pattern) {
			getter = g.nextTemp("mapGet")
			break
		}
	}
	if getter == "" {
		return func() {}
	}
	prev := g.mapGetters
	next := map[string]string{}
	for key, value := range prev {
		next[key] = value
	}
	next[subject] = getter
	g.mapGetters = next
	return func() {
		g.mapGetters = prev
	}
}

func (g *generator) mapPatternGetterPrelude(subject string, branches []ir.PatternBranch) []string {
	var pattern *ir.MapPattern
	for _, branch := range branches {
		pattern = firstMapLikeGetPattern(branch.Pattern)
		if pattern != nil {
			break
		}
	}
	if pattern == nil {
		return nil
	}
	getter := ""
	if g.mapGetters != nil {
		getter = g.mapGetters[subject]
	}
	if getter == "" {
		getter = g.nextTemp("mapGet")
	}
	keyType := checker.Unknown
	if len(pattern.Entries) > 0 {
		keyType = pattern.Entries[0].Key.ResultType()
	}
	seen := g.nextTemp("mapSeen")
	cache := g.nextTemp("mapCache")
	key := g.nextTemp("key")
	value := g.nextTemp("value")
	return []string{
		fmt.Sprintf("%s := map[%s]bool{}", seen, goType(keyType)),
		fmt.Sprintf("%s := map[%s]any{}", cache, goType(keyType)),
		fmt.Sprintf("%s := func(%s %s) any { if %s[%s] { return %s[%s] }; %s := %s.%s(%s); %s[%s] = true; %s[%s] = %s; return %s }",
			getter, key, goType(keyType), seen, key, cache, key, value, subject, mangleIdent("get"), key, seen, key, cache, key, value, value),
	}
}

func patternHasMapLikeGet(pattern ir.Pattern) bool {
	return firstMapLikeGetPattern(pattern) != nil
}

func firstMapLikeGetPattern(pattern ir.Pattern) *ir.MapPattern {
	switch p := pattern.(type) {
	case *ir.MapPattern:
		if p.Access == "get" {
			return p
		}
	case *ir.AsPattern:
		return firstMapLikeGetPattern(p.Pattern)
	case *ir.OrPattern:
		for _, alternative := range p.Alternatives {
			if found := firstMapLikeGetPattern(alternative); found != nil {
				return found
			}
		}
	case *ir.ConstructorPattern:
		for _, arg := range p.Args {
			if found := firstMapLikeGetPattern(arg); found != nil {
				return found
			}
		}
	case *ir.ArrayPattern:
		for _, elem := range p.Elements {
			if found := firstMapLikeGetPattern(elem); found != nil {
				return found
			}
		}
	case *ir.BitPattern:
		return firstMapLikeGetPattern(p.Value)
	case *ir.TuplePattern:
		for _, elem := range p.Elements {
			if found := firstMapLikeGetPattern(elem); found != nil {
				return found
			}
		}
	case *ir.ObjectPattern:
		for _, field := range p.Fields {
			if found := firstMapLikeGetPattern(field.Pattern); found != nil {
				return found
			}
		}
	}
	return nil
}

func patternNeedsSubjectTemp(pattern ir.Pattern) bool {
	switch p := pattern.(type) {
	case *ir.BindingPattern, *ir.ConstructorPattern, *ir.MapPattern, *ir.ObjectPattern, *ir.ArrayPattern, *ir.AsPattern:
		return true
	case *ir.OrPattern:
		for _, alternative := range p.Alternatives {
			if patternNeedsSubjectTemp(alternative) {
				return true
			}
		}
	case *ir.TuplePattern:
		for _, elem := range p.Elements {
			if patternNeedsSubjectTemp(elem) {
				return true
			}
		}
	}
	return false
}

func (g *generator) patternBinding(subject string, pattern ir.Pattern) string {
	var parts []string
	g.appendPatternBindings(&parts, subject, pattern)
	return strings.Join(parts, " ")
}

func (g *generator) appendPatternBindings(parts *[]string, subject string, pattern ir.Pattern) {
	switch p := pattern.(type) {
	case *ir.BindingPattern:
		if p.Constant {
			return
		}
		*parts = append(*parts, fmt.Sprintf("%s := %s;", mangleIdent(p.Name), subject))
	case *ir.TuplePattern:
		for idx, elem := range p.Elements {
			g.appendPatternBindings(parts, fmt.Sprintf("%s.F%d", subject, idx), elem)
		}
	case *ir.ArrayPattern:
		for idx, elem := range p.Elements {
			if bit, ok := elem.(*ir.BitPattern); ok {
				g.appendPatternBindings(parts, g.bitPatternValueExpr(subject, p.SubjectType, bitPatternOffset(p, idx), bit), bit.Value)
				continue
			}
			if _, ok := elem.(*ir.SequenceSpreadPattern); ok {
				continue
			}
			g.appendPatternBindings(parts, goSequenceIndex(subject, p.SubjectType, g.arrayPatternElementIndex(subject, p, idx)), elem)
		}
		if p.RestBinding != "" {
			if irArrayPatternHasBits(p) {
				start := bitPatternRequiredBits(p) / 8
				*parts = append(*parts, fmt.Sprintf("%s := %s;", mangleIdent(p.RestBinding), goSequenceSlice(subject, p.SubjectType, start, 0)))
			} else {
				start := g.arrayPatternPrefixWidth(p, p.RestIndex)
				end := goSubtractExpr(goSequenceLength(subject, p.SubjectType), g.arrayPatternSuffixWidth(p, p.RestIndex))
				*parts = append(*parts, fmt.Sprintf("%s := %s;", mangleIdent(p.RestBinding), goSequenceSliceExpr(subject, p.SubjectType, start, end)))
			}
		}
	case *ir.BitPattern:
		g.appendPatternBindings(parts, subject, p.Value)
	case *ir.AsPattern:
		g.appendPatternBindings(parts, subject, p.Pattern)
		*parts = append(*parts, fmt.Sprintf("%s := %s;", mangleIdent(p.Name), subject))
	case *ir.OrPattern:
		if len(p.Alternatives) > 0 {
			g.appendPatternBindings(parts, subject, p.Alternatives[0])
		}
	case *ir.ConstructorPattern:
		for idx, arg := range p.Args {
			payload := g.constructorPayload(subject, p, idx)
			if payload == "" {
				continue
			}
			g.appendPatternBindings(parts, payload, arg)
		}
	case *ir.MapPattern:
		for _, entry := range p.Entries {
			value := fmt.Sprintf("%s[%s]", subject, g.expr(entry.Key))
			if p.Access == "get" {
				value = g.goMapLikeGet(subject, p.SubjectType, g.expr(entry.Key))
				if !entry.Optional && p.ValueType != checker.Unknown {
					value = fmt.Sprintf("%s.(%s)", value, goType(p.ValueType))
				}
			} else if p.Access == "object" || p.SubjectType == checker.Object {
				if key, ok := entry.Key.(*ir.StringLiteral); ok {
					value = fmt.Sprintf("%s.(map[string]any)[%q]", subject, key.Value)
				}
				if entry.Optional {
					value = fmt.Sprintf("func() any { obj := %s.(map[string]any); return obj[%s] }()", subject, g.expr(entry.Key))
				}
			} else if entry.Optional {
				patternValue := g.nextTemp("pattern")
				ok := g.nextTemp("ok")
				value = fmt.Sprintf("func() any { %s, %s := %s[%s]; if !%s { return any(nil) }; return %s }()", patternValue, ok, subject, g.expr(entry.Key), ok, patternValue)
			}
			g.appendPatternBindings(parts, value, entry.Pattern)
		}
	case *ir.ObjectPattern:
		for _, field := range p.Fields {
			if field.Optional && !field.Exists {
				continue
			}
			value := goObjectFieldAccess(subject, field.Name)
			if p.SubjectType == checker.Object || p.SubjectType == checker.Unknown {
				value = fmt.Sprintf("%s.(map[string]any)[%q]", subject, field.Name)
			}
			g.appendPatternBindings(parts, value, field.Pattern)
		}
	}
}

func goPatternIndex(idx int, pattern *ir.ArrayPattern) string {
	if pattern.RestIndex < 0 || idx < pattern.RestIndex {
		return fmt.Sprintf("%d", idx)
	}
	tail := len(pattern.Elements) - idx
	return fmt.Sprintf("%s - %d", goSequenceLength("_", pattern.SubjectType), tail)
}

func (g *generator) arrayPatternElementIndex(subject string, pattern *ir.ArrayPattern, idx int) string {
	if pattern.RestIndex < 0 || idx < pattern.RestIndex {
		return g.arrayPatternPrefixWidth(pattern, idx)
	}
	return goSubtractExpr(goSequenceLength(subject, pattern.SubjectType), g.arrayPatternSuffixWidth(pattern, idx))
}

func (g *generator) arrayPatternRequiredWidth(pattern *ir.ArrayPattern) string {
	return g.arrayPatternSum(pattern, 0, len(pattern.Elements))
}

func (g *generator) arrayPatternPrefixWidth(pattern *ir.ArrayPattern, end int) string {
	return g.arrayPatternSum(pattern, 0, end)
}

func (g *generator) arrayPatternSuffixWidth(pattern *ir.ArrayPattern, start int) string {
	return g.arrayPatternSum(pattern, start, len(pattern.Elements))
}

func (g *generator) arrayPatternSum(pattern *ir.ArrayPattern, start int, end int) string {
	constants := 0
	terms := []string{}
	for idx := start; idx < end; idx++ {
		elem := pattern.Elements[idx]
		if spread, ok := elem.(*ir.SequenceSpreadPattern); ok {
			terms = append(terms, goSequenceLength(g.expr(spread.Value), spread.Type))
			continue
		}
		constants++
	}
	if constants > 0 || len(terms) == 0 {
		terms = append([]string{fmt.Sprintf("%d", constants)}, terms...)
	}
	return strings.Join(terms, " + ")
}

func goSequenceLength(subject string, typ checker.Type) string {
	switch typ {
	case checker.String:
		return fmt.Sprintf("len([]rune(%s))", subject)
	case checker.Bytes:
		return subject + ".ByteLength()"
	default:
		return "len(" + subject + ")"
	}
}

func goSubtractExpr(left string, right string) string {
	if right == "0" {
		return left
	}
	if strings.ContainsAny(right, "+-*/ ") {
		return fmt.Sprintf("%s-(%s)", left, right)
	}
	return fmt.Sprintf("%s-%s", left, right)
}

func goSequenceIndex(subject string, typ checker.Type, index string) string {
	index = strings.ReplaceAll(index, goSequenceLength("_", typ), goSequenceLength(subject, typ))
	switch typ {
	case checker.String:
		return fmt.Sprintf("[]rune(%s)[%s]", subject, index)
	case checker.Bytes:
		return fmt.Sprintf("%s.GetUInt8(%s)", subject, index)
	default:
		return fmt.Sprintf("%s[%s]", subject, index)
	}
}

func goSequenceSlice(subject string, typ checker.Type, restIndex int, tailCount int) string {
	end := goSequenceLength(subject, typ)
	if tailCount > 0 {
		end = fmt.Sprintf("%s - %d", end, tailCount)
	}
	return goSequenceSliceExpr(subject, typ, fmt.Sprintf("%d", restIndex), end)
}

func goSequenceSliceExpr(subject string, typ checker.Type, start string, end string) string {
	switch typ {
	case checker.String:
		return fmt.Sprintf("func() string { runes := []rune(%s); return string(runes[%s:%s]) }()", subject, start, end)
	case checker.Bytes:
		return fmt.Sprintf("%s.Slice(%s, %s)", subject, start, end)
	default:
		return fmt.Sprintf("append([]%s{}, %s[%s:%s]...)", goType(arrayElemOrUnknown(typ, typ)), subject, start, end)
	}
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
			if isNamespaceValue(s.Value) {
				continue
			}
			parts = append(parts, fmt.Sprintf("%s := %s", mangleIdent(s.Name), g.expr(s.Value)))
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
	g.signalDeps = append(g.signalDeps, map[string][]string{})
}

func (g *generator) popSignalScope() {
	g.signals = g.signals[:len(g.signals)-1]
	g.signalDeps = g.signalDeps[:len(g.signalDeps)-1]
}

func (g *generator) addSignal(name string, typ checker.Type) {
	if len(g.signals) == 0 {
		g.pushSignalScope()
	}
	g.signals[len(g.signals)-1][name] = typ
}

func (g *generator) setSignalDeps(name string, deps []string) {
	if len(g.signalDeps) == 0 {
		g.pushSignalScope()
	}
	g.signalDeps[len(g.signalDeps)-1][name] = append([]string(nil), deps...)
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

func (g *generator) effectSignalDeps(expr ir.Expr) []string {
	deps := g.exprSignalDeps(expr)
	drop := map[string]bool{}
	for _, dep := range deps {
		for _, other := range deps {
			if dep != other && g.signalDependsOn(other, dep, map[string]bool{}) {
				drop[dep] = true
			}
		}
	}
	out := make([]string, 0, len(deps))
	for _, dep := range deps {
		if !drop[dep] {
			out = append(out, dep)
		}
	}
	return out
}

func (g *generator) signalDependsOn(name string, target string, seen map[string]bool) bool {
	if seen[name] {
		return false
	}
	seen[name] = true
	for _, dep := range g.lookupSignalDeps(name) {
		if dep == target || g.signalDependsOn(dep, target, seen) {
			return true
		}
	}
	return false
}

func (g *generator) lookupSignalDeps(name string) []string {
	for i := len(g.signalDeps) - 1; i >= 0; i-- {
		if deps, ok := g.signalDeps[i][name]; ok {
			return deps
		}
	}
	return nil
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
	case lexer.BitOr:
		return 3
	case lexer.BitXor:
		return 4
	case lexer.BitAnd:
		return 5
	case lexer.EqualEqual, lexer.BangEqual:
		return 6
	case lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
		return 7
	case lexer.ShiftLeft, lexer.ShiftRight, lexer.UnsignedShiftRight:
		return 8
	case lexer.Plus, lexer.Minus:
		return 9
	case lexer.Star, lexer.Slash, lexer.Percent:
		return 10
	default:
		return 0
	}
}

func goBinaryOp(op lexer.Kind) string {
	if op == lexer.UnsignedShiftRight {
		return ">>"
	}
	return op.String()
}
