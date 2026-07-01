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
		if fn := g.lookupFunction(e.Name); fn != nil {
			return FunctionSymbolName(fn)
		}
		if g.isSignal(e.Name) || g.isReactive(e.Name) {
			return mangleIdent(e.Name) + ".get()"
		}
		return mangleIdent(e.Name)
	case *ir.AtExpr:
		return "{}"
	case *ir.ThisExpr:
		if name := g.currentThisName(); name != "" {
			return name
		}
		return mangleIdent("this")
	case *ir.IntegerLiteral:
		return strconv.Itoa(e.Value)
	case *ir.DoubleLiteral:
		if e.Raw != "" {
			return e.Raw
		}
		return strconv.FormatFloat(e.Value, 'f', -1, 64)
	case *ir.BigIntLiteral:
		return e.Value + "n"
	case *ir.StringLiteral:
		return strconv.Quote(e.Value)
	case *ir.TemplateLiteral:
		return g.templateLiteral(e)
	case *ir.CharLiteral:
		return strconv.Quote(string(e.Value))
	case *ir.RegexLiteral:
		return e.Raw
	case *ir.BoolLiteral:
		if e.Value {
			return "true"
		}
		return "false"
	case *ir.NullLiteral:
		return "null"
	case *ir.UnaryExpr:
		s := fmt.Sprintf("%s%s", e.Op, g.exprPrec(e.Expr, 6))
		if 5 < parentPrec {
			return "(" + s + ")"
		}
		return s
	case *ir.PostfixExpr:
		return g.postfixExpr(e)
	case *ir.ResultUnwrapExpr:
		return "undefined /* result unwrap is only supported in statement position */"
	case *ir.BinaryExpr:
		prec := tsPrecedence(e.Op)
		op := tsBinaryOp(e.Op)
		s := fmt.Sprintf("%s %s %s", g.exprPrec(e.Left, prec), op, g.exprPrec(e.Right, prec+1))
		if e.Op == lexer.Slash && tsIntegerResultType(e.ResultType()) {
			s = fmt.Sprintf("Math.trunc(%s)", s)
		}
		if tsArithmeticOp(e.Op) && isTSGenericResultType(e.ResultType()) {
			s = fmt.Sprintf("(%s) as %s", s, tsType(e.ResultType()))
		}
		if prec < parentPrec {
			return "(" + s + ")"
		}
		return s
	case *ir.TernaryExpr:
		condition := g.exprPrec(e.Condition, 1)
		alternative := "undefined"
		if e.Alternative != nil {
			alternative = g.exprPrec(e.Alternative, 0)
		}
		s := fmt.Sprintf("%s ? %s : %s", condition, g.expr(e.Consequence), alternative)
		if parentPrec > 0 {
			return "(" + s + ")"
		}
		return s
	case *ir.AssignExpr:
		if target, ok := e.Target.(*ir.IndexExpr); ok {
			if expr, ok := g.indexAssignExpr(target, e.Value); ok {
				return expr
			}
		}
		if target, ok := e.Target.(*ir.SelectorExpr); ok {
			if expr, ok := g.selectorAssignExpr(target, e.Value); ok {
				return expr
			}
		}
		if g.isSignal(e.Name) {
			return fmt.Sprintf("%s.set(%s)", mangleIdent(e.Name), g.expr(e.Value))
		}
		if g.isReactive(e.Name) {
			return fmt.Sprintf("%s.set(%s)", mangleIdent(e.Name), g.expr(e.Value))
		}
		if e.Target != nil && e.Name == "" {
			return fmt.Sprintf("%s = %s", g.expr(e.Target), g.expr(e.Value))
		}
		return fmt.Sprintf("%s = %s", mangleIdent(e.Name), g.expr(e.Value))
	case *ir.CallExpr:
		return g.callExpr(e)
	case *ir.LambdaExpr:
		return g.lambda(e)
	case *ir.SelectorExpr:
		if e.Static {
			if ident, ok := e.Receiver.(*ir.Identifier); ok {
				return mangleMethod(ident.Name, e.Name)
			}
		}
		if member, ok := g.enumMemberSelector(e); ok {
			return member
		}
		if _, ok := checker.ImportNamespacePath(e.Receiver.ResultType()); ok {
			return mangleIdent(selectorResolvedName(e))
		}
		if at, ok := e.Receiver.(*ir.AtExpr); ok {
			return "@" + at.Name + "." + e.Name
		}
		return tsPropertyAccess(g.selectorReceiverExpr(e.Receiver), e.Name)
	case *ir.IndexExpr:
		if _, ok := checker.TupleElements(e.Receiver.ResultType()); ok {
			return fmt.Sprintf("%s[%s]", g.expr(e.Receiver), g.expr(e.Index))
		}
		if _, _, ok := checker.MapKeyValue(e.Receiver.ResultType()); ok {
			mapName := g.nextTemp("map")
			keyName := g.nextTemp("key")
			return fmt.Sprintf("((%s, %s) => %s.has(%s) ? %s.get(%s)! : null)(%s, %s)", mapName, keyName, mapName, keyName, mapName, keyName, g.expr(e.Receiver), g.expr(e.Index))
		}
		return fmt.Sprintf("%s[%s]", g.selectorReceiverExpr(e.Receiver), g.expr(e.Index))
	case *ir.ArrayLiteral:
		elems := make([]string, 0, len(e.Elements))
		for _, elem := range e.Elements {
			elems = append(elems, g.expr(elem))
		}
		return "[" + strings.Join(elems, ", ") + "]"
	case *ir.TupleLiteral:
		elems := make([]string, 0, len(e.Elements))
		for _, elem := range e.Elements {
			elems = append(elems, g.expr(elem))
		}
		return "[" + strings.Join(elems, ", ") + "]"
	case *ir.MapLiteral:
		return g.mapLiteral(e)
	case *ir.SpreadExpr:
		return "..." + g.expr(e.Expr)
	case *ir.ReactiveLiteral:
		return g.reactiveLiteral(e)
	case *ir.StructLiteral:
		fields := make([]string, 0, len(e.Fields))
		for _, field := range e.Fields {
			fields = append(fields, fmt.Sprintf("%s: %s", tsPropertyName(field.Name), g.expr(field.Value)))
		}
		return "{" + strings.Join(fields, ", ") + "}"
	case *ir.AnonymousObjectLiteral:
		fields := make([]string, 0, len(e.Fields))
		for _, field := range e.Fields {
			fields = append(fields, fmt.Sprintf("%s: %s", tsPropertyName(field.Name), g.expr(field.Value)))
		}
		return "{" + strings.Join(fields, ", ") + "}"
	case *ir.BlockExpr:
		return fmt.Sprintf("(() => { %s })()", g.blockInline(e, e.ResultType()))
	case *ir.PatternBlock:
		return "undefined /* pattern blocks are only supported as function bodies */"
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

