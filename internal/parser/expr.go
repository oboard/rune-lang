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
			name := ""
			if ident, ok := left.(*ast.Identifier); ok {
				name = ident.Name
			}
			p.skipNewlines()
			left = &ast.AssignExpr{Name: name, Target: left, Value: p.parseExpression(1), Pos: left.Position()}
			continue
		}
		if minPrec <= 1 && p.check(lexer.Question) && p.questionIsPostfixUnwrap() {
			question := p.advance()
			left = &ast.ResultUnwrapExpr{Expr: left, Pos: question.Pos}
			continue
		}
		if minPrec <= 1 && p.match(lexer.Question) {
			p.skipNewlines()
			consequence := p.parseExpression(1)
			p.skipNewlines()
			p.consume(lexer.Colon, "expected ':' after ternary consequence")
			p.skipNewlines()
			left = &ast.TernaryExpr{
				Condition:   left,
				Consequence: consequence,
				Alternative: p.parseExpression(1),
				Pos:         left.Position(),
			}
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
		ret := p.parseLambdaReturnType()
		p.consume(lexer.FatArrow, "expected '=>' after watch handler parameter")
		p.skipNewlines()
		return &ast.LambdaExpr{
			Params:            params.names,
			ParamPos:          params.positions,
			ParamTypes:        params.types,
			ParamTypeDisplays: params.typeDisplays,
			ReturnType:        ret.canonical,
			ReturnDisplay:     ret.display,
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
	ret := p.parseLambdaReturnType()
	p.consume(lexer.FatArrow, "expected '=>' after lambda parameter")
	p.skipNewlines()
	body := p.parseBody()
	return &ast.LambdaExpr{
		Params:            params.names,
		ParamPos:          params.positions,
		ParamTypes:        params.types,
		ParamTypeDisplays: params.typeDisplays,
		ReturnType:        ret.canonical,
		ReturnDisplay:     ret.display,
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

func (p *Parser) parseLambdaReturnType() parsedType {
	p.skipNewlines()
	if !p.match(lexer.Arrow) {
		return parsedType{}
	}
	p.skipNewlines()
	ret := p.parseTypeName()
	p.skipNewlines()
	return ret
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
	if p.match(lexer.Minus, lexer.Bang, lexer.Tilde) {
		op := p.previous()
		return &ast.UnaryExpr{Op: op.Kind, Expr: p.parseExpression(11), Pos: op.Pos}
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
		return &ast.DoubleLiteral{Value: value, Raw: tok.Lexeme, Pos: tok.Pos}
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
	case lexer.Char:
		p.advance()
		value, err := strconv.Unquote(tok.Lexeme)
		runes := []rune(value)
		if err != nil || len(runes) != 1 {
			p.errorAt(tok, "invalid char literal")
			return &ast.CharLiteral{Value: 0, Pos: tok.Pos}
		}
		return &ast.CharLiteral{Value: runes[0], Pos: tok.Pos}
	case lexer.Regex:
		p.advance()
		pattern, flags, ok := splitRegexLiteral(tok.Lexeme)
		if !ok {
			p.errorAt(tok, "invalid regex literal")
		}
		return &ast.RegexLiteral{Pattern: pattern, Flags: flags, Raw: tok.Lexeme, Pos: tok.Pos}
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
		return p.parseBraceLiteral()
	case lexer.LParen:
		start := p.advance()
		p.skipNewlines()
		expr := p.parseExpression(1)
		p.skipNewlines()
		if p.match(lexer.Comma) {
			elems := []ast.Expr{expr}
			p.skipNewlines()
			for !p.check(lexer.RParen) && !p.check(lexer.EOF) {
				elems = append(elems, p.parseExpression(1))
				p.skipNewlines()
				if !p.match(lexer.Comma) {
					break
				}
				p.skipNewlines()
			}
			p.consume(lexer.RParen, "expected ')' after tuple literal")
			return &ast.TupleLiteral{Elements: elems, Pos: start.Pos}
		}
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

func splitRegexLiteral(raw string) (string, string, bool) {
	if len(raw) < 2 || raw[0] != '/' {
		return "", "", false
	}
	escaped := false
	inClass := false
	for idx := 1; idx < len(raw); idx++ {
		ch := raw[idx]
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		switch ch {
		case '[':
			inClass = true
		case ']':
			inClass = false
		case '/':
			if !inClass {
				return raw[1:idx], raw[idx+1:], true
			}
		}
	}
	return "", "", false
}

func (p *Parser) parseReactiveLiteral() ast.Expr {
	start := p.consume(lexer.Dollar, "expected '$'")
	switch p.peek().Kind {
	case lexer.LBracket:
		return &ast.ReactiveLiteral{Value: p.parseArrayLiteral(), Pos: start.Pos}
	case lexer.LBrace:
		return &ast.ReactiveLiteral{Value: p.parseBraceLiteral(), Pos: start.Pos}
	default:
		p.errorAt(p.peek(), "expected '[' or '{' after '$'")
		return &ast.Identifier{Name: "<error>", Pos: start.Pos}
	}
}

func (p *Parser) parseBraceLiteral() ast.Expr {
	if p.looksLikeMapLiteralBody() {
		return p.parseMapLiteral()
	}
	return p.parseAnonymousObjectLiteral()
}

func (p *Parser) parseMapLiteral() ast.Expr {
	start := p.consume(lexer.LBrace, "expected '{'")
	lit := &ast.MapLiteral{Pos: start.Pos}
	p.skipNewlines()
	for !p.check(lexer.RBrace) && !p.check(lexer.EOF) {
		key := p.parseExpression(1)
		p.consume(lexer.Colon, "expected ':' after map key")
		p.skipNewlines()
		value := p.parseExpression(1)
		lit.Entries = append(lit.Entries, ast.MapEntry{Key: key, Value: value, Pos: key.Position()})
		p.consumeStatementEnd()
		p.match(lexer.Comma)
		p.skipNewlines()
	}
	p.consume(lexer.RBrace, "expected '}' after map literal")
	return lit
}

func (p *Parser) parseAnonymousObjectLiteral() ast.Expr {
	start := p.consume(lexer.LBrace, "expected '{'")
	lit := &ast.AnonymousObjectLiteral{Pos: start.Pos}
	p.skipNewlines()
	for !p.check(lexer.RBrace) && !p.check(lexer.EOF) {
		private := p.parsePrivateModifier()
		if p.looksLikeFunctionDecl() {
			lit.Fields = append(lit.Fields, p.parseAnonymousObjectMethod(private))
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
			Name:    fieldName.Lexeme,
			Private: private,
			Value:   value,
			Pos:     fieldName.Pos,
		})
		p.consumeStatementEnd()
		p.match(lexer.Comma)
		p.skipNewlines()
	}
	p.consume(lexer.RBrace, "expected '}' after object literal")
	return lit
}

func (p *Parser) parseAnonymousObjectMethod(private bool) ast.FieldValue {
	fn := p.parseFunctionWithReceiver("", private)
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
		Name:    fn.Name,
		Private: private,
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
	case lexer.OrOr:
		return 1
	case lexer.AndAnd:
		return 2
	case lexer.BitOr:
		return 3
	case lexer.BitXor:
		return 4
	case lexer.BitAnd:
		return 5
	case lexer.EqualEqual, lexer.BangEqual:
		return 6
	case lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
		return 7
	case lexer.ShiftLeft, lexer.ShiftRight, lexer.UnsignedShiftRight:
		return 8
	case lexer.Plus, lexer.Minus:
		return 9
	case lexer.Star, lexer.Slash, lexer.Percent:
		return 10
	default:
		return 0
	}
}
