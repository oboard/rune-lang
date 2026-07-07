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
		start := compileTimeRangeStart(text, comptime)
		ranges = append(ranges, map[string]any{
			"range": map[string]any{
				"start": positionAtOffset(text, start),
				"end":   positionFromLexer(comptime.MarkPos),
			},
		})
	})
	return ranges
}

func compileTimeRangeStart(text string, comptime *ast.CompileTimeExpr) int {
	start := max(comptime.Pos.Offset, 0)
	for {
		i := start - 1
		for i >= 0 {
			switch text[i] {
			case ' ', '\t', '\r', '\n':
				i--
				continue
			}
			break
		}
		if i < 0 || text[i] != '(' {
			return start
		}
		start = i
	}
}
