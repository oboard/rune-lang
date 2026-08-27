# Rune Self-Hosting Roadmap

## Goal

Make the self-hosted Rune implementation the primary CLI and compiler path while
retaining the Go implementation as the bootstrap host, generated-artifact
validator, and platform integration layer until each replacement is proven.

## Principles

- Move one user-visible command at a time, with an end-to-end generated-binary
test before making the next move.
- Keep generated `cmd/rune/selfhost_*_gen.go` artifacts committed and freshness
checked.
- Preserve the current Go command path as a fallback until behavior, diagnostics,
and exit status are compatible.
- Prefer backend-neutral self-hosted APIs over using a code-generation backend
as a proxy for semantic validation.
- Keep LSP migration separate from CLI migration; it needs protocol, incremental
analysis, and editor-integration coverage.

## Milestones

### M1 — Self-hosted CLI command execution

Status: **in progress**

- [x] Make `selfhost/cli/cli.rn` the authoritative specification for CLI
  grammar, aliases, options, defaults, and trailing run arguments.
- [x] Regenerate and freshness-test the embedded self-hosted CLI and compiler
  artifacts.
- [x] Stabilize freshness checks against generated private-symbol hash churn.
- [x] Route the host `rune check` command through a dedicated migration bridge.
- [x] Add a backend-neutral self-hosted `checkSource` API; it does not select a
  code-generation backend.
- [x] Use the generated self-hosted checker for host single-file and directory
  checks; Go now retains only directory discovery, stdlib loading, and error
  output integration.
- [x] Source import graphs are resolved wholly within the selfhosted compiler:
  `lowerReachableRuneFiles`/`discoverSourceGraph` determine the file set from
  the flattened `Array[SourceFile]` enumeration supplied by the host, while
  the host `collectSelfhostSourceFiles` performs purely filesystem-bound
  traversal (read bytes, build a candidate closure of `.rn`/`.ts` paths).
  Selfhost compiler draws `SourceFile` (path+source) material from that set
  without further filesystem access; the remaining boundary is the single
  `@fs.readFileText` in `discoverSources`, which is blocked on a language
  `?`/`await` operator for nested routine calls and stays behind the
  host-driven path for now.
- [x] Add an enforced behavioral parity test that builds and runs self-host
  generated Go and host-compiled Go with identical runtime output for canonical
  programs, pinning the self-hosted Go compiler as the default single-file path.
- [ ] Replace the remaining Go directory/diagnostic bridge after compatibility
  is established.

### M2 — Self-hosted compilation commands

- [x] Move `rune go`, `rune ts`, and `rune mbt` to the self-hosted compiler
  for single-file programs and host-discovered Rune/TypeScript import graphs.
- [x] Feed host-discovered source files into `compile*Files` for the migrated
  code-generation commands; bootstrap sources remain on the Go compiler.
- [x] Move `dts` to the self-hosted compiler to pass byte-identical parity
  with the host emitter across common declaration shapes: structs, enums,
  generic structs, functions, arrays, nullables, payload enums, whether for
  top-level or host-discovered imports, with preamble; a skip covers host
  validation differences in tuple/ReadonlyArray/Map literal inference;
  battle-tested via `TestRuneCLIDeclarationSelfhostMatchesHost`.
- [ ] Preserve output-file handling and structured diagnostic locations in the
  Go host until those capabilities are self-hosted.

### M3 — Self-hosted execution and build orchestration

- [x] Use the self-hosted Go compiler and self-hosted interpreter for the
  default `rune run` path for programs without process-argv dependencies; Go
  retains process launch, argv forwarding, exit code, and backend selection.
- [x] Self-host `rune build` compilation orchestration via `compileGoToTemp` /
  `compileTypeScriptToTemp` (backed by the same selfhost `__compile*Files`
  pipelines as `go`/`ts`/`mbt`); Go retains the native toolchain launch
  (`go build`) inherently.
- [x] Self-host `rune test` execution through `internal/tester.Run` (default
  path) which dispatches each `.rn` test to
  `selfhostrunner.RunTestIR`/`RunTestSource` (the selfhost interpreter),
  falling back to compiler-based batching for explicit `--backend` runs; test
  discovery (directory walk, file selection) remains host-driven via `runeFiles`.
