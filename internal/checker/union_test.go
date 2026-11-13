package checker

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/parser"
)

func TestUnionTypeResolvesAndChecksMembers(t *testing.T) {
	file, errs := parser.Parse(`take(value: String | Double) -> Void => {}
main() => {
  take("text")
  take(1.5)
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("Check() diagnostics = %v", diags)
	}
	params := info.FunctionDecls[file.Functions[0]].Params
	if got := params[0].Type; got != Type("String|Double") {
		t.Fatalf("param type = %s, want String|Double", got)
	}
}

func TestUnionTypeRejectsNonMember(t *testing.T) {
	file, errs := parser.Parse(`take(value: String | Double) -> Void => {}
main() => take(true)`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	_, diags := Check(file)
	if len(diags) == 0 || !strings.Contains(diags[0].Message, "expected String|Double") {
		t.Fatalf("diagnostics = %v, want union member rejection", diags)
	}
}

func TestTypeofNarrowsUnionBranch(t *testing.T) {
	file, errs := parser.Parse(`describe(value: String | Double) -> String => (typeof value == "string") ? {
  value + "!"
} : {
  "number"
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if _, diags := Check(file); len(diags) > 0 {
		t.Fatalf("Check() diagnostics = %v", diags)
	}
}

func TestTypeofNegatedGuardNarrowsElse(t *testing.T) {
	file, errs := parser.Parse(`describe(value: String | Double) -> String => (typeof value != "string") ? {
  "number"
} : {
  value + "!"
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if _, diags := Check(file); len(diags) > 0 {
		t.Fatalf("Check() diagnostics = %v", diags)
	}
}

func TestUndefinedSatisfiesNullableField(t *testing.T) {
	file, errs := parser.Parse(`User: {
  name: String
  nickname: String?
}
main() => {
  user := User { name: "Ada", nickname: undefined }
  @io.println(user.name)
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if _, diags := Check(file); len(diags) > 0 {
		t.Fatalf("Check() diagnostics = %v", diags)
	}
}

func TestTypeofUndefinedGuard(t *testing.T) {
	file, errs := parser.Parse(`isMissing(value: String | Undefined) -> Bool => typeof value == "undefined"`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if _, diags := Check(file); len(diags) > 0 {
		t.Fatalf("Check() diagnostics = %v", diags)
	}
}
