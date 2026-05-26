package checker

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/parser"
)

func TestIterHelpersAreNotModuleFunctions(t *testing.T) {
	file, parseErrs := parser.Parse(`main() => {
  @iter.toArray(@iter.range(0, 1), [])
}
`)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	_, diags := Check(file)
	if !hasDiagnostic(diags, "unknown module function @iter.toArray") {
		t.Fatalf("diagnostics = %#v, want hidden @iter.toArray", diags)
	}
}

func TestIterInternalMethodsAreNotPublic(t *testing.T) {
	file, parseErrs := parser.Parse(`main() => {
  @iter.range(0, 1).toArrayFrom([])
}
`)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	_, diags := Check(file)
	if !hasDiagnostic(diags, `type Iter[Int] has no method "toArrayFrom"`) {
		t.Fatalf("diagnostics = %#v, want hidden Iter.toArrayFrom", diags)
	}
}

func hasDiagnostic(diags []Diagnostic, want string) bool {
	for _, diag := range diags {
		if strings.Contains(diag.Message, want) {
			return true
		}
	}
	return false
}
