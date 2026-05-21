package parser

import (
	"fmt"
	"strconv"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

func (p *Parser) parseExpression(minPrec int) ast.Expr {
	if minPrec <= 1 && p.check(lexer.Ident) && p.checkNext(lexer.FatArrow) {
		return p.parseLambda()
	}
	left := p.parseUnary()
	for {
		if p.check(lexer.LBrace) {
			ident, ok := left.(*ast.Identifier)
			if !ok {
				break
			}
			left = p.parseStructLiteral(ident)
			continue
		}
		if p.match(lexer.LParen) {
			call := &ast.CallExpr{Callee: left, Pos: left.Position()}
			p.skipNewlines()
			if !p.check(lexer.RParen) {
				for {
					call.Args = append(call.Args, p.parseExpression(1))
					p.skipNewlines()
					if !p.match(lexer.Comma) {
						break
					}
					p.skipNewlines()
				}
			}
			p.consume(lexer.RParen, "expected ')' after arguments")
			left = call
			continue
		}
		if p.match(lexer.LBracket) {
			index := p.parseExpression(1)
			p.consume(lexer.RBracket, "expected ']' after index")
			left = &ast.IndexExpr{Receiver: left, Index: index, Pos: left.Position()}
			continue
		}
		if p.match(lexer.Dot) {
			name := p.consume(lexer.Ident, "expected selector name after '.'")
			left = &ast.SelectorExpr{Receiver: left, Name: name.Lexeme, Pos: left.Position()}
			continue
		}

		prec := precedence(p.peek().Kind)
		if prec < minPrec {
			break
		}
		op := p.advance()
		right := p.parseExpression(prec + 1)
		left = &ast.BinaryExpr{Left: left, Op: op.Kind, Right: right, Pos: left.Position()}
	}
	return left
}

func (p *Parser) parseLambda() ast.Expr {
	param := p.consume(lexer.Ident, "expected lambda parameter")
	p.consume(lexer.FatArrow, "expected '=>' after lambda parameter")
	p.skipNewlines()
	return &ast.LambdaExpr{
		Params: []string{param.Lexeme},
		Body:   p.parseBody(),
		Pos:    param.Pos,
	}
}

func (p *Parser) parseStructLiteral(typeName *ast.Identifier) ast.Expr {
	lit := &ast.StructLiteral{TypeName: typeName.Name, Pos: typeName.Pos}
	p.consume(lexer.LBrace, "expected '{' after type name")
	p.skipNewlines()
	for !p.check(lexer.RBrace) && !p.check(lexer.EOF) {
		fieldName := p.consume(lexer.Ident, "expected field name")
		p.consume(lexer.Colon, "expected ':' after field name")
		p.skipNewlines()
		value := p.parseExpression(1)
		lit.Fields = append(lit.Fields, ast.FieldValue{
			Name:  fieldName.Lexeme,
			Value: value,
			Pos:   fieldName.Pos,
		})
		p.consumeStatementEnd()
		p.match(lexer.Comma)
		p.skipNewlines()
	}
	p.consume(lexer.RBrace, "expected '}' after struct literal")
	return lit
}

func (p *Parser) parseUnary() ast.Expr {
	if p.match(lexer.Minus, lexer.Bang) {
		op := p.previous()
		return &ast.UnaryExpr{Op: op.Kind, Expr: p.parseUnary(), Pos: op.Pos}
	}
	return p.parsePrimary()
}

func (p *Parser) parsePrimary() ast.Expr {
	tok := p.peek()
	switch tok.Kind {
	case lexer.Int:
		p.advance()
		value, err := strconv.Atoi(tok.Lexeme)
		if err != nil {
			p.errorAt(tok, "invalid integer literal")
		}
		return &ast.IntegerLiteral{Value: value, Pos: tok.Pos}
	case lexer.String:
		p.advance()
		value, err := strconv.Unquote(tok.Lexeme)
		if err != nil {
			p.errorAt(tok, "invalid string literal")
			value = tok.Lexeme
		}
		return &ast.StringLiteral{Value: value, Pos: tok.Pos}
	case lexer.Ident:
		p.advance()
		if tok.Lexeme == "true" {
			return &ast.BoolLiteral{Value: true, Pos: tok.Pos}
		}
		if tok.Lexeme == "false" {
			return &ast.BoolLiteral{Value: false, Pos: tok.Pos}
		}
		return &ast.Identifier{Name: tok.Lexeme, Pos: tok.Pos}
	case lexer.At:
		at := p.advance()
		name := p.consume(lexer.Ident, "expected module name after '@'")
		return &ast.AtExpr{Name: name.Lexeme, Pos: at.Pos}
	case lexer.Dot:
		dot := p.advance()
		name := p.consume(lexer.Ident, "expected field name after '.'")
		return &ast.SelectorExpr{
			Receiver: &ast.ThisExpr{Pos: dot.Pos},
			Name:     name.Lexeme,
			Pos:      dot.Pos,
		}
	case lexer.LBracket:
		return p.parseArrayLiteral()
	case lexer.LParen:
		p.advance()
		p.skipNewlines()
		expr := p.parseExpression(1)
		p.skipNewlines()
		p.consume(lexer.RParen, "expected ')' after expression")
		return expr
	default:
		p.errorAt(tok, fmt.Sprintf("expected expression, got %s", tok.Kind))
		p.advance()
		return &ast.Identifier{Name: "<error>", Pos: tok.Pos}
	}
}

func (p *Parser) parseArrayLiteral() ast.Expr {
	start := p.consume(lexer.LBracket, "expected '['")
	lit := &ast.ArrayLiteral{Pos: start.Pos}
	p.skipNewlines()
	if !p.check(lexer.RBracket) {
		for {
			lit.Elements = append(lit.Elements, p.parseExpression(1))
			p.skipNewlines()
			if !p.match(lexer.Comma) {
				break
			}
			p.skipNewlines()
		}
	}
	p.consume(lexer.RBracket, "expected ']' after array literal")
	return lit
}

func precedence(kind lexer.Kind) int {
	switch kind {
	case lexer.EqualEqual, lexer.BangEqual:
		return 1
	case lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
		return 2
	case lexer.Plus, lexer.Minus:
		return 3
	case lexer.Star, lexer.Slash, lexer.Percent:
		return 4
	default:
		return 0
	}
}
