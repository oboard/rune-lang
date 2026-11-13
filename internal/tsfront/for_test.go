package tsfront

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/oboard/rune-lang/internal/astprint"
	"github.com/oboard/rune-lang/internal/checker"
	gocodegen "github.com/oboard/rune-lang/internal/codegen/go"
	"github.com/oboard/rune-lang/internal/ir"
	"github.com/oboard/rune-lang/internal/parser"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func TestConvertClassicForWithContinue(t *testing.T) {
	source := `function sum(limit: number): number {
  let total = 0
  for (let i = 0; i < limit; i++) {
    if (i === 2) {
      continue
    }
    total = total + i
  }
  return total
}

function count(): number {
  let i = 0
  for (;;) {
    if (i === 3) {
      break
    }
    i = i + 1
  }
  return i
}

function main(): void {
  console.log(sum(4))
  console.log(count())
}`
	file, diags := ConvertSource("for.ts", source)
	if len(diags) > 0 {
		t.Fatalf("ConvertSource() diagnostics = %#v", diags)
	}
	printed := astprint.File(file)
	reparsed, errs := parser.Parse(printed)
	if len(errs) > 0 {
		t.Fatalf("reparse errors = %v\n%s", errs, printed)
	}
	reg, err := stdlib.LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault() error = %v", err)
	}
	info, checkDiags := checker.CheckWithStdlib(reparsed, reg)
	if len(checkDiags) > 0 {
		t.Fatalf("check diagnostics = %v\n%s", checkDiags, printed)
	}
	goSource, err := gocodegen.GenerateIR(ir.LowerFile(reparsed, info))
	if err != nil {
		t.Fatalf("GenerateIR() error = %v", err)
	}
	dir := t.TempDir()
	goFile := filepath.Join(dir, "main.go")
	if err := os.WriteFile(goFile, []byte(goSource), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	out, err := exec.Command("go", "run", goFile).CombinedOutput()
	if err != nil {
		t.Fatalf("go run error = %v\n%s", err, out)
	}
	if string(out) != "4\n3\n" {
		t.Fatalf("go run output = %q", out)
	}
}
