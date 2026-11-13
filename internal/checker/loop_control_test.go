package checker

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/parser"
)

func TestBreakOutsideLoopIsRejected(t *testing.T) {
	file, errs := parser.Parse("main() => {\n  break\n}")
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	_, diags := Check(file)
	if len(diags) == 0 || !strings.Contains(diags[0].Message, "break outside loop") {
		t.Fatalf("diagnostics = %v, want break outside loop", diags)
	}
}

func TestContinueOutsideLoopIsRejected(t *testing.T) {
	file, errs := parser.Parse("main() => {\n  continue\n}")
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	_, diags := Check(file)
	if len(diags) == 0 || !strings.Contains(diags[0].Message, "continue outside loop") {
		t.Fatalf("diagnostics = %v, want continue outside loop", diags)
	}
}
