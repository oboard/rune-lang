package conformance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/tsfront"
)

// TestNegativeProfileFixtures freezes RFC 0001 D7: every dynamic construct
// must be rejected with the documented stable code instead of being lowered
// into an approximation.
func TestNegativeProfileFixtures(t *testing.T) {
	root := filepath.Join("..", "..", "tests", "conformance", "negative")
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
		t.Fatalf("walk negative fixtures: %v", err)
	}
	if len(paths) == 0 {
		t.Fatalf("no negative fixtures found under %s", root)
	}
	for _, path := range paths {
		path := path
		t.Run(strings.TrimSuffix(filepath.Base(path), ".ts"), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ReadFile(%s) error = %v", path, err)
			}
			want := fixtureCodes(string(data))
			if len(want) == 0 {
				t.Fatalf("%s: missing // code: marker", path)
			}
			_, diags := tsfront.ConvertSource(path, string(data))
			if len(diags) == 0 {
				t.Fatalf("%s: expected rejection with %v, got none", path, want)
			}
			got := map[string]bool{}
			for _, diag := range diags {
				if diag.Code == "" {
					t.Errorf("%s:%d: diagnostic %q has no profile code", path, diag.Line, diag.Message)
					continue
				}
				got[diag.Code] = true
			}
			for _, code := range want {
				if !got[code] {
					t.Errorf("%s: missing diagnostic %s; got %v", path, code, diagCodes(diags))
				}
			}
		})
	}
}

func fixtureCodes(source string) []string {
	var codes []string
	for _, line := range strings.Split(source, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "// code:") {
			code := strings.TrimSpace(strings.TrimPrefix(line, "// code:"))
			if code != "" {
				codes = append(codes, code)
			}
		}
	}
	return codes
}

func diagCodes(diags []tsfront.Diagnostic) []string {
	codes := make([]string, 0, len(diags))
	for _, diag := range diags {
		codes = append(codes, diag.Code)
	}
	return codes
}
