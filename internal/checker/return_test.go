package checker

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/parser"
)

func TestReturnMatchesDeclaredType(t *testing.T) {
	file, errs := parser.Parse(`choose() -> Int32 => {
  return 42
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	_, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("Check() diagnostics = %v", diags)
	}
}

func TestReturnInfersFunctionType(t *testing.T) {
	file, errs := parser.Parse(`choose() => {
  return 42
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("Check() diagnostics = %v", diags)
	}
	if got := info.FunctionDecls[file.Functions[0]].Return; got != Int32 {
		t.Fatalf("return type = %s, want Int32", got)
	}
}

func TestReturnRejectsMismatchedType(t *testing.T) {
	file, errs := parser.Parse(`choose() -> Int32 => {
  return "x"
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	_, diags := Check(file)
	if len(diags) == 0 || !strings.Contains(diags[0].Message, "return has type") {
		t.Fatalf("diagnostics = %v, want return type error", diags)
	}
}
