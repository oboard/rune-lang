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
	got := generateForTest(t, src)
	wantParts := []string{
		`function __render(): HTMLElement`,
		`const __count = runeSignal(0);`,
		`document.createElement("div")`,
		`document.createElement("h1")`,
		`document.createTextNode("Counter Example")`,
		`document.createTextNode("Count: ")`,
		`document.createTextNode(String(__count.get()))`,
		`runeWatch(__count, () => { __text`,
		`.addEventListener("click", () => { __count.set(__count.get() + 1); });`,
		`export { __render as render };`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateElementArrayChild(t *testing.T) {
	src := `render() => {
  list := ["Item 1", "Item 2", "Item 3"]

  <ul>
    {list.map((item) => (
        <li>{item}</li>
    ))}
  </ul>
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`const __children`,
		`__list.map((__item: string): HTMLElement =>`,
		`for (const __child`,
		`.appendChild(__child`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "String(__list.map") {
		t.Fatalf("generated TypeScript stringifies element array:\n%s", got)
	}
}

func TestGenerateReactiveElementArrayChild(t *testing.T) {
	src := `render() => {
  list := $["Item 1", "Item 2", "Item 3"]

  <ul>
    {list.map((item) => (
        <li>{item}</li>
    ))}
    <button @click={list.push("New Item")}>Add Item</button>
  </ul>
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`function runeReactiveArray<T>(initial: T[]): T[]`,
		`const __list = runeReactiveArray(["Item 1", "Item 2", "Item 3"]);`,
		`const __start`,
		`const __render`,
		`runeWatch(__list, __render`,
		`.insertBefore(__child`,
		`.addEventListener("click", () => { __list.push("New Item"); });`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "String(__list.map") {
		t.Fatalf("generated TypeScript stringifies reactive element array:\n%s", got)
	}
}

func TestGenerateArraySpread(t *testing.T) {
	src := `main() => {
  items := ["Item 1"]
  next := [...items, "New Item"]
  @io.println(next.length())
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`const __next = [...__items, "New Item"];`,
		`console.log(__next.length);`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateSignalAssignmentExpression(t *testing.T) {
	src := `render() => {
  list $= ["Item 1"]
  <button @click={list = [...list, "New Item"]}>Add Item</button>
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`const __list = runeSignal(["Item 1"]);`,
		`.addEventListener("click", () => { __list.set([...__list.get(), "New Item"]); });`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
}

func generateForTest(t *testing.T, src string) string {
	t.Helper()
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
	return got
}
