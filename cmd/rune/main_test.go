package main

import (
	"bytes"
	"go/format"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/compiler"
	runefmt "github.com/oboard/rune-lang/internal/format"
	"github.com/oboard/rune-lang/internal/parser"
)

func TestRuneCLIDeclarationUsesSelfhostEmitter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mod.rn")
	writeTestFile(t, path, `+ User: { name: String }
+ makeUser(name: String) -> User => User { name: name }
`)

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := runRuneCLI([]string{"dts", path}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("runRuneCLI(dts) error = %v, stderr = %s", err, errOut.String())
	}
	got := out.String()
	if !strings.Contains(got, "type __User = {") {
		t.Fatalf("selfhost declaration missing __User struct: %q", got)
	}
	if !strings.Contains(got, "declare function __makeUser(__name: string): __User;") {
		t.Fatalf("selfhost declaration missing __makeUser signature: %q", got)
	}
	if !strings.Contains(got, "export type User = __User;") {
		t.Fatalf("selfhost declaration missing User type alias export: %q", got)
	}
	if !strings.Contains(got, "export declare const makeUser: typeof __makeUser;") {
		t.Fatalf("selfhost declaration missing makeUser value alias export: %q", got)
	}
}

// TestRuneCLIDeclarationSelfhostMatchesHost validates that the selfhost dts
// emitter produces byte-identical output to the host declaration generator for
// common declaration shapes: structs, enums, generic structs, functions with
// arrays, nullable returns, and routines.
func TestRuneCLIDeclarationSelfhostMatchesHost(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"empty", "main() => 0\n"},
		{"simple_fn", "+ answer() -> Int => 42\n"},
		{"struct", "+ User: { name: String, active: Bool }\n"},
		{"nested_struct", `+ Address: { street: String, city: String }
+ User: { name: String, address: Address }
`},
		{"generic_box", "+ Box[T]: { value: T }\n"},
		{"generic_constrained", "+ Pair[L, R]: { left: L, right: R }\n"},
		{"array_return", "+ nums() -> Array[Int] => [1, 2, 3]\n"},
		{"arr_of_struct", "+ users() -> Array[User] => []\n+ User: { name: String }\n"},
		{"nullable", "+ maybe() -> String? => null\n"},
		{"enum", "+ Color: { Red, Green, Blue }\n"},
		{"enum_value", "+ Priority: { Low = 1, Medium = 5, High = 10 }\n"},
		{"payload_enum", "+ Shape: { Circle(radius: Double), Square(side: Double) }\n"},
		{"map_of_string", "+ config() -> Map[String, String] => {}\n"},
		{"readonly_arr", "+ bits() -> ReadonlyArray[Int] => [0, 1]\n"},
		{"tuple2", "+ pipe() -> Tuple[Int, String] => [1, \"a\"]\n"},
		{"multi_fn", `+ add(a: Int, b: Int) -> Int => a + b
+ sub(a: Int, b: Int) -> Int => a - b
+ multiply(a: Int, b: Int) -> Int => a * b
`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "m.rn")
			writeTestFile(t, path, tc.src)

			host, diags := compiler.GenerateTypeScriptDeclarationFile(path)
			if len(diags) > 0 {
				t.Skipf("host declaration also reports errors: %v", diags)
			}
			out := __compileDeclarations(tc.src)
			if !out.__ok {
				t.Fatalf("selfhost compileDeclarations failed: %v", out.__errors)
			}
			if out.__output != host {
				t.Fatalf("selfhost declaration mismatch\nselfhost:\n%s\nhost:\n%s", out.__output, host)
			}
		})
	}
}

func TestRuneCLITypeScriptUsesSelfhostCompilerImportGraph(t *testing.T) {
	dir := t.TempDir()
	entry := filepath.Join(dir, "main.rn")
	writeTestFile(t, filepath.Join(dir, "helper.rn"), "+ answer() -> Int => 42\n")
	writeTestFile(t, entry, "@\"helper.rn\"\nmain() => @io.println(answer())\n")

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := runRuneCLI([]string{"ts", entry}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("runRuneCLI(ts import graph) error = %v, stderr = %s", err, errOut.String())
	}
	if got := out.String(); !strings.Contains(got, "function __answer(): number") {
		t.Fatalf("self-host generated TypeScript = %q, want imported answer", got)
	}
}

