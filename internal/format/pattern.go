package format

import (
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
)

func (f *formatter) pattern(pattern ast.Pattern) string {
	switch p := pattern.(type) {
	case *ast.WildcardPattern:
		return "_"
	case *ast.LiteralPattern:
		return f.expr(p.Value)
	case *ast.ComparePattern:
		return p.Op.String() + f.expr(p.Value)
	case *ast.TuplePattern:
		parts := make([]string, 0, len(p.Elements))
		for _, elem := range p.Elements {
			parts = append(parts, f.pattern(elem))
		}
		return "(" + strings.Join(parts, ", ") + ")"
	default:
		return "_"
	}
}
