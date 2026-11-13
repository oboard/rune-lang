# Rune Conformance Fixtures

These fixtures are the executable form of RFC 0001's conformance matrix. Each
`.rn` file declares its expected stdout and the targets it covers:

```rune
// expect: 7
// targets: interpreter,go
main() => @io.println(true ? 7 : 9)
```

- `expect:` may appear multiple times; lines are concatenated in order and each
  contributes a trailing newline.
- `targets:` defaults to `interpreter,go`. Available: `interpreter`, `go`, `ts`
  (node), `mbt` (moon).
- Fixtures with no `expect:` must still produce empty stdout.

Current directories:

| Directory | Scope |
| --- | --- |
| `es/` | ES built-in semantics (Int32, strings, arrays, JSON) |
| `control/` | Control flow and expression evaluation |
| `types/` | Records, enums, optionals, union/typeof, class lowering |
| `negative/` | `.ts` programs rejected with stable `RUNE-DYN-*` / `RUNE-PROFILE-*` codes |
| `roundtrip/` | `ts -> rune -> ts` L1 semantic checks |

The harness lives in `internal/conformance`. It runs:

- `interpreter`: host IR interpreter
- `go`: host Go codegen, then `go run`
- `ts`: host TypeScript codegen, then `node`
- `mbt`: host MoonBit codegen, then `moon run`

MoonBit codegen rejects union types explicitly until the native tagged-union
lowering lands, so union fixtures list only `interpreter,go,ts`.
