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

const inheritanceSource = `class Base {
  x: number
  constructor(x: number) {
    this.x = x
  }
  value(): number {
    return this.x
  }
  label(): string {
    return "base:" + this.x
  }
  describe(): string {
    return "base"
  }
  static kind(): string {
    return "base"
  }
  static readonly KIND: string = "base-kind"
}

class Child extends Base {
  y: number
  constructor(x: number, y: number) {
    super(x)
    this.y = y
  }
  value(): number {
    return this.x + this.y
  }
  describe(): string {
    return "child+" + super.describe()
  }
  static tag(): string {
    return "child"
  }
  static readonly KIND: string = "child-kind"
}

function main(): void {
  const child = new Child(1, 2)
  console.log(child.value())
  console.log(child.label())
  console.log(child.x)
  console.log(child.y)
  console.log(child.describe())
  console.log(Child.kind())
  console.log(Child.tag())
  console.log(Base.KIND)
  console.log(Child.KIND)
}`

func TestClassInheritanceFlattensFieldsAndMethods(t *testing.T) {
	file, diags := ConvertSource("inherit.ts", inheritanceSource)
	if len(diags) > 0 {
		t.Fatalf("ConvertSource() diagnostics = %#v", diags)
	}
	printed := astprint.File(file)
	if !strings.Contains(printed, "__base_Child := Base_new(x)") {
		t.Fatalf("printed source missing base factory call:\n%s", printed)
	}
	if !strings.Contains(printed, "label() -> String") {
		t.Fatalf("printed source missing inherited method:\n%s", printed)
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
	if string(out) != "3\nbase:1\n1\n2\nchild+base\nbase\nchild\nbase-kind\nchild-kind\n" {
		t.Fatalf("go run output = %q", out)
	}
}

func TestRejectMissingBaseClass(t *testing.T) {
	_, diags := ConvertSource("inherit.ts", `class Child extends Missing {
  constructor() {}
}`)
	if !hasDiagnosticCode(diags, CodeProfileClass) {
		t.Fatalf("diagnostics = %#v, want %s", diags, CodeProfileClass)
	}
}

func TestRejectMissingSuperCall(t *testing.T) {
	source := `class Base {
  x: number
  constructor(x: number) { this.x = x }
}
class Child extends Base {
  y: number
  constructor(y: number) { this.y = y }
}`
	_, diags := ConvertSource("inherit.ts", source)
	if !hasDiagnosticCode(diags, CodeProfileClass) {
		t.Fatalf("diagnostics = %#v, want %s", diags, CodeProfileClass)
	}
}

func TestRejectStaticFieldWithoutInitializer(t *testing.T) {
	_, diags := ConvertSource("class.ts", `class Config {
  static VALUE: number
}`)
	if !hasDiagnosticCode(diags, CodeProfileClass) {
		t.Fatalf("diagnostics = %#v, want %s", diags, CodeProfileClass)
	}
}

func TestSuperMethodCallUsesBaseImplementation(t *testing.T) {
	source := `class Base {
  value(): number { return 1 }
}
class Child extends Base {
  value(): number { return super.value() }
}
function main(): void {
  const child = new Child()
  console.log(child.value())
}`
	file, diags := ConvertSource("inherit.ts", source)
	if len(diags) > 0 {
		t.Fatalf("ConvertSource() diagnostics = %#v", diags)
	}
	printed := astprint.File(file)
	if !strings.Contains(printed, ".__super_value()") {
		t.Fatalf("printed source missing super alias call:\n%s", printed)
	}
}
