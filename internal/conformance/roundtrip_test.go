package conformance

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/astprint"
	"github.com/oboard/rune-lang/internal/checker"
	tscodegen "github.com/oboard/rune-lang/internal/codegen/typescript"
	"github.com/oboard/rune-lang/internal/ir"
	"github.com/oboard/rune-lang/internal/parser"
	"github.com/oboard/rune-lang/internal/stdlib"
	"github.com/oboard/rune-lang/internal/tsfront"
)

// TestTypeScriptRoundtripFixtures covers RFC 0001 L1: TypeScript source is
// converted to Rune, printed, re-parsed, and compiled back to TypeScript. The
// printed Rune must also be a fixed point of the printer.
func TestTypeScriptRoundtripFixtures(t *testing.T) {
	root := filepath.Join("..", "..", "tests", "conformance", "roundtrip")
	var paths []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".ts" {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		t.Fatalf("walk roundtrip fixtures: %v", err)
	}
	if len(paths) == 0 {
		t.Fatalf("no roundtrip fixtures found under %s", root)
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not found; skipping TypeScript roundtrip")
	}
	registry, err := stdlib.LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault() error = %v", err)
	}
	for _, path := range paths {
		path := path
		t.Run(strings.TrimSuffix(filepath.Base(path), ".ts"), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ReadFile(%s) error = %v", path, err)
			}
			want := strings.Join(fixtureLines(string(data), "// expect:"), "")
			runeFile, diags := tsfront.ConvertSource(path, string(data))
			if len(diags) > 0 {
				t.Fatalf("tsfront diagnostics: %v", diags)
			}
			printed := astprint.File(runeFile)
			reparsed, parseErrs := parser.Parse(printed)
			if len(parseErrs) > 0 {
				t.Fatalf("reparse errors: %v\n%s", parseErrs, printed)
			}
			if again := astprint.File(reparsed); again != printed {
				t.Fatalf("Rune printer is not a fixed point:\nfirst:\n%s\nsecond:\n%s", printed, again)
			}
			info, checkDiags := checker.CheckWithStdlib(reparsed, registry)
			if len(checkDiags) > 0 {
				t.Fatalf("check diagnostics: %v\n%s", checkDiags, printed)
			}
			tsSource, err := tscodegen.GenerateIR(ir.LowerFile(reparsed, info))
			if err != nil {
				t.Fatalf("ts codegen error = %v", err)
			}
			tsSource += "\nvoid (async () => {\n  if (typeof main === \"function\") {\n    await main();\n  }\n})();\n"
			dir := t.TempDir()
			tsFile := filepath.Join(dir, "main.ts")
			if err := os.WriteFile(tsFile, []byte(tsSource), 0o644); err != nil {
				t.Fatalf("WriteFile() error = %v", err)
			}
			var stdout, stderr bytes.Buffer
			cmd := exec.Command(node, tsFile)
			cmd.Dir = dir
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("node error = %v\n%s", err, stderr.String())
			}
			if got := stdout.String(); got != want {
				t.Fatalf("roundtrip output = %q, want %q\nRune:\n%s\nTS:\n%s", got, want, printed, tsSource)
			}
		})
	}
}

func fixtureLines(source, prefix string) []string {
	var out []string
	for _, line := range strings.Split(source, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			out = append(out, strings.TrimSpace(strings.TrimPrefix(line, prefix))+"\n")
		}
	}
	return out
}
