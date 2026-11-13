# TypeScript Compatibility Profile

Rune is being repositioned as a statically-checkable TypeScript subset plus
Go/MoonBit backends (RFC 0001). This page is the maintained list of what the
profile currently accepts, how types map across backends, and which constructs
are rejected on purpose.

## Accepted syntax

| TypeScript | Rune | Status |
| --- | --- | --- |
| `function f(a: T): R` | `f(a: T) -> R => ...` | `ts2rune` |
| `const x = v` / `let x = v` | `x := v` / `x :=: v` | immutable by default; `:=:` is mutable |
| `if / else if / else` | `cond ? { ... } : { ... }` | lowered to ternary blocks |
| `switch` | pattern block | TS switch not converted yet; original `1 => ...` syntax remains |
| `for (const x of xs)` | same | arrays; `^` / `^>` supported |
| `while` | `* (cond) { ... }` | supported |
| `return` / `break` / `continue` | `return` / `^` / `^>` | supported |
| `A \| B` type annotations | same | supported |
| `typeof x === "string"` | same | supported; narrows unions |
| arrow functions | same | supported as callbacks |
| template strings | same | supported |
| static `import` / `export` | `@"./m.rune"` | supported |
| classes (fields + constructor + methods + static) | struct + `Class_new` | supported subset |
| class inheritance (`extends`, `super(...)`, overrides) | flattened fields/methods | supported subset |
| static fields with initializer | top-level `Class_field` constants | supported |
| `get` accessors | rewritten to `get_<property>()` methods | supported |
| `set` accessors, mutable static fields | — | rejected with `RUNE-PROFILE-002` |
| `try/catch/finally`, `throw` | `{ } !? { e => } -- { }`, `!! value` | symbol form; Go rejects control flow inside |
| `for(init; cond; update)` | lowers to `* (cond)` | supported by `ts2rune` |
| `do { } while (cond)` | lowers to `* (true)` + `^` | supported by `ts2rune` |

### Class lowering

A TypeScript class is lowered to a Rune struct plus a factory function:

```ts
class Point {
  x: number
  constructor(x: number) { this.x = x }
  sum(): number { return this.x + this.x }
  static origin(): Point { return new Point(0) }
}
```

becomes `Point: { x: Double; sum() -> Double => ... }`, `Point_new(x: Double)
-> Point`, and `Point_origin() -> Point`. `new Point(...)` and `Point.origin()`
are rewritten to the factory/free function. The constructor body may only
contain `this.field = value` assignments (and, in a derived class, a leading
`super(...)` call); anything else is rejected rather than approximated.

Single inheritance is flattened: the derived struct contains base fields, and
base methods are copied unless overridden. `Child_new` calls `Base_new` once
and copies base fields into the derived struct. `super.method()` calls a
generated `__super_<method>` alias on the derived type; static methods are
inherited as `Child_<method>` wrappers. This is sound for direct `new Child()`
usage. Getter accessors are rewritten to `get_<property>()` calls when the
receiver's class is statically known (`this`, a typed parameter, or a `new C()`
local). Base-typed polymorphism, setter accessors, mutable static fields, and
decorators are not fully modelled: static fields with initializers become
top-level constants (reads only), while setters, mutable static fields, and
decorators are rejected with `RUNE-PROFILE-002`.

### try / catch / finally

`throw` is written `!! value`. A try block is a bare block followed by symbol
handlers: `!? { err => ... }` catches with a binding, `?! { ... }` catches
without a binding, and `-- { ... }` is finally.

```rune
{
  !! "boom"
} !? {
  err => @io.println(err)
} -- {
  @io.println("done")
}
```

Interpreter and TypeScript backend implement full statement semantics. The Go
backend uses `panic`/`recover` inside a closure and therefore rejects
`return` / `^` / `^>` **inside** try/catch/finally (a clear codegen error)
instead of silently changing their target. MoonBit currently rejects the
construct explicitly (`try/catch/throw yet`).

`rune test` detects these constructs and uses the host interpreter, so test
files do not depend on the self-hosted interpreter's feature set.

### Classic for loops

`ts2rune` lowers `for (init; cond; update)` to the init statement plus a
`* (cond)` loop. `^>` is rewritten to run `update` before jumping, so
post-iteration semantics match TypeScript:

```ts
for (let i = 0; i < n; i++) { ... continue ... }
```

becomes a `* (cond)` loop with `i++` duplicated at the `^>` site and at the end
of the body. The resulting Rune is accepted by every backend. Rune source itself
has no native classic-for syntax.

