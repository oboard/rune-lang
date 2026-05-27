package ir

import (
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/lexer"
	"github.com/oboard/rune-lang/internal/stdlib"
)

type Package struct {
	Name    string
	Modules []*Module
}

type Module struct {
	Name  string
	Files []*File
}

type File struct {
	GoImports []GoImport
	TSImports []TSImport
	Types     []*StructType
	Enums     []*EnumType
	Functions []*Function
	Tests     []*Test
	Stdlib    *stdlib.Registry
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
	Name string
}

type TSValue struct {
	Name string
}

type StructType struct {
	Name     string
	Private  bool
	Generics []string
	Fields   []Field
	Methods  []*Function
	Pos      lexer.Position
	NamePos  lexer.Position
}

type Field struct {
	Name    string
	Private bool
	Type    checker.Type
	Pos     lexer.Position
}

type EnumType struct {
	Name    string
	Private bool
	Members []EnumMember
	Pos     lexer.Position
	NamePos lexer.Position
}

type EnumMember struct {
	Name    string
	Private bool
	Value   int
	Pos     lexer.Position
}

type Function struct {
	Name         string
	Private      bool
	Routine      bool
	Generics     []string
	ReceiverType checker.Type
	Params       []Param
	Return       checker.Type
	Body         Expr
	Pos          lexer.Position
	NamePos      lexer.Position
}

type Param struct {
	Name string
	Type checker.Type
	Pos  lexer.Position
}

type Test struct {
	Name string
	Body Expr
	Pos  lexer.Position
}
