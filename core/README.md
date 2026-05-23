# Rune Core Library

`/core` is the declaration source for built-in Rune modules. The compiler does
not invent `@module.function` calls: checker and codegen load declarations from
these module files.

Current modules:

* `array` - declared placeholder for array APIs
* `map` - declared placeholder for map APIs
* `io` - I/O declarations backed by Go `fmt`
* `json` - JSON serialization helpers
* `go` - inline Go FFI declarations

Each module directory contains a `<module>.rn` stub file. A stub is still Rune
syntax: function signatures declare the callable surface, and the string body
names the backend intrinsic or binding.

Example:

```rune
println(value: Any) => "%go:fmt.Println"
```

Backends use those declarations to type-check calls and lower them to
target-specific implementations.
