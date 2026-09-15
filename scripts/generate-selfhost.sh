#!/usr/bin/env bash
set -euo pipefail

# Regenerates the self-hosted bootstrap artifacts checked into cmd/rune/.
#
#   internal/selfhost/generated.go <- rune go ./selfhost/compiler/compiler.rn
#                                    (includes the formatter and compiler)
#   selfhost_cli_gen.go       <- rune go ./selfhost/cli/cli.rn
#   internal/selfhostformat/generated.go <- rune go ./selfhost/format/format.rn
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

echo "==> Regenerating ambient DOM event stub from lib.dom.d.ts"
go run ./internal/tsdecl/cmd/gen-dom-stub \
  -lib "/Applications/Visual Studio Code.app/Contents/Resources/app/extensions/node_modules/typescript/lib/lib.dom.d.ts" \
  -out "$ROOT/selfhost/compiler/dom_types.rn"

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

gen selfhost/compiler/compiler.rn internal/selfhost/generated.go ""
sed -i.bak '1s/package main/package selfhost/' "$ROOT/internal/selfhost/generated.go"
rm -f "$ROOT/internal/selfhost/generated.go.bak"
cat >> "$ROOT/internal/selfhost/generated.go" <<'EOF'

// SourceFileInput is a source file consumed by the generated self-host compiler.
type SourceFileInput struct { Path string; Source string }
// CompileOutput is the transport-safe result of generated self-host compilation.
type CompileOutput struct { Ok bool; Output string; Errors []string }
func toSourceFiles(inputs []SourceFileInput) []SourceFile { out := make([]SourceFile, len(inputs)); for i, input := range inputs { out[i] = SourceFile{path: input.Path, source: input.Source} }; return out }
func toCompileOutput(result CompileResult) CompileOutput { return CompileOutput{Ok: result.ok, Output: result.output, Errors: result.errors} }
func FormatSource(source string) string { return formatSource(source) }
func CompileGo(source string) CompileOutput { return toCompileOutput(compileGo(source)) }
func CompileGoFiles(files []SourceFileInput) CompileOutput { return toCompileOutput(compileGoFiles(toSourceFiles(files))) }
func CompileTypeScriptFiles(files []SourceFileInput) CompileOutput { return toCompileOutput(compileTypeScriptFiles(toSourceFiles(files))) }
func CompileMoonBitFiles(files []SourceFileInput) CompileOutput { return toCompileOutput(compileMoonBitFiles(toSourceFiles(files))) }
func CompileDeclarations(source string) CompileOutput { return toCompileOutput(compileDeclarations(source)) }
func CompileDeclarationsFiles(files []SourceFileInput) CompileOutput { return toCompileOutput(compileDeclarationsFiles(toSourceFiles(files))) }
func CheckSourceWithCore(source, sourcePath string, files []SourceFileInput) CompileOutput { return toCompileOutput(checkSourceWithCore(source, sourcePath, toSourceFiles(files))) }
EOF
gofmt -w "$ROOT/internal/selfhost/generated.go"
gen selfhost/cli/cli.rn cmd/rune/selfhost_cli_gen.go "strip"

echo "==> rebuilding rune with generated artifacts"
go build -o "$BIN_DIR/rune" ./cmd/rune

echo
echo "Generated bootstrap artifacts are up to date:"
echo "  internal/selfhost/generated.go (includes formatter and compiler)"
echo "  cmd/rune/selfhost_cli_gen.go"
