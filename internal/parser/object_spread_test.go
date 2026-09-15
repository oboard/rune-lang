package parser

import (
	"testing"

	"github.com/oboard/rune-lang/internal/ast"
)

func TestParseObjectLiteralSpread(t *testing.T) {
	file, errs := Parse(`main() => {
  item := { name: "old" }
  updated := {..item, name: "new"}
}`)
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	block := file.Functions[0].Body.(*ast.BlockExpr)
	updated := block.Statements[1].(*ast.LetStmt).Value.(*ast.AnonymousObjectLiteral)
	if len(updated.Fields) != 2 || !updated.Fields[0].Spread || updated.Fields[1].Name != "name" {
		t.Fatalf("object fields = %#v, want spread then name", updated.Fields)
	}
}
