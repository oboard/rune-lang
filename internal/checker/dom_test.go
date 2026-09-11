package checker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/parser"
)

func TestDOMInterfacesResolved(t *testing.T) {
	if locateDOMLib() == "" {
		t.Skip("lib.dom.d.ts not available on this machine")
	}
	ifaces := []string{
		"Document", "Element", "HTMLElement", "HTMLDivElement", "HTMLInputElement",
		"KeyboardEvent", "MouseEvent", "Event", "EventTarget", "UIEvent",
		"Window", "DocumentFragment", "Text",
	}
	src := `
main() => {
	document.createElement("div")
}
`
	file := parseSource(t, src)
	info, diags := CheckWithStdlibForPath(file, nil, "test.rn")
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	for _, name := range ifaces {
		if info.Types[name] == nil {
			t.Errorf("checker missing DOM interface %q", name)
		}
	}
	if info.valuesByName["document"] == nil {
		t.Error("document global not registered")
	}
	if info.valuesByName["window"] == nil {
		t.Error("window global not registered")
	}
	fetched := info.functionsByName["fetch"]
	if len(fetched) == 0 {
		t.Error("fetch global function not registered")
	}
}

func TestDOMInheritance(t *testing.T) {
	if locateDOMLib() == "" {
		t.Skip("lib.dom.d.ts not available on this machine")
	}
	src := `
main() => {
	el := document.createElement("div")
	el.classList.add("active")
	el.querySelector(".foo")
	el.appendChild(document.createElement("span"))
}
`
	file := parseSource(t, src)
	_, diags := CheckWithStdlibForPath(file, nil, "test.rn")
	if len(diags) != 0 {
		t.Logf("element chain: %#v", diags)
		t.Fail()
	}
	src2 := `
main() => {
	el := document.createElement("div")
	el.noSuchMethod()
}
`
	file2 := parseSource(t, src2)
	info, diags2 := CheckWithStdlibForPath(file2, nil, "test.rn")
	if len(diags2) == 0 {
		// Check what type createElement resolved to
		// Find the call expression in the AST
		if declared := info.Types["HTMLElement"]; declared != nil {
			m := declared.Methods["noSuchMethod"]
			t.Logf("noSuchMethod on HTMLElement: %v", m)
		}
		// Check Document.createElement return
		docInfo := info.Types["Document"]
		if docInfo != nil {
			m := docInfo.Methods["createElement"]
			if m != nil {
				t.Logf("createElement Return: %s  MinRequired: %d  Params len: %d", m.Return, m.MinRequired, len(m.Params))
			}
		}
		t.Error("expected type error for unknown method")
	} else {
		for _, d := range diags2 {
			t.Logf("diag: %s", d.Message)
		}
	}
}

func TestDOMGlobals(t *testing.T) {
	if locateDOMLib() == "" {
		t.Skip("lib.dom.d.ts not available on this machine")
	}
	src := `
main() => {
	fetch("/api")
	setTimeout(() => {}, 100.0)
	clearTimeout(0)
	alert("hi")
}
`
	file := parseSource(t, src)
	_, diags := CheckWithStdlibForPath(file, nil, "test.rn")
	for _, d := range diags {
		t.Logf("diag: %s", d.Message)
	}
}

func parseSource(t *testing.T, src string) *ast.File {
	t.Helper()
	f, errs := parser.Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse error: %v", errs)
	}
	return f
}

func TestDOMEventHandlerParamTyping(t *testing.T) {
	if locateDOMLib() == "" {
		t.Skip("lib.dom.d.ts not available on this machine")
	}
	src := `
main() => {
  <div>
    <input @keydown={(e) => e.key == "Enter"} />
    <button @click={(e) => e.clientX} />
  </div>
}
`
	file := parseSource(t, src)
	info, diags := CheckWithStdlibForPath(file, nil, "test.rn")
	for _, d := range diags {
		t.Logf("diag: %s", d.Message)
	}
	// The keydown handler's `e` must resolve to KeyboardEvent (so `.key`
	// type-checks), and the click handler's `e` to PointerEvent (`.clientX`).
	if kb := info.Types["KeyboardEvent"]; kb == nil {
		t.Error("KeyboardEvent interface absent from ambient DOM")
	} else if kb.ByName["key"].Name == "" {
		t.Error("KeyboardEvent.key member absent")
	}
	if pe := info.Types["PointerEvent"]; pe == nil {
		t.Error("PointerEvent interface absent from ambient DOM")
	} else if pe.ByName["clientX"].Name == "" {
		t.Error("PointerEvent.clientX member absent")
	}
}

func TestDOMEventHandlerWrongMemberErrors(t *testing.T) {
	if locateDOMLib() == "" {
		t.Skip("lib.dom.d.ts not available on this machine")
	}
	src := `
main() => {
  <input @keydown={(e) => e.noSuchThing} />
}
`
	file := parseSource(t, src)
	_, diags := CheckWithStdlibForPath(file, nil, "test.rn")
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "noSuchThing") {
			found = true
		}
	}
	if !found {
		t.Fatalf("diagnostics = %#v, want noSuchThing error on KeyboardEvent", diags)
	}
}

func TestDOMLibPath(t *testing.T) {
	path := locateDOMLib()
	if path == "" {
		t.Skip("lib.dom.d.ts missing")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("locateDOMLib returned %q but stat failed: %v", path, err)
	}
	abs, _ := filepath.Abs(path)
	if _, err := os.Stat(abs); err == nil {
		t.Logf("locateDOMLib: %s", path)
	}
}
