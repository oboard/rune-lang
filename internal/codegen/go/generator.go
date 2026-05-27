package gocodegen

import (
	"bytes"
	"fmt"

	goformat "go/format"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	codeusage "github.com/oboard/rune-lang/internal/codegen/usage"
	"github.com/oboard/rune-lang/internal/ir"
)

func Generate(file *ast.File, info *checker.Info) (string, error) {
	return GenerateIR(ir.LowerFile(file, info))
}

func GenerateIR(file *ir.File) (string, error) {
	g := &generator{file: file, imports: map[string]bool{}}
	usage := codeusage.Collect(file)
	for _, imp := range file.GoImports {
		g.imports[imp.Path] = true
	}
	if fileUsesType(usage, checker.BigInt) {
		g.imports["math/big"] = true
	}
	if fileUsesType(usage, checker.Regex) {
		g.imports["regexp"] = true
	}
	if fileUsesBinaryRuntime(usage) {
		g.imports["encoding/binary"] = true
		g.imports["math"] = true
	}
	if fileUsesFSRuntime(usage) {
		g.imports["os"] = true
	}
	if fileUsesPathRuntime(usage) {
		g.imports["path/filepath"] = true
	}
	if fileUsesProcessRuntime(usage) {
		g.imports["os"] = true
		g.imports["runtime"] = true
	}
	if fileUsesStringBufferRuntime(usage) {
		g.imports["strings"] = true
	}
	if fileUsesCompressRuntime(usage) {
		g.imports["bytes"] = true
		g.imports["compress/gzip"] = true
		g.imports["compress/zlib"] = true
		g.imports["github.com/andybalholm/brotli"] = true
		g.imports["github.com/klauspost/compress/zstd"] = true
		g.imports["io"] = true
	}
	if fileUsesNetRuntime(usage) {
		g.imports["net"] = true
	}
	if fileUsesTaskRuntime(usage) {
		g.imports["sync"] = true
	}
	for _, fn := range file.Functions {
		ir.WalkExpr(fn.Body, func(expr ir.Expr) {
			g.collectExprImports(expr)
		})
	}
	for _, typ := range file.Types {
		for _, method := range typ.Methods {
			ir.WalkExpr(method.Body, func(expr ir.Expr) {
				g.collectExprImports(expr)
			})
		}
	}
	g.line("package main")
	g.line("")
	if len(g.imports) > 0 {
		g.line("import (")
		for name := range g.imports {
			g.linef("\t%q", name)
		}
		g.line(")")
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
	if fileUsesType(usage, checker.BigInt) {
		g.bigIntRuntime()
		if len(file.Functions) > 0 || len(file.Types) > 0 {
			g.line("")
		}
	}
	if fileUsesType(usage, checker.Regex) {
		g.regexRuntime()
		if len(file.Functions) > 0 || len(file.Types) > 0 {
			g.line("")
		}
	}
	if fileUsesBinaryRuntime(usage) {
		g.binaryRuntime()
		if len(file.Functions) > 0 || len(file.Types) > 0 {
			g.line("")
		}
	}
	if fileUsesTaskRuntime(usage) {
		g.taskRuntime()
		if len(file.Functions) > 0 || len(file.Types) > 0 {
			g.line("")
		}
	}
	if fileUsesResultRuntime(usage) {
		g.resultRuntime()
		if len(file.Functions) > 0 || len(file.Types) > 0 {
			g.line("")
		}
	}
	if fileUsesErrorRuntime(usage) {
		g.errorRuntime()
		if len(file.Functions) > 0 || len(file.Types) > 0 {
			g.line("")
		}
	}
	if fileUsesPathRuntime(usage) {
		g.pathRuntime()
		if len(file.Functions) > 0 || len(file.Types) > 0 {
			g.line("")
		}
	}
	if fileUsesProcessRuntime(usage) {
		g.processRuntime()
		if len(file.Functions) > 0 || len(file.Types) > 0 {
			g.line("")
		}
	}
	if fileUsesStringBufferRuntime(usage) {
		g.stringBufferRuntime()
		if len(file.Functions) > 0 || len(file.Types) > 0 {
			g.line("")
		}
	}
	if fileUsesIterRuntime(usage) {
		g.iterRuntime()
		if len(file.Functions) > 0 || len(file.Types) > 0 {
			g.line("")
		}
	}
	if fileUsesFSRuntime(usage) {
		g.fsRuntime()
		if len(file.Functions) > 0 || len(file.Types) > 0 {
			g.line("")
		}
	}
	if fileUsesCompressRuntime(usage) {
		g.compressRuntime()
		if len(file.Functions) > 0 || len(file.Types) > 0 {
			g.line("")
		}
	}
	if fileUsesNetRuntime(usage) {
		g.netRuntime()
		if len(file.Functions) > 0 || len(file.Types) > 0 {
			g.line("")
		}
	}
	if fileUsesSignals(usage) {
		g.signalRuntime()
		if len(file.Functions) > 0 {
			g.line("")
		}
	}
	for i, fn := range file.Functions {
		if i > 0 {
			g.line("")
		}
		if err := g.function(fn); err != nil {
			return "", err
		}
	}
	if hasMain(file) {
		if len(file.Functions) > 0 || len(file.Types) > 0 {
			g.line("")
		}
		g.line("func main() {")
		g.indent++
		if mainFn := mainFunction(file); mainFn != nil && mainFn.Routine {
			g.linef("runeAwait(%s())", mangleIdent("main"))
		} else {
			g.linef("%s()", mangleIdent("main"))
		}
		if fileUsesTaskRuntime(usage) {
			g.line("runeWaitAll()")
		}
		g.indent--
		g.line("}")
	}
	if err := g.codegenError(); err != nil {
		return g.buf.String(), err
	}
	formatted, err := goformat.Source(g.buf.Bytes())
	if err != nil {
		return g.buf.String(), err
	}
	return string(formatted), nil
}

func fileUsesType(usage codeusage.Usage, typ checker.Type) bool {
	return usage.HasType(typ)
}

func fileUsesBinaryRuntime(usage codeusage.Usage) bool {
	return fileUsesType(usage, checker.Binary) ||
		fileUsesType(usage, checker.Buffer) ||
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

func (g *generator) bigIntRuntime() {
	g.line("func runeBigInt(src string) *big.Int {")
	g.indent++
	g.line("value, ok := new(big.Int).SetString(src, 10)")
	g.line("if !ok {")
	g.indent++
	g.line("panic(\"invalid BigInt literal\")")
	g.indent--
	g.line("}")
	g.line("return value")
	g.indent--
	g.line("}")
}

func (g *generator) taskRuntime() {
	g.line("type runeUnit struct{}")
	g.line("")
	g.line("type runeTask[T any] <-chan T")
	g.line("")
	g.line("var runeTasks sync.WaitGroup")
	g.line("")
	g.line("func runeGo[T any](work func() T) runeTask[T] {")
	g.indent++
	g.line("runeTasks.Add(1)")
	g.line("ch := make(chan T, 1)")
	g.line("go func() {")
	g.indent++
	g.line("defer runeTasks.Done()")
	g.line("ch <- work()")
	g.indent--
	g.line("}()")
	g.line("return ch")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeWaitAll() {")
	g.indent++
	g.line("runeTasks.Wait()")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeAwait[T any](task runeTask[T]) T {")
	g.indent++
	g.line("return <-task")
	g.indent--
	g.line("}")
}

func (g *generator) resultRuntime() {
	g.line("type runeResult[T any, E any] struct {")
	g.indent++
	g.line("ok bool")
	g.line("value T")
	g.line("err E")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeOk[T any, E any](value T) runeResult[T, E] {")
	g.indent++
	g.line("return runeResult[T, E]{ok: true, value: value}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeErr[T any, E any](err E) runeResult[T, E] {")
	g.indent++
	g.line("return runeResult[T, E]{err: err}")
	g.indent--
	g.line("}")
}

func (g *generator) errorRuntime() {
	g.line("type runeError struct {")
	g.indent++
	g.line("__code int")
	g.line("__message string")
	g.line("__cause *runeError")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeErrorFrom(err error) *runeError {")
	g.indent++
	g.line("if err == nil {")
	g.indent++
	g.line("return nil")
	g.indent--
	g.line("}")
	g.line("return &runeError{__code: 1, __message: err.Error()}")
	g.indent--
	g.line("}")
}

func (g *generator) pathRuntime() {
	g.line("func runePathBasename(path string) string { return filepath.Base(path) }")
	g.line("func runePathDirname(path string) string { return filepath.Dir(path) }")
	g.line("func runePathExtname(path string) string { return filepath.Ext(path) }")
	g.line("func runePathJoin(parts []string) string { return filepath.Join(parts...) }")
	g.line("func runePathNormalize(path string) string { return filepath.Clean(path) }")
	g.line("func runePathResolve(parts []string) string {")
	g.indent++
	g.line("joined := filepath.Join(parts...)")
	g.line("abs, err := filepath.Abs(joined)")
	g.line("if err != nil {")
	g.indent++
	g.line("return joined")
	g.indent--
	g.line("}")
	g.line("return abs")
	g.indent--
	g.line("}")
	g.line("func runePathRelative(from string, to string) string {")
	g.indent++
	g.line("value, err := filepath.Rel(from, to)")
	g.line("if err != nil {")
	g.indent++
	g.line("return to")
	g.indent--
	g.line("}")
	g.line("return value")
	g.indent--
	g.line("}")
	g.line("func runePathIsAbsolute(path string) bool { return filepath.IsAbs(path) }")
}

func (g *generator) processRuntime() {
	g.line("func runeProcessArgv() []string { return append([]string(nil), os.Args...) }")
	g.line("func runeProcessCwd() string {")
	g.indent++
	g.line("cwd, err := os.Getwd()")
	g.line("if err != nil {")
	g.indent++
	g.line("return \"\"")
	g.indent--
	g.line("}")
	g.line("return cwd")
	g.indent--
	g.line("}")
	g.line("func runeProcessEnv(name string) any {")
	g.indent++
	g.line("value, ok := os.LookupEnv(name)")
	g.line("if !ok {")
	g.indent++
	g.line("return any(nil)")
	g.indent--
	g.line("}")
	g.line("return value")
	g.indent--
	g.line("}")
	g.line("func runeProcessExit(code int) struct{} { os.Exit(code); return struct{}{} }")
	g.line("func runeProcessPlatform() string { return runtime.GOOS }")
}

func (g *generator) stringBufferRuntime() {
	g.line("type runeStringBuffer struct {")
	g.indent++
	g.line("builder strings.Builder")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func newRuneStringBuffer() *runeStringBuffer { return &runeStringBuffer{} }")
	g.line("func newRuneStringBufferFromString(value string) *runeStringBuffer {")
	g.indent++
	g.line("b := &runeStringBuffer{}")
	g.line("b.builder.WriteString(value)")
	g.line("return b")
	g.indent--
	g.line("}")
	g.line("func (b *runeStringBuffer) Length() int { return len([]rune(b.builder.String())) }")
	g.line("func (b *runeStringBuffer) Clear() { b.builder.Reset() }")
	g.line("func (b *runeStringBuffer) Append(value string) *runeStringBuffer { b.builder.WriteString(value); return b }")
	g.line("func (b *runeStringBuffer) AppendLine(value string) *runeStringBuffer { b.builder.WriteString(value); b.builder.WriteByte('\\n'); return b }")
	g.line("func (b *runeStringBuffer) ToString() string { return b.builder.String() }")
}

func (g *generator) iterRuntime() {
	g.line("type runeIter[T any] struct {")
	g.indent++
	g.line("__next func() struct{ F0 T; F1 bool }")
	g.indent--
	g.line("}")
}

func (g *generator) fsRuntime() {
	g.line("type runeFileStat struct {")
	g.indent++
	g.line("__size int")
	g.line("__isFile bool")
	g.line("__isDirectory bool")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeReadFile(path string) runeTask[runeResult[[]byte, *runeError]] { return runeFsReadFile(path) }")
	g.line("")
	g.line("func runeFsReadFile(path string) runeTask[runeResult[[]byte, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[[]byte, *runeError] {")
	g.indent++
	g.line("data, err := os.ReadFile(path)")
	g.line("if err != nil {")
	g.indent++
	g.line("return runeErr[[]byte, *runeError](runeErrorFrom(err))")
	g.indent--
	g.line("}")
	g.line("return runeOk[[]byte, *runeError](data)")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeFsReadFileText(path string) runeTask[runeResult[string, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[string, *runeError] {")
	g.indent++
	g.line("data, err := os.ReadFile(path)")
	g.line("if err != nil {")
	g.indent++
	g.line("return runeErr[string, *runeError](runeErrorFrom(err))")
	g.indent--
	g.line("}")
	g.line("return runeOk[string, *runeError](string(data))")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeFsWriteFile(path string, data []byte) runeTask[runeResult[struct{}, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[struct{}, *runeError] {")
	g.indent++
	g.line("if err := os.WriteFile(path, data, 0644); err != nil {")
	g.indent++
	g.line("return runeErr[struct{}, *runeError](runeErrorFrom(err))")
	g.indent--
	g.line("}")
	g.line("return runeOk[struct{}, *runeError](struct{}{})")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeFsWriteFileText(path string, data string) runeTask[runeResult[struct{}, *runeError]] {")
	g.indent++
	g.line("return runeFsWriteFile(path, []byte(data))")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeFsExists(path string) runeTask[runeResult[bool, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[bool, *runeError] {")
	g.indent++
	g.line("_, err := os.Stat(path)")
	g.line("if err == nil {")
	g.indent++
	g.line("return runeOk[bool, *runeError](true)")
	g.indent--
	g.line("}")
	g.line("if os.IsNotExist(err) {")
	g.indent++
	g.line("return runeOk[bool, *runeError](false)")
	g.indent--
	g.line("}")
	g.line("return runeErr[bool, *runeError](runeErrorFrom(err))")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeFsReaddir(path string) runeTask[runeResult[[]string, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[[]string, *runeError] {")
	g.indent++
	g.line("entries, err := os.ReadDir(path)")
	g.line("if err != nil {")
	g.indent++
	g.line("return runeErr[[]string, *runeError](runeErrorFrom(err))")
	g.indent--
	g.line("}")
	g.line("names := make([]string, 0, len(entries))")
	g.line("for _, entry := range entries {")
	g.indent++
	g.line("names = append(names, entry.Name())")
	g.indent--
	g.line("}")
	g.line("return runeOk[[]string, *runeError](names)")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeFsMkdir(path string) runeTask[runeResult[struct{}, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[struct{}, *runeError] {")
	g.indent++
	g.line("if err := os.Mkdir(path, 0755); err != nil {")
	g.indent++
	g.line("return runeErr[struct{}, *runeError](runeErrorFrom(err))")
	g.indent--
	g.line("}")
	g.line("return runeOk[struct{}, *runeError](struct{}{})")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeFsRemove(path string) runeTask[runeResult[struct{}, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[struct{}, *runeError] {")
	g.indent++
	g.line("if err := os.Remove(path); err != nil {")
	g.indent++
	g.line("return runeErr[struct{}, *runeError](runeErrorFrom(err))")
	g.indent--
	g.line("}")
	g.line("return runeOk[struct{}, *runeError](struct{}{})")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeFsStat(path string) runeTask[runeResult[*runeFileStat, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[*runeFileStat, *runeError] {")
	g.indent++
	g.line("info, err := os.Stat(path)")
	g.line("if err != nil {")
	g.indent++
	g.line("return runeErr[*runeFileStat, *runeError](runeErrorFrom(err))")
	g.indent--
	g.line("}")
	g.line("stat := &runeFileStat{__size: int(info.Size()), __isFile: info.Mode().IsRegular(), __isDirectory: info.IsDir()}")
	g.line("return runeOk[*runeFileStat, *runeError](stat)")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
}

func (g *generator) compressRuntime() {
	g.line("func runeCompressWrite(data []byte, writer io.WriteCloser, out *bytes.Buffer) runeResult[[]byte, *runeError] {")
	g.indent++
	g.line("if _, err := writer.Write(data); err != nil {")
	g.indent++
	g.line("_ = writer.Close()")
	g.line("return runeErr[[]byte, *runeError](runeErrorFrom(err))")
	g.indent--
	g.line("}")
	g.line("if err := writer.Close(); err != nil {")
	g.indent++
	g.line("return runeErr[[]byte, *runeError](runeErrorFrom(err))")
	g.indent--
	g.line("}")
	g.line("return runeOk[[]byte, *runeError](out.Bytes())")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeCompressRead(reader io.Reader) runeResult[[]byte, *runeError] {")
	g.indent++
	g.line("out, err := io.ReadAll(reader)")
	g.line("if err != nil {")
	g.indent++
	g.line("return runeErr[[]byte, *runeError](runeErrorFrom(err))")
	g.indent--
	g.line("}")
	g.line("return runeOk[[]byte, *runeError](out)")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeCompressGzip(data []byte) runeTask[runeResult[[]byte, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[[]byte, *runeError] {")
	g.indent++
	g.line("var out bytes.Buffer")
	g.line("writer := gzip.NewWriter(&out)")
	g.line("return runeCompressWrite(data, writer, &out)")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeCompressGunzip(data []byte) runeTask[runeResult[[]byte, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[[]byte, *runeError] {")
	g.indent++
	g.line("reader, err := gzip.NewReader(bytes.NewReader(data))")
	g.line("if err != nil {")
	g.indent++
	g.line("return runeErr[[]byte, *runeError](runeErrorFrom(err))")
	g.indent--
	g.line("}")
	g.line("defer reader.Close()")
	g.line("return runeCompressRead(reader)")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeCompressDeflate(data []byte) runeTask[runeResult[[]byte, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[[]byte, *runeError] {")
	g.indent++
	g.line("var out bytes.Buffer")
	g.line("writer := zlib.NewWriter(&out)")
	g.line("return runeCompressWrite(data, writer, &out)")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeCompressInflate(data []byte) runeTask[runeResult[[]byte, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[[]byte, *runeError] {")
	g.indent++
	g.line("reader, err := zlib.NewReader(bytes.NewReader(data))")
	g.line("if err != nil {")
	g.indent++
	g.line("return runeErr[[]byte, *runeError](runeErrorFrom(err))")
	g.indent--
	g.line("}")
	g.line("defer reader.Close()")
	g.line("return runeCompressRead(reader)")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeCompressBrotli(data []byte) runeTask[runeResult[[]byte, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[[]byte, *runeError] {")
	g.indent++
	g.line("var out bytes.Buffer")
	g.line("writer := brotli.NewWriter(&out)")
	g.line("return runeCompressWrite(data, writer, &out)")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeCompressUnbrotli(data []byte) runeTask[runeResult[[]byte, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[[]byte, *runeError] {")
	g.indent++
	g.line("reader := brotli.NewReader(bytes.NewReader(data))")
	g.line("return runeCompressRead(reader)")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeCompressZstd(data []byte) runeTask[runeResult[[]byte, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[[]byte, *runeError] {")
	g.indent++
	g.line("var out bytes.Buffer")
	g.line("writer, err := zstd.NewWriter(&out)")
	g.line("if err != nil {")
	g.indent++
	g.line("return runeErr[[]byte, *runeError](runeErrorFrom(err))")
	g.indent--
	g.line("}")
	g.line("return runeCompressWrite(data, writer, &out)")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeCompressUnzstd(data []byte) runeTask[runeResult[[]byte, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[[]byte, *runeError] {")
	g.indent++
	g.line("reader, err := zstd.NewReader(bytes.NewReader(data))")
	g.line("if err != nil {")
	g.indent++
	g.line("return runeErr[[]byte, *runeError](runeErrorFrom(err))")
	g.indent--
	g.line("}")
	g.line("defer reader.Close()")
	g.line("return runeCompressRead(reader)")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeCompressGzipText(value string) runeTask[runeResult[[]byte, *runeError]] { return runeCompressGzip([]byte(value)) }")
	g.line("func runeCompressGunzipText(data []byte) runeTask[runeResult[string, *runeError]] {")
	g.indent++
	g.line("return runeCompressDecodeText(data, runeCompressGunzip)")
	g.indent--
	g.line("}")
	g.line("func runeCompressBrotliText(value string) runeTask[runeResult[[]byte, *runeError]] { return runeCompressBrotli([]byte(value)) }")
	g.line("func runeCompressUnbrotliText(data []byte) runeTask[runeResult[string, *runeError]] {")
	g.indent++
	g.line("return runeCompressDecodeText(data, runeCompressUnbrotli)")
	g.indent--
	g.line("}")
	g.line("func runeCompressZstdText(value string) runeTask[runeResult[[]byte, *runeError]] { return runeCompressZstd([]byte(value)) }")
	g.line("func runeCompressUnzstdText(data []byte) runeTask[runeResult[string, *runeError]] {")
	g.indent++
	g.line("return runeCompressDecodeText(data, runeCompressUnzstd)")
	g.indent--
	g.line("}")
	g.line("func runeCompressDecodeText(data []byte, decode func([]byte) runeTask[runeResult[[]byte, *runeError]]) runeTask[runeResult[string, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[string, *runeError] {")
	g.indent++
	g.line("result := runeAwait(decode(data))")
	g.line("if !result.ok {")
	g.indent++
	g.line("return runeErr[string, *runeError](result.err)")
	g.indent--
	g.line("}")
	g.line("return runeOk[string, *runeError](string(result.value))")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
}

func (g *generator) netRuntime() {
	g.line("type runeTCPConnection struct { conn net.Conn }")
	g.line("type runeTCPListener struct { listener net.Listener }")
	g.line("")
	g.line("func runeNetConnect(address string) runeTask[runeResult[*runeTCPConnection, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[*runeTCPConnection, *runeError] {")
	g.indent++
	g.line("conn, err := net.Dial(\"tcp\", address)")
	g.line("if err != nil {")
	g.indent++
	g.line("return runeErr[*runeTCPConnection, *runeError](runeErrorFrom(err))")
	g.indent--
	g.line("}")
	g.line("return runeOk[*runeTCPConnection, *runeError](&runeTCPConnection{conn: conn})")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeNetListen(address string) runeTask[runeResult[*runeTCPListener, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[*runeTCPListener, *runeError] {")
	g.indent++
	g.line("listener, err := net.Listen(\"tcp\", address)")
	g.line("if err != nil {")
	g.indent++
	g.line("return runeErr[*runeTCPListener, *runeError](runeErrorFrom(err))")
	g.indent--
	g.line("}")
	g.line("return runeOk[*runeTCPListener, *runeError](&runeTCPListener{listener: listener})")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func (c *runeTCPConnection) Read(length int) runeTask[runeResult[[]byte, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[[]byte, *runeError] {")
	g.indent++
	g.line("if length < 0 {")
	g.indent++
	g.line("return runeErr[[]byte, *runeError](&runeError{__code: 1, __message: \"net read length out of range\"})")
	g.indent--
	g.line("}")
	g.line("buf := make([]byte, length)")
	g.line("n, err := c.conn.Read(buf)")
	g.line("if err != nil {")
	g.indent++
	g.line("return runeErr[[]byte, *runeError](runeErrorFrom(err))")
	g.indent--
	g.line("}")
	g.line("return runeOk[[]byte, *runeError](buf[:n])")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
	g.line("func (c *runeTCPConnection) Write(data []byte) runeTask[runeResult[int, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[int, *runeError] {")
	g.indent++
	g.line("n, err := c.conn.Write(data)")
	g.line("if err != nil {")
	g.indent++
	g.line("return runeErr[int, *runeError](runeErrorFrom(err))")
	g.indent--
	g.line("}")
	g.line("return runeOk[int, *runeError](n)")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
	g.line("func (c *runeTCPConnection) Close() runeTask[runeResult[struct{}, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[struct{}, *runeError] {")
	g.indent++
	g.line("if err := c.conn.Close(); err != nil {")
	g.indent++
	g.line("return runeErr[struct{}, *runeError](runeErrorFrom(err))")
	g.indent--
	g.line("}")
	g.line("return runeOk[struct{}, *runeError](struct{}{})")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
	g.line("func (l *runeTCPListener) Address() string { return l.listener.Addr().String() }")
	g.line("func (l *runeTCPListener) Accept() runeTask[runeResult[*runeTCPConnection, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[*runeTCPConnection, *runeError] {")
	g.indent++
	g.line("conn, err := l.listener.Accept()")
	g.line("if err != nil {")
	g.indent++
	g.line("return runeErr[*runeTCPConnection, *runeError](runeErrorFrom(err))")
	g.indent--
	g.line("}")
	g.line("return runeOk[*runeTCPConnection, *runeError](&runeTCPConnection{conn: conn})")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
	g.line("func (l *runeTCPListener) Close() runeTask[runeResult[struct{}, *runeError]] {")
	g.indent++
	g.line("return runeGo(func() runeResult[struct{}, *runeError] {")
	g.indent++
	g.line("if err := l.listener.Close(); err != nil {")
	g.indent++
	g.line("return runeErr[struct{}, *runeError](runeErrorFrom(err))")
	g.indent--
	g.line("}")
	g.line("return runeOk[struct{}, *runeError](struct{}{})")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
}

func fileUsesSignals(usage codeusage.Usage) bool {
	return usage.Signal
}

type generator struct {
	buf       bytes.Buffer
	file      *ir.File
	imports   map[string]bool
	indent    int
	temp      int
	errors    []error
	thisNames []string
	signals   []map[string]checker.Type
}

func (g *generator) nextTemp(prefix string) string {
	g.temp++
	return fmt.Sprintf("%s%d", mangleIdent(prefix), g.temp)
}
