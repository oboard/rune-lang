package checker

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
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
			c.bindings[s.Name] = s.Value
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

func (c *checker) inferMatchExpr(match *ast.MatchExpr, env map[string]Type) Type {
	subject := c.inferExpr(match.Subject, env)
	patternSubject := c.inferPatternSubject(match.Branches)
	if subject == Unknown && patternSubject != Unknown {
		if ident, ok := match.Subject.(*ast.Identifier); ok {
			env[ident.Name] = patternSubject
			subject = patternSubject
			c.info.ExprTypes[ident] = patternSubject
		}
	}
	if subject != Unknown && patternSubject != Unknown && subject != patternSubject {
		c.errorf(match.Subject.Position(), "match subject has type %s, expected %s", subject, patternSubject)
	}
	result := Unknown
	var branchExprs []ast.Expr
	for _, branch := range match.Branches {
		c.checkPattern(branch.Pattern, env)
		typ := c.inferExpr(branch.Expr, env)
		branchExprs = append(branchExprs, branch.Expr)
		if result == Unknown {
			result = typ
			continue
		}
		result = c.mergeBranchTypes(result, typ, branch.Expr.Position())
	}
	for _, expr := range branchExprs {
		c.applyExpectedType(expr, result)
	}
	return result
}

func (c *checker) applyExpectedType(expr ast.Expr, typ Type) {
	if expr == nil || typ == Unknown {
		return
	}
	c.info.ExprTypes[expr] = typ
	switch e := expr.(type) {
	case *ast.LambdaExpr:
		_, ret, ok := parseFuncType(string(typ))
		if ok {
			c.applyExpectedType(e.Body, Type(ret))
		}
	case *ast.AnonymousObjectLiteral:
		for _, field := range e.Fields {
			if info := c.info.Types[baseTypeName(typ)]; info != nil {
				if fieldInfo, ok := info.ByName[field.Name]; ok {
					c.applyExpectedType(field.Value, fieldInfo.Type)
				}
			}
		}
	case *ast.MatchExpr:
		for _, branch := range e.Branches {
			c.applyExpectedType(branch.Expr, typ)
		}
	}
}

func (c *checker) inferPatternSubject(branches []ast.PatternBranch) Type {
	result := Unknown
	for _, branch := range branches {
		typ := c.patternLiteralType(branch.Pattern)
		if typ == Unknown {
			continue
		}
		if result == Unknown {
			result = typ
			continue
		}
		if result != typ {
			return Unknown
		}
	}
	return result
}

func (c *checker) patternLiteralType(pattern ast.Pattern) Type {
	switch p := pattern.(type) {
	case *ast.LiteralPattern:
		switch p.Value.(type) {
		case *ast.BoolLiteral:
			return Bool
		case *ast.IntegerLiteral:
			return Int
		case *ast.StringLiteral:
			return String
		}
	}
	return Unknown
}

func (c *checker) mergeBranchTypes(left Type, right Type, pos lexer.Position) Type {
	if typ, ok := c.mergeFunctionValueTypes(left, right); ok {
		return typ
	}
	typ, ok := c.unifyTypes(left, right)
	if !ok {
		c.errorf(pos, "match branch returns %s, expected %s", right, left)
		return left
	}
	return typ
}

func (c *checker) mergeFunctionValueTypes(left Type, right Type) (Type, bool) {
	leftParams, leftRet, leftOK := parseFuncType(string(left))
	rightParams, rightRet, rightOK := parseFuncType(string(right))
	if !leftOK || !rightOK || len(leftParams) != len(rightParams) {
		return Unknown, false
	}
	params := make([]Type, 0, len(leftParams))
	for i := range leftParams {
		param, ok := c.mergeParameterRequirement(Type(leftParams[i]), Type(rightParams[i]))
		if !ok {
			return Unknown, false
		}
		params = append(params, param)
	}
	ret, ok := c.unifyTypes(Type(leftRet), Type(rightRet))
	if !ok {
		return Unknown, false
	}
	return FuncOfTypes(params, ret), true
}

func (c *checker) mergeParameterRequirement(left Type, right Type) (Type, bool) {
	if left == Unknown {
		return right, true
	}
	if right == Unknown || left == right {
		return left, true
	}
	if isObjectType(left) && isObjectType(right) {
		return c.unionObjectTypes(left, right)
	}
	return c.unifyTypes(left, right)
}

func (c *checker) unionObjectTypes(left Type, right Type) (Type, bool) {
	leftInfo := c.info.Types[baseTypeName(left)]
	rightInfo := c.info.Types[baseTypeName(right)]
	if leftInfo == nil || rightInfo == nil {
		return Unknown, false
	}
	fields := append([]FieldInfo(nil), leftInfo.Fields...)
	seen := map[string]bool{}
	for _, field := range fields {
		seen[field.Name] = true
	}
	for _, field := range rightInfo.Fields {
		if seen[field.Name] {
			continue
		}
		fields = append(fields, field)
	}
	typ := ObjectOf(fields)
	byName := map[string]FieldInfo{}
	for _, field := range fields {
		byName[field.Name] = field
	}
	c.registerAnonymousObjectType(typ, fields, byName)
	return typ, true
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
