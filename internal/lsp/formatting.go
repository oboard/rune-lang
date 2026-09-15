package lsp

import (
	"strings"

	"github.com/oboard/rune-lang/internal/selfhost"
)

func (s *server) formatting(uri string, options *formattingOptions) any {
	text, ok := s.docs[uri]
	if !ok {
		return nil
	}
	formatted := selfhost.FormatSource(text)
	formatted = applyFormattingIndent(formatted, options)
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

func applyFormattingIndent(formatted string, options *formattingOptions) string {
	if options == nil || (options.InsertSpaces && options.TabSize == 2) {
		return formatted
	}
	unit := "  "
	if !options.InsertSpaces {
		unit = "\t"
	} else if options.TabSize > 0 {
		unit = strings.Repeat(" ", options.TabSize)
	}
	return strings.ReplaceAll(formatted, "  ", unit)
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
