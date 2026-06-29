package ast

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

type BindingPattern struct {
	Name     string
	Type     string
	Constant bool
	LinkName string
	Pos      lexer.Position
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
	Start     Expr
	End       Expr
	Inclusive bool
	Pos       lexer.Position
}

func (*RangePattern) patternNode() {}
func (p *RangePattern) Position() lexer.Position {
	return p.Pos
}

type OrPattern struct {
	Alternatives []Pattern
	Pos          lexer.Position
}

func (*OrPattern) patternNode() {}
func (p *OrPattern) Position() lexer.Position {
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

type ArrayPattern struct {
	Elements    []Pattern
	RestIndex   int
	RestBinding string
	RestType    string
	RestPos     lexer.Position
	SubjectType string
	Pos         lexer.Position
}

func (*ArrayPattern) patternNode() {}
func (p *ArrayPattern) Position() lexer.Position {
	return p.Pos
}

type SequenceSpreadPattern struct {
	Value Expr
	Type  string
	Pos   lexer.Position
}

func (*SequenceSpreadPattern) patternNode() {}
func (p *SequenceSpreadPattern) Position() lexer.Position {
	return p.Pos
}

type BitPattern struct {
	Width  int
	Signed bool
	Endian string
	Value  Pattern
	Pos    lexer.Position
}

func (*BitPattern) patternNode() {}
func (p *BitPattern) Position() lexer.Position {
	return p.Pos
}

type AsPattern struct {
	Pattern Pattern
	Name    string
	Type    string
	NamePos lexer.Position
	Pos     lexer.Position
}

func (*AsPattern) patternNode() {}
func (p *AsPattern) Position() lexer.Position {
	return p.Pos
}

type ConstructorPattern struct {
	Name        string
	Args        []Pattern
	Rest        bool
	RestPos     lexer.Position
	Binding     string
	BindingPos  lexer.Position
	SubjectType string
	Pos         lexer.Position
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
	Entries     []MapPatternEntry
	Rest        bool
	SubjectType string
	ValueType   string
	Access      string
	Pos         lexer.Position
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
	Type     string
	Pos      lexer.Position
}

type ObjectPattern struct {
	Fields      []ObjectPatternField
	Rest        bool
	SubjectType string
	Pos         lexer.Position
}

func (*ObjectPattern) patternNode() {}
func (p *ObjectPattern) Position() lexer.Position {
	return p.Pos
}
