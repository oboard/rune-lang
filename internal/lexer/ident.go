package lexer

import "github.com/oboard/rune-lang/internal/syntax"

func isIdentStart(ch rune) bool {
	return syntax.IsIdentStart(ch)
}

func isIdentContinue(ch rune) bool {
	return syntax.IsIdentContinue(ch)
}
