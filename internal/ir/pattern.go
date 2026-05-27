package ir

import (
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/lexer"
)

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

type BindingPattern struct {
	Name string
	Type checker.Type
	Pos  lexer.Position
}

func (*BindingPattern) patternNode() {}
func (p *BindingPattern) Position() lexer.Position {
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

type RangePattern struct {
	Start Expr
	End   Expr
	Pos   lexer.Position
}

func (*RangePattern) patternNode() {}
func (p *RangePattern) Position() lexer.Position {
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

type ConstructorPattern struct {
	Name       string
	Binding    string
	BindingPos lexer.Position
	Pos        lexer.Position
}

func (*ConstructorPattern) patternNode() {}
func (p *ConstructorPattern) Position() lexer.Position {
	return p.Pos
}

type MapPatternEntry struct {
	Key      Expr
	Pattern  Pattern
	Optional bool
	Pos      lexer.Position
}

type MapPattern struct {
	Entries []MapPatternEntry
	Rest    bool
	Pos     lexer.Position
}

func (*MapPattern) patternNode() {}
func (p *MapPattern) Position() lexer.Position {
	return p.Pos
}

type ObjectPatternField struct {
	Name     string
	Pattern  Pattern
	Optional bool
	Exists   bool
	Type     checker.Type
	Pos      lexer.Position
}

type ObjectPattern struct {
	Fields []ObjectPatternField
	Rest   bool
	Pos    lexer.Position
}

func (*ObjectPattern) patternNode() {}
func (p *ObjectPattern) Position() lexer.Position {
	return p.Pos
}
