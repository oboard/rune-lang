package compiler

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/checker"
)

func TestAnalyzeFileWithWarningsSkipsPublicCliAPI(t *testing.T) {
	_, diags := AnalyzeFileWithWarnings(filepath.Join("..", "..", "core", "cli", "cli.rn"))
	for _, diag := range diags {
		if diag.Severity != checker.SeverityWarning {
			continue
		}
		if diag.Kind == "unused_field" || diag.Kind == "unused_constructor" {
			t.Fatalf("AnalyzeFileWithWarnings() diagnostics = %#v, do not want public API warning %q", diags, diag.Message)
		}
	}
}

func TestAnalyzeFileWithWarningsPreservesImportedUnusedFunctionPath(t *testing.T) {
	dir := t.TempDir()
	dependencyPath := filepath.Join(dir, "dependency.rn")
	entryPath := filepath.Join(dir, "entry.rn")
	writeRuneFile(t, dependencyPath, "unused() => 1\n")
	writeRuneFile(t, entryPath, "@\"dependency.rn\"\n\nmain() => 1\n")

	_, diags := AnalyzeFileWithWarnings(entryPath)
	for _, diag := range diags {
		if diag.Kind != "unused_value" {
			continue
		}
		if diag.Path != dependencyPath {
			t.Fatalf("unused function warning path = %q, want %q", diag.Path, dependencyPath)
		}
		if diag.Pos.Line != 1 || diag.Pos.Column != 1 {
			t.Fatalf("unused function warning position = %d:%d, want 1:1", diag.Pos.Line, diag.Pos.Column)
		}
		return
	}
	t.Fatalf("AnalyzeFileWithWarnings() diagnostics = %#v, want unused function warning", diags)
}

func TestAnalyzeFileWithWarningsPreservesImportedDiagnosticPath(t *testing.T) {
	dir := t.TempDir()
	dependencyPath := filepath.Join(dir, "dependency.rn")
	entryPath := filepath.Join(dir, "entry.rn")
	writeRuneFile(t, dependencyPath, `Choice: {
  First
  Second
}

choose(value: Choice) -> Int => value {
  First => 1
  Second => 2
  _ => 3
}
`)
	writeRuneFile(t, entryPath, "@\"dependency.rn\"\n\nmain() => 1\n")

	_, diags := AnalyzeFileWithWarnings(entryPath)
	for _, diag := range diags {
		if diag.Kind != "unreachable_code" {
			continue
		}
		if diag.Path != dependencyPath {
			t.Fatalf("unreachable warning path = %q, want %q", diag.Path, dependencyPath)
		}
		return
	}
	t.Fatalf("AnalyzeFileWithWarnings() diagnostics = %#v, want unreachable warning", diags)
}

func TestAnalyzeSourceWithWarningsKeepsAnalyzeSourceQuiet(t *testing.T) {
	src := `Token: {
  value: Int
}

unused() => 1

main() => {
  token := Token { value: 1 }
  token
}`
	_, diags := AnalyzeSource("warnings.rn", src)
	if len(diags) > 0 {
		t.Fatalf("AnalyzeSource() diagnostics = %#v", diags)
	}
	_, diags = AnalyzeSourceWithWarnings("warnings.rn", src)
	var found bool
	for _, diag := range diags {
		if diag.Severity == checker.SeverityWarning && strings.Contains(diag.Message, "unused_value") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("AnalyzeSourceWithWarnings() diagnostics = %#v, want warning", diags)
	}
}
