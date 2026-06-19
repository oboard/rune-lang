package lsp

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/compiler"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func TestLocalMacroAnnotationLanguageFeatures(t *testing.T) {
	src := `@syntax

#tag(tree: SyntaxFile, context: MacroContext, name: String, enabled: Bool) -> SyntaxFile => tree

#tag("command", true)
Args: {
  verbose: Bool
}
`
	uri := "file:///tmp/macro_features.rn"
	session := NewSession()
	session.SetDocument(uri, src)

	hover := session.Hover(uri, 2, 2).(map[string]any)
	if got := hoverValue(hover); !strings.Contains(got, "name: String") ||
		!strings.Contains(got, "enabled: Bool") ||
		strings.Contains(got, "tree: SyntaxFile") ||
		strings.Contains(got, "context: MacroContext") {
		t.Fatalf("hover = %q, want macro parameter names", got)
	}

	definition := session.Definition(uri, 2, 2).(map[string]any)
	targetRange := definition["range"].(map[string]any)
	start := targetRange["start"].(position)
	if start.Line != 2 || start.Character != 1 {
		t.Fatalf("definition start = %#v, want macro function declaration", start)
	}

	hints := session.InlayHints(uri).([]map[string]any)
	if !containsHintLabel(hints, "name=") || !containsHintLabel(hints, "enabled=") {
		t.Fatalf("inlay hints = %#v, want macro argument names", hints)
	}
	if containsHintLabel(hints, "tree=") || containsHintLabel(hints, "context=") {
		t.Fatalf("inlay hints = %#v, hidden macro context leaked", hints)
	}
}

func TestCLIMacroInlayHintsIgnoreGeneratedNodes(t *testing.T) {
	src := `#cli.command("ship", "Ship a build artifact", "1.0.0")
Args: {
  #cli.flag("v", "enable verbose output")
  verbose: Bool
  #cli.option("o", "FILE", "write output file", "dist/app")
  output: String
  #cli.arg("target name")
  target: String
}

#cli.main
main(args: Args) => {
  @io.println("target: " + args.target)
}
`
	uri := "file:///tmp/cli_macro_inlay.rn"
	session := NewSession()
	session.SetDocument(uri, src)

	hints := session.InlayHints(uri).([]map[string]any)
	for _, hint := range hints {
		if hint["position"] == (position{Line: 0, Character: 0}) {
			t.Fatalf("generated hint leaked at document origin: %#v", hints)
		}
	}
	for _, label := range []string{
		"name=", "about=", "version=", "short=", "valueName=",
		"help=", "defaultValue=",
	} {
		if !containsHintLabel(hints, label) {
			t.Fatalf("inlay hints = %#v, want %q", hints, label)
		}
	}
	for i := 1; i < len(hints); i++ {
		previous := hints[i-1]["position"].(position)
		current := hints[i]["position"].(position)
		if current.Line < previous.Line ||
			(current.Line == previous.Line && current.Character < previous.Character) {
			t.Fatalf("inlay hints are not ordered: %#v", hints)
		}
	}
}

func TestMacroCompletionOnlyIncludesMacros(t *testing.T) {
	reg, err := stdlib.LoadSources(map[string]string{
		"cli/cli.rn": `#flag(tree: SyntaxFile, context: MacroContext, short: String) -> SyntaxFile => tree

parse(value: String) -> String => value
`,
	})
	if err != nil {
		t.Fatalf("LoadSources() error = %v", err)
	}
	items := stdlibMacroCompletion(&checker.Info{Stdlib: reg}, "cli")
	if len(items) != 1 || items[0]["label"] != "flag" {
		t.Fatalf("completion items = %#v, want only flag", items)
	}
	module, ok := annotationCompletionModule("#cli.fl", position{Line: 0, Character: 7})
	if !ok || module != "cli" {
		t.Fatalf("annotation module = %q, %v", module, ok)
	}
}

func TestLocalMacroCompletionStaysOutOfRuntimeCompletion(t *testing.T) {
	src := `@syntax

#tag(tree: SyntaxFile, context: MacroContext, name: String) -> SyntaxFile => tree

main() => null
`
	uri := "file:///tmp/macro_completion.rn"
	session := NewSession()
	session.SetDocument(uri, src)

	annotationItems := macroCompletion(mustAnalyzeProgram(t, session, uri))
	if !completionContains(annotationItems, "tag") {
		t.Fatalf("annotation completion = %#v, want tag", annotationItems)
	}
	if !looksLikeAnnotationCompletion("#ta", position{Line: 0, Character: 3}) {
		t.Fatal("#ta was not recognized as macro completion")
	}
	runtimeItems := globalCompletion(mustAnalyzeProgram(t, session, uri))
	if completionContains(runtimeItems, "tag") {
		t.Fatalf("runtime completion = %#v, macro tag must be hidden", runtimeItems)
	}
}

func TestExpandedMacroShowsFinalDocument(t *testing.T) {
	src := `@syntax

#rename(
  tree: SyntaxFile,
  context: MacroContext,
  name: String
) -> SyntaxFile => {
  current := tree.types[0]
  renamed := SyntaxStruct {
    id: current.id
    name: name
    private: current.private
    generics: current.generics
    annotations: current.annotations
    fields: current.fields
    methods: current.methods
    sourcePath: current.sourcePath
  }
  SyntaxFile {
    types: [renamed]
    enums: tree.enums
    functions: tree.functions
  }
}

#rename("Generated")
Original: {
  value: Int
}
`
	uri := "file:///tmp/macro_expansion.rn"
	session := NewSession()
	session.SetDocument(uri, src)

	result := session.server.expandedMacro(uri).(map[string]any)
	if result["error"] != nil {
		t.Fatalf("expandedMacro() error = %v", result["error"])
	}
	source := result["source"].(string)
	if !strings.Contains(source, "Generated: {") {
		t.Fatalf("expanded source = %q, want generated declaration", source)
	}
	if strings.Contains(source, "#rename") || strings.Contains(source, "Original: {") {
		t.Fatalf("expanded source = %q, want consumed macro removed", source)
	}
}

func TestCodeLensIncludesMacroExpansion(t *testing.T) {
	uri := "file:///tmp/macro_lens.rn"
	src := `#macro.renameDeclaration("Generated")
Original: {
  value: Int
}
`
	s := &server{docs: map[string]string{uri: src}}

	lenses := s.codeLenses(uri).([]map[string]any)
	if len(lenses) != 1 {
		t.Fatalf("code lenses = %d, want 1: %#v", len(lenses), lenses)
	}
	command := lenses[0]["command"].(map[string]any)
	if command["command"] != "rune.showMacroExpansion" {
		t.Fatalf("command = %#v, want macro expansion command", command)
	}
	start := lenses[0]["range"].(map[string]any)["start"].(position)
	if start.Line != 0 || start.Character != 0 {
		t.Fatalf("lens start = %+v, want line 0 char 0", start)
	}
}

func containsHintLabel(hints []map[string]any, label string) bool {
	for _, hint := range hints {
		if hint["label"] == label {
			return true
		}
	}
	return false
}

func mustAnalyzeProgram(t *testing.T, session *Session, uri string) *compiler.Program {
	t.Helper()
	prog, diags := session.server.analyze(uri)
	if prog == nil || len(diags) > 0 {
		t.Fatalf("analyze() = %#v, %v", prog, diags)
	}
	return prog
}
