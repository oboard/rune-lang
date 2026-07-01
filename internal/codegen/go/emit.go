package gocodegen

import (
	"fmt"
	"strings"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
)

func (g *generator) structType(typ *ir.StructType) {
	g.linef("type %s struct {", mangleIdent(typ.Name))
	g.indent++
	for _, field := range typ.Fields {
		g.linef("%s %s", mangleIdent(field.Name), goType(field.Type))
	}
	g.indent--
	g.line("}")
}

func (g *generator) enumType(enum *ir.EnumType) {
	if enumHasPayload(enum) {
		g.linef("type %s struct {", mangleIdent(enum.Name))
		g.indent++
		g.line("__tag int")
		g.line("__payload []any")
		g.indent--
		g.line("}")
		if len(enum.Members) == 0 {
			return
		}
		g.line("")
		g.line("const (")
		g.indent++
		for i, member := range enum.Members {
			g.linef("%s = %d", mangleEnumMember(enum.Name, member.Name), i)
		}
		g.indent--
		g.line(")")
		return
	}
	g.linef("type %s int", mangleIdent(enum.Name))
	if len(enum.Members) == 0 {
		return
	}
	g.line("")
	g.line("const (")
	g.indent++
	for i, member := range enum.Members {
		value := i
		if member.HasValue {
			value = member.Value
		}
		g.linef("%s %s = %d", mangleEnumMember(enum.Name, member.Name), mangleIdent(enum.Name), value)
	}
	g.indent--
	g.line(")")
}

func (g *generator) constDecl(constant *ir.ConstDecl) {
	g.linef("var %s %s = %s", mangleIdent(constant.Name), goType(constant.Type), g.expr(constant.Value))
}

func (g *generator) function(fn *ir.Function) error {
	params := make([]string, 0, len(fn.Params))
	for _, param := range fn.Params {
		params = append(params, fmt.Sprintf("%s %s", mangleIdent(param.Name), goType(param.Type)))
	}
	if fn.Routine {
		return g.routineFunction(fn, params)
	}
	ret := ""
	if fn.Return != checker.Void {
		ret = " " + goType(fn.Return)
	}
	g.linef("func %s%s(%s)%s {", mangleIdent(fn.Name), goGenerics(fn.Generics, fn.GenericConstraints), strings.Join(params, ", "), ret)
	g.indent++
	g.pushSignalScope()
	if err := g.body(fn, fn.Body, fn.Return); err != nil {
		return err
	}
	g.popSignalScope()
	g.indent--
	g.line("}")
	return nil
}

func (g *generator) routineFunction(fn *ir.Function, params []string) error {
	ret := goType(fn.Return)
	if fn.Return == checker.Void {
		ret = "runeUnit"
	}
	g.linef("func %s(%s) runeTask[%s] {", mangleIdent(fn.Name), strings.Join(params, ", "), ret)
	g.indent++
	g.linef("return runeGo(func() %s {", ret)
	g.indent++
	g.pushSignalScope()
	if err := g.body(fn, fn.Body, fn.Return); err != nil {
		return err
	}
	if fn.Return == checker.Void {
		g.line("return runeUnit{}")
	}
	g.popSignalScope()
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
	return nil
}

func (g *generator) method(typeName string, fn *ir.Function) error {
	params := make([]string, 0, len(fn.Params))
	for _, param := range fn.Params {
		params = append(params, fmt.Sprintf("%s %s", mangleIdent(param.Name), goType(param.Type)))
	}
	ret := ""
	if fn.Return != checker.Void {
		ret = " " + goType(fn.Return)
	}
	if fn.Routine {
		if fn.Return == checker.Void {
			ret = " runeTask[runeUnit]"
		} else {
			ret = " runeTask[" + goType(fn.Return) + "]"
		}
	}
	if fn.Static {
		g.linef("func %s%s(%s)%s {", mangleIdent(typeName+"_"+fn.Name), goGenerics(fn.Generics, fn.GenericConstraints), strings.Join(params, ", "), ret)
	} else {
		g.linef("func (%s %s) %s%s(%s)%s {", mangleIdent("this"), mangleIdent(typeName), mangleIdent(fn.Name), goGenerics(fn.Generics, fn.GenericConstraints), strings.Join(params, ", "), ret)
	}
	g.indent++
	g.pushSignalScope()
	if fn.Routine {
		routineRet := goType(fn.Return)
		if fn.Return == checker.Void {
			routineRet = "runeUnit"
		}
		g.linef("return runeGo(func() %s {", routineRet)
		g.indent++
		if err := g.body(fn, fn.Body, fn.Return); err != nil {
			return err
		}
		if fn.Return == checker.Void {
			g.line("return runeUnit{}")
		}
		g.indent--
		g.line("})")
	} else {
		if err := g.body(fn, fn.Body, fn.Return); err != nil {
			return err
		}
	}
	g.popSignalScope()
	g.indent--
	g.line("}")
	return nil
}

