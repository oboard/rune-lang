package tscodegen

import (
	"fmt"
	"strings"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
)

func (g *generator) structType(typ *ir.StructType) {
	g.linef("type %s = {", mangleIdent(typ.Name))
	g.indent++
	for _, field := range typ.Fields {
		g.linef("%s: %s;", tsPropertyName(field.Name), tsType(field.Type))
	}
	g.indent--
	g.line("};")
}

func (g *generator) enumType(enum *ir.EnumType) {
	if enumHasPayload(enum) {
		g.linef("type %s = { tag: number; payload: any[] };", mangleIdent(enum.Name))
		g.linef("const %s = {", mangleIdent(enum.Name))
		g.indent++
		for i, member := range enum.Members {
			g.linef("%s: %d,", tsPropertyName(member.Name), i)
		}
		g.indent--
		g.line("} as const;")
		return
	}
	g.linef("type %s = number;", mangleIdent(enum.Name))
	g.linef("const %s = {", mangleIdent(enum.Name))
	g.indent++
	for i, member := range enum.Members {
		value := i
		if member.HasValue {
			value = member.Value
		}
		g.linef("%s: %d,", tsPropertyName(member.Name), value)
	}
	g.indent--
	g.line("} as const;")
}

func (g *generator) constDecl(constant *ir.ConstDecl) {
	g.linef("const %s: %s = %s;", mangleIdent(constant.Name), tsType(constant.Type), g.expr(constant.Value))
}

func (g *generator) function(fn *ir.Function) error {
	params := make([]string, 0, len(fn.Params))
	for _, param := range fn.Params {
		params = append(params, fmt.Sprintf("%s: %s", mangleIdent(param.Name), tsType(param.Type)))
	}
	if fn.Return == checker.WebComponent && !fn.Routine {
		return g.webComponentFunction(FunctionSymbolName(fn), params, fn)
	}
	if fn.Routine {
		return g.routineFunction(fn, params)
	}
	prefix := "function"
	ret := tsType(fn.Return)
	g.linef("%s %s%s(%s): %s {", prefix, FunctionSymbolName(fn), tsGenerics(fn.Generics, fn.GenericConstraints), strings.Join(params, ", "), ret)
	g.indent++
	g.pushSignalScope()
	g.pushReactiveScope()
	if err := g.body(fn, fn.Body, fn.Return); err != nil {
		return err
	}
	g.popReactiveScope()
	g.popSignalScope()
	g.indent--
	g.line("}")
	return nil
}

func (g *generator) routineFunction(fn *ir.Function, params []string) error {
	ret := tsType(fn.Return)
	g.linef("function %s(%s): Promise<%s> {", FunctionSymbolName(fn), strings.Join(params, ", "), ret)
	g.indent++
	g.linef("return runeGo(async (): Promise<%s> => {", ret)
	g.indent++
	g.pushSignalScope()
	g.pushReactiveScope()
	if err := g.body(fn, fn.Body, fn.Return); err != nil {
		return err
	}
	g.popReactiveScope()
	g.popSignalScope()
	g.indent--
	g.line("});")
	g.indent--
	g.line("}")
	return nil
}

func (g *generator) method(typeName string, fn *ir.Function) error {
	params := []string{}
	if !fn.Static {
		params = append(params, fmt.Sprintf("%s: %s", mangleIdent("this"), mangleIdent(typeName)))
	}
	for _, param := range fn.Params {
		params = append(params, fmt.Sprintf("%s: %s", mangleIdent(param.Name), tsType(param.Type)))
	}
	if fn.Return == checker.WebComponent && !fn.Routine {
		return g.webComponentFunction(mangleMethod(typeName, fn.Name), params, fn)
	}
	if fn.Routine {
		return g.routineMethod(typeName, fn, params)
	}
	prefix := "function"
	ret := tsType(fn.Return)
	g.linef("%s %s%s(%s): %s {", prefix, mangleMethod(typeName, fn.Name), tsGenerics(fn.Generics, fn.GenericConstraints), strings.Join(params, ", "), ret)
	g.indent++
	g.pushSignalScope()
	g.pushReactiveScope()
	if !fn.Static {
		g.thisNames = append(g.thisNames, mangleIdent("this"))
	}
	if err := g.body(fn, fn.Body, fn.Return); err != nil {
		return err
	}
	if !fn.Static {
		g.thisNames = g.thisNames[:len(g.thisNames)-1]
	}
	g.popReactiveScope()
	g.popSignalScope()
	g.indent--
	g.line("}")
	return nil
}

