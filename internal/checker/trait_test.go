package checker

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/parser"
)

func TestStructuralTraitAcceptsExtraMembersAndSelf(t *testing.T) {
	file, parseErrs := parser.Parse(`&Cloneable: {
  value: Int
  clone() -> Self
}

Item: {
  value: Int
  extra: String
  clone() -> Item => Item { value: .value, extra: .extra }
}

accept(value: &Cloneable) -> &Cloneable => value

main() => accept(Item { value: 1, extra: "ok" })
`)
	if len(parseErrs) > 0 {
		t.Fatalf("Parse() errors = %v", parseErrs)
	}
	_, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("Check() diagnostics = %v", diags)
	}
}

func TestStructuralTraitRejectsMissingOrWrongMembers(t *testing.T) {
	file, parseErrs := parser.Parse(`&Cloneable: {
  value: Int
  clone() -> Self
}

Missing: {
  value: Int
}

Wrong: {
  value: String
  clone() -> Wrong => Wrong { value: .value }
}

accept(value: &Cloneable) => value

main() => {
  accept(Missing { value: 1 })
  accept(Wrong { value: "bad" })
}
`)
	if len(parseErrs) > 0 {
		t.Fatalf("Parse() errors = %v", parseErrs)
	}
	_, diags := Check(file)
	var messages []string
	for _, diag := range diags {
		messages = append(messages, diag.Message)
	}
	got := strings.Join(messages, "\n")
	if strings.Count(got, "expected &Cloneable") != 2 {
		t.Fatalf("diagnostics = %v, want two trait mismatch errors", diags)
	}
}

func TestTraitMembersAreAvailableThroughTraitType(t *testing.T) {
	file, parseErrs := parser.Parse(`&Named: {
  name: String
  rename(value: String) -> Self
}

update(value: &Named) -> &Named => value.rename(value.name)
`)
	if len(parseErrs) > 0 {
		t.Fatalf("Parse() errors = %v", parseErrs)
	}
	_, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("Check() diagnostics = %v", diags)
	}
}

func TestStructuralTraitWorksInContainersAndDefaultsMethodsToVoid(t *testing.T) {
	file, parseErrs := parser.Parse(`&Runnable: {
  run()
}

RunnableJob: {
  run() => {}
}

accept(values: Array[&Runnable]) => values

main() => accept([RunnableJob {}])
`)
	if len(parseErrs) > 0 {
		t.Fatalf("Parse() errors = %v", parseErrs)
	}
	_, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("Check() diagnostics = %v", diags)
	}
}

func TestJSONParseRequiresFromJson(t *testing.T) {
	file, parseErrs := parser.Parse(`Plain: {
  name: String
}

main() => {
  value := @json.parse("{\"name\":\"Rune\"}") : Plain
  value
}
`)
	if len(parseErrs) > 0 {
		t.Fatalf("Parse() errors = %v", parseErrs)
	}
	_, diags := Check(file)
	if !hasDiagnostic(diags, "does not implement &FromJson") {
		t.Fatalf("Check() diagnostics = %v, want FromJson mismatch", diags)
	}
}

func TestJSONParseUsesDeclaredReturnType(t *testing.T) {
	file, parseErrs := parser.Parse(`#json.object
User: {
  name: String
}

decode(text: String) -> User => @json.parse(text)
`)
	if len(parseErrs) > 0 {
		t.Fatalf("Parse() errors = %v", parseErrs)
	}
	_, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("Check() diagnostics = %v", diags)
	}
}

func TestJSONParseRequiresExpectedType(t *testing.T) {
	file, parseErrs := parser.Parse(`main() => @json.parse("{}")`)
	if len(parseErrs) > 0 {
		t.Fatalf("Parse() errors = %v", parseErrs)
	}
	_, diags := Check(file)
	if !hasDiagnostic(diags, "target type cannot be inferred") {
		t.Fatalf("Check() diagnostics = %v, want missing target type", diags)
	}
}

func TestJSONParseRejectsWrongGeneratedMethodOverride(t *testing.T) {
	file, parseErrs := parser.Parse(`#json.object
Wrong: {
  name: String
  ::fromJson(value: Int) -> Wrong => Wrong { name: "" }
}

main() => {
  value := @json.parse("{\"name\":\"Rune\"}") : Wrong
  value
}
`)
	if len(parseErrs) > 0 {
		t.Fatalf("Parse() errors = %v", parseErrs)
	}
	_, diags := Check(file)
	if !hasDiagnostic(diags, "does not implement &FromJson") {
		t.Fatalf("Check() diagnostics = %v, want FromJson mismatch", diags)
	}
}

func TestJSONParseRejectsInstanceFromJsonOverride(t *testing.T) {
	file, parseErrs := parser.Parse(`#json.object
Wrong: {
  name: String
  fromJson(text: String) -> Wrong => this
}

main() => {
  value := @json.parse("{\"name\":\"Rune\"}") : Wrong
  value
}
`)
	if len(parseErrs) > 0 {
		t.Fatalf("Parse() errors = %v", parseErrs)
	}
	_, diags := Check(file)
	if !hasDiagnostic(diags, "does not implement &FromJson") {
		t.Fatalf("Check() diagnostics = %v, want static FromJson mismatch", diags)
	}
}

func TestStaticMethodCallRequiresDoubleColon(t *testing.T) {
	file, parseErrs := parser.Parse(`User: {
  ::create() -> User => User {}
}

main() => User.create()
`)
	if len(parseErrs) > 0 {
		t.Fatalf("Parse() errors = %v", parseErrs)
	}
	_, diags := Check(file)
	if !hasDiagnostic(diags, "must be called with '::'") {
		t.Fatalf("Check() diagnostics = %v, want static selector diagnostic", diags)
	}
}

func TestStructLiteralSpreadStillChecksExplicitDuplicateFields(t *testing.T) {
	file, parseErrs := parser.Parse(`User: {
  name: String
  age: Int
}

main(existing: User) => User {
  ...existing,
  age: 41,
  age: 42,
}
`)
	if len(parseErrs) > 0 {
		t.Fatalf("Parse() errors = %v", parseErrs)
	}
	_, diags := Check(file)
	if !hasDiagnostic(diags, `duplicate field value "age"`) {
		t.Fatalf("Check() diagnostics = %v, want duplicate field diagnostic", diags)
	}
}
