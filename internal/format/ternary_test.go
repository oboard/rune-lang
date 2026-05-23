package format

import (
	"testing"

	"github.com/oboard/rune-lang/internal/parser"
)

func TestTernaryExpressionFormatting(t *testing.T) {
	file, errs := parser.Parse(`main()=>{value:=flag?1:other?2:3 total:=(flag?1:2)+3}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}

	got := File(file)
	want := `main() => {
  value := flag ? 1 : other ? 2 : 3
  total := (flag ? 1 : 2) + 3
}
`
	if got != want {
		t.Fatalf("File() =\n%s\nwant:\n%s", got, want)
	}
}
