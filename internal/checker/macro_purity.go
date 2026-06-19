package checker

import (
	"fmt"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func (c *checker) checkMacroPurity(file *ast.File) {
	for _, fn := range file.Functions {
		if !fn.Macro {
			continue
		}
		if message := c.localMacroPurityError(fn, map[*ast.Function]bool{}); message != "" {
			c.errorf(fn.NamePos, "macro %s is not pure: %s", fn.Name, message)
		}
	}
}

func (c *checker) localMacroPurityError(fn *ast.Function, visiting map[*ast.Function]bool) string {
	if fn == nil || fn.Body == nil || visiting[fn] {
		return ""
	}
	visiting[fn] = true
	defer delete(visiting, fn)
	return c.exprPurityError(fn.Body, visiting, map[*stdlib.Function]bool{})
}

func (c *checker) stdlibMacroPurityError(fn *stdlib.Function) string {
	if fn == nil {
		return "macro body is not available"
	}
	if c.stdlibMacroPurity == nil {
		c.stdlibMacroPurity = map[*stdlib.Function]string{}
	}
	if message, ok := c.stdlibMacroPurity[fn]; ok {
		return message
	}
	message := c.stdlibFunctionPurityError(fn, map[*stdlib.Function]bool{})
	c.stdlibMacroPurity[fn] = message
	return message
}

func (c *checker) stdlibFunctionPurityError(fn *stdlib.Function, visiting map[*stdlib.Function]bool) string {
	if fn == nil {
		return "function is not declared"
	}
	if visiting[fn] {
		return ""
	}
	if fn.Body == nil {
		return fmt.Sprintf("calls intrinsic %s", stdlibFunctionName(fn))
	}
	visiting[fn] = true
	defer delete(visiting, fn)
	return c.exprPurityError(fn.Body, map[*ast.Function]bool{}, visiting)
}

func (c *checker) exprPurityError(expr ast.Expr, localVisiting map[*ast.Function]bool, stdlibVisiting map[*stdlib.Function]bool) string {
	var message string
	ast.WalkExpr(expr, func(candidate ast.Expr) {
		if message != "" {
			return
		}
		call, ok := candidate.(*ast.CallExpr)
		if !ok {
			return
		}
		switch callee := call.Callee.(type) {
		case *ast.Identifier:
			if c.isPureConstructorCall(callee.Name) {
				return
			}
			fn := c.info.ResolvedFunctions[callee]
			if fn == nil || fn.Node == nil {
				message = fmt.Sprintf("cannot prove call %s is pure", callee.Name)
				return
			}
			message = c.localMacroPurityError(fn.Node, localVisiting)
		case *ast.SelectorExpr:
			at, ok := callee.Receiver.(*ast.AtExpr)
			if !ok || at.Name == "" {
				if pureMacroMethod(callee.Name) {
					return
				}
				message = fmt.Sprintf("cannot prove method call %s is pure", callee.Name)
				return
			}
			fn, ok := c.info.Stdlib.Function(at.Name, callee.Name)
			if !ok {
				message = fmt.Sprintf("calls unknown function @%s.%s", at.Name, callee.Name)
				return
			}
			if fn.Body == nil {
				message = fmt.Sprintf("calls impure function @%s.%s", at.Name, callee.Name)
				return
			}
			message = c.stdlibFunctionPurityError(fn, stdlibVisiting)
		default:
			message = "cannot prove dynamic call is pure"
		}
	})
	return message
}

func (c *checker) isPureConstructorCall(name string) bool {
	if pureSyntaxConstructor(name) {
		return true
	}
	if len(c.info.Constructors[name]) > 0 {
		return true
	}
	if c.info.Stdlib == nil {
		return false
	}
	for _, typ := range c.info.Stdlib.Types {
		for _, constructor := range typ.Constructors {
			if constructor.Name == name {
				return true
			}
		}
	}
	return false
}

func pureSyntaxConstructor(name string) bool {
	switch name {
	case "IdentifierExpr", "ModuleExpr", "StringExpr", "BoolExpr", "NullExpr",
		"SelectorExpr", "StaticSelectorExpr", "CallExpr", "StructExpr", "BlockExpr":
		return true
	default:
		return false
	}
}

func pureMacroMethod(name string) bool {
	switch name {
	case "map", "foldl", "length", "isEmpty":
		return true
	default:
		return false
	}
}

func stdlibFunctionName(fn *stdlib.Function) string {
	if fn.Intrinsic != "" {
		return fn.Intrinsic
	}
	return fn.Name
}
