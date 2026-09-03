package lsp

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/oboard/rune-lang/internal/compiler"
	"github.com/oboard/rune-lang/internal/parser"
)

type analyzeSourceFunc func(string, string) (*compiler.Program, []compiler.Diagnostic)
type selfhostCheckSourceFunc func(string, string) SelfhostCompileResult

var analyzeSource = compiler.AnalyzeSource
var selfhostCheckSource selfhostCheckSourceFunc

type SelfhostCompileResult struct {
	Ok     bool
	Output string
	Errors []string
}

func Serve(in io.Reader, out io.Writer) error {
	s := &server{
		reader: bufio.NewReader(in),
		out:    out,
		docs:   map[string]string{},
		cache:  map[string]programCacheEntry{},
	}
	for {
		msg, err := s.readMessage()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if err := s.handle(msg); err != nil {
			return err
		}
	}
}

type server struct {
	reader *bufio.Reader
	out    io.Writer
	docs   map[string]string
	cache  map[string]programCacheEntry
}

func RegisterSelfhostAnalyzer(check selfhostCheckSourceFunc, analyze analyzeSourceFunc) {
	selfhostCheckSource = check
	if analyze != nil {
		analyzeSource = analyze
	}
}

type programCacheEntry struct {
	text     string
	prog     *compiler.Program
	diags    []compiler.Diagnostic
	warnings bool
}

func (s *server) setDocument(uri string, text string) {
	if s.docs == nil {
		s.docs = map[string]string{}
	}
	s.docs[uri] = text
	s.clearProgramCache()
}

func (s *server) closeDocument(uri string) {
	delete(s.docs, uri)
	s.clearProgramCache()
}

func (s *server) clearDocuments() {
	s.docs = map[string]string{}
	s.clearProgramCache()
}

func (s *server) clearProgramCache() {
	if s.cache == nil {
		return
	}
	clear(s.cache)
}

func (s *server) analyze(uri string) (*compiler.Program, []compiler.Diagnostic) {
	return s.analyzeWithWarnings(uri, false)
}

func (s *server) analyzeWithWarnings(uri string, includeWarnings bool) (*compiler.Program, []compiler.Diagnostic) {
	text := s.docs[uri]
	traceLSPAnalyze(uri, text)
	if s.cache != nil {
		if entry, ok := s.cache[uri]; ok && entry.text == text && (!includeWarnings || entry.warnings) {
			return entry.prog, entry.diags
		}
	} else {
		s.cache = map[string]programCacheEntry{}
	}
	// Core/selfhost modules and tests that import selfhost modules rely on the
	// host import graph to bring bootstrap declarations into scope. The selfhost
	// single-source precheck cannot see those imports from an open editor buffer,
	// so skip it and let the host analyzer report authoritative diagnostics.
	if selfhostCheckSource != nil && !isBootstrapSourceURI(uri) && !sourceImportsBootstrap(text) {
		checked := selfhostCheckSource(text, uri)
		if !checked.Ok {
			diags := selfhostDiagnostics(uri, checked.Errors)
			s.cache[uri] = programCacheEntry{text: text, diags: diags, warnings: includeWarnings}
			return nil, diags
		}
	}
	prog, diags := analyzeSource(uri, text)
	if includeWarnings && len(diags) == 0 {
		_, warningDiags := compiler.AnalyzeSourceWithWarnings(uri, text)
		for _, diag := range warningDiags {
			if diag.Severity == "warning" {
				diags = append(diags, diag)
			}
		}
	}
	s.cache[uri] = programCacheEntry{text: text, prog: prog, diags: diags, warnings: includeWarnings}
	return prog, diags
}

func isBootstrapSourceURI(uri string) bool {
	path := uri
	if strings.HasPrefix(path, "file://") {
		if parsed, err := url.Parse(path); err == nil {
			path = parsed.Path
		}
	}
	path = filepath.ToSlash(filepath.Clean(path))
	return strings.Contains(path, "/core/") || strings.Contains(path, "/selfhost/")
}

func sourceImportsBootstrap(text string) bool {
	file, errs := parser.Parse(text)
	if len(errs) > 0 || file == nil {
		return false
	}
	for _, imp := range file.Imports {
		path := filepath.ToSlash(filepath.Clean(imp.Path))
		if strings.Contains(path, "selfhost/") || strings.Contains(path, "core/") {
			return true
		}
	}
	return false
}

func selfhostDiagnostics(uri string, messages []string) []compiler.Diagnostic {
	diags := make([]compiler.Diagnostic, 0, len(messages))
	for _, message := range messages {
		diags = append(diags, compiler.Diagnostic{Message: message, Path: uri})
	}
	return diags
}

func traceLSPAnalyze(uri string, text string) {
	path := os.Getenv("RUNE_LSP_TRACE_PATH")
	if path == "" {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "URI=%s\nTEXT=%q\n---\n", uri, text)
}
