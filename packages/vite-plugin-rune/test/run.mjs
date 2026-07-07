import { execFile } from "node:child_process";
import { mkdtemp, readFile, realpath, rm, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { promisify } from "node:util";
import { build } from "vite";
import { rune } from "../index.js";

const execFileAsync = promisify(execFile);
const repoRoot = path.resolve(new URL("../../..", import.meta.url).pathname);
const tmp = await realpath(await mkdtemp(path.join(os.tmpdir(), "rune-vite-plugin-")));
const runeBin = path.join(tmp, process.platform === "win32" ? "rune.exe" : "rune");
const appDir = path.join(tmp, "app");

try {
  await execFileAsync("go", ["build", "-o", runeBin, "./cmd/rune"], { cwd: repoRoot });
  await writeFixture(appDir);

  await build({
    root: appDir,
    logLevel: "silent",
    configFile: false,
    plugins: [rune({ runeBin, runeRoot: repoRoot, runtime: "node" })],
    build: {
      outDir: "dist",
      emptyOutDir: true
    }
  });

  await execFileAsync("tsc", ["-p", appDir], { cwd: appDir });

  const dts = await readFile(path.join(appDir, "src", "math.rn.d.ts"), "utf8");
  assertIncludes(dts, "export type User = __User;");
  assertIncludes(dts, "export declare const Status: typeof __Status;");
  assertIncludes(dts, "export declare const scaled: typeof __scaled;");
} finally {
  await rm(tmp, { recursive: true, force: true });
}

async function writeFixture(root) {
  await mkdirp(root);
  await writeFile(path.join(root, "index.html"), `<div id="app"></div><script type="module" src="/src/main.ts"></script>\n`);
  await mkdirp(path.join(root, "src"));
  await writeFile(path.join(root, "package.json"), `{"type":"module"}\n`);
  await writeFile(path.join(root, "tsconfig.json"), JSON.stringify({
    compilerOptions: {
      target: "ES2020",
      module: "ESNext",
      moduleResolution: "Bundler",
      strict: true,
      noEmit: true,
      lib: ["ES2020", "DOM"]
    },
    include: ["src"]
  }, null, 2));
  await writeFile(path.join(root, "src", "scale.ts"), `export function scale(value: number): number {\n  return value * 2;\n}\n`);
  await writeFile(path.join(root, "src", "math.rn"), `@"./scale.ts"

+ User: {
  name: String
  age: Int
}

+ Status: {
  Ready = 1
  Done = 2
}

+ const answer: Int = 42

+ add(a: Int, b: Int) -> Int => a + b

+ scaled(value: Double) -> Double => scale(value)
`);
  await writeFile(path.join(root, "src", "main.ts"), `import { Status, add, answer, scaled, type User } from "./math.rn";

const user: User = { name: "Rune", age: add(20, 2) };
document.querySelector("#app")!.textContent = [
  user.name,
  user.age,
  answer,
  Status.Ready,
  scaled(2)
].join(":");
`);
}

async function mkdirp(dir) {
  await import("node:fs/promises").then(({ mkdir }) => mkdir(dir, { recursive: true }));
}

function assertIncludes(source, value) {
  if (!source.includes(value)) {
    throw new Error(`expected generated declaration to include ${JSON.stringify(value)}\n${source}`);
  }
}
