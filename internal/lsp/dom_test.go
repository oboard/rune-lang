package lsp

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/checker"
)

// TestDOMMemberCompletion verifies that DOM ambient types resolve through
// prog.Info.Types and provide member completions.
func TestDOMMemberCompletion(t *testing.T) {
	if checker.LocateDOMLib() == "" {
		t.Skip("lib.dom.d.ts not on this machine")
	}
	uri := "file:///tmp/main.rn"
	src := `main() => {
	el := document.createElement("div")
	el.
}
`
	s := &server{docs: map[string]string{uri: src}}
	pos := positionOf(src, "el.", ".")
	pos.Character++
	items := s.completion(uri, pos).([]map[string]any)
	for _, label := range []string{"classList", "appendChild", "querySelector", "textContent", "id"} {
		if !domCompletionContains(items, label) {
			t.Errorf("missing completion %q; sample items: %.200v", label, items[:min(5, len(items))])
		}
	}
}

// TestDOMHoverDocument checks that hovering `document` returns a typed hover
// pointing at the ambient Document interface.
func TestDOMHoverDocument(t *testing.T) {
	if checker.LocateDOMLib() == "" {
		t.Skip("lib.dom.d.ts not on this machine")
	}
	uri := "file:///tmp/main.rn"
	src := `main() => {
	document
}
`
	s := &server{docs: map[string]string{uri: src}}
	hover := s.hover(uri, positionOf(src, "document", "document"))
	hoverMap, ok := hover.(map[string]any)
	if !ok {
		t.Fatalf("hover = %#v", hover)
	}
	val := hoverValue(hoverMap)
	if !strings.Contains(val, "Document") {
		t.Errorf("hover = %q, want Document", val)
	}
}

// TestDOMErrorDetection shows that calling an unknown method on a DOM
// receiver surfaces via diagnostics (not silently passes).
func TestDOMErrorDetection(t *testing.T) {
	if checker.LocateDOMLib() == "" {
		t.Skip("lib.dom.d.ts not on this machine")
	}
	uri := "file:///tmp/main.rn"
	src := `main() => {
	el := document.createElement("div")
	el.noSuchMethod()
}
`
	s := &server{docs: map[string] string{uri: src}}
	_, diags := s.analyze(uri)
	found := false
	for _, d := range diags {
		if strings.Contains(string(d.Message), "noSuchMethod") {
			found = true
		}
	}
	if !found {
		t.Errorf("diagnostics = %#v, want noSuchMethod error", diags)
	}
}

func domCompletionContains(items []map[string]any, needle string) bool {
	for _, item := range items {
		if item["label"] == needle {
			return true
		}
	}
	return false
}
