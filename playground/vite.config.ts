import { execFileSync } from "node:child_process";
import { copyFileSync, existsSync, mkdirSync } from "node:fs";
import { dirname, join } from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";
import type { Plugin } from "vite";
import { defineConfig } from "vite-plus";
import react from "@vitejs/plugin-react";

const playgroundDir = dirname(fileURLToPath(import.meta.url));
const repoDir = join(playgroundDir, "..");
const publicDir = join(playgroundDir, "public");
const playgroundBase = process.env.PLAYGROUND_BASE ?? "/";

function buildRuneWasm() {
  mkdirSync(publicDir, { recursive: true });
  if (!commandExists("go")) {
    throw new Error("Go is required to build Rune wasm. Run scripts/build-docs.sh or install Go before building.");
  }
  const goroot = execFileSync("go", ["env", "GOROOT"], { encoding: "utf8" }).trim();
  const wasmExec = [join(goroot, "misc", "wasm", "wasm_exec.js"), join(goroot, "lib", "wasm", "wasm_exec.js")].find(
    (candidate) => existsSync(candidate),
  );
  if (!wasmExec) {
    throw new Error(`wasm_exec.js was not found under ${goroot}`);
  }
  copyFileSync(wasmExec, join(publicDir, "wasm_exec.js"));
  execFileSync("go", ["build", "-o", join(publicDir, "rune.wasm"), "./cmd/rune-wasm"], {
    cwd: repoDir,
    env: {
      ...process.env,
      GOARCH: "wasm",
      GOOS: "js",
    },
    stdio: "inherit",
  });
}

function commandExists(command: string) {
  try {
    execFileSync(command, ["version"], { stdio: "ignore" });
    return true;
  } catch {
    return false;
  }
}

function runeWasmPlugin(): Plugin {
  return {
    name: "rune-wasm",
    buildStart() {
      buildRuneWasm();
    },
    configureServer(server) {
      buildRuneWasm();
      const watched = [join(repoDir, "cmd", "rune-wasm"), join(repoDir, "core"), join(repoDir, "internal")];
      server.watcher.add(watched);
      server.watcher.on("change", (file) => {
        if (watched.some((path) => file.startsWith(path))) {
          buildRuneWasm();
        }
      });
    },
  };
}

// https://vite.dev/config/
export default defineConfig({
  base: playgroundBase,
  build: {
    rolldownOptions: {
      output: {
        strictExecutionOrder: true,
        codeSplitting: {
          maxSize: 450_000,
          groups: [
            {
              name(id) {
                return vendorChunkName(id);
              },
              test(id) {
                return id.includes("node_modules");
              },
            },
          ],
        },
      },
    },
  },
  fmt: {},
  lint: {
    plugins: ["oxc", "typescript", "unicorn", "react"],
    categories: {
      correctness: "warn",
    },
    env: {
      builtin: true,
    },
    ignorePatterns: ["dist", "public/wasm_exec.js"],
    overrides: [
      {
        files: ["**/*.{ts,tsx}"],
        rules: {
          "constructor-super": "error",
          "for-direction": "error",
          "getter-return": "error",
          "no-async-promise-executor": "error",
          "no-case-declarations": "error",
          "no-class-assign": "error",
          "no-compare-neg-zero": "error",
          "no-cond-assign": "error",
          "no-const-assign": "error",
          "no-constant-binary-expression": "error",
          "no-constant-condition": "error",
          "no-control-regex": "error",
          "no-debugger": "error",
          "no-delete-var": "error",
          "no-dupe-class-members": "error",
          "no-dupe-else-if": "error",
          "no-dupe-keys": "error",
          "no-duplicate-case": "error",
          "no-empty": "error",
          "no-empty-character-class": "error",
          "no-empty-pattern": "error",
          "no-empty-static-block": "error",
          "no-ex-assign": "error",
          "no-extra-boolean-cast": "error",
          "no-fallthrough": "error",
          "no-func-assign": "error",
          "no-global-assign": "error",
          "no-import-assign": "error",
          "no-invalid-regexp": "error",
          "no-irregular-whitespace": "error",
          "no-loss-of-precision": "error",
          "no-misleading-character-class": "error",
          "no-new-native-nonconstructor": "error",
          "no-nonoctal-decimal-escape": "error",
          "no-obj-calls": "error",
          "no-prototype-builtins": "error",
          "no-redeclare": "error",
          "no-regex-spaces": "error",
          "no-self-assign": "error",
          "no-setter-return": "error",
          "no-shadow-restricted-names": "error",
          "no-sparse-arrays": "error",
          "no-this-before-super": "error",
          "no-unassigned-vars": "error",
          "no-undef": "error",
          "no-unexpected-multiline": "error",
          "no-unreachable": "error",
          "no-unsafe-finally": "error",
          "no-unsafe-negation": "error",
          "no-unsafe-optional-chaining": "error",
          "no-unused-labels": "error",
          "no-unused-private-class-members": "error",
          "no-unused-vars": "error",
          "no-useless-assignment": "error",
          "no-useless-backreference": "error",
          "no-useless-catch": "error",
          "no-useless-escape": "error",
          "no-with": "error",
          "preserve-caught-error": "error",
          "require-yield": "error",
          "use-isnan": "error",
          "valid-typeof": "error",
          "no-array-constructor": "error",
          "no-unused-expressions": "error",
          "typescript/ban-ts-comment": "error",
          "typescript/no-duplicate-enum-values": "error",
          "typescript/no-empty-object-type": "error",
          "typescript/no-explicit-any": "error",
          "typescript/no-extra-non-null-assertion": "error",
          "typescript/no-misused-new": "error",
          "typescript/no-namespace": "error",
          "typescript/no-non-null-asserted-optional-chain": "error",
          "typescript/no-require-imports": "error",
          "typescript/no-this-alias": "error",
          "typescript/no-unnecessary-type-constraint": "error",
          "typescript/no-unsafe-declaration-merging": "error",
          "typescript/no-unsafe-function-type": "error",
          "typescript/no-wrapper-object-types": "error",
          "typescript/prefer-as-const": "error",
          "typescript/prefer-namespace-keyword": "error",
          "typescript/triple-slash-reference": "error",
          "react/rules-of-hooks": "error",
          "react/exhaustive-deps": "warn",
          "react/only-export-components": [
            "error",
            {
              allowConstantExport: true,
            },
          ],
        },
        env: {
          browser: true,
        },
      },
    ],
    options: {
      typeAware: true,
      typeCheck: true,
    },
  },
  plugins: [runeWasmPlugin(), react()],
});

