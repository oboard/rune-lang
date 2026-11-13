package parser

import (
	"testing"

	"github.com/oboard/rune-lang/internal/ast"
)

func TestParseWhileExpression(t *testing.T) {
	file, errs := Parse(`main() => {
  * (keepGoing) {
    @io.println("loop")
  }
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	block := file.Functions[0].Body.(*ast.BlockExpr)
	stmt := block.Statements[0].(*ast.ExprStmt)
	loop, ok := stmt.Expr.(*ast.WhileExpr)
	if !ok {
		t.Fatalf("expression = %T, want WhileExpr", stmt.Expr)
	}
	if loop.Cond == nil || loop.Body == nil {
		t.Fatalf("while node = %#v, want condition and body", loop)
	}
}

func TestParseLegacyWhileKeywordAlias(t *testing.T) {
	file, errs := Parse(`main() => {
  while (keepGoing) {
    break
    continue
  }
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	block := file.Functions[0].Body.(*ast.BlockExpr)
	loop := block.Statements[0].(*ast.ExprStmt).Expr.(*ast.WhileExpr)
	body := loop.Body.(*ast.BlockExpr)
	if _, ok := body.Statements[0].(*ast.BreakStmt); !ok {
		t.Fatalf("first statement = %T, want BreakStmt", body.Statements[0])
	}
	if _, ok := body.Statements[1].(*ast.ContinueStmt); !ok {
		t.Fatalf("second statement = %T, want ContinueStmt", body.Statements[1])
	}
}
