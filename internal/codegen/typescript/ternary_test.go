package tscodegen

import (
	"strings"
	"testing"
)

func TestGenerateTernaryExpression(t *testing.T) {
	src := `choose(flag: Bool) -> String => flag ? "yes" : "no"`
	got := generateForTest(t, src)
	wantParts := []string{
		`function __choose(__flag: boolean): string`,
		`return __flag ? "yes" : "no";`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateVoidTernaryExpression(t *testing.T) {
	src := `main() => {
  true ? @io.println("a") : @io.println("b")
}`
	got := generateForTest(t, src)
	want := `true ? console.log("a") : console.log("b");`
	if !strings.Contains(got, want) {
		t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
	}
}

func TestGenerateConditionalExpressionWithoutElse(t *testing.T) {
	src := `main() => {
 handled ~= false
  true ? handled = true
}`
	got := generateForTest(t, src)
	want := `if (true) {
    __handled = true;
  }`
	if !strings.Contains(got, want) {
		t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
	}
	if strings.Contains(got, `__handled = __handled`) {
		t.Fatalf("generated TypeScript contains redundant else assignment:\n%s", got)
	}
	if strings.Contains(got, `: undefined`) {
		t.Fatalf("generated TypeScript contains unnecessary else fallback:\n%s", got)
	}
}

func TestGenerateTernaryLambdaCallee(t *testing.T) {
	src := `fun(flag: Bool) => {
  (flag ? (x) => {
    k: x.a + 1
  } : (y) => {
    k: y.b + 1
  })({
    b: 2,
    z: false,
    a: 1
  }).k
}`
	got := generateForTest(t, src)
	for _, want := range []string{
		`(__flag ?`,
		`__x: { b: number; z: boolean; a: number }`,
		`__y: { b: number; z: boolean; a: number }`,
		`=> ({k: __x.a + 1})`,
		`=> ({k: __y.b + 1})`,
		`({b: 2, z: false, a: 1})).k`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
}
