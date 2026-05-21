package ir

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
	case *BlockExpr:
		for _, stmt := range e.Statements {
			WalkStmt(stmt, visit)
		}
	case *PatternBlock:
		for _, branch := range e.Branches {
			WalkPattern(branch.Pattern, visit)
			WalkExpr(branch.Expr, visit)
		}
	}
}

func WalkStmt(stmt Stmt, visit func(Expr)) {
	switch s := stmt.(type) {
	case *LetStmt:
		WalkExpr(s.Value, visit)
	case *AssignStmt:
		WalkExpr(s.Value, visit)
	case *ExprStmt:
		WalkExpr(s.Expr, visit)
	}
}

func WalkPattern(pattern Pattern, visit func(Expr)) {
	switch p := pattern.(type) {
	case *LiteralPattern:
		WalkExpr(p.Value, visit)
	case *ComparePattern:
		WalkExpr(p.Value, visit)
	case *TuplePattern:
		for _, elem := range p.Elements {
			WalkPattern(elem, visit)
		}
	}
}
