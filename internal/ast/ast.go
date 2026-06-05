package ast

import (
	"strings"

	"github.com/oboard/rune-lang/internal/lexer"
)

type File struct {
	Imports   []Import
	GoImports []GoImport
	TSImports []TSImport
	Types     []*StructType
	Enums     []*EnumType
	Functions []*Function
	Tests     []*Test
}

type Import struct {
	Path string
	Pos  lexer.Position
}

type GoImport struct {
	Path string
	Pos  lexer.Position
}

type TSImport struct {
	Path      string
	Specifier string
	Pos       lexer.Position
	Functions []TSFunction
	Values    []TSValue
}

type TSFunction struct {
	Name       string
	Routine    bool
	Params     []Param
	ReturnType Type
	Pos        lexer.Position
	NamePos    lexer.Position
	SourcePath string
}

type TSValue struct {
	Name       string
	Type       Type
	Pos        lexer.Position
	NamePos    lexer.Position
	SourcePath string
}

type Annotation struct {
	Name  string
	Value string
	Pos   lexer.Position
}

type StructType struct {
	Name       string
	Private    bool
	Generics   []string
	Fields     []Field
	Methods    []*Function
	Pos        lexer.Position
	NamePos    lexer.Position
	SourcePath string
}

type Field struct {
	Name    string
	Private bool
	Type    Type
	Pos     lexer.Position
}

type EnumType struct {
	Name       string
	Private    bool
	Generics   []string
	Members    []EnumMember
	Pos        lexer.Position
	NamePos    lexer.Position
	SourcePath string
}

type EnumMember struct {
	Name     string
	Private  bool
	Value    int
	HasValue bool
	Params   []Param
	Pos      lexer.Position
}

type Function struct {
	Name         string
	Private      bool
	Routine      bool
	Generics     []string
	Annotations  []Annotation
	ReceiverType string
	Params       []Param
	ReturnType   Type
	Body         Expr
	Pos          lexer.Position
	NamePos      lexer.Position
	SourcePath   string
}

func (f *Function) Signature() string {
	var b strings.Builder
	if !f.Private {
		b.WriteString("+ ")
	}
	if f.Routine {
		b.WriteString("~ ")
	}
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
		b.WriteString(param.Type.Display())
	}
	b.WriteByte(')')
	return b.String()
}

type Param struct {
	Name string
	Type Type
	Pos  lexer.Position
}

type Test struct {
	Name       string
	Body       Expr
	Pos        lexer.Position
	NamePos    lexer.Position
	SourcePath string
}
