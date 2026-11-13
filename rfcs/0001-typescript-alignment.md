# RFC 0001：Rune 方向调整 —— 对齐 TypeScript，保留 Go / MoonBit 后端

- 状态：提案（Draft）
- 日期：2026-11-12
- 影响范围：语言语法、类型系统、标准库、IR、三个 codegen（Go / TypeScript / MoonBit）、
  CLI、检查器、自举代码（`selfhost/`）、文档

---

## 0. TL;DR

把 Rune 重新定位为 **TypeScript 的静态同构子集 + 多后端编译器**：

1. **语言不再自带标准库**。类型与 API 的唯一真实来源是 TypeScript 官方的
   `lib.*.d.ts`（ES + DOM），可选叠加 `@types/node` 等声明包。仓库里 `core/`
   的 43 个模块、253 个 `%intrinsic` 名字逐步退役。
2. **语法与 TypeScript 双向可逆**。定义一个小而稳的 `Rune TS Profile`，
   在 `TS 源 ⇄ Rune AST ⇄ Rune 源` 上做到可往返，而不是"长得像"。
3. **Go 与 MoonBit 后端保留**，通过一层 **JS 运行时垫片（runtime shim）** 提供
   ES 内建语义（`Array`/`Map`/`String`/`JSON`/`Math`…），DOM 这类宿主 API
   只对 TS 后端可用。
4. **最终链路**：`TS → Rune → Go | MoonBit`，并保证同一份 TS 输入在
   `→ Rune → TS` 后能回到等价语义，在 `→ Go / MoonBit` 后符合同一套符合性用例。
5. **只支持静态可判定的 TS**。`eval`、`new Function`、`with`、运行时代码生成
   等一切动态特性在语言层一律不支持，用诊断明确拒绝，不做降级。

一句话概括方向：**Rune = 可往返的 TypeScript 静态核心 + Go/MoonBit 代码生成器**。

---

## 1. 现状盘点

| 维度 | 现状（可核对的事实） |
| --- | --- |
| 流水线 | `lexer → parser → AST → checker → IR → interpreter / Go codegen / TS codegen / MoonBit codegen` |
| 标准库 | `core/` 下 43 个模块目录，全部靠 `"%module.intrinsic"` 字符串体声明，共约 253 个唯一 intrinsic 名 |
| 后端分发 | `internal/codegen/{go,typescript,moonbit}/intrinsic.go` 各自维护一份 intrinsic → 目标代码的映射（约 190 / 210 / 320 个 `case`） |
| 模块调用 | 用户侧写法 `@io.println`、`@json.stringify`、`@fs.readFile`；检查器只承认 `core/<m>/<m>.rn` 里声明过的形状 |
| 整数类型 | `Int32` 在 Go 端是平台字长的 `int`（64 位平台上为 64 位），MoonBit 端是 `Int32`，TS 端是 `number`；MoonBit 上游已把该类型改名为 `Int32` |
| 文件导入 | `@"./helper.rn"` + `+` 前缀控制跨文件可见性，没有 `import/export` |
| TS 摄入 | `internal/compiler/ts_import.go` 已用 `go-typescript-eslint` 解析 `.ts`，但只抽取 <code>export function</code> / <code>export const/let/var</code> 的**签名**，不导入语句体 |
| TS 声明 | `internal/tsdecl/` 已能解析 `lib.dom.d.ts`，并生成 `selfhost/compiler/dom_types.rn`（事件名、标签名、事件接口形状），说明"从 .d.ts 拿类型"这条路已经验证过 |
| 控制流 | IR 里有 `IfStmt` / `LoopExpr` / `ContinueStmt`，但**只有 Go 后端**处理，parser 产不出它们，属于历史残留 |
| 自举 | `selfhost/` 约 3.4 万行 `.rn`，`@module` 调用点只有 ~77 处（`@cli` 48、`@go` 9、`@io` 7、`@map` 6、`@json` 3、其余各 1），迁移成本可控 |

### 1.1 现状与新方向的冲突点

1. `core/string`、`core/array` 的 API 形状**已经接近** TS/JS（`split`、`replaceAll`、`trimStart`、
   `map`、`filter`…），但调用方式不同：Rune 用 `value.split(",")` 可以，`@string.fromChars` 不行；
   TS 的 `Array.length` 是属性，Rune 是 `length()` 方法。对齐需要逐个 API 收敛，而不是推倒重来。
2. `core/` 的存在使"标准库等于 TypeScript"无法成立：同一功能两套来源，检查器、LSP、三个后端都要维护两套。
3. Rune 目前**没有** `if` / 经典 `for` / `class` / `try` / `union type` / `import/export`，
   循环写作 `* (cond) { ... }`、`^` / `^>` 控制跳出与继续；这些都是 TS 的日常语法。
   不补齐，就不可能"TS 与 Rune 互转"。
4. 当前 TS 后端是在 **IR 上**直接生成文本，没有"保真打印"能力；做往返必须在**AST 层**加 emitter。

---

## 2. 目标与非目标

### 2.1 目标

- **G1**：标准库来源单一化 —— 只有 TypeScript 的 `.d.ts`，Rune 不定义任何内建 API。
- **G2**：`TS → Rune` 与 `Rune → TS` 在一个显式定义的子集上**可往返**（有自动测试）。
- **G3**：`Rune → Go` 与 `Rune → MoonBit` 保留，且能编译"带 ES 内建 API"的程序（垫片实现）。
- **G4**：端到端打通 `TS → Rune → Go` 与 `TS → Rune → MoonBit`，并可用 conformance 套件验证行为一致。
- **G5**：动态特性一律拒绝：**静态可判定**是语言的第一约束，诊断优先于兼容。
- **G6**：现有工具链（`check` / `fmt` / `run` / `test` / `lsp` / `repl` / 自举）不退化。

### 2.2 非目标

- 不追求 100% 兼容 TypeScript 语言规范，只承诺 `Rune TS Profile` 内的兼容。
- 不实现 TS 的类型级图灵完备能力（条件类型、映射类型、模板字面量类型的完整求值）。
- 不为 Go / MoonBit 后端模拟 DOM；DOM/Web API 只对 TS 后端可用。
- 不保留 `core/` 的长期兼容层（只保留一个有期限的过渡 alias）。
- 不引入 `any` 逃逸口带来的动态行为：`any` 可以出现在**互操作边界**上，但不能据此做动态取值。

---

## 3. 核心决策

### D1 标准库 = TypeScript 的 `lib.*.d.ts`

**决策**：删除 `core/` 作为 API 定义源，改为"内置声明集（lib set）"机制。

```text
lib set = lib.es5.d.ts + lib.es2015..2023.*.d.ts + (dom ? lib.dom.d.ts : []) + 用户声明的 @types/*
```

- 检查器（`internal/checker`）、LSP、三个 codegen 共用同一个 `TypeEnv`，由 `internal/tsdecl` 加载。
- 内置 API 的"存在性"由声明决定；后端是否"能实现"由一个独立的**能力矩阵**决定（见 §7）。
- 参考实现路径已经存在：`internal/tsdecl/cmd/gen-dom-stub` 从 `lib.dom.d.ts` 生成
  `selfhost/compiler/dom_types.rn`。本 RFC 把这套机制从"DOM 专用"推广为"标准库通用"。
- 版本策略：**锁定 TypeScript 版本**（写进仓库，如 `tools/ts-lib/VERSION`），lib 文件作为 vendored 资产，
  升级走一次显式的 snapshot 更新 + 差异评审，避免上游漂移导致后端悄悄失效。

**收益**：一套 API、一套文档、一套补全/悬停/签名帮助；新增 API 只需升级 lib 版本。

