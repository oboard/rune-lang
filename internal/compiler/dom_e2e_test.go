package compiler

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/checker"
)

// TestDOMEndToEndAnalysis verifies that the compiler's AnalyzeSource picks up
// ambient DOM declarations (interfaces, globals, functions) and surfaces DOM
// type errors while letting valid DOM code through cleanly.
func TestDOMEndToEndAnalysis(t *testing.T) {
	if checker.LocateDOMLib() == "" {
		t.Skip("lib.dom.d.ts not available on this machine")
	}

	// Valid: chained DOM usage must not produce any diagnostics.
	_, diags := AnalyzeSource("main.rn", `
main() => {
  el := document.createElement("div")
  el.classList.add("active")
  el.querySelector(".foo")
  document.body.appendChild(el)
}
`)
	if len(diags) != 0 {
		for _, d := range diags {
			t.Logf("unexpected diag: %s @%d:%d", d.Message, d.Pos.Line, d.Pos.Column)
		}
		t.FailNow()
	}

	// Structure: member access of a known-but-untyped field should error.
	_, badDiags := AnalyzeSource("main.rn", `
main() => {
  el := document.createElement("div")
  el.noSuchMethod()
}
`)
	if len(badDiags) == 0 {
		t.Fatalf("expected diagnostic for noSuchMethod, got %#v", badDiags)
	}
	found := false
	for _, d := range badDiags {
		if strings.Contains(d.Message, "noSuchMethod") {
			found = true
		}
	}
	if !found {
		t.Errorf("diagnostics = %#v, want noSuchMethod error", badDiags)
	}

	// Global functions: fetch is ambient and callable.
	prog, fetchDiags := AnalyzeSource("main.rn", `
main() => {
  fetch("/api")
}
`)
	if len(fetchDiags) != 0 {
		t.Fatalf("fetch call produced unexpected diagnostics: %#v", fetchDiags)
	}
	if prog.Info.Functions["fetch"] == nil {
		t.Error("fetch ambient function not registered in Info")
	}
}

// TestDOMShadowedByUser wins the test contract: user-defined structs shadow
// ambient DOM names so local stub types keep full control.
func TestDOMShadowedByUser(t *testing.T) {
	if checker.LocateDOMLib() == "" {
		t.Skip("lib.dom.d.ts not available on this machine")
	}
	src := `
+ Document: {
  title: String
}

main() => {
  doc := Document { title: "x" }
  doc.title
}
`
	_, resDiags := AnalyzeSource("main.rn", src)
	for _, d := range resDiags {
		if strings.Contains(d.Message, "duplicate type") {
			t.Errorf("duplicate type despite explicit user struct: %s", d.Message)
		}
	}
}
