package stdlib

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

type Registry struct {
	Modules map[string]*Module
	Types   map[string]*Type
}

type Module struct {
	Name      string
	Functions []Function
	Types     []Type

	byName     map[string]*Function
	byReceiver map[string]map[string]*Function
	byAlias    map[string]*Function
}

type Type struct {
	Name         string
	SourcePath   string
	Pos          lexer.Position
	Generics     []string
	Fields       []Field
	Constructors []Constructor
}

type Field struct {
	Name string
	Type string
	Pos  lexer.Position
}

type Constructor struct {
	Name       string
	ParamNames []string
	Params     []string
	Pos        lexer.Position
}

type Function struct {
	Name         string
	Routine      bool
	SourcePath   string
	Pos          lexer.Position
	Receiver     string
	Generics     []string
	ParamNames   []string
	Params       []string
	Return       string
	Alias        string
	Variadic     bool
	TopLevelOnly bool
	Intrinsic    string
	Body         ast.Expr
	Go           *GoBinding
}

type GoBinding struct {
	Import string
	Symbol string
}