func TestRuneCLITypeScriptUsesSelfhostCompiler(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.rn")
	writeTestFile(t, path, "main() => @io.println(42)\n")

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := runRuneCLI([]string{"ts", path}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("runRuneCLI(ts) error = %v, stderr = %s", err, errOut.String())
	}
	if got := out.String(); !strings.Contains(got, "console.log(42)") {
		t.Fatalf("self-host generated TypeScript = %q, want console.log(42)", got)
	}
}

func TestRuneCLIGoUsesSelfhostCompilerImportGraph(t *testing.T) {
	dir := t.TempDir()
	entry := filepath.Join(dir, "main.rn")
	writeTestFile(t, filepath.Join(dir, "helper.rn"), "+ answer() -> Int => 42\n")
	writeTestFile(t, entry, "@\"helper.rn\"\nmain() => @io.println(answer())\n")

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := runRuneCLI([]string{"go", entry}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("runRuneCLI(go import graph) error = %v, stderr = %s", err, errOut.String())
	}
	if got := out.String(); !strings.Contains(got, "func __answer() int") {
		t.Fatalf("self-host generated Go = %q, want imported answer", got)
	}
}

func TestGeneratedSelfhostCompilerInfersComplexStructuralTypes(t *testing.T) {
	source := `fun(flag) => {
  (flag ? (x) => { k: x.a + 1 } : (y) => { k: y.b + 1 })({ b: 2, z: false, a: 1 }).k
}
main() => @io.println(fun(true) + fun(false))
`
	result := __compileGo(source)
	if !result.__ok {
		t.Fatalf("__compileGo() errors = %v", result.__errors)
	}
	if !strings.Contains(result.__output, "func __fun(__flag bool) int") {
		t.Fatalf("selfhost generated Go = %q, want inferred bool-to-int signature", result.__output)
	}
	if !strings.Contains(result.__output, "struct { __k int;") {
		t.Fatalf("selfhost generated Go = %q, want inferred object result", result.__output)
	}
	formatted, err := format.Source([]byte(result.__output))
	if err != nil {
		t.Fatalf("format selfhost generated Go: %v\n%s", err, result.__output)
	}
	goFile := filepath.Join(t.TempDir(), "main.go")
	if err := os.WriteFile(goFile, formatted, 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command("go", "run", goFile).CombinedOutput()
	if err != nil {
		t.Fatalf("run selfhost generated Go: %v\n%s", err, out)
	}
	if got := string(out); got != "5\n" {
		t.Fatalf("selfhost generated output = %q, want 5", got)
	}
}

func TestRuneCLIGoUsesSelfhostCompiler(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.rn")
	writeTestFile(t, path, "main() => @io.println(42)\n")

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := runRuneCLI([]string{"go", path}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("runRuneCLI(go) error = %v, stderr = %s", err, errOut.String())
	}
	if got := out.String(); !strings.Contains(got, "fmt.Println(42)") {
		t.Fatalf("self-host generated Go = %q, want fmt.Println(42)", got)
	}
}

// TestGeneratedSelfhostCompilerMatchesGoHostCompiled validates that the
// self-hosted compiler is the mainline path: for canonical single-file
// programs it must build and run with output identical to the Go-host
// compiler, pinning the bridge without constraining symbol naming.
func TestGeneratedSelfhostCompilerMatchesGoHostCompiled(t *testing.T) {
	cases := []struct {
		name   string
		source string
	}{
		{"print", "main() => @io.println(42)\n"},
		{"fib", "fib(n) => {\n  0 => 0\n  1 => 1\n  _ => fib(n - 1) + fib(n - 2)\n}\n\nmain() => {\n  @io.println(fib(10))\n}\n"},
		{"typed", "User: {\n  name: String\n  active: Bool\n}\n\nmakeUser(name: String) -> User => {\n  name: name,\n  active: true\n}\n\nmain() => @io.println(makeUser(\"Rune\").name)\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			selfhostOut := __compileGo(tc.source)
			if !selfhostOut.__ok {
				t.Fatalf("__compileGo() errors = %v", selfhostOut.__errors)
			}
			path := filepath.Join(t.TempDir(), "main.rn")
			if err := os.WriteFile(path, []byte(tc.source), 0o644); err != nil {
				t.Fatalf("WriteFile error = %v", err)
			}
			hostOut, diags := compiler.GenerateGoFile(path)
			if len(diags) > 0 {
				t.Fatalf("host GenerateGoFile() diags = %v", diags)
			}
			runBuilt := func(goSrc string) string {
				dir := t.TempDir()
				formatted, err := format.Source([]byte(goSrc))
				if err != nil {
					t.Fatalf("format generated Go error = %v", err)
				}
				goFile := filepath.Join(dir, "main.go")
				if err := os.WriteFile(goFile, formatted, 0o644); err != nil {
					t.Fatalf("WriteFile(%s) error = %v", goFile, err)
				}
				exe := filepath.Join(dir, "prog")
				build := exec.Command("go", "build", "-o", exe, goFile)
				if out, err := build.CombinedOutput(); err != nil {
					t.Fatalf("go build failed: %v\n%s", err, out)
				}
				out, err := exec.Command(exe).CombinedOutput()
				if err != nil {
					t.Fatalf("run failed: %v\n%s", err, out)
				}
				return string(out)
			}
			if got, want := runBuilt(selfhostOut.__output), runBuilt(hostOut); got != want {
				t.Fatalf("self-host runtime output != host runtime output: got %q want %q", got, want)
			}
		})
	}
}

