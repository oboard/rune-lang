package lsp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/compiler"
)

func TestMethodDefinitionAndRename(t *testing.T) {
	uri := "file:///tmp/main.rn"
	src := `User: {
    age: Int

    isAdult() => .age >= 18
}

main() => {
    user := User {
        age: 22
    }
    user.isAdult()
}
`
	s := &server{docs: map[string]string{uri: src}}

	def := s.definition(uri, positionOf(src, "user.isAdult", "isAdult")).(map[string]any)
	if got := def["uri"]; got != uri {
		t.Fatalf("definition uri = %v, want %s", got, uri)
	}
	start := def["range"].(map[string]any)["start"].(position)
	if start.Line != 3 || start.Character != 4 {
		t.Fatalf("definition start = %+v, want line 3 char 4", start)
	}

	edit := s.rename(uri, positionOf(src, "user.isAdult", "isAdult"), "adult").(map[string]any)
	changes := edit["changes"].(map[string]any)[uri].([]map[string]any)
	if len(changes) != 2 {
		t.Fatalf("rename edits = %d, want 2: %#v", len(changes), changes)
	}

	hover := s.hover(uri, positionOf(src, "user.isAdult", "isAdult")).(map[string]any)
	if got := hoverValue(hover); !strings.Contains(got, "User.isAdult() -> Bool") {
		t.Fatalf("hover = %q, want User.isAdult signature", got)
	}
}

func TestImportDefinitionJumpsToFile(t *testing.T) {
	dir := t.TempDir()
	helperPath := filepath.Join(dir, "helper.rn")
	writeLSPFile(t, helperPath, `helper() -> Int => 1
`)
	mainPath := filepath.Join(dir, "main.rn")
	uri := fileURI(mainPath)
	src := `@"helper.rn"

main() => helper()
`
	s := &server{docs: map[string]string{uri: src}}

	def := s.definition(uri, positionOf(src, `@"helper.rn"`, "helper.rn")).(map[string]any)
	if got := def["uri"].(string); got != fileURI(helperPath) {
		t.Fatalf("definition uri = %s, want %s", got, fileURI(helperPath))
	}
	start := def["range"].(map[string]any)["start"].(position)
	if start.Line != 0 || start.Character != 0 {
		t.Fatalf("definition start = %+v, want file start", start)
	}
}

func TestImportedFunctionDefinitionJumpsToSourceFile(t *testing.T) {
	dir := t.TempDir()
	helperPath := filepath.Join(dir, "helper.rn")
	writeLSPFile(t, helperPath, `helper() -> Int => 1
`)
	mainPath := filepath.Join(dir, "main.rn")
	uri := fileURI(mainPath)
	src := `@"helper.rn"

main() => helper()
`
	s := &server{docs: map[string]string{uri: src}}

	def := s.definition(uri, positionOf(src, "helper()", "helper")).(map[string]any)
	if got := def["uri"].(string); got != fileURI(helperPath) {
		t.Fatalf("definition uri = %s, want %s", got, fileURI(helperPath))
	}
	start := def["range"].(map[string]any)["start"].(position)
	if start.Line != 0 || start.Character != 0 {
		t.Fatalf("definition start = %+v, want helper declaration", start)
	}
}

func TestStructLiteralTypeHoverAndSemanticToken(t *testing.T) {
	uri := "file:///tmp/main.rn"
	src := `User: {
    id: Int
    name: String
    age: Int

    isAdult() -> Bool => .age >= 18
}

main() => {
    user := User {
        id: 1
        name: "oboard"
        age: 22
    }
}
`
	s := &server{docs: map[string]string{uri: src}}

	hover := s.hover(uri, positionOf(src, "User {\n        id", "User")).(map[string]any)
	value := hoverValue(hover)
	for _, want := range []string{"User: {", "id: Int", "name: String", "age: Int", "isAdult() -> Bool"} {
		if !strings.Contains(value, want) {
			t.Fatalf("hover = %q, want %q", value, want)
		}
	}

	resp := s.semanticTokens(uri).(map[string]any)
	got := decodeSemanticTokenTypes(resp["data"].([]int))
	pos := positionOf(src, "User {\n        id", "User")
	if got[pos] != semanticTokenTypeType {
		t.Fatalf("semantic token at %+v = %d, want type; all tokens %#v", pos, got[pos], got)
	}
}

