# 基础语法

本页按语言基础参考文档的方式组织，内容以当前 Rune parser、checker、
examples 和 tests 已支持的语法为准。

## 源文件结构

一个 Rune 源文件由顶层声明组成：

```rune
@go.import("fmt")

User: {
  name: String
  age: Int
}

isAdult(user: User) -> Bool => user.age >= 18

? "adult check" {
  @assert.eq(isAdult(User { name: "Ada", age: 36 }), true)
}
```

当前支持的顶层形式包括：

- `@go.import("pkg")`：声明 Go 后端需要的 import。
- `Name[T]: { ... }`：类型声明。
- `name[T](params) -> Return => body`：函数声明。
- `? "name" { ... }`：测试声明。

目前还没有通用的用户级 import 语法。内置模块通过 `@module.function(...)`
调用，并从 `core/` 目录加载声明。

## 词法规则

标识符以 Unicode 字母或 `_` 开头，后续可以包含字母、数字或 `_`。

```rune
value := 1
_private := "ok"
名字 := "Rune"
```

支持行注释和块注释：

```rune
// 单行注释
/*
  多行注释
*/
```

换行用于分隔声明和语句。在参数列表、数组、对象字面量、结构体字面量和
match 分支等位置，逗号也可以作为分隔符。Rune 风格通常在对象字段中使用换
行，在紧凑列表中使用逗号。

## 字面量

Rune 支持整数、双精度浮点、大整数、字符串、布尔值和 null：

```rune
intValue := 42
doubleValue := 6.25e-1
bigValue := 9007199254740993n
text := "hello\nRune"
ch := 'R'
ok := true
missing := null
```

整数字面量类型是 `Int`。带小数点或指数的字面量类型是 `Double`。以 `n`
结尾的字面量类型是 `BigInt`。单引号字符字面量类型是 `Char`。`true` 和
`false` 类型是 `Bool`。`null` 的类型是 `Null`，可以流入 `Int?`、
`String?` 这样的可空类型。

## 内置类型

面向用户的内置标量和特殊类型包括：

```text
Int
Double
BigInt
String
Char
Bool
Void
Object
Symbol
HTMLElement
Dynamic
Data
Error
Result[T, E]
Task[T]
```

`Dynamic` 可以出现在确实需要动态检查的声明里。能写出具体类型或泛型类型参数
时，优先使用具体类型或泛型。

可空类型使用后缀 `?`：

```rune
maybeName(value: String?) -> String? => value

main() => {
  @io.println(maybeName(null))
  @io.println(maybeName("Rune"))
}
```

泛型类型应用使用方括号：

```rune
values: Array[Int]
scores: Map[String, Int]
seen: Set[String]
```

函数类型使用 `(params) -> Return`。函数类型里的参数名主要用于文档展示和
回调检查：

```rune
each[R](callback: (value: Int, index?: Int, array?: Array[Int]) -> R) -> Void
```

## 运算符

一元运算符：

```rune
-value
!flag
~value
```

后缀自增：

```rune
count++
```

二元运算符按优先级从低到高为：

```text
||
&&
|
^
&
== !=
< <= > >=
<< >> >>>
+ -
* / %
```

数值运算要求两侧是相同的数值类型。`%` 支持 `Int` 和 `BigInt`，不支持
`Double`。`+` 也用于字符串拼接，但只要任意一侧是字符串，两侧都必须是
`String`。

```rune
@assert.eq(1 + 2 * 3, 7)
@assert.eq("ru" + "ne", "rune")
@assert.eq(22 / 7, 3)
@assert.eq(22n % 7n, 1n)
```

`&&` 和 `||` 会短路求值，并要求操作数是布尔值。有序比较支持类型相同的
`Int`、`Double`、`BigInt`、`String` 或 `Char`。

位运算符 `~`、`&`、`|`、`^`、`<<`、`>>` 和 `>>>` 要求操作数是匹配的整数
类型。`>>>` 要求左操作数是无符号整数。

## 绑定与代码块

代码块是表达式。代码块返回最后一个表达式；如果最后一条语句没有值，则返回
`Void`：

```rune
sum(a: Int, b: Int) -> Int => {
  result := a + b
  result
}
```

块内可以声明绑定：

```rune
value := 1
mutable ~= 1
signal $= 0
```

`:=` 创建普通局部绑定。`~=` 表示可变意图。`$=` 创建用于响应式代码的 signal
绑定。赋值使用 `=`：

```rune
count ~= 0
count = count + 1
```

赋值也可以作为表达式出现，常见于事件处理器：

```rune
<button @click={count = count + 1}>Add</button>
```

## 函数

函数把参数映射到一个函数体表达式：

```rune
add(a: Int, b: Int) -> Int => a + b
```