func (g *generator) routineMethod(typeName string, fn *ir.Function, params []string) error {
	ret := tsType(fn.Return)
	g.linef("function %s(%s): Promise<%s> {", mangleMethod(typeName, fn.Name), strings.Join(params, ", "), ret)
	g.indent++
	g.linef("return runeGo(async (): Promise<%s> => {", ret)
	g.indent++
	g.pushSignalScope()
	g.pushReactiveScope()
	g.thisNames = append(g.thisNames, mangleIdent("this"))
	if err := g.body(fn, fn.Body, fn.Return); err != nil {
		return err
	}
	g.thisNames = g.thisNames[:len(g.thisNames)-1]
	g.popReactiveScope()
	g.popSignalScope()
	g.indent--
	g.line("});")
	g.indent--
	g.line("}")
	return nil
}

func (g *generator) webComponentFunction(name string, params []string, fn *ir.Function) error {
	g.linef("function %s(%s): %s {", name, strings.Join(params, ", "), tsType(fn.Return))
	g.indent++
	g.line("return class extends HTMLElement {")
	g.indent++
	g.line("connectedCallback(): void {")
	g.indent++
	self := g.nextTemp("__self")
	g.linef("const %s = this as HTMLElement & { __runeMounted?: boolean };", self)
	g.linef("if (%s.__runeMounted) {", self)
	g.indent++
	g.line("return;")
	g.indent--
	g.line("}")
	g.linef("%s.__runeMounted = true;", self)
	root := g.nextTemp("__root")
	g.linef("const %s = ((): HTMLElement => {", root)
	g.indent++
	g.pushSignalScope()
	g.pushReactiveScope()
	err := g.body(fn, fn.Body, checker.HTMLElement)
	g.popReactiveScope()
	g.popSignalScope()
	if err != nil {
		return err
	}
	g.indent--
	g.line("})();")
	g.linef("this.appendChild(%s);", root)
	g.indent--
	g.line("}")
	g.indent--
	g.line("};")
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
				g.line(expr + ";")
			}
		} else {
			g.lineExpr("return ", g.returnExpr(expr, ret), ";")
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
			if reactive, ok := s.Value.(*ir.ReactiveLiteral); ok {
				kind := "const"
				if s.Mutable {
					kind = "let"
				}
				g.linef("%s %s = %s;", kind, mangleIdent(s.Name), g.expr(reactive))
				g.addReactive(s.Name, s.Value.ResultType())
				continue
			}
			if s.Signal || g.exprUsesSignal(s.Value) {
				g.linef("const %s = %s;", mangleIdent(s.Name), g.signalInitialValue(s.Value))
				g.addSignal(s.Name, s.Value.ResultType())
				deps := g.exprSignalDeps(s.Value)
				g.setSignalDeps(s.Name, deps)
				for _, dep := range deps {
					g.linef("runeWatch(%s, () => { %s.set(%s); });", mangleIdent(dep), mangleIdent(s.Name), g.expr(s.Value))
				}
				continue
			}
			kind := "let"
			value := g.expr(s.Value)
			if _, ok := s.Value.(*ir.AnonymousObjectLiteral); ok {
				value = g.withThisName(mangleIdent(s.Name), func() string {
					return g.expr(s.Value)
				})
			}
			g.linef("%s %s = %s;", kind, mangleIdent(s.Name), value)
		case *ir.ObjectDestructureStmt:
			if unwrap, ok := s.Value.(*ir.ResultUnwrapExpr); ok {
				g.resultUnwrapObjectDestructure(s, unwrap, ret)
				continue
			}
			g.objectDestructure(s, g.expr(s.Value))
		case *ir.AssignStmt:
			if g.isSignal(s.Name) {
				g.linef("%s.set(%s);", mangleIdent(s.Name), g.expr(s.Value))
				continue
			}
			if g.isReactive(s.Name) {
				g.linef("%s.set(%s);", mangleIdent(s.Name), g.expr(s.Value))
				continue
			}
			g.linef("%s = %s;", mangleIdent(s.Name), g.expr(s.Value))
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
			expr := g.expr(s.Expr)
			if expr == "" {
				continue
			}
			if last && ret != checker.Void {
				g.lineExpr("return ", g.returnExpr(s.Expr, ret), ";")
			} else {
				g.line(g.stmtExpr(s.Expr) + ";")
			}
		}
	}
	if ret != checker.Void && len(block.Statements) == 0 {
		g.linef("return %s;", g.zeroValue(ret))
	}
	return nil
}

