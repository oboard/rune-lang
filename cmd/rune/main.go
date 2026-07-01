package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/oboard/rune-lang/internal/checker"
	gocodegen "github.com/oboard/rune-lang/internal/codegen/go"
	tscodegen "github.com/oboard/rune-lang/internal/codegen/typescript"
	"github.com/oboard/rune-lang/internal/compiler"
	runefmt "github.com/oboard/rune-lang/internal/format"
	"github.com/oboard/rune-lang/internal/ir"
	"github.com/oboard/rune-lang/internal/lsp"
	"github.com/oboard/rune-lang/internal/parser"
	"github.com/oboard/rune-lang/internal/repl"
	"github.com/oboard/rune-lang/internal/stdlib"
	"github.com/oboard/rune-lang/internal/tester"
)

func main() {
	if err := runRuneCLI(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		if code, ok := exitCode(err); ok {
			os.Exit(code)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func exitCode(err error) (int, bool) {
	var exit interface{ ExitCode() int }
	if errors.As(err, &exit) {
		return exit.ExitCode(), true
	}
	return 0, false
}

func runRuneCLI(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	invocation := __parseCli(args)
	if invocation.__help {
		fmt.Fprint(stdout, invocation.__helpText)
		return nil
	}
	if !invocation.__ok {
		for _, message := range invocation.__errors {
			fmt.Fprintln(stderr, message)
		}
		return fmt.Errorf("rune command failed")
	}
	return executeInvocation(invocation, invocation.__backendExplicit, stdin, stdout, stderr)
}

func executeInvocation(invocation __RuneCliInvocation, explicitBackend bool, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	switch invocation.__command {
	case "run":
		entry, diags, err := resolveRunEntry(invocation.__path)
		if len(diags) > 0 {
			printDiagnostics(invocation.__path, diags)
			return fmt.Errorf("run failed")
		}
		if err != nil {
			return err
		}
		runBackend := selectRunBackend(entry, invocation.__backend, explicitBackend)
		if err := validateBackend(runBackend); err != nil {
			return err
		}
		return runEntry(entry, runBackend, invocation.__target, invocation.__runArgs, stdin, stdout, stderr)
	case "build":
		if invocation.__backend == "mbt" {
			return buildMoonBit(invocation.__path, invocation.__target, invocation.__output)
		}
		return buildGo(invocation.__path, invocation.__target, invocation.__output, stdin, stdout, stderr)
	case "go":
		return emitGo(invocation.__path, invocation.__output, stdout)
	case "ts":
		return emitTypeScript(invocation.__path, invocation.__output, stdout)
	case "mbt":
		return emitMoonBit(invocation.__path, invocation.__output, stdout)
	case "check":
		return checkTarget(invocation.__path, stdout)
	case "test":
		if explicitBackend {
			_, err := tester.RunWithBackend(invocation.__path, invocation.__pattern, invocation.__backend, stdout)
			return err
		}
		_, err := tester.Run(invocation.__path, invocation.__pattern, stdout)
		return err
	case "fmt":
		return formatTarget(invocation.__path, invocation.__checkOnly, invocation.__stdout, stdout)
	case "repl":
		return repl.Serve(stdin, stdout)
	case "lsp":
		return lsp.Serve(stdin, stdout)
	default:
		return fmt.Errorf("unknown command %s", invocation.__command)
	}
}

func runEntry(entry string, runBackend string, runTarget string, programArgs []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	switch runBackend {
	case "go":
		exe, cleanup, err := compileGoExecutableToTemp(entry, stdout, stderr)
		if err != nil {
			return err
		}
		defer cleanup()
		run := exec.Command(exe, programArgs...)
		run.Stdout = stdout
		run.Stderr = stderr
		run.Stdin = stdin
		return run.Run()
	case "ts":
		tsFile, runDir, cleanup, err := compileTypeScriptToTemp(entry)
		if err != nil {
			return err
		}
		defer cleanup()
		run, err := typeScriptRuntimeCommand(tsFile, programArgs)
		if err != nil {
			return err
		}
		run.Stdout = stdout
		run.Stderr = stderr
		run.Stdin = stdin
		run.Dir = runDir
		return run.Run()
	case "mbt":
		if runTarget == "" {
			runTarget = "native"
		}
		if err := validateMoonBitTarget(runTarget); err != nil {
			return err
		}
		runDir, cleanup, err := compileMoonBitProjectToTemp(entry)
		if err != nil {
			return err
		}
		defer cleanup()
		args := []string{"run", "--target", runTarget, "."}
		if len(programArgs) > 0 {
			args = append(args, "--")
			args = append(args, programArgs...)
		}
		run := exec.Command("moon", args...)
		run.Stdout = stdout
		run.Stderr = stderr
		run.Stdin = stdin
		run.Dir = runDir
		return run.Run()
	default:
		return validateBackend(runBackend)
	}
}

func selectRunBackend(entry string, requested string, explicit bool) string {
	if requested != "go" || explicit {
		return requested
	}
	prog, _ := compiler.AnalyzeFile(entry)
	if prog != nil && len(prog.File.TSImports) > 0 {
		return "ts"
	}
	return requested
}

func buildGo(path string, target string, output string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	goFile, cleanup, err := compileGoToTemp(path)
	if err != nil {
		return err
	}
	defer cleanup()

	env, err := buildEnv(target)
	if err != nil {
		return err
	}

	out := output
	if out == "" {
		out = defaultBinaryName(path)
	}
	build := exec.Command("go", "build", "-o", out, goFile)
	build.Env = env
	build.Stdout = stdout
	build.Stderr = stderr
	build.Stdin = stdin
	return build.Run()
}

func emitGo(path string, output string, stdout io.Writer) error {
	src, diags := compiler.GenerateGoFile(path)
	if len(diags) > 0 {
		printDiagnostics(path, diags)
		return fmt.Errorf("compile failed")
	}
	if output == "" {
		fmt.Fprint(stdout, src)
		return nil
	}
	return os.WriteFile(output, []byte(src), 0o644)
}

func emitTypeScript(path string, output string, stdout io.Writer) error {
	src, diags := compiler.GenerateTypeScriptFile(path)
	if len(diags) > 0 {
		printDiagnostics(path, diags)
		return fmt.Errorf("compile failed")
	}
	if output == "" {
		fmt.Fprint(stdout, src)
		return nil
	}
	return os.WriteFile(output, []byte(src), 0o644)
}

func emitMoonBit(path string, output string, stdout io.Writer) error {
	src, diags := compiler.GenerateMoonBitFile(path)
	if len(diags) > 0 {
		printDiagnostics(path, diags)
		return fmt.Errorf("compile failed")
	}
	if output == "" {
		fmt.Fprint(stdout, src)
		return nil
	}
	return os.WriteFile(output, []byte(src), 0o644)
}

func checkTarget(path string, out io.Writer) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return checkFile(path, out)
	}
	if isStdlibRoot(path) {
		if _, err := stdlib.Load(path); err != nil {
			return err
		}
		fmt.Fprintf(out, "ok %s\n", path)
		return nil
	}
	files, err := runeFiles(path)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no Rune files found in %s", path)
	}
	failed := false
	for _, file := range files {
		_, diags := compiler.AnalyzeFileWithWarnings(file)
		if len(diags) > 0 {
			printDiagnostics(file, diags)
		}
		if hasErrorDiagnostics(diags) {
			failed = true
		}
	}
	if failed {
		return fmt.Errorf("check failed")
	}
	fmt.Fprintf(out, "ok %s\n", path)
	return nil
}