声明默认是 private。在函数、类型、字段、方法或枚举成员前加 `+`，可以让其他
Rune 文件访问它。

如果类型推断有足够信息，参数和返回类型可以省略：

```rune
fib(n) => n {
  0 => 0
  1 => 1
  _ => fib(n - 1) + fib(n - 2)
}
```

泛型函数在函数名后声明类型参数：

```rune
identity[T](value: T) -> T => value
```

`main` 是特殊函数：它总是按返回 `Void` 进行检查。

## Routine 与 Result

在函数前加 `~` 可以把它标记为 routine：

```rune
~ test(count: Int) => {
  @io.println("Hello World" + count.toString())
}

main() => {
  test(1)
  test(2)
  test(3)
  @io.println("Hello World")
}
```

在普通函数中调用 routine 会启动它，并返回 `Task[T]`。生成后的程序会在退出前
等待未完成的 task。在 routine 中调用另一个 routine 会自动等待，源码里不需要
写 `await`。

错误处理使用内置的 `Result[T, E]` enum 形状：

```rune
Result[T, E]: {
  Ok(value: T)
  Err(error: E)
}

Error: {
  code: Int
  message: String
  cause: Error?
}
```

在 routine 内部，后缀 `?` 会解包 `Result[T, E]`。遇到 `Err` 时会提前返回，
并把错误提升到 routine 的返回类型中：

```rune
~ read() => {
  file := @fs.readFile("1.txt")?
  file
}
```

推导出的类型是 `~ read() -> Result[@io.Data, Error]`。

## Lambda

Lambda 参数必须写在括号中：

```rune
values.map((value) => value * 2)
values.map((value: Int, index: Int) => value + index)
```

下面的写法是无效的：

```rune
values.map(value => value * 2)
```

Lambda 函数体可以是表达式，也可以是代码块：

```rune
handle := (value: Int) => {
  next := value + 1
  next
}
```

如果回调类型允许更多参数，而实际 lambda 只声明了更少参数，Rune 会接受较短
的 lambda。数组回调因此可以只写 `value`，即使声明类型还包含 `index` 和
`array`。

## 模式函数体与 Match

函数体可以是一个模式块。这种形式主要用于单参数函数：

```rune
fib(n: Int) -> Int => {
  0 => 0
  1 => 1
  _ => fib(n - 1) + fib(n - 2)
}
```

任意表达式都可以用 `subject { pattern => expression }` 进行匹配：

```rune
label(value: Int) -> String => value {
  <0 => "negative"
  0 => "zero"
  _ => "positive"
}
```

支持的模式形式包括：

```rune
_          // 通配
1          // Int 字面量
1.5        // Double 字面量
1n         // BigInt 字面量
"text"     // String 字面量
'c'        // Char 字面量
true       // Bool 字面量
false
null
< 10       // 比较模式
<= 10
> 10
>= 10
(1, _)     // 元组模式语法，预留给 tuple-like subject
Ok(value)  // Result 构造器模式
Err(error)
```

所有非 `Void` 分支应该返回兼容类型。嵌套 match 只是普通表达式：

```rune
x {
  1 => y {
    2 => "both"
    _ => "only x"
  }
  _ => "none"
}
```

`Result` 值也可以手动 match：

```rune
readUser("user.json") {
  Ok(user) => @io.println(user.name)
  Err(e) => @io.println(e.message)
}
```

## 三元表达式

Rune 支持 C 风格条件表达式：

```rune
choose(flag: Bool) -> Int => flag ? 1 : 2
```

条件必须是 `Bool`。真分支和假分支必须能统一为同一个结果类型。三元表达式会
短路求值，只执行被选中的分支。

## 数组

数组字面量会推断元素类型：

```rune
empty := []
values := [1, 2, 3]
matrix := [[1, 2], [3, 4]]
nullable := [1, null, 3]
```

展开语法会把另一个数组插入新字面量：

```rune
middle := [2, 3]
values := [1, ...middle, 4]
```

索引使用 `Int`：

```rune
first := values[0]
```

数组方法声明在 `core/array` 中：

```rune
values.push(4)
values.set(1, 20)
values.length()
values.isEmpty()
values.first()
values.last()
values.slice(1, 3)
values.clone()
values.reverse()
values.contains(20)
values.each((value, index) => @io.println(value))
doubled := values.map((value) => value * 2)
```

## 结构体类型

命名的记录式类型用 `Name: { ... }` 声明：

```rune
User: {
  id: Int
  name: String
  age: Int

  isAdult() -> Bool => .age >= 18
  label(prefix: String) -> String => prefix + .name
}
```

结构体字面量必须提供所有已声明字段：

```rune
user := User {
  id: 1
  name: "Ada"
  age: 36
}
```

字段和方法通过 selector 访问：

