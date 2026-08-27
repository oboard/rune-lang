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
- [ ] Move directory traversal, source import graphs, warning locations, and
  standalone CLI execution into self-hosted APIs.
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
- [ ] Move `dts` with parity tests per backend.
- [ ] Preserve output-file handling and structured diagnostic locations in the
  Go host until those capabilities are self-hosted.

### M3 — Self-hosted execution and build orchestration

- [x] Use the self-hosted Go compiler for the default `rune run` compilation
  path; Go retains process launch and backend selection.
- [ ] Move `rune build`, retaining Go only for native process/toolchain bridges
  where necessary.
- [ ] Move test discovery and `rune test` execution.
- [x] Add a self-hosted formatter and use it through a byte-equivalence bridge
  for its supported comment-free subset.
- [ ] Add lossless trivia and AST-printing coverage before replacing the Go
  formatter fallback, then move REPL.

### M4 — Host minimization

- [ ] Reduce `cmd/rune` to bootstrap, artifact validation, and platform bridges.
- [ ] Document supported fallback behavior and remove duplicated command parsing.
- [ ] Establish a release build that validates generated artifacts from a clean
  checkout.

### M5 — LSP migration (separate track)

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
