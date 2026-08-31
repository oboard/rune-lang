package lsp

import (
	"io"

	"github.com/oboard/rune-lang/internal/compiler"
)

type Session struct {
	server *server
}

func NewSession() *Session {
	return &Session{
		server: &server{
			out:   io.Discard,
			docs:  map[string]string{},
			cache: map[string]programCacheEntry{},
		},
	}
}

func (s *Session) SetDocument(uri string, text string) {
	s.server.setDocument(uri, text)
}

func (s *Session) Diagnostics(uri string) []map[string]any {
	_, diags := s.server.analyze(uri)
	return diagnosticsToLSP(diags)
}

func diagnosticsToLSP(diags []compiler.Diagnostic) []map[string]any {
	items := make([]map[string]any, 0, len(diags))
	for _, diag := range diags {
		severity := 1 // Error
		if diag.Severity == "warning" {
			severity = 2 // Warning
		}
		items = append(items, map[string]any{
			"range":    lspRange(diag.Pos),
			"severity": severity,
			"source":   "rune",
			"message":  diag.Message,
		})
	}
	return items
}

func (s *Session) Hover(uri string, line int, character int) any {
	return s.server.hover(uri, position{Line: line, Character: character})
}

func (s *Session) Completion(uri string, line int, character int) any {
	return s.server.completion(uri, position{Line: line, Character: character})
}

func (s *Session) Definition(uri string, line int, character int) any {
	return s.server.definition(uri, position{Line: line, Character: character})
}

func (s *Session) References(uri string, line int, character int, includeDeclaration bool) any {
	return s.server.references(uri, position{Line: line, Character: character}, includeDeclaration)
}

func (s *Session) CodeLenses(uri string) any {
	return s.server.codeLenses(uri)
}

func (s *Session) DocumentSymbols(uri string) any {
	return s.server.documentSymbols(uri)
}

func (s *Session) Formatting(uri string, tabSize int, insertSpaces bool) any {
	return s.server.formatting(uri, &formattingOptions{TabSize: tabSize, InsertSpaces: insertSpaces})
}

func (s *Session) InlayHints(uri string) any {
	return s.server.inlayHints(uri)
}

func (s *Session) SemanticTokens(uri string) any {
	return s.server.semanticTokens(uri)
}

func (s *Session) Rename(uri string, line int, character int, newName string) any {
	return s.server.rename(uri, position{Line: line, Character: character}, newName)
}
