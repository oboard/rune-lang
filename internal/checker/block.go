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
	if !c.checkPrivateAccess("type", lit.TypeName, structInfo.Private, structInfo.SourcePath, lit.Pos) {
		for _, field := range lit.Fields {
			c.inferExpr(field.Value, env)
		}
		return Unknown
	}

	seen := map[string]bool{}
	typeBindings := c.structLiteralTypeBindings(structInfo)
	for _, field := range lit.Fields {
		fieldInfo, ok := structInfo.ByName[field.Name]
		if !ok {
			c.errorf(field.Pos, "type %s has no field %q", lit.TypeName, field.Name)
			c.inferExpr(field.Value, env)
			continue
		}
		if !c.checkPrivateAccess("field", lit.TypeName+"."+field.Name, fieldInfo.Private, fieldInfo.SourcePath, field.Pos) {
			c.inferExpr(field.Value, env)
			seen[field.Name] = true
			continue
		}
		if seen[field.Name] {
			c.errorf(field.Pos, "duplicate field value %q", field.Name)
		}
		seen[field.Name] = true
		valueType := c.inferExpr(field.Value, env)
		expectedType := fieldInfo.Type
		if len(typeBindings) > 0 {
			c.bindTypeParams(expectedType, valueType, typeBindings)
			expectedType = substituteTypeParams(expectedType, typeBindings)
		}
		if valueType != Unknown && expectedType != Unknown && !typesCompatible(expectedType, valueType, nil) {
			c.errorf(field.Value.Position(), "field %s.%s has type %s, expected %s", lit.TypeName, field.Name, valueType, expectedType)
		}
	}
	for _, field := range structInfo.Fields {
		if !seen[field.Name] {
			c.errorf(lit.Pos, "missing field %s.%s", lit.TypeName, field.Name)
		}
	}
	return c.structLiteralResultType(lit.TypeName, structInfo, typeBindings)
}

func (c *checker) structLiteralTypeBindings(info *StructInfo) map[string]Type {
	if info == nil || len(info.Generics) == 0 {
		return nil
	}
	bindings := make(map[string]Type, len(info.Generics))
	for _, name := range info.Generics {
		bindings[name] = Unknown
	}
	return bindings
}

func (c *checker) structLiteralResultType(name string, info *StructInfo, bindings map[string]Type) Type {
	if info == nil || len(info.Generics) == 0 {
		return Type(name)
	}
	args := make([]Type, 0, len(info.Generics))
	for _, param := range info.Generics {
		typ := bindings[param]
		if typ == "" || typ == Unknown {
			return Type(name)
		}
		args = append(args, typ)
	}
	return genericTypeOf(name, args)
}

func (c *checker) bindTypeParams(expected Type, actual Type, bindings map[string]Type) {
	if len(bindings) == 0 || expected == "" || actual == Unknown {
		return
	}
	if _, ok := bindings[string(expected)]; ok {
		if bindings[string(expected)] == Unknown {
			bindings[string(expected)] = actual
			return
		}
		if unified, ok := c.unifyTypes(bindings[string(expected)], actual); ok {
			bindings[string(expected)] = unified
		}
		return
	}
	if elem, ok := parseArrayType(string(expected)); ok {
		if actualElem, ok := ArrayElement(actual); ok {
			c.bindTypeParams(Type(elem), actualElem, bindings)
		}
		return
	}
	if params, ret, ok := parseFuncType(string(expected)); ok {
		actualParams, actualRet, actualFunc := parseFuncType(string(actual))
		if !actualFunc || len(actualParams) != len(params) {
			return
		}
		for i, param := range params {
			c.bindTypeParams(Type(param), Type(actualParams[i]), bindings)
		}
		c.bindTypeParams(Type(ret), Type(actualRet), bindings)
		return
	}
	if base, args, ok := parseGenericType(string(expected)); ok {
		actualBase, actualArgs, actualGeneric := parseGenericType(string(actual))
		if !actualGeneric || actualBase != base || len(actualArgs) != len(args) {
			return
		}
		for i, arg := range args {
			c.bindTypeParams(Type(arg), Type(actualArgs[i]), bindings)
		}
		return
	}
	if inner, ok := parseNullableType(string(expected)); ok {
		if actualInner, actualNullable := parseNullableType(string(actual)); actualNullable {
			c.bindTypeParams(Type(inner), Type(actualInner), bindings)
		} else {
			c.bindTypeParams(Type(inner), actual, bindings)
		}
	}
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
			fieldInfo := FieldInfo{Name: field.Name, Private: field.Private, Type: Unknown}
			fields = append(fields, fieldInfo)
			byName[field.Name] = fieldInfo
			continue
		}
		valueType := c.inferExpr(field.Value, env)
		fieldInfo := FieldInfo{Name: field.Name, Private: field.Private, Type: valueType}
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
			fieldInfo := FieldInfo{Name: field.Name, Private: field.Private, Type: valueType}
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
		branchEnv := cloneEnv(env)
		c.checkPatternForSubject(branch.Pattern, subject, branchEnv)
		typ := c.inferExpr(branch.Expr, branchEnv)
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
		params, ret, ok := parseFuncType(string(typ))
		if ok {
			c.applyLambdaParamTypes(e, params)
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
	case *ast.TernaryExpr:
		c.applyExpectedType(e.Consequence, typ)
		c.applyExpectedType(e.Alternative, typ)
	}
}

