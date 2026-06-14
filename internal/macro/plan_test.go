package macro

import (
	"testing"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/parser"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func TestPlanUsesLexicalTargetAndAnnotationOrder(t *testing.T) {
	reg, err := stdlib.LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault() error = %v", err)
	}
	file, parseErrs := parser.Parse(`#first(tree: SyntaxFile, context: MacroContext) -> SyntaxFile => tree

#second(tree: SyntaxFile, context: MacroContext) -> SyntaxFile => tree

#first
#second
Args: {
  #second
  verbose: Bool
}

#first
run() => null
`)
	if len(parseErrs) > 0 {
		t.Fatalf("Parse() errors = %v", parseErrs)
	}
	info, diags := checker.CheckWithStdlib(file, reg)
	if len(diags) > 0 {
		t.Fatalf("CheckWithStdlib() diagnostics = %v", diags)
	}
	invocations := Plan(file, info)
	if len(invocations) != 4 {
		t.Fatalf("invocations = %#v, want 4", invocations)
	}
	wantTargets := []string{"Args", "Args", "verbose", "run"}
	wantMacros := []string{"first", "second", "second", "first"}
	for i, invocation := range invocations {
		if invocation.Order != i {
			t.Fatalf("invocation %d order = %d", i, invocation.Order)
		}
		if invocation.Target.Name != wantTargets[i] {
			t.Fatalf("invocation %d target = %q, want %q", i, invocation.Target.Name, wantTargets[i])
		}
		if invocation.LocalMacro == nil || invocation.LocalMacro.Name != wantMacros[i] {
			t.Fatalf("invocation %d macro = %#v, want %q", i, invocation.LocalMacro, wantMacros[i])
		}
	}
}
