package tscodegen

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/parser"
)

func TestGenerateCounterDOMProgram(t *testing.T) {
	src := `render() -> HTMLElement => {
  count $= 0

  <div>
    <h1>Counter Example</h1>
    <p>Count: {count}</p>
    <button @click={count++}>Click Me</button>
  </div>
}
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	info, diags := checker.Check(file)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
	got, err := Generate(file, info)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	wantParts := []string{
		`function __render(): HTMLElement`,
		`const __count = runeSignal(0);`,
		`document.createElement("div")`,
		`document.createElement("h1")`,
		`document.createTextNode("Counter Example")`,
		`document.createTextNode("Count: ")`,
		`document.createTextNode(String(__count.get()))`,
		`__count.watch(() => { __text`,
		`.addEventListener("click", () => { __count.set(__count.get() + 1); });`,
		`export { __render as render };`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
}
