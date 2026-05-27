package parser

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

func (p *Parser) parsePattern() ast.Pattern {
	pattern := p.parsePatternAtom()
	if p.match(lexer.DotDotEqual) {
		literal, ok := pattern.(*ast.LiteralPattern)
		if !ok {
			p.errorAt(p.previous(), "range pattern start must be a literal")
			return pattern
		}
		end := p.parsePrimary()
		return &ast.RangePattern{Start: literal.Value, End: end, Pos: pattern.Position()}
	}
	return pattern
}

func (p *Parser) parsePatternAtom() ast.Pattern {
	tok := p.peek()
	switch tok.Kind {
	case lexer.Underscore:
		p.advance()
		return &ast.WildcardPattern{Pos: tok.Pos}
	case lexer.Int, lexer.Double, lexer.BigInt, lexer.String, lexer.Char:
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
		if p.check(lexer.RParen) {
			p.consume(lexer.RParen, "expected ')' after tuple pattern")
			return &ast.TuplePattern{Pos: start.Pos}
		}
		first := p.parsePattern()
		p.skipNewlines()
		if !p.match(lexer.Comma) {
			p.consume(lexer.RParen, "expected ')' after pattern")
			return first
		}
		elems := []ast.Pattern{first}
		p.skipNewlines()
		for !p.check(lexer.RParen) && !p.check(lexer.EOF) {
			elems = append(elems, p.parsePattern())
			p.skipNewlines()
			if !p.match(lexer.Comma) {
				break
			}
			p.skipNewlines()
		}
		p.consume(lexer.RParen, "expected ')' after tuple pattern")
		return &ast.TuplePattern{Elements: elems, Pos: start.Pos}
	}

	p.errorAt(tok, "expected pattern")
	p.advance()
	return &ast.WildcardPattern{Pos: tok.Pos}
}
