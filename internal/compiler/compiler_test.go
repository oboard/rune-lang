package compiler

import (
	"path/filepath"
	"strings"
	"testing"
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
