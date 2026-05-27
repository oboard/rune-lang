# CLI and Editor

The Rune CLI lives at `cmd/rune`.

```text
rune check <path>      Parse and type-check a file or directory
rune fmt <path>        Format a file or directory (alias: format)
rune run <path>        Run a file, or a directory with one main
rune build <file.rn>   Compile a Rune program to an executable
rune ts <file.rn>      Compile a Rune program to TypeScript
rune repl              Start the Rune REPL
rune lsp               Start the Rune language server
```

Use it through `go run` during development:

```sh
go run ./cmd/rune check examples/fib.rn
go run ./cmd/rune check core
go run ./cmd/rune run examples/fib.rn
go run ./cmd/rune ts examples/counter.rn
```

Or build a binary:

```sh
go build -o .bin/rune ./cmd/rune
```

## Formatting

```sh
go run ./cmd/rune fmt examples/fib.rn
```

The formatter normalizes declarations, object fields, arrays, blocks, match
branches, XML expressions, and comments.

## REPL

```sh
go run ./cmd/rune repl
```

The REPL evaluates expressions and declarations through the interpreter path and
does not require a `main` function.

## Language Server

```sh
go run ./cmd/rune lsp
```

`rune lsp` serves over stdio by default and also accepts `--stdio`.

The VSCode extension in `vscode-rune/` provides:

- TextMate highlighting.
- Snippets.
- Diagnostics.
- Hover.
- Completion.
- Go to definition.
- Rename.
- Document symbols.
- Formatting.
- Inlay hints for inferred function and lambda types.
- Run and debug code lenses for `main`.

Local VSCode development:

```sh
scripts/dev.sh --vscode
```
