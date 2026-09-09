package lexer

import "unicode"

func (l *Lexer) Next() Token {
	var tok Token
	switch l.mode {
	case modeXMLTag:
		tok = l.nextXMLTag()
	case modeXMLText:
		tok = l.nextXMLText()
	case modeXMLExpr:
		tok = l.nextXMLExpr()
	default:
		tok = l.nextCode()
	}
	return l.finishToken(tok)
}

func (l *Lexer) finishToken(tok Token) Token {
	switch tok.Kind {
	case Ident, Int, Double, BigInt, String, TemplateString, Char, Regex, XMLText, RParen, RBracket:
		l.canStartRegex = false
		l.canStartXML = false
	case RBrace:
		l.canStartRegex = false
		// A closing brace of an XML expression may be followed by the outer element's closing tag.
		l.canStartXML = l.xmlDepth > 0
	case Less:
		l.canStartRegex = !(l.peek() == '/' || isIdentStart(l.peek()))
		l.canStartXML = true
	case Greater:
		// Only XML tag '>' allows subsequent '<' to start a new element
		l.canStartXML = l.mode == modeXMLTag || l.mode == modeXMLText
	case EOF:
		l.canStartXML = true
	default:
		l.canStartRegex = true
		l.canStartXML = true
	}
	return tok
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
	case '#':
		return l.token(Hash)
	case '.':
		if l.match('.') {
			if l.match('.') {
				return l.token(DotDotDot)
			}
			if l.match('<') {
				return l.token(DotDotLess)
			}
			if l.match('=') {
				return l.token(DotDotEqual)
			}
			return l.token(DotDot)
		}
		return l.token(Dot)
	case ',':
		return l.token(Comma)
	case ':':
		if l.match('=') {
			if l.match(':') {
				return l.token(MutDeclare)
			}
			return l.token(Declare)
		}
		if l.match(':') {
			return l.token(DoubleColon)
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
		if l.match('?') {
			return l.token(QuestionQuestion)
		}
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
		if l.canStartRegex {
			return l.regex()
		}
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
		return l.token(BitAnd)
	case '|':
		if l.match('|') {
			return l.token(OrOr)
		}
		return l.token(BitOr)
	case '^':
		return l.token(BitXor)
	case '=':
		if l.match('>') {
			return l.token(FatArrow)
		}
		if l.match('=') {
			return l.token(EqualEqual)
		}
		return l.token(Assign)
	case '~':
		return l.token(Tilde)
	case '$':
		return l.token(Dollar)
	case '<':
		if l.match('=') {
			return l.token(LessEqual)
		}
		if l.match('<') {
			return l.token(ShiftLeft)
		}
		if l.canStartXML && (l.peek() == '/' || isIdentStart(l.peek())) {
			l.mode = modeXMLTag
			l.xmlClosing = l.peek() == '/'
			l.xmlSelfClosed = false
		}
		return l.token(Less)
	case '>':
		if l.match('=') {
			return l.token(GreaterEqual)
		}
		if l.match('>') {
			if l.match('>') {
				return l.token(UnsignedShiftRight)
			}
			return l.token(ShiftRight)
		}
		return l.token(Greater)
	case '"':
		return l.string()
	case '`':
		return l.templateString()
	case '\'':
		if !l.canStartRegex {
			return l.token(Apostrophe)
		}
		return l.char()
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
		if l.xmlExprDepth == 0 {
			l.xmlSavedDepth = l.xmlDepth
		}
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
		if l.xmlExprDepth > 0 && l.xmlDepth == l.xmlSavedDepth {
			// Closing tag finished the top-level element inside an expression; stay in expression code mode.
			l.mode = modeCode
		} else if l.xmlDepth > 0 {
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
		if l.xmlExprDepth == 0 {
			l.xmlSavedDepth = l.xmlDepth
			l.xmlExprDepth = 1
			l.xmlExprMode = modeXMLText
		} else {
			l.xmlExprDepth++
		}
		l.mode = modeXMLExpr
		return l.token(LBrace)
	case '}':
		if l.xmlExprDepth > 0 {
			l.advance()
			l.xmlExprDepth--
			l.mode = l.xmlExprMode
			return l.token(RBrace)
		}
	}
	for !l.isAtEnd() && l.peek() != '<' && l.peek() != '{' {
		if l.xmlExprDepth > 0 && l.peek() == '}' {
			break
		}
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
		if l.match('.') {
			if l.match('.') {
				return l.token(DotDotDot)
			}
			if l.match('=') {
				return l.token(DotDotEqual)
			}
			return l.token(DotDot)
		}
		return l.token(Dot)
	case ',':
		return l.token(Comma)
	case ':':
		if l.match('=') {
			if l.match(':') {
				return l.token(MutDeclare)
			}
			return l.token(Declare)
		}
		if l.match(':') {
			return l.token(DoubleColon)
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
		}
		return tok
	case '?':
		if l.match('?') {
			return l.token(QuestionQuestion)
		}
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
		if l.canStartRegex {
			return l.regex()
		}
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
		return l.token(BitAnd)
	case '|':
		if l.match('|') {
			return l.token(OrOr)
		}
		return l.token(BitOr)
	case '^':
		return l.token(BitXor)
	case '=':
		if l.match('>') {
			return l.token(FatArrow)
		}
		if l.match('=') {
			return l.token(EqualEqual)
		}
		return l.token(Assign)
	case '~':
		return l.token(Tilde)
	case '$':
		return l.token(Dollar)
	case '<':
		if l.match('=') {
			return l.token(LessEqual)
		}
		if l.match('<') {
			return l.token(ShiftLeft)
		}
		if l.canStartXML && (l.peek() == '/' || isIdentStart(l.peek())) {
			l.mode = modeXMLTag
			l.xmlClosing = l.peek() == '/'
			l.xmlSelfClosed = false
		}
		return l.token(Less)
	case '>':
		if l.match('=') {
			return l.token(GreaterEqual)
		}
		if l.match('>') {
			if l.match('>') {
				return l.token(UnsignedShiftRight)
			}
			return l.token(ShiftRight)
		}
		return l.token(Greater)
	case '"':
		return l.string()
	case '`':
		return l.templateString()
	case '\'':
		return l.char()
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
