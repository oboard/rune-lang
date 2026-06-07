package tscodegen

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/parser"
)

func TestGenerateCounterDOMProgram(t *testing.T) {
	src := `+ render() -> HTMLElement => {
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
	if strings.Contains(got, "RuneResult") || strings.Contains(got, "RuneError") || strings.Contains(got, "runeOk") || strings.Contains(got, "runeErr") {
		t.Fatalf("generated TypeScript should not include Result/Error runtime:\n%s", got)
	}
}

func TestGenerateWebComponentFromXMLLiteral(t *testing.T) {
	src := `+ HelloWorld() -> WebComponent => {
  <div>hello world</div>
}

+ render() -> HTMLElement => {
  <div>
    <HelloWorld />
  </div>
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`function runeDefineWebComponent(name: string, factory: () => CustomElementConstructor): string`,
		`function __HelloWorld(): CustomElementConstructor`,
		`return class extends HTMLElement`,
		`connectedCallback(): void`,
		`const __root`,
		`document.createElement("div")`,
		`document.createTextNode("hello world")`,
		`function __render(): HTMLElement`,
		`document.createElement(runeDefineWebComponent("HelloWorld", __HelloWorld))`,
		`export { __HelloWorld as HelloWorld, __render as render };`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "function __HelloWorld(): HTMLElement") {
		t.Fatalf("generated TypeScript should return a WebComponent constructor:\n%s", got)
	}
}

func TestGeneratePatternPredicateRange(t *testing.T) {
	src := `isDigit(ch: Char) -> Bool => ('0'..='9')
`
	got := generateForTest(t, src)
	want := `if ((__ch >= "0" && __ch <= "9")) {`
	if !strings.Contains(got, want) {
		t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
	}
}