func (g *generator) effectScope(block *ir.BlockExpr) {
	name := g.nextTemp("__effect")
	g.linef("const %s = () => {", name)
	g.indent++
	_ = g.block(block, checker.Void)
	g.indent--
	g.line("};")
	pending := g.nextTemp("__effectPending")
	schedule := g.nextTemp("__scheduleEffect")
	g.linef("let %s = false;", pending)
	g.linef("const %s = () => {", schedule)
	g.indent++
	g.linef("if (%s) return;", pending)
	g.linef("%s = true;", pending)
	g.linef("runeScheduleEffect(() => { %s = false; %s(); });", pending, name)
	g.indent--
	g.line("};")
	g.linef("%s();", name)
	for _, dep := range g.effectSignalDeps(block) {
		g.linef("runeWatch(%s, %s);", mangleIdent(dep), schedule)
	}
}

func (g *generator) conditionalExprStmt(expr *ir.TernaryExpr) {
	g.linef("if (%s) {", g.expr(expr.Condition))
	g.indent++
	if consequence := g.stmtExpr(expr.Consequence); consequence != "" {
		g.line(consequence + ";")
	}
	g.indent--
	g.line("}")
}

func (g *generator) resultUnwrapLet(name string, unwrap *ir.ResultUnwrapExpr, ret checker.Type) {
	tmp := g.nextTemp("__result")
	g.linef("const %s = %s;", tmp, g.expr(unwrap.Expr))
	g.linef("if (!%s.ok) {", tmp)
	g.indent++
	g.linef("return %s;", g.resultErrReturn(ret, tmp+".error"))
	g.indent--
	g.line("}")
	g.linef("const %s = %s.value;", mangleIdent(name), tmp)
}

func (g *generator) resultUnwrapObjectDestructure(stmt *ir.ObjectDestructureStmt, unwrap *ir.ResultUnwrapExpr, ret checker.Type) {
	tmp := g.nextTemp("__result")
	g.linef("const %s = %s;", tmp, g.expr(unwrap.Expr))
	g.linef("if (!%s.ok) {", tmp)
	g.indent++
	g.linef("return %s;", g.resultErrReturn(ret, tmp+".error"))
	g.indent--
	g.line("}")
	g.objectDestructure(stmt, tmp+".value")
}

