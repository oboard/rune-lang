# Rune Language

Rune is an expression-oriented language toolchain written in Go. The current
implementation parses Rune source, checks types, lowers to IR, and can either
interpret, compile to Go, or emit TypeScript.

```text
Rune source
  -> lexer/parser
  -> AST
  -> semantic/type check
  -> IR
  -> interpreter, Go codegen, or TypeScript codegen
  -> go run / go build
```

The fastest path for Rune is still: build a practical language that compiles to
Go and reuses Go's runtime, GC, module system, and cross compilation.

## Quick Start

```sh
scripts/dev.sh
scripts/dev.sh --shell
scripts/dev.sh --vscode
```

Or run the toolchain directly:

```sh
go run ./cmd/rune check examples/fib.rn
go run ./cmd/rune fmt examples/fib.rn
go run ./cmd/rune run examples/fib.rn
go run ./cmd/rune build -o /tmp/rune-fib examples/fib.rn
go run ./cmd/rune ts examples/counter.rn
go run ./cmd/rune repl
go run ./cmd/rune lsp
```

Build the local CLI used by the VSCode extension:

```sh
go build -o .bin/rune ./cmd/rune
```

## CLI

```text
rune check <file.rn>   Parse and type-check a Rune program
rune fmt <file.rn>     Format a Rune source file
rune run <file.rn>     Compile and run a Rune program
rune build <file.rn>   Compile a Rune program to an executable
rune ts <file.rn>      Compile a Rune program to TypeScript
rune repl              Start the Rune REPL
rune lsp               Start the Rune language server
```

`rune lsp` serves over stdio by default and also accepts `--stdio`.

## Project Layout

```text
cmd/rune/              CLI entrypoint
internal/lexer/        Lexer
internal/parser/       Parser
internal/ast/          AST
internal/checker/      Type checking and inference
internal/ir/           Shared IR
internal/interpreter/  IR interpreter
internal/codegen/go/   Rune -> Go backend
internal/codegen/typescript/
                       Rune -> TypeScript backend
internal/format/       Formatter
internal/lsp/          Language server
internal/repl/         REPL
core/                  Rune core library stubs
examples/              Runnable and type-checking examples
selfhost/              Rune implementations of bootstrap components
vscode-rune/           VSCode extension
tree-sitter-rune/      Future Tree-sitter grammar scaffold
```

Rune files can import other Rune files with Dart-like file imports:

```rune
@"./helper.rn"

main() => @io.println(helper())
```

The self-hosted lexer lives at `selfhost/lexer/lexer.rn`; see
`examples/lexer_bootstrap.rn` and `tests/lexer_bootstrap.rn` for the current
bootstrap entry points.

## Language Snapshot

Rune functions are mappings from parameters to expressions:

```rune
add(a: Int, b: Int) -> Int => a + b

main() => {
  @io.println(add(1, 2))
}
```

Blocks return their final expression:

```rune
sum(a: Int, b: Int) => {
  result := a + b
  result
}
```

Pattern bodies are supported for single-parameter functions:

```rune
fib(n: Int) -> Int => {
  0 => 0
  1 => 1
  _ => fib(n - 1) + fib(n - 2)
}
```

Match expressions use the same pattern syntax:

```rune
value {
  true => "yes"
  false => "no"
}
```

## Types

The built-in scalar types are:

```text
Int
String
Bool
Void
HTMLElement
```

Struct types use symbolic object syntax:

```rune
User: {
  id: Int
  name: String
  age: Int
}

main() => {
  user := User {
    id: 1,
    name: "oboard",
    age: 22,
  }

  @io.println(user.name)
}
```

Methods live inside type declarations. Inside a method, `.field` means
`this.field`:

```rune
User: {
  age: Int

  isAdult() -> Bool => .age >= 18
}
```

Enum types use the same declaration shape with integer members:

```rune
Status: {
  Completed = 0
  Fail = 1
}

main() => {
  status := Status.Completed
  @io.println(status == Status.Fail)
}
```

Anonymous records are expressions:

```rune
obj := {
  name: "Alice",
  age: 30,

  nextAge() => .age + 1
  greetText() => "Hello, " + .name
}
```

## Type Inference

Rune uses a Hindley-Milner-style inference direction with closed anonymous
records. There is intentionally no row polymorphism: two anonymous record types
only unify when their fields match as closed types.

Function values can still be refined by call sites. If a function value only
uses `x.a` and is called with `{ b, z, a }`, the call site can refine the
parameter to the full argument shape.

