package checker

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

func (c *checker) rewritePatternPredicateBody(fn *ast.Function, info *FuncInfo, env map[string]Type) {
	if fn == nil || info == nil || len(info.Params) == 0 {
		return
	}
	expr, pos, ok := patternPredicateSourceExpr(fn.Body)
	if !ok {
		return
	}
	patterns, ok := c.patternsFromPredicateExpr(expr)
	if !ok {
		return
	}
	if info.ReturnDeclared {
		if info.Return == Bool {
			c.bindPatternPredicateSubject(info, env, patterns)
			fn.Body = c.patternPredicateBlock(pos, patterns)
		}
		return
	}
	if c.bitOrExprSupported(expr, env) {
		return
	}
	c.bindPatternPredicateSubject(info, env, patterns)
	fn.Body = c.patternPredicateBlock(pos, patterns)
}

func patternPredicateSourceExpr(body ast.Expr) (ast.Expr, lexer.Position, bool) {
	if body == nil {
		return nil, lexer.Position{}, false
	}
	if block, ok := body.(*ast.BlockExpr); ok {
		if len(block.Statements) != 1 {
			return nil, lexer.Position{}, false
		}
		stmt, ok := block.Statements[0].(*ast.ExprStmt)
		if !ok {
			return nil, lexer.Position{}, false
		}
		return stmt.Expr, block.Pos, true
	}
	return body, body.Position(), true
}

func (c *checker) bindPatternPredicateSubject(info *FuncInfo, env map[string]Type, patterns []ast.Pattern) {
	if info == nil || len(info.Params) == 0 {
		return
	}
	subject := c.patternsSubjectType(patterns)
	if subject == Unknown {
		return
	}
	param := info.Params[0].Name
	if current := env[param]; current == Unknown || current == "" {
		env[param] = subject
	} else if current != subject {
		c.errorf(patterns[0].Position(), "pattern has type %s, expected %s", subject, current)
	}
}

func (c *checker) patternsSubjectType(patterns []ast.Pattern) Type {
	result := Unknown
	for _, pattern := range patterns {
		typ := c.patternLiteralType(pattern)
		if typ == Unknown {
			return Unknown
		}
		if result == Unknown {
			result = typ
			continue
		}
		if result != typ {
			return Unknown
		}
	}
	return result
}

func (c *checker) patternPredicateBlock(pos lexer.Position, patterns []ast.Pattern) *ast.PatternBlock {
	block := &ast.PatternBlock{Pos: pos}
	for _, pattern := range patterns {
		block.Branches = append(block.Branches, ast.PatternBranch{
			Pattern: pattern,
			Expr:    &ast.BoolLiteral{Value: true, Pos: pattern.Position()},
			Pos:     pattern.Position(),
		})
	}
	block.Branches = append(block.Branches, ast.PatternBranch{
		Pattern: &ast.WildcardPattern{Pos: pos},
		Expr:    &ast.BoolLiteral{Value: false, Pos: pos},
		Pos:     pos,
	})
	return block
}

func (c *checker) patternsFromPredicateExpr(expr ast.Expr) ([]ast.Pattern, bool) {
	if binary, ok := expr.(*ast.BinaryExpr); ok && binary.Op == lexer.BitOr {
		return c.patternsFromPredicatePart(binary)
	}
	if pattern, ok := c.rangePatternFromExpr(expr); ok {
		return []ast.Pattern{pattern}, true
	}
	return nil, false
}

func (c *checker) patternsFromPredicatePart(expr ast.Expr) ([]ast.Pattern, bool) {
	if binary, ok := expr.(*ast.BinaryExpr); ok && binary.Op == lexer.BitOr {
		left, ok := c.patternsFromPredicatePart(binary.Left)
		if !ok {
			return nil, false
		}
		right, ok := c.patternsFromPredicatePart(binary.Right)
		if !ok {
			return nil, false
		}
		return append(left, right...), true
	}
	if pattern, ok := c.rangePatternFromExpr(expr); ok {
		return []ast.Pattern{pattern}, true
	}
	pattern, ok := c.literalPatternFromExpr(expr)
	if !ok {
		return nil, false
	}
	return []ast.Pattern{pattern}, true
}

func (c *checker) rangePatternFromExpr(expr ast.Expr) (ast.Pattern, bool) {
	binary, ok := expr.(*ast.BinaryExpr)
	if !ok || binary.Op != lexer.DotDotEqual {
		return nil, false
	}
	if !patternPredicateRangeEndpoint(binary.Left) || !patternPredicateRangeEndpoint(binary.Right) {
		return nil, false
	}
	return &ast.RangePattern{Start: binary.Left, End: binary.Right, Inclusive: true, Pos: binary.Pos}, true
}

func (c *checker) literalPatternFromExpr(expr ast.Expr) (ast.Pattern, bool) {
	switch expr.(type) {
	case *ast.BoolLiteral, *ast.IntegerLiteral, *ast.DoubleLiteral, *ast.BigIntLiteral, *ast.StringLiteral, *ast.CharLiteral, *ast.NullLiteral, *ast.SelectorExpr:
		return &ast.LiteralPattern{Value: expr, Pos: expr.Position()}, true
	default:
		return nil, false
	}
}

func patternPredicateRangeEndpoint(expr ast.Expr) bool {
	switch expr.(type) {
	case *ast.BoolLiteral, *ast.IntegerLiteral, *ast.DoubleLiteral, *ast.BigIntLiteral, *ast.StringLiteral, *ast.CharLiteral, *ast.NullLiteral, *ast.SelectorExpr:
		return true
	default:
		return false
	}
}

func (c *checker) bitOrExprSupported(expr ast.Expr, env map[string]Type) bool {
	typ, ok := c.bitOrExprType(expr, env)
	return ok && isBitwiseType(typ)
}

func (c *checker) bitOrExprType(expr ast.Expr, env map[string]Type) (Type, bool) {
	if binary, ok := expr.(*ast.BinaryExpr); ok && binary.Op == lexer.BitOr {
		left, leftOK := c.bitOrExprType(binary.Left, env)
		right, rightOK := c.bitOrExprType(binary.Right, env)
		if !leftOK || !rightOK || !isBitwiseType(left) || !isBitwiseType(right) || left != right {
			return Unknown, false
		}
		return left, true
	}
	return c.bitOrLeafType(expr, env)
}

func (c *checker) bitOrLeafType(expr ast.Expr, env map[string]Type) (Type, bool) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return Int, true
	case *ast.BigIntLiteral:
		return BigInt, true
	case *ast.Identifier:
		typ, ok := env[e.Name]
		return typ, ok
	default:
		return Unknown, false
	}
}
