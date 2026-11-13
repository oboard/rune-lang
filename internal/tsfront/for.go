package tsfront

import "github.com/oboard/rune-lang/internal/ast"

// forStatement lowers a TypeScript classic `for (init; cond; update)` into a
// Rune block containing the init statement and a while loop. `continue` is
// rewritten to run the update expression before jumping, which preserves
// TypeScript's post-iteration semantics with the existing while backend.
func (c *converter) forStatement(node map[string]any) (ast.Stmt, bool) {
	init, ok := c.forInit(node)
	if !ok {
		return nil, false
	}
	cond := ast.Expr(&ast.BoolLiteral{Value: true})
	if test, ok := mapField(node, "test"); ok {
		converted, ok := c.expression(test)
		if !ok {
			return nil, false
		}
		cond = converted
	}
	var update ast.Expr
	if raw, ok := mapField(node, "update"); ok {
		converted, ok := c.expression(raw)
		if !ok {
			return nil, false
		}
		update = converted
	}
	bodyNode, ok := mapField(node, "body")
	if !ok {
		return nil, false
	}
	body, ok := c.block(bodyNode)
	if !ok {
		return nil, false
	}
	if update != nil {
		body = rewriteBlockForContinues(body, update)
		body.Statements = append(body.Statements, &ast.ExprStmt{Expr: update})
	}
	statements := make([]ast.Stmt, 0, 2)
	if init != nil {
		statements = append(statements, init)
	}
	statements = append(statements, &ast.ExprStmt{Expr: &ast.WhileExpr{Cond: cond, Body: body}})
	return &ast.ExprStmt{Expr: &ast.BlockExpr{Statements: statements}}, true
}

func (c *converter) forInit(node map[string]any) (ast.Stmt, bool) {
	raw, ok := mapField(node, "init")
	if !ok {
		return nil, true
	}
	switch stringField(raw, "type") {
	case "VariableDeclaration", "ExpressionStatement":
		return c.statement(raw)
	default:
		expr, ok := c.expression(raw)
		if !ok {
			return nil, false
		}
		return &ast.ExprStmt{Expr: expr}, true
	}
}

func rewriteBlockForContinues(block *ast.BlockExpr, update ast.Expr) *ast.BlockExpr {
	out := &ast.BlockExpr{Pos: block.Pos}
	for _, stmt := range block.Statements {
		out.Statements = append(out.Statements, rewriteStmtForContinues(stmt, update)...)
	}
	return out
}

func rewriteStmtForContinues(stmt ast.Stmt, update ast.Expr) []ast.Stmt {
	switch s := stmt.(type) {
	case *ast.ContinueStmt:
		return []ast.Stmt{
			&ast.ExprStmt{Expr: update},
			&ast.ContinueStmt{Pos: s.Pos},
		}
	case *ast.TryStmt:
		out := *s
		if s.Body != nil {
			out.Body = rewriteBlockForContinues(s.Body, update)
		}
		if s.Catch != nil {
			out.Catch = rewriteBlockForContinues(s.Catch, update)
		}
		if s.Finally != nil {
			out.Finally = rewriteBlockForContinues(s.Finally, update)
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
			return []ast.Stmt{&ast.ExprStmt{Expr: rewriteBlockForContinues(block, update)}}
		}
		if ifExpr, ok := s.Expr.(*ast.IfExpr); ok {
			out := *ifExpr
			out.Then = rewriteExprForContinues(ifExpr.Then, update)
			out.Else = rewriteExprForContinues(ifExpr.Else, update)
			return []ast.Stmt{&ast.ExprStmt{Expr: &out}}
		}
	}
	return []ast.Stmt{stmt}
}

func rewriteExprForContinues(expr ast.Expr, update ast.Expr) ast.Expr {
	switch e := expr.(type) {
	case *ast.BlockExpr:
		return rewriteBlockForContinues(e, update)
	case *ast.IfExpr:
		out := *e
		out.Then = rewriteExprForContinues(e.Then, update)
		out.Else = rewriteExprForContinues(e.Else, update)
		return &out
	default:
		return expr
	}
}
