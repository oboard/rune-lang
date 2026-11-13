package ast

import "github.com/oboard/rune-lang/internal/lexer"

type Stmt interface {
	stmtNode()
	Position() lexer.Position
}

type LetStmt struct {
	Name    string
	Mutable bool
	Signal  bool
	Value   Expr
	Type    Type
	Pos     lexer.Position
}

func (*LetStmt) stmtNode() {}
func (s *LetStmt) Position() lexer.Position {
	return s.Pos
}

type ObjectBindingField struct {
	Field    string
	Name     string
	FieldPos lexer.Position
	NamePos  lexer.Position
}

type ObjectDestructureStmt struct {
	Fields  []ObjectBindingField
	Mutable bool
	Signal  bool
	Value   Expr
	Pos     lexer.Position
}

func (*ObjectDestructureStmt) stmtNode() {}
func (s *ObjectDestructureStmt) Position() lexer.Position {
	return s.Pos
}

type AssignStmt struct {
	Name         string
	SignalPrefix bool
	Value        Expr
	Pos          lexer.Position
}

func (*AssignStmt) stmtNode() {}
func (s *AssignStmt) Position() lexer.Position {
	return s.Pos
}

// ReturnStmt returns from the enclosing function. Value is nil for a bare
// `return`.
type ReturnStmt struct {
	Value Expr
	Pos   lexer.Position
}

func (*ReturnStmt) stmtNode() {}
func (s *ReturnStmt) Position() lexer.Position {
	return s.Pos
}

type BreakStmt struct {
	Pos lexer.Position
}

func (*BreakStmt) stmtNode() {}
func (s *BreakStmt) Position() lexer.Position {
	return s.Pos
}

type ContinueStmt struct {
	Pos lexer.Position
}

// ThrowStmt raises a value. `try` catches it as an ordinary value; the
// interpreter and TS backend preserve the thrown value instead of wrapping it
// in a string message.
type ThrowStmt struct {
	Value Expr
	Pos   lexer.Position
}

func (*ThrowStmt) stmtNode() {}
func (s *ThrowStmt) Position() lexer.Position {
	return s.Pos
}

// TryStmt is statement-form `try { } catch (e) { } finally { }`. Try is not an
// expression, matching TypeScript; a function should return after the block.
type TryStmt struct {
	Body      *BlockExpr
	CatchName string
	Catch     *BlockExpr
	Finally   *BlockExpr
	Pos       lexer.Position
}

func (*TryStmt) stmtNode() {}
func (s *TryStmt) Position() lexer.Position {
	return s.Pos
}

func (*ContinueStmt) stmtNode() {}
func (s *ContinueStmt) Position() lexer.Position {
	return s.Pos
}

type ExprStmt struct {
	Expr Expr
	Pos  lexer.Position
}

type WatchExpr struct {
	Target  Expr
	Handler Expr
	Pos     lexer.Position
}

func (*WatchExpr) exprNode() {}
func (e *WatchExpr) Position() lexer.Position {
	return e.Pos
}

func (*ExprStmt) stmtNode() {}
func (s *ExprStmt) Position() lexer.Position {
	return s.Pos
}
