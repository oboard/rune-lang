package ast

import "fmt"

func WalkExpr(expr Expr, visit func(Expr)) {
	if expr == nil {
		return
	}
	visit(expr)
	switch e := expr.(type) {
	case *UnaryExpr:
		WalkExpr(e.Expr, visit)
	case *BinaryExpr:
		WalkExpr(e.Left, visit)
		WalkExpr(e.Right, visit)
	case *CallExpr:
		WalkExpr(e.Callee, visit)
		for _, arg := range e.Args {
			WalkExpr(arg, visit)
		}
	case *LambdaExpr:
		WalkExpr(e.Body, visit)
	case *SelectorExpr:
		WalkExpr(e.Receiver, visit)
	case *IndexExpr:
		WalkExpr(e.Receiver, visit)
		WalkExpr(e.Index, visit)
	case *ArrayLiteral:
		for _, elem := range e.Elements {
			WalkExpr(elem, visit)
		}
	case *StructLiteral:
		for _, field := range e.Fields {
			WalkExpr(field.Value, visit)
		}
	case *AnonymousObjectLiteral:
		for _, field := range e.Fields {
			WalkExpr(field.Value, visit)
		}
	case *BlockExpr:
		for _, stmt := range e.Statements {
			switch s := stmt.(type) {
			case *LetStmt:
				WalkExpr(s.Value, visit)
			case *AssignStmt:
				WalkExpr(s.Value, visit)
			case *ExprStmt:
				WalkExpr(s.Expr, visit)
			}
		}
	case *PatternBlock:
		for _, branch := range e.Branches {
			WalkPattern(branch.Pattern, func(p Pattern) {
				switch p := p.(type) {
				case *LiteralPattern:
					WalkExpr(p.Value, visit)
				case *ComparePattern:
					WalkExpr(p.Value, visit)
				}
			})
			WalkExpr(branch.Expr, visit)
		}
	case *MatchExpr:
		WalkExpr(e.Subject, visit)
		for _, branch := range e.Branches {
			WalkPattern(branch.Pattern, func(p Pattern) {
				switch p := p.(type) {
				case *LiteralPattern:
					WalkExpr(p.Value, visit)
				case *ComparePattern:
					WalkExpr(p.Value, visit)
				}
			})
			WalkExpr(branch.Expr, visit)
		}
	}
}

func WalkPattern(pattern Pattern, visit func(Pattern)) {
	if pattern == nil {
		return
	}
	visit(pattern)
	if tuple, ok := pattern.(*TuplePattern); ok {
		for _, elem := range tuple.Elements {
			WalkPattern(elem, visit)
		}
	}
}

func ExprName(expr Expr) string {
	switch e := expr.(type) {
	case *Identifier:
		return e.Name
	case *AtExpr:
		return "@" + e.Name
	case *ThisExpr:
		return "this"
	case *SelectorExpr:
		return fmt.Sprintf("%s.%s", ExprName(e.Receiver), e.Name)
	case *IndexExpr:
		return fmt.Sprintf("%s[]", ExprName(e.Receiver))
	default:
		return ""
	}
}
