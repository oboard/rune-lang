package parser

import (
	"testing"

	"github.com/oboard/rune-lang/internal/ast"
)

func TestParseSymbolTryCatchFinally(t *testing.T) {
	file, errs := Parse(`run() => {
  {
    !! "boom"
  } !? {
    err => @io.println(err)
  } -- {
    @io.println("finally")
  }
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	body := file.Functions[0].Body.(*ast.BlockExpr)
	tryStmt, ok := body.Statements[0].(*ast.TryStmt)
	if !ok {
		t.Fatalf("statement = %T, want TryStmt", body.Statements[0])
	}
	if tryStmt.CatchName != "err" || tryStmt.Catch == nil || tryStmt.Finally == nil {
		t.Fatalf("try = %#v", tryStmt)
	}
	if len(tryStmt.Body.Statements) != 1 {
		t.Fatalf("body statements = %d, want 1", len(tryStmt.Body.Statements))
	}
	if _, ok := tryStmt.Body.Statements[0].(*ast.ThrowStmt); !ok {
		t.Fatalf("throw = %T, want ThrowStmt", tryStmt.Body.Statements[0])
	}
}

func TestParseSymbolCatchWithoutBinding(t *testing.T) {
	file, errs := Parse(`main() => {
  {
    !! 1
  } ?! {
    @io.println("caught")
  }
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	body := file.Functions[0].Body.(*ast.BlockExpr)
	tryStmt, ok := body.Statements[0].(*ast.TryStmt)
	if !ok || tryStmt.CatchName != "" || tryStmt.Catch == nil {
		t.Fatalf("try = %#v", body.Statements[0])
	}
}
