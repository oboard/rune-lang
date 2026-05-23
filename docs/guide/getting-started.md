# Getting Started

Rune is developed as a Go module. From the repository root, the fastest way to
exercise the toolchain is:

```sh
go run ./cmd/rune check examples/fib.rn
go run ./cmd/rune fmt examples/fib.rn
go run ./cmd/rune run examples/fib.rn
go run ./cmd/rune build -o /tmp/rune-fib examples/fib.rn
go run ./cmd/rune ts examples/counter.rn
go run ./cmd/rune repl
go run ./cmd/rune lsp
```

The development helper wraps the common setup:

```sh
scripts/dev.sh
scripts/dev.sh --shell
scripts/dev.sh --vscode
```

Build a local CLI binary for editor integration with:

```sh
go build -o .bin/rune ./cmd/rune
```

## Minimal Program

```rune
main() => {
  @io.println("Hello, Rune")
}
```

`main` is the process entrypoint for compiled or run programs. It must return
`Void`; if no return type is declared, Rune treats `main` as `Void`.

## Running Tests

Rune source files can contain test declarations:

```rune
? "arithmetic" {
  @assert.eq(1 + 2, 3)
}
```

The Go test suite drives the Rune tester today:

```sh
go test ./...
```

## Documentation Site

This documentation is a VitePress site stored entirely in `docs/`.

```sh
cd docs
npm install
npm run dev
npm run build
```
