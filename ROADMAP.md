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
- [x] Use the generated self-hosted checker for host single-file checks, with
  compatibility tests for success, errors, output, and exit behavior.
- [ ] Move directory traversal, source import graphs, warning locations, and
  standalone CLI execution into self-hosted APIs.
- [ ] Replace the remaining Go directory/diagnostic bridge after compatibility
  is established.

### M2 — Self-hosted compilation commands

- [ ] Move `rune go` to the self-hosted compiler for single-file programs.
- [ ] Add source-file discovery and imports through `compile*Files`.
- [ ] Move `ts`, `dts`, and `mbt` with parity tests per backend.
- [ ] Preserve output-file handling and structured diagnostic locations in the
  Go host until those capabilities are self-hosted.

### M3 — Self-hosted execution and build orchestration

- [ ] Move backend selection and `rune run` orchestration.
- [ ] Move `rune build`, retaining Go only for native process/toolchain bridges
  where necessary.
- [ ] Move test discovery and `rune test` execution.
- [ ] Move formatting and REPL only after their self-hosted APIs are stable.

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

Finish M1 by adding generated-standalone CLI tests for `check` success and
failure, then extract a backend-neutral compiler `check` API and compare its
output and exit behavior with the existing Go host.
