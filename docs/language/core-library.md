# Core Library

Rune built-ins are declaration-driven. The checker and backends load stubs from
`core/<module>/<module>.rn`; a module call is valid only if the declaration is
present there.

```rune
@io.println("hello")
@json.stringify({ name: "Rune" })
```

## assert

```rune
@assert.eq(actual, expected)
```

`eq[T](actual: T, expected: T)` is used by Rune tests.

## io

```rune
@io.print(value)
@io.println(value)
@io.printf(format, value)
```

The Go backend lowers these helpers to `fmt` calls. The TypeScript backend
lowers printing to console output.

`@io.Data` is the byte-data type returned by async file APIs. It maps to
`[]byte` on the Go backend and `Uint8Array` on the TypeScript backend.

## fs

File APIs are declared as routines and use `Result` instead of callbacks:

```rune
~ read() => {
  file := @fs.readFile("1.txt")?
  file
}
```

`@fs.readFile(path: String)` has the same shape as Node.js `readFile` with the
callback converted to an async result:

```rune
@fs.readFile(path) -> Result[@io.Data, Error]
```

Inside a routine, the call is awaited automatically. Outside a routine, the
call starts a task.

The `fs` module also declares `readFileText`, `writeFile`, `writeFileText`,
`exists`, `readdir`, `mkdir`, `remove`, and `stat`. `stat` returns `FileStat`
with `size`, `isFile`, and `isDirectory` fields.

## compress

Compression helpers are routines and return `Result`:

```rune
~ roundtrip() => {
  data := @compress.gzipText("hello")?
  @compress.gunzipText(data)?

  packed := @compress.zstdText("hello")?
  @compress.unzstdText(packed)?
}
```

The module declares `gzip`, `gunzip`, `deflate`, `inflate`, `brotli`,
`unbrotli`, `zstd`, `unzstd`, plus `gzipText`, `gunzipText`, `brotliText`,
`unbrotliText`, `zstdText`, and `unzstdText`.

## path

```rune
@path.basename("src/main.rn")
@path.dirname("src/main.rn")
@path.extname("src/main.rn")
@path.join(["src", "main.rn"])
@path.normalize("src/../main.rn")
@path.resolve(["."])
@path.relative("src", "src/main.rn")
@path.isAbsolute("src/main.rn")
```

## process

```rune
@process.argv()
@process.cwd()
@process.env("HOME")
@process.platform()
```

`@process.exit(code)` returns `Never`.

## iter

`Iter[T]` is a normal core type with a `next: () -> (T, Bool)` field. Each
`next()` call returns a tuple where `[0]` is the value and `[1]` is the ok flag.

```rune
iter := @iter.range(0, 3)
item := iter.next()
@io.println(item[0])
@io.println(item[1])

letters := @iter.fromArray(["a", "b"])
wrapped := @iter.new(letters.next)
@io.println(wrapped.next()[0])

values := @iter.range(1, 4).toArray()
@iter.range(1, 4).each((value) => @io.println(value))
mapped := @iter.range(1, 4).map((value) => value * 2)
```

`@iter.rangeStep(start, end, step)`, `@iter.repeat(value, count)`, and
`@iter.empty(valueType)` return `Iter[T]`. Iterators expose `toArray`, `each`,
and `map` receiver methods.

## stringbuffer

```rune
buffer := @stringbuffer.new()
buffer.append("Rune")
buffer.appendLine(" core")
buffer.length()
buffer.isEmpty()
buffer.toString()
buffer.clear()
```

`@stringbuffer.from(value)` creates a buffer initialized with a string.

## net

TCP helpers are routine-based:

```rune
~ open() => {
  conn := @net.connect("127.0.0.1:8080")?
  data := @fs.readFile("payload.bin")?
  conn.write(data)?
  conn.close()?
}
```

`@net.listen(address)` returns a `TCPListener`; listeners expose `address`,
`accept`, and `close`. Connections expose `read`, `write`, and `close`.

## array

Array methods are called on array values:

```rune
values := [1, 2, 3]
values.length()
values.isEmpty()
values.push(4)
values.set(1, 20)
values.pop()
values.first()
values.last()
values.slice(1, 3)
values.clone()
values.reverse()
values.contains(20)
values.each((value, index, array) => @io.println(value))
values.each((value) => @io.println(value))
mapped := values.map((value) => value * 2)
value := values[0]
```

`values[0]` is declared through the `_[_]` alias in `core/array`.

## string

`String` has receiver methods:

```rune
"rune".length()
"rune".isEmpty()
"rune".toString()
"rune".at(0)
"rune".charAt(0)
"rune".slice(1, 3)
"ru".concat("ne")
"rune".includes("un")
"rune".startsWith("ru")
"rune".endsWith("ne")
"banana".indexOf("na")
"banana".lastIndexOf("na")
"Rune".toLowerCase()
"Rune".toUpperCase()
"  rune  ".trim()
"  rune  ".trimStart()
"  rune  ".trimEnd()
"ha".repeat(3)
"one one".replace("one", "1")
"one one".replaceAll("one", "1")
"r,u,n,e".split(",")
```

String indexing and slicing operate on user-visible Unicode characters in the
current runtime tests.

## bool

```rune
true.not()
true.xor(false)
true.toString()
```

Boolean operators `!`, `&&`, and `||` are language operators; these receiver
methods are convenience APIs.

## int, double, bigint

Numeric conversion helpers:

```rune
@int.toDouble(1)
@int.toBigInt(1)
@double.trunc(1.5)
@double.floor(1.5)
@double.ceil(1.5)
@double.round(1.5)
@bigint.fromInt(1)
@bigint.toDouble(1n)
@bigint.toString(1n)
```

