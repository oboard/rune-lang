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
		g.linef("%s: %s;", mangleIdent(field.Name), tsType(field.Type))
	}
	g.indent--
	g.line("};")
}

func (g *generator) function(fn *ir.Function) error {
	params := make([]string, 0, len(fn.Params))
	for _, param := range fn.Params {
		params = append(params, fmt.Sprintf("%s: %s", mangleIdent(param.Name), tsType(param.Type)))
	}
	g.linef("function %s(%s): %s {", mangleIdent(fn.Name), strings.Join(params, ", "), tsType(fn.Return))
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

func (g *generator) method(typ *ir.StructType, fn *ir.Function) error {
	params := []string{fmt.Sprintf("%s: %s", mangleIdent("this"), mangleIdent(typ.Name))}
	for _, param := range fn.Params {
		params = append(params, fmt.Sprintf("%s: %s", mangleIdent(param.Name), tsType(param.Type)))
	}
	g.linef("function %s(%s): %s {", mangleMethod(typ.Name, fn.Name), strings.Join(params, ", "), tsType(fn.Return))
	g.indent++
	g.pushSignalScope()
	g.thisNames = append(g.thisNames, mangleIdent("this"))
	if err := g.body(fn, fn.Body, fn.Return); err != nil {
		return err
	}
	g.thisNames = g.thisNames[:len(g.thisNames)-1]
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
		if ret == checker.Void {
			if expr := g.expr(expr); expr != "" {
				g.line(expr + ";")
			}
		} else {
			g.lineExpr("return ", g.expr(expr), ";")
		}
	}
	return nil
}

func (g *generator) block(block *ir.BlockExpr, ret checker.Type) error {
	for i, stmt := range block.Statements {
		last := i == len(block.Statements)-1
		switch s := stmt.(type) {
		case *ir.LetStmt:
			if s.Signal || g.exprUsesSignal(s.Value) {
				g.linef("const %s = runeSignal(%s);", mangleIdent(s.Name), g.expr(s.Value))
				g.addSignal(s.Name, s.Value.ResultType())
				for _, dep := range g.exprSignalDeps(s.Value) {
					g.linef("%s.watch(() => { %s.set(%s); });", mangleIdent(dep), mangleIdent(s.Name), g.expr(s.Value))
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
		case *ir.AssignStmt:
			if g.isSignal(s.Name) {
				g.linef("%s.set(%s);", mangleIdent(s.Name), g.expr(s.Value))
				continue
			}
			g.linef("%s = %s;", mangleIdent(s.Name), g.expr(s.Value))
		case *ir.ExprStmt:
			expr := g.expr(s.Expr)
			if expr == "" {
				continue
			}
			if last && ret != checker.Void {
				g.lineExpr("return ", expr, ";")
			} else {
				g.line(g.stmtExpr(s.Expr) + ";")
			}
		}
	}
	if ret != checker.Void && len(block.Statements) == 0 {
		g.linef("return %s;", zeroValue(ret))
	}
	return nil
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
			g.linef("return %s;", g.expr(branch.Expr))
		}
		g.indent--
		g.line("}")
	}
	if ret != checker.Void && !hasDefault {
		g.linef("return %s;", zeroValue(ret))
	}
	return nil
}

func (g *generator) patternCondition(subject string, pattern ir.Pattern) string {
	switch p := pattern.(type) {
	case *ir.LiteralPattern:
		return fmt.Sprintf("%s === %s", subject, g.expr(p.Value))
	case *ir.ComparePattern:
		return fmt.Sprintf("%s %s %s", subject, p.Op, g.expr(p.Value))
	case *ir.TuplePattern:
		parts := make([]string, 0, len(p.Elements))
		for i, elem := range p.Elements {
			parts = append(parts, g.patternCondition(fmt.Sprintf("%s[%d]", subject, i), elem))
		}
		return strings.Join(parts, " && ")
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
}
