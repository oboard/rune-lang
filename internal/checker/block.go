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
		start := Unknown
		if p.Start != nil {
			start = c.patternExprType(p.Start)
		}
		end := Unknown
		if p.End != nil {
			end = c.patternExprType(p.End)
		}
		if start == Unknown && end == Unknown {
			return Unknown
		}
		if start != Unknown && end != Unknown && start != end {
			return Unknown
		}
		if start != Unknown {
			return start
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
		return Unknown
	case *ast.ArrayPattern:
		return Unknown
	case *ast.AsPattern:
		return c.patternLiteralType(p.Pattern)
	}
	return Unknown
}

func (c *checker) checkPatternForSubject(pattern ast.Pattern, subject Type, env map[string]Type) {
	c.checkPatternBindingUniqueness(pattern)
	switch pattern.(type) {
	case *ast.MapPattern, *ast.ObjectPattern, *ast.ArrayPattern:
	default:
		patternSubject := c.patternLiteralType(pattern)
		if subject != Unknown && patternSubject != Unknown && subject != patternSubject {
			c.errorf(pattern.Position(), "pattern has type %s, expected %s", patternSubject, subject)
		}
	}
	c.checkPatternWithSubject(pattern, subject, env, false)
}

func (c *checker) checkPatternBindingUniqueness(pattern ast.Pattern) {
	seen := map[string]lexer.Position{}
	var visit func(ast.Pattern)
	visit = func(pattern ast.Pattern) {
		switch p := pattern.(type) {
		case *ast.BindingPattern:
			if c.patternConstantInfo(p) != nil {
				return
			}
			if prev, ok := seen[p.Name]; ok {
				c.errorf(p.Pos, "pattern binding %q was already bound at %s", p.Name, prev)
				return
			}
			seen[p.Name] = p.Pos
		case *ast.AsPattern:
			visit(p.Pattern)
			if prev, ok := seen[p.Name]; ok {
				c.errorf(p.NamePos, "pattern binding %q was already bound at %s", p.Name, prev)
				return
			}
			seen[p.Name] = p.NamePos
		case *ast.OrPattern:
			for _, alt := range p.Alternatives {
				c.checkPatternBindingUniqueness(alt)
			}
		case *ast.TuplePattern:
			for _, elem := range p.Elements {
				visit(elem)
			}
		case *ast.ArrayPattern:
			for _, elem := range p.Elements {
				visit(elem)
			}
		case *ast.SequenceSpreadPattern:
		case *ast.BitPattern:
			visit(p.Value)
		case *ast.ConstructorPattern:
			for _, arg := range p.Args {
				visit(arg)
			}
		case *ast.MapPattern:
			for _, entry := range p.Entries {
				visit(entry.Pattern)
			}
		case *ast.ObjectPattern:
			for _, field := range p.Fields {
				visit(field.Pattern)
			}
		}
	}
	visit(pattern)
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
		if c.checkEnumMemberBindingPattern(p, subject) {
			return
		}
		if value := c.patternConstantInfo(p); value != nil {
			p.Constant = true
			p.Type = string(value.Type)
			p.LinkName = value.LinkName
			if subject != Unknown && value.Type != Unknown && !c.typesCompatible(subject, value.Type, nil) {
				c.errorf(p.Pos, "constant pattern %s has type %s, expected %s", p.Name, value.Type, subject)
			}
			return
		}
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
		start := Unknown
		if p.Start != nil {
			c.checkRangePatternBound(p.Start)
			start = c.inferExpr(p.Start, env)
		}
		end := Unknown
		if p.End != nil {
			c.checkRangePatternBound(p.End)
			end = c.inferExpr(p.End, env)
		}
		if !rangePatternType(start) && start != Unknown {
			c.errorf(p.Start.Position(), "range pattern expects an integer or Char bound")
		}
		if !rangePatternType(end) && end != Unknown {
			c.errorf(p.End.Position(), "range pattern expects an integer or Char bound")
		}
		if start != Unknown && end != Unknown && start != end {
			c.errorf(p.Pos, "range pattern bounds must have matching types, got %s and %s", start, end)
		}
	case *ast.OrPattern:
		c.checkOrPattern(p, subject, env, optional)
	case *ast.TuplePattern:
		for _, elem := range p.Elements {
			c.checkPatternWithSubject(elem, Unknown, env, optional)
		}
	case *ast.ArrayPattern:
		c.checkArrayPattern(p, subject, env, optional)
	case *ast.SequenceSpreadPattern:
		elemType, _, _ := c.sequencePatternTypes(subject)
		c.checkSequenceSpreadPattern(p, subject, elemType, env)
	case *ast.BitPattern:
		c.checkBitPattern(p, env, optional)
	case *ast.AsPattern:
		c.checkPatternWithSubject(p.Pattern, subject, env, optional)
		if optional {
			c.errorf(p.NamePos, "optional pattern binding %q is not available when the key is absent", p.Name)
			return
		}
		p.Type = string(subject)
		env[p.Name] = subject
	case *ast.ConstructorPattern:
		c.checkConstructorPattern(p, subject, env, optional)
	case *ast.MapPattern:
		c.checkMapPattern(p, subject, env, optional)
	case *ast.ObjectPattern:
		c.checkObjectPattern(p, subject, env, optional)
	}
}

