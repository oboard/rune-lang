package parser

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

func (p *Parser) parsePattern() ast.Pattern {
	pattern := p.parseAliasPattern()
	if !p.match(lexer.BitOr) {
		return pattern
	}
	out := &ast.OrPattern{Alternatives: []ast.Pattern{pattern}, Pos: pattern.Position()}
	for {
		p.skipNewlines()
		out.Alternatives = append(out.Alternatives, p.parseAliasPattern())
		p.skipNewlines()
		if !p.match(lexer.BitOr) {
			break
		}
	}
	return out
}

func (p *Parser) parseAliasPattern() ast.Pattern {
	pattern := p.parseRangePattern()
	if p.match(lexer.At) {
		name := p.consume(lexer.Ident, "expected binding name after '@'")
		return &ast.AliasPattern{Pattern: pattern, Name: name.Lexeme, NamePos: name.Pos, Pos: pattern.Position()}
	}
	return pattern
}

func (p *Parser) parseRangePattern() ast.Pattern {
	pattern := p.parsePatternAtom()
	if p.match(lexer.DotDotEqual) {
		start, ok := p.rangeBoundFromPattern(pattern)
		if !ok {
			p.errorAt(p.previous(), "range pattern start must be a literal or '_'")
			return pattern
		}
		end := p.parseRangeBound()
		return &ast.RangePattern{Start: start, End: end, Inclusive: true, Pos: pattern.Position()}
	}
	if p.match(lexer.DotDotLess) || p.match(lexer.DotDot) {
		if p.previous().Kind == lexer.DotDot && !p.match(lexer.Less) {
			p.errorAt(p.previous(), "expected '<' after '..' in range pattern")
			return pattern
		}
		start, ok := p.rangeBoundFromPattern(pattern)
		if !ok {
			p.errorAt(p.previous(), "range pattern start must be a literal or '_'")
			return pattern
		}
		end := p.parseRangeBound()
		return &ast.RangePattern{Start: start, End: end, Inclusive: false, Pos: pattern.Position()}
	}
	return pattern
}

func (p *Parser) rangeBoundFromPattern(pattern ast.Pattern) (ast.Expr, bool) {
	switch b := pattern.(type) {
	case *ast.LiteralPattern:
		return b.Value, true
	case *ast.WildcardPattern:
		return nil, true
	case *ast.BindingPattern:
		return &ast.Identifier{Name: b.Name, Pos: b.Pos}, true
	default:
		return nil, false
	}
}

func (p *Parser) parseRangeBound() ast.Expr {
	if p.match(lexer.Underscore) {
		return nil
	}
	return p.parsePrimary()
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
			return p.parseConstructorPattern()
		}
		if p.checkNext(lexer.Dot) {
			value := p.parseExpression(1)
			return &ast.LiteralPattern{Value: value, Pos: tok.Pos}
		}
		if isLiteralIdentifier(tok.Lexeme) {
			return &ast.LiteralPattern{Value: p.parsePrimary(), Pos: tok.Pos}
		}
		p.advance()
		return &ast.BindingPattern{Name: tok.Lexeme, Pos: tok.Pos}
	case lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
		op := p.advance()
		value := p.parsePrimary()
		return &ast.ComparePattern{Op: op.Kind, Value: value, Pos: op.Pos}
	case lexer.LBrace:
		return p.parseMapOrObjectPattern()
	case lexer.LBracket:
		return p.parseArrayPattern()
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

func (p *Parser) parseConstructorPattern() ast.Pattern {
	name := p.advance()
	p.consume(lexer.LParen, "expected '(' after pattern constructor")
	pattern := &ast.ConstructorPattern{Name: name.Lexeme, BindingPos: name.Pos, Pos: name.Pos}
	p.skipNewlines()
	for !p.check(lexer.RParen) && !p.check(lexer.EOF) {
		if p.match(lexer.DotDot) {
			pattern.Rest = true
			pattern.RestPos = p.previous().Pos
			p.skipNewlines()
			p.match(lexer.Comma)
			p.skipNewlines()
			break
		}
		arg := p.parsePattern()
		pattern.Args = append(pattern.Args, arg)
		if binding, ok := arg.(*ast.BindingPattern); ok && len(pattern.Args) == 1 {
			pattern.Binding = binding.Name
			pattern.BindingPos = binding.Pos
		} else if wildcard, ok := arg.(*ast.WildcardPattern); ok && len(pattern.Args) == 1 {
			pattern.BindingPos = wildcard.Pos
		}
		p.skipNewlines()
		if !p.match(lexer.Comma) {
			break
		}
		p.skipNewlines()
	}
	p.consume(lexer.RParen, "expected ')' after constructor pattern")
	return pattern
}

func (p *Parser) parseArrayPattern() ast.Pattern {
	start := p.consume(lexer.LBracket, "expected '[' before array pattern")
	pattern := &ast.ArrayPattern{RestIndex: -1, Pos: start.Pos}
	p.skipNewlines()
	for !p.check(lexer.RBracket) && !p.check(lexer.EOF) {
		if p.match(lexer.DotDot) {
			if p.check(lexer.String) {
				spread := p.parsePrimary()
				pattern.Elements = append(pattern.Elements, &ast.SequenceSpreadPattern{Value: spread, Pos: spread.Position()})
				p.skipNewlines()
				if !p.match(lexer.Comma) {
					break
				}
				p.skipNewlines()
				continue
			}
			if p.check(lexer.Ident) && isPatternSpreadIdentifier(p.peek().Lexeme) {
				name := p.advance()
				pattern.Elements = append(pattern.Elements, &ast.SequenceSpreadPattern{
					Value: &ast.Identifier{Name: name.Lexeme, Pos: name.Pos},
					Pos:   name.Pos,
				})
				p.skipNewlines()
				if !p.match(lexer.Comma) {
					break
				}
				p.skipNewlines()
				continue
			}
			if pattern.RestIndex >= 0 {
				p.errorAt(p.previous(), "array pattern can contain at most one '..'")
			}
			pattern.RestIndex = len(pattern.Elements)
			pattern.RestPos = p.previous().Pos
			if p.match(lexer.Ident) {
				pattern.RestBinding = p.previous().Lexeme
				pattern.RestPos = p.previous().Pos
			}
		} else {
			if bit, ok := p.tryParseBitPattern(); ok {
				pattern.Elements = append(pattern.Elements, bit)
				p.skipNewlines()
				if !p.match(lexer.Comma) {
					break
				}
				p.skipNewlines()
				continue
			}
			pattern.Elements = append(pattern.Elements, p.parsePattern())
		}
		p.skipNewlines()
		if !p.match(lexer.Comma) {
			break
		}
		p.skipNewlines()
	}
	p.consume(lexer.RBracket, "expected ']' after array pattern")
	return pattern
}