func checkFile(path string, out io.Writer) error {
	_, diags := compiler.AnalyzeFileWithWarnings(path)
	if len(diags) > 0 {
		printDiagnostics(path, diags)
	}
	if hasErrorDiagnostics(diags) {
		return fmt.Errorf("check failed")
	}
	fmt.Fprintf(out, "ok %s\n", path)
	return nil
}

func formatTarget(path string, checkOnly bool, stdout bool, out io.Writer) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return formatFile(path, checkOnly, stdout, out)
	}
	if stdout {
		return fmt.Errorf("fmt --stdout only supports a single file")
	}
	files, err := runeFiles(path)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no Rune files found in %s", path)
	}
	for _, file := range files {
		if isStdlibPrelude(file) {
			continue
		}
		if err := formatFile(file, checkOnly, false, out); err != nil {
			return err
		}
	}
	return nil
}

func formatFile(path string, checkOnly bool, stdout bool, out io.Writer) error {
	original, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	file, errs := parser.Parse(string(original))
	if len(errs) > 0 {
		printDiagnostics(path, parseDiagnostics(errs))
		return fmt.Errorf("format failed")
	}
	formatted := runefmt.Source(file, string(original))
	if stdout {
		fmt.Fprint(out, formatted)
		return nil
	}
	if string(original) == formatted {
		return nil
	}
	if checkOnly {
		return fmt.Errorf("%s is not formatted", path)
	}
	return os.WriteFile(path, []byte(formatted), 0o644)
}