Named struct arguments are preserved:

```rune
Return: {
  b: Int
  z: Bool
  a: Int
}

fun(flag) => {
  (flag {
    true => (x: Return) => {
      k: x.a + 1,
    }
    false => (y: Return) => {
      k: y.b + 1,
    }
  })({
    b: 2,
    z: false,
    a: 1,
  }).k
}
```

Here `x` and `y` are typed as `Return`; the anonymous argument is accepted
because it structurally satisfies `Return`, and Go codegen emits a `Return`
literal.

## Lambdas

Lambda parameters must be parenthesized:

```rune
arr.map((value) => value * 2)
```

This is invalid:

```rune
arr.map(value => value * 2)
```

Lambda parameter annotations are supported:

```rune
(value: Int) => value + 1
(user: User) => user.name
```

## Core Library

Core library calls are declaration-driven. The checker and backend load stubs
from `core/<module>/<module>.rn`; compiler-invented module calls are rejected.

Currently included modules:

```text
core/array
core/binary
core/buffer
core/compress
core/fs
core/go
core/io
core/iter
core/json
core/map
core/net
core/path
core/process
core/reader
core/set
core/stringbuffer
core/writer
```

I/O:

```rune
@io.print(value)
@io.println(value)
@io.printf(format, value)
```

Arrays:

```rune
main() => {
  arr := [1, 2, 3]
  arr.push(4)
  arr.each((value) => @io.println(value))
  doubled := arr.map((value) => value * 2)
  @io.println(doubled[0])
}
```

JSON:

```rune
json := @json.stringify({
  name: "Rune"
  version: 1
})
```

Inline Go FFI:

```rune
@go.import("fmt")

isAdult(age: Int) -> Bool => @go.expr("$age >= 18")

main() => {
  name := "oboard"
  @go.stmt("fmt.Println($name)")
  @io.println(isAdult(22))
}
```

FFI strings can reference Rune identifiers with `$name`; the Go backend expands
them after identifier mangling.

## Go Codegen

Rune-defined identifiers are prefixed with `__` in generated Go to avoid
collisions with Go keywords and runtime names. Rune `main` becomes `__main`,
with a small Go `main` wrapper as the process entrypoint.

Anonymous records are emitted as Go struct literals. Named structs remain named
Go structs. Function values are emitted as Go function values.

## TypeScript Codegen

The TypeScript backend emits DOM-oriented TypeScript from the shared IR. It
does not support `@go` FFI. XML elements are Rune expressions and embedded
`{expr}` children become text nodes unless the expression returns
`HTMLElement`.

```rune
render() -> HTMLElement => {
  count $= 0

  <div>
    <p>Count: {count}</p>
    <button @click={count++}>Click Me</button>
  </div>
}
```

Compile it with:

```sh
rune ts examples/counter.rn
```

## REPL

Start an interactive session:

```sh
rune repl
```

The REPL evaluates expressions and definitions through the interpreter path, so
it does not require a `main` function.

## VSCode

The VSCode extension in `vscode-rune/` provides:

* TextMate highlighting
* snippets
* `rune lsp` integration
* diagnostics
* hover
* completion
* go to definition
* rename
* document symbols
* formatting
* inlay hints for inferred function and lambda types
* Run / Debug code lenses for `main`

For local development:

```sh
scripts/dev.sh --vscode
```

Inlay hints show inferred parameter and return types, including anonymous
lambda types:

```rune
fun(flag) => {
  (flag {
    true => (x) => {
      k: x.a + 1,
    }
    false => (y) => {
      k: y.b + 1,
    }
  })({
    b: 2,
    z: false,
    a: 1,
  }).k
}
```

The editor can display `flag: Bool`, `x: { ... }`, and return hints inline.

## Examples

```sh
rune run examples/fib.rn
rune run examples/array.rn
rune run examples/anonymous_object.rn
rune run examples/complex_type.rn
rune run examples/ffi.rn
```

`examples/complex_type2.rn` is a static type-inference example. It intentionally
contains recursive functions that should be checked, not run.

## Design Notes

Rune currently prioritizes:

* expression-first syntax
* deterministic parsing
* strong editor support
* shared IR for interpreter and compiler
* Go backend pragmatism
* closed record types without row polymorphism

Rune intentionally does not yet implement LLVM, a custom VM, GC, JIT, package
manager, package imports, traits, or classes. Those can come after the language
core and LSP are stable.
