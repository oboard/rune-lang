package parser

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

func (p *Parser) parseStatement() ast.Stmt {
	if p.check(lexer.LBrace) && p.looksLikeObjectDestructureDecl() {
		return p.parseObjectDestructureStatement()
	}
	if p.check(lexer.LBrace) {
		expr := p.parseBlock()
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
		if p.kindAt(p.curr+2) == lexer.Assign {
			p.advance()
			name := p.advance()
			p.advance()
			p.skipNewlines()
			return &ast.AssignStmt{
				Name:         name.Lexeme,
				SignalPrefix: true,
				Value:        p.parseExpression(1),
				Pos:          name.Pos,
			}
		}
	}
	if p.check(lexer.Ident) {
		if p.checkNext(lexer.Declare) || p.checkNext(lexer.MutDeclare) || p.checkNext(lexer.SignalDeclare) {
			name := p.advance()
			mutable := false
			signal := false
			if p.match(lexer.MutDeclare) {
				mutable = true
			} else if p.match(lexer.SignalDeclare) {
				signal = true
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
	} else if p.match(lexer.SignalDeclare) {
		stmt.Signal = true
		stmt.Mutable = true
	} else {
		p.consume(lexer.Declare, "expected ':=' after object destructuring")
	}
	p.skipNewlines()
	stmt.Value = p.parseExpression(1)
	return stmt
}