func (g *generator) body(fn *ir.Function, expr ir.Expr, ret checker.Type) error {
	switch e := expr.(type) {
	case *ir.PatternBlock:
		return g.patternBlock(fn, e, ret)
	case *ir.BlockExpr:
		return g.block(e, ret)
	default:
		if unwrap, ok := expr.(*ir.ResultUnwrapExpr); ok {
			g.resultUnwrapExprStmt(unwrap, ret, true)
			return nil
		}
		if ret == checker.Void {
			if expr := g.expr(expr); expr != "" {
				g.line(expr)
			}
		} else {
			g.linef("return %s", g.returnExpr(expr, ret))
		}
	}
	return nil
}

func (g *generator) block(block *ir.BlockExpr, ret checker.Type) error {
	for i, stmt := range block.Statements {
		last := i == len(block.Statements)-1
		switch s := stmt.(type) {
		case *ir.LetStmt:
			if unwrap, ok := s.Value.(*ir.ResultUnwrapExpr); ok {
				g.resultUnwrapLet(s.Name, unwrap, ret)
				continue
			}
			if isNamespaceValue(s.Value) {
				continue
			}
			if s.Signal || g.exprUsesSignal(s.Value) {
				g.linef("%s := newRuneSignal(%s)", mangleIdent(s.Name), g.expr(s.Value))
				g.addSignal(s.Name, s.Value.ResultType())
				deps := g.exprSignalDeps(s.Value)
				g.setSignalDeps(s.Name, deps)
				for _, dep := range deps {
					depName := mangleIdent(dep)
					name := mangleIdent(s.Name)
					g.linef("%s.Watch(func(_, _ %s) { %s.Set(%s) })", depName, goType(g.signalType(dep)), name, g.expr(s.Value))
				}
				continue
			}
			if obj, ok := s.Value.(*ir.AnonymousObjectLiteral); ok {
				g.linef("var %s %s", mangleIdent(s.Name), anonymousObjectType(obj))
				value := g.withThisName(mangleIdent(s.Name), func() string {
					return g.expr(s.Value)
				})
				g.linef("%s = %s", mangleIdent(s.Name), value)
				continue
			}
			g.linef("%s := %s", mangleIdent(s.Name), g.expr(s.Value))
		case *ir.ObjectDestructureStmt:
			if unwrap, ok := s.Value.(*ir.ResultUnwrapExpr); ok {
				g.resultUnwrapObjectDestructure(s, unwrap, ret)
				continue
			}
			g.objectDestructure(s, g.expr(s.Value))
		case *ir.AssignStmt:
			if g.isSignal(s.Name) {
				g.linef("%s.Set(%s)", mangleIdent(s.Name), g.expr(s.Value))
				continue
			}
			g.linef("%s = %s", mangleIdent(s.Name), g.expr(s.Value))
		case *ir.ExprStmt:
			if unwrap, ok := s.Expr.(*ir.ResultUnwrapExpr); ok {
				g.resultUnwrapExprStmt(unwrap, ret, last)
				continue
			}
			if block, ok := s.Expr.(*ir.BlockExpr); ok && !(last && ret != checker.Void) && g.exprUsesSignal(block) {
				g.effectScope(block)
				continue
			}
			if ternary, ok := s.Expr.(*ir.TernaryExpr); ok && ternary.Alternative == nil && !(last && ret != checker.Void) {
				g.conditionalExprStmt(ternary)
				continue
			}
			if stmt, ok := g.arrayEachStmt(s.Expr); ok {
				g.line(stmt)
				continue
			}
			if stmt, ok := g.arrayPushStmt(s.Expr); ok {
				g.line(stmt)
				continue
			}
			expr := g.expr(s.Expr)
			if expr == "" {
				continue
			}
			if last && ret != checker.Void {
				g.linef("return %s", g.returnExpr(s.Expr, ret))
			} else {
				g.line(expr)
			}
		}
	}
	if ret != checker.Void && len(block.Statements) == 0 {
		g.linef("return %s", g.zeroValue(ret))
	}
	return nil
}

