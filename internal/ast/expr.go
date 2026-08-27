package ast

import "github.com/oboard/rune-lang/internal/lexer"

type Expr interface {
	exprNode()
	Position() lexer.Position
}

type Identifier struct {
	Name         string
	SignalPrefix bool
	Pos          lexer.Position
}

func (*Identifier) exprNode() {}
func (e *Identifier) Position() lexer.Position {
	return e.Pos
}

type AtExpr struct {
	Name       string
	Path       string
	SourcePath string
	Pos        lexer.Position
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

type DoubleLiteral struct {
	Value float64
	Raw   string
	Pos   lexer.Position
}

func (*DoubleLiteral) exprNode() {}
func (e *DoubleLiteral) Position() lexer.Position {
	return e.Pos
}

type BigIntLiteral struct {
	Value string
	Pos   lexer.Position
}

func (*BigIntLiteral) exprNode() {}
func (e *BigIntLiteral) Position() lexer.Position {
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

type TemplateLiteral struct {
	Parts []TemplatePart
	Pos   lexer.Position
}

func (*TemplateLiteral) exprNode() {}
func (e *TemplateLiteral) Position() lexer.Position {
	return e.Pos
}

type TemplatePart struct {
	Text string
	Expr Expr
	Pos  lexer.Position
}

type CharLiteral struct {
	Value rune
	Pos   lexer.Position
}

func (*CharLiteral) exprNode() {}
func (e *CharLiteral) Position() lexer.Position {
	return e.Pos
}

type RegexLiteral struct {
	Pattern string
	Flags   string
	Raw     string
	Pos     lexer.Position
}

func (*RegexLiteral) exprNode() {}
func (e *RegexLiteral) Position() lexer.Position {
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

type NullLiteral struct {
	Pos lexer.Position
}

func (*NullLiteral) exprNode() {}
func (e *NullLiteral) Position() lexer.Position {
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

type PostfixExpr struct {
	Op   lexer.Kind
	Expr Expr
	Pos  lexer.Position
}

func (*PostfixExpr) exprNode() {}
func (e *PostfixExpr) Position() lexer.Position {
	return e.Pos
}

type ResultUnwrapExpr struct {
	Expr Expr
	Op   lexer.Kind
	Pos  lexer.Position
}

func (*ResultUnwrapExpr) exprNode() {}
func (e *ResultUnwrapExpr) Position() lexer.Position {
	return e.Pos
}

type CompileTimeExpr struct {
	Expr    Expr
	Pos     lexer.Position
	MarkPos lexer.Position
}

func (*CompileTimeExpr) exprNode() {}
func (e *CompileTimeExpr) Position() lexer.Position {
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

type TernaryExpr struct {
	Condition   Expr
	Consequence Expr
	Alternative Expr
	Pos         lexer.Position
}

func (*TernaryExpr) exprNode() {}
func (e *TernaryExpr) Position() lexer.Position {
	return e.Pos
}

type AssignExpr struct {
	Name   string
	Target Expr
	Value  Expr
	Pos    lexer.Position
}

func (*AssignExpr) exprNode() {}
func (e *AssignExpr) Position() lexer.Position {
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
	Params     []string
	ParamPos   []lexer.Position
	ParamTypes []Type
	ReturnType Type
	Implicit   bool
	Body       Expr
	Pos        lexer.Position
}

func (*LambdaExpr) exprNode() {}
func (e *LambdaExpr) Position() lexer.Position {
	return e.Pos
}

type SelectorExpr struct {
	Receiver Expr
	Name     string
	Static   bool
	Pos      lexer.Position
	NamePos  lexer.Position
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

type TupleLiteral struct {
	Elements []Expr
	Pos      lexer.Position
}

func (*TupleLiteral) exprNode() {}
func (e *TupleLiteral) Position() lexer.Position {
	return e.Pos
}

type SpreadExpr struct {
	Expr Expr
	Pos  lexer.Position
}

func (*SpreadExpr) exprNode() {}
func (e *SpreadExpr) Position() lexer.Position {
	return e.Pos
}

type ReactiveLiteral struct {
	Value Expr
	Pos   lexer.Position
}

func (*ReactiveLiteral) exprNode() {}
func (e *ReactiveLiteral) Position() lexer.Position {
	return e.Pos
}

type MapEntry struct {
	Key   Expr
	Value Expr
	Pos   lexer.Position
}

type MapLiteral struct {
	Entries []MapEntry
	Pos     lexer.Position
}

func (*MapLiteral) exprNode() {}
func (e *MapLiteral) Position() lexer.Position {
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

type AnonymousObjectLiteral struct {
	Fields []FieldValue
	Pos    lexer.Position
}

func (*AnonymousObjectLiteral) exprNode() {}
func (e *AnonymousObjectLiteral) Position() lexer.Position {
	return e.Pos
}

type FieldValue struct {
	Name         string
	Private      bool
	Spread       bool
	MissingComma bool
	Value        Expr
	Pos          lexer.Position
}

type XMLElement struct {
	Tag      string
	Attrs    []XMLAttr
	Children []XMLChild
	Pos      lexer.Position
}

func (*XMLElement) exprNode() {}
func (e *XMLElement) Position() lexer.Position {
	return e.Pos
}

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

type MatchExpr struct {
	Subject  Expr
	Branches []PatternBranch
	Pos      lexer.Position
}

func (*MatchExpr) exprNode() {}
func (e *MatchExpr) Position() lexer.Position {
	return e.Pos
}
