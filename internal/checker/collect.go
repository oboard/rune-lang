package checker

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
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
	for _, trait := range file.Traits {
		if c.info.Traits[trait.Name] != nil || c.info.Types[trait.Name] != nil || c.info.Enums[trait.Name] != nil {
			c.errorf(trait.NamePos, "duplicate type %q", trait.Name)
			continue
		}
		c.info.Traits[trait.Name] = &TraitInfo{
			Name:          trait.Name,
			SourcePath:    trait.SourcePath,
			ByName:        map[string]FieldInfo{},
			Methods:       map[string]*FuncInfo{},
			StaticMethods: map[string]*FuncInfo{},
			Node:          trait,
		}
	}
	for _, typ := range file.Types {
		if _, exists := c.info.Types[typ.Name]; exists || c.info.Traits[typ.Name] != nil {
			c.errorf(typ.NamePos, "duplicate type %q", typ.Name)
			continue
		}
		if _, exists := c.info.Enums[typ.Name]; exists {
			c.errorf(typ.NamePos, "duplicate type %q", typ.Name)
			continue
		}
		c.info.Types[typ.Name] = &StructInfo{
			Name:               typ.Name,
			Private:            typ.Private,
			SourcePath:         typ.SourcePath,
			Generics:           append([]string(nil), typ.Generics...),
			GenericConstraints: c.collectGenericConstraints(typ.GenericConstraints, genericSet(typ.Generics...), typ.NamePos),
			ByName:             map[string]FieldInfo{},
			Methods:            map[string]*FuncInfo{},
			StaticMethods:      map[string]*FuncInfo{},
			Node:               typ,
		}
	}
	for _, enum := range file.Enums {
		if _, exists := c.info.Types[enum.Name]; exists || c.info.Traits[enum.Name] != nil {
			c.errorf(enum.NamePos, "duplicate type %q", enum.Name)
			continue
		}
		if _, exists := c.info.Enums[enum.Name]; exists {
			c.errorf(enum.NamePos, "duplicate type %q", enum.Name)
			continue
		}
		c.info.Enums[enum.Name] = &EnumInfo{Name: enum.Name, Private: enum.Private, SourcePath: enum.SourcePath, Generics: append([]string(nil), enum.Generics...), GenericConstraints: c.collectGenericConstraints(enum.GenericConstraints, genericSet(enum.Generics...), enum.NamePos), ByName: map[string]EnumMemberInfo{}, Node: enum}
	}
	for _, trait := range file.Traits {
		info := c.info.Traits[trait.Name]
		if info == nil || info.Node != trait {
			continue
		}
		self := genericSet("Self")
		for _, field := range trait.Fields {
			if info.ByName[field.Name].Name != "" || info.Methods[field.Name] != nil {
				c.errorf(field.Pos, "duplicate trait member %s.%s", trait.Name, field.Name)
				continue
			}
			typ := c.resolveTypeWithGenerics(field.Type.Canonical(), self)
			if typ == Unknown {
				c.reportUnknownOrPrivateType(field.Pos, field.Type.Canonical())
			}
			member := FieldInfo{Name: field.Name, Type: typ, SourcePath: trait.SourcePath}
			info.Fields = append(info.Fields, member)
			info.ByName[field.Name] = member
		}
		for _, method := range trait.Methods {
			if info.ByName[method.Name].Name != "" || info.Methods[method.Name] != nil || info.StaticMethods[method.Name] != nil {
				c.errorf(method.NamePos, "duplicate trait member %s.%s", trait.Name, method.Name)
				continue
			}
			method.SourcePath = trait.SourcePath
			member := c.collectFunction(method, []string{"Self"}, nil)
			if method.ReturnType.IsZero() {
				member.Return = Void
				member.ReturnDeclared = true
			}
			member.ReceiverType = Type("&" + trait.Name)
			if method.Static {
				info.StaticMethods[method.Name] = member
			} else {
				info.Methods[method.Name] = member
			}
		}
	}
	for _, enum := range file.Enums {
		info := c.info.Enums[enum.Name]
		if info == nil || info.Node != enum {
			continue
		}
		enumGenerics := genericSet(enum.Generics...)
		for _, member := range enum.Members {
			if _, exists := info.ByName[member.Name]; exists {
				c.errorf(member.Pos, "duplicate enum member %s.%s", enum.Name, member.Name)
				continue
			}
			if member.HasValue && len(enum.Generics) > 0 {
				c.errorf(member.Pos, "integer enum member %s.%s cannot belong to generic enum %s", enum.Name, member.Name, enum.Name)
			}
			memberInfo := EnumMemberInfo{Name: member.Name, Private: member.Private, SourcePath: enum.SourcePath, Value: member.Value, HasValue: member.HasValue, Pos: member.Pos}
			for _, param := range member.Params {
				paramName := param.Type.Canonical()
				paramType := c.resolveTypeWithGenerics(paramName, enumGenerics)
				if paramType == Unknown && !isDynamicTypeName(paramName) {
					c.reportUnknownOrPrivateType(param.Pos, paramName)
				}
				memberInfo.Params = append(memberInfo.Params, ParamInfo{Name: param.Name, Type: paramType})
			}
			info.Members = append(info.Members, memberInfo)
			info.ByName[member.Name] = memberInfo
			if !member.HasValue {
				c.info.Constructors[member.Name] = append(c.info.Constructors[member.Name], EnumConstructorInfo{Enum: info, Member: memberInfo})
			}
		}
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
			fieldName := field.Type.Canonical()
			c.withSourcePath(typ.SourcePath, func() {
				fieldType = c.resolveTypeWithGenerics(fieldName, typeGenerics)
			})
			if fieldType == Unknown && !isDynamicTypeName(fieldName) {
				c.reportUnknownOrPrivateType(field.Pos, fieldName)
			}
			fieldInfo := FieldInfo{Name: field.Name, Private: field.Private, SourcePath: typ.SourcePath, Type: fieldType}
			info.Fields = append(info.Fields, fieldInfo)
			info.ByName[field.Name] = fieldInfo
		}
		for _, method := range typ.Methods {
			if _, exists := info.Methods[method.Name]; exists || info.StaticMethods[method.Name] != nil {
				c.errorf(method.NamePos, "duplicate method %s.%s", typ.Name, method.Name)
				continue
			}
			methodInfo := c.collectFunction(method, typ.Generics, info.GenericConstraints)
			methodInfo.ReceiverType = receiverType(typ)
			if method.Static {
				info.StaticMethods[method.Name] = methodInfo
			} else {
				info.Methods[method.Name] = methodInfo
			}
		}
	}
	for _, fn := range file.Functions {
		info := c.collectFunction(fn, nil, nil)
		if !c.addFunction(fn, info) {
			continue
		}
	}
	for _, imp := range file.TSImports {
		for _, fn := range imp.Functions {
			info := c.collectTypeScriptFunction(fn)
			if !c.addFunctionInfo(fn.NamePos, nil, info) {
				continue
			}
			c.info.ExternalFunctions = append(c.info.ExternalFunctions, info)
		}
		for _, value := range imp.Values {
			info := c.collectTypeScriptValue(value)
			if !c.addExternalValue(value.NamePos, info) {
				continue
			}
			c.info.ExternalValues = append(c.info.ExternalValues, info)
		}
	}
}

