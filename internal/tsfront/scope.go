package tsfront

import "github.com/oboard/rune-lang/internal/ast"

import "strings"

func collectTopLevelDeclarations(statements []map[string]any, out map[string]bool) {
	for _, stmt := range statements {
		collectStatementDeclarations(stmt, out)
	}
}

func collectStatementDeclarations(node map[string]any, out map[string]bool) {
	if node == nil {
		return
	}
	if stringField(node, "type") == "ExportNamedDeclaration" {
		if decl, ok := mapField(node, "declaration"); ok {
			collectStatementDeclarations(decl, out)
		}
		return
	}
	switch stringField(node, "type") {
	case "FunctionDeclaration", "ClassDeclaration", "TSInterfaceDeclaration", "TSTypeAliasDeclaration", "TSEnumDeclaration":
		if id, ok := mapField(node, "id"); ok {
			if name := stringField(id, "name"); name != "" {
				out[name] = true
			}
		}
	case "VariableDeclaration":
		for _, raw := range arrayField(node, "declarations") {
			decl, ok := asObject(raw)
			if !ok {
				continue
			}
			if id, ok := mapField(decl, "id"); ok {
				if name := stringField(id, "name"); name != "" {
					out[name] = true
				}
			}
		}
	case "ImportDeclaration":
		for _, raw := range arrayField(node, "specifiers") {
			spec, ok := asObject(raw)
			if !ok {
				continue
			}
			if local, ok := mapField(spec, "local"); ok {
				if name := stringField(local, "name"); name != "" {
					out[name] = true
				}
			}
		}
	}
}

func (c *converter) pushScope(params []astParamInfo) func() {
	declared := map[string]bool{}
	anyVars := map[string]bool{}
	vars := map[string]string{}
	for _, param := range params {
		declared[param.name] = true
		if param.any {
			anyVars[param.name] = true
		}
		if param.class != "" {
			vars[param.name] = param.class
		}
	}
	c.scopes = append(c.scopes, declared)
	c.anyScopes = append(c.anyScopes, anyVars)
	c.varTypes = append(c.varTypes, vars)
	return func() {
		c.scopes = c.scopes[:len(c.scopes)-1]
		c.anyScopes = c.anyScopes[:len(c.anyScopes)-1]
		c.varTypes = c.varTypes[:len(c.varTypes)-1]
	}
}

type astParamInfo struct {
	name  string
	any   bool
	class string
}

func (c *converter) declare(name string) {
	if name == "" {
		return
	}
	if len(c.scopes) > 0 {
		c.scopes[len(c.scopes)-1][name] = true
		return
	}
	c.declared[name] = true
}

func (c *converter) markAny(name string) {
	if name == "" {
		return
	}
	if len(c.anyScopes) == 0 {
		c.globalAny[name] = true
		return
	}
	c.anyScopes[len(c.anyScopes)-1][name] = true
}

func (c *converter) isDeclared(name string) bool {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if c.scopes[i][name] {
			return true
		}
	}
	return c.declared[name]
}

func (c *converter) isAny(name string) bool {
	for i := len(c.anyScopes) - 1; i >= 0; i-- {
		if c.anyScopes[i][name] {
			return true
		}
	}
	return c.globalAny[name]
}

func (c *converter) declareClassVar(name, className string) {
	if name == "" || className == "" || len(c.varTypes) == 0 {
		return
	}
	c.varTypes[len(c.varTypes)-1][name] = className
}

func (c *converter) varClass(name string) string {
	for i := len(c.varTypes) - 1; i >= 0; i-- {
		if className := c.varTypes[i][name]; className != "" {
			return className
		}
	}
	return ""
}

func (c *converter) classNameOf(typ ast.Type) string {
	if typ.Kind != ast.TypeName {
		return ""
	}
	if c.classes[typ.Name] {
		return typ.Name
	}
	return ""
}

func (c *converter) currentClass() string {
	if len(c.classStack) == 0 {
		return ""
	}
	return c.classStack[len(c.classStack)-1]
}

func (c *converter) classOfValue(expr ast.Expr) string {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return ""
	}
	ident, ok := call.Callee.(*ast.Identifier)
	if !ok || !strings.HasSuffix(ident.Name, "_new") {
		return ""
	}
	base := strings.TrimSuffix(ident.Name, "_new")
	if c.classes[base] {
		return base
	}
	return ""
}

func isDynamicType(typ ast.Type) bool {
	return typ.Canonical() == "Dynamic"
}

func (c *converter) typedParam(param map[string]any) (ast.Type, bool, bool) {
	name := stringField(param, "name")
	annotation, ok := mapField(param, "typeAnnotation")
	if !ok {
		c.errorCodef(param, CodeDynamicImplicitAny, "parameter %q requires an explicit type", name)
		return ast.Type{}, false, false
	}
	converted, ok := c.typeAnnotation(annotation)
	if !ok {
		return ast.Type{}, false, false
	}
	return converted, isDynamicType(converted), true
}
