---
layout: home

hero:
  name: Rune
  text: Expression-oriented language toolchain
  tagline: Parse, check, interpret, compile to Go, and emit TypeScript from a compact language surface.
  image:
    src: /rune-icon.svg
    alt: Rune
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
    - theme: alt
      text: Language
      link: /language/fundamentals

features:
  - title: Compact syntax
    details: Functions, blocks, objects, arrays, pattern bodies, and match expressions share a small expression-first model.
  - title: Multiple runtimes
    details: Run programs through the interpreter, compile to Go, or emit TypeScript for DOM-style targets.
  - title: Built-in tooling
    details: The repository includes parser, checker, formatter, LSP, REPL, core stubs, editor integration, and core library docs.
---

## Overview

Rune is an expression-oriented language toolchain written in Go. It parses and
checks Rune source, lowers it to a shared IR, and can interpret it, compile it
to Go, or emit TypeScript for DOM-style programs.

```rune
add(a: Int, b: Int) -> Int => a + b

main() => {
  @io.println(add(1, 2))
}
```

The language keeps the surface small: functions map parameters to expressions,
blocks return their final expression, data is represented with arrays and
record-like objects, and control flow is built from match expressions, pattern
bodies, ternary expressions, and ordinary function calls.

## Documentation

- [Getting Started](/guide/getting-started) shows the local commands.
- [Fundamentals](/language/fundamentals) documents the full supported syntax.
- [Core Library](/language/core-library) lists the built-in modules.
- [GitHub: oboard/rune-lang](https://github.com/oboard/rune-lang) hosts the
  source code.

Chinese documentation is available at [简体中文](/zh/).
