package checker

import (
	"testing"

	"github.com/oboard/rune-lang/internal/ast"
)

func TestOptionalSelectorProducesNullableFieldType(t *testing.T) {
	file := parseSource(t, `
+ User: { name: String }
displayName(user: User?) -> String => user?.name ?? "anonymous"
`)
	info, diags := Check(file)
	for _, diag := range diags {
		t.Logf("diag: %s", diag.Message)
	}
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	body := file.Functions[0].Body.(*ast.BinaryExpr)
	if got := info.ExprTypes[body.Left]; got != Type("String?") {
		t.Fatalf("user?.name type = %s, want String?", got)
	}
}
