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
	"github.com/oboard/rune-lang/internal/stdlib"
)

func TestConvertFunctionsAndConsoleLog(t *testing.T) {
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
	if len(file.Functions) != 2 {
		t.Fatalf("functions = %d, want 2", len(file.Functions))
	}
	add := file.Functions[0]
	if add.Name != "add" || len(add.Params) != 2 {
		t.Fatalf("add function = %#v", add)
	}
	if got := add.Params[0].Type.Canonical(); got != "Double" {
		t.Fatalf("param type = %s, want Double", got)
	}
	if got := add.ReturnType.Canonical(); got != "Double" {
		t.Fatalf("return type = %s, want Double", got)
	}

	reg, err := stdlib.LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault() error = %v", err)
	}
	info, checkDiags := checker.CheckWithStdlib(file, reg)
	if len(checkDiags) > 0 {
		t.Fatalf("CheckWithStdlib() diagnostics = %v", checkDiags)
	}
	goSource, err := gocodegen.GenerateIR(ir.LowerFile(file, info))
	if err != nil {
		t.Fatalf("GenerateIR() error = %v", err)
	}
	if !strings.Contains(goSource, "func add(a float64, b float64) float64") {
		t.Fatalf("generated Go missing add signature:\n%s", goSource)
	}
	if !strings.Contains(goSource, "fmt.Println(add(1, 2))") {
		t.Fatalf("generated Go missing console.log lowering:\n%s", goSource)
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
	if string(out) != "3\n" {
		t.Fatalf("go run output = %q, want 3", out)
	}
}

func TestConvertIfWhileAndLocals(t *testing.T) {
	source := `export function sum(limit: number): number {
  let i = 0
  let total = 0
  while (i < limit) {
    i = i + 1
    if (i > 3) {
      break
    }
    total = total + i
  }
  return total
}`
	file, diags := ConvertSource("sample.ts", source)
	if len(diags) > 0 {
		t.Fatalf("ConvertSource() diagnostics = %v", diags)
	}
	reg, err := stdlib.LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault() error = %v", err)
	}
	_, checkDiags := checker.CheckWithStdlib(file, reg)
	if len(checkDiags) > 0 {
		t.Fatalf("CheckWithStdlib() diagnostics = %v", checkDiags)
	}
}

func TestConvertArraysIndexesAndForOf(t *testing.T) {
	source := "export function main(): void {\n" +
		"  const values = [1, 2, 3]\n" +
		"  console.log(values[1])\n" +
		"  for (const value of values) {\n" +
		"    console.log(value)\n" +
		"  }\n" +
		"}"
	file, diags := ConvertSource("sample.ts", source)
	if len(diags) > 0 {
		t.Fatalf("ConvertSource() diagnostics = %v", diags)
	}
	reg, err := stdlib.LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault() error = %v", err)
	}
	_, checkDiags := checker.CheckWithStdlib(file, reg)
	if len(checkDiags) > 0 {
		t.Fatalf("CheckWithStdlib() diagnostics = %v", checkDiags)
	}
}

func TestConvertObjectLiteral(t *testing.T) {
	source := `export function main(): void {
  const user = { name: "Ada", age: 36 }
  console.log(user.name, user.age)
}`
	file, diags := ConvertSource("sample.ts", source)
	if len(diags) > 0 {
		t.Fatalf("ConvertSource() diagnostics = %v", diags)
	}
	reg, err := stdlib.LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault() error = %v", err)
	}
	_, checkDiags := checker.CheckWithStdlib(file, reg)
	if len(checkDiags) > 0 {
		t.Fatalf("CheckWithStdlib() diagnostics = %v", checkDiags)
	}
}

func TestConvertTemplateLiteral(t *testing.T) {
	source := "export function main(): void {\n" +
		"  const name = \"Ada\"\n" +
		"  console.log(`hello ${name}!`)\n" +
		"}"
	file, diags := ConvertSource("sample.ts", source)
	if len(diags) > 0 {
		t.Fatalf("ConvertSource() diagnostics = %v", diags)
	}
	reg, err := stdlib.LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault() error = %v", err)
	}
	_, checkDiags := checker.CheckWithStdlib(file, reg)
	if len(checkDiags) > 0 {
		t.Fatalf("CheckWithStdlib() diagnostics = %v", checkDiags)
	}
}

func TestConvertArrowFunctionCallback(t *testing.T) {
	source := `export function main(): void {
  const values = [1, 2, 3]
  const doubled = values.map((value) => value * 2)
  console.log(doubled[0], doubled[2])
}`
	file, diags := ConvertSource("sample.ts", source)
	if len(diags) > 0 {
		t.Fatalf("ConvertSource() diagnostics = %v", diags)
	}
	reg, err := stdlib.LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault() error = %v", err)
	}
	_, checkDiags := checker.CheckWithStdlib(file, reg)
	if len(checkDiags) > 0 {
		t.Fatalf("CheckWithStdlib() diagnostics = %v", checkDiags)
	}
}

func TestConvertImports(t *testing.T) {
	source := `import { add } from "./math"

export function main(): void {
  console.log(add(1, 2))
}`
	file, diags := ConvertSource("sample.ts", source)
	if len(diags) > 0 {
		t.Fatalf("ConvertSource() diagnostics = %v", diags)
	}
	if len(file.Imports) != 1 || file.Imports[0].Path != "./math.rune" {
		t.Fatalf("imports = %#v, want ./math.rune", file.Imports)
	}
	printed := astprint.File(file)
	if !strings.Contains(printed, `@"./math.rune"`) {
		t.Fatalf("printed source missing import:\n%s", printed)
	}
}

func TestConvertUnionType(t *testing.T) {
	file, diags := ConvertSource("sample.ts", `export function f(value: string | number): void {}`)
	if len(diags) > 0 {
		t.Fatalf("ConvertSource() diagnostics = %v", diags)
	}
	if len(file.Functions) != 1 || len(file.Functions[0].Params) != 1 {
		t.Fatalf("functions = %#v", file.Functions)
	}
	paramType := file.Functions[0].Params[0].Type
	if paramType.Canonical() != "String|Double" {
		t.Fatalf("param type = %q, want String|Double", paramType.Canonical())
	}
	printed := astprint.File(file)
	if !strings.Contains(printed, "String | Double") {
		t.Fatalf("printed source missing union:\n%s", printed)
	}
}

func TestConvertUnionTypeofRunsOnGo(t *testing.T) {
	source := `function describe(value: string | number | boolean): string {
  if (typeof value === "string") {
    return "string:" + value
  }
  if (typeof value === "boolean") {
    return "boolean"
  }
  return "number"
}

function main(): void {
  console.log(describe("Ada"))
  console.log(describe(42))
  console.log(describe(true))
}`
	file, diags := ConvertSource("sample.ts", source)
	if len(diags) > 0 {
		t.Fatalf("ConvertSource() diagnostics = %v", diags)
	}
	reg, err := stdlib.LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault() error = %v", err)
	}
	info, checkDiags := checker.CheckWithStdlib(file, reg)
	if len(checkDiags) > 0 {
		t.Fatalf("CheckWithStdlib() diagnostics = %v", checkDiags)
	}
	goSource, err := gocodegen.GenerateIR(ir.LowerFile(file, info))
	if err != nil {
		t.Fatalf("GenerateIR() error = %v", err)
	}
	if !strings.Contains(goSource, "value any") {
		t.Fatalf("generated Go missing union parameter:\n%s", goSource)
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
	if string(out) != "string:Ada\nnumber\nboolean\n" {
		t.Fatalf("go run output = %q", out)
	}
}
