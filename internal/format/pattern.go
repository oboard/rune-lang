package format

import (
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
)

func (f *formatter) pattern(pattern ast.Pattern) string {
	switch p := pattern.(type) {
	case *ast.WildcardPattern:
		return "_"
	case *ast.BindingPattern:
		return p.Name
	case *ast.LiteralPattern:
		return f.expr(p.Value)
	case *ast.ComparePattern:
		return p.Op.String() + f.expr(p.Value)
	case *ast.RangePattern:
		return f.expr(p.Start) + "..=" + f.expr(p.End)
	case *ast.TuplePattern:
		parts := make([]string, 0, len(p.Elements))
		for _, elem := range p.Elements {
			parts = append(parts, f.pattern(elem))
		}
		return "(" + strings.Join(parts, ", ") + ")"
	case *ast.ConstructorPattern:
		binding := p.Binding
		if binding == "" {
			binding = "_"
		}
		return p.Name + "(" + binding + ")"
	case *ast.MapPattern:
		parts := make([]string, 0, len(p.Entries)+1)
		for _, entry := range p.Entries {
			key := f.expr(entry.Key)
			if entry.Optional {
				key += "?"
			}
			parts = append(parts, key+": "+f.pattern(entry.Pattern))
		}
		if p.Rest {
			parts = append(parts, "..")
		}
		return "{ " + strings.Join(parts, ", ") + " }"
	case *ast.ObjectPattern:
		parts := make([]string, 0, len(p.Fields)+1)
		for _, field := range p.Fields {
			name := field.Name
			if field.Optional {
				name += "?"
			}
			if binding, ok := field.Pattern.(*ast.BindingPattern); ok && binding.Name == field.Name && !field.Optional {
				parts = append(parts, field.Name)
				continue
			}
			parts = append(parts, name+": "+f.pattern(field.Pattern))
		}
		if p.Rest {
			parts = append(parts, "..")
		}
		return "{ " + strings.Join(parts, ", ") + " }"
	default:
		return "_"
	}
}
