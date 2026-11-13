package parser

import (
	"testing"

	"github.com/oboard/rune-lang/internal/ast"
)

func TestParseReturnStatement(t *testing.T) {
	file, errs := Parse(`choose() -> Int32 => {
  return 42
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	block := file.Functions[0].Body.(*ast.BlockExpr)
	stmt, ok := block.Statements[0].(*ast.ReturnStmt)
	if !ok {
		t.Fatalf("statement = %T, want ReturnStmt", block.Statements[0])
	}
	lit, ok := stmt.Value.(*ast.IntegerLiteral)
	if !ok || lit.Value != 42 {
		t.Fatalf("return value = %#v, want 42", stmt.Value)
	}
}

func TestParseBareReturnStatement(t *testing.T) {
	file, errs := Parse(`stop() => {
  return
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	block := file.Functions[0].Body.(*ast.BlockExpr)
	stmt := block.Statements[0].(*ast.ReturnStmt)
	if stmt.Value != nil {
		t.Fatalf("return value = %#v, want nil", stmt.Value)
	}
}
