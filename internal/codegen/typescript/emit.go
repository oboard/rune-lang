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
	g.linef("type %s = number;", mangleIdent(enum.Name))
	g.linef("const %s = {", mangleIdent(enum.Name))
	g.indent++
	for _, member := range enum.Members {
		g.linef("%s: %d,", tsPropertyName(member.Name), member.Value)
	}
	g.indent--
	g.line("} as const;")
}

func (g *generator) function(fn *ir.Function) error {
	params := make([]string, 0, len(fn.Params))
	for _, param := range fn.Params {
		params = append(params, fmt.Sprintf("%s: %s", mangleIdent(param.Name), tsType(param.Type)))
	}
	if fn.Routine {
		return g.routineFunction(fn, params)
	}
	prefix := "function"
	ret := tsType(fn.Return)
	g.linef("%s %s(%s): %s {", prefix, mangleIdent(fn.Name), strings.Join(params, ", "), ret)
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
	g.linef("function %s(%s): Promise<%s> {", mangleIdent(fn.Name), strings.Join(params, ", "), ret)
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

func (g *generator) method(typ *ir.StructType, fn *ir.Function) error {
	params := []string{fmt.Sprintf("%s: %s", mangleIdent("this"), mangleIdent(typ.Name))}
	for _, param := range fn.Params {
		params = append(params, fmt.Sprintf("%s: %s", mangleIdent(param.Name), tsType(param.Type)))
	}
	if fn.Routine {
		return g.routineMethod(typ, fn, params)
	}
	prefix := "function"
	ret := tsType(fn.Return)
	g.linef("%s %s(%s): %s {", prefix, mangleMethod(typ.Name, fn.Name), strings.Join(params, ", "), ret)
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
	g.line("}")
	return nil
}

func (g *generator) routineMethod(typ *ir.StructType, fn *ir.Function, params []string) error {
	ret := tsType(fn.Return)
	g.linef("function %s(%s): Promise<%s> {", mangleMethod(typ.Name, fn.Name), strings.Join(params, ", "), ret)
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
	_ = typ
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
				g.linef("const %s = runeSignal(%s);", mangleIdent(s.Name), g.expr(s.Value))
				g.addSignal(s.Name, s.Value.ResultType())
				for _, dep := range g.exprSignalDeps(s.Value) {
					g.linef("runeWatch(%s, () => { %s.set(%s); });", mangleIdent(dep), mangleIdent(s.Name), g.expr(s.Value))
				}
				continue
			}
			kind := "const"
			if s.Mutable {
				kind = "let"
			}
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
			g.linef("%s = %s;", mangleIdent(s.Name), g.expr(s.Value))
		case *ir.ExprStmt:
			if unwrap, ok := s.Expr.(*ir.ResultUnwrapExpr); ok {
				g.resultUnwrapExprStmt(unwrap, ret, last)
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
	return g.returnRawExpr(expr, ret, g.expr(expr))
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
	case *ir.LiteralPattern:
		return fmt.Sprintf("%s === %s", subject, g.expr(p.Value))
	case *ir.ComparePattern:
		return fmt.Sprintf("%s %s %s", subject, p.Op, g.expr(p.Value))
	case *ir.RangePattern:
		return fmt.Sprintf("(%s >= %s && %s <= %s)", subject, g.expr(p.Start), subject, g.expr(p.End))
	case *ir.TuplePattern:
		parts := make([]string, 0, len(p.Elements))
		for i, elem := range p.Elements {
			parts = append(parts, g.patternCondition(fmt.Sprintf("%s[%d]", subject, i), elem))
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
	default:
		return "true"
	}
}

func (g *generator) signalRuntime() {
	g.line("type RuneSignal<T> = {")
	g.indent++
	g.line("get(): T;")
	g.line("set(value: T): void;")
	g.line("watch(watcher: (oldValue: T, newValue: T) => void): void;")
	g.indent--
	g.line("};")
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
	g.line("for (const watcher of watchers) watcher(old, next);")
	g.indent--
	g.line("},")
	g.line("watch: (watcher: (oldValue: T, newValue: T) => void) => { watchers.push(watcher); },")
	g.indent--
	g.line("};")
	g.indent--
	g.line("}")
	g.line("")
	g.line("const runeReactiveWatchers = new WeakMap<object, Array<() => void>>();")
	g.line("")
	g.line("function runeWatch(source: any, watcher: () => void): void {")
	g.indent++
	g.line("if (source && typeof source.watch === \"function\") {")
	g.indent++
	g.line("source.watch(() => watcher());")
	g.line("return;")
	g.indent--
	g.line("}")
	g.line("if (source && typeof source === \"object\") {")
	g.indent++
	g.line("const watchers = runeReactiveWatchers.get(source);")
	g.line("if (watchers) watchers.push(watcher);")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("function runeNotify(source: object): void {")
	g.indent++
	g.line("const watchers = runeReactiveWatchers.get(source);")
	g.line("if (!watchers) return;")
	g.line("for (const watcher of watchers) watcher();")
	g.indent--
	g.line("}")
	g.line("")
	g.line("function runeReactiveArray<T>(initial: T[]): T[] {")
	g.indent++
	g.line("let proxy: T[];")
	g.line("const mutators = new Set([\"copyWithin\", \"fill\", \"pop\", \"push\", \"reverse\", \"shift\", \"sort\", \"splice\", \"unshift\"]);")
	g.line("proxy = new Proxy(initial, {")
	g.indent++
	g.line("get(target, prop, receiver) {")
	g.indent++
	g.line("const value = Reflect.get(target, prop, receiver);")
	g.line("if (typeof prop === \"string\" && mutators.has(prop) && typeof value === \"function\") {")
	g.indent++
	g.line("return (...args: unknown[]) => {")
	g.indent++
	g.line("const result = value.apply(target, args);")
	g.line("runeNotify(proxy);")
	g.line("return result;")
	g.indent--
	g.line("};")
	g.indent--
	g.line("}")
	g.line("return value;")
	g.indent--
	g.line("},")
	g.line("set(target, prop, value, receiver) {")
	g.indent++
	g.line("const old = Reflect.get(target, prop, receiver);")
	g.line("const ok = Reflect.set(target, prop, value, receiver);")
	g.line("if (!Object.is(old, value)) runeNotify(proxy);")
	g.line("return ok;")
	g.indent--
	g.line("},")
	g.indent--
	g.line("});")
	g.line("runeReactiveWatchers.set(proxy, []);")
	g.line("return proxy;")
	g.indent--
	g.line("}")
	g.line("")
	g.line("function runeReactiveObject<T extends object>(initial: T): T {")
	g.indent++
	g.line("let proxy: T;")
	g.line("proxy = new Proxy(initial, {")
	g.indent++
	g.line("set(target, prop, value, receiver) {")
	g.indent++
	g.line("const old = Reflect.get(target, prop, receiver);")
	g.line("const ok = Reflect.set(target, prop, value, receiver);")
	g.line("if (!Object.is(old, value)) runeNotify(proxy as object);")
	g.line("return ok;")
	g.indent--
	g.line("},")
	g.line("deleteProperty(target, prop) {")
	g.indent++
	g.line("const ok = Reflect.deleteProperty(target, prop);")
	g.line("if (ok) runeNotify(proxy as object);")
	g.line("return ok;")
	g.indent--
	g.line("},")
	g.indent--
	g.line("});")
	g.line("runeReactiveWatchers.set(proxy as object, []);")
	g.line("return proxy;")
	g.indent--
	g.line("}")
}
