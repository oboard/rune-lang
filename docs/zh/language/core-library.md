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
```

Go 后端会把这些 helper 降低为 `fmt` 调用。TypeScript 后端会降低为 console
输出。

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
values.forEach((value, index, array) => @io.println(value))
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

当前运行时测试中，字符串索引和切片按用户可见的 Unicode 字符工作。

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

Rune 也提供面向二进制处理的数值类型转换 helper，供 `Binary`、`Reader` 和
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

## binary、buffer、reader 与 writer

`Binary` 是固定长度的字节视图。多字节读写需要显式传入 `littleEndian`
参数，单字节读写不需要：

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

`@binary.fromInts(values)` 可以从字节数组创建 `Binary`。`Binary` 支持
`length`、`byteLength`、`clone`、`slice`、`toInts`，以及面向 `Int4`、
`Int8`、`UInt8`、`Int16`、`UInt16`、`Int`/`UInt`、`Int64`/`UInt64`、
`Float` 和 `Double` 的 `get*`、`set*` 方法。`Uint*`、`Int32`、
`Float32`、`Float64`、`BigInt64`/`BigUInt64` 这些别名也已声明。

需要可增长字节序列时使用 `Buffer`：

```rune
buffer := @buffer.new()
buffer.append(@uint8.fromInt(1))
buffer.appendInt(2)
buffer.appendBinary(@binary.fromInts([3, 4]))

copy := buffer.clone()
binary := copy.toBinary()
ints := copy.toInts()
```

`@buffer.fromBinary(binary)` 会创建一个可变 buffer 副本。`Buffer` 支持
`length`、`byteLength`、`isEmpty`、`clear`、`clone`、`toBinary`、`toInts`、
`append`、`appendInt`、`appendBinary`、`reader` 和 `writer`。

`Reader` 用于从 `Binary` 顺序读取：

```rune
reader := @reader.new(binary)
first := reader.readUInt8()
next := reader.readInt16(true)
chunk := reader.readBinary(4)
reader.seek(0)
reader.skip(1)
```

`Reader` 支持 `length`、`byteLength`、`position`、`remaining`、`isEmpty`、
`seek`、`skip`、`read`/`readBinary`，以及和 `Binary` 读取侧对应的 `read*`
方法。

`Writer` 用于顺序写入字节：

```rune
writer := @writer.new()
writer.writeUInt8(@uint8.fromInt(255))
writer.writeInt16(@int16.fromInt(0 - 1234), true)
writer.writeFloat(@float.fromDouble(1.5), true)
out := writer.toBinary()
```

`@writer.withCapacity(capacity)` 可以预分配空间。`Writer` 支持 `length`、
`byteLength`、`position`、`clear`、`toBinary`、`toInts`、
`write`/`writeBinary`，以及和 `Binary` 写入侧对应的 `write*` 方法。

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
scores.forEach((value, key, map) => @io.println(value))
scores.clear()

seen := @set.new("")
seen.size()
seen.has("rune")
seen.add("rune")
seen.delete("rune")
seen.values()
seen.forEach((value) => @io.println(value))
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