func (c *checker) checkEnumMemberBindingPattern(pattern *ast.BindingPattern, subject Type) bool {
	enum := c.info.Enums[baseTypeName(subject)]
	if enum == nil {
		return false
	}
	member, ok := enum.ByName[pattern.Name]
	if !ok || len(member.Params) > 0 {
		return false
	}
	if !c.checkPrivateAccess("enum member", enum.Name+"."+pattern.Name, member.Private, member.SourcePath, pattern.Pos) {
		return true
	}
	pattern.Constant = true
	pattern.Type = string(subject)
	return true
}

func (c *checker) patternConstantInfo(pattern *ast.BindingPattern) *ExternalValueInfo {
	if pattern == nil {
		return nil
	}
	value := c.resolveExternalValue(pattern.Name)
	if value == nil || value.Const == nil {
		return nil
	}
	return value
}

func (c *checker) checkRangePatternBound(expr ast.Expr) {
	if c.patternLiteralOrConst(expr) {
		return
	}
	c.errorf(expr.Position(), "range pattern bound must be a literal, const, or '_'")
}

func rangePatternType(typ Type) bool {
	return isIntegerType(typ) || typ == Char
}

func (c *checker) checkOrPattern(pattern *ast.OrPattern, subject Type, env map[string]Type, optional bool) {
	var expected map[string]Type
	for _, alternative := range pattern.Alternatives {
		branchEnv := cloneEnv(env)
		c.checkPatternWithSubject(alternative, subject, branchEnv, optional)
		bindings := patternBindingTypes(alternative, branchEnv)
		if expected == nil {
			expected = bindings
			continue
		}
		c.compareOrPatternBindings(alternative.Position(), expected, bindings)
		for name, typ := range expected {
			if other, ok := bindings[name]; ok {
				if merged, ok := c.unifyTypes(typ, other); ok {
					expected[name] = merged
				}
			}
		}
	}
	for name, typ := range expected {
		env[name] = typ
	}
}

func (c *checker) compareOrPatternBindings(pos lexer.Position, expected map[string]Type, actual map[string]Type) {
	for name := range expected {
		if _, ok := actual[name]; !ok {
			c.errorf(pos, "or pattern alternative must bind %q", name)
		}
	}
	for name := range actual {
		if _, ok := expected[name]; !ok {
			c.errorf(pos, "or pattern alternative binds extra name %q", name)
		}
	}
}

func patternBindingTypes(pattern ast.Pattern, env map[string]Type) map[string]Type {
	out := map[string]Type{}
	var visit func(ast.Pattern)
	visit = func(pattern ast.Pattern) {
		switch p := pattern.(type) {
		case *ast.BindingPattern:
			if p.Constant {
				return
			}
			out[p.Name] = env[p.Name]
		case *ast.AsPattern:
			visit(p.Pattern)
			out[p.Name] = env[p.Name]
		case *ast.OrPattern:
		case *ast.TuplePattern:
			for _, elem := range p.Elements {
				visit(elem)
			}
		case *ast.ArrayPattern:
			for _, elem := range p.Elements {
				visit(elem)
			}
		case *ast.SequenceSpreadPattern:
		case *ast.ConstructorPattern:
			for _, arg := range p.Args {
				visit(arg)
			}
		case *ast.MapPattern:
			for _, entry := range p.Entries {
				visit(entry.Pattern)
			}
		case *ast.ObjectPattern:
			for _, field := range p.Fields {
				visit(field.Pattern)
			}
		}
	}
	visit(pattern)
	return out
}

