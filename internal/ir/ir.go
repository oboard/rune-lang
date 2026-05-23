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
	Types     []*StructType
	Functions []*Function
	Tests     []*Test
	Stdlib    *stdlib.Registry
}

type GoImport struct {
	Path string
	Pos  lexer.Position
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
	Name string
	Type checker.Type
	Pos  lexer.Position
}

type Function struct {
	Name         string
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
