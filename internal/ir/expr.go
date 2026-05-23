package ir

import (
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/lexer"
)

type Expr interface {
	exprNode()
	Position() lexer.Position
	ResultType() checker.Type
}

type ExprBase struct {
	Pos  lexer.Position
	Type checker.Type
}

func (b ExprBase) Position() lexer.Position {
	return b.Pos
}

func (b ExprBase) ResultType() checker.Type {
	return b.Type
}

type Identifier struct {
	ExprBase
	Name string
}

func (*Identifier) exprNode() {}

type AtExpr struct {
	ExprBase
	Name string
}

func (*AtExpr) exprNode() {}

type ThisExpr struct {
	ExprBase
}

func (*ThisExpr) exprNode() {}

type IntegerLiteral struct {
	ExprBase
	Value int
}

func (*IntegerLiteral) exprNode() {}

type StringLiteral struct {
	ExprBase
	Value string
}

func (*StringLiteral) exprNode() {}

type BoolLiteral struct {
	ExprBase
	Value bool
}

func (*BoolLiteral) exprNode() {}

type UnaryExpr struct {
	ExprBase
	Op   lexer.Kind
	Expr Expr
}

func (*UnaryExpr) exprNode() {}

type PostfixExpr struct {
	ExprBase
	Op   lexer.Kind
	Expr Expr
}

func (*PostfixExpr) exprNode() {}

type BinaryExpr struct {
	ExprBase
	Left  Expr
	Op    lexer.Kind
	Right Expr
}

func (*BinaryExpr) exprNode() {}

type CallExpr struct {
	ExprBase
	Callee Expr
	Args   []Expr
}

func (*CallExpr) exprNode() {}

type LambdaExpr struct {
	ExprBase
	Params []string
	Body   Expr
}

func (*LambdaExpr) exprNode() {}

type SelectorExpr struct {
	ExprBase
	Receiver Expr
	Name     string
}

func (*SelectorExpr) exprNode() {}

type IndexExpr struct {
	ExprBase
	Receiver Expr
	Index    Expr
}

func (*IndexExpr) exprNode() {}

type ArrayLiteral struct {
	ExprBase
	Elements []Expr
}

func (*ArrayLiteral) exprNode() {}

type StructLiteral struct {
	ExprBase
	TypeName string
	Fields   []FieldValue
}

func (*StructLiteral) exprNode() {}

type AnonymousObjectLiteral struct {
	ExprBase
	Fields []FieldValue
}

func (*AnonymousObjectLiteral) exprNode() {}

type FieldValue struct {
	Name  string
	Value Expr
	Pos   lexer.Position
}

type XMLElement struct {
	ExprBase
	Tag      string
	Attrs    []XMLAttr
	Children []XMLChild
}

func (*XMLElement) exprNode() {}

type XMLAttr struct {
	Name  string
	Event bool
	Value Expr
	Pos   lexer.Position
}

type XMLChild struct {
	Text string
	Expr Expr
	Pos  lexer.Position
}

type BlockExpr struct {
	ExprBase
	Statements []Stmt
}

func (*BlockExpr) exprNode() {}

type PatternBlock struct {
	ExprBase
	Branches []PatternBranch
}

func (*PatternBlock) exprNode() {}

type MatchExpr struct {
	ExprBase
	Subject  Expr
	Branches []PatternBranch
}

func (*MatchExpr) exprNode() {}

type WatchExpr struct {
	ExprBase
	Target  Expr
	Handler Expr
}

func (*WatchExpr) exprNode() {}