func isPatternSpreadIdentifier(name string) bool {
	ch, _ := utf8.DecodeRuneInString(name)
	return ch != utf8.RuneError && unicode.IsUpper(ch)
}

func (p *Parser) tryParseBitPattern() (ast.Pattern, bool) {
	if !p.check(lexer.Ident) || !p.checkNext(lexer.LParen) {
		return nil, false
	}
	tok := p.peek()
	signed, width, endian, ok := parseBitPatternName(tok.Lexeme)
	if !ok {
		return nil, false
	}
	p.advance()
	p.consume(lexer.LParen, "expected '(' after bit pattern")
	p.skipNewlines()
	value := p.parsePattern()
	p.skipNewlines()
	p.consume(lexer.RParen, "expected ')' after bit pattern")
	return &ast.BitPattern{Width: width, Signed: signed, Endian: endian, Value: value, Pos: tok.Pos}, true
}

func parseBitPatternName(name string) (bool, int, string, bool) {
	if len(name) < 4 {
		return false, 0, "", false
	}
	signed := name[0] == 'i'
	if !signed && name[0] != 'u' {
		return false, 0, "", false
	}
	endian := ""
	switch {
	case strings.HasSuffix(name, "be"):
		endian = "be"
	case strings.HasSuffix(name, "le"):
		endian = "le"
	default:
		return false, 0, "", false
	}
	widthText := name[1 : len(name)-len(endian)]
	width, err := strconv.Atoi(widthText)
	if err != nil {
		return false, 0, "", false
	}
	return signed, width, endian, true
}

func (p *Parser) parseMapOrObjectPattern() ast.Pattern {
	start := p.consume(lexer.LBrace, "expected '{' before pattern")
	p.skipNewlines()
	if p.check(lexer.RBrace) {
		p.consume(lexer.RBrace, "expected '}' after pattern")
		return &ast.ObjectPattern{Pos: start.Pos}
	}
	if p.check(lexer.Ident) && p.patternLooksLikeObjectField() {
		return p.parseObjectPattern(start.Pos)
	}
	return p.parseMapPattern(start.Pos)
}

func (p *Parser) parseObjectPattern(pos lexer.Position) ast.Pattern {
	pattern := &ast.ObjectPattern{Pos: pos}
	for !p.check(lexer.RBrace) && !p.check(lexer.EOF) {
		if p.match(lexer.DotDot) {
			pattern.Rest = true
			p.skipNewlines()
			p.match(lexer.Comma)
			p.skipNewlines()
			break
		}
		field := p.consume(lexer.Ident, "expected object pattern field")
		optional := p.match(lexer.Question)
		var value ast.Pattern
		if p.match(lexer.Colon) {
			p.skipNewlines()
			value = p.parsePattern()
		} else {
			value = &ast.BindingPattern{Name: field.Lexeme, Pos: field.Pos}
		}
		pattern.Fields = append(pattern.Fields, ast.ObjectPatternField{
			Name:     field.Lexeme,
			Pattern:  value,
			Optional: optional,
			Pos:      field.Pos,
		})
		p.skipNewlines()
		if !p.match(lexer.Comma) {
			break
		}
		p.skipNewlines()
	}
	p.consume(lexer.RBrace, "expected '}' after object pattern")
	return pattern
}

func (p *Parser) parseMapPattern(pos lexer.Position) ast.Pattern {
	pattern := &ast.MapPattern{Pos: pos}
	for !p.check(lexer.RBrace) && !p.check(lexer.EOF) {
		if p.match(lexer.DotDot) {
			pattern.Rest = true
			p.skipNewlines()
			p.match(lexer.Comma)
			p.skipNewlines()
			break
		}
		key := p.parsePrimary()
		optional := p.match(lexer.Question)
		p.consume(lexer.Colon, "expected ':' after map pattern key")
		p.skipNewlines()
		value := p.parsePattern()
		pattern.Entries = append(pattern.Entries, ast.MapPatternEntry{
			Key:      key,
			Pattern:  value,
			Optional: optional,
			Pos:      key.Position(),
		})
		p.skipNewlines()
		if !p.match(lexer.Comma) {
			break
		}
		p.skipNewlines()
	}
	p.consume(lexer.RBrace, "expected '}' after map pattern")
	return pattern
}

func (p *Parser) patternLooksLikeObjectField() bool {
	if p.check(lexer.Ident) && isPatternSpreadIdentifier(p.peek().Lexeme) {
		return false
	}
	if p.curr+1 >= len(p.tokens) {
		return false
	}
	switch p.tokens[p.curr+1].Kind {
	case lexer.Colon, lexer.Question, lexer.Comma, lexer.RBrace:
		return true
	default:
		return false
	}
}
