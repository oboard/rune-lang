package lexer

import "testing"

func TestLexTemplateString(t *testing.T) {
	tokens := Lex("value := `hello, ${name}`\n")
	if tokens[2].Kind != TemplateString || tokens[2].Lexeme != "`hello, ${name}`" {
		t.Fatalf("tokens[2] = %s, want TemplateString; tokens = %#v", tokens[2], tokens)
	}
}
