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
			Name:       typ.Name,
			Private:    typ.Private,
			SourcePath: typ.SourcePath,
			Generics:   append([]string(nil), typ.Generics...),
			ByName:     map[string]FieldInfo{},
			Methods:    map[string]*FuncInfo{},
			Node:       typ,
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
		info := &EnumInfo{Name: enum.Name, Private: enum.Private, SourcePath: enum.SourcePath, ByName: map[string]EnumMemberInfo{}, Node: enum}
		for _, member := range enum.Members {
			if _, exists := info.ByName[member.Name]; exists {
				c.errorf(member.Pos, "duplicate enum member %s.%s", enum.Name, member.Name)
				continue
			}
			memberInfo := EnumMemberInfo{Name: member.Name, Private: member.Private, SourcePath: enum.SourcePath, Value: member.Value}
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
			var fieldType Type
			c.withSourcePath(typ.SourcePath, func() {
				fieldType = c.resolveTypeWithGenerics(field.Type, typeGenerics)
			})
			if fieldType == Unknown && !isDynamicTypeName(field.Type) {
				c.reportUnknownOrPrivateType(field.Pos, field.Type)
			}
			fieldInfo := FieldInfo{Name: field.Name, Private: field.Private, SourcePath: typ.SourcePath, Type: fieldType}
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
		info := c.collectFunction(fn, nil)
		if !c.addFunction(fn, info) {
			continue
		}
	}
}

func (c *checker) addFunction(fn *ast.Function, info *FuncInfo) bool {
	for _, existing := range c.info.functionsByName[fn.Name] {
		if sameSourcePath(existing.SourcePath, info.SourcePath) || (!existing.Private && !info.Private) {
			c.errorf(fn.NamePos, "duplicate function %q", fn.Name)
			return false
		}
	}
	c.info.functionsByName[fn.Name] = append(c.info.functionsByName[fn.Name], info)
	c.info.FunctionDecls[fn] = info
	if existing := c.info.Functions[fn.Name]; existing == nil || (existing.Private && !info.Private) {
		c.info.Functions[fn.Name] = info
	}
	return true
}

func (c *checker) collectFunction(fn *ast.Function, inheritedGenerics []string) *FuncInfo {
	generics := append([]string(nil), inheritedGenerics...)
	generics = append(generics, fn.Generics...)
	genericTypes := genericSet(generics...)
	info := &FuncInfo{Name: fn.Name, LinkName: fn.Name, Private: fn.Private, SourcePath: fn.SourcePath, Routine: fn.Routine, Generics: generics, Node: fn, Return: Unknown}
	if fn.Private {
		info.LinkName = privateLinkName(fn.SourcePath, fn.Name)
	}
	seenParams := map[string]bool{}
	c.withSourcePath(fn.SourcePath, func() {
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
				c.reportUnknownOrPrivateType(param.Pos, param.Type)
			}
			info.Params = append(info.Params, ParamInfo{Name: param.Name, Type: typ})
		}
		if fn.ReturnType != "" {
			typ := c.resolveTypeWithGenerics(fn.ReturnType, genericTypes)
			if typ == Unknown && !isDynamicTypeName(fn.ReturnType) {
				if privateName, ok := c.inaccessibleTypeName(fn.ReturnType); ok {
					c.errorf(fn.NamePos, "return type %q is private", privateName)
				} else {
					c.errorf(fn.NamePos, "unknown return type %q", fn.ReturnType)
				}
			}
			info.Return = typ
			info.ReturnDeclared = true
		}
	})
	return info
}
