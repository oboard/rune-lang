package ast

import "github.com/oboard/rune-lang/internal/lexer"

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

type LambdaExpr struct {
	Params []string
	Body   Expr
	Pos    lexer.Position
}

func (*LambdaExpr) exprNode() {}
func (e *LambdaExpr) Position() lexer.Position {
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

type IndexExpr struct {
	Receiver Expr
	Index    Expr
	Pos      lexer.Position
}

func (*IndexExpr) exprNode() {}
func (e *IndexExpr) Position() lexer.Position {
	return e.Pos
}

type ArrayLiteral struct {
	Elements []Expr
	Pos      lexer.Position
}

func (*ArrayLiteral) exprNode() {}
func (e *ArrayLiteral) Position() lexer.Position {
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