**代价**：lib 里大量成员在 Go/MoonBit 上不可实现，必须显式维护"可用性"信息。见 D8。

### D2 语法向 TS 收敛，但保留"可降糖"的 Rune 特性

**决策**：把语言分成两层。

| 层 | 内容 | 用途 |
| --- | --- | --- |
| **Rune Core** | TS Profile 的同构子集：`x := v`（不可变）/`x :=: v`（可变）、`cond ? { } : { }`、`for (const x of xs)`、`* (cond) { ... }`、`^`/`^>`、`return`、`function`、`class`、`try/catch/finally/throw`、`interface`、`type`、`enum`、`import/export`（过渡期 `@"./m.rune"`）、可选 `async/await` | TS 互转的落点；Go/MoonBit 的直接输入 |
| **Rune Sugar** | 现有表达力特性：`x { pattern => ... }`、XML 表达式、`$signal`、`~routine`、`?` 解包、`T: Trait` 结构约束 | 手写 Rune 时的人体工学；**单向降糖到 Core**，不参与 TS 往返 |

**关键约束**：Sugar 不允许改变 Core 的语义。任何 Sugar 必须在 `internal/ir/lower.go` 里
有明确的 Core 展开规则，并保证 `lower(sugar) ≡ lower(desugared)`。

“同构”指 **AST/语义同构**，不是文本同构：TS 的 `while` 由 tsfront 降为 Rune 的
`* (cond)`，`break` / `continue` 分别降为 `^` / `^>`；反向由 astprint 还原。

**收益**：往返只需要守住一个窄核心；Rune 的差异化表达力不用丢。

### D3 模块系统改为 ESM 语义

**决策**：`@"./helper.rn"` + `+` 可见性 → `import` / `export`；模块调用 `@io.println` → 全局对象或导入命名空间。

```rune
// 之前
@"./helper.rn"
main() => @io.println(helper())

// 之后
import { helper } from "./helper.rune"

export function main(): void {
  console.log(helper())
}
```

- `export` 显式化，删除 `+` 前缀与隐式私有默认。
- 模块解析遵循 TS 的相对路径规则；扩展名 `.rune`（见 §10 Q7）。
- `@go.*` / `@mbt.*` FFI 保留，但降级为**语言级扩展**，不参与 TS 往返：
  出现在 TS 侧源码里视为错误，出现在 `rune → ts` 时输出诊断。

### D4 数值模型：`number` 为主，`Int32` 为可选定宽扩展

TS 里**没有**整数类型，`number` 是 f64。Rune 现在的 `Int32` 在 Go 端映射到平台字长的 `int`
（64 位平台上即 64 位），在 MoonBit 端映射到 `Int32`，与 TS 冲突。

同时，MoonBit 上游最近把整数类型改名为 **`Int32`**。Rune 跟随上游，把 `Int` **重命名为 `Int32`**，
并明确它是 32 位二补码整数、溢出按位回绕（与 MoonBit `Int32`、JS `x | 0` 对齐）。

| 方案 | 说明 | 结论 |
| --- | --- | --- |
| A. TS `number` → Rune `Int32` | 精度丢失（`2^53` 以上、位运算、除法语义全变） | 否 |
| B. TS `number` → Rune `Double`，`Int32` 保留为扩展类型 | 往返默认无损；手写 Rune 仍可用定宽整数 | **采纳** |
| C. 完全去掉整数族 | 现有 `core/*int*`、位运算、自举代码全要改，且 Go 端性能变差 | 否 |

规则：

- `number` ⇄ `Double`（f64）为**默认**双向映射。
- `Int32` 是 Rune 扩展类型，`rune → ts` 时按 `number` 输出。
- 位运算只对 `Int32`、`Int64`、`BigInt` 等整数类型开放；TS 侧 `x | 0` 这类"用位运算截断"写法
  在 `ts → rune` 时映射为 `Int32`，回程还原为同样的位运算表达式。
- 迁移动作：
  - 类型名 `Int32` → `Int32`（`internal/checker` 的 `Int32` 常量、`core/int` → `core/int32`、
    `%int.*` → `%int32.*`，过渡期用，M4 后随 `core/` 退役）；
  - Go 后端保持 `int` 作为承载类型（避免 `len`/索引的全量转换），但 Int32 的算术与
    位运算在生成处按 32 位回绕（`int(int32(...))`），语义与名字一致；
  - MoonBit 后端类型从 `Int32` 改为 `Int32`（对齐上游改名）；
  - `Int8` / `Int16` / `Int64` / `UInt*` 保持不变。

#### D4.1 已接受的往返损失

`Int32 → TS number → Rune` 会得到 `Double`，**这是有意接受的单向损失**，不做自动还原：

```text
Int32  ──rune → ts──►  number  ──ts → rune──►  Double   // 类型标注丢失，终态
Double ──rune → ts──►  number  ──ts → rune──►  Double   // 稳定
```

- **值不丢**：所有 `Int32` 值都能被 f64 精确表示，所以 `ts → rune → ts` 的**运行时行为等价**
  （L1 仍然成立）。
- **类型会丢**：一次往返后，原本依赖 `Int32` 的位运算、重载或定宽 API 会类型不匹配。
  这是"TS 没有整数"的必然结果，文档与诊断中直接说明，不用注释元数据去挽救。
- 因此：**`rune → ts` 是"定宽语义出口"**。需要保留定宽语义的代码，应以 `.rune` 为主源码，
  只在需要交给 TS 生态时单向导出。
- L3 文本保真只承诺"无定宽整数标注"的子集；含 `Int32` 的源码按 L1 处理。

### D5 字符串语义：UTF-16 为准

TS 的 `String` 索引、`length`、`charCodeAt` 都是 **UTF-16 码元**语义。Go 的 `string` 是 UTF-8 字节序列，
MoonBit 的 `String` 本身是 UTF-16 —— 三者中 MoonBit 天然对齐，Go 需要垫片。

| 后端 | 策略 |
| --- | --- |
| TS | 原生 |
| MoonBit | 原生（UTF-16） |
| Go | `runtime/shims/jsstring`：内部 `[]uint16` 或 `utf16.Encode` 包装，索引/长度/码元操作走该类型；`==`、`+`、范围遍历保持可用 |

**性能开关**：`--string=utf16|native`。`native` 用 Go 原生 string（更快，但 `s[i]`、
`length` 语义与 TS 不一致）；默认 `utf16`，保证语义正确。

### D6 类型系统补齐：union / undefined / class / 异常

TS 的日常代码离不开 Union 与 class。缺一样，往返就不成立。

1. **Union 与收窄**
   - 类型语法：`A | B`；
   - 收窄：`typeof`、`instanceof`、判别字段（discriminant）、`null/undefined` 检查、
     `switch` 穷尽分支；
   - 表示：TS 后端原生；Go 后端用**标记联合（tagged union）**生成；
     MoonBit 后端优先映射到 MoonBit 原生枚举，无法映射时退回 tagged union；
   - 与现有 `Result[T, E]` 的互操作：`Result` 是 Union 的特例，不重复建模。
2. **`null` / `undefined`**
   - 新增 `Undefined` 类型，`T?` 精确含义为 `T | null | undefined`（严格模式下 `?` 等价于可选字段标记）；
   - Go 后端：`undefined` 用零值 + `js.Undefined` 哨兵区分（仅在需要区分时生成），
     MoonBit 用 `Option`。
3. **`class`**
   - 支持：字段、方法、构造器、继承、`static`、`public/private/protected`、`readonly`、
     访问器（get/set）、抽象类；
   - 后端：Go 用 struct + 方法集 + 接口表实现虚派发；MoonBit 用 struct + trait（或已有对象模型）实现；
   - 首版可延后：装饰器、`#private`、复杂 mixin 模式。
