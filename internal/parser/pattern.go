package parser

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

func (p *Parser) parsePattern() ast.Pattern {
	tok := p.peek()
	switch tok.Kind {
	case lexer.Underscore:
		p.advance()
		return &ast.WildcardPattern{Pos: tok.Pos}
	case lexer.Int, lexer.Double, lexer.BigInt, lexer.String:
		return &ast.LiteralPattern{Value: p.parsePrimary(), Pos: tok.Pos}
	case lexer.Ident:
		if p.checkNext(lexer.LParen) {
			name := p.advance()
			p.consume(lexer.LParen, "expected '(' after pattern constructor")
			p.skipNewlines()
			binding := ""
			bindingPos := name.Pos
			if p.match(lexer.Ident) {
				binding = p.previous().Lexeme
				bindingPos = p.previous().Pos
			} else if p.match(lexer.Underscore) {
				bindingPos = p.previous().Pos
			} else {
				p.errorAt(p.peek(), "expected binding name or '_' in constructor pattern")
				p.advance()
			}
			p.skipNewlines()
			p.consume(lexer.RParen, "expected ')' after constructor pattern")
			return &ast.ConstructorPattern{Name: name.Lexeme, Binding: binding, BindingPos: bindingPos, Pos: name.Pos}
		}
		if p.checkNext(lexer.Dot) {
			value := p.parseExpression(1)
			return &ast.LiteralPattern{Value: value, Pos: tok.Pos}
		}
		if isLiteralIdentifier(tok.Lexeme) {
			return &ast.LiteralPattern{Value: p.parsePrimary(), Pos: tok.Pos}
		}
	case lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
		op := p.advance()
		value := p.parsePrimary()
		return &ast.ComparePattern{Op: op.Kind, Value: value, Pos: op.Pos}
	case lexer.LParen:
		start := p.advance()
		p.skipNewlines()
		var elems []ast.Pattern
		if !p.check(lexer.RParen) {
			for {
				elems = append(elems, p.parsePattern())
				p.skipNewlines()
				if !p.match(lexer.Comma) {
					break
				}
				p.skipNewlines()
			}
		}
		p.consume(lexer.RParen, "expected ')' after tuple pattern")
		return &ast.TuplePattern{Elements: elems, Pos: start.Pos}
	}

	p.errorAt(tok, "expected pattern")
	p.advance()
	return &ast.WildcardPattern{Pos: tok.Pos}
}