func (c *checker) addFunction(fn *ast.Function, info *FuncInfo) bool {
	return c.addFunctionInfo(fn.NamePos, fn, info)
}

func (c *checker) addFunctionInfo(pos lexer.Position, fn *ast.Function, info *FuncInfo) bool {
	for _, existing := range c.info.functionsByName[info.Name] {
		if sameSourcePath(existing.SourcePath, info.SourcePath) || (!existing.Private && !info.Private) {
			c.errorf(pos, "duplicate function %q", info.Name)
			return false
		}
	}
	c.info.functionsByName[info.Name] = append(c.info.functionsByName[info.Name], info)
	if fn != nil {
		c.info.FunctionDecls[fn] = info
	}
	if existing := c.info.Functions[info.Name]; existing == nil || (existing.Private && !info.Private) {
		c.info.Functions[info.Name] = info
	}
	return true
}

func (c *checker) collectFunction(fn *ast.Function, inheritedGenerics []string, inheritedConstraints map[string]string) *FuncInfo {
	generics := append([]string(nil), inheritedGenerics...)
	generics = append(generics, fn.Generics...)
	genericTypes := genericSet(generics...)
	constraints := cloneStringMap(inheritedConstraints)
	for name, constraint := range c.collectGenericConstraints(fn.GenericConstraints, genericSet(fn.Generics...), fn.NamePos) {
		constraints[name] = constraint
	}
	info := &FuncInfo{Name: fn.Name, LinkName: fn.Name, Private: fn.Private, Static: fn.Static, Macro: fn.Macro, SourcePath: fn.SourcePath, Routine: fn.Routine, Generics: generics, GenericConstraints: constraints, Node: fn, Return: Unknown, Pos: fn.Pos, NamePos: fn.NamePos}
	if fn.Private && fn.SourcePath != "" && fn.ReceiverType == "" && fn.Name != "main" {
		info.LinkName = privateLinkName(fn.SourcePath, fn.Name)
	}
	seenParams := map[string]bool{}
	c.withSourcePath(fn.SourcePath, func() {
		for _, param := range fn.Params {
			if seenParams[param.Name] {
				c.errorf(param.Pos, "duplicate parameter %q", param.Name)
			}
			seenParams[param.Name] = true
			paramName := param.Type.Canonical()
			if paramName == "" {
				info.Params = append(info.Params, ParamInfo{Name: param.Name, Type: Unknown})
				continue
			}
			typ := c.resolveTypeWithGenerics(paramName, genericTypes)
			if typ == Unknown && !isDynamicTypeName(paramName) {
				c.reportUnknownOrPrivateType(param.Pos, paramName)
			}
			info.Params = append(info.Params, ParamInfo{Name: param.Name, Type: typ})
		}
		returnName := fn.ReturnType.Canonical()
		if returnName != "" {
			typ := c.resolveTypeWithGenerics(returnName, genericTypes)
			if typ == Unknown && !isDynamicTypeName(returnName) {
				if privateName, ok := c.inaccessibleTypeName(returnName); ok {
					c.errorf(fn.NamePos, "return type %q is private", privateName)
				} else {
					c.errorf(fn.NamePos, "unknown return type %q", returnName)
				}
			}
			info.Return = typ
			info.ReturnDeclared = true
		}
	})
	return info
}

