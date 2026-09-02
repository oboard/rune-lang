package macro

import (
	"testing"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/parser"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func runCliMainExpand(t *testing.T, src string) {
	t.Helper()
	reg, err := stdlib.LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault() error = %v", err)
	}
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("Parse() errors = %v", parseErrs)
	}
	info, diags := checker.CheckWithStdlib(file, reg)
	for _, d := range diags {
		t.Fatalf("diag: %v", d.Message)
	}
	changed, macroDiags := Expand(file, info)
	for _, d := range macroDiags {
		t.Fatalf("expand diag: %v", d.Message)
	}
	if !changed {
		t.Fatal("no change")
	}
}

func TestCliMainExpansionExact(t *testing.T) {
	runCliMainExpand(t, `@syntax

#cli.command("demo", "Demo CLI", "1.0.0")

#cli.option("o", "FILE", "write output", "dist/app")
Args: {
  #cli.arg("target name")
  target: String
}

#cli.main
main(args: Args) => args.target
`)
}
