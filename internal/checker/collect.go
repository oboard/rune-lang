package checker

import (
	"github.com/oboard/rune-lang/internal/ast"
)

func (c *checker) checkGoImports(file *ast.File) {
	for _, imp := range file.GoImports {
		fn, ok := c.info.Stdlib.Function("go", "import")
		if !ok || fn.Intrinsic != "go.import" || !fn.TopLevelOnly {
			c.errorf(imp.Pos, "@go.import is not declared in core/go")
		}
	}
}

func (c *checker) collect(file *ast.File) {
	for _, typ := range file.Types {
		if _, exists := c.info.Types[typ.Name]; exists {
			c.errorf(typ.NamePos, "duplicate type %q", typ.Name)
			continue
		}
		if _, exists := c.info.Enums[typ.Name]; exists {
			c.errorf(typ.NamePos, "duplicate type %q", typ.Name)
			continue
		}
		c.info.Types[typ.Name] = &StructInfo{
			Name:     typ.Name,
			Generics: append([]string(nil), typ.Generics...),
			ByName:   map[string]FieldInfo{},
			Methods:  map[string]*FuncInfo{},
			Node:     typ,
		}
	}
	for _, enum := range file.Enums {
		if _, exists := c.info.Types[enum.Name]; exists {
			c.errorf(enum.NamePos, "duplicate type %q", enum.Name)
			continue
		}
		if _, exists := c.info.Enums[enum.Name]; exists {
			c.errorf(enum.NamePos, "duplicate type %q", enum.Name)
			continue
		}
		info := &EnumInfo{Name: enum.Name, ByName: map[string]EnumMemberInfo{}, Node: enum}
		for _, member := range enum.Members {
			if _, exists := info.ByName[member.Name]; exists {
				c.errorf(member.Pos, "duplicate enum member %s.%s", enum.Name, member.Name)
				continue
			}
			memberInfo := EnumMemberInfo{Name: member.Name, Value: member.Value}
			info.Members = append(info.Members, memberInfo)
			info.ByName[member.Name] = memberInfo
		}
		c.info.Enums[enum.Name] = info
	}
	for _, typ := range file.Types {
		info := c.info.Types[typ.Name]
		if info == nil {
			continue
		}
		typeGenerics := genericSet(typ.Generics...)
		for _, field := range typ.Fields {
			if _, exists := info.ByName[field.Name]; exists {
				c.errorf(field.Pos, "duplicate field %q", field.Name)
				continue
			}
			fieldType := c.resolveTypeWithGenerics(field.Type, typeGenerics)
			if fieldType == Unknown && !isDynamicTypeName(field.Type) {
				c.errorf(field.Pos, "unknown type %q", field.Type)
			}
			fieldInfo := FieldInfo{Name: field.Name, Type: fieldType}
			info.Fields = append(info.Fields, fieldInfo)
			info.ByName[field.Name] = fieldInfo
		}
		for _, method := range typ.Methods {
			if _, exists := info.Methods[method.Name]; exists {
				c.errorf(method.NamePos, "duplicate method %s.%s", typ.Name, method.Name)
				continue
			}
			methodInfo := c.collectFunction(method, typ.Generics)
			methodInfo.ReceiverType = receiverType(typ)
			info.Methods[method.Name] = methodInfo
		}
	}
	for _, fn := range file.Functions {
		if _, exists := c.info.Functions[fn.Name]; exists {
			c.errorf(fn.NamePos, "duplicate function %q", fn.Name)
			continue
		}
		c.info.Functions[fn.Name] = c.collectFunction(fn, nil)
	}
}

func (c *checker) collectFunction(fn *ast.Function, inheritedGenerics []string) *FuncInfo {
	generics := append([]string(nil), inheritedGenerics...)
	generics = append(generics, fn.Generics...)
	genericTypes := genericSet(generics...)
	info := &FuncInfo{Name: fn.Name, Generics: generics, Node: fn, Return: Unknown}
	seenParams := map[string]bool{}
	for _, param := range fn.Params {
		if seenParams[param.Name] {
			c.errorf(param.Pos, "duplicate parameter %q", param.Name)
		}
		seenParams[param.Name] = true
		if param.Type == "" {
			info.Params = append(info.Params, ParamInfo{Name: param.Name, Type: Unknown})
			continue
		}
		typ := c.resolveTypeWithGenerics(param.Type, genericTypes)
		if typ == Unknown && !isDynamicTypeName(param.Type) {
			c.errorf(param.Pos, "unknown type %q", param.Type)
		}
		info.Params = append(info.Params, ParamInfo{Name: param.Name, Type: typ})
	}
	if fn.ReturnType != "" {
		typ := c.resolveTypeWithGenerics(fn.ReturnType, genericTypes)
		if typ == Unknown && !isDynamicTypeName(fn.ReturnType) {
			c.errorf(fn.NamePos, "unknown return type %q", fn.ReturnType)
		}
		info.Return = typ
		info.ReturnDeclared = true
	}
	return info
}
