package tsfront

import (
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/astprint"
	"github.com/oboard/rune-lang/internal/parser"
)

type classInfo struct {
	name       string
	base       string
	fields     []classField
	staticVals []classStaticField
	methods    []*ast.Function
	statics    []*ast.Function
	ctor       map[string]any
	ctorParams []ast.Param
	accessors  map[string]accessorInfo
}

type accessorInfo struct {
	get bool
	set bool
}

type classStaticField struct {
	name    string
	private bool
}

// class registers a TypeScript class and converts it (and its base classes) on
// demand. Inheritance is flattened: the derived Rune struct contains base
// fields and copies of base methods, which is sound for direct `new Child()`
// usage. Base-typed polymorphism is intentionally not modelled yet.
func (c *converter) class(node map[string]any, file *ast.File) {
	id, ok := mapField(node, "id")
	if !ok {
		c.errorf(node, "class declaration requires a name")
		return
	}
	name := stringField(id, "name")
	if name == "" {
		c.errorf(node, "class declaration requires a name")
		return
	}
	c.file = file
	if _, exists := c.classNodes[name]; !exists {
		c.classNodes[name] = node
		c.classes[name] = true
	}
	_, _ = c.ensureClass(name)
}

func (c *converter) ensureClass(name string) (*classInfo, bool) {
	if info, ok := c.classInfos[name]; ok {
		return info, true
	}
	if c.converting[name] {
		c.errorf(c.classNodes[name], "cyclic class inheritance involving %q", name)
		return nil, false
	}
	node, ok := c.classNodes[name]
	if !ok {
		c.errorf(nil, "unknown class %q", name)
		return nil, false
	}
	c.converting[name] = true
	defer delete(c.converting, name)

	baseName, ok := c.classBaseName(node)
	if !ok {
		return nil, false
	}
	info := &classInfo{name: name, base: baseName, accessors: map[string]accessorInfo{}}
	c.classInfos[name] = info
	success := false
	defer func() {
		if !success {
			delete(c.classInfos, name)
		}
	}()
	var baseInfo *classInfo
	if baseName != "" {
		baseInfo, ok = c.ensureClass(baseName)
		if !ok {
			return nil, false
		}
	}
	if !c.parseClassInfo(name, node, info) {
		return nil, false
	}
	if baseName != "" {
		info.fields = mergeClassFields(baseInfo.fields, info.fields)
		info.methods = mergeClassMethods(info.methods, baseInfo.methods, name)
		info.methods = append(info.methods, classSuperAliases(info.methods, baseInfo.methods, name)...)
		info.statics = c.mergeClassStatics(info, baseInfo)
	}
	c.file.Types = append(c.file.Types, &ast.StructType{
		Name:    name,
		Fields:  buildClassFields(info.fields),
		Methods: info.methods,
		Pos:     lexerPosition(node),
		NamePos: lexerPosition(classID(node)),
	})
	var factory *ast.Function
	if baseInfo != nil {
		factory, ok = c.derivedClassFactory(info, baseInfo)
	} else {
		factory, ok = c.classConstructor(name, info.ctor, info.fields)
	}
	if !ok {
		return nil, false
	}
	c.file.Functions = append(c.file.Functions, factory)
	c.classInfos[name] = info
	success = true
	return info, true
}

func (c *converter) mergeClassStatics(info *classInfo, base *classInfo) []*ast.Function {
	out := append([]*ast.Function(nil), info.statics...)
	seen := map[string]bool{}
	for _, fn := range out {
		seen[fn.Name] = true
	}
	for _, fn := range base.statics {
		original := strings.TrimPrefix(fn.Name, base.name+"_")
		name := info.name + "_" + original
		if seen[name] {
			continue
		}
		seen[name] = true
		clone := cloneStaticMethod(fn, name)
		out = append(out, clone)
		c.file.Functions = append(c.file.Functions, clone)
	}
	return out
}

func cloneStaticMethod(fn *ast.Function, name string) *ast.Function {
	copyFn := *fn
	copyFn.Name = name
	copyFn.ReceiverType = ""
	copyFn.Static = true
	source := astprint.File(&ast.File{Functions: []*ast.Function{&copyFn}})
	parsed, errs := parser.Parse(source)
	if len(errs) == 0 && len(parsed.Functions) == 1 {
		out := parsed.Functions[0]
		out.Name = name
		out.ReceiverType = ""
		out.Static = true
		return out
	}
	return &copyFn
}

