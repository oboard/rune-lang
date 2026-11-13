package lsp

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/compiler"
	"testing"
)

func TestDOMEventTargetCompletion(t *testing.T) {
	uri := "file:///tmp/test.rn"
	src := `main() => {
  <input @change={(e) => e.target.value} />
}
`
	prog, diags := compiler.AnalyzeSource(uri, src)
	if len(diags) > 0 {
		t.Fatalf("diags: %v", diags)
	}
	// Debug: print ExprTypes for selectors
	for expr, typ := range prog.Info.ExprTypes {
		if sel, ok := expr.(*ast.SelectorExpr); ok {
			t.Logf("SelectorExpr name=%s type=%s", sel.Name, typ)
		}
	}
}
