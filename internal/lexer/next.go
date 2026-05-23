package lexer

import "unicode"

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
	case '[':
		return l.token(LBracket)
	case ']':
		return l.token(RBracket)
	case '{':
		return l.token(LBrace)
	case '}':
		return l.token(RBrace)
	case '+':
		return l.token(Plus)
	case '-':
		if l.match('>') {
			return l.token(Arrow)
		}
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
	case '$':
		if l.match('=') {
			return l.token(SignalDeclare)
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
