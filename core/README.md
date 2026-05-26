# Rune Core Library

`/core` is the declaration source for built-in Rune modules. The compiler does
not invent `@module.function` calls: checker and codegen load declarations from
these module files.

Current modules:

* `array` - declared placeholder for array APIs
* `binary` - fixed byte views and numeric reads/writes
* `buffer` - mutable byte buffers
* `compress` - async gzip/zlib/brotli/zstd helpers
* `fs` - async filesystem helpers returning `Result`
* `map` - map construction and receiver APIs
* `net` - async TCP connection/listener declarations
* `path` - path manipulation helpers
* `process` - argv/cwd/env/platform helpers
* `reader` - sequential binary readers
* `set` - set construction and receiver APIs
* `stringbuffer` - mutable string builder
* `iter` - tuple-returning iterator helpers
* `writer` - sequential binary writers
* `io` - I/O declarations backed by Go `fmt`
* `json` - JSON serialization helpers
* `go` - inline Go FFI declarations

Each module directory contains a `<module>.rn` stub file. A stub is still Rune
syntax: function signatures declare the callable surface, and the string body
names the backend intrinsic or binding.

Example:

```rune
println[T](value: T) => "%go:fmt.Println"
```

Backends use those declarations to type-check calls and lower them to
target-specific implementations.