4. **异常**
   - 新增 `throw` / `try` / `catch` / `finally`（TS 语义）；
   - Go 后端：`panic`/`recover` 包一层 `js.Throw`，`finally` 用 defer；
   - MoonBit 后端：`Result` + 显式错误传播，`throw` 编译为返回 `Err`（与现有 `~routine` 机制合流）；
   - 与 `?` 的关系：`?` 是 Rune Sugar，展开为 `try/catch` 或 `Result` 传播，二者语义必须一致。
5. **`switch`**
   - 新增语句形式（现有 `subject { ... }` 保留为 Sugar），fallthrough 规则与 TS 一致。
6. **`if` / 循环**
   - 条件仍以 `cond ? { ... } : { ... }` 表达式表达；循环语句为
     `* (cond) { ... }`，`^` 跳出、`^>` 继续；
   - IR 的 `IfStmt` / `LoopExpr` 复活并补齐 TS / MoonBit / interpreter 的支持。

### D7 只支持静态 TS：动态特性一律拒绝

**这是本 RFC 的硬边界。** 任何依赖运行时动态查找 / 代码生成 / 篡改语义的特性，都不进入语言，
也不做"尽力而为"的降级实现。

| 特性 | 处置 |
| --- | --- |
| `eval(...)` | 不支持。全局 `eval` 从 lib 视图中移除，调用即编译错误 `RUNE-DYN-001` |
| `new Function(...)` / `Function(...)` | 不支持，`Function` 构造函数从 lib 视图移除（`Function` **类型**保留，用于函数签名） |
| `with` | 不支持 |
| `Proxy` / `Reflect` 的 trap 能力 | 不支持；`Reflect.get/set` 等简单转发可作为 host-only 扩展，不进入 TS Profile |
| 动态改原型：`Object.setPrototypeOf`、`__proto__`、运行时增删方法 | 不支持；class 形状必须静态确定 |
| `Symbol.toPrimitive` 等钩子 | 仅允许垫片内部使用，用户代码禁用 |
| `import()` 动态导入 | 不支持；只支持静态 `import ... from "..."` |
| `require()` | 不支持；Node 互操作走静态 `import` |
| `arguments` 对象 | 不支持；用 rest 参数替代 |
| `debugger` | 不支持 |
| 非静态 `this`（`call/apply/bind` 改绑） | 受限：`bind` 仅在编译期可确定签名时允许；`call/apply` 不支持 |
| 赋值给未声明的全局、隐式 `any` | 严格模式报错（`noImplicitAny` 语义） |
| 运行时字符串反射（`globalThis[name]`、`window[eventName]`） | 不支持；必须用静态分支 |
| `JSON.parse` 结果当结构化类型 | 不自动推导；必须显式标注或校验后收窄 |
| 标签模板 + 动态标签 | 只支持静态已知标签函数 |
| generator / `yield` | 首版不支持（可在 L3 之后评估） |

**命名与诊断**：所有拒绝都要有稳定错误码（`RUNE-DYN-001`…），并在
`rune check --profile` 里可一次性列出源码中所有越界点，便于使用者判断迁移成本。

**注意**：拒绝的是"动态行为"，不是"动态类型"。`any` / `unknown` 仍是**类型**，
但在 Rune 里的用法受限：`any` 只能整体传递或整体断言，不能 `x.foo` 任意取成员
（这会让检查器失去形状信息，也让 Go/MoonBit 无法生成代码）。

### D8 后端能力用"能力集"表达，而不是散落的 `case`

lib 里有 `document.querySelector`，但 Go 没有 DOM。这类差异必须**显式、可查询**，否则
用户会在 codegen 阶段才撞墙。

```text
capability set = { es.core, es.promise, es.async, dom, node.fs, node.path, ... }
backend       = { ts: all, go: es.core + es.promise + node.*, mbt: es.core + node.* }
target triple = <backend> × <capability set>
```

- 每个 API 在声明层（而不是实现层）标注能力归属，例如 `console.log` 属于 `es.core`（各端都支持），
  `document` 属于 `dom`（仅 TS）。
- `rune check --target go` 在**类型检查阶段**就报 `RUNE-CAP-xxx: document is only available on target ts`，
  而不是等 codegen。
- 现有 `internal/codegen/*/intrinsic.go` 的巨型 switch 逐步替换为"垫片库 + 能力表"，见 §7。

### D9 往返保真度分级

不承诺所有情况都逐字节往返。定义三个级别，测试和文档都按级别表述。

| 级别 | 承诺 | 实现方式 | 验收 |
| --- | --- | --- | --- |
| **L1 语义等价** | `ts → rune → ts` 后再编译，行为等价 | AST 级 emitter + 类型重写 | conformance 用例全绿 |
| **L2 结构等价** | 声明、类型、控制流结构一致（允许格式/命名辅助差异） | L1 + 稳定命名 + 保留注释 | AST diff 用例全绿 |
| **L3 文本保真** | 除注释外逐 token 一致；Rune 扩展用 `// @rune:` 元数据承载 | trivia 保留 + 元数据通道 | 快照测试 |

**明确不承诺**：Sugar 特性（XML、signal、pattern body、routine）反向生成 TS 源码文本；
它们只能通过 L1 语义降糖往返（例如 XML → `document.createElement` 调用序列）。

---

## 4. 新流水线

```text
              ┌───────────────┐
   .ts  ────► │  tsfront      │  TS AST → Rune AST（带位置、注释、类型）
              └──────┬────────┘
                     │
   .rune ────────────┴────► parser ──► Rune AST
                                          │
                                   ┌──────┴───────┐
                                   │  checker      │  TypeEnv from lib.*.d.ts
                                   │  + capability │  能力校验
                                   └──────┬───────┘
                                          │
                                        IR（补齐语句：if/loop/switch/try/class）
                                     ┌────┴────┬─────────┬──────────┐
                                     │         │         │          │
                                  interpreter  ts AST   go emit   mbt emit
                                              printer     │          │
                                                 │   js runtime    js runtime
                                                 │     (Go)        (MoonBit)
                                               .ts
```

新增组件：

| 组件 | 位置（建议） | 职责 |
| --- | --- | --- |
| `tsfront` | `internal/tsfront/` | 完整 TS → Rune AST 转换（复用现有 `go-typescript-eslint` 解析器） |
| `rune/astprint` | `internal/astprint/` | Rune AST → 源码，保留 trivia；fmt 与 `ts` 输出共用 |
| `libenv` | `internal/libenv/` | lib set 加载、版本锁定、能力表、快照 |
| `shim/go` | `runtime/go/` | Go 版 JS 运行时（ES 内建 + Node 最小集） |
| `shim/mbt` | `runtime/mbt/` | MoonBit 版 JS 运行时 |
| `conformance` | `tests/conformance/` | 四端共享的行为用例（源 + 期望输出） |
| CLI | `cmd/rune` | `rune ts2rune` / `rune rune2ts` / <code>rune build --target go&#124;mbt</code> |

**为什么 tsfront 走 AST 而不是走 IR**：IR 丢失注释、泛型写法、类型别名、声明顺序等信息，
做不到 L2/L3；AST 级转换是往返的唯一可行路径。IR 仍然是 L1 语义和 Go/MoonBit 后端的输入。

---

## 5. 语法映射表

### 5.1 声明与模块

