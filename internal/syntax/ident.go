package syntax

import "unicode"

func IsIdentStart(ch rune) bool {
	if ch == '_' {
		return true
	}
	if unicode.IsDigit(ch) || unicode.IsMark(ch) {
		return false
	}
	return isIdentText(ch)
}

func IsIdentContinue(ch rune) bool {
	return IsIdentStart(ch) || unicode.IsDigit(ch) || unicode.IsMark(ch) || ch == '\u200c' || ch == '\u200d'
}

func isIdentText(ch rune) bool {
	if ch == 0 || unicode.IsSpace(ch) || unicode.IsControl(ch) {
		return false
	}
	if isSyntaxPunctuation(ch) {
		return false
	}
	return unicode.IsLetter(ch) || unicode.IsNumber(ch) || unicode.IsSymbol(ch) || unicode.IsPunct(ch)
}

func isSyntaxPunctuation(ch rune) bool {
	switch ch {
	case '@', '$', '.', ',', ':', '(', ')', '[', ']', '{', '}', '?',
		'+', '-', '*', '/', '%', '!', '&', '|', '^', '=', '~', '<', '>',
		'"', '\'', '\\', ';', '#', '`':
		return true
	default:
		return false
	}
}