func TestArrayMethodDefinitionUsesCoreStub(t *testing.T) {
	uri := "file:///tmp/main.rn"
	src := `main() => {
    arr := [1, 2, 3]
    arr.map((value) => value + 1)
}
`
	s := &server{docs: map[string]string{uri: src}}

	def := s.definition(uri, positionOf(src, "arr.map", "map")).(map[string]any)
	defURI := def["uri"].(string)
	if !strings.HasSuffix(defURI, "/core/array/array.rn") {
		t.Fatalf("definition uri = %s, want core/array/array.rn", defURI)
	}
	start := def["range"].(map[string]any)["start"].(position)
	if start.Line != 31 || start.Character != 2 {
		t.Fatalf("definition start = %+v, want line 31 char 2", start)
	}
	if got := s.rename(uri, positionOf(src, "arr.map", "map"), "collect"); got != nil {
		t.Fatalf("array method rename = %#v, want nil", got)
	}
	hover := s.hover(uri, positionOf(src, "arr.map", "map")).(map[string]any)
	value := hoverValue(hover)
	if !strings.Contains(value, "Array.map[U](callbackfn: (T, Int, Array[T]) -> U) -> Array[U]") {
		t.Fatalf("hover = %q, want array map signature", value)
	}
	if !strings.Contains(value, "core/array/array.rn") {
		t.Fatalf("hover = %q, want core source path", value)
	}
}

func TestRegexMethodDefinitionUsesCoreStub(t *testing.T) {
	uri := "file:///tmp/main.rn"
	src := `main() => {
    word := /rune\s+(\d+)/ig
    word.match("Rune 123")
}
`
	s := &server{docs: map[string]string{uri: src}}

	def := s.definition(uri, positionOf(src, "word.match", "match")).(map[string]any)
	defURI := def["uri"].(string)
	if !strings.HasSuffix(defURI, "/core/regex/regex.rn") {
		t.Fatalf("definition uri = %s, want core/regex/regex.rn", defURI)
	}
	start := def["range"].(map[string]any)["start"].(position)
	if start.Line != 7 || start.Character != 2 {
		t.Fatalf("definition start = %+v, want line 7 char 2", start)
	}
	if got := s.rename(uri, positionOf(src, "word.match", "match"), "matches"); got != nil {
		t.Fatalf("regex method rename = %#v, want nil", got)
	}
	hover := s.hover(uri, positionOf(src, "word.match", "match")).(map[string]any)
	value := hoverValue(hover)
	if !strings.Contains(value, "Regex.match(string: String) -> Array[String]") {
		t.Fatalf("hover = %q, want regex match signature", value)
	}
	if !strings.Contains(value, "core/regex/regex.rn") {
		t.Fatalf("hover = %q, want core source path", value)
	}
}

func TestCodeLensIncludesSingletonTests(t *testing.T) {
	uri := "file:///tmp/data.rn"
	src := `? "int" {
  @assert.eq(1, 1)
}
`
	s := &server{docs: map[string]string{uri: src}}

	lenses := s.codeLenses(uri).([]map[string]any)
	if len(lenses) != 1 {
		t.Fatalf("code lenses = %d, want 1: %#v", len(lenses), lenses)
	}
	start := lenses[0]["range"].(map[string]any)["start"].(position)
	if start.Line != 0 || start.Character != 0 {
		t.Fatalf("lens start = %+v, want line 0 char 0", start)
	}
	command := lenses[0]["command"].(map[string]any)
	if command["command"] != "rune.runTest" {
		t.Fatalf("command = %#v, want rune.runTest", command)
	}
	args := command["arguments"].([]any)
	spec := args[0].(map[string]any)
	if spec["name"] != "int" {
		t.Fatalf("test spec = %#v, want int", spec)
	}
}

