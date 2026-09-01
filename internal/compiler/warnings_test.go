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
		if diag.Kind == "unused_value" || diag.Kind == "unused_field" || diag.Kind == "unused_constructor" {
			t.Fatalf("AnalyzeFileWithWarnings() diagnostics = %#v, do not want public API warning %q", diags, diag.Message)
		}
	}
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
