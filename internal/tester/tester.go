package tester

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/oboard/rune-lang/internal/checker"
	gocodegen "github.com/oboard/rune-lang/internal/codegen/go"
	moonbitcodegen "github.com/oboard/rune-lang/internal/codegen/moonbit"
	tscodegen "github.com/oboard/rune-lang/internal/codegen/typescript"
	"github.com/oboard/rune-lang/internal/compiler"
	"github.com/oboard/rune-lang/internal/ir"
	"github.com/oboard/rune-lang/internal/selfhostrunner"
)

var (
	selfhostCompileTypeScriptFiles func([]SourceFile) CompileResult
	selfhostCompileMoonBitFiles    func([]SourceFile) CompileResult
)

func RegisterSelfhostCompilers(ts func([]SourceFile) CompileResult, mbt func([]SourceFile) CompileResult) {
	selfhostCompileTypeScriptFiles = ts
	selfhostCompileMoonBitFiles = mbt
}

func registerSelfhostCompilers(ts func([]SourceFile) CompileResult, mbt func([]SourceFile) CompileResult) func() {
	prevTS := selfhostCompileTypeScriptFiles
	prevMBT := selfhostCompileMoonBitFiles
	RegisterSelfhostCompilers(ts, mbt)
	return func() {
		selfhostCompileTypeScriptFiles = prevTS
		selfhostCompileMoonBitFiles = prevMBT
	}
}

type SourceFile struct {
	Path   string
	Source string
}

type CompileResult struct {
	Ok     bool
	Output string
	Errors []string
}

type Summary struct {
	Passed  int
	Failed  int
	Skipped int
	Files   int
}

func Run(path string, pattern string, out io.Writer) (Summary, error) {
	return run(path, pattern, out, runSelfhostTest)
}

func RunWithBackend(path string, pattern string, backend string, out io.Writer) (Summary, error) {
	switch backend {
	case "go":
		return runGoBatch(path, pattern, out)
	case "ts":
		return run(path, pattern, out, runTypeScriptTest)
	case "mbt":
		return run(path, pattern, out, runMoonBitTest)
	default:
		return Summary{}, fmt.Errorf("invalid backend %q, expected go, ts, or mbt", backend)
	}
}

type testRunner func(*ir.File, *ir.Test) selfhostrunner.Result

func run(path string, pattern string, out io.Writer, runner testRunner) (Summary, error) {
	if path == "" {
		path = "tests"
	}
	files, err := runeFiles(path)
	if err != nil {
		return Summary{}, err
	}
	var match *regexp.Regexp
	if pattern != "" {
		match, err = regexp.Compile(pattern)
		if err != nil {
			return Summary{}, fmt.Errorf("invalid test pattern: %w", err)
		}
	}

	start := time.Now()
	summary := Summary{Files: len(files)}
	for _, file := range files {
		prog, diags := compiler.AnalyzeFile(file)
		if len(diags) > 0 {
			summary.Failed++
			fmt.Fprintf(out, "=== RUN %s\n", file)
			fmt.Fprintf(out, "--- FAIL %s (compile)\n", file)
			for _, diag := range diags {
				diagPath := file
				if diag.Path != "" {
					diagPath = diag.Path
				}
				if diag.Pos.Line > 0 {
					fmt.Fprintf(out, "    %s:%d:%d: %s\n", diagPath, diag.Pos.Line, diag.Pos.Column, diag.Message)
				} else {
					fmt.Fprintf(out, "    %s: %s\n", diagPath, diag.Message)
				}
			}
			continue
		}
		for _, test := range prog.IR.Tests {
			fullName := file + " " + test.Name
			if match != nil && !match.MatchString(test.Name) && !match.MatchString(fullName) {
				summary.Skipped++
				continue
			}
			fmt.Fprintf(out, "=== RUN %s ? %s\n", file, test.Name)
			testStart := time.Now()
			result := runner(prog.IR, test)
			elapsed := time.Since(testStart)
			if result.Err != nil {
				summary.Failed++
				fmt.Fprintf(out, "--- FAIL %s ? %s (%s)\n", file, test.Name, elapsed.Round(time.Microsecond))
				if result.Output != "" {
					fmt.Fprintf(out, "    output:\n%s", indent(result.Output, "      "))
				}
				fmt.Fprintf(out, "    error: %s\n", result.Err)
				continue
			}
			summary.Passed++
			fmt.Fprintf(out, "--- PASS %s ? %s (%s)\n", file, test.Name, elapsed.Round(time.Microsecond))
			if result.Output != "" {
				fmt.Fprintf(out, "    output:\n%s", indent(result.Output, "      "))
			}
		}
	}
	total := summary.Passed + summary.Failed + summary.Skipped
	status := "PASS"
	if summary.Failed > 0 {
		status = "FAIL"
	}
	fmt.Fprintf(out, "%s %d tests, %d passed, %d failed, %d skipped in %s\n", status, total, summary.Passed, summary.Failed, summary.Skipped, time.Since(start).Round(time.Millisecond))
	if summary.Failed > 0 {
		return summary, fmt.Errorf("test failed")
	}
	if total == 0 {
		return summary, fmt.Errorf("no tests matched")
	}
	return summary, nil
}

