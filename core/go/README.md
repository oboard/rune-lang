# core/go

Declared inline Go FFI module.

```rune
@go.import("fmt")
@go.stmt("fmt.Println($name)")
@go.expr("$age >= 18")
```

`@go.import` is file-level only. `@go.stmt` and `@go.expr` may appear inside
function bodies. `@go.expr` returns a dynamic value, so the containing Rune
function should use `-> Type` when that expression defines the return value.
`$name` is rewritten to the Go backend's mangled Rune identifier.
