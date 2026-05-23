package parser

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

func (p *Parser) parseStatement() ast.Stmt {
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
			return &ast.LetStmt{
				Name:    name.Lexeme,
				Mutable: mutable,
				Signal:  signal,
				Value:   p.parseExpression(1),
				Pos:     name.Pos,
			}
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
