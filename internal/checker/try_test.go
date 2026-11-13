package checker

import (
	"testing"

	"github.com/oboard/rune-lang/internal/parser"
)

func TestTryCatchFinallyChecks(t *testing.T) {
	file, errs := parser.Parse(`run() => {
  try {
    throw "boom"
  } catch (err) {
    @io.println(err)
  } finally {
    @io.println("done")
  }
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if _, diags := Check(file); len(diags) > 0 {
		t.Fatalf("Check() diagnostics = %v", diags)
	}
}

func TestTryBranchesThatReturnSatisfyFunctionReturn(t *testing.T) {
	file, errs := parser.Parse(`choose(flag: Bool) -> Int32 => {
  try {
    throw "boom"
  } catch (err) {
    return 2
  }
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("Check() diagnostics = %v", diags)
	}
	if got := info.FunctionDecls[file.Functions[0]].Return; got != Int32 {
		t.Fatalf("return type = %s, want Int32", got)
	}
}
