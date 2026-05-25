# Language Fundamentals

This page follows the shape of a language-fundamentals reference and documents
the syntax that the current Rune parser, checker, examples, and tests support.

## Source Files

A Rune source file contains top-level declarations:

```rune
@go.import("fmt")

User: {
  name: String
  age: Int
}

isAdult(user: User) -> Bool => user.age >= 18

? "adult check" {
  @assert.eq(isAdult(User { name: "Ada", age: 36 }), true)
}
```

Supported top-level forms are:

- `@go.import("pkg")` for Go backend imports.
- `Name[T]: { ... }` type declarations.
- `name[T](params) -> Return => body` function declarations.
- `? "name" { ... }` test declarations.

There is no general user import syntax yet. Built-in modules are addressed with
`@module.function(...)` and are loaded from `core/`.

## Lexical Rules

Identifiers start with a Unicode letter or `_` and continue with letters,
digits, or `_`.

```rune
value := 1
_private := "ok"
名字 := "Rune"
```

Line comments and block comments are supported:

```rune
// one line
/*
  multiple lines
*/
```

Newlines separate declarations and statements. Commas are accepted in argument
lists, arrays, object literals, struct literals, and match branches where the
parser expects them. Rune style usually prefers newlines for object fields and
commas for compact lists.

## Literals

Rune has scalar literals for integers, doubles, big integers, strings, booleans,
and null:

```rune
intValue := 42
doubleValue := 6.25e-1
bigValue := 9007199254740993n
text := "hello\nRune"
ok := true
missing := null
```

Integer literals have type `Int`. Decimal or exponent literals have type
`Double`. Literals ending in `n` have type `BigInt`. `true` and `false` have
type `Bool`. `null` has type `Null` and can flow into nullable types such as
`Int?` or `String?`.

## Built-In Types

The user-facing built-in scalar and special types are:

```text
Int
Double
BigInt
String
Bool
Void
Object
Symbol
HTMLElement
Any
Dynamic
Data
Error
Result[T, E]
Task[T]
```

`Any` and `Dynamic` are accepted in declarations and are treated as dynamic
types by the checker. Prefer concrete types when possible.

Nullable types use a suffix `?`:

```rune
maybeName(value: String?) -> String? => value

main() => {
  @io.println(maybeName(null))
  @io.println(maybeName("Rune"))
}
```

Generic type application uses brackets:

```rune
values: Array[Int]
scores: Map[String, Int]
seen: Set[String]
```

Function types use `(params) -> Return`. Parameter names inside function types
are optional documentation for call sites and callback checking:

```rune
callback: (value: Int, index?: Int, array?: Array[Int]) -> Any
```

## Operators

Unary operators:

```rune
-value
!flag
~value
```

Postfix increment:

```rune
count++
```

Binary operators, from low to high precedence:

```text
||
&&
|
^
&
== !=
< <= > >=
<< >> >>>
+ -
* / %
```

Numeric arithmetic requires both sides to have the same numeric type. `%` is
available for `Int` and `BigInt`, not `Double`. `+` also concatenates strings,
but both sides must be `String` when either side is a string.

```rune
@assert.eq(1 + 2 * 3, 7)
@assert.eq("ru" + "ne", "rune")
@assert.eq(22 / 7, 3)
@assert.eq(22n % 7n, 1n)
```

`&&` and `||` short-circuit and require boolean operands. Ordered comparison
works on matching `Int`, `Double`, `BigInt`, or `String` values.

Bitwise operators `~`, `&`, `|`, `^`, `<<`, `>>`, and `>>>` require matching
integer operands. `>>>` requires an unsigned integer left operand.

## Bindings and Blocks

Blocks are expressions. A block returns the final expression, or `Void` if the
last statement has no value:

```rune
sum(a: Int, b: Int) -> Int => {
  result := a + b
  result
}
```

Bindings are written inside blocks:

```rune
value := 1
mutable ~= 1
signal $= 0
```

`:=` creates an ordinary local binding. `~=` marks a binding as mutable intent.
`$=` creates a signal binding for reactive code. Assignment uses `=`:

```rune
count ~= 0
count = count + 1
```

Assignment can also appear as an expression, especially in event handlers:

```rune
<button @click={count = count + 1}>Add</button>
```

## Functions

Functions map parameters to a body expression:

