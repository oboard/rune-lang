package checker

import (
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
	switch fn.Intrinsic {
	case "array.each":
		return c.inferArrayEach(elem, sel, call, env)
	case "array.map":
		return c.inferArrayMap(elem, sel, call, env)
	}
	c.checkArrayArgs(sel.Name, fn, call.Args, argTypes, elem, sel.Pos)
	if fn.Return == "T" {
		return elem
	}
	return c.resolveDeclaredReturn(fn.Return)
}

func (c *checker) inferArrayEach(elem Type, sel *ast.SelectorExpr, call *ast.CallExpr, env map[string]Type) Type {
	lambda, ok := c.arrayLambdaArg("each", sel, call)
	if !ok {
		return Void
	}
	ret := c.inferLambda(lambda, elem, env)
	if ret != Void && ret != Unknown {
		c.errorf(lambda.Body.Position(), "array.each lambda returns %s, expected Void", ret)
	}
	return Void
}

func (c *checker) inferArrayMap(elem Type, sel *ast.SelectorExpr, call *ast.CallExpr, env map[string]Type) Type {
	lambda, ok := c.arrayLambdaArg("map", sel, call)
	if !ok {
		return ArrayOf(Unknown)
	}
	ret := c.inferLambda(lambda, elem, env)
	if ret == Void {
		c.errorf(lambda.Body.Position(), "array.map lambda must return a value")
		return ArrayOf(Unknown)
	}
	return ArrayOf(ret)
}

func (c *checker) arrayLambdaArg(functionName string, sel *ast.SelectorExpr, call *ast.CallExpr) (*ast.LambdaExpr, bool) {
	if len(call.Args) != 1 {
		c.errorf(sel.Pos, "array.%s expects 1 args, got %d", functionName, len(call.Args))
		return nil, false
	}
	lambda, ok := call.Args[0].(*ast.LambdaExpr)
	if !ok {
		c.errorf(call.Args[0].Position(), "array.%s expects a lambda", functionName)
		return nil, false
	}
	if len(lambda.Params) != 1 {
		c.errorf(lambda.Pos, "array.%s lambda expects exactly 1 parameter", functionName)
		return nil, false
	}
	return lambda, true
}

func (c *checker) inferLambda(lambda *ast.LambdaExpr, elem Type, env map[string]Type) Type {
	local := cloneEnv(env)
	local[lambda.Params[0]] = elem
	ret := c.inferExpr(lambda.Body, local)
	c.info.ExprTypes[lambda] = FuncOf(elem, ret)
	return ret
}

func (c *checker) checkArrayArgs(functionName string, fn *stdlib.Function, args []ast.Expr, argTypes []Type, elem Type, pos lexer.Position) {
	if len(fn.Params) != len(args) {
		c.errorf(pos, "array.%s expects %d args, got %d", functionName, len(fn.Params), len(args))
	}
	limit := min(len(fn.Params), len(args))
	for i := 0; i < limit; i++ {
		expected := fn.Params[i]
		if expected == "T" {
			if argTypes[i] != Unknown && elem != Unknown && argTypes[i] != elem {
				c.errorf(args[i].Position(), "argument %d to array.%s has type %s, expected %s", i+1, functionName, argTypes[i], elem)
			}
			continue
		}
		c.checkDeclaredArg("array", functionName, i, expected, args[i], argTypes[i])
	}
}
