package parser

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

func (p *Parser) parseStatement() ast.Stmt {
	if p.check(lexer.Star) && p.checkNext(lexer.LParen) {
		star := p.peek()
		return &ast.ExprStmt{Expr: p.parseStarLoopExpr(), Pos: star.Pos}
	}
	if p.checkPair(lexer.BitXor, lexer.Greater) {
		symbol := p.advance()
		p.advance()
		p.skipNewlines()
		return &ast.ContinueStmt{Pos: symbol.Pos}
	}
	if p.check(lexer.BitXor) {
		symbol := p.advance()
		p.skipNewlines()
		return &ast.BreakStmt{Pos: symbol.Pos}
	}
	if p.checkPair(lexer.Bang, lexer.Bang) {
		start := p.advance()
		p.advance()
		p.skipNewlines()
		return &ast.ThrowStmt{Value: p.parseExpression(1), Pos: start.Pos}
	}
	if p.checkKeyword("return") {
		keyword := p.advance()
		if p.check(lexer.Newline) || p.check(lexer.RBrace) || p.check(lexer.EOF) {
			return &ast.ReturnStmt{Pos: keyword.Pos}
		}
		return &ast.ReturnStmt{Value: p.parseExpression(1), Pos: keyword.Pos}
	}
	if p.checkKeyword("break") {
		keyword := p.advance()
		return &ast.BreakStmt{Pos: keyword.Pos}
	}
	if p.checkKeyword("continue") {
		keyword := p.advance()
		return &ast.ContinueStmt{Pos: keyword.Pos}
	}
	if p.check(lexer.LBrace) && p.looksLikeObjectDestructureDecl() {
		return p.parseObjectDestructureStatement()
	}
	if p.check(lexer.LBrace) {
		start := p.peek()
		expr := p.parseBlock()
		p.skipNewlines()
		if p.checkPair(lexer.Bang, lexer.Question) || p.checkPair(lexer.Question, lexer.Bang) || p.checkPair(lexer.Minus, lexer.Minus) {
			return p.parseTryAfterBlock(start, expr)
		}
		return &ast.ExprStmt{Expr: expr, Pos: expr.Position()}
	}
	if p.check(lexer.Dollar) && p.checkNext(lexer.Ident) {
		if p.kindAt(p.curr+2) == lexer.Declare {
			p.advance()
			name := p.advance()
			p.advance()
			p.skipNewlines()
			stmt := &ast.LetStmt{
				Name:    name.Lexeme,
				Mutable: true,
				Signal:  true,
				Value:   p.parseExpression(1),
				Pos:     name.Pos,
			}
			if p.match(lexer.Colon) {
				p.skipNewlines()
				stmt.Type = p.parseTypeName()
			}
			return stmt
		}
	}
	if p.check(lexer.Ident) {
		if p.checkNext(lexer.Declare) || p.checkNext(lexer.MutDeclare) {
			name := p.advance()
			mutable := false
			signal := false
			if p.match(lexer.MutDeclare) {
				mutable = true
			} else {
				p.consume(lexer.Declare, "expected ':=' after binding name")
			}
			p.skipNewlines()
			stmt := &ast.LetStmt{
				Name:    name.Lexeme,
				Mutable: mutable,
				Signal:  signal,
				Value:   p.parseExpression(1),
				Pos:     name.Pos,
			}
			if p.match(lexer.Colon) {
				p.skipNewlines()
				stmt.Type = p.parseTypeName()
			}
			return stmt
		}
		if p.checkNext(lexer.Assign) {
			name := p.advance()
			p.advance()
			p.skipNewlines()
			return &ast.AssignStmt{
				Name:  name.Lexeme,
				Value: p.parseExpression(1),
				Pos:   name.Pos,
			}
		}
	}
	expr := p.parseExpression(1)
	return &ast.ExprStmt{Expr: expr, Pos: expr.Position()}
}

func (p *Parser) checkPair(first, second lexer.Kind) bool {
	return p.kindAt(p.curr) == first && p.kindAt(p.curr+1) == second
}

func (p *Parser) parseTryAfterBlock(start lexer.Token, expr ast.Expr) ast.Stmt {
	body, _ := expr.(*ast.BlockExpr)
	stmt := &ast.TryStmt{Body: body, Pos: start.Pos}
	if p.checkPair(lexer.Bang, lexer.Question) {
		p.advance()
		p.advance()
		p.skipNewlines()
		stmt.CatchName, stmt.Catch = p.parseCatchBindingBlock()
		p.skipNewlines()
	} else if p.checkPair(lexer.Question, lexer.Bang) {
		p.advance()
		p.advance()
		p.skipNewlines()
		stmt.Catch, _ = p.parseBlock().(*ast.BlockExpr)
		p.skipNewlines()
	}
	if p.checkPair(lexer.Minus, lexer.Minus) {
		p.advance()
		p.advance()
		p.skipNewlines()
		stmt.Finally, _ = p.parseBlock().(*ast.BlockExpr)
	}
	if stmt.Catch == nil && stmt.Finally == nil {
		p.errorAt(start, "block handler requires !? / ?! / --")
	}
	return stmt
}

func (p *Parser) parseCatchBindingBlock() (string, *ast.BlockExpr) {
	p.consume(lexer.LBrace, "expected '{' after !?")
	p.skipNewlines()
	name := p.consume(lexer.Ident, "expected catch binding name")
	p.consume(lexer.FatArrow, "expected '=>' after catch binding")
	p.skipNewlines()
	var body ast.Expr
	if p.check(lexer.LBrace) {
		body = p.parseBlock()
	} else {
		body = p.parseExpression(1)
	}
	p.skipNewlines()
	p.consume(lexer.RBrace, "expected '}' after catch handler")
	if block, ok := body.(*ast.BlockExpr); ok {
		return name.Lexeme, block
	}
	return name.Lexeme, &ast.BlockExpr{Statements: []ast.Stmt{&ast.ExprStmt{Expr: body, Pos: body.Position()}}}
}

func (p *Parser) parseObjectDestructureStatement() ast.Stmt {
	start := p.consume(lexer.LBrace, "expected '{' before object destructuring")
	stmt := &ast.ObjectDestructureStmt{Pos: start.Pos}
	p.skipNewlines()
	for !p.check(lexer.RBrace) && !p.check(lexer.EOF) {
		field := p.consume(lexer.Ident, "expected field name in object destructuring")
		name := field
		if p.match(lexer.Colon) {
			p.skipNewlines()
			name = p.consume(lexer.Ident, "expected binding name after ':'")
		}
		stmt.Fields = append(stmt.Fields, ast.ObjectBindingField{
			Field:    field.Lexeme,
			Name:     name.Lexeme,
			FieldPos: field.Pos,
			NamePos:  name.Pos,
		})
		p.skipNewlines()
		if !p.match(lexer.Comma) {
			break
		}
		p.skipNewlines()
	}
	p.consume(lexer.RBrace, "expected '}' after object destructuring")
	p.skipNewlines()
	if p.match(lexer.MutDeclare) {
		stmt.Mutable = true
	} else {
		p.consume(lexer.Declare, "expected ':=' after object destructuring")
	}
	p.skipNewlines()
	stmt.Value = p.parseExpression(1)
	return stmt
}
