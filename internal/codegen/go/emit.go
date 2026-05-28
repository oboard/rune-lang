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
	if !enumHasValueMembers(enum) {
		return
	}
	g.linef("type %s int", mangleIdent(enum.Name))
	if len(enum.Members) == 0 {
		return
	}
	g.line("")
	g.line("const (")
	g.indent++
	for _, member := range enum.Members {
		if !member.HasValue {
			continue
		}
		g.linef("%s %s = %d", mangleEnumMember(enum.Name, member.Name), mangleIdent(enum.Name), member.Value)
	}
	g.indent--
	g.line(")")
}

func enumHasValueMembers(enum *ir.EnumType) bool {
	for _, member := range enum.Members {
		if member.HasValue {
			return true
		}
	}
	return false
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
	g.linef("func %s(%s)%s {", mangleIdent(fn.Name), strings.Join(params, ", "), ret)
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

func (g *generator) method(typ *ir.StructType, fn *ir.Function) error {
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
	g.linef("func (%s %s) %s(%s)%s {", mangleIdent("this"), mangleIdent(typ.Name), mangleIdent(fn.Name), strings.Join(params, ", "), ret)
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
			if s.Signal || g.exprUsesSignal(s.Value) {
				g.linef("%s := newRuneSignal(%s)", mangleIdent(s.Name), g.expr(s.Value))
				g.linef("_ = %s", mangleIdent(s.Name))
				g.addSignal(s.Name, s.Value.ResultType())
				for _, dep := range g.exprSignalDeps(s.Value) {
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
				g.linef("_ = %s", mangleIdent(s.Name))
				continue
			}
			g.linef("%s := %s", mangleIdent(s.Name), g.expr(s.Value))
			g.linef("_ = %s", mangleIdent(s.Name))
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
	g.linef("_ = %s", mangleIdent(name))
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
			g.linef("_ = %s", name)
			g.addSignal(field.Name, field.Type)
			continue
		}
		g.linef("%s := %s", name, value)
		g.linef("_ = %s", name)
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
		return "true"
	case *ir.LiteralPattern:
		return fmt.Sprintf("%s == %s", subject, g.expr(p.Value))
	case *ir.ComparePattern:
		return fmt.Sprintf("%s %s %s", subject, p.Op, g.expr(p.Value))
	case *ir.RangePattern:
		return fmt.Sprintf("(%s >= %s && %s <= %s)", subject, g.expr(p.Start), subject, g.expr(p.End))
	case *ir.TuplePattern:
		parts := make([]string, 0, len(p.Elements))
		for i, elem := range p.Elements {
			parts = append(parts, g.patternCondition(fmt.Sprintf("%s.F%d", subject, i), elem))
		}
		return strings.Join(parts, " && ")
	case *ir.ConstructorPattern:
		switch p.Name {
		case "Ok":
			return subject + ".ok"
		case "Err":
			return "!" + subject + ".ok"
		default:
			return "false"
		}
	case *ir.MapPattern:
		return g.mapPatternCondition(subject, p)
	case *ir.ObjectPattern:
		return g.objectPatternCondition(subject, p)
	default:
		return "true"
	}
}

func (g *generator) mapPatternCondition(subject string, pattern *ir.MapPattern) string {
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
			parts = append(parts, fmt.Sprintf("func() bool { %s, %s := %s[%s]; return !%s || (%s) }()", target, ok, subject, key, ok, condition))
			continue
		}
		parts = append(parts, fmt.Sprintf("func() bool { %s, %s := %s[%s]; return %s && (%s) }()", target, ok, subject, key, ok, condition))
	}
	if len(parts) == 0 {
		return "true"
	}
	return strings.Join(parts, " && ")
}

func (g *generator) objectPatternCondition(subject string, pattern *ir.ObjectPattern) string {
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

func (g *generator) signalRuntime() {
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
	g.line("for _, watcher := range s.watchers {")
	g.indent++
	g.line("watcher(old, value)")
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
