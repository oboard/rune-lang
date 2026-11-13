# TypeScript 兼容档（Rune TS Profile）

Rune 的方向是「可静态判定的 TypeScript 子集 + Go/MoonBit 后端」（RFC 0001）。
本页维护当前接受的语法、跨端类型映射，以及**故意拒绝**的动态特性。

## 已接受的语法

| TypeScript | Rune | 状态 |
| --- | --- | --- |
| `function f(a: T): R` | `f(a: T) -> R => ...` | `ts2rune` 支持 |
| `const x = v` / `let x = v` | `x := v` / `x :=: v` | 默认不可变；`:=:` 才是可变绑定 |
| `if / else if / else` | `cond ? { ... } : { ... }` | 降为三元块表达式 |
| `switch` | 模式分支块 | TS switch 暂未转换；保留原有 `1 => ...` 写法 |
| `for (const x of xs)` | 同形 | 数组；支持 `^`/`^>` |
| `while` | `* (cond) { ... }` | 支持 |
| `return` / `break` / `continue` | `return` / `^` / `^>` | 支持 |
| `A \| B` 类型标注 | 同形 | 支持 |
| `typeof x === "string"` | 同形 | 支持，并收窄 union |
| 箭头函数 | 同形 | 回调场景支持 |
| 模板字符串 | 同形 | 支持 |
| 静态 `import` / `export` | `@"./m.rune"` | 支持 |
| class（字段 + 构造函数 + 方法 + static） | struct + `Class_new` | 支持子集 |
| class 继承（`extends`、`super(...)`、override） | 扁平化字段/方法 | 支持子集 |
| 带初始化的 static 字段 | 顶层 `Class_field` 常量 | 支持 |
| `get` 访问器 | 改写为 `get_<property>()` 方法 | 支持 |
| `set` 访问器、可变 static 字段 | — | 以 `RUNE-PROFILE-002` 拒绝 |
| `try/catch/finally`、`throw` | `{ } !? { e => } -- { }`、`!! value` | 符号形式；Go 端拒绝内部的控制流 |
| `for(init; cond; update)` | 降为 `* (cond)` | `ts2rune` 支持 |
| `do { } while (cond)` | 降为 `* (true)` + `^` | `ts2rune` 支持 |

### class 降糖

TypeScript class 会降成 Rune struct 加工厂函数：

```ts
class Point {
  x: number
  constructor(x: number) { this.x = x }
  sum(): number { return this.x + this.x }
  static origin(): Point { return new Point(0) }
}
```

输出 `Point: { x: Double; sum() -> Double => ... }`、`Point_new(x: Double) -> Point`
与 `Point_origin() -> Point`，并把 `new Point(...)` / `Point.origin()` 重写为对应调用。
构造函数体只允许 `this.field = value` 赋值（派生类可在开头调用一次
`super(...)`）；分支、校验等一律拒绝，不做有损近似。

单继承会扁平化：派生 struct 包含基类字段，未 override 的基类方法会复制到派生
类型；`Child_new` 调用一次 `Base_new` 后把基类字段拷入派生 struct。
`super.method()` 调用派生类型上的 `__super_<method>` 别名；static 方法以
`Child_<method>` 形式继承。带初始化的 static 字段降为顶层常量，只支持读取。
当接收者类型静态可知（`this`、有类型标注的参数、`new C()` 局部变量）时，
`get` 访问器改写为 `get_<property>()` 调用。setter、可变 static 字段、装饰器
报 `RUNE-PROFILE-002`；基类类型多态仍不支持。

### try / catch / finally

`throw` 写作 `!! value`；try 块是裸块，后接符号处理器：`!? { err => ... }`
是带绑定 catch，`?! { ... }` 是不带绑定 catch，`-- { ... }` 是 finally。

```rune
{
  !! "boom"
} !? {
  err => @io.println(err)
} -- {
  @io.println("done")
}
```

解释器与 TypeScript 后端是完整语句语义。Go 后端用闭包内的 `panic`/`recover`
实现，因此 **try/catch/finally 内部**的 `return` / `^` / `^>` 会给出明确
codegen 错误，而不是悄悄改变跳转目标。MoonBit 当前显式拒绝
（`try/catch/throw yet`）。

`rune test` 会检测这些构造并切换到宿主解释器，测试文件不依赖自举解释器的
特性覆盖。

### 经典 for

`ts2rune` 把 `for (init; cond; update)` 降为 init 语句 + `* (cond)`。
`^>` 会被改写为先执行 `update` 再跳转，从而与 TypeScript 的后置迭代语义一致：

```ts
for (let i = 0; i < n; i++) { ... continue ... }
```

会生成在 `^>` 处和循环体末尾各有一份 `i++` 的 `* (cond)` 循环。Rune 源本身
尚无原生经典 for 语法。

