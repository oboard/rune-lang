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

## map 与 set

Map 和 Set 通过模块函数创建，再通过 receiver 方法使用：

```rune
scores := @map.newMap("", 0)
scores.size()
scores.has("rune")
scores.getOr("rune", 0)
scores.set("rune", 10)
scores.delete("rune")
scores.keys()
scores.values()
scores.forEach((value, key, map) => @io.println(value))
scores.clear()

seen := @map.newSet("")
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
