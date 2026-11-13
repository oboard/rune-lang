#!/usr/bin/env bash
set -euo pipefail

# Copies a TypeScript lib set into tools/ts-lib, preserving the pinned VERSION
# file. Usage: scripts/sync-ts-lib.sh /path/to/typescript/lib

if [[ $# -ne 1 ]]; then
  echo "usage: $0 /path/to/typescript/lib" >&2
  exit 1
fi

src="$1"
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
dst="$root/tools/ts-lib"

if [[ ! -f "$src/lib.es5.d.ts" ]]; then
  echo "missing $src/lib.es5.d.ts" >&2
  exit 1
fi

mkdir -p "$dst"
for file in lib.es5.d.ts lib.es2015.d.ts lib.dom.d.ts; do
  if [[ -f "$src/$file" ]]; then
    cp "$src/$file" "$dst/$file"
    echo "copied $file"
  fi
done

if [[ -f "$src/../package.json" ]]; then
  version="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["version"])' "$src/../package.json")"
  printf '%s\n' "$version" > "$dst/VERSION"
  echo "pinned TypeScript $version"
fi