func TestGeneratedSelfhostCompilerEmitsPatternFunction(t *testing.T) {
	source := `fib(n) => {
  0 => 0
  1 => 1
  _ => fib(n - 1) + fib(n - 2)
}

main() => {
  @io.println(fib(10))
}
`
	out := __compileGo(source)
	if !out.__ok {
		t.Fatalf("__compileGo() errors = %v", out.__errors)
	}
	if !strings.Contains(out.__output, "func __fib(__n int) int") {
		t.Fatalf("generated Go =\n%s\nwant func __fib(__n int) int", out.__output)
	}
	if !strings.Contains(out.__output, "switch __n {") {
		t.Fatalf("generated Go =\n%s\nwant switch __n", out.__output)
	}
}

func TestExecuteSelfhostCheckDirectory(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "foo", "foo.rn"), "Foo: {}\n")

	var out bytes.Buffer
	if err := checkTarget(dir, &out); err != nil {
		t.Fatalf("checkTarget() error = %v", err)
	}
	if got := out.String(); got != "ok "+dir+"\n" {
		t.Fatalf("checkTarget() output = %q", got)
	}
}

func TestGeneratedSelfhostFormatterMatchesHostForComments(t *testing.T) {
	for _, src := range []string{
		"main()=>{\nvalue:=1 // value\n@io.println(value) // print\n}",
		"main()=>{\nvalue:=1 /* value */\n@io.println(value)}",
	} {
		file, errs := parser.Parse(src)
		if len(errs) > 0 {
			t.Fatalf("Parse(%q) errors = %v", src, errs)
		}
		if got, want := __fmt_formatSource(src), runefmt.Source(file, src); got != want {
			t.Fatalf("__fmt_formatSource(%q) = %q, want %q", src, got, want)
		}
	}
}