func (c *checker) checkArrayPattern(pattern *ast.ArrayPattern, subject Type, env map[string]Type, optional bool) {
	if arrayPatternHasBits(pattern) {
		c.checkBitArrayPattern(pattern, subject, env, optional)
		return
	}
	elemType, restType, ok := c.sequencePatternTypes(subject)
	if !ok && subject != Unknown {
		c.errorf(pattern.Pos, "array pattern expects Array, String, or Bytes, got %s", subject)
	}
	pattern.SubjectType = string(subject)
	pattern.RestType = string(restType)
	if optional && pattern.RestBinding != "" {
		c.errorf(pattern.RestPos, "optional pattern binding %q is not available when the key is absent", pattern.RestBinding)
	}
	for _, elem := range pattern.Elements {
		if spread, ok := elem.(*ast.SequenceSpreadPattern); ok {
			c.checkSequenceSpreadPattern(spread, subject, elemType, env)
			continue
		}
		c.checkPatternWithSubject(elem, elemType, env, optional)
	}
	if pattern.RestBinding != "" && !optional {
		env[pattern.RestBinding] = restType
	}
}

func (c *checker) checkSequenceSpreadPattern(pattern *ast.SequenceSpreadPattern, subject Type, elemType Type, env map[string]Type) {
	typ := c.inferExpr(pattern.Value, env)
	pattern.Type = string(typ)
	if typ == Unknown || subject == Unknown {
		return
	}
	switch subject {
	case String:
		if typ != String {
			c.errorf(pattern.Pos, "string pattern spread expects String, got %s", typ)
		}
	case Bytes:
		if typ != Bytes {
			c.errorf(pattern.Pos, "bytes pattern spread expects Bytes, got %s", typ)
		}
	default:
		if spreadElem, ok := ArrayElement(typ); ok {
			if elemType != Unknown && spreadElem != Unknown && !c.typesCompatible(elemType, spreadElem, nil) {
				c.errorf(pattern.Pos, "array pattern spread has element type %s, expected %s", spreadElem, elemType)
			}
			return
		}
		c.errorf(pattern.Pos, "array pattern spread expects Array, String, or Bytes, got %s", typ)
	}
}

func arrayPatternHasBits(pattern *ast.ArrayPattern) bool {
	for _, elem := range pattern.Elements {
		if _, ok := elem.(*ast.BitPattern); ok {
			return true
		}
	}
	return false
}

func (c *checker) checkBitArrayPattern(pattern *ast.ArrayPattern, subject Type, env map[string]Type, optional bool) {
	restType, ok := c.bitPatternSubjectType(subject)
	if !ok && subject != Unknown {
		c.errorf(pattern.Pos, "bitstring pattern expects Bytes or Array[UInt8], got %s", subject)
	}
	pattern.SubjectType = string(subject)
	pattern.RestType = string(restType)
	if optional && pattern.RestBinding != "" {
		c.errorf(pattern.RestPos, "optional pattern binding %q is not available when the key is absent", pattern.RestBinding)
	}
	bitsBeforeRest := 0
	bitOffset := 0
	for idx, elem := range pattern.Elements {
		bit, ok := elem.(*ast.BitPattern)
		if !ok {
			c.errorf(elem.Position(), "bitstring pattern cannot mix bit fields with element patterns")
			c.checkPatternWithSubject(elem, UInt8, env, optional)
			continue
		}
		if bit.Endian == "le" && bitOffset%8 != 0 {
			c.errorf(bit.Pos, "little-endian bitstring pattern must start on a byte boundary")
		}
		c.checkBitPattern(bit, env, optional)
		if pattern.RestIndex < 0 || idx < pattern.RestIndex {
			bitsBeforeRest += bit.Width
		}
		bitOffset += bit.Width
	}
	if pattern.RestBinding != "" && bitsBeforeRest%8 != 0 {
		c.errorf(pattern.RestPos, "bitstring rest must start on a byte boundary")
	}
	if pattern.RestBinding != "" && !optional {
		env[pattern.RestBinding] = restType
	}
}