| TypeScript | Rune | 备注 |
| --- | --- | --- |
| `function f(a: T): R {}` | `function f(a: T): R {}` | 完全同形 |
| `const f = (a: T): R => {}` | 同形 | 保留箭头函数 |
| `interface I { a: T }` | `interface I { a: T }` | 结构类型，与现有匿名对象模型统一 |
| <code>type X = A &#124; B</code> | <code>type X = A &#124; B</code> | 新增 union |
| `enum E { A, B }` | `enum E { A, B }` | 现有 enum 语法对齐 |
| `class C extends B {}` | 同形 | 新增 class |
| `import {x} from "./m"` | 同形（扩展名 `.rune`） | D3 |
| `export default f` | 同形 | |
| `declare` / `namespace` | 不进入 Profile | 仅 `.d.ts` 摄入，不能在 `.rune` 里写 |
| `const x: T = v` | `x := v : T` | 不可变绑定 |
| `let x: T = v` / `var x` | `x :=: v : T` | 可变绑定；`:=:` 才允许后续 `=` 赋值 |
| `@decorator` | 首版不支持 | 诊断拒绝 |

### 5.2 语句

| TypeScript | Rune Core | Sugar（Rune 独有） |
| --- | --- | --- |
| `if/else` | `cond ? { ... } : { ... }` | 三元块表达式（已有） |
| `for(;;)` / `for-of` / `for-in` | 新增 | `.each` / `.map`（现有 API） |
| `while` / `do-while` | `* (cond) { ... }`（`do-while` 由 tsfront 降糖） | 递归 |
| `switch` | 新增 | `subject { p => ... }`（现有） |
| `try/catch/finally`、`throw` | `{ } !? { e => } -- { }`、`!! value` | `?` 解包（Routine 内） |
| `break/continue[/label]` | `^` / `^>`（label 首版不支持） | — |
| `return` | 新增（`=>` 尾部表达式为 Sugar） | — |

### 5.3 表达式

| TypeScript | Rune | 备注 |
| --- | --- | --- |
| `a?.b`、`a ?? b` | 已支持 | 现有 optional chaining |
| 模板字符串 / `${}` | 已支持（Rune 用 `\(...)`，TS 侧转换） | 需在 L2 保留原写法 |
| 空值合并、可选链、解构 | 已支持（`internal/ast` 有 `ObjectPattern` 等） | 复用 |
| `new C()` | 新增（class） | |
| 生成器、`yield` | 不支持 | D7 |
| 标签模板 | 静态标签 | 动态拒绝 |

### 5.4 库里"方法 vs 属性"的差异（会导致往返失败，必须处理）

示例（`core/array/array.rn` 与 TS 的差异）：

```text
TS:    xs.length        xs.forEach(f)      xs.push(x)
Rune:  xs.length()      xs.each(f)         xs.push(x)
```

处理策略：

- 以 **TS 形状为准**：`length` 作为属性进入类型模型（LSP/检查器层面），
  Rune 源码仍可写 `length()` 作为 Sugar，lower 到同一 IR 节点；
- `each` 作为 `forEach` 的 Sugar 别名保留一个版本周期，之后移除；
- 所有类似差异汇总进 `docs/language/ts-compat.md`，作为必须维护的"差异清单"。

---

## 6. 标准库落地方案

### 6.1 三层划分

| 层 | 内容 | TS 后端 | Go 后端 | MoonBit 后端 |
| --- | --- | --- | --- | --- |
| **S0 ES 内建** | `Array`、`String`、`Number`、`Boolean`、`Object`、`Math`、`JSON`、`Map`、`Set`、`Date`、`RegExp`、`Promise`、`Error` | 原生 | 垫片 | 垫片 |
| **S1 宿主 API** | DOM、`fetch`、`localStorage`、`requestAnimationFrame` | 原生 | 拒绝（capability 诊断） | 拒绝 |
| **S2 平台扩展** | `@types/node`：`fs`、`process`、`path`、`Buffer`、`net` | 原生 | 垫片（已有 `core/fs` 等可直接复用） | 垫片（同理） |

关键点：**S0/S2 的 API 必须由垫片提供，而不是继续用 `core/*` 描述**。`core/fs` 里已有的
`readFile/readFileText/exists/readdir/mkdir/remove/stat` 实现可以**平移**成
`runtime/go/nodefs` 里的 `fs.promises.*` 实现，声明则来自 `@types/node`。

### 6.2 垫片设计原则

1. **按类型实现，不按函数名实现**：`Array<T>`、`Map<K,V>`、`String` 各自是一个类型 + 一组方法，
   而不是 253 个散落的 intrinsic 名。
2. **与现有 intrinsic 机制共存一个周期**：`core/` 退役期用 alias 表把 `@io.println` 指向 `console.log`，
   让老代码、`selfhost/` 能继续编译（见 §9 M4）。
3. **能力标注在声明层**：lib 快照里每个成员带 `cap` 标记，由 `internal/libenv` 生成。
4. **测试从第一天就是四端共享的**：每个垫片行为都要有 conformance 用例，不允许"只有 Go 测过"。

---

## 7. 运行时垫片

### 7.1 Go 侧（`runtime/go`）

| 包 | 提供 |
| --- | --- |
| `js/value` | `undefined` 哨兵、`null` 区分、类型断言辅助、结构化相等 |
| `js/array` | 增长数组，方法完整实现（`map/filter/reduce/sort/splice/slice/find/...`），回调语义（`this` 固定为 `undefined`） |
| `js/string` | UTF-16 语义字符串（D5） |
| `js/map`、`js/set` | 保持插入顺序，键比较遵循 SameValueZero |
| `js/json` | `JSON.parse`/`stringify`（`undefined`、函数、循环引用行为对齐） |
| `js/math`、`js/number`、`js/date`、`js/regexp` | f64 语义（`Math.*` 用 `math` 包，`NaN`/`-0` 对齐） |
| `js/promise` | 微任务队列 + `Promise.all/race/allSettled` |
| `js/throw` | TS 异常 → Go panic 的桥接与 `finally` 展开 |
| `node/fs`、`node/path`、`node/process`、`node/net` | Node 最小集，尽量平移现有 `core/` 实现 |

### 7.2 MoonBit 侧（`runtime/mbt`）

- 语义目标与 Go 侧一致，实现方式按 MoonBit 的能力选（枚举优先、必要时 struct + 方法表）。
- MoonBit 的 `String` 已是 UTF-16，`String` 相关成本最低。
- 需要显式覆盖：`Array` 的稀疏语义（TS 允许 hole，首版可拒绝稀疏数组）、`Map` 的插入序、
  `Promise`（映射到 `async fn` + 现有 `Task` 机制）、异常（`throw` → `Result` 传播）。

### 7.3 不实现的语义（列为已知差异）

- `__proto__`、`Object.setPrototypeOf`、`Proxy`、`Reflect` trap：D7 拒绝。
- 原型链上的动态方法查找：class 方法必须静态可知。
- `arguments`、`with`、`eval`：D7 拒绝。
- 稀疏数组（`new Array(5)` 的 hole 语义）：首版拒绝，diagnostic 说明。

---

## 8. 符合性测试

四端（interpreter / Go / MoonBit / TS）共享一套用例，输入是 `.rune`（或 `.ts`），期望输出与错误码固定。

```text
tests/conformance/
  es/            Array/String/Map/Set/JSON/Math 行为
  types/         union 收窄、可选链、null/undefined、class 继承
  control/       if/for-of/loop-symbols/switch/try-finally
  node/          fs/path/process
  roundtrip/     ts → rune → ts（L1/L2）、rune → ts → rune（L3）
  negative/      D7 拒绝用例（每个错误码至少一条）
```

CI 矩阵（必须全绿才算阶段完成）：

| 用例集 | interpreter | Go | MoonBit | TS |
| --- | --- | --- | --- | --- |
| es | ✔ | ✔ | ✔ | ✔ |
| types | ✔ | ✔ | ✔ | ✔ |
| control | ✔ | ✔ | ✔ | ✔ |
| node | ✔ | ✔ | ✔ | ✔ |
| roundtrip L1/L2 | — | — | — | ✔ |
| negative | 诊断一致 | 诊断一致 | 诊断一致 | 诊断一致 |

