package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

func (p *Parser) parseXMLElement() ast.Expr {
	start := p.consume(lexer.Less, "expected '<'")
	if p.match(lexer.Slash) {
		p.errorAt(start, "unexpected XML closing tag")
		return &ast.Identifier{Name: "<error>", Pos: start.Pos}
	}
	name := p.consume(lexer.Ident, "expected XML tag name")
	elem := &ast.XMLElement{Tag: name.Lexeme, Pos: start.Pos}

	selfClosing := p.parseXMLAttributes(elem)
	if selfClosing {
		return elem
	}

	for !p.check(lexer.EOF) {
		if p.check(lexer.XMLText) {
			text := normalizeXMLText(p.advance().Lexeme)
			if text != "" {
				elem.Children = append(elem.Children, ast.XMLChild{Text: text, Pos: p.previous().Pos})
			}
			continue
		}
		if p.check(lexer.LBrace) {
			open := p.advance()
			p.skipNewlines()
			expr := p.parseExpression(1)
			p.skipNewlines()
			p.consume(lexer.RBrace, "expected '}' after XML expression")
			elem.Children = append(elem.Children, ast.XMLChild{Expr: expr, Pos: open.Pos})
			continue
		}
		if p.check(lexer.Less) {
			if p.checkNext(lexer.Slash) {
				p.parseXMLClose(elem.Tag)
				return elem
			}
			child := p.parseXMLElement()
			elem.Children = append(elem.Children, ast.XMLChild{Expr: child, Pos: child.Position()})
			continue
		}
		p.errorAt(p.peek(), fmt.Sprintf("expected XML child, got %s", p.peek().Kind))
		p.advance()
	}
	p.errorAt(start, "unterminated XML element")
	return elem
}

func (p *Parser) parseXMLAttributes(elem *ast.XMLElement) bool {
	for !p.check(lexer.Greater) && !p.check(lexer.EOF) {
		if p.match(lexer.Slash) {
			p.consume(lexer.Greater, "expected '>' after '/' in XML tag")
			return true
		}

		event := p.match(lexer.At)
		name := p.consume(lexer.Ident, "expected XML attribute name")
		attr := ast.XMLAttr{Name: name.Lexeme, Event: event, Pos: name.Pos}
		if p.match(lexer.Assign) {
			attr.Value = p.parseXMLAttributeValue()
		}
		elem.Attrs = append(elem.Attrs, attr)
	}
	p.consume(lexer.Greater, "expected '>' after XML tag")
	return false
}

func (p *Parser) parseXMLAttributeValue() ast.Expr {
	if p.check(lexer.String) {
		tok := p.advance()
		value, err := strconv.Unquote(tok.Lexeme)
		if err != nil {
			p.errorAt(tok, "invalid XML attribute string")
			value = tok.Lexeme
		}
		return &ast.StringLiteral{Value: value, Pos: tok.Pos}
	}
	if p.match(lexer.LBrace) {
		open := p.previous()
		p.skipNewlines()
		expr := p.parseExpression(1)
		p.skipNewlines()
		p.consume(lexer.RBrace, "expected '}' after XML attribute expression")
		if expr == nil {
			return &ast.Identifier{Name: "<error>", Pos: open.Pos}
		}
		return expr
	}
	tok := p.consume(lexer.Ident, "expected XML attribute value")
	return &ast.StringLiteral{Value: tok.Lexeme, Pos: tok.Pos}
}

func (p *Parser) parseXMLClose(tag string) {
	start := p.consume(lexer.Less, "expected '</'")
	p.consume(lexer.Slash, "expected '/' in XML closing tag")
	name := p.consume(lexer.Ident, "expected XML closing tag name")
	if name.Lexeme != tag {
		p.errorAt(name, fmt.Sprintf("mismatched XML closing tag %q, expected %q", name.Lexeme, tag))
	}
	p.consume(lexer.Greater, "expected '>' after XML closing tag")
	_ = start
}

func normalizeXMLText(text string) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	containsNewline := strings.Contains(text, "\n") || strings.Contains(text, "\r")
	out := strings.Join(strings.Fields(text), " ")
	if !containsNewline && len(text) > 0 {
		if isXMLSpace(rune(text[0])) {
			out = " " + out
		}
		if isXMLSpace(rune(text[len(text)-1])) {
			out += " "
		}
	}
	return out
}

func isXMLSpace(ch rune) bool {
	switch ch {
	case ' ', '\t', '\n', '\r':
		return true
	default:
		return false
	}
}
