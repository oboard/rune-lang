package tsfront

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/astprint"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/parser"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func TestConvertedSourceReparsesAndChecks(t *testing.T) {
	source := `export function add(a: number, b: number): number {
  return a + b
}

export function main(): void {
  console.log(add(1, 2))
}`
	file, diags := ConvertSource("sample.ts", source)
	if len(diags) > 0 {
		t.Fatalf("ConvertSource() diagnostics = %v", diags)
	}
	runeSource := astprint.File(file)
	if !strings.Contains(runeSource, "add(a: Double, b: Double) -> Double") {
		t.Fatalf("printed Rune source missing add signature:\n%s", runeSource)
	}
	reparsed, errs := parser.Parse(runeSource)
	if len(errs) > 0 {
		t.Fatalf("reparse errors = %v\nsource:\n%s", errs, runeSource)
	}
	reg, err := stdlib.LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault() error = %v", err)
	}
	if _, checkDiags := checker.CheckWithStdlib(reparsed, reg); len(checkDiags) > 0 {
		t.Fatalf("reparsed CheckWithStdlib() diagnostics = %v\nsource:\n%s", checkDiags, runeSource)
	}
}