func runSelfhostTest(file *ir.File, test *ir.Test) selfhostrunner.Result {
	return selfhostrunner.RunTestIR(file, test.Name)
}

const goBatchTestMarker = "__RUNE_TEST_START__"

func runGoBatch(path string, pattern string, out io.Writer) (Summary, error) {
	if path == "" {
		path = "tests"
	}
	files, err := runeFiles(path)
	if err != nil {
		return Summary{}, err
	}
	var match *regexp.Regexp
	if pattern != "" {
		match, err = regexp.Compile(pattern)
		if err != nil {
			return Summary{}, fmt.Errorf("invalid test pattern: %w", err)
		}
	}

	start := time.Now()
	summary := Summary{Files: len(files)}
	for _, file := range files {
		prog, diags := compiler.AnalyzeFile(file)
		if len(diags) > 0 {
			summary.Failed++
			fmt.Fprintf(out, "=== RUN %s\n", file)
			fmt.Fprintf(out, "--- FAIL %s (compile)\n", file)
			for _, diag := range diags {
				diagPath := file
				if diag.Path != "" {
					diagPath = diag.Path
				}
				if diag.Pos.Line > 0 {
					fmt.Fprintf(out, "    %s:%d:%d: %s\n", diagPath, diag.Pos.Line, diag.Pos.Column, diag.Message)
				} else {
					fmt.Fprintf(out, "    %s: %s\n", diagPath, diag.Message)
				}
			}
			continue
		}
		tests := make([]*ir.Test, 0, len(prog.IR.Tests))
		for _, test := range prog.IR.Tests {
			fullName := file + " " + test.Name
			if match != nil && !match.MatchString(test.Name) && !match.MatchString(fullName) {
				summary.Skipped++
				continue
			}
			tests = append(tests, test)
			fmt.Fprintf(out, "=== RUN %s ? %s\n", file, test.Name)
		}
		if len(tests) == 0 {
			continue
		}
		testStart := time.Now()
		result := runGoTestBatch(prog.IR, tests)
		elapsed := time.Since(testStart)
		outputs, lastStarted := splitGoBatchOutput(result.Output, len(tests))
		if result.Err != nil {
			failed := lastStarted
			if failed < 0 || failed >= len(tests) {
				failed = 0
			}
			for idx, test := range tests {
				switch {
				case idx < failed:
					summary.Passed++
					fmt.Fprintf(out, "--- PASS %s ? %s (%s)\n", file, test.Name, elapsed.Round(time.Microsecond))
					if outputs[idx] != "" {
						fmt.Fprintf(out, "    output:\n%s", indent(outputs[idx], "      "))
					}
				case idx == failed:
					summary.Failed++
					fmt.Fprintf(out, "--- FAIL %s ? %s (%s)\n", file, test.Name, elapsed.Round(time.Microsecond))
					if outputs[idx] != "" {
						fmt.Fprintf(out, "    output:\n%s", indent(outputs[idx], "      "))
					}
					fmt.Fprintf(out, "    error: %s\n", result.Err)
				default:
					summary.Skipped++
				}
			}
			continue
		}
		for idx, test := range tests {
			summary.Passed++
			fmt.Fprintf(out, "--- PASS %s ? %s (%s)\n", file, test.Name, elapsed.Round(time.Microsecond))
			if outputs[idx] != "" {
				fmt.Fprintf(out, "    output:\n%s", indent(outputs[idx], "      "))
			}
		}
	}
	total := summary.Passed + summary.Failed + summary.Skipped
	status := "PASS"
	if summary.Failed > 0 {
		status = "FAIL"
	}
	fmt.Fprintf(out, "%s %d tests, %d passed, %d failed, %d skipped in %s\n", status, total, summary.Passed, summary.Failed, summary.Skipped, time.Since(start).Round(time.Millisecond))
	if summary.Failed > 0 {
		return summary, fmt.Errorf("test failed")
	}
	if total == 0 {
		return summary, fmt.Errorf("no tests matched")
	}
	return summary, nil
}

