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

func (g *generator) function(fn *ir.Function) error {
	params := make([]string, 0, len(fn.Params))
	for _, param := range fn.Params {
		params = append(params, fmt.Sprintf("%s %s", mangleIdent(param.Name), goType(param.Type)))
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

func (g *generator) method(typ *ir.StructType, fn *ir.Function) error {
	params := make([]string, 0, len(fn.Params))
	for _, param := range fn.Params {
		params = append(params, fmt.Sprintf("%s %s", mangleIdent(param.Name), goType(param.Type)))
	}
	ret := ""
	if fn.Return != checker.Void {
		ret = " " + goType(fn.Return)
	}
	g.linef("func (%s %s) %s(%s)%s {", mangleIdent("this"), mangleIdent(typ.Name), mangleIdent(fn.Name), strings.Join(params, ", "), ret)
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

func (g *generator) body(fn *ir.Function, expr ir.Expr, ret checker.Type) error {
	switch e := expr.(type) {
	case *ir.PatternBlock:
		return g.patternBlock(fn, e, ret)
	case *ir.BlockExpr:
		return g.block(e, ret)
	default:
		if ret == checker.Void {
			if expr := g.expr(expr); expr != "" {
				g.line(expr)
			}
		} else {
			g.linef("return %s", g.expr(expr))
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
		case *ir.AssignStmt:
			if g.isSignal(s.Name) {
				g.linef("%s.Set(%s)", mangleIdent(s.Name), g.expr(s.Value))
				continue
			}
			g.linef("%s = %s", mangleIdent(s.Name), g.expr(s.Value))
		case *ir.ExprStmt:
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
				g.linef("return %s", expr)
			} else {
				g.line(expr)
			}
		}
	}
	if ret != checker.Void && len(block.Statements) == 0 {
		g.linef("return %s", zeroValue(ret))
	}
	return nil
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
		if ret == checker.Void {
			if expr := g.expr(branch.Expr); expr != "" {
				g.line(expr)
			}
		} else {
			g.linef("return %s", g.expr(branch.Expr))
		}
		g.indent--
	}
	g.indent--
	g.line("}")
	if ret != checker.Void && !hasDefault {
		g.linef("return %s", zeroValue(ret))
	}
	return nil
}

func (g *generator) patternCondition(subject string, pattern ir.Pattern) string {
	switch p := pattern.(type) {
	case *ir.LiteralPattern:
		return fmt.Sprintf("%s == %s", subject, g.expr(p.Value))
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
