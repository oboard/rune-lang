package format

import (
	"testing"

	"github.com/oboard/rune-lang/internal/parser"
)

func TestAnonymousObjectLiteralFormatting(t *testing.T) {
	src := `main()=>{obj:={name:"Alice",age:30,greet()=>@io.println(.greetText()),nextAge()=>.age+1,greetText()=> "Hello, my name is "+.name}@io.println(obj.name)}`
	file, errs := parser.Parse(src)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}

	got := File(file)
	want := `main() => {
  obj := {
    name: "Alice",
    age: 30,

    greet() => @io.println(.greetText())
    nextAge() => .age + 1
    greetText() => "Hello, my name is " + .name
  }

  @io.println(obj.name)
}
`
	if got != want {
		t.Fatalf("File() =\n%s\nwant:\n%s", got, want)
	}
}

func TestFunctionReturningAnonymousObjectFormatting(t *testing.T) {
	src := `left(x:Input)=>({k:x.a+1,onlyLeft:true})`
	file, errs := parser.Parse(src)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}

	got := File(file)
	want := `left(x: Input) => ({
  k: x.a + 1,
  onlyLeft: true,
})
`
	if got != want {
		t.Fatalf("File() =\n%s\nwant:\n%s", got, want)
	}
	if _, errs := parser.Parse(got); len(errs) > 0 {
		t.Fatalf("formatted source does not parse: %v\n%s", errs, got)
	}
}

func TestSourcePreservesLineComments(t *testing.T) {
	src := `main()=>{
// construct a user
obj:={
name:"Alice", // display name
age:30, // years old
greet()=>@io.println(.name) // print greeting
}
@io.println(obj.name) // prints "Alice"
}`
	file, errs := parser.Parse(src)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}

	got := Source(file, src)
	want := `main() => {
  // construct a user
  obj := {
    name: "Alice", // display name
    age: 30, // years old

    greet() => @io.println(.name) // print greeting
  }

  @io.println(obj.name) // prints "Alice"
}
`
	if got != want {
		t.Fatalf("Source() =\n%s\nwant:\n%s", got, want)
	}
}