func (c *checker) bitPatternSubjectType(subject Type) (Type, bool) {
	if subject == Bytes || subject == Unknown {
		return subject, true
	}
	if elem, ok := ArrayElement(subject); ok && elem == UInt8 {
		return subject, true
	}
	return Unknown, false
}

func (c *checker) checkBitPattern(pattern *ast.BitPattern, env map[string]Type, optional bool) {
	if pattern.Width < 1 || pattern.Width > 64 {
		c.errorf(pattern.Pos, "bitstring pattern width must be between 1 and 64")
	}
	if pattern.Endian == "le" && pattern.Width%8 != 0 {
		c.errorf(pattern.Pos, "little-endian bitstring pattern width must be byte-aligned")
	}
	c.checkPatternWithSubject(pattern.Value, bitPatternValueType(pattern), env, optional)
}

func bitPatternValueType(pattern *ast.BitPattern) Type {
	if pattern.Width > 32 {
		if pattern.Signed {
			return Int64
		}
		return UInt64
	}
	if pattern.Signed {
		return Int
	}
	return UInt
}

func (c *checker) sequencePatternTypes(subject Type) (Type, Type, bool) {
	if elem, ok := ArrayElement(subject); ok {
		return elem, subject, true
	}
	switch subject {
	case String:
		return Char, String, true
	case Bytes:
		return UInt8, Bytes, true
	case Unknown:
		return Unknown, Unknown, true
	default:
		return Unknown, Unknown, false
	}
}

func (c *checker) checkConstructorPattern(pattern *ast.ConstructorPattern, subject Type, env map[string]Type, optional bool) {
	pattern.SubjectType = string(subject)
	if c.checkEnumConstructorPattern(pattern, subject, env, optional) {
		return
	}
	if c.checkJSONConstructorPattern(pattern, subject, env, optional) {
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
	}
	c.checkConstructorArgs(pattern, []Type{bindingType}, env, optional)
	if pattern.Binding != "" && len(pattern.Args) == 0 {
		env[pattern.Binding] = bindingType
	}
}

func (c *checker) checkJSONConstructorPattern(pattern *ast.ConstructorPattern, subject Type, env map[string]Type, optional bool) bool {
	if subject != Object && subject != Unknown {
		return false
	}
	var payload Type
	switch pattern.Name {
	case "Array":
		payload = ArrayOf(Object)
	case "Object":
		payload = Object
	case "String":
		payload = String
	case "Bool":
		payload = Bool
	case "Number":
		payload = Unknown
	case "Null":
		payload = Null
	default:
		return false
	}
	if pattern.Name == "Null" {
		if len(pattern.Args) > 0 {
			c.errorf(pattern.Pos, "JSON Null pattern expects no args")
		}
		return true
	}
	c.checkConstructorArgs(pattern, []Type{payload}, env, optional)
	return true
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
	if optional {
		c.errorf(pattern.BindingPos, "optional pattern binding %q is not available when the key is absent", pattern.Binding)
		return true
	}
	paramTypes := make([]Type, 0, len(member.Params))
	for _, param := range member.Params {
		paramTypes = append(paramTypes, substituteTypeParams(param.Type, typeParamBindingsForEnum(enum, subject)))
	}
	if pattern.Binding != "" && len(pattern.Args) == 0 && len(member.Params) == 0 {
		c.errorf(pattern.BindingPos, "constructor %s.%s does not bind a value", enum.Name, pattern.Name)
		return true
	}
	if pattern.Binding != "" && len(pattern.Args) == 0 && len(member.Params) > 1 {
		c.errorf(pattern.BindingPos, "constructor %s.%s binds multiple values", enum.Name, pattern.Name)
		return true
	}
	c.checkConstructorArgs(pattern, paramTypes, env, optional)
	if pattern.Binding != "" && len(pattern.Args) == 0 {
		env[pattern.Binding] = paramTypes[0]
	}
	return true
}

