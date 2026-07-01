package interpreter_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/compiler"
	"github.com/oboard/rune-lang/internal/interpreter"
)

func TestInterpreterRunsNestedVoidMatch(t *testing.T) {
	src := `nestedMatch() => {
  x := 1
  y := 2
  x {
    1 => y {
      2 => @io.println("x is 1 and y is 2")
      _ => @io.println("x is 1 and y is not 2")
    }
    _ => @io.println("x is not 1")
  }
}

main() => {
  nestedMatch()
}
`
	prog, diags := compiler.AnalyzeSource("nested_match.rn", src)
	if len(diags) > 0 {
		t.Fatalf("diagnostics = %#v, want none", diags)
	}

	var out bytes.Buffer
	interp := interpreter.New(prog.IR, interpreter.WithOutput(&out))
	if err := interp.RunMain(); err != nil {
		t.Fatalf("RunMain() error = %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "x is 1 and y is 2" {
		t.Fatalf("output = %q, want nested match output", got)
	}
}

func TestInterpreterRunsPatternPredicateExpression(t *testing.T) {
	src := `Kind: {
  Ident = 0
  Slash = 1
}

canEnd(kind: Kind) -> Bool => kind ~ (Ident)

main() => {
  @io.println(canEnd(Kind.Ident))
  @io.println(canEnd(Kind.Slash))
}
`
	prog, diags := compiler.AnalyzeSource("pattern_predicate.rn", src)
	if len(diags) > 0 {
		t.Fatalf("diagnostics = %#v, want none", diags)
	}

	var out bytes.Buffer
	interp := interpreter.New(prog.IR, interpreter.WithOutput(&out))
	if err := interp.RunMain(); err != nil {
		t.Fatalf("RunMain() error = %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "true\nfalse" {
		t.Fatalf("output = %q, want predicate match result", got)
	}
}

func TestInterpreterRunsMapLiteral(t *testing.T) {
	src := `main() => {
  values := {
    "a": 1,
    "b": 2
  }
  @io.println(values["a"])
  values["b"] = 3
  @io.println(values["b"])
  @io.println(values["missing"] ?? 7)
  @io.println(values["missing"])
}
`
	prog, diags := compiler.AnalyzeSource("map_literal.rn", src)
	if len(diags) > 0 {
		t.Fatalf("diagnostics = %#v, want none", diags)
	}

	var out bytes.Buffer
	interp := interpreter.New(prog.IR, interpreter.WithOutput(&out))
	if err := interp.RunMain(); err != nil {
		t.Fatalf("RunMain() error = %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "1\n3\n7\nnull" {
		t.Fatalf("output = %q, want map literal output", got)
	}
}

func TestInterpreterRunsTemplateLiteral(t *testing.T) {
	src := "main() => {\n  count := 3\n  ch := 'x'\n  ok := true\n  @io.println(`count \\(count) char \\(ch) ok \\(ok)`)\n}\n"
	prog, diags := compiler.AnalyzeSource("template_literal.rn", src)
	if len(diags) > 0 {
		t.Fatalf("diagnostics = %#v, want none", diags)
	}

	var out bytes.Buffer
	interp := interpreter.New(prog.IR, interpreter.WithOutput(&out))
	if err := interp.RunMain(); err != nil {
		t.Fatalf("RunMain() error = %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "count 3 char x ok true" {
		t.Fatalf("output = %q, want template output", got)
	}
}

func TestInterpreterRunsEnumConstructorPayload(t *testing.T) {
	src := `Box: {
  Item(value: String)
}

unwrap(box: Box) -> String => box {
  Item(value) => value
  _ => ""
}

main() => {
  @io.println(unwrap(Item("payload")))
}
`
	prog, diags := compiler.AnalyzeSource("enum_constructor_payload.rn", src)
	if len(diags) > 0 {
		t.Fatalf("diagnostics = %#v, want none", diags)
	}

	var out bytes.Buffer
	interp := interpreter.New(prog.IR, interpreter.WithOutput(&out))
	if err := interp.RunMain(); err != nil {
		t.Fatalf("RunMain() error = %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "payload" {
		t.Fatalf("output = %q, want enum payload output", got)
	}
}

func TestInterpreterRunsQualifiedEnumConstructorPayload(t *testing.T) {
	src := `+ Box: {
  Empty
  Item(value: String)
}

unwrap(box: Box) -> String => box {
  Empty => ""
  Item(value) => value
}

main() => {
  @io.println(unwrap(Box.Item("payload")))
}
`
	prog, diags := compiler.AnalyzeSource("qualified_enum_constructor_payload.rn", src)
	if len(diags) > 0 {
		t.Fatalf("diagnostics = %#v, want none", diags)
	}

	var out bytes.Buffer
	interp := interpreter.New(prog.IR, interpreter.WithOutput(&out))
	if err := interp.RunMain(); err != nil {
		t.Fatalf("RunMain() error = %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "payload" {
		t.Fatalf("output = %q, want enum payload output", got)
	}
}

func TestInterpreterReadsInputIntrinsics(t *testing.T) {
	src := `main() => {
  @io.println(@io.scan())
  @io.println(@io.scanLine())
  @io.println(@io.readAll())
}
`
	prog, diags := compiler.AnalyzeSource("io_input.rn", src)
	if len(diags) > 0 {
		t.Fatalf("diagnostics = %#v, want none", diags)
	}

	var out bytes.Buffer
	interp := interpreter.New(
		prog.IR,
		interpreter.WithInput(strings.NewReader("alpha beta gamma\nrest")),
		interpreter.WithOutput(&out),
	)
	if err := interp.RunMain(); err != nil {
		t.Fatalf("RunMain() error = %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "alpha\n beta gamma\nrest" {
		t.Fatalf("output = %q, want scanned stdin output", got)
	}
}

func TestInterpreterRunsObjectDestructure(t *testing.T) {
	src := `Point: {
  x: Int
  y: Int
}

main() => {
	  point := Point {
	    x: 20,
	    y: 22
	  }
  { x, y } := point
  @io.println(x + y)
}
`
	prog, diags := compiler.AnalyzeSource("object_destructure.rn", src)
	if len(diags) > 0 {
		t.Fatalf("diagnostics = %#v, want none", diags)
	}

	var out bytes.Buffer
	interp := interpreter.New(prog.IR, interpreter.WithOutput(&out))
	if err := interp.RunMain(); err != nil {
		t.Fatalf("RunMain() error = %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "42" {
		t.Fatalf("output = %q, want 42", got)
	}
}

func TestInterpreterRunsInferredPatternPredicate(t *testing.T) {
	src := `isSpace(ch) => ' ' | '\t' | '\r'
isDigit(ch) => ('0'..='9')
isAlpha(ch: Char) -> Bool => ('a'..='z') | ('A'..='Z')

main() => {
  @io.println(isSpace(' '))
  @io.println(isSpace('x'))
  @io.println(isDigit('5'))
  @io.println(isDigit('x'))
  @io.println(isAlpha('R'))
  @io.println(isAlpha('7'))
}
`
	prog, diags := compiler.AnalyzeSource("pattern_predicate.rn", src)
	if len(diags) > 0 {
		t.Fatalf("diagnostics = %#v, want none", diags)
	}

	var out bytes.Buffer
	interp := interpreter.New(prog.IR, interpreter.WithOutput(&out))
	if err := interp.RunMain(); err != nil {
		t.Fatalf("RunMain() error = %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "true\nfalse\ntrue\nfalse\ntrue\nfalse" {
		t.Fatalf("output = %q, want predicate output", got)
	}
}

func TestInterpreterRunsDestructuringPatterns(t *testing.T) {
	src := `Point: {
  x: Int
  y: Int
}

pointScore(point: Point) -> Int => point {
  { x, y: yy, .. } => x + yy
  _ => 0
}

optionalPoint(point: Point) -> Int => point {
  { z?: 1, .. } => 7
  _ => 0
}

mapScore(values: Map[String, Int]) -> Int => values {
  { "a": value, .. } => value
  { "missing"?: null, .. } => 7
  _ => 0
}

main() => {
  point := Point { x: 20, y: 22 }
  first := { "a": 41 }
  second := { "b": 1 }
  @io.println(pointScore(point))
  @io.println(optionalPoint(point))
  @io.println(mapScore(first))
  @io.println(mapScore(second))
}
`
	prog, diags := compiler.AnalyzeSource("destructuring_patterns.rn", src)
	if len(diags) > 0 {
		t.Fatalf("diagnostics = %#v, want none", diags)
	}

	var out bytes.Buffer
	interp := interpreter.New(prog.IR, interpreter.WithOutput(&out))
	if err := interp.RunMain(); err != nil {
		t.Fatalf("RunMain() error = %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "42\n7\n41\n7" {
		t.Fatalf("output = %q, want destructuring pattern output", got)
	}
}