func runGoTestBatch(file *ir.File, tests []*ir.Test) selfhostrunner.Result {
	src, err := gocodegen.GenerateIR(goBatchTestFile(file, tests))
	if err != nil {
		return selfhostrunner.Result{Err: err}
	}
	bin, err := cachedGoTestBinary(src)
	if err != nil {
		return selfhostrunner.Result{Err: err}
	}
	cmd := exec.Command(bin)
	out, err := cmd.CombinedOutput()
	return selfhostrunner.Result{Output: string(out), Err: err}
}

func cachedGoTestBinary(src string) (string, error) {
	dir, err := goTestCacheDir()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte("rune-go-test-v1\n" + src))
	key := hex.EncodeToString(sum[:])[:16]
	binPath := filepath.Join(dir, "test-"+key)
	if _, err := os.Stat(binPath); err == nil {
		return binPath, nil
	}
	srcPath := filepath.Join(dir, "test-"+key+".go")
	if err := os.WriteFile(srcPath, []byte(src), 0o644); err != nil {
		return "", err
	}
	tmpFile, err := os.CreateTemp(dir, "test-"+key+"-*.tmp")
	if err != nil {
		return "", err
	}
	tmpBin := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		return "", err
	}
	defer os.Remove(tmpBin)
	cmd := exec.Command("go", "build", "-p=1", "-gcflags=all=-l", "-o", tmpBin, srcPath)
	if cwd, err := os.Getwd(); err == nil {
		cmd.Dir = cwd
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("build Go test binary: %w\n%s", err, out)
	}
	if err := os.Rename(tmpBin, binPath); err != nil {
		return "", err
	}
	cleanupGoTestCache(dir, binPath)
	return binPath, nil
}

func goTestCacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		base = os.TempDir()
	}
	dir := filepath.Join(base, "rune-lang", "go-test-bin")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func cleanupGoTestCache(dir string, keepPath string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	type cachedFile struct {
		path string
		info fs.FileInfo
	}
	var bins []cachedFile
	keepPath = filepath.Clean(keepPath)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "test-") || strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), ".tmp") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if filepath.Clean(path) == keepPath {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		bins = append(bins, cachedFile{path: path, info: info})
	}
	sort.Slice(bins, func(i, j int) bool {
		return bins[i].info.ModTime().After(bins[j].info.ModTime())
	})
	for idx, file := range bins {
		if idx < 8 {
			continue
		}
		_ = os.Remove(file.path)
		_ = os.Remove(file.path + ".go")
	}
}

