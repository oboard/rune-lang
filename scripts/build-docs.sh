#!/usr/bin/env bash
set -euo pipefail

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_dir"

ensure_go() {
  if [[ "${GO_FORCE_INSTALL:-0}" != "1" ]] && command -v go >/dev/null 2>&1; then
    go version
    return
  fi

  local go_version="${GO_VERSION:-}"
  if [[ -z "$go_version" ]]; then
    go_version="$(awk '$1 == "go" { print $2; exit }' go.mod)"
  fi
  go_version="${go_version#go}"
  if [[ -z "$go_version" ]]; then
    echo "Could not determine Go version from GO_VERSION or go.mod." >&2
    exit 1
  fi

  local os arch
  case "$(uname -s)" in
    Linux) os="linux" ;;
    Darwin) os="darwin" ;;
    *)
      echo "Unsupported OS for automatic Go install: $(uname -s)" >&2
      exit 1
      ;;
  esac

  case "$(uname -m)" in
    x86_64 | amd64) arch="amd64" ;;
    arm64 | aarch64) arch="arm64" ;;
    *)
      echo "Unsupported architecture for automatic Go install: $(uname -m)" >&2
      exit 1
      ;;
  esac

  local cache_base="${VERCEL_CACHE_DIR:-}"
  if [[ -z "$cache_base" ]]; then
    cache_base="${HOME:-$repo_dir}/.cache/rune-lang"
  fi

  local install_root="${GO_INSTALL_ROOT:-$cache_base/go}"
  local go_root="$install_root/go${go_version}.${os}-${arch}"
  local archive_name="go${go_version}.${os}-${arch}.tar.gz"
  local download_url="${GO_DOWNLOAD_URL:-https://go.dev/dl/$archive_name}"

  if [[ ! -x "$go_root/bin/go" ]]; then
    local tmp_dir
    tmp_dir="$(mktemp -d)"
    trap 'rm -rf "$tmp_dir"' EXIT

    mkdir -p "$install_root"
    echo "Installing Go $go_version for $os/$arch..."
    if command -v curl >/dev/null 2>&1; then
      curl -fsSL "$download_url" -o "$tmp_dir/$archive_name"
    elif command -v wget >/dev/null 2>&1; then
      wget -qO "$tmp_dir/$archive_name" "$download_url"
    else
      echo "curl or wget is required to install Go." >&2
      exit 1
    fi

    rm -rf "$go_root"
    mkdir -p "$go_root"
    tar -C "$go_root" --strip-components=1 -xzf "$tmp_dir/$archive_name"
  fi

  export GOROOT="$go_root"
  export PATH="$GOROOT/bin:$PATH"
  go version
}

ensure_go
exec vitepress build docs
