package tscodegen

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	codeusage "github.com/oboard/rune-lang/internal/codegen/usage"
	"github.com/oboard/rune-lang/internal/ir"
)

func Generate(file *ast.File, info *checker.Info) (string, error) {
	return GenerateIR(ir.LowerFile(file, info))
}

func GenerateIR(file *ir.File) (string, error) {
	usage := codeusage.Collect(file)
	if len(file.GoImports) > 0 {
		return "", fmt.Errorf("TypeScript backend does not support @go.import")
	}
	if usesGoFFI(usage) {
		return "", fmt.Errorf("TypeScript backend does not support @go FFI")
	}
	g := &generator{file: file}
	if g.typeScriptImports() {
		g.line("")
	}
	if fileUsesTaskRuntime(usage) {
		g.taskRuntime()
		g.line("")
	}
	if fileUsesResultRuntime(usage) {
		g.resultRuntime()
		g.line("")
	}
	if fileUsesErrorRuntime(usage) {
		g.errorRuntime()
		g.line("")
	}
	if fileUsesPathRuntime(usage) {
		g.pathRuntime()
		g.line("")
	}
	if fileUsesProcessRuntime(usage) {
		g.processRuntime()
		g.line("")
	}
	if fileUsesStringBufferRuntime(usage) {
		g.stringBufferRuntime()
		g.line("")
	}
	if fileUsesIterRuntime(usage) {
		g.iterRuntime()
		g.line("")
	}
	if fileUsesFSRuntime(usage) {
		g.fsRuntime()
		g.line("")
	}
	if fileUsesCompressRuntime(usage) {
		g.compressRuntime()
		g.line("")
	}
	if fileUsesNetRuntime(usage) {
		g.netRuntime()
		g.line("")
	}
	if fileUsesBytesRuntime(usage) {
		g.bytesRuntime()
		g.line("")
	}
	if fileUsesSignals(usage) {
		g.signalRuntime()
		g.line("")
	}
	for i, enum := range file.Enums {
		if i > 0 {
			g.line("")
		}
		g.enumType(enum)
	}
	if len(file.Enums) > 0 && len(file.Types) > 0 {
		g.line("")
	}
	for i, typ := range file.Types {
		if i > 0 {
			g.line("")
		}
		g.structType(typ)
		for _, method := range typ.Methods {
			g.line("")
			if err := g.method(typ, method); err != nil {
				return "", err
			}
		}
	}
	if (len(file.Types) > 0 || len(file.Enums) > 0) && len(file.Functions) > 0 {
		g.line("")
	}
	for i, fn := range file.Functions {
		if i > 0 {
			g.line("")
		}
		if err := g.function(fn); err != nil {
			return "", err
		}
	}
	if len(file.Functions) > 0 {
		g.line("")
		exports := make([]string, 0, len(file.Functions))
		for _, fn := range file.Functions {
			if fn.Private {
				continue
			}
			exports = append(exports, fmt.Sprintf("%s as %s", mangleIdent(fn.Name), tsExportName(fn.Name)))
		}
		if len(exports) > 0 {
			g.linef("export { %s };", join(exports, ", "))
		}
	}
	if err := g.codegenError(); err != nil {
		return g.buf.String(), err
	}
	return g.buf.String(), nil
}

func (g *generator) typeScriptImports() bool {
	byPath := map[string]map[string]bool{}
	var paths []string
	for _, imp := range g.file.TSImports {
		if len(imp.Functions) == 0 && len(imp.Values) == 0 {
			continue
		}
		if byPath[imp.Path] == nil {
			byPath[imp.Path] = map[string]bool{}
			paths = append(paths, imp.Path)
		}
		for _, fn := range imp.Functions {
			byPath[imp.Path][fn.Name] = true
		}
		for _, value := range imp.Values {
			byPath[imp.Path][value.Name] = true
		}
	}
	if len(paths) == 0 {
		return false
	}
	for _, path := range paths {
		names := make([]string, 0, len(byPath[path]))
		for name := range byPath[path] {
			names = append(names, name)
		}
		sort.Strings(names)
		specs := make([]string, 0, len(names))
		for _, name := range names {
			specs = append(specs, fmt.Sprintf("%s as %s", name, mangleIdent(name)))
		}
		g.linef("import { %s } from %s;", join(specs, ", "), strconv.Quote(typeScriptFileSpecifier(path)))
	}
	return true
}