func TestFunctionReferencesFindCallSites(t *testing.T) {
	uri := "file:///tmp/main.rn"
	src := `helper(value: Int) => value + 1

main() => {
    helper(1)
    other := helper(2)
}
`
	s := &server{docs: map[string]string{uri: src}}

	refs := s.references(uri, positionOf(src, "helper(value", "helper"), false).([]map[string]any)
	if len(refs) != 2 {
		t.Fatalf("references = %d, want 2: %#v", len(refs), refs)
	}
	if refStart(refs[0]) != (position{Line: 3, Character: 4}) {
		t.Fatalf("first reference = %+v, want first helper call", refStart(refs[0]))
	}
	if refStart(refs[1]) != (position{Line: 4, Character: 13}) {
		t.Fatalf("second reference = %+v, want second helper call", refStart(refs[1]))
	}

	refs = s.references(uri, positionOf(src, "helper(1)", "helper"), true).([]map[string]any)
	if len(refs) != 3 {
		t.Fatalf("references with declaration = %d, want 3: %#v", len(refs), refs)
	}
	if refStart(refs[0]) != (position{Line: 0, Character: 0}) {
		t.Fatalf("declaration reference = %+v, want helper definition", refStart(refs[0]))
	}
}

func TestMethodReferencesFilterByReceiverType(t *testing.T) {
	uri := "file:///tmp/main.rn"
	src := `User: {
    name: String
    label() => .name
}

Team: {
    label() => "team"
}

main() => {
    user := User { name: "A" }
    team := Team {}
    user.label()
    team.label()
}
`
	s := &server{docs: map[string]string{uri: src}}

	refs := s.references(uri, positionOf(src, "label() => .name", "label"), false).([]map[string]any)
	if len(refs) != 1 {
		t.Fatalf("method references = %d, want 1: %#v", len(refs), refs)
	}
	if refStart(refs[0]) != (position{Line: 12, Character: 9}) {
		t.Fatalf("method reference = %+v, want user.label call", refStart(refs[0]))
	}
}

func TestAnonymousObjectHover(t *testing.T) {
	uri := "file:///tmp/main.rn"
	src := `main() => {
    obj := {
        name: "Alice"
        age: 30

        greet() => @io.println("Hello, my name is " + obj.name)
        nextAge() => .age + 1
    }

    @io.println(obj.name)
    @io.println(obj.age)
    @io.println(obj.nextAge())
    obj.greet()
}
`
	s := &server{docs: map[string]string{uri: src}}

	objHover := s.hover(uri, positionOf(src, "obj :=", "obj")).(map[string]any)
	objValue := hoverValue(objHover)
	wantObject := `obj: {
  name: String
  age: Int
  greet: () -> Void
  nextAge: () -> Int
}`
	if !strings.Contains(objValue, wantObject) {
		t.Fatalf("hover = %q, want %s", objValue, wantObject)
	}

	nameHover := s.hover(uri, positionOf(src, "@io.println(obj.name)", "name")).(map[string]any)
	if got := hoverValue(nameHover); !strings.Contains(got, "name: String") {
		t.Fatalf("hover = %q, want name field type", got)
	}

	nextAgeHover := s.hover(uri, positionOf(src, "obj.nextAge", "nextAge")).(map[string]any)
	if got := hoverValue(nextAgeHover); !strings.Contains(got, "nextAge: () -> Int") {
		t.Fatalf("hover = %q, want nextAge function field type", got)
	}
}

func TestAnonymousObjectDefinition(t *testing.T) {
	uri := "file:///tmp/main.rn"
	src := `main() => {
    obj := {
        name: "Alice"
        age: 30

        greet() => @io.println(.greetText())
        nextAge() => .age + 1
        greetText() => "Hello, my name is " + .name
    }

    obj2 := {
        parent: obj
        name: "Bob"
    }

    @io.println(obj.name)
    @io.println(obj2.parent.name)
}
`
	s := &server{docs: map[string]string{uri: src}}

	objDef := s.definition(uri, positionOf(src, "@io.println(obj.name)", "obj")).(map[string]any)
	if got := objDef["uri"]; got != uri {
		t.Fatalf("obj definition uri = %v, want %s", got, uri)
	}
	objStart := objDef["range"].(map[string]any)["start"].(position)
	if objStart.Line != 1 || objStart.Character != 4 {
		t.Fatalf("obj definition start = %+v, want line 1 char 4", objStart)
	}

	nameDef := s.definition(uri, positionOf(src, "@io.println(obj.name)", "name")).(map[string]any)
	nameStart := nameDef["range"].(map[string]any)["start"].(position)
	if nameStart.Line != 2 || nameStart.Character != 8 {
		t.Fatalf("name definition start = %+v, want line 2 char 8", nameStart)
	}

	parentDef := s.definition(uri, positionOf(src, "obj2.parent", "parent")).(map[string]any)
	parentStart := parentDef["range"].(map[string]any)["start"].(position)
	if parentStart.Line != 11 || parentStart.Character != 8 {
		t.Fatalf("parent definition start = %+v, want line 11 char 8", parentStart)
	}

	nestedNameDef := s.definition(uri, positionOf(src, "obj2.parent.name", "name")).(map[string]any)
	nestedNameStart := nestedNameDef["range"].(map[string]any)["start"].(position)
	if nestedNameStart.Line != 2 || nestedNameStart.Character != 8 {
		t.Fatalf("nested name definition start = %+v, want line 2 char 8", nestedNameStart)
	}

	implicitNameDef := s.definition(uri, positionOf(src, `" + .name`, "name")).(map[string]any)
	implicitNameStart := implicitNameDef["range"].(map[string]any)["start"].(position)
	if implicitNameStart.Line != 2 || implicitNameStart.Character != 8 {
		t.Fatalf("implicit name definition start = %+v, want line 2 char 8", implicitNameStart)
	}
}

