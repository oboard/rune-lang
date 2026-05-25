package gocodegen

import (
	"fmt"
	"strings"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
)

func (g *generator) mapModuleCall(call *ir.CallExpr) (string, bool) {
	fn, ok := g.stdlibFunctionFromCall(call)
	if !ok {
		return "", false
	}
	switch fn.Intrinsic {
	case "map.new", "set.new":
		return goType(call.ResultType()) + "{}", true
	default:
		return "", false
	}
}

func (g *generator) mapMethodCall(call *ir.CallExpr) (string, bool) {
	sel, ok := call.Callee.(*ir.SelectorExpr)
	if !ok {
		return "", false
	}
	base, args, ok := parseGoGenericType(string(sel.Receiver.ResultType()))
	if !ok {
		return "", false
	}
	switch base {
	case "Map", "WeakMap":
		return g.mapReceiverCall(base, args, sel, call), true
	case "Set", "WeakSet":
		return g.setReceiverCall(base, args, sel, call), true
	default:
		return "", false
	}
}

func (g *generator) mapReceiverCall(base string, typeArgs []string, sel *ir.SelectorExpr, call *ir.CallExpr) string {
	if len(typeArgs) != 2 {
		return "/* invalid map type */"
	}
	fn, ok := g.file.Stdlib.ReceiverFunction("map", base, sel.Name)
	if !ok {
		return "/* unsupported map method */"
	}
	receiver := g.expr(sel.Receiver)
	args := make([]string, 0, len(call.Args))
	for _, arg := range call.Args {
		args = append(args, g.expr(arg))
	}
	keyType := goType(checker.Type(typeArgs[0]))
	valueType := goType(checker.Type(typeArgs[1]))
	mapType := goType(sel.Receiver.ResultType())
	switch fn.Intrinsic {
	case "map.size":
		return fmt.Sprintf("len(%s)", receiver)
	case "map.has", "weakMap.has":
		if len(args) != 1 {
			return "/* invalid map.has */"
		}
		return fmt.Sprintf("func() bool { _, ok := %s[%s]; return ok }()", receiver, args[0])
	case "map.getOr", "weakMap.getOr":
		if len(args) != 2 {
			return "/* invalid map.getOr */"
		}
		return fmt.Sprintf("func() %s { value, ok := %s[%s]; if ok { return value }; return %s }()", valueType, receiver, args[0], args[1])
	case "map.set", "weakMap.set":
		if len(args) != 2 {
			return "/* invalid map.set */"
		}
		return fmt.Sprintf("func() %s { %s[%s] = %s; return %s }()", mapType, receiver, args[0], args[1], receiver)
	case "map.delete", "weakMap.delete":
		if len(args) != 1 {
			return "/* invalid map.delete */"
		}
		return fmt.Sprintf("func() bool { _, ok := %s[%s]; delete(%s, %s); return ok }()", receiver, args[0], receiver, args[0])
	case "map.clear":
		return fmt.Sprintf("func() { clear(%s) }()", receiver)
	case "map.keys":
		return fmt.Sprintf("func() []%s { out := make([]%s, 0, len(%s)); for key := range %s { out = append(out, key) }; return out }()", keyType, keyType, receiver, receiver)
	case "map.values":
		return fmt.Sprintf("func() []%s { out := make([]%s, 0, len(%s)); for _, value := range %s { out = append(out, value) }; return out }()", valueType, valueType, receiver, receiver)
	case "map.forEach":
		return g.mapForEachExpr(receiver, call)
	default:
		return g.unsupportedIntrinsic(fn, call.ResultType())
	}
}

func (g *generator) setReceiverCall(base string, typeArgs []string, sel *ir.SelectorExpr, call *ir.CallExpr) string {
	if len(typeArgs) != 1 {
		return "/* invalid set type */"
	}
	fn, ok := g.file.Stdlib.ReceiverFunction("set", base, sel.Name)
	if !ok {
		return "/* unsupported set method */"
	}
	receiver := g.expr(sel.Receiver)
	args := make([]string, 0, len(call.Args))
	for _, arg := range call.Args {
		args = append(args, g.expr(arg))
	}
	elemType := goType(checker.Type(typeArgs[0]))
	setType := goType(sel.Receiver.ResultType())
	switch fn.Intrinsic {
	case "set.size":
		return fmt.Sprintf("len(%s)", receiver)
	case "set.has", "weakSet.has":
		if len(args) != 1 {
			return "/* invalid set.has */"
		}
		return fmt.Sprintf("func() bool { _, ok := %s[%s]; return ok }()", receiver, args[0])
	case "set.add", "weakSet.add":
		if len(args) != 1 {
			return "/* invalid set.add */"
		}
		return fmt.Sprintf("func() %s { %s[%s] = struct{}{}; return %s }()", setType, receiver, args[0], receiver)
	case "set.delete", "weakSet.delete":
		if len(args) != 1 {
			return "/* invalid set.delete */"
		}
		return fmt.Sprintf("func() bool { _, ok := %s[%s]; delete(%s, %s); return ok }()", receiver, args[0], receiver, args[0])
	case "set.clear":
		return fmt.Sprintf("func() { clear(%s) }()", receiver)
	case "set.values":
		return fmt.Sprintf("func() []%s { out := make([]%s, 0, len(%s)); for value := range %s { out = append(out, value) }; return out }()", elemType, elemType, receiver, receiver)
	case "set.forEach":
		return g.setForEachExpr(receiver, call)
	default:
		return g.unsupportedIntrinsic(fn, call.ResultType())
	}
}

func (g *generator) mapForEachExpr(receiver string, call *ir.CallExpr) string {
	if len(call.Args) != 1 {
		return "/* invalid map.forEach */"
	}
	lambda, ok := call.Args[0].(*ir.LambdaExpr)
	if !ok || len(lambda.Params) == 0 || len(lambda.Params) > 3 {
		return "/* invalid map.forEach */"
	}
	callArgs := []string{"value"}
	if len(lambda.Params) >= 2 {
		callArgs = append(callArgs, "key")
	}
	if len(lambda.Params) >= 3 {
		callArgs = append(callArgs, receiver)
	}
	return fmt.Sprintf("func() { for key, value := range %s { %s(%s) } }()", receiver, g.expr(lambda), strings.Join(callArgs, ", "))
}

func (g *generator) setForEachExpr(receiver string, call *ir.CallExpr) string {
	if len(call.Args) != 1 {
		return "/* invalid set.forEach */"
	}
	lambda, ok := call.Args[0].(*ir.LambdaExpr)
	if !ok || len(lambda.Params) == 0 || len(lambda.Params) > 3 {
		return "/* invalid set.forEach */"
	}
	callArgs := []string{"value"}
	if len(lambda.Params) >= 2 {
		callArgs = append(callArgs, "value")
	}
	if len(lambda.Params) >= 3 {
		callArgs = append(callArgs, receiver)
	}
	return fmt.Sprintf("func() { for value := range %s { %s(%s) } }()", receiver, g.expr(lambda), strings.Join(callArgs, ", "))
}
