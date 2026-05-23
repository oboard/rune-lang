package checker

import (
	"sort"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

func (c *checker) inferMethods(typ *ast.StructType) {
	structInfo := c.info.Types[typ.Name]
	if structInfo == nil {
		return
	}
	for _, method := range typ.Methods {
		info := structInfo.Methods[method.Name]
		if info == nil {
			continue
		}
		if isIntrinsicStub(method) {
			if !info.ReturnDeclared {
				info.Return = Void
			}
			continue
		}
		env := map[string]Type{"this": info.ReceiverType}
		for _, param := range info.Params {
			env[param.Name] = param.Type
		}
		ret := c.inferExpr(method.Body, env)
		c.finishFunctionReturn(info, ret, method)
	}
}

func (c *checker) inferFunction(fn *ast.Function) {
	info := c.info.Functions[fn.Name]
	if info == nil {
		return
	}
	if isIntrinsicStub(fn) {
		if !info.ReturnDeclared {
			info.Return = Void
		}
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
	ret := c.inferExpr(fn.Body, env)
	for idx, param := range info.Params {
		if param.Type == Unknown {
			if inferred := env[param.Name]; inferred != "" && inferred != Unknown {
				info.Params[idx].Type = inferred
			} else if inferred := c.inferredParamUseType(fn.Body, param.Name); inferred != Unknown {
				info.Params[idx].Type = inferred
			}
		}
	}
	if fn.Name == "main" {
		if info.ReturnDeclared && info.Return != Void {
			c.errorf(fn.NamePos, "main must return Void, got %s", info.Return)
		}
		info.Return = Void
		return
	}
	c.finishFunctionReturn(info, ret, fn)
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

func (c *checker) finishFunctionReturn(info *FuncInfo, ret Type, fn *ast.Function) {
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
	case *ast.StringLiteral:
		return String
	case *ast.BoolLiteral:
		return Bool
	case *ast.Identifier:
		if typ, ok := env[e.Name]; ok {
			return typ
		}
		if fn, ok := c.info.Functions[e.Name]; ok {
			typ := FuncOfTypes(paramTypes(fn.Params), fn.Return)
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
		receiver := c.inferExpr(e.Receiver, env)
		structInfo := c.info.Types[baseTypeName(receiver)]
		if structInfo == nil {
			if receiver != Unknown {
				c.errorf(e.Pos, "type %s has no fields", receiver)
			}
			return Unknown
		}
		field, ok := structInfo.ByName[e.Name]
		if !ok {
			c.errorf(e.Pos, "type %s has no field %q", receiver, e.Name)
			return Unknown
		}
		return field.Type
	case *ast.StructLiteral:
		return c.inferStructLiteral(e, env)
	case *ast.AnonymousObjectLiteral:
		return c.inferAnonymousObjectLiteral(e, env)
	case *ast.ArrayLiteral:
		return c.inferArrayLiteral(e, env)
	case *ast.IndexExpr:
		return c.inferIndexExpr(e, env)
	case *ast.UnaryExpr:
		typ := c.inferExpr(e.Expr, env)
		switch e.Op {
		case lexer.Minus:
			if typ != Int && typ != Unknown {
				c.errorf(e.Pos, "operator '-' expects Int, got %s", typ)
			}
			return Int
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
			if typ != Int && typ != Unknown {
				c.errorf(e.Pos, "operator '++' expects Int, got %s", typ)
			}
			return Int
		default:
			return Unknown
		}
	case *ast.BinaryExpr:
		left := c.inferExpr(e.Left, env)
		right := c.inferExpr(e.Right, env)
		switch e.Op {
		case lexer.Plus:
			if left == String || right == String {
				if left != String && left != Unknown {
					c.errorf(e.Left.Position(), "string concatenation expects String, got %s", left)
				}
				if right != String && right != Unknown {
					c.errorf(e.Right.Position(), "string concatenation expects String, got %s", right)
				}
				return String
			}
			if left != Int && left != Unknown {
				c.errorf(e.Left.Position(), "arithmetic expects Int, got %s", left)
			}
			if right != Int && right != Unknown {
				c.errorf(e.Right.Position(), "arithmetic expects Int, got %s", right)
			}
			return Int
		case lexer.Minus, lexer.Star, lexer.Slash, lexer.Percent:
			if left != Int && left != Unknown {
				c.errorf(e.Left.Position(), "arithmetic expects Int, got %s", left)
			}
			if right != Int && right != Unknown {
				c.errorf(e.Right.Position(), "arithmetic expects Int, got %s", right)
			}
			return Int
		case lexer.EqualEqual, lexer.BangEqual, lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
			return Bool
		default:
			return Unknown
		}
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

func (c *checker) inferCall(call *ast.CallExpr, env map[string]Type) Type {
	argTypes := make([]Type, 0, len(call.Args))
	for _, arg := range call.Args {
		argTypes = append(argTypes, c.inferExpr(arg, env))
	}
	if sel, ok := call.Callee.(*ast.SelectorExpr); ok {
		if at, ok := sel.Receiver.(*ast.AtExpr); ok {
			if fn, ok := c.info.Stdlib.Function(at.Name, sel.Name); ok {
				return c.inferStdlibCall(at.Name, sel, call, argTypes, fn)
			}
			c.errorf(sel.Pos, "unknown module function @%s.%s", at.Name, sel.Name)
			return Unknown
		}
		receiver := c.inferExpr(sel.Receiver, env)
		if elem, ok := ArrayElement(receiver); ok {
			return c.inferArrayMethodCall(elem, sel, call, argTypes, env)
		}
		structInfo := c.info.Types[baseTypeName(receiver)]
		if structInfo == nil {
			if receiver != Unknown {
				c.errorf(sel.Pos, "type %s has no methods", receiver)
			}
			return Unknown
		}
		method := structInfo.Methods[sel.Name]
		if field, ok := structInfo.ByName[sel.Name]; ok {
			ret, _ := c.inferFunctionValueCall(sel.Name, field.Type, call.Args, argTypes, sel.Pos)
			return ret
		}
		if method == nil {
			c.errorf(sel.Pos, "type %s has no method %q", receiver, sel.Name)
			return Unknown
		}
		c.checkArgs(sel.Name, method.Params, call.Args, argTypes, sel.Pos)
		return method.Return
	}
	if ident, ok := call.Callee.(*ast.Identifier); ok {
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
		fn := c.info.Functions[ident.Name]
		if fn == nil {
			c.errorf(ident.Pos, "undefined function %q", ident.Name)
			return Unknown
		}
		for i, argType := range argTypes {
			if i < len(fn.Params) && fn.Params[i].Type == Unknown && argType != Unknown {
				fn.Params[i].Type = argType
			}
		}
		c.info.ExprTypes[ident] = FuncOfTypes(paramTypes(fn.Params), fn.Return)
		c.checkArgs(ident.Name, fn.Params, call.Args, argTypes, ident.Pos)
		return fn.Return
	}
	calleeType := c.inferExpr(call.Callee, env)
	ret, refined := c.inferFunctionValueCall("<expr>", calleeType, call.Args, argTypes, call.Callee.Position())
	if refined != Unknown && refined != calleeType {
		c.info.ExprTypes[call.Callee] = refined
		c.applyExpectedType(call.Callee, refined)
	}
	return ret
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
			paramType = c.resolveType(lambda.ParamTypes[i])
			if paramType == Unknown && !isDynamicTypeName(lambda.ParamTypes[i]) {
				c.errorf(lambda.Pos, "unknown type %q", lambda.ParamTypes[i])
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
	if lambda.ReturnType != "" {
		declared := c.resolveDeclaredReturn(lambda.ReturnType)
		if declared == Unknown && !isDynamicTypeName(lambda.ReturnType) {
			c.errorf(lambda.Pos, "unknown return type %q", lambda.ReturnType)
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
			case lexer.Plus, lexer.Minus, lexer.Star, lexer.Slash, lexer.Percent:
				walk(e.Left, Int)
				walk(e.Right, Int)
			default:
				walk(e.Left, Unknown)
				walk(e.Right, Unknown)
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
	params, ret, ok := parseFuncType(string(typ))
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
			continue
		}
		if !typesCompatible(expected, argTypes[i], nil) {
			c.errorf(args[i].Position(), "argument %d to %q has type %s, expected %s", i+1, name, argTypes[i], expected)
		}
	}
	return Type(ret), FuncOfTypes(refinedParams, Type(ret))
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
