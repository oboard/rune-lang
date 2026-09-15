package lexer

import "testing"

func TestLexOptionalChain(t *testing.T) {
	tokens := Lex("event.target?.value ?? \"\"")
	kinds := []Kind{Ident, Dot, Ident, QuestionDot, Ident, QuestionQuestion, String, EOF}
	if len(tokens) != len(kinds) {
		t.Fatalf("tokens = %#v, want %d tokens", tokens, len(kinds))
	}
	for i, kind := range kinds {
		if tokens[i].Kind != kind {
			t.Fatalf("tokens[%d] = %s, want %s; tokens = %#v", i, tokens[i].Kind, kind, tokens)
		}
	}
}
