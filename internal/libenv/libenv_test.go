package libenv

import (
	"os"
	"path/filepath"
	"testing"
)

const esStub = `interface Console {
  log(message?: any, ...optionalParams: any[]): void;
  error(message?: any, ...optionalParams: any[]): void;
}

declare var console: Console;
declare var JSON: JSON;

declare function parseInt(string: string, radix?: number): number;
`

const domStub = `interface Console {
  table(data?: any): void;
}

interface Document {
  title: string;
}

declare var document: Document;
`

func writeLib(t *testing.T, dir, name, contents string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
	return path
}

func TestLoadMergesDeclarationsAndCapabilities(t *testing.T) {
	dir := t.TempDir()
	esPath := writeLib(t, dir, "lib.es5.d.ts", esStub)
	domPath := writeLib(t, dir, "lib.dom.d.ts", domStub)

	set, err := Load(esPath, domPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := set.Interface("Console"); got == nil || len(got.Members) != 3 {
		t.Fatalf("Console members = %#v, want log/error/table", got)
	}
	if got := set.Global("console"); got == nil {
		t.Fatalf("console global missing")
	}
	if got := set.Function("parseInt"); got == nil {
		t.Fatalf("parseInt declaration missing")
	}
	if got := set.CapabilityOf("Console"); got != CapDOM {
		t.Fatalf("Console capability = %q, want %q", got, CapDOM)
	}
	if got := set.CapabilityOf("parseInt"); got != CapESCore {
		t.Fatalf("parseInt capability = %q, want %q", got, CapESCore)
	}
	if len(set.Files) != 2 {
		t.Fatalf("Files = %#v, want two entries", set.Files)
	}
}

func TestLoadDefaultUsesEnvDirAndVersion(t *testing.T) {
	dir := t.TempDir()
	writeLib(t, dir, "lib.es5.d.ts", esStub)
	writeLib(t, dir, "VERSION", "6.0.3\n")
	t.Setenv("RUNE_TS_LIB_DIR", dir)

	set, err := LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault() error = %v", err)
	}
	if set.Version != "6.0.3" {
		t.Fatalf("Version = %q, want 6.0.3", set.Version)
	}
	if set.Interface("Console") == nil {
		t.Fatalf("Console declaration missing")
	}
}

func TestLoadDefaultReportsMissingLib(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("RUNE_TS_LIB_DIR", dir)
	if _, err := LoadDefault(); err == nil {
		t.Fatalf("LoadDefault() error = nil, want missing lib error")
	}
}

func TestCapabilityDefaultsToESCore(t *testing.T) {
	set := New()
	if got := set.CapabilityOf("Unknown"); got != CapESCore {
		t.Fatalf("unknown capability = %q, want %q", got, CapESCore)
	}
}