func tsIntegerResultType(typ checker.Type) bool {
	switch typ {
	case checker.Int, checker.Int4, checker.Int8, checker.Int16, checker.UInt, checker.UInt8, checker.UInt16:
		return true
	default:
		return false
	}
}

func (g *generator) selectorReceiverExpr(expr ir.Expr) string {
	out := g.expr(expr)
	switch expr.(type) {
	case *ir.BinaryExpr, *ir.TernaryExpr, *ir.AssignExpr, *ir.CallExpr, *ir.BlockExpr:
		return "(" + out + ")"
	default:
		return out
	}
}

func (g *generator) templateLiteral(lit *ir.TemplateLiteral) string {
	var b strings.Builder
	b.WriteByte('`')
	for _, part := range lit.Parts {
		if part.Text != "" {
			b.WriteString(escapeTemplateText(part.Text))
		}
		if part.Expr != nil {
			b.WriteString("${")
			b.WriteString(g.expr(part.Expr))
			b.WriteByte('}')
		}
	}
	b.WriteByte('`')
	return b.String()
}

func escapeTemplateText(text string) string {
	var b strings.Builder
	for i, ch := range text {
		switch ch {
		case '\\':
			b.WriteString(`\\`)
		case '`':
			b.WriteString("\\`")
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '$':
			if i+1 < len(text) && text[i+1] == '{' {
				b.WriteString(`\$`)
			} else {
				b.WriteRune(ch)
			}
		default:
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func (g *generator) indexAssignExpr(target *ir.IndexExpr, value ir.Expr) (string, bool) {
	if name, ok := g.reactiveIdentifier(target.Receiver); ok {
		return fmt.Sprintf("%s.mutate((__value) => (__value[%s] = %s))", name, g.expr(target.Index), g.expr(value)), true
	}
	if _, _, ok := checker.MapKeyValue(target.Receiver.ResultType()); !ok {
		return "", false
	}
	return fmt.Sprintf("%s.set(%s, %s)", g.expr(target.Receiver), g.expr(target.Index), g.expr(value)), true
}

func (g *generator) selectorAssignExpr(target *ir.SelectorExpr, value ir.Expr) (string, bool) {
	name, ok := g.reactiveIdentifier(target.Receiver)
	if !ok {
		return "", false
	}
	return fmt.Sprintf("%s.mutate((__value) => (%s = %s))", name, tsPropertyAccess("__value", target.Name), g.expr(value)), true
}

func (g *generator) mapLiteral(lit *ir.MapLiteral) string {
	keyType, valueType, ok := checker.MapKeyValue(lit.ResultType())
	if !ok {
		keyType = checker.Unknown
		valueType = checker.Unknown
	}
	entries := make([]string, 0, len(lit.Entries))
	for _, entry := range lit.Entries {
		entries = append(entries, fmt.Sprintf("[%s, %s]", g.expr(entry.Key), g.expr(entry.Value)))
	}
	return fmt.Sprintf("new Map<%s, %s>([%s])", tsType(keyType), tsType(valueType), strings.Join(entries, ", "))
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
					return fmt.Sprintf("{ tag: %s, payload: [] }", tsPropertyAccess(mangleIdent(enum.Name), member.Name)), true
				}
				return tsPropertyAccess(mangleIdent(enum.Name), member.Name), true
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
	name := baseTypeName(typ)
	for _, enum := range g.file.Enums {
		if enum.Name == name {
			return enum
		}
	}
	return nil
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

func selectorResolvedName(sel *ir.SelectorExpr) string {
	if sel.ResolvedName != "" {
		return sel.ResolvedName
	}
	return sel.Name
}

func (g *generator) callExpr(e *ir.CallExpr) string {
	raw := g.callExprRaw(e)
	if e.Await {
		return "await " + raw
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
			return fmt.Sprintf("runeOk<%s, %s>(%s)", tsType(okType), tsType(errType), g.expr(e.Args[0]))
		case "Err":
			if len(e.Args) != 1 {
				return g.zeroValue(e.ResultType())
			}
			okType, errType := resultTypeArgs(e.ResultType())
			return fmt.Sprintf("runeErr<%s, %s>(%s)", tsType(okType), tsType(errType), g.expr(e.Args[0]))
		}
	}
	if constructor, ok := g.enumConstructorCall(e); ok {
		return constructor
	}
	if call, ok := g.moduleIntrinsicCall(e); ok {
		return call
	}
	if call, ok := g.receiverIntrinsicCall(e); ok {
		return call
	}
	if call, ok := g.arrayMethodCall(e); ok {
		return call
	}
	if call, ok := g.iterMethodCall(e); ok {
		return call
	}
	if call, ok := g.primitiveMethodCall(e); ok {
		return call
	}
	if call, ok := g.methodCall(e); ok {
		return call
	}
	args := make([]string, 0, len(e.Args))
	for _, arg := range e.Args {
		args = append(args, g.expr(arg))
	}
	return fmt.Sprintf("%s(%s)", g.callCalleeExpr(e.Callee), strings.Join(args, ", "))
}

func (g *generator) callCalleeExpr(expr ir.Expr) string {
	out := g.expr(expr)
	switch expr.(type) {
	case *ir.TernaryExpr, *ir.AssignExpr, *ir.BlockExpr:
		return "(" + out + ")"
	default:
		return out
	}
}

func (g *generator) enumConstructorCall(call *ir.CallExpr) (string, bool) {
	name := ""
	if ident, ok := call.Callee.(*ir.Identifier); ok {
		name = ident.Name
	} else if sel, ok := call.Callee.(*ir.SelectorExpr); ok {
		name = sel.Name
	}
	if name == "" {
		return "", false
	}
	enum, member, ok := g.enumMemberForConstructor(call.ResultType(), name)
	if !ok || !enumHasPayload(enum) {
		return "", false
	}
	args := make([]string, 0, len(call.Args))
	for _, arg := range call.Args {
		args = append(args, g.expr(arg))
	}
	return fmt.Sprintf("{ tag: %s, payload: [%s] }", tsPropertyAccess(mangleIdent(enum.Name), member.Name), strings.Join(args, ", ")), true
}

func (g *generator) iterMethodCall(call *ir.CallExpr) (string, bool) {
	sel, ok := call.Callee.(*ir.SelectorExpr)
	if !ok {
		return "", false
	}
	if _, ok := checker.IterValue(sel.Receiver.ResultType()); !ok {
		return "", false
	}
	receiver := g.expr(sel.Receiver)
	switch sel.Name {
	case "toArray":
		return fmt.Sprintf("(() => { const iter = %s; const out = []; for (;;) { const item = iter.next(); if (!item[1]) return out; out.push(item[0]); } })()", receiver), true
	case "each":
		if len(call.Args) != 1 {
			return g.zeroValue(call.ResultType()), true
		}
		callback, arity := g.iterCallback(call.Args[0])
		return fmt.Sprintf("(() => { const iter = %s; let index = 0; for (;;) { const item = iter.next(); if (!item[1]) return null; %s; index += 1; } })()", receiver, iterCallbackCall(callback, arity, "item[0]", "index", "iter")), true
	case "map":
		if len(call.Args) != 1 {
			return g.zeroValue(call.ResultType()), true
		}
		callback, arity := g.iterCallback(call.Args[0])
		return fmt.Sprintf("(() => { const iter = %s; const out = []; let index = 0; for (;;) { const item = iter.next(); if (!item[1]) return out; out.push(%s); index += 1; } })()", receiver, iterCallbackCall(callback, arity, "item[0]", "index", "iter")), true
	default:
		return "", false
	}
}

func (g *generator) iterCallback(expr ir.Expr) (string, int) {
	if lambda, ok := expr.(*ir.LambdaExpr); ok {
		return g.expr(lambda), len(lambda.Params)
	}
	params, _, ok := parseTSFuncType(string(expr.ResultType()))
	if !ok {
		return g.expr(expr), 3
	}
	return g.expr(expr), len(params)
}

func iterCallbackCall(callback string, arity int, value string, index string, iter string) string {
	args := []string{value, index, iter}
	if arity < len(args) {
		args = args[:arity]
	}
	return fmt.Sprintf("(%s)(%s)", callback, strings.Join(args, ", "))
}

func resultTypeArgs(typ checker.Type) (checker.Type, checker.Type) {
	base, args, ok := parseTSGenericType(string(typ))
	if !ok || base != "Result" || len(args) != 2 {
		return checker.Unknown, checker.Unknown
	}
	return checker.Type(args[0]), checker.Type(args[1])
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

func (g *generator) signalInitialValue(expr ir.Expr) string {
	switch expr.(type) {
	case *ir.ArrayLiteral:
		return "runeReactiveArray(" + g.expr(expr) + ")"
	case *ir.AnonymousObjectLiteral:
		return "runeReactiveObject(" + g.expr(expr) + ")"
	default:
		return "runeSignal(" + g.expr(expr) + ")"
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
	if name, ok := g.reactiveIdentifier(sel.Receiver); ok {
		switch sel.Name {
		case "push":
			return fmt.Sprintf("%s.mutate((__value) => __value.push(%s))", name, strings.Join(args, ", ")), true
		}
	}
	switch sel.Name {
	case "length":
		return receiver + ".length", true
	case "isEmpty":
		return receiver + ".length === 0", true
	case "push":
		return fmt.Sprintf("%s.push(%s)", receiver, strings.Join(args, ", ")), true
	case "each":
		if len(args) != 1 {
			return "undefined", true
		}
		index := g.nextTemp("__arrayIndex")
		value := g.nextTemp("__arrayValue")
		return fmt.Sprintf("(() => { for (const [%s, %s] of %s.entries()) { (%s)(%s, %s, %s); } })()", index, value, receiver, args[0], value, index, receiver), true
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
			return receiver + ".length", true
		case "isEmpty":
			return receiver + ".length === 0", true
		case "toString":
			return receiver, true
		case "concat", "includes", "startsWith", "endsWith", "indexOf", "lastIndexOf", "toLowerCase", "toUpperCase", "trim", "trimStart", "trimEnd", "repeat", "replace", "replaceAll", "split":
			return fmt.Sprintf("%s.%s(%s)", receiver, sel.Name, strings.Join(args, ", ")), true
		}
	case checker.Char:
		switch sel.Name {
		case "toString":
			return receiver, true
		}
	case checker.Bool:
		switch sel.Name {
		case "not":
			return "!" + receiver, true
		case "xor":
			if len(args) != 1 {
				return "undefined", true
			}
			return fmt.Sprintf("%s !== %s", receiver, args[0]), true
		case "toString":
			return receiver + ".toString()", true
		}
	case checker.Regex:
		return g.regexMethodCall(receiver, sel.Name, args)
	}
	return "", false
}

func (g *generator) regexMethodCall(receiver string, name string, args []string) (string, bool) {
	switch name {
	case "exec":
		if len(args) != 1 {
			return "undefined", true
		}
		return fmt.Sprintf("((__match) => __match ? Array.from(__match, (__value) => __value ?? \"\") : [])(%s.exec(%s))", receiver, args[0]), true
	case "match":
		if len(args) != 1 {
			return "undefined", true
		}
		return fmt.Sprintf("((__match) => __match ? Array.from(__match, (__value) => __value ?? \"\") : [])(%s.match(%s))", args[0], receiver), true
	case "matchAll":
		if len(args) != 1 {
			return "undefined", true
		}
		return fmt.Sprintf("((__regex) => Array.from(%s.matchAll(__regex.global ? __regex : new RegExp(__regex.source, __regex.flags + \"g\")), (__match) => Array.from(__match, (__value) => __value ?? \"\")))(%s)", args[0], receiver), true
	case "test":
		if len(args) != 1 {
			return "undefined", true
		}
		return fmt.Sprintf("%s.test(%s)", receiver, args[0]), true
	case "replace":
		if len(args) != 2 {
			return "undefined", true
		}
		return fmt.Sprintf("%s.replace(%s, %s)", args[0], receiver, args[1]), true
	case "replaceAll":
		if len(args) != 2 {
			return "undefined", true
		}
		return fmt.Sprintf("((__regex) => %s.replaceAll(__regex.global ? __regex : new RegExp(__regex.source, __regex.flags + \"g\"), %s))(%s)", args[0], args[1], receiver), true
	case "search":
		if len(args) != 1 {
			return "undefined", true
		}
		return fmt.Sprintf("%s.search(%s)", args[0], receiver), true
	case "split":
		if len(args) != 1 {
			return "undefined", true
		}
		return fmt.Sprintf("%s.split(%s)", args[0], receiver), true
	case "source", "flags", "global", "ignoreCase", "multiline", "dotAll", "unicode", "sticky", "hasIndices", "lastIndex":
		if len(args) != 0 {
			return "undefined", true
		}
		return fmt.Sprintf("%s.%s", receiver, regexPropertyName(name)), true
	case "unicodeSets":
		if len(args) != 0 {
			return "undefined", true
		}
		return fmt.Sprintf("(%s as RegExp & { unicodeSets?: boolean }).unicodeSets ?? false", receiver), true
	case "setLastIndex":
		if len(args) != 1 {
			return "undefined", true
		}
		return fmt.Sprintf("(%s.lastIndex = %s)", receiver, args[0]), true
	default:
		return "", false
	}
}

func regexPropertyName(name string) string {
	switch name {
	case "hasIndices":
		return "hasIndices"
	default:
		return name
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
		if method.Name != sel.Name || method.Static != sel.Static {
			continue
		}
		args := []string{}
		if !method.Static {
			args = append(args, g.expr(sel.Receiver))
		}
		for _, arg := range call.Args {
			args = append(args, g.expr(arg))
		}
		return fmt.Sprintf("%s(%s)", mangleMethod(typeName, sel.Name), strings.Join(args, ", ")), true
	}
	return "", false
}

func (g *generator) lookupFunction(name string) *ir.Function {
	if g.file == nil {
		return nil
	}
	for _, fn := range g.file.Functions {
		if fn.Name == name {
			return fn
		}
	}
	return nil
}

func (g *generator) matchExpr(match *ir.MatchExpr) string {
	ret := tsType(match.ResultType())
	subject := g.expr(match.Subject)
	var b strings.Builder
	b.WriteString("((): ")
	b.WriteString(ret)
	b.WriteString(" => { ")
	if matchNeedsSubjectTemp(match) {
		subject = g.nextTemp("__match")
		b.WriteString("const ")
		b.WriteString(subject)
		b.WriteString(" = ")
		b.WriteString(g.expr(match.Subject))
		b.WriteString("; ")
	}
	restoreMapGetters := g.pushMapPatternGetters(subject, match.Branches)
	defer restoreMapGetters()
	for _, line := range g.mapPatternGetterPrelude(subject, match.Branches) {
		b.WriteString(line)
		b.WriteString("; ")
	}
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
		b.WriteString(g.patternBinding(subject, branch.Pattern))
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
		b.WriteString(g.zeroValue(match.ResultType()))
		b.WriteString("; ")
	}
	b.WriteString("})()")
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
			getter = g.nextTemp("__mapGet")
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
		getter = g.nextTemp("__mapGet")
	}
	keyType := checker.Unknown
	if len(pattern.Entries) > 0 {
		keyType = pattern.Entries[0].Key.ResultType()
	}
	cache := g.nextTemp("__mapCache")
	key := g.nextTemp("__key")
	value := g.nextTemp("__value")
	return []string{
		fmt.Sprintf("const %s = new Map<%s, any>()", cache, tsType(keyType)),
		fmt.Sprintf("const %s = (%s: %s): any => { if (%s.has(%s)) return %s.get(%s); const %s = %s(%s, %s); %s.set(%s, %s); return %s }",
			getter, key, tsType(keyType), cache, key, cache, key, value, mangleMethod(baseTypeName(pattern.SubjectType), "get"), subject, key, cache, key, value, value),
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
		*parts = append(*parts, fmt.Sprintf("const %s = %s;", mangleIdent(p.Name), subject))
	case *ir.TuplePattern:
		for idx, elem := range p.Elements {
			g.appendPatternBindings(parts, fmt.Sprintf("%s[%d]", subject, idx), elem)
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
			g.appendPatternBindings(parts, tsSequenceIndex(subject, p.SubjectType, g.arrayPatternElementIndex(subject, p, idx)), elem)
		}
		if p.RestBinding != "" {
			if irArrayPatternHasBits(p) {
				start := bitPatternRequiredBits(p) / 8
				*parts = append(*parts, fmt.Sprintf("const %s = %s;", mangleIdent(p.RestBinding), tsSequenceSlice(subject, p.SubjectType, start, 0)))
			} else {
				start := g.arrayPatternPrefixWidth(p, p.RestIndex)
				end := tsSubtractExpr(tsSequenceLength(subject, p.SubjectType), g.arrayPatternSuffixWidth(p, p.RestIndex))
				*parts = append(*parts, fmt.Sprintf("const %s = %s;", mangleIdent(p.RestBinding), tsSequenceSliceExpr(subject, p.SubjectType, start, end)))
			}
		}
	case *ir.BitPattern:
		g.appendPatternBindings(parts, subject, p.Value)
	case *ir.AsPattern:
		g.appendPatternBindings(parts, subject, p.Pattern)
		*parts = append(*parts, fmt.Sprintf("const %s = %s;", mangleIdent(p.Name), subject))
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
			value := fmt.Sprintf("%s.get(%s)!", subject, g.expr(entry.Key))
			if p.Access == "get" {
				value = g.tsMapLikeGet(subject, p.SubjectType, g.expr(entry.Key))
			} else if p.Access == "object" || p.SubjectType == checker.Object {
				if key, ok := entry.Key.(*ir.StringLiteral); ok {
					value = fmt.Sprintf("(%s as any)[%q]", subject, key.Value)
				}
			} else if entry.Optional {
				value = fmt.Sprintf("(%s.has(%s) ? %s.get(%s)! : null)", subject, g.expr(entry.Key), subject, g.expr(entry.Key))
			}
			g.appendPatternBindings(parts, value, entry.Pattern)
		}
	case *ir.ObjectPattern:
		for _, field := range p.Fields {
			if field.Optional && !field.Exists {
				continue
			}
			value := tsPropertyAccess(subject, field.Name)
			if p.SubjectType == checker.Object || p.SubjectType == checker.Unknown {
				value = fmt.Sprintf("(%s as any)[%q]", subject, field.Name)
			}
			g.appendPatternBindings(parts, value, field.Pattern)
		}
	}
}

