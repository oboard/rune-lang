package tscodegen

import (
	"errors"
	"fmt"
	"strings"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func (g *generator) codegenError() error {
	return errors.Join(g.errors...)
}

func (g *generator) addError(err error) {
	g.errors = append(g.errors, err)
}

func (g *generator) unsupportedIntrinsic(fn *stdlib.Function, typ checker.Type) string {
	g.addError(fmt.Errorf("TypeScript backend does not support intrinsic %s", fn.Intrinsic))
	return g.zeroValue(typ)
}

func (g *generator) moduleIntrinsicCall(call *ir.CallExpr) (string, bool) {
	fn, ok := g.stdlibFunctionFromCall(call)
	if !ok || fn.Intrinsic == "" {
		return "", false
	}
	args := g.intrinsicArgs(call.Args)
	switch fn.Intrinsic {
	case "io.print", "io.println", "io.printf":
		return "console.log(" + strings.Join(args, ", ") + ")", true
	case "json.stringify":
		if len(call.Args) != 1 {
			return "undefined", true
		}
		return "JSON.stringify(" + g.jsonValueExpr(call.Args[0]) + ")", true
	case "regex.new":
		if len(args) != 2 {
			return "undefined", true
		}
		return fmt.Sprintf("new RegExp(%s, %s)", args[0], args[1]), true
	case "regex.escape":
		if len(args) != 1 {
			return "undefined", true
		}
		return fmt.Sprintf("((__value: string): string => (RegExp as any).escape ? (RegExp as any).escape(__value) : __value.replace(/[\\\\^$.*+?()[\\]{}|]/g, \"\\\\$&\"))(%s)", args[0]), true
	case "map.new", "set.new":
		return "new " + tsType(call.ResultType()) + "()", true
	case "go.import", "go.stmt", "go.expr":
		return g.unsupportedIntrinsic(fn, call.ResultType()), true
	default:
		return g.unsupportedIntrinsic(fn, call.ResultType()), true
	}
}

func (g *generator) receiverIntrinsicCall(call *ir.CallExpr) (string, bool) {
	sel, fn, ok := g.stdlibReceiverFunctionFromCall(call)
	if !ok || fn.Intrinsic == "" {
		return "", false
	}
	receiver := g.expr(sel.Receiver)
	args := g.intrinsicArgs(call.Args)
	switch {
	case strings.HasPrefix(fn.Intrinsic, "array."):
		return g.arrayIntrinsicCall(fn, receiver, args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "string."):
		return g.stringIntrinsicCall(fn, receiver, args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "bool."):
		return g.boolIntrinsicCall(fn, receiver, args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "regex."):
		name := strings.TrimPrefix(fn.Intrinsic, "regex.")
		if expr, ok := g.regexMethodCall(receiver, name, args); ok {
			return expr, true
		}
		return g.unsupportedIntrinsic(fn, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "map."), strings.HasPrefix(fn.Intrinsic, "weakMap."):
		return g.mapIntrinsicCall(fn, receiver, args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "set."), strings.HasPrefix(fn.Intrinsic, "weakSet."):
		return g.setIntrinsicCall(fn, receiver, args, call.ResultType()), true
	default:
		return g.unsupportedIntrinsic(fn, call.ResultType()), true
	}
}

func (g *generator) stdlibFunctionFromCall(call *ir.CallExpr) (*stdlib.Function, bool) {
	sel, ok := call.Callee.(*ir.SelectorExpr)
	if !ok {
		return nil, false
	}
	at, ok := sel.Receiver.(*ir.AtExpr)
	if !ok || g.file.Stdlib == nil {
		return nil, false
	}
	return g.file.Stdlib.Function(at.Name, sel.Name)
}

func (g *generator) stdlibReceiverFunctionFromCall(call *ir.CallExpr) (*ir.SelectorExpr, *stdlib.Function, bool) {
	sel, ok := call.Callee.(*ir.SelectorExpr)
	if !ok || g.file.Stdlib == nil {
		return nil, nil, false
	}
	if _, ok := checker.ArrayElement(sel.Receiver.ResultType()); ok {
		fn, ok := g.file.Stdlib.Function("array", sel.Name)
		return sel, fn, ok
	}
	moduleName, receiverName, ok := checker.StdlibReceiverModule(sel.Receiver.ResultType())
	if !ok {
		return nil, nil, false
	}
	fn, ok := g.file.Stdlib.ReceiverFunction(moduleName, receiverName, sel.Name)
	return sel, fn, ok
}

func (g *generator) intrinsicArgs(args []ir.Expr) []string {
	out := make([]string, 0, len(args))
	for _, arg := range args {
		out = append(out, g.expr(arg))
	}
	return out
}

func (g *generator) arrayIntrinsicCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "array.len":
		return receiver + ".length"
	case "array.push":
		return fmt.Sprintf("%s.push(%s)", receiver, strings.Join(args, ", "))
	case "array.set":
		if len(args) != 2 {
			return "undefined"
		}
		return fmt.Sprintf("(%s[%s] = %s)", receiver, args[0], args[1])
	case "array.pop":
		return fmt.Sprintf("%s.pop() as %s", receiver, tsType(resultType))
	case "array.first":
		return fmt.Sprintf("%s[0]", receiver)
	case "array.last":
		return fmt.Sprintf("%s[%s.length - 1]", receiver, receiver)
	case "array.slice":
		return fmt.Sprintf("%s.slice(%s)", receiver, strings.Join(args, ", "))
	case "array.clone":
		return fmt.Sprintf("%s.slice()", receiver)
	case "array.reverse":
		return fmt.Sprintf("%s.slice().reverse()", receiver)
	case "array.contains":
		if len(args) != 1 {
			return "undefined"
		}
		return fmt.Sprintf("%s.includes(%s)", receiver, args[0])
	case "array.each", "array.forEach":
		if len(args) != 1 {
			return "undefined"
		}
		return fmt.Sprintf("%s.forEach(%s)", receiver, args[0])
	case "array.map":
		if len(args) != 1 {
			return "undefined"
		}
		return fmt.Sprintf("%s.map(%s)", receiver, args[0])
	case "array.at", "array.get":
		if len(args) != 1 {
			return "undefined"
		}
		return fmt.Sprintf("%s.at(%s)", receiver, args[0])
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}