func (g *generator) effectScope(block *ir.BlockExpr) {
	name := g.nextTemp("effect")
	g.linef("%s := func() {", name)
	g.indent++
	_ = g.block(block, checker.Void)
	g.indent--
	g.line("}")
	pending := g.nextTemp("effectPending")
	schedule := g.nextTemp("scheduleEffect")
	g.linef("%s := false", pending)
	g.linef("%s := func() {", schedule)
	g.indent++
	g.linef("if %s {", pending)
	g.indent++
	g.line("return")
	g.indent--
	g.line("}")
	g.linef("%s = true", pending)
	g.line("runeScheduleEffect(func() {")
	g.indent++
	g.linef("%s = false", pending)
	g.linef("%s()", name)
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
	g.linef("%s()", name)
	for _, dep := range g.effectSignalDeps(block) {
		depName := mangleIdent(dep)
		g.linef("%s.Watch(func(_, _ %s) { %s() })", depName, goType(g.signalType(dep)), schedule)
	}
}

func (g *generator) conditionalExprStmt(expr *ir.TernaryExpr) {
	g.linef("if %s {", g.expr(expr.Condition))
	g.indent++
	if consequence := g.expr(expr.Consequence); consequence != "" {
		g.line(consequence)
	}
	g.indent--
	g.line("}")
}

func (g *generator) resultUnwrapLet(name string, unwrap *ir.ResultUnwrapExpr, ret checker.Type) {
	tmp := g.nextTemp("result")
	g.linef("%s := %s", tmp, g.expr(unwrap.Expr))
	g.linef("if !%s.ok {", tmp)
	g.indent++
	g.linef("return %s", g.resultErrReturn(ret, tmp+".err"))
	g.indent--
	g.line("}")
	g.linef("%s := %s.value", mangleIdent(name), tmp)
}

func (g *generator) resultUnwrapObjectDestructure(stmt *ir.ObjectDestructureStmt, unwrap *ir.ResultUnwrapExpr, ret checker.Type) {
	tmp := g.nextTemp("result")
	g.linef("%s := %s", tmp, g.expr(unwrap.Expr))
	g.linef("if !%s.ok {", tmp)
	g.indent++
	g.linef("return %s", g.resultErrReturn(ret, tmp+".err"))
	g.indent--
	g.line("}")
	g.objectDestructure(stmt, tmp+".value")
}

func (g *generator) objectDestructure(stmt *ir.ObjectDestructureStmt, source string) {
	tmp := g.nextTemp("destructure")
	g.linef("%s := %s", tmp, source)
	signal := stmt.Signal || g.exprUsesSignal(stmt.Value)
	for _, field := range stmt.Fields {
		name := mangleIdent(field.Name)
		value := goObjectFieldAccess(tmp, field.Field)
		if signal {
			g.linef("%s := newRuneSignal(%s)", name, value)
			g.addSignal(field.Name, field.Type)
			g.setSignalDeps(field.Name, g.exprSignalDeps(stmt.Value))
			continue
		}
		g.linef("%s := %s", name, value)
	}
	if !signal {
		return
	}
	for _, dep := range g.exprSignalDeps(stmt.Value) {
		depName := mangleIdent(dep)
		for _, field := range stmt.Fields {
			name := mangleIdent(field.Name)
			g.linef("%s.Watch(func(_, _ %s) { %s.Set(%s) })", depName, goType(g.signalType(dep)), name, goObjectFieldAccess(g.expr(stmt.Value), field.Field))
		}
	}
}