func TestInferredFunctionSignatureHover(t *testing.T) {
	uri := "file:///tmp/complex_type.rn"
	src := `fun(flag) => {
  f := (flag {
    true => ((x) => {
      k: x.value + 1,
      label: "left"
    })
    false => ((y) => {
      k: y.value + 2,
      label: "right"
    })
  })

  f({
    value: 1
  }).k
}

main() => {
  @io.println(fun(true))
}
`
	s := &server{docs: map[string]string{uri: src}}

	hover := s.hover(uri, positionOf(src, "fun(flag)", "fun")).(map[string]any)
	value := hoverValue(hover)
	if !strings.Contains(value, "fun(flag: Bool) -> Int") {
		t.Fatalf("hover = %q, want inferred fun signature", value)
	}

	paramHover := s.hover(uri, positionOf(src, "fun(flag)", "flag")).(map[string]any)
	if got := hoverValue(paramHover); !strings.Contains(got, "flag: Bool") {
		t.Fatalf("param hover = %q, want inferred Bool", got)
	}

	completion := s.completion(uri).([]map[string]any)
	var detail string
	for _, item := range completion {
		if item["label"] == "fun" {
			detail = item["detail"].(string)
			break
		}
	}
	if !strings.Contains(detail, "fun(flag: Bool) -> Int") {
		t.Fatalf("completion detail = %q, want inferred fun signature", detail)
	}
}

func TestRoutineFunctionHoverUsesRuneSignature(t *testing.T) {
	uri := "file:///tmp/routine_hover.rn"
	src := `~ test(count: Int) => {
  @io.println(count.toString())
}

main() => {
  test(1)
}
`
	s := &server{docs: map[string]string{uri: src}}

	for _, pos := range []position{
		positionOf(src, "~ test", "test"),
		positionOf(src, "test(1)", "test"),
	} {
		hover := s.hover(uri, pos).(map[string]any)
		if got := hoverValue(hover); !strings.Contains(got, "test: ~ (count: Int) -> Void") {
			t.Fatalf("hover = %q, want routine function signature", got)
		}
	}
}

func TestSemanticTokensMarkSignalVariables(t *testing.T) {
	uri := "file:///tmp/signal.rn"
	src := `main() => {
  count $= 0
  double := count * 2
  double -> {
    @io.println(double)
  }
  count -> (old, new) => {
    @io.println(old)
    @io.println(new)
  }
  count = count + 1
}
`
	s := &server{docs: map[string]string{uri: src}}
	resp := s.semanticTokens(uri).(map[string]any)
	got := decodeSemanticTokenRanges(resp["data"].([]int))
	want := map[position]int{
		{Line: 1, Character: 2}:   5,
		{Line: 2, Character: 2}:   6,
		{Line: 2, Character: 12}:  5,
		{Line: 3, Character: 2}:   6,
		{Line: 4, Character: 16}:  6,
		{Line: 6, Character: 2}:   5,
		{Line: 10, Character: 2}:  5,
		{Line: 10, Character: 10}: 5,
	}
	for pos, length := range want {
		if got[pos] != length {
			t.Fatalf("semantic token at %+v = %d, want %d; all tokens %#v", pos, got[pos], length, got)
		}
	}
	if got[position{Line: 6, Character: 12}] != 0 {
		t.Fatalf("old parameter was marked as signal: %#v", got)
	}
}

