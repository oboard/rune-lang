package lsp

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/compiler"
)

func TestInitializeAdvertisesDocumentClose(t *testing.T) {
	var out bytes.Buffer
	s := &server{out: &out, docs: map[string]string{}}
	if err := s.handle(request{ID: json.RawMessage(`1`), Method: "initialize"}); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, `"openClose":true`) {
		t.Fatalf("initialize response = %s, want openClose sync", got)
	}
	if !strings.Contains(got, `"change":1`) {
		t.Fatalf("initialize response = %s, want full text sync", got)
	}
}

func TestDidCloseDropsCachedDocument(t *testing.T) {
	uri := "file:///tmp/main.rn"
	var out bytes.Buffer
	s := &server{out: &out, docs: map[string]string{uri: "main() => 1"}, cache: map[string]programCacheEntry{}}
	s.cache[uri] = programCacheEntry{text: "main() => 1"}
	params := map[string]any{
		"textDocument": map[string]any{"uri": uri},
	}
	body, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.handle(request{Method: "textDocument/didClose", Params: body}); err != nil {
		t.Fatalf("didClose: %v", err)
	}
	if _, ok := s.docs[uri]; ok {
		t.Fatalf("closed document is still cached")
	}
	if len(s.cache) != 0 {
		t.Fatalf("program cache after close = %d, want 0", len(s.cache))
	}
	got := out.String()
	if !strings.Contains(got, `"diagnostics":[]`) {
		t.Fatalf("didClose notification = %s, want cleared diagnostics", got)
	}
}

func TestShutdownReleasesCachedDocuments(t *testing.T) {
	var out bytes.Buffer
	s := &server{out: &out, docs: map[string]string{
		"file:///tmp/a.rn": "main() => 1",
		"file:///tmp/b.rn": "main() => 2",
	}, cache: map[string]programCacheEntry{"file:///tmp/a.rn": {text: "main() => 1"}}}
	if err := s.handle(request{ID: json.RawMessage(`1`), Method: "shutdown"}); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
	if len(s.docs) != 0 {
		t.Fatalf("docs after shutdown = %d, want 0", len(s.docs))
	}
	if len(s.cache) != 0 {
		t.Fatalf("program cache after shutdown = %d, want 0", len(s.cache))
	}
}

func TestProgramCacheInvalidatesOnDocumentChange(t *testing.T) {
	uri := "file:///tmp/main.rn"
	s := &server{docs: map[string]string{}, cache: map[string]programCacheEntry{}}
	s.setDocument(uri, "main() => 1")
	if _, diags := s.analyze(uri); len(diags) != 0 {
		t.Fatalf("AnalyzeSource() diagnostics = %#v", diags)
	}
	if len(s.cache) != 1 {
		t.Fatalf("program cache entries = %d, want 1", len(s.cache))
	}

	s.setDocument(uri, "main() => missing")
	if len(s.cache) != 0 {
		t.Fatalf("program cache after change = %d, want 0", len(s.cache))
	}
	_, diags := s.analyze(uri)
	if len(diags) == 0 || !strings.Contains(diags[0].Message, `undefined name "missing"`) {
		t.Fatalf("diagnostics after change = %#v, want undefined name", diags)
	}
}

func TestAnalyzeSkipsSelfhostPrecheckForBootstrapSources(t *testing.T) {
	prevCheck := selfhostCheckSource
	prevAnalyze := analyzeSource
	t.Cleanup(func() {
		selfhostCheckSource = prevCheck
		analyzeSource = prevAnalyze
	})
	selfhostCheckSource = func(string, string) SelfhostCompileResult {
		return SelfhostCompileResult{Errors: []string{"bootstrap checker error"}}
	}
	analyzeSource = func(path string, text string) (*compiler.Program, []compiler.Diagnostic) {
		return nil, nil
	}
	for _, tc := range []struct {
		uri  string
		text string
	}{
		{uri: "file:///workspace/core/cli/cli.rn", text: "main() => 1"},
		{uri: "file:///workspace/selfhost/cli/cli.rn", text: "main() => 1"},
		{uri: "file:///workspace/tests/compiler_bootstrap.rn", text: "@\"../selfhost/compiler/compiler.rn\"\n\n? \"ok\" {\n  result := compileGo(`main() => 1`)\n}\n"},
	} {
		s := &server{docs: map[string]string{tc.uri: tc.text}, cache: map[string]programCacheEntry{}}
		if _, diags := s.analyze(tc.uri); len(diags) != 0 {
			t.Fatalf("analyze(%s) diagnostics = %#v, want none", tc.uri, diags)
		}
	}
}

