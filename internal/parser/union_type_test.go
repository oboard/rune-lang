package parser

import "testing"

func TestParseUnionTypeAnnotation(t *testing.T) {
	file, errs := Parse(`describe(value: String | Double) -> String => (typeof value == "string") ? {
  value
} : {
  "number"
}
`)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	if len(file.Functions) != 1 {
		t.Fatalf("functions = %d", len(file.Functions))
	}
	paramType := file.Functions[0].Params[0].Type
	if paramType.Canonical() != "String|Double" {
		t.Fatalf("param type = %q, want String|Double", paramType.Canonical())
	}
}
