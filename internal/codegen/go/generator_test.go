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
		`func fib(n int) int`,
		`case n == 0:`,
		`return fib(n-1) + fib(n-2)`,
		`func main_()`,
		`fmt.Println(fib(10))`,
		`func main()`,
		`main_()`,
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
		`func add[T runeNumber](a T, b T) T`,
		`return a + b`,
		`fmt.Println(add(1, 2))`,
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
	want := `case (ch >= '0' && ch <= '9'):`
	if !strings.Contains(got, want) {
		t.Fatalf("generated Go missing %q:\n%s", want, got)
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
		`match1 := values`,
		`case len(match1) >= 2 && true && true:`,
		`head := match1[0]`,
		`tail := match1[len(match1)-1]`,
		`rest := append([]int{}, match1[1:len(match1)-1]...)`,
		`whole := match1`,
		`case len(match1) == 0:`,
		`case (value < 0):`,
		`case (value >= 1):`,
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
		`case (typeName == "") || (typeName == "Void"):`,
		`return "void"`,
		`case (typeName == "Int") || (typeName == "Double"):`,
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
	src := "label(count: Int, ch: Char) -> String => `count \\(count) char \\(ch)`\n"
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
		`return "count " + runeTemplateString(count) + " char " + runeTemplateString(ch)`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "template_literal.go", got, 0); err != nil {
		t.Fatalf("generated Go does not parse: %v\n%s", err, got)
	}

	file, parseErrs = parser.Parse("message(name: String) -> String => `hello\n\\(name)`\n")
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
	if want := `return "hello\n" + name`; !strings.Contains(got, want) {
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
		`x := match`,
		`yy := match`,
		`return x + yy`,
		`func() bool { _, ok`,
		`value := match`,
		`return value`,
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
	    id: 1,
	    name: "oboard",
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
		`type User struct`,
		`id   int`,
		`name string`,
		`age  int`,
		`func (this User) isAdult() bool`,
		`return this.age >= 18`,
		`user := User{id: 1, name: "oboard", age: 22}`,
		`fmt.Println(user.name)`,
		`fmt.Println(user.isAdult())`,
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
	    x: 20,
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
		`destructure1 := point`,
		`x := destructure1.x`,
		`y := destructure1.y`,
		`fmt.Println(x + y)`,
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
		`type Status int`,
		`Status_Completed Status = 0`,
		`Status_Fail`,
		`= 1`,
		`func statusText(status Status) string`,
		`case status == Status_Completed:`,
		`case status == Status_Fail:`,
		`func fallback(flag bool) Status`,
		`return Status(0)`,
		`status := Status_Completed`,
		`fmt.Println(statusText(status))`,
		`fmt.Println(fallback(false))`,
		`Status := Container{Completed: 42}`,
		`fmt.Println(Status.Completed)`,
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
		`func isAdult(age int) bool`,
		`return age >= 18`,
		`name := "oboard"`,
		`fmt.Println(name)`,
		`fmt.Println(age)`,
		mangleIdent("分数💯") + ` := 42`,
		`fmt.Println(` + mangleIdent("分数💯") + `)`,
		`fmt.Println(isAdult(age))`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateGoPackageImportFFI(t *testing.T) {
	src := `isNaN(value: Double) -> Bool => @"go:math".IsNaN(value)

main() => {
  fmt := @"go:fmt"
  fmt.Println("hello")
  @io.println(isNaN(0.0))
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
		`"math"`,
		`func isNaN(value float64) bool`,
		`return math.IsNaN(value)`,
		`fmt.Println("hello")`,
		`fmt.Println(isNaN(0.0))`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, `__fmt :=`) {
		t.Fatalf("generated Go should not bind package namespace locally:\n%s", got)
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "main.go", got, 0); err != nil {
		t.Fatalf("generated Go parse error: %v\n%s", err, got)
	}
}

func TestGenerateArraySpread(t *testing.T) {
	src := `main() => {
  items := ["Item 1"]
  next := [..items, "New Item"]
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
		`next := func() []string`,
		`__rune_spread_out = append(__rune_spread_out, items...)`,
		`__rune_spread_out = append(__rune_spread_out, "New Item")`,
		`fmt.Println(len(next))`,
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
		`range array1`,
		`return accumulator + value`,
		`return result2`,
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
  $count := 0
  $double := $count * 2
  {
    @io.println($count)
    @io.println($double)
  }
  $count -> (old, new) => {
    @io.println(old)
    @io.println(new)
  }
  $count = $count + 1
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
		`count := newRuneSignal(0)`,
		`double := newRuneSignal(count.Get() * 2)`,
		`count.Watch(func(_, _ int) { double.Set(count.Get() * 2) })`,
		`effect1 := func() {`,
		`fmt.Println(count.Get())`,
		`fmt.Println(double.Get())`,
		`effectPending2 := false`,
		`scheduleEffect3 := func() {`,
		`runeScheduleEffect(func() {`,
		`effect1()`,
		`double.Watch(func(_, _ int) { scheduleEffect3() })`,
		`count.Watch(func(old int, new int) { fmt.Println(old); fmt.Println(new) })`,
		`count.Set(count.Get() + 1)`,
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
		`arr := []int{1, 2, 3}`,
		`fmt.Println(arr[0])`,
		`arr = append(arr, 4)`,
		`fmt.Println(arr[3])`,
		`fmt.Println(len(arr))`,
		`fmt.Println(len(arr) == 0)`,
		`for _, value := range arr`,
		`fmt.Println(value)`,
		`mapped := func() []int`,
		`result = append(result, value+1)`,
		`fmt.Println(mapped[0])`,
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
  total :=: 0
  scores.each((value) => total = total + value)

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
		`scores := map[string]int{}`,
		`scores["rune"] = 10`,
		`fmt.Println(func() int {`,
		`value, ok := scores["rune"]`,
		`return 0`,
		`for _, value := range scores`,
		`seen := map[string]struct{}{}`,
		`seen["rune"] = struct{}{}`,
		`fmt.Println(func() bool { _, ok := seen["rune"]; return ok }())`,
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
		`scores := map[string]int{"a": 1, "b": 2}`,
		`if !ok`,
		`return coalesce`,
		`.(int)`,
		`scores["b"] = 3`,
		`return 7`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateAnonymousObjectFieldsStableOrder(t *testing.T) {
	src := `readScore(row) => row.points + row.bonus

main() => {
  row := {
    points: 30,
    bonus: 12
  }
  @io.println(readScore(row))
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
		"row struct {\n\t\tbonus  int\n\t\tpoints int\n\t}",
		"row = struct {\n\t\tbonus  int\n\t\tpoints int\n\t}{bonus: 12, points: 30}",
		"fmt.Println(readScore(row))",
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing stable anonymous object order %q:\n%s", want, got)
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
		`bytes := newRuneBytes(16)`,
		`bytes.SetUInt8(0, func() uint8 { n := int(255); return uint8(n) }())`,
		`bytes.SetInt16(1, func() int16 { n := int(0 - 1234); return int16(n) }(), true)`,
		`bytes.SetUInt64(4, uint64(123456), false)`,
		`bytes.SetFloat(12, float32(1.5), true)`,
		`fmt.Println(bytes.GetUInt8(0))`,
		`fmt.Println(bytes.GetInt16(1, true))`,
		`fmt.Println(bytes.GetUInt64(4, false))`,
		`fmt.Println(bytes.GetFloat(12, true))`,
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
		`runeCompressUnbrotliText(brotli)`,
		`runeCompressZstdText(brotliText)`,
		`runeCompressUnzstdText(zstd)`,
		`fmt.Println(text)`,
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
	    name: "Alice",
	    age: 30,

	    greet() => @io.println("Hello, my name is " + obj.name),
	    nextAge() => .age + 1
	  }

	  obj2 := {
	    parent: obj,
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
		`var obj struct`,
		`name    string`,
		`age     int`,
		`greet   func()`,
		`nextAge func() int`,
		`fmt.Println("Hello, my name is " + obj.name)`,
		`return obj.age + 1`,
		`var obj2 struct`,
		`parent struct`,
		`name   string`,
		`fmt.Println(obj.name)`,
		`fmt.Println(obj2.parent.name)`,
		`fmt.Println(obj.age)`,
		`fmt.Println(obj.nextAge())`,
		`obj.greet()`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
}

func TestGeneratePathBodyHelper(t *testing.T) {
	src := `main() => {
  @io.println(@path.basename("/tmp/example.txt"))
  @io.println(@process.platform())
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
		`func path_basename(path string) string`,
		`func path_normalize(path string) string`,
		`fmt.Println(path_basename("/tmp/example.txt"))`,
		`fmt.Println(runeProcessPlatform())`,
		`append(__rune_spread_out, out...)`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "struct{}{}.__basename") {
		t.Fatalf("generated Go should call stdlib body helper directly:\n%s", got)
	}
	if strings.Contains(got, "spread is only supported") {
		t.Fatalf("generated Go should lower spread array literals:\n%s", got)
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "path_body_helper.go", got, 0); err != nil {
		t.Fatalf("generated Go parse error: %v\n%s", err, got)
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
		`F0: v.name`,
		`v := v.user`,
		`return json1{F0: v.name, F1: v.age}`,
		`F2: v.tags`,
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
	if strings.Contains(got, "__this") {
		t.Fatalf("generated Go should not emit method bodies for omitted JSON function fields:\n%s", got)
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
		`out.name = raw.F0`,
		`out.scores =`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, `json:"password"`) || strings.Contains(got, `out.password =`) {
		t.Fatalf("generated Go should leave ignored fields at their zero value:\n%s", got)
	}
}

func TestGenerateSymbolIntrinsicProgram(t *testing.T) {
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
	if err != nil {
		t.Fatalf("Generate() error = %v\n%s", err, got)
	}
	wantParts := []string{
		`"sync/atomic"`,
		`type runeSymbol struct`,
		`runeSymbolToString(runeSymbolCreate("x"))`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateNetUsesRuneBytes(t *testing.T) {
	src := `~ main() => {
  listener := @net.listen("127.0.0.1:0")?
  conn := listener.accept()?
  data := conn.read(1024)?
  conn.write(data)?
  conn.close()?
  listener.close()?
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
		`func (c *runeTCPConnection) Read(length int) runeTask[runeResult[*runeBytes, *runeError]]`,
		`return runeOk[*runeBytes, *runeError](&runeBytes{data: buf[:n]})`,
		`func (c *runeTCPConnection) Write(data *runeBytes) runeTask[runeResult[int, *runeError]]`,
		`n, err := c.conn.Write(data.data)`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
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
		`func test(count int) runeTask[runeUnit]`,
		`strconv.Itoa(count)`,
		`test(1)`,
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
		`z bool`,
		`}) struct{ k int }`,
		`{a: 1, b: 2, z: false}).k`,
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
	if !strings.Contains(got, `Return{a: 1, b: 2, z: false}`) {
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
		`re := newRuneRegex("rune\\s+(\\d+)", "ig")`,
		`built := newRuneRegex("\\d+", "g")`,
		`fmt.Println(re.match("Rune 123 rune 456"))`,
		`fmt.Println(built.replaceAll("a1 b22", "[$1]"))`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
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
		`type Expr struct`,
		`__tag`,
		`int`,
		`__payload []any`,
		`Expr{__tag: Expr_Add, __payload: []any{`,
		`match1`,
		`.__tag == Expr_Add`,
		`.__payload[0].(Expr)`,
		`.__payload[0].(int)`,
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
		`map[string]any{}`,
		`func(key`,
		`.get(key`,
		`x := mapGet`,
		`b := mapGet`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
}

func TestMangleIdentRemovesPrefixAndEscapesGoKeywords(t *testing.T) {
	for input, want := range map[string]string{"answer": "answer", "type": "type_", "1st": "rune_1st"} {
		if got := mangleIdent(input); got != want {
			t.Errorf("mangleIdent(%q) = %q, want %q", input, got, want)
		}
	}
}