func TestDiagnosticsHandlesCompilerBootstrapImports(t *testing.T) {
	root := repoRootForLSPTest(t)
	path := filepath.Join(root, "tests", "compiler_bootstrap.rn")
	uri := "file://" + filepath.ToSlash(path)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}

	prevCheck := selfhostCheckSource
	selfhostCheckSource = func(string, string) SelfhostCompileResult {
		return SelfhostCompileResult{Ok: false, Errors: []string{"undefined function compileGo"}}
	}
	defer func() { selfhostCheckSource = prevCheck }()

	s := NewSession()
	s.SetDocument(uri, string(data))
	for _, diag := range s.Diagnostics(uri) {
		message, _ := diag["message"].(string)
		if strings.Contains(message, "compileTypeScript") || strings.Contains(message, "checkSource") || strings.Contains(message, "compileGo") || strings.Contains(message, "SourceFile") {
			t.Fatalf("Diagnostics(%s) = %#v, want imported compiler declarations resolved", uri, s.Diagnostics(uri))
		}
	}
}

func repoRootForLSPTest(t *testing.T) string {
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
			t.Fatalf("could not find repo root from %s", dir)
		}
		dir = parent
	}
}

func TestDiagnosticsIncludesWarnings(t *testing.T) {
	uri := "file:///tmp/main.rn"
	src := `choose(aaa: Bool, bbb: Int, ccc: Int) -> Int => (aaa) {
  true => bbb
  _ => ccc
}
`
	s := NewSession()
	s.SetDocument(uri, src)
	diags := s.Diagnostics(uri)
	if len(diags) == 0 {
		t.Fatalf("Diagnostics() = none, want warning")
	}
	found := false
	for _, diag := range diags {
		if strings.Contains(diag["message"].(string), "prefer_ternary") {
			found = true
			if diag["severity"] != 2 {
				t.Fatalf("warning severity = %v, want 2", diag["severity"])
			}
			rangeValue := diag["range"].(map[string]any)
			start := rangeValue["start"].(position)
			if start.Line != 0 || start.Character != 49 {
				t.Fatalf("warning start = %d:%d, want 0:49", start.Line, start.Character)
			}
		}
	}
	if !found {
		t.Fatalf("Diagnostics() = %#v, want prefer_ternary warning", diags)
	}
}

func TestDiagnosticsReportMissingStructLiteralComma(t *testing.T) {
	uri := "file:///tmp/main.rn"
	src := `User: {
  name: String
  age: Int
}

main() => User {
  name: "oboard"
  age: 42
}
`
	s := NewSession()
	s.SetDocument(uri, src)
	diags := s.Diagnostics(uri)
	if len(diags) == 0 {
		t.Fatalf("Diagnostics() = none, want missing comma")
	}
	if got := diags[0]["message"]; got != "expected ',' between struct literal fields" {
		t.Fatalf("Diagnostics()[0] = %v, want missing comma", got)
	}
}

func TestAnalyzeUsesOverrideHook(t *testing.T) {
	uri := "file:///tmp/main.rn"
	called := 0
	prevAnalyze := analyzeSource
	prevCheck := selfhostCheckSource
	analyzeSource = func(path string, text string) (*compiler.Program, []compiler.Diagnostic) {
		called++
		if path != uri {
			t.Fatalf("analyze path = %q, want %q", path, uri)
		}
		if text != "main() => 1" {
			t.Fatalf("analyze text = %q, want main source", text)
		}
		return &compiler.Program{}, nil
	}
	selfhostCheckSource = nil
	defer func() {
		analyzeSource = prevAnalyze
		selfhostCheckSource = prevCheck
	}()

	s := &server{docs: map[string]string{uri: "main() => 1"}, cache: map[string]programCacheEntry{}}
	if _, diags := s.analyze(uri); len(diags) != 0 {
		t.Fatalf("analyze() diagnostics = %#v", diags)
	}
	if called != 1 {
		t.Fatalf("analyze override calls = %d, want 1", called)
	}
	if _, diags := s.analyze(uri); len(diags) != 0 {
		t.Fatalf("cached analyze() diagnostics = %#v", diags)
	}
	if called != 1 {
		t.Fatalf("cached analyze override calls = %d, want 1", called)
	}
}