func isNamespaceValue(expr ir.Expr) bool {
	if _, ok := expr.(*ir.AtExpr); ok {
		return true
	}
	if _, ok := checker.ModuleNamespaceName(expr.ResultType()); ok {
		return true
	}
	if _, ok := checker.ImportNamespacePath(expr.ResultType()); ok {
		return true
	}
	return false
}

func goObjectFieldAccess(source string, field string) string {
	return source + "." + mangleIdent(field)
}

func (g *generator) resultUnwrapExprStmt(unwrap *ir.ResultUnwrapExpr, ret checker.Type, last bool) {
	tmp := g.nextTemp("result")
	g.linef("%s := %s", tmp, g.expr(unwrap.Expr))
	g.linef("if !%s.ok {", tmp)
	g.indent++
	g.linef("return %s", g.resultErrReturn(ret, tmp+".err"))
	g.indent--
	g.line("}")
	if last && ret != checker.Void {
		g.linef("return %s", g.returnRawExpr(unwrap, ret, tmp+".value"))
	}
}

func (g *generator) resultErrReturn(ret checker.Type, errExpr string) string {
	okType, errType := resultTypeArgs(ret)
	if okType == checker.Unknown {
		return g.zeroValue(ret)
	}
	return fmt.Sprintf("runeErr[%s, %s](%s)", goType(okType), goType(errType), errExpr)
}

func (g *generator) returnExpr(expr ir.Expr, ret checker.Type) string {
	return g.returnRawExpr(expr, ret, g.expr(expr))
}

func (g *generator) returnRawExpr(expr ir.Expr, ret checker.Type, raw string) string {
	okType, errType := resultTypeArgs(ret)
	if okType == checker.Unknown {
		return raw
	}
	if expr != nil && expr.ResultType() == okType {
		return fmt.Sprintf("runeOk[%s, %s](%s)", goType(okType), goType(errType), raw)
	}
	return raw
}

func (g *generator) patternBlock(fn *ir.Function, block *ir.PatternBlock, ret checker.Type) error {
	if len(fn.Params) != 1 {
		return fmt.Errorf("%s: pattern blocks currently require exactly one parameter", block.Pos)
	}
	subject := mangleIdent(fn.Params[0].Name)
	restoreMapGetters := g.pushMapPatternGetters(subject, block.Branches)
	defer restoreMapGetters()
	for _, line := range g.mapPatternGetterPrelude(subject, block.Branches) {
		g.line(line)
	}
	g.line("switch {")
	g.indent++
	hasDefault := false
	for _, branch := range block.Branches {
		if _, ok := branch.Pattern.(*ir.WildcardPattern); ok {
			hasDefault = true
			g.line("default:")
		} else {
			g.linef("case %s:", g.patternCondition(subject, branch.Pattern))
		}
		g.indent++
		if binding := g.patternBinding(subject, branch.Pattern); binding != "" {
			g.line(binding)
		}
		if ret == checker.Void {
			if expr := g.expr(branch.Expr); expr != "" {
				g.line(expr)
			}
		} else {
			g.linef("return %s", g.returnExpr(branch.Expr, ret))
		}
		g.indent--
	}
	g.indent--
	g.line("}")
	if ret != checker.Void && !hasDefault {
		g.linef("return %s", g.zeroValue(ret))
	}
	return nil
}

