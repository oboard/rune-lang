# Rune Language

A minimal expression-oriented programming language built around pattern selection, implicit returns, and symbolic structure.

## Current MVP

This repository now contains the first Go-based Rune toolchain slice:

```text
Rune source
  -> lexer/parser
  -> AST
  -> semantic/type check
  -> Go source
  -> go run / go build
```

Implemented commands:

```sh
go run ./cmd/rune check examples/fib.rn
go run ./cmd/rune fmt examples/fib.rn
go run ./cmd/rune run examples/fib.rn
go run ./cmd/rune build -o /tmp/rune-fib examples/fib.rn
go run ./cmd/rune lsp
```

Start the local development environment:

```sh
scripts/dev.sh
scripts/dev.sh --shell
scripts/dev.sh --vscode
```

Install the CLI locally:

```sh
go install ./cmd/rune
```

The MVP compiler currently supports:

* function mappings: `name(args) => expr`
* `Int`, `String`, `Bool`, and `Void`
* explicit return annotations: `name(args) -> Type => expr`
* block bodies with final-expression returns
* immutable and mutable bindings: `:=`, `~=`
* assignment: `=`
* arithmetic and comparison expressions
* struct declarations: `Name: { field: Type }`
* struct literals and field access: `User { id: 1 }`, `user.id`
* struct methods and implicit receivers: `isAdult() => .age >= 18`
* single-parameter pattern blocks with literals, comparisons, and `_`
* calls to Rune functions
* core library declarations under `/core`: `array`, `map`, `io`, and `go`
* declared I/O calls: `@io.println`, `@io.print`, `@io.printf`
* inline Go FFI: `@go.import`, `@go.stmt`, and `@go.expr`

The Go backend prefixes Rune-defined identifiers with `__` during code
generation to avoid collisions with Go keywords and runtime names. Rune
`main` is generated as `__main`, with a small Go `main` wrapper as the process
entrypoint.

Inline Go FFI is intentionally string-based for the first implementation, so
Rune does not need to parse Go syntax. Use `$name` inside FFI strings to refer
to Rune identifiers after backend name mangling:

```rune
@go.import("fmt")

isAdult(age: Int) -> Bool => @go.expr("$age >= 18")

main() => {
  name := "oboard"
  @go.stmt("fmt.Println($name)")
  @io.println(isAdult(22))
}
```

Standard library calls are not compiler-invented names. The checker and Go
backend load module declarations from `/core/<module>/<module>.rn` stub files;
an undeclared
module call such as `@fmt.println(...)` is rejected.

Editor scaffolding is included under:

* `vscode-rune/` for TextMate highlighting, snippets, and `rune lsp`
* `tree-sitter-rune/` as the later incremental parser starting point

⸻

1. Introduction

Rune is a modern programming language designed around one core idea:

Most programming syntax is structural noise.

Traditional languages rely heavily on ceremonial keywords:

fn
function
return
if
else
match
import
class
new

Rune removes most of them and replaces them with:

* expressions
* pattern selection
* symbolic structure
* implicit semantics

Rune aims to be:

* concise
* expressive
* parser-friendly
* IDE-friendly
* compiler-friendly
* readable at scale

It is not a symbolic joke language like Brainfuck or APL.

Instead, Rune is intended to be a practical modern systems/application language.

⸻

1. Core Philosophy

Rune is built on several principles.

⸻

2.1 Functions Are Mappings

A function is defined as:

name(args) => expression

Example:

add(a: Int, b: Int) => a + b

There is no:

* fn
* function
* def
* return

The expression itself is the return value.

⸻

2.2 Control Flow Is Pattern Selection

Rune does not use traditional if/else.

Instead, all branching is expressed through pattern selection.

Example:

abs(x: Int) => {
    <0 => -x
    _  => x
}

This means:

match x {
    <0 => -x
    _  => x
}

without explicitly writing match.

⸻

2.3 One Function = One Definition

Rune does not allow multi-definition functions.

Invalid:

