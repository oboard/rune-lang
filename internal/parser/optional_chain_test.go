package parser

import (
	"testing"

	"github.com/oboard/rune-lang/internal/ast"
)

func TestParseOptionalSelector(t *testing.T) {
	file, errs := Parse("main() => event.target?.value ?? \"\"\n")
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	body := file.Functions[0].Body.(*ast.BinaryExpr)
	selector, ok := body.Left.(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("left = %T, want *ast.SelectorExpr", body.Left)
	}
	if !selector.Optional || selector.Name != "value" {
		t.Fatalf("selector = %#v, want optional .value", selector)
	}
}
