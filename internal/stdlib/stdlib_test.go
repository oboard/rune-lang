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

	reduceFn, ok := reg.Function("array", "reduce")
	if !ok {
		t.Fatal("missing core/array reduce declaration")
	}
	if reduceFn.Intrinsic != "array.reduce" || reduceFn.Return != "U" || len(reduceFn.Params) != 2 || reduceFn.Params[0] != "U" || reduceFn.Params[1] != "Func[U,T,Int,Array[T],U]" {
		t.Fatalf("unexpected array reduce declaration: %#v", reduceFn)
	}

	regexNew, ok := reg.Function("regex", "new")
	if !ok {
		t.Fatal("missing core/regex new declaration")
	}
	if regexNew.Intrinsic != "regex.new" || regexNew.Return != "Regex" || len(regexNew.Params) != 2 || regexNew.Params[0] != "String" || regexNew.Params[1] != "String" {
		t.Fatalf("unexpected regex.new declaration: %#v", regexNew)
	}

	matchAll, ok := reg.ReceiverFunction("regex", "Regex", "matchAll")
	if !ok {
		t.Fatal("missing core/regex matchAll declaration")
	}
	if matchAll.Intrinsic != "regex.matchAll" || matchAll.Return != "Array[Array[String]]" {
		t.Fatalf("unexpected regex.matchAll declaration: %#v", matchAll)
	}

	mapNew, ok := reg.Function("map", "new")
	if !ok {
		t.Fatal("missing core/map new declaration")
	}
	if mapNew.Intrinsic != "map.new" || mapNew.Return != "Map[K,V]" || len(mapNew.Params) != 2 {
		t.Fatalf("unexpected map.new declaration: %#v", mapNew)
	}

	setNew, ok := reg.Function("set", "new")
	if !ok {
		t.Fatal("missing core/set new declaration")
	}
	if setNew.Intrinsic != "set.new" || setNew.Return != "Set[T]" || len(setNew.Params) != 1 {
		t.Fatalf("unexpected set.new declaration: %#v", setNew)
	}

	setAdd, ok := reg.ReceiverFunction("set", "Set", "add")
	if !ok {
		t.Fatal("missing core/set Set.add declaration")
	}
	if setAdd.Intrinsic != "set.add" || setAdd.Return != "Set[T]" {
		t.Fatalf("unexpected set.add declaration: %#v", setAdd)
	}
}

func TestLoadDefaultCachesSuccessfulLoad(t *testing.T) {
	defaultRegistryMu.Lock()
	previous := defaultRegistry
	defaultRegistry = nil
	defaultRegistryMu.Unlock()
	defer func() {
		defaultRegistryMu.Lock()
		defaultRegistry = previous
		defaultRegistryMu.Unlock()
	}()

	first, err := LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault() first error = %v", err)
	}
	second, err := LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault() second error = %v", err)
	}
	if first != second {
		t.Fatal("LoadDefault() did not return the cached registry")
	}
}

func TestParseMultilineFunctionTypeStub(t *testing.T) {
	mod, err := parseModule("array", "array.rn", `Array[T]: {
  each[R](
    callbackfn: (
      value: T,
      index?: Int,
      array?: Array[T]
    ) -> R
  ) -> Void => "%array.each"
}
`)
	if err != nil {
		t.Fatalf("parseModule() error = %v", err)
	}
	fn := mod.byName["each"]
	if fn == nil {
		t.Fatal("missing each declaration")
	}
	if len(fn.Params) != 1 || fn.Params[0] != "Func[T,Int,Array[T],R]" {
		t.Fatalf("each params = %v, want callback Func", fn.Params)
	}
	if fn.Return != "Void" {
		t.Fatalf("each return = %s, want Void", fn.Return)
	}
}
