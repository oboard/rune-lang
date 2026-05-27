package checker

import (
	"sort"
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

func (c *checker) inferMethods(typ *ast.StructType) {
	structInfo := c.info.Types[typ.Name]
	if structInfo == nil {
		return
	}
	c.withSourcePath(typ.SourcePath, func() {
		for _, method := range typ.Methods {
			info := structInfo.Methods[method.Name]
			if info == nil {
				continue
			}
			c.withSourcePath(method.SourcePath, func() {
				c.withGenericTypes(info.Generics, func() {
					if isIntrinsicStub(method) {
						if !info.ReturnDeclared {
							info.Return = Void
						}
						return
					}
					env := map[string]Type{"this": info.ReceiverType}
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
		c.withGenericTypes(info.Generics, func() {
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
			if fn.Name == "main" {
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

func (c *checker) withGenericTypes(names []string, fn func()) {
	prev := c.genericTypes
	if len(names) > 0 {
		c.genericTypes = genericSet(names...)
	} else {
		c.genericTypes = nil
	}
	fn()
	c.genericTypes = prev
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
		if inferred := c.inferredParamUseType(body, param.Name); inferred != Unknown {
			info.Params[idx].Type = inferred
		}
	}
}

func (c *checker) inferFunctionBody(fn *ast.Function, info *FuncInfo, env map[string]Type) Type {
	if info != nil && info.Routine {
		c.routineDepth++
		c.unwrapErrors = append(c.unwrapErrors, Unknown)
	}
	ret := c.inferExpr(fn.Body, env)
	if info == nil || !info.Routine {
		c.unwrapErrors = append(c.unwrapErrors, Unknown)
	}
	return ret
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

func (c *checker) inferredParamUseType(body ast.Expr, name string) Type {
	result := Unknown
	ast.WalkExpr(body, func(expr ast.Expr) {
		if result != Unknown {
			return
		}
		if ident, ok := expr.(*ast.Identifier); ok && ident.Name == name {
			if typ := c.info.ExprTypes[ident]; typ != "" && typ != Unknown {
				result = typ
			}
		}
	})
	return result
}

func (c *checker) finishFunctionReturn(info *FuncInfo, ret Type, unwrapErr Type, fn *ast.Function) {
	if info.Routine && unwrapErr != Unknown {
		if _, _, ok := parseResultType(ret); !ok {
			ret = ResultOf(ret, unwrapErr)
		}
	}
	if info.ReturnDeclared {
		if !typesCompatible(info.Return, ret, info.Generics) {
			c.errorf(fn.Body.Position(), "function %q returns %s, expected %s", fn.Name, ret, info.Return)
		}
		return
	}
	if ret == Unknown {
		ret = Void
	}
	info.Return = ret
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
		if e.Name != "<error>" {
			c.errorf(e.Pos, "undefined name %q", e.Name)
		}
		return Unknown
	case *ast.AtExpr:
		return Unknown
	case *ast.ThisExpr:
		typ, ok := env["this"]
		if !ok {
			c.errorf(e.Pos, "implicit this selector can only be used inside a method")
			return Unknown
		}
		return typ
	case *ast.SelectorExpr:
		if typ, ok := c.inferEnumMemberSelector(e, env); ok {
			return typ
		}
		receiver := c.inferExpr(e.Receiver, env)
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
	case *ast.BinaryExpr:
		left := c.inferExpr(e.Left, env)
		right := c.inferExpr(e.Right, env)
		switch e.Op {
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
		default:
			return Unknown
		}
	case *ast.TernaryExpr:
		condition := c.inferExpr(e.Condition, env)
		if condition != Bool && condition != Unknown {
			c.errorf(e.Condition.Position(), "ternary condition expects Bool, got %s", condition)
		}
		consequence := c.inferExpr(e.Consequence, env)
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
		if _, exists := env[e.Name]; !exists {
			c.errorf(e.Pos, "cannot assign undefined name %q", e.Name)
		}
		if e.Target != nil {
			c.inferExpr(e.Target, env)
		}
		c.inferExpr(e.Value, env)
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
	argTypes := make([]Type, 0, len(call.Args))
	for _, arg := range call.Args {
		argTypes = append(argTypes, c.inferExpr(arg, env))
	}
	if sel, ok := call.Callee.(*ast.SelectorExpr); ok {
		if at, ok := sel.Receiver.(*ast.AtExpr); ok {
			if fn, ok := c.info.Stdlib.Function(at.Name, sel.Name); ok {
				return c.inferStdlibCall(at.Name, sel, call, argTypes, fn, env)
			}
			c.errorf(sel.Pos, "unknown module function @%s.%s", at.Name, sel.Name)
			return Unknown
		}
		receiver := c.inferExpr(sel.Receiver, env)
		if elem, ok := ArrayElement(receiver); ok {
			return c.inferArrayMethodCall(elem, sel, call, argTypes, env)
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
					ret, _ := c.inferFunctionValueCall(sel.Name, fieldType, call.Args, argTypes, sel.Pos)
					return ret
				}
			}
		}
		if ret, ok := c.inferStdlibReceiverMethodCall(receiver, sel, call, argTypes, env); ok {
			return ret
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
			ret, _ := c.inferFunctionValueCall(sel.Name, structFieldType(structInfo, receiver, field), call.Args, argTypes, sel.Pos)
			return ret
		}
		if method == nil {
			c.errorf(sel.Pos, "type %s has no method %q", receiver, sel.Name)
			return Unknown
		}
		if !c.checkPrivateAccess("method", structInfo.Name+"."+sel.Name, method.Private, method.SourcePath, sel.Pos) {
			return Unknown
		}
		c.checkArgs(sel.Name, method.Params, call.Args, argTypes, sel.Pos)
		return c.finishRoutineCall(call, method.Routine, method.Return)
	}
	if ident, ok := call.Callee.(*ast.Identifier); ok {
		if ident.Name == "Ok" || ident.Name == "Err" {
			return c.inferResultConstructor(ident.Name, call, argTypes)
		}
		if localType, ok := env[ident.Name]; ok {
			ret, refined := c.inferFunctionValueCall(ident.Name, localType, call.Args, argTypes, ident.Pos)
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
		if fn.Node != nil {
			c.inferFunction(fn.Node)
		}
		for i, argType := range argTypes {
			if i < len(fn.Params) && fn.Params[i].Type == Unknown && argType != Unknown {
				fn.Params[i].Type = argType
			}
		}
		c.refineCallArgsFromParams(fn.Params, call.Args, argTypes, env)
		c.info.ResolvedFunctions[ident] = fn
		c.info.ExprTypes[ident] = functionType(fn)
		c.checkArgs(ident.Name, fn.Params, call.Args, argTypes, ident.Pos)
		return c.finishRoutineCall(call, fn.Routine, fn.Return)
	}
	calleeType := c.inferExpr(call.Callee, env)
	ret, refined := c.inferFunctionValueCall("<expr>", calleeType, call.Args, argTypes, call.Callee.Position())
	if refined != Unknown && refined != calleeType {
		c.info.ExprTypes[call.Callee] = refined
		c.applyExpectedType(call.Callee, refined)
	}
	return ret
}

func functionType(fn *FuncInfo) Type {
	if fn.Routine {
		return AsyncFuncOfTypes(paramTypes(fn.Params), fn.Return)
	}
	return FuncOfTypes(paramTypes(fn.Params), fn.Return)
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

func (c *checker) refineNumericBinaryOperands(expr *ast.BinaryExpr, left Type, right Type, env map[string]Type) (Type, Type) {
	if left == Unknown && isNumericType(right) {
		left = c.refineUnknownIdentifierType(expr.Left, left, right, env)
	}
	if right == Unknown && isNumericType(left) {
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
	if left == Unknown && isNumericType(right) {
		return right
	}
	if right == Unknown && isNumericType(left) {
		return left
	}
	if !isNumericType(left) {
		c.errorf(expr.Left.Position(), "arithmetic expects numeric operands, got %s", left)
	}
	if !isNumericType(right) {
		c.errorf(expr.Right.Position(), "arithmetic expects numeric operands, got %s", right)
	}
	if isNumericType(left) && isNumericType(right) && left != right {
		c.errorf(expr.Pos, "arithmetic requires matching numeric types, got %s and %s", left, right)
		return Unknown
	}
	if expr.Op == lexer.Percent && isFloatType(left) && isFloatType(right) {
		c.errorf(expr.Pos, "operator '%%' expects integer operands, got %s", left)
		return Unknown
	}
	if isNumericType(left) {
		return left
	}
	return Unknown
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
		if i < len(lambda.ParamTypes) && lambda.ParamTypes[i] != "" {
			paramType = c.resolveTypeWithGenerics(lambda.ParamTypes[i], c.genericTypes)
			if paramType == Unknown && !isDynamicTypeName(lambda.ParamTypes[i]) {
				c.reportUnknownOrPrivateType(lambda.Pos, lambda.ParamTypes[i])
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
	if lambda.ReturnType != "" {
		declared := c.resolveTypeWithGenerics(lambda.ReturnType, c.genericTypes)
		if declared == Unknown && !isDynamicTypeName(lambda.ReturnType) {
			if privateName, ok := c.inaccessibleTypeName(lambda.ReturnType); ok {
				c.errorf(lambda.Pos, "return type %q is private", privateName)
			} else {
				c.errorf(lambda.Pos, "unknown return type %q", lambda.ReturnType)
			}
		}
		if !typesCompatible(declared, ret, nil) {
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
		if inferred := c.inferredParamUseType(body, name); inferred != Unknown {
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
			walk(e.Alternative, expected)
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

func (c *checker) inferFunctionValueCall(name string, typ Type, args []ast.Expr, argTypes []Type, pos lexer.Position) (Type, Type) {
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
		if refined, ok := c.refineArgumentType(expected, argTypes[i]); ok {
			refinedParams[i] = refined
			if shouldApplyExpectedType(argTypes[i], expected) {
				argTypes[i] = refined
				c.applyExpectedType(args[i], refined)
			}
			continue
		}
		if !typesCompatible(expected, argTypes[i], nil) {
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
