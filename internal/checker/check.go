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
			ResolvedMacros:            map[*ast.Annotation]*stdlib.Function{},
			ResolvedMacroFunctions:    map[*ast.Annotation]*FuncInfo{},
			Types:                     map[string]*StructInfo{},
			Traits:                    map[string]*TraitInfo{},
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
		expectedType:       Unknown,
		sourcePath:         normalizeSourcePath(sourcePath),
		currentSourcePath:  normalizeSourcePath(sourcePath),
	}
	c.collectCoreTraits()
	c.collectCoreTypes()
	c.checkGoImports(file)
	c.collect(file)
	c.checkMacros(file)
	for _, typ := range file.Types {
		c.inferMethods(typ)
	}
	for _, fn := range file.Functions {
		c.inferFunction(fn)
	}
	for _, test := range file.Tests {
		c.inferTest(test)
	}
	c.checkMacroPurity(file)
	return c.info, c.diags
}

type checker struct {
	info               *Info
	diags              []Diagnostic
	bindings           map[string]ast.Expr
	inferredFunctions  map[*ast.Function]bool
	inferringFunctions map[*ast.Function]bool
	genericTypes       map[string]bool
	genericConstraints map[string]string
	expectedType       Type
	sourcePath         string
	currentSourcePath  string
	routineDepth       int
	unwrapErrors       []Type
	stdlibMacroPurity  map[*stdlib.Function]string
}

func (c *checker) collectCoreTraits() {
	if c.info.Stdlib == nil {
		return
	}
	for _, trait := range c.info.Stdlib.Traits {
		if c.isSourceType(trait.SourcePath) {
			continue
		}
		c.info.Traits[trait.Name] = &TraitInfo{
			Name:          trait.Name,
			SourcePath:    trait.SourcePath,
			ByName:        map[string]FieldInfo{},
			Methods:       map[string]*FuncInfo{},
			StaticMethods: map[string]*FuncInfo{},
		}
	}
	for _, trait := range c.info.Stdlib.Traits {
		if c.isSourceType(trait.SourcePath) {
			continue
		}
		info := c.info.Traits[trait.Name]
		if info == nil {
			continue
		}
		generics := genericSet("Self")
		for _, field := range trait.Fields {
			fieldType := c.resolveTypeWithGenerics(field.Type, generics)
			member := FieldInfo{Name: field.Name, SourcePath: trait.SourcePath, Type: fieldType}
			info.Fields = append(info.Fields, member)
			info.ByName[field.Name] = member
		}
		for _, method := range trait.Methods {
			member := &FuncInfo{
				Name:           method.Name,
				LinkName:       method.Name,
				Static:         method.Static,
				SourcePath:     trait.SourcePath,
				Return:         c.resolveTypeWithGenerics(method.Return, generics),
				ReturnDeclared: true,
				ReceiverType:   Type("&" + trait.Name),
				Pos:            method.Pos,
				NamePos:        method.Pos,
			}
			for idx, paramType := range method.Params {
				name := ""
				if idx < len(method.ParamNames) {
					name = method.ParamNames[idx]
				}
				member.Params = append(member.Params, ParamInfo{
					Name: name,
					Type: c.resolveTypeWithGenerics(paramType, generics),
				})
			}
			if method.Static {
				info.StaticMethods[method.Name] = member
			} else {
				info.Methods[method.Name] = member
			}
		}
	}
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
			Name:               typ.Name,
			Private:            false,
			SourcePath:         typ.SourcePath,
			Generics:           append([]string(nil), typ.Generics...),
			GenericConstraints: cloneStringMap(typ.GenericConstraints),
			ByName:             map[string]FieldInfo{},
			Methods:            map[string]*FuncInfo{},
			StaticMethods:      map[string]*FuncInfo{},
		}
	}
	for _, typ := range c.info.Stdlib.Types {
		if c.isSourceType(typ.SourcePath) {
			continue
		}
		typeGenerics := genericSet(typ.Generics...)
		info := &StructInfo{
			Name:               typ.Name,
			Private:            false,
			SourcePath:         typ.SourcePath,
			Generics:           append([]string(nil), typ.Generics...),
			GenericConstraints: cloneStringMap(typ.GenericConstraints),
			ByName:             map[string]FieldInfo{},
			Methods:            map[string]*FuncInfo{},
			StaticMethods:      map[string]*FuncInfo{},
		}
		for _, field := range typ.Fields {
			fieldType := c.resolveTypeWithGenerics(field.Type, typeGenerics)
			fieldInfo := FieldInfo{Name: field.Name, SourcePath: typ.SourcePath, Type: fieldType}
			info.Fields = append(info.Fields, fieldInfo)
			info.ByName[field.Name] = fieldInfo
		}
		c.info.Types[typ.Name] = info
	}
	for _, mod := range c.info.Stdlib.Modules {
		for i := range mod.Functions {
			fn := &mod.Functions[i]
			if fn.Receiver == "" {
				continue
			}
			typeInfo := c.info.Types[fn.Receiver]
			if typeInfo == nil {
				continue
			}
			method := c.coreFunctionInfo(fn, typeInfo)
			typeInfo.Methods[fn.Name] = method
		}
	}
}

func (c *checker) coreFunctionInfo(fn *stdlib.Function, receiver *StructInfo) *FuncInfo {
	generics := append([]string(nil), receiver.Generics...)
	generics = append(generics, fn.Generics...)
	constraints := cloneStringMap(receiver.GenericConstraints)
	for name, constraint := range fn.GenericConstraints {
		constraints[name] = constraint
	}
	genericTypes := genericSet(generics...)
	info := &FuncInfo{
		Name:               fn.Name,
		LinkName:           fn.Name,
		SourcePath:         fn.SourcePath,
		Routine:            fn.Routine,
		Generics:           generics,
		GenericConstraints: constraints,
		ReceiverType:       coreReceiverType(receiver),
		Return:             c.resolveTypeWithGenerics(fn.Return, genericTypes),
		ReturnDeclared:     true,
		Pos:                fn.Pos,
		NamePos:            fn.Pos,
	}
	for idx, paramType := range fn.Params {
		name := ""
		if idx < len(fn.ParamNames) {
			name = fn.ParamNames[idx]
		}
		info.Params = append(info.Params, ParamInfo{
			Name: name,
			Type: c.resolveTypeWithGenerics(paramType, genericTypes),
		})
	}
	return info
}

func coreReceiverType(info *StructInfo) Type {
	if info == nil || len(info.Generics) == 0 {
		if info == nil {
			return Unknown
		}
		return Type(info.Name)
	}
	return Type(info.Name + "[" + strings.Join(info.Generics, ",") + "]")
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
