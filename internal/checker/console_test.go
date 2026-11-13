package checker

import (
	"testing"

	"github.com/oboard/rune-lang/internal/parser"
)

func TestConsoleLogAliasesPrintln(t *testing.T) {
	file, errs := parser.Parse(`main() => {
  console.log("hello")
  console.log("a", 1)
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if _, diags := Check(file); len(diags) > 0 {
		t.Fatalf("Check() diagnostics = %v", diags)
	}
}
