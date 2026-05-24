package lsp

import (
	"strings"

	runefmt "github.com/oboard/rune-lang/internal/format"
	"github.com/oboard/rune-lang/internal/parser"
)

func (s *server) formatting(uri string, options *formattingOptions) any {
	text, ok := s.docs[uri]
	if !ok {
		return nil
	}
	file, errs := parser.Parse(text)
	if len(errs) > 0 {
		return nil
	}
	formatted := runefmt.SourceWithOptions(file, text, formatOptions(options))
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

func formatOptions(options *formattingOptions) runefmt.Options {
	if options == nil {
		return runefmt.Options{}
	}
	if !options.InsertSpaces {
		return runefmt.Options{Indent: "\t"}
	}
	tabSize := options.TabSize
	if tabSize <= 0 {
		return runefmt.Options{}
	}
	return runefmt.Options{Indent: strings.Repeat(" ", tabSize)}
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
