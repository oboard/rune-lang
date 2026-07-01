package moonbitcodegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
	"github.com/oboard/rune-lang/internal/lexer"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func (g *generator) expr(expr ir.Expr) string {
	return g.exprPrec(expr, 0)
}

func (g *generator) exprPrec(expr ir.Expr, parentPrec int) string {
	switch e := expr.(type) {
	case *ir.Identifier:
		return mangleIdent(e.Name)
	case *ir.AtExpr:
		return "()"
	case *ir.ThisExpr:
		if len(g.thisNames) > 0 {
			return g.thisNames[len(g.thisNames)-1]
		}
		return "self"
	case *ir.IntegerLiteral:
		if e.Value > 2147483647 || e.Value < -2147483648 {
			return fmt.Sprintf("BigInt::from_string(%s)", quoteString(strconv.Itoa(e.Value)))
		}
		return strconv.Itoa(e.Value)
	case *ir.DoubleLiteral:
		if e.Raw != "" && !strings.ContainsAny(e.Raw, "eE") {
			return e.Raw
		}
		return moonBitDoubleLiteral(e.Value)
	case *ir.BigIntLiteral:
		return fmt.Sprintf("BigInt::from_string(%s)", quoteString(e.Value))
	case *ir.StringLiteral:
		return quoteString(e.Value)
	case *ir.TemplateLiteral:
		return g.templateLiteral(e)
	case *ir.CharLiteral:
		return quoteChar(e.Value)
	case *ir.RegexLiteral:
		g.useRegex = true
		return fmt.Sprintf("rune_regex_new(%s, %s)", quoteString(e.Pattern), quoteString(e.Flags))
	case *ir.BoolLiteral:
		if e.Value {
			return "true"
		}
		return "false"
	case *ir.NullLiteral:
		return "None"
	case *ir.UnaryExpr:
		if e.Op == lexer.Bang {
			out := g.mbtNotExpr(e.Expr)
			if 6 < parentPrec {
				return "(" + out + ")"
			}
			return out
		}
		if e.Op == lexer.Tilde {
			return fmt.Sprintf("(%s).lnot()", g.expr(e.Expr))
		}
		op := e.Op.String()
		out := op + g.exprPrec(e.Expr, 6)
		if 6 < parentPrec {
			return "(" + out + ")"
		}
		return out
	case *ir.PostfixExpr:
		switch e.Op {
		case lexer.PlusPlus:
			return g.expr(e.Expr) + " += 1"
		default:
			g.addError(fmt.Errorf("MoonBit backend does not support postfix operator %s", e.Op))
			return g.expr(e.Expr)
		}
	case *ir.ResultUnwrapExpr:
		g.addError(fmt.Errorf("MoonBit backend does not support result unwrap yet"))
		return g.expr(e.Expr)
	case *ir.BinaryExpr:
		if e.Op == lexer.QuestionQuestion {
			return fmt.Sprintf("%s.unwrap_or(%s)", g.exprPrec(e.Left, 6), g.expr(e.Right))
		}
		if e.Left.ResultType() == checker.String && e.Right.ResultType() == checker.String && mbtComparisonOp(e.Op) {
			return g.stringComparisonExpr(e, parentPrec)
		}
		if e.Op == lexer.UnsignedShiftRight {
			out := fmt.Sprintf("((%s).reinterpret_as_uint() >> %s).reinterpret_as_int()", g.expr(e.Left), g.expr(e.Right))
			if 5 < parentPrec {
				return "(" + out + ")"
			}
			return out
		}
		if e.Left.ResultType() == checker.BigInt && (e.Op == lexer.ShiftLeft || e.Op == lexer.ShiftRight) {
			right := g.expr(e.Right)
			if e.Right.ResultType() == checker.BigInt {
				right = fmt.Sprintf("(%s).to_int()", right)
			}
			out := fmt.Sprintf("%s %s %s", g.expr(e.Left), mbtBinaryOp(e.Op), right)
			if 5 < parentPrec {
				return "(" + out + ")"
			}
			return out
		}
		prec := mbtPrecedence(e.Op)
		out := fmt.Sprintf("%s %s %s", g.mbtBinaryOperand(e.Left, e.Op, prec), mbtBinaryOp(e.Op), g.mbtBinaryOperand(e.Right, e.Op, prec+1))
		if prec < parentPrec {
			return "(" + out + ")"
		}
		return out
	case *ir.TernaryExpr:
		alt := "()"
		if e.Alternative != nil {
			alt = g.expr(e.Alternative)
		}
		out := fmt.Sprintf("if %s { %s } else { %s }", g.expr(e.Condition), g.expr(e.Consequence), alt)
		if parentPrec > 0 {
			return "(" + out + ")"
		}
		return out
	case *ir.AssignExpr:
		return g.assignExpr(e, true)
	case *ir.CallExpr:
		return g.callExpr(e)
	case *ir.LambdaExpr:
		return g.lambda(e)
	case *ir.SelectorExpr:
		if e.ResolvedName != "" {
			return mangleIdent(e.ResolvedName)
		}
		if ident, ok := e.Receiver.(*ir.Identifier); ok && g.importAliases[ident.Name] {
			return mangleIdent(e.Name)
		}
		if member, ok := g.enumMemberSelector(e); ok {
			return member
		}
		if e.Static {
			if ident, ok := e.Receiver.(*ir.Identifier); ok {
				return mangleMethod(ident.Name, e.Name)
			}
		}
		if _, ok := checker.ImportNamespacePath(e.Receiver.ResultType()); ok {
			return mangleIdent(selectorResolvedName(e))
		}
		if at, ok := e.Receiver.(*ir.AtExpr); ok {
			return "@" + at.Name + "." + e.Name
		}
		return g.selectorReceiverExpr(e.Receiver) + "." + mangleIdent(e.Name)
	case *ir.IndexExpr:
		if _, ok := checker.TupleElements(e.Receiver.ResultType()); ok {
			if index, ok := e.Index.(*ir.IntegerLiteral); ok {
				return fmt.Sprintf("%s.%d", g.selectorReceiverExpr(e.Receiver), index.Value)
			}
		}
		if _, _, ok := checker.MapKeyValue(e.Receiver.ResultType()); ok {
			return fmt.Sprintf("%s.get(%s)", g.selectorReceiverExpr(e.Receiver), g.expr(e.Index))
		}
		if e.Receiver.ResultType() == checker.String && e.ResultType() == checker.Char {
			return fmt.Sprintf("%s[%s].unsafe_to_char()", g.selectorReceiverExpr(e.Receiver), g.expr(e.Index))
		}
		return fmt.Sprintf("%s[%s]", g.selectorReceiverExpr(e.Receiver), g.expr(e.Index))
	case *ir.ArrayLiteral:
		return g.arrayLiteralAs(e, e.ResultType())
	case *ir.TupleLiteral:
		elems := make([]string, 0, len(e.Elements))
		for _, elem := range e.Elements {
			elems = append(elems, g.expr(elem))
		}
		return "(" + strings.Join(elems, ", ") + ")"
	case *ir.MapLiteral:
		if len(e.Entries) == 0 {
			return "{}"
		}
		parts := make([]string, 0, len(e.Entries))
		for _, entry := range e.Entries {
			parts = append(parts, fmt.Sprintf("%s: %s", g.expr(entry.Key), g.expr(entry.Value)))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case *ir.SpreadExpr:
		g.addError(fmt.Errorf("MoonBit backend does not support spread expressions yet"))
		return g.expr(e.Expr)
	case *ir.ReactiveLiteral:
		g.addError(fmt.Errorf("MoonBit backend does not support reactive literals"))
		return g.expr(e.Value)
	case *ir.StructLiteral:
		if e.TypeName == string(checker.Error) {
			return g.errorLiteral(e)
		}
		if e.TypeName == "Map" && len(e.Fields) == 0 {
			return "{}"
		}
		fields := make([]string, 0, len(e.Fields))
		for _, field := range e.Fields {
			value := g.expr(field.Value)
			if fieldType, ok := g.structFieldType(e.TypeName, field.Name); ok {
				value = g.exprAs(field.Value, fieldType)
			}
			fields = append(fields, fmt.Sprintf("%s: %s", mangleIdent(field.Name), value))
		}
		return fmt.Sprintf("%s::{ %s }", mangleType(e.TypeName), strings.Join(fields, ", "))
	case *ir.AnonymousObjectLiteral:
		return g.anonymousObjectLiteral(e)
	case *ir.BlockExpr:
		return "{ " + g.blockInline(e, e.ResultType()) + " }"
	case *ir.PatternBlock:
		g.addError(fmt.Errorf("MoonBit backend only supports pattern blocks as function bodies"))
		return "()"
	case *ir.MatchExpr:
		return g.matchExpr(e)
	case *ir.WatchExpr:
		g.addError(fmt.Errorf("MoonBit backend does not support watch expressions"))
		return "()"
	case *ir.XMLElement:
		g.addError(fmt.Errorf("MoonBit backend does not support XML expressions"))
		return "()"
	default:
		return zeroValue(expr.ResultType())
	}
}

func moonBitDoubleLiteral(value float64) string {
	out := strconv.FormatFloat(value, 'f', -1, 64)
	if !strings.Contains(out, ".") {
		out += ".0"
	}
	return out
}

func (g *generator) stringComparisonExpr(e *ir.BinaryExpr, parentPrec int) string {
	g.useString = true
	cmp := fmt.Sprintf("rune_string_compare(%s, %s)", g.expr(e.Left), g.expr(e.Right))
	var out string
	switch e.Op {
	case lexer.EqualEqual:
		out = fmt.Sprintf("%s == 0", cmp)
	case lexer.BangEqual:
		out = fmt.Sprintf("%s != 0", cmp)
	case lexer.Less:
		out = fmt.Sprintf("%s < 0", cmp)
	case lexer.LessEqual:
		out = fmt.Sprintf("%s <= 0", cmp)
	case lexer.Greater:
		out = fmt.Sprintf("%s > 0", cmp)
	case lexer.GreaterEqual:
		out = fmt.Sprintf("%s >= 0", cmp)
	default:
		return zeroValue(e.ResultType())
	}
	if mbtPrecedence(e.Op) < parentPrec {
		return "(" + out + ")"
	}
	return out
}

func (g *generator) mbtBinaryOperand(expr ir.Expr, parentOp lexer.Kind, parentPrec int) string {
	if child, ok := expr.(*ir.BinaryExpr); ok && mbtNeedsBinaryOperandParens(parentOp, child.Op) {
		return "(" + g.expr(child) + ")"
	}
	return g.exprPrec(expr, parentPrec)
}

func (g *generator) mbtNotExpr(expr ir.Expr) string {
	switch expr.(type) {
	case *ir.BinaryExpr, *ir.TernaryExpr, *ir.AssignExpr:
		return "!(" + g.expr(expr) + ")"
	default:
		return "!" + g.exprPrec(expr, 6)
	}
}

func mbtNeedsBinaryOperandParens(parent lexer.Kind, child lexer.Kind) bool {
	if (parent == lexer.AndAnd || parent == lexer.OrOr) && parent != child {
		return child == lexer.AndAnd || child == lexer.OrOr || mbtComparisonOp(child)
	}
	return mbtComparisonOp(parent) && mbtComparisonOp(child)
}

func mbtComparisonOp(op lexer.Kind) bool {
	switch op {
	case lexer.EqualEqual, lexer.BangEqual, lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
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

func (g *generator) withThisName(name string, fn func() string) string {
	g.thisNames = append(g.thisNames, name)
	defer func() {
		g.thisNames = g.thisNames[:len(g.thisNames)-1]
	}()
	return fn()
}

func (g *generator) anonymousObjectLiteral(lit *ir.AnonymousObjectLiteral) string {
	return g.anonymousObjectLiteralAs(lit, lit.ResultType())
}

func (g *generator) anonymousObjectLiteralAs(lit *ir.AnonymousObjectLiteral, typ checker.Type) string {
	expected := map[string]checker.Type{}
	if fields, ok := parseObjectType(string(typ)); ok {
		for _, field := range fields {
			expected[field.name] = checker.Type(field.typ)
		}
	}
	fields := make([]string, 0, len(lit.Fields))
	for _, field := range lit.Fields {
		value := g.expr(field.Value)
		if fieldType, ok := expected[field.Name]; ok {
			value = g.exprAs(field.Value, fieldType)
		}
		fields = append(fields, fmt.Sprintf("%s: %s", mangleIdent(field.Name), value))
	}
	return fmt.Sprintf("%s::{ %s }", g.anonymousTypeName(typ), strings.Join(fields, ", "))
}

func (g *generator) anonymousObjectLiteralWithFunctionPlaceholders(lit *ir.AnonymousObjectLiteral) string {
	fields := make([]string, 0, len(lit.Fields))
	for _, field := range lit.Fields {
		value := g.expr(field.Value)
		if replacement, ok := zeroFunctionValue(field.Value.ResultType()); ok {
			value = replacement
		}
		fields = append(fields, fmt.Sprintf("%s: %s", mangleIdent(field.Name), value))
	}
	return fmt.Sprintf("%s::{ %s }", g.anonymousTypeName(lit.ResultType()), strings.Join(fields, ", "))
}

func (g *generator) structFieldType(typeName string, fieldName string) (checker.Type, bool) {
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

func anonymousObjectHasFunctionFields(lit *ir.AnonymousObjectLiteral) bool {
	for _, field := range lit.Fields {
		if _, ok := zeroFunctionValue(field.Value.ResultType()); ok {
			return true
		}
	}
	return false
}

func (g *generator) arrayLiteral(lit *ir.ArrayLiteral) string {
	return g.arrayLiteralAs(lit, lit.ResultType())
}

func (g *generator) arrayLiteralAs(lit *ir.ArrayLiteral, expected checker.Type) string {
	elemType := checker.Unknown
	if elem, ok := checker.ArrayElement(expected); ok {
		elemType = elem
	} else if elem, ok := checker.ArrayElement(lit.ResultType()); ok {
		elemType = elem
	}
	elems := make([]string, 0, len(lit.Elements))
	for _, elem := range lit.Elements {
		if spread, ok := elem.(*ir.SpreadExpr); ok {
			elems = append(elems, ".."+g.expr(spread.Expr))
			continue
		}
		elems = append(elems, g.exprAs(elem, elemType))
	}
	return "[" + strings.Join(elems, ", ") + "]"
}

func (g *generator) exprAs(expr ir.Expr, expected checker.Type) string {
	if ternary, ok := expr.(*ir.TernaryExpr); ok {
		return g.ternaryExprAs(ternary, expected)
	}
	if expr.ResultType() == expected {
		return g.expr(expr)
	}
	if expected == checker.Object {
		return g.jsonValueFromSource(g.expr(expr), expr.ResultType())
	}
	if inner, ok := parseNullableType(string(expected)); ok {
		if _, ok := expr.(*ir.NullLiteral); ok {
			return "None"
		}
		return "Some(" + g.exprAs(expr, checker.Type(inner)) + ")"
	}
	if lit, ok := expr.(*ir.ArrayLiteral); ok {
		return g.arrayLiteralAs(lit, expected)
	}
	if lit, ok := expr.(*ir.AnonymousObjectLiteral); ok {
		if _, ok := parseObjectType(string(expected)); ok {
			return g.anonymousObjectLiteralAs(lit, expected)
		}
	}
	if _, ok := parseObjectType(string(expected)); ok {
		if _, sourceOk := parseObjectType(string(expr.ResultType())); sourceOk {
			return g.objectAsExpected(expr, expected)
		}
	}
	return g.expr(expr)
}

func (g *generator) ternaryExprAs(expr *ir.TernaryExpr, expected checker.Type) string {
	alt := zeroValue(expected)
	if expr.Alternative != nil {
		alt = g.exprAs(expr.Alternative, expected)
	}
	return fmt.Sprintf("if %s { %s } else { %s }", g.expr(expr.Condition), g.exprAs(expr.Consequence, expected), alt)
}

func (g *generator) objectAsExpected(expr ir.Expr, expected checker.Type) string {
	fields, _ := parseObjectType(string(expected))
	values := make([]string, 0, len(fields))
	for _, field := range fields {
		values = append(values, fmt.Sprintf("%s: %s", mangleIdent(field.name), g.exprAs(&ir.SelectorExpr{
			ExprBase: ir.ExprBase{Type: checker.Type(field.typ)},
			Receiver: expr,
			Name:     field.name,
		}, checker.Type(field.typ))))
	}
	return fmt.Sprintf("%s::{ %s }", g.anonymousTypeName(expected), strings.Join(values, ", "))
}

func (g *generator) errorLiteral(lit *ir.StructLiteral) string {
	for _, field := range lit.Fields {
		if field.Name == "message" {
			return g.expr(field.Value)
		}
	}
	return quoteString("")
}

func (g *generator) templateLiteral(lit *ir.TemplateLiteral) string {
	parts := make([]string, 0, len(lit.Parts))
	for _, part := range lit.Parts {
		if part.Text != "" {
			parts = append(parts, quoteString(part.Text))
		}
		if part.Expr != nil {
			parts = append(parts, g.showExpr(part.Expr))
		}
	}
	if len(parts) == 0 {
		return "\"\""
	}
	return strings.Join(parts, " + ")
}

func (g *generator) showExpr(expr ir.Expr) string {
	if expr.ResultType() == checker.String {
		return g.expr(expr)
	}
	return fmt.Sprintf("(%s).to_string()", g.expr(expr))
}

func (g *generator) assignExpr(e *ir.AssignExpr, keepValue bool) string {
	tail := "()"
	if keepValue {
		tail = "__value"
	}
	if target, ok := e.Target.(*ir.IndexExpr); ok {
		value := g.nextTemp("__value")
		tail = "()"
		if keepValue {
			tail = value
		}
		if _, _, ok := checker.MapKeyValue(target.Receiver.ResultType()); ok {
			return fmt.Sprintf("{ let %s = %s; let _ = %s.set(%s, %s); %s }", value, g.exprAs(e.Value, target.ResultType()), g.selectorReceiverExpr(target.Receiver), g.expr(target.Index), value, tail)
		}
		return fmt.Sprintf("{ let %s = %s; let _ = { %s[%s] = %s }; %s }", value, g.exprAs(e.Value, target.ResultType()), g.expr(target.Receiver), g.expr(target.Index), value, tail)
	}
	if sel, ok := e.Target.(*ir.SelectorExpr); ok {
		if ident, ok := sel.Receiver.(*ir.Identifier); ok {
			if fieldType, found := g.structFieldType(string(sel.Receiver.ResultType()), sel.Name); found {
				value := g.nextTemp("__value")
				tail = "()"
				if keepValue {
					tail = value
				}
				return fmt.Sprintf("{ let %s = %s; %s = %s::{ ..%s, %s: %s }; %s }", value, g.exprAs(e.Value, fieldType), mangleIdent(ident.Name), mangleType(string(sel.Receiver.ResultType())), mangleIdent(ident.Name), mangleIdent(sel.Name), value, tail)
			}
		}
	}
	if e.Target != nil && e.Name == "" {
		return fmt.Sprintf("%s = %s", g.expr(e.Target), g.expr(e.Value))
	}
	return fmt.Sprintf("%s = %s", mangleIdent(e.Name), g.expr(e.Value))
}

func selectorResolvedName(sel *ir.SelectorExpr) string {
	if sel.ResolvedName != "" {
		return sel.ResolvedName
	}
	return sel.Name
}

func (g *generator) callExpr(call *ir.CallExpr) string {
	if out, ok := g.moduleIntrinsicCall(call); ok {
		return out
	}
	if out, ok := g.iterMethodCall(call); ok {
		return out
	}
	if out, ok := g.receiverIntrinsicCall(call); ok {
		return out
	}
	args := g.callArgs(call)
	if ident, ok := call.Callee.(*ir.Identifier); ok {
		if (ident.Name == "Ok" || ident.Name == "Err") && isResultType(call.ResultType()) {
			return ident.Name + "(" + strings.Join(args, ", ") + ")"
		}
		if ctor, ok := g.enumConstructorCall(ident.Name, call.ResultType()); ok {
			return ctor + "(" + strings.Join(args, ", ") + ")"
		}
	}
	if sel, ok := call.Callee.(*ir.SelectorExpr); ok {
		if ctor, ok := g.enumSelectorConstructorCall(sel, call.ResultType()); ok {
			return ctor + "(" + strings.Join(args, ", ") + ")"
		}
		if sel.Static {
			if ident, ok := sel.Receiver.(*ir.Identifier); ok {
				return mangleMethod(ident.Name, sel.Name) + "(" + strings.Join(args, ", ") + ")"
			}
			return g.expr(sel) + "(" + strings.Join(args, ", ") + ")"
		}
		if sel.ResolvedName != "" {
			return mangleIdent(sel.ResolvedName) + "(" + strings.Join(args, ", ") + ")"
		}
		if ident, ok := sel.Receiver.(*ir.Identifier); ok && g.importAliases[ident.Name] {
			return mangleIdent(sel.Name) + "(" + strings.Join(args, ", ") + ")"
		}
		if typ := g.structTypeFromReceiver(sel.Receiver); typ != nil {
			args = append([]string{g.expr(sel.Receiver)}, args...)
			return mangleMethod(typ.Name, sel.Name) + "(" + strings.Join(args, ", ") + ")"
		}
		if fieldType, ok := g.selectorFieldType(sel.Receiver.ResultType(), sel.Name); ok {
			if isFuncType(fieldType) {
				return "(" + g.expr(sel) + ")(" + strings.Join(args, ", ") + ")"
			}
		}
		return g.expr(sel.Receiver) + "." + mangleIdent(sel.Name) + "(" + strings.Join(args, ", ") + ")"
	}
	return g.callCalleeExpr(call.Callee) + "(" + strings.Join(args, ", ") + ")"
}

func (g *generator) enumSelectorConstructorCall(sel *ir.SelectorExpr, typ checker.Type) (string, bool) {
	ident, ok := sel.Receiver.(*ir.Identifier)
	if !ok {
		return "", false
	}
	enum := g.enumFromType(typ)
	if enum == nil || enum.Name != ident.Name {
		return "", false
	}
	for _, member := range enum.Members {
		if member.Name == sel.Name {
			return mangleType(enum.Name) + "::" + mangleType(member.Name), true
		}
	}
	return "", false
}

func (g *generator) selectorFieldType(receiver checker.Type, fieldName string) (checker.Type, bool) {
	if fields, ok := parseObjectType(string(receiver)); ok {
		for _, field := range fields {
			if field.name == fieldName {
				return checker.Type(field.typ), true
			}
		}
	}
	return g.structFieldType(string(receiver), fieldName)
}

func (g *generator) callArgs(call *ir.CallExpr) []string {
	params := g.callParamTypes(call)
	args := make([]string, 0, len(call.Args))
	for i, arg := range call.Args {
		if i < len(params) {
			args = append(args, g.exprAs(arg, params[i]))
			continue
		}
		args = append(args, g.expr(arg))
	}
	return args
}

func (g *generator) callParamTypes(call *ir.CallExpr) []checker.Type {
	if ident, ok := call.Callee.(*ir.Identifier); ok {
		for _, fn := range g.file.Functions {
			if fn.Name != ident.Name || len(fn.Params) != len(call.Args) {
				continue
			}
			params := make([]checker.Type, 0, len(fn.Params))
			for _, param := range fn.Params {
				params = append(params, param.Type)
			}
			return params
		}
	}
	return nil
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

func (g *generator) iterMethodCall(call *ir.CallExpr) (string, bool) {
	sel, ok := call.Callee.(*ir.SelectorExpr)
	if !ok {
		return "", false
	}
	if _, ok := checker.IterValue(sel.Receiver.ResultType()); !ok {
		return "", false
	}
	fn := &stdlib.Function{Name: sel.Name, Intrinsic: "iter." + sel.Name}
	return g.iterIntrinsicCall(call, fn, g.expr(sel.Receiver), g.intrinsicArgs(call.Args)), true
}

func (g *generator) structTypeFromReceiver(receiver ir.Expr) *ir.StructType {
	name := string(receiver.ResultType())
	for _, typ := range g.file.Types {
		if typ.Name == name {
			return typ
		}
	}
	return nil
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
			if member.Name != sel.Name {
				continue
			}
			return mangleType(enum.Name) + "::" + mangleType(member.Name), true
		}
	}
	return "", false
}

func (g *generator) enumConstructorCall(name string, typ checker.Type) (string, bool) {
	enum := g.enumFromType(typ)
	if enum == nil {
		return "", false
	}
	for _, member := range enum.Members {
		if member.Name == name {
			return mangleType(member.Name), true
		}
	}
	return "", false
}

func (g *generator) lambda(lambda *ir.LambdaExpr) string {
	params := make([]string, 0, len(lambda.Params))
	paramTypes, _ := lambdaParamTypes(lambda.ResultType())
	for i, param := range lambda.Params {
		out := mangleIdent(param)
		if i < len(paramTypes) {
			out = out + " : " + g.mbtType(checker.Type(paramTypes[i]))
		}
		params = append(params, out)
	}
	if block, ok := lambda.Body.(*ir.BlockExpr); ok {
		return "(" + strings.Join(params, ", ") + ") => { " + g.blockInline(block, block.ResultType()) + " }"
	}
	return "(" + strings.Join(params, ", ") + ") => " + g.expr(lambda.Body)
}

func lambdaParamTypes(typ checker.Type) ([]string, string) {
	name := string(typ)
	if params, ret, ok := parseFuncType(name); ok {
		return params, ret
	}
	if base, args, ok := parseGenericType(name); ok && (base == "Func" || base == "AsyncFunc") && len(args) > 0 {
		return args[:len(args)-1], args[len(args)-1]
	}
	return nil, ""
}

func (g *generator) blockInline(block *ir.BlockExpr, ret checker.Type) string {
	parts := make([]string, 0, len(block.Statements))
	for i, stmt := range block.Statements {
		last := i == len(block.Statements)-1
		switch s := stmt.(type) {
		case *ir.LetStmt:
			parts = append(parts, fmt.Sprintf("let %s = %s", mangleIdent(s.Name), g.expr(s.Value)))
		case *ir.AssignStmt:
			parts = append(parts, fmt.Sprintf("%s = %s", mangleIdent(s.Name), g.expr(s.Value)))
		case *ir.ExprStmt:
			if last && ret != checker.Void {
				parts = append(parts, g.expr(s.Expr))
			} else {
				parts = append(parts, g.discardExpr(s.Expr, g.expr(s.Expr)))
			}
		}
	}
	return strings.Join(parts, "; ")
}

func (g *generator) matchExpr(match *ir.MatchExpr) string {
	if !patternsRequireConditionLowering(match.Branches) {
		subject := g.nextTemp("__match")
		branches := make([]string, 0, len(match.Branches))
		covered, exhaustive := map[string]bool{}, false
		enum := g.enumFromType(match.Subject.ResultType())
		for _, branch := range match.Branches {
			if exhaustive {
				continue
			}
			branches = append(branches, fmt.Sprintf("%s => %s", g.patternFor(subject, branch.Pattern), g.expr(branch.Expr)))
			if enum != nil {
				exhaustive = updateEnumCoverage(enum, covered, branch.Pattern)
			}
		}
		return fmt.Sprintf("{ let %s = %s; match %s { %s } }", subject, g.expr(match.Subject), subject, strings.Join(branches, "; "))
	}
	subject := g.nextTemp("__match")
	parts := []string{fmt.Sprintf("let %s = %s", subject, g.expr(match.Subject))}
	chain := make([]string, 0, len(match.Branches)+1)
	hasDefault := false
	for idx, branch := range match.Branches {
		patterns := expandedBranchPatterns(branch.Pattern)
		for altIdx, pattern := range patterns {
			branchIdx := len(chain)
			prefix := "if"
			if branchIdx > 0 {
				prefix = "else if"
			}
			if _, ok := pattern.(*ir.WildcardPattern); ok {
				if altIdx > 0 {
					continue
				}
				_ = idx
				hasDefault = true
				if branchIdx == 0 {
					chain = append(chain, fmt.Sprintf("{ %s%s }", g.patternBinding(subject, match.Subject.ResultType(), pattern), g.expr(branch.Expr)))
				} else {
					chain = append(chain, fmt.Sprintf("else { %s%s }", g.patternBinding(subject, match.Subject.ResultType(), pattern), g.expr(branch.Expr)))
				}
				continue
			}
			condition := g.patternConditionAs(subject, match.Subject.ResultType(), pattern)
			chain = append(chain, fmt.Sprintf("%s %s { %s%s }", prefix, condition, g.patternBinding(subject, match.Subject.ResultType(), pattern), g.expr(branch.Expr)))
		}
		if _, ok := branch.Pattern.(*ir.WildcardPattern); ok {
			hasDefault = true
		}
	}
	if !hasDefault {
		chain = append(chain, fmt.Sprintf("else { %s }", g.zeroValue(match.ResultType())))
	}
	parts = append(parts, strings.Join(chain, " "))
	return "{ " + strings.Join(parts, "; ") + " }"
}

func (g *generator) patternBlock(fn *ir.Function, block *ir.PatternBlock, ret checker.Type) {
	if len(fn.Params) != 1 {
		g.addError(fmt.Errorf("%s: pattern blocks currently require exactly one parameter", block.Pos))
		g.line(zeroValue(ret))
		return
	}
	subject := mangleIdent(fn.Params[0].Name)
	if !patternsRequireConditionLowering(block.Branches) {
		enum := g.enumFromType(fn.Params[0].Type)
		covered, exhaustive := map[string]bool{}, false
		g.linef("match %s {", subject)
		g.indent++
		for _, branch := range block.Branches {
			if exhaustive {
				continue
			}
			g.linef("%s => %s", g.patternFor(subject, branch.Pattern), g.expr(branch.Expr))
			if enum != nil {
				exhaustive = updateEnumCoverage(enum, covered, branch.Pattern)
			}
		}
		g.indent--
		g.line("}")
		return
	}
	g.line("{")
	g.indent++
	hasDefault := false
	emitted := 0
	for _, branch := range block.Branches {
		for _, pattern := range expandedBranchPatterns(branch.Pattern) {
			if hasDefault {
				continue
			}
			prefix := "if"
			if emitted > 0 {
				prefix = "else if"
			}
			if _, ok := pattern.(*ir.WildcardPattern); ok {
				hasDefault = true
				if emitted == 0 {
					g.line("{")
				} else {
					g.line("else {")
				}
			} else {
				g.linef("%s %s {", prefix, g.patternConditionAs(subject, fn.Params[0].Type, pattern))
			}
			g.indent++
			if binding := g.patternBinding(subject, fn.Params[0].Type, pattern); binding != "" {
				for _, stmt := range splitInlineStatements(binding) {
					g.line(stmt)
				}
			}
			g.line(g.returnExpr(branch.Expr, ret))
			g.indent--
			g.line("}")
			emitted++
			if _, ok := pattern.(*ir.WildcardPattern); ok {
				hasDefault = true
			}
		}
	}
	if !hasDefault {
		g.linef("else { %s }", zeroValue(ret))
	}
	g.indent--
	g.line("}")
}

func (g *generator) enumFromType(typ checker.Type) *ir.EnumType {
	name := string(typ)
	for _, enum := range g.file.Enums {
		if enum.Name == name {
			return enum
		}
	}
	return nil
}

func updateEnumCoverage(enum *ir.EnumType, covered map[string]bool, pattern ir.Pattern) bool {
	members, all := enumPatternCoverage(enum, pattern)
	if all {
		return true
	}
	for _, member := range members {
		covered[member] = true
	}
	return len(covered) == len(enum.Members)
}

func enumPatternCoverage(enum *ir.EnumType, pattern ir.Pattern) ([]string, bool) {
	switch p := pattern.(type) {
	case *ir.WildcardPattern, *ir.BindingPattern:
		return nil, true
	case *ir.LiteralPattern:
		if sel, ok := p.Value.(*ir.SelectorExpr); ok {
			if ident, ok := sel.Receiver.(*ir.Identifier); ok && ident.Name == enum.Name {
				for _, member := range enum.Members {
					if member.Name == sel.Name {
						return []string{member.Name}, false
					}
				}
			}
		}
	case *ir.ConstructorPattern:
		for _, member := range enum.Members {
			if member.Name == p.Name {
				return []string{member.Name}, false
			}
		}
	case *ir.OrPattern:
		var out []string
		for _, alt := range p.Alternatives {
			members, all := enumPatternCoverage(enum, alt)
			if all {
				return nil, true
			}
			out = append(out, members...)
		}
		return out, false
	}
	return nil, false
}

func patternsRequireConditionLowering(branches []ir.PatternBranch) bool {
	for _, branch := range branches {
		if patternRequiresConditionLowering(branch.Pattern) {
			return true
		}
	}
	return false
}

func patternRequiresConditionLowering(pattern ir.Pattern) bool {
	switch p := pattern.(type) {
	case *ir.BindingPattern:
		return p.Constant
	case *ir.ComparePattern, *ir.RangePattern, *ir.MapPattern, *ir.ObjectPattern, *ir.BitPattern, *ir.SequenceSpreadPattern:
		return true
	case *ir.ArrayPattern:
		for _, elem := range p.Elements {
			if patternRequiresConditionLowering(elem) {
				return true
			}
		}
		return p.SubjectType == checker.String || p.SubjectType == checker.Bytes
	case *ir.AsPattern:
		return patternRequiresConditionLowering(p.Pattern)
	case *ir.OrPattern:
		for _, alt := range p.Alternatives {
			if patternRequiresConditionLowering(alt) {
				return true
			}
		}
	case *ir.TuplePattern:
		for _, elem := range p.Elements {
			if patternRequiresConditionLowering(elem) {
				return true
			}
		}
	case *ir.ConstructorPattern:
		if mbtJSONConstructorPattern(p.Name) || len(p.Args) > 0 || p.Rest {
			return true
		}
		for _, arg := range p.Args {
			if patternRequiresConditionLowering(arg) {
				return true
			}
		}
	}
	return false
}

func expandedBranchPatterns(pattern ir.Pattern) []ir.Pattern {
	if or, ok := pattern.(*ir.OrPattern); ok {
		return or.Alternatives
	}
	return []ir.Pattern{pattern}
}

func mbtJSONConstructorPattern(name string) bool {
	switch name {
	case "Array", "Object", "String", "Number", "Bool", "Null":
		return true
	default:
		return false
	}
}

func (g *generator) pattern(pattern ir.Pattern) string {
	return g.patternFor("_", pattern)
}

func (g *generator) patternFor(subject string, pattern ir.Pattern) string {
	switch p := pattern.(type) {
	case *ir.WildcardPattern:
		return "_"
	case *ir.BindingPattern:
		return mangleIdent(p.Name)
	case *ir.LiteralPattern:
		return g.expr(p.Value)
	case *ir.ComparePattern:
		return subjectGuard(subject, fmt.Sprintf("%s %s %s", subject, mbtBinaryOp(p.Op), g.expr(p.Value)))
	case *ir.RangePattern:
		return subjectGuard(subject, g.patternCondition(subject, p))
	case *ir.OrPattern:
		conditions := make([]string, 0, len(p.Alternatives))
		for _, alt := range p.Alternatives {
			conditions = append(conditions, g.patternCondition(subject, alt))
		}
		return subjectGuard(subject, strings.Join(conditions, " || "))
	case *ir.TuplePattern:
		parts := make([]string, 0, len(p.Elements))
		for _, elem := range p.Elements {
			parts = append(parts, g.patternFor(subject, elem))
		}
		return "(" + strings.Join(parts, ", ") + ")"
	case *ir.ArrayPattern:
		parts := make([]string, 0, len(p.Elements)+1)
		for idx, elem := range p.Elements {
			if p.RestIndex == idx {
				parts = append(parts, mbtRestPattern(p.RestBinding))
			}
			parts = append(parts, g.patternFor(subject, elem))
		}
		if p.RestIndex == len(p.Elements) {
			parts = append(parts, mbtRestPattern(p.RestBinding))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case *ir.AsPattern:
		return fmt.Sprintf("%s as %s", g.patternFor(subject, p.Pattern), mangleIdent(p.Name))
	case *ir.ConstructorPattern:
		if len(p.Args) == 0 && !p.Rest {
			return p.Name
		}
		parts := make([]string, 0, len(p.Args)+1)
		for _, arg := range p.Args {
			parts = append(parts, g.patternFor(subject, arg))
		}
		if p.Rest {
			parts = append(parts, "..")
		}
		return fmt.Sprintf("%s(%s)", p.Name, strings.Join(parts, ", "))
	default:
		g.addError(fmt.Errorf("MoonBit backend does not support complex pattern %T yet", pattern))
		return "_"
	}
}

func subjectGuard(subject string, condition string) string {
	if subject == "_" {
		return "_"
	}
	return "_ if " + condition
}

func mbtRestPattern(binding string) string {
	if binding == "" {
		return ".."
	}
	return ".." + mangleIdent(binding)
}

func (g *generator) patternCondition(subject string, pattern ir.Pattern) string {
	return g.patternConditionAs(subject, checker.Unknown, pattern)
}

func (g *generator) patternConditionAs(subject string, subjectType checker.Type, pattern ir.Pattern) string {
	switch p := pattern.(type) {
	case *ir.WildcardPattern:
		return "true"
	case *ir.BindingPattern:
		if p.Constant {
			if condition, ok := g.enumBindingPatternCondition(subject, p); ok {
				return condition
			}
			return fmt.Sprintf("%s == %s", subject, mangleIdent(p.Name))
		}
		return "true"
	case *ir.LiteralPattern:
		return fmt.Sprintf("%s == %s", subject, g.patternLiteralExpr(subjectType, p.Value))
	case *ir.ComparePattern:
		return fmt.Sprintf("%s %s %s", subject, mbtBinaryOp(p.Op), g.expr(p.Value))
	case *ir.RangePattern:
		parts := []string{}
		if p.Start != nil {
			parts = append(parts, fmt.Sprintf("%s >= %s", subject, g.expr(p.Start)))
		}
		if p.End != nil {
			op := "<"
			if p.Inclusive {
				op = "<="
			}
			parts = append(parts, fmt.Sprintf("%s %s %s", subject, op, g.expr(p.End)))
		}
		if len(parts) == 0 {
			return "true"
		}
		return strings.Join(parts, " && ")
	case *ir.OrPattern:
		conditions := make([]string, 0, len(p.Alternatives))
		for _, alt := range p.Alternatives {
			conditions = append(conditions, g.patternConditionAs(subject, subjectType, alt))
		}
		return "(" + strings.Join(conditions, " || ") + ")"
	case *ir.TuplePattern:
		parts := make([]string, 0, len(p.Elements))
		for idx, elem := range p.Elements {
			parts = append(parts, g.patternConditionAs(fmt.Sprintf("%s.%d", subject, idx), checker.Unknown, elem))
		}
		return joinConditions(parts)
	case *ir.ArrayPattern:
		return g.arrayPatternCondition(subject, p)
	case *ir.SequenceSpreadPattern:
		return g.sequenceSpreadPatternCondition(subject, subjectType, "0", p)
	case *ir.BitPattern:
		return g.patternConditionAs(g.bitPatternValueExpr(subject, subjectType, 0, p), checker.Int, p.Value)
	case *ir.AsPattern:
		return g.patternConditionAs(subject, subjectType, p.Pattern)
	case *ir.ConstructorPattern:
		if condition, ok := g.enumConstructorPatternCondition(subject, p); ok {
			return condition
		}
		if condition, ok := g.jsonConstructorPatternCondition(subject, p); ok {
			return condition
		}
		switch p.Name {
		case "Ok", "Err":
			parts := []string{fmt.Sprintf("%s is %s(_)", subject, p.Name)}
			for idx, arg := range p.Args {
				payload := g.constructorPayload(subject, p, idx)
				if payload != "" {
					parts = append(parts, g.patternConditionAs(payload, checker.Unknown, arg))
				}
			}
			return joinConditions(parts)
		default:
			return "false"
		}
	case *ir.MapPattern:
		return g.mapPatternCondition(subject, p)
	case *ir.ObjectPattern:
		return g.objectPatternCondition(subject, p)
	default:
		return "false"
	}
}

func (g *generator) patternLiteralExpr(subjectType checker.Type, value ir.Expr) string {
	if mbtJSONValueType(subjectType) {
		return g.jsonValueExpr(value)
	}
	return g.expr(value)
}

func (g *generator) enumBindingPatternCondition(subject string, pattern *ir.BindingPattern) (string, bool) {
	enum := g.enumFromType(pattern.Type)
	if enum == nil {
		return "", false
	}
	for _, member := range enum.Members {
		if member.Name == pattern.Name {
			return fmt.Sprintf("%s is %s", subject, mangleType(member.Name)), true
		}
	}
	return "", false
}

func (g *generator) enumConstructorPatternCondition(subject string, pattern *ir.ConstructorPattern) (string, bool) {
	_, member, ok := g.enumMemberForConstructor(pattern.SubjectType, pattern.Name)
	if !ok || len(member.Params) == 0 {
		return "", false
	}
	parts := []string{fmt.Sprintf("(match %s { %s => true; _ => false })", subject, constructorMatchPattern(member.Name, len(member.Params)))}
	for idx, arg := range pattern.Args {
		payload := g.constructorPayload(subject, pattern, idx)
		if payload == "" {
			continue
		}
		parts = append(parts, g.patternConditionAs(payload, member.Params[idx].Type, arg))
	}
	return joinConditions(parts), true
}

func (g *generator) enumMemberForConstructor(typ checker.Type, name string) (*ir.EnumType, *ir.EnumMember, bool) {
	enum := g.enumFromType(typ)
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

func (g *generator) constructorPayload(subject string, pattern *ir.ConstructorPattern, idx int) string {
	if _, member, ok := g.enumMemberForConstructor(pattern.SubjectType, pattern.Name); ok && len(member.Params) > 0 {
		names := make([]string, len(member.Params))
		for i := range names {
			names[i] = "_"
		}
		if idx < 0 || idx >= len(names) {
			return ""
		}
		value := g.nextTemp("__payload")
		names[idx] = value
		return fmt.Sprintf("(match %s { %s(%s) => %s; _ => abort(\"unreachable pattern payload\") })", subject, mangleType(member.Name), strings.Join(names, ", "), value)
	}
	if idx != 0 {
		return ""
	}
	switch pattern.Name {
	case "Ok":
		value := g.nextTemp("__ok")
		return fmt.Sprintf("(match %s { Ok(%s) => %s; _ => %s })", subject, value, value, zeroValue(checker.Unknown))
	case "Err":
		value := g.nextTemp("__err")
		return fmt.Sprintf("(match %s { Err(%s) => %s; _ => %s })", subject, value, value, zeroValue(checker.Unknown))
	case "Array":
		value := g.nextTemp("__json_array")
		return fmt.Sprintf("(match %s { Array(%s) => %s; _ => [] })", subject, value, value)
	case "Object":
		value := g.nextTemp("__json_object")
		return fmt.Sprintf("(match %s { Object(%s) => %s; _ => {} })", subject, value, value)
	case "String":
		value := g.nextTemp("__json_string")
		return fmt.Sprintf("(match %s { String(%s) => %s; _ => \"\" })", subject, value, value)
	case "Number":
		value := g.nextTemp("__json_number")
		return fmt.Sprintf("(match %s { Number(%s, ..) => %s; _ => 0.0 })", subject, value, value)
	case "Bool":
		return fmt.Sprintf("(match %s { True => true; False => false; _ => false })", subject)
	default:
		return ""
	}
}

func constructorMatchPattern(name string, arity int) string {
	if arity <= 0 {
		return mangleType(name)
	}
	return fmt.Sprintf("%s(%s)", mangleType(name), strings.TrimSuffix(strings.Repeat("_, ", arity), ", "))
}

func (g *generator) jsonConstructorPatternCondition(subject string, pattern *ir.ConstructorPattern) (string, bool) {
	var check string
	switch pattern.Name {
	case "Array":
		check = fmt.Sprintf("%s is Array(_)", subject)
	case "Object":
		check = fmt.Sprintf("%s is Object(_)", subject)
	case "String":
		check = fmt.Sprintf("%s is String(_)", subject)
	case "Number":
		check = fmt.Sprintf("%s is Number(_)", subject)
	case "Null":
		return fmt.Sprintf("%s is Null", subject), true
	case "Bool":
		check = fmt.Sprintf("(%s is True || %s is False)", subject, subject)
	default:
		return "", false
	}
	if len(pattern.Args) == 0 {
		return check, true
	}
	payload := g.constructorPayload(subject, pattern, 0)
	return fmt.Sprintf("(%s && %s)", check, g.patternConditionAs(payload, g.jsonConstructorPayloadType(pattern.Name), pattern.Args[0])), true
}

func (g *generator) jsonConstructorPayloadType(name string) checker.Type {
	switch name {
	case "Array":
		return checker.Type("Array[Object]")
	case "Object":
		return checker.Type("Map[String,Object]")
	case "String":
		return checker.String
	case "Number":
		return checker.Double
	case "Bool":
		return checker.Bool
	default:
		return checker.Unknown
	}
}

func (g *generator) arrayPatternCondition(subject string, pattern *ir.ArrayPattern) string {
	if pattern.SubjectType == checker.String {
		g.useString = true
	}
	if irArrayPatternHasBits(pattern) {
		return g.bitArrayPatternCondition(subject, pattern)
	}
	length := mbtSequenceLength(subject, pattern.SubjectType)
	required := g.arrayPatternRequiredWidth(pattern)
	parts := []string{}
	if pattern.RestIndex >= 0 {
		parts = append(parts, fmt.Sprintf("%s >= %s", length, required))
	} else {
		parts = append(parts, fmt.Sprintf("%s == %s", length, required))
	}
	for idx, elem := range pattern.Elements {
		if spread, ok := elem.(*ir.SequenceSpreadPattern); ok {
			if pattern.RestIndex >= 0 {
				parts = append(parts, g.sequenceSpreadSearchPatternCondition(subject, pattern.SubjectType, spread))
				continue
			}
			parts = append(parts, g.sequenceSpreadPatternCondition(subject, pattern.SubjectType, g.arrayPatternElementIndex(subject, pattern, idx), spread))
			continue
		}
		parts = append(parts, g.patternConditionAs(mbtSequenceIndex(subject, pattern.SubjectType, g.arrayPatternElementIndex(subject, pattern, idx)), g.sequenceElementType(pattern.SubjectType), elem))
	}
	return joinConditions(parts)
}

func (g *generator) sequenceSpreadPatternCondition(subject string, subjectType checker.Type, start string, pattern *ir.SequenceSpreadPattern) string {
	if subjectType == checker.String || pattern.Type == checker.String {
		g.useString = true
	}
	spread := g.nextTemp("__spread")
	idx := g.nextTemp("__idx")
	return fmt.Sprintf("{ let %s = %s; let mut %s = 0; let mut __ok = true; while %s < %s { if %s != %s { __ok = false }; %s = %s + 1 }; __ok }",
		spread,
		g.expr(pattern.Value),
		idx,
		idx,
		mbtSequenceLength(spread, pattern.Type),
		mbtSequenceIndex(subject, subjectType, fmt.Sprintf("(%s + %s)", start, idx)),
		mbtSequenceIndex(spread, pattern.Type, idx),
		idx,
		idx,
	)
}

func (g *generator) sequenceSpreadSearchPatternCondition(subject string, subjectType checker.Type, pattern *ir.SequenceSpreadPattern) string {
	if subjectType == checker.String || pattern.Type == checker.String {
		g.useString = true
	}
	spread := g.nextTemp("__spread")
	offset := g.nextTemp("__offset")
	idx := g.nextTemp("__idx")
	subjectLen := mbtSequenceLength(subject, subjectType)
	spreadLen := mbtSequenceLength(spread, pattern.Type)
	return fmt.Sprintf("{ let %s = %s; let mut %s = 0; let mut __found = false; while %s <= %s - %s { let mut %s = 0; let mut __ok = true; while %s < %s { if %s != %s { __ok = false }; %s = %s + 1 }; if __ok { __found = true }; %s = %s + 1 }; __found }",
		spread,
		g.expr(pattern.Value),
		offset,
		offset,
		subjectLen,
		spreadLen,
		idx,
		idx,
		spreadLen,
		mbtSequenceIndex(subject, subjectType, fmt.Sprintf("(%s + %s)", offset, idx)),
		mbtSequenceIndex(spread, pattern.Type, idx),
		idx,
		idx,
		offset,
		offset,
	)
}

func irArrayPatternHasBits(pattern *ir.ArrayPattern) bool {
	for _, elem := range pattern.Elements {
		if _, ok := elem.(*ir.BitPattern); ok {
			return true
		}
	}
	return false
}

func (g *generator) bitArrayPatternCondition(subject string, pattern *ir.ArrayPattern) string {
	lengthBits := mbtSequenceLength(subject, pattern.SubjectType) + " * 8"
	requiredBits := bitPatternRequiredBits(pattern)
	parts := []string{}
	if pattern.RestIndex >= 0 {
		parts = append(parts, fmt.Sprintf("%s >= %d", lengthBits, requiredBits))
	} else {
		parts = append(parts, fmt.Sprintf("%s == %d", lengthBits, requiredBits))
	}
	for idx, elem := range pattern.Elements {
		bit, ok := elem.(*ir.BitPattern)
		if !ok {
			continue
		}
		offset := bitPatternOffset(pattern, idx)
		parts = append(parts, g.patternConditionAs(g.bitPatternValueExpr(subject, pattern.SubjectType, offset, bit), checker.Int, bit.Value))
	}
	return joinConditions(parts)
}

func bitPatternRequiredBits(pattern *ir.ArrayPattern) int {
	total := 0
	for idx, elem := range pattern.Elements {
		if pattern.RestIndex >= 0 && idx >= pattern.RestIndex {
			break
		}
		if bit, ok := elem.(*ir.BitPattern); ok {
			total += bit.Width
		}
	}
	return total
}

func bitPatternOffset(pattern *ir.ArrayPattern, idx int) int {
	if pattern.RestIndex < 0 || idx < pattern.RestIndex {
		offset := 0
		for i := 0; i < idx; i++ {
			if bit, ok := pattern.Elements[i].(*ir.BitPattern); ok {
				offset += bit.Width
			}
		}
		return offset
	}
	tail := 0
	for i := idx; i < len(pattern.Elements); i++ {
		if bit, ok := pattern.Elements[i].(*ir.BitPattern); ok {
			tail += bit.Width
		}
	}
	return -tail
}

func (g *generator) bitPatternValueExpr(subject string, typ checker.Type, offset int, pattern *ir.BitPattern) string {
	start := fmt.Sprintf("%d", offset)
	if offset < 0 {
		start = fmt.Sprintf("%s * 8 - %d", mbtSequenceLength(subject, typ), -offset)
	}
	if pattern.Endian == "le" && pattern.Width%8 == 0 {
		byteIndex := g.nextTemp("__byte_index")
		value := "__out"
		if pattern.Signed {
			value = fmt.Sprintf("{ let mut __value = __out; if %d < 32 && (__value & (1 << %d)) != 0 { __value = __value - (1 << %d) }; __value }", pattern.Width, pattern.Width-1, pattern.Width)
		}
		return fmt.Sprintf("{ let __start = (%s) / 8; let mut __out = 0; let mut __i = 0; while __i < %d { let %s = __start + __i; __out = __out | (%s << (8 * __i)); __i = __i + 1 }; %s }", start, pattern.Width/8, byteIndex, g.bitPatternByteExpr(subject, typ, byteIndex), value)
	}
	value := "__out"
	if pattern.Signed {
		value = fmt.Sprintf("{ let mut __value = __out; if %d < 32 && (__value & (1 << %d)) != 0 { __value = __value - (1 << %d) }; __value }", pattern.Width, pattern.Width-1, pattern.Width)
	}
	return fmt.Sprintf("{ let mut __out = 0; let mut __i = 0; while __i < %d { let __bit_index = %s + __i; let __bit = (%s >> (7 - (__bit_index %% 8))) & 1; __out = (__out << 1) | __bit; __i = __i + 1 }; %s }", pattern.Width, start, g.bitPatternByteExpr(subject, typ, "__bit_index / 8"), value)
}

func (g *generator) bitPatternByteExpr(subject string, typ checker.Type, index string) string {
	return mbtSequenceIndex(subject, typ, index)
}

func (g *generator) mapPatternCondition(subject string, pattern *ir.MapPattern) string {
	if pattern.Access == "object" || pattern.SubjectType == checker.Object {
		return g.objectMapPatternCondition(subject, pattern)
	}
	parts := make([]string, 0, len(pattern.Entries))
	for _, entry := range pattern.Entries {
		key := g.nextTemp("__key")
		value := g.nextTemp("__pattern")
		keyExpr := g.expr(entry.Key)
		valueExpr := fmt.Sprintf("%s.get(%s)", subject, key)
		condition := g.patternConditionAs(value, g.optionalPatternType(pattern.ValueType, entry.Optional), entry.Pattern)
		if entry.Optional {
			parts = append(parts, fmt.Sprintf("{ let %s = %s; let %s = %s; %s }", key, keyExpr, value, valueExpr, condition))
		} else if pattern.Access == "get" {
			parts = append(parts, fmt.Sprintf("{ let %s = %s; let %s = %s; %s != None && { let %s = %s.unwrap(); %s } }", key, keyExpr, value, valueExpr, value, value, value, g.patternConditionAs(value, pattern.ValueType, entry.Pattern)))
		} else {
			parts = append(parts, fmt.Sprintf("{ let %s = %s; %s.contains(%s) && { let %s = %s.unwrap(); %s } }", key, keyExpr, subject, key, value, valueExpr, g.patternConditionAs(value, pattern.ValueType, entry.Pattern)))
		}
	}
	return joinConditions(parts)
}

func (g *generator) optionalPatternType(typ checker.Type, optional bool) checker.Type {
	if optional {
		return checker.Type(string(typ) + "?")
	}
	return typ
}

func (g *generator) objectMapPatternCondition(subject string, pattern *ir.MapPattern) string {
	obj := g.nextTemp("__json_object")
	parts := []string{fmt.Sprintf("%s is Object(_)", subject)}
	inner := make([]string, 0, len(pattern.Entries))
	for _, entry := range pattern.Entries {
		key, ok := entry.Key.(*ir.StringLiteral)
		if !ok {
			inner = append(inner, "false")
			continue
		}
		value := fmt.Sprintf("%s.get(%s).unwrap_or(Json::null())", obj, quoteString(key.Value))
		condition := g.patternConditionAs(value, checker.Object, entry.Pattern)
		if entry.Optional {
			inner = append(inner, condition)
		} else {
			inner = append(inner, fmt.Sprintf("%s.contains(%s) && %s", obj, quoteString(key.Value), condition))
		}
	}
	if len(inner) == 0 {
		return parts[0]
	}
	return fmt.Sprintf("(%s && (match %s { Object(%s) => %s; _ => false }))", parts[0], subject, obj, joinConditions(inner))
}

func (g *generator) objectPatternCondition(subject string, pattern *ir.ObjectPattern) string {
	parts := make([]string, 0, len(pattern.Fields))
	for _, field := range pattern.Fields {
		if field.Optional && !field.Exists {
			continue
		}
		parts = append(parts, g.patternConditionAs(subject+"."+mangleIdent(field.Name), field.Type, field.Pattern))
	}
	return joinConditions(parts)
}

func (g *generator) patternBinding(subject string, subjectType checker.Type, pattern ir.Pattern) string {
	var parts []string
	g.appendPatternBindings(&parts, subject, subjectType, pattern)
	return strings.Join(parts, " ")
}

func (g *generator) appendPatternBindings(parts *[]string, subject string, subjectType checker.Type, pattern ir.Pattern) {
	switch p := pattern.(type) {
	case *ir.BindingPattern:
		if !p.Constant {
			*parts = append(*parts, fmt.Sprintf("let %s = %s;", mangleIdent(p.Name), subject))
		}
	case *ir.TuplePattern:
		for idx, elem := range p.Elements {
			g.appendPatternBindings(parts, fmt.Sprintf("%s.%d", subject, idx), checker.Unknown, elem)
		}
	case *ir.ArrayPattern:
		if p.SubjectType == checker.String {
			g.useString = true
		}
		for idx, elem := range p.Elements {
			if bit, ok := elem.(*ir.BitPattern); ok {
				g.appendPatternBindings(parts, g.bitPatternValueExpr(subject, p.SubjectType, bitPatternOffset(p, idx), bit), checker.Int, bit.Value)
				continue
			}
			if _, ok := elem.(*ir.SequenceSpreadPattern); ok {
				continue
			}
			g.appendPatternBindings(parts, mbtSequenceIndex(subject, p.SubjectType, g.arrayPatternElementIndex(subject, p, idx)), g.sequenceElementType(p.SubjectType), elem)
		}
		if p.RestBinding != "" {
			if irArrayPatternHasBits(p) {
				start := bitPatternRequiredBits(p) / 8
				*parts = append(*parts, fmt.Sprintf("let %s = %s;", mangleIdent(p.RestBinding), mbtSequenceSliceExpr(subject, p.SubjectType, fmt.Sprintf("%d", start), mbtSequenceLength(subject, p.SubjectType))))
			} else {
				start := g.arrayPatternPrefixWidth(p, p.RestIndex)
				end := mbtSubtractExpr(mbtSequenceLength(subject, p.SubjectType), g.arrayPatternSuffixWidth(p, p.RestIndex))
				*parts = append(*parts, fmt.Sprintf("let %s = %s;", mangleIdent(p.RestBinding), mbtSequenceSliceExpr(subject, p.SubjectType, start, end)))
			}
		}
	case *ir.BitPattern:
		g.appendPatternBindings(parts, subject, subjectType, p.Value)
	case *ir.AsPattern:
		g.appendPatternBindings(parts, subject, subjectType, p.Pattern)
		*parts = append(*parts, fmt.Sprintf("let %s = %s;", mangleIdent(p.Name), subject))
	case *ir.OrPattern:
		if len(p.Alternatives) > 0 {
			g.appendPatternBindings(parts, subject, subjectType, p.Alternatives[0])
		}
	case *ir.ConstructorPattern:
		if _, member, ok := g.enumMemberForConstructor(p.SubjectType, p.Name); ok {
			for idx, arg := range p.Args {
				payload := g.constructorPayload(subject, p, idx)
				if payload != "" && idx < len(member.Params) {
					g.appendPatternBindings(parts, payload, member.Params[idx].Type, arg)
				}
			}
			return
		}
		for idx, arg := range p.Args {
			payload := g.constructorPayload(subject, p, idx)
			if payload != "" {
				g.appendPatternBindings(parts, payload, g.jsonConstructorPayloadType(p.Name), arg)
			}
		}
	case *ir.MapPattern:
		if p.Access == "object" || p.SubjectType == checker.Object {
			for _, entry := range p.Entries {
				key, ok := entry.Key.(*ir.StringLiteral)
				if !ok {
					continue
				}
				value := fmt.Sprintf("(match %s { Object(__obj) => __obj.get(%s).unwrap_or(Json::null()); _ => Json::null() })", subject, quoteString(key.Value))
				g.appendPatternBindings(parts, value, checker.Object, entry.Pattern)
			}
			return
		}
		for _, entry := range p.Entries {
			value := fmt.Sprintf("%s.get(%s)", subject, g.expr(entry.Key))
			if !entry.Optional {
				value += ".unwrap()"
			}
			g.appendPatternBindings(parts, value, g.optionalPatternType(p.ValueType, entry.Optional), entry.Pattern)
		}
	case *ir.ObjectPattern:
		for _, field := range p.Fields {
			if field.Optional && !field.Exists {
				continue
			}
			g.appendPatternBindings(parts, subject+"."+mangleIdent(field.Name), field.Type, field.Pattern)
		}
	}
}

func (g *generator) sequenceElementType(typ checker.Type) checker.Type {
	if typ == checker.String {
		return checker.Char
	}
	if typ == checker.Bytes {
		return checker.Int
	}
	if elem, ok := checker.ArrayElement(typ); ok {
		return elem
	}
	return checker.Unknown
}

func (g *generator) arrayPatternElementIndex(subject string, pattern *ir.ArrayPattern, idx int) string {
	if pattern.RestIndex < 0 || idx < pattern.RestIndex {
		return g.arrayPatternPrefixWidth(pattern, idx)
	}
	return mbtSubtractExpr(mbtSequenceLength(subject, pattern.SubjectType), g.arrayPatternSuffixWidth(pattern, idx))
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
			terms = append(terms, mbtSequenceLength(g.expr(spread.Value), spread.Type))
			continue
		}
		constants++
	}
	if constants > 0 || len(terms) == 0 {
		terms = append([]string{fmt.Sprintf("%d", constants)}, terms...)
	}
	return strings.Join(terms, " + ")
}

func mbtSequenceLength(subject string, typ checker.Type) string {
	switch typ {
	case checker.String:
		return fmt.Sprintf("rune_string_length(%s)", subject)
	default:
		return subject + ".length()"
	}
}

func mbtSubtractExpr(left string, right string) string {
	if right == "0" {
		return left
	}
	if strings.ContainsAny(right, "+-*/ ") {
		return fmt.Sprintf("%s - (%s)", left, right)
	}
	return fmt.Sprintf("%s - %s", left, right)
}

func mbtSequenceIndex(subject string, typ checker.Type, index string) string {
	index = strings.ReplaceAll(index, mbtSequenceLength("_", typ), mbtSequenceLength(subject, typ))
	switch typ {
	case checker.String:
		return fmt.Sprintf("rune_string_at(%s, %s)", subject, index)
	default:
		return fmt.Sprintf("%s[%s]", subject, index)
	}
}

func mbtSequenceSliceExpr(subject string, typ checker.Type, start string, end string) string {
	switch typ {
	case checker.String:
		return fmt.Sprintf("rune_string_slice(%s, %s, %s)", subject, start, end)
	default:
		return fmt.Sprintf("%s[%s:%s].to_owned()", subject, start, end)
	}
}

func joinConditions(parts []string) string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" || part == "true" {
			continue
		}
		out = append(out, part)
	}
	if len(out) == 0 {
		return "true"
	}
	return strings.Join(out, " && ")
}

func splitInlineStatements(src string) []string {
	chunks := strings.Split(src, ";")
	out := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		if trimmed := strings.TrimSpace(chunk); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
