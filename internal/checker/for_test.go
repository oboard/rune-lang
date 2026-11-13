package checker

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/parser"
)

func TestForOfBindsArrayElement(t *testing.T) {
	file, errs := parser.Parse(`main(values: Array[Int32]) => {
  for (const value of values) {
    @io.println(value)
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

func TestForOfRejectsNonArray(t *testing.T) {
	file, errs := parser.Parse(`main() => {
  for (const value of 1) {
    @io.println(value)
  }
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	_, diags := Check(file)
	if len(diags) == 0 || !strings.Contains(diags[0].Message, "for-of expects Array") {
		t.Fatalf("diagnostics = %v, want for-of array error", diags)
	}
}