func goBatchTestFile(file *ir.File, tests []*ir.Test) *ir.File {
	copy := *file
	functions := make([]*ir.Function, 0, len(file.Functions)+len(tests)+1)
	for _, fn := range file.Functions {
		if fn.Name == "main" && len(fn.Params) == 0 && len(fn.Generics) == 0 {
			continue
		}
		functions = append(functions, fn)
	}
	mainStatements := make([]ir.Stmt, 0, len(tests))
	for idx, test := range tests {
		name := goBatchTestFunctionName(idx)
		functions = append(functions, &ir.Function{
			Name:       name,
			SourceName: name,
			Private:    true,
			Return:     checker.Void,
			Body: &ir.BlockExpr{ExprBase: ir.ExprBase{Type: checker.Void}, Statements: []ir.Stmt{
				&ir.ExprStmt{Expr: goBatchMarkerCall(idx)},
				&ir.ExprStmt{Expr: test.Body},
			}},
			Pos:     test.Pos,
			NamePos: test.Pos,
		})
		mainStatements = append(mainStatements, &ir.ExprStmt{Expr: &ir.CallExpr{
			ExprBase: ir.ExprBase{Type: checker.Void},
			Callee:   &ir.Identifier{ExprBase: ir.ExprBase{Type: checker.Type("Func[Void]")}, Name: name},
		}})
	}
	functions = append(functions, &ir.Function{
		Name:       "main",
		SourceName: "main",
		Return:     checker.Void,
		Body:       &ir.BlockExpr{ExprBase: ir.ExprBase{Type: checker.Void}, Statements: mainStatements},
	})
	copy.Functions = functions
	copy.Tests = nil
	return &copy
}

func goBatchMarkerCall(idx int) ir.Expr {
	return &ir.CallExpr{
		ExprBase: ir.ExprBase{Type: checker.Void},
		Callee: &ir.SelectorExpr{
			ExprBase: ir.ExprBase{Type: checker.Void},
			Receiver: &ir.AtExpr{
				ExprBase: ir.ExprBase{Type: checker.Type("@io")},
				Name:     "io",
			},
			Name: "println",
		},
		Args: []ir.Expr{&ir.StringLiteral{ExprBase: ir.ExprBase{Type: checker.String}, Value: fmt.Sprintf("%s%d", goBatchTestMarker, idx)}},
	}
}

func goBatchTestFunctionName(idx int) string {
	return fmt.Sprintf("__rune_test_%d", idx)
}

func splitGoBatchOutput(output string, testCount int) ([]string, int) {
	out := make([]string, testCount)
	current := -1
	lastStarted := -1
	for _, line := range strings.SplitAfter(output, "\n") {
		trimmed := strings.TrimRight(line, "\r\n")
		if strings.HasPrefix(trimmed, goBatchTestMarker) {
			idxText := strings.TrimPrefix(trimmed, goBatchTestMarker)
			var idx int
			if _, err := fmt.Sscanf(idxText, "%d", &idx); err == nil && idx >= 0 && idx < testCount {
				current = idx
				lastStarted = idx
			}
			continue
		}
		if current >= 0 && current < testCount {
			out[current] += line
		}
	}
	return out, lastStarted
}

func testMainFile(file *ir.File, test *ir.Test) *ir.File {
	copy := *file
	functions := make([]*ir.Function, 0, len(file.Functions)+1)
	for _, fn := range file.Functions {
		if fn.Name == "main" && len(fn.Params) == 0 && len(fn.Generics) == 0 {
			continue
		}
		functions = append(functions, fn)
	}
	functions = append(functions, &ir.Function{
		Name:       "main",
		SourceName: "main",
		Return:     "Void",
		Body:       test.Body,
		Pos:        test.Pos,
		NamePos:    test.Pos,
	})
	copy.Functions = functions
	copy.Tests = nil
	return &copy
}