func (g *generator) patternCondition(subject string, pattern ir.Pattern) string {
	switch p := pattern.(type) {
	case *ir.BindingPattern:
		if p.Constant {
			if condition, ok := g.enumBindingPatternCondition(subject, p); ok {
				return condition
			}
			return fmt.Sprintf("%s == %s", subject, mangleIdent(p.Name))
		}
		return "true"
	case *ir.LiteralPattern:
		return fmt.Sprintf("%s == %s", subject, g.expr(p.Value))
	case *ir.ComparePattern:
		return fmt.Sprintf("%s %s %s", subject, p.Op, g.expr(p.Value))
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
		return "(" + strings.Join(parts, " && ") + ")"
	case *ir.OrPattern:
		parts := make([]string, 0, len(p.Alternatives))
		for _, alternative := range p.Alternatives {
			parts = append(parts, "("+g.patternCondition(subject, alternative)+")")
		}
		return strings.Join(parts, " || ")
	case *ir.TuplePattern:
		parts := make([]string, 0, len(p.Elements))
		for i, elem := range p.Elements {
			parts = append(parts, g.patternCondition(fmt.Sprintf("%s.F%d", subject, i), elem))
		}
		return strings.Join(parts, " && ")
	case *ir.ArrayPattern:
		return g.arrayPatternCondition(subject, p)
	case *ir.BitPattern:
		return g.patternCondition(g.bitPatternValueExpr(subject, checker.Unknown, 0, p), p.Value)
	case *ir.AsPattern:
		return g.patternCondition(subject, p.Pattern)
	case *ir.ConstructorPattern:
		if condition, ok := g.enumConstructorPatternCondition(subject, p); ok {
			return condition
		}
		if condition, ok := g.jsonConstructorPatternCondition(subject, p); ok {
			return condition
		}
		parts := []string{}
		switch p.Name {
		case "Ok":
			parts = append(parts, subject+".ok")
		case "Err":
			parts = append(parts, "!"+subject+".ok")
		default:
			return "false"
		}
		for idx, arg := range p.Args {
			payload := g.constructorPayload(subject, p, idx)
			if payload == "" {
				continue
			}
			parts = append(parts, g.patternCondition(payload, arg))
		}
		return strings.Join(parts, " && ")
	case *ir.MapPattern:
		return g.mapPatternCondition(subject, p)
	case *ir.ObjectPattern:
		return g.objectPatternCondition(subject, p)
	default:
		return "true"
	}
}

func (g *generator) enumBindingPatternCondition(subject string, pattern *ir.BindingPattern) (string, bool) {
	enum := g.enumForType(pattern.Type)
	if enum == nil {
		return "", false
	}
	for _, member := range enum.Members {
		if member.Name != pattern.Name {
			continue
		}
		if enumHasPayload(enum) {
			return fmt.Sprintf("%s.__tag == %s", subject, mangleEnumMember(enum.Name, member.Name)), true
		}
		return fmt.Sprintf("%s == %s", subject, mangleEnumMember(enum.Name, member.Name)), true
	}
	return "", false
}

func (g *generator) enumConstructorPatternCondition(subject string, pattern *ir.ConstructorPattern) (string, bool) {
	enum, member, ok := g.enumMemberForConstructor(pattern.SubjectType, pattern.Name)
	if !ok || !enumHasPayload(enum) {
		return "", false
	}
	parts := []string{fmt.Sprintf("%s.__tag == %s", subject, mangleEnumMember(enum.Name, member.Name))}
	if pattern.Rest {
		parts = append(parts, fmt.Sprintf("len(%s.__payload) >= %d", subject, len(pattern.Args)))
	} else {
		parts = append(parts, fmt.Sprintf("len(%s.__payload) == %d", subject, len(pattern.Args)))
	}
	for idx, arg := range pattern.Args {
		parts = append(parts, g.patternCondition(g.constructorPayload(subject, pattern, idx), arg))
	}
	return strings.Join(parts, " && "), true
}

func (g *generator) constructorPayload(subject string, pattern *ir.ConstructorPattern, idx int) string {
	if enum, member, ok := g.enumMemberForConstructor(pattern.SubjectType, pattern.Name); ok && enumHasPayload(enum) {
		if idx < 0 || idx >= len(member.Params) {
			return ""
		}
		return fmt.Sprintf("%s.__payload[%d].(%s)", subject, idx, goType(member.Params[idx].Type))
	}
	if idx != 0 {
		return ""
	}
	switch pattern.Name {
	case "Ok":
		return subject + ".value"
	case "Err":
		return subject + ".err"
	case "Array":
		return subject + ".([]any)"
	case "Object":
		return subject + ".(map[string]any)"
	case "String":
		return subject + ".(string)"
	case "Bool":
		return subject + ".(bool)"
	case "Number":
		return subject
	default:
		return ""
	}
}

