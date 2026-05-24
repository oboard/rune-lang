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

func TestMultilineTernaryCalleeFormatting(t *testing.T) {
	file, errs := parser.Parse(`fun(flag)=>{(flag?(x)=>{k:x.a+1}:(y)=>{k:y.b+1})(value).k}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}

	got := File(file)
	want := `fun(flag) => {
  (flag ? (x) => {
    k: x.a + 1,
  } : (y) => {
    k: y.b + 1,
  })(value).k
}
`
	if got != want {
		t.Fatalf("File() =\n%s\nwant:\n%s", got, want)
	}
	if _, errs := parser.Parse(got); len(errs) > 0 {
		t.Fatalf("formatted source does not parse: %v\n%s", errs, got)
	}
}

func TestTernarySelectorReceiverFormatting(t *testing.T) {
	file, errs := parser.Parse(`main()=>{(flag?left:right).map()}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}

	got := File(file)
	want := `main() => {
  (flag ? left : right).map()
}
`
	if got != want {
		t.Fatalf("File() =\n%s\nwant:\n%s", got, want)
	}
	if _, errs := parser.Parse(got); len(errs) > 0 {
		t.Fatalf("formatted source does not parse: %v\n%s", errs, got)
	}
}