```rune
add(a: Int, b: Int) -> Int => a + b
```

Parameter and return annotations are optional when inference has enough
information:

```rune
fib(n) => n {
  0 => 0
  1 => 1
  _ => fib(n - 1) + fib(n - 2)
}
```

Generic functions declare type parameters after the name:

```rune
identity[T](value: T) -> T => value
```

`main` is special: it is always checked as returning `Void`.

## Routines and Result

Prefix a function with `~` to mark it as a routine:

```rune
~ test(count: Int) => {
  @io.println("Hello World" + count.toString())
}

main() => {
  test(1)
  test(2)
  test(3)
  @io.println("Hello World")
}
```

Calling a routine from an ordinary function starts it and returns a `Task[T]`.
The generated program waits for outstanding tasks before exit. Calling a
routine from inside another routine waits automatically, so source code does
not use an `await` keyword.

Errors use the built-in `Result[T, E]` enum shape:

```rune
Result[T, E]: {
  Ok(value: T)
  Err(error: E)
}

Error: {
  code: Int
  message: String
  cause: Error?
}
```

Inside a routine, postfix `?` unwraps `Result[T, E]`. On `Err`, it returns early
and lifts the error into the routine return type:

```rune
~ read() => {
  file := @fs.readFile("1.txt")?
  file
}
```

The inferred type is `~ read() -> Result[@io.Data, Error]`.

## Lambdas

Lambda parameters must be parenthesized:

```rune
values.map((value) => value * 2)
values.map((value: Int, index: Int) => value + index)
```

This is intentionally invalid:

```rune
values.map(value => value * 2)
```

Lambda bodies can be expressions or blocks:

```rune
handle := (value: Int) => {
  next := value + 1
  next
}
```

When a callback type allows more parameters than a lambda provides, Rune accepts
the shorter lambda. This is how array callbacks can use only `value` even though
the declared callback type also includes `index` and `array`.

## Pattern Bodies and Match

A function body can be a pattern block. This form is intended for
single-parameter functions:

```rune
fib(n: Int) -> Int => {
  0 => 0
  1 => 1
  _ => fib(n - 1) + fib(n - 2)
}
```

Any expression can be matched with `subject { pattern => expression }`:

```rune
label(value: Int) -> String => value {
  <0 => "negative"
  0 => "zero"
  _ => "positive"
}
```

Supported pattern forms are:

```rune
_          // wildcard
1          // Int literal
1.5        // Double literal
1n         // BigInt literal
"text"     // String literal
true       // Bool literal
false
null
< 10       // comparison pattern
<= 10
> 10
>= 10
(1, _)     // tuple pattern syntax, reserved for tuple-like subjects
Ok(value)  // Result constructor pattern
Err(error)
```

All non-void branches should return compatible types. Nested matches are just
expressions:

```rune
x {
  1 => y {
    2 => "both"
    _ => "only x"
  }
  _ => "none"
}
```

`Result` values can also be matched manually:

```rune
readUser("user.json") {
  Ok(user) => @io.println(user.name)
  Err(e) => @io.println(e.message)
}
```

## Ternary Expressions

Rune also supports C-style conditional expressions:

```rune
choose(flag: Bool) -> Int => flag ? 1 : 2
```

The condition must be `Bool`. The consequence and alternative branches must
unify to one result type. Ternaries short-circuit, so only the selected branch
is evaluated.

## Arrays

Array literals infer their element type:

```rune
empty := []
values := [1, 2, 3]
matrix := [[1, 2], [3, 4]]
nullable := [1, null, 3]
```

Spread inserts another array into a new literal:

```rune
middle := [2, 3]
values := [1, ...middle, 4]
```

Indexing uses an `Int` index:

```rune
first := values[0]
```

Array methods are declared in `core/array`:

```rune
values.push(4)
values.set(1, 20)
values.length()
values.isEmpty()
values.first()
values.last()
values.slice(1, 3)
values.clone()
values.reverse()
values.contains(20)
values.forEach((value, index) => @io.println(value))
doubled := values.map((value) => value * 2)
```

## Struct Types

Named record-like types are declared with `Name: { ... }`:

```rune
User: {
  id: Int
  name: String
  age: Int

  isAdult() -> Bool => .age >= 18
  label(prefix: String) -> String => prefix + .name
}
```

