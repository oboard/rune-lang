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

const classSource = `class Point {
  x: number
  y: number
  constructor(x: number, y: number) {
    this.x = x
    this.y = y
  }
  sum(): number {
    return this.x + this.y
  }
  static origin(): Point {
    return new Point(0, 0)
  }
}
function main(): void {
  const p = new Point(1, 2)
  console.log(p.sum())
  const o = Point.origin()
  console.log(o.sum())
}`

func TestConvertClassToStructAndFactory(t *testing.T) {
	file, diags := ConvertSource("class.ts", classSource)
	if len(diags) > 0 {
		t.Fatalf("ConvertSource() diagnostics = %#v", diags)
	}
	if len(file.Types) != 1 || file.Types[0].Name != "Point" {
		t.Fatalf("types = %#v", file.Types)
	}
	if len(file.Functions) != 3 || file.Functions[0].Name != "Point_origin" || file.Functions[1].Name != "Point_new" || file.Functions[2].Name != "main" {
		t.Fatalf("functions = %#v", file.Functions)
	}
	printed := astprint.File(file)
	if !strings.Contains(printed, "Point: {") || !strings.Contains(printed, "Point_new(") {
		t.Fatalf("printed source:\n%s", printed)
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
	if string(out) != "3\n0\n" {
		t.Fatalf("go run output = %q", out)
	}
}

func TestRejectStaticClassAccessor(t *testing.T) {
	source := `class Acc {
  static get value(): number { return 1 }
}`
	_, diags := ConvertSource("class.ts", source)
	if !hasDiagnosticCode(diags, CodeProfileClass) {
		t.Fatalf("diagnostics = %#v, want %s", diags, CodeProfileClass)
	}
}

func TestRejectConstructorStatements(t *testing.T) {
	source := `class Point {
  x: number
  constructor(x: number) {
    if (x < 0) {
      x = 0
    }
    this.x = x
  }
}`
	_, diags := ConvertSource("class.ts", source)
	if !hasDiagnosticCode(diags, CodeProfileClass) {
		t.Fatalf("diagnostics = %#v, want %s", diags, CodeProfileClass)
	}
}

func hasDiagnosticCode(diags []Diagnostic, code string) bool {
	for _, diag := range diags {
		if diag.Code == code {
			return true
		}
	}
	return false
}
