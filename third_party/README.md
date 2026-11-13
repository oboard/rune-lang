# Vendored Dependencies

## `go-typescript-eslint`

Source: `github.com/kdy1/go-typescript-eslint` pinned at
`v0.0.0-20251023225804-e122a8dd8931`.

`go.mod` uses a `replace` directive so the TypeScript frontend can carry two
local fixes required by RFC 0001's `tsfront`:

1. `internal/parser/expressions.go` — object literal properties no longer try
   to parse their value as a TypeScript type annotation (`{ a: 1 }` now parses
   as an object expression).
2. `internal/parser/expressions.go` — identifiers followed by `<` no longer
   unconditionally attempt generic instantiation, so `a < b` parses as a
   comparison instead of failing with `expected GTR`.

3. `internal/lexer` / `internal/parser` — added template continuation scanning
   (`ScanTemplateContinuation` + parser resume) so `${...}` templates produce
   TemplateMiddle/Tail tokens instead of failing with
   `expected template middle or tail`.
4. `internal/parser/parser.go` — comments are consumed inside `nextToken` while
   still being collected. Previously the scanner returned COMMENT tokens but
   the grammar never skipped them, so a file starting with `//` failed to
   parse.
5. `internal/parser` — `import(...)` is parsed as an expression rather than a
   static import declaration, `require` is accepted as an expression
   identifier, and keyword tokens are accepted as property names
   (`Reflect.get`, `x.delete`). These keep RFC 0001's stable `RUNE-DYN-*`
   diagnostics reachable instead of failing at generic parse errors.
6. `internal/parser/functions.go` / `expressions.go` — class members may use
   keyword names (`constructor`), and `new C(args)` is normalized from the
   parser's eager `CallExpression` callee to a `NewExpression` with the
   constructor arguments attached.
7. `internal/parser/statements.go` — `throw` compares the argument line against
   the `throw` token line. The upstream check compared it against the following
   token, so every multi-token throw expression failed with a spurious
   "line break is not allowed" error.
8. `internal/parser/statements.go` — regular-for variable initializers consume
   their first semicolon, matching expression initializers; otherwise the test
   clause was skipped and `for (let i = 0; i < n; i++)` failed with
   `expected RPAREN, got SEMICOLON`.
