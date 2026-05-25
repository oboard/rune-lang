package interpreter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/compiler"
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
	interp := New(prog.IR, WithOutput(&out))
	if err := interp.RunMain(); err != nil {
		t.Fatalf("RunMain() error = %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "x is 1 and y is 2" {
		t.Fatalf("output = %q, want nested match output", got)
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
}
`
	prog, diags := compiler.AnalyzeSource("map_literal.rn", src)
	if len(diags) > 0 {
		t.Fatalf("diagnostics = %#v, want none", diags)
	}

	var out bytes.Buffer
	interp := New(prog.IR, WithOutput(&out))
	if err := interp.RunMain(); err != nil {
		t.Fatalf("RunMain() error = %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "1\n3" {
		t.Fatalf("output = %q, want map literal output", got)
	}
}
