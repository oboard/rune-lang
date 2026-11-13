package tsfront

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/astprint"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/parser"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func TestClassGetterRewritesMemberAccess(t *testing.T) {
	source := `class Temp {
  celsius: number
  constructor(c: number) { this.celsius = c }
  get fahrenheit(): number { return this.celsius * 9 + 32 }
}
function main(): void {
  const temp = new Temp(0)
  console.log(temp.fahrenheit)
}`
	file, diags := ConvertSource("getter.ts", source)
	if len(diags) > 0 {
		t.Fatalf("ConvertSource() diagnostics = %#v", diags)
	}
	printed := astprint.File(file)
	if !strings.Contains(printed, "get_fahrenheit() -> Double") || !strings.Contains(printed, "temp.get_fahrenheit()") {
		t.Fatalf("printed source missing getter lowering:\n%s", printed)
	}
	reparsed, errs := parser.Parse(printed)
	if len(errs) > 0 {
		t.Fatalf("reparse errors = %v\n%s", errs, printed)
	}
	reg, err := stdlib.LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault() error = %v", err)
	}
	if _, checkDiags := checker.CheckWithStdlib(reparsed, reg); len(checkDiags) > 0 {
		t.Fatalf("check diagnostics = %v\n%s", checkDiags, printed)
	}
}

func TestRejectSetterAccessor(t *testing.T) {
	source := `class Temp {
  celsius: number
  constructor(c: number) { this.celsius = c }
  set fahrenheit(v: number) { this.celsius = v }
}`
	_, diags := ConvertSource("setter.ts", source)
	if !hasDiagnosticCode(diags, CodeProfileClass) {
		t.Fatalf("diagnostics = %#v, want %s", diags, CodeProfileClass)
	}
}