func typeScriptFileSpecifier(path string) string {
	return (&url.URL{Scheme: "file", Path: path}).String()
}

func usesGoFFI(usage codeusage.Usage) bool {
	return usage.GoFFI
}

func (g *generator) taskRuntime() {
	g.line("const runeTasks: Promise<unknown>[] = [];")
	g.line("")
	g.line("function runeGo<T>(work: () => T | Promise<T>): Promise<T> {")
	g.indent++
	g.line("const task = Promise.resolve().then(work);")
	g.line("runeTasks.push(task);")
	g.line("task.finally(() => {")
	g.indent++
	g.line("const index = runeTasks.indexOf(task);")
	g.line("if (index >= 0) {")
	g.indent++
	g.line("runeTasks.splice(index, 1);")
	g.indent--
	g.line("}")
	g.indent--
	g.line("});")
	g.line("return task;")
	g.indent--
	g.line("}")
	g.line("")
	g.line("async function runeWaitAll(): Promise<void> {")
	g.indent++
	g.line("while (runeTasks.length > 0) {")
	g.indent++
	g.line("await Promise.allSettled([...runeTasks]);")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
}

func (g *generator) resultRuntime() {
	g.line("type RuneResult<T, E> = { ok: true; value: T } | { ok: false; error: E };")
	g.line("")
	g.line("function runeOk<T, E>(value: T): RuneResult<T, E> {")
	g.indent++
	g.line("return { ok: true, value };")
	g.indent--
	g.line("}")
	g.line("")
	g.line("function runeErr<T, E>(error: E): RuneResult<T, E> {")
	g.indent++
	g.line("return { ok: false, error };")
	g.indent--
	g.line("}")
}

func (g *generator) errorRuntime() {
	g.line("type RuneError = { code: number; message: string; cause: RuneError | null };")
	g.line("")
	g.line("function runeErrorFrom(error: unknown): RuneError {")
	g.indent++
	g.line("return { code: 1, message: error instanceof Error ? error.message : String(error), cause: null };")
	g.indent--
	g.line("}")
}

func (g *generator) pathRuntime() {
	g.line("function runePathParts(path: string): string[] {")
	g.indent++
	g.line("return path.replace(/\\\\+/g, \"/\").split(\"/\").filter((part) => part.length > 0);")
	g.indent--
	g.line("}")
	g.line("function runePathBasename(path: string): string { const parts = runePathParts(path); return parts.length === 0 ? \".\" : parts[parts.length - 1]; }")
	g.line("function runePathDirname(path: string): string { const normalized = runePathNormalize(path); const index = normalized.lastIndexOf(\"/\"); return index <= 0 ? (normalized.startsWith(\"/\") ? \"/\" : \".\") : normalized.slice(0, index); }")
	g.line("function runePathExtname(path: string): string { const base = runePathBasename(path); const index = base.lastIndexOf(\".\"); return index <= 0 ? \"\" : base.slice(index); }")
	g.line("function runePathJoin(parts: string[]): string { return runePathNormalize(parts.join(\"/\")); }")
	g.line("function runePathNormalize(path: string): string {")
	g.indent++
	g.line("const absolute = path.startsWith(\"/\");")
	g.line("const out: string[] = [];")
	g.line("for (const part of path.replace(/\\\\+/g, \"/\").split(\"/\")) {")
	g.indent++
	g.line("if (part === \"\" || part === \".\") continue;")
	g.line("if (part === \"..\") out.pop(); else out.push(part);")
	g.indent--
	g.line("}")
	g.line("const joined = out.join(\"/\");")
	g.line("return (absolute ? \"/\" : \"\") + (joined === \"\" ? (absolute ? \"\" : \".\") : joined);")
	g.indent--
	g.line("}")
	g.line("function runePathProcessObject(): any { return (globalThis as any).process ?? {}; }")
	g.line("function runePathResolve(parts: string[]): string { const cwd = runePathProcessObject().cwd?.() ?? \".\"; const path = parts.length === 0 || !runePathIsAbsolute(parts[0]) ? [cwd, ...parts] : parts; return runePathJoin(path); }")
	g.line("function runePathRelative(from: string, to: string): string { const fromParts = runePathParts(runePathResolve([from])); const toParts = runePathParts(runePathResolve([to])); while (fromParts.length > 0 && toParts.length > 0 && fromParts[0] === toParts[0]) { fromParts.shift(); toParts.shift(); } return [...fromParts.map(() => \"..\"), ...toParts].join(\"/\") || \".\"; }")
	g.line("function runePathIsAbsolute(path: string): boolean { return path.startsWith(\"/\") || /^[A-Za-z]:[\\\\/]/.test(path); }")
}

