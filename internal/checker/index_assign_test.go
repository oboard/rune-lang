package checker

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/parser"
)

func TestArrayIndexAssignmentUsesAlias(t *testing.T) {
	src := `main() => {
  arr := [1, 2, 3]
  arr[1] = 10
}
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	_, diags := Check(file)
	for _, diag := range diags {
		if strings.Contains(diag.Message, "not assignable by index") {
			t.Fatalf("index assignment reported error: %v", diag)
		}
	}
}

func TestArrayIndexAssignmentTypeMismatch(t *testing.T) {
	src := `main() => {
  arr := [1, 2, 3]
  arr[1] = "nope"
}
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	_, diags := Check(file)
	found := false
	for _, diag := range diags {
		if strings.Contains(diag.Message, "array assignment has type") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected type mismatch diagnostic, got %v", diags)
	}
}
