package checker

import (
	"fmt"

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
	c.checkStructLiteralCommas(lit)
	c.checkExplicitStructLiteralDuplicateFields(lit)
	fields := c.expandStructLiteralFields(lit, structInfo, typeBindings, env)
	for _, field := range fields {
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
		seen[field.Name] = true
		valueType := c.inferExpr(field.Value, env)
		expectedType := fieldInfo.Type
		if len(typeBindings) > 0 {
			c.bindTypeParams(expectedType, valueType, typeBindings)
			expectedType = substituteTypeParams(expectedType, typeBindings)
		}
		if expectedType != Unknown && shouldApplyExpectedType(valueType, expectedType) {
			c.applyExpectedType(field.Value, expectedType)
		}
		if valueType != Unknown && expectedType != Unknown && !c.typesCompatible(expectedType, valueType, nil) {
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

func (c *checker) checkStructLiteralCommas(lit *ast.StructLiteral) {
	for _, field := range lit.Fields {
		if field.MissingComma {
			c.errorf(field.Pos, "expected ',' between struct literal fields")
		}
	}
}

func (c *checker) checkExplicitStructLiteralDuplicateFields(lit *ast.StructLiteral) {
	seen := map[string]bool{}
	for _, field := range lit.Fields {
		if field.Spread || field.Name == "" {
			continue
		}
		if seen[field.Name] {
			c.errorf(field.Pos, "duplicate field value %q", field.Name)
			continue
		}
		seen[field.Name] = true
	}
}

func (c *checker) expandStructLiteralFields(lit *ast.StructLiteral, structInfo *StructInfo, typeBindings map[string]Type, env map[string]Type) []ast.FieldValue {
	if len(lit.Fields) == 0 {
		return lit.Fields
	}
	var expanded []ast.FieldValue
	hasSpread := false
	for _, field := range lit.Fields {
		if !field.Spread {
			expanded = append(expanded, field)
			continue
		}
		hasSpread = true
		spreadType := c.inferExpr(field.Value, env)
		expectedType := c.structLiteralResultType(lit.TypeName, structInfo, typeBindings)
		if spreadType != Unknown && expectedType != Unknown && !c.typesCompatible(expectedType, spreadType, nil) {
			c.errorf(field.Pos, "spread expects %s, got %s", expectedType, spreadType)
			continue
		}
		for _, spreadField := range structInfo.Fields {
			expanded = append(expanded, ast.FieldValue{
				Name:    spreadField.Name,
				Private: spreadField.Private,
				Value: &ast.SelectorExpr{
					Receiver: field.Value,
					Name:     spreadField.Name,
					Pos:      field.Value.Position(),
					NamePos:  field.Pos,
				},
				Pos: field.Pos,
			})
		}
	}
	if !hasSpread {
		return lit.Fields
	}
	return dedupeStructLiteralFields(expanded)
}

func dedupeStructLiteralFields(fields []ast.FieldValue) []ast.FieldValue {
	keep := make([]bool, len(fields))
	seen := map[string]bool{}
	for idx := len(fields) - 1; idx >= 0; idx-- {
		name := fields[idx].Name
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		keep[idx] = true
	}
	out := make([]ast.FieldValue, 0, len(seen))
	for idx, field := range fields {
		if keep[idx] {
			out = append(out, field)
		}
	}
	return out
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
	if selfName == "" && c.expectedType != Unknown {
		if typ, ok := c.inferAnonymousObjectLiteralAsStruct(lit, c.expectedType, env); ok {
			return typ
		}
	}
	var fields []FieldInfo
	byName := map[string]FieldInfo{}
	for _, field := range lit.Fields {
		if field.MissingComma {
			c.errorf(field.Pos, "expected ',' between object literal fields")
		}
		if _, exists := byName[field.Name]; exists {
			c.errorf(field.Pos, "duplicate field value %q", field.Name)
		}
		if lambda, isLambda := field.Value.(*ast.LambdaExpr); isLambda && selfName != "" {
			fieldInfo := FieldInfo{Name: field.Name, Private: field.Private, Type: c.lambdaSignaturePlaceholder(lambda)}
			c.info.ExprTypes[lambda] = fieldInfo.Type
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

func (c *checker) inferAnonymousObjectLiteralAsStruct(lit *ast.AnonymousObjectLiteral, expected Type, env map[string]Type) (Type, bool) {
	if expected == Unknown || isObjectType(expected) {
		return Unknown, false
	}
	typeName := baseTypeName(expected)
	structInfo := c.info.Types[typeName]
	if structInfo == nil {
		return Unknown, false
	}
	if !c.checkPrivateAccess("type", typeName, structInfo.Private, structInfo.SourcePath, lit.Pos) {
		for _, field := range lit.Fields {
			c.inferExpr(field.Value, env)
		}
		return Unknown, true
	}

	c.checkAnonymousObjectLiteralCommas(lit)
	c.checkAnonymousObjectLiteralDuplicateFields(lit)
	fields := c.expandStructLiteralFields(&ast.StructLiteral{
		TypeName: typeName,
		Fields:   lit.Fields,
		Pos:      lit.Pos,
	}, structInfo, typeParamBindingsForStruct(structInfo, expected), env)
	seen := map[string]bool{}
	for _, field := range fields {
		fieldInfo, ok := structInfo.ByName[field.Name]
		if !ok {
			c.errorf(field.Pos, "type %s has no field %q", typeName, field.Name)
			c.inferExpr(field.Value, env)
			continue
		}
		if !c.checkPrivateAccess("field", typeName+"."+field.Name, fieldInfo.Private, fieldInfo.SourcePath, field.Pos) {
			c.inferExpr(field.Value, env)
			seen[field.Name] = true
			continue
		}
		seen[field.Name] = true
		expectedType := structFieldType(structInfo, expected, fieldInfo)
		valueType := c.inferExprExpected(field.Value, env, expectedType)
		if valueType != Unknown && expectedType != Unknown && !c.typesCompatible(expectedType, valueType, nil) {
			c.errorf(field.Value.Position(), "field %s.%s has type %s, expected %s", typeName, field.Name, valueType, expectedType)
		}
	}
	for _, field := range structInfo.Fields {
		if !seen[field.Name] {
			c.errorf(lit.Pos, "missing field %s.%s", typeName, field.Name)
		}
	}
	c.info.ExprTypes[lit] = expected
	return expected, true
}

func (c *checker) checkAnonymousObjectLiteralCommas(lit *ast.AnonymousObjectLiteral) {
	for _, field := range lit.Fields {
		if field.MissingComma {
			c.errorf(field.Pos, "expected ',' between object literal fields")
		}
	}
}

func (c *checker) checkAnonymousObjectLiteralDuplicateFields(lit *ast.AnonymousObjectLiteral) {
	seen := map[string]bool{}
	for _, field := range lit.Fields {
		if field.Spread || field.Name == "" {
			continue
		}
		if seen[field.Name] {
			c.errorf(field.Pos, "duplicate field value %q", field.Name)
			continue
		}
		seen[field.Name] = true
	}
}

func (c *checker) lambdaSignaturePlaceholder(lambda *ast.LambdaExpr) Type {
	params := make([]Type, 0, len(lambda.Params))
	for i := range lambda.Params {
		paramType := Unknown
		if i < len(lambda.ParamTypes) && !lambda.ParamTypes[i].IsZero() {
			paramType = c.resolveTypeWithGenerics(lambda.ParamTypes[i].Canonical(), c.genericTypes)
		}
		params = append(params, paramType)
	}
	ret := Unknown
	if !lambda.ReturnType.IsZero() {
		ret = c.resolveTypeWithGenerics(lambda.ReturnType.Canonical(), c.genericTypes)
	}
	return FuncOfTypes(params, ret)
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
	blockExpected := c.expectedType
	c.expectedType = Unknown
	defer func() {
		c.expectedType = blockExpected
	}()
	for index, stmt := range block.Statements {
		isLast := index == len(block.Statements)-1
		switch s := stmt.(type) {
		case *ast.LetStmt:
			if _, exists := local[s.Name]; exists {
				c.errorf(s.Pos, "name %q is already defined", s.Name)
			}
			c.bindings[s.Name] = s.Value
			declared := Unknown
			if !s.Type.IsZero() {
				declared = c.resolveTypeWithGenerics(s.Type.Canonical(), c.genericTypes)
				if declared == Unknown {
					c.reportUnknownOrPrivateType(s.Pos, s.Type.Canonical())
				}
			}
			if lit, ok := s.Value.(*ast.AnonymousObjectLiteral); ok && declared == Unknown {
				typ := c.inferAnonymousObjectLiteralWithSelf(lit, local, s.Name)
				c.info.ExprTypes[lit] = typ
				local[s.Name] = typ
			} else {
				valueType := c.inferExprExpected(s.Value, local, declared)
				if declared != Unknown && valueType != Unknown && !c.typesCompatible(declared, valueType, nil) {
					c.errorf(s.Value.Position(), "binding %q has type %s, expected %s", s.Name, valueType, declared)
				}
				if declared != Unknown {
					local[s.Name] = declared
				} else {
					local[s.Name] = valueType
				}
			}
			result = Void
		case *ast.ObjectDestructureStmt:
			c.inferObjectDestructureStmt(s, local)
			result = Void
		case *ast.AssignStmt:
			if _, exists := local[s.Name]; !exists {
				c.errorf(s.Pos, "cannot assign undefined name %q", s.Name)
			}
			c.inferExpr(s.Value, local)
			result = Void
		case *ast.ExprStmt:
			expected := Unknown
			if isLast {
				expected = blockExpected
			}
			result = c.inferExprExpected(s.Expr, local, expected)
		}
	}
	return result
}

func (c *checker) inferObjectDestructureStmt(stmt *ast.ObjectDestructureStmt, local map[string]Type) {
	valueType := c.inferExpr(stmt.Value, local)
	structInfo := c.info.Types[baseTypeName(valueType)]
	if structInfo == nil {
		if valueType != Unknown {
			c.errorf(stmt.Value.Position(), "type %s has no fields", valueType)
		}
		for _, field := range stmt.Fields {
			if _, exists := local[field.Name]; exists {
				c.errorf(field.NamePos, "name %q is already defined", field.Name)
				continue
			}
			local[field.Name] = Unknown
		}
		return
	}
	if !c.checkPrivateAccess("type", structInfo.Name, structInfo.Private, structInfo.SourcePath, stmt.Pos) {
		return
	}
	seenFields := map[string]bool{}
	for _, binding := range stmt.Fields {
		if seenFields[binding.Field] {
			c.errorf(binding.FieldPos, "duplicate destructured field %q", binding.Field)
			continue
		}
		seenFields[binding.Field] = true
		if _, exists := local[binding.Name]; exists {
			c.errorf(binding.NamePos, "name %q is already defined", binding.Name)
			continue
		}
		field, ok := structInfo.ByName[binding.Field]
		if !ok {
			c.errorf(binding.FieldPos, "type %s has no field %q", valueType, binding.Field)
			local[binding.Name] = Unknown
			continue
		}
		if !c.checkPrivateAccess("field", structInfo.Name+"."+binding.Field, field.Private, field.SourcePath, binding.FieldPos) {
			local[binding.Name] = Unknown
			continue
		}
		local[binding.Name] = structFieldType(structInfo, valueType, field)
		c.bindings[binding.Name] = &ast.SelectorExpr{
			Receiver: stmt.Value,
			Name:     binding.Field,
			Pos:      binding.FieldPos,
			NamePos:  binding.FieldPos,
		}
	}
}

func (c *checker) inferPatternBlock(block *ast.PatternBlock, env map[string]Type) Type {
	return c.inferPatternBlockForSubject(block, Unknown, env)
}

func (c *checker) inferPatternBlockForSubject(block *ast.PatternBlock, subject Type, env map[string]Type) Type {
	result := Unknown
	var branchExprs []ast.Expr
	for _, branch := range block.Branches {
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
		if e.Alternative != nil {
			c.applyExpectedType(e.Alternative, typ)
		}
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
		return c.patternExprType(p.Value)
	case *ast.RangePattern:
		start := c.patternExprType(p.Start)
		end := c.patternExprType(p.End)
		if start == Unknown || end == Unknown || start != end {
			return Unknown
		}
		return start
	case *ast.OrPattern:
		result := Unknown
		for _, alternative := range p.Alternatives {
			typ := c.patternLiteralType(alternative)
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
	case *ast.MapPattern:
		keyType := Unknown
		valueType := Unknown
		for _, entry := range p.Entries {
			key := c.patternExprType(entry.Key)
			if keyType == Unknown {
				keyType = key
			} else if key != Unknown && keyType != key {
				return Unknown
			}
			value := c.patternLiteralType(entry.Pattern)
			if valueType == Unknown {
				valueType = value
			} else if value != Unknown && valueType != value {
				return Unknown
			}
		}
		if keyType != Unknown && valueType != Unknown {
			return MapOf(keyType, valueType)
		}
	}
	return Unknown
}

func (c *checker) checkPatternForSubject(pattern ast.Pattern, subject Type, env map[string]Type) {
	switch pattern.(type) {
	case *ast.MapPattern, *ast.ObjectPattern:
	default:
		patternSubject := c.patternLiteralType(pattern)
		if subject != Unknown && patternSubject != Unknown && subject != patternSubject {
			c.errorf(pattern.Position(), "pattern has type %s, expected %s", patternSubject, subject)
		}
	}
	c.checkPatternWithSubject(pattern, subject, env, false)
}

func (c *checker) patternExprType(expr ast.Expr) Type {
	switch expr.(type) {
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
	case *ast.CharLiteral:
		return Char
	case *ast.NullLiteral:
		return Null
	case *ast.SelectorExpr:
		if typ, ok := c.enumMemberType(expr); ok {
			return typ
		}
	}
	return Unknown
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
	c.checkPatternWithSubject(pattern, Unknown, env, false)
}

func (c *checker) checkPatternWithSubject(pattern ast.Pattern, subject Type, env map[string]Type, optional bool) {
	switch p := pattern.(type) {
	case *ast.WildcardPattern:
	case *ast.BindingPattern:
		if optional {
			c.errorf(p.Pos, "optional pattern binding %q is not available when the key is absent", p.Name)
			return
		}
		p.Type = string(subject)
		env[p.Name] = subject
	case *ast.LiteralPattern:
		c.inferExpr(p.Value, env)
	case *ast.ComparePattern:
		typ := c.inferExpr(p.Value, env)
		if !orderedComparisonType(typ) && typ != Unknown {
			c.errorf(p.Pos, "comparison pattern expects Int, Double, BigInt, String, or Char literal")
		}
	case *ast.RangePattern:
		start := c.inferExpr(p.Start, env)
		end := c.inferExpr(p.End, env)
		if !orderedComparisonType(start) && start != Unknown {
			c.errorf(p.Start.Position(), "range pattern expects Int, Double, BigInt, String, or Char literal")
		}
		if !orderedComparisonType(end) && end != Unknown {
			c.errorf(p.End.Position(), "range pattern expects Int, Double, BigInt, String, or Char literal")
		}
		if start != Unknown && end != Unknown && start != end {
			c.errorf(p.Pos, "range pattern bounds must have matching types, got %s and %s", start, end)
		}
	case *ast.OrPattern:
		for _, alternative := range p.Alternatives {
			c.checkPatternWithSubject(alternative, subject, cloneEnv(env), optional)
		}
	case *ast.TuplePattern:
		for _, elem := range p.Elements {
			c.checkPatternWithSubject(elem, Unknown, env, optional)
		}
	case *ast.ConstructorPattern:
		c.checkConstructorPattern(p, subject, env, optional)
	case *ast.MapPattern:
		c.checkMapPattern(p, subject, env, optional)
	case *ast.ObjectPattern:
		c.checkObjectPattern(p, subject, env, optional)
	}
}

func (c *checker) checkConstructorPattern(pattern *ast.ConstructorPattern, subject Type, env map[string]Type, optional bool) {
	if c.checkEnumConstructorPattern(pattern, subject, env, optional) {
		return
	}
	okType, errType, result := parseResultType(subject)
	if !result {
		if subject != Unknown {
			c.errorf(pattern.Pos, "constructor pattern %s expects Result, got %s", pattern.Name, subject)
		} else {
			c.errorf(pattern.Pos, "constructor pattern requires a Result subject")
		}
		return
	}
	var bindingType Type
	switch pattern.Name {
	case "Ok":
		bindingType = okType
	case "Err":
		bindingType = errType
	default:
		c.errorf(pattern.Pos, "unknown result constructor %q", pattern.Name)
		return
	}
	if pattern.Binding != "" {
		if optional {
			c.errorf(pattern.BindingPos, "optional pattern binding %q is not available when the key is absent", pattern.Binding)
			return
		}
		env[pattern.Binding] = bindingType
	}
}

func (c *checker) checkEnumConstructorPattern(pattern *ast.ConstructorPattern, subject Type, env map[string]Type, optional bool) bool {
	enumName := baseTypeName(subject)
	enum := c.info.Enums[enumName]
	if enum == nil {
		return false
	}
	member, ok := enum.ByName[pattern.Name]
	if !ok || member.HasValue {
		c.errorf(pattern.Pos, "enum %s has no constructor %q", enum.Name, pattern.Name)
		return true
	}
	if !c.checkPrivateAccess("enum constructor", enum.Name+"."+pattern.Name, member.Private, member.SourcePath, pattern.Pos) {
		return true
	}
	if pattern.Binding == "" {
		return true
	}
	if optional {
		c.errorf(pattern.BindingPos, "optional pattern binding %q is not available when the key is absent", pattern.Binding)
		return true
	}
	if len(member.Params) == 0 {
		c.errorf(pattern.BindingPos, "constructor %s.%s does not bind a value", enum.Name, pattern.Name)
		return true
	}
	if len(member.Params) > 1 {
		c.errorf(pattern.BindingPos, "constructor %s.%s binds multiple values", enum.Name, pattern.Name)
		return true
	}
	env[pattern.Binding] = substituteTypeParams(member.Params[0].Type, typeParamBindingsForEnum(enum, subject))
	return true
}

func (c *checker) checkMapPattern(pattern *ast.MapPattern, subject Type, env map[string]Type, optional bool) {
	if !pattern.Rest {
		c.errorf(pattern.Pos, "map pattern requires '..' to ignore unmatched keys")
	}
	keyType, valueType, ok := MapKeyValue(subject)
	if !ok {
		if subject != Unknown {
			c.errorf(pattern.Pos, "map pattern expects Map, got %s", subject)
		}
		keyType = Unknown
		valueType = Unknown
	}
	seen := map[string]lexer.Position{}
	for _, entry := range pattern.Entries {
		key := c.inferExpr(entry.Key, env)
		if keyType != Unknown && key != Unknown && !c.typesCompatible(keyType, key, nil) {
			c.errorf(entry.Key.Position(), "map pattern key has type %s, expected %s", key, keyType)
		}
		if prev, exists := seen[patternKeyText(entry.Key)]; exists {
			c.errorf(entry.Pos, "duplicate map pattern key also used at %s", prev)
		}
		seen[patternKeyText(entry.Key)] = entry.Pos
		c.checkPatternWithSubject(entry.Pattern, valueType, env, entry.Optional)
	}
}

func (c *checker) checkObjectPattern(pattern *ast.ObjectPattern, subject Type, env map[string]Type, optional bool) {
	if !pattern.Rest {
		c.errorf(pattern.Pos, "object pattern requires '..' to ignore unmatched fields")
	}
	structInfo := c.info.Types[baseTypeName(subject)]
	if structInfo == nil {
		if subject != Unknown && subject != Object {
			c.errorf(pattern.Pos, "object pattern expects an object, got %s", subject)
		}
	}
	seen := map[string]lexer.Position{}
	for idx := range pattern.Fields {
		field := &pattern.Fields[idx]
		if prev, exists := seen[field.Name]; exists {
			c.errorf(field.Pos, "duplicate object pattern field %q also used at %s", field.Name, prev)
		}
		seen[field.Name] = field.Pos
		fieldType := Unknown
		field.Exists = structInfo == nil || subject == Object || subject == Unknown
		if structInfo != nil {
			info, ok := structInfo.ByName[field.Name]
			field.Exists = ok
			if !ok {
				if !field.Optional {
					c.errorf(field.Pos, "type %s has no field %q", subject, field.Name)
				}
			} else if c.checkPrivateAccess("field", structInfo.Name+"."+field.Name, info.Private, info.SourcePath, field.Pos) {
				fieldType = structFieldType(structInfo, subject, info)
			}
		}
		field.Type = string(fieldType)
		if !field.Exists && field.Optional {
			continue
		}
		c.checkPatternWithSubject(field.Pattern, fieldType, env, field.Optional)
	}
}

func patternKeyText(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.StringLiteral:
		return "string:" + e.Value
	case *ast.CharLiteral:
		return fmt.Sprintf("char:%d", e.Value)
	case *ast.IntegerLiteral:
		return fmt.Sprintf("int:%d", e.Value)
	case *ast.BoolLiteral:
		return fmt.Sprintf("bool:%t", e.Value)
	default:
		return ast.ExprName(expr)
	}
}