func (g *generator) objectDestructure(stmt *ir.ObjectDestructureStmt, source string) {
	signal := stmt.Signal || g.exprUsesSignal(stmt.Value)
	if !signal {
		kind := "const"
		if stmt.Mutable {
			kind = "let"
		}
		bindings := make([]string, 0, len(stmt.Fields))
		for _, field := range stmt.Fields {
			bindings = append(bindings, fmt.Sprintf("%s: %s", tsPropertyName(field.Field), mangleIdent(field.Name)))
		}
		g.linef("%s { %s } = %s;", kind, strings.Join(bindings, ", "), source)
		return
	}
	tmp := g.nextTemp("__destructure")
	g.linef("const %s = %s;", tmp, source)
	for _, field := range stmt.Fields {
		g.linef("const %s = runeSignal(%s);", mangleIdent(field.Name), tsPropertyAccess(tmp, field.Field))
		g.addSignal(field.Name, field.Type)
		g.setSignalDeps(field.Name, g.exprSignalDeps(stmt.Value))
	}
	for _, dep := range g.exprSignalDeps(stmt.Value) {
		for _, field := range stmt.Fields {
			g.linef("runeWatch(%s, () => { %s.set(%s); });", mangleIdent(dep), mangleIdent(field.Name), tsPropertyAccess(g.expr(stmt.Value), field.Field))
		}
	}
}

func (g *generator) resultUnwrapExprStmt(unwrap *ir.ResultUnwrapExpr, ret checker.Type, last bool) {
	tmp := g.nextTemp("__result")
	g.linef("const %s = %s;", tmp, g.expr(unwrap.Expr))
	g.linef("if (!%s.ok) {", tmp)
	g.indent++
	g.linef("return %s;", g.resultErrReturn(ret, tmp+".error"))
	g.indent--
	g.line("}")
	if last && ret != checker.Void {
		g.linef("return %s;", g.returnRawExpr(unwrap, ret, tmp+".value"))
	}
}

func (g *generator) resultErrReturn(ret checker.Type, errExpr string) string {
	okType, errType := resultTypeArgs(ret)
	if okType == checker.Unknown {
		return g.zeroValue(ret)
	}
	return fmt.Sprintf("runeErr<%s, %s>(%s)", tsType(okType), tsType(errType), errExpr)
}

func (g *generator) returnExpr(expr ir.Expr, ret checker.Type) string {
	expected := ret
	if okType, _ := resultTypeArgs(ret); okType != checker.Unknown && expr != nil && expr.ResultType() == okType {
		expected = okType
	}
	return g.returnRawExpr(expr, ret, g.exprWithExpected(expr, expected))
}

func (g *generator) returnRawExpr(expr ir.Expr, ret checker.Type, raw string) string {
	okType, errType := resultTypeArgs(ret)
	if okType == checker.Unknown {
		return raw
	}
	if expr != nil && expr.ResultType() == okType {
		return fmt.Sprintf("runeOk<%s, %s>(%s)", tsType(okType), tsType(errType), raw)
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
		g.line(line + ";")
	}
	hasDefault := false
	for i, branch := range block.Branches {
		prefix := "if"
		if i > 0 {
			prefix = "else if"
		}
		if _, ok := branch.Pattern.(*ir.WildcardPattern); ok {
			hasDefault = true
			if i == 0 {
				g.line("{")
			} else {
				g.line("else {")
			}
		} else {
			g.linef("%s (%s) {", prefix, g.patternCondition(subject, branch.Pattern))
		}
		g.indent++
		if binding := g.patternBinding(subject, branch.Pattern); binding != "" {
			g.line(binding)
		}
		if ret == checker.Void {
			if expr := g.expr(branch.Expr); expr != "" {
				g.line(g.stmtExpr(branch.Expr) + ";")
			}
		} else {
			g.linef("return %s;", g.returnExpr(branch.Expr, ret))
		}
		g.indent--
		g.line("}")
	}
	if ret != checker.Void && !hasDefault {
		g.linef("return %s;", g.zeroValue(ret))
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
			return fmt.Sprintf("%s === %s", subject, mangleIdent(p.Name))
		}
		return "true"
	case *ir.LiteralPattern:
		return fmt.Sprintf("%s === %s", subject, g.expr(p.Value))
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
			parts = append(parts, g.patternCondition(fmt.Sprintf("%s[%d]", subject, i), elem))
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
		memberExpr := tsPropertyAccess(mangleIdent(enum.Name), member.Name)
		if enumHasPayload(enum) {
			return fmt.Sprintf("%s.tag === %s", subject, memberExpr), true
		}
		return fmt.Sprintf("%s === %s", subject, memberExpr), true
	}
	return "", false
}