func (c *checker) checkConstructorArgs(pattern *ast.ConstructorPattern, paramTypes []Type, env map[string]Type, optional bool) {
	if pattern.Rest {
		if len(pattern.Args) > len(paramTypes) {
			c.errorf(pattern.Pos, "constructor %s pattern expects at most %d args before '..', got %d", pattern.Name, len(paramTypes), len(pattern.Args))
			return
		}
	} else if len(pattern.Args) != len(paramTypes) {
		c.errorf(pattern.Pos, "constructor %s pattern expects %d args, got %d", pattern.Name, len(paramTypes), len(pattern.Args))
		return
	}
	for idx, arg := range pattern.Args {
		expected := Unknown
		if idx < len(paramTypes) {
			expected = paramTypes[idx]
		}
		c.checkPatternWithSubject(arg, expected, env, optional)
	}
}

func (c *checker) checkMapPattern(pattern *ast.MapPattern, subject Type, env map[string]Type, optional bool) {
	if !pattern.Rest {
		c.errorf(pattern.Pos, "map pattern requires '..' to ignore unmatched keys")
	}
	pattern.SubjectType = string(subject)
	keyType, valueType, access, ok := c.mapPatternTypes(subject, pattern.Pos)
	pattern.Access = access
	pattern.ValueType = string(valueType)
	if !ok {
		keyType = Unknown
		valueType = Unknown
		pattern.ValueType = string(Unknown)
	}
	seen := map[string]lexer.Position{}
	for _, entry := range pattern.Entries {
		c.checkMapPatternKey(entry.Key)
		key := c.inferExpr(entry.Key, env)
		if keyType != Unknown && key != Unknown && !c.typesCompatible(keyType, key, nil) {
			c.errorf(entry.Key.Position(), "map pattern key has type %s, expected %s", key, keyType)
		}
		if prev, exists := seen[patternKeyText(entry.Key)]; exists {
			c.errorf(entry.Pos, "duplicate map pattern key also used at %s", prev)
		}
		seen[patternKeyText(entry.Key)] = entry.Pos
		entryType := valueType
		entryOptional := entry.Optional
		if entry.Optional && valueType != Unknown {
			entryType = NullableOf(valueType)
			entryOptional = false
		}
		c.checkPatternWithSubject(entry.Pattern, entryType, env, entryOptional)
	}
}

func (c *checker) checkMapPatternKey(expr ast.Expr) {
	if c.patternLiteralOrConst(expr) {
		return
	}
	c.errorf(expr.Position(), "map pattern key must be a literal or const")
}

func (c *checker) patternLiteralOrConst(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.StringLiteral, *ast.CharLiteral, *ast.IntegerLiteral, *ast.DoubleLiteral,
		*ast.BigIntLiteral, *ast.BoolLiteral, *ast.NullLiteral:
		return true
	case *ast.Identifier:
		value := c.resolveExternalValue(e.Name)
		return value != nil && value.Const != nil
	}
	return false
}

func (c *checker) mapPatternTypes(subject Type, pos lexer.Position) (Type, Type, string, bool) {
	if key, value, ok := MapKeyValue(subject); ok {
		return key, value, "map", true
	}
	switch subject {
	case Object:
		return String, Unknown, "object", true
	case Unknown:
		return Unknown, Unknown, "", true
	}
	structInfo := c.info.Types[baseTypeName(subject)]
	if structInfo == nil {
		c.errorf(pos, "map pattern expects Map or a type with get(key) -> V?, got %s", subject)
		return Unknown, Unknown, "", false
	}
	method := structInfo.Methods["get"]
	if method == nil || method.Static || len(method.Params) != 1 {
		c.errorf(pos, "map pattern expects type %s to define get(key) -> V?", subject)
		return Unknown, Unknown, "", false
	}
	if !c.checkPrivateAccess("method", structInfo.Name+".get", method.Private, method.SourcePath, pos) {
		return Unknown, Unknown, "", false
	}
	bindings := typeParamBindingsForStruct(structInfo, subject)
	keyType := substituteTypeParams(method.Params[0].Type, bindings)
	returnType := substituteTypeParams(method.Return, bindings)
	inner, ok := parseNullableType(string(returnType))
	if !ok {
		c.errorf(pos, "map pattern expects get(key) to return nullable value, got %s", returnType)
		return Unknown, Unknown, "", false
	}
	return keyType, Type(inner), "get", true
}

func (c *checker) checkObjectPattern(pattern *ast.ObjectPattern, subject Type, env map[string]Type, optional bool) {
	if !pattern.Rest {
		c.errorf(pattern.Pos, "object pattern requires '..' to ignore unmatched fields")
	}
	pattern.SubjectType = string(subject)
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
