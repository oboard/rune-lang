package lsp

import (
	"strings"

	"github.com/oboard/rune-lang/internal/compiler"
	runefmt "github.com/oboard/rune-lang/internal/format"
)

func (s *server) formatting(uri string) any {
	text, ok := s.docs[uri]
	if !ok {
		return nil
	}
	prog, diags := compiler.AnalyzeSource(uri, text)
	if prog == nil || len(diags) > 0 {
		return nil
	}
	formatted := runefmt.File(prog.File)
	if formatted == text {
		return []map[string]any{}
	}
	return []map[string]any{
		{
			"range":   fullDocumentRange(text),
			"newText": formatted,
		},
	}
}

func fullDocumentRange(text string) map[string]any {
	lines := strings.Split(text, "\n")
	lastLine := len(lines) - 1
	lastChar := 0
	if lastLine >= 0 {
		lastChar = len([]rune(lines[lastLine]))
	}
	return map[string]any{
		"start": position{Line: 0, Character: 0},
		"end":   position{Line: max(lastLine, 0), Character: lastChar},
	}
}