func (g *generator) jsonConstructorPatternCondition(subject string, pattern *ir.ConstructorPattern) (string, bool) {
	kind := ""
	switch pattern.Name {
	case "Array":
		kind = "[]any"
	case "Object":
		kind = "map[string]any"
	case "String":
		kind = "string"
	case "Bool":
		kind = "bool"
	case "Number":
		if len(pattern.Args) == 0 {
			return fmt.Sprintf("func() bool { switch %s.(type) { case int, int8, int16, int64, uint, uint8, uint16, uint64, float32, float64: return true; default: return false } }()", subject), true
		}
		payload := g.nextTemp("json")
		condition := g.patternCondition(payload, pattern.Args[0])
		if condition == "true" {
			return fmt.Sprintf("func() bool { switch %s.(type) { case int, int8, int16, int64, uint, uint8, uint16, uint64, float32, float64: return true; default: return false } }()", subject), true
		}
		return fmt.Sprintf("func() bool { switch %s := %s.(type) { case int, int8, int16, int64, uint, uint8, uint16, uint64, float32, float64: return %s; default: return false } }()", payload, subject, condition), true
	case "Null":
		return subject + " == nil", true
	default:
		return "", false
	}
	if len(pattern.Args) == 0 {
		return fmt.Sprintf("func() bool { _, ok := %s.(%s); return ok }()", subject, kind), true
	}
	payload := g.nextTemp("json")
	condition := g.patternCondition(payload, pattern.Args[0])
	if condition == "true" {
		return fmt.Sprintf("func() bool { _, ok := %s.(%s); return ok }()", subject, kind), true
	}
	return fmt.Sprintf("func() bool { %s, ok := %s.(%s); if !ok { return false }; return %s }()", payload, subject, kind, condition), true
}

func (g *generator) arrayPatternCondition(subject string, pattern *ir.ArrayPattern) string {
	if irArrayPatternHasBits(pattern) {
		return g.bitArrayPatternCondition(subject, pattern)
	}
	length := goSequenceLength(subject, pattern.SubjectType)
	required := g.arrayPatternRequiredWidth(pattern)
	parts := []string{}
	if pattern.RestIndex >= 0 {
		parts = append(parts, fmt.Sprintf("%s >= %s", length, required))
	} else {
		parts = append(parts, fmt.Sprintf("%s == %s", length, required))
	}
	for idx, elem := range pattern.Elements {
		if spread, ok := elem.(*ir.SequenceSpreadPattern); ok {
			parts = append(parts, g.sequenceSpreadPatternCondition(subject, pattern.SubjectType, g.arrayPatternElementIndex(subject, pattern, idx), spread))
			continue
		}
		parts = append(parts, g.patternCondition(goSequenceIndex(subject, pattern.SubjectType, g.arrayPatternElementIndex(subject, pattern, idx)), elem))
	}
	return strings.Join(parts, " && ")
}

func (g *generator) sequenceSpreadPatternCondition(subject string, subjectType checker.Type, start string, pattern *ir.SequenceSpreadPattern) string {
	spread := g.nextTemp("spread")
	idx := g.nextTemp("idx")
	return fmt.Sprintf("func() bool { %s := %s; for %s := 0; %s < %s; %s++ { if %s != %s { return false } }; return true }()",
		spread,
		g.expr(pattern.Value),
		idx,
		idx,
		goSequenceLength(spread, pattern.Type),
		idx,
		goSequenceIndex(subject, subjectType, fmt.Sprintf("(%s + %s)", start, idx)),
		goSequenceIndex(spread, pattern.Type, idx),
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
	lengthBits := goSequenceLength(subject, pattern.SubjectType) + " * 8"
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
			parts = append(parts, "false")
			continue
		}
		offset := bitPatternOffset(pattern, idx)
		parts = append(parts, g.patternCondition(g.bitPatternValueExpr(subject, pattern.SubjectType, offset, bit), bit.Value))
	}
	return strings.Join(parts, " && ")
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
	ret := bitPatternGoType(pattern)
	start := fmt.Sprintf("%d", offset)
	if offset < 0 {
		start = fmt.Sprintf("%s * 8 - %d", goSequenceLength(subject, typ), -offset)
	}
	value := "__out"
	if pattern.Signed {
		value = fmt.Sprintf("func(__value uint64) %s { if %d < 64 && (__value & (uint64(1) << uint(%d))) != 0 { __mask := ^uint64(0); __value |= __mask << uint(%d) }; return %s(__value) }(__out)", ret, pattern.Width, pattern.Width-1, pattern.Width, ret)
	} else {
		value = fmt.Sprintf("%s(__out)", ret)
	}
	if pattern.Endian == "le" {
		byteExpr := g.bitPatternByteExpr(subject, typ, "__byteIndex")
		return fmt.Sprintf("func() %s { var __out uint64; __start := (%s) / 8; for __i := 0; __i < %d; __i++ { __byteIndex := __start + __i; __out |= uint64(%s) << uint(8 * __i) }; return %s }()", ret, start, pattern.Width/8, byteExpr, value)
	}
	byteExpr := g.bitPatternByteExpr(subject, typ, "__bitIndex / 8")
	step := "__out = (__out << 1) | __bit"
	return fmt.Sprintf("func() %s { var __out uint64; for __i := 0; __i < %d; __i++ { __bitIndex := %s + __i; __bit := (uint64(%s) >> uint(7 - (__bitIndex %% 8))) & 1; %s }; return %s }()", ret, pattern.Width, start, byteExpr, step, value)
}

