package format

import (
	"fmt"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
)

func (f *formatter) stmt(stmt ast.Stmt) string {
	switch s := stmt.(type) {
	case *ast.LetStmt:
		op := ":="
		if s.Mutable {
			op = ":=:"
		}
		if s.Signal {
			out := fmt.Sprintf("$%s := %s", s.Name, f.expr(s.Value))
			if !s.Type.IsZero() {
				out += " : " + formatType(s.Type)
			}
			return out
		}
		out := fmt.Sprintf("%s %s %s", s.Name, op, f.expr(s.Value))
		if !s.Type.IsZero() {
			out += " : " + formatType(s.Type)
		}
		return out
	case *ast.ObjectDestructureStmt:
		op := ":="
		if s.Mutable {
			op = ":=:"
		}
		fields := make([]string, 0, len(s.Fields))
		for _, field := range s.Fields {
			if field.Field == field.Name {
				fields = append(fields, field.Field)
				continue
			}
			fields = append(fields, fmt.Sprintf("%s: %s", field.Field, field.Name))
		}
		return fmt.Sprintf("{ %s } %s %s", strings.Join(fields, ", "), op, f.expr(s.Value))
	case *ast.AssignStmt:
		name := s.Name
		if s.SignalPrefix {
			name = "$" + name
		}
		return fmt.Sprintf("%s = %s", name, f.expr(s.Value))
	case *ast.ExprStmt:
		return f.expr(s.Expr)
	default:
		return ""
	}
}
