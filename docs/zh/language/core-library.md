# 核心库

Rune 的内置能力由声明驱动。类型检查器和后端会从
`core/<module>/<module>.rn` 加载 stub；只有存在声明的模块调用才是合法的。

```rune
@io.println("hello")
@json.stringify({ name: "Rune" })
```

## assert

```rune
@assert.eq(actual, expected)
```

`eq[T](actual: T, expected: T)` 用于 Rune 测试。

## io

```rune
@io.print(value)
@io.println(value)
@io.printf(format, value)
@io.scan()
@io.scanLine()
@io.readAll()
```

Go 后端会把这些 helper 降低为 `fmt` 调用。TypeScript 后端会降低为 console
输出。

`scan` 读取下一个以空白分隔的 stdin token，`scanLine` 读取下一行，
`readAll` 读取剩余 stdin 文本。`scan` 和 `scanLine` 在 EOF 时返回 `null`。

`@io.Data` 是异步文件 API 返回的字节数据类型。Go 后端映射为 `[]byte`，
TypeScript 后端映射为 `Uint8Array`。

## fs

文件 API 声明为 routine，并使用 `Result` 代替回调：

```rune
~ read() => {
  file := @fs.readFile("1.txt")?
  file
}
```

`@fs.readFile(path: String)` 的形状与 Node.js `readFile` 对齐，但回调被转换
成异步结果：

```rune
@fs.readFile(path) -> Result[@io.Data, Error]
```

在 routine 内调用会自动等待。在普通函数中调用会启动一个 task。

`fs` 模块还声明了 `readFileText`、`writeFile`、`writeFileText`、`exists`、
`readdir`、`mkdir`、`remove` 和 `stat`。`stat` 返回 `FileStat`，包含
`size`、`isFile` 和 `isDirectory` 字段。

## compress

压缩 helper 都是 routine，并返回 `Result`：

```rune
~ roundtrip() => {
  data := @compress.gzipText("hello")?
  @compress.gunzipText(data)?

  packed := @compress.zstdText("hello")?
  @compress.unzstdText(packed)?
}
```

模块声明了 `gzip`、`gunzip`、`deflate`、`inflate`、`brotli`、
`unbrotli`、`zstd`、`unzstd`，以及 `gzipText`、`gunzipText`、
`brotliText`、`unbrotliText`、`zstdText` 和 `unzstdText`。

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

`@process.exit(code)` 返回 `Never`。

## cli

`cli` 模块提供命令、选项、参数、解析结果和帮助文本 helper，用于构建命令行
程序：

```rune
cmd ~= @cli.command("ship", "Ship a build artifact")
cmd = @cli.withVersion(cmd, "1.0.0")
cmd = @cli.withOption(cmd, @cli.flag("verbose", "v", "enable verbose output"))
cmd = @cli.withOption(cmd, @cli.option("output", "o", "FILE", "write output", false, "dist/app"))
cmd = @cli.withArgument(cmd, @cli.argument("target", "target name", true))

result := @cli.parse(cmd)
@io.println(result.values.getOr("output", ""))
@io.println(result.flags.getOr("verbose", false))
@io.println(result.positionals.getOr("target", ""))
```

测试中可以用 `@cli.parseArgs(cmd, args)` 解析显式参数数组。`@cli.help(cmd)`
返回生成的 usage 文本。解析错误放在 `result.error`；`-h` 和 `--help` 会设置
`result.help`。

## iter

`Iter[T]` 是普通 core 类型，包含 `next: () -> (T, Bool)` 字段。每次
`next()` 调用都会返回一个 tuple，其中 `[0]` 是值，`[1]` 是 ok 标志。

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

`@iter.rangeStep(start, end, step)`、`@iter.repeat(value, count)` 和
`@iter.empty(valueType)` 都返回 `Iter[T]`。迭代器提供 `toArray`、`each`
和 `map` receiver 方法。

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

`@stringbuffer.from(value)` 会创建一个带初始字符串的 buffer。

## net

TCP helper 基于 routine：

```rune
~ open() => {
  conn := @net.connect("127.0.0.1:8080")?
  data := @fs.readFile("payload.bin")?
  conn.write(data)?
  conn.close()?
}
```

`@net.listen(address)` 返回 `TCPListener`；listener 提供 `address`、`accept`
和 `close`。connection 提供 `read`、`write` 和 `close`。