func TestGenerateOrPatternBlock(t *testing.T) {
	src := `tsType(typeName: String) -> String => {
  "" | "Void" => "void"
  "Int" | "Double" => "number"
  _ => typeName
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`if ((__typeName === "") || (__typeName === "Void")) {`,
		`return "void";`,
		`else if ((__typeName === "Int") || (__typeName === "Double")) {`,
		`return "number";`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateUnicodeIdentifiers(t *testing.T) {
	src := `+ 计算✅(数值🐉: Int) -> Int => {
  增量📈 := 1
  数值🐉 + 增量📈
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`function ` + mangleIdent("计算✅") + `(` + mangleIdent("数值🐉") + `: number): number`,
		`const ` + mangleIdent("增量📈") + ` = 1;`,
		`return ` + mangleIdent("数值🐉") + ` + ` + mangleIdent("增量📈") + `;`,
		`export { ` + mangleIdent("计算✅") + ` as "计算✅" };`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateTemplateLiteral(t *testing.T) {
	src := "label(count: Int, ch: Char) -> String => `count ${count} char ${ch}`\n"
	got := generateForTest(t, src)
	want := "return `count ${__count} char ${__ch}`;"
	if !strings.Contains(got, want) {
		t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
	}

	got = generateForTest(t, "message(name: String) -> String => `hello\n${name}`\n")
	want = "return `hello\\n${__name}`;"
	if !strings.Contains(got, want) {
		t.Fatalf("generated TypeScript missing multiline template %q:\n%s", want, got)
	}
}

func TestGenerateDestructuringPatterns(t *testing.T) {
	src := `Point: {
  x: Int
  y: Int
}

pointScore(point: Point) -> Int => point {
  { x, y: yy, .. } => x + yy
  _ => 0
}

mapScore(values: Map[String, Int]) -> Int => values {
  { "a": value, .. } => value
  _ => 0
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`const __x = __match`,
		`const __yy = __match`,
		`return __x + __yy;`,
		`.has(__key`,
		`const __value = __match`,
		`return __value;`,
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

func TestGenerateArrayFoldr(t *testing.T) {
	src := `sum(values: Array[Int]) -> Int => values.foldr(0, (accumulator, value) => accumulator + value)
`
	got := generateForTest(t, src)
	wantParts := []string{
		`__values.reduceRight`,
		`(__accumulator: number, __value: number): number => __accumulator + __value`,
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
  scores := @map.new("", 0)
  scores.set("rune", 10)
  @io.println(scores.getOr("rune", 0))

  seen := @set.new("")
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

func TestGenerateMapLiteralProgram(t *testing.T) {
	src := `main() => {
  scores := {
    "a": 1,
    "b": 2
  }
  @io.println(scores["a"] ?? 0)
  scores["b"] = 3
  @io.println(scores["b"] ?? 0)
  @io.println(scores["missing"] ?? 7)
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`const __scores = new Map<string, number>([["a", 1], ["b", 2]]);`,
		`.has(`,
		` ?? 0`,
		`__scores.set("b", 3);`,
		` ?? 7`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateBytesIntrinsicProgram(t *testing.T) {
	src := `main() => {
  bytes := @bytes.new(16)
  bytes.setUint8(0, @uint8.fromInt(255))
  bytes.setInt16(1, @int16.fromInt(0 - 1234), true)
  bytes.setBigUint64(4, @uint64.fromInt(123456), false)
  bytes.setFloat32(12, @float.fromDouble(1.5), true)

  @io.println(bytes.getUint8(0))
  @io.println(bytes.getInt16(1, true))
  @io.println(bytes.getBigUint64(4, false))
  @io.println(bytes.getFloat32(12, true))
  @io.println(@uint.toInt(@uint.fromInt(16) >>> @uint.fromInt(2)))
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`const __bytes = new DataView(new ArrayBuffer(16));`,
		`__bytes.setUint8(0, __value); return __value; })((255 & 0xff))`,
		`__bytes.setInt16(1, __value, true); return __value; })`,
		`__bytes.setBigUint64(4, __value, false); return __value; })`,
		`__bytes.setFloat32(12, __value, true); return __value; })(Math.fround(1.5))`,
		`console.log(__bytes.getUint8(0));`,
		`console.log(__bytes.getInt16(1, true));`,
		`console.log(__bytes.getBigUint64(4, false));`,
		`console.log(__bytes.getFloat32(12, true));`,
		`console.log((16 >>> 0) >>> (2 >>> 0));`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateCompressIntrinsicProgram(t *testing.T) {
	src := `~ main() => {
  brotli := @compress.brotliText("hello")?
  brotliText := @compress.unbrotliText(brotli)?
  zstd := @compress.zstdText(brotliText)?
  text := @compress.unzstdText(zstd)?
  @io.println(text)
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`return runeCompressCall("brotliCompress", data);`,
		`return runeCompressCall("brotliDecompress", data);`,
		`return runeCompressCall("zstdCompress", data);`,
		`return runeCompressCall("zstdDecompress", data);`,
		`runeCompressBrotliText("hello")`,
		`runeCompressUnbrotliText(__brotli)`,
		`runeCompressZstdText(__brotliText)`,
		`runeCompressUnzstdText(__zstd)`,
		`console.log(__text);`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
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

func TestGenerateRoutineCallsUseTrackedPromises(t *testing.T) {
	src := `~ test(count: Int) => {
  @io.println("Hello World" + count.toString())
}

main() => {
  test(1)
  test(2)
  test(3)
  @io.println("Hello World")
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`const runeTasks: Promise<unknown>[] = [];`,
		`function runeGo<T>(work: () => T | Promise<T>): Promise<T>`,
		`async function runeWaitAll(): Promise<void>`,
		`function __test(__count: number): Promise<void>`,
		`return runeGo(async (): Promise<void> => {`,
		`__count.toString()`,
		`__test(1);`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
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

func TestGenerateObjectDestructureProgram(t *testing.T) {
	src := `Point: {
  x: Int
  y: Int
}

main() => {
  point := Point {
    x: 20
    y: 22
  }
  { x, y } := point
  @io.println(x + y)
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`const { x: __x, y: __y } = __point;`,
		`console.log(__x + __y);`,
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
