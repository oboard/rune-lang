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
values.forEach((value, index, array) => @io.println(value))
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

## map and set

Maps and sets are created through module functions and then used through
receiver methods:

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