func (g *generator) enumConstructorPatternCondition(subject string, pattern *ir.ConstructorPattern) (string, bool) {
	enum, member, ok := g.enumMemberForConstructor(pattern.SubjectType, pattern.Name)
	if !ok || !enumHasPayload(enum) {
		return "", false
	}
	parts := []string{fmt.Sprintf("%s.tag === %s", subject, tsPropertyAccess(mangleIdent(enum.Name), member.Name))}
	if pattern.Rest {
		parts = append(parts, fmt.Sprintf("%s.payload.length >= %d", subject, len(pattern.Args)))
	} else {
		parts = append(parts, fmt.Sprintf("%s.payload.length === %d", subject, len(pattern.Args)))
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
		return fmt.Sprintf("(%s.payload[%d] as %s)", subject, idx, tsType(member.Params[idx].Type))
	}
	if idx != 0 {
		return ""
	}
	switch pattern.Name {
	case "Ok":
		return subject + ".value"
	case "Err":
		return subject + ".error"
	case "Array":
		return subject + " as any[]"
	case "Object":
		return subject + " as Record<string, any>"
	case "String":
		return subject + " as string"
	case "Bool":
		return subject + " as boolean"
	case "Number":
		return subject + " as number"
	default:
		return ""
	}
}

func (g *generator) jsonConstructorPatternCondition(subject string, pattern *ir.ConstructorPattern) (string, bool) {
	check := ""
	switch pattern.Name {
	case "Array":
		check = fmt.Sprintf("Array.isArray(%s)", subject)
	case "Object":
		check = fmt.Sprintf("%s !== null && typeof %s === \"object\" && !Array.isArray(%s)", subject, subject, subject)
	case "String":
		check = fmt.Sprintf("typeof %s === \"string\"", subject)
	case "Bool":
		check = fmt.Sprintf("typeof %s === \"boolean\"", subject)
	case "Number":
		check = fmt.Sprintf("typeof %s === \"number\"", subject)
	case "Null":
		return subject + " === null", true
	default:
		return "", false
	}
	if len(pattern.Args) == 0 {
		return check, true
	}
	payload := g.constructorPayload(subject, pattern, 0)
	condition := g.patternCondition(payload, pattern.Args[0])
	return fmt.Sprintf("(%s && (%s))", check, condition), true
}

func (g *generator) arrayPatternCondition(subject string, pattern *ir.ArrayPattern) string {
	if irArrayPatternHasBits(pattern) {
		return g.bitArrayPatternCondition(subject, pattern)
	}
	length := tsSequenceLength(subject, pattern.SubjectType)
	required := g.arrayPatternRequiredWidth(pattern)
	parts := []string{}
	if pattern.RestIndex >= 0 {
		parts = append(parts, fmt.Sprintf("%s >= %s", length, required))
	} else {
		parts = append(parts, fmt.Sprintf("%s === %s", length, required))
	}
	for idx, elem := range pattern.Elements {
		if spread, ok := elem.(*ir.SequenceSpreadPattern); ok {
			parts = append(parts, g.sequenceSpreadPatternCondition(subject, pattern.SubjectType, g.arrayPatternElementIndex(subject, pattern, idx), spread))
			continue
		}
		parts = append(parts, g.patternCondition(tsSequenceIndex(subject, pattern.SubjectType, g.arrayPatternElementIndex(subject, pattern, idx)), elem))
	}
	return strings.Join(parts, " && ")
}