func TestSignalWatchInlayHintsReturnOnce(t *testing.T) {
	uri := "file:///tmp/signal.rn"
	src := `main() => {
  count $= 0
  double := count * 2
  double -> {
    @io.println(double)
  }
  count -> (old, new) => {
    @io.println(old)
    @io.println(new)
  }
}
`
	s := &server{docs: map[string]string{uri: src}}
	hints := s.inlayHints(uri).([]map[string]any)
	returnHintsAtWatch := 0
	returnHintsAtShorthand := 0
	for _, hint := range hints {
		pos := hint["position"].(position)
		if pos.Line == 6 && hint["label"] == "-> Void " {
			returnHintsAtWatch++
		}
		if pos.Line >= 2 && pos.Line <= 4 && hint["label"] == "-> Void " {
			returnHintsAtShorthand++
		}
	}
	if returnHintsAtWatch != 1 {
		t.Fatalf("watch return hints = %d, want 1; all hints %#v", returnHintsAtWatch, hints)
	}
	if returnHintsAtShorthand != 0 {
		t.Fatalf("shorthand watch return hints = %d, want 0; all hints %#v", returnHintsAtShorthand, hints)
	}
}

func TestInlayHintsIgnoreImportedDeclarations(t *testing.T) {
	dir := t.TempDir()
	helperPath := filepath.Join(dir, "import_helper.rn")
	writeLSPFile(t, helperPath, `greeting(name) => private(name)

- private(name) => "hello, " + name
`)
	mainPath := filepath.Join(dir, "import_main.rn")
	src := `@"import_helper.rn"

private() => "this is real function"

main() => {
  @io.println(greeting("Rune"))
  @io.println(private())
}
`
	uri := fileURI(mainPath)
	s := &server{docs: map[string]string{uri: src}}

	hints := s.inlayHints(uri).([]map[string]any)
	privateArrow := positionOf(src, "private() =>", "=>")
	mainArrow := positionOf(src, "main() =>", "=>")

	if got := countHintAt(hints, privateArrow, "-> String "); got != 1 {
		t.Fatalf("private return hints = %d, want 1; all hints %#v", got, hints)
	}
	if got := countHintAt(hints, mainArrow, "-> Void "); got != 1 {
		t.Fatalf("main return hints = %d, want 1; all hints %#v", got, hints)
	}
	if got := countHintAt(hints, mainArrow, "-> String "); got != 0 {
		t.Fatalf("main String return hints = %d, want 0; all hints %#v", got, hints)
	}
}

func TestSignalHoverShowsDependencyChain(t *testing.T) {
	uri := "file:///tmp/signal.rn"
	src := `main() => {
  count $= 0
  double := count * 2
  quadruple := double * 2
  @io.println(quadruple)
}
`
	s := &server{docs: map[string]string{uri: src}}
	countHover := hoverValue(s.hover(uri, positionOf(src, "count $=", "count")).(map[string]any))
	if !strings.Contains(countHover, "count: Int") || !strings.Contains(countHover, "signal") {
		t.Fatalf("count hover = %q, want signal type", countHover)
	}
	doubleHover := hoverValue(s.hover(uri, positionOf(src, "double :=", "double")).(map[string]any))
	if !strings.Contains(doubleHover, "double: Int") ||
		!strings.Contains(doubleHover, "computed") ||
		!strings.Contains(doubleHover, "deps: double -> count") {
		t.Fatalf("double hover = %q, want computed dependency chain", doubleHover)
	}
	quadHover := hoverValue(s.hover(uri, positionOf(src, "@io.println(quadruple)", "quadruple")).(map[string]any))
	if !strings.Contains(quadHover, "quadruple: Int") ||
		!strings.Contains(quadHover, "computed") ||
		!strings.Contains(quadHover, "deps: quadruple -> double -> count") {
		t.Fatalf("quadruple hover = %q, want nested dependency chain", quadHover)
	}
}