func resolveRunEntry(path string) (string, []compiler.Diagnostic, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", nil, err
	}
	if !info.IsDir() {
		return path, nil, nil
	}
	files, err := runeFiles(path)
	if err != nil {
		return "", nil, err
	}
	if len(files) == 0 {
		return "", nil, fmt.Errorf("no Rune files found in %s", path)
	}
	return findMainFile(path, files)
}

type mainDecl struct {
	path   string
	line   int
	column int
}

func findMainFile(root string, files []string) (string, []compiler.Diagnostic, error) {
	var mains []mainDecl
	var diags []compiler.Diagnostic
	for _, filePath := range files {
		if isStdlibPrelude(filePath) {
			continue
		}
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", nil, err
		}
		file, errs := parser.Parse(string(data))
		if len(errs) > 0 {
			diags = append(diags, parseDiagnosticsForPath(filePath, errs)...)
			continue
		}
		for _, fn := range file.Functions {
			if fn.Name != "main" || fn.ReceiverType != "" {
				continue
			}
			mains = append(mains, mainDecl{
				path:   filePath,
				line:   fn.NamePos.Line,
				column: fn.NamePos.Column,
			})
		}
	}
	if len(diags) > 0 {
		return "", diags, fmt.Errorf("run failed")
	}
	if len(mains) == 0 {
		return "", nil, fmt.Errorf("no main function found in %s", root)
	}
	if len(mains) > 1 {
		return "", nil, fmt.Errorf("multiple main functions found in %s:\n  %s", root, strings.Join(formatMainDecls(mains), "\n  "))
	}
	return mains[0].path, nil, nil
}

func formatMainDecls(mains []mainDecl) []string {
	decls := make([]string, 0, len(mains))
	for _, main := range mains {
		if main.line > 0 {
			decls = append(decls, fmt.Sprintf("%s:%d:%d", main.path, main.line, main.column))
			continue
		}
		decls = append(decls, main.path)
	}
	sort.Strings(decls)
	return decls
}

func runeFiles(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{path}, nil
	}
	var files []string
	err = filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".rn" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func isStdlibRoot(path string) bool {
	if _, err := os.Stat(filepath.Join(path, "prelude.rn")); err != nil {
		return false
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		modulePath := filepath.Join(path, entry.Name(), entry.Name()+".rn")
		if _, err := os.Stat(modulePath); err == nil {
			return true
		}
	}
	return false
}

func isStdlibPrelude(path string) bool {
	if filepath.Base(path) != "prelude.rn" {
		return false
	}
	return isStdlibRoot(filepath.Dir(path))
}

