package astprint

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/parser"
)

func TestPrintSourceRoundtrips(t *testing.T) {
	source := `add(a: Int32, b: Int32) -> Int32 => a + b

main() => {
  i :=: 0
  * (i < 3) {
    if (i == 1) {
      i = i + 1
      continue
    }
    i = i + 1
  }
  for (const value of [1, 2]) {
    @io.println(value)
  }
}
`
	file, errs := parser.Parse(source)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	printed := File(file)
	if !strings.Contains(printed, "* (i < 3)") {
		t.Fatalf("printed source missing loop:\n%s", printed)
	}
	if !strings.Contains(printed, "for (const value of [1, 2])") {
		t.Fatalf("printed source missing for-of:\n%s", printed)
	}
	if _, errs := parser.Parse(printed); len(errs) > 0 {
		t.Fatalf("reparse errors = %v\nsource:\n%s", errs, printed)
	}
}

func TestPrintReturn(t *testing.T) {
	file, errs := parser.Parse(`choose() -> Int32 => {
  return 42
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	printed := File(file)
	if !strings.Contains(printed, "return 42") {
		t.Fatalf("printed source missing return:\n%s", printed)
	}
}

func TestPrintLoopSymbols(t *testing.T) {
	file, errs := parser.Parse(`main() => {
  i :=: 0
  * (i < 3) {
    i = i + 1
    (i == 1) ? { ^> } : { }
    (i == 2) ? { ^ } : { }
  }
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	printed := File(file)
	for _, want := range []string{"i :=: 0", "* (i < 3)"} {
		if !strings.Contains(printed, want) {
			t.Fatalf("printed source missing %q:\n%s", want, printed)
		}
	}
	var sawBreak, sawContinue bool
	for _, line := range strings.Split(printed, "\n") {
		switch strings.TrimSpace(line) {
		case "^":
			sawBreak = true
		case "^>":
			sawContinue = true
		}
	}
	if !sawBreak || !sawContinue {
		t.Fatalf("printed source missing loop control symbols:\n%s", printed)
	}
	if _, errs := parser.Parse(printed); len(errs) > 0 {
		t.Fatalf("reparse errors = %v\nsource:\n%s", errs, printed)
	}
}
