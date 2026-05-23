package lexer

import "unicode"

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

func (l *Lexer) skipXMLSpaces() {
	for !l.isAtEnd() {
		switch l.peek() {
		case ' ', '\t', '\r', '\n':
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

func (l *Lexer) xmlIdentifier() Token {
	for isIdentContinue(l.peek()) || l.peek() == '-' || l.peek() == ':' {
		l.advance()
	}
	return l.token(Ident)
}
