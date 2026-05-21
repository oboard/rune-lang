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

	lenFn, ok := reg.Function("array", "length")
	if !ok {
		t.Fatal("missing core/array length declaration")
	}
	if lenFn.Receiver != "Array" || lenFn.Intrinsic != "array.len" || lenFn.Return != "Int" {
		t.Fatalf("unexpected array.length declaration: %#v", lenFn)
	}

	atFn, ok := reg.FunctionByAlias("array", "_[_]")
	if !ok {
		t.Fatal("missing core/array index alias")
	}
	if atFn.Name != "at" || atFn.Intrinsic != "array.get" || atFn.Return != "T" {
		t.Fatalf("unexpected array index declaration: %#v", atFn)
	}

	isEmpty, ok := reg.Function("array", "isEmpty")
	if !ok {
		t.Fatal("missing core/array isEmpty declaration")
	}
	if isEmpty.Body == nil || isEmpty.Return != "Bool" {
		t.Fatalf("unexpected array.isEmpty declaration: %#v", isEmpty)
	}

	eachFn, ok := reg.Function("array", "each")
	if !ok {
		t.Fatal("missing core/array each declaration")
	}
	if eachFn.Intrinsic != "array.each" || len(eachFn.Params) != 1 || eachFn.Params[0] != "Func[T,Void]" {
		t.Fatalf("unexpected array.each declaration: %#v", eachFn)
	}

	mapFn, ok := reg.Function("array", "map")
	if !ok {
		t.Fatal("missing core/array map declaration")
	}
	if mapFn.Intrinsic != "array.map" || mapFn.Return != "Array[U]" || len(mapFn.Params) != 1 || mapFn.Params[0] != "Func[T,U]" {
		t.Fatalf("unexpected array.map declaration: %#v", mapFn)
	}
}
