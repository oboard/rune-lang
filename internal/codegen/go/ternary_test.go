package gocodegen

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/parser"
)

func TestGenerateTernaryExpression(t *testing.T) {
	src := `choose(flag: Bool) -> Int => flag ? 1 : 2

main() => {
  @io.println(choose(true))
}
`
	got := generateGoForTest(t, src)
	wantParts := []string{
		`func __choose(__flag bool) int`,
		`return func() int {`,
		`if __flag {`,
		`return 1`,
		`return 2`,
		`fmt.Println(__choose(true))`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateVoidTernaryExpression(t *testing.T) {
	src := `main() => {
  true ? @io.println("a") : @io.println("b")
}
`
	got := generateGoForTest(t, src)
	wantParts := []string{
		`func() {`,
		`if true {`,
		`fmt.Println("a")`,
		`return`,
		`fmt.Println("b")`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateConditionalExpressionWithoutElse(t *testing.T) {
	src := `main() => {
  handled :=: false
  true ? handled = true
}
`
	got := generateGoForTest(t, src)
	wantParts := []string{
		`if true {`,
		`__handled = true`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated Go missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, `__handled = __handled`) {
		t.Fatalf("generated Go contains redundant else assignment:\n%s", got)
	}
	if strings.Contains(got, `func()`) {
		t.Fatalf("generated Go contains unnecessary ternary thunk:\n%s", got)
	}
}

func generateGoForTest(t *testing.T, src string) string {
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
