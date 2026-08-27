package main

import (
	"bytes"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/compiler"
)

func TestParseTarget(t *testing.T) {
	goos, goarch, err := parseTarget("linux-amd64")
	if err != nil {
		t.Fatalf("parseTarget() error = %v", err)
	}
	if goos != "linux" || goarch != "amd64" {
		t.Fatalf("parseTarget() = %q, %q", goos, goarch)
	}
}

func TestParseTargetRejectsInvalidTarget(t *testing.T) {
	for _, target := range []string{"", "linux", "linux-", "-amd64", "linux-amd64-extra"} {
		if _, _, err := parseTarget(target); err == nil {
			t.Fatalf("parseTarget(%q) succeeded, want error", target)
		}
	}
}

func TestValidateBackend(t *testing.T) {
	for _, backend := range []string{"go", "ts", "mbt"} {
		if err := validateBackend(backend); err != nil {
			t.Fatalf("validateBackend(%q) error = %v", backend, err)
		}
	}
	if err := validateBackend("js"); err == nil {
		t.Fatal("validateBackend(\"js\") succeeded, want error")
	}
}

func TestSelectRunBackendDefaultsToTypeScriptForTypeScriptImports(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "greet.ts"), "export function greet(name: string): string { return name }\n")
	mainPath := filepath.Join(dir, "main.rn")
	writeTestFile(t, mainPath, `@"greet.ts"

main() => @io.println(greet("Rune"))
`)

	if got := selectRunBackend(mainPath, "go", false); got != "ts" {
		t.Fatalf("selectRunBackend() = %q, want ts", got)
	}
	if got := selectRunBackend(mainPath, "go", true); got != "go" {
		t.Fatalf("selectRunBackend(explicit go) = %q, want go", got)
	}
	if got := selectRunBackend(mainPath, "ts", true); got != "ts" {
		t.Fatalf("selectRunBackend(explicit ts) = %q, want ts", got)
	}
}

func TestSelectRunBackendKeepsGoForRuneOnlyProgram(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.rn")
	writeTestFile(t, mainPath, "main() => @io.println(\"Rune\")\n")

	if got := selectRunBackend(mainPath, "go", false); got != "go" {
		t.Fatalf("selectRunBackend() = %q, want go", got)
	}
}

func TestRunRuneCLIUsesSelfhostInterpreter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.rn")
	writeTestFile(t, path, `main() => {
  values := [1, 2, 3]
  doubled := values.map((value) => value * 2)
  @io.println(doubled.reduce(0, (sum, value) => sum + value))
}
`)

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := runRuneCLI([]string{"run", path}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("runRuneCLI(run) error = %v, stderr = %s", err, errOut.String())
	}
	if got, want := strings.TrimSpace(out.String()), "12"; got != want {
		t.Fatalf("selfhost interpreter run output = %q, want %q", got, want)
	}
}

func TestRunEntryGoForwardsProgramArgs(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.rn")
	writeTestFile(t, mainPath, `main() => {
  argv := @process.argv()
  @io.println(argv[1])
  @io.println(argv[2])
}
`)

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := runEntry(mainPath, "go", "", []string{"-v", "target"}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("runEntry() error = %v, stderr = %s", err, errOut.String())
	}
	if got, want := out.String(), "-v\ntarget\n"; got != want {
		t.Fatalf("runEntry() output = %q, want %q", got, want)
	}
}

func TestRunEntryGoPreservesProgramExitCode(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.rn")
	writeTestFile(t, mainPath, "main() => @process.exit(2)\n")

	var out bytes.Buffer
	var errOut bytes.Buffer
	err := runEntry(mainPath, "go", "", nil, strings.NewReader(""), &out, &errOut)
	code, ok := exitCode(err)
	if !ok || code != 2 {
		t.Fatalf("runEntry() error = %v, exit code = %d/%v, want 2", err, code, ok)
	}
	if strings.Contains(errOut.String(), "exit status") {
		t.Fatalf("stderr = %q, want no go run exit wrapper", errOut.String())
	}
}

func TestGeneratedCLIParsesFormatAlias(t *testing.T) {
	invocation := __parseCli([]string{"format", "--check", "main.rn"})
	if !invocation.__ok {
		t.Fatalf("__parseCli() errors = %v", invocation.__errors)
	}
	if invocation.__command != "fmt" || !invocation.__checkOnly || invocation.__path != "main.rn" {
		t.Fatalf("__parseCli(format) = %#v", invocation)
	}
}