func TestAnalyzeUsesSelfhostCheckOverride(t *testing.T) {
	uri := "file:///tmp/main.rn"
	prevAnalyze := analyzeSource
	prevCheck := selfhostCheckSource
	called := 0
	analyzeSource = func(path string, text string) (*compiler.Program, []compiler.Diagnostic) {
		called++
		return &compiler.Program{}, nil
	}
	selfhostCheckSource = func(source string, path string) SelfhostCompileResult {
		if source != "broken" {
			t.Fatalf("check source = %q, want broken", source)
		}
		if path != uri {
			t.Fatalf("check path = %q, want %q", path, uri)
		}
		return SelfhostCompileResult{Ok: false, Errors: []string{"selfhost parse failed"}}
	}
	defer func() {
		analyzeSource = prevAnalyze
		selfhostCheckSource = prevCheck
	}()

	s := &server{docs: map[string]string{uri: "broken"}, cache: map[string]programCacheEntry{}}
	prog, diags := s.analyze(uri)
	if prog != nil {
		t.Fatalf("analyze() program = %#v, want nil on selfhost failure", prog)
	}
	if called != 0 {
		t.Fatalf("host analyze calls = %d, want 0 when selfhost check fails", called)
	}
	if len(diags) != 1 || diags[0].Message != "selfhost parse failed" {
		t.Fatalf("analyze() diagnostics = %#v, want selfhost error", diags)
	}
}

func TestDiagnosticsToLSPMapsSelfhostDiagnosticDefaults(t *testing.T) {
	diags := diagnosticsToLSP([]compiler.Diagnostic{{Message: "selfhost parse failed"}})
	if len(diags) != 1 {
		t.Fatalf("diagnosticsToLSP() len = %d, want 1", len(diags))
	}
	if got := diags[0]["message"]; got != "selfhost parse failed" {
		t.Fatalf("diagnosticsToLSP()[0].message = %v, want selfhost parse failed", got)
	}
	if got := diags[0]["severity"]; got != 1 {
		t.Fatalf("diagnosticsToLSP()[0].severity = %v, want 1", got)
	}
}

func TestDiagnosticsReportSelfhostPrecheckFailure(t *testing.T) {
	uri := "file:///tmp/main.rn"
	prevAnalyze := analyzeSource
	prevCheck := selfhostCheckSource
	analyzeSource = func(path string, text string) (*compiler.Program, []compiler.Diagnostic) {
		return &compiler.Program{}, nil
	}
	selfhostCheckSource = func(source string, path string) SelfhostCompileResult {
		return SelfhostCompileResult{Ok: false, Errors: []string{"selfhost parse failed"}}
	}
	defer func() {
		analyzeSource = prevAnalyze
		selfhostCheckSource = prevCheck
	}()

	s := NewSession()
	s.SetDocument(uri, "broken")
	diags := s.Diagnostics(uri)
	if len(diags) != 1 {
		t.Fatalf("Diagnostics() len = %d, want 1", len(diags))
	}
	if got := diags[0]["message"]; got != "selfhost parse failed" {
		t.Fatalf("Diagnostics()[0].message = %v, want selfhost parse failed", got)
	}
}

func TestDefinitionSkipsHostAnalyzeWhenSelfhostPrecheckFails(t *testing.T) {
	uri := "file:///tmp/main.rn"
	prevAnalyze := analyzeSource
	prevCheck := selfhostCheckSource
	called := 0
	analyzeSource = func(path string, text string) (*compiler.Program, []compiler.Diagnostic) {
		called++
		return &compiler.Program{}, nil
	}
	selfhostCheckSource = func(source string, path string) SelfhostCompileResult {
		return SelfhostCompileResult{Ok: false, Errors: []string{"selfhost parse failed"}}
	}
	defer func() {
		analyzeSource = prevAnalyze
		selfhostCheckSource = prevCheck
	}()

	s := &server{docs: map[string]string{uri: "broken"}, cache: map[string]programCacheEntry{}}
	if got := s.definition(uri, position{}); got != nil {
		t.Fatalf("definition() = %#v, want nil on selfhost precheck failure", got)
	}
	if called != 0 {
		t.Fatalf("host analyze calls = %d, want 0 when selfhost precheck fails", called)
	}
}

func TestHoverSkipsHostAnalyzeWhenSelfhostPrecheckFails(t *testing.T) {
	uri := "file:///tmp/main.rn"
	prevAnalyze := analyzeSource
	prevCheck := selfhostCheckSource
	called := 0
	analyzeSource = func(path string, text string) (*compiler.Program, []compiler.Diagnostic) {
		called++
		return &compiler.Program{}, nil
	}
	selfhostCheckSource = func(source string, path string) SelfhostCompileResult {
		return SelfhostCompileResult{Ok: false, Errors: []string{"selfhost parse failed"}}
	}
	defer func() {
		analyzeSource = prevAnalyze
		selfhostCheckSource = prevCheck
	}()

	s := &server{docs: map[string]string{uri: "broken"}, cache: map[string]programCacheEntry{}}
	if got := s.hover(uri, position{}); got != nil {
		t.Fatalf("hover() = %#v, want nil on selfhost precheck failure", got)
	}
	if called != 0 {
		t.Fatalf("host analyze calls = %d, want 0 when selfhost precheck fails", called)
	}
}

