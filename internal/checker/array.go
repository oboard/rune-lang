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
		if elemType == Unknown {
			elemType = typ
			continue
		}
		if typ != Unknown && typ != elemType {
			c.errorf(elem.Position(), "array element has type %s, expected %s", typ, elemType)
		}
	}
	return ArrayOf(elemType)
}

func (c *checker) inferIndexExpr(expr *ast.IndexExpr, env map[string]Type) Type {
	if _, ok := c.info.Stdlib.FunctionByAlias("array", "_[_]"); !ok {
		c.errorf(expr.Pos, "array index operator is not declared in core/array")
	}
	receiver := c.inferExpr(expr.Receiver, env)
	index := c.inferExpr(expr.Index, env)
	if index != Int && index != Unknown {
		c.errorf(expr.Index.Position(), "array index expects Int, got %s", index)
	}
	elem, ok := ArrayElement(receiver)
	if !ok {
		if receiver != Unknown {
			c.errorf(expr.Pos, "type %s is not indexable", receiver)
		}
		return Unknown
	}
	return elem
}

func (c *checker) inferArrayMethodCall(elem Type, sel *ast.SelectorExpr, call *ast.CallExpr, argTypes []Type, env map[string]Type) Type {
	fn, ok := c.info.Stdlib.Function("array", sel.Name)
	if !ok {
		c.errorf(sel.Pos, "type Array has no method %q", sel.Name)
		return Unknown
	}
	bindings := c.arrayTypeBindings(fn, elem)
	c.checkArrayDeclaredArgs(sel.Name, fn, call.Args, argTypes, bindings, env, sel.Pos)
	return c.resolveDeclaredType(fn.Return, bindings)
}

func (c *checker) arrayTypeBindings(fn *stdlib.Function, elem Type) map[string]Type {
	bindings := map[string]Type{}
	for _, name := range fn.Generics {
		bindings[name] = Unknown
	}
	bindings["T"] = elem
	return bindings
}

func (c *checker) checkArrayDeclaredArgs(functionName string, fn *stdlib.Function, args []ast.Expr, argTypes []Type, bindings map[string]Type, env map[string]Type, pos lexer.Position) {
	if len(fn.Params) != len(args) {
		c.errorf(pos, "array.%s expects %d args, got %d", functionName, len(fn.Params), len(args))
	}
	limit := min(len(fn.Params), len(args))
	for i := 0; i < limit; i++ {
		c.checkDeclaredGenericArg("array", functionName, i, fn.Params[i], args[i], argTypes[i], bindings, env)
	}
}

func (c *checker) checkDeclaredGenericArg(moduleName string, functionName string, index int, expected string, arg ast.Expr, actual Type, bindings map[string]Type, env map[string]Type) {
	if expected == "Any" || expected == "Dynamic" {
		return
	}
	if params, ret, ok := parseFuncType(expected); ok {
		lambda, ok := arg.(*ast.LambdaExpr)
		if !ok {
			expectedType := c.resolveDeclaredType(expected, bindings)
			if actual != Unknown && expectedType != Unknown && actual != expectedType {
				c.errorf(arg.Position(), "argument %d to @%s.%s has type %s, expected %s", index+1, moduleName, functionName, actual, expectedType)
			}
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
		if bindings[expected] != actual {
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
	expectedType := c.resolveDeclaredType(expected, bindings)
	if actual != Unknown && expectedType != Unknown && actual != expectedType {
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
	if elem, ok := parseArrayType(name); ok {
		elemType := c.resolveDeclaredType(elem, bindings)
		if elemType == Unknown {
			return Unknown
		}
		return ArrayOf(elemType)
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

func parseArrayType(name string) (string, bool) {
	if !strings.HasPrefix(name, "Array[") || !strings.HasSuffix(name, "]") {
		return "", false
	}
	return strings.TrimSuffix(strings.TrimPrefix(name, "Array["), "]"), true
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
