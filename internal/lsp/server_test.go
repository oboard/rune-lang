package lsp

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
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

func TestDependencyChainStopsAtCycles(t *testing.T) {
	got := dependencyChain("a", map[string][]string{
		"a": []string{"b"},
		"b": []string{"a"},
	})
	if got != "a -> b -> a (cycle)" {
		t.Fatalf("dependencyChain = %q, want cycle marker", got)
	}
}
