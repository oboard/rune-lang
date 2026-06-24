package moonbitcodegen_test

import (
	"strings"
	"testing"

	moonbitcodegen "github.com/oboard/rune-lang/internal/codegen/moonbit"
	"github.com/oboard/rune-lang/internal/compiler"
)

func TestGenerateMainPrintln(t *testing.T) {
	src := `main() => {
  @io.println("Hello")
}`
	got := generateSource(t, src)
	if !strings.Contains(got, "fn main {\n") {
		t.Fatalf("generated main =\n%s\nwant MoonBit main without parameter list", got)
	}
	if !strings.Contains(got, `println("Hello")`) {
		t.Fatalf("generated source =\n%s\nwant println call", got)
	}
}

func TestGeneratePatternBlockFib(t *testing.T) {
	src := `fib(n: Int) -> Int => {
  0 => 0
  1 => 1
  _ => fib(n - 1) + fib(n - 2)
}`
	got := generateSource(t, src)
	for _, want := range []string{
		"(n : Int) -> Int {",
		"match n {",
		"0 => 0",
		"_ =>",
		"(n - 1) +",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated source =\n%s\nmissing %q", got, want)
		}
	}
}

func TestGenerateRejectsGoFFI(t *testing.T) {
	src := `main() => {
  @go.stmt("println")
}`
	prog, diags := compiler.AnalyzeSource("test.rn", src)
	if len(diags) > 0 {
		t.Fatalf("AnalyzeSource() diagnostics = %#v", diags)
	}
	_, err := moonbitcodegen.GenerateIR(prog.IR)
	if err == nil || !strings.Contains(err.Error(), "MoonBit backend does not support @go FFI") {
		t.Fatalf("GenerateIR() error = %v, want @go FFI diagnostic", err)
	}
}

func TestGenerateMapGetOr(t *testing.T) {
	src := `main() => {
  values := @map.new("", "")
  @io.println(values.getOr("name", "Rune"))
}`
	got := generateSource(t, src)
	if !strings.Contains(got, `.get_or_default("name", "Rune")`) {
		t.Fatalf("generated source =\n%s\nwant Map::get_or_default", got)
	}
}

func TestGenerateMapIndexFallback(t *testing.T) {
	src := `main() => {
  values := {
    "name": "Rune",
  }
  @io.println(values["missing"] ?? "fallback")
}`
	got := generateSource(t, src)
	if !strings.Contains(got, `.get("missing").unwrap_or("fallback")`) {
		t.Fatalf("generated source =\n%s\nwant Map::get with Option unwrap_or", got)
	}
	if strings.Contains(got, "??") {
		t.Fatalf("generated source still contains Rune fallback operator:\n%s", got)
	}
}

func TestGenerateParenthesizesTernaryInBinaryExpr(t *testing.T) {
	src := `main() => @io.println("hello " + (true ? "Rune" : "MoonBit"))`
	got := generateSource(t, src)
	if !strings.Contains(got, `"hello " + (if true { "Rune" } else { "MoonBit" })`) {
		t.Fatalf("generated source =\n%s\nwant parenthesized if expression in binary expr", got)
	}
}

func TestGenerateStringTrimReturnsString(t *testing.T) {
	src := `trimmed() -> String => " Rune ".trim()`
	got := generateSource(t, src)
	if !strings.Contains(got, `" Rune ".trim().to_owned()`) {
		t.Fatalf("generated source =\n%s\nwant trim converted back to String", got)
	}
}

func TestGenerateStringSearchUsesCurrentMoonBitNames(t *testing.T) {
	src := `startsAndEnds(value: String) -> Bool => value.startsWith("Ru") && value.endsWith("ne")`
	got := generateSource(t, src)
	for _, want := range []string{
		`value.has_prefix("Ru")`,
		`value.has_suffix("ne")`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated source =\n%s\nmissing %q", got, want)
		}
	}
	for _, deprecated := range []string{"starts_with", "ends_with"} {
		if strings.Contains(got, deprecated) {
			t.Fatalf("generated source contains deprecated %q:\n%s", deprecated, got)
		}
	}
}