```rune
@io.println(user.name)
@io.println(user.isAdult())
```

方法内部的 `.field` 是 `this.field` 的简写。显式的 `this` 绑定也可以在方法
中使用。

可以声明泛型类型：

```rune
Box[T]: {
  value: T
}
```

当前最完整的泛型行为集中在内置泛型容器中，例如 `Array[T]`、`Map[K, V]` 和
`Set[T]`。

## 枚举类型

枚举沿用顶层 `Name: { ... }` 的声明形式，成员值必须是整数：

```rune
Status: {
  Completed = 0
  Fail = 1
}
```

枚举成员通过 selector 读取。成员本身带有枚举类型，因此 `Status.Completed`
的类型是 `Status`：

```rune
status := Status.Completed
```

枚举值可以和同一枚举类型的值比较，也可以用于 match pattern：

```rune
statusText(status: Status) -> String => status {
  Status.Completed => "completed"
  Status.Fail => "fail"
  _ => "unknown"
}
```

## 匿名对象

匿名对象是表达式：

```rune
account := {
  name: "core"
  age: 4

  nextAge() -> Int => .age + 1
  title(prefix: String) -> String => prefix + .name
}
```

匿名对象会推断为封闭对象类型。两个匿名对象类型只有在字段作为封闭类型匹配
时才会统一。命名结构体参数仍然可以在调用点接收满足该结构的匿名对象。

函数值字段可以直接调用：

```rune
@io.println(account.title("module:"))
```

`@json.stringify` 生成 JSON 时会忽略函数值字段。

## 模块与调用

模块函数使用 `@module.name(...)`：

```rune
@io.println("hello")
json := @json.stringify({ name: "Rune" })
scores := @map.new("", 0)
```

编译器不会凭空创造模块函数。只有 `core/<module>/<module>.rn` 中存在匹配声明
时，模块调用才是合法的。

Rune 源文件可以用 `@"path"` 导入同项目里的其他文件：

```rune
@"./helper.rn"

main() => @io.println(helper())
```

被导入文件里的声明默认是 private。如果其他文件需要调用 `helper.rn` 里的
API，需要给对应声明加 `+`：

```rune
+ helper() -> Int => 42
```

相对路径基于当前导入文件解析。导入路径必须显式包含文件后缀。

## 响应式与 Watch

Signal 绑定使用 `$=`：

```rune
render() => {
  count $= 0
  double := count * 2

  count -> (old, new) => {
    @io.println(old)
    @io.println(new)
  }

  count = count + 1
}
```

`target -> handler` 注册 watcher。handler 可以是零参数代码块，也可以是接收
`(old, new)` 两个参数的 lambda。

响应式数组和对象字面量使用 `$[...]` 和 `${...}`：

```rune
items := $["Item 1", "Item 2"]
state := ${ count: 0 }
```

TypeScript 后端会为 signal 和响应式数组生成辅助封装。

## XML 与 DOM 表达式

类 XML 元素是表达式，返回 `HTMLElement`：

```rune
render() -> HTMLElement => {
  count $= 0

  <div class="counter">
    <p>Count: {count}</p>
    <button @click={count++}>Click Me</button>
  </div>
}
```

支持的 XML 语法包括：

- 普通元素：`<div>...</div>`。
- 自闭合元素：`<input />`。
- 字符串属性：`class="counter"`。
- 裸属性：`disabled`。
- 表达式属性：`value={count}`。
- 带 `@` 的事件属性：`@click={handler}`。
- 文本子节点。
- 嵌入表达式：`{expr}`。
- 嵌套元素。

XML 面向 TypeScript 后端。Go 后端目前会为 XML 表达式生成占位注释。

## 内联 Go FFI

Go 后端支持通过 `@go` 内联 Go：

```rune
@go.import("fmt")

isAdult(age: Int) -> Bool => @go.expr("$age >= 18")

main() => {
  name := "Ada"
  @go.stmt("fmt.Println($name)")
  @io.println(isAdult(36))
}
```

`@go.import` 只能出现在顶层。`@go.stmt` 和 `@go.expr` 可以出现在函数体中。
Go 字符串里的 `$name` 会被重写为 Rune 绑定对应的生成后 Go 标识符。

## 测试

测试声明以 `?` 开始：

```rune
? "string split" {
  parts := "r,u,n,e".split(",")
  @assert.eq(parts[1], "u")
}
```

测试体是普通 Rune 代码块。标准断言 helper 是 `@assert.eq(actual, expected)`。

## 当前语法边界

Rune 目前刻意保持较小的语法面。当前 parser 中没有 `if`、`for`、`while`、
`trait`、`impl`、`class`、包导入、异常或宏语法。对应能力请使用已支持的
match 表达式、三元表达式、递归、数组回调和核心模块。
