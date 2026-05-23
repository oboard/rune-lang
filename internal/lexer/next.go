package lexer

import "unicode"

func (l *Lexer) Next() Token {
	switch l.mode {
	case modeXMLTag:
		return l.nextXMLTag()
	case modeXMLText:
		return l.nextXMLText()
	case modeXMLExpr:
		return l.nextXMLExpr()
	default:
		return l.nextCode()
	}
}

func (l *Lexer) nextCode() Token {
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
		if l.match('.') && l.match('.') {
			return l.token(DotDotDot)
		}
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
	case '?':
		return l.token(Question)
	case '+':
		if l.match('+') {
			return l.token(PlusPlus)
		}
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
	case '&':
		if l.match('&') {
			return l.token(AndAnd)
		}
		return l.token(Illegal)
	case '|':
		if l.match('|') {
			return l.token(OrOr)
		}
		return l.token(Illegal)
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
		return l.token(Dollar)
	case '<':
		if l.match('=') {
			return l.token(LessEqual)
		}
		if l.peek() == '/' || isIdentStart(l.peek()) {
			l.mode = modeXMLTag
			l.xmlClosing = l.peek() == '/'
			l.xmlSelfClosed = false
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

func (l *Lexer) nextXMLTag() Token {
	l.skipXMLSpaces()
	l.start = l.curr
	if l.isAtEnd() {
		return l.token(EOF)
	}

	ch := l.advance()
	switch ch {
	case '@':
		return l.token(At)
	case '=':
		return l.token(Assign)
	case '"':
		return l.string()
	case '{':
		l.mode = modeXMLExpr
		l.xmlExprMode = modeXMLTag
		l.xmlExprDepth = 1
		return l.token(LBrace)
	case '}':
		return l.token(RBrace)
	case '/':
		if l.peek() == '>' {
			l.xmlSelfClosed = true
		}
		return l.token(Slash)
	case '>':
		tok := l.token(Greater)
		if l.xmlClosing {
			if l.xmlDepth > 0 {
				l.xmlDepth--
			}
		} else if !l.xmlSelfClosed {
			l.xmlDepth++
		}
		l.xmlClosing = false
		l.xmlSelfClosed = false
		if l.xmlDepth > 0 {
			l.mode = modeXMLText
		} else {
			l.mode = modeCode
		}
		return tok
	default:
		if isIdentStart(ch) {
			return l.xmlIdentifier()
		}
		return l.token(Illegal)
	}
}

func (l *Lexer) nextXMLText() Token {
	l.start = l.curr
	if l.isAtEnd() {
		return l.token(EOF)
	}
	switch l.peek() {
	case '<':
		l.advance()
		l.mode = modeXMLTag
		l.xmlClosing = l.peek() == '/'
		l.xmlSelfClosed = false
		return l.token(Less)
	case '{':
		l.advance()
		l.mode = modeXMLExpr
		l.xmlExprMode = modeXMLText
		l.xmlExprDepth = 1
		return l.token(LBrace)
	}
	for !l.isAtEnd() && l.peek() != '<' && l.peek() != '{' {
		l.advance()
	}
	return l.token(XMLText)
}

func (l *Lexer) nextXMLExpr() Token {
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
		if l.match('.') && l.match('.') {
			return l.token(DotDotDot)
		}
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
		l.xmlExprDepth++
		return l.token(LBrace)
	case '}':
		l.xmlExprDepth--
		tok := l.token(RBrace)
		if l.xmlExprDepth <= 0 {
			l.mode = l.xmlExprMode
			l.xmlExprMode = modeCode
		}
		return tok
	case '?':
		return l.token(Question)
	case '+':
		if l.match('+') {
			return l.token(PlusPlus)
		}
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
	case '&':
		if l.match('&') {
			return l.token(AndAnd)
		}
		return l.token(Illegal)
	case '|':
		if l.match('|') {
			return l.token(OrOr)
		}
		return l.token(Illegal)
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
		return l.token(Dollar)
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
