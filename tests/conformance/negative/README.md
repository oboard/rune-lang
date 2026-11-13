# Negative Fixtures

Each `.ts` fixture declares the stable diagnostic it expects with a
`// code: RUNE-...` line. `internal/conformance/negative_test.go` converts the
file through `internal/tsfront` and fails if the documented code is missing.

Covered today: `eval`, `Function` construction, `with`, `Proxy`/`Reflect`,
runtime prototype mutation, dynamic `import()`, `arguments`, `call`/`apply`
this-rebinding, `globalThis[...]`, `debugger`, `require()`,
`declare`/`namespace` profile escapes, implicit-`any` parameters/fields and
undeclared global writes (`RUNE-DYN-008`), and member access on `any`/`unknown`
values (`RUNE-DYN-009`).
