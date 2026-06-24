package moonbitcodegen

import (
	"fmt"
	"strings"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func (g *generator) moduleIntrinsicCall(call *ir.CallExpr) (string, bool) {
	fn, ok := g.stdlibFunctionFromCall(call)
	if !ok {
		return "", false
	}
	if sel, ok := call.Callee.(*ir.SelectorExpr); ok {
		if at, ok := sel.Receiver.(*ir.AtExpr); ok && at.Name == "iter" && fn.Intrinsic == "" {
			return g.iterModuleCall(fn, call.Args, g.intrinsicArgs(call.Args), call.ResultType()), true
		}
	}
	args := g.intrinsicArgs(call.Args)
	if fn.Intrinsic == "" {
		if sel, ok := call.Callee.(*ir.SelectorExpr); ok {
			if at, ok := sel.Receiver.(*ir.AtExpr); ok && at.Name == "cli" && fn.Body != nil {
				return g.cliModuleCall(fn, args, call.ResultType()), true
			}
		}
		return "", false
	}
	switch fn.Intrinsic {
	case "io.print":
		return "print(" + g.printArgs(call.Args) + ")", true
	case "io.println":
		return "println(" + g.printArgs(call.Args) + ")", true
	case "io.printf":
		return "println(" + g.printArgs(call.Args) + ")", true
	case "io.scan", "io.scanLine":
		return "None", true
	case "io.readAll":
		return quoteString(""), true
	case "int.toString", "int4.toInt", "int8.toInt", "int16.toInt", "int64.toInt",
		"uint.toInt", "uint8.toInt", "uint16.toInt", "uint64.toInt",
		"float.toDouble", "float.fromDouble":
		if len(args) == 1 {
			return args[0], true
		}
	case "int4.fromInt", "int8.fromInt", "int16.fromInt", "int64.fromInt",
		"uint.fromInt", "uint8.fromInt", "uint16.fromInt", "uint64.fromInt":
		if len(args) == 1 {
			return args[0], true
		}
	case "map.new":
		return "{}", true
	case "set.new":
		return "Set::new()", true
	case "stringbuffer.new", "stringbuffer.from":
		return g.stringBufferModuleCall(fn, args, call.ResultType()), true
	case "iter.new", "iter.fromArray", "iter.range", "iter.rangeStep", "iter.repeat", "iter.empty":
		return g.iterModuleCall(fn, call.Args, args, call.ResultType()), true
	case "json.stringify":
		return fmt.Sprintf("%s.to_string()", args[0]), true
	case "json.parse":
		return "()", true
	case "assert.eq":
		if len(args) == 2 {
			return fmt.Sprintf("assert_eq!(%s, %s)", args[0], args[1]), true
		}
	case "process.argv":
		return "@env.args()", true
	case "process.cwd":
		return quoteString("."), true
	case "process.env":
		return "None", true
	case "process.exit":
		return "abort(\"process.exit\")", true
	case "process.platform":
		return quoteString("moonbit"), true
	case "go.import", "go.stmt", "go.expr":
		g.addError(fmt.Errorf("MoonBit backend does not support @go FFI"))
		return zeroValue(call.ResultType()), true
	default:
		return runtimeTrap(fn.Intrinsic), true
	}
	return runtimeTrap(fn.Intrinsic), true
}

func (g *generator) printArgs(args []ir.Expr) string {
	if len(args) == 0 {
		return quoteString("")
	}
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		parts = append(parts, g.showExpr(arg))
	}
	return strings.Join(parts, " + \" \" + ")
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
	case strings.HasPrefix(fn.Intrinsic, "int."), strings.HasPrefix(fn.Intrinsic, "char."), strings.HasPrefix(fn.Intrinsic, "bool."):
		return g.primitiveIntrinsicCall(fn, receiver, args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "map."), strings.HasPrefix(fn.Intrinsic, "weakMap."):
		return g.mapIntrinsicCall(fn, receiver, args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "set."), strings.HasPrefix(fn.Intrinsic, "weakSet."):
		return g.setIntrinsicCall(fn, receiver, args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "stringbuffer."):
		return g.stringBufferIntrinsicCall(fn, receiver, args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "iter."):
		return g.iterIntrinsicCall(call, fn, receiver, args), true
	default:
		return runtimeTrap(fn.Intrinsic), true
	}
}