Struct literals require all declared fields:

```rune
user := User {
  id: 1
  name: "Ada"
  age: 36
}
```

Fields are read with selectors:

```rune
@io.println(user.name)
@io.println(user.isAdult())
```

Inside methods, `.field` is shorthand for `this.field`. The explicit `this`
binding is also available inside methods.

Generic type declarations are accepted:

```rune
Box[T]: {
  value: T
}
```

The most complete generic behavior today is in built-in generic containers such
as `Array[T]`, `Map[K, V]`, and `Set[T]`.

## Enum Types

Enums use the same top-level `Name: { ... }` declaration shape with integer
members:

```rune
Status: {
  Completed = 0
  Fail = 1
}
```

Enum members are read with selectors. A member has the enum type, so
`Status.Completed` is typed as `Status`:

```rune
status := Status.Completed
```

Enum values compare with other values of the same enum type and can be used in
match patterns:

```rune
statusText(status: Status) -> String => status {
  Status.Completed => "completed"
  Status.Fail => "fail"
  _ => "unknown"
}
```

## Anonymous Objects

Anonymous objects are expressions:

```rune
account := {
  name: "core"
  age: 4

  nextAge() -> Int => .age + 1
  title(prefix: String) -> String => prefix + .name
}
```

They infer closed object types. Two anonymous object types unify only when their
fields match as closed types. A named struct parameter can still accept an
anonymous object when the anonymous object structurally satisfies the expected
shape at the call site.

Function-valued object fields are callable:

```rune
@io.println(account.title("module:"))
```

`@json.stringify` omits function-valued fields when producing JSON.

## Modules and Calls

Module functions use `@module.name(...)`:

```rune
@io.println("hello")
json := @json.stringify({ name: "Rune" })
scores := @map.new("", 0)
```

The compiler does not invent module functions. Calls are accepted only when a
matching declaration exists in `core/<module>/<module>.rn`.

## Reactivity and Watch

Signal bindings use `$=`:

```rune
render() => {
  count $= 0
  double := count * 2

  count -> (old, new) => {
    @io.println(old)
    @io.println(new)
  }

  count = count + 1
}
```

`target -> handler` registers a watcher. A handler can be a zero-argument block
or a lambda with two parameters `(old, new)`.

Reactive array and object literals use `$[...]` and `${...}`:

```rune
items := $["Item 1", "Item 2"]
state := ${ count: 0 }
```

The TypeScript backend emits helper wrappers for signals and reactive arrays.

## XML and DOM Expressions

XML-like elements are expressions that return `HTMLElement`:

```rune
render() -> HTMLElement => {
  count $= 0

  <div class="counter">
    <p>Count: {count}</p>
    <button @click={count++}>Click Me</button>
  </div>
}
```

Supported XML syntax includes:

- Normal elements: `<div>...</div>`.
- Self-closing elements: `<input />`.
- String attributes: `class="counter"`.
- Bare attributes: `disabled`.
- Expression attributes: `value={count}`.
- Event attributes with `@`: `@click={handler}`.
- Text children.
- Embedded expressions: `{expr}`.
- Nested elements.

XML is intended for the TypeScript backend. The Go backend currently emits a
placeholder comment for XML expressions.

## Inline Go FFI

The Go backend supports inline Go through `@go`:

```rune
@go.import("fmt")

isAdult(age: Int) -> Bool => @go.expr("$age >= 18")

main() => {
  name := "Ada"
  @go.stmt("fmt.Println($name)")
  @io.println(isAdult(36))
}
```

`@go.import` is top-level only. `@go.stmt` and `@go.expr` can appear in
function bodies. `$name` inside the Go string is rewritten to the generated Go
identifier for the Rune binding.

## Tests

Test declarations start with `?`:

```rune
? "string split" {
  parts := "r,u,n,e".split(",")
  @assert.eq(parts[1], "u")
}
```

Tests are ordinary Rune blocks. The standard assertion helper is
`@assert.eq(actual, expected)`.

## Current Syntax Boundaries

Rune intentionally has a small surface today. There is no `if`, `for`, `while`,
`trait`, `impl`, `class`, package import, exception, or macro syntax in the
current parser. Use match expressions, ternary expressions, recursion, array
callbacks, and core modules for the supported equivalents.
