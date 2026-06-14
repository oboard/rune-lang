package checker

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/parser"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func TestResolveCoreMacroAnnotation(t *testing.T) {
	reg, err := stdlib.LoadSources(map[string]string{
		"cli/cli.rn": `#flag(tree: SyntaxFile, context: MacroContext, short: String, help: String) -> SyntaxFile => tree
`,
	})
	if err != nil {
		t.Fatalf("LoadSources() error = %v", err)
	}
	file, parseErrs := parser.Parse(`Args: {
  #cli.flag("v", "show verbose output")
  verbose: Bool
}
`)
	if len(parseErrs) > 0 {
		t.Fatalf("Parse() errors = %v", parseErrs)
	}
	info, diags := CheckWithStdlib(file, reg)
	if len(diags) > 0 {
		t.Fatalf("CheckWithStdlib() diagnostics = %v", diags)
	}
	annotation := &file.Types[0].Fields[0].Annotations[0]
	if info.ResolvedMacros[annotation] == nil {
		t.Fatal("macro annotation was not resolved")
	}
}

func TestRejectNonMacroAnnotationFunction(t *testing.T) {
	reg, err := stdlib.LoadSources(map[string]string{
		"cli/cli.rn": `flag(short: String) -> String => short
`,
	})
	if err != nil {
		t.Fatalf("LoadSources() error = %v", err)
	}
	file, parseErrs := parser.Parse(`#cli.flag("v")
Args: {
  verbose: Bool
}
`)
	if len(parseErrs) > 0 {
		t.Fatalf("Parse() errors = %v", parseErrs)
	}
	_, diags := CheckWithStdlib(file, reg)
	if !diagnosticsContain(diags, "not a macro") {
		t.Fatalf("diagnostics = %v, want non-macro error", diags)
	}
}

func TestRejectRuntimeMacroCall(t *testing.T) {
	reg, err := stdlib.LoadSources(map[string]string{
		"cli/cli.rn": `#flag(tree: SyntaxFile, context: MacroContext, short: String) -> SyntaxFile => tree
`,
	})
	if err != nil {
		t.Fatalf("LoadSources() error = %v", err)
	}
	file, parseErrs := parser.Parse(`main() => @cli.flag("v")
`)
	if len(parseErrs) > 0 {
		t.Fatalf("Parse() errors = %v", parseErrs)
	}
	_, diags := CheckWithStdlib(file, reg)
	if !diagnosticsContain(diags, "can only be used with '#'") {
		t.Fatalf("diagnostics = %v, want runtime macro error", diags)
	}
}

func TestResolveLocalMacroAnnotation(t *testing.T) {
	reg, err := stdlib.LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault() error = %v", err)
	}
	file, parseErrs := parser.Parse(`#tag(tree: SyntaxFile, context: MacroContext, name: String) -> SyntaxFile => tree

#tag("command")
Args: {
  verbose: Bool
}
`)
	if len(parseErrs) > 0 {
		t.Fatalf("Parse() errors = %v", parseErrs)
	}
	info, diags := CheckWithStdlib(file, reg)
	if len(diags) > 0 {
		t.Fatalf("CheckWithStdlib() diagnostics = %v", diags)
	}
	annotation := &file.Types[0].Annotations[0]
	if info.ResolvedMacroFunctions[annotation] == nil {
		t.Fatal("local macro annotation was not resolved")
	}
}

func TestRejectLegacyMacroSignature(t *testing.T) {
	file, parseErrs := parser.Parse(`#tag(name: String) -> String => name
`)
	if len(parseErrs) > 0 {
		t.Fatalf("Parse() errors = %v", parseErrs)
	}
	_, diags := CheckWithStdlib(file, nil)
	if !diagnosticsContain(diags, "must accept SyntaxFile and MacroContext first and return SyntaxFile") {
		t.Fatalf("diagnostics = %v, want macro signature error", diags)
	}
}

func diagnosticsContain(diags []Diagnostic, text string) bool {
	for _, diag := range diags {
		if strings.Contains(diag.Message, text) {
			return true
		}
	}
	return false
}
