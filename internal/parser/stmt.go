package parser

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

func (p *Parser) parseStatement() ast.Stmt {
	if p.check(lexer.Ident) {
		if p.checkNext(lexer.Declare) || p.checkNext(lexer.MutDeclare) {
			name := p.advance()
			mutable := p.match(lexer.MutDeclare)
			if !mutable {
				p.consume(lexer.Declare, "expected ':=' after binding name")
			}
			p.skipNewlines()
			return &ast.LetStmt{
				Name:    name.Lexeme,
				Mutable: mutable,
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
