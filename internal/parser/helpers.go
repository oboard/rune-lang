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
	p.match(lexer.Tilde)
	p.skipNewlines()
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

func (p *Parser) looksLikeMacroFunctionDecl() bool {
	saved := p.curr
	defer func() { p.curr = saved }()
	return p.match(lexer.Hash) && p.looksLikeFunctionDecl()
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
	if p.match(lexer.Arrow) {
		p.skipTypeNameTokens()
		p.skipNewlines()
	}
	return p.check(lexer.FatArrow)
}

func (p *Parser) questionIsPostfixUnwrap() bool {
	if p.curr+1 >= len(p.tokens) {
		return true
	}
	switch p.tokens[p.curr+1].Kind {
	case lexer.EOF, lexer.Newline, lexer.RParen, lexer.RBracket, lexer.RBrace, lexer.Comma:
		return true
	default:
		return false
	}
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
	if i >= len(p.tokens) || p.tokens[i].Kind != lexer.Ident || isLiteralIdentifier(p.tokens[i].Lexeme) {
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

func (p *Parser) looksLikeObjectDestructureDecl() bool {
	saved := p.curr
	defer func() { p.curr = saved }()
	if !p.match(lexer.LBrace) {
		return false
	}
	p.skipNewlines()
	if p.check(lexer.RBrace) {
		return false
	}
	for {
		if !p.match(lexer.Ident) {
			return false
		}
		if p.match(lexer.Colon) {
			p.skipNewlines()
			if !p.match(lexer.Ident) {
				return false
			}
		}
		p.skipNewlines()
		if !p.match(lexer.Comma) {
			break
		}
		p.skipNewlines()
		if p.check(lexer.RBrace) {
			break
		}
	}
	if !p.match(lexer.RBrace) {
		return false
	}
	p.skipNewlines()
	return p.check(lexer.Declare) || p.check(lexer.MutDeclare)
}

func (p *Parser) looksLikeMapLiteralBody() bool {
	if !p.check(lexer.LBrace) {
		return false
	}
	i := p.curr + 1
	for i < len(p.tokens) && p.tokens[i].Kind == lexer.Newline {
		i++
	}
	if i >= len(p.tokens) || p.tokens[i].Kind == lexer.RBrace {
		return false
	}
	if p.tokens[i].Kind == lexer.Ident && !isLiteralIdentifier(p.tokens[i].Lexeme) {
		return false
	}
	depth := 0
	for ; i < len(p.tokens); i++ {
		switch p.tokens[i].Kind {
		case lexer.LParen, lexer.LBracket, lexer.LBrace:
			depth++
		case lexer.RParen, lexer.RBracket:
			if depth > 0 {
				depth--
			}
		case lexer.RBrace:
			if depth == 0 {
				return false
			}
			depth--
		case lexer.Newline, lexer.Question, lexer.FatArrow:
			if depth == 0 {
				return false
			}
		case lexer.Colon:
			if depth == 0 {
				return true
			}
		}
	}
	return false
}

func (p *Parser) tokensLookLikePatternBranch(i int) bool {
	for i < len(p.tokens) && p.tokens[i].Kind == lexer.Newline {
		i++
	}
	i = p.skipPatternLookahead(i)
	if i < 0 {
		return false
	}
	for i < len(p.tokens) && p.tokens[i].Kind == lexer.Newline {
		i++
	}
	return i < len(p.tokens) && p.tokens[i].Kind == lexer.FatArrow
}

func (p *Parser) skipPatternLookahead(i int) int {
	i = p.skipSinglePatternLookahead(i)
	if i < 0 {
		return -1
	}
	for i < len(p.tokens) && p.tokens[i].Kind == lexer.BitOr {
		i = p.skipSinglePatternLookahead(i + 1)
		if i < 0 {
			return -1
		}
	}
	return i
}

func (p *Parser) skipSinglePatternLookahead(i int) int {
	if i >= len(p.tokens) {
		return -1
	}
	rangeStart := false
	switch p.tokens[i].Kind {
	case lexer.Underscore, lexer.Int, lexer.Double, lexer.BigInt, lexer.String, lexer.Char:
		i++
		rangeStart = p.tokens[i-1].Kind != lexer.Underscore
	case lexer.Ident:
		if i+1 < len(p.tokens) && p.tokens[i+1].Kind == lexer.LParen {
			i += 2
			depth := 1
			for i < len(p.tokens) && depth > 0 {
				switch p.tokens[i].Kind {
				case lexer.LParen:
					depth++
				case lexer.RParen:
					depth--
				}
				i++
			}
			break
		}
		if i+2 < len(p.tokens) && p.tokens[i+1].Kind == lexer.Dot && p.tokens[i+2].Kind == lexer.Ident {
			i += 3
			break
		}
		rangeStart = isLiteralIdentifier(p.tokens[i].Lexeme)
		i++
	case lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
		i++
		if i >= len(p.tokens) {
			return -1
		}
		if p.tokens[i].Kind != lexer.Int && p.tokens[i].Kind != lexer.Double && p.tokens[i].Kind != lexer.BigInt && p.tokens[i].Kind != lexer.String && p.tokens[i].Kind != lexer.Char && p.tokens[i].Kind != lexer.Ident {
			return -1
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
		rangeStart = true
	case lexer.LBrace:
		depth := 1
		i++
		for i < len(p.tokens) && depth > 0 {
			switch p.tokens[i].Kind {
			case lexer.LBrace:
				depth++
			case lexer.RBrace:
				depth--
			}
			i++
		}
	default:
		return -1
	}
	if rangeStart && i < len(p.tokens) && p.tokens[i].Kind == lexer.DotDotEqual {
		i++
		if i >= len(p.tokens) {
			return -1
		}
		switch p.tokens[i].Kind {
		case lexer.Int, lexer.Double, lexer.BigInt, lexer.String, lexer.Char:
			i++
		case lexer.Ident:
			if i+2 < len(p.tokens) && p.tokens[i+1].Kind == lexer.Dot && p.tokens[i+2].Kind == lexer.Ident {
				i += 3
			} else {
				i++
			}
		default:
			return -1
		}
	}
	return i
}

func isLiteralIdentifier(name string) bool {
	switch name {
	case "true", "false", "null":
		return true
	default:
		return false
	}
}

func (p *Parser) consumeStatementEnd() {
	if p.match(lexer.Newline) {
		p.skipNewlines()
	}
}

func (p *Parser) consumeFieldSeparator(close lexer.Kind) bool {
	p.consumeStatementEnd()
	if p.check(close) || p.check(lexer.EOF) {
		return false
	}
	if p.match(lexer.Comma) {
		p.skipNewlines()
		return false
	}
	return true
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
	if !p.check(lexer.EOF) {
		p.advance()
	}
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

func (p *Parser) kindAt(index int) lexer.Kind {
	if index >= len(p.tokens) {
		return lexer.EOF
	}
	return p.tokens[index].Kind
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
		case lexer.Ident, lexer.At, lexer.BitAnd, lexer.Dot, lexer.Comma, lexer.Colon, lexer.Question, lexer.Arrow:
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
