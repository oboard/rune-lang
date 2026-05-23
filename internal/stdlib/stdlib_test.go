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

	stringify, ok := reg.Function("json", "stringify")
	if !ok {
		t.Fatal("missing core/json stringify declaration")
	}
	if stringify.Intrinsic != "json.stringify" || stringify.Return != "String" || len(stringify.Params) != 1 || stringify.Params[0] != "Object" {
		t.Fatalf("unexpected json.stringify declaration: %#v", stringify)
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
	if atFn.Name != "at" || atFn.Intrinsic != "array.at" || atFn.Return != "T" {
		t.Fatalf("unexpected array index declaration: %#v", atFn)
	}

	isEmpty, ok := reg.Function("array", "isEmpty")
	if !ok {
		t.Fatal("missing core/array isEmpty declaration")
	}
	if isEmpty.Body == nil || isEmpty.Return != "Bool" {
		t.Fatalf("unexpected array.isEmpty declaration: %#v", isEmpty)
	}

	mapFn, ok := reg.Function("array", "map")
	if !ok {
		t.Fatal("missing core/array map declaration")
	}
	if mapFn.Intrinsic != "array.map" || mapFn.Return != "Array[U]" || len(mapFn.Params) != 1 || mapFn.Params[0] != "Func[T,Int,Array[T],U]" {
		t.Fatalf("unexpected array map declaration: %#v", mapFn)
	}
}

func TestParseMultilineFunctionTypeStub(t *testing.T) {
	mod, err := parseModule("array", "array.rn", `Array[T]: {
  forEach(
    callbackfn: (
      value: T,
      index?: Int,
      array?: Array[T]
    ) -> Any
  ) => "%array.forEach"
}
`)
	if err != nil {
		t.Fatalf("parseModule() error = %v", err)
	}
	fn := mod.byName["forEach"]
	if fn == nil {
		t.Fatal("missing forEach declaration")
	}
	if len(fn.Params) != 1 || fn.Params[0] != "Func[T,Int,Array[T],Any]" {
		t.Fatalf("forEach params = %v, want callback Func", fn.Params)
	}
}