func TestGeneratedSelfhostFormatterMatchesHostCorpus(t *testing.T) {
	cases := []struct {
		src      string
		eligible bool
	}{
		{"1 + 2", true},
		{"let x: Int = 3\nx + 1", true},
		{`let s: String = "hello"`, true},
		{"User: { name: String, age: Int }\n\nmain() => User { name: \"a\", age: 1 }.name", false},
		{"Box[T]: { value: T }\n\nmain() => Box[Int] { value: 1 }", false},
		{"+ Task: { id: Int, label: String, done: Bool }", true},
		{"fib(n: Int) -> Int => (n < 2 ? n : fib(n - 1) + fib(n - 2))", false},
		{`Color: { Red, Green, Blue }

describe(color: Color) -> String => color {
	Red => "r"
	Blue => "b"
	_ => "g"
}`, false},
		{`Circle: { radius: Double }
Square: { side: Double }

main() => 0`, false},
		{"\tif true { 1 } else { 2 }", true},
		{"main() => {\n\ta := 1\n\tb := 2\n\ta + b\n}", false},
		{"main() => @io.println(\"hi\")", true},
		{"+ answer() -> Int => 42", true},
		{"main() => item.length()", true},
		{"data := 1 + 2 * 3", true},
		{"square := (n: Int) -> Int => n * n", true},
		{`TaggedWrapper: {
  Int: Int
  String: String
}`, true},
		{`main() => {
	a := 1
	b := 2
	a ? b
}`, false},
	}
	for _, tc := range cases {
		t.Run(tc.src[0:min(len(tc.src), 30)], func(t *testing.T) {
			file, errs := parser.Parse(tc.src)
			if len(errs) > 0 {
				t.Skipf("parser rejected input: %v", errs)
			}
			if got := selfhostFormatterEligible(tc.src); got != tc.eligible {
				t.Fatalf("selfhostFormatterEligible(%q) = %v, want %v", tc.src, got, tc.eligible)
			}
			if got, want := formatWithSelfhostBridge(file, tc.src), runefmt.Source(file, tc.src); got != want {
				t.Fatalf("formatWithSelfhostBridge(%q) = %q (selfhost: %q), want %q", tc.src, got, __fmt_formatSource(tc.src), want)
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestGeneratedSelfhostFormatterFormatsSimpleProgram(t *testing.T) {
	if got, want := __fmt_formatSource("main()=>1"), "main() => 1\n"; got != want {
		t.Fatalf("__fmt_formatSource() = %q, want %q", got, want)
	}
}

func TestFormatTargetFormatsDirectory(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "sub", "main.rn")
	writeTestFile(t, sourcePath, "main()=>1\n")

	var out bytes.Buffer
	if err := formatTarget(dir, false, false, &out); err != nil {
		t.Fatalf("formatTarget() error = %v", err)
	}
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if got := string(data); got != "main() => 1\n" {
		t.Fatalf("formatted source = %q", got)
	}
}

func TestFormatTargetRejectsStdoutDirectory(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "main.rn"), "main()=>1\n")

	var out bytes.Buffer
	err := formatTarget(dir, false, true, &out)
	if err == nil || !strings.Contains(err.Error(), "fmt --stdout only supports a single file") {
		t.Fatalf("formatTarget() error = %v, want stdout directory error", err)
	}
}

func TestResolveRunEntryFindsUniqueMain(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "helper.rn"), "helper() => 1\n")
	mainPath := filepath.Join(dir, "nested", "main.rn")
	writeTestFile(t, mainPath, "main() => @io.println(\"ok\")\n")

	entry, diags, err := resolveRunEntry(dir)
	if err != nil {
		t.Fatalf("resolveRunEntry() error = %v", err)
	}
	if len(diags) > 0 {
		t.Fatalf("resolveRunEntry() diagnostics = %v", diags)
	}
	if entry != mainPath {
		t.Fatalf("resolveRunEntry() = %q, want %q", entry, mainPath)
	}
}

func TestResolveRunEntryRejectsDuplicateMain(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "a.rn"), "main() => 1\n")
	writeTestFile(t, filepath.Join(dir, "b.rn"), "main() => 2\n")

	_, diags, err := resolveRunEntry(dir)
	if err == nil {
		t.Fatal("resolveRunEntry() succeeded, want duplicate main error")
	}
	if len(diags) > 0 {
		t.Fatalf("resolveRunEntry() diagnostics = %v", diags)
	}
	if !strings.Contains(err.Error(), "multiple main functions found") {
		t.Fatalf("resolveRunEntry() error = %v, want duplicate main error", err)
	}
}