补充门禁：

- **往返不动点**：`ts → rune → ts → rune` 第二次结果与第一次相同（幂等）。
- **诊断一致性**：同一非法程序在四个路径上报同一错误码。
- **自举不得回退**：`selfhost/` 全部 `.rn` 必须能被新前端解析/检查/生成。

---

## 9. 实施路线

每个里程碑都以"可验证的产物"结束，不做纯重构阶段。

### M0：冻结与基线（1 周）

- [ ] 冻结 `core/`：不再新增模块与 intrinsic。
- [ ] 建立 `tests/conformance/` 骨架，把现有 `tests/*.rn` 中可复用用例登记进去。
- [ ] 锁定 TypeScript 版本，vendored lib 入库（`tools/ts-lib/`）。
- [ ] 记录基线：四端各自当前通过的用例数，作为后续不退化基线。

**完成判据**：CI 有 conformance 任务，基线报表可查。

### M1：lib 成为唯一类型来源（2–3 周）

- [ ] 新增 `internal/libenv`：加载 lib set，生成统一 `TypeEnv`，替代 checker 对 `core/` 的依赖。
- [ ] checker / LSP 改读 `TypeEnv`；`@module.fn` 调用只作为 transition alias 解析。
- [ ] 能力表初版（`es.core / dom / node.*`），`rune check --target` 接入。

**完成判据**：`rune check examples/fib.rn` 不再读取 `core/` 即可通过；`document` 在 `--target go` 下报 capability 诊断。

### M2：语法补齐（4–6 周，可与 M3 并行）

- [ ] parser/checker/AST：`import/export`、`if/else`、`for/for-of/for-in`、`* (cond)` 循环、
      `^`/`^>`、`switch`、`return`、`union`、`Undefined`、`class`、
      `try/catch/finally/throw`、`new`。
- [ ] 整数改名：`Int32` → `Int32`（checker 常量、`core/int` → `core/int32`、`%int.*` → `%int32.*`），
      Go 后端改出 `int32`、MoonBit 后端改出 `Int32`；整数语义固定为 32 位二补码回绕。
- [ ] lowering：上述结构进 IR；`LoopExpr`/`IfStmt` 补齐 TS / MoonBit / interpreter。
- [ ] 语法与类型映射文档：`docs/language/ts-compat.md`（含差异清单）。
- [ ] 负例：D7 全部拒绝项 + 错误码。

**完成判据**：`tests/conformance/{types,control,negative}` 在四端全绿。

### M3：TS 前端（tsfront）（4–6 周）

- [ ] `internal/tsfront`：完整 AST 转换（声明、语句、表达式、类型、注释）。
- [ ] `rune ts2rune in.ts -o out.rune`；`rune check --profile` 列出越界点。
- [ ] `internal/astprint` 保真打印（trivia、注释、原始写法）。
- [ ] L1 往返：`ts → rune → ts` 语义等价用例。

**完成判据**：取 300–500 行真实 TS 文件（不含 D7 特性）完成往返，conformance 通过；
D7 用例全部报稳定错误码。

### M4：垫片与自举迁移（6–8 周）

- [ ] Go 垫片：`Array/String/Map/Set/JSON/Math/Promise/Error` + `node/fs/path/process`。
- [ ] MoonBit 垫片：同集合，按 MoonBit 能力实现。
- [ ] `core/` alias 层：`@io.*` → `console.*`、`@fs.*` → `fs.promises.*`、`@json.*` → `JSON.*`、
      `@map/@set/@string/@array` → 对应内建；有效期一个版本。
- [ ] `selfhost/*.rn`（~77 处 `@module` 调用）迁移到新 API 并保持自举绿灯。

**完成判据**：`tests/conformance/{es,node}` 在四端全绿；自举测试全绿；`core/` 只被 alias 层引用。

### M5：端到端目标链路（3–4 周）

- [ ] `rune build --target go in.ts`、`rune build --target mbt in.ts` 端到端可用。
- [ ] L2 往返：结构等价用例；L3 文本保真落在"无 Rune 扩展"子集上。
- [ ] 性能基线：`Array/String` 密集用例与手写 Go 的差距记录（目标：≤ 2x，无下限承诺）。

**完成判据**：用同一份 TS 程序产出可运行的 Go 与 MoonBit 二进制，输出一致。

### M6：清理与文档（2 周）

- [ ] 删除 `core/` 与 `@module` 语法（或标记为 deprecated + 移除计划）。
- [ ] 删除三个 `intrinsic.go` 中已由垫片取代的分支。
- [ ] 文档：README、`docs/language/*`、`docs/zh/*`、playground 示例全部改为新写法。
- [ ] VSCode 扩展：语法高亮、snippet、LSP 语义同步。

**完成判据**：仓库内不存在对 `core/` 的运行期依赖；文档无遗留 `@io.` 写法。

---

## 10. 影响面与破坏性变更

| 影响 | 规模 | 处置 |
| --- | --- | --- |
| `core/` 43 模块 / 253 intrinsic | 全量 | 退役；S2 实现平移进垫片，S1 直接删除 |
| `examples/` / `tests/` 里的 `@io`/`@map`/`@json` 写法 | 中 | 批量迁移脚本 + 手工复核 |
| `selfhost/` 3.4 万行 | ~77 处调用点 | M4 迁移，逐文件绿灯推进 |
| 三份 `intrinsic.go` 巨型 switch | 约 720 个 case | 逐步替换为能力表 + 垫片，M6 收尾 |
| TS 后端（IR → 文本） | 需要双通道 | 保留现有快速通道；新增 AST 保真通道用于往返 |
| LSP/completion | 类型来源切换 | M1 一次性切换，靠现有 parity 测试守门 |
| Playground / 文档 / 网站 | 全量示例 | M6 统一更新 |

**用户可见的破坏性变更**（需要在 README 顶部与 release note 明确标注）：

1. `@io.println(x)` → `console.log(x)`
2. `@json.stringify(x)` → `JSON.stringify(x)`
3. `@"./a.rn"` + `+` → `import/export`
4. `Int` 重命名为 `Int32`，溢出行为变为 32 位回绕（Go 后端用 `int` 承载、运算处收窄）
5. `xs.length()` → `xs.length`（方法形式保留为 Sugar，一个版本后移除）
6. D7 动态特性从"可能能用"变为"确定不支持"
7. 已接受的有损往返：`Int32 → TS number → Double`（值不丢，类型标注丢失，见 D4.1）

---

## 11. 风险与对策

| 风险 | 等级 | 对策 |
| --- | --- | --- |
| 各端垫片行为漂移 | 高 | conformance 用例作为唯一标准；新 API 必须有四端用例才能合并 |
| Go 端 UTF-16 字符串性能 | 中 | 默认严格、`native` 逃生；热点用例单独 profile |
| `Int32`/`Double` 双模型导致类型错误 | 中 | 默认 `Double`；`Int32` 只在显式标注或位运算下出现；诊断给出两选一建议 |
| 往返在真实 TS 代码上失败 | 高 | 从第一天就用真实文件做语料；D7 越界点用 `--profile` 提前暴露 |
| lib 版本漂移导致能力表失效 | 中 | 版本锁 + snapshot 生成器 + 差异评审；CI 有 lib 一致性检查 |
| 语法扩展过多，Core/Sugar 边界失控 | 中 | Sugar 必须给出 lowering 规则才允许存在；Core 只接受能映射到 TS 的语法 |
| 自举迁移拖慢 | 中 | alias 层先行，M4 允许自举暂走 alias，M6 再清除 |

---

## 12. 待决问题