func (c *checker) applyLambdaParamTypes(lambda *ast.LambdaExpr, params []string) {
	types := map[string]Type{}
	for i, name := range lambda.Params {
		if i < len(params) {
			types[name] = Type(params[i])
		}
	}
	if len(types) == 0 {
		return
	}
	ast.WalkExpr(lambda.Body, func(expr ast.Expr) {
		ident, ok := expr.(*ast.Identifier)
		if !ok {
			return
		}
		if typ, ok := types[ident.Name]; ok {
			c.info.ExprTypes[ident] = typ
		}
	})
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
		case *ast.DoubleLiteral:
			return Double
		case *ast.BigIntLiteral:
			return BigInt
		case *ast.StringLiteral:
			return String
		case *ast.NullLiteral:
			return Null
		case *ast.SelectorExpr:
			if typ, ok := c.enumMemberType(p.Value); ok {
				return typ
			}
		}
	}
	return Unknown
}

func (c *checker) checkPatternForSubject(pattern ast.Pattern, subject Type, env map[string]Type) {
	if constructor, ok := pattern.(*ast.ConstructorPattern); ok {
		okType, errType, result := parseResultType(subject)
		if !result {
			if subject != Unknown {
				c.errorf(constructor.Pos, "constructor pattern %s expects Result, got %s", constructor.Name, subject)
			}
			return
		}
		var bindingType Type
		switch constructor.Name {
		case "Ok":
			bindingType = okType
		case "Err":
			bindingType = errType
		default:
			c.errorf(constructor.Pos, "unknown result constructor %q", constructor.Name)
			return
		}
		if constructor.Binding != "" {
			env[constructor.Binding] = bindingType
		}
		return
	}
	c.checkPattern(pattern, env)
}

func (c *checker) enumMemberType(expr ast.Expr) (Type, bool) {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return Unknown, false
	}
	ident, ok := sel.Receiver.(*ast.Identifier)
	if !ok {
		return Unknown, false
	}
	enum := c.info.Enums[ident.Name]
	if enum == nil {
		return Unknown, false
	}
	if !c.checkPrivateAccess("enum", enum.Name, enum.Private, enum.SourcePath, ident.Pos) {
		return Unknown, false
	}
	member, ok := enum.ByName[sel.Name]
	if !ok {
		return Unknown, false
	}
	if !c.checkPrivateAccess("enum member", enum.Name+"."+sel.Name, member.Private, member.SourcePath, sel.NamePos) {
		return Unknown, false
	}
	return Type(enum.Name), true
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
	leftParams, leftRet, leftOK := parseCallableType(string(left))
	rightParams, rightRet, rightOK := parseCallableType(string(right))
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
	if _, _, leftAsync := parseAsyncFuncType(string(left)); leftAsync {
		return AsyncFuncOfTypes(params, ret), true
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
		if !orderedComparisonType(typ) && typ != Unknown {
			c.errorf(p.Pos, "comparison pattern expects Int, Double, BigInt, or String literal")
		}
	case *ast.TuplePattern:
		for _, elem := range p.Elements {
			c.checkPattern(elem, env)
		}
	case *ast.ConstructorPattern:
		c.errorf(p.Pos, "constructor pattern requires a Result subject")
	}
}
