package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
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
	for _, backend := range []string{"go", "ts"} {
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

func TestRunProgramArgsUsesDashBoundary(t *testing.T) {
	args := []string{"main.rn", "-v", "target"}
	if got, want := runProgramArgs(args, 1), []string{"-v", "target"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("runProgramArgs() = %#v, want %#v", got, want)
	}
	if got, want := runProgramArgs(args, -1), []string{"-v", "target"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("runProgramArgs(no dash) = %#v, want %#v", got, want)
	}
	if got, want := runProgramArgs(args, 0), []string{"-v", "target"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("runProgramArgs(dash before path) = %#v, want %#v", got, want)
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
	if err := runEntry(mainPath, "go", []string{"-v", "target"}, strings.NewReader(""), &out, &errOut); err != nil {
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
	err := runEntry(mainPath, "go", nil, strings.NewReader(""), &out, &errOut)
	code, ok := exitCode(err)
	if !ok || code != 2 {
		t.Fatalf("runEntry() error = %v, exit code = %d/%v, want 2", err, code, ok)
	}
	if strings.Contains(errOut.String(), "exit status") {
		t.Fatalf("stderr = %q, want no go run exit wrapper", errOut.String())
	}
}

func TestFmtCommandHasFormatAlias(t *testing.T) {
	cmd := fmtCmd()
	for _, alias := range cmd.Aliases {
		if alias == "format" {
			return
		}
	}
	t.Fatalf("fmt aliases = %v, want format", cmd.Aliases)
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
