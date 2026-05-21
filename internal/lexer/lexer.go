package lexer

import "unicode"

type Lexer struct {
	input  []rune
	start  int
	curr   int
	line   int
	column int
}

func Lex(src string) []Token {
	l := New(src)
	var tokens []Token
	for {
		tok := l.Next()
		tokens = append(tokens, tok)
		if tok.Kind == EOF {
			return tokens
		}
	}
}

func New(src string) *Lexer {
	return &Lexer{
		input:  []rune(src),
		line:   1,
		column: 1,
	}
}

func (l *Lexer) Next() Token {
	for {
		l.skipSpaces()
		l.start = l.curr
		if l.isAtEnd() {
			return l.token(EOF)
		}

		if l.matchComment() {
			continue
		}
		break
	}

	l.start = l.curr
	ch := l.advance()
	switch ch {
	case '\n':
		return l.token(Newline)
	case '@':
		return l.token(At)
	case '.':
		return l.token(Dot)
	case ',':
		return l.token(Comma)
	case ':':
		if l.match('=') {
			return l.token(Declare)
		}
		return l.token(Colon)
	case '(':
		return l.token(LParen)
	case ')':
		return l.token(RParen)
	case '{':
		return l.token(LBrace)
	case '}':
		return l.token(RBrace)
	case '+':
		return l.token(Plus)
	case '-':
		return l.token(Minus)
	case '*':
		return l.token(Star)
	case '/':
		return l.token(Slash)
	case '%':
		return l.token(Percent)
	case '!':
		if l.match('=') {
			return l.token(BangEqual)
		}
		return l.token(Bang)
	case '=':
		if l.match('>') {
			return l.token(FatArrow)
		}
		if l.match('=') {
			return l.token(EqualEqual)
		}
		return l.token(Assign)
	case '~':
		if l.match('=') {
			return l.token(MutDeclare)
		}
		return l.token(Illegal)
	case '<':
		if l.match('=') {
			return l.token(LessEqual)
		}
		return l.token(Less)
	case '>':
		if l.match('=') {
			return l.token(GreaterEqual)
		}
		return l.token(Greater)
	case '"':
		return l.string()
	case '_':
		if isIdentContinue(l.peek()) {
			return l.identifier()
		}
		return l.token(Underscore)
	default:
		if unicode.IsDigit(ch) {
			return l.number()
		}
		if isIdentStart(ch) {
			return l.identifier()
		}
		return l.token(Illegal)
	}
}

func (l *Lexer) skipSpaces() {
	for !l.isAtEnd() {
		switch l.peek() {
		case ' ', '\t', '\r':
			l.advance()
		default:
			return
		}
	}
}

func (l *Lexer) matchComment() bool {
	if l.peek() != '/' || l.peekNext() == 0 {
		return false
	}
	switch l.peekNext() {
	case '/':
		for !l.isAtEnd() && l.peek() != '\n' {
			l.advance()
		}
		return true
	case '*':
		l.advance()
		l.advance()
		for !l.isAtEnd() {
			if l.peek() == '*' && l.peekNext() == '/' {
				l.advance()
				l.advance()
				return true
			}
			l.advance()
		}
		return true
	default:
		return false
	}
}

func (l *Lexer) string() Token {
	escaped := false
	for !l.isAtEnd() {
		ch := l.advance()
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '"' {
			return l.token(String)
		}
	}
	return l.token(Illegal)
}

func (l *Lexer) number() Token {
	for unicode.IsDigit(l.peek()) {
		l.advance()
	}
	return l.token(Int)
}

func (l *Lexer) identifier() Token {
	for isIdentContinue(l.peek()) {
		l.advance()
	}
	return l.token(Ident)
}

func (l *Lexer) token(kind Kind) Token {
	return Token{
		Kind:   kind,
		Lexeme: string(l.input[l.start:l.curr]),
		Pos: Position{
			Offset: l.start,
			Line:   l.lineAt(l.start),
			Column: l.columnAt(l.start),
		},
	}
}

func (l *Lexer) lineAt(offset int) int {
	line := 1
	for i := 0; i < offset && i < len(l.input); i++ {
		if l.input[i] == '\n' {
			line++
		}
	}
	return line
}

func (l *Lexer) columnAt(offset int) int {
	col := 1
	for i := offset - 1; i >= 0 && i < len(l.input); i-- {
		if l.input[i] == '\n' {
			break
		}
		col++
	}
	return col
}

func (l *Lexer) advance() rune {
	if l.isAtEnd() {
		return 0
	}
	ch := l.input[l.curr]
	l.curr++
	if ch == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	return ch
}

func (l *Lexer) match(ch rune) bool {
	if l.isAtEnd() || l.input[l.curr] != ch {
		return false
	}
	l.advance()
	return true
}

func (l *Lexer) peek() rune {
	if l.isAtEnd() {
		return 0
	}
	return l.input[l.curr]
}

func (l *Lexer) peekNext() rune {
	if l.curr+1 >= len(l.input) {
		return 0
	}
	return l.input[l.curr+1]
}

func (l *Lexer) isAtEnd() bool {
	return l.curr >= len(l.input)
}

func isIdentStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func isIdentContinue(ch rune) bool {
	return isIdentStart(ch) || unicode.IsDigit(ch)
}