func (g *generator) primitiveIntrinsicCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "int.toString", "char.toString", "bool.toString":
		return receiver + ".to_string()"
	case "int.add", "int16.add", "int8.add", "int4.add", "int64.add", "uint.add", "uint16.add", "uint8.add", "uint64.add", "double.add", "float.add", "bigint.add":
		if len(args) == 1 {
			return receiver + " + " + args[0]
		}
	case "int.sub", "int16.sub", "int8.sub", "int4.sub", "int64.sub", "uint.sub", "uint16.sub", "uint8.sub", "uint64.sub", "double.sub", "float.sub", "bigint.sub":
		if len(args) == 1 {
			return receiver + " - " + args[0]
		}
	case "int.mul", "int16.mul", "int8.mul", "int4.mul", "int64.mul", "uint.mul", "uint16.mul", "uint8.mul", "uint64.mul", "double.mul", "float.mul", "bigint.mul":
		if len(args) == 1 {
			return receiver + " * " + args[0]
		}
	case "int.div", "int16.div", "int8.div", "int4.div", "int64.div", "uint.div", "uint16.div", "uint8.div", "uint64.div", "double.div", "float.div", "bigint.div":
		if len(args) == 1 {
			return receiver + " / " + args[0]
		}
	default:
		return runtimeTrap(fn.Intrinsic)
	}
	_ = resultType
	return runtimeTrap(fn.Intrinsic)
}

func (g *generator) arrayIntrinsicCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "array.len":
		return receiver + ".length()"
	case "array.push":
		return fmt.Sprintf("%s.push(%s)", receiver, strings.Join(args, ", "))
	case "array.set":
		if len(args) == 2 {
			return fmt.Sprintf("%s[%s] = %s", receiver, args[0], args[1])
		}
	case "array.pop":
		return receiver + ".pop()"
	case "array.first":
		return receiver + "[0]"
	case "array.last":
		return fmt.Sprintf("%s[%s.length() - 1]", receiver, receiver)
	case "array.slice":
		if len(args) == 2 {
			return fmt.Sprintf("%s[%s:%s].to_owned()", receiver, args[0], args[1])
		}
	case "array.clone":
		return receiver + ".copy()"
	case "array.reverse":
		return receiver + ".rev()"
	case "array.contains":
		if len(args) == 1 {
			return fmt.Sprintf("%s.contains(%s)", receiver, args[0])
		}
	case "array.at":
		if len(args) == 1 {
			return fmt.Sprintf("%s[%s]", receiver, args[0])
		}
	case "array.map", "array.each":
		if len(args) == 1 {
			return fmt.Sprintf("%s.%s(%s)", receiver, strings.TrimPrefix(fn.Intrinsic, "array."), args[0])
		}
	}
	return runtimeTrap(fn.Intrinsic)
}

