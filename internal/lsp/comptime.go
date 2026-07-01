package lsp

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/parser"
)

func (s *server) compileTimeRanges(uri string) any {
	text := s.docs[uri]
	file, errs := parser.Parse(text)
	if file == nil || len(errs) > 0 {
		return []any{}
	}
	ranges := []any{}
	walkFileExprs(file, func(expr ast.Expr) {
		comptime, ok := expr.(*ast.CompileTimeExpr)
		if !ok {
			return
		}
		ranges = append(ranges, map[string]any{
			"range": map[string]any{
				"start": positionFromLexer(comptime.Pos),
				"end":   positionFromLexer(comptime.MarkPos),
			},
		})
	})
	return ranges
}
