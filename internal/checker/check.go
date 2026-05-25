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
			Functions:  map[string]*FuncInfo{},
			Types:      map[string]*StructInfo{},
			Enums:      map[string]*EnumInfo{},
			Stdlib:     reg,
			ExprTypes:  map[ast.Expr]Type{},
			AsyncCalls: map[*ast.CallExpr]bool{},
			AwaitCalls: map[*ast.CallExpr]bool{},
		},
		bindings: map[string]ast.Expr{},
	}
	if err != nil {
		c.errorf(lexer.Position{}, "%s", err.Error())
	}
	c.collectCoreTypes()
	c.checkGoImports(file)
	c.collect(file)
	for _, typ := range file.Types {
		c.inferMethods(typ)
	}
	for _, fn := range file.Functions {
		c.inferFunction(fn)
	}
	for _, test := range file.Tests {
		c.inferTest(test)
	}
	return c.info, c.diags
}

type checker struct {
	info         *Info
	diags        []Diagnostic
	bindings     map[string]ast.Expr
	routineDepth int
	unwrapErrors []Type
}

func (c *checker) collectCoreTypes() {
	if c.info.Stdlib == nil {
		return
	}
	for _, typ := range c.info.Stdlib.Types {
		c.info.Types[typ.Name] = &StructInfo{
			Name:     typ.Name,
			Generics: append([]string(nil), typ.Generics...),
			ByName:   map[string]FieldInfo{},
			Methods:  map[string]*FuncInfo{},
		}
	}
	for _, typ := range c.info.Stdlib.Types {
		typeGenerics := genericSet(typ.Generics...)
		info := &StructInfo{
			Name:     typ.Name,
			Generics: append([]string(nil), typ.Generics...),
			ByName:   map[string]FieldInfo{},
			Methods:  map[string]*FuncInfo{},
		}
		for _, field := range typ.Fields {
			fieldType := c.resolveTypeWithGenerics(field.Type, typeGenerics)
			fieldInfo := FieldInfo{Name: field.Name, Type: fieldType}
			info.Fields = append(info.Fields, fieldInfo)
			info.ByName[field.Name] = fieldInfo
		}
		c.info.Types[typ.Name] = info
	}
}
