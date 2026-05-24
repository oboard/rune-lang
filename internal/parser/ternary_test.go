package parser

import (
	"testing"

	"github.com/oboard/rune-lang/internal/ast"
)

func TestParseTernaryExpression(t *testing.T) {
	file, errs := Parse(`main() => {
  value := flag ? 1 : other ? 2 : 3
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	block, ok := file.Functions[0].Body.(*ast.BlockExpr)
	if !ok || len(block.Statements) != 1 {
		t.Fatalf("main body = %#v, want one block statement", file.Functions[0].Body)
	}
	let, ok := block.Statements[0].(*ast.LetStmt)
	if !ok {
		t.Fatalf("statement = %T, want LetStmt", block.Statements[0])
	}
	ternary, ok := let.Value.(*ast.TernaryExpr)
	if !ok {
		t.Fatalf("let value = %T, want TernaryExpr", let.Value)
	}
	if _, ok := ternary.Alternative.(*ast.TernaryExpr); !ok {
		t.Fatalf("alternative = %T, want nested TernaryExpr", ternary.Alternative)
	}
}

func TestParseTernaryLowerThanOr(t *testing.T) {
	file, errs := Parse(`main() => {
  value := false || true ? 1 : 2
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	block := file.Functions[0].Body.(*ast.BlockExpr)
	let := block.Statements[0].(*ast.LetStmt)
	ternary, ok := let.Value.(*ast.TernaryExpr)
	if !ok {
		t.Fatalf("let value = %T, want TernaryExpr", let.Value)
	}
	if _, ok := ternary.Condition.(*ast.BinaryExpr); !ok {
		t.Fatalf("condition = %T, want BinaryExpr", ternary.Condition)
	}
}

func TestParseInvalidTernaryCalleeDoesNotHang(t *testing.T) {
	_, errs := Parse(`Return: {
  b: Int
  z: Bool
  a: Int
}

fun(flag) => {
  flag ? (x) => {
    k: x.a + 1,
  } : (y) => {
    k: y.b + 1,
  }
  }(Return { b: 2, z: false, a: 1 }).k
}`)
	if len(errs) == 0 {
		t.Fatalf("Parse() errors = 0, want errors for invalid ternary callee")
	}
}