func (g *generator) bitPatternByteExpr(subject string, typ checker.Type, index string) string {
	switch typ {
	case checker.Bytes:
		return fmt.Sprintf("%s.GetUInt8(%s)", subject, index)
	default:
		return fmt.Sprintf("%s[%s]", subject, index)
	}
}

func bitPatternGoType(pattern *ir.BitPattern) string {
	if pattern.Width > 32 {
		if pattern.Signed {
			return "int64"
		}
		return "uint64"
	}
	if pattern.Signed {
		return "int"
	}
	return "uint"
}

func (g *generator) mapPatternCondition(subject string, pattern *ir.MapPattern) string {
	if pattern.Access == "object" || pattern.SubjectType == checker.Object {
		return g.objectMapPatternCondition(subject, pattern)
	}
	if pattern.Access == "get" {
		return g.mapLikePatternCondition(subject, pattern)
	}
	parts := make([]string, 0, len(pattern.Entries))
	for _, entry := range pattern.Entries {
		key := g.expr(entry.Key)
		value := g.nextTemp("pattern")
		ok := g.nextTemp("ok")
		condition := g.patternCondition(value, entry.Pattern)
		target := value
		if condition == "true" {
			target = "_"
		}
		if entry.Optional {
			if condition == "true" {
				parts = append(parts, "true")
				continue
			}
			raw := g.nextTemp("pattern")
			parts = append(parts, fmt.Sprintf("func() bool { %s, %s := %s[%s]; var %s any; if %s { %s = %s } else { %s = any(nil) }; return %s }()", raw, ok, subject, key, value, ok, value, raw, value, condition))
			continue
		}
		parts = append(parts, fmt.Sprintf("func() bool { %s, %s := %s[%s]; return %s && (%s) }()", target, ok, subject, key, ok, condition))
	}
	if len(parts) == 0 {
		return "true"
	}
	return strings.Join(parts, " && ")
}

func (g *generator) mapLikePatternCondition(subject string, pattern *ir.MapPattern) string {
	parts := make([]string, 0, len(pattern.Entries))
	for _, entry := range pattern.Entries {
		key := g.expr(entry.Key)
		value := g.nextTemp("pattern")
		condition := g.patternCondition(value, entry.Pattern)
		if entry.Optional {
			if condition == "true" {
				parts = append(parts, "true")
				continue
			}
			parts = append(parts, fmt.Sprintf("func() bool { %s := %s; return %s }()", value, g.goMapLikeGet(subject, pattern.SubjectType, key), condition))
			continue
		}
		parts = append(parts, fmt.Sprintf("func() bool { %s := %s; return %s != nil && (%s) }()", value, g.goMapLikeGet(subject, pattern.SubjectType, key), value, condition))
	}
	if len(parts) == 0 {
		return "true"
	}
	return strings.Join(parts, " && ")
}

func (g *generator) goMapLikeGet(subject string, typ checker.Type, key string) string {
	if g.mapGetters != nil {
		if getter := g.mapGetters[subject]; getter != "" {
			return fmt.Sprintf("%s(%s)", getter, key)
		}
	}
	return fmt.Sprintf("%s.%s(%s)", subject, mangleIdent("get"), key)
}