func compileGoToTemp(path string) (string, func(), error) {
	prog, diags := compiler.AnalyzeFile(path)
	if len(diags) > 0 {
		printDiagnostics(path, diags)
		return "", func() {}, fmt.Errorf("compile failed")
	}
	src, err := gocodegen.GenerateIR(prog.IR)
	if err != nil {
		return "", func() {}, err
	}
	dir, err := os.MkdirTemp("", "rune-*")
	if err != nil {
		return "", func() {}, err
	}
	cleanup := func() {
		_ = os.RemoveAll(dir)
	}
	goFile := filepath.Join(dir, "main.go")
	if err := os.WriteFile(goFile, []byte(src), 0o644); err != nil {
		cleanup()
		return "", func() {}, err
	}
	return goFile, cleanup, nil
}

func compileGoExecutableToTemp(path string, stdout io.Writer, stderr io.Writer) (string, func(), error) {
	goFile, cleanup, err := compileGoToTemp(path)
	if err != nil {
		return "", cleanup, err
	}
	exe := filepath.Join(filepath.Dir(goFile), "main")
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	build := exec.Command("go", "build", "-o", exe, goFile)
	build.Stdout = stdout
	build.Stderr = stderr
	if err := build.Run(); err != nil {
		cleanup()
		return "", func() {}, err
	}
	return exe, cleanup, nil
}

func compileTypeScriptToTemp(path string) (string, string, func(), error) {
	prog, diags := compiler.AnalyzeFile(path)
	if len(diags) > 0 {
		printDiagnostics(path, diags)
		return "", "", func() {}, fmt.Errorf("compile failed")
	}
	src, err := tscodegen.GenerateIR(typeScriptRuntimeIR(prog.IR))
	if err != nil {
		return "", "", func() {}, err
	}
	runDir, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		return "", "", func() {}, err
	}
	file, err := os.CreateTemp(runDir, ".rune-*.ts")
	if err != nil {
		return "", "", func() {}, err
	}
	tsFile := file.Name()
	cleanup := func() {
		_ = os.Remove(tsFile)
	}
	src += "\nif (typeof __main === \"function\") {\n  const __runeMainResult = __main();\n  if (__runeMainResult && typeof __runeMainResult.then === \"function\") {\n    await __runeMainResult;\n  }\n}\nif (typeof runeWaitAll === \"function\") {\n  await runeWaitAll();\n}\n"
	if _, err := file.WriteString(src); err != nil {
		_ = file.Close()
		cleanup()
		return "", "", func() {}, err
	}
	if err := file.Close(); err != nil {
		cleanup()
		return "", "", func() {}, err
	}
	return tsFile, runDir, cleanup, nil
}

func compileMoonBitProjectToTemp(path string) (string, func(), error) {
	src, diags := compiler.GenerateMoonBitFile(path)
	if len(diags) > 0 {
		printDiagnostics(path, diags)
		return "", func() {}, fmt.Errorf("compile failed")
	}
	dir, err := os.MkdirTemp("", "rune-mbt-*")
	if err != nil {
		return "", func() {}, err
	}
	cleanup := func() {
		_ = os.RemoveAll(dir)
	}
	files := map[string]string{
		"main.mbt": src,
		"moon.mod": moonBitMod(src),
		"moon.pkg": moonBitPkg(src),
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			cleanup()
			return "", func() {}, err
		}
	}
	return dir, cleanup, nil
}

func moonBitMod(src string) string {
	if !moonBitUsesAsync(src) {
		return "name = \"oboard/rune_mbt\"\n"
	}
	var b strings.Builder
	b.WriteString("name = \"oboard/rune_mbt\"\n\nimport {\n")
	b.WriteString("  \"moonbitlang/async@0.19.4\",\n")
	if moonBitUsesBikallemCompress(src) {
		b.WriteString("  \"bikallem/compress@0.3.4\",\n")
	}
	b.WriteString("}\n\npreferred_target = \"native\"\n")
	return b.String()
}

