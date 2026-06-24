package checker

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/parser"
)

func TestLintMoonBitStyleWarnings(t *testing.T) {
	src := `Token: {
  kind: TokenKind
  offset: Int
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
		"Warning [0006] (unused_constructor): Variant \"XMLText\" is never constructed",
		"Warning [0007] (unused_field): Field \"offset\" is never read",
		"Warning [0001] (unused_value): Function \"tokenKindName\" is never used",
		"Warning [0012] (unreachable_code): Unreachable pattern branch",
	} {
		if !hasWarning(warnings, want) {
			t.Fatalf("warnings = %#v, want %q", warnings, want)
		}
	}
}

func hasWarning(diags []Diagnostic, want string) bool {
	for _, diag := range diags {
		if diag.Severity == SeverityWarning && strings.Contains(diag.Message, want) {
			return true
		}
	}
	return false
}