1. **已决：`Int32` 改名 `Int32`**（对齐 MoonBit 上游），并长期保留为可选扩展；接受
   `Int32 → ts → Double` 的类型标注损失。新的待决点是：`Int32` 与 `Double` 之间是否允许**隐式**互转？
   建议：只允许字面量与显式转换，避免 32 位回绕在不经意间发生。
2. **异常 vs `Result`**：TS 转来的代码天然用 `try/catch`；手写 Rune 更习惯 `?` + `Result`。
   两者并存需要明确"推荐写法"，建议：**互操作边界用异常，库内部用 `Result`**。
3. **lib set 的默认集合**：只 `lib.es*` + `lib.dom`？还是默认含 `@types/node`？
   建议：CLI 默认 `es + dom`，`--lib node` 显式开启。
4. **文件扩展名**：`.rune` 还是继续 `.rn`？建议 `.rune`（明确、不与其它格式冲突），迁移期双识别。
5. **异步模型**：TS 的 `async/await` 是否映射到现有 `~routine`，还是直接实现 `Promise<T>`？
   建议：`Promise<T>` 为类型，`async/await` 为语法；`~routine` 保留为 Sugar（等于 `async`）。
6. **是否支持 `.js` 输入**：建议不直接支持，先 `js → ts`（用 Babel/typescript 的解析器）再进 `tsfront`。
7. **DOM 测试策略**：TS 后端需要无头浏览器；建议只在 `tests/conformance/dom` 里跑，且不进入四端矩阵。

---

## 13. 与其他语言的对照（不作为依据，仅作参考）

- **AssemblyScript / Static TypeScript**：证明了"TS 语法 + 静态子集 + 多后端"可行，且必须接受"子集"事实。
- **Porffor**：TS 语义直接编译到 wasm，说明"不做 eval 也能覆盖大量真实代码"。
- **MoonBit 的 JS 后端**：反方向的映射，其 runtime 设计可借鉴本 RFC 的垫片分层。

共同结论：**先把子集定义清楚，再谈覆盖度**。本 RFC 的 `Rune TS Profile` + D7 黑名单就是那个子集边界。

---

## 14. 立即可以开工的 5 件事

1. 在 `rfcs/` 提交本文件，评审 D1/D4/D7 三个最关键的决策。
2. 搭 `tests/conformance/` 骨架与 CI 矩阵（空用例也要跑通）。
3. 做 `internal/libenv` 的最小版本：只加载 `lib.es5.d.ts` + `lib.dom.d.ts`，让 `rune check --target go` 能报 capability 错误。
4. 给 parser 加 `if/else` 与 `for-of` 两个语法（PR 最小、可独立验证），跑通 IR → Go/TS/MoonBit。
5. 在 `docs/language/ts-compat.md` 落第一版语法/标准库差异清单（哪怕只有 20 条）。

---

## 附录 A：D7 错误码草案

| 错误码 | 触发 |
| --- | --- |
| `RUNE-DYN-001` | 使用 `eval` |
| `RUNE-DYN-002` | 使用 `new Function` / `Function(...)` |
| `RUNE-DYN-003` | 使用 `with` |
| `RUNE-DYN-004` | 使用 `Proxy` / `Reflect` trap |
| `RUNE-DYN-005` | 动态原型修改（`__proto__`、`Object.setPrototypeOf`） |
| `RUNE-DYN-006` | 动态 `import()` |
| `RUNE-DYN-007` | 使用 `arguments` |
| `RUNE-DYN-008` | 对未声明全局赋值 / 隐式 `any` |
| `RUNE-DYN-009` | 对 `any` 值做成员访问 |
| `RUNE-DYN-010` | 运行时字符串索引反射（`globalThis[name]`） |
| `RUNE-DYN-011` | 使用 `debugger` |
| `RUNE-DYN-012` | 使用 `require()` |
| `RUNE-DYN-013` | 使用 `call` / `apply` 改变 this 绑定 |
| `RUNE-CAP-001` | API 在目标后端不可用（如 DOM on go） |
| `RUNE-PROFILE-001` | 使用 `declare` / `namespace` 等 Profile 外语法 |

## 附录 B：示例对照

```ts
// input.ts
export interface User {
  name: string
  age: number
}

export function adults(users: User[]): string[] {
  const out: string[] = []
  for (const u of users) {
    if (u.age >= 18) {
      out.push(u.name)
    }
  }
  return out
}

console.log(adults([{ name: "Ada", age: 36 }, { name: "Bob", age: 12 }]))
```

```rune
// input.rune（ts2rune 输出，L2 结构等价）
export interface User {
  name: String
  age: Double
}

export function adults(users: Array[User]): Array[String] {
  out := []
  for (const u of users) {
    if (u.age >= 18.0) {
      out.push(u.name)
    }
  }
  return out
}

console.log(adults([{ name: "Ada", age: 36.0 }, { name: "Bob", age: 12.0 }]))
```

```go
// go 后端生成（示意）
func adults(users []User) []string {
  out := jsarray.New[string]()
  for _, u := range users {
    if u.Age >= 18 {
      out.Push(u.Name)
    }
  }
  return out.Unwrap()
}
```

## 附录 C：本次实现状态（2026-11-12）

已落地：

- `Int` → `Int32` 全链路重命名：checker 常量与类型解析、`core/int` → `core/int32`、
  `%int.*` → `%int32.*`、Go / TypeScript / MoonBit 三端 intrinsic 分发。
- 宿主 Go 后端与自举解释器对 Int32 算术/位运算做 32 位回绕；整数字面量超出 32 位时按回绕输出。
- `selfhost/` 源码与 bootstrap 产物同步重命名；`rune check/run/test/go/ts/mbt` 自举路径可用。
- 文档、示例、测试、VSCode 语法与 snippet 同步到 Int32。
- 顺带修复了两个原本失败的用例：对象方法体的 lookahead（支持 `name(params) -> R =>`）、
  空 `{}` 在 `Map` 上下文中的类型推断。
- M0：`tests/conformance/` 骨架与 `internal/conformance` 执行器落地，首批覆盖
  Int32 回绕、三元表达式、记录类型、`if/else`；执行器支持 interpreter / Go / TS / MoonBit
  四个 target（缺少 node/moon 时自动跳过）。
- M1 基础：`internal/libenv` 可以合并加载 `lib.es5.d.ts` / `lib.dom.d.ts`，带能力标注
  （`es.core` / `dom` / `node`）与版本探测；`tools/ts-lib/VERSION` 锁定 TypeScript 6.0.3；
  `core/` 标记为冻结，不再新增模块或 intrinsic。
- Conformance 首轮跑出的两个跨端差异已修：Go codegen 用 `runeInt32Zero` 阻止常量折叠
  提前报 int32 溢出；解释器 `println` 多参数改为空格分隔，与 Go/TS/MoonBit 对齐。
- M2 第一步修正：`if` 语句同样偏离了 Rune 原有语法，已从宿主 parser 移除。
  原有条件表达式仍为 `cond ? { ... } : { ... }`（支持块分支）；TS `if` 由 tsfront
  降为同样的三元块表达式。
- M2 第二步修正：此前加入的 `switch (value) { case ... }` 语法偏离了 Rune 原有
  模式分支写法，已移除。原始语法仍为 `{ 1 => "one"; _ => "other" }`，四端
  conformance 使用该形式；TS `switch` 的转换（建议降为 if/else 或模式分支）待后续实现。
- M2 第三步：`for (const value of array) { ... }` 已支持（`let` 同样接受）。
  宿主将其降为 `array.each((value) => { ... })`，四端 conformance 通过；selfhost
  parser 同样降为 Call/Lambda，formatter 保留 for-of 写法，`check/run/test/fmt` 已验证。
  当前仅支持 Array，且不支持 `break` / `continue`。
