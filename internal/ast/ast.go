package ast

import (
	"strings"

	"github.com/oboard/rune-lang/internal/lexer"
)

type File struct {
	Imports   []Import
	GoImports []GoImport
	TSImports []TSImport
	Traits    []*TraitDecl
	Types     []*StructType
	Enums     []*EnumType
	Constants []*ConstDecl
	Functions []*Function
	Tests     []*Test
}

type TraitDecl struct {
	Name       string
	Fields     []Field
	Methods    []*Function
	Pos        lexer.Position
	NamePos    lexer.Position
	SourcePath string
}

type Import struct {
	Path   string
	Module bool
	Pos    lexer.Position
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
	Module    string
	Name      string
	Args      []Expr
	HasParens bool
	Pos       lexer.Position
	NamePos   lexer.Position
}

type StructType struct {
	Name               string
	Private            bool
	Generics           []string
	GenericConstraints map[string]Type
	Annotations        []Annotation
	Fields             []Field
	Methods            []*Function
	Pos                lexer.Position
	NamePos            lexer.Position
	SourcePath         string
}

type Field struct {
	Name        string
	Private     bool
	Annotations []Annotation
	Type        Type
	Pos         lexer.Position
}

type EnumType struct {
	Name               string
	Private            bool
	Generics           []string
	GenericConstraints map[string]Type
	Annotations        []Annotation
	Members            []EnumMember
	Pos                lexer.Position
	NamePos            lexer.Position
	SourcePath         string
}

type EnumMember struct {
	Name        string
	Private     bool
	Annotations []Annotation
	Value       int
	HasValue    bool
	Params      []Param
	Pos         lexer.Position
}

type ConstDecl struct {
	Name       string
	Private    bool
	Type       Type
	Value      Expr
	Pos        lexer.Position
	NamePos    lexer.Position
	SourcePath string
}

type Function struct {
	Name               string
	Private            bool
	Static             bool
	Routine            bool
	Macro              bool
	Generics           []string
	GenericConstraints map[string]Type
	Annotations        []Annotation
	ReceiverType       string
	Params             []Param
	ReturnType         Type
	Body               Expr
	Pos                lexer.Position
	NamePos            lexer.Position
	SourcePath         string
}

func (f *Function) Signature() string {
	var b strings.Builder
	if !f.Private {
		b.WriteString("+ ")
	}
	if f.Routine {
		b.WriteString("~ ")
	}
	if f.Static {
		b.WriteString("::")
	}
	b.WriteString(f.Name)
	if len(f.Generics) > 0 {
		b.WriteByte('[')
		b.WriteString(formatGenericSignature(f.Generics, f.GenericConstraints))
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

func formatGenericSignature(names []string, constraints map[string]Type) string {
	parts := make([]string, 0, len(names))
	for _, name := range names {
		part := name
		if constraint, ok := constraints[name]; ok && !constraint.IsZero() {
			part += ": " + constraint.Display()
		}
		parts = append(parts, part)
	}
	return strings.Join(parts, ", ")
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