func TestCompletionSkipsHostAnalyzeWhenSelfhostPrecheckFails(t *testing.T) {
	uri := "file:///tmp/main.rn"
	prevAnalyze := analyzeSource
	prevCheck := selfhostCheckSource
	called := 0
	analyzeSource = func(path string, text string) (*compiler.Program, []compiler.Diagnostic) {
		called++
		return &compiler.Program{}, nil
	}
	selfhostCheckSource = func(source string, path string) SelfhostCompileResult {
		return SelfhostCompileResult{Ok: false, Errors: []string{"selfhost parse failed"}}
	}
	defer func() {
		analyzeSource = prevAnalyze
		selfhostCheckSource = prevCheck
	}()

	s := &server{docs: map[string]string{uri: "broken"}, cache: map[string]programCacheEntry{}}
	items := s.completion(uri, position{}).([]map[string]any)
	if len(items) != 0 {
		t.Fatalf("completion() = %#v, want empty on selfhost precheck failure", items)
	}
	if called != 0 {
		t.Fatalf("host analyze calls = %d, want 0 when selfhost precheck fails", called)
	}
}

func TestRenameSkipsHostAnalyzeWhenSelfhostPrecheckFails(t *testing.T) {
	uri := "file:///tmp/main.rn"
	prevAnalyze := analyzeSource
	prevCheck := selfhostCheckSource
	called := 0
	analyzeSource = func(path string, text string) (*compiler.Program, []compiler.Diagnostic) {
		called++
		return &compiler.Program{}, nil
	}
	selfhostCheckSource = func(source string, path string) SelfhostCompileResult {
		return SelfhostCompileResult{Ok: false, Errors: []string{"selfhost parse failed"}}
	}
	defer func() {
		analyzeSource = prevAnalyze
		selfhostCheckSource = prevCheck
	}()

	s := &server{docs: map[string]string{uri: "brokenName"}, cache: map[string]programCacheEntry{}}
	if got := s.rename(uri, position{Line: 0, Character: 1}, "newName"); got != nil {
		t.Fatalf("rename() = %#v, want nil on selfhost precheck failure", got)
	}
	if called != 0 {
		t.Fatalf("host analyze calls = %d, want 0 when selfhost precheck fails", called)
	}
}

func TestSemanticTokensSkipHostAnalyzeWhenSelfhostPrecheckFails(t *testing.T) {
	uri := "file:///tmp/main.rn"
	prevAnalyze := analyzeSource
	prevCheck := selfhostCheckSource
	called := 0
	analyzeSource = func(path string, text string) (*compiler.Program, []compiler.Diagnostic) {
		called++
		return &compiler.Program{}, nil
	}
	selfhostCheckSource = func(source string, path string) SelfhostCompileResult {
		return SelfhostCompileResult{Ok: false, Errors: []string{"selfhost parse failed"}}
	}
	defer func() {
		analyzeSource = prevAnalyze
		selfhostCheckSource = prevCheck
	}()

	s := &server{docs: map[string]string{uri: "broken"}, cache: map[string]programCacheEntry{}}
	got := s.semanticTokens(uri).(map[string]any)
	data := got["data"].([]int)
	if len(data) != 0 {
		t.Fatalf("semanticTokens() data = %#v, want empty on selfhost precheck failure", data)
	}
	if called != 0 {
		t.Fatalf("host analyze calls = %d, want 0 when selfhost precheck fails", called)
	}
}

func TestCodeLensesSkipParserWhenSelfhostPrecheckFails(t *testing.T) {
	uri := "file:///tmp/main.rn"
	prevCheck := selfhostCheckSource
	selfhostCheckSource = func(source string, path string) SelfhostCompileResult {
		if source != `? "sample" { @assert.eq(1, 1) }` {
			t.Fatalf("check source = %q, want sample test", source)
		}
		if path != uri {
			t.Fatalf("check path = %q, want %q", path, uri)
		}
		return SelfhostCompileResult{Ok: false, Errors: []string{"selfhost parse failed"}}
	}
	defer func() { selfhostCheckSource = prevCheck }()

	s := &server{docs: map[string]string{uri: `? "sample" { @assert.eq(1, 1) }`}}
	lenses := s.codeLenses(uri).([]map[string]any)
	if len(lenses) != 0 {
		t.Fatalf("codeLenses() = %#v, want empty on selfhost precheck failure", lenses)
	}
}

func TestDependencyChainStopsAtCycles(t *testing.T) {
	got := dependencyChain("a", map[string][]string{
		"a": {"b"},
		"b": {"a"},
	})
	if got != "a -> b -> a (cycle)" {
		t.Fatalf("dependencyChain = %q, want cycle marker", got)
	}
}
