package tsfront

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

// classField is the subset of a TypeScript property declaration the Rune
// struct model can represent.
type classField struct {
	name     string
	typ      ast.Type
	init     ast.Expr
	private  bool
	optional bool
}

// classLegacy is the pre-inheritance class lowering. The active path is in
// class_inherit.go; this copy remains as a reference for the base-less form.
// The supported subset is deliberately explicit: fields, a constructor made of
// `this.field = value` assignments, instance methods, and static methods.
// Inheritance, accessors, static fields and arbitrary constructor statements
// are rejected with RUNE-PROFILE-002 instead of being approximated.
func (c *converter) classLegacy(node map[string]any, file *ast.File) {
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
	if super, ok := mapField(node, "superClass"); ok && super != nil {
		c.errorCodef(node, CodeProfileClass, "class inheritance is not supported yet")
		return
	}
	body, ok := mapField(node, "body")
	if !ok {
		c.errorf(node, "class body is required")
		return
	}
	var fields []classField
	var methods []*ast.Function
	var constructor map[string]any
	for _, raw := range arrayField(body, "body") {
		member, ok := asObject(raw)
		if !ok {
			continue
		}
		switch stringField(member, "type") {
		case "PropertyDefinition", "ClassProperty":
			field, ok := c.classProperty(member)
			if !ok {
				continue
			}
			fields = append(fields, field)
		case "MethodDefinition", "TSAbstractMethodDefinition":
			kind := stringField(member, "kind")
			memberName := classMemberName(member)
			if kind == "constructor" {
				constructor, _ = mapField(member, "value")
				continue
			}
			if kind != "method" {
				c.errorCodef(member, CodeProfileClass, "class accessors and %q members are not supported", kind)
				continue
			}
			value, ok := mapField(member, "value")
			if !ok {
				continue
			}
			value = cloneNode(value)
			value["id"] = map[string]any{"type": "Identifier", "name": memberName}
			private := stringField(member, "accessibility") == "private"
			fn, ok := c.function(value, private)
			if !ok {
				continue
			}
			if boolField(member, "static") {
				// Static methods become free functions (`Class_method`). This
				// keeps generated Rune readable by the self-hosted interpreter,
				// which does not resolve `Type::method` yet.
				fn.Name = name + "_" + fn.Name
				file.Functions = append(file.Functions, fn)
				continue
			}
			fn.ReceiverType = name
			methods = append(methods, fn)
		default:
			c.errorCodef(member, CodeProfileClass, "unsupported class member %q", stringField(member, "type"))
		}
	}
	file.Types = append(file.Types, &ast.StructType{
		Name:    name,
		Fields:  buildClassFields(fields),
		Methods: methods,
		Pos:     lexerPosition(node),
		NamePos: lexerPosition(id),
	})
	if factory, ok := c.classConstructor(name, constructor, fields); ok {
		file.Functions = append(file.Functions, factory)
	}
}

func (c *converter) classProperty(member map[string]any) (classField, bool) {
	name := classMemberName(member)
	if name == "" {
		c.errorf(member, "class property requires a name")
		return classField{}, false
	}
	if boolField(member, "static") {
		c.errorCodef(member, CodeProfileClass, "static fields are not supported")
		return classField{}, false
	}
	typ := ast.NamedType("Dynamic")
	if annotation, ok := mapField(member, "typeAnnotation"); ok {
		converted, ok := c.typeAnnotation(annotation)
		if !ok {
			return classField{}, false
		}
		typ = converted
	} else if _, hasInit := mapField(member, "value"); !hasInit {
		c.errorCodef(member, CodeDynamicImplicitAny, "class field %q requires an explicit type", name)
		return classField{}, false
	}
	field := classField{
		name:     name,
		typ:      typ,
		private:  stringField(member, "accessibility") == "private",
		optional: boolField(member, "optional"),
	}
	if value, ok := mapField(member, "value"); ok && value != nil {
		init, ok := c.expression(value)
		if !ok {
			return classField{}, false
		}
		field.init = init
	}
	return field, true
}

func buildClassFields(fields []classField) []ast.Field {
	out := make([]ast.Field, 0, len(fields))
	for _, field := range fields {
		out = append(out, ast.Field{Name: field.name, Private: field.private, Type: field.typ})
	}
	return out
}

