package stdlib

import "github.com/oboard/rune-lang/internal/ast"

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
	Receiver     string
	Generics     []string
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
