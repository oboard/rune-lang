package stdlib

import (
	"path/filepath"
	"testing"
)

func TestLoadCoreStubs(t *testing.T) {
	reg, err := Load(filepath.Join("..", "..", "core"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	expr, ok := reg.Function("go", "expr")
	if !ok {
		t.Fatal("missing core/go expr declaration")
	}
	if expr.Intrinsic != "go.expr" || expr.Return != "Dynamic" || len(expr.Params) != 1 || expr.Params[0] != "String" {
		t.Fatalf("unexpected go.expr declaration: %#v", expr)
	}

	println, ok := reg.Function("io", "println")
	if !ok {
		t.Fatal("missing core/io println declaration")
	}
	if println.Go == nil || println.Go.Import != "fmt" || println.Go.Symbol != "fmt.Println" || !println.Variadic {
		t.Fatalf("unexpected io.println declaration: %#v", println)
	}

	if _, ok := reg.Function("fmt", "println"); ok {
		t.Fatal("unexpected undeclared fmt.println declaration")
	}
}
