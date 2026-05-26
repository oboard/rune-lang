package tscodegen

import (
	"fmt"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
)

func Generate(file *ast.File, info *checker.Info) (string, error) {
	return GenerateIR(ir.LowerFile(file, info))
}

func GenerateIR(file *ir.File) (string, error) {
	if len(file.GoImports) > 0 {
		return "", fmt.Errorf("TypeScript backend does not support @go.import")
	}
	if usesGoFFI(file) {
		return "", fmt.Errorf("TypeScript backend does not support @go FFI")
	}
	g := &generator{file: file}
	if fileUsesTaskRuntime(file) {
		g.taskRuntime()
		g.line("")
	}
	if fileUsesResultRuntime(file) {
		g.resultRuntime()
		g.line("")
	}
	if fileUsesErrorRuntime(file) {
		g.errorRuntime()
		g.line("")
	}
	if fileUsesPathRuntime(file) {
		g.pathRuntime()
		g.line("")
	}
	if fileUsesProcessRuntime(file) {
		g.processRuntime()
		g.line("")
	}
	if fileUsesStringBufferRuntime(file) {
		g.stringBufferRuntime()
		g.line("")
	}
	if fileUsesIterRuntime(file) {
		g.iterRuntime()
		g.line("")
	}
	if fileUsesFSRuntime(file) {
		g.fsRuntime()
		g.line("")
	}
	if fileUsesCompressRuntime(file) {
		g.compressRuntime()
		g.line("")
	}
	if fileUsesNetRuntime(file) {
		g.netRuntime()
		g.line("")
	}
	if fileUsesBinaryRuntime(file) {
		g.binaryRuntime()
		g.line("")
	}
	if fileUsesSignals(file) {
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
			exports = append(exports, fmt.Sprintf("%s as %s", mangleIdent(fn.Name), fn.Name))
		}
		g.linef("export { %s };", join(exports, ", "))
	}
	if err := g.codegenError(); err != nil {
		return g.buf.String(), err
	}
	return g.buf.String(), nil
}

func usesGoFFI(file *ir.File) bool {
	found := false
	for _, fn := range file.Functions {
		ir.WalkExpr(fn.Body, func(expr ir.Expr) {
			if selectorUsesGo(expr) {
				found = true
			}
		})
	}
	for _, typ := range file.Types {
		for _, method := range typ.Methods {
			ir.WalkExpr(method.Body, func(expr ir.Expr) {
				if selectorUsesGo(expr) {
					found = true
				}
			})
		}
	}
	return found
}

