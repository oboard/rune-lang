package gocodegen

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
	g.addError(fmt.Errorf("Go backend does not support intrinsic %s", fn.Intrinsic))
	return g.zeroValue(typ)
}

func (g *generator) moduleIntrinsicCall(call *ir.CallExpr) (string, bool) {
	fn, ok := g.stdlibFunctionFromCall(call)
	if !ok || (fn.Intrinsic == "" && fn.Go == nil) {
		return "", false
	}
	if fn.Go != nil && fn.Go.Symbol != "" {
		return fmt.Sprintf("%s(%s)", fn.Go.Symbol, strings.Join(g.intrinsicArgs(call.Args), ", ")), true
	}
	args := g.intrinsicArgs(call.Args)
	switch fn.Intrinsic {
	case "int4.fromInt", "int8.fromInt", "int16.fromInt", "int64.fromInt",
		"uint.fromInt", "uint8.fromInt", "uint16.fromInt", "uint64.fromInt",
		"float.fromDouble", "int4.toInt", "int8.toInt", "int16.toInt", "int64.toInt",
		"uint.toInt", "uint8.toInt", "uint16.toInt", "uint64.toInt", "float.toDouble":
		return g.numericIntrinsicCall(fn, args, call.ResultType()), true
	case "binary.new", "binary.fromInts":
		return g.binaryModuleCall(fn, args, call.ResultType()), true
	case "json.stringify":
		return g.jsonStringifyCall(call)
	case "regex.new", "regex.escape":
		return g.regexModuleCall(call)
	case "map.new", "set.new":
		return g.mapModuleCall(call)
	case "go.import", "go.stmt", "go.expr":
		return g.goFFICall(call)
	default:
		return g.unsupportedIntrinsic(fn, call.ResultType()), true
	}
}

func (g *generator) receiverIntrinsicCall(call *ir.CallExpr) (string, bool) {
	sel, fn, ok := g.stdlibReceiverFunctionFromCall(call)
	if !ok || fn.Intrinsic == "" {
		return "", false
	}
	switch {
	case strings.HasPrefix(fn.Intrinsic, "array."):
		return g.arrayMethodCall(call)
	case strings.HasPrefix(fn.Intrinsic, "map."), strings.HasPrefix(fn.Intrinsic, "weakMap."), strings.HasPrefix(fn.Intrinsic, "set."), strings.HasPrefix(fn.Intrinsic, "weakSet."):
		return g.mapMethodCall(call)
	case strings.HasPrefix(fn.Intrinsic, "string."), strings.HasPrefix(fn.Intrinsic, "bool."), strings.HasPrefix(fn.Intrinsic, "regex."):
		return g.primitiveIntrinsicCall(fn, g.expr(sel.Receiver), g.intrinsicArgs(call.Args), call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "binary."):
		return g.binaryReceiverCall(fn, g.expr(sel.Receiver), g.intrinsicArgs(call.Args), call.ResultType()), true
	default:
		return g.unsupportedIntrinsic(fn, call.ResultType()), true
	}
}

func (g *generator) numericIntrinsicCall(fn *stdlib.Function, args []string, resultType checker.Type) string {
	if len(args) != 1 {
		return g.zeroValue(resultType)
	}
	value := args[0]
	switch fn.Intrinsic {
	case "int4.fromInt":
		return fmt.Sprintf("func() int8 { n := (%s) & 0xf; if n >= 8 { return int8(n - 16) }; return int8(n) }()", value)
	case "int8.fromInt":
		return fmt.Sprintf("int8(%s)", value)
	case "int16.fromInt":
		return fmt.Sprintf("int16(%s)", value)
	case "int64.fromInt":
		return fmt.Sprintf("int64(%s)", value)
	case "uint.fromInt":
		return fmt.Sprintf("uint(%s)", value)
	case "uint8.fromInt":
		return fmt.Sprintf("uint8(%s)", value)
	case "uint16.fromInt":
		return fmt.Sprintf("uint16(%s)", value)
	case "uint64.fromInt":
		return fmt.Sprintf("uint64(%s)", value)
	case "float.fromDouble":
		return fmt.Sprintf("float32(%s)", value)
	case "int4.toInt", "int8.toInt", "int16.toInt", "int64.toInt", "uint.toInt", "uint8.toInt", "uint16.toInt", "uint64.toInt":
		return fmt.Sprintf("int(%s)", value)
	case "float.toDouble":
		return fmt.Sprintf("float64(%s)", value)
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}

func (g *generator) binaryModuleCall(fn *stdlib.Function, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "binary.new":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("newRuneBinary(%s)", args[0])
	case "binary.fromInts":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("runeBinaryFromInts(%s)", args[0])
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}

func (g *generator) binaryReceiverCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "binary.length":
		return fmt.Sprintf("%s.ByteLength()", receiver)
	case "binary.clone":
		return fmt.Sprintf("%s.Clone()", receiver)
	case "binary.slice":
		if len(args) != 2 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("%s.Slice(%s, %s)", receiver, args[0], args[1])
	case "binary.toInts":
		return fmt.Sprintf("%s.ToInts()", receiver)
	case "binary.getInt4":
		return fmt.Sprintf("%s.GetInt4(%s)", receiver, args[0])
	case "binary.setInt4":
		return fmt.Sprintf("%s.SetInt4(%s, %s)", receiver, args[0], args[1])
	}
	methods := map[string]string{
		"binary.getInt8":   "GetInt8",
		"binary.getUInt8":  "GetUInt8",
		"binary.getInt16":  "GetInt16",
		"binary.getUInt16": "GetUInt16",
		"binary.getInt":    "GetInt",
		"binary.getUInt":   "GetUInt",
		"binary.getInt64":  "GetInt64",
		"binary.getUInt64": "GetUInt64",
		"binary.getFloat":  "GetFloat",
		"binary.getDouble": "GetDouble",
		"binary.setInt8":   "SetInt8",
		"binary.setUInt8":  "SetUInt8",
		"binary.setInt16":  "SetInt16",
		"binary.setUInt16": "SetUInt16",
		"binary.setInt":    "SetInt",
		"binary.setUInt":   "SetUInt",
		"binary.setInt64":  "SetInt64",
		"binary.setUInt64": "SetUInt64",
		"binary.setFloat":  "SetFloat",
		"binary.setDouble": "SetDouble",
	}
	method, ok := methods[fn.Intrinsic]
	if !ok {
		return g.unsupportedIntrinsic(fn, resultType)
	}
	if strings.HasPrefix(method, "Get") {
		return fmt.Sprintf("%s.%s(%s)", receiver, method, strings.Join(args, ", "))
	}
	return fmt.Sprintf("%s.%s(%s)", receiver, method, strings.Join(args, ", "))
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