func patternIndex(idx int, pattern *ir.ArrayPattern) string {
	if pattern.RestIndex < 0 || idx < pattern.RestIndex {
		return fmt.Sprintf("%d", idx)
	}
	tail := len(pattern.Elements) - idx
	return fmt.Sprintf("%s - %d", tsSequenceLength("_", pattern.SubjectType), tail)
}

func (g *generator) arrayPatternElementIndex(subject string, pattern *ir.ArrayPattern, idx int) string {
	if pattern.RestIndex < 0 || idx < pattern.RestIndex {
		return g.arrayPatternPrefixWidth(pattern, idx)
	}
	return tsSubtractExpr(tsSequenceLength(subject, pattern.SubjectType), g.arrayPatternSuffixWidth(pattern, idx))
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
			terms = append(terms, tsSequenceLength(g.expr(spread.Value), spread.Type))
			continue
		}
		constants++
	}
	if constants > 0 || len(terms) == 0 {
		terms = append([]string{fmt.Sprintf("%d", constants)}, terms...)
	}
	return strings.Join(terms, " + ")
}

func tsSequenceLength(subject string, typ checker.Type) string {
	switch typ {
	case checker.String:
		return fmt.Sprintf("Array.from(%s).length", subject)
	case checker.Bytes:
		return subject + ".byteLength"
	default:
		return subject + ".length"
	}
}

