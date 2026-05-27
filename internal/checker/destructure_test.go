package checker

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/parser"
)

func TestObjectDestructureInfersFieldTypes(t *testing.T) {
	src := `Advanced: {
  state: Int
  ch: Char
}

read(step: Advanced) -> Int => {
  { state, ch } := step
  ch == 'a' ? state + 1 : state
}
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	_, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
}

func TestObjectDestructureRequiresExistingField(t *testing.T) {
	src := `Advanced: {
  state: Int
}

main(step: Advanced) => {
  { ch } := step
}
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	_, diags := Check(file)
	var found bool
	for _, diag := range diags {
		if strings.Contains(diag.Message, `type Advanced has no field "ch"`) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("diagnostics = %#v, want missing field diagnostic", diags)
	}
}
