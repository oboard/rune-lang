package tscodegen

import (
	"strings"
	"testing"
)

func TestOptionalChain(t *testing.T) {
	got := generateForTest(t, `
+ User: { name: String }
+ displayName(user: User?) -> String => user?.name ?? "anonymous"
`)
	if !strings.Contains(got, "user?.name") {
		t.Fatalf("generated TypeScript missing optional selector:\n%s", got)
	}
}

func TestInputEventTargetIsNarrowed(t *testing.T) {
	got := generateForTest(t, `
main() => <input @change={(e) => e.target?.value ?? ""} @keydown={(e) => e.target?.value = ""} />
`)
	for _, want := range []string{
		"Event & { target: HTMLInputElement; currentTarget: HTMLInputElement }",
		"KeyboardEvent & { target: HTMLInputElement; currentTarget: HTMLInputElement }",
		"e as Event & { target: HTMLInputElement; currentTarget: HTMLInputElement }",
		"e.target?.value",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated TypeScript missing %q:\n%s", want, got)
		}
	}
}