func TestSelfhostCLIGeneratedGoIsCurrent(t *testing.T) {
	root := repoRootForCommandTest(t)
	cmd := exec.Command("go", "run", "./cmd/rune", "go", "selfhost/cli/cli.rn")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generate selfhost CLI failed: %v\n%s", err, out)
	}
	got := normalizeGeneratedCLIForHost(string(out))
	wantPath := filepath.Join(root, "cmd", "rune", "selfhost_cli_gen.go")
	want, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", wantPath, err)
	}
	if got != normalizeGeneratedCLIForHost(string(want)) {
		t.Fatalf("selfhost_cli_gen.go is stale; regenerate it with rune go selfhost/cli/cli.rn")
	}
}

func TestGeneratedSelfhostCLIStandaloneParsesArgs(t *testing.T) {
	root := repoRootForCommandTest(t)
	cmd := exec.Command("go", "run", "./cmd/rune", "go", "selfhost/cli/cli.rn")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generate selfhost CLI failed: %v\n%s", err, out)
	}
	formatted, err := format.Source(out)
	if err != nil {
		t.Fatalf("format generated CLI error = %v\n%s", err, out)
	}
	dir := t.TempDir()
	goFile := filepath.Join(dir, "main.go")
	if err := os.WriteFile(goFile, formatted, 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", goFile, err)
	}
	run := exec.Command("go", "run", goFile, "--backend=go", "format", "--check", "main.rn")
	run.Dir = root
	got, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("generated CLI run failed: %v\n%s", err, got)
	}
	if string(got) != "fmt\n" {
		t.Fatalf("generated CLI output = %q, want fmt", got)
	}
}

func TestSelfhostCompilerGeneratedGoIsCurrent(t *testing.T) {
	root := repoRootForCommandTest(t)
	cmd := exec.Command("go", "run", "./cmd/rune", "go", "selfhost/compiler/compiler.rn")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generate selfhost compiler failed: %v\n%s", err, out)
	}
	wantPath := filepath.Join(root, "cmd", "rune", "selfhost_compiler_gen.go")
	want, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", wantPath, err)
	}
	formatted, err := format.Source(out)
	if err != nil {
		t.Fatalf("format generated compiler error = %v", err)
	}
	if string(formatted) != string(want) {
		t.Fatalf("selfhost_compiler_gen.go is stale; regenerate it with rune go selfhost/compiler/compiler.rn")
	}
}

func TestSelfhostFormatterGeneratedGoIsCurrent(t *testing.T) {
	root := repoRootForCommandTest(t)
	cmd := exec.Command("go", "run", "./cmd/rune", "go", "selfhost/format/format.rn")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generate selfhost formatter failed: %v\n%s", err, out)
	}
	formatted, err := format.Source(out)
	if err != nil {
		t.Fatalf("format generated formatter error = %v", err)
	}
	wantPath := filepath.Join(root, "cmd", "rune", "selfhost_format_gen.go")
	want, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", wantPath, err)
	}
	if got := strings.ReplaceAll(string(formatted), "__", "__fmt_"); got != string(want) {
		t.Fatalf("selfhost_format_gen.go is stale; regenerate it with rune go selfhost/format/format.rn")
	}
}

func TestRuneCLIRunUsesSelfhostCompiler(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.rn")
	writeTestFile(t, path, "main() => @io.println(42)\n")

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := runRuneCLI([]string{"run", path}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("runRuneCLI(run) error = %v, stderr = %s", err, errOut.String())
	}
	if got, want := out.String(), "42\n"; got != want {
		t.Fatalf("runRuneCLI(run) output = %q, want %q", got, want)
	}
}

func TestRuneCLIMoonBitUsesSelfhostCompilerImportGraph(t *testing.T) {
	dir := t.TempDir()
	entry := filepath.Join(dir, "main.rn")
	writeTestFile(t, filepath.Join(dir, "helper.rn"), "+ answer() -> Int => 42\n")
	writeTestFile(t, entry, "@\"helper.rn\"\nmain() => @io.println(answer())\n")

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := runRuneCLI([]string{"mbt", entry}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("runRuneCLI(mbt import graph) error = %v, stderr = %s", err, errOut.String())
	}
	if got := out.String(); !strings.Contains(got, "fn __answer() -> Int") {
		t.Fatalf("self-host generated MoonBit = %q, want imported answer", got)
	}
}

func TestRuneCLIMoonBitUsesSelfhostCompiler(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.rn")
	writeTestFile(t, path, "main() => @io.println(42)\n")

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := runRuneCLI([]string{"mbt", path}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("runRuneCLI(mbt) error = %v, stderr = %s", err, errOut.String())
	}
	if got := out.String(); !strings.Contains(got, "println((42).to_string())") {
		t.Fatalf("self-host generated MoonBit = %q, want generated println", got)
	}
}

