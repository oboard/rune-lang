#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="${RUNE_BIN_DIR:-"$ROOT/.bin"}"
HOST_PATH="$PATH"

RUN_TESTS=1
RUN_EXAMPLE=1
INSTALL_VSCODE_DEPS=0
INSTALL_VSCODE_EXTENSION=0
START_LSP=0
OPEN_SHELL=0

usage() {
  cat <<'EOF'
Usage: scripts/dev.sh [options]

Prepare a local Rune development environment.

Options:
  --no-test       Skip go test ./...
  --no-example    Skip examples/fib.rn check/run smoke test
  --vscode        Install VSCode extension for local development
  --lsp           Start rune lsp after preparing the environment
  --shell         Open a shell with local rune on PATH
  -h, --help      Show this help

Environment:
  RUNE_BIN_DIR    Where the local rune binary is built. Default: ./.bin

Examples:
  scripts/dev.sh
  scripts/dev.sh --shell
  scripts/dev.sh --vscode
  scripts/dev.sh --lsp
EOF
}

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

shell_quote() {
  printf "%q" "$1"
}

open_dev_shell() {
  local shell_path shell_name tmp rc zshenv
  shell_path="${SHELL:-/bin/sh}"
  shell_name="$(basename "$shell_path")"

  case "$shell_name" in
    zsh)
      tmp="$(mktemp -d "${TMPDIR:-/tmp}/rune-dev-zsh.XXXXXX")"
      zshenv="$tmp/.zshenv"
      rc="$tmp/.zshrc"
      cat >"$zshenv" <<EOF
if [ -f "\$HOME/.zshenv" ]; then
  source "\$HOME/.zshenv"
fi
export ZDOTDIR=$(shell_quote "$tmp")
EOF
      cat >"$rc" <<EOF
if [ -f "\$HOME/.zshrc" ]; then
  source "\$HOME/.zshrc"
fi
export RUNE_ROOT=$(shell_quote "$ROOT")
export PATH=$(shell_quote "$BIN_DIR:$HOST_PATH"):\$PATH
hash -r 2>/dev/null || true
trap 'rm -rf $(shell_quote "$tmp")' EXIT
EOF
      exec env ZDOTDIR="$tmp" "$shell_path" -i
      ;;
    bash)
      tmp="$(mktemp -d "${TMPDIR:-/tmp}/rune-dev-bash.XXXXXX")"
      rc="$tmp/bashrc"
      cat >"$rc" <<EOF
if [ -f "\$HOME/.bashrc" ]; then
  source "\$HOME/.bashrc"
fi
export RUNE_ROOT=$(shell_quote "$ROOT")
export PATH=$(shell_quote "$BIN_DIR:$HOST_PATH"):\$PATH
hash -r 2>/dev/null || true
trap 'rm -rf $(shell_quote "$tmp")' EXIT
EOF
      exec "$shell_path" --rcfile "$rc" -i
      ;;
    *)
      export RUNE_ROOT="$ROOT"
      export PATH="$BIN_DIR:$HOST_PATH:$PATH"
      exec "$shell_path"
      ;;
  esac
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --no-test)
      RUN_TESTS=0
      ;;
    --no-example)
      RUN_EXAMPLE=0
      ;;
    --vscode)
      INSTALL_VSCODE_DEPS=1
      INSTALL_VSCODE_EXTENSION=1
      ;;
    --lsp)
      START_LSP=1
      ;;
    --shell)
      OPEN_SHELL=1
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown option: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
  shift
done

cd "$ROOT"

need go
mkdir -p "$BIN_DIR"

echo "==> Rune dev root: $ROOT"
echo "==> Building local rune CLI"
go mod download
go build -o "$BIN_DIR/rune" ./cmd/rune

export RUNE_ROOT="$ROOT"
export PATH="$BIN_DIR:$PATH"

if [ "$RUN_TESTS" -eq 1 ]; then
  echo "==> Running Go tests"
  go test ./...
fi

if [ "$RUN_EXAMPLE" -eq 1 ]; then
  echo "==> Checking Fibonacci example"
  rune check examples/fib.rn

  echo "==> Running Fibonacci example"
  output="$(rune run examples/fib.rn)"
  echo "$output"
  if [ "$output" != "55" ]; then
    echo "unexpected examples/fib.rn output: $output" >&2
    exit 1
  fi
fi

if command -v node >/dev/null 2>&1; then
  echo "==> Validating tree-sitter grammar JavaScript"
  node -c tree-sitter-rune/grammar.js
else
  echo "==> Skipping tree-sitter grammar validation; node not found"
fi

if [ "$INSTALL_VSCODE_DEPS" -eq 1 ]; then
  need npm
  echo "==> Installing VSCode extension dependencies"
  npm install --prefix vscode-rune
fi

if [ "$INSTALL_VSCODE_EXTENSION" -eq 1 ]; then
  if command -v code >/dev/null 2>&1; then
    VSCODE_CLI="code"
  elif command -v codium >/dev/null 2>&1; then
    VSCODE_CLI="codium"
  elif command -v cursor >/dev/null 2>&1; then
    VSCODE_CLI="cursor"
  else
    echo "missing VSCode CLI: install the 'code' command from VSCode first" >&2
    exit 1
  fi

  vsix="$BIN_DIR/vscode-rune.vsix"
  echo "==> Packaging VSCode extension"
  (
    cd vscode-rune
    npm exec -- vsce package --allow-missing-repository --out "$vsix"
  )

  echo "==> Installing VSCode extension with $VSCODE_CLI"
  "$VSCODE_CLI" --install-extension "$vsix" --force
fi

echo
echo "Rune development environment is ready."
echo "Local rune: $(command -v rune)"
echo "Try: rune run examples/fib.rn"
echo "LSP: rune lsp"
if [ "$INSTALL_VSCODE_EXTENSION" -eq 1 ]; then
  echo "VSCode: reload the window after installing/updating the Rune extension"
fi
echo

if [ "$START_LSP" -eq 1 ]; then
  echo "==> Starting rune lsp"
  exec rune lsp
fi

if [ "$OPEN_SHELL" -eq 1 ]; then
  echo "==> Opening shell with local rune on PATH"
  open_dev_shell
fi