func (g *generator) stringIntrinsicCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "string.length":
		return receiver + ".length()"
	case "string.toString":
		return receiver
	case "string.at":
		if len(args) == 1 {
			if resultType == checker.Char {
				return fmt.Sprintf("%s[%s].unsafe_to_char()", receiver, args[0])
			}
			return fmt.Sprintf("%s[%s]", receiver, args[0])
		}
	case "string.slice":
		if len(args) == 2 {
			return fmt.Sprintf("%s[%s:%s].to_owned()", receiver, args[0], args[1])
		}
	case "string.concat":
		if len(args) == 1 {
			return receiver + " + " + args[0]
		}
	case "string.includes":
		if len(args) == 1 {
			return fmt.Sprintf("%s.contains(%s)", receiver, args[0])
		}
	case "string.startsWith":
		if len(args) == 1 {
			return fmt.Sprintf("%s.has_prefix(%s)", receiver, args[0])
		}
	case "string.endsWith":
		if len(args) == 1 {
			return fmt.Sprintf("%s.has_suffix(%s)", receiver, args[0])
		}
	case "string.toLowerCase":
		return receiver + ".to_lower()"
	case "string.toUpperCase":
		return receiver + ".to_upper()"
	case "string.trim":
		return receiver + ".trim().to_owned()"
	case "string.trimStart":
		return receiver + ".trim_start().to_owned()"
	case "string.trimEnd":
		return receiver + ".trim_end().to_owned()"
	case "string.indexOf":
		if len(args) == 1 {
			return fmt.Sprintf("%s.find(%s).unwrap_or(-1)", receiver, args[0])
		}
	case "string.lastIndexOf":
		if len(args) == 1 {
			return fmt.Sprintf("%s.rev_find(%s).unwrap_or(-1)", receiver, args[0])
		}
	case "string.repeat":
		if len(args) == 1 {
			return fmt.Sprintf("%s.repeat(%s)", receiver, args[0])
		}
	case "string.replace":
		if len(args) == 2 {
			return fmt.Sprintf("%s.replace(old=%s, new=%s)", receiver, args[0], args[1])
		}
	case "string.replaceAll":
		if len(args) == 2 {
			return fmt.Sprintf("%s.replace_all(old=%s, new=%s)", receiver, args[0], args[1])
		}
	case "string.split":
		if len(args) == 1 {
			return fmt.Sprintf("%s.split(%s).map(fn(x) { x.to_owned() }).collect()", receiver, args[0])
		}
	}
	return runtimeTrap(fn.Intrinsic)
}

func (g *generator) mapIntrinsicCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "map.size":
		return receiver + ".length()"
	case "map.has":
		return fmt.Sprintf("%s.contains(%s)", receiver, args[0])
	case "map.getOr", "weakMap.getOr":
		if len(args) == 2 {
			return fmt.Sprintf("%s.get_or_default(%s, %s)", receiver, args[0], args[1])
		}
	case "map.get":
		if len(args) == 1 {
			return fmt.Sprintf("%s[%s]", receiver, args[0])
		}
	case "map.set":
		if len(args) == 2 {
			return fmt.Sprintf("{ %s[%s] = %s; %s }", receiver, args[0], args[1], receiver)
		}
	case "map.delete", "weakMap.delete":
		if len(args) == 1 {
			return fmt.Sprintf("{ let __had = %s.contains(%s); %s.remove(%s); __had }", receiver, args[0], receiver, args[0])
		}
	case "map.clear":
		return receiver + ".clear()"
	case "map.keys":
		return receiver + ".keys().collect()"
	case "map.values":
		return receiver + ".values().collect()"
	case "map.each":
		if len(args) == 1 {
			return fmt.Sprintf("%s.each((k, v) => { %s(v, k, %s) })", receiver, args[0], receiver)
		}
	case "weakMap.has":
		if len(args) == 1 {
			return fmt.Sprintf("%s.contains(%s)", receiver, args[0])
		}
	case "weakMap.set":
		if len(args) == 2 {
			return fmt.Sprintf("{ %s[%s] = %s; %s }", receiver, args[0], args[1], receiver)
		}
	}
	return runtimeTrap(fn.Intrinsic)
}

func (g *generator) setIntrinsicCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "set.size":
		return receiver + ".length()"
	case "set.has", "weakSet.has":
		return fmt.Sprintf("%s.contains(%s)", receiver, args[0])
	case "set.add", "weakSet.add":
		return fmt.Sprintf("{ %s.add(%s); %s }", receiver, args[0], receiver)
	case "set.delete", "weakSet.delete":
		return fmt.Sprintf("{ let __had = %s.contains(%s); %s.remove(%s); __had }", receiver, args[0], receiver, args[0])
	case "set.clear":
		return receiver + ".clear()"
	case "set.values":
		return receiver + ".to_array()"
	case "set.each":
		if len(args) == 1 {
			return fmt.Sprintf("%s.each((v) => { %s(v, v, %s) })", receiver, args[0], receiver)
		}
	}
	return runtimeTrap(fn.Intrinsic)
}