`do { body } while (cond)` 降为内部循环，在每轮末尾检查 `cond`；body 内的
`^>` 会在跳转前执行同一个检查，因此每次迭代只求值一次 `cond`，
即使条件带副作用。

### 内建别名

Rune 源可直接写 `console.log(...)`，四端都映射到 `@io.println`；多参数按空格
分隔输出在同一行。这是文档/示例逐步去掉 `@io.` 写法的迁移路径。

`for-of` 降为基于下标的循环（可迭代对象只求值一次），因此
`^`/`^>` 作用于真实循环而不是回调；`^>` 会先推进下标。

## 类型映射

| Rune | TypeScript | Go | MoonBit |
| --- | --- | --- | --- |
| `Int32` | `number` | `int`（32 位回绕） | `Int32` |
| `Int64` | `bigint` | `int64` | `Int64` |
| `Double` | `number` | `float64` | `Double` |
| `String` | `string` | `string` | `String` |
| `Bool` | `boolean` | `bool` | `Bool` |
| `Null` | `null` | `any(nil)` | `Unit` |
| `Undefined` | `undefined` | `any`（nil） | `Unit` |
| `Array[T]` | `T[]` | `[]T` | `Array[T]` |
| `A \| B` | 原生 `A \| B` | `any` + 收窄转换 | 暂不支持（待原生枚举） |

`Int32` 与 `Double` 在 TypeScript 里都是 `number`，转回来无法区分，
因此往返可能把 `Int32` 变成 `Double`。这是接受的 L1 限制。

## union 收窄

`typeof` 返回 `"string" | "number" | "boolean" | "object" | "function" |
"undefined"`。`typeof x == "tag"`（含 `===`、`!=`、`!==` 与左右交换）会在
受保护分支内收窄 `x`；如果 then 分支总是 `return`，后续语句自动应用取反后的收窄：

```rune
describe(value: String | Double) -> String => (typeof value == "string") ? {
  value
} : {
  "number"
}
```

`undefined` 是一等值/字面量，在 `typeof` 下与 `null` 区分；`T?` 同时接受
`null` 与 `undefined`。

Go 后端把 union 表示为 `any`，在收窄使用点生成带检查的数值/字符串转换；
MoonBit 后端目前直接拒绝 union（提示 `union types yet`），计划是原生枚举或 tagged union。

## 拒绝的动态特性（D7）

`ts2rune` 会给出稳定的 `RUNE-DYN-*` 错误码；负例在 `tests/conformance/negative/`。

| 错误码 | 构造 |
| --- | --- |
| `RUNE-DYN-001` | `eval` |
| `RUNE-DYN-002` | `Function(...)` / `new Function(...)` |
| `RUNE-DYN-003` | `with` |
| `RUNE-DYN-004` | `Proxy`、`Reflect` trap |
| `RUNE-DYN-005` | `__proto__`、`Object.setPrototypeOf`、`defineProperty` |
| `RUNE-DYN-006` | 动态 `import()` |
| `RUNE-DYN-007` | `arguments` |
| `RUNE-DYN-008` | 隐式 `any` 参数/字段、未声明全局赋值 |
| `RUNE-DYN-009` | 对 `any` / `unknown` 值做成员访问 |
| `RUNE-DYN-010` | `globalThis[name]`、`window[name]` |
| `RUNE-DYN-011` | `debugger` |
| `RUNE-DYN-012` | `require()` |
| `RUNE-DYN-013` | `call` / `apply` 改绑 this |
| `RUNE-PROFILE-001` | `declare`、`namespace`、generator、装饰器 |
| `RUNE-PROFILE-002` | 不支持的 class 特性（继承、访问器、static 字段、复杂构造函数） |

模块只支持静态 `import ... from`。`JSON.parse` 的结果不会自动获得结构化类型，
需要显式标注并收窄。

## 能力检查

`rune check --target go|mbt file.rn` 在 codegen 之前用 `RUNE-CAP-001` 拒绝 DOM
专属全局；`--target ts` 放行。当前覆盖 `document`、`window`、`navigator`、
`localStorage`、`sessionStorage`、`history`、`location`。目录级 check 暂不支持
`--target`。

## 往返级别

- **L1（语义等价）**：`ts -> rune -> go`、`ts -> rune -> ts` 已有
  `internal/tsfront` 测试与 conformance 用例覆盖。
- **L2（结构等价）**：暂不承诺；Rune 独有语法（struct、signal、模式体）会被降糖。
- **L3（文本保真）**：当前不在范围内。

## 过渡 bridge

自举前端/解释器尚未解析 union 与 `typeof`，因此包含这些构造的源码在
`rune check`、`rune run`、`rune build` 时走宿主参考实现；其余语言行为不受影响。
该 bridge 是临时的，进度记录在 RFC 0001 附录 C。
