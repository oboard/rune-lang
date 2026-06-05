package checker

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/parser"
)

func TestBinaryOperandsInferUnknownParameterTypes(t *testing.T) {
	src := `message(name) => "hello, " + name
suffix(text) => text + "!"
forward(name) => helper(name)
helper(name) => "hello, " + name
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

func TestArrayEachLambdaUsesExpectedElementTypeBeforeFieldInference(t *testing.T) {
	src := `LocalOption: {
  name: String
  valueName: String
  defaultValue: String?
}

main() => {
  options := [
    LocalOption {
      name: "output"
      valueName: "file"
      defaultValue: null
    }
  ]
  values := @map.new("", "")
  options.each((option) => {
    useDefault := !option.valueName.isEmpty() && (option.defaultValue != null)
    (
      useDefault ? values.set(option.name, option.defaultValue ?? "")
        : values
    )
  })
}
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	_, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
}

func TestLambdaArgumentToNonFunctionStdlibParamIsRejected(t *testing.T) {
	src := `main() => {
  values := [1]
  values.push((value) => value)
}
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	_, diags := Check(file)
	for _, diag := range diags {
		if strings.Contains(diag.Message, "argument 1 to @array.push has type") &&
			strings.Contains(diag.Message, "expected Int") {
			return
		}
	}
	t.Fatalf("diagnostics = %#v, want lambda argument mismatch", diags)
}

func TestArrayPushRefinesEmptyArrayReceiver(t *testing.T) {
	src := `RuntimeValue: {
  text: String
}

stringValue(value: String) -> RuntimeValue => RuntimeValue {
  text: value
}

arrayValue(values: Array[RuntimeValue]) -> RuntimeValue => RuntimeValue {
  text: ""
}

main() => {
  out := []
  out.push(stringValue("x"))
  arrayValue(out)
}
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}

	var arrayValueArg ast.Expr
	ast.WalkExpr(file.Functions[len(file.Functions)-1].Body, func(expr ast.Expr) {
		call, ok := expr.(*ast.CallExpr)
		if !ok || len(call.Args) != 1 {
			return
		}
		ident, ok := call.Callee.(*ast.Identifier)
		if !ok || ident.Name != "arrayValue" {
			return
		}
		arrayValueArg = call.Args[0]
	})
	if arrayValueArg == nil {
		t.Fatalf("missing arrayValue(out) call")
	}
	if got, want := info.ExprTypes[arrayValueArg], ArrayOf(Type("RuntimeValue")); got != want {
		t.Fatalf("arrayValue argument type = %s, want %s", got, want)
	}
}
