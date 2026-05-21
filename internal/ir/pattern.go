package ir

import "github.com/oboard/rune-lang/internal/lexer"

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
