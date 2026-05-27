package main

import (
	"bytes"
	"os"
	"path/filepath"
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

func writeTestFile(t *testing.T, path string, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