func (g *generator) stringBufferModuleCall(fn *stdlib.Function, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "stringbuffer.new":
		return "StringBuilder()"
	case "stringbuffer.from":
		if len(args) == 1 {
			buf := g.nextTemp("__buf")
			return fmt.Sprintf("{ let %s = StringBuilder(); %s.write_string(%s); %s }", buf, buf, args[0], buf)
		}
	}
	_ = resultType
	return runtimeTrap(fn.Intrinsic)
}

func (g *generator) stringBufferIntrinsicCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "stringbuffer.length":
		return receiver + ".to_string().length()"
	case "stringbuffer.clear":
		return receiver + ".reset()"
	case "stringbuffer.append":
		if len(args) == 1 {
			return fmt.Sprintf("{ %s.write_string(%s); %s }", receiver, args[0], receiver)
		}
	case "stringbuffer.appendLine":
		if len(args) == 1 {
			return fmt.Sprintf("{ %s.write_string(%s); %s.write_char('\\n'); %s }", receiver, args[0], receiver, receiver)
		}
	case "stringbuffer.toString":
		return receiver + ".to_string()"
	}
	_ = resultType
	return runtimeTrap(fn.Intrinsic)
}

func (g *generator) iterModuleCall(fn *stdlib.Function, rawArgs []ir.Expr, args []string, resultType checker.Type) string {
	elem, ok := checker.IterValue(resultType)
	if !ok {
		elem = checker.Unknown
	}
	switch fn.Name {
	case "new":
		if len(args) != 1 {
			return zeroValue(resultType)
		}
		item := g.nextTemp("__item")
		return fmt.Sprintf("Iter::new(fn() { let %s = %s(); if %s.1 { Some(%s.0) } else { None } })", item, args[0], item, item)
	case "fromArray":
		if len(args) != 1 {
			return zeroValue(resultType)
		}
		values := g.nextTemp("__values")
		index := g.nextTemp("__index")
		return fmt.Sprintf("{ let %s = %s; let mut %s = -1; Iter::new(fn() { %s = %s + 1; if %s >= %s.length() { None } else { Some(%s[%s]) } }) }", values, args[0], index, index, index, index, values, values, index)
	case "range":
		if len(args) != 2 {
			return zeroValue(resultType)
		}
		return g.iterRangeExpr(args[0], args[1], "1")
	case "rangeStep":
		if len(args) != 3 {
			return zeroValue(resultType)
		}
		return g.iterRangeExpr(args[0], args[1], args[2])
	case "repeat":
		if len(args) != 2 {
			return zeroValue(resultType)
		}
		value := g.nextTemp("__value")
		count := g.nextTemp("__count")
		index := g.nextTemp("__index")
		return fmt.Sprintf("{ let %s = %s; let %s = %s; let mut %s = -1; Iter::new(fn() { %s = %s + 1; if %s < %s { Some(%s) } else { None } }) }", value, args[0], count, args[1], index, index, index, index, count, value)
	case "empty":
		if len(args) != 1 {
			return zeroValue(resultType)
		}
		return fmt.Sprintf("{ ignore(%s); Iter::new(fn() { None }) }", args[0])
	}
	_ = rawArgs
	_ = elem
	return runtimeTrap(fn.Intrinsic)
}

