package compiler

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/interpreter"
)

func TestAnalyzeCoreSourceDoesNotDuplicateOwnTypes(t *testing.T) {
	_, diags := AnalyzeFile(filepath.Join("..", "..", "core", "set", "set.rn"))
	for _, diag := range diags {
		if strings.Contains(diag.Message, `duplicate type "Set"`) ||
			strings.Contains(diag.Message, `duplicate type "WeakSet"`) {
			t.Fatalf("AnalyzeFile() diagnostics include own stdlib type duplicate: %#v", diags)
		}
	}
	if len(diags) > 0 {
		t.Fatalf("AnalyzeFile() diagnostics = %#v", diags)
	}
}

func TestAnalyzeFileLoadsRuneImports(t *testing.T) {
	dir := t.TempDir()
	writeRuneFile(t, filepath.Join(dir, "math.rn"), `inc(value: Int) -> Int => value + 1
`)
	writeRuneFile(t, filepath.Join(dir, "main.rn"), `@"math.rn"

main() => @io.println(inc(41))
`)

	prog, diags := AnalyzeFile(filepath.Join(dir, "main.rn"))
	if len(diags) > 0 {
		t.Fatalf("AnalyzeFile() diagnostics = %#v", diags)
	}
	if prog.Info.Functions["inc"] == nil {
		t.Fatalf("imported function inc was not collected")
	}

	var out bytes.Buffer
	runner := interpreter.New(prog.IR, interpreter.WithOutput(&out))
	if err := runner.RunMain(); err != nil {
		t.Fatalf("RunMain() error = %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "42" {
		t.Fatalf("output = %q, want 42", got)
	}
}

func TestAnalyzeFileLoadsTransitiveRuneImports(t *testing.T) {
	dir := t.TempDir()
	writeRuneFile(t, filepath.Join(dir, "base.rn"), `base() -> Int => 40
`)
	writeRuneFile(t, filepath.Join(dir, "math.rn"), `@"base.rn"

inc2(value: Int) -> Int => value + 2
`)
	writeRuneFile(t, filepath.Join(dir, "main.rn"), `@"math.rn"

main() => @io.println(inc2(base()))
`)

	prog, diags := AnalyzeFile(filepath.Join(dir, "main.rn"))
	if len(diags) > 0 {
		t.Fatalf("AnalyzeFile() diagnostics = %#v", diags)
	}
	if prog.Info.Functions["base"] == nil || prog.Info.Functions["inc2"] == nil {
		t.Fatalf("transitive imports were not collected: %#v", prog.Info.Functions)
	}
}

func TestAnalyzeFileKeepsPrivateImportsAvailableToOwner(t *testing.T) {
	dir := t.TempDir()
	writeRuneFile(t, filepath.Join(dir, "math.rn"), `- add(a: Int, b: Int) -> Int => a + b

sum(a: Int, b: Int) -> Int => add(a, b)
`)
	writeRuneFile(t, filepath.Join(dir, "main.rn"), `@"math.rn"

main() => @io.println(sum(20, 22))
`)

	prog, diags := AnalyzeFile(filepath.Join(dir, "main.rn"))
	if len(diags) > 0 {
		t.Fatalf("AnalyzeFile() diagnostics = %#v", diags)
	}
	if prog.Info.Functions["add"] == nil || !prog.Info.Functions["add"].Private {
		t.Fatalf("private function add was not collected: %#v", prog.Info.Functions["add"])
	}

	var out bytes.Buffer
	runner := interpreter.New(prog.IR, interpreter.WithOutput(&out))
	if err := runner.RunMain(); err != nil {
		t.Fatalf("RunMain() error = %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "42" {
		t.Fatalf("output = %q, want 42", got)
	}
}

func TestAnalyzeFileRejectsPrivateImportedFunctionUse(t *testing.T) {
	dir := t.TempDir()
	writeRuneFile(t, filepath.Join(dir, "math.rn"), `- add(a: Int, b: Int) -> Int => a + b
`)
	writeRuneFile(t, filepath.Join(dir, "main.rn"), `@"math.rn"

main() => add(20, 22)
`)

	_, diags := AnalyzeFile(filepath.Join(dir, "main.rn"))
	for _, diag := range diags {
		if strings.Contains(diag.Message, `function "add" is private`) {
			return
		}
	}
	t.Fatalf("AnalyzeFile() diagnostics = %#v, want private function diagnostic", diags)
}

func TestAnalyzeFileAllowsLocalFunctionToShadowImportedPrivate(t *testing.T) {
	dir := t.TempDir()
	writeRuneFile(t, filepath.Join(dir, "helper.rn"), `- private() => "this is private function"
`)
	writeRuneFile(t, filepath.Join(dir, "main.rn"), `@"helper.rn"

private() => "this is real function"

main() => @io.println(private())
`)

	prog, diags := AnalyzeFile(filepath.Join(dir, "main.rn"))
	if len(diags) > 0 {
		t.Fatalf("AnalyzeFile() diagnostics = %#v", diags)
	}
	var out bytes.Buffer
	runner := interpreter.New(prog.IR, interpreter.WithOutput(&out))
	if err := runner.RunMain(); err != nil {
		t.Fatalf("RunMain() error = %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "this is real function" {
		t.Fatalf("output = %q, want local function", got)
	}
}

func TestAnalyzeFileRejectsPrivateImportedMembers(t *testing.T) {
	dir := t.TempDir()
	writeRuneFile(t, filepath.Join(dir, "user.rn"), `User: {
  name: String
  - token: String

  - secret() -> String => .token
}

Status: {
  Ready = 0
  - Hidden = 1
}
`)
	writeRuneFile(t, filepath.Join(dir, "main.rn"), `@"user.rn"

main() => {
  user := User { name: "Rune", token: "secret" }
  @io.println(user.token)
  @io.println(user.secret())
  Status.Hidden
}
`)

	_, diags := AnalyzeFile(filepath.Join(dir, "main.rn"))
	want := []string{
		`field "User.token" is private`,
		`method "User.secret" is private`,
		`enum member "Status.Hidden" is private`,
	}
	for _, msg := range want {
		found := false
		for _, diag := range diags {
			if strings.Contains(diag.Message, msg) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("AnalyzeFile() diagnostics = %#v, want %s", diags, msg)
		}
	}
}

func TestAnalyzeFileRequiresRuneImportExtension(t *testing.T) {
	dir := t.TempDir()
	writeRuneFile(t, filepath.Join(dir, "math.rn"), `inc(value: Int) -> Int => value + 1
`)
	writeRuneFile(t, filepath.Join(dir, "main.rn"), `@"math"

main() => inc(1)
`)

	_, diags := AnalyzeFile(filepath.Join(dir, "main.rn"))
	for _, diag := range diags {
		if strings.Contains(diag.Message, "must include a file extension") {
			return
		}
	}
	t.Fatalf("AnalyzeFile() diagnostics = %#v, want missing extension diagnostic", diags)
}

func writeRuneFile(t *testing.T, path string, src string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}
