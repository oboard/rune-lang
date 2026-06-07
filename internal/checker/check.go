package checker

import (
	"net/url"
	"path/filepath"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func Check(file *ast.File) (*Info, []Diagnostic) {
	reg, err := stdlib.LoadDefault()
	info, diags := CheckWithStdlibForPath(file, reg, "")
	if err != nil {
		diags = append([]Diagnostic{{Message: err.Error()}}, diags...)
	}
	return info, diags
}

func CheckWithStdlib(file *ast.File, reg *stdlib.Registry) (*Info, []Diagnostic) {
	return CheckWithStdlibForPath(file, reg, "")
}

func CheckWithStdlibForPath(file *ast.File, reg *stdlib.Registry, sourcePath string) (*Info, []Diagnostic) {
	c := &checker{
		info: &Info{
			Functions:                 map[string]*FuncInfo{},
			FunctionDecls:             map[*ast.Function]*FuncInfo{},
			ResolvedFunctions:         map[*ast.Identifier]*FuncInfo{},
			ResolvedValues:            map[*ast.Identifier]*ExternalValueInfo{},
			ResolvedSelectorFunctions: map[*ast.SelectorExpr]*FuncInfo{},
			ResolvedSelectorValues:    map[*ast.SelectorExpr]*ExternalValueInfo{},
			Types:                     map[string]*StructInfo{},
			Enums:                     map[string]*EnumInfo{},
			Constructors:              map[string][]EnumConstructorInfo{},
			Stdlib:                    reg,
			ExprTypes:                 map[ast.Expr]Type{},
			AsyncCalls:                map[*ast.CallExpr]bool{},
			AwaitCalls:                map[*ast.CallExpr]bool{},
			functionsByName:           map[string][]*FuncInfo{},
			valuesByName:              map[string]*ExternalValueInfo{},
		},
		bindings:           map[string]ast.Expr{},
		inferredFunctions:  map[*ast.Function]bool{},
		inferringFunctions: map[*ast.Function]bool{},
		sourcePath:         normalizeSourcePath(sourcePath),
		currentSourcePath:  normalizeSourcePath(sourcePath),
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
	info               *Info
	diags              []Diagnostic
	bindings           map[string]ast.Expr
	inferredFunctions  map[*ast.Function]bool
	inferringFunctions map[*ast.Function]bool
	genericTypes       map[string]bool
	sourcePath         string
	currentSourcePath  string
	routineDepth       int
	unwrapErrors       []Type
}

func (c *checker) collectCoreTypes() {
	if c.info.Stdlib == nil {
		return
	}
	for _, typ := range c.info.Stdlib.Types {
		if c.isSourceType(typ.SourcePath) {
			continue
		}
		c.info.Types[typ.Name] = &StructInfo{
			Name:       typ.Name,
			Private:    false,
			SourcePath: typ.SourcePath,
			Generics:   append([]string(nil), typ.Generics...),
			ByName:     map[string]FieldInfo{},
			Methods:    map[string]*FuncInfo{},
		}
	}
	for _, typ := range c.info.Stdlib.Types {
		if c.isSourceType(typ.SourcePath) {
			continue
		}
		typeGenerics := genericSet(typ.Generics...)
		info := &StructInfo{
			Name:       typ.Name,
			Private:    false,
			SourcePath: typ.SourcePath,
			Generics:   append([]string(nil), typ.Generics...),
			ByName:     map[string]FieldInfo{},
			Methods:    map[string]*FuncInfo{},
		}
		for _, field := range typ.Fields {
			fieldType := c.resolveTypeWithGenerics(field.Type, typeGenerics)
			fieldInfo := FieldInfo{Name: field.Name, SourcePath: typ.SourcePath, Type: fieldType}
			info.Fields = append(info.Fields, fieldInfo)
			info.ByName[field.Name] = fieldInfo
		}
		c.info.Types[typ.Name] = info
	}
}

func (c *checker) isSourceType(sourcePath string) bool {
	return c.sourcePath != "" && c.sourcePath == normalizeSourcePath(sourcePath)
}

func normalizeSourcePath(path string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "file://") {
		if uri, err := url.Parse(path); err == nil {
			path = uri.Path
		}
	}
	clean := filepath.Clean(path)
	if abs, err := filepath.Abs(clean); err == nil {
		clean = abs
	}
	return clean
}
