package lsp

import "testing"

func TestFormattingReturnsWholeDocumentEdit(t *testing.T) {
	uri := "file:///tmp/main.rn"
	src := "main()=>{\nname:=\"oboard\"\n@io.println(name) // print name\n}\n"
	s := &server{docs: map[string]string{uri: src}}

	edits := s.formatting(uri).([]map[string]any)
	if len(edits) != 1 {
		t.Fatalf("formatting edits = %d, want 1", len(edits))
	}
	got := edits[0]["newText"].(string)
	want := "main() => {\n  name := \"oboard\"\n  @io.println(name) // print name\n}\n"
	if got != want {
		t.Fatalf("formatted text = %q, want %q", got, want)
	}
}

func TestFormattingReturnsNoEditsWhenUnchanged(t *testing.T) {
	uri := "file:///tmp/main.rn"
	src := "main() => {\n  @io.println(\"ok\")\n}\n"
	s := &server{docs: map[string]string{uri: src}}

	edits := s.formatting(uri).([]map[string]any)
	if len(edits) != 0 {
		t.Fatalf("formatting edits = %d, want 0", len(edits))
	}
}
