package checker

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/parser"
)

func TestMapLiteralInfersKeyAndValueTypes(t *testing.T) {
	src := `main() => {
  values := {
    "a": 1,
    "b": 2
  }
  values["b"] = 3
  values["a"]
}
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
	block := file.Functions[0].Body.(*ast.BlockExpr)
	let := block.Statements[0].(*ast.LetStmt)
	lit := let.Value.(*ast.MapLiteral)
	if got, want := DisplayType(info.ExprTypes[lit]), "Map[String, Int]"; got != want {
		t.Fatalf("map literal type = %q, want %q", got, want)
	}
	last := block.Statements[2].(*ast.ExprStmt)
	index := last.Expr.(*ast.IndexExpr)
	if got, want := info.ExprTypes[index], NullableOf(Int); got != want {
		t.Fatalf("index type = %s, want %s", got, want)
	}
}

func TestMapIndexNullCoalesceReturnsValueType(t *testing.T) {
	src := `main() => {
  values := {
    "a": 1
  }
  values["missing"] ?? 7
}
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
	block := file.Functions[0].Body.(*ast.BlockExpr)
	stmt := block.Statements[1].(*ast.ExprStmt)
	if got, want := info.ExprTypes[stmt.Expr], Int; got != want {
		t.Fatalf("coalesce type = %s, want %s", got, want)
	}
}

func TestMapLiteralInfersIntegerKeys(t *testing.T) {
	src := `main() => {
  values := {
    1: 2,
    2: 4
  }
  values[1]
}
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
	block := file.Functions[0].Body.(*ast.BlockExpr)
	let := block.Statements[0].(*ast.LetStmt)
	lit := let.Value.(*ast.MapLiteral)
	if got, want := DisplayType(info.ExprTypes[lit]), "Map[Int, Int]"; got != want {
		t.Fatalf("map literal type = %q, want %q", got, want)
	}
}

func TestMapIndexRequiresKeyType(t *testing.T) {
	src := `main() => {
  values := {
    "a": 1
  }
  values[1]
}
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	_, diags := Check(file)
	var found bool
	for _, diag := range diags {
		if strings.Contains(diag.Message, "map index has type Int, expected String") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("diagnostics = %#v, want map key mismatch", diags)
	}
}