func (g *generator) objectMapPatternCondition(subject string, pattern *ir.MapPattern) string {
	parts := make([]string, 0, len(pattern.Entries))
	for _, entry := range pattern.Entries {
		key, ok := entry.Key.(*ir.StringLiteral)
		if !ok {
			parts = append(parts, "false")
			continue
		}
		value := g.nextTemp("field")
		exists := g.nextTemp("ok")
		condition := g.patternCondition(value, entry.Pattern)
		if entry.Optional {
			parts = append(parts, fmt.Sprintf("func() bool { obj, ok := %s.(map[string]any); if !ok { return false }; %s, _ := obj[%q]; return %s }()", subject, value, key.Value, condition))
			continue
		}
		parts = append(parts, fmt.Sprintf("func() bool { obj, ok := %s.(map[string]any); if !ok { return false }; %s, %s := obj[%q]; if !%s { return false }; return %s }()", subject, value, exists, key.Value, exists, condition))
	}
	if len(parts) == 0 {
		return fmt.Sprintf("func() bool { _, ok := %s.(map[string]any); return ok }()", subject)
	}
	return strings.Join(parts, " && ")
}

func (g *generator) objectPatternCondition(subject string, pattern *ir.ObjectPattern) string {
	if pattern.SubjectType == checker.Object || pattern.SubjectType == checker.Unknown {
		return g.dynamicObjectPatternCondition(subject, pattern)
	}
	parts := make([]string, 0, len(pattern.Fields))
	for _, field := range pattern.Fields {
		if field.Optional && !field.Exists {
			continue
		}
		condition := g.patternCondition(goObjectFieldAccess(subject, field.Name), field.Pattern)
		if condition != "true" {
			parts = append(parts, condition)
		}
	}
	if len(parts) == 0 {
		return "true"
	}
	return strings.Join(parts, " && ")
}

func (g *generator) dynamicObjectPatternCondition(subject string, pattern *ir.ObjectPattern) string {
	parts := make([]string, 0, len(pattern.Fields))
	for _, field := range pattern.Fields {
		value := g.nextTemp("field")
		ok := g.nextTemp("ok")
		condition := g.patternCondition(value, field.Pattern)
		parts = append(parts, fmt.Sprintf("func() bool { obj, ok := %s.(map[string]any); if !ok { return false }; %s, %s := obj[%q]; if !%s { return false }; return %s }()", subject, value, ok, field.Name, ok, condition))
	}
	if len(parts) == 0 {
		return fmt.Sprintf("func() bool { _, ok := %s.(map[string]any); return ok }()", subject)
	}
	return strings.Join(parts, " && ")
}

func (g *generator) signalRuntime() {
	g.line("var runeSignalDepth int")
	g.line("var runeEffectQueue []func()")
	g.line("")
	g.line("func runeScheduleEffect(effect func()) {")
	g.indent++
	g.line("runeEffectQueue = append(runeEffectQueue, effect)")
	g.line("runeFlushEffects()")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeFlushEffects() {")
	g.indent++
	g.line("for len(runeEffectQueue) > 0 {")
	g.indent++
	g.line("effect := runeEffectQueue[0]")
	g.line("runeEffectQueue = runeEffectQueue[1:]")
	g.line("effect()")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("type runeSignal[T comparable] struct {")
	g.indent++
	g.line("value T")
	g.line("watchers []func(T, T)")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func newRuneSignal[T comparable](value T) *runeSignal[T] {")
	g.indent++
	g.line("return &runeSignal[T]{value: value}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func (s *runeSignal[T]) Get() T {")
	g.indent++
	g.line("return s.value")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func (s *runeSignal[T]) Set(value T) {")
	g.indent++
	g.line("old := s.value")
	g.line("if old == value {")
	g.indent++
	g.line("return")
	g.indent--
	g.line("}")
	g.line("s.value = value")
	g.line("runeSignalDepth++")
	g.line("for _, watcher := range s.watchers {")
	g.indent++
	g.line("watcher(old, value)")
	g.indent--
	g.line("}")
	g.line("runeSignalDepth--")
	g.line("if runeSignalDepth == 0 {")
	g.indent++
	g.line("runeFlushEffects()")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func (s *runeSignal[T]) Watch(watcher func(T, T)) {")
	g.indent++
	g.line("s.watchers = append(s.watchers, watcher)")
	g.indent--
	g.line("}")
}