func TestGenerateEscapesReservedIdentifiers(t *testing.T) {
	src := `Token: {
  module: String
  static: Int
}

useReserved(method: String) -> String => {
  member := Token { module: method, static: 1 }
  member.module
}`
	got := generateSource(t, src)
	for _, want := range []string{
		"module_ : String",
		"static_ : Int",
		"method_ : String",
		"let member_ = Token::{ module_: method_, static_: 1 }",
		"member_.module_",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated source =\n%s\nmissing escaped identifier %q", got, want)
		}
	}
}

func TestGenerateEnumShowUsesOrdinalValues(t *testing.T) {
	src := `Kind: {
  A
  B
}

main() => @io.println(Kind.B)`
	got := generateSource(t, src)
	for _, want := range []string{
		"} derive(Eq)",
		"pub impl Show for Kind with fn output(self, logger) {",
		"B => 1",
		"}).output(logger)",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated source =\n%s\nmissing %q", got, want)
		}
	}
}

func TestGeneratePrivateKeepaliveForMoonBitWarnings(t *testing.T) {
	src := `Token: {
  kind: Kind
  offset: Int
}

Kind: {
  A
  B
}

unused(kind: Kind) -> String => kind {
  Kind.A => "a"
  Kind.B => "b"
  _ => "unknown"
}

main() => {
  token := Token { kind: Kind.A, offset: 1 }
  @io.println(token.kind)
}`
	got := generateSource(t, src)
	for _, want := range []string{
		"fn __rune_keepalive() -> Unit {",
		"match Kind::B { B => (); _ => () }",
		"let __rune_keep_rune_Token = Token::{ kind: Kind::A, offset: 0 }",
		"ignore(__rune_keep_rune_Token.offset)",
		"__rune_keepalive()",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated source =\n%s\nmissing %q", got, want)
		}
	}
	if !strings.Contains(got, "ignore(__rune_private_") || !strings.Contains(got, "_unused)") {
		t.Fatalf("generated source =\n%s\nmissing private unused function keepalive", got)
	}
	for _, bad := range []string{"pub fn", "pub let", "pub(all)", "_ => \"unknown\""} {
		if strings.Contains(got, bad) {
			t.Fatalf("generated source contains %q:\n%s", bad, got)
		}
	}
}

func TestGenerateIterAndStringBuffer(t *testing.T) {
	src := `main() => {
  values := @iter.range(1, 3)
  first := values.next()
  @io.println(first[0])
  @iter.range(4, 6).map((value) => value * 2).each((value) => @io.println(value))
  buffer := @stringbuffer.from("Rune")
  buffer.append(" core")
  @io.println(buffer.toString())
}`
	got := generateSource(t, src)
	for _, want := range []string{
		"Iter::new(fn()",
		"values.next()",
		"first.0",
		".map((value) => value * 2).to_array().each",
		"StringBuilder()",
		`write_string(" core")`,
		"ignore({ buffer.write_string",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated source =\n%s\nmissing %q", got, want)
		}
	}
	if strings.Contains(got, "().range") || strings.Contains(got, "??") {
		t.Fatalf("generated source contains stale Rune/MoonBit lowering:\n%s", got)
	}
}

func generateSource(t *testing.T, src string) string {
	t.Helper()
	prog, diags := compiler.AnalyzeSource("test.rn", src)
	if len(diags) > 0 {
		t.Fatalf("AnalyzeSource() diagnostics = %#v", diags)
	}
	got, err := moonbitcodegen.GenerateIR(prog.IR)
	if err != nil {
		t.Fatalf("GenerateIR() error = %v\nsource:\n%s", err, got)
	}
	return got
}