func decodeSemanticTokenRanges(data []int) map[position]int {
	out := map[position]int{}
	line := 0
	character := 0
	for i := 0; i+4 < len(data); i += 5 {
		line += data[i]
		if data[i] == 0 {
			character += data[i+1]
		} else {
			character = data[i+1]
		}
		out[position{Line: line, Character: character}] = data[i+2]
	}
	return out
}

func decodeSemanticTokenTypes(data []int) map[position]int {
	out := map[position]int{}
	line := 0
	character := 0
	for i := 0; i+4 < len(data); i += 5 {
		line += data[i]
		if data[i] == 0 {
			character += data[i+1]
		} else {
			character = data[i+1]
		}
		out[position{Line: line, Character: character}] = data[i+3]
	}
	return out
}

func TestLocalFunctionTypeRefinesFromCall(t *testing.T) {
	uri := "file:///tmp/complex_type.rn"
	src := `fun(flag) => {
  f := flag {
    true => (x) => {
      k: x.a + 1,
    }
    false => (y) => {
      k: y.b + 1,
    }
  }

  f({
    b: 2,
    z: false,
    a: 1,
  }).k
}

main() => {
  @io.println(fun(true) + fun(false))
}
`
	s := &server{docs: map[string]string{uri: src}}
	_, diags := compiler.AnalyzeSource(uri, src)
	if len(diags) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diags)
	}

	hover := s.hover(uri, positionOf(src, "f :=", "f")).(map[string]any)
	value := hoverValue(hover)
	want := "f: {\n  b: Int\n  z: Bool\n  a: Int\n} -> {\n  k: Int\n}"
	if !strings.Contains(value, want) {
		t.Fatalf("hover = %q, want %q", value, want)
	}

	hints := s.inlayHints(uri).([]map[string]any)
	if !inlayLabelsContain(hints, ": Bool") ||
		!inlayLabelsContain(hints, "-> Int ") ||
		!inlayLabelsContain(hints, ": { b: Int z: Bool a: Int }") ||
		!inlayLabelsContain(hints, "-> { k: Int } ") {
		t.Fatalf("inlay hints = %#v, want inferred named and anonymous function hints", hints)
	}
}

func TestLambdaParamsRefineToNamedStructArgument(t *testing.T) {
	uri := "file:///tmp/named_return.rn"
	src := `Return: {
  b: Int
  z: Bool
  a: Int
}

fun(flag) => {
  flag {
    true => (x) => {
      k: x.a + 1,
    }
    false => (y) => {
      k: y.b + 1,
    }
  }(Return {
    b: 2,
    z: false,
    a: 1,
  }).k
}
`
	s := &server{docs: map[string]string{uri: src}}
	_, diags := compiler.AnalyzeSource(uri, src)
	if len(diags) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diags)
	}

	xHover := s.hover(uri, positionOf(src, "x.a", "x")).(map[string]any)
	if got := hoverValue(xHover); !strings.Contains(got, "x: Return") {
		t.Fatalf("x hover = %q, want Return", got)
	}
	yHover := s.hover(uri, positionOf(src, "y.b", "y")).(map[string]any)
	if got := hoverValue(yHover); !strings.Contains(got, "y: Return") {
		t.Fatalf("y hover = %q, want Return", got)
	}

	def := s.definition(uri, positionOf(src, "Return {", "Return")).(map[string]any)
	start := def["range"].(map[string]any)["start"].(position)
	if start.Line != 0 || start.Character != 0 {
		t.Fatalf("Return definition start = %+v, want line 0 char 0", start)
	}

	edit := s.rename(uri, positionOf(src, "Return {", "Return"), "Result").(map[string]any)
	changes := edit["changes"].(map[string]any)[uri].([]map[string]any)
	if len(changes) != 2 {
		t.Fatalf("rename edits = %d, want 2: %#v", len(changes), changes)
	}

	hints := s.inlayHints(uri).([]map[string]any)
	if !inlayLabelsContain(hints, ": Return") {
		t.Fatalf("inlay hints = %#v, want lambda param Return hints", hints)
	}
}

