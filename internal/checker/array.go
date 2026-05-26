package checker

import (
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func (c *checker) inferArrayLiteral(lit *ast.ArrayLiteral, env map[string]Type) Type {
	if len(lit.Elements) == 0 {
		return ArrayOf(Unknown)
	}
	elemType := Unknown
	for _, elem := range lit.Elements {
		typ := c.inferExpr(elem, env)
		if spread, ok := elem.(*ast.SpreadExpr); ok {
			spreadElem, ok := ArrayElement(typ)
			if !ok {
				if typ != Unknown {
					c.errorf(spread.Pos, "spread expects Array, got %s", typ)
				}
				continue
			}
			typ = spreadElem
		}
		if elemType == Unknown {
			elemType = typ
			continue
		}
		if typ != Unknown {
			unified, ok := c.unifyTypes(elemType, typ)
			if !ok {
				c.errorf(elem.Position(), "array element has type %s, expected %s", typ, elemType)
				continue
			}
			elemType = unified
		}
	}
	return ArrayOf(elemType)
}

func (c *checker) inferTupleLiteral(lit *ast.TupleLiteral, env map[string]Type) Type {
	elems := make([]Type, 0, len(lit.Elements))
	for _, elem := range lit.Elements {
		elems = append(elems, c.inferExpr(elem, env))
	}
	return genericTypeOf("Tuple", elems)
}

func (c *checker) inferMapLiteral(lit *ast.MapLiteral, env map[string]Type) Type {
	keyType := Unknown
	valueType := Unknown
	for _, entry := range lit.Entries {
		key := c.inferExpr(entry.Key, env)
		value := c.inferExpr(entry.Value, env)
		if keyType == Unknown {
			keyType = key
		} else if key != Unknown {
			unified, ok := c.unifyTypes(keyType, key)
			if !ok {
				c.errorf(entry.Key.Position(), "map key has type %s, expected %s", key, keyType)
			} else {
				keyType = unified
			}
		}
		if valueType == Unknown {
			valueType = value
		} else if value != Unknown {
			unified, ok := c.unifyTypes(valueType, value)
			if !ok {
				c.errorf(entry.Value.Position(), "map value has type %s, expected %s", value, valueType)
			} else {
				valueType = unified
			}
		}
	}
	if !mapKeyType(keyType) && keyType != Unknown {
		c.errorf(lit.Pos, "map key type %s is not supported", keyType)
	}
	return MapOf(keyType, valueType)
}

func (c *checker) inferIndexExpr(expr *ast.IndexExpr, env map[string]Type) Type {
	receiver := c.inferExpr(expr.Receiver, env)
	index := c.inferExpr(expr.Index, env)
	if elems, ok := TupleElements(receiver); ok {
		lit, ok := expr.Index.(*ast.IntegerLiteral)
		if !ok {
			c.errorf(expr.Index.Position(), "tuple index expects an integer literal")
			return Unknown
		}
		if lit.Value < 0 || lit.Value >= len(elems) {
			c.errorf(expr.Index.Position(), "tuple index %d out of range", lit.Value)
			return Unknown
		}
		return elems[lit.Value]
	}
	if elem, ok := ArrayElement(receiver); ok {
		if _, ok := c.info.Stdlib.FunctionByAlias("array", "_[_]"); !ok {
			c.errorf(expr.Pos, "array index operator is not declared in core/array")
		}
		if index != Int && index != Unknown {
			c.errorf(expr.Index.Position(), "array index expects Int, got %s", index)
		}
		return elem
	}
	if key, value, ok := MapKeyValue(receiver); ok {
		if !typesCompatible(key, index, nil) {
			c.errorf(expr.Index.Position(), "map index has type %s, expected %s", index, key)
		}
		return value
	}
	if receiver != Unknown {
		c.errorf(expr.Pos, "type %s is not indexable", receiver)
	}
	return Unknown
}

func (c *checker) inferIndexAssign(target *ast.IndexExpr, value ast.Expr, env map[string]Type) Type {
	receiver := c.inferExpr(target.Receiver, env)
	index := c.inferExpr(target.Index, env)
	actual := c.inferExpr(value, env)
	if key, expected, ok := MapKeyValue(receiver); ok {
		if !typesCompatible(key, index, nil) {
			c.errorf(target.Index.Position(), "map index has type %s, expected %s", index, key)
		}
		if !typesCompatible(expected, actual, nil) {
			c.errorf(value.Position(), "map assignment has type %s, expected %s", actual, expected)
		}
		return Void
	}
	if receiver != Unknown {
		c.errorf(target.Pos, "type %s is not assignable by index", receiver)
	}
	return Void
}

func mapKeyType(typ Type) bool {
	switch typ {
	case String, Bool, Int, Int4, Int8, Int16, Int64, UInt, UInt8, UInt16, UInt64, Double, Float:
		return true
	default:
		return false
	}
}

func (c *checker) inferArrayMethodCall(elem Type, sel *ast.SelectorExpr, call *ast.CallExpr, argTypes []Type, env map[string]Type) Type {
	fn, ok := c.info.Stdlib.Function("array", sel.Name)
	if !ok {
		c.errorf(sel.Pos, "type Array has no method %q", sel.Name)
		return Unknown
	}
	bindings := c.arrayTypeBindings(fn, elem)
	c.checkDeclaredReceiverArgs("array", sel.Name, fn, call.Args, argTypes, bindings, env, sel.Pos)
	return c.finishRoutineCall(call, fn.Routine, c.resolveDeclaredType(fn.Return, bindings))
}

func (c *checker) inferStdlibReceiverMethodCall(receiver Type, sel *ast.SelectorExpr, call *ast.CallExpr, argTypes []Type, env map[string]Type) (Type, bool) {
	moduleName, receiverName, ok := StdlibReceiverModule(receiver)
	if !ok {
		return Unknown, false
	}
	fn, ok := c.info.Stdlib.ReceiverFunction(moduleName, receiverName, sel.Name)
	if !ok || (!isPublicStdlibReceiverMethod(receiverName, sel.Name) && !c.isSourceType(fn.SourcePath)) {
		c.errorf(sel.Pos, "type %s has no method %q", receiver, sel.Name)
		return Unknown, true
	}
	bindings := c.receiverTypeBindings(fn, receiver)
	c.checkDeclaredReceiverArgs(moduleName, sel.Name, fn, call.Args, argTypes, bindings, env, sel.Pos)
	return c.finishRoutineCall(call, fn.Routine, c.resolveDeclaredType(fn.Return, bindings)), true
}

func isPublicStdlibReceiverMethod(receiverName string, methodName string) bool {
	if receiverName != "Iter" {
		return true
	}
	switch methodName {
	case "toArray", "each", "map":
		return true
	default:
		return false
	}
}

func StdlibReceiverModule(receiver Type) (string, string, bool) {
	base := baseTypeName(receiver)
	switch base {
	case "Int":
		return "int", "Int", true
	case "String":
		return "string", "String", true
	case "Bool":
		return "bool", "Bool", true
	case "Regex":
		return "regex", "Regex", true
	case "Binary":
		return "binary", "Binary", true
	case "Buffer":
		return "buffer", "Buffer", true
	case "Reader":
		return "reader", "Reader", true
	case "Writer":
		return "writer", "Writer", true
	case "StringBuffer":
		return "stringbuffer", "StringBuffer", true
	case "TCPConnection", "TCPListener":
		return "net", base, true
	case "Iter":
		return "iter", "Iter", true
	case "Map", "WeakMap":
		return "map", base, true
	case "Set", "WeakSet":
		return "set", base, true
	default:
		return "", "", false
	}
}

func (c *checker) receiverTypeBindings(fn *stdlib.Function, receiver Type) map[string]Type {
	bindings := map[string]Type{}
	for _, name := range fn.Generics {
		bindings[name] = Unknown
	}
	if base, args, ok := parseGenericType(string(receiver)); ok {
		switch base {
		case "Map", "WeakMap":
			if len(args) >= 2 {
				bindings["K"] = Type(args[0])
				bindings["V"] = Type(args[1])
			}
		case "Set", "WeakSet":
			if len(args) >= 1 {
				bindings["T"] = Type(args[0])
			}
		case "Iter":
			if len(args) >= 1 {
				bindings["T"] = Type(args[0])
			}
		}
	}
	return bindings
}

func (c *checker) arrayTypeBindings(fn *stdlib.Function, elem Type) map[string]Type {
	bindings := map[string]Type{}
	for _, name := range fn.Generics {
		bindings[name] = Unknown
	}
	bindings["T"] = elem
	return bindings
}

func (c *checker) checkDeclaredReceiverArgs(moduleName string, functionName string, fn *stdlib.Function, args []ast.Expr, argTypes []Type, bindings map[string]Type, env map[string]Type, pos lexer.Position) {
	if len(fn.Params) != len(args) {
		c.errorf(pos, "%s.%s expects %d args, got %d", moduleName, functionName, len(fn.Params), len(args))
	}
	limit := min(len(fn.Params), len(args))
	for i := 0; i < limit; i++ {
		c.checkDeclaredGenericArg(moduleName, functionName, i, fn.Params[i], args[i], argTypes[i], bindings, env)
	}
}

func (c *checker) checkDeclaredGenericArg(moduleName string, functionName string, index int, expected string, arg ast.Expr, actual Type, bindings map[string]Type, env map[string]Type) {
	if expected == "Any" || expected == "Dynamic" {
		return
	}
	if params, ret, ok := parseFuncType(expected); ok {
		lambda, ok := arg.(*ast.LambdaExpr)
		if !ok {
			c.bindDeclaredType(expected, actual, bindings, arg.Position(), moduleName, functionName, index)
			return
		}
		if len(lambda.Params) > len(params) {
			c.errorf(lambda.Pos, "argument %d to @%s.%s expects lambda with at most %d parameters, got %d", index+1, moduleName, functionName, len(params), len(lambda.Params))
			return
		}
		local := cloneEnv(env)
		paramTypes := make([]Type, 0, len(params))
		for i, param := range params[:len(lambda.Params)] {
			paramType := c.resolveDeclaredType(param, bindings)
			paramTypes = append(paramTypes, paramType)
			local[lambda.Params[i]] = paramType
		}
		retType := c.inferExpr(lambda.Body, local)
		c.info.ExprTypes[lambda] = funcTypeOf(paramTypes, retType)
		c.bindDeclaredType(ret, retType, bindings, arg.Position(), moduleName, functionName, index)
		return
	}
	c.bindDeclaredType(expected, actual, bindings, arg.Position(), moduleName, functionName, index)
}

func (c *checker) bindDeclaredType(expected string, actual Type, bindings map[string]Type, pos lexer.Position, moduleName string, functionName string, index int) {
	if expected == "" || expected == "Any" || expected == "Dynamic" || actual == Unknown {
		return
	}
	if _, ok := bindings[expected]; ok {
		if bindings[expected] == Unknown {
			bindings[expected] = actual
			return
		}
		if unified, ok := c.unifyTypes(bindings[expected], actual); ok {
			bindings[expected] = unified
			return
		}
		if !typesCompatible(bindings[expected], actual, nil) {
			c.errorf(pos, "argument %d to @%s.%s has type %s, expected %s", index+1, moduleName, functionName, actual, bindings[expected])
		}
		return
	}
	if elem, ok := parseArrayType(expected); ok {
		actualElem, ok := ArrayElement(actual)
		if !ok {
			c.errorf(pos, "argument %d to @%s.%s has type %s, expected %s", index+1, moduleName, functionName, actual, c.resolveDeclaredType(expected, bindings))
			return
		}
		c.bindDeclaredType(elem, actualElem, bindings, pos, moduleName, functionName, index)
		return
	}
	if params, ret, ok := parseFuncType(expected); ok {
		actualParams, actualRet, actualFunc := parseFuncType(string(actual))
		if !actualFunc || len(actualParams) != len(params) {
			c.errorf(pos, "argument %d to @%s.%s has type %s, expected %s", index+1, moduleName, functionName, actual, c.resolveDeclaredType(expected, bindings))
			return
		}
		for i, param := range params {
			c.bindDeclaredType(param, Type(actualParams[i]), bindings, pos, moduleName, functionName, index)
		}
		c.bindDeclaredType(ret, Type(actualRet), bindings, pos, moduleName, functionName, index)
		return
	}
	if base, args, ok := parseGenericType(expected); ok {
		actualBase, actualArgs, actualGeneric := parseGenericType(string(actual))
		if !actualGeneric || actualBase != base || len(actualArgs) != len(args) {
			c.errorf(pos, "argument %d to @%s.%s has type %s, expected %s", index+1, moduleName, functionName, actual, c.resolveDeclaredType(expected, bindings))
			return
		}
		for i, arg := range args {
			c.bindDeclaredType(arg, Type(actualArgs[i]), bindings, pos, moduleName, functionName, index)
		}
		return
	}
	expectedType := c.resolveDeclaredType(expected, bindings)
	if actual != Unknown && expectedType != Unknown && !typesCompatible(expectedType, actual, nil) {
		c.errorf(pos, "argument %d to @%s.%s has type %s, expected %s", index+1, moduleName, functionName, actual, expectedType)
	}
}

func (c *checker) resolveDeclaredType(name string, bindings map[string]Type) Type {
	if name == "" {
		return Void
	}
	if name == "Dynamic" || name == "Any" {
		return Unknown
	}
	if typ, ok := bindings[name]; ok {
		return typ
	}
	if inner, ok := parseNullableType(name); ok {
		innerType := c.resolveDeclaredType(inner, bindings)
		if innerType == Unknown {
			return Unknown
		}
		return NullableOf(innerType)
	}
	if elem, ok := parseArrayType(name); ok {
		elemType := c.resolveDeclaredType(elem, bindings)
		if elemType == Unknown {
			return Unknown
		}
		return ArrayOf(elemType)
	}
	if base, args, ok := parseGenericType(name); ok && (isBuiltinGenericType(base) || c.coreTypeExists(base)) {
		resolved := make([]Type, 0, len(args))
		for _, arg := range args {
			resolved = append(resolved, c.resolveDeclaredType(arg, bindings))
		}
		return genericTypeOf(base, resolved)
	}
	if params, ret, ok := parseFuncType(name); ok {
		types := make([]Type, 0, len(params)+1)
		for _, param := range params {
			types = append(types, c.resolveDeclaredType(param, bindings))
		}
		types = append(types, c.resolveDeclaredType(ret, bindings))
		return funcTypeOf(types[:len(types)-1], types[len(types)-1])
	}
	return c.resolveDeclaredReturn(name)
}

func parseNullableType(name string) (string, bool) {
	if !strings.HasSuffix(name, "?") || name == "?" {
		return "", false
	}
	return strings.TrimSuffix(name, "?"), true
}

func parseArrayType(name string) (string, bool) {
	if !strings.HasPrefix(name, "Array[") || !strings.HasSuffix(name, "]") {
		return "", false
	}
	return strings.TrimSuffix(strings.TrimPrefix(name, "Array["), "]"), true
}

func parseGenericType(name string) (string, []string, bool) {
	idx := strings.IndexByte(name, '[')
	if idx <= 0 || !strings.HasSuffix(name, "]") {
		return "", nil, false
	}
	base := name[:idx]
	args := splitTypeList(strings.TrimSuffix(name[idx+1:], "]"))
	if len(args) == 0 {
		return "", nil, false
	}
	return base, args, true
}

func isBuiltinGenericType(base string) bool {
	switch base {
	case "ReadonlyArray", "Tuple", "ReadonlyTuple", "Map", "Set", "WeakMap", "WeakSet", "Record", "Result", "Task", "Iter":
		return true
	default:
		return false
	}
}

func parseFuncType(name string) ([]string, string, bool) {
	if !strings.HasPrefix(name, "Func[") || !strings.HasSuffix(name, "]") {
		return nil, "", false
	}
	parts := splitTypeList(strings.TrimSuffix(strings.TrimPrefix(name, "Func["), "]"))
	if len(parts) == 0 {
		return nil, "", false
	}
	return parts[:len(parts)-1], parts[len(parts)-1], true
}

func parseAsyncFuncType(name string) ([]string, string, bool) {
	if !strings.HasPrefix(name, "AsyncFunc[") || !strings.HasSuffix(name, "]") {
		return nil, "", false
	}
	parts := splitTypeList(strings.TrimSuffix(strings.TrimPrefix(name, "AsyncFunc["), "]"))
	if len(parts) == 0 {
		return nil, "", false
	}
	return parts[:len(parts)-1], parts[len(parts)-1], true
}

func parseCallableType(name string) ([]string, string, bool) {
	if params, ret, ok := parseFuncType(name); ok {
		return params, ret, true
	}
	return parseAsyncFuncType(name)
}

func parseResultType(typ Type) (Type, Type, bool) {
	base, args, ok := parseGenericType(string(typ))
	if !ok || base != "Result" || len(args) != 2 {
		return Unknown, Unknown, false
	}
	return Type(args[0]), Type(args[1]), true
}

func parseTaskType(typ Type) (Type, bool) {
	base, args, ok := parseGenericType(string(typ))
	if !ok || base != "Task" || len(args) != 1 {
		return Unknown, false
	}
	return Type(args[0]), true
}

func splitTypeList(src string) []string {
	var out []string
	depth := 0
	start := 0
	for i, ch := range src {
		switch ch {
		case '[', '{', '(':
			depth++
		case ']', '}', ')':
			depth--
		case ',':
			if depth == 0 {
				out = append(out, strings.TrimSpace(src[start:i]))
				start = i + 1
			}
		}
	}
	out = append(out, strings.TrimSpace(src[start:]))
	return out
}

func funcTypeOf(params []Type, ret Type) Type {
	parts := make([]string, 0, len(params)+1)
	for _, param := range params {
		parts = append(parts, string(param))
	}
	parts = append(parts, string(ret))
	return Type("Func[" + strings.Join(parts, ",") + "]")
}