func moonBitPkg(src string) string {
	imports := []string{
		"moonbitlang/core/env",
		"moonbitlang/core/string",
	}
	if moonBitUsesBigInt(src) {
		imports = append(imports, "moonbitlang/core/bigint")
	}
	if moonBitUsesJSON(src) {
		imports = append(imports, "moonbitlang/core/json")
	}
	if moonBitUsesAsync(src) {
		imports = append(imports, "moonbitlang/async")
	}
	if moonBitUsesAsyncFS(src) {
		imports = append(imports, "moonbitlang/async/fs")
	}
	if moonBitUsesAsyncIO(src) {
		imports = append(imports, "moonbitlang/async/io")
	}
	if moonBitUsesAsyncGzip(src) {
		imports = append(imports, "moonbitlang/async/gzip")
	}
	if moonBitUsesBikallemCompress(src) {
		imports = append(imports,
			"moonbitlang/core/encoding/utf8",
			"bikallem/compress/flate",
			"bikallem/compress/brotli",
			"bikallem/compress/zstd",
		)
	}
	var b strings.Builder
	b.WriteString("import {\n")
	for i, imp := range imports {
		b.WriteString("  \"")
		b.WriteString(imp)
		b.WriteString("\"")
		if i < len(imports)-1 {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
	b.WriteString("}\n\n")
	b.WriteString("warnings = \"-1-7-23-67\"\n\n")
	if moonBitUsesAsyncFS(src) || moonBitUsesAsyncGzip(src) {
		b.WriteString("supported_targets = \"+native\"\n\n")
	}
	b.WriteString("options(\n  \"is-main\": true,\n)\n")
	return b.String()
}

func moonBitUsesAsync(src string) bool {
	return strings.Contains(src, "async fn") ||
		moonBitUsesAsyncFS(src) ||
		moonBitUsesAsyncIO(src) ||
		moonBitUsesAsyncGzip(src)
}

func moonBitUsesAsyncFS(src string) bool {
	return strings.Contains(src, "@fs.")
}

func moonBitUsesAsyncIO(src string) bool {
	return strings.Contains(src, "@io.")
}

func moonBitUsesAsyncGzip(src string) bool {
	return strings.Contains(src, "@gzip.")
}

func moonBitUsesBigInt(src string) bool {
	return strings.Contains(src, "BigInt")
}

func moonBitUsesJSON(src string) bool {
	return strings.Contains(src, "@json.") ||
		strings.Contains(src, "Json::") ||
		strings.Contains(src, "Object(") ||
		strings.Contains(src, "Array(")
}

func moonBitUsesBikallemCompress(src string) bool {
	return strings.Contains(src, "@flate.") ||
		strings.Contains(src, "@brotli.") ||
		strings.Contains(src, "@zstd.") ||
		strings.Contains(src, "@utf8.")
}

func buildMoonBit(path string, target string, output string) error {
	if target == "" {
		target = "native"
	}
	if err := validateMoonBitTarget(target); err != nil {
		return err
	}
	if output != "" && !filepath.IsAbs(output) {
		abs, err := filepath.Abs(output)
		if err != nil {
			return err
		}
		output = abs
	}
	dir, cleanup, err := compileMoonBitProjectToTemp(path)
	if err != nil {
		return err
	}
	defer cleanup()
	args := []string{"build", "--target", target, "--release"}
	if output != "" {
		args = append(args, "--target-dir", output)
	}
	build := exec.Command("moon", args...)
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	build.Stdin = os.Stdin
	build.Dir = dir
	return build.Run()
}

func validateMoonBitTarget(target string) error {
	switch target {
	case "wasm", "wasm-gc", "js", "native", "llvm", "all":
		return nil
	default:
		return fmt.Errorf("invalid MoonBit target %q, expected wasm, wasm-gc, js, native, llvm, or all", target)
	}
}

func typeScriptRuntimeIR(file *ir.File) *ir.File {
	out := *file
	out.TSImports = append([]ir.TSImport(nil), file.TSImports...)
	for i := range out.TSImports {
		out.TSImports[i].Specifier = typeScriptRuntimeSpecifier(out.TSImports[i].Specifier)
	}
	return &out
}

func typeScriptRuntimeSpecifier(specifier string) string {
	if specifier == "" {
		return specifier
	}
	if parsed, err := url.Parse(specifier); err == nil && parsed.Scheme != "" {
		return specifier
	}
	if filepath.IsAbs(specifier) {
		return (&url.URL{Scheme: "file", Path: specifier}).String()
	}
	if strings.HasPrefix(specifier, "./") || strings.HasPrefix(specifier, "../") {
		return specifier
	}
	return "./" + specifier
}

func typeScriptRuntimeCommand(path string, args []string) (*exec.Cmd, error) {
	return typeScriptRuntimeCommandWithLookPath(path, args, exec.LookPath)
}

func typeScriptRuntimeCommandWithLookPath(path string, args []string, lookPath func(string) (string, error)) (*exec.Cmd, error) {
	if runner, err := lookPath("bun"); err == nil {
		return exec.Command(runner, append([]string{path}, args...)...), nil
	}
	if runner, err := lookPath("deno"); err == nil {
		return exec.Command(runner, append([]string{"run", path}, args...)...), nil
	}
	if runner, err := lookPath("node"); err == nil {
		return exec.Command(runner, append([]string{path}, args...)...), nil
	}
	if runner, err := lookPath("ts-node"); err == nil {
		return exec.Command(runner, append([]string{"--esm", path}, args...)...), nil
	}
	return nil, fmt.Errorf("TypeScript backend requires bun, deno, node, or ts-node")
}

func buildEnv(target string) ([]string, error) {
	env := os.Environ()
	if target == "" {
		return env, nil
	}
	goos, goarch, err := parseTarget(target)
	if err != nil {
		return nil, err
	}
	return append(env, "GOOS="+goos, "GOARCH="+goarch), nil
}

func parseTarget(target string) (string, string, error) {
	parts := strings.Split(target, "-")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid target %q, expected GOOS-GOARCH", target)
	}
	return parts[0], parts[1], nil
}

func validateBackend(value string) error {
	switch value {
	case "go", "ts", "mbt":
		return nil
	default:
		return fmt.Errorf("invalid backend %q, expected go, ts, or mbt", value)
	}
}

func printDiagnostics(path string, diags []compiler.Diagnostic) {
	for _, diag := range diags {
		diagPath := path
		if diag.Path != "" {
			diagPath = diag.Path
		}
		prefix := ""
		if diag.Severity == checker.SeverityWarning {
			prefix = "warning: "
		}
		if diag.Pos.Line > 0 {
			fmt.Fprintf(os.Stderr, "%s:%d:%d: %s%s\n", diagPath, diag.Pos.Line, diag.Pos.Column, prefix, diag.Message)
		} else {
			fmt.Fprintf(os.Stderr, "%s: %s%s\n", diagPath, prefix, diag.Message)
		}
	}
}

func hasErrorDiagnostics(diags []compiler.Diagnostic) bool {
	for _, diag := range diags {
		if diag.Severity != checker.SeverityWarning {
			return true
		}
	}
	return false
}

func parseDiagnostics(errs []parser.Error) []compiler.Diagnostic {
	diags := make([]compiler.Diagnostic, 0, len(errs))
	for _, err := range errs {
		diags = append(diags, compiler.Diagnostic{
			Message: err.Message,
			Pos:     err.Pos,
		})
	}
	return diags
}

func parseDiagnosticsForPath(path string, errs []parser.Error) []compiler.Diagnostic {
	diags := parseDiagnostics(errs)
	for i := range diags {
		diags[i].Path = path
	}
	return diags
}

func defaultBinaryName(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	if ext != "" {
		base = base[:len(base)-len(ext)]
	}
	if base == "" {
		return "main"
	}
	return base
}
