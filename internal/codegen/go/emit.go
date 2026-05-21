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
	if err := g.body(fn, fn.Body, fn.Return); err != nil {
		return err
	}
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
	if err := g.body(fn, fn.Body, fn.Return); err != nil {
		return err
	}
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
			g.linef("%s := %s", mangleIdent(s.Name), g.expr(s.Value))
		case *ir.AssignStmt:
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
