package lexer

import "unicode"

func isIdentStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func isIdentContinue(ch rune) bool {
	return isIdentStart(ch) || unicode.IsDigit(ch)
}
