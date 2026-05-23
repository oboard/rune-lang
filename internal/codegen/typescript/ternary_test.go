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