func runGoTest(file *ir.File, test *ir.Test) selfhostrunner.Result {
	src, err := gocodegen.GenerateIR(testMainFile(file, test))
	if err != nil {
		return selfhostrunner.Result{Err: err}
	}
	bin, err := cachedGoTestBinary(src)
	if err != nil {
		return selfhostrunner.Result{Err: err}
	}
	cmd := exec.Command(bin)
	out, err := cmd.CombinedOutput()
	return selfhostrunner.Result{Output: string(out), Err: err}
}

func runTypeScriptTest(file *ir.File, test *ir.Test) selfhostrunner.Result {
	src, err := typeScriptTestRuntimeSource(testMainFile(file, test))
	if err != nil {
		return selfhostrunner.Result{Err: err}
	}
	src += "\nif (typeof __main === \"function\") {\n  const __runeMainResult = __main();\n  if (__runeMainResult && typeof __runeMainResult.then === \"function\") {\n    await __runeMainResult;\n  }\n}\nif (typeof runeWaitAll === \"function\") {\n  await runeWaitAll();\n}\n"
	dir, err := os.MkdirTemp("", "rune-test-ts-*")
	if err != nil {
		return selfhostrunner.Result{Err: err}
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "main.ts")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		return selfhostrunner.Result{Err: err}
	}
	cmd, err := typeScriptRuntimeCommand(path)
	if err != nil {
		return selfhostrunner.Result{Err: err}
	}
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return selfhostrunner.Result{Output: string(out), Err: err}
}

func typeScriptTestRuntimeSource(file *ir.File) (string, error) {
	if selfhostCompileTypeScriptFiles != nil {
		files, err := irSourceFiles(file)
		if err != nil {
			return "", err
		}
		result := selfhostCompileTypeScriptFiles(files)
		if result.Ok && result.Output != "" && typeScriptTestSelfhostEligible(file) {
			return result.Output, nil
		}
	}
	return tscodegen.GenerateIR(testTypeScriptRuntimeIR(file))
}

func testTypeScriptRuntimeIR(file *ir.File) *ir.File {
	copy := *file
	copy.TSImports = append([]ir.TSImport(nil), file.TSImports...)
	for i := range copy.TSImports {
		copy.TSImports[i].Specifier = typeScriptRuntimeSpecifier(copy.TSImports[i].Specifier)
	}
	return &copy
}

func typeScriptRuntimeSpecifier(specifier string) string {
	if specifier == "" || strings.HasPrefix(specifier, "./") || strings.HasPrefix(specifier, "../") || strings.HasPrefix(specifier, "file:") {
		return specifier
	}
	if filepath.IsAbs(specifier) {
		return "file://" + specifier
	}
	return "./" + specifier
}

func typeScriptRuntimeCommand(path string) (*exec.Cmd, error) {
	if runner, err := exec.LookPath("bun"); err == nil {
		return exec.Command(runner, path), nil
	}
	if runner, err := exec.LookPath("deno"); err == nil {
		return exec.Command(runner, "run", path), nil
	}
	if runner, err := exec.LookPath("node"); err == nil {
		return exec.Command(runner, path), nil
	}
	if runner, err := exec.LookPath("ts-node"); err == nil {
		return exec.Command(runner, "--esm", path), nil
	}
	return nil, fmt.Errorf("TypeScript backend requires bun, deno, node, or ts-node")
}

func runMoonBitTest(file *ir.File, test *ir.Test) selfhostrunner.Result {
	src, err := moonBitTestRuntimeSource(testMainFile(file, test))
	if err != nil {
		return selfhostrunner.Result{Err: err}
	}
	dir, err := os.MkdirTemp("", "rune-test-mbt-*")
	if err != nil {
		return selfhostrunner.Result{Err: err}
	}
	defer os.RemoveAll(dir)
	if err := writeMoonBitPackage(dir, src); err != nil {
		return selfhostrunner.Result{Err: err}
	}
	cmd := exec.Command("moon", "run", "--target", "native", ".")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return selfhostrunner.Result{Output: string(out), Err: err}
}

