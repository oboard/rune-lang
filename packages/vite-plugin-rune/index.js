import { execFile } from "node:child_process";
import { access, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { promisify } from "node:util";
import { transformWithEsbuild } from "vite";

const execFileAsync = promisify(execFile);

export function rune(options = {}) {
  const include = matcher(options.include ?? /\.rn$/);
  const exclude = matcher(options.exclude);
  const declaration = options.declaration ?? true;
  const runtime = options.runtime ?? "browser";
  const runeBin = options.runeBin ?? "rune";
  const runeRoot = options.runeRoot;
  let root = process.cwd();
  let sourcemap = false;

  return {
    name: "vite-plugin-rune",
    enforce: "pre",
    configResolved(config) {
      root = config.root;
      sourcemap = Boolean(config.build.sourcemap);
    },
    async resolveId(source, importer) {
      const cleanSource = cleanId(source);
      if (!cleanSource.endsWith(".rn")) {
        return null;
      }
      if (path.isAbsolute(cleanSource)) {
        return cleanSource;
      }
      if (!importer || !isRelative(cleanSource)) {
        return null;
      }
      const resolved = path.resolve(path.dirname(cleanId(importer)), cleanSource);
      try {
        await access(resolved);
        return resolved;
      } catch {
        return null;
      }
    },
    async load(id) {
      const clean = cleanId(id);
      if (!shouldHandle(clean, include, exclude)) {
        return null;
      }
      const code = await runRune(runeBin, ["ts", clean], root, runeRoot);
      if (declaration) {
        const dts = await runRune(runeBin, ["dts", clean], root, runeRoot);
        await writeFile(`${clean}.d.ts`, dts);
      }
      if (runtime === "browser" && code.includes("node:")) {
        this.warn(`Rune module ${path.relative(root, clean)} uses Node-only runtime APIs.`);
      }
      return transformWithEsbuild(code, `${clean}.ts`, {
        loader: "ts",
        sourcemap
      });
    }
  };
}

export default rune;

async function runRune(runeBin, args, cwd, runeRoot) {
  const command = Array.isArray(runeBin) ? runeBin[0] : runeBin;
  const prefixArgs = Array.isArray(runeBin) ? runeBin.slice(1) : [];
  try {
    const { stdout } = await execFileAsync(command, [...prefixArgs, ...args], {
      cwd,
      env: runeRoot ? { ...process.env, RUNE_ROOT: runeRoot } : process.env,
      maxBuffer: 32 * 1024 * 1024
    });
    return stdout;
  } catch (error) {
    const stderr = typeof error.stderr === "string" ? error.stderr.trim() : "";
    const stdout = typeof error.stdout === "string" ? error.stdout.trim() : "";
    const detail = [stderr, stdout].filter(Boolean).join("\n");
    throw new Error(detail || error.message);
  }
}

function shouldHandle(id, include, exclude) {
  return id.endsWith(".rn") && include(id) && !exclude(id);
}

function matcher(value) {
  const values = Array.isArray(value) ? value : value == null ? [] : [value];
  if (values.length === 0) {
    return () => false;
  }
  const matchers = values.map((item) => {
    if (typeof item === "function") {
      return item;
    }
    if (item instanceof RegExp) {
      return (id) => item.test(id);
    }
    const pattern = globToRegExp(item);
    return (id) => pattern.test(id);
  });
  return (id) => matchers.some((fn) => fn(id));
}

function globToRegExp(glob) {
  let source = "";
  for (let i = 0; i < glob.length; i++) {
    const ch = glob[i];
    if (ch === "*") {
      if (glob[i + 1] === "*") {
        source += ".*";
        i++;
      } else {
        source += "[^/]*";
      }
      continue;
    }
    if (ch === "?") {
      source += ".";
      continue;
    }
    source += escapeRegExp(ch);
  }
  return new RegExp(source + "$");
}

function escapeRegExp(value) {
  return value.replace(/[|\\{}()[\]^$+?.]/g, "\\$&");
}

function cleanId(id) {
  return id.startsWith("file://") ? fileURLToPath(id) : id.split("?", 1)[0];
}

function isRelative(id) {
  return id.startsWith("./") || id.startsWith("../");
}