Arithmetic itself is handled by language operators.

## fixed-width numeric types

Rune also exposes conversion helpers for the binary-oriented numeric types used
by `Binary`, `Reader`, and `Writer`:

```rune
@int4.fromInt(15)
@int4.toInt(@int4.fromInt(15))
@int8.fromInt(130)
@int16.fromInt(65535)
@int64.fromInt(123456)
@uint.fromInt(123456)
@uint8.fromInt(255)
@uint16.fromInt(65535)
@uint64.fromInt(123456)
@float.fromDouble(1.5)
@float.toDouble(@float.fromDouble(1.5))
```

Signed conversions wrap to their target width. Unsigned conversions mask to the
target width. `Float` is a 32-bit floating-point value.

## binary, buffer, reader, and writer

`Binary` is a fixed byte view. Multi-byte reads and writes take an explicit
`littleEndian` flag, while byte reads do not:

```rune
bytes := @binary.new(16)
bytes.setInt4(0, @int4.fromInt(0 - 1))
bytes.setUInt8(1, @uint8.fromInt(255))
bytes.setInt16(2, @int16.fromInt(0 - 1234), true)
bytes.setUInt(4, @uint.fromInt(123456), false)
bytes.setFloat(8, @float.fromDouble(1.5), true)

@assert.eq(@int4.toInt(bytes.getInt4(0)), 0 - 1)
@assert.eq(@uint8.toInt(bytes.getUint8(1)), 255)
@assert.eq(@int16.toInt(bytes.getInt16(2, true)), 0 - 1234)
@assert.eq(@uint.toInt(bytes.getUInt(4, false)), 123456)
@assert.eq(@float.toDouble(bytes.getFloat32(8, true)), 1.5)
```

`@binary.fromInts(values)` builds a `Binary` from byte values. `Binary`
supports `length`, `byteLength`, `clone`, `slice`, `toInts`, `get*`, and
`set*` methods for `Int4`, `Int8`, `UInt8`, `Int16`, `UInt16`, `Int`/`UInt`,
`Int64`/`UInt64`, `Float`, and `Double`. `Uint*`, `Int32`, `Float32`,
`Float64`, and `BigInt64`/`BigUInt64` aliases are also declared.

Use `Buffer` when bytes need to grow:

```rune
buffer := @buffer.new()
buffer.append(@uint8.fromInt(1))
buffer.appendInt(2)
buffer.appendBinary(@binary.fromInts([3, 4]))

copy := buffer.clone()
binary := copy.toBinary()
ints := copy.toInts()
```

`@buffer.fromBinary(binary)` creates a mutable buffer copy. `Buffer` supports
`length`, `byteLength`, `isEmpty`, `clear`, `clone`, `toBinary`, `toInts`,
`append`, `appendInt`, `appendBinary`, `reader`, and `writer`.

`Reader` consumes bytes sequentially from a `Binary`:

```rune
reader := @reader.new(binary)
first := reader.readUInt8()
next := reader.readInt16(true)
chunk := reader.readBinary(4)
reader.seek(0)
reader.skip(1)
```

`Reader` supports `length`, `byteLength`, `position`, `remaining`, `isEmpty`,
`seek`, `skip`, `read`/`readBinary`, and `read*` methods matching the `Binary`
numeric read surface.

`Writer` builds bytes sequentially:

```rune
writer := @writer.new()
writer.writeUInt8(@uint8.fromInt(255))
writer.writeInt16(@int16.fromInt(0 - 1234), true)
writer.writeFloat(@float.fromDouble(1.5), true)
out := writer.toBinary()
```

`@writer.withCapacity(capacity)` preallocates space. `Writer` supports
`length`, `byteLength`, `position`, `clear`, `toBinary`, `toInts`,
`write`/`writeBinary`, and `write*` methods matching the `Binary` numeric write
surface.

## map and set

Maps and sets are created through module functions and then used through
receiver methods:

```rune
scores := @map.new("", 0)
scores.size()
scores.has("rune")
scores.getOr("rune", 0)
scores.set("rune", 10)
scores.delete("rune")
scores.keys()
scores.values()
scores.each((value, key, map) => @io.println(value))
scores.clear()

seen := @set.new("")
seen.size()
seen.has("rune")
seen.add("rune")
seen.delete("rune")
seen.values()
seen.each((value) => @io.println(value))
seen.clear()
```

The core declarations also include `WeakMap[K, V]` and `WeakSet[T]` receiver
surfaces.

## json

```rune
JsonUser: {
  name: String
  age: Int
}

main() => {
  user := JsonUser { name: "Ada", age: 36 }
  @io.println(@json.stringify(user))
  @io.println(@json.stringify({
    name: "Rune"
    user: user
    tags: ["compiler", "json"]
    greet() => @io.println(.name)
  }))
}
```

`@json.stringify` serializes object-like values and arrays. Function-valued
object fields are omitted.

## symbol

```rune
value := @symbol.create("name")
unique := @symbol.unique("id")
shared := @symbol.for("global")
@symbol.keyFor(shared)
@symbol.description(value)
@symbol.toString(value)
```

`keyFor` and `description` return nullable strings.

## go

```rune
@go.import("fmt")

main() => {
  name := "Rune"
  @go.stmt("fmt.Println($name)")
}

isAdult(age: Int) -> Bool => @go.expr("$age >= 18")
```

`@go.import` is top-level only. `@go.stmt` and `@go.expr` are Go-backend FFI
escape hatches and are not supported by the TypeScript backend.
