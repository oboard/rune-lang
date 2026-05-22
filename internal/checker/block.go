package checker

import (
	"github.com/oboard/rune-lang/internal/ast"
)

func (c *checker) inferStructLiteral(lit *ast.StructLiteral, env map[string]Type) Type {
	structInfo := c.info.Types[lit.TypeName]
	if structInfo == nil {
		c.errorf(lit.Pos, "unknown type %q", lit.TypeName)
		for _, field := range lit.Fields {
			c.inferExpr(field.Value, env)
		}
		return Unknown
	}

	seen := map[string]bool{}
	for _, field := range lit.Fields {
		fieldInfo, ok := structInfo.ByName[field.Name]
		if !ok {
			c.errorf(field.Pos, "type %s has no field %q", lit.TypeName, field.Name)
			c.inferExpr(field.Value, env)
			continue
		}
		if seen[field.Name] {
			c.errorf(field.Pos, "duplicate field value %q", field.Name)
		}
		seen[field.Name] = true
		valueType := c.inferExpr(field.Value, env)
		if valueType != Unknown && fieldInfo.Type != Unknown && valueType != fieldInfo.Type {
			c.errorf(field.Value.Position(), "field %s.%s has type %s, expected %s", lit.TypeName, field.Name, valueType, fieldInfo.Type)
		}
	}
	for _, field := range structInfo.Fields {
		if !seen[field.Name] {
			c.errorf(lit.Pos, "missing field %s.%s", lit.TypeName, field.Name)
		}
	}
	return Type(lit.TypeName)
}

func (c *checker) inferAnonymousObjectLiteral(lit *ast.AnonymousObjectLiteral, env map[string]Type) Type {
	return c.inferAnonymousObjectLiteralWithSelf(lit, env, "")
}

func (c *checker) inferAnonymousObjectLiteralWithSelf(lit *ast.AnonymousObjectLiteral, env map[string]Type, selfName string) Type {
	var fields []FieldInfo
	byName := map[string]FieldInfo{}
	for _, field := range lit.Fields {
		if _, exists := byName[field.Name]; exists {
			c.errorf(field.Pos, "duplicate field value %q", field.Name)
		}
		if _, isLambda := field.Value.(*ast.LambdaExpr); isLambda && selfName != "" {
			fieldInfo := FieldInfo{Name: field.Name, Type: Unknown}
			fields = append(fields, fieldInfo)
			byName[field.Name] = fieldInfo
			continue
		}
		valueType := c.inferExpr(field.Value, env)
		fieldInfo := FieldInfo{Name: field.Name, Type: valueType}
		fields = append(fields, fieldInfo)
		byName[field.Name] = fieldInfo
	}
	typ := ObjectOf(fields)
	if selfName != "" {
		c.registerAnonymousObjectType(typ, fields, byName)
		selfEnv := cloneEnv(env)
		selfEnv[selfName] = typ
		selfEnv["this"] = typ
		fields = fields[:0]
		byName = map[string]FieldInfo{}
		for _, field := range lit.Fields {
			valueType := c.inferExpr(field.Value, selfEnv)
			fieldInfo := FieldInfo{Name: field.Name, Type: valueType}
			fields = append(fields, fieldInfo)
			byName[field.Name] = fieldInfo
		}
		typ = ObjectOf(fields)
	}
	c.registerAnonymousObjectType(typ, fields, byName)
	return typ
}

func (c *checker) registerAnonymousObjectType(typ Type, fields []FieldInfo, byName map[string]FieldInfo) {
	c.info.Types[string(typ)] = &StructInfo{
		Name:   string(typ),
		Fields: append([]FieldInfo(nil), fields...),
		ByName: byName,
	}
}

func (c *checker) inferBlock(block *ast.BlockExpr, env map[string]Type) Type {
	local := cloneEnv(env)
	result := Void
	for _, stmt := range block.Statements {
		switch s := stmt.(type) {
		case *ast.LetStmt:
			if _, exists := local[s.Name]; exists {
				c.errorf(s.Pos, "name %q is already defined", s.Name)
			}
			if lit, ok := s.Value.(*ast.AnonymousObjectLiteral); ok {
				typ := c.inferAnonymousObjectLiteralWithSelf(lit, local, s.Name)
				c.info.ExprTypes[lit] = typ
				local[s.Name] = typ
			} else {
				local[s.Name] = c.inferExpr(s.Value, local)
			}
			result = Void
		case *ast.AssignStmt:
			if _, exists := local[s.Name]; !exists {
				c.errorf(s.Pos, "cannot assign undefined name %q", s.Name)
			}
			c.inferExpr(s.Value, local)
			result = Void
		case *ast.ExprStmt:
			result = c.inferExpr(s.Expr, local)
		}
	}
	return result
}

func (c *checker) inferPatternBlock(block *ast.PatternBlock, env map[string]Type) Type {
	result := Unknown
	for _, branch := range block.Branches {
		c.checkPattern(branch.Pattern, env)
		typ := c.inferExpr(branch.Expr, env)
		if result == Unknown {
			result = typ
			continue
		}
		if typ != Unknown && result != typ {
			c.errorf(branch.Expr.Position(), "pattern branch returns %s, expected %s", typ, result)
		}
	}
	return result
}

func (c *checker) checkPattern(pattern ast.Pattern, env map[string]Type) {
	switch p := pattern.(type) {
	case *ast.LiteralPattern:
		c.inferExpr(p.Value, env)
	case *ast.ComparePattern:
		typ := c.inferExpr(p.Value, env)
		if typ != Int && typ != Unknown {
			c.errorf(p.Pos, "comparison pattern expects Int literal")
		}
	case *ast.TuplePattern:
		for _, elem := range p.Elements {
			c.checkPattern(elem, env)
		}
	}
}