func moonBitTestRuntimeSource(file *ir.File) (string, error) {
	if selfhostCompileMoonBitFiles != nil {
		files, err := irSourceFiles(file)
		if err != nil {
			return "", err
		}
		result := selfhostCompileMoonBitFiles(files)
		if result.Ok && result.Output != "" && moonBitTestSelfhostEligible(file) {
			return result.Output, nil
		}
	}
	return moonbitcodegen.GenerateIR(file)
}

func irSourceFiles(file *ir.File) ([]SourceFile, error) {
	return []SourceFile{{Path: "main.rn", Source: "main() => 0\n"}}, nil
}

func typeScriptTestSelfhostEligible(file *ir.File) bool {
	if file == nil || len(file.TSImports) != 0 {
		return false
	}
	return hasOnlyMainFunction(file.Functions)
}

func moonBitTestSelfhostEligible(file *ir.File) bool {
	if file == nil || len(file.TSImports) != 0 {
		return false
	}
	return hasOnlyMainFunction(file.Functions)
}

func hasOnlyMainFunction(functions []*ir.Function) bool {
	if len(functions) == 0 {
		return false
	}
	for _, fn := range functions {
		if fn == nil {
			return false
		}
		if fn.Name != "main" {
			return false
		}
		if len(fn.Params) != 0 || len(fn.Generics) != 0 {
			return false
		}
	}
	return true
}

func writeMoonBitPackage(dir string, src string) error {
	files := map[string]string{
		"main.mbt": src,
		"moon.mod": moonBitMod(src),
		"moon.pkg": moonBitPkg(src),
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func moonBitMod(src string) string {
	if !moonBitUsesAsync(src) {
		return "name = \"oboard/rune_test_mbt\"\n"
	}
	var b strings.Builder
	b.WriteString("name = \"oboard/rune_test_mbt\"\n\nimport {\n")
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
		"moonbitlang/core/bigint",
		"moonbitlang/core/json",
	}
	if moonBitUsesAsync(src) {
		imports = append(imports, "moonbitlang/async")
	}
	if strings.Contains(src, "@fs.") {
		imports = append(imports, "moonbitlang/async/fs")
	}
	if strings.Contains(src, "@io.") {
		imports = append(imports, "moonbitlang/async/io")
	}
	if strings.Contains(src, "@socket.") {
		imports = append(imports, "moonbitlang/async/socket")
	}
	if strings.Contains(src, "@gzip.") {
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
		b.WriteByte('\n')
	}
	b.WriteString("}\n\nwarnings = \"-1-7-23-67\"\n\n")
	if strings.Contains(src, "@fs.") || strings.Contains(src, "@gzip.") || strings.Contains(src, "@socket.") {
		b.WriteString("supported_targets = \"+native\"\n\n")
	}
	b.WriteString("options(\n  \"is-main\": true,\n)\n")
	return b.String()
}

func moonBitUsesBikallemCompress(src string) bool {
	return strings.Contains(src, "@flate.") ||
		strings.Contains(src, "@brotli.") ||
		strings.Contains(src, "@zstd.") ||
		strings.Contains(src, "@utf8.")
}

func moonBitUsesAsync(src string) bool {
	return strings.Contains(src, "async fn") ||
		strings.Contains(src, "@fs.") ||
		strings.Contains(src, "@io.") ||
		strings.Contains(src, "@gzip.")
}

func runeFiles(path string) ([]string, error) {
	info, err := fs.Stat(osDirFS{}, path)
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

func indent(s string, prefix string) string {
	var b bytes.Buffer
	for _, ch := range s {
		if b.Len() == 0 || b.Bytes()[b.Len()-1] == '\n' {
			b.WriteString(prefix)
		}
		b.WriteRune(ch)
	}
	if len(s) > 0 && s[len(s)-1] != '\n' {
		b.WriteByte('\n')
	}
	return b.String()
}

type osDirFS struct{}

func (osDirFS) Open(name string) (fs.File, error) {
	return os.Open(name)
}