func (g *generator) primitiveIntrinsicCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "string.length":
		return fmt.Sprintf("len([]rune(%s))", receiver)
	case "string.toString":
		return receiver
	case "string.at":
		if len(args) != 1 {
			return "/* invalid string.at */"
		}
		return fmt.Sprintf("string([]rune(%s)[%s])", receiver, args[0])
	case "string.slice":
		if len(args) != 2 {
			return "/* invalid string.slice */"
		}
		return fmt.Sprintf("func() string { runes := []rune(%s); return string(runes[%s:%s]) }()", receiver, args[0], args[1])
	case "string.concat":
		if len(args) != 1 {
			return "/* invalid string.concat */"
		}
		return fmt.Sprintf("%s + %s", receiver, args[0])
	case "string.includes":
		return fmt.Sprintf("strings.Contains(%s, %s)", receiver, args[0])
	case "string.startsWith":
		return fmt.Sprintf("strings.HasPrefix(%s, %s)", receiver, args[0])
	case "string.endsWith":
		return fmt.Sprintf("strings.HasSuffix(%s, %s)", receiver, args[0])
	case "string.indexOf":
		return fmt.Sprintf("strings.Index(%s, %s)", receiver, args[0])
	case "string.lastIndexOf":
		return fmt.Sprintf("strings.LastIndex(%s, %s)", receiver, args[0])
	case "string.toLowerCase":
		return fmt.Sprintf("strings.ToLower(%s)", receiver)
	case "string.toUpperCase":
		return fmt.Sprintf("strings.ToUpper(%s)", receiver)
	case "string.trim":
		return fmt.Sprintf("strings.TrimSpace(%s)", receiver)
	case "string.trimStart":
		return fmt.Sprintf("strings.TrimLeftFunc(%s, unicode.IsSpace)", receiver)
	case "string.trimEnd":
		return fmt.Sprintf("strings.TrimRightFunc(%s, unicode.IsSpace)", receiver)
	case "string.repeat":
		return fmt.Sprintf("strings.Repeat(%s, %s)", receiver, args[0])
	case "string.replace":
		return fmt.Sprintf("strings.Replace(%s, %s, %s, 1)", receiver, args[0], args[1])
	case "string.replaceAll":
		return fmt.Sprintf("strings.ReplaceAll(%s, %s, %s)", receiver, args[0], args[1])
	case "string.split":
		return fmt.Sprintf("func() []string { parts := strings.Split(%s, %s); return parts }()", receiver, args[0])
	case "bool.not":
		return "!" + receiver
	case "bool.xor":
		if len(args) != 1 {
			return "/* invalid bool.xor */"
		}
		return fmt.Sprintf("%s != %s", receiver, args[0])
	case "bool.toString":
		return fmt.Sprintf("strconv.FormatBool(%s)", receiver)
	case "regex.exec", "regex.match", "regex.matchAll", "regex.test", "regex.replace", "regex.replaceAll", "regex.search", "regex.split":
		name := strings.TrimPrefix(fn.Intrinsic, "regex.")
		return fmt.Sprintf("%s.%s(%s)", receiver, name, strings.Join(args, ", "))
	case "regex.source":
		return receiver + ".source"
	case "regex.flags":
		return receiver + ".flags"
	case "regex.global":
		return fmt.Sprintf("regexHasFlag(%s.flags, 'g')", receiver)
	case "regex.ignoreCase":
		return fmt.Sprintf("regexHasFlag(%s.flags, 'i')", receiver)
	case "regex.multiline":
		return fmt.Sprintf("regexHasFlag(%s.flags, 'm')", receiver)
	case "regex.dotAll":
		return fmt.Sprintf("regexHasFlag(%s.flags, 's')", receiver)
	case "regex.unicode":
		return fmt.Sprintf("regexHasFlag(%s.flags, 'u')", receiver)
	case "regex.unicodeSets":
		return fmt.Sprintf("regexHasFlag(%s.flags, 'v')", receiver)
	case "regex.sticky":
		return fmt.Sprintf("regexHasFlag(%s.flags, 'y')", receiver)
	case "regex.hasIndices":
		return fmt.Sprintf("regexHasFlag(%s.flags, 'd')", receiver)
	case "regex.lastIndex":
		return receiver + ".lastIndex"
	case "regex.setLastIndex":
		if len(args) != 1 {
			return "/* invalid regex.setLastIndex */"
		}
		return fmt.Sprintf("func() int { %s.lastIndex = %s; return %s.lastIndex }()", receiver, args[0], receiver)
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}
