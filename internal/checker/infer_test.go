package checker

import (
	"testing"

	"github.com/oboard/rune-lang/internal/parser"
)

func TestBinaryOperandsInferUnknownParameterTypes(t *testing.T) {
	src := `message(name) => "hello, " + name
suffix(text) => text + "!"
forward(name) => helper(name)
- helper(name) => "hello, " + name
increment(value) => value + 1
decrement(value) => value - 1
scale(value) => 2 * value
half(value) => value / 2
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}

	for _, tc := range []struct {
		name       string
		paramType  Type
		returnType Type
	}{
		{"message", String, String},
		{"suffix", String, String},
		{"forward", String, String},
		{"helper", String, String},
		{"increment", Int, Int},
		{"decrement", Int, Int},
		{"scale", Int, Int},
		{"half", Int, Int},
	} {
		fn := info.Functions[tc.name]
		if fn == nil {
			t.Fatalf("missing function %q", tc.name)
		}
		if len(fn.Params) != 1 {
			t.Fatalf("%s params = %d, want 1", tc.name, len(fn.Params))
		}
		if got := fn.Params[0].Type; got != tc.paramType {
			t.Fatalf("%s param type = %s, want %s", tc.name, got, tc.paramType)
		}
		if got := fn.Return; got != tc.returnType {
			t.Fatalf("%s return type = %s, want %s", tc.name, got, tc.returnType)
		}
	}
}
