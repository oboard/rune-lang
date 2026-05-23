package format

import (
	"fmt"

	"github.com/oboard/rune-lang/internal/ast"
)

func (f *formatter) stmt(stmt ast.Stmt) string {
	switch s := stmt.(type) {
	case *ast.LetStmt:
		op := ":="
		if s.Mutable {
			op = "~="
		}
		if s.Signal {
			op = "$="
		}
		return fmt.Sprintf("%s %s %s", s.Name, op, f.expr(s.Value))
	case *ast.AssignStmt:
		return fmt.Sprintf("%s = %s", s.Name, f.expr(s.Value))
	case *ast.ExprStmt:
		return f.expr(s.Expr)
	default:
		return ""
	}
}