// classSuperAliases exposes every base method on the derived type under
// `__super_<name>`, so `super.method()` can be rewritten to
// `this.__super_method()` without introducing free-function receiver calls.
func classSuperAliases(existing []*ast.Function, base []*ast.Function, receiver string) []*ast.Function {
	seen := map[string]bool{}
	for _, method := range existing {
		seen[method.Name] = true
	}
	out := make([]*ast.Function, 0, len(base))
	for _, method := range base {
		name := method.Name
		if !strings.HasPrefix(name, "__super_") {
			name = "__super_" + name
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		clone := cloneClassMethod(method, receiver)
		clone.Name = name
		out = append(out, clone)
	}
	return out
}

func (c *converter) parseClassInfo(name string, node map[string]any, info *classInfo) bool {
	body, ok := mapField(node, "body")
	if !ok {
		c.errorf(node, "class body is required")
		return false
	}
	for _, raw := range arrayField(body, "body") {
		member, ok := asObject(raw)
		if !ok || stringField(member, "type") != "MethodDefinition" {
			continue
		}
		kind := stringField(member, "kind")
		if kind != "get" {
			continue
		}
		property := classMemberName(member)
		accessor := info.accessors[property]
		accessor.get = true
		info.accessors[property] = accessor
	}
	for _, raw := range arrayField(body, "body") {
		member, ok := asObject(raw)
		if !ok {
			continue
		}
		switch stringField(member, "type") {
		case "PropertyDefinition", "ClassProperty":
			if boolField(member, "static") {
				if !c.classStaticField(name, member, info) {
					continue
				}
				continue
			}
			field, ok := c.classProperty(member)
			if !ok {
				continue
			}
			info.fields = append(info.fields, field)
		case "MethodDefinition", "TSAbstractMethodDefinition":
			kind := stringField(member, "kind")
			memberName := classMemberName(member)
			if kind == "constructor" {
				info.ctor, _ = mapField(member, "value")
				if info.ctor != nil {
					params, ok := c.functionParams(info.ctor)
					if !ok {
						return false
					}
					info.ctorParams = params
				}
				continue
			}
			if kind == "get" || kind == "set" {
				if boolField(member, "static") {
					c.errorCodef(member, CodeProfileClass, "static accessors are not supported")
					continue
				}
				if kind == "set" {
					c.errorCodef(member, CodeProfileClass, "set accessors are not supported until field mutation lands")
					continue
				}
				value, ok := mapField(member, "value")
				if !ok {
					continue
				}
				value = cloneNode(value)
				methodName := kind + "_" + memberName
				value["id"] = map[string]any{"type": "Identifier", "name": methodName}
				c.classStack = append(c.classStack, name)
				fn, ok := c.function(value, false)
				c.classStack = c.classStack[:len(c.classStack)-1]
				if !ok {
					continue
				}
				fn.ReceiverType = name
				info.methods = append(info.methods, fn)
				accessor := info.accessors[memberName]
				if kind == "get" {
					accessor.get = true
				} else {
					accessor.set = true
				}
				info.accessors[memberName] = accessor
				continue
			}
			if kind != "method" {
				c.errorCodef(member, CodeProfileClass, "class member kind %q is not supported", kind)
				continue
			}
			value, ok := mapField(member, "value")
			if !ok {
				continue
			}
			value = cloneNode(value)
			value["id"] = map[string]any{"type": "Identifier", "name": memberName}
			private := stringField(member, "accessibility") == "private"
			c.classStack = append(c.classStack, name)
			fn, ok := c.function(value, private)
			c.classStack = c.classStack[:len(c.classStack)-1]
			if !ok {
				continue
			}
			if boolField(member, "static") {
				fn.Name = name + "_" + fn.Name
				info.statics = append(info.statics, fn)
				c.file.Functions = append(c.file.Functions, fn)
				continue
			}
			fn.ReceiverType = name
			info.methods = append(info.methods, fn)
		default:
			c.errorCodef(member, CodeProfileClass, "unsupported class member %q", stringField(member, "type"))
		}
	}
	return true
}

func (c *converter) classStaticField(className string, member map[string]any, info *classInfo) bool {
	name := classMemberName(member)
	if name == "" {
		c.errorf(member, "static field requires a name")
		return false
	}
	valueNode, ok := mapField(member, "value")
	if !ok || valueNode == nil {
		c.errorCodef(member, CodeProfileClass, "static field %q requires an initializer", name)
		return false
	}
	value, ok := c.expression(valueNode)
	if !ok {
		return false
	}
	private := stringField(member, "accessibility") == "private"
	constName := className + "_" + name
	c.file.Constants = append(c.file.Constants, &ast.ConstDecl{Name: constName, Private: private, Value: value})
	info.staticVals = append(info.staticVals, classStaticField{name: name, private: private})
	return true
}

func (c *converter) classBaseName(node map[string]any) (string, bool) {
	super, ok := mapField(node, "superClass")
	if !ok || super == nil {
		return "", true
	}
	if stringField(super, "type") != "Identifier" {
		c.errorCodef(node, CodeProfileClass, "only simple identifier class inheritance is supported")
		return "", false
	}
	name := stringField(super, "name")
	if _, ok := c.classNodes[name]; !ok {
		c.errorCodef(node, CodeProfileClass, "base class %q must be declared in this module", name)
		return "", false
	}
	return name, true
}

func mergeClassFields(base []classField, derived []classField) []classField {
	out := append([]classField(nil), base...)
	index := map[string]int{}
	for i, field := range out {
		index[field.name] = i
	}
	for _, field := range derived {
		if at, ok := index[field.name]; ok {
			out[at] = field
			continue
		}
		index[field.name] = len(out)
		out = append(out, field)
	}
	return out
}

func mergeClassMethods(derived []*ast.Function, base []*ast.Function, receiver string) []*ast.Function {
	out := append([]*ast.Function(nil), derived...)
	seen := map[string]bool{}
	for _, method := range out {
		seen[method.Name] = true
	}
	for _, method := range base {
		if seen[method.Name] {
			continue
		}
		out = append(out, cloneClassMethod(method, receiver))
	}
	return out
}

func cloneClassMethod(fn *ast.Function, receiver string) *ast.Function {
	copyFn := *fn
	source := astprint.File(&ast.File{Functions: []*ast.Function{&copyFn}})
	parsed, errs := parser.Parse(source)
	if len(errs) == 0 && len(parsed.Functions) == 1 {
		out := parsed.Functions[0]
		out.ReceiverType = receiver
		out.Static = fn.Static
		out.Private = fn.Private
		return out
	}
	copyFn.ReceiverType = receiver
	return &copyFn
}

func (c *converter) derivedClassFactory(info *classInfo, base *classInfo) (*ast.Function, bool) {
	params, assignments, baseArgs, ok := c.derivedConstructorParts(info, base)
	if !ok {
		return nil, false
	}
	temp := "__base_" + info.name
	statements := []ast.Stmt{&ast.LetStmt{
		Name:  temp,
		Value: &ast.CallExpr{Callee: &ast.Identifier{Name: base.name + "_new"}, Args: baseArgs},
	}}
	baseFields := map[string]bool{}
	for _, field := range base.fields {
		baseFields[field.name] = true
	}
	values := make([]ast.FieldValue, 0, len(info.fields))
	for _, field := range info.fields {
		value, ok := assignments[field.name]
		if !ok {
			value = field.init
		}
		if value == nil && baseFields[field.name] {
			value = &ast.SelectorExpr{Receiver: &ast.Identifier{Name: temp}, Name: field.name}
		}
		if value == nil {
			if field.optional {
				value = &ast.UndefinedLiteral{}
			} else {
				c.errorCodef(info.ctor, CodeProfileClass, "field %q must be initialized in the constructor or have a field initializer", field.name)
				return nil, false
			}
		}
		values = append(values, ast.FieldValue{Name: field.name, Value: value})
	}
	statements = append(statements, &ast.ReturnStmt{Value: &ast.StructLiteral{TypeName: info.name, Fields: values}})
	return &ast.Function{
		Name:       info.name + "_new",
		Params:     params,
		ReturnType: ast.NamedType(info.name),
		Body:       &ast.BlockExpr{Statements: statements},
	}, true
}

func (c *converter) derivedConstructorParts(info *classInfo, base *classInfo) ([]ast.Param, map[string]ast.Expr, []ast.Expr, bool) {
	assignments := map[string]ast.Expr{}
	baseArgs := []ast.Expr{}
	params := []ast.Param{}
	sawSuper := false
	if info.ctor != nil {
		converted, ok := c.functionParams(info.ctor)
		if !ok {
			return nil, nil, nil, false
		}
		params = converted
		body, _ := mapField(info.ctor, "body")
		for _, raw := range arrayField(body, "body") {
			stmt, ok := asObject(raw)
			if !ok {
				continue
			}
			expr, ok := mapField(stmt, "expression")
			if !ok || stringField(stmt, "type") != "ExpressionStatement" {
				c.errorCodef(stmt, CodeProfileClass, "derived constructor body is limited to super(...) and this.field = value")
				return nil, nil, nil, false
			}
			if stringField(expr, "type") == "CallExpression" {
				callee, _ := mapField(expr, "callee")
				if stringField(callee, "type") == "Super" {
					if sawSuper {
						c.errorCodef(stmt, CodeProfileClass, "super(...) may only be called once")
						return nil, nil, nil, false
					}
					sawSuper = true
					for _, argRaw := range arrayField(expr, "arguments") {
						arg, ok := asObject(argRaw)
						if !ok {
							continue
						}
						converted, ok := c.expression(arg)
						if !ok {
							return nil, nil, nil, false
						}
						baseArgs = append(baseArgs, converted)
					}
					continue
				}
			}
			fieldName, value, ok := c.constructorFieldAssignment(expr, info.fields)
			if !ok {
				return nil, nil, nil, false
			}
			assignments[fieldName] = value
		}
	}
	if !sawSuper && len(base.ctorParams) > 0 {
		c.errorCodef(info.ctor, CodeProfileClass, "derived constructor must call super(...) for base %q", base.name)
		return nil, nil, nil, false
	}
	return params, assignments, baseArgs, true
}

func (c *converter) constructorFieldAssignment(expr map[string]any, fields []classField) (string, ast.Expr, bool) {
	if stringField(expr, "type") != "AssignmentExpression" || stringField(expr, "operator") != "=" {
		c.errorCodef(expr, CodeProfileClass, "derived constructor body is limited to super(...) and this.field = value")
		return "", nil, false
	}
	left, ok := mapField(expr, "left")
	if !ok || stringField(left, "type") != "MemberExpression" {
		c.errorCodef(expr, CodeProfileClass, "constructor assignments must target `this.field`")
		return "", nil, false
	}
	object, _ := mapField(left, "object")
	if stringField(object, "type") != "ThisExpression" {
		c.errorCodef(expr, CodeProfileClass, "constructor assignments must target `this.field`")
		return "", nil, false
	}
	name := classPropertyName(left)
	known := false
	for _, field := range fields {
		if field.name == name {
			known = true
			break
		}
	}
	if !known {
		c.errorCodef(expr, CodeProfileClass, "constructor assigns unknown field %q", name)
		return "", nil, false
	}
	valueNode, ok := mapField(expr, "right")
	if !ok {
		return "", nil, false
	}
	value, ok := c.expression(valueNode)
	if !ok {
		return "", nil, false
	}
	return name, value, true
}

func classID(node map[string]any) map[string]any {
	id, _ := mapField(node, "id")
	return id
}

func (c *converter) staticFieldName(className, field string) (string, bool) {
	info, ok := c.ensureClass(className)
	if !ok {
		return "", false
	}
	for _, static := range info.staticVals {
		if static.name == field {
			return info.name + "_" + static.name, true
		}
	}
	if info.base != "" {
		return c.staticFieldName(info.base, field)
	}
	return "", false
}

func (c *converter) classAccessor(className, property string) (bool, bool) {
	info, ok := c.ensureClass(className)
	if !ok {
		return false, false
	}
	get, set := false, false
	if accessor, ok := info.accessors[property]; ok {
		get = get || accessor.get
		set = set || accessor.set
	}
	if info.base != "" {
		baseGet, baseSet := c.classAccessor(info.base, property)
		get = get || baseGet
		set = set || baseSet
	}
	return get, set
}

func (c *converter) classBase(className string) string {
	info, ok := c.ensureClass(className)
	if !ok {
		return ""
	}
	return info.base
}
