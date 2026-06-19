package lexer

import "testing"

func TestLexUnicodeIdentifiers(t *testing.T) {
	tokens := Lex("变量💡 := 1\n用户👩‍💻 := 2\n_emoji✨ := 3\n姓名·字段 := 4")
	want := []struct {
		index  int
		kind   Kind
		lexeme string
	}{
		{0, Ident, "变量💡"},
		{1, Declare, ":="},
		{2, Int, "1"},
		{4, Ident, "用户👩‍💻"},
		{5, Declare, ":="},
		{6, Int, "2"},
		{8, Ident, "_emoji✨"},
		{9, Declare, ":="},
		{10, Int, "3"},
		{12, Ident, "姓名·字段"},
		{13, Declare, ":="},
		{14, Int, "4"},
	}

	for _, tc := range want {
		if tokens[tc.index].Kind != tc.kind || tokens[tc.index].Lexeme != tc.lexeme {
			t.Fatalf("tokens[%d] = %s, want %s(%q); tokens = %#v", tc.index, tokens[tc.index], tc.kind, tc.lexeme, tokens)
		}
	}
}

func TestLexTemplateString(t *testing.T) {
	tokens := Lex("value := `hello, \\(name)`\n")
	if tokens[2].Kind != TemplateString || tokens[2].Lexeme != "`hello, \\(name)`" {
		t.Fatalf("tokens[2] = %s, want TemplateString; tokens = %#v", tokens[2], tokens)
	}

	tokens = Lex("value := `hello\nworld`\nnext := 1\n")
	if tokens[2].Kind != TemplateString || tokens[2].Lexeme != "`hello\nworld`" {
		t.Fatalf("tokens[2] = %s, want multiline TemplateString; tokens = %#v", tokens[2], tokens)
	}
	if tokens[3].Kind != Newline || tokens[3].Pos.Line != 2 {
		t.Fatalf("tokens[3] = %#v, want newline after multiline template on line 2", tokens[3])
	}
}

func TestLexStaticSelector(t *testing.T) {
	tokens := Lex("User::fromJson(text)")
	if tokens[1].Kind != DoubleColon || tokens[1].Lexeme != "::" {
		t.Fatalf("tokens[1] = %s, want DoubleColon; tokens = %#v", tokens[1], tokens)
	}
}
