# Roundtrip Fixtures

Each `.ts` fixture is an L1 (semantic) round-trip case. The harness in
`internal/conformance/roundtrip_test.go` runs:

1. `ts -> rune` (`internal/tsfront`) and prints the Rune source.
2. Re-parses the printed Rune and asserts the printer is a fixed point.
3. `rune -> ts` (checker + IR + TypeScript backend), runs it with `node`, and
   compares stdout with the `// expect:` lines.

Current fixtures cover union/`typeof`, the class lowering subset, and
if/while/break control flow. L2 (structural) and L3 (textual) fidelity are not
asserted yet.