func TestRuneCLITypeScriptDeclarationUsesSelfhostCompiler(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.rn")
	writeTestFile(t, path, `+ User: {
  name: String
  active: Bool
}

+ makeUser(name: String) -> User => {
  name: name,
  active: true
}

main() => @io.println(makeUser("Rune").name)
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
	if !strings.Contains(got, "export { __User as User, __makeUser as makeUser };") {
		t.Fatalf("selfhost declaration missing export aliases: %q", got)
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
	result := __compileGo(source)
	if !result.__ok {
		t.Fatalf("__compileGo() errors = %v", result.__errors)
	}
	if !strings.Contains(result.__output, "func __fib(__n int) int") ||
		!strings.Contains(result.__output, "} else if __n == 1 {") {
		t.Fatalf("__compileGo() output did not lower pattern function:\n%s", result.__output)
	}
}

func TestGeneratedSelfhostCompilerBuildsNullCoalesceGo(t *testing.T) {
	assertGeneratedSelfhostGoBuilds(t, `fallback(value: String?) -> String => value ?? "missing"

main() => @io.println(fallback(null))
	`)
}

func TestGeneratedSelfhostCompilerBuildsCollectionMethodsGo(t *testing.T) {
	assertGeneratedSelfhostGoBuilds(t, `main() => {
  values := ["go"]
  values.push("ts")
  names := @map.new("", "")
  names.set("backend", values.at(0))
  backend := names.getOr("backend", "go")
  flags := @map.new("", false)
  flags.set("check", true)
  check := flags.getOr("check", false)
  @io.println(backend, check)
}
		`)
}

func TestGeneratedSelfhostCompilerBuildsTypedObjectGo(t *testing.T) {
	assertGeneratedSelfhostGoBuilds(t, `User: {
  name: String
  active: Bool
}

makeUser(name: String) -> User => {
  name: name,
  active: true
}

main() => @io.println(makeUser("Rune").name)
		`)
}

func TestGeneratedSelfhostCompilerBuildsJSONObjectGo(t *testing.T) {
	assertGeneratedSelfhostGoBuilds(t, `#json.object
User: {
  #json.name("display_name")
  name: String
  #json.ignore
  password: String
}

main() => {
  user := User::fromJson("{\"display_name\":\"Ada\",\"password\":\"drop\"}")
  @io.println(user.name)
}
		`)
}

func assertGeneratedSelfhostGoBuilds(t *testing.T, source string) {
	t.Helper()
	result := __compileGo(source)
	if !result.__ok {
		t.Fatalf("__compileGo() errors = %v", result.__errors)
	}
	assertGeneratedGoBuilds(t, result.__output)
}

func assertGeneratedGoBuilds(t *testing.T, source string) {
	t.Helper()
	formatted, err := format.Source([]byte(source))
	if err != nil {
		t.Fatalf("format generated Go error = %v\n%s", err, source)
	}
	dir := t.TempDir()
	goFile := filepath.Join(dir, "main.go")
	if err := os.WriteFile(goFile, formatted, 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", goFile, err)
	}
	cmd := exec.Command("go", "test", goFile)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generated Go build failed: %v\n%s\n%s", err, out, formatted)
	}
}

func normalizeGeneratedCLIForHost(src string) string {
	src = strings.ReplaceAll(src, "func __main()", "func selfhostCliGeneratedMain()")
	src = strings.Replace(src, "func main() {\n\t__main()\n}\n", "func selfhostCliGeneratedEntrypoint() {\n\tselfhostCliGeneratedMain()\n}\n", 1)
	formatted, err := format.Source([]byte(src))
	if err != nil {
		return src
	}
	return normalizeGeneratedPrivateSymbols(string(formatted))
}

func normalizeGeneratedPrivateSymbols(src string) string {
	return regexp.MustCompile(`____rune_private_[0-9a-f]+_`).ReplaceAllString(src, "____rune_private_HASH_")
}

func TestRunRuneCLIForwardsProgramArgs(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.rn")
	writeTestFile(t, mainPath, `main() => {
  argv := @process.argv()
  @io.println(argv[1])
  @io.println(argv[2])
}
`)

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := runRuneCLI([]string{"run", mainPath, "--", "-v", "target"}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("runRuneCLI() error = %v, stderr = %s", err, errOut.String())
	}
	if got, want := out.String(), "-v\ntarget\n"; got != want {
		t.Fatalf("runRuneCLI() output = %q, want %q", got, want)
	}
}

func TestRunRuneCLICheckUsesSelfhostMigrationBridge(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.rn")
	writeTestFile(t, path, "main() => 42\n")

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := runRuneCLI([]string{"check", path}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("runRuneCLI(check) error = %v, stderr = %s", err, errOut.String())
	}
	if got, want := out.String(), "ok "+path+"\n"; got != want {
		t.Fatalf("runRuneCLI(check) output = %q, want %q", got, want)
	}
}

func TestRunRuneCLICheckReportsSelfhostErrors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.rn")
	writeTestFile(t, path, "main() -> Int => \"wrong\"\n")

	var out bytes.Buffer
	var errOut bytes.Buffer
	err := runRuneCLI([]string{"check", path}, strings.NewReader(""), &out, &errOut)
	if err == nil || err.Error() != "check failed" {
		t.Fatalf("runRuneCLI(check) error = %v, want check failed", err)
	}
	if got, want := errOut.String(), path+": function \"main\" returns String, expected Int\n"; got != want {
		t.Fatalf("runRuneCLI(check) stderr = %q, want %q", got, want)
	}
	if got := out.String(); got != "" {
		t.Fatalf("runRuneCLI(check) stdout = %q, want empty", got)
	}
}

func TestCheckTargetAcceptsStdlibRoot(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "prelude.rn"), "@foo\n")
	writeTestFile(t, filepath.Join(dir, "foo", "foo.rn"), "Foo: {}\n")

	var out bytes.Buffer
	if err := checkTarget(dir, &out); err != nil {
		t.Fatalf("checkTarget() error = %v", err)
	}
	if got := out.String(); got != "ok "+dir+"\n" {
		t.Fatalf("checkTarget() output = %q", got)
	}
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

	got, diags, err := resolveRunEntry(dir)
	if err != nil {
		t.Fatalf("resolveRunEntry() error = %v", err)
	}
	if len(diags) > 0 {
		t.Fatalf("resolveRunEntry() diagnostics = %v", diags)
	}
	if got != mainPath {
		t.Fatalf("resolveRunEntry() = %q, want %q", got, mainPath)
	}
}

func TestResolveRunEntryRejectsMultipleMains(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "a.rn"), "main() => 1\n")
	writeTestFile(t, filepath.Join(dir, "nested", "b.rn"), "main() => 2\n")

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
	want := `from "./greet.ts"`
	if !strings.Contains(string(data), want) {
		t.Fatalf("runtime TypeScript = %q, want %q", data, want)
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
	err := runEntry(mainPath, "mbt", "linux-amd64", nil, strings.NewReader(""), &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), `invalid MoonBit target "linux-amd64"`) {
		t.Fatalf("runEntry() error = %v, want invalid MoonBit target", err)
	}
}

func TestValidateMoonBitTarget(t *testing.T) {
	for _, target := range []string{"wasm", "wasm-gc", "js", "native", "llvm", "all"} {
		if err := validateMoonBitTarget(target); err != nil {
			t.Fatalf("validateMoonBitTarget(%q) error = %v", target, err)
		}
	}
	if err := validateMoonBitTarget("linux-amd64"); err == nil {
		t.Fatal("validateMoonBitTarget(\"linux-amd64\") succeeded, want error")
	}
}

func TestTypeScriptRuntimeSpecifier(t *testing.T) {
	cases := map[string]string{
		"greet.ts":       "./greet.ts",
		"lib/greet.ts":   "./lib/greet.ts",
		"./greet.ts":     "./greet.ts",
		"../greet.ts":    "../greet.ts",
		"https://x/y.ts": "https://x/y.ts",
	}
	for input, want := range cases {
		if got := typeScriptRuntimeSpecifier(input); got != want {
			t.Fatalf("typeScriptRuntimeSpecifier(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestTypeScriptRuntimeCommandSupportsAvailableRuntimes(t *testing.T) {
	cases := []struct {
		name string
		want []string
	}{
		{name: "bun", want: []string{"/bin/bun", "main.ts", "arg"}},
		{name: "deno", want: []string{"/bin/deno", "run", "main.ts", "arg"}},
		{name: "node", want: []string{"/bin/node", "main.ts", "arg"}},
		{name: "ts-node", want: []string{"/bin/ts-node", "--esm", "main.ts", "arg"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd, err := typeScriptRuntimeCommandWithLookPath("main.ts", []string{"arg"}, func(name string) (string, error) {
				if name == tc.name {
					return "/bin/" + name, nil
				}
				return "", exec.ErrNotFound
			})
			if err != nil {
				t.Fatalf("typeScriptRuntimeCommandWithLookPath() error = %v", err)
			}
			if !reflect.DeepEqual(cmd.Args, tc.want) {
				t.Fatalf("command args = %#v, want %#v", cmd.Args, tc.want)
			}
		})
	}
}

func writeTestFile(t *testing.T, path string, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func repoRootForCommandTest(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root not found")
		}
		dir = parent
	}
}