func (c *converter) classConstructor(name string, constructor map[string]any, fields []classField) (*ast.Function, bool) {
	params := []ast.Param{}
	assignments := map[string]ast.Expr{}
	if constructor != nil {
		converted, ok := c.functionParams(constructor)
		if !ok {
			return nil, false
		}
		params = converted
		parsed, ok := c.constructorAssignments(constructor, fields)
		if !ok {
			return nil, false
		}
		assignments = parsed
	}
	values := make([]ast.FieldValue, 0, len(fields))
	for _, field := range fields {
		value, ok := assignments[field.name]
		if !ok {
			value = field.init
		}
		if value == nil {
			if field.optional {
				value = &ast.UndefinedLiteral{}
			} else {
				c.errorCodef(constructor, CodeProfileClass, "field %q must be initialized in the constructor or have a field initializer", field.name)
				return nil, false
			}
		}
		values = append(values, ast.FieldValue{Name: field.name, Value: value})
	}
	return &ast.Function{
		Name:       name + "_new",
		Params:     params,
		ReturnType: ast.NamedType(name),
		Body: &ast.BlockExpr{Statements: []ast.Stmt{&ast.ReturnStmt{
			Value: &ast.StructLiteral{TypeName: name, Fields: values},
		}}},
	}, true
}

func (c *converter) functionParams(node map[string]any) ([]ast.Param, bool) {
	params := make([]ast.Param, 0)
	for _, raw := range arrayField(node, "params") {
		param, ok := asObject(raw)
		if !ok || stringField(param, "type") != "Identifier" {
			c.errorf(node, "only identifier parameters are supported")
			return nil, false
		}
		paramType, _, ok := c.typedParam(param)
		if !ok {
			return nil, false
		}
		params = append(params, ast.Param{Name: stringField(param, "name"), Type: paramType})
	}
	return params, true
}

func (c *converter) constructorAssignments(constructor map[string]any, fields []classField) (map[string]ast.Expr, bool) {
	known := map[string]bool{}
	for _, field := range fields {
		known[field.name] = true
	}
	assignments := map[string]ast.Expr{}
	body, ok := mapField(constructor, "body")
	if !ok {
		return assignments, true
	}
	for _, raw := range arrayField(body, "body") {
		stmt, ok := asObject(raw)
		if !ok {
			continue
		}
		if stringField(stmt, "type") != "ExpressionStatement" {
			c.errorCodef(stmt, CodeProfileClass, "constructor body is limited to `this.field = value` assignments")
			return nil, false
		}
		expr, ok := mapField(stmt, "expression")
		if !ok || stringField(expr, "type") != "AssignmentExpression" || stringField(expr, "operator") != "=" {
			c.errorCodef(stmt, CodeProfileClass, "constructor body is limited to `this.field = value` assignments")
			return nil, false
		}
		left, ok := mapField(expr, "left")
		if !ok || stringField(left, "type") != "MemberExpression" {
			c.errorCodef(stmt, CodeProfileClass, "constructor assignments must target `this.field`")
			return nil, false
		}
		object, _ := mapField(left, "object")
		if stringField(object, "type") != "ThisExpression" {
			c.errorCodef(stmt, CodeProfileClass, "constructor assignments must target `this.field`")
			return nil, false
		}
		fieldName := classPropertyName(left)
		if !known[fieldName] {
			c.errorCodef(stmt, CodeProfileClass, "constructor assigns unknown field %q", fieldName)
			return nil, false
		}
		valueNode, ok := mapField(expr, "right")
		if !ok {
			return nil, false
		}
		value, ok := c.expression(valueNode)
		if !ok {
			return nil, false
		}
		assignments[fieldName] = value
	}
	return assignments, true
}

func classMemberName(member map[string]any) string {
	key, ok := mapField(member, "key")
	if !ok {
		return ""
	}
	if name := stringField(key, "name"); name != "" {
		return name
	}
	return stringField(key, "value")
}

func classPropertyName(member map[string]any) string {
	property, ok := mapField(member, "property")
	if !ok {
		return ""
	}
	if name := stringField(property, "name"); name != "" {
		return name
	}
	return stringField(property, "value")
}

func cloneNode(node map[string]any) map[string]any {
	out := make(map[string]any, len(node))
	for key, value := range node {
		out[key] = value
	}
	return out
}

func boolField(node map[string]any, key string) bool {
	value, _ := node[key].(bool)
	return value
}

func lexerPosition(node map[string]any) lexer.Position {
	line, column := nodePosition(node)
	return lexer.Position{Line: line, Column: column + 1}
}
