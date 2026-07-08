package checker

import (
	"sort"
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func (c *checker) inferMethods(typ *ast.StructType) {
	structInfo := c.info.Types[typ.Name]
	if structInfo == nil {
		return
	}
	c.withSourcePath(typ.SourcePath, func() {
		for _, method := range typ.Methods {
			info := structInfo.Methods[method.Name]
			if method.Static {
				info = structInfo.StaticMethods[method.Name]
			}
			if info == nil {
				continue
			}
			c.withSourcePath(method.SourcePath, func() {
				c.withGenericTypes(info.Generics, info.GenericConstraints, func() {
					if isIntrinsicStub(method) {
						if !info.ReturnDeclared {
							info.Return = Void
						}
						return
					}
					env := map[string]Type{}
					if !info.Static {
						env["this"] = info.ReceiverType
					}
					for _, param := range info.Params {
						env[param.Name] = param.Type
					}
					c.rewritePatternPredicateBody(method, info, env)
					ret := c.inferFunctionBody(method, info, env)
					c.finishInferredParams(info, method.Body, env)
					c.finishFunctionReturn(info, ret, c.popRoutineErrors(), method)
				})
			})
		}
	})
}

func (c *checker) inferEnumMethods(enum *ast.EnumType) {
	enumInfo := c.info.Enums[enum.Name]
	if enumInfo == nil {
		return
	}
	c.withSourcePath(enum.SourcePath, func() {
		for _, method := range enum.Methods {
			info := enumInfo.Methods[method.Name]
			if method.Static {
				info = enumInfo.StaticMethods[method.Name]
			}
			if info == nil {
				continue
			}
			c.withSourcePath(method.SourcePath, func() {
				c.withGenericTypes(info.Generics, info.GenericConstraints, func() {
					env := map[string]Type{}
					if !info.Static {
						env["this"] = info.ReceiverType
					}
					for _, param := range info.Params {
						env[param.Name] = param.Type
					}
					c.rewritePatternPredicateBody(method, info, env)
					ret := c.inferFunctionBody(method, info, env)
					c.finishInferredParams(info, method.Body, env)
					c.finishFunctionReturn(info, ret, c.popRoutineErrors(), method)
				})
			})
		}
	})
}

func (c *checker) inferFunction(fn *ast.Function) {
	if c.inferredFunctions[fn] {
		return
	}
	if c.inferringFunctions[fn] {
		return
	}
	info := c.info.FunctionDecls[fn]
	if info == nil {
		return
	}
	c.inferringFunctions[fn] = true
	defer delete(c.inferringFunctions, fn)
	c.withSourcePath(fn.SourcePath, func() {
		c.withGenericTypes(info.Generics, info.GenericConstraints, func() {
			if isIntrinsicStub(fn) {
				if !info.ReturnDeclared {
					info.Return = Void
				}
				c.inferredFunctions[fn] = true
				return
			}
			env := map[string]Type{}
			inferredFields := inferParamFields(fn.Body, paramNames(info.Params))
			for _, param := range info.Params {
				paramType := param.Type
				if paramType == Unknown {
					if inferred := c.objectTypeFromFields(inferredFields[param.Name]); inferred != Unknown {
						paramType = inferred
					}
				}
				env[param.Name] = paramType
			}
			c.rewritePatternPredicateBody(fn, info, env)
			ret := c.inferFunctionBody(fn, info, env)
			c.finishInferredParams(info, fn.Body, env)
			unwrapErr := c.popRoutineErrors()
			if fn.Name == "main" && !fn.Macro {
				if info.ReturnDeclared && info.Return != Void {
					c.errorf(fn.NamePos, "main must return Void, got %s", info.Return)
				}
				info.Return = Void
				c.inferredFunctions[fn] = true
				return
			}
			c.finishFunctionReturn(info, ret, unwrapErr, fn)
			c.inferredFunctions[fn] = true
		})
	})
}

func (c *checker) withGenericTypes(names []string, constraints map[string]string, fn func()) {
	prev := c.genericTypes
	prevConstraints := c.genericConstraints
	if len(names) > 0 {
		c.genericTypes = genericSet(names...)
	} else {
		c.genericTypes = nil
	}
	c.genericConstraints = constraints
	fn()
	c.genericTypes = prev
	c.genericConstraints = prevConstraints
}

func (c *checker) finishInferredParams(info *FuncInfo, body ast.Expr, env map[string]Type) {
	for idx, param := range info.Params {
		if param.Type != Unknown {
			continue
		}
		if inferred := env[param.Name]; inferred != "" && inferred != Unknown {
			info.Params[idx].Type = inferred
			continue
		}
		if inferred := c.inferredParamUseType(body, param.Name, info.Name); inferred != Unknown {
			info.Params[idx].Type = inferred
		}
	}
}

func (c *checker) inferFunctionBody(fn *ast.Function, info *FuncInfo, env map[string]Type) Type {
	if info != nil && info.Routine {
		c.routineDepth++
		c.unwrapErrors = append(c.unwrapErrors, Unknown)
	}
	if block, ok := fn.Body.(*ast.PatternBlock); ok && info != nil && len(info.Params) > 0 {
		ret := c.inferPatternBlockForSubject(block, info.Params[0].Type, env)
		if info == nil || !info.Routine {
			c.unwrapErrors = append(c.unwrapErrors, Unknown)
		}
		return ret
	}
	ret := Unknown
	if info != nil && info.ReturnDeclared {
		ret = c.inferExprExpected(fn.Body, env, info.Return)
	} else {
		ret = c.inferExpr(fn.Body, env)
	}
	if info == nil || !info.Routine {
		c.unwrapErrors = append(c.unwrapErrors, Unknown)
	}
	return ret
}

func (c *checker) inferConstDecl(constant *ast.ConstDecl) Type {
	info := c.info.ConstDecls[constant]
	if info == nil || constant.Value == nil {
		return Unknown
	}
	var typ Type
	c.withSourcePath(constant.SourcePath, func() {
		env := map[string]Type{}
		if info.Type != Unknown {
			typ = c.inferExprExpected(constant.Value, env, info.Type)
		} else {
			typ = c.inferExpr(constant.Value, env)
		}
		if info.Type != Unknown && typ != Unknown && !c.typesCompatible(info.Type, typ, nil) {
			c.errorf(constant.Value.Position(), "constant %s has type %s, expected %s", constant.Name, typ, info.Type)
		}
		if info.Type == Unknown {
			info.Type = typ
		}
	})
	return info.Type
}

func (c *checker) inferExprExpected(expr ast.Expr, env map[string]Type, expected Type) Type {
	previous := c.expectedType
	c.expectedType = expected
	typ := c.inferExpr(expr, env)
	c.expectedType = previous
	if typ == Unknown && expected != Unknown {
		c.info.ExprTypes[expr] = expected
		return expected
	}
	return typ
}

func (c *checker) popRoutineErrors() Type {
	if len(c.unwrapErrors) == 0 {
		return Unknown
	}
	errType := c.unwrapErrors[len(c.unwrapErrors)-1]
	c.unwrapErrors = c.unwrapErrors[:len(c.unwrapErrors)-1]
	if c.routineDepth > 0 {
		c.routineDepth--
	}
	return errType
}

func (c *checker) inferTest(test *ast.Test) {
	if test.Body == nil {
		return
	}
	c.withSourcePath(test.SourcePath, func() {
		c.inferExpr(test.Body, map[string]Type{})
	})
}

func (c *checker) inferredParamUseType(body ast.Expr, name string, selfFunction string) Type {
	result := Unknown
	var walk func(ast.Expr, Type)
	walk = func(expr ast.Expr, expected Type) {
		if result != Unknown {
			return
		}
		switch e := expr.(type) {
		case *ast.Identifier:
			if e.Name != name {
				return
			}
			if expected != Unknown {
				result = expected
				return
			}
			if typ := c.info.ExprTypes[e]; typ != "" && typ != Unknown {
				result = typ
			}
		case *ast.BinaryExpr:
			switch e.Op {
			case lexer.Plus, lexer.Minus, lexer.Star, lexer.Slash, lexer.Percent, lexer.BitAnd, lexer.BitOr, lexer.BitXor, lexer.ShiftLeft, lexer.ShiftRight, lexer.UnsignedShiftRight:
				walk(e.Left, Int)
				walk(e.Right, Int)
			case lexer.AndAnd, lexer.OrOr:
				walk(e.Left, Bool)
				walk(e.Right, Bool)
			default:
				walk(e.Left, Unknown)
				walk(e.Right, Unknown)
			}
		case *ast.TernaryExpr:
			walk(e.Condition, Bool)
			walk(e.Consequence, expected)
			if e.Alternative != nil {
				walk(e.Alternative, expected)
			}
		case *ast.CallExpr:
			if ident, ok := e.Callee.(*ast.Identifier); ok && selfFunction != "" && ident.Name == selfFunction && expected != Unknown {
				for _, arg := range e.Args {
					walk(arg, expected)
				}
				return
			}
			walk(e.Callee, Unknown)
			for _, arg := range e.Args {
				walk(arg, Unknown)
			}
		default:
			ast.WalkExpr(expr, func(child ast.Expr) {
				if child != expr {
					walk(child, Unknown)
				}
			})
		}
	}
	walk(body, Unknown)
	return result
}

func (c *checker) finishFunctionReturn(info *FuncInfo, ret Type, unwrapErr Type, fn *ast.Function) {
	if info.Routine && unwrapErr != Unknown {
		if _, _, ok := parseResultType(ret); !ok {
			ret = ResultOf(ret, unwrapErr)
		}
	}
	if info.ReturnDeclared {
		if info.Return == WebComponent && ret == HTMLElement && exprCanBuildWebComponent(fn.Body) {
			return
		}
		if !c.typesCompatible(info.Return, ret, info.Generics) {
			c.errorf(fn.Body.Position(), "function %q returns %s, expected %s", fn.Name, ret, info.Return)
		}
		return
	}
	if ret == Unknown {
		ret = Void
	}
	info.Return = ret
}

func exprCanBuildWebComponent(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.XMLElement:
		return true
	case *ast.BlockExpr:
		if len(e.Statements) == 0 {
			return false
		}
		last, ok := e.Statements[len(e.Statements)-1].(*ast.ExprStmt)
		return ok && exprCanBuildWebComponent(last.Expr)
	case *ast.TernaryExpr:
		return e.Alternative != nil &&
			exprCanBuildWebComponent(e.Consequence) &&
			exprCanBuildWebComponent(e.Alternative)
	case *ast.MatchExpr:
		if len(e.Branches) == 0 {
			return false
		}
		for _, branch := range e.Branches {
			if !exprCanBuildWebComponent(branch.Expr) {
				return false
			}
		}
		return true
	case *ast.PatternBlock:
		if len(e.Branches) == 0 {
			return false
		}
		for _, branch := range e.Branches {
			if !exprCanBuildWebComponent(branch.Expr) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (c *checker) inferExpr(expr ast.Expr, env map[string]Type) Type {
	typ := c.inferExprType(expr, env)
	if expr != nil {
		c.info.ExprTypes[expr] = typ
	}
	return typ
}

func (c *checker) inferExprType(expr ast.Expr, env map[string]Type) Type {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return Int
	case *ast.DoubleLiteral:
		return Double
	case *ast.BigIntLiteral:
		return BigInt
	case *ast.StringLiteral:
		return String
	case *ast.TemplateLiteral:
		for _, part := range e.Parts {
			c.inferExpr(part.Expr, env)
		}
		return String
	case *ast.CharLiteral:
		return Char
	case *ast.RegexLiteral:
		if err := validateRegexFlags(e.Flags); err != "" {
			c.errorf(e.Pos, err)
		}
		return Regex
	case *ast.BoolLiteral:
		return Bool
	case *ast.NullLiteral:
		return Null
	case *ast.Identifier:
		if typ, ok := env[e.Name]; ok {
			return typ
		}
		if fn, ok := c.resolveFunction(e.Name, e.Pos); ok {
			if fn == nil {
				return Unknown
			}
			typ := functionType(fn)
			c.info.ResolvedFunctions[e] = fn
			c.info.ExprTypes[e] = typ
			return typ
		}
		if value := c.resolveExternalValue(e.Name); value != nil {
			c.info.ResolvedValues[e] = value
			return value.Type
		}
		if e.Name != "<error>" {
			c.errorf(e.Pos, "undefined name %q", e.Name)
		}
		return Unknown
	case *ast.AtExpr:
		return c.inferAtExpr(e)
	case *ast.ThisExpr:
		typ, ok := env["this"]
		if !ok {
			c.errorf(e.Pos, "implicit this selector can only be used inside a method")
			return Unknown
		}
		return typ
	case *ast.SelectorExpr:
		if e.Static {
			return c.inferStaticSelector(e, env)
		}
		if typ, ok := c.inferEnumMemberSelector(e, env); ok {
			return typ
		}
		receiver := c.inferExpr(e.Receiver, env)
		if typ, ok := c.inferNamespaceSelector(e, receiver); ok {
			return typ
		}
		if trait := c.traitView(receiver); trait != nil {
			field, ok := trait.ByName[e.Name]
			if !ok {
				c.errorf(e.Pos, "trait %s has no field %q", trait.Name, e.Name)
				return Unknown
			}
			return substituteSelfType(field.Type, receiver)
		}
		structInfo := c.info.Types[baseTypeName(receiver)]
		if structInfo == nil {
			if receiver != Unknown {
				c.errorf(e.Pos, "type %s has no fields", receiver)
			}
			return Unknown
		}
		if !c.checkPrivateAccess("type", structInfo.Name, structInfo.Private, structInfo.SourcePath, e.Pos) {
			return Unknown
		}
		field, ok := structInfo.ByName[e.Name]
		if !ok {
			c.errorf(e.Pos, "type %s has no field %q", receiver, e.Name)
			return Unknown
		}
		if !c.checkPrivateAccess("field", structInfo.Name+"."+e.Name, field.Private, field.SourcePath, e.Pos) {
			return Unknown
		}
		return structFieldType(structInfo, receiver, field)
	case *ast.StructLiteral:
		return c.inferStructLiteral(e, env)
	case *ast.AnonymousObjectLiteral:
		return c.inferAnonymousObjectLiteral(e, env)
	case *ast.ArrayLiteral:
		return c.inferArrayLiteral(e, env)
	case *ast.TupleLiteral:
		return c.inferTupleLiteral(e, env)
	case *ast.SpreadExpr:
		return c.inferExpr(e.Expr, env)
	case *ast.ReactiveLiteral:
		return c.inferExpr(e.Value, env)
	case *ast.MapLiteral:
		return c.inferMapLiteral(e, env)
	case *ast.IndexExpr:
		return c.inferIndexExpr(e, env)
	case *ast.UnaryExpr:
		typ := c.inferExpr(e.Expr, env)
		switch e.Op {
		case lexer.Minus:
			if !isNumericType(typ) && typ != Unknown {
				c.errorf(e.Pos, "operator '-' expects a numeric type, got %s", typ)
			}
			return typ
		case lexer.Tilde:
			if !isBitwiseType(typ) && typ != Unknown {
				c.errorf(e.Pos, "operator '~' expects an integer type, got %s", typ)
			}
			return typ
		case lexer.Bang:
			if typ != Bool && typ != Unknown {
				c.errorf(e.Pos, "operator '!' expects Bool, got %s", typ)
			}
			return Bool
		default:
			return Unknown
		}
	case *ast.PostfixExpr:
		typ := c.inferExpr(e.Expr, env)
		switch e.Op {
		case lexer.PlusPlus:
			if !isNumericType(typ) && typ != Unknown {
				c.errorf(e.Pos, "operator '++' expects a numeric type, got %s", typ)
			}
			return typ
		default:
			return Unknown
		}
	case *ast.ResultUnwrapExpr:
		result := c.inferExpr(e.Expr, env)
		value, errType, ok := parseResultType(result)
		if !ok {
			if result != Unknown {
				c.errorf(e.Pos, "operator '?' expects Result, got %s", result)
			}
			return Unknown
		}
		if c.routineDepth == 0 {
			c.errorf(e.Pos, "operator '?' can only be used inside a routine")
		} else if len(c.unwrapErrors) > 0 {
			current := c.unwrapErrors[len(c.unwrapErrors)-1]
			if current == Unknown {
				c.unwrapErrors[len(c.unwrapErrors)-1] = errType
			} else if unified, ok := c.unifyTypes(current, errType); ok {
				c.unwrapErrors[len(c.unwrapErrors)-1] = unified
			} else {
				c.errorf(e.Pos, "operator '?' lifts %s, expected %s", errType, current)
			}
		}
		return value
	case *ast.CompileTimeExpr:
		return c.inferExpr(e.Expr, env)
	case *ast.BinaryExpr:
		left := c.inferExpr(e.Left, env)
		right := c.inferExpr(e.Right, env)
		switch e.Op {
		case lexer.QuestionQuestion:
			return c.nullCoalesceType(e, left, right)
		case lexer.AndAnd, lexer.OrOr:
			if left != Bool && left != Unknown {
				c.errorf(e.Left.Position(), "operator '%s' expects Bool, got %s", e.Op, left)
			}
			if right != Bool && right != Unknown {
				c.errorf(e.Right.Position(), "operator '%s' expects Bool, got %s", e.Op, right)
			}
			return Bool
		case lexer.Plus:
			if left == String || right == String {
				left = c.refineUnknownIdentifierType(e.Left, left, String, env)
				right = c.refineUnknownIdentifierType(e.Right, right, String, env)
				if left != String && left != Unknown {
					c.errorf(e.Left.Position(), "string concatenation expects String, got %s", left)
				}
				if right != String && right != Unknown {
					c.errorf(e.Right.Position(), "string concatenation expects String, got %s", right)
				}
				return String
			}
			left, right = c.refineNumericBinaryOperands(e, left, right, env)
			return c.numericBinaryType(e, left, right)
		case lexer.Minus, lexer.Star, lexer.Slash, lexer.Percent:
			left, right = c.refineNumericBinaryOperands(e, left, right, env)
			return c.numericBinaryType(e, left, right)
		case lexer.BitAnd, lexer.BitOr, lexer.BitXor, lexer.ShiftLeft, lexer.ShiftRight, lexer.UnsignedShiftRight:
			return c.bitwiseBinaryType(e, left, right)
		case lexer.EqualEqual, lexer.BangEqual:
			if !typesComparable(left, right) {
				c.errorf(e.Pos, "cannot compare %s and %s", left, right)
			}
			return Bool
		case lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
			if !orderedComparisonType(left) && left != Unknown {
				c.errorf(e.Left.Position(), "ordered comparison expects a numeric type or String, got %s", left)
			}
			if !orderedComparisonType(right) && right != Unknown {
				c.errorf(e.Right.Position(), "ordered comparison expects a numeric type or String, got %s", right)
			}
			if left != Unknown && right != Unknown && left != right {
				c.errorf(e.Pos, "ordered comparison requires matching types, got %s and %s", left, right)
			}
			return Bool
		case lexer.DotDotEqual:
			c.errorf(e.Pos, "range pattern can only be used as a pattern predicate")
			return Unknown
		default:
			return Unknown
		}
	case *ast.TernaryExpr:
		condition := c.inferExpr(e.Condition, env)
		if condition != Bool && condition != Unknown {
			c.errorf(e.Condition.Position(), "ternary condition expects Bool, got %s", condition)
		}
		consequence := c.inferExpr(e.Consequence, env)
		if e.Alternative == nil {
			return Void
		}
		alternative := c.inferExpr(e.Alternative, env)
		result, ok := c.mergeFunctionValueTypes(consequence, alternative)
		if !ok {
			result, ok = c.unifyTypes(consequence, alternative)
		}
		if !ok {
			c.errorf(e.Pos, "ternary branches return %s and %s", consequence, alternative)
			return Unknown
		}
		return result
	case *ast.AssignExpr:
		if target, ok := e.Target.(*ast.IndexExpr); ok {
			return c.inferIndexAssign(target, e.Value, env)
		}
		if e.Name != "" {
			if _, exists := env[e.Name]; !exists {
				c.errorf(e.Pos, "cannot assign undefined name %q", e.Name)
			}
			if e.Target != nil {
				c.inferExpr(e.Target, env)
			}
			c.inferExpr(e.Value, env)
			return Void
		}
		if e.Target == nil {
			c.errorf(e.Pos, "cannot assign undefined name %q", e.Name)
			c.inferExpr(e.Value, env)
			return Void
		}
		expected := c.inferExpr(e.Target, env)
		actual := c.inferExpr(e.Value, env)
		if expected != Unknown && actual != Unknown && !c.typesCompatible(expected, actual, nil) {
			c.errorf(e.Value.Position(), "assignment has type %s, expected %s", actual, expected)
		}
		return Void
	case *ast.CallExpr:
		return c.inferCall(e, env)
	case *ast.LambdaExpr:
		return c.inferLambda(e, env)
	case *ast.BlockExpr:
		return c.inferBlock(e, env)
	case *ast.PatternBlock:
		return c.inferPatternBlock(e, env)
	case *ast.MatchExpr:
		return c.inferMatchExpr(e, env)
	case *ast.XMLElement:
		return c.inferXMLElement(e, env)
	case *ast.WatchExpr:
		target := c.inferExpr(e.Target, env)
		if lambda, ok := e.Handler.(*ast.LambdaExpr); ok && len(lambda.Params) == 2 && target != Unknown {
			c.applyExpectedType(lambda, FuncOfTypes([]Type{target, target}, Void))
		}
		handler := c.inferExpr(e.Handler, env)
		if handler != Unknown {
			params, _, ok := parseFuncType(string(handler))
			if !ok {
				c.errorf(e.Handler.Position(), "watch handler must be a function, got %s", handler)
			} else if len(params) != 0 && len(params) != 2 {
				c.errorf(e.Handler.Position(), "watch handler must accept zero or two parameters")
			}
		}
		return Void
	default:
		return Unknown
	}
}

func (c *checker) nullCoalesceType(expr *ast.BinaryExpr, left Type, right Type) Type {
	if left == Unknown || right == Unknown {
		return Unknown
	}
	if left == Null {
		return right
	}
	inner, ok := parseNullableType(string(left))
	if !ok {
		c.errorf(expr.Left.Position(), "operator '??' expects a nullable left operand, got %s", left)
		return left
	}
	result, ok := c.unifyTypes(Type(inner), right)
	if !ok {
		c.errorf(expr.Pos, "operator '??' fallback has type %s, expected %s", right, Type(inner))
		return Unknown
	}
	return result
}

func (c *checker) inferEnumMemberSelector(sel *ast.SelectorExpr, env map[string]Type) (Type, bool) {
	ident, ok := sel.Receiver.(*ast.Identifier)
	if !ok {
		return Unknown, false
	}
	if _, exists := env[ident.Name]; exists {
		return Unknown, false
	}
	enum := c.info.Enums[ident.Name]
	if enum == nil {
		return Unknown, false
	}
	if !c.checkPrivateAccess("enum", enum.Name, enum.Private, enum.SourcePath, ident.Pos) {
		return Unknown, true
	}
	member, ok := enum.ByName[sel.Name]
	if !ok {
		c.errorf(sel.NamePos, "enum %s has no member %q", enum.Name, sel.Name)
		return Unknown, true
	}
	if !c.checkPrivateAccess("enum member", enum.Name+"."+sel.Name, member.Private, member.SourcePath, sel.NamePos) {
		return Unknown, true
	}
	c.info.ExprTypes[ident] = Type(enum.Name)
	return Type(enum.Name), true
}

func (c *checker) inferStaticSelector(sel *ast.SelectorExpr, env map[string]Type) Type {
	ident, ok := sel.Receiver.(*ast.Identifier)
	if !ok {
		c.errorf(sel.Pos, "static selector receiver must be a type")
		return Unknown
	}
	if _, shadowed := env[ident.Name]; shadowed {
		c.errorf(sel.Pos, "static selector receiver %q is a value, not a type", ident.Name)
		return Unknown
	}
	structInfo := c.info.Types[ident.Name]
	enumInfo := c.info.Enums[ident.Name]
	if structInfo == nil && enumInfo == nil {
		c.errorf(sel.Pos, "unknown type %q", ident.Name)
		return Unknown
	}
	var method *FuncInfo
	if structInfo != nil {
		method = structInfo.StaticMethods[sel.Name]
	} else {
		method = enumInfo.StaticMethods[sel.Name]
	}
	if method == nil {
		if structInfo != nil && sel.Name == "fromJson" &&
			structInfo.Methods[sel.Name] == nil &&
			structInfo.Node != nil &&
			hasJSONAnnotation(structInfo.Node.Annotations, "object") {
			c.info.ExprTypes[ident] = Type(ident.Name)
			return FuncOfTypes([]Type{String}, Type(ident.Name))
		}
		c.errorf(sel.Pos, "type %s has no static method %q", ident.Name, sel.Name)
		return Unknown
	}
	if !c.checkPrivateAccess("static method", ident.Name+"::"+sel.Name, method.Private, method.SourcePath, sel.Pos) {
		return Unknown
	}
	c.info.ExprTypes[ident] = Type(ident.Name)
	c.info.ResolvedSelectorFunctions[sel] = method
	return functionType(method)
}

func (c *checker) inferAtExpr(expr *ast.AtExpr) Type {
	if expr.Path != "" {
		if goPath, ok := GoPackageImportPath(expr.Path); ok {
			return GoPackageNamespaceOf(goPath)
		}
		if expr.SourcePath == "" {
			c.errorf(expr.Pos, "unresolved import %q", expr.Path)
			return ImportNamespaceOf(expr.Path)
		}
		return ImportNamespaceOf(expr.SourcePath)
	}
	if expr.Name == "" {
		return Unknown
	}
	if c.info.Stdlib == nil || c.info.Stdlib.Modules[expr.Name] == nil {
		c.errorf(expr.Pos, "unknown module @%s", expr.Name)
		return Unknown
	}
	return ModuleNamespaceOf(expr.Name)
}

func (c *checker) inferNamespaceSelector(sel *ast.SelectorExpr, receiver Type) (Type, bool) {
	if moduleName, ok := ModuleNamespaceName(receiver); ok {
		fn, ok := c.info.Stdlib.Function(moduleName, sel.Name)
		if !ok || fn.Receiver != "" || fn.TopLevelOnly {
			c.errorf(sel.Pos, "unknown module function @%s.%s", moduleName, sel.Name)
			return Unknown, true
		}
		return c.stdlibFunctionValueType(fn), true
	}
	if importPath, ok := ImportNamespacePath(receiver); ok {
		if fn := c.functionInSource(importPath, sel.Name); fn != nil {
			if !c.checkPrivateAccess("function", sel.Name, fn.Private, fn.SourcePath, sel.NamePos) {
				return Unknown, true
			}
			if fn.Node != nil {
				c.inferFunction(fn.Node)
			}
			c.info.ResolvedSelectorFunctions[sel] = fn
			return functionType(fn), true
		}
		if value := c.externalValueInSource(importPath, sel.Name); value != nil {
			c.info.ResolvedSelectorValues[sel] = value
			return value.Type, true
		}
		c.errorf(sel.Pos, "import %q has no member %q", importPath, sel.Name)
		return Unknown, true
	}
	if _, ok := GoPackageNamespacePath(receiver); ok {
		return FuncOfTypes(nil, Unknown), true
	}
	return Unknown, false
}

func (c *checker) stdlibFunctionValueType(fn *stdlib.Function) Type {
	bindings := c.stdlibTypeBindings(fn)
	params := make([]Type, 0, len(fn.Params))
	for _, param := range fn.Params {
		params = append(params, c.resolveDeclaredType(param, bindings))
	}
	ret := c.resolveDeclaredType(fn.Return, bindings)
	if fn.Routine {
		return AsyncFuncOfTypes(params, ret)
	}
	return FuncOfTypes(params, ret)
}

func (c *checker) inferImportedNamespaceCall(importPath string, sel *ast.SelectorExpr, call *ast.CallExpr, argTypes []Type, env map[string]Type) Type {
	fn := c.functionInSource(importPath, sel.Name)
	if fn == nil {
		if c.externalValueInSource(importPath, sel.Name) != nil {
			c.errorf(sel.Pos, "import member %q is not callable", sel.Name)
			return Unknown
		}
		c.errorf(sel.Pos, "import %q has no function %q", importPath, sel.Name)
		return Unknown
	}
	if !c.checkPrivateAccess("function", sel.Name, fn.Private, fn.SourcePath, sel.NamePos) {
		return Unknown
	}
	if fn.Node != nil {
		c.inferFunction(fn.Node)
	}
	c.refineCallArgsFromParams(fn.Params, call.Args, argTypes, env)
	c.info.ResolvedSelectorFunctions[sel] = fn
	c.info.ExprTypes[sel] = functionType(fn)
	c.checkArgs(sel.Name, fn.Params, call.Args, argTypes, sel.Pos)
	return c.finishRoutineCall(call, fn.Routine, fn.Return)
}

func (c *checker) functionInSource(sourcePath string, name string) *FuncInfo {
	for _, fn := range c.info.functionsByName[name] {
		if sameSourcePath(fn.SourcePath, sourcePath) {
			return fn
		}
	}
	return nil
}

func (c *checker) externalValueInSource(sourcePath string, name string) *ExternalValueInfo {
	value := c.info.valuesByName[name]
	if value == nil || !sameSourcePath(value.SourcePath, sourcePath) {
		return nil
	}
	return value
}

func validateRegexFlags(flags string) string {
	seen := map[rune]bool{}
	for _, flag := range flags {
		switch flag {
		case 'd', 'g', 'i', 'm', 's', 'u', 'v', 'y':
		default:
			return "invalid regex flag " + strconv.QuoteRune(flag)
		}
		if seen[flag] {
			return "duplicate regex flag " + strconv.QuoteRune(flag)
		}
		seen[flag] = true
	}
	if seen['u'] && seen['v'] {
		return "regex flags 'u' and 'v' cannot be used together"
	}
	return ""
}

func (c *checker) inferCall(call *ast.CallExpr, env map[string]Type) Type {
	callExpected := c.expectedType
	c.expectedType = Unknown
	argTypes := make([]Type, 0, len(call.Args))
	for _, arg := range call.Args {
		if _, ok := arg.(*ast.LambdaExpr); ok {
			argTypes = append(argTypes, Unknown)
			continue
		}
		argTypes = append(argTypes, c.inferExpr(arg, env))
	}
	c.expectedType = callExpected
	if sel, ok := call.Callee.(*ast.SelectorExpr); ok {
		if typ, ok := c.inferQualifiedEnumConstructor(sel, call, argTypes, env); ok {
			return typ
		}
		if sel.Static {
			ident, ok := sel.Receiver.(*ast.Identifier)
			if !ok {
				c.errorf(sel.Pos, "static selector receiver must be a type")
				return Unknown
			}
			if _, shadowed := env[ident.Name]; shadowed {
				c.errorf(sel.Pos, "static selector receiver %q is a value, not a type", ident.Name)
				return Unknown
			}
			structInfo := c.info.Types[ident.Name]
			enumInfo := c.info.Enums[ident.Name]
			if structInfo == nil && enumInfo == nil {
				c.errorf(sel.Pos, "unknown type %q", ident.Name)
				return Unknown
			}
			var method *FuncInfo
			if structInfo != nil {
				method = structInfo.StaticMethods[sel.Name]
			} else {
				method = enumInfo.StaticMethods[sel.Name]
			}
			if method == nil {
				if structInfo != nil && sel.Name == "fromJson" &&
					structInfo.Methods[sel.Name] == nil &&
					structInfo.Node != nil &&
					hasJSONAnnotation(structInfo.Node.Annotations, "object") {
					params := []ParamInfo{{Name: "text", Type: String}}
					c.info.ExprTypes[ident] = Type(ident.Name)
					c.refineCallArgsFromParams(params, call.Args, argTypes, env)
					c.checkArgs(sel.Name, params, call.Args, argTypes, sel.Pos)
					return Type(ident.Name)
				}
				c.errorf(sel.Pos, "type %s has no static method %q", ident.Name, sel.Name)
				return Unknown
			}
			if !c.checkPrivateAccess("static method", ident.Name+"::"+sel.Name, method.Private, method.SourcePath, sel.Pos) {
				return Unknown
			}
			c.info.ExprTypes[ident] = Type(ident.Name)
			c.info.ExprTypes[sel] = functionType(method)
			c.info.ResolvedSelectorFunctions[sel] = method
			c.refineCallArgsFromParams(method.Params, call.Args, argTypes, env)
			c.checkArgs(sel.Name, method.Params, call.Args, argTypes, sel.Pos)
			return c.finishRoutineCall(call, method.Routine, method.Return)
		}
		if ident, ok := sel.Receiver.(*ast.Identifier); ok {
			if _, shadowed := env[ident.Name]; !shadowed {
				if structInfo := c.info.Types[ident.Name]; structInfo != nil && structInfo.StaticMethods[sel.Name] != nil {
					c.errorf(sel.Pos, "static method %s::%s must be called with '::'", ident.Name, sel.Name)
					return Unknown
				}
				if enumInfo := c.info.Enums[ident.Name]; enumInfo != nil && enumInfo.StaticMethods[sel.Name] != nil {
					c.errorf(sel.Pos, "static method %s::%s must be called with '::'", ident.Name, sel.Name)
					return Unknown
				}
			}
		}
		receiver := c.inferExpr(sel.Receiver, env)
		if moduleName, ok := ModuleNamespaceName(receiver); ok {
			if fn, ok := c.info.Stdlib.Function(moduleName, sel.Name); ok {
				return c.inferStdlibCall(moduleName, sel, call, argTypes, fn, env)
			}
			c.errorf(sel.Pos, "unknown module function @%s.%s", moduleName, sel.Name)
			return Unknown
		}
		if importPath, ok := ImportNamespacePath(receiver); ok {
			return c.inferImportedNamespaceCall(importPath, sel, call, argTypes, env)
		}
		if _, ok := GoPackageNamespacePath(receiver); ok {
			if callExpected != Unknown {
				c.info.ExprTypes[sel] = FuncOfTypes(nil, callExpected)
				return callExpected
			}
			c.info.ExprTypes[sel] = FuncOfTypes(nil, Unknown)
			return Unknown
		}
		if elem, ok := ArrayElement(receiver); ok {
			return c.inferArrayMethodCall(elem, sel, call, argTypes, env)
		}
		if trait := c.traitView(receiver); trait != nil {
			if field, ok := trait.ByName[sel.Name]; ok {
				ret, _ := c.inferFunctionValueCall(
					sel.Name,
					substituteSelfType(field.Type, receiver),
					call.Args,
					argTypes,
					env,
					sel.Pos,
				)
				return ret
			}
			method := trait.Methods[sel.Name]
			if method == nil {
				c.errorf(sel.Pos, "trait %s has no method %q", trait.Name, sel.Name)
				return Unknown
			}
			params := make([]ParamInfo, 0, len(method.Params))
			for _, param := range method.Params {
				params = append(params, ParamInfo{Name: param.Name, Type: substituteSelfType(param.Type, receiver)})
			}
			c.refineCallArgsFromParams(params, call.Args, argTypes, env)
			c.checkArgs(sel.Name, params, call.Args, argTypes, sel.Pos)
			return c.finishRoutineCall(call, method.Routine, substituteSelfType(method.Return, receiver))
		}
		if structInfo := c.info.Types[baseTypeName(receiver)]; structInfo != nil {
			if !c.checkPrivateAccess("type", structInfo.Name, structInfo.Private, structInfo.SourcePath, sel.Pos) {
				return Unknown
			}
			if field, ok := structInfo.ByName[sel.Name]; ok {
				if !c.checkPrivateAccess("field", structInfo.Name+"."+sel.Name, field.Private, field.SourcePath, sel.Pos) {
					return Unknown
				}
				fieldType := structFieldType(structInfo, receiver, field)
				if _, _, ok := parseCallableType(string(fieldType)); ok {
					c.info.ExprTypes[sel] = fieldType
					ret, _ := c.inferFunctionValueCall(sel.Name, fieldType, call.Args, argTypes, env, sel.Pos)
					return ret
				}
			}
		}
		if ret, ok := c.inferStdlibReceiverMethodCall(receiver, sel, call, argTypes, env); ok {
			return ret
		}
		if enumInfo := c.info.Enums[baseTypeName(receiver)]; enumInfo != nil {
			if !c.checkPrivateAccess("enum", enumInfo.Name, enumInfo.Private, enumInfo.SourcePath, sel.Pos) {
				return Unknown
			}
			method := enumInfo.Methods[sel.Name]
			if method == nil {
				c.errorf(sel.Pos, "type %s has no method %q", receiver, sel.Name)
				return Unknown
			}
			if !c.checkPrivateAccess("method", enumInfo.Name+"."+sel.Name, method.Private, method.SourcePath, sel.Pos) {
				return Unknown
			}
			bindings := typeParamBindingsForEnum(enumInfo, receiver)
			params := make([]ParamInfo, 0, len(method.Params))
			for _, param := range method.Params {
				params = append(params, ParamInfo{Name: param.Name, Type: substituteTypeParams(param.Type, bindings)})
			}
			ret := substituteTypeParams(method.Return, bindings)
			c.refineCallArgsFromParams(params, call.Args, argTypes, env)
			c.checkArgs(sel.Name, params, call.Args, argTypes, sel.Pos)
			c.info.ResolvedSelectorFunctions[sel] = method
			return c.finishRoutineCall(call, method.Routine, ret)
		}
		structInfo := c.info.Types[baseTypeName(receiver)]
		if structInfo == nil {
			if receiver != Unknown {
				c.errorf(sel.Pos, "type %s has no methods", receiver)
			}
			return Unknown
		}
		if !c.checkPrivateAccess("type", structInfo.Name, structInfo.Private, structInfo.SourcePath, sel.Pos) {
			return Unknown
		}
		method := structInfo.Methods[sel.Name]
		if field, ok := structInfo.ByName[sel.Name]; ok {
			if !c.checkPrivateAccess("field", structInfo.Name+"."+sel.Name, field.Private, field.SourcePath, sel.Pos) {
				return Unknown
			}
			ret, _ := c.inferFunctionValueCall(sel.Name, structFieldType(structInfo, receiver, field), call.Args, argTypes, env, sel.Pos)
			return ret
		}
		if method == nil {
			c.errorf(sel.Pos, "type %s has no method %q", receiver, sel.Name)
			return Unknown
		}
		if !c.checkPrivateAccess("method", structInfo.Name+"."+sel.Name, method.Private, method.SourcePath, sel.Pos) {
			return Unknown
		}
		c.refineCallArgsFromParams(method.Params, call.Args, argTypes, env)
		c.checkArgs(sel.Name, method.Params, call.Args, argTypes, sel.Pos)
		return c.finishRoutineCall(call, method.Routine, method.Return)
	}
	if ident, ok := call.Callee.(*ast.Identifier); ok {
		if typ, ok := c.inferEnumConstructor(ident.Name, call, argTypes); ok {
			return typ
		}
		if ident.Name == "Ok" || ident.Name == "Err" {
			c.inferDeferredCallArgs(call.Args, argTypes, env)
			return c.inferResultConstructor(ident.Name, call, argTypes)
		}
		if localType, ok := env[ident.Name]; ok {
			ret, refined := c.inferFunctionValueCall(ident.Name, localType, call.Args, argTypes, env, ident.Pos)
			if refined != Unknown && refined != localType {
				env[ident.Name] = refined
				if binding := c.bindings[ident.Name]; binding != nil {
					c.info.ExprTypes[binding] = refined
					c.applyExpectedType(binding, refined)
				}
				localType = refined
			}
			c.info.ExprTypes[ident] = localType
			return ret
		}
		fn, ok := c.resolveFunction(ident.Name, ident.Pos)
		if !ok {
			c.errorf(ident.Pos, "undefined function %q", ident.Name)
			return Unknown
		}
		if fn == nil {
			return Unknown
		}
		if fn.Macro {
			c.errorf(ident.Pos, "%s is a macro and can only be used with '#'", ident.Name)
			return Unknown
		}
		if fn.Node != nil {
			c.inferFunction(fn.Node)
		}
		bindings := c.genericFunctionBindings(fn, argTypes)
		c.checkGenericCallConstraints(fn, bindings, ident.Pos)
		params := substituteParamInfos(fn.Params, bindings)
		c.refineCallArgsFromParams(params, call.Args, argTypes, env)
		c.info.ResolvedFunctions[ident] = fn
		c.info.ExprTypes[ident] = functionType(fn)
		c.checkArgs(ident.Name, params, call.Args, argTypes, ident.Pos)
		return c.finishRoutineCall(call, fn.Routine, substituteTypeParams(fn.Return, bindings))
	}
	calleeType := c.inferExpr(call.Callee, env)
	ret, refined := c.inferFunctionValueCall("<expr>", calleeType, call.Args, argTypes, env, call.Callee.Position())
	if refined != Unknown && refined != calleeType {
		c.info.ExprTypes[call.Callee] = refined
		c.applyExpectedType(call.Callee, refined)
	}
	return ret
}

func (c *checker) inferQualifiedEnumConstructor(sel *ast.SelectorExpr, call *ast.CallExpr, argTypes []Type, env map[string]Type) (Type, bool) {
	ident, ok := sel.Receiver.(*ast.Identifier)
	if !ok || sel.Static {
		return Unknown, false
	}
	if _, shadowed := env[ident.Name]; shadowed {
		return Unknown, false
	}
	enum := c.info.Enums[ident.Name]
	if enum == nil {
		return Unknown, false
	}
	member, ok := enum.ByName[sel.Name]
	if !ok || len(member.Params) == 0 {
		return Unknown, false
	}
	if !c.checkPrivateAccess("enum", enum.Name, enum.Private, enum.SourcePath, sel.Pos) {
		return Unknown, true
	}
	if !c.checkPrivateAccess("enum constructor", enum.Name+"."+sel.Name, member.Private, member.SourcePath, sel.NamePos) {
		return Unknown, true
	}
	if len(member.Params) != len(call.Args) {
		c.errorf(call.Pos, "constructor %q expects %d args, got %d", sel.Name, len(member.Params), len(call.Args))
		return Unknown, true
	}
	bindings := enumConstructorBindings(enum.Generics, member.Params, argTypes)
	for idx, param := range member.Params {
		expected := substituteTypeParams(param.Type, bindings)
		if idx < len(argTypes) && !c.typesCompatible(expected, argTypes[idx], nil) {
			c.errorf(call.Args[idx].Position(), "argument %d to %q has type %s, expected %s", idx+1, sel.Name, argTypes[idx], expected)
		}
	}
	c.info.ExprTypes[ident] = Type(enum.Name)
	if len(enum.Generics) == 0 {
		return Type(enum.Name), true
	}
	args := make([]Type, 0, len(enum.Generics))
	for _, generic := range enum.Generics {
		if typ, ok := bindings[generic]; ok {
			args = append(args, typ)
		} else {
			args = append(args, Unknown)
		}
	}
	return genericTypeOf(enum.Name, args), true
}

func (c *checker) traitInfo(typ Type) *TraitInfo {
	name := string(typ)
	if !strings.HasPrefix(name, "&") {
		return nil
	}
	return c.info.Traits[strings.TrimPrefix(name, "&")]
}

func (c *checker) traitView(typ Type) *TraitInfo {
	if trait := c.traitInfo(typ); trait != nil {
		return trait
	}
	if c.genericTypes == nil || !c.genericTypes[string(typ)] {
		return nil
	}
	constraint := c.genericConstraints[string(typ)]
	if constraint == "" || c.info == nil {
		return nil
	}
	return c.info.Traits[constraint]
}

func functionType(fn *FuncInfo) Type {
	if fn.Routine {
		return AsyncFuncOfTypes(paramTypes(fn.Params), fn.Return)
	}
	return FuncOfTypes(paramTypes(fn.Params), fn.Return)
}

func (c *checker) genericFunctionBindings(fn *FuncInfo, argTypes []Type) map[string]Type {
	if len(fn.Generics) == 0 {
		return nil
	}
	bindings := make(map[string]Type, len(fn.Generics))
	generics := genericSet(fn.Generics...)
	for _, name := range fn.Generics {
		bindings[name] = Unknown
	}
	limit := min(len(fn.Params), len(argTypes))
	for i := 0; i < limit; i++ {
		bindTypeParams(fn.Params[i].Type, argTypes[i], generics, bindings)
	}
	return bindings
}

func (c *checker) checkGenericCallConstraints(fn *FuncInfo, bindings map[string]Type, pos lexer.Position) {
	for name, constraint := range fn.GenericConstraints {
		actual := bindings[name]
		if actual == "" || actual == Unknown {
			continue
		}
		if !c.genericConstraintSatisfied(actual, constraint) {
			c.errorf(pos, "type %s does not satisfy generic constraint %s", actual, constraint)
		}
	}
}

func substituteParamInfos(params []ParamInfo, bindings map[string]Type) []ParamInfo {
	if len(bindings) == 0 {
		return params
	}
	out := make([]ParamInfo, 0, len(params))
	for _, param := range params {
		out = append(out, ParamInfo{Name: param.Name, Type: substituteTypeParams(param.Type, bindings)})
	}
	return out
}

func (c *checker) inferResultConstructor(name string, call *ast.CallExpr, argTypes []Type) Type {
	if len(call.Args) != 1 {
		c.errorf(call.Pos, "%s expects 1 arg, got %d", name, len(call.Args))
		return Unknown
	}
	switch name {
	case "Ok":
		return ResultOf(argTypes[0], Unknown)
	case "Err":
		return ResultOf(Unknown, argTypes[0])
	default:
		return Unknown
	}
}

func (c *checker) finishRoutineCall(call *ast.CallExpr, routine bool, ret Type) Type {
	if !routine {
		return ret
	}
	if call != nil {
		c.info.AsyncCalls[call] = true
	}
	if c.routineDepth > 0 {
		if call != nil {
			c.info.AwaitCalls[call] = true
		}
		return ret
	}
	return TaskOf(ret)
}

func (c *checker) refineCallArgsFromParams(params []ParamInfo, args []ast.Expr, argTypes []Type, env map[string]Type) {
	limit := min(len(params), len(argTypes))
	for i := 0; i < limit; i++ {
		expected := params[i].Type
		if _, ok := args[i].(*ast.LambdaExpr); ok {
			argTypes[i] = c.inferLambdaArg(args[i], expected, env)
			continue
		}
		if expected == "" || expected == Unknown {
			continue
		}
		argTypes[i] = c.refineUnknownIdentifierType(args[i], argTypes[i], expected, env)
		if shouldApplyExpectedType(argTypes[i], expected) {
			refined, _ := c.refineArgumentType(expected, argTypes[i])
			argTypes[i] = refined
			c.applyExpectedType(args[i], refined)
		}
	}
}

func (c *checker) inferDeferredCallArgs(args []ast.Expr, argTypes []Type, env map[string]Type) {
	limit := min(len(args), len(argTypes))
	for i := 0; i < limit; i++ {
		if _, ok := args[i].(*ast.LambdaExpr); !ok || argTypes[i] != Unknown {
			continue
		}
		argTypes[i] = c.inferLambdaArg(args[i], Unknown, env)
	}
}

func (c *checker) inferLambdaArg(arg ast.Expr, expected Type, env map[string]Type) Type {
	if expected != "" && expected != Unknown {
		c.applyExpectedType(arg, expected)
	}
	return c.inferExpr(arg, env)
}

func (c *checker) inferEnumConstructor(name string, call *ast.CallExpr, argTypes []Type) (Type, bool) {
	constructors := c.info.Constructors[name]
	if len(constructors) == 0 {
		return Unknown, false
	}
	constructor := constructors[0]
	if len(constructors) > 1 {
		c.errorf(call.Pos, "ambiguous enum constructor %q", name)
		return Unknown, true
	}
	if !c.checkPrivateAccess("enum", constructor.Enum.Name, constructor.Enum.Private, constructor.Enum.SourcePath, call.Pos) {
		return Unknown, true
	}
	if !c.checkPrivateAccess("enum constructor", constructor.Enum.Name+"."+name, constructor.Member.Private, constructor.Member.SourcePath, call.Pos) {
		return Unknown, true
	}
	if len(constructor.Member.Params) != len(call.Args) {
		c.errorf(call.Pos, "constructor %q expects %d args, got %d", name, len(constructor.Member.Params), len(call.Args))
		return Unknown, true
	}
	bindings := enumConstructorBindings(constructor.Enum.Generics, constructor.Member.Params, argTypes)
	for idx, param := range constructor.Member.Params {
		expected := substituteTypeParams(param.Type, bindings)
		if idx < len(argTypes) && !c.typesCompatible(expected, argTypes[idx], nil) {
			c.errorf(call.Args[idx].Position(), "argument %d to %q has type %s, expected %s", idx+1, name, argTypes[idx], expected)
		}
	}
	if len(constructor.Enum.Generics) == 0 {
		return Type(constructor.Enum.Name), true
	}
	args := make([]Type, 0, len(constructor.Enum.Generics))
	for _, generic := range constructor.Enum.Generics {
		if typ, ok := bindings[generic]; ok {
			args = append(args, typ)
		} else {
			args = append(args, Unknown)
		}
	}
	return genericTypeOf(constructor.Enum.Name, args), true
}

func enumConstructorBindings(generics []string, params []ParamInfo, argTypes []Type) map[string]Type {
	bindings := make(map[string]Type, len(generics))
	genericNames := genericSet(generics...)
	for _, generic := range generics {
		bindings[generic] = Unknown
	}
	limit := min(len(params), len(argTypes))
	for i := 0; i < limit; i++ {
		bindTypeParams(params[i].Type, argTypes[i], genericNames, bindings)
	}
	return bindings
}

func bindTypeParams(pattern Type, actual Type, generics map[string]bool, bindings map[string]Type) {
	if actual == Unknown || pattern == Unknown {
		return
	}
	if generics[string(pattern)] {
		if current := bindings[string(pattern)]; current == "" || current == Unknown {
			bindings[string(pattern)] = actual
		}
		return
	}
	if inner, ok := parseNullableType(string(pattern)); ok {
		if actualInner, actualNullable := parseNullableType(string(actual)); actualNullable {
			bindTypeParams(Type(inner), Type(actualInner), generics, bindings)
		}
		return
	}
	if elem, ok := parseArrayType(string(pattern)); ok {
		if actualElem, actualArray := parseArrayType(string(actual)); actualArray {
			bindTypeParams(Type(elem), Type(actualElem), generics, bindings)
		}
		return
	}
	if base, args, ok := parseGenericType(string(pattern)); ok {
		actualBase, actualArgs, actualGeneric := parseGenericType(string(actual))
		if !actualGeneric || base != actualBase || len(args) != len(actualArgs) {
			return
		}
		for i := range args {
			bindTypeParams(Type(args[i]), Type(actualArgs[i]), generics, bindings)
		}
		return
	}
	if params, ret, ok := parseFuncType(string(pattern)); ok {
		actualParams, actualRet, actualFunc := parseFuncType(string(actual))
		if !actualFunc || len(params) != len(actualParams) {
			return
		}
		for i := range params {
			bindTypeParams(Type(params[i]), Type(actualParams[i]), generics, bindings)
		}
		bindTypeParams(Type(ret), Type(actualRet), generics, bindings)
	}
}

func (c *checker) refineNumericBinaryOperands(expr *ast.BinaryExpr, left Type, right Type, env map[string]Type) (Type, Type) {
	if left == Unknown && c.isArithmeticLikeType(right, expr.Op) {
		left = c.refineUnknownIdentifierType(expr.Left, left, right, env)
	}
	if right == Unknown && c.isArithmeticLikeType(left, expr.Op) {
		right = c.refineUnknownIdentifierType(expr.Right, right, left, env)
	}
	return left, right
}

func (c *checker) refineUnknownIdentifierType(expr ast.Expr, actual Type, expected Type, env map[string]Type) Type {
	if actual != Unknown || expected == Unknown {
		return actual
	}
	ident, ok := expr.(*ast.Identifier)
	if !ok {
		return actual
	}
	current, ok := env[ident.Name]
	if !ok || current != Unknown {
		return actual
	}
	env[ident.Name] = expected
	c.info.ExprTypes[ident] = expected
	return expected
}

func (c *checker) numericBinaryType(expr *ast.BinaryExpr, left Type, right Type) Type {
	if left == Unknown && right == Unknown {
		return Int
	}
	if left == Unknown && c.isArithmeticLikeType(right, expr.Op) {
		return right
	}
	if right == Unknown && c.isArithmeticLikeType(left, expr.Op) {
		return left
	}
	leftNumeric := c.isArithmeticLikeType(left, expr.Op)
	rightNumeric := c.isArithmeticLikeType(right, expr.Op)
	if !leftNumeric {
		c.errorf(expr.Left.Position(), "arithmetic expects numeric operands, got %s", left)
	}
	if !rightNumeric {
		c.errorf(expr.Right.Position(), "arithmetic expects numeric operands, got %s", right)
	}
	if leftNumeric && rightNumeric && left != right {
		c.errorf(expr.Pos, "arithmetic requires matching numeric types, got %s and %s", left, right)
		return Unknown
	}
	if expr.Op == lexer.Percent && (c.isGenericArithmeticType(left) || c.isGenericArithmeticType(right)) {
		c.errorf(expr.Pos, "operator '%%' cannot be used with generic arithmetic operands")
		return Unknown
	}
	if expr.Op == lexer.Percent && isFloatType(left) && isFloatType(right) {
		c.errorf(expr.Pos, "operator '%%' expects integer operands, got %s", left)
		return Unknown
	}
	if leftNumeric {
		return left
	}
	return Unknown
}

func (c *checker) isNumericLikeType(typ Type) bool {
	return isNumericType(typ) || c.isGenericArithmeticType(typ)
}

func (c *checker) isArithmeticLikeType(typ Type, op lexer.Kind) bool {
	if isNumericType(typ) {
		return true
	}
	if op == lexer.Percent {
		return false
	}
	return c.genericSupportsBinaryOp(typ, op)
}

func (c *checker) isGenericArithmeticType(typ Type) bool {
	if c.genericTypes == nil || !c.genericTypes[string(typ)] {
		return false
	}
	constraint := c.genericConstraints[string(typ)]
	if constraint == "" || c.info == nil {
		return false
	}
	trait := c.info.Traits[constraint]
	return traitHasSelfBinaryMethod(trait, "add") ||
		traitHasSelfBinaryMethod(trait, "sub") ||
		traitHasSelfBinaryMethod(trait, "mul") ||
		traitHasSelfBinaryMethod(trait, "div")
}

func (c *checker) genericSupportsBinaryOp(typ Type, op lexer.Kind) bool {
	if c.genericTypes == nil || !c.genericTypes[string(typ)] {
		return false
	}
	method := arithmeticTraitMethod(op)
	if method == "" {
		return false
	}
	constraint := c.genericConstraints[string(typ)]
	if constraint == "" || c.info == nil {
		return false
	}
	return traitHasSelfBinaryMethod(c.info.Traits[constraint], method)
}

func arithmeticTraitMethod(op lexer.Kind) string {
	switch op {
	case lexer.Plus:
		return "add"
	case lexer.Minus:
		return "sub"
	case lexer.Star:
		return "mul"
	case lexer.Slash:
		return "div"
	default:
		return ""
	}
}

func traitHasSelfBinaryMethod(trait *TraitInfo, name string) bool {
	if trait == nil {
		return false
	}
	method := trait.Methods[name]
	if method == nil || method.Routine || len(method.Params) != 1 {
		return false
	}
	return method.Params[0].Type == Type("Self") && method.Return == Type("Self")
}

func (c *checker) bitwiseBinaryType(expr *ast.BinaryExpr, left Type, right Type) Type {
	if left == Unknown && right == Unknown {
		return Int
	}
	if left == Unknown && isBitwiseType(right) {
		return right
	}
	if right == Unknown && isBitwiseType(left) {
		return left
	}
	if !isBitwiseType(left) {
		c.errorf(expr.Left.Position(), "bitwise operator expects integer operands, got %s", left)
	}
	if !isBitwiseType(right) {
		c.errorf(expr.Right.Position(), "bitwise operator expects integer operands, got %s", right)
	}
	if expr.Op == lexer.UnsignedShiftRight && !isUnsignedIntegerType(left) && left != Unknown {
		c.errorf(expr.Left.Position(), "operator '>>>' expects an unsigned integer left operand, got %s", left)
	}
	if isBitwiseType(left) && isBitwiseType(right) && left != right {
		c.errorf(expr.Pos, "bitwise operator requires matching integer types, got %s and %s", left, right)
		return Unknown
	}
	if isBitwiseType(left) {
		return left
	}
	return Unknown
}

func isNumericType(typ Type) bool {
	return isIntegerType(typ) || isFloatType(typ) || typ == BigInt
}

func isIntegerType(typ Type) bool {
	return isSignedIntegerType(typ) || isUnsignedIntegerType(typ)
}

func isSignedIntegerType(typ Type) bool {
	switch typ {
	case Int, Int4, Int8, Int16, Int64:
		return true
	default:
		return false
	}
}

func isUnsignedIntegerType(typ Type) bool {
	switch typ {
	case UInt, UInt8, UInt16, UInt64:
		return true
	default:
		return false
	}
}

func isFloatType(typ Type) bool {
	return typ == Float || typ == Double
}

func isBitwiseType(typ Type) bool {
	return isIntegerType(typ) || typ == BigInt
}

func orderedComparisonType(typ Type) bool {
	return isNumericType(typ) || typ == String || typ == Char
}

func typesComparable(left Type, right Type) bool {
	if left == Unknown || right == Unknown || left == right {
		return true
	}
	if left == Never || right == Never {
		return true
	}
	if leftInner, ok := parseNullableType(string(left)); ok {
		return right == Null || typesComparable(Type(leftInner), right)
	}
	if rightInner, ok := parseNullableType(string(right)); ok {
		return left == Null || typesComparable(left, Type(rightInner))
	}
	return false
}

func (c *checker) inferLambda(lambda *ast.LambdaExpr, env map[string]Type) Type {
	local := cloneEnv(env)
	params := make([]Type, 0, len(lambda.Params))
	inferredFields := inferParamFields(lambda.Body, lambda.Params)
	var expectedParams []string
	if expected := c.info.ExprTypes[lambda]; expected != "" {
		if parsed, _, ok := parseFuncType(string(expected)); ok {
			expectedParams = parsed
		}
	}
	for i, name := range lambda.Params {
		paramType := Unknown
		if i < len(lambda.ParamTypes) && !lambda.ParamTypes[i].IsZero() {
			paramName := lambda.ParamTypes[i].Canonical()
			paramType = c.resolveTypeWithGenerics(paramName, c.genericTypes)
			if paramType == Unknown && !isDynamicTypeName(paramName) {
				c.reportUnknownOrPrivateType(lambda.Pos, paramName)
			}
		} else if i < len(expectedParams) {
			paramType = Type(expectedParams[i])
		} else if fields := inferredFields[name]; len(fields) > 0 {
			paramType = c.objectTypeFromFields(fields)
		}
		params = append(params, paramType)
		local[name] = paramType
	}
	ret := c.inferExpr(lambda.Body, local)
	c.finishInferredLambdaParams(params, lambda.Params, lambda.Body, local)
	returnName := lambda.ReturnType.Canonical()
	if returnName != "" {
		declared := c.resolveTypeWithGenerics(returnName, c.genericTypes)
		if declared == Unknown && !isDynamicTypeName(returnName) {
			if privateName, ok := c.inaccessibleTypeName(returnName); ok {
				c.errorf(lambda.Pos, "return type %q is private", privateName)
			} else {
				c.errorf(lambda.Pos, "unknown return type %q", returnName)
			}
		}
		if !c.typesCompatible(declared, ret, nil) {
			c.errorf(lambda.Body.Position(), "lambda returns %s, expected %s", ret, declared)
		}
		return FuncOfTypes(params, declared)
	}
	if ret == Unknown {
		ret = Void
	}
	return FuncOfTypes(params, ret)
}

func (c *checker) finishInferredLambdaParams(params []Type, names []string, body ast.Expr, env map[string]Type) {
	for idx, name := range names {
		if idx >= len(params) || params[idx] != Unknown {
			continue
		}
		if inferred := env[name]; inferred != "" && inferred != Unknown {
			params[idx] = inferred
			continue
		}
		if inferred := c.inferredParamUseType(body, name, ""); inferred != Unknown {
			params[idx] = inferred
		}
	}
}

func paramNames(params []ParamInfo) []string {
	out := make([]string, 0, len(params))
	for _, param := range params {
		out = append(out, param.Name)
	}
	return out
}

func (c *checker) objectTypeFromFields(fields map[string]FieldInfo) Type {
	if len(fields) == 0 {
		return Unknown
	}
	var infos []FieldInfo
	names := make([]string, 0, len(fields))
	for fieldName := range fields {
		names = append(names, fieldName)
	}
	sort.Strings(names)
	for _, fieldName := range names {
		field := fields[fieldName]
		infos = append(infos, field)
	}
	typ := ObjectOf(infos)
	byName := map[string]FieldInfo{}
	for _, field := range infos {
		byName[field.Name] = field
	}
	c.registerAnonymousObjectType(typ, infos, byName)
	return typ
}

func inferParamFields(body ast.Expr, names []string) map[string]map[string]FieldInfo {
	params := map[string]bool{}
	for _, name := range names {
		params[name] = true
	}
	out := map[string]map[string]FieldInfo{}
	var walk func(ast.Expr, Type)
	walk = func(expr ast.Expr, expected Type) {
		switch e := expr.(type) {
		case *ast.SelectorExpr:
			if ident, ok := e.Receiver.(*ast.Identifier); ok && params[ident.Name] {
				fields := out[ident.Name]
				if fields == nil {
					fields = map[string]FieldInfo{}
					out[ident.Name] = fields
				}
				if expected == Unknown {
					expected = Int
				}
				fields[e.Name] = FieldInfo{Name: e.Name, Type: expected}
				return
			}
			walk(e.Receiver, Unknown)
		case *ast.BinaryExpr:
			switch e.Op {
			case lexer.Plus, lexer.Minus, lexer.Star, lexer.Slash, lexer.Percent, lexer.BitAnd, lexer.BitOr, lexer.BitXor, lexer.ShiftLeft, lexer.ShiftRight, lexer.UnsignedShiftRight:
				walk(e.Left, Int)
				walk(e.Right, Int)
			case lexer.AndAnd, lexer.OrOr:
				walk(e.Left, Bool)
				walk(e.Right, Bool)
			default:
				walk(e.Left, Unknown)
				walk(e.Right, Unknown)
			}
		case *ast.TernaryExpr:
			walk(e.Condition, Bool)
			walk(e.Consequence, expected)
			if e.Alternative != nil {
				walk(e.Alternative, expected)
			}
		case *ast.UnaryExpr:
			walk(e.Expr, expected)
		case *ast.PostfixExpr:
			walk(e.Expr, expected)
		case *ast.CallExpr:
			walk(e.Callee, Unknown)
			for _, arg := range e.Args {
				walk(arg, Unknown)
			}
		case *ast.AnonymousObjectLiteral:
			for _, field := range e.Fields {
				walk(field.Value, Unknown)
			}
		case *ast.ReactiveLiteral:
			walk(e.Value, Unknown)
		case *ast.XMLElement:
			for _, attr := range e.Attrs {
				walk(attr.Value, Unknown)
			}
			for _, child := range e.Children {
				walk(child.Expr, Unknown)
			}
		case *ast.BlockExpr:
			for _, stmt := range e.Statements {
				switch s := stmt.(type) {
				case *ast.LetStmt:
					walk(s.Value, Unknown)
				case *ast.ObjectDestructureStmt:
					walk(s.Value, Unknown)
				case *ast.AssignStmt:
					walk(s.Value, Unknown)
				case *ast.ExprStmt:
					walk(s.Expr, Unknown)
				}
			}
		case *ast.MatchExpr:
			walk(e.Subject, Unknown)
			for _, branch := range e.Branches {
				walk(branch.Expr, Unknown)
			}
		}
	}
	walk(body, Unknown)
	return out
}

func (c *checker) inferFunctionValueCall(name string, typ Type, args []ast.Expr, argTypes []Type, env map[string]Type, pos lexer.Position) (Type, Type) {
	params, ret, ok := parseCallableType(string(typ))
	if !ok {
		if typ != Unknown {
			c.errorf(pos, "type %s is not callable", typ)
		}
		return Unknown, typ
	}
	if len(params) != len(args) {
		c.errorf(pos, "function %q expects %d args, got %d", name, len(params), len(args))
	}
	limit := min(len(params), len(argTypes))
	refinedParams := make([]Type, 0, len(params))
	for _, param := range params {
		refinedParams = append(refinedParams, Type(param))
	}
	for i := 0; i < limit; i++ {
		expected := Type(params[i])
		if _, ok := args[i].(*ast.LambdaExpr); ok {
			argTypes[i] = c.inferLambdaArg(args[i], expected, env)
		}
		if refined, ok := c.refineArgumentType(expected, argTypes[i]); ok {
			refinedParams[i] = refined
			if shouldApplyExpectedType(argTypes[i], expected) {
				argTypes[i] = refined
				c.applyExpectedType(args[i], refined)
			}
			continue
		}
		if !c.typesCompatible(expected, argTypes[i], nil) {
			c.errorf(args[i].Position(), "argument %d to %q has type %s, expected %s", i+1, name, argTypes[i], expected)
		}
	}
	if _, _, async := parseAsyncFuncType(string(typ)); async {
		return c.finishRoutineCall(nil, true, Type(ret)), AsyncFuncOfTypes(refinedParams, Type(ret))
	}
	return Type(ret), FuncOfTypes(refinedParams, Type(ret))
}

func shouldApplyExpectedType(actual Type, expected Type) bool {
	if expected == Unknown {
		return false
	}
	if actualElem, ok := ArrayElement(actual); ok && actualElem == Unknown {
		_, expectedArray := ArrayElement(expected)
		return expectedArray
	}
	return false
}

func (c *checker) refineArgumentType(expected Type, actual Type) (Type, bool) {
	if expected == Unknown || actual == Unknown {
		return expected, true
	}
	if unified, ok := c.unifyTypes(expected, actual); ok {
		return unified, true
	}
	if c.objectHasFields(actual, expected) {
		if isObjectType(expected) {
			return actual, true
		}
		return expected, true
	}
	return Unknown, false
}

func paramTypes(params []ParamInfo) []Type {
	out := make([]Type, 0, len(params))
	for _, param := range params {
		out = append(out, param.Type)
	}
	return out
}

func isIntrinsicStub(fn *ast.Function) bool {
	lit, ok := fn.Body.(*ast.StringLiteral)
	return ok && strings.HasPrefix(lit.Value, "%")
}
