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

func (l *Lexer) char() Token {
	escaped := false
	for !l.isAtEnd() {
		ch := l.advance()
		if ch == '\n' {
			return l.token(Illegal)
		}
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '\'' {
			return l.token(Char)
		}
	}
	return l.token(Illegal)
}

func (l *Lexer) regex() Token {
	escaped := false
	inClass := false
	for !l.isAtEnd() {
		ch := l.advance()
		if ch == '\n' {
			return l.token(Illegal)
		}
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
				for isRegexFlag(l.peek()) {
					l.advance()
				}
				return l.token(Regex)
			}
		}
	}
	return l.token(Illegal)
}

func isRegexFlag(ch rune) bool {
	return (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z')
}

func (l *Lexer) number() Token {
	for unicode.IsDigit(l.peek()) {
		l.advance()
	}
	isDouble := false
	if l.peek() == 'n' {
		l.advance()
		return l.token(BigInt)
	}
	if l.peek() == '.' && unicode.IsDigit(l.peekNext()) {
		isDouble = true
		l.advance()
		for unicode.IsDigit(l.peek()) {
			l.advance()
		}
	}
	if l.peek() == 'e' || l.peek() == 'E' {
		isDouble = true
		l.advance()
		if l.peek() == '+' || l.peek() == '-' {
			l.advance()
		}
		for unicode.IsDigit(l.peek()) {
			l.advance()
		}
	}
	if isDouble {
		return l.token(Double)
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
