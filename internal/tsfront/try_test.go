package tsfront

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/astprint"
	"github.com/oboard/rune-lang/internal/checker"
	gocodegen "github.com/oboard/rune-lang/internal/codegen/go"
	"github.com/oboard/rune-lang/internal/ir"
	"github.com/oboard/rune-lang/internal/parser"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func TestConvertTryCatchFinallyRunsOnGo(t *testing.T) {
	source := `function run(): void {
  try {
    throw "boom"
  } catch (err) {
    console.log(err)
  } finally {
    console.log("finally")
  }
}

function main(): void {
  run()
}`
	file, diags := ConvertSource("try.ts", source)
	if len(diags) > 0 {
		t.Fatalf("ConvertSource() diagnostics = %#v", diags)
	}
	printed := astprint.File(file)
	if !strings.Contains(printed, "try {") || !strings.Contains(printed, "catch (err)") || !strings.Contains(printed, "finally {") {
		t.Fatalf("printed source missing try/catch/finally:\n%s", printed)
	}
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
	if string(out) != "boom\nfinally\n" {
		t.Fatalf("go run output = %q", out)
	}
}
