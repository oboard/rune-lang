package lexer

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
