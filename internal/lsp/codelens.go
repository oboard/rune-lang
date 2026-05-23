package lsp

import "github.com/oboard/rune-lang/internal/parser"

func (s *server) codeLenses(uri string) any {
	file, _ := parser.Parse(s.docs[uri])
	if file == nil {
		return []map[string]any{}
	}
	lenses := make([]map[string]any, 0, len(file.Tests))
	for _, test := range file.Tests {
		line := max(test.Pos.Line-1, 0)
		character := max(test.Pos.Column-1, 0)
		lenses = append(lenses, map[string]any{
			"range": symbolRange(test.Pos, 1),
			"command": map[string]any{
				"title":   "▶ Run Test",
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
