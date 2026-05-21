package ir

import "github.com/oboard/rune-lang/internal/lexer"

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