fib(0) => 0
fib(1) => 1

Correct:

fib(n: Int) => {
    0 => 0
    1 => 1
    _ => fib(n - 1) + fib(n - 2)
}

This keeps:

* AST structure simple
* tooling predictable
* symbol ownership explicit
* incremental compilation efficient

⸻

2.4 Expressions Over Statements

Everything possible should be an expression.

Blocks return values.

Pattern branches return values.

Functions return values.

There are very few statement-like constructs.

⸻

1. Imports

Rune uses symbolic path imports.

@std.io
@std.math
@net.http.Client

Optional aliasing:

@std.io as io

Imports are declarations, not executable statements.

⸻

1. Variables

Immutable bindings:

x := 10
name := "oboard"

Mutable bindings:

count ~= 0

Mutation:

count = count + 1

⸻

1. Functions

⸻

5.1 Single Expression

square(x: Int) => x * x

⸻

5.2 Block Body

sum(a: Int, b: Int) => {
    result := a + b
    result
}

The final expression becomes the return value.

⸻

1. Pattern Bodies

If a function body contains pattern branches, the parameter list automatically becomes the match target.

⸻

6.1 Single Parameter

abs(x: Int) => {
    <0 => -x
    _  => x
}

Equivalent semantic form:

match x

⸻

6.2 Multiple Parameters

point(x: Int, y: Int) => {
    (0, 0) => "origin"
    (_, 0) => "x-axis"
    (0,_) => "y-axis"
    _      => "normal"
}

Equivalent semantic form:

match (x, y)

⸻

1. Patterns

Rune supports:

⸻

7.1 Literal Patterns

0
1
"hello"
true

⸻

7.2 Wildcard

_

Matches anything.

⸻

7.3 Comparison Patterns

<0
>=18
>100

⸻

7.4 Tuple Patterns

(x, y)

⸻

7.5 Object Patterns

{ name, age }

⸻

7.6 Enum Patterns

Ok(v)
Err(e)

⸻

1. Types

⸻

8.1 Struct Types

User: {
    id: Int
    name: String
    age: Int
}

⸻

8.2 Algebraic Types

Result[T, E]: {
    Ok(T)
    Err(E)
}

⸻

1. Lambdas

Single parameter:

x => x * 2

Multiple parameters:

(a, b) => a + b

⸻

1. Pipelines

Rune supports pipeline composition.

data
    |> filter(x => x > 0)
    |> map(x => x * 2)
    |> sum()

⸻

1. Error Propagation

Rune uses postfix ?.

read(path: String) => {
    text := fs.read(path)?
    parse(text)?
}

⸻

1. Methods

Rune avoids traditional classes.

Methods are implemented through extension scopes.

Point: {
    x: Int
    y: Int
}
Point:len(self) => math.sqrt(self.x *self.x + self.y* self.y)

⸻

1. Comments

// single line
/*
multi line
*/

⸻

1. Example Program

User: {
    id: Int
    name: String
    age: Int
}
abs(x: Int) => {
    <0 => -x
    _=> x
}
fib(n: Int) => {
    0 => 0
    1 => 1
    _ => fib(n - 1) + fib(n - 2)
}
isAdult(user: User) => user.age {
    >=18 => true
    _    => false
}
main() => {
    user := User {
        id: 1
        name: "Luo Yuhang"
        age: 22
    }
    @io.println(user.name)
    @io.println(abs(-10))
    @io.println(fib(10))
}

⸻

1. Design Goals

Rune prioritizes:

* low syntax noise
* unified expression semantics
* minimal keywords
* strong pattern matching
* modern tooling support
* deterministic parsing
* compiler simplicity
* readability under scale

Rune intentionally avoids:

* statement-heavy syntax
* keyword ceremony
* declaration duplication
* implicit global mutation
* parser ambiguity explosion

⸻

1. Language Summary

Rune can be summarized in two rules:

function(args) => expression

and:

value {
    pattern => expression
}

Everything else emerges from these two structures.
