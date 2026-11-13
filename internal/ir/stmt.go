package ir

import (
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/lexer"
)

type Stmt interface {
	stmtNode()
	Position() lexer.Position
}

type LetStmt struct {
	Name    string
	Mutable bool
	Signal  bool
	Value   Expr
	Type    checker.Type
	Pos     lexer.Position
}

func (*LetStmt) stmtNode() {}
func (s *LetStmt) Position() lexer.Position {
	return s.Pos
}

type ObjectBindingField struct {
	Field    string
	Name     string
	Type     checker.Type
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

// ThrowStmt raises a value that `try` can catch.
type ThrowStmt struct {
	Value Expr
	Pos   lexer.Position
}

func (*ThrowStmt) stmtNode() {}
func (s *ThrowStmt) Position() lexer.Position {
	return s.Pos
}

// TryStmt is the statement form of try/catch/finally.
type TryStmt struct {
	Body      []Stmt
	CatchName string
	Catch     []Stmt
	Finally   []Stmt
	Pos       lexer.Position
}

func (*TryStmt) stmtNode() {}
func (s *TryStmt) Position() lexer.Position {
	return s.Pos
}

func (*ExprStmt) stmtNode() {}
func (s *ExprStmt) Position() lexer.Position {
	return s.Pos
}
