package gocodegen

import (
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

func TestGenerateInlineGoFFI(t *testing.T) {
	src := `@go.import("fmt")

isAdult(age: Int) -> Bool => @go.expr("$age >= 18")

main() => {
  name := "oboard"
  age := 22
  @go.stmt("fmt.Println($name)")
  @go.stmt("fmt.Println($age)")
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
		`fmt.Println(__isAdult(__age))`,
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
  arr.each(value => @io.println(value))
  mapped := arr.map(value => value + 1)
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

func TestGenerateAnonymousObjectProgram(t *testing.T) {
	src := `main() => {
  obj := {
    name: "Alice"
    age: 30

    greet() => @io.println("Hello, my name is " + obj.name)
    nextAge() => .age + 1
  }

  @io.println(obj.name)
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
		`fmt.Println(__obj.__name)`,
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
