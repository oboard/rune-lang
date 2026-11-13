# Pinned TypeScript Library Set

RFC 0001 makes TypeScript declaration files the single source of truth for
Rune's built-in API surface. The pinned version lives in `VERSION`.

Expected files in this directory:

- `lib.es5.d.ts` (required)
- `lib.es2015.d.ts` (optional today)
- `lib.dom.d.ts` (optional; required for DOM targets)

Populate the directory from a TypeScript installation:

```sh
scripts/sync-ts-lib.sh "/path/to/typescript/lib"
```

The files are not committed by default because `lib.dom.d.ts` is ~2 MB. CI and
release builds should vendor the exact `VERSION` before freezing artifacts;
`internal/libenv` falls back to a local `node_modules/typescript/lib` or the
VS Code bundled TypeScript when this directory is empty.
