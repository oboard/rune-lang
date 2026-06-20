package macro

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/parser"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func TestExpandRunsRuneMacroInAnnotationOrder(t *testing.T) {
	reg, err := stdlib.LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault() error = %v", err)
	}
	file, parseErrs := parser.Parse(`#macro.renameDeclaration("Intermediate")
#macro.renameDeclaration("FinalArgs")
Args: {
  verbose: Bool
}
`)
	if len(parseErrs) > 0 {
		t.Fatalf("Parse() errors = %v", parseErrs)
	}
	info, diags := checker.CheckWithStdlib(file, reg)
	if len(diags) > 0 {
		t.Fatalf("CheckWithStdlib() diagnostics = %v", diags)
	}
	changed, macroDiags := Expand(file, info)
	if len(macroDiags) > 0 {
		t.Fatalf("Expand() diagnostics = %v", macroDiags)
	}
	if !changed || file.Types[0].Name != "FinalArgs" {
		t.Fatalf("expanded type name = %q, changed = %v", file.Types[0].Name, changed)
	}
}

func TestExpandRejectsCompileTimeIO(t *testing.T) {
	reg, err := stdlib.LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault() error = %v", err)
	}
	file, parseErrs := parser.Parse(`@syntax

#bad(tree: SyntaxFile, context: MacroContext) -> SyntaxFile => {
  @io.println("side effect")
  tree
}

#bad
Args: {
  verbose: Bool
}
`)
	if len(parseErrs) > 0 {
		t.Fatalf("Parse() errors = %v", parseErrs)
	}
	info, diags := checker.CheckWithStdlib(file, reg)
	if len(diags) != 1 || !strings.Contains(diags[0].Message, "calls impure function @io.println") {
		t.Fatalf("CheckWithStdlib() diagnostics = %v, want static purity error", diags)
	}
	changed, macroDiags := Expand(file, info)
	if changed || len(macroDiags) != 1 || !strings.Contains(macroDiags[0].Message, "compile-time macro cannot call @io.println") {
		t.Fatalf("Expand() = %v, %v", changed, macroDiags)
	}
}

func TestSyntaxFileOmitsMacroFunctions(t *testing.T) {
	file, parseErrs := parser.Parse(`@syntax

#transform(tree: SyntaxFile, context: MacroContext) -> SyntaxFile => tree

run() => null
`)
	if len(parseErrs) > 0 {
		t.Fatalf("Parse() errors = %v", parseErrs)
	}
	value := syntaxFileValue(file, newSyntaxRefs())
	root, err := expectStruct(value, "SyntaxFile")
	if err != nil {
		t.Fatal(err)
	}
	functions, err := structArrayField(root, "functions")
	if err != nil {
		t.Fatal(err)
	}
	if len(functions) != 1 {
		t.Fatalf("SyntaxFile.functions = %#v, want only the runtime function", functions)
	}
	fn, err := expectStruct(functions[0], "SyntaxFunction")
	if err != nil {
		t.Fatal(err)
	}
	if fn.Fields["name"] != "run" {
		t.Fatalf("SyntaxFile.functions[0].name = %#v, want run", fn.Fields["name"])
	}
}

func TestExpandSyntaxMacroReplacesDeclarationTree(t *testing.T) {
	reg, err := stdlib.LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault() error = %v", err)
	}
	file, parseErrs := parser.Parse(`@syntax

#renameFirst(
  tree: SyntaxFile,
  context: MacroContext,
  name: String
) -> SyntaxFile => {
  current := tree.types[0]
  selectedName := context.targetID == current.id ? name : current.name
	  generatedField := SyntaxField {
	    id: "",
	    name: "generated",
	    private: false,
	    annotations: [],
	    type: current.fields[0].type
	  }
	  renamed := SyntaxStruct {
	    id: current.id,
	    name: selectedName,
	    private: current.private,
	    generics: current.generics,
	    annotations: current.annotations,
	    fields: [...current.fields, generatedField],
	    methods: current.methods,
	    sourcePath: current.sourcePath
	  }
	  SyntaxFile {
	    types: [renamed],
	    enums: tree.enums,
	    functions: tree.functions
	  }
}

#renameFirst("FinalArgs")
Args: {
  verbose: Bool
}

run() => null
`)
	if len(parseErrs) > 0 {
		t.Fatalf("Parse() errors = %v", parseErrs)
	}
	info, diags := checker.CheckWithStdlib(file, reg)
	if len(diags) > 0 {
		t.Fatalf("CheckWithStdlib() diagnostics = %v", diags)
	}
	changed, macroDiags := Expand(file, info)
	if len(macroDiags) > 0 {
		t.Fatalf("Expand() diagnostics = %v", macroDiags)
	}
	if !changed || len(file.Types) != 1 || file.Types[0].Name != "FinalArgs" {
		t.Fatalf("expanded types = %#v, changed = %v", file.Types, changed)
	}
	if len(file.Types[0].Fields) != 2 || file.Types[0].Fields[1].Name != "generated" {
		t.Fatalf("expanded fields = %#v, want generated field", file.Types[0].Fields)
	}
	if len(file.Functions) != 2 || !file.Functions[0].Macro || file.Functions[1].Name != "run" {
		t.Fatalf("expanded functions = %#v, want hidden macro plus run", file.Functions)
	}
}

func TestCoreRenameMacroHandlesEveryDeclarationTarget(t *testing.T) {
	reg, err := stdlib.LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault() error = %v", err)
	}
	file, parseErrs := parser.Parse(`#macro.renameDeclaration("RenamedArgs")
Args: {
  #macro.renameDeclaration("renamedField")
  value: Int

  #macro.renameDeclaration("renamedMethod")
  read() -> Int => .value
}

#macro.renameDeclaration("RenamedCommand")
Command: {
  #macro.renameDeclaration("RenamedRun")
  Run
}

#macro.renameDeclaration("renamedFunction")
execute() => null
`)
	if len(parseErrs) > 0 {
		t.Fatalf("Parse() errors = %v", parseErrs)
	}
	info, diags := checker.CheckWithStdlib(file, reg)
	if len(diags) > 0 {
		t.Fatalf("CheckWithStdlib() diagnostics = %v", diags)
	}
	changed, macroDiags := Expand(file, info)
	if len(macroDiags) > 0 {
		t.Fatalf("Expand() diagnostics = %v", macroDiags)
	}
	if !changed ||
		file.Types[0].Name != "RenamedArgs" ||
		file.Types[0].Fields[0].Name != "renamedField" ||
		file.Types[0].Methods[0].Name != "renamedMethod" ||
		file.Enums[0].Name != "RenamedCommand" ||
		file.Enums[0].Members[0].Name != "RenamedRun" ||
		file.Functions[0].Name != "renamedFunction" {
		t.Fatalf("expanded file = %#v", file)
	}
}