function vendorChunkName(id: string) {
  if (!id.includes("node_modules")) {
    return undefined;
  }
  if (id.includes("@monaco-editor/react") || id.includes("@monaco-editor/loader")) {
    return "vendor-monaco-react";
  }
  if (id.includes("monaco-editor")) {
    return monacoChunkName(id);
  }
  if (id.includes("react") || id.includes("scheduler")) {
    return "vendor-react";
  }
  if (id.includes("esbuild-wasm")) {
    return "vendor-esbuild";
  }
  if (id.includes("lucide-react")) {
    return "vendor-icons";
  }
  return "vendor";
}

function monacoChunkName(id: string) {
  const detailedContrib = id.match(/monaco-editor\/esm\/vs\/editor\/contrib\/([^/]+)\/(?:([^/]+)\/)?([^/]+)\.js/);
  if (detailedContrib && ["codeAction", "gotoError"].includes(detailedContrib[1])) {
    return `monaco-contrib-${chunkSafeName(detailedContrib[1])}-${chunkSafeName(detailedContrib[3])}`;
  }

  const contrib = id.match(/monaco-editor\/esm\/vs\/editor\/contrib\/([^/]+)/);
  if (contrib) {
    return `monaco-contrib-${chunkSafeName(contrib[1])}`;
  }

  const standalone = id.match(/monaco-editor\/esm\/vs\/editor\/standalone\/([^/]+)/);
  if (standalone) {
    return `monaco-standalone-${chunkSafeName(standalone[1])}`;
  }

  const detailedBaseBrowser = id.match(/monaco-editor\/esm\/vs\/base\/browser\/(?:(ui|dompurify)\/)?([^/]+)\.js/);
  if (detailedBaseBrowser) {
    return detailedBaseBrowser[1]
      ? `monaco-base-browser-${chunkSafeName(detailedBaseBrowser[1])}-${chunkSafeName(detailedBaseBrowser[2])}`
      : `monaco-base-browser-${chunkSafeName(detailedBaseBrowser[2])}`;
  }

  const editorBrowser = id.match(/monaco-editor\/esm\/vs\/editor\/browser\/([^/]+)/);
  if (editorBrowser) {
    return `monaco-editor-browser-${chunkSafeName(editorBrowser[1])}`;
  }

  const editorCommon = id.match(/monaco-editor\/esm\/vs\/editor\/common\/([^/]+)/);
  if (editorCommon) {
    return `monaco-editor-common-${chunkSafeName(editorCommon[1])}`;
  }

  const platform = id.match(/monaco-editor\/esm\/vs\/platform\/([^/]+)/);
  if (platform) {
    return `monaco-platform-${chunkSafeName(platform[1])}`;
  }

  const base = id.match(/monaco-editor\/esm\/vs\/base\/([^/]+)/);
  if (base) {
    return `monaco-base-${chunkSafeName(base[1])}`;
  }

  return "monaco-core";
}

function chunkSafeName(value: string) {
  return value.replaceAll(/[^A-Za-z0-9_-]/g, "-");
}
