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
  value := (
    flag ? 1
    : other ? 2
    : 3
  )
  total := (
    flag ? 1
      : 2
  ) + 3
}
`
	if got != want {
		t.Fatalf("File() =\n%s\nwant:\n%s", got, want)
	}
	if _, errs := parser.Parse(got); len(errs) > 0 {
		t.Fatalf("formatted source does not parse: %v\n%s", errs, got)
	}
}

func TestMultilineTernaryCalleeFormatting(t *testing.T) {
	file, errs := parser.Parse(`fun(flag)=>{(flag?(x)=>{k:x.a+1}:(y)=>{k:y.b+1})(value).k}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}

	got := File(file)
	want := `fun(flag) => {
  (
    flag ? (x) => {
      k: x.a + 1
    }
      : (y) => {
        k: y.b + 1
      }
  )(value).k
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
  (
    flag ? left
      : right
  ).map()
}
`
	if got != want {
		t.Fatalf("File() =\n%s\nwant:\n%s", got, want)
	}
	if _, errs := parser.Parse(got); len(errs) > 0 {
		t.Fatalf("formatted source does not parse: %v\n%s", errs, got)
	}
}

func TestXMLTernaryChainStayFlat(t *testing.T) {
	file, errs := parser.Parse(`+ render() => {
  $score := 85
  <div>
    <h2>Grade</h2>
    <p>{($score >= 90) ? (<b>A+</b>) : ($score >= 80) ? (<b>A</b>) : ($score >= 70) ? (<b>B</b>) : ($score >= 60) ? (<b>C</b>) : (<em>F</em>)}</p>
  </div>
}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}

	got := File(file)
	want := `+ render() => {
  $score := 85

  <div>
    <h2>Grade</h2>
    <p>
      {(
        $score >= 90 ? <b>A+</b>
        : $score >= 80 ? <b>A</b>
        : $score >= 70 ? <b>B</b>
        : $score >= 60 ? <b>C</b>
        : <em>F</em>
      )}
    </p>
  </div>
}
`
	if got != want {
		t.Fatalf("File() =\n%s\nwant:\n%s", got, want)
	}
	if _, errs := parser.Parse(got); len(errs) > 0 {
		t.Fatalf("formatted source does not parse: %v\n%s", errs, got)
	}
}

func TestConditionalExpressionWithoutElseFormatting(t *testing.T) {
	file, errs := parser.Parse(`main()=>{isHelp ? handled = true}`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}

	got := File(file)
	want := `main() => {
  (
    isHelp ? handled = true
  )
}
`
	if got != want {
		t.Fatalf("File() =\n%s\nwant:\n%s", got, want)
	}
}
