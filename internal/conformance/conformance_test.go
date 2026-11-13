// Package conformance runs the cross-backend specification fixtures under
// tests/conformance. It is the executable part of RFC 0001's conformance
// matrix: every fixture states its expected stdout and the targets it covers.
package conformance

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/compiler"
	"github.com/oboard/rune-lang/internal/interpreter"
)

type fixture struct {
	path    string
	source  string
	expects []string
	targets []string
}

func TestConformanceFixtures(t *testing.T) {
	root := filepath.Join("..", "..", "tests", "conformance")
	var paths []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".rn" {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		t.Fatalf("walk conformance fixtures: %v", err)
	}
	if len(paths) == 0 {
		t.Fatalf("no conformance fixtures found under %s", root)
	}
	for _, path := range paths {
		path := path
		t.Run(relativeName(root, path), func(t *testing.T) {
			fx := loadFixture(t, path)
			want := strings.Join(fx.expects, "")
			for _, target := range fx.targets {
				target := target
				t.Run(target, func(t *testing.T) {
					got, err := runTarget(t, target, fx.path)
					if err != nil {
						t.Fatalf("%s: %v", target, err)
					}
					if got != want {
						t.Fatalf("%s output = %q, want %q", target, got, want)
					}
				})
			}
		})
	}
}

func relativeName(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return strings.TrimSuffix(filepath.ToSlash(rel), ".rn")
}

func loadFixture(t *testing.T, path string) fixture {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	fx := fixture{path: path, source: string(data)}
	for _, line := range strings.Split(fx.source, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "// expect:"):
			value := strings.TrimSpace(strings.TrimPrefix(line, "// expect:"))
			fx.expects = append(fx.expects, value+"\n")
		case strings.HasPrefix(line, "// targets:"):
			raw := strings.TrimSpace(strings.TrimPrefix(line, "// targets:"))
			for _, target := range strings.Split(raw, ",") {
				if target = strings.TrimSpace(target); target != "" {
					fx.targets = append(fx.targets, target)
				}
			}
		}
	}
	if len(fx.targets) == 0 {
		fx.targets = []string{"interpreter", "go"}
	}
	return fx
}

func runTarget(t *testing.T, target, path string) (string, error) {
	t.Helper()
	switch target {
	case "interpreter":
		return runInterpreter(t, path)
	case "go":
		return runGo(t, path)
	case "ts":
		return runTypeScript(t, path)
	case "mbt":
		return runMoonBit(t, path)
	default:
		t.Skipf("unknown conformance target %q", target)
		return "", nil
	}
}

func runInterpreter(t *testing.T, path string) (string, error) {
	t.Helper()
	prog, diags := compiler.AnalyzeFile(path)
	if len(diags) > 0 {
		t.Fatalf("interpreter diagnostics for %s: %v", path, diags)
	}
	var out bytes.Buffer
	runner := interpreter.New(prog.IR, interpreter.WithOutput(&out))
	if err := runner.RunMain(); err != nil {
		return "", err
	}
	return out.String(), nil
}

func runGo(t *testing.T, path string) (string, error) {
	t.Helper()
	source, diags := compiler.GenerateGoFile(path)
	if len(diags) > 0 {
		t.Fatalf("go codegen diagnostics for %s: %v", path, diags)
	}
	dir := t.TempDir()
	goFile := filepath.Join(dir, "main.go")
	if err := os.WriteFile(goFile, []byte(source), 0o644); err != nil {
		return "", err
	}
	cmd := exec.Command("go", "run", goFile)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", &commandError{target: "go run", err: err, output: string(out)}
	}
	return string(out), nil
}

func runTypeScript(t *testing.T, path string) (string, error) {
	t.Helper()
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not found; skipping TypeScript conformance target")
	}
	source, diags := compiler.GenerateTypeScriptFile(path)
	if len(diags) > 0 {
		t.Fatalf("ts codegen diagnostics for %s: %v", path, diags)
	}
	source += `
void (async () => {
  if (typeof main === "function") {
    await main();
  }
})();
`
	dir := t.TempDir()
	tsFile := filepath.Join(dir, "main.ts")
	if err := os.WriteFile(tsFile, []byte(source), 0o644); err != nil {
		return "", err
	}
	cmd := exec.Command(node, tsFile)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", &commandError{target: "node", err: err, output: string(out)}
	}
	return string(out), nil
}

func runMoonBit(t *testing.T, path string) (string, error) {
	t.Helper()
	moon, err := exec.LookPath("moon")
	if err != nil {
		t.Skip("moon not found; skipping MoonBit conformance target")
	}
	source, diags := compiler.GenerateMoonBitFile(path)
	if len(diags) > 0 {
		t.Fatalf("mbt codegen diagnostics for %s: %v", path, diags)
	}
	dir := t.TempDir()
	files := map[string]string{
		"main.mbt": source,
		"moon.mod": "name = \"oboard/rune_conformance\"\n",
		"moon.pkg": "warnings = \"-1-7-23-67\"\n\noptions(\n  \"is-main\": true,\n)\n",
	}
	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644); err != nil {
			return "", err
		}
	}
	cmd := exec.Command(moon, "run", "--target", "native", ".")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", &commandError{target: "moon run", err: err, output: string(out)}
	}
	return string(out), nil
}

type commandError struct {
	target string
	err    error
	output string
}

func (e *commandError) Error() string {
	return e.target + ": " + e.err.Error() + "\n" + e.output
}
