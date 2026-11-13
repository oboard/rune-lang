package parser

import (
	"testing"

	"github.com/oboard/rune-lang/internal/ast"
)

func TestParseBreakAndContinue(t *testing.T) {
	file, errs := Parse(`main() => {
  * (true) {
    ^
  }
  * (false) {
    ^>
  }
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	block := file.Functions[0].Body.(*ast.BlockExpr)
	first := block.Statements[0].(*ast.ExprStmt).Expr.(*ast.WhileExpr)
	if _, ok := first.Body.(*ast.BlockExpr).Statements[0].(*ast.BreakStmt); !ok {
		t.Fatalf("first loop body = %T, want BreakStmt", first.Body.(*ast.BlockExpr).Statements[0])
	}
	second := block.Statements[1].(*ast.ExprStmt).Expr.(*ast.WhileExpr)
	if _, ok := second.Body.(*ast.BlockExpr).Statements[0].(*ast.ContinueStmt); !ok {
		t.Fatalf("second loop body = %T, want ContinueStmt", second.Body.(*ast.BlockExpr).Statements[0])
	}
}
