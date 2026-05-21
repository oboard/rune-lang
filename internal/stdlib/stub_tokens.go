package stdlib

import (
	"fmt"

	"github.com/oboard/rune-lang/internal/lexer"
)

func (p *stubParser) skipNewlines() {
	for p.match(lexer.Newline) {
	}
}

func (p *stubParser) consume(kind lexer.Kind, message string) lexer.Token {
	if p.check(kind) {
		return p.advance()
	}
	tok := p.peek()
	return lexer.Token{Kind: lexer.Illegal, Lexeme: message, Pos: tok.Pos}
}

func (p *stubParser) match(kind lexer.Kind) bool {
	if p.check(kind) {
		p.advance()
		return true
	}
	return false
}

func (p *stubParser) check(kind lexer.Kind) bool {
	return p.peek().Kind == kind
}

func (p *stubParser) advance() lexer.Token {
	if !p.check(lexer.EOF) {
		p.curr++
	}
	return p.previous()
}

func (p *stubParser) peek() lexer.Token {
	if p.curr >= len(p.tokens) {
		return lexer.Token{Kind: lexer.EOF}
	}
	return p.tokens[p.curr]
}

func (p *stubParser) previous() lexer.Token {
	if p.curr == 0 {
		return lexer.Token{Kind: lexer.EOF}
	}
	return p.tokens[p.curr-1]
}

func (p *stubParser) hasErrorToken(tok lexer.Token) bool {
	return tok.Kind == lexer.Illegal || tok.Kind == lexer.EOF
}

func (p *stubParser) errorf(tok lexer.Token, format string, args ...any) error {
	return fmt.Errorf("%s:%s: %s", p.path, tok.Pos, fmt.Sprintf(format, args...))
}

func unquote(src string) string {
	if len(src) >= 2 && src[0] == '"' && src[len(src)-1] == '"' {
		return src[1 : len(src)-1]
	}
	return src
}
