package gocodegen

import (
	goparser "go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/parser"
)

func TestGenerateFibProgram(t *testing.T) {
	src := `fib(n: Int) => {
  0 => 0
  1 => 1
  _ => fib(n - 1) + fib(n - 2)
}

main() => {
  @io.println(fib(10))
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
		`"fmt"`,
		`func __fib(__n int) int`,
		`case __n == 0:`,
		`return __fib(__n-1) + __fib(__n-2)`,
		`func __main()`,
		`fmt.Println(__fib(10))`,
		`func main()`,
		`__main()`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "runeResult") || strings.Contains(got, "runeError") {
		t.Fatalf("generated Go should not include Result/Error runtime:\n%s", got)
	}
}

func TestGenerateGenericTraitConstraintFunction(t *testing.T) {
	src := `add[T: Number](a: T, b: T) -> T => a + b

main() => @io.println(add(1, 2))
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
		`type runeNumber interface`,
		`func __add[__T runeNumber](__a __T, __b __T) __T`,
		`return __a + __b`,
		`fmt.Println(__add(1, 2))`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "main.go", got, 0); err != nil {
		t.Fatalf("generated Go parse error: %v\n%s", err, got)
	}
}

func TestGeneratePatternPredicateRange(t *testing.T) {
	src := `isDigit(ch: Char) -> Bool => ('0'..='9')
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
	want := `case (__ch >= '0' && __ch <= '9'):`
	if !strings.Contains(got, want) {
		t.Fatalf("generated Go missing %q:\n%s", want, got)
	}
}

func TestGenerateOrPatternBlock(t *testing.T) {
	src := `tsType(typeName: String) -> String => {
  "" | "Void" => "void"
  "Int" | "Double" => "number"
  _ => typeName
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
		`case (__typeName == "") || (__typeName == "Void"):`,
		`return "void"`,
		`case (__typeName == "Int") || (__typeName == "Double"):`,
		`return "number"`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateUnicodeIdentifiers(t *testing.T) {
	src := `计算✅(数值🐉: Int) -> Int => {
  增量📈 := 1
  数值🐉 + 增量📈
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
		`func ` + mangleIdent("计算✅") + `(` + mangleIdent("数值🐉") + ` int) int`,
		mangleIdent("增量📈") + ` := 1`,
		`return ` + mangleIdent("数值🐉") + ` + ` + mangleIdent("增量📈"),
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "unicode_identifiers.go", got, 0); err != nil {
		t.Fatalf("generated Go does not parse: %v\n%s", err, got)
	}
}

func TestGenerateTemplateLiteral(t *testing.T) {
	src := "label(count: Int, ch: Char) -> String => `count ${count} char ${ch}`\n"
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
		`"fmt"`,
		`func runeTemplateString(value any) string`,
		`return "count " + runeTemplateString(__count) + " char " + runeTemplateString(__ch)`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "template_literal.go", got, 0); err != nil {
		t.Fatalf("generated Go does not parse: %v\n%s", err, got)
	}

	file, parseErrs = parser.Parse("message(name: String) -> String => `hello\n${name}`\n")
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	info, diags = checker.Check(file)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
	got, err = Generate(file, info)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if want := `return "hello\n" + __name`; !strings.Contains(got, want) {
		t.Fatalf("generated Go missing multiline template %q:\n%s", want, got)
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
		`__x := __match`,
		`__yy := __match`,
		`return __x + __yy`,
		`func() bool { _, __ok`,
		`__value := __match`,
		`return __value`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateStructProgram(t *testing.T) {
	src := `User: {
  id: Int
  name: String
  age: Int

  isAdult() => .age >= 18
}

main() => {
  user := User {
    id: 1
    name: "oboard"
    age: 22
  }
  @io.println(user.name)
  @io.println(user.isAdult())
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
		`type __User struct`,
		`__id   int`,
		`__name string`,
		`__age  int`,
		`func (__this __User) __isAdult() bool`,
		`return __this.__age >= 18`,
		`__user := __User{__id: 1, __name: "oboard", __age: 22}`,
		`fmt.Println(__user.__name)`,
		`fmt.Println(__user.__isAdult())`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
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
		`__destructure1 := __point`,
		`__x := __destructure1.__x`,
		`__y := __destructure1.__y`,
		`fmt.Println(__x + __y)`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
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
		`type __Status int`,
		`__Status_Completed __Status = 0`,
		`__Status_Fail`,
		`= 1`,
		`func __statusText(__status __Status) string`,
		`case __status == __Status_Completed:`,
		`case __status == __Status_Fail:`,
		`func __fallback(__flag bool) __Status`,
		`return __Status(0)`,
		`__status := __Status_Completed`,
		`fmt.Println(__statusText(__status))`,
		`fmt.Println(__fallback(false))`,
		`__Status := __Container{__Completed: 42}`,
		`fmt.Println(__Status.__Completed)`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateInlineGoFFI(t *testing.T) {
	src := `@go.import("fmt")

isAdult(age: Int) -> Bool => @go.expr("$age >= 18")

main() => {
  name := "oboard"
  age := 22
  分数💯 := 42
  @go.stmt("fmt.Println($name)")
  @go.stmt("fmt.Println($age)")
  @go.stmt("fmt.Println($分数💯)")
  @io.println(isAdult(age))
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
		`"fmt"`,
		`func __isAdult(__age int) bool`,
		`return __age >= 18`,
		`__name := "oboard"`,
		`fmt.Println(__name)`,
		`fmt.Println(__age)`,
		mangleIdent("分数💯") + ` := 42`,
		`fmt.Println(` + mangleIdent("分数💯") + `)`,
		`fmt.Println(__isAdult(__age))`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateArraySpread(t *testing.T) {
	src := `main() => {
  items := ["Item 1"]
  next := [...items, "New Item"]
  @io.println(next.length())
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
		`__next := func() []string`,
		`out = append(out, __items...)`,
		`out = append(out, "New Item")`,
		`fmt.Println(len(__next))`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateArrayReduce(t *testing.T) {
	src := `sum(values: Array[Int]) -> Int => values.reduce(0, (accumulator, value) => accumulator + value)
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
		`range __array`,
		`return __accumulator + __value`,
		`return __result`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "array_reduce.go", got, 0); err != nil {
		t.Fatalf("generated Go does not parse: %v\n%s", err, got)
	}
}

func TestGenerateSignalProgram(t *testing.T) {
	src := `main() => {
  count $= 0
  double := count * 2
  double -> {
    @io.println(double)
  }
  count -> (old, new) => {
    @io.println(old)
    @io.println(new)
  }
  count = count + 1
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
		`type runeSignal[T comparable] struct`,
		`__count := newRuneSignal(0)`,
		`__double := newRuneSignal(__count.Get() * 2)`,
		`__count.Watch(func(_, _ int) { __double.Set(__count.Get() * 2) })`,
		`__double.Watch(func(_ int, _ int) { fmt.Println(__double.Get()) })`,
		`__count.Watch(func(__old int, __new int) { fmt.Println(__old); fmt.Println(__new) })`,
		`__count.Set(__count.Get() + 1)`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateArrayProgram(t *testing.T) {
	src := `main() => {
  arr := [1, 2, 3]
  @io.println(arr[0])
  arr.push(4)
  @io.println(arr[3])
  @io.println(arr.length())
  @io.println(arr.isEmpty())
  arr.each((value) => @io.println(value))
  mapped := arr.map((value) => value + 1)
  @io.println(mapped[0])
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
		`__arr := []int{1, 2, 3}`,
		`fmt.Println(__arr[0])`,
		`__arr = append(__arr, 4)`,
		`fmt.Println(__arr[3])`,
		`fmt.Println(len(__arr))`,
		`fmt.Println(len(__arr) == 0)`,
		`for _, __value := range __arr`,
		`fmt.Println(__value)`,
		`__mapped := func() []int`,
		`__result = append(__result, __value+1)`,
		`fmt.Println(__mapped[0])`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
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
		t.Fatalf("Generate() error = %v\n%s", err, got)
	}
	wantParts := []string{
		`__scores := map[string]int{}`,
		`__scores["rune"] = 10`,
		`fmt.Println(func() int {`,
		`value, ok := __scores["rune"]`,
		`return 0`,
		`__seen := map[string]struct{}{}`,
		`__seen["rune"] = struct{}{}`,
		`fmt.Println(func() bool { _, ok := __seen["rune"]; return ok }())`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
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
		t.Fatalf("Generate() error = %v\n%s", err, got)
	}
	wantParts := []string{
		`__scores := map[string]int{"a": 1, "b": 2}`,
		`if !__ok`,
		`return __coalesce`,
		`.(int)`,
		`__scores["b"] = 3`,
		`return 7`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
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
		t.Fatalf("Generate() error = %v\n%s", err, got)
	}
	wantParts := []string{
		`"encoding/binary"`,
		`"math"`,
		`type runeBytes struct`,
		`__bytes := newRuneBytes(16)`,
		`__bytes.SetUInt8(0, uint8(255))`,
		`__bytes.SetInt16(1, int16(0-1234), true)`,
		`__bytes.SetUInt64(4, uint64(123456), false)`,
		`__bytes.SetFloat(12, float32(1.5), true)`,
		`fmt.Println(__bytes.GetUInt8(0))`,
		`fmt.Println(__bytes.GetInt16(1, true))`,
		`fmt.Println(__bytes.GetUInt64(4, false))`,
		`fmt.Println(__bytes.GetFloat(12, true))`,
		`fmt.Println(int(uint(16) >> uint(2)))`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
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
		t.Fatalf("Generate() error = %v\n%s", err, got)
	}
	wantParts := []string{
		`"github.com/andybalholm/brotli"`,
		`"github.com/klauspost/compress/zstd"`,
		`func runeCompressBrotli(data []byte)`,
		`func runeCompressUnbrotli(data []byte)`,
		`func runeCompressZstd(data []byte)`,
		`func runeCompressUnzstd(data []byte)`,
		`runeCompressBrotliText("hello")`,
		`runeCompressUnbrotliText(__brotli)`,
		`runeCompressZstdText(__brotliText)`,
		`runeCompressUnzstdText(__zstd)`,
		`fmt.Println(__text)`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateAnonymousObjectProgram(t *testing.T) {
	src := `main() => {
  obj := {
    name: "Alice"
    age: 30

    greet() => @io.println("Hello, my name is " + obj.name)
    nextAge() => .age + 1
  }

  obj2 := {
    parent: obj
    name: "Bob"
  }

  @io.println(obj.name)
  @io.println(obj2.parent.name)
  @io.println(obj.age)
  @io.println(obj.nextAge())
  obj.greet()
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
		t.Fatalf("Generate() error = %v\n%s", err, got)
	}

	wantParts := []string{
		`var __obj struct`,
		`__name    string`,
		`__age     int`,
		`__greet   func()`,
		`__nextAge func() int`,
		`fmt.Println("Hello, my name is " + __obj.__name)`,
		`return __obj.__age + 1`,
		`var __obj2 struct`,
		`__parent struct`,
		`__name string`,
		`_ = __obj2`,
		`fmt.Println(__obj.__name)`,
		`fmt.Println(__obj2.__parent.__name)`,
		`fmt.Println(__obj.__age)`,
		`fmt.Println(__obj.__nextAge())`,
		`__obj.__greet()`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
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
		t.Fatalf("Generate() error = %v\n%s", err, got)
	}

	wantParts := []string{
		`"encoding/json"`,
		"type json1 struct",
		"type json0 struct",
		"F0 string `json:\"display_name\"`",
		"F1 json1",
		"`json:\"user\"`",
		"F2 []string",
		"`json:\"tags\"`",
		"json.Marshal(func() json0",
		"return json0{",
		`F0: v.__name`,
		`v := v.__user`,
		`return json1{F0: v.__name, F1: v.__age}`,
		`F2: v.__tags`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "`json:\"password\"`") {
		t.Fatalf("generated Go should omit ignored JSON field:\n%s", got)
	}
	if strings.Contains(got, `json:"greet"`) {
		t.Fatalf("generated Go should omit function fields:\n%s", got)
	}
	if strings.Contains(got, "func() struct") {
		t.Fatalf("generated Go should reuse named json types:\n%s", got)
	}
	if strings.Contains(got, "__rune_json") {
		t.Fatalf("generated Go should use short json temporaries:\n%s", got)
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
		t.Fatalf("Generate() error = %v\n%s", err, got)
	}
	wantParts := []string{
		`"encoding/json"`,
		`json.Unmarshal([]byte(`,
		`F0 string ` + "`json:\"display_name\"`",
		`F1 []int`,
		"`json:\"scores\"`",
		`out.__name = raw.F0`,
		`out.__scores =`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, `json:"password"`) || strings.Contains(got, `out.__password =`) {
		t.Fatalf("generated Go should leave ignored fields at their zero value:\n%s", got)
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
	if !strings.Contains(err.Error(), "Go backend does not support intrinsic symbol.toString") {
		t.Fatalf("Generate() error = %v, want symbol intrinsic error", err)
	}
	if strings.Contains(got, "symbol.__") {
		t.Fatalf("generated Go leaked symbol fallback call:\n%s", got)
	}
}

func TestGenerateRoutineCallsWaitAtProgramExit(t *testing.T) {
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
		`"sync"`,
		`"strconv"`,
		`var runeTasks sync.WaitGroup`,
		`func __test(__count int) runeTask[runeUnit]`,
		`strconv.Itoa(__count)`,
		`__test(1)`,
		`runeWaitAll()`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateInlineFunctionValueCall(t *testing.T) {
	src := `fun(flag) => {
  (flag {
    true => (x) => {
      k: x.a + 1,
    }
    false => (y) => {
      k: y.b + 1,
    }
  }) ({
    b: 2,
    z: false,
    a: 1,
  }).k
}

main() => {
  @io.println(fun(true) + fun(false))
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
		t.Fatalf("Generate() error = %v\n%s", err, got)
	}

	wantParts := []string{
		`func() func(struct {`,
		`__z bool`,
		`}) struct{ __k int }`,
		`{__b: 2, __z: false, __a: 1}).__k`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateAnnotatedFunctionValueCallWithAnonymousArgument(t *testing.T) {
	src := `Return: {
  b: Int
  z: Bool
  a: Int
}

fun(flag) => {
  (flag {
    true => (x: Return) => {
      k: x.a + 1,
    }
    false => (y: Return) => {
      k: y.b + 1,
    }
  })({
    b: 2,
    z: false,
    a: 1,
  }).k
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
		t.Fatalf("Generate() error = %v\n%s", err, got)
	}
	if !strings.Contains(got, `__Return{__b: 2, __z: false, __a: 1}`) {
		t.Fatalf("generated Go missing __Return literal conversion:\n%s", got)
	}
}

func TestGenerateNestedVoidMatch(t *testing.T) {
	src := `nestedMatch() => {
  x := 1
  y := 2
  x {
    1 => y {
      2 => @io.println("x is 1 and y is 2")
      _ => @io.println("x is 1 and y is not 2")
    }
    _ => @io.println("x is not 1")
  }
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
		t.Fatalf("Generate() error = %v\n%s", err, got)
	}
	if strings.Contains(got, "__Void") {
		t.Fatalf("generated Go should not mention __Void:\n%s", got)
	}
	if !strings.Contains(got, `fmt.Println("x is 1 and y is 2")`) {
		t.Fatalf("generated Go missing nested println:\n%s", got)
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
		t.Fatalf("Generate() error = %v\n%s", err, got)
	}
	wantParts := []string{
		`"regexp"`,
		`type runeRegex struct`,
		`__re := newRuneRegex("rune\\s+(\\d+)", "ig")`,
		`__built := newRuneRegex("\\d+", "g")`,
		`fmt.Println(__re.match("Rune 123 rune 456"))`,
		`fmt.Println(__built.replaceAll("a1 b22", "[$1]"))`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
}
