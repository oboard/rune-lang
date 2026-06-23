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
