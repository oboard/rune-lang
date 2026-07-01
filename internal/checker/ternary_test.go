package checker

import (
	"testing"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/parser"
)

func TestTernaryLambdaCalleeUsesArgumentObjectType(t *testing.T) {
	file, errs := parser.Parse(`fun(flag) => {
  (flag ? (x) => {
    k: x.a + 1,
  } : (y) => {
    k: y.b + 1,
  })({
    b: 2,
    z: false,
    a: 1,
  }).k
}

main() => {
  @io.println(fun(true) + fun(false))
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}

	info, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("Check() diagnostics = %v", diags)
	}

	want := Type("{b: Int, z: Bool, a: Int}")
	for _, name := range []string{"x", "y"} {
		if got := identifierType(info, file, name); got != want {
			t.Fatalf("%s type = %s, want %s", name, got, want)
		}
	}
	if got := info.Functions["fun"].Params[0].Type; got != Bool {
		t.Fatalf("flag type = %s, want %s", got, Bool)
	}
}

func identifierType(info *Info, file *ast.File, name string) Type {
	var found Type
	visit := func(expr ast.Expr) {
		if found != "" {
			return
		}
		ident, ok := expr.(*ast.Identifier)
		if !ok || ident.Name != name {
			return
		}
		found = info.ExprTypes[ident]
	}
	for _, fn := range file.Functions {
		ast.WalkExpr(fn.Body, visit)
	}
	for _, test := range file.Tests {
		ast.WalkExpr(test.Body, visit)
	}
	return found
}