func tsSubtractExpr(left string, right string) string {
	if right == "0" {
		return left
	}
	if strings.ContainsAny(right, "+-*/ ") {
		return fmt.Sprintf("%s - (%s)", left, right)
	}
	return fmt.Sprintf("%s - %s", left, right)
}

func tsSequenceIndex(subject string, typ checker.Type, index string) string {
	index = strings.ReplaceAll(index, tsSequenceLength("_", typ), tsSequenceLength(subject, typ))
	switch typ {
	case checker.String:
		return fmt.Sprintf("(Array.from(%s)[%s] ?? \"\")", subject, index)
	case checker.Bytes:
		return fmt.Sprintf("%s.getUint8(%s)", subject, index)
	default:
		return fmt.Sprintf("%s[%s]", subject, index)
	}
}

func tsSequenceSlice(subject string, typ checker.Type, restIndex int, tailCount int) string {
	end := tsSequenceLength(subject, typ)
	if tailCount > 0 {
		end = fmt.Sprintf("%s - %d", end, tailCount)
	}
	return tsSequenceSliceExpr(subject, typ, fmt.Sprintf("%d", restIndex), end)
}

func tsSequenceSliceExpr(subject string, typ checker.Type, start string, end string) string {
	switch typ {
	case checker.String:
		return fmt.Sprintf("Array.from(%s).slice(%s, %s).join(\"\")", subject, start, end)
	case checker.Bytes:
		return fmt.Sprintf("new DataView(%s.buffer.slice(%s.byteOffset + %s, %s.byteOffset + %s))", subject, subject, start, subject, end)
	default:
		return fmt.Sprintf("%s.slice(%s, %s)", subject, start, end)
	}
}

