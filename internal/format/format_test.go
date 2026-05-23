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

func TestStructLiteralFormatting(t *testing.T) {
	src := `User:{id:Int name:String age:Int} main()=>{user:=User { id: 1, name: "oboard", age: 22 }}`
	file, errs := parser.Parse(src)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}

	got := File(file)
	want := `User: {
  id: Int
  name: String
  age: Int
}

main() => {
  user := User {
    id: 1
    name: "oboard"
    age: 22
  }
}
`
	if got != want {
		t.Fatalf("File() =\n%s\nwant:\n%s", got, want)
	}
}

func TestSignalWatchFormatting(t *testing.T) {
	src := `main()=>{count $= 0 double:=count*2 double->{@io.println(double)} count->(old,new)=>{@io.println(old) @io.println(new)} count=count+1}`
	file, errs := parser.Parse(src)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}

	got := File(file)
	want := `main() => {
  count $= 0
  double := count * 2
  double -> {
    @io.println(double)
  }
  count -> (old, new) => {
    @io.println(old)
    @io.println(new)
  }
  count = count + 1
}
`
	if got != want {
		t.Fatalf("File() =\n%s\nwant:\n%s", got, want)
	}
	if _, errs := parser.Parse(got); len(errs) > 0 {
		t.Fatalf("formatted source does not parse: %v\n%s", errs, got)
	}
}

func TestXMLElementFormatting(t *testing.T) {
	src := `render() => {
  list := ["Item 1", "Item 2", "Item 3"]

  <div>
    <h1>List Example</h1>
    <ul>
      {list.map((item) => (
          <li>{item}</li>
      ))}
    </ul>
  </div>
}`
	file, errs := parser.Parse(src)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}

	got := File(file)
	want := `render() => {
  list := ["Item 1", "Item 2", "Item 3"]

  <div>
    <h1>List Example</h1>
    <ul>
      {list.map((item) => (
          <li>{item}</li>
      ))}
    </ul>
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

func TestSourcePreservesCommentsAroundStructLiteralExpansion(t *testing.T) {
	src := `User:{id:Int name:String age:Int}
main()=>{
// create user
user:=User { id: 1, name: "oboard", age: 22 } // user literal
@io.println(user.name) // print name
}`
	file, errs := parser.Parse(src)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}

	got := Source(file, src)
	want := `User: {
  id: Int
  name: String
  age: Int
}

main() => {
  // create user
  user := User {
    id: 1
    name: "oboard"
    age: 22
  } // user literal
  @io.println(user.name) // print name
}
`
	if got != want {
		t.Fatalf("Source() =\n%s\nwant:\n%s", got, want)
	}
}

func TestSourcePreservesLeadingCommentsAcrossDeclarationReflow(t *testing.T) {
	src := `// file header
maxInt(a:Int,b:Int)->Int=>(a>b){true=>a false=>b}

// 001. Split Signature
lc001TwoSum(nums:Array[Int],target:Int)->Array[Int]=>nums`
	file, errs := parser.Parse(src)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}

	got := Source(file, src)
	want := `// file header
maxInt(a: Int, b: Int) -> Int => a > b {
  true => a
  false => b
}

// 001. Split Signature
lc001TwoSum(
  nums: Array[Int],
  target: Int
) -> Array[Int] => nums
`
	if got != want {
		t.Fatalf("Source() =\n%s\nwant:\n%s", got, want)
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

func TestSourceWrapsLongChainAndKeepsInlineComment(t *testing.T) {
	src := `main()=>{
arr.map((value)=>value*2).each((value,index)=>@io.println(value)) // prints 2, 4, 6, 8, 10, 12
}`
	file, errs := parser.Parse(src)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}

	got := Source(file, src)
	want := `main() => {
  arr.map((value) => value * 2)
       .each((value, index) => @io.println(value)) // prints 2, 4, 6, 8, 10, 12
}
`
	if got != want {
		t.Fatalf("Source() =\n%s\nwant:\n%s", got, want)
	}
	if _, errs := parser.Parse(got); len(errs) > 0 {
		t.Fatalf("formatted source does not parse: %v\n%s", errs, got)
	}
}

func TestSourceWrapsFunctionTypeSignature(t *testing.T) {
	src := `Array[T]: {
forEach(callbackfn: (value: T, index?: Int, array?: Array[T]) -> Void) => "%array.forEach"
}`
	file, errs := parser.Parse(src)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}

	got := Source(file, src)
	want := `Array[T]: {
  forEach(
    callbackfn: (
      value: T,
      index?: Int,
      array?: Array[T]
    ) -> Void
  ) => "%array.forEach"
}
`
	if got != want {
		t.Fatalf("Source() =\n%s\nwant:\n%s", got, want)
	}
	if _, errs := parser.Parse(got); len(errs) > 0 {
		t.Fatalf("formatted source does not parse: %v\n%s", errs, got)
	}
}
