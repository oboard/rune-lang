package tsfront

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

// doWhileStatement lowers `do { body } while (cond)` to a `* (true)` loop
// with an explicit cond check at the end of the body. A `continue` inside the
// body is rewritten to run the same check before jumping, so each iteration
// evaluates the condition exactly once.
func (c *converter) doWhileStatement(node map[string]any) (ast.Stmt, bool) {
	bodyNode, ok := mapField(node, "body")
	if !ok {
		return nil, false
	}
	body, ok := c.block(bodyNode)
	if !ok {
		return nil, false
	}
	test, ok := mapField(node, "test")
	if !ok {
		return nil, false
	}
	continueCond, ok := c.expression(test)
	if !ok {
		return nil, false
	}
	finalCond, ok := c.expression(test)
	if !ok {
		return nil, false
	}
	body = rewriteBlockForContinueCond(body, continueCond)
	body.Statements = append(body.Statements, &ast.ExprStmt{Expr: &ast.IfExpr{
		Cond: negateExpr(finalCond),
		Then: &ast.BlockExpr{Statements: []ast.Stmt{&ast.BreakStmt{}}},
	}})
	return &ast.ExprStmt{Expr: &ast.WhileExpr{Cond: &ast.BoolLiteral{Value: true}, Body: body}}, true
}

func rewriteBlockForContinueCond(block *ast.BlockExpr, cond ast.Expr) *ast.BlockExpr {
	out := &ast.BlockExpr{Pos: block.Pos}
	for _, stmt := range block.Statements {
		out.Statements = append(out.Statements, rewriteStmtForContinueCond(stmt, cond)...)
	}
	return out
}

func rewriteStmtForContinueCond(stmt ast.Stmt, cond ast.Expr) []ast.Stmt {
	switch s := stmt.(type) {
	case *ast.ContinueStmt:
		return []ast.Stmt{
			&ast.ExprStmt{Expr: &ast.IfExpr{
				Cond: negateExpr(cond),
				Then: &ast.BlockExpr{Statements: []ast.Stmt{&ast.BreakStmt{}}},
			}},
			&ast.ContinueStmt{Pos: s.Pos},
		}
	case *ast.TryStmt:
		out := *s
		if s.Body != nil {
			out.Body = rewriteBlockForContinueCond(s.Body, cond)
		}
		if s.Catch != nil {
			out.Catch = rewriteBlockForContinueCond(s.Catch, cond)
		}
		if s.Finally != nil {
			out.Finally = rewriteBlockForContinueCond(s.Finally, cond)
		}
		return []ast.Stmt{&out}
	case *ast.ExprStmt:
		if _, ok := s.Expr.(*ast.WhileExpr); ok {
			return []ast.Stmt{stmt}
		}
		if _, ok := s.Expr.(*ast.ForOfExpr); ok {
			return []ast.Stmt{stmt}
		}
		if block, ok := s.Expr.(*ast.BlockExpr); ok {
			return []ast.Stmt{&ast.ExprStmt{Expr: rewriteBlockForContinueCond(block, cond)}}
		}
		if ifExpr, ok := s.Expr.(*ast.IfExpr); ok {
			out := *ifExpr
			out.Then = rewriteExprForContinueCond(ifExpr.Then, cond)
			out.Else = rewriteExprForContinueCond(ifExpr.Else, cond)
			return []ast.Stmt{&ast.ExprStmt{Expr: &out}}
		}
	}
	return []ast.Stmt{stmt}
}

func rewriteExprForContinueCond(expr ast.Expr, cond ast.Expr) ast.Expr {
	switch e := expr.(type) {
	case *ast.BlockExpr:
		return rewriteBlockForContinueCond(e, cond)
	case *ast.IfExpr:
		out := *e
		out.Then = rewriteExprForContinueCond(e.Then, cond)
		out.Else = rewriteExprForContinueCond(e.Else, cond)
		return &out
	default:
		return expr
	}
}

func negateExpr(expr ast.Expr) ast.Expr {
	return &ast.UnaryExpr{Op: lexer.Bang, Expr: expr}
}
