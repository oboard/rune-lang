package ast

import (
	"fmt"
	"strings"

	"github.com/oboard/rune-lang/internal/lexer"
)

type File struct {
	Types     []*StructType
	Functions []*Function
}

type StructType struct {
	Name    string
	Fields  []Field
	Methods []*Function
	Pos     lexer.Position
	NamePos lexer.Position
}

type Field struct {
	Name string
	Type string
	Pos  lexer.Position
}

type Function struct {
	Name         string
	ReceiverType string
	Params       []Param
	Body         Expr
	Pos          lexer.Position
	NamePos      lexer.Position
}

func (f *Function) Signature() string {
	var b strings.Builder
	b.WriteString(f.Name)
	b.WriteByte('(')
	for i, param := range f.Params {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(param.Name)
		b.WriteString(": ")
		b.WriteString(param.Type)
	}
	b.WriteByte(')')
	return b.String()
}

type Param struct {
	Name string
	Type string
	Pos  lexer.Position
}

type Expr interface {
	exprNode()
	Position() lexer.Position
}

type Identifier struct {
	Name string
	Pos  lexer.Position
}

func (*Identifier) exprNode() {}
func (e *Identifier) Position() lexer.Position {
	return e.Pos
}

type AtExpr struct {
	Name string
	Pos  lexer.Position
}

func (*AtExpr) exprNode() {}
func (e *AtExpr) Position() lexer.Position {
	return e.Pos
}

type ThisExpr struct {
	Pos lexer.Position
}

func (*ThisExpr) exprNode() {}
func (e *ThisExpr) Position() lexer.Position {
	return e.Pos
}

type IntegerLiteral struct {
	Value int
	Pos   lexer.Position
}

func (*IntegerLiteral) exprNode() {}
func (e *IntegerLiteral) Position() lexer.Position {
	return e.Pos
}

type StringLiteral struct {
	Value string
	Pos   lexer.Position
}

func (*StringLiteral) exprNode() {}
func (e *StringLiteral) Position() lexer.Position {
	return e.Pos
}

type BoolLiteral struct {
	Value bool
	Pos   lexer.Position
}

func (*BoolLiteral) exprNode() {}
func (e *BoolLiteral) Position() lexer.Position {
	return e.Pos
}

type UnaryExpr struct {
	Op   lexer.Kind
	Expr Expr
	Pos  lexer.Position
}

func (*UnaryExpr) exprNode() {}
func (e *UnaryExpr) Position() lexer.Position {
	return e.Pos
}

type BinaryExpr struct {
	Left  Expr
	Op    lexer.Kind
	Right Expr
	Pos   lexer.Position
}

func (*BinaryExpr) exprNode() {}
func (e *BinaryExpr) Position() lexer.Position {
	return e.Pos
}

type CallExpr struct {
	Callee Expr
	Args   []Expr
	Pos    lexer.Position
}

func (*CallExpr) exprNode() {}
func (e *CallExpr) Position() lexer.Position {
	return e.Pos
}

type SelectorExpr struct {
	Receiver Expr
	Name     string
	Pos      lexer.Position
}

func (*SelectorExpr) exprNode() {}
func (e *SelectorExpr) Position() lexer.Position {
	return e.Pos
}

type StructLiteral struct {
	TypeName string
	Fields   []FieldValue
	Pos      lexer.Position
}

func (*StructLiteral) exprNode() {}
func (e *StructLiteral) Position() lexer.Position {
	return e.Pos
}

type FieldValue struct {
	Name  string
	Value Expr
	Pos   lexer.Position
}

type BlockExpr struct {
	Statements []Stmt
	Pos        lexer.Position
}

func (*BlockExpr) exprNode() {}
func (e *BlockExpr) Position() lexer.Position {
	return e.Pos
}

type PatternBlock struct {
	Branches []PatternBranch
	Pos      lexer.Position
}

func (*PatternBlock) exprNode() {}
func (e *PatternBlock) Position() lexer.Position {
	return e.Pos
}

type Stmt interface {
	stmtNode()
	Position() lexer.Position
}

type LetStmt struct {
	Name    string
	Mutable bool
	Value   Expr
	Pos     lexer.Position
}

func (*LetStmt) stmtNode() {}
func (s *LetStmt) Position() lexer.Position {
	return s.Pos
}

type AssignStmt struct {
	Name  string
	Value Expr
	Pos   lexer.Position
}

func (*AssignStmt) stmtNode() {}
func (s *AssignStmt) Position() lexer.Position {
	return s.Pos
}

type ExprStmt struct {
	Expr Expr
	Pos  lexer.Position
}

func (*ExprStmt) stmtNode() {}
func (s *ExprStmt) Position() lexer.Position {
	return s.Pos
}

type PatternBranch struct {
	Pattern Pattern
	Expr    Expr
	Pos     lexer.Position
}

type Pattern interface {
	patternNode()
	Position() lexer.Position
}

type WildcardPattern struct {
	Pos lexer.Position
}

func (*WildcardPattern) patternNode() {}
func (p *WildcardPattern) Position() lexer.Position {
	return p.Pos
}

type LiteralPattern struct {
	Value Expr
	Pos   lexer.Position
}

func (*LiteralPattern) patternNode() {}
func (p *LiteralPattern) Position() lexer.Position {
	return p.Pos
}

type ComparePattern struct {
	Op    lexer.Kind
	Value Expr
	Pos   lexer.Position
}

func (*ComparePattern) patternNode() {}
func (p *ComparePattern) Position() lexer.Position {
	return p.Pos
}

type TuplePattern struct {
	Elements []Pattern
	Pos      lexer.Position
}

func (*TuplePattern) patternNode() {}
func (p *TuplePattern) Position() lexer.Position {
	return p.Pos
}

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
	case *SelectorExpr:
		WalkExpr(e.Receiver, visit)
	case *StructLiteral:
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
	default:
		return ""
	}
}
