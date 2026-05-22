package checker

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func (c *checker) inferStdlibCall(moduleName string, sel *ast.SelectorExpr, call *ast.CallExpr, argTypes []Type, fn *stdlib.Function) Type {
	if fn.TopLevelOnly {
		c.errorf(sel.Pos, "@%s.%s must be a top-level declaration", moduleName, sel.Name)
		return Void
	}

	switch fn.Intrinsic {
	case "go.stmt":
		c.checkStdlibArgs(moduleName, sel.Name, fn, call.Args, argTypes, sel.Pos)
		if len(call.Args) == 1 {
			if _, ok := call.Args[0].(*ast.StringLiteral); !ok {
				c.errorf(call.Args[0].Position(), "@go.stmt argument must be a string literal")
			}
		}
		return Void
	case "go.expr":
		c.checkStdlibArgs(moduleName, sel.Name, fn, call.Args, argTypes, sel.Pos)
		if len(call.Args) != 1 {
			return Unknown
		}
		if _, ok := call.Args[0].(*ast.StringLiteral); !ok {
			c.errorf(call.Args[0].Position(), "@go.expr body must be a string literal")
		}
		return c.resolveDeclaredReturn(fn.Return)
	}

	c.checkStdlibArgs(moduleName, sel.Name, fn, call.Args, argTypes, sel.Pos)
	return c.resolveDeclaredReturn(fn.Return)
}

func (c *checker) checkStdlibArgs(moduleName string, functionName string, fn *stdlib.Function, args []ast.Expr, argTypes []Type, pos lexer.Position) {
	if fn.Variadic {
		minArgs := len(fn.Params)
		if minArgs > 0 {
			minArgs--
		}
		if len(args) < minArgs {
			c.errorf(pos, "@%s.%s expects at least %d args, got %d", moduleName, functionName, minArgs, len(args))
		}
	} else if len(fn.Params) != len(args) {
		c.errorf(pos, "@%s.%s expects %d args, got %d", moduleName, functionName, len(fn.Params), len(args))
	}

	for i := 0; i < len(args) && i < len(fn.Params); i++ {
		expected := fn.Params[i]
		if fn.Variadic && i >= len(fn.Params)-1 {
			expected = fn.Params[len(fn.Params)-1]
		}
		c.checkDeclaredArg(moduleName, functionName, i, expected, args[i], argTypes[i])
	}
}

func (c *checker) checkDeclaredArg(moduleName string, functionName string, index int, expected string, arg ast.Expr, actual Type) {
	if expected == "Any" || expected == "Dynamic" {
		return
	}
	expectedType := c.resolveType(expected)
	if expectedType == Unknown {
		return
	}
	if !typesCompatible(expectedType, actual, nil) {
		c.errorf(arg.Position(), "argument %d to @%s.%s has type %s, expected %s", index+1, moduleName, functionName, actual, expectedType)
	}
}

func (c *checker) checkArgs(name string, params []ParamInfo, args []ast.Expr, argTypes []Type, pos lexer.Position) {
	if len(params) != len(args) {
		c.errorf(pos, "function %q expects %d args, got %d", name, len(params), len(args))
	}
	limit := min(len(params), len(argTypes))
	for i := 0; i < limit; i++ {
		if !typesCompatible(params[i].Type, argTypes[i], nil) {
			c.errorf(args[i].Position(), "argument %d to %q has type %s, expected %s", i+1, name, argTypes[i], params[i].Type)
		}
	}
}
