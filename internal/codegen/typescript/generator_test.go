package tscodegen

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/parser"
)

func TestGenerateCounterDOMProgram(t *testing.T) {
	src := `+ render() -> HTMLElement => {
  $count := 0

  <div>
    <h1>Counter Example</h1>
    <p>Count: {$count}</p>
    <button @click={$count++}>Click Me</button>
  </div>
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`function render(): HTMLElement`,
		`const count = runeSignal(0);`,
		`document.createElement("div")`,
		`document.createElement("h1")`,
		`document.createTextNode("Counter Example")`,
		`document.createTextNode("Count: ")`,
		`document.createTextNode(String(count.get()))`,
		`runeWatch(count, () => { __text`,
		`.addEventListener("click", () => { count.set(count.get() + 1); });`,
		`export { render as render };`,
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

func TestGenerateGenericTraitConstraintFunction(t *testing.T) {
	src := `add[T: Number](a: T, b: T) -> T => a + b

main() => @io.println(add(1, 2))
`
	got := generateForTest(t, src)
	wantParts := []string{
		`function add<T extends number>(a: T, b: T): T`,
		`return (a + b) as T`,
		`console.log(add(1, 2));`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateExportsPublicTypesEnumsAndConstants(t *testing.T) {
	src := `+ User: {
  name: String
}

+ Status: {
  Ready = 1
  Done = 2
}

+ answer := 42

+ add(a: Int, b: Int) -> Int => a + b
`
	got := generateForTest(t, src)
	wantParts := []string{
		`type User = {`,
		`const Status = {`,
		`const answer: number = 42;`,
		`function add(a: number, b: number): number`,
		`export type { User as User, Status as Status };`,
		`export { Status as Status, answer as answer, add as add };`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
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
		`function HelloWorld(): CustomElementConstructor`,
		`return class extends HTMLElement`,
		`connectedCallback(): void`,
		`const __root`,
		`document.createElement("div")`,
		`document.createTextNode("hello world")`,
		`function render(): HTMLElement`,
		`document.createElement(runeDefineWebComponent("HelloWorld", HelloWorld))`,
		`export { HelloWorld as HelloWorld, render as render };`,
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

func TestGenerateWebComponentTagPassesAttributeParams(t *testing.T) {
	src := `+ HelloWorld(text: String) -> WebComponent => {
  <div>{text}</div>
}

+ render() -> HTMLElement => {
  <div>
    <HelloWorld text="hello" />
  </div>
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`function HelloWorld(text: string): CustomElementConstructor`,
		`document.createElement(runeDefineWebComponent("HelloWorld", () => HelloWorld("hello")))`,
		`.setAttribute("text", String("hello"));`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, `document.createElement("HelloWorld")`) {
		t.Fatalf("generated TypeScript created an unregistered component tag:\n%s", got)
	}
}

func TestGeneratePrivateWebComponentTagUsesSourceName(t *testing.T) {
	src := `HelloWorld() -> WebComponent => {
  <div>hello world</div>
}

render() -> HTMLElement => {
  <div>
    <HelloWorld />
  </div>
}
`
	got := generateForTestWithSourcePath(t, "playground.rn", src)
	wantParts := []string{
		`function HelloWorld`,
		`(): CustomElementConstructor`,
		`document.createElement(runeDefineWebComponent("HelloWorld", HelloWorld))`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, `__rune_private_`) {
		t.Fatalf("generated TypeScript leaked private link prefix:\n%s", got)
	}
}

func TestGeneratePrivateLinkNameWithPathUsesSourceName(t *testing.T) {
	src := `+ greeting(name: String) -> String => privateGreeting(name)

privateGreeting(name: String) -> String => "hello, " + name
`
	got := generateForTestWithSourcePath(t, "helper.rn", src)
	wantParts := []string{
		`function greeting(name: string): string`,
		`return privateGreeting(name);`,
		`function privateGreeting(name: string): string`,
		`export { greeting as greeting };`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, `__rune_private_`) {
		t.Fatalf("generated TypeScript leaked private link prefix:\n%s", got)
	}
}

func TestGeneratePatternPredicateRange(t *testing.T) {
	src := `isDigit(ch: Char) -> Bool => ('0'..='9')
`
	got := generateForTest(t, src)
	want := `if ((ch >= "0" && ch <= "9")) {`
	if !strings.Contains(got, want) {
		t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
	}
}

func TestGenerateArrayAliasAndOpenRangePatterns(t *testing.T) {
	src := `score(values: Array[Int]) -> Int => values {
  [head, ..rest, tail] @ whole => head + tail + rest.length() + whole.length()
  [] => 0
  _ => 1
}

sign(value: Int) -> Int => value {
  _..<0 => -1
  0 => 0
  1..<_ => 1
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`const __match1 = values;`,
		`if (__match1.length >= 2 && true && true)`,
		`const head = __match1[0];`,
		`const tail = __match1[__match1.length - 1];`,
		`const rest = __match1.slice(1, __match1.length - 1);`,
		`const whole = __match1;`,
		`else if (__match1.length === 0)`,
		`if ((value < 0))`,
		`else if ((value >= 1))`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
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
		`if ((typeName === "") || (typeName === "Void")) {`,
		`return "void";`,
		`else if ((typeName === "Int") || (typeName === "Double")) {`,
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
		`let ` + mangleIdent("增量📈") + ` = 1;`,
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
	src := "label(count: Int, ch: Char) -> String => `count \\(count) char \\(ch)`\n"
	got := generateForTest(t, src)
	want := "return `count ${count} char ${ch}`;"
	if !strings.Contains(got, want) {
		t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
	}

	got = generateForTest(t, "message(name: String) -> String => `hello\n\\(name)`\n")
	want = "return `hello\\n${name}`;"
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
		`const x = __match`,
		`const yy = __match`,
		`return x + yy;`,
		`.has(__key`,
		`const value = __match`,
		`return value;`,
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
		`list.map((item: string): HTMLElement =>`,
		`for (const __child`,
		`.appendChild(__child`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "String(list.map") {
		t.Fatalf("generated TypeScript stringifies element array:\n%s", got)
	}
}

func TestGenerateReactiveElementArrayChild(t *testing.T) {
	src := `render() => {
  $list := ["Item 1", "Item 2", "Item 3"]

  <ul>
    {$list.map((item) => (
        <li>{item}</li>
    ))}
    <button @click={$list.push("New Item")}>Add Item</button>
  </ul>
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`function runeReactiveArray<T>(initial: T[]): RuneSignal<T[]>`,
		`const list = runeReactiveArray(["Item 1", "Item 2", "Item 3"]);`,
		`const __start`,
		`const __render`,
		`runeWatch(list, __render`,
		`.insertBefore(__child`,
		`.addEventListener("click", () => { list.mutate((__value) => __value.push("New Item")); });`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "Proxy") {
		t.Fatalf("generated TypeScript should not use Proxy for reactive arrays:\n%s", got)
	}
	if strings.Contains(got, "String(list.map") {
		t.Fatalf("generated TypeScript stringifies reactive element array:\n%s", got)
	}
}

func TestGenerateArraySpread(t *testing.T) {
	src := `main() => {
  items := ["Item 1"]
  next := [..items, "New Item"]
  @io.println(next.length())
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`let next = [...items, "New Item"];`,
		`console.log(next.length);`,
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
		`values.reduceRight`,
		`(accumulator: number, value: number): number => accumulator + value`,
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
		`type Status = number;`,
		`const Status = {`,
		`Completed: 0,`,
		`Fail: 1,`,
		`function statusText(status: Status): string`,
		`status === Status.Completed`,
		`status === Status.Fail`,
		`function fallback(flag: boolean): Status`,
		`return 0 as Status;`,
		`let status = Status.Completed;`,
		`console.log(statusText(status));`,
		`console.log(fallback(false));`,
		`let Status = {Completed: 42};`,
		`console.log(Status.Completed);`,
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
  @io.println(@regex.escape("a+b?"))
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`let re = /rune\s+(\d+)/ig;`,
		`let built = new RegExp("\\d+", "g");`,
		`"Rune 123 rune 456".match(re)`,
		`"a1 b22".replaceAll(__regex.global ? __regex : new RegExp(__regex.source, __regex.flags + "g"), "[$1]"))(built)`,
		`__value.replace(/[\\^$.*+?()[\]{}|]/g, "\\$&")`,
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
		`let scores = new Map<string, number>();`,
		`scores.set("rune", 10);`,
		`((__map, __key) => __map.has(__key) ? __map.get(__key)! : 0)(scores, "rune")`,
		`let seen = new Set<string>();`,
		`seen.add("rune");`,
		`console.log(seen.has("rune"));`,
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
		`let scores = new Map<string, number>([["a", 1], ["b", 2]]);`,
		`.has(`,
		` ?? 0`,
		`scores.set("b", 3);`,
		` ?? 7`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateEmptyMapFieldFromStructContext(t *testing.T) {
	src := `Env: {
  bindings: Map[Int, String]
}

emptyEnv() -> Env => {
  bindings: {}
}
`
	got := generateForTest(t, src)
	want := `bindings: new Map<number, string>()`
	if !strings.Contains(got, want) {
		t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
	}
}

func TestGenerateArrayEachAvoidsUserIndexShadow(t *testing.T) {
	src := `main() => {
  index := 0
  [1].each((value, position, array) => {
    index = position + value + array.length()
  })
  @io.println(index)
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`let index = 0;`,
		`for (const [__arrayIndex`,
		`index = position + value + array.length`,
		`console.log(index);`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, `for (const [__index,`) {
		t.Fatalf("generated TypeScript shadowed user index:\n%s", got)
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
		`let bytes = new DataView(new ArrayBuffer(16));`,
		`bytes.setUint8(0, __value); return __value; })((255 & 0xff))`,
		`bytes.setInt16(1, __value, true); return __value; })`,
		`bytes.setBigUint64(4, __value, false); return __value; })`,
		`bytes.setFloat32(12, __value, true); return __value; })(Math.fround(1.5))`,
		`console.log(bytes.getUint8(0));`,
		`console.log(bytes.getInt16(1, true));`,
		`console.log(bytes.getBigUint64(4, false));`,
		`console.log(bytes.getFloat32(12, true));`,
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
		`runeCompressUnbrotliText(brotli)`,
		`runeCompressZstdText(brotliText)`,
		`runeCompressUnzstdText(zstd)`,
		`console.log(text);`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateJSONStringifyObject(t *testing.T) {
	src := `#json.object
User: {
  #json.name("display_name")
  name: String
  #json.ignore
  password: String
  age: Int
}

main() => {
  user := User { name: "Ada", password: "secret", age: 36 }
	  obj := {
	    name: "Rune",
	    user: user,
	    tags: ["compiler", "json"],
	    greet() => @io.println(.name)
	  }

	  @io.println(@json.stringify(obj))
	  @io.println(@json.stringify({
	    name: "Direct",
	    greet() => @io.println("skip")
	  }))
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`JSON.stringify(((rune_json_value) => ({ name: rune_json_value.name`,
		`user: ((rune_json_value) => ({ display_name: rune_json_value.name, age: rune_json_value.age }))(rune_json_value.user)`,
		`tags: rune_json_value.tags`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, `"greet"`) {
		t.Fatalf("generated TypeScript should omit function fields:\n%s", got)
	}
	if strings.Contains(got, "password: rune_json_value.password") {
		t.Fatalf("generated TypeScript should omit ignored JSON field:\n%s", got)
	}
}

func TestGenerateJSONParseObject(t *testing.T) {
	src := `#json.object
User: {
  #json.name("display_name")
  name: String
  #json.ignore
  password: String
  scores: Array[Int]
}

main() => {
  user := @json.parse("{\"display_name\":\"Ada\",\"password\":\"drop\",\"scores\":[3,5]}") : User
  @io.println(user.name)
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`JSON.parse("{\"display_name\":\"Ada\",\"password\":\"drop\",\"scores\":[3,5]}")`,
		`...({ name: "", password: "", scores: [] })`,
		`name: rune_json_raw["display_name"]`,
		`scores: rune_json_raw["scores"].map`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, `password: rune_json_raw`) {
		t.Fatalf("generated TypeScript should leave ignored fields at their zero value:\n%s", got)
	}
}

func TestGenerateSymbolIntrinsicProgram(t *testing.T) {
	src := `main() => {
  @io.println(@symbol.toString(@symbol.create("x")))
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`String(Symbol("x"))`,
		`console.log`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateNetUsesDataViewBytes(t *testing.T) {
	src := `~ main() => {
  listener := @net.listen("127.0.0.1:0")?
  conn := listener.accept()?
  data := conn.read(1024)?
  conn.write(data)?
  conn.close()?
  listener.close()?
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`function runeNetConnectionRead(connection: RuneTCPConnection, length: number): Promise<RuneResult<DataView, RuneError>>`,
		`resolve(runeOk<DataView, RuneError>(runeDataViewFromBytes(Array.from(chunk.slice(0, length)))))`,
		`function runeNetConnectionWrite(connection: RuneTCPConnection, data: DataView): Promise<RuneResult<number, RuneError>>`,
		`const bytes = new Uint8Array(data.buffer, data.byteOffset, data.byteLength);`,
		`connection.socket.write(bytes`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
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
		`function test(count: number): Promise<void>`,
		`return runeGo(async (): Promise<void> => {`,
		`count.toString()`,
		`test(1);`,
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
		`let freedom = {"return": 0, func: 1, def: 2};`,
		`console.log(freedom["return"]);`,
		`console.log(freedom.func);`,
		`console.log(freedom.def);`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateSignalAssignmentExpression(t *testing.T) {
	src := `render() => {
  $list := ["Item 1"]
  <button @click={$list = [..$list, "New Item"]}>Add Item</button>
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`const list = runeReactiveArray(["Item 1"]);`,
		`.addEventListener("click", () => { list.set([...list.get(), "New Item"]); });`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateSignalObjectInitializer(t *testing.T) {
	src := `render() => {
  $state := {count: 0}
  <button @click={$state.count = $state.count + 1}>Add Item</button>
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`const state = runeReactiveObject({count: 0});`,
		`.addEventListener("click", () => { state.mutate((__value) => (__value.count = state.get().count + 1)); });`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateSignalEffectScope(t *testing.T) {
	src := `main() => {
  $count := 0
  $double := $count * 2
  {
    $count + $double
  }
  $count = $count + 1
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`const __effect1 = () => {`,
		`count.get() + double.get();`,
		`let __effectPending2 = false;`,
		`const __scheduleEffect3 = () => {`,
		`runeScheduleEffect(() => { __effectPending2 = false; __effect1(); });`,
		`__effect1();`,
		`runeWatch(double, __scheduleEffect3);`,
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
	    x: 20,
	    y: 22
	  }
  { x, y } := point
  @io.println(x + y)
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`const { x: x, y: y } = point;`,
		`console.log(x + y);`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
}

func TestGeneratePayloadEnumPatterns(t *testing.T) {
	src := `Expr: {
  Lit(value: Int)
  Add(left: Expr, right: Expr)
}

eval(expr: Expr) -> Int => expr {
  Add(Lit(left), Lit(0)) => left
  Add(left, right) => eval(left) + eval(right)
  Lit(value) => value
}

main() => @io.println(eval(Add(Lit(2), Lit(0))))
`
	got := generateForTest(t, src)
	wantParts := []string{
		`type Expr = { tag: number; payload: any[] };`,
		`{ tag: Expr.Add, payload: [`,
		`.tag === Expr.Add`,
		`(__match`,
		`.payload[0] as Expr)`,
		`.payload[0] as number`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateMapLikePatternCachesGet(t *testing.T) {
	src := `Lookup: {
  entries: Map[String, Int]

  get(key: String) -> Int? => .entries[key]
}

score(values: Lookup) -> Int => values {
  { "b"? : null, "a": x, .. } => x
  { "b"? : b, .. } => b ?? 0
}
`
	got := generateForTest(t, src)
	wantParts := []string{
		`new Map<string, any>()`,
		`const __mapGet`,
		`Lookup_get(__match`,
		`const x = __mapGet`,
		`const b = __mapGet`,
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

func generateForTestWithSourcePath(t *testing.T, path string, src string) string {
	t.Helper()
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	for _, fn := range file.Functions {
		fn.SourcePath = path
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

func TestMangleIdentRemovesPrefixAndEscapesTypeScriptKeywords(t *testing.T) {
	for input, want := range map[string]string{"answer": "answer", "class": "class_", "1st": "rune_1st"} {
		if got := mangleIdent(input); got != want {
			t.Errorf("mangleIdent(%q) = %q, want %q", input, got, want)
		}
	}
}
