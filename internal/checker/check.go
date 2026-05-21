package checker

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func Check(file *ast.File) (*Info, []Diagnostic) {
	reg, err := stdlib.LoadDefault()
	c := &checker{
		info: &Info{
			Functions: map[string]*FuncInfo{},
			Types:     map[string]*StructInfo{},
			Stdlib:    reg,
			ExprTypes: map[ast.Expr]Type{},
		},
	}
	if err != nil {
		c.errorf(lexer.Position{}, "%s", err.Error())
	}
	c.checkGoImports(file)
	c.collect(file)
	for _, typ := range file.Types {
		c.inferMethods(typ)
	}
	for _, fn := range file.Functions {
		c.inferFunction(fn)
	}
	return c.info, c.diags
}

type checker struct {
	info  *Info
	diags []Diagnostic
}
