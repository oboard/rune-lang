package checker

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/parser"
)

func TestLintMoonBitStyleWarnings(t *testing.T) {
	src := `Token: {
  kind: TokenKind
  - offset: Int
}

TokenKind: {
  EOF
  Ident
  XMLText
}

tokenKindName(kind: TokenKind) -> String => kind {
  TokenKind.EOF => "EOF"
  TokenKind.Ident => "Ident"
  TokenKind.XMLText => "XMLText"
  _ => "Unknown"
}

makeToken() -> Token => {
  kind: TokenKind.Ident,
  offset: 0
}

main() => {
  token := makeToken()
  @io.println(token.kind)
  @io.println('\x00')
}`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
	warnings := Lint(file, info)
	for _, want := range []string{
		"Warning [0001] (unused_value): Function \"tokenKindName\" is never used",
		"Warning [0012] (unreachable_code): Unreachable pattern branch",
	} {
		if !hasWarning(warnings, want) {
			t.Fatalf("warnings = %#v, want %q", warnings, want)
		}
	}
	for _, warning := range warnings {
		if warning.Kind != "unreachable_code" {
			continue
		}
		if warning.Pos.Line != 16 || warning.Pos.Column != 3 {
			t.Fatalf("unreachable pattern position = %d:%d, want 16:3", warning.Pos.Line, warning.Pos.Column)
		}
		return
	}
	t.Fatalf("warnings = %#v, want unreachable pattern warning", warnings)
}

func TestLintDoesNotTreatEnumMemberBindingAsCatchAll(t *testing.T) {
	src := `RuntimeValue: {
  Void
  Null
  Bool(value: Bool)
}

runtimeKind(value: RuntimeValue) -> Int => value {
  Void => 0
  Null => 1
  Bool(_) => 2
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
	for _, warning := range Lint(file, info) {
		if warning.Kind == "unreachable_code" {
			t.Fatalf("warnings = %#v, do not want unreachable warning", Lint(file, info))
		}
	}
}

func TestLintSkipsPublicDeclarations(t *testing.T) {
	src := `+ Token: {
  offset: Int
}

+ TokenKind: {
  EOF
}

+ tokenKindName(kind: TokenKind) -> String => "EOF"
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
	warnings := Lint(file, info)
	for _, unwanted := range []string{"unused_field", "unused_constructor", "unused_value"} {
		if hasWarning(warnings, unwanted) {
			t.Fatalf("warnings = %#v, do not want public %s warning", warnings, unwanted)
		}
	}
}

func TestLintRejectsChainedEqualityTernary(t *testing.T) {
	src := `dispatch(op: String) -> Int => op == "-" ? 1 : (op == "*" ? 2 : (op == "/" ? 3 : 0))
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
	diags = LintErrors(file, info)
	if !hasLintDiagnostic(diags, "Use pattern matching instead of a chained equality ternary") {
		t.Fatalf("diagnostics = %#v, want chained equality ternary error", diags)
	}
}

func TestLintRejectsChainedEqualityOr(t *testing.T) {
	src := `TokenKind: {
  Ident
  Int
  RParen
}

canEndValueToken(kind: TokenKind) -> Bool => ((kind == TokenKind.Ident) || (kind == TokenKind.Int)) || (kind == TokenKind.RParen)
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
	diags = LintErrors(file, info)
	if !hasLintDiagnostic(diags, "Use '~' with an or-pattern instead of chained equality checks") {
		t.Fatalf("diagnostics = %#v, want chained equality or error", diags)
	}
}

func TestLintWarnsPreferRecordSpread(t *testing.T) {
	src := `CliCommand: {
  name: String
  version: String?
  about: String
  options: Array[String]
  arguments: Array[String]
  commands: Array[CliCommand]
  aliases: Array[String]
}

withVersion(command: CliCommand, version: String) -> CliCommand => {
  name: command.name,
  version: version,
  about: command.about,
  options: command.options,
  arguments: command.arguments,
  commands: command.commands,
  aliases: command.aliases
}
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	info, diags := CheckWithStdlib(file, nil)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
	diags = Lint(file, info)
	if !hasWarning(diags, "Warning [0014] (prefer_record_spread): Use '..command' in this record update instead of copying fields manually") {
		t.Fatalf("diagnostics = %#v, want prefer record spread warning", diags)
	}
}

func TestLintDoesNotWarnRecordSpreadForDifferentSourceType(t *testing.T) {
	src := `SyntaxField: {
  name: String
  value: Int
}

SyntaxExprField: {
  name: String
  value: String
}

make(field: SyntaxField) -> SyntaxExprField => {
  name: field.name,
  value: "expr"
}
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	info, diags := CheckWithStdlib(file, nil)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
	diags = Lint(file, info)
	if hasWarning(diags, "prefer_record_spread") {
		t.Fatalf("diagnostics = %#v, do not want prefer record spread warning", diags)
	}
}

func TestLintWarnsPreferRecordSpreadForInferredAnonymousObject(t *testing.T) {
	src := `CliCommand: {
  name: String
  version: String?
  about: String
  options: Array[String]
  arguments: Array[String]
  commands: Array[CliCommand]
  aliases: Array[String]
}

withAliases(command: CliCommand, aliases: Array[String]) => {
  ..command,
  aliases: aliases
}

`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	info, diags := CheckWithStdlib(file, nil)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
	if got := info.Functions["withAliases"].Return; got != Type("CliCommand") {
		t.Fatalf("withAliases return = %s, want CliCommand", got)
	}
}

func TestLintWarnsPreferTernaryForBoolPatternBlock(t *testing.T) {
	src := `choose(aaa: Bool, bbb: Int, ccc: Int) -> Int => aaa {
  true => bbb
  _ => ccc
}
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	info, diags := CheckWithStdlib(file, nil)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
	diags = Lint(file, info)
	var found bool
	for _, diag := range diags {
		if diag.Kind != "prefer_ternary" {
			continue
		}
		found = true
		if diag.Pos.Line != 1 || diag.Pos.Column != 49 {
			t.Fatalf("prefer_ternary position = %d:%d, want 1:49", diag.Pos.Line, diag.Pos.Column)
		}
	}
	if !found {
		t.Fatalf("diagnostics = %#v, want prefer ternary warning", diags)
	}
}

func hasLintDiagnostic(diags []Diagnostic, want string) bool {
	for _, diag := range diags {
		if diag.Severity != SeverityWarning && strings.Contains(diag.Message, want) {
			return true
		}
	}
	return false
}

func hasWarning(diags []Diagnostic, want string) bool {
	for _, diag := range diags {
		if diag.Severity == SeverityWarning && strings.Contains(diag.Message, want) {
			return true
		}
	}
	return false
}
