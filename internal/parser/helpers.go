package parser

import "github.com/oboard/rune-lang/internal/lexer"

func (p *Parser) looksLikeTypeDecl() bool {
	saved := p.curr
	defer func() { p.curr = saved }()
	if !p.match(lexer.Ident) {
		return false
	}
	p.parseGenericNames()
	p.skipNewlines()
	return p.check(lexer.Colon)
}

func (p *Parser) looksLikeFunctionDecl() bool {
	saved := p.curr
	defer func() { p.curr = saved }()
	if !p.match(lexer.Ident) {
		return false
	}
	p.parseGenericNames()
	if !p.match(lexer.LParen) {
		return false
	}
	depth := 1
	for !p.check(lexer.EOF) && depth > 0 {
		tok := p.advance()
		switch tok.Kind {
		case lexer.LParen:
			depth++
		case lexer.RParen:
			depth--
		}
	}
	p.skipNewlines()
	if p.match(lexer.Arrow) {
		p.skipTypeNameTokens()
	}
	p.skipNewlines()
	return p.check(lexer.FatArrow)
}

func (p *Parser) looksLikeLambda() bool {
	saved := p.curr
	defer func() { p.curr = saved }()
	if !p.match(lexer.LParen) {
		return false
	}
	depth := 1
	for !p.check(lexer.EOF) && depth > 0 {
		tok := p.advance()
		switch tok.Kind {
		case lexer.LParen:
			depth++
		case lexer.RParen:
			depth--
		}
	}
	p.skipNewlines()
	return p.check(lexer.FatArrow)
}

func (p *Parser) looksLikePatternBranch() bool {
	i := p.curr
	return p.tokensLookLikePatternBranch(i)
}

func (p *Parser) looksLikePatternBlockAfterSubject() bool {
	if !p.check(lexer.LBrace) {
		return false
	}
	i := p.curr + 1
	return p.tokensLookLikePatternBranch(i)
}

func (p *Parser) looksLikeObjectLiteralBody() bool {
	if !p.check(lexer.LBrace) {
		return false
	}
	i := p.curr + 1
	for i < len(p.tokens) && p.tokens[i].Kind == lexer.Newline {
		i++
	}
	if i >= len(p.tokens) || p.tokens[i].Kind != lexer.Ident {
		return false
	}
	if i+1 >= len(p.tokens) {
		return false
	}
	switch p.tokens[i+1].Kind {
	case lexer.Colon:
		return true
	default:
		return false
	}
}

func (p *Parser) tokensLookLikePatternBranch(i int) bool {
	for i < len(p.tokens) && p.tokens[i].Kind == lexer.Newline {
		i++
	}
	if i >= len(p.tokens) {
		return false
	}
	switch p.tokens[i].Kind {
	case lexer.Underscore, lexer.Int, lexer.String:
		i++
	case lexer.Ident:
		if p.tokens[i].Lexeme != "true" && p.tokens[i].Lexeme != "false" {
			return false
		}
		i++
	case lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
		i++
		if i >= len(p.tokens) {
			return false
		}
		if p.tokens[i].Kind != lexer.Int && p.tokens[i].Kind != lexer.String && p.tokens[i].Kind != lexer.Ident {
			return false
		}
		i++
	case lexer.LParen:
		depth := 1
		i++
		for i < len(p.tokens) && depth > 0 {
			switch p.tokens[i].Kind {
			case lexer.LParen:
				depth++
			case lexer.RParen:
				depth--
			}
			i++
		}
	default:
		return false
	}
	for i < len(p.tokens) && p.tokens[i].Kind == lexer.Newline {
		i++
	}
	return i < len(p.tokens) && p.tokens[i].Kind == lexer.FatArrow
}

func (p *Parser) consumeStatementEnd() {
	if p.match(lexer.Newline) {
		p.skipNewlines()
	}
}

func (p *Parser) skipNewlines() {
	for p.match(lexer.Newline) {
	}
}

func (p *Parser) consume(kind lexer.Kind, message string) lexer.Token {
	if p.check(kind) {
		return p.advance()
	}
	tok := p.peek()
	p.errorAt(tok, message)
	return tok
}

func (p *Parser) match(kinds ...lexer.Kind) bool {
	for _, kind := range kinds {
		if p.check(kind) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) check(kind lexer.Kind) bool {
	if p.curr >= len(p.tokens) {
		return kind == lexer.EOF
	}
	return p.peek().Kind == kind
}

func (p *Parser) checkNext(kind lexer.Kind) bool {
	if p.curr+1 >= len(p.tokens) {
		return kind == lexer.EOF
	}
	return p.tokens[p.curr+1].Kind == kind
}

func (p *Parser) advance() lexer.Token {
	if !p.check(lexer.EOF) {
		p.curr++
	}
	return p.previous()
}

func (p *Parser) peek() lexer.Token {
	if p.curr >= len(p.tokens) {
		return lexer.Token{Kind: lexer.EOF}
	}
	return p.tokens[p.curr]
}

func (p *Parser) previous() lexer.Token {
	if p.curr == 0 {
		return lexer.Token{Kind: lexer.EOF}
	}
	return p.tokens[p.curr-1]
}

func (p *Parser) errorAt(tok lexer.Token, message string) {
	p.errors = append(p.errors, Error{Message: message, Pos: tok.Pos})
}

func (p *Parser) skipTypeNameTokens() {
	depth := 0
	for !p.check(lexer.EOF) {
		switch p.peek().Kind {
		case lexer.Ident, lexer.Comma, lexer.Arrow:
			p.advance()
		case lexer.LBracket, lexer.LParen:
			depth++
			p.advance()
		case lexer.RBracket, lexer.RParen:
			if depth == 0 {
				return
			}
			depth--
			p.advance()
		default:
			return
		}
	}
}