func (g *generator) processRuntime() {
	g.line("function runeProcessObject(): any { return (globalThis as any).process ?? {}; }")
	g.line("function runeProcessArgv(): string[] { return [...(runeProcessObject().argv ?? [])]; }")
	g.line("function runeProcessCwd(): string { return runeProcessObject().cwd?.() ?? \"\"; }")
	g.line("function runeProcessEnv(name: string): string | null { const value = runeProcessObject().env?.[name]; return value === undefined ? null : String(value); }")
	g.line("function runeProcessExit(code: number): never { const proc = runeProcessObject(); if (typeof proc.exit === \"function\") proc.exit(code); throw new Error(`process.exit(${code})`); }")
	g.line("function runeProcessPlatform(): string { return runeProcessObject().platform ?? \"unknown\"; }")
}

func (g *generator) stringBufferRuntime() {
	g.line("class RuneStringBuffer {")
	g.indent++
	g.line("private parts: string[] = [];")
	g.line("constructor(value = \"\") { if (value.length > 0) this.parts.push(value); }")
	g.line("length(): number { return Array.from(this.toString()).length; }")
	g.line("clear(): void { this.parts = []; }")
	g.line("append(value: string): RuneStringBuffer { this.parts.push(value); return this; }")
	g.line("appendLine(value: string): RuneStringBuffer { this.parts.push(value, \"\\n\"); return this; }")
	g.line("toString(): string { return this.parts.join(\"\"); }")
	g.indent--
	g.line("}")
}

func (g *generator) iterRuntime() {
	g.line("type RuneIter<T> = { next: () => [T, boolean] };")
}

