package ast

import (
	"fmt"
	"strconv"
)

func WalkExpr(expr Expr, visit func(Expr)) {
	if expr == nil {
		return
	}
	visit(expr)
	switch e := expr.(type) {
	case *Identifier, *AtExpr, *ThisExpr, *IntegerLiteral, *DoubleLiteral, *BigIntLiteral,
		*StringLiteral, *CharLiteral, *RegexLiteral, *BoolLiteral, *NullLiteral:
	case *TemplateLiteral:
		for _, part := range e.Parts {
			WalkExpr(part.Expr, visit)
		}
	case *UnaryExpr:
		WalkExpr(e.Expr, visit)
	case *PostfixExpr:
		WalkExpr(e.Expr, visit)
	case *ResultUnwrapExpr:
		WalkExpr(e.Expr, visit)
	case *CompileTimeExpr:
		WalkExpr(e.Expr, visit)
	case *BinaryExpr:
		WalkExpr(e.Left, visit)
		WalkExpr(e.Right, visit)
	case *TernaryExpr:
		WalkExpr(e.Condition, visit)
		WalkExpr(e.Consequence, visit)
		WalkExpr(e.Alternative, visit)
	case *AssignExpr:
		WalkExpr(e.Target, visit)
		WalkExpr(e.Value, visit)
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
	case *TupleLiteral:
		for _, elem := range e.Elements {
			WalkExpr(elem, visit)
		}
	case *SpreadExpr:
		WalkExpr(e.Expr, visit)
	case *ReactiveLiteral:
		WalkExpr(e.Value, visit)
	case *MapLiteral:
		for _, entry := range e.Entries {
			WalkExpr(entry.Key, visit)
			WalkExpr(entry.Value, visit)
		}
	case *StructLiteral:
		for _, field := range e.Fields {
			WalkExpr(field.Value, visit)
		}
	case *AnonymousObjectLiteral:
		for _, field := range e.Fields {
			WalkExpr(field.Value, visit)
		}
	case *XMLElement:
		for _, attr := range e.Attrs {
			WalkExpr(attr.Value, visit)
		}
		for _, child := range e.Children {
			WalkExpr(child.Expr, visit)
		}
	case *BlockExpr:
		for _, stmt := range e.Statements {
			switch s := stmt.(type) {
			case *LetStmt:
				WalkExpr(s.Value, visit)
			case *ObjectDestructureStmt:
				WalkExpr(s.Value, visit)
			case *AssignStmt:
				WalkExpr(s.Value, visit)
			case *ExprStmt:
				WalkExpr(s.Expr, visit)
			}
		}
	case *PatternBlock:
		for _, branch := range e.Branches {
			WalkPatternExprs(branch.Pattern, visit)
			WalkExpr(branch.Expr, visit)
		}
	case *MatchExpr:
		WalkExpr(e.Subject, visit)
		for _, branch := range e.Branches {
			WalkPatternExprs(branch.Pattern, visit)
			WalkExpr(branch.Expr, visit)
		}
	case *WatchExpr:
		WalkExpr(e.Target, visit)
		WalkExpr(e.Handler, visit)
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
	if array, ok := pattern.(*ArrayPattern); ok {
		for _, elem := range array.Elements {
			WalkPattern(elem, visit)
		}
	}
	if bit, ok := pattern.(*BitPattern); ok {
		WalkPattern(bit.Value, visit)
	}
	if as, ok := pattern.(*AsPattern); ok {
		WalkPattern(as.Pattern, visit)
	}
	if constructor, ok := pattern.(*ConstructorPattern); ok {
		for _, arg := range constructor.Args {
			WalkPattern(arg, visit)
		}
	}
	if or, ok := pattern.(*OrPattern); ok {
		for _, alternative := range or.Alternatives {
			WalkPattern(alternative, visit)
		}
	}
	if mapPattern, ok := pattern.(*MapPattern); ok {
		for _, entry := range mapPattern.Entries {
			WalkPattern(entry.Pattern, visit)
		}
	}
	if objectPattern, ok := pattern.(*ObjectPattern); ok {
		for _, field := range objectPattern.Fields {
			WalkPattern(field.Pattern, visit)
		}
	}
}

func WalkPatternExprs(pattern Pattern, visit func(Expr)) {
	switch p := pattern.(type) {
	case *LiteralPattern:
		WalkExpr(p.Value, visit)
	case *ComparePattern:
		WalkExpr(p.Value, visit)
	case *RangePattern:
		if p.Start != nil {
			WalkExpr(p.Start, visit)
		}
		if p.End != nil {
			WalkExpr(p.End, visit)
		}
	case *OrPattern:
		for _, alternative := range p.Alternatives {
			WalkPatternExprs(alternative, visit)
		}
	case *TuplePattern:
		for _, elem := range p.Elements {
			WalkPatternExprs(elem, visit)
		}
	case *ArrayPattern:
		for _, elem := range p.Elements {
			WalkPatternExprs(elem, visit)
		}
	case *SequenceSpreadPattern:
		WalkExpr(p.Value, visit)
	case *BitPattern:
		WalkPatternExprs(p.Value, visit)
	case *AsPattern:
		WalkPatternExprs(p.Pattern, visit)
	case *ConstructorPattern:
		for _, arg := range p.Args {
			WalkPatternExprs(arg, visit)
		}
	case *MapPattern:
		for _, entry := range p.Entries {
			WalkExpr(entry.Key, visit)
			WalkPatternExprs(entry.Pattern, visit)
		}
	case *ObjectPattern:
		for _, field := range p.Fields {
			WalkPatternExprs(field.Pattern, visit)
		}
	}
}

func ExprName(expr Expr) string {
	switch e := expr.(type) {
	case *Identifier:
		return e.Name
	case *AtExpr:
		if e.Path != "" {
			return "@" + strconv.Quote(e.Path)
		}
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
