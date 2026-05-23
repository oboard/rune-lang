package parser

import (
	"fmt"
	"strconv"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

func (p *Parser) parseExpression(minPrec int) ast.Expr {
	if minPrec <= 1 && p.check(lexer.LParen) && p.looksLikeLambda() {
		return p.parseLambda()
	}
	left := p.parseUnary()
	for {
		if p.check(lexer.LBrace) {
			if p.looksLikePatternBlockAfterSubject() {
				left = p.parseMatchExpr(left)
				continue
			}
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
			left = &ast.SelectorExpr{Receiver: left, Name: name.Lexeme, Pos: left.Position(), NamePos: name.Pos}
			continue
		}
		if p.match(lexer.PlusPlus) {
			left = &ast.PostfixExpr{Expr: left, Op: lexer.PlusPlus, Pos: left.Position()}
			continue
		}
		if minPrec <= 1 && p.match(lexer.Arrow) {
			p.skipNewlines()
			handler := p.parseWatchHandler()
			left = &ast.WatchExpr{Target: left, Handler: handler, Pos: left.Position()}
			continue
		}
		if minPrec <= 1 && p.match(lexer.Assign) {
			ident, ok := left.(*ast.Identifier)
			if !ok {
				p.errorAt(lexer.Token{Pos: left.Position()}, "assignment target must be an identifier")
				left = p.parseExpression(1)
				continue
			}
			p.skipNewlines()
			left = &ast.AssignExpr{Name: ident.Name, Value: p.parseExpression(1), Pos: ident.Pos}
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

func (p *Parser) parseWatchHandler() ast.Expr {
	if p.check(lexer.LParen) && p.looksLikeLambda() {
		params := p.parseLambdaParams()
		p.consume(lexer.FatArrow, "expected '=>' after watch handler parameter")
		p.skipNewlines()
		return &ast.LambdaExpr{
			Params:            params.names,
			ParamPos:          params.positions,
			ParamTypes:        params.types,
			ParamTypeDisplays: params.typeDisplays,
			Body:              p.parseBody(),
			Pos:               params.pos,
		}
	}
	if p.check(lexer.LBrace) {
		return &ast.LambdaExpr{
			Implicit: true,
			Body:     p.parseBody(),
			Pos:      p.peek().Pos,
		}
	}
	return p.parseExpression(1)
}

func (p *Parser) parseMatchExpr(subject ast.Expr) ast.Expr {
	start := p.consume(lexer.LBrace, "expected '{' after match subject")
	match := &ast.MatchExpr{Subject: subject, Pos: start.Pos}
	p.skipNewlines()
	for !p.check(lexer.RBrace) && !p.check(lexer.EOF) {
		pattern := p.parsePattern()
		p.consume(lexer.FatArrow, "expected '=>' after pattern")
		p.skipNewlines()
		expr := p.parseExpression(1)
		match.Branches = append(match.Branches, ast.PatternBranch{
			Pattern: pattern,
			Expr:    expr,
			Pos:     pattern.Position(),
		})
		p.consumeStatementEnd()
		p.match(lexer.Comma)
		p.skipNewlines()
	}
	p.consume(lexer.RBrace, "expected '}' after match expression")
	return match
}

func (p *Parser) parseLambda() ast.Expr {
	params := p.parseLambdaParams()
	p.consume(lexer.FatArrow, "expected '=>' after lambda parameter")
	p.skipNewlines()
	var body ast.Expr
	if p.check(lexer.LBrace) && !p.looksLikePatternBranch() {
		body = p.parseAnonymousObjectLiteral()
	} else {
		body = p.parseBody()
	}
	return &ast.LambdaExpr{
		Params:            params.names,
		ParamPos:          params.positions,
		ParamTypes:        params.types,
		ParamTypeDisplays: params.typeDisplays,
		Body:              body,
		Pos:               params.pos,
	}
}

type lambdaParams struct {
	names        []string
	positions    []lexer.Position
	types        []string
	typeDisplays []string
	pos          lexer.Position
}

func (p *Parser) parseLambdaParams() lambdaParams {
	if p.match(lexer.LParen) {
		pos := p.previous().Pos
		var names []string
		var positions []lexer.Position
		var types []string
		var typeDisplays []string
		p.skipNewlines()
		for !p.check(lexer.RParen) && !p.check(lexer.EOF) {
			name := p.consume(lexer.Ident, "expected lambda parameter")
			if name.Kind == lexer.Ident {
				names = append(names, name.Lexeme)
				positions = append(positions, name.Pos)
				paramType := parsedType{}
				if p.match(lexer.Colon) {
					paramType = p.parseTypeName()
				}
				types = append(types, paramType.canonical)
				typeDisplays = append(typeDisplays, paramType.display)
			}
			p.skipNewlines()
			if !p.match(lexer.Comma) {
				break
			}
			p.skipNewlines()
		}
		p.consume(lexer.RParen, "expected ')' after lambda parameters")
		return lambdaParams{names: names, positions: positions, types: types, typeDisplays: typeDisplays, pos: pos}
	}
	tok := p.peek()
	p.errorAt(tok, "lambda parameters must be parenthesized")
	return lambdaParams{pos: tok.Pos}
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
	case lexer.Double:
		p.advance()
		value, err := strconv.ParseFloat(tok.Lexeme, 64)
		if err != nil {
			p.errorAt(tok, "invalid double literal")
		}
		return &ast.DoubleLiteral{Value: value, Pos: tok.Pos}
	case lexer.BigInt:
		p.advance()
		return &ast.BigIntLiteral{Value: tok.Lexeme[:len(tok.Lexeme)-1], Pos: tok.Pos}
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
		if tok.Lexeme == "null" {
			return &ast.NullLiteral{Pos: tok.Pos}
		}
		if tok.Lexeme == "undefined" {
			return &ast.UndefinedLiteral{Pos: tok.Pos}
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
			NamePos:  name.Pos,
		}
	case lexer.LBracket:
		return p.parseArrayLiteral()
	case lexer.Dollar:
		return p.parseReactiveLiteral()
	case lexer.LBrace:
		return p.parseAnonymousObjectLiteral()
	case lexer.LParen:
		p.advance()
		p.skipNewlines()
		expr := p.parseExpression(1)
		p.skipNewlines()
		p.consume(lexer.RParen, "expected ')' after expression")
		return expr
	case lexer.Less:
		return p.parseXMLElement()
	default:
		p.errorAt(tok, fmt.Sprintf("expected expression, got %s", tok.Kind))
		p.advance()
		return &ast.Identifier{Name: "<error>", Pos: tok.Pos}
	}
}

func (p *Parser) parseReactiveLiteral() ast.Expr {
	start := p.consume(lexer.Dollar, "expected '$'")
	switch p.peek().Kind {
	case lexer.LBracket:
		return &ast.ReactiveLiteral{Value: p.parseArrayLiteral(), Pos: start.Pos}
	case lexer.LBrace:
		return &ast.ReactiveLiteral{Value: p.parseAnonymousObjectLiteral(), Pos: start.Pos}
	default:
		p.errorAt(p.peek(), "expected '[' or '{' after '$'")
		return &ast.Identifier{Name: "<error>", Pos: start.Pos}
	}
}

func (p *Parser) parseAnonymousObjectLiteral() ast.Expr {
	start := p.consume(lexer.LBrace, "expected '{'")
	lit := &ast.AnonymousObjectLiteral{Pos: start.Pos}
	p.skipNewlines()
	for !p.check(lexer.RBrace) && !p.check(lexer.EOF) {
		if p.looksLikeFunctionDecl() {
			lit.Fields = append(lit.Fields, p.parseAnonymousObjectMethod())
			p.consumeStatementEnd()
			p.match(lexer.Comma)
			p.skipNewlines()
			continue
		}
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
	p.consume(lexer.RBrace, "expected '}' after object literal")
	return lit
}

func (p *Parser) parseAnonymousObjectMethod() ast.FieldValue {
	fn := p.parseFunctionWithReceiver("")
	params := make([]string, 0, len(fn.Params))
	paramPos := make([]lexer.Position, 0, len(fn.Params))
	paramTypes := make([]string, 0, len(fn.Params))
	paramTypeDisplays := make([]string, 0, len(fn.Params))
	for _, param := range fn.Params {
		params = append(params, param.Name)
		paramPos = append(paramPos, param.Pos)
		paramTypes = append(paramTypes, param.Type)
		paramTypeDisplays = append(paramTypeDisplays, param.TypeDisplay)
	}
	return ast.FieldValue{
		Name: fn.Name,
		Value: &ast.LambdaExpr{
			Params:            params,
			ParamPos:          paramPos,
			ParamTypes:        paramTypes,
			ParamTypeDisplays: paramTypeDisplays,
			ReturnType:        fn.ReturnType,
			ReturnDisplay:     fn.ReturnDisplay,
			Body:              fn.Body,
			Pos:               fn.Pos,
		},
		Pos: fn.NamePos,
	}
}

func (p *Parser) parseArrayLiteral() ast.Expr {
	start := p.consume(lexer.LBracket, "expected '['")
	lit := &ast.ArrayLiteral{Pos: start.Pos}
	p.skipNewlines()
	if !p.check(lexer.RBracket) {
		for {
			if p.match(lexer.DotDotDot) {
				lit.Elements = append(lit.Elements, &ast.SpreadExpr{Expr: p.parseExpression(1), Pos: p.previous().Pos})
			} else {
				lit.Elements = append(lit.Elements, p.parseExpression(1))
			}
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