func TestCompileTypeScriptToTempRunsFromEntryDirectory(t *testing.T) {
	dir := t.TempDir()
	tsPath := filepath.Join(dir, "greet.ts")
	writeTestFile(t, tsPath, "export function greet(name: string): string { return name }\n")
	mainPath := filepath.Join(dir, "main.rn")
	writeTestFile(t, mainPath, `@"greet.ts"

main() => @io.println(greet("Rune"))
`)

	tsFile, runDir, cleanup, err := compileTypeScriptToTemp(mainPath)
	if err != nil {
		t.Fatalf("compileTypeScriptToTemp() error = %v", err)
	}
	defer cleanup()
	if filepath.Dir(tsFile) != dir {
		t.Fatalf("compiled TypeScript dir = %q, want %q", filepath.Dir(tsFile), dir)
	}
	if runDir != dir {
		t.Fatalf("run dir = %q, want %q", runDir, dir)
	}
	data, err := os.ReadFile(tsFile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(data)
	if !(strings.Contains(text, `from "./greet.ts"`) || strings.Contains(text, `from "`+filepath.ToSlash(tsPath)+`"`)) {
		t.Fatalf("runtime TypeScript = %q, want relative or entry-rooted greet import", data)
	}
}

func TestSelfhostTypeScriptRuntimeSourceUsesSelfhostCompilerImportGraph(t *testing.T) {
	dir := t.TempDir()
	entry := filepath.Join(dir, "main.rn")
	writeTestFile(t, filepath.Join(dir, "helper.rn"), "+ answer() -> Int => 42\n")
	writeTestFile(t, entry, "@\"helper.rn\"\nmain() => @io.println(answer())\n")

	src, diags := selfhostTypeScriptRuntimeSource(entry)
	if len(diags) > 0 {
		t.Fatalf("selfhostTypeScriptRuntimeSource() diagnostics = %v", diags)
	}
	if !strings.Contains(src, "function __answer(): number") {
		t.Fatalf("runtime TypeScript = %q, want imported answer", src)
	}
}

func TestCompileMoonBitProjectToTempWritesPackage(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.rn")
	writeTestFile(t, mainPath, `main() => {
  @io.println("Rune")
}
`)

	projectDir, cleanup, err := compileMoonBitProjectToTemp(mainPath)
	if err != nil {
		t.Fatalf("compileMoonBitProjectToTemp() error = %v", err)
	}
	defer cleanup()
	for _, name := range []string{"main.mbt", "moon.mod", "moon.pkg"} {
		if _, err := os.Stat(filepath.Join(projectDir, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
	data, err := os.ReadFile(filepath.Join(projectDir, "main.mbt"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if got := string(data); !strings.Contains(got, "fn main {\n") || !strings.Contains(got, `println("Rune")`) {
		t.Fatalf("main.mbt =\n%s", got)
	}
	selfhost, diags := selfhostMoonBitRuntimeSource(mainPath)
	if len(diags) > 0 {
		t.Fatalf("selfhostMoonBitRuntimeSource() diagnostics = %v", diags)
	}
	if string(data) != selfhost {
		t.Fatalf("compileMoonBitProjectToTemp main.mbt != selfhost runtime source\nproject:\n%s\nselfhost:\n%s", string(data), selfhost)
	}
	pkg, err := os.ReadFile(filepath.Join(projectDir, "moon.pkg"))
	if err != nil {
		t.Fatalf("ReadFile(moon.pkg) error = %v", err)
	}
	for _, want := range []string{
		`"moonbitlang/core/env"`,
		`"moonbitlang/core/string"`,
	} {
		if !strings.Contains(string(pkg), want) {
			t.Fatalf("moon.pkg =\n%s\nmissing %s", pkg, want)
		}
	}
}

func TestSelfhostMoonBitRuntimeSourceUsesSelfhostCompilerImportGraph(t *testing.T) {
	dir := t.TempDir()
	entry := filepath.Join(dir, "main.rn")
	writeTestFile(t, filepath.Join(dir, "helper.rn"), "+ answer() -> Int => 42\n")
	writeTestFile(t, entry, "@\"helper.rn\"\nmain() => @io.println(answer())\n")

	src, diags := selfhostMoonBitRuntimeSource(entry)
	if len(diags) > 0 {
		t.Fatalf("selfhostMoonBitRuntimeSource() diagnostics = %v", diags)
	}
	if !strings.Contains(src, "fn __answer() -> Int") {
		t.Fatalf("runtime MoonBit = %q, want imported answer", src)
	}
}

func TestRunEntryMoonBit(t *testing.T) {
	if _, err := exec.LookPath("moon"); err != nil {
		t.Skip("moon command not available")
	}
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.rn")
	writeTestFile(t, mainPath, `main() => {
  @io.println("Rune")
}
`)

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := runEntry(mainPath, "mbt", "native", nil, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("runEntry() error = %v, stderr = %s", err, errOut.String())
	}
	if got, want := out.String(), "Rune\n"; got != want {
		t.Fatalf("runEntry() output = %q, want %q", got, want)
	}
}

func TestRunEntryMoonBitCLIArgs(t *testing.T) {
	if _, err := exec.LookPath("moon"); err != nil {
		t.Skip("moon command not available")
	}
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.rn")
	writeTestFile(t, mainPath, `#cli.command("ship", "Ship a build artifact", "1.0.0")
Args: {
  #cli.flag("v", "enable verbose output")
  verbose: Bool
  #cli.option("o", "FILE", "write output file", "dist/app")
  output: String
  #cli.arg("target name")
  target: String
}

#cli.main
main(args: Args) => {
  @io.println(args.target)
  @io.println(args.output)
  @io.println(args.verbose.toString())
}
`)

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := runEntry(mainPath, "mbt", "native", []string{"wasm", "-v", "-o", "out.js"}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("runEntry() error = %v, stderr = %s", err, errOut.String())
	}
	if got, want := out.String(), "wasm\nout.js\ntrue\n"; got != want {
		t.Fatalf("runEntry() output = %q, want %q", got, want)
	}
}

func TestRunEntryMoonBitRejectsInvalidTarget(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.rn")
	writeTestFile(t, mainPath, `main() => {
  @io.println("Rune")
}
`)

	var out bytes.Buffer
	var errOut bytes.Buffer
	err := runEntry(mainPath, "mbt", "invalid", nil, strings.NewReader(""), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "invalid MoonBit target") {
		t.Fatalf("runEntry() error = %v, want invalid target", err)
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}

func TestTypeScriptRuntimeCommandWithLookPathPrefersRunners(t *testing.T) {
	cases := []struct {
		name      string
		available map[string]string
		wantPath  string
		wantArgs  []string
	}{
		{"bun", map[string]string{"bun": "/bin/bun"}, "/bin/bun", []string{"main.ts", "arg"}},
		{"deno", map[string]string{"deno": "/bin/deno"}, "/bin/deno", []string{"run", "main.ts", "arg"}},
		{"node", map[string]string{"node": "/bin/node"}, "/bin/node", []string{"main.ts", "arg"}},
		{"ts-node", map[string]string{"ts-node": "/bin/ts-node"}, "/bin/ts-node", []string{"--esm", "main.ts", "arg"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd, err := typeScriptRuntimeCommandWithLookPath("main.ts", []string{"arg"}, func(name string) (string, error) {
				path, ok := tc.available[name]
				if !ok {
					return "", fs.ErrNotExist
				}
				return path, nil
			})
			if err != nil {
				t.Fatalf("typeScriptRuntimeCommandWithLookPath() error = %v", err)
			}
			if cmd.Path != tc.wantPath {
				t.Fatalf("cmd.Path = %q, want %q", cmd.Path, tc.wantPath)
			}
			if got := cmd.Args[1:]; strings.Join(got, "\x00") != strings.Join(tc.wantArgs, "\x00") {
				t.Fatalf("cmd.Args = %v, want %v", cmd.Args, append([]string{tc.wantPath}, tc.wantArgs...))
			}
		})
	}
}

func TestTypeScriptRuntimeCommandWithLookPathErrorsWithoutRunner(t *testing.T) {
	_, err := typeScriptRuntimeCommandWithLookPath("main.ts", nil, func(string) (string, error) {
		return "", fs.ErrNotExist
	})
	if err == nil || !strings.Contains(err.Error(), "TypeScript backend requires") {
		t.Fatalf("typeScriptRuntimeCommandWithLookPath() error = %v, want runner requirement", err)
	}
}

func TestBuildEnvDefaultTarget(t *testing.T) {
	env, err := buildEnv("")
	if err != nil {
		t.Fatalf("buildEnv() error = %v", err)
	}
	if len(env) == 0 {
		t.Fatal("buildEnv() returned empty environment")
	}
}

func TestBuildEnvCrossCompileTarget(t *testing.T) {
	env, err := buildEnv("linux/amd64")
	if err != nil {
		t.Fatalf("buildEnv() error = %v", err)
	}
	joined := strings.Join(env, "\n")
	if !strings.Contains(joined, "GOOS=linux") || !strings.Contains(joined, "GOARCH=amd64") {
		t.Fatalf("buildEnv() = %v, want GOOS=linux GOARCH=amd64", env)
	}
}

func TestDefaultBinaryName(t *testing.T) {
	path := filepath.Join("/tmp", "demo.rn")
	got := defaultBinaryName(path)
	want := "demo"
	if runtime.GOOS == "windows" {
		want = "demo.exe"
	}
	if got != want {
		t.Fatalf("defaultBinaryName() = %q, want %q", got, want)
	}
}