func (c *checker) collectGenericConstraints(raw map[string]ast.Type, generics map[string]bool, pos lexer.Position) map[string]string {
	if len(raw) == 0 {
		return nil
	}
	out := map[string]string{}
	for name, constraintType := range raw {
		if !generics[name] {
			continue
		}
		constraint := genericConstraintName(constraintType)
		if constraint == "" {
			continue
		}
		if !c.genericConstraintExists(constraint) {
			c.errorf(pos, "unknown generic constraint %q", constraintType.Display())
			continue
		}
		out[name] = constraint
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func cloneStringMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	if len(in) == 0 {
		return out
	}
	for key, value := range in {
		out[key] = value
	}
	return out
}

func (c *checker) collectTypeScriptFunction(fn ast.TSFunction) *FuncInfo {
	info := &FuncInfo{
		Name:           fn.Name,
		LinkName:       fn.Name,
		External:       true,
		SourcePath:     fn.SourcePath,
		Routine:        fn.Routine,
		Return:         Unknown,
		ReturnDeclared: !fn.ReturnType.IsZero(),
		Pos:            fn.Pos,
		NamePos:        fn.NamePos,
	}
	for _, param := range fn.Params {
		info.Params = append(info.Params, ParamInfo{Name: param.Name, Type: c.resolveExternalType(param.Type.Canonical())})
	}
	if returnName := fn.ReturnType.Canonical(); returnName != "" {
		info.Return = c.resolveExternalType(returnName)
	}
	return info
}

func (c *checker) collectTypeScriptValue(value ast.TSValue) *ExternalValueInfo {
	return &ExternalValueInfo{
		Name:       value.Name,
		LinkName:   value.Name,
		SourcePath: value.SourcePath,
		Type:       c.resolveExternalType(value.Type.Canonical()),
		Pos:        value.Pos,
		NamePos:    value.NamePos,
	}
}

func (c *checker) addExternalValue(pos lexer.Position, info *ExternalValueInfo) bool {
	if c.info.valuesByName[info.Name] != nil || len(c.info.functionsByName[info.Name]) > 0 {
		c.errorf(pos, "duplicate declaration %q", info.Name)
		return false
	}
	c.info.valuesByName[info.Name] = info
	return true
}

func (c *checker) resolveExternalType(name string) Type {
	if name == "" || isDynamicTypeName(name) {
		return Unknown
	}
	return c.resolveType(name)
}