`do { body } while (cond)` lowers to an internal loop that checks `cond` at the
end of each iteration. A `^>` inside the body runs the same check before
jumping, so each iteration evaluates `cond` exactly once, including when the
condition has side effects.

### Built-in aliases

`console.log(...)` is accepted in Rune source and maps to `@io.println` on all
backends. Multiple arguments are printed space-separated on one line. This is
the migration path away from `@io.` in examples and docs.

`for-of` is lowered to an index-based loop over the array (evaluating the
iterable once), so `^`/`^>` affect the real loop instead of a callback. `^>`
advances the index before jumping.

## Type mapping

| Rune | TypeScript | Go | MoonBit |
| --- | --- | --- | --- |
| `Int32` | `number` | `int` with 32-bit wrap | `Int32` |
| `Int64` | `bigint` | `int64` | `Int64` |
| `Double` | `number` | `float64` | `Double` |
| `String` | `string` | `string` | `String` |
| `Bool` | `boolean` | `bool` | `Bool` |
| `Null` | `null` | `any(nil)` | `Unit` |
| `Undefined` | `undefined` | `any` (nil) | `Unit` |
| `Array[T]` | `T[]` | `[]T` | `Array[T]` |
| `A \| B` | native `A \| B` | `any` + checked narrowing | rejected (native enum pending) |

`Int32`/`Double` both become `number` in TypeScript. Converting back cannot
distinguish them, so a round trip may turn `Int32` into `Double`; this is an
accepted L1 limitation.

## Union narrowing

`typeof value` returns `"string" | "number" | "boolean" | "object" |
"function" | "undefined"`. Guards of the form `typeof x == "tag"`,
`typeof x === "tag"`, `typeof x != "tag"`, and the reversed operand order
narrow `x` inside the guarded branch. Early returns propagate the negated guard
to the statements after the `if`:

```rune
describe(value: String | Double) -> String => (typeof value == "string") ? {
  value
} : {
  "number"
}
```

`undefined` is a first-class value/literal, distinct from `null` for `typeof`;
`T?` accepts both `null` and `undefined`.

The Go backend represents a union as `any` and emits checked numeric/string
conversions at narrowed uses. MoonBit currently rejects union types at codegen
time (`union types yet`); the planned lowering is a native enum or tagged union.

## Rejected dynamic features (D7)

The `ts2rune` frontend rejects these with stable `RUNE-DYN-*` codes. The
fixtures live in `tests/conformance/negative/`.

| Code | Construct |
| --- | --- |
| `RUNE-DYN-001` | `eval` |
| `RUNE-DYN-002` | `Function(...)` / `new Function(...)` |
| `RUNE-DYN-003` | `with` |
| `RUNE-DYN-004` | `Proxy`, `Reflect` traps |
| `RUNE-DYN-005` | `__proto__`, `Object.setPrototypeOf`, `defineProperty` |
| `RUNE-DYN-006` | dynamic `import()` |
| `RUNE-DYN-007` | `arguments` |
| `RUNE-DYN-008` | implicit `any` parameters/fields, undeclared global writes |
| `RUNE-DYN-009` | member access on `any` / `unknown` values |
| `RUNE-DYN-010` | `globalThis[name]`, `window[name]` |
| `RUNE-DYN-011` | `debugger` |
| `RUNE-DYN-012` | `require()` |
| `RUNE-DYN-013` | `call` / `apply` this-rebinding |
| `RUNE-PROFILE-001` | `declare`, `namespace`, generators, decorators |
| `RUNE-PROFILE-002` | unsupported class feature (inheritance, accessors, static fields, complex constructors) |

Static `import ... from` remains the only module mechanism. `JSON.parse`
results are not given a structural type automatically; annotate and narrow them
explicitly.

## Capability checks

`rune check --target go|mbt file.rn` rejects DOM-only globals with
`RUNE-CAP-001` before code generation. `--target ts` allows them. Current
coverage: `document`, `window`, `navigator`, `localStorage`, `sessionStorage`,
`history`, `location`. Directory-level checks currently do not accept
`--target`.

## Round-trip levels

- **L1 (semantic)**: `ts -> rune -> go` and `ts -> rune -> ts` are covered by
  `internal/tsfront` tests and conformance fixtures.
- **L2 (structural)**: not guaranteed yet. Rune-only sugar (structs, signals,
  pattern bodies) is lowered, not preserved.
- **L3 (textual)**: out of scope for now.

## Transitional bridges

The self-hosted frontend/interpreter does not parse union types or `typeof`
yet. Sources that use them are routed to the host reference implementation for
`rune check`, `rune run`, and `rune build`; self-hosted behavior is unchanged
for the rest of the language. This bridge is temporary and tracked in RFC 0001
Appendix C.
