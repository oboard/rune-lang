package checker

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/parser"
)

func TestWhileRequiresBoolCondition(t *testing.T) {
	file, errs := parser.Parse(`main() => {
  * (1) {
    @io.println("x")
  }
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	_, diags := Check(file)
	if len(diags) == 0 || !strings.Contains(diags[0].Message, "loop condition expects Bool") {
		t.Fatalf("diagnostics = %v, want Bool condition error", diags)
	}
}

func TestWhileBodyScopesBindings(t *testing.T) {
	file, errs := parser.Parse(`loop(flag: Bool) => {
  * (flag) {
    @io.println("x")
  }
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	_, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("Check() diagnostics = %v", diags)
	}
}