func (g *generator) iterRangeExpr(start string, end string, step string) string {
	startName := g.nextTemp("__start")
	endName := g.nextTemp("__end")
	stepName := g.nextTemp("__step")
	value := g.nextTemp("__value")
	return fmt.Sprintf("{ let %s = %s; let %s = %s; let %s = %s; let mut %s = %s - %s; Iter::new(fn() { %s = %s + %s; if %s == 0 { None } else if %s > 0 { if %s < %s { Some(%s) } else { None } } else if %s > %s { Some(%s) } else { None } }) }", startName, start, endName, end, stepName, step, value, startName, stepName, value, value, stepName, stepName, stepName, value, endName, value, value, endName, value)
}

func (g *generator) iterIntrinsicCall(call *ir.CallExpr, fn *stdlib.Function, receiver string, args []string) string {
	switch fn.Intrinsic {
	case "iter.next":
		elem := checker.Unknown
		if value, ok := checker.IterValue(call.Callee.(*ir.SelectorExpr).Receiver.ResultType()); ok {
			elem = value
		}
		item := g.nextTemp("__item")
		return fmt.Sprintf("{ let %s = %s.next(); match %s { Some(value) => (value, true); None => (%s, false) } }", item, receiver, item, zeroValue(elem))
	case "iter.toArray":
		return receiver + ".to_array()"
	case "iter.each":
		if len(args) == 1 {
			callback, arity := g.iterCallback(call.Args[0])
			return g.iterEachExpr(receiver, callback, arity)
		}
	case "iter.map":
		if len(args) == 1 {
			callback, arity := g.iterCallback(call.Args[0])
			return g.iterMapExpr(receiver, callback, arity)
		}
	}
	return runtimeTrap(fn.Intrinsic)
}

func (g *generator) iterEachExpr(receiver string, callback string, arity int) string {
	switch arity {
	case 0:
		return fmt.Sprintf("%s.each(fn(_) { %s() })", receiver, callback)
	case 1:
		return fmt.Sprintf("%s.each(%s)", receiver, callback)
	case 2:
		return fmt.Sprintf("%s.eachi(fn(index, value) { %s(value, index) })", receiver, callback)
	default:
		iter := g.nextTemp("__iter")
		return fmt.Sprintf("{ let %s = %s; %s.eachi(fn(index, value) { %s(value, index, %s) }) }", iter, receiver, iter, callback, iter)
	}
}

func (g *generator) iterMapExpr(receiver string, callback string, arity int) string {
	switch arity {
	case 0:
		return fmt.Sprintf("%s.map(fn(_) { %s() }).to_array()", receiver, callback)
	case 1:
		return fmt.Sprintf("%s.map(%s).to_array()", receiver, callback)
	case 2:
		return fmt.Sprintf("%s.mapi(fn(index, value) { %s(value, index) }).to_array()", receiver, callback)
	default:
		iter := g.nextTemp("__iter")
		return fmt.Sprintf("{ let %s = %s; %s.mapi(fn(index, value) { %s(value, index, %s) }).to_array() }", iter, receiver, iter, callback, iter)
	}
}

func (g *generator) iterCallback(expr ir.Expr) (string, int) {
	if lambda, ok := expr.(*ir.LambdaExpr); ok {
		return g.expr(lambda), len(lambda.Params)
	}
	params, _, ok := parseFuncType(string(expr.ResultType()))
	if !ok {
		return g.expr(expr), 3
	}
	return g.expr(expr), len(params)
}

func runtimeTrap(name string) string {
	return fmt.Sprintf("abort(%q)", "MoonBit intrinsic "+name+" is not implemented")
}

func (g *generator) stdlibFunctionFromCall(call *ir.CallExpr) (*stdlib.Function, bool) {
	sel, ok := call.Callee.(*ir.SelectorExpr)
	if !ok || g.file.Stdlib == nil {
		return nil, false
	}
	if at, ok := sel.Receiver.(*ir.AtExpr); ok && at.Name != "" {
		return g.file.Stdlib.Function(at.Name, sel.Name)
	}
	if moduleName, ok := checker.ModuleNamespaceName(sel.Receiver.ResultType()); ok {
		return g.file.Stdlib.Function(moduleName, sel.Name)
	}
	return nil, false
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