func (g *generator) watchExpr(watch *ir.WatchExpr) string {
	target, ok := watch.Target.(*ir.Identifier)
	if !ok || (!g.isSignal(target.Name) && !g.isReactive(target.Name)) {
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

func (g *generator) reactiveIdentifier(expr ir.Expr) (string, bool) {
	ident, ok := expr.(*ir.Identifier)
	if !ok {
		return "", false
	}
	if g.isReactive(ident.Name) {
		return mangleIdent(ident.Name), true
	}
	if g.isSignal(ident.Name) {
		if _, ok := checker.ArrayElement(expr.ResultType()); ok {
			return mangleIdent(ident.Name), true
		}
		typeName := string(expr.ResultType())
		if strings.HasPrefix(typeName, "{") && strings.HasSuffix(typeName, "}") {
			return mangleIdent(ident.Name), true
		}
	}
	return "", false
}

func (g *generator) letStmtValue(stmt *ir.LetStmt) string {
	if stmt.Signal || g.exprUsesSignal(stmt.Value) {
		g.addSignal(stmt.Name, stmt.Value.ResultType())
		deps := g.exprSignalDeps(stmt.Value)
		g.setSignalDeps(stmt.Name, deps)
		return g.signalInitialValue(stmt.Value)
	}
	if _, ok := stmt.Value.(*ir.AnonymousObjectLiteral); ok {
		return g.withThisName(mangleIdent(stmt.Name), func() string {
			return g.expr(stmt.Value)
		})
	}
	return g.expr(stmt.Value)
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
			value := g.letStmtValue(s)
			if _, ok := s.Value.(*ir.ReactiveLiteral); ok {
				parts = append(parts, fmt.Sprintf("%s %s = %s", kind, mangleIdent(s.Name), value))
				continue
			}
			parts = append(parts, fmt.Sprintf("%s %s = %s", kind, mangleIdent(s.Name), value))
		case *ir.AssignStmt:
			if g.isSignal(s.Name) {
				parts = append(parts, fmt.Sprintf("%s.set(%s)", mangleIdent(s.Name), g.expr(s.Value)))
			} else if g.isReactive(s.Name) {
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
	case lexer.QuestionQuestion, lexer.OrOr:
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
