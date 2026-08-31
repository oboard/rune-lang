package lsp

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/parser"
)

func (s *server) codeLenses(uri string) any {
	text := s.docs[uri]
	if selfhostCheckSource != nil {
		checked := selfhostCheckSource(text, uri)
		if !checked.Ok {
			return []map[string]any{}
		}
	}
	file, _ := parser.Parse(text)
	if file == nil {
		return []map[string]any{}
	}
	lenses := make([]map[string]any, 0, len(file.Tests))
	walkFileAnnotations(file, func(annotation *ast.Annotation) {
		line := max(annotation.Pos.Line-1, 0)
		character := max(annotation.Pos.Column-1, 0)
		lenses = append(lenses, map[string]any{
			"range": symbolRange(annotation.Pos, 1),
			"command": map[string]any{
				"title":   "$(unfold) Show Macro Expansion",
				"command": "rune.showMacroExpansion",
				"arguments": []any{map[string]any{
					"uri":       uri,
					"line":      line,
					"character": character,
				}},
			},
		})
	})
	for _, test := range file.Tests {
		line := max(test.Pos.Line-1, 0)
		character := max(test.Pos.Column-1, 0)
		lenses = append(lenses, map[string]any{
			"range": symbolRange(test.Pos, 1),
			"command": map[string]any{
				"title":   "$(play) Run Test",
				"command": "rune.runTest",
				"arguments": []any{map[string]any{
					"uri":       uri,
					"name":      test.Name,
					"line":      line,
					"character": character,
				}},
			},
		})
	}
	return lenses
}