func (g *generator) fsRuntime() {
	g.line("type RuneFileStat = { size: number; isFile: boolean; isDirectory: boolean };")
	g.line("")
	g.line("async function runeReadFile(path: string): Promise<RuneResult<Uint8Array, RuneError>> { return runeFsReadFile(path); }")
	g.line("")
	g.line("async function runeFsReadFile(path: string): Promise<RuneResult<Uint8Array, RuneError>> {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("const fs = await import(\"node:fs/promises\");")
	g.line("return runeOk<Uint8Array, RuneError>(await fs.readFile(path));")
	g.indent--
	g.line("} catch (error) {")
	g.indent++
	g.line("return runeErr<Uint8Array, RuneError>(runeErrorFrom(error));")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("async function runeFsReadFileText(path: string): Promise<RuneResult<string, RuneError>> {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("const fs = await import(\"node:fs/promises\");")
	g.line("return runeOk<string, RuneError>(await fs.readFile(path, \"utf8\"));")
	g.indent--
	g.line("} catch (error) {")
	g.indent++
	g.line("return runeErr<string, RuneError>(runeErrorFrom(error));")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("async function runeFsWriteFile(path: string, data: Uint8Array): Promise<RuneResult<void, RuneError>> {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("const fs = await import(\"node:fs/promises\");")
	g.line("await fs.writeFile(path, data);")
	g.line("return runeOk<void, RuneError>(undefined);")
	g.indent--
	g.line("} catch (error) {")
	g.indent++
	g.line("return runeErr<void, RuneError>(runeErrorFrom(error));")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("async function runeFsWriteFileText(path: string, data: string): Promise<RuneResult<void, RuneError>> {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("const fs = await import(\"node:fs/promises\");")
	g.line("await fs.writeFile(path, data, \"utf8\");")
	g.line("return runeOk<void, RuneError>(undefined);")
	g.indent--
	g.line("} catch (error) {")
	g.indent++
	g.line("return runeErr<void, RuneError>(runeErrorFrom(error));")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("async function runeFsExists(path: string): Promise<RuneResult<boolean, RuneError>> {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("const fs = await import(\"node:fs/promises\");")
	g.line("await fs.stat(path);")
	g.line("return runeOk<boolean, RuneError>(true);")
	g.indent--
	g.line("} catch (error: any) {")
	g.indent++
	g.line("if (error?.code === \"ENOENT\") return runeOk<boolean, RuneError>(false);")
	g.line("return runeErr<boolean, RuneError>(runeErrorFrom(error));")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("async function runeFsReaddir(path: string): Promise<RuneResult<string[], RuneError>> {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("const fs = await import(\"node:fs/promises\");")
	g.line("return runeOk<string[], RuneError>(await fs.readdir(path));")
	g.indent--
	g.line("} catch (error) {")
	g.indent++
	g.line("return runeErr<string[], RuneError>(runeErrorFrom(error));")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("async function runeFsMkdir(path: string): Promise<RuneResult<void, RuneError>> {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("const fs = await import(\"node:fs/promises\");")
	g.line("await fs.mkdir(path);")
	g.line("return runeOk<void, RuneError>(undefined);")
	g.indent--
	g.line("} catch (error) {")
	g.indent++
	g.line("return runeErr<void, RuneError>(runeErrorFrom(error));")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("async function runeFsRemove(path: string): Promise<RuneResult<void, RuneError>> {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("const fs = await import(\"node:fs/promises\");")
	g.line("await fs.rm(path);")
	g.line("return runeOk<void, RuneError>(undefined);")
	g.indent--
	g.line("} catch (error) {")
	g.indent++
	g.line("return runeErr<void, RuneError>(runeErrorFrom(error));")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("async function runeFsStat(path: string): Promise<RuneResult<RuneFileStat, RuneError>> {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("const fs = await import(\"node:fs/promises\");")
	g.line("const stat = await fs.stat(path);")
	g.line("return runeOk<RuneFileStat, RuneError>({ size: Number(stat.size), isFile: stat.isFile(), isDirectory: stat.isDirectory() });")
	g.indent--
	g.line("} catch (error) {")
	g.indent++
	g.line("return runeErr<RuneFileStat, RuneError>(runeErrorFrom(error));")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
}

func (g *generator) compressRuntime() {
	g.line("type RuneZlibFunction = (data: Uint8Array, callback: (error: unknown, result: Uint8Array) => void) => void;")
	g.line("async function runeCompressCall(method: string, data: Uint8Array): Promise<RuneResult<Uint8Array, RuneError>> {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("const zlib = (await import(\"node:zlib\")) as Record<string, unknown>;")
	g.line("const fn = zlib[method] as RuneZlibFunction | undefined;")
	g.line("if (typeof fn !== \"function\") { return runeErr<Uint8Array, RuneError>(runeErrorFrom(new Error(\"node:zlib \" + method + \" is not available\"))); }")
	g.line("return await new Promise<RuneResult<Uint8Array, RuneError>>((resolve) => fn(data, (error, result) => error ? resolve(runeErr<Uint8Array, RuneError>(runeErrorFrom(error))) : resolve(runeOk<Uint8Array, RuneError>(result))));")
	g.indent--
	g.line("} catch (error) {")
	g.indent++
	g.line("return runeErr<Uint8Array, RuneError>(runeErrorFrom(error));")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("async function runeCompressGzip(data: Uint8Array): Promise<RuneResult<Uint8Array, RuneError>> {")
	g.indent++
	g.line("return runeCompressCall(\"gzip\", data);")
	g.indent--
	g.line("}")
	g.line("async function runeCompressGunzip(data: Uint8Array): Promise<RuneResult<Uint8Array, RuneError>> {")
	g.indent++
	g.line("return runeCompressCall(\"gunzip\", data);")
	g.indent--
	g.line("}")
	g.line("async function runeCompressDeflate(data: Uint8Array): Promise<RuneResult<Uint8Array, RuneError>> {")
	g.indent++
	g.line("return runeCompressCall(\"deflate\", data);")
	g.indent--
	g.line("}")
	g.line("async function runeCompressInflate(data: Uint8Array): Promise<RuneResult<Uint8Array, RuneError>> {")
	g.indent++
	g.line("return runeCompressCall(\"inflate\", data);")
	g.indent--
	g.line("}")
	g.line("async function runeCompressBrotli(data: Uint8Array): Promise<RuneResult<Uint8Array, RuneError>> {")
	g.indent++
	g.line("return runeCompressCall(\"brotliCompress\", data);")
	g.indent--
	g.line("}")
	g.line("async function runeCompressUnbrotli(data: Uint8Array): Promise<RuneResult<Uint8Array, RuneError>> {")
	g.indent++
	g.line("return runeCompressCall(\"brotliDecompress\", data);")
	g.indent--
	g.line("}")
	g.line("async function runeCompressZstd(data: Uint8Array): Promise<RuneResult<Uint8Array, RuneError>> {")
	g.indent++
	g.line("return runeCompressCall(\"zstdCompress\", data);")
	g.indent--
	g.line("}")
	g.line("async function runeCompressUnzstd(data: Uint8Array): Promise<RuneResult<Uint8Array, RuneError>> {")
	g.indent++
	g.line("return runeCompressCall(\"zstdDecompress\", data);")
	g.indent--
	g.line("}")
	g.line("function runeTextEncoder(): TextEncoder { return new TextEncoder(); }")
	g.line("function runeTextDecoder(): TextDecoder { return new TextDecoder(); }")
	g.line("async function runeCompressGzipText(value: string): Promise<RuneResult<Uint8Array, RuneError>> { return runeCompressGzip(runeTextEncoder().encode(value)); }")
	g.line("async function runeCompressGunzipText(data: Uint8Array): Promise<RuneResult<string, RuneError>> {")
	g.indent++
	g.line("return runeCompressDecodeText(data, runeCompressGunzip);")
	g.indent--
	g.line("}")
	g.line("async function runeCompressBrotliText(value: string): Promise<RuneResult<Uint8Array, RuneError>> { return runeCompressBrotli(runeTextEncoder().encode(value)); }")
	g.line("async function runeCompressUnbrotliText(data: Uint8Array): Promise<RuneResult<string, RuneError>> {")
	g.indent++
	g.line("return runeCompressDecodeText(data, runeCompressUnbrotli);")
	g.indent--
	g.line("}")
	g.line("async function runeCompressZstdText(value: string): Promise<RuneResult<Uint8Array, RuneError>> { return runeCompressZstd(runeTextEncoder().encode(value)); }")
	g.line("async function runeCompressUnzstdText(data: Uint8Array): Promise<RuneResult<string, RuneError>> {")
	g.indent++
	g.line("return runeCompressDecodeText(data, runeCompressUnzstd);")
	g.indent--
	g.line("}")
	g.line("async function runeCompressDecodeText(data: Uint8Array, decode: (data: Uint8Array) => Promise<RuneResult<Uint8Array, RuneError>>): Promise<RuneResult<string, RuneError>> {")
	g.indent++
	g.line("const result = await decode(data);")
	g.line("return result.ok ? runeOk<string, RuneError>(runeTextDecoder().decode(result.value)) : runeErr<string, RuneError>(result.error);")
	g.indent--
	g.line("}")
}

func (g *generator) netRuntime() {
	g.line("type RuneTCPConnection = { socket: any };")
	g.line("type RuneTCPListener = { server: any; address: string };")
	g.line("function runeSplitAddress(address: string): { host: string; port: number } {")
	g.indent++
	g.line("const index = address.lastIndexOf(\":\");")
	g.line("return index < 0 ? { host: \"127.0.0.1\", port: Number(address) } : { host: address.slice(0, index) || \"127.0.0.1\", port: Number(address.slice(index + 1)) };")
	g.indent--
	g.line("}")
	g.line("async function runeNetConnect(address: string): Promise<RuneResult<RuneTCPConnection, RuneError>> {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("const net = await import(\"node:net\");")
	g.line("const target = runeSplitAddress(address);")
	g.line("return await new Promise<RuneResult<RuneTCPConnection, RuneError>>((resolve) => {")
	g.indent++
	g.line("const socket = net.createConnection(target.port, target.host, () => resolve(runeOk<RuneTCPConnection, RuneError>({ socket })));")
	g.line("socket.once(\"error\", (error: unknown) => resolve(runeErr<RuneTCPConnection, RuneError>(runeErrorFrom(error))));")
	g.indent--
	g.line("});")
	g.indent--
	g.line("} catch (error) { return runeErr<RuneTCPConnection, RuneError>(runeErrorFrom(error)); }")
	g.indent--
	g.line("}")
	g.line("async function runeNetListen(address: string): Promise<RuneResult<RuneTCPListener, RuneError>> {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("const net = await import(\"node:net\");")
	g.line("const target = runeSplitAddress(address);")
	g.line("return await new Promise<RuneResult<RuneTCPListener, RuneError>>((resolve) => {")
	g.indent++
	g.line("const server = net.createServer();")
	g.line("server.once(\"error\", (error: unknown) => resolve(runeErr<RuneTCPListener, RuneError>(runeErrorFrom(error))));")
	g.line("server.listen(target.port, target.host, () => resolve(runeOk<RuneTCPListener, RuneError>({ server, address: String(server.address()) })));")
	g.indent--
	g.line("});")
	g.indent--
	g.line("} catch (error) { return runeErr<RuneTCPListener, RuneError>(runeErrorFrom(error)); }")
	g.indent--
	g.line("}")
	g.line("function runeNetConnectionRead(connection: RuneTCPConnection, length: number): Promise<RuneResult<Uint8Array, RuneError>> {")
	g.indent++
	g.line("return new Promise((resolve) => {")
	g.indent++
	g.line("const onData = (chunk: Uint8Array) => { cleanup(); resolve(runeOk<Uint8Array, RuneError>(chunk.slice(0, length))); };")
	g.line("const onError = (error: unknown) => { cleanup(); resolve(runeErr<Uint8Array, RuneError>(runeErrorFrom(error))); };")
	g.line("const cleanup = () => { connection.socket.off(\"data\", onData); connection.socket.off(\"error\", onError); };")
	g.line("connection.socket.once(\"data\", onData);")
	g.line("connection.socket.once(\"error\", onError);")
	g.indent--
	g.line("});")
	g.indent--
	g.line("}")
	g.line("function runeNetConnectionWrite(connection: RuneTCPConnection, data: Uint8Array): Promise<RuneResult<number, RuneError>> {")
	g.indent++
	g.line("return new Promise((resolve) => connection.socket.write(data, (error: unknown) => error ? resolve(runeErr<number, RuneError>(runeErrorFrom(error))) : resolve(runeOk<number, RuneError>(data.byteLength))));")
	g.indent--
	g.line("}")
	g.line("function runeNetConnectionClose(connection: RuneTCPConnection): Promise<RuneResult<void, RuneError>> { connection.socket.end(); return Promise.resolve(runeOk<void, RuneError>(undefined)); }")
	g.line("function runeNetListenerAddress(listener: RuneTCPListener): string { const value = listener.server.address(); return typeof value === \"string\" ? value : `${value.address}:${value.port}`; }")
	g.line("function runeNetListenerAccept(listener: RuneTCPListener): Promise<RuneResult<RuneTCPConnection, RuneError>> {")
	g.indent++
	g.line("return new Promise((resolve) => {")
	g.indent++
	g.line("const onConnection = (socket: any) => { cleanup(); resolve(runeOk<RuneTCPConnection, RuneError>({ socket })); };")
	g.line("const onError = (error: unknown) => { cleanup(); resolve(runeErr<RuneTCPConnection, RuneError>(runeErrorFrom(error))); };")
	g.line("const cleanup = () => { listener.server.off(\"connection\", onConnection); listener.server.off(\"error\", onError); };")
	g.line("listener.server.once(\"connection\", onConnection);")
	g.line("listener.server.once(\"error\", onError);")
	g.indent--
	g.line("});")
	g.indent--
	g.line("}")
	g.line("function runeNetListenerClose(listener: RuneTCPListener): Promise<RuneResult<void, RuneError>> { return new Promise((resolve) => listener.server.close((error: unknown) => error ? resolve(runeErr<void, RuneError>(runeErrorFrom(error))) : resolve(runeOk<void, RuneError>(undefined)))); }")
}

func join(parts []string, sep string) string {
	out := ""
	for i, part := range parts {
		if i > 0 {
			out += sep
		}
		out += part
	}
	return out
}

func fileUsesBytesRuntime(usage codeusage.Usage) bool {
	return fileUsesType(usage, checker.Buffer) ||
		fileUsesType(usage, checker.Reader) ||
		fileUsesType(usage, checker.Writer)
}

func fileUsesPathRuntime(usage codeusage.Usage) bool {
	return usage.HasIntrinsicPrefix("path.")
}

func fileUsesProcessRuntime(usage codeusage.Usage) bool {
	return usage.HasIntrinsicPrefix("process.")
}

func fileUsesStringBufferRuntime(usage codeusage.Usage) bool {
	return fileUsesType(usage, checker.StringBuffer) || usage.HasIntrinsicPrefix("stringbuffer.")
}

func fileUsesIterRuntime(usage codeusage.Usage) bool {
	return usage.HasGeneric("Iter") || usage.HasIntrinsicPrefix("iter.")
}

func fileUsesCompressRuntime(usage codeusage.Usage) bool {
	return usage.HasIntrinsicPrefix("compress.")
}

func fileUsesNetRuntime(usage codeusage.Usage) bool {
	return fileUsesType(usage, checker.TCPConnection) ||
		fileUsesType(usage, checker.TCPListener) ||
		usage.HasIntrinsicPrefix("net.")
}

func fileUsesType(usage codeusage.Usage, typ checker.Type) bool {
	return usage.HasType(typ)
}

func fileUsesSignals(usage codeusage.Usage) bool {
	return usage.Signal
}

func fileUsesTaskRuntime(usage codeusage.Usage) bool {
	return usage.HasGeneric("Task") ||
		fileUsesFSRuntime(usage) ||
		fileUsesCompressRuntime(usage) ||
		fileUsesNetRuntime(usage) ||
		usage.UsesAsyncRuntime()
}

func fileUsesResultRuntime(usage codeusage.Usage) bool {
	return usage.HasGeneric("Result") || fileUsesFSRuntime(usage) || fileUsesCompressRuntime(usage) || fileUsesNetRuntime(usage)
}

func fileUsesErrorRuntime(usage codeusage.Usage) bool {
	return fileUsesType(usage, checker.Error) || fileUsesFSRuntime(usage) || fileUsesCompressRuntime(usage) || fileUsesNetRuntime(usage)
}

func fileUsesFSRuntime(usage codeusage.Usage) bool {
	return fileUsesType(usage, checker.FileStat) || usage.HasIntrinsicPrefix("fs.")
}
