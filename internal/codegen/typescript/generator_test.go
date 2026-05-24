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

func TestGenerateEnumProgram(t *testing.T) {
	src := `Status: {
  Completed = 0
  Fail = 1
}

Container: {
  Completed: Int
}

statusText(status: Status) -> String => status {
  Status.Completed => "completed"
  Status.Fail => "fail"
  _ => "unknown"
}

fallback(flag: Bool) -> Status => flag {
  true => Status.Fail
}

main() => {
  status := Status.Completed
  @io.println(statusText(status))
  @io.println(fallback(false))
  Status := Container { Completed: 42 }
  @io.println(Status.Completed)
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`type __Status = number;`,
		`const __Status = {`,
		`Completed: 0,`,
		`Fail: 1,`,
		`function __statusText(__status: __Status): string`,
		`__status === __Status.Completed`,
		`__status === __Status.Fail`,
		`function __fallback(__flag: boolean): __Status`,
		`return 0 as __Status;`,
		`const __status = __Status.Completed;`,
		`console.log(__statusText(__status));`,
		`console.log(__fallback(false));`,
		`const __Status = {Completed: 42};`,
		`console.log(__Status.Completed);`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateRegexProgram(t *testing.T) {
	src := `main() => {
  re := /rune\s+(\d+)/ig
  built := @regex.new("\\d+", "g")
  @io.println(re.match("Rune 123 rune 456"))
  @io.println(built.replaceAll("a1 b22", "[$1]"))
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`const __re = /rune\s+(\d+)/ig;`,
		`const __built = new RegExp("\\d+", "g");`,
		`"Rune 123 rune 456".match(__re)`,
		`"a1 b22".replaceAll(__regex.global ? __regex : new RegExp(__regex.source, __regex.flags + "g"), "[$1]"))(__built)`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateMapIntrinsicProgram(t *testing.T) {
	src := `main() => {
  scores := @map.newMap("", 0)
  scores.set("rune", 10)
  @io.println(scores.getOr("rune", 0))

  seen := @map.newSet("")
  seen.add("rune")
  @io.println(seen.has("rune"))
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`const __scores = new Map<string, number>();`,
		`__scores.set("rune", 10);`,
		`((__map, __key) => __map.has(__key) ? __map.get(__key)! : 0)(__scores, "rune")`,
		`const __seen = new Set<string>();`,
		`__seen.add("rune");`,
		`console.log(__seen.has("rune"));`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "@map") {
		t.Fatalf("generated TypeScript leaked @map intrinsic:\n%s", got)
	}
}

func TestGenerateJSONStringifyObject(t *testing.T) {
	src := `User: {
  name: String
  age: Int
}

main() => {
  user := User { name: "Ada", age: 36 }
  obj := {
    name: "Rune"
    user: user
    tags: ["compiler", "json"]
    greet() => @io.println(.name)
  }

  @io.println(@json.stringify(obj))
  @io.println(@json.stringify({
    name: "Direct"
    greet() => @io.println("skip")
  }))
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`JSON.stringify(((__rune_json_value) => ({ name: __rune_json_value.name`,
		`user: ((__rune_json_value) => ({ name: __rune_json_value.name, age: __rune_json_value.age }))(__rune_json_value.user)`,
		`tags: __rune_json_value.tags`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, `"greet"`) {
		t.Fatalf("generated TypeScript should omit function fields:\n%s", got)
	}
}

func TestGenerateUnsupportedModuleIntrinsicError(t *testing.T) {
	src := `main() => {
  @io.println(@symbol.toString(@symbol.create("x")))
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
	if err == nil {
		t.Fatalf("Generate() expected unsupported intrinsic error:\n%s", got)
	}
	if !strings.Contains(err.Error(), "TypeScript backend does not support intrinsic symbol.toString") ||
		!strings.Contains(err.Error(), "TypeScript backend does not support intrinsic symbol.create") {
		t.Fatalf("Generate() error = %v, want symbol intrinsic errors", err)
	}
	if strings.Contains(got, "@symbol") {
		t.Fatalf("generated TypeScript leaked @symbol intrinsic:\n%s", got)
	}
}

func TestGenerateKeywordObjectFields(t *testing.T) {
	src := `Println: {
  return: Int
  func: Int
  def: Int
}

main() => {
  freedom := Println { return: 0, func: 1, def: 2 }
  @io.println(freedom.return)
  @io.println(freedom.func)
  @io.println(freedom.def)
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`"return": number;`,
		`func: number;`,
		`def: number;`,
		`const __freedom = {"return": 0, func: 1, def: 2};`,
		`console.log(__freedom["return"]);`,
		`console.log(__freedom.func);`,
		`console.log(__freedom.def);`,
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
