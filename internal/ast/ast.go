package ast

import (
	"strings"

	"github.com/oboard/rune-lang/internal/lexer"
)

type File struct {
	GoImports []GoImport
	Types     []*StructType
	Enums     []*EnumType
	Functions []*Function
	Tests     []*Test
}

type GoImport struct {
	Path string
	Pos  lexer.Position
}

type Annotation struct {
	Name  string
	Value string
	Pos   lexer.Position
}

type StructType struct {
	Name     string
	Generics []string
	Fields   []Field
	Methods  []*Function
	Pos      lexer.Position
	NamePos  lexer.Position
}

type Field struct {
	Name        string
	Type        string
	TypeDisplay string
	Pos         lexer.Position
}

type EnumType struct {
	Name    string
	Members []EnumMember
	Pos     lexer.Position
	NamePos lexer.Position
}

type EnumMember struct {
	Name  string
	Value int
	Pos   lexer.Position
}

type Function struct {
	Name          string
	Generics      []string
	Annotations   []Annotation
	ReceiverType  string
	Params        []Param
	ReturnType    string
	ReturnDisplay string
	Body          Expr
	Pos           lexer.Position
	NamePos       lexer.Position
}

func (f *Function) Signature() string {
	var b strings.Builder
	b.WriteString(f.Name)
	if len(f.Generics) > 0 {
		b.WriteByte('[')
		b.WriteString(strings.Join(f.Generics, ", "))
		b.WriteByte(']')
	}
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
	Name        string
	Type        string
	TypeDisplay string
	Pos         lexer.Position
}

type Test struct {
	Name    string
	Body    Expr
	Pos     lexer.Position
	NamePos lexer.Position
}