func TestAnnotatedLambdaParamsAcceptAnonymousStructArgument(t *testing.T) {
	uri := "file:///tmp/annotated_lambda_anon_arg.rn"
	src := `Return: {
  b: Int
  z: Bool
  a: Int
}

fun(flag) => {
  (flag {
    true => (x: Return) => {
      k: x.a + 1,
    }
    false => (y: Return) => {
      k: y.b + 1,
    }
  })({
    b: 2,
    z: false,
    a: 1,
  }).k
}
`
	s := &server{docs: map[string]string{uri: src}}
	_, diags := compiler.AnalyzeSource(uri, src)
	if len(diags) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diags)
	}

	xHover := s.hover(uri, positionOf(src, "x.a", "x")).(map[string]any)
	if got := hoverValue(xHover); !strings.Contains(got, "x: Return") {
		t.Fatalf("x hover = %q, want Return", got)
	}
	hints := s.inlayHints(uri).([]map[string]any)
	if !inlayLabelsContain(hints, "-> { k: Int } ") {
		t.Fatalf("inlay hints = %#v, want anonymous lambda return hint", hints)
	}
}

func TestComplexTypeFunctionSignaturesRejectMismatch(t *testing.T) {
	uri := "file:///tmp/complex_type2.rn"
	src := `h(x) => h(x) + 1

f(x) => {
  k: x.a + h(1),
  onlyLeft: true,
}

g(y) => {
  onlyRight: h(0),
  k: y.b + 1,
}

r(flag) => {
  j := flag {
    true => f
    false => g
  }

  j({
    b: 2,
    z: false,
    a: 1,
  }).k
}
`
	s := &server{docs: map[string]string{uri: src}}
	_, diags := compiler.AnalyzeSource(uri, src)
	if !diagnosticsContain(diags, "match branch returns {b: Int} -> {onlyRight: Int, k: Int}, expected {a: Int} -> {k: Int, onlyLeft: Bool}") {
		t.Fatalf("diagnostics = %#v, want function branch mismatch", diags)
	}

	cases := []struct {
		name string
		want string
	}{
		{name: "h", want: "h(x: Int) -> Int"},
		{name: "f", want: "f(x: {\n  a: Int\n}) -> {\n  k: Int\n  onlyLeft: Bool\n}"},
		{name: "g", want: "g(y: {\n  b: Int\n}) -> {\n  onlyRight: Int\n  k: Int\n}"},
	}
	for _, tc := range cases {
		hover := s.hover(uri, positionOf(src, tc.name+"(", tc.name)).(map[string]any)
		if got := hoverValue(hover); !strings.Contains(got, tc.want) {
			t.Fatalf("%s hover = %q, want %q", tc.name, got, tc.want)
		}
	}

	symbols := s.documentSymbols(uri).([]map[string]any)
	details := map[string]string{}
	for _, symbol := range symbols {
		details[symbol["name"].(string)] = symbol["detail"].(string)
	}
	for _, tc := range cases {
		if !strings.Contains(details[tc.name], tc.want) {
			t.Fatalf("%s symbol detail = %q, want %q", tc.name, details[tc.name], tc.want)
		}
	}
}

func inlayLabelsContain(hints []map[string]any, want string) bool {
	for _, hint := range hints {
		if hint["label"] == want {
			return true
		}
	}
	return false
}

func countHintAt(hints []map[string]any, pos position, label string) int {
	count := 0
	for _, hint := range hints {
		if hint["position"] == pos && hint["label"] == label {
			count++
		}
	}
	return count
}

func diagnosticsContain(diags []compiler.Diagnostic, want string) bool {
	for _, diag := range diags {
		if strings.Contains(diag.Message, want) {
			return true
		}
	}
	return false
}

func hoverValue(hover map[string]any) string {
	return hover["contents"].(map[string]any)["value"].(string)
}

func refStart(ref map[string]any) position {
	return ref["range"].(map[string]any)["start"].(position)
}

func writeLSPFile(t *testing.T, path string, src string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}

func positionOf(src string, context string, needle string) position {
	contextStart := strings.Index(src, context)
	if contextStart < 0 {
		panic("context not found")
	}
	start := contextStart + strings.Index(src[contextStart:], needle)
	if start < contextStart {
		panic("needle not found")
	}
	line := 0
	lineStart := 0
	for i, ch := range src[:start] {
		if ch == '\n' {
			line++
			lineStart = i + 1
		}
	}
	return position{Line: line, Character: start - lineStart}
}