- M2 第四步：`while (cond) { ... }` 已支持。宿主新增 `ast.WhileExpr` / `ir.WhileExpr`，
  checker 校验 Bool 条件，解释器与 Go/TS/MoonBit codegen 均实现；selfhost parser 解析为
  `ExprKind.Loop`，checker/infer/interpreter/formatter 同步支持。四端 conformance 通过。
- M2 第五步：`return expr` / 裸 `return` 已支持。宿主新增 `ast.ReturnStmt` / `ir.ReturnStmt`，
  解释器通过 return 信号传播到函数边界；非尾位置的 `if` 降为 `ir.IfStmt`，避免
  return 被三元 IIFE 吞掉。自举 parser 新增 `ExprKind.Return`，checker/infer/interpreter/
  formatter 同步支持，四端 conformance 通过。
- M2 第六步：`break` / `continue` 已支持（当前作用于 `while`）。宿主新增
  `ast.BreakStmt` / `ast.ContinueStmt` / `ir.BreakStmt`，解释器通过控制信号处理；
  自举 parser 新增 `ExprKind.Break` / `Continue`，解释器与 formatter 同步支持。
  四端 conformance 通过。
- M3 起步：新增 `internal/tsfront`，可把 TypeScript 的导出函数、局部变量、
  `return`、`if`、`while`、`break`/`continue`、`console.log`、二元/一元表达式转成
  Rune AST；`number` 映射为 `Double`。转换结果已通过 checker → IR → Go codegen
  并实际 `go run` 验证。union / class / try-catch / 泛型仍会明确拒绝。
- M3 第二步：新增 `internal/astprint`（Rune AST → 源码）与 `rune ts2rune`
  命令；已打通 `ts -> rune -> check -> run -> build`。同时修复自举解释器
  Double 运算结果打印为空的问题（通过 `@go.expr("fmt.Sprint($value)")` 格式化）。
- M3 第三步：tsfront 增加数组字面量、数组索引、for-of、逻辑/条件表达式、
  update 表达式等；自举解释器 Double 输出统一走 `fmt.Sprint`，与 Go/TS 的
  `1` 而非 `1.0` 对齐。
- M3 第四步：将 TypeScript 解析器依赖 vendored 到 `third_party/go-typescript-eslint`
  并修补两个上游 bug：对象字面量属性值被误当类型注解、标识符后的 `<` 被误判为
  泛型实参。tsfront 现已支持对象字面量和裸 `<` 比较。
- M3 第五步：修补模板字符串 lexer 续扫（`}` 后恢复 TemplateMiddle/Tail），
  `${...}` 模板字符串现已支持；`rune ts2rune` 产物可运行并输出正确字符串。
- M3 第六步：tsfront 支持箭头函数/函数表达式回调，`Array.map` 等回调 API
  已可用于 TS → Rune → run。
- M3 第七步：tsfront 支持基础 `import`/`export`，TS 模块路径自动映射为
  `.rune` 扩展名；`ts2rune` 产出的多文件工程可直接 `rune run`。
  含 `if (` 的源码在 `go/ts/mbt` 生成时走宿主编译器 bridge，避免自举 emitter 的
  无类型三元表达式问题。

2026-11-13 增量：

- D7 负例落地：`internal/tsfront` 新增稳定的 `RUNE-DYN-001…013` 与
  `RUNE-PROFILE-001` 诊断，覆盖 `eval`、`Function`、`with`、`Proxy`/`Reflect`、
  原型篡改、动态 `import()`、`arguments`、`globalThis[...]`、`debugger`、
  `require()`、`call`/`apply`、`declare`/`namespace`/generator/decorator。
  `tests/conformance/negative/*.ts` 与 `internal/conformance/negative_test.go`
  逐条断言错误码。D7 扫描在转换前执行，保证不会退化为泛化的「unsupported node」。
- 顺带修复 vendored TS parser 的三个阻塞问题：文件首行注释导致解析失败、
  `import(...)` 被当成静态 import 声明、关键字不能作为属性名/`require` 表达式。
- union 类型：AST 新增 `TypeUnion`，parser 支持 `A | B`（含泛型/函数类型成员），
  checker 支持联合解析、赋值兼容、`UnionOf` 去重展平与 `NullableOf` 合并。
- `typeof` 表达式与收窄：新增 `ast.TypeOfExpr` / `ir.TypeOfExpr`；
  `typeof x == "tag"`、`===`、`!=`、`!==` 与左右交换都能收窄 union；
  `if` 无 `else` 且 then 分支为 `Never` 时，后续语句自动应用取反收窄。
- 三端 union 行为：TS 原生 `A | B`；Go 用 `any` + `runeTypeof`/`runeAs*`
  检查转换；解释器动态执行；MoonBit 在 codegen 前以
  `MoonBit backend does not support union types yet` 明确拒绝（原生枚举待做）。
- `internal/astprint` 输出 `A | B`；`ts2rune` 新增 union/`typeof` 转换；
  新增 `tests/conformance/types/union_typeof.rn`，interpreter/go/ts 三端通过；
  `internal/tsfront` 新增 TS union + typeof → Go 实际运行测试。
- 过渡 bridge：含 ` | ` 或 `typeof ` 的源码在 `rune check` / `rune run` /
  `rune build` 走宿主参考实现，避免自举前端尚未实现的语法造成错误；
  这是临时方案，后续需在 selfhost parser/checker/interpreter 补齐。
- 文档：新增 `docs/language/ts-compat.md` 与 `docs/zh/language/ts-compat.md`，
  记录接受语法、类型映射、union 收窄、D7 错误码与往返级别。
- class 子集：`tsfront` 把 class 降为 Rune struct + `<Class>_new` 工厂 +
  `<Class>_<method>` 静态函数；支持字段、构造函数（仅 `this.field = value`）、
  实例方法与 static 方法，`new C(...)` / `C.m()` 自动重写。继承、访问器、
  static 字段、复杂构造函数以 `RUNE-PROFILE-002` 明确拒绝。
  新增 `tests/conformance/types/class_struct.rn`，四端通过；`tsfront` 测试
  完成 TS → Rune → Go 实际运行（输出 3 / 0）。
- L1 往返：新增 `tests/conformance/roundtrip/{union,class,control}.ts` 与
  `internal/conformance/roundtrip_test.go`，跑通 `ts -> rune -> 打印 -> 重解析
  -> ts -> node`，并断言 Rune printer 是不动点。
- 顺带修复 `astprint` 的存量问题：可变绑定应打印 `:=:`（原 `~=` 无法解析）、
  公开类型/枚举需打印 `+` 前缀（否则重解析后变成私有）、类型/枚举/方法此前
  完全没有输出。
- 能力诊断起步：`rune check --target go|mbt` 现在在 check 阶段对 DOM 专属全局
  （`document`/`window`/`navigator`/`localStorage`/`sessionStorage`/`history`/
  `location`）报 `RUNE-CAP-001`，`--target ts` 放行；目录级 check 暂未接线。
- `undefined` 值：新增 `UndefinedLiteral`，checker/IR/解释器/三端 codegen 与
  `typeof` 全线打通；Go 用 `runeUndefined` 哨兵、TS 原生 `undefined`、
  MoonBit 用 `None`。`typeof undefined == "undefined"` 已加入 union conformance。
- `try/catch/finally` 与 `throw`：使用符号语法——裸块后接 `!? { e => ... }`
  （带绑定 catch）、`?! { ... }`（无绑定 catch）、`-- { ... }`（finally），
  `!! value` 表示 throw；新增语句级 `TryStmt` / `ThrowStmt`；
  解释器用可捕获的 `throwSignal`，TS 原生 `try/catch`，Go 用闭包内
  `panic`/`recover` + `runeThrown`。Go 目前明确拒绝 try 内部的
  `return`/`break`/`continue`（闭包无法跳回外层函数）。
  新增 `tests/conformance/control/try_catch.rn`（interpreter/go/ts 通过）与
  `roundtrip/try*.ts`（含 return-in-try 的 TS 语义），MoonBit 显式拒绝。
