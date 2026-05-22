package stdlib

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

type Registry struct {
	Modules map[string]*Module
}

type Module struct {
	Name      string
	Functions []Function

	byName  map[string]*Function
	byAlias map[string]*Function
}

type Function struct {
	Name         string
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