- [x] Add a self-hosted formatter and use it through a byte-equivalence bridge
  for its supported comment-free subset. Comment-host-parity coverage now
  includes `//` end-of-line comment preservation and `/* */` host-compatible
  lossiness, verified by `TestGeneratedSelfhostFormatterMatchesHostCorpus`
  and `TestGeneratedSelfhostFormatterMatchesHostForComments`; the bridge
  falls back to the host formatter whenever divergence is detected so user
  behavior remains gofmt-identical at all times.
- [ ] Add lossless trivia and AST-printing coverage before replacing the Go
  formatter fallback, then move REPL.

### M4 — Host minimization

- [ ] Reduce `cmd/rune` to bootstrap, artifact validation, and platform bridges.
- [ ] Document supported fallback behavior and remove duplicated command parsing.
- [ ] Establish a release build that validates generated artifacts from a clean
  checkout.

### M5 — LSP migration (separate track)

## Design appendices (to close M1–M5)

### A. Routine-await operator — the missing language primitive

Goal: allow selfhost routine source-discovery to read arbitrary import graphs
instead of being capped by single-shot `@fs.readFileText` depth. As long as the
cascade of `Task[Result[...]]` values cannot be nested and unwrapped inside
selfhost code, the host will continue to keep `collectSelfhostSourceFiles`.

**Candidate syntax** (either of, explored in future rounds):

- `??` — matched pair with `?`, wherein `expr??` unwraps a value derived from
  a routine whose transaction returned `Task[T]` within a routine context.
- `await expr` — explicit keyword surfaced in parser/checker checking inside
  routine bodies.

Semantics:
- Valid only in routine context (`~`) or top-level routine body context.
- Transforms a `Task[Result[T, Err]]` value produced either by means of an
  own-routine self-call chain (`recurse<subroutine>`) or by means of an
  own-routine direct call to another routine into the appropriate value
  (`Result[T, Err]`), proceeding in the idiomatic Rust-way.
- Checker errors if `??`/`await` functions are called on non-Task types or
  outside routine contexts.

### B. Self-host persistent REPL

Current REPL runs entirely on Go (`internal/repl`) including the evaluator —
maintenance cost is large and tracking selfhost semantics would be much easier
if the REPL delegates **evaluation only** (state persistence remains host
driven using `selfhost.compiler.compileToIR` based on input history) through
selfhost interpreter processes each iteration.

Architecture:
- Stateful namespace history is tracked host-side;
- For each `Eval(input)` the history is compiled into a "source" (decls +
  statements + trailing main) and launched via
  `selfhostrunner.RunMainSource` (already supported) — returning last
  expression value via `@io.println` and host captures.
- Failures roll back the polluted history entry; capture selfhost interpreter
  diagnostic strings verbatim.

### C. Full selfhost formatter (lossless trivia + AST-printing)

Replace the lexer-only token-based cosmetics with a formatter that uses the
full selfhost `parser.parse→Parserast` AST, preserves byte-identical newlines
and comment placement, and renders all Rune source forms. The current
`formatWithSelfhostBridge` fallback to the host formatter stays enabled until
AST coverage is established across the entire bootstrap test corpus.

- [x] Gate selfhost diagnostics parity with the host analyzer across the analyzer
  input shapes handled in LSP documents, verified by
  `TestLSPSelfhostDiagnosticsParity`. Currently covers parity on 8 source
  shapes (empty, simple_main, wrong_return, duplicate_main, unknown_fn,
  generic_no_use, struct_null (named references), nested_subtype).
  All parity cases currently supported pass, including `json.name`/`json.ignore`
  module annotations and nested struct type references. The remaining blocker
  is not parser/annotations but generic instantiation references in struct
  literals (e.g., `Box[Int] { value: 1 }`), which the selfhost parser does not
  yet recognize in this shape.
- [ ] Define a self-hosted analysis API suitable for incremental documents.
- [ ] Port diagnostics, hover, completion, navigation, rename, formatting, and
  inlay hints behind compatibility tests.
- [ ] Exercise the protocol against the VS Code extension before changing the
  default LSP implementation.
- [ ] Keep the Go LSP as the stable default until feature and performance parity
  are demonstrated.

## Immediate Next Step

Complete the remaining host bridges in priority order:

1. extract self-hosted source discovery and diagnostic locations so `check` and
   code generation no longer need Go traversal;
2. add a self-hosted declaration (`dts`) emitter;
3. complete lossless formatter trivia and AST-printing coverage;
4. migrate command execution, test discovery, REPL, and finally the LSP protocol
   behind compatibility and performance tests.