func (g *generator) sequenceSpreadPatternCondition(subject string, subjectType checker.Type, start string, pattern *ir.SequenceSpreadPattern) string {
	spread := g.nextTemp("spread")
	idx := g.nextTemp("idx")
	return fmt.Sprintf("(() => { const %s = %s; for (let %s = 0; %s < %s; %s++) { if (%s !== %s) { return false; } } return true; })()",
		spread,
		g.expr(pattern.Value),
		idx,
		idx,
		tsSequenceLength(spread, pattern.Type),
		idx,
		tsSequenceIndex(subject, subjectType, fmt.Sprintf("(%s + %s)", start, idx)),
		tsSequenceIndex(spread, pattern.Type, idx),
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
	lengthBits := tsSequenceLength(subject, pattern.SubjectType) + " * 8"
	requiredBits := bitPatternRequiredBits(pattern)
	parts := []string{}
	if pattern.RestIndex >= 0 {
		parts = append(parts, fmt.Sprintf("%s >= %d", lengthBits, requiredBits))
	} else {
		parts = append(parts, fmt.Sprintf("%s === %d", lengthBits, requiredBits))
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
	ret := bitPatternTSType(pattern)
	start := fmt.Sprintf("%d", offset)
	if offset < 0 {
		start = fmt.Sprintf("%s * 8 - %d", tsSequenceLength(subject, typ), -offset)
	}
	if ret == "bigint" {
		if pattern.Endian == "le" {
			byteExpr := g.bitPatternByteExpr(subject, typ, "__byteIndex")
			value := "__out"
			if pattern.Signed {
				value = fmt.Sprintf("(() => { let __value = __out; if (%d < 64 && (__out & (1n << BigInt(%d))) !== 0n) __value -= 1n << BigInt(%d); return __value; })()", pattern.Width, pattern.Width-1, pattern.Width)
			}
			return fmt.Sprintf("((): bigint => { let __out = 0n; const __start = Math.floor((%s) / 8); for (let __i = 0; __i < %d; __i++) { const __byteIndex = __start + __i; __out |= BigInt(%s) << BigInt(8 * __i); } return %s; })()", start, pattern.Width/8, byteExpr, value)
		}
		byteExpr := g.bitPatternByteExpr(subject, typ, "__bitIndex >> 3")
		step := "__out = (__out << 1n) | __bit"
		value := "__out"
		if pattern.Signed {
			value = fmt.Sprintf("(() => { let __value = __out; if (%d < 64 && (__out & (1n << BigInt(%d))) !== 0n) __value -= 1n << BigInt(%d); return __value; })()", pattern.Width, pattern.Width-1, pattern.Width)
		}
		return fmt.Sprintf("((): bigint => { let __out = 0n; for (let __i = 0; __i < %d; __i++) { const __bitIndex = %s + __i; const __bit = BigInt((%s >> (7 - (__bitIndex %% 8))) & 1); %s; } return %s; })()", pattern.Width, start, byteExpr, step, value)
	}
	if pattern.Endian == "le" {
		byteExpr := g.bitPatternByteExpr(subject, typ, "__byteIndex")
		value := "__out"
		if pattern.Signed {
			value = fmt.Sprintf("(() => { let __value = __out; if (%d < 32 && (__value & (1 << %d)) !== 0) __value -= 1 << %d; return __value; })()", pattern.Width, pattern.Width-1, pattern.Width)
		}
		return fmt.Sprintf("((): number => { let __out = 0; const __start = Math.floor((%s) / 8); for (let __i = 0; __i < %d; __i++) { const __byteIndex = __start + __i; __out |= %s << (8 * __i); } return %s; })()", start, pattern.Width/8, byteExpr, value)
	}
	byteExpr := g.bitPatternByteExpr(subject, typ, "__bitIndex >> 3")
	step := "__out = (__out << 1) | __bit"
	value := "__out"
	if pattern.Signed {
		value = fmt.Sprintf("(() => { let __value = __out; if (%d < 32 && (__value & (1 << %d)) !== 0) __value -= 1 << %d; return __value; })()", pattern.Width, pattern.Width-1, pattern.Width)
	}
	return fmt.Sprintf("((): %s => { let __out = 0; for (let __i = 0; __i < %d; __i++) { const __bitIndex = %s + __i; const __bit = (%s >> (7 - (__bitIndex %% 8))) & 1; %s; } return %s; })()", ret, pattern.Width, start, byteExpr, step, value)
}

func (g *generator) bitPatternByteExpr(subject string, typ checker.Type, index string) string {
	switch typ {
	case checker.Bytes:
		return fmt.Sprintf("%s.getUint8(%s)", subject, index)
	default:
		return fmt.Sprintf("%s[%s]", subject, index)
	}
}

func bitPatternTSType(pattern *ir.BitPattern) string {
	if pattern.Width > 32 {
		return "bigint"
	}
	return "number"
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
		key := g.nextTemp("__key")
		value := g.nextTemp("__pattern")
		condition := g.patternCondition(value, entry.Pattern)
		valueDecl := ""
		if condition != "true" {
			valueDecl = fmt.Sprintf(" const %s = %s.get(%s)!;", value, subject, key)
		}
		if entry.Optional {
			parts = append(parts, fmt.Sprintf("(() => { const %s = %s; const %s = %s.has(%s) ? %s.get(%s)! : null; return %s; })()", key, g.expr(entry.Key), value, subject, key, subject, key, condition))
			continue
		}
		parts = append(parts, fmt.Sprintf("(() => { const %s = %s; if (!%s.has(%s)) return false;%s return %s; })()", key, g.expr(entry.Key), subject, key, valueDecl, condition))
	}
	if len(parts) == 0 {
		return "true"
	}
	return strings.Join(parts, " && ")
}

func (g *generator) mapLikePatternCondition(subject string, pattern *ir.MapPattern) string {
	parts := make([]string, 0, len(pattern.Entries))
	for _, entry := range pattern.Entries {
		key := g.nextTemp("__key")
		value := g.nextTemp("__pattern")
		condition := g.patternCondition(value, entry.Pattern)
		if entry.Optional {
			parts = append(parts, fmt.Sprintf("(() => { const %s = %s; const %s = %s; return %s; })()", key, g.expr(entry.Key), value, g.tsMapLikeGet(subject, pattern.SubjectType, key), condition))
			continue
		}
		parts = append(parts, fmt.Sprintf("(() => { const %s = %s; const %s = %s; return %s !== null && (%s); })()", key, g.expr(entry.Key), value, g.tsMapLikeGet(subject, pattern.SubjectType, key), value, condition))
	}
	if len(parts) == 0 {
		return "true"
	}
	return strings.Join(parts, " && ")
}

func (g *generator) tsMapLikeGet(subject string, typ checker.Type, key string) string {
	if g.mapGetters != nil {
		if getter := g.mapGetters[subject]; getter != "" {
			return fmt.Sprintf("%s(%s)", getter, key)
		}
	}
	return fmt.Sprintf("%s(%s, %s)", mangleMethod(baseTypeName(typ), "get"), subject, key)
}

func (g *generator) objectMapPatternCondition(subject string, pattern *ir.MapPattern) string {
	parts := make([]string, 0, len(pattern.Entries))
	for _, entry := range pattern.Entries {
		key, ok := entry.Key.(*ir.StringLiteral)
		if !ok {
			parts = append(parts, "false")
			continue
		}
		value := fmt.Sprintf("(%s as any)[%q]", subject, key.Value)
		objectCheck := fmt.Sprintf("%s !== null && typeof %s === \"object\" && !Array.isArray(%s)", subject, subject, subject)
		condition := g.patternCondition(value, entry.Pattern)
		if entry.Optional {
			parts = append(parts, fmt.Sprintf("(%s && (%s))", objectCheck, condition))
			continue
		}
		exists := fmt.Sprintf("Object.prototype.hasOwnProperty.call(%s as any, %q)", subject, key.Value)
		parts = append(parts, fmt.Sprintf("(%s && %s && (%s))", objectCheck, exists, condition))
	}
	if len(parts) == 0 {
		return fmt.Sprintf("(%s !== null && typeof %s === \"object\" && !Array.isArray(%s))", subject, subject, subject)
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
		condition := g.patternCondition(tsPropertyAccess(subject, field.Name), field.Pattern)
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
		value := fmt.Sprintf("(%s as any)[%q]", subject, field.Name)
		exists := fmt.Sprintf("Object.prototype.hasOwnProperty.call(%s as any, %q)", subject, field.Name)
		condition := g.patternCondition(value, field.Pattern)
		parts = append(parts, fmt.Sprintf("(%s !== null && typeof %s === \"object\" && %s && (%s))", subject, subject, exists, condition))
	}
	if len(parts) == 0 {
		return fmt.Sprintf("(%s !== null && typeof %s === \"object\" && !Array.isArray(%s))", subject, subject, subject)
	}
	return strings.Join(parts, " && ")
}

func (g *generator) signalRuntime() {
	g.line("type RuneSignal<T> = {")
	g.indent++
	g.line("get(): T;")
	g.line("set(value: T): void;")
	g.line("mutate<R>(mutator: (value: T) => R): R;")
	g.line("watch(watcher: (oldValue: T, newValue: T) => void): void;")
	g.indent--
	g.line("};")
	g.line("")
	g.line("let runeSignalDepth = 0;")
	g.line("const runeEffectQueue: Array<() => void> = [];")
	g.line("")
	g.line("function runeScheduleEffect(effect: () => void): void {")
	g.indent++
	g.line("runeEffectQueue.push(effect);")
	g.line("runeFlushEffects();")
	g.indent--
	g.line("}")
	g.line("")
	g.line("function runeFlushEffects(): void {")
	g.indent++
	g.line("while (runeEffectQueue.length > 0) {")
	g.indent++
	g.line("runeEffectQueue.shift()!();")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("function runeSignal<T>(initial: T): RuneSignal<T> {")
	g.indent++
	g.line("let value = initial;")
	g.line("const watchers: Array<(oldValue: T, newValue: T) => void> = [];")
	g.line("return {")
	g.indent++
	g.line("get: () => value,")
	g.line("set: (next: T) => {")
	g.indent++
	g.line("const old = value;")
	g.line("if (Object.is(old, next)) return;")
	g.line("value = next;")
	g.line("runeSignalDepth++;")
	g.line("try {")
	g.indent++
	g.line("for (const watcher of watchers) watcher(old, next);")
	g.indent--
	g.line("} finally {")
	g.indent++
	g.line("runeSignalDepth--;")
	g.line("runeFlushEffects();")
	g.indent--
	g.line("}")
	g.indent--
	g.line("},")
	g.line("mutate: <R>(mutator: (value: T) => R): R => {")
	g.indent++
	g.line("const old = value;")
	g.line("const result = mutator(value);")
	g.line("runeSignalDepth++;")
	g.line("try {")
	g.indent++
	g.line("for (const watcher of watchers) watcher(old, value);")
	g.indent--
	g.line("} finally {")
	g.indent++
	g.line("runeSignalDepth--;")
	g.line("if (runeSignalDepth === 0) runeFlushEffects();")
	g.indent--
	g.line("}")
	g.line("return result;")
	g.indent--
	g.line("},")
	g.line("watch: (watcher: (oldValue: T, newValue: T) => void) => { watchers.push(watcher); },")
	g.indent--
	g.line("};")
	g.indent--
	g.line("}")
	g.line("")
	g.line("function runeWatch(source: any, watcher: () => void): void {")
	g.indent++
	g.line("if (source && typeof source.watch === \"function\") {")
	g.indent++
	g.line("source.watch(() => watcher());")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("function runeReactiveArray<T>(initial: T[]): RuneSignal<T[]> {")
	g.indent++
	g.line("return runeSignal(initial);")
	g.indent--
	g.line("}")
	g.line("")
	g.line("function runeReactiveObject<T extends object>(initial: T): RuneSignal<T> {")
	g.indent++
	g.line("return runeSignal(initial);")
	g.indent--
	g.line("}")
}