func selectorUsesGo(expr ir.Expr) bool {
	sel, ok := expr.(*ir.SelectorExpr)
	if !ok {
		return false
	}
	at, ok := sel.Receiver.(*ir.AtExpr)
	return ok && at.Name == "go"
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
	g.line("function runeIterRange(start: number, end: number): number[] { return runeIterRangeStep(start, end, 1); }")
	g.line("function runeIterRangeStep(start: number, end: number, step: number): number[] {")
	g.indent++
	g.line("if (step === 0) throw new RangeError(\"iter step cannot be zero\");")
	g.line("const out: number[] = [];")
	g.line("if (step > 0) { for (let value = start; value < end; value += step) out.push(value); }")
	g.line("else { for (let value = start; value > end; value += step) out.push(value); }")
	g.line("return out;")
	g.indent--
	g.line("}")
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
	g.line("async function runeCompressGzip(data: Uint8Array): Promise<RuneResult<Uint8Array, RuneError>> {")
	g.indent++
	g.line("try { const zlib = await import(\"node:zlib\"); return await new Promise((resolve) => zlib.gzip(data, (error, result) => error ? resolve(runeErr<Uint8Array, RuneError>(runeErrorFrom(error))) : resolve(runeOk<Uint8Array, RuneError>(result)))); }")
	g.line("catch (error) { return runeErr<Uint8Array, RuneError>(runeErrorFrom(error)); }")
	g.indent--
	g.line("}")
	g.line("async function runeCompressGunzip(data: Uint8Array): Promise<RuneResult<Uint8Array, RuneError>> {")
	g.indent++
	g.line("try { const zlib = await import(\"node:zlib\"); return await new Promise((resolve) => zlib.gunzip(data, (error, result) => error ? resolve(runeErr<Uint8Array, RuneError>(runeErrorFrom(error))) : resolve(runeOk<Uint8Array, RuneError>(result)))); }")
	g.line("catch (error) { return runeErr<Uint8Array, RuneError>(runeErrorFrom(error)); }")
	g.indent--
	g.line("}")
	g.line("async function runeCompressDeflate(data: Uint8Array): Promise<RuneResult<Uint8Array, RuneError>> {")
	g.indent++
	g.line("try { const zlib = await import(\"node:zlib\"); return await new Promise((resolve) => zlib.deflate(data, (error, result) => error ? resolve(runeErr<Uint8Array, RuneError>(runeErrorFrom(error))) : resolve(runeOk<Uint8Array, RuneError>(result)))); }")
	g.line("catch (error) { return runeErr<Uint8Array, RuneError>(runeErrorFrom(error)); }")
	g.indent--
	g.line("}")
	g.line("async function runeCompressInflate(data: Uint8Array): Promise<RuneResult<Uint8Array, RuneError>> {")
	g.indent++
	g.line("try { const zlib = await import(\"node:zlib\"); return await new Promise((resolve) => zlib.inflate(data, (error, result) => error ? resolve(runeErr<Uint8Array, RuneError>(runeErrorFrom(error))) : resolve(runeOk<Uint8Array, RuneError>(result)))); }")
	g.line("catch (error) { return runeErr<Uint8Array, RuneError>(runeErrorFrom(error)); }")
	g.indent--
	g.line("}")
	g.line("function runeTextEncoder(): TextEncoder { return new TextEncoder(); }")
	g.line("function runeTextDecoder(): TextDecoder { return new TextDecoder(); }")
	g.line("async function runeCompressGzipText(value: string): Promise<RuneResult<Uint8Array, RuneError>> { return runeCompressGzip(runeTextEncoder().encode(value)); }")
	g.line("async function runeCompressGunzipText(data: Uint8Array): Promise<RuneResult<string, RuneError>> {")
	g.indent++
	g.line("const result = await runeCompressGunzip(data);")
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

func fileUsesBinaryRuntime(file *ir.File) bool {
	return fileUsesType(file, checker.Buffer) ||
		fileUsesType(file, checker.Reader) ||
		fileUsesType(file, checker.Writer)
}

func fileUsesPathRuntime(file *ir.File) bool {
	return fileUsesIntrinsicPrefix(file, "path.")
}

func fileUsesProcessRuntime(file *ir.File) bool {
	return fileUsesIntrinsicPrefix(file, "process.")
}

func fileUsesStringBufferRuntime(file *ir.File) bool {
	return fileUsesType(file, checker.StringBuffer) || fileUsesIntrinsicPrefix(file, "stringbuffer.")
}

func fileUsesIterRuntime(file *ir.File) bool {
	return fileUsesIntrinsicPrefix(file, "iter.")
}

func fileUsesCompressRuntime(file *ir.File) bool {
	return fileUsesIntrinsicPrefix(file, "compress.")
}

func fileUsesNetRuntime(file *ir.File) bool {
	return fileUsesType(file, checker.TCPConnection) ||
		fileUsesType(file, checker.TCPListener) ||
		fileUsesIntrinsicPrefix(file, "net.")
}

func fileUsesType(file *ir.File, typ checker.Type) bool {
	found := false
	check := func(candidate checker.Type) {
		if found || typeContains(candidate, typ) {
			found = true
		}
	}
	for _, fn := range file.Functions {
		check(fn.Return)
		for _, param := range fn.Params {
			check(param.Type)
		}
		ir.WalkExpr(fn.Body, func(expr ir.Expr) {
			check(expr.ResultType())
		})
	}
	for _, test := range file.Tests {
		ir.WalkExpr(test.Body, func(expr ir.Expr) {
			check(expr.ResultType())
		})
	}
	for _, typDecl := range file.Types {
		for _, field := range typDecl.Fields {
			check(field.Type)
		}
		for _, method := range typDecl.Methods {
			check(method.Return)
			for _, param := range method.Params {
				check(param.Type)
			}
			ir.WalkExpr(method.Body, func(expr ir.Expr) {
				check(expr.ResultType())
			})
		}
	}
	return found
}

func typeContains(candidate checker.Type, typ checker.Type) bool {
	if candidate == typ {
		return true
	}
	return strings.Contains(string(candidate), string(typ))
}

func fileUsesSignals(file *ir.File) bool {
	for _, fn := range file.Functions {
		if blockUsesSignals(fn.Body) {
			return true
		}
	}
	for _, typ := range file.Types {
		for _, method := range typ.Methods {
			if blockUsesSignals(method.Body) {
				return true
			}
		}
	}
	return false
}

func fileUsesTaskRuntime(file *ir.File) bool {
	if fileUsesGenericType(file, "Task") || fileUsesFSRuntime(file) || fileUsesCompressRuntime(file) || fileUsesNetRuntime(file) {
		return true
	}
	for _, fn := range file.Functions {
		if fn.Routine || exprUsesAsync(fn.Body) {
			return true
		}
	}
	for _, typ := range file.Types {
		for _, method := range typ.Methods {
			if method.Routine || exprUsesAsync(method.Body) {
				return true
			}
		}
	}
	return false
}

func fileUsesResultRuntime(file *ir.File) bool {
	return fileUsesGenericType(file, "Result") || fileUsesFSRuntime(file) || fileUsesCompressRuntime(file) || fileUsesNetRuntime(file)
}

func fileUsesErrorRuntime(file *ir.File) bool {
	return fileUsesType(file, checker.Error) || fileUsesFSRuntime(file) || fileUsesCompressRuntime(file) || fileUsesNetRuntime(file)
}

func fileUsesGenericType(file *ir.File, base string) bool {
	found := false
	check := func(candidate checker.Type) {
		if found {
			return
		}
		found = typeUsesGeneric(candidate, base)
	}
	for _, fn := range file.Functions {
		check(fn.Return)
		for _, param := range fn.Params {
			check(param.Type)
		}
		ir.WalkExpr(fn.Body, func(expr ir.Expr) {
			check(expr.ResultType())
		})
	}
	for _, typ := range file.Types {
		for _, field := range typ.Fields {
			check(field.Type)
		}
		for _, method := range typ.Methods {
			check(method.Return)
			for _, param := range method.Params {
				check(param.Type)
			}
			ir.WalkExpr(method.Body, func(expr ir.Expr) {
				check(expr.ResultType())
			})
		}
	}
	return found
}

func typeUsesGeneric(candidate checker.Type, base string) bool {
	name := string(candidate)
	return strings.HasPrefix(name, base+"[") || strings.Contains(name, ","+base+"[") || strings.Contains(name, "["+base+"[")
}

func exprUsesAsync(expr ir.Expr) bool {
	found := false
	ir.WalkExpr(expr, func(expr ir.Expr) {
		if call, ok := expr.(*ir.CallExpr); ok && call.Async {
			found = true
		}
		if _, ok := expr.(*ir.ResultUnwrapExpr); ok {
			found = true
		}
	})
	return found
}

func fileUsesFSRuntime(file *ir.File) bool {
	return fileUsesType(file, checker.FileStat) || fileUsesIntrinsicPrefix(file, "fs.")
}

func fileUsesIntrinsicPrefix(file *ir.File, prefix string) bool {
	found := false
	check := func(expr ir.Expr) {
		if found {
			return
		}
		call, ok := expr.(*ir.CallExpr)
		if !ok || file.Stdlib == nil {
			return
		}
		sel, ok := call.Callee.(*ir.SelectorExpr)
		if !ok {
			return
		}
		at, ok := sel.Receiver.(*ir.AtExpr)
		if ok {
			fn, ok := file.Stdlib.Function(at.Name, sel.Name)
			if ok && strings.HasPrefix(fn.Intrinsic, prefix) {
				found = true
			}
			return
		}
		moduleName, receiverName, ok := checker.StdlibReceiverModule(sel.Receiver.ResultType())
		if !ok {
			return
		}
		fn, ok := file.Stdlib.ReceiverFunction(moduleName, receiverName, sel.Name)
		if ok && strings.HasPrefix(fn.Intrinsic, prefix) {
			found = true
		}
	}
	for _, fn := range file.Functions {
		ir.WalkExpr(fn.Body, check)
	}
	for _, typ := range file.Types {
		for _, method := range typ.Methods {
			ir.WalkExpr(method.Body, check)
		}
	}
	return found
}

func blockUsesSignals(expr ir.Expr) bool {
	found := false
	ir.WalkExpr(expr, func(expr ir.Expr) {
		if _, ok := expr.(*ir.WatchExpr); ok {
			found = true
		}
		if _, ok := expr.(*ir.ReactiveLiteral); ok {
			found = true
		}
	})
	if found {
		return true
	}
	if block, ok := expr.(*ir.BlockExpr); ok {
		for _, stmt := range block.Statements {
			if let, ok := stmt.(*ir.LetStmt); ok && let.Signal {
				return true
			}
		}
	}
	return false
}
