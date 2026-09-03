package moonbitcodegen

import (
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
)

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
