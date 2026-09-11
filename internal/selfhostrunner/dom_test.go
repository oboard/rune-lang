package selfhostrunner

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/compiler"
)

// selfhostCheckErrors runs the self-hosted checker over source and returns its
// error messages (or nil when it reports ok). This drives the canonical
// `rune check` / LSP selfhost precheck path, so DOM event typing is asserted
// where it actually ships.
func selfhostCheckErrors(t *testing.T, source string) []string {
	t.Helper()
	compilerPath := filepath.Join(repoRootForTest(t), "selfhost", "compiler", "compiler.rn")
	program := "@" + quoteRuneString(compilerPath) + "\n\nmain() => {\n  source := `" + source + "`\n  result := checkSource(source)\n  @io.println(result.ok.toString())\n  result.errors.each((e) => @io.println(e))\n}\n"
	prog, diags := compiler.AnalyzeSource("selfhost_check_errors.rn", program)
	if len(diags) > 0 {
		t.Fatalf("AnalyzeSource diagnostics = %v\nprogram:\n%s", diags, program)
	}
	result := RunMainIR(prog.IR)
	if result.Err != nil {
		t.Fatalf("RunMainIR error = %v, output = %q", result.Err, result.Output)
	}
	lines := strings.Split(strings.TrimRight(result.Output, "\n"), "\n")
	if len(lines) == 0 {
		t.Fatalf("no output from checkSource driver")
	}
	if strings.TrimSpace(lines[0]) != "true" {
		// ok=false: remaining lines are the error messages
		return lines[1:]
	}
	return nil
}

func TestSelfhostDOMEventParamTyping(t *testing.T) {
	// A keydown handler's `e` must resolve to KeyboardEvent so `.key` checks,
	// while an unknown member still reports a field error.
	valid := `main() => {
  <input @keydown={(e) => e.key == "Enter"} />
}
`
	if errs := selfhostCheckErrors(t, valid); errs != nil {
		t.Fatalf("keydown handler produced errors: %v", errs)
	}

	bad := `main() => {
  <input @keydown={(e) => e.noSuchThing} />
}
`
	errs := selfhostCheckErrors(t, bad)
	if errs == nil {
		t.Fatalf("expected a field error for e.noSuchThing, got ok")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e, "noSuchThing") && strings.Contains(e, "KeyboardEvent") {
			found = true
		}
	}
	if !found {
		t.Fatalf("errors = %v, want 'KeyboardEvent has no field noSuchThing'", errs)
	}
}

func TestSelfhostDOMEventTypeSpecificity(t *testing.T) {
	// click -> PointerEvent (not plain Event) per lib.dom.d.ts; keydown ->
	// KeyboardEvent. Each resolves its own members without cross-talk.
	src := `main() => {
  <button @click={(e) => e.clientX} />
  <input @keydown={(e) => e.code} />
}
`
	if errs := selfhostCheckErrors(t, src); errs != nil {
		t.Fatalf("click/keydown handlers produced errors: %v", errs)
	}
}

func TestSelfhostDOMEventTargetNarrowing(t *testing.T) {
	// React ChangeEvent<T> parity: on form-control tags the handler's e.target
	// narrows to the element interface, so e.target.value / .checked resolve.
	valid := `main() => {
  <input @change={(e) => e.target.value} />
  <input @change={(e) => e.target.checked} />
  <input @change={(e) => e.target.value = ""} />
  <button @click={(e) => e.target.value} />
}
`
	if errs := selfhostCheckErrors(t, valid); errs != nil {
		t.Fatalf("narrowed form-control handlers produced errors: %v", errs)
	}

	// A non-form-control tag keeps vanilla EventTarget: e.target.value must
	// still be an error (div has no .value in lib.dom.d.ts).
	nonForm := `main() => {
  <div @click={(e) => e.target.value} />
}
`
	errs := selfhostCheckErrors(t, nonForm)
	if errs == nil {
		t.Fatal("expected EventTarget has no field value for a <div> handler, got ok")
	}
	if !errorsContain(errs, "EventTarget") || !errorsContain(errs, "value") {
		t.Fatalf("errors = %v, want EventTarget/value", errs)
	}

	// The narrowed target is precise: an unknown member on HTMLInputElement
	// still reports a field error (naming the narrowed element type).
	badMember := `main() => {
  <input @change={(e) => e.target.noSuchThing} />
}
`
	errs = selfhostCheckErrors(t, badMember)
	if errs == nil {
		t.Fatal("expected HTMLInputElement has no field noSuchThing, got ok")
	}
	if !errorsContain(errs, "HTMLInputElement") || !errorsContain(errs, "noSuchThing") {
		t.Fatalf("errors = %v, want HTMLInputElement/noSuchThing", errs)
	}
}

func errorsContain(errs []string, want string) bool {
	for _, e := range errs {
		if strings.Contains(e, want) {
			return true
		}
	}
	return false
}