- `rune test` 增加 host bridge：含 try/throw/union/typeof/undefined 的测试文件
  改走宿主参考解释器，避免自举解释器的不支持语法导致假失败。
- D7 收尾：`RUNE-DYN-008`（未标注参数/字段的隐式 any、未声明全局赋值）与
  `RUNE-DYN-009`（对 `any`/`unknown` 值做成员访问）已接入 tsfront；负例覆盖
  implicit_any / undeclared_global / any_member / any_member_param /
  class_field_implicit_any。
- 经典 `for(init; cond; update)`：`tsfront` 将其降为 init + `while`，并把
  `continue` 改写为「先执行 update 再 continue」，保持 TS 后置迭代语义。
  同时修复 Go/TS/MoonBit 的嵌套 BlockExpr 内联、循环体内 `return` 的类型化、
  Double `i++` 的 Go/MoonBit 生成，以及 vendored TS parser 的 for-var 分号问题。
  新增 `roundtrip/for_loop.ts` 与 tsfront → Go 实际运行测试；Rune 源仍无
  原生经典 for 语法。
- `do { } while (cond)`：`tsfront` 降为 `while (true)`，continue 改写为
  「先检查 cond，再 continue」，保证每次迭代只求值一次条件；新增
  `roundtrip/do_while.ts`，TS 端通过，Go/MoonBit 手工验证输出 1/3/1。
- 修复尾位置 `if` 含 `break/continue/return/throw` 时被错误降为三元表达式的问题；
  TS 一元 `!` 现在总是输出 `!(...)`，避免 `!i < 3` 的优先级错误。
- class 单继承：派生 struct 扁平化基类字段，未 override 的基类方法复制到派生
  类型；`Child_new` 调用一次 `Base_new` 后拷贝基类字段。支持 `super(...)`、
  override、继承方法调用；`super.method()` 降为 `__super_<method>` 别名调用，
  static 方法以 `Child_<method>` 继承；带初始化的 static 字段降为顶层常量。
  基类类型多态、可变 static 字段、访问器仍拒绝。
  新增 `roundtrip/class_inherit.ts` 与 tsfront → Go 实际运行测试。
- `console.log(...)` 内建别名：checker 将 Rune 源中的 `console.log` 归一化为
  `@io.println`，支持多参数空格分隔；新增 `es/console_log.rn` 四端 conformance。
  这是 M6 文档去 `@io.` 的第一步。
- `for-of` 改为下标式 `while` 降糖（可迭代对象只求值一次），
  `break`/`continue` 现在作用于真实循环；新增
  `control/for_of_break_continue.rn` 四端用例与 `roundtrip/for_of.ts`。
- class `get` 访问器：降为 `get_<property>()` 方法；当接收者类型静态可知时，
  `obj.prop` 改写为 `obj.get_prop()`。`set` 访问器因字段可变性尚未跨端落地而显式
  拒绝。新增 `roundtrip/class_getter.ts` 与 tsfront 测试。
- 循环与绑定语法校正：Rune 源码循环统一为 `* (cond) { ... }`，`^` / `^>` 表示
  `break` / `continue`；`while` / `break` / `continue` 关键字仅保留为兼容别名。
  `:=` 是不可变绑定，`:=:` 才是可变绑定；自举 formatter 已修复可变绑定被打印成
  `:=` 的问题。宿主 parser/astprint、自举 parser/formatter、conformance 与文档
  已同步，新增 `control/loop_symbols.rn` 覆盖 `^` / `^>` 的四端行为。

已知差距（后续项）：

- 自举 Go emitter（`selfhost/compiler/go.rn`）尚未按表达式类型做 Int32 收窄；
  目前回绕语义在宿主 Go codegen 与自举解释器中生效，`rune build` 走自举 emitter 时
  仍以宿主字长 `int` 运算。
- 自举 Go/TS/MoonBit emitter 尚未为 `if` 生成类型化分支，`go/ts/mbt` 目前通过
  `requiresHostCompilerBridge` 的 `if (` 检测回到宿主编译器；下一步是让自举 emitter
  直接消费 `If` 节点，去掉该 bridge。
- 自举 Go/TS/MoonBit emitter 也尚未实现 `Match`，因此 `switch (` 同样通过 bridge
  回到宿主编译器；TS `switch` 的 fallthrough 与 case 内 `break` 语义尚未实现。
- `for-of` 已改为下标式 `* (cond)` 循环降糖，支持 `^`/`^>`；仍只支持 Array，
  不支持 Set 等非 Array iterable。
- 自举 Go/TS/MoonBit emitter 对循环内 `^` / `^>` 的作用域与“循环内重新赋值”的
  可变性推导仍不完整；宿主 parser 已支持，含循环的源码在 `go/ts/mbt` 生成时继续走
  宿主编译器 bridge。
- `return` 在自举 emitter 中尚未处理“if 分支块内 return”的类型化作用域，
  因此包含 `return` 的源码在 `go/ts/mbt` 生成时走宿主编译器 bridge。
- `^` / `^>` 在自举 emitter 的嵌套分支中仍有作用域问题，代码生成走 host bridge。
- `tsfront` 已提供 `rune ts2rune` 与源码打印器，但仍只覆盖保守子集。
  缺少分号的单行 `return` / `^` / `^>` 仍会预处理补分号。
  try-catch、`for(;;)`、do-while、泛型、JSX 暂不支持；union、`typeof` 与
  class 基础子集已支持。
- class 继承为字段/方法扁平化，不支持基类类型多态；setter 访问器、可变 static
  字段、装饰器仍未实现。
- Go 后端的 try/catch/finally 是闭包 + panic/recover，内部 `return` / `^` / `^>`
  会明确报 codegen 错误；后续需要信号变量或语句级展开来补齐。
- `rune fmt` 仍走自举 formatter；try/catch 可被重新输出，但 `} catch` 的
  规范空白与自举 formatter 的既有输出尚未完全对齐。
- union 的 MoonBit 后端尚未落地原生 tagged union；当前为显式 codegen 错误。
- 收窄仅覆盖 `typeof` 判别；判别字段（discriminant）、`instanceof`、`in`
  与 `switch` 穷尽收窄尚未实现。
- `import` 目前只保留模块路径，不校验 named/default specifier；Rune 的文件
  导入会暴露目标文件所有 `+` 声明，没有 TS 的符号级导入语义。
- 自举解释器内部用 Int32 承载所有整数，`Int64` 超出 32 位的值暂时无法完整表示。
- `TestSelfhostCompilerTargetOutputsTypecheck` 需要较长超时（本机约 55–70 分钟），
  这是自举编译器解释执行的固有成本，不是断言失败。

验证（2026-11-13）：

- `go build ./...` 与 `git diff --check`：通过。
- `go test $(go list ./... | grep -v internal/selfhostrunner) -count=1 -timeout 20m`：全部通过。
- `go test ./internal/selfhostrunner -run 'TestRunSelfhostBootstrapTestsIR|TestRunFunctionIRReturnsStructuredValue'`：
  通过（约 14s）。
- `go test ./internal/selfhostrunner -run 'TestRunCoreModuleTestsIR|TestRunSelfhostBootstrapTestsIR'`：
  通过（约 516s）。
- `go test ./internal/selfhostrunner -run '^TestSelfhostCompilerTargetOutputsTypecheck$' -timeout 150m`：
  通过（约 3464s）。