## array

数组方法在数组值上调用：

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

`values[0]` 通过 `core/array` 中的 `_[_]` alias 声明。

## string

`String` 提供 receiver 方法：

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

当前运行时测试中，字符串索引和切片按用户可见的 Unicode 字符工作。`at` 和
`charAt` 返回 `Char`；需要 `String` 时可以调用 `.toString()`。

## char

```rune
'r'.toString()
```

## bool

```rune
true.not()
true.xor(false)
true.toString()
```

布尔运算符 `!`、`&&` 和 `||` 是语言运算符；这些 receiver 方法是便利 API。

## int、double、bigint

数值转换 helper：

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

算术运算本身由语言运算符处理。

## 定宽数值类型

Rune 也提供面向二进制处理的数值类型转换 helper，供 `Bytes`、`Reader` 和
`Writer` 使用：

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

有符号转换会按目标宽度回绕，无符号转换会按目标宽度截断。`Float` 表示 32
位浮点数。

## bytes、buffer、reader 与 writer

`Bytes` 是固定长度的字节视图。多字节读写需要显式传入 `littleEndian`
参数，单字节读写不需要：

```rune
bytes := @bytes.new(16)
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

`@bytes.fromInts(values)` 可以从字节数组创建 `Bytes`。`Bytes` 支持
`length`、`byteLength`、`clone`、`slice`、`toInts`，以及面向 `Int4`、
`Int8`、`UInt8`、`Int16`、`UInt16`、`Int`/`UInt`、`Int64`/`UInt64`、
`Float` 和 `Double` 的 `get*`、`set*` 方法。`Uint*`、`Int32`、
`Float32`、`Float64`、`BigInt64`/`BigUInt64` 这些别名也已声明。

需要可增长字节序列时使用 `Buffer`：

```rune
buffer := @buffer.new()
buffer.append(@uint8.fromInt(1))
buffer.appendInt(2)
buffer.appendBytes(@bytes.fromInts([3, 4]))

copy := buffer.clone()
data := copy.toBytes()
ints := copy.toInts()
```

`@buffer.fromBytes(data)` 会创建一个可变 buffer 副本。`Buffer` 支持
`length`、`byteLength`、`isEmpty`、`clear`、`clone`、`toBytes`、`toInts`、
`append`、`appendInt`、`appendBytes`、`reader` 和 `writer`。

`Reader` 用于从 `Bytes` 顺序读取：

```rune
reader := @reader.new(data)
first := reader.readUInt8()
next := reader.readInt16(true)
chunk := reader.readBytes(4)
reader.seek(0)
reader.skip(1)
```

`Reader` 支持 `length`、`byteLength`、`position`、`remaining`、`isEmpty`、
`seek`、`skip`、`read`/`readBytes`，以及和 `Bytes` 读取侧对应的 `read*`
方法。

`Writer` 用于顺序写入字节：

```rune
writer := @writer.new()
writer.writeUInt8(@uint8.fromInt(255))
writer.writeInt16(@int16.fromInt(0 - 1234), true)
writer.writeFloat(@float.fromDouble(1.5), true)
out := writer.toBytes()
```

`@writer.withCapacity(capacity)` 可以预分配空间。`Writer` 支持 `length`、
`byteLength`、`position`、`clear`、`toBytes`、`toInts`、
`write`/`writeBytes`，以及和 `Bytes` 写入侧对应的 `write*` 方法。

## map 与 set

Map 和 Set 通过模块函数创建，再通过 receiver 方法使用：

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

核心声明中也包含 `WeakMap[K, V]` 和 `WeakSet[T]` 的 receiver 方法面。

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

`@json.stringify` 会序列化 object-like 值和数组。函数值对象字段会被忽略。

## symbol

```rune
value := @symbol.create("name")
unique := @symbol.unique("id")
shared := @symbol.for("global")
@symbol.keyFor(shared)
@symbol.description(value)
@symbol.toString(value)
```

`keyFor` 和 `description` 返回可空字符串。

## go

```rune
@go.import("fmt")

main() => {
  name := "Rune"
  @go.stmt("fmt.Println($name)")
}

isAdult(age: Int) -> Bool => @go.expr("$age >= 18")
```

`@go.import` 只能出现在顶层。`@go.stmt` 和 `@go.expr` 是 Go 后端 FFI escape
hatch，TypeScript 后端不支持。