func (g *generator) stringIntrinsicCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "string.length":
		return fmt.Sprintf("Array.from(%s).length", receiver)
	case "string.toString":
		return receiver
	case "string.at":
		if len(args) != 1 {
			return "undefined"
		}
		return fmt.Sprintf("(Array.from(%s)[%s] ?? \"\")", receiver, args[0])
	case "string.slice":
		if len(args) != 2 {
			return "undefined"
		}
		return fmt.Sprintf("Array.from(%s).slice(%s, %s).join(\"\")", receiver, args[0], args[1])
	case "string.concat", "string.includes", "string.startsWith", "string.endsWith", "string.indexOf", "string.lastIndexOf", "string.toLowerCase", "string.toUpperCase", "string.trim", "string.trimStart", "string.trimEnd", "string.repeat", "string.replace", "string.replaceAll", "string.split":
		name := strings.TrimPrefix(fn.Intrinsic, "string.")
		return fmt.Sprintf("%s.%s(%s)", receiver, name, strings.Join(args, ", "))
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}

func (g *generator) boolIntrinsicCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "bool.toString":
		return receiver + ".toString()"
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}

func (g *generator) mapIntrinsicCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "map.size":
		return receiver + ".size"
	case "map.has", "weakMap.has":
		if len(args) != 1 {
			return "undefined"
		}
		return fmt.Sprintf("%s.has(%s)", receiver, args[0])
	case "map.getOr", "weakMap.getOr":
		if len(args) != 2 {
			return "undefined"
		}
		return fmt.Sprintf("((__map, __key) => __map.has(__key) ? __map.get(__key)! : %s)(%s, %s)", args[1], receiver, args[0])
	case "map.set", "weakMap.set":
		if len(args) != 2 {
			return "undefined"
		}
		return fmt.Sprintf("%s.set(%s, %s)", receiver, args[0], args[1])
	case "map.delete", "weakMap.delete":
		if len(args) != 1 {
			return "undefined"
		}
		return fmt.Sprintf("%s.delete(%s)", receiver, args[0])
	case "map.clear":
		return fmt.Sprintf("%s.clear()", receiver)
	case "map.keys":
		return fmt.Sprintf("Array.from(%s.keys())", receiver)
	case "map.values":
		return fmt.Sprintf("Array.from(%s.values())", receiver)
	case "map.forEach":
		if len(args) != 1 {
			return "undefined"
		}
		return fmt.Sprintf("%s.forEach(%s)", receiver, args[0])
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}

func (g *generator) setIntrinsicCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "set.size":
		return receiver + ".size"
	case "set.has", "weakSet.has":
		if len(args) != 1 {
			return "undefined"
		}
		return fmt.Sprintf("%s.has(%s)", receiver, args[0])
	case "set.add", "weakSet.add":
		if len(args) != 1 {
			return "undefined"
		}
		return fmt.Sprintf("%s.add(%s)", receiver, args[0])
	case "set.delete", "weakSet.delete":
		if len(args) != 1 {
			return "undefined"
		}
		return fmt.Sprintf("%s.delete(%s)", receiver, args[0])
	case "set.clear":
		return fmt.Sprintf("%s.clear()", receiver)
	case "set.values":
		return fmt.Sprintf("Array.from(%s.values())", receiver)
	case "set.forEach":
		if len(args) != 1 {
			return "undefined"
		}
		return fmt.Sprintf("%s.forEach(%s)", receiver, args[0])
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}
