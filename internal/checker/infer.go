package checker

import (
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
		env := map[string]Type{"this": Type(typ.Name)}
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
	env := map[string]Type{}
	for _, param := range info.Params {
		env[param.Name] = param.Type
	}
	ret := c.inferExpr(fn.Body, env)
	if fn.Name == "main" {
		if info.ReturnDeclared && info.Return != Void {
			c.errorf(fn.NamePos, "main must return Void, got %s", info.Return)
		}
		info.Return = Void
		return
	}
	c.finishFunctionReturn(info, ret, fn)
}

func (c *checker) finishFunctionReturn(info *FuncInfo, ret Type, fn *ast.Function) {
	if info.ReturnDeclared {
		if ret != Unknown && ret != info.Return {
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
		if _, ok := c.info.Functions[e.Name]; ok {
			return Unknown
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
		structInfo := c.info.Types[string(receiver)]
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
	case *ast.BinaryExpr:
		left := c.inferExpr(e.Left, env)
		right := c.inferExpr(e.Right, env)
		switch e.Op {
		case lexer.Plus, lexer.Minus, lexer.Star, lexer.Slash, lexer.Percent:
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
		return Unknown
	case *ast.BlockExpr:
		return c.inferBlock(e, env)
	case *ast.PatternBlock:
		return c.inferPatternBlock(e, env)
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
		structInfo := c.info.Types[string(receiver)]
		if structInfo == nil {
			if receiver != Unknown {
				c.errorf(sel.Pos, "type %s has no methods", receiver)
			}
			return Unknown
		}
		method := structInfo.Methods[sel.Name]
		if method == nil {
			c.errorf(sel.Pos, "type %s has no method %q", receiver, sel.Name)
			return Unknown
		}
		c.checkArgs(sel.Name, method.Params, call.Args, argTypes, sel.Pos)
		return method.Return
	}
	if ident, ok := call.Callee.(*ast.Identifier); ok {
		fn := c.info.Functions[ident.Name]
		if fn == nil {
			c.errorf(ident.Pos, "undefined function %q", ident.Name)
			return Unknown
		}
		c.checkArgs(ident.Name, fn.Params, call.Args, argTypes, ident.Pos)
		return fn.Return
	}
	c.inferExpr(call.Callee, env)
	return Unknown
}
