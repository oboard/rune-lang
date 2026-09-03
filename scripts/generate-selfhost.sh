#!/usr/bin/env bash
set -euo pipefail

# Regenerates the self-hosted bootstrap artifacts checked into cmd/rune/.
#
#   selfhost_compiler_gen.go  <- rune go ./selfhost/compiler/compiler.rn
#   selfhost_cli_gen.go       <- rune go ./selfhost/cli/cli.rn
#
# This is self-bootstrapping: the `rune go` subcommand uses the currently built
# .bin/rune (host compiler/selfhost bridge) to compile the Rune bootstrap source
# to a Go emitter. Run it after changing files under selfhost/ to keep the
# checked-in generated Go in sync with the Rune sources.

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="${RUNE_BIN_DIR:-"$ROOT/.bin"}"

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

need go
need gofmt

mkdir -p "$BIN_DIR"

echo "==> Building host rune CLI"
go mod download
go build -o "$BIN_DIR/rune" ./cmd/rune

gen() {
  local entry="$1" out="$2" strip="$3" tmp
  tmp="$(mktemp)"
  echo "==> generating $out from $entry"
  "$BIN_DIR/rune" go "$entry" >"$tmp"
  if [ -n "$strip" ]; then
    # The host cmd/rune/main.go provides `main` and drives the CLI through
    # __parseCli directly, so drop the emitted `func main() { __main() }`
    # wrapper that rune go appends when the bootstrap source has a main().
    sed '/^func main() {$/,$d' "$tmp" > "$tmp.stripped"
    mv "$tmp.stripped" "$tmp"
  fi
  gofmt -w "$tmp"
  mv "$tmp" "$ROOT/$out"
}

gen selfhost/compiler/compiler.rn cmd/rune/selfhost_compiler_gen.go ""
gen selfhost/cli/cli.rn cmd/rune/selfhost_cli_gen.go "strip"

echo "==> rebuilding rune with generated artifacts"
go build -o "$BIN_DIR/rune" ./cmd/rune

echo
echo "Generated bootstrap artifacts are up to date:"
echo "  cmd/rune/selfhost_compiler_gen.go"
echo "  cmd/rune/selfhost_cli_gen.go"