package moonbitcodegen

import (
	"fmt"
	"strings"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/codegen/stdlibhelpers"
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
	if fn.Intrinsic == "" {
		args := g.intrinsicArgs(call.Args)
		if sel, ok := call.Callee.(*ir.SelectorExpr); ok {
			if at, ok := sel.Receiver.(*ir.AtExpr); ok && at.Name == "cli" && fn.Body != nil {
				return g.cliModuleCall(fn, args, call.ResultType()), true
			}
		}
		if fn.Body != nil {
			if moduleName, ok := stdlibCallModuleName(call); ok {
				return fmt.Sprintf("%s(%s)", mangleIdent(stdlibhelpers.HelperName(moduleName, fn.Name)), strings.Join(args, ", ")), true
			}
		}
		return "", false
	}
	switch fn.Intrinsic {
	case "json.stringify":
		return g.jsonStringifyCall(call)
	case "json.parse":
		return g.jsonParseCall(call)
	}
	args := g.intrinsicArgs(call.Args)
	switch fn.Intrinsic {
	case "io.print":
		return "println(" + g.printArgs(call.Args) + ")", true
	case "io.println":
		return "println(" + g.printArgs(call.Args) + ")", true
	case "io.printf":
		return "println(" + g.printArgs(call.Args) + ")", true
	case "io.scan", "io.scanLine":
		return "None", true
	case "io.readAll":
		return quoteString(""), true
	case "int.toString", "bigint.toString":
		if len(args) == 1 {
			return fmt.Sprintf("(%s).to_string()", args[0]), true
		}
	case "int.toDouble", "float.toDouble":
		if len(args) == 1 {
			return fmt.Sprintf("(%s).to_double()", args[0]), true
		}
	case "bigint.toDouble":
		if len(args) == 1 {
			return fmt.Sprintf("(%s).to_int().to_double()", args[0]), true
		}
	case "float.fromDouble":
		if len(args) == 1 {
			return fmt.Sprintf("Float::from_double(%s)", args[0]), true
		}
	case "int4.fromInt":
		if len(args) == 1 {
			return fmt.Sprintf("{ let __n = (%s) & 15; if __n >= 8 { __n - 16 } else { __n } }", args[0]), true
		}
	case "int8.fromInt":
		if len(args) == 1 {
			return fmt.Sprintf("{ let __n = (%s) & 255; if __n >= 128 { __n - 256 } else { __n } }", args[0]), true
		}
	case "int16.fromInt":
		if len(args) == 1 {
			return fmt.Sprintf("{ let __n = (%s) & 65535; if __n >= 32768 { __n - 65536 } else { __n } }", args[0]), true
		}
	case "int64.fromInt":
		if len(args) == 1 {
			return args[0], true
		}
	case "uint.fromInt", "uint8.fromInt", "uint16.fromInt":
		if len(args) == 1 {
			return args[0], true
		}
	case "uint64.fromInt":
		if len(args) == 1 {
			return args[0], true
		}
	case "int4.toInt", "int8.toInt", "int16.toInt":
		if len(args) == 1 {
			return args[0], true
		}
	case "uint.toInt", "uint8.toInt", "uint16.toInt":
		if len(args) == 1 {
			return args[0], true
		}
	case "int64.toInt", "uint64.toInt":
		if len(args) == 1 {
			return args[0], true
		}
	case "int.toBigInt", "bigint.fromInt":
		if len(args) == 1 {
			return fmt.Sprintf("BigInt::from_int(%s)", args[0]), true
		}
	case "double.trunc":
		if len(args) == 1 {
			return fmt.Sprintf("(%s).trunc()", args[0]), true
		}
	case "double.floor":
		if len(args) == 1 {
			return fmt.Sprintf("(%s).floor()", args[0]), true
		}
	case "double.ceil":
		if len(args) == 1 {
			return fmt.Sprintf("(%s).ceil()", args[0]), true
		}
	case "double.round":
		if len(args) == 1 {
			return fmt.Sprintf("(%s).round()", args[0]), true
		}
	case "bytes.new", "bytes.fromInts":
		return g.bytesModuleCall(fn, args, call.ResultType()), true
	case "buffer.new", "buffer.fromBytes", "reader.new", "writer.new", "writer.withCapacity":
		return g.streamModuleCall(fn, args, call.ResultType()), true
	case "fs.readFile", "fs.readFileText", "fs.writeFile", "fs.writeFileText", "fs.exists", "fs.readdir", "fs.mkdir", "fs.remove", "fs.stat":
		return g.fsModuleCall(fn, args, call.ResultType()), true
	case "compress.gzip", "compress.gunzip", "compress.gzipText", "compress.gunzipText",
		"compress.deflate", "compress.inflate", "compress.brotli", "compress.unbrotli",
		"compress.zstd", "compress.unzstd", "compress.brotliText", "compress.unbrotliText",
		"compress.zstdText", "compress.unzstdText":
		return g.compressModuleCall(fn, args, call.ResultType()), true
	case "map.new":
		return "{}", true
	case "set.new":
		return "Set::new()", true
	case "stringbuffer.new", "stringbuffer.from":
		return g.stringBufferModuleCall(fn, args, call.ResultType()), true
	case "symbol.create", "symbol.unique", "symbol.for", "symbol.keyFor", "symbol.description", "symbol.toString":
		return g.symbolModuleCall(fn, args, call.ResultType()), true
	case "regex.new", "regex.escape":
		return g.regexModuleCall(fn, args, call.ResultType()), true
	case "iter.new", "iter.fromArray", "iter.range", "iter.rangeStep", "iter.repeat", "iter.empty":
		return g.iterModuleCall(fn, call.Args, args, call.ResultType()), true
	case "assert.eq":
		if len(call.Args) == 2 {
			left := g.equalityArg(call.Args[0], call.Args[1].ResultType())
			right := g.equalityArg(call.Args[1], call.Args[0].ResultType())
			return fmt.Sprintf("if (%s) != (%s) { abort(\"assert.eq failed\") }", left, right), true
		}
	case "process.argv":
		return "@env.args()", true
	case "process.cwd":
		return "@env.current_dir().unwrap_or(\"\")", true
	case "process.env":
		if len(args) == 1 {
			return fmt.Sprintf("@env.get_env_var(%s)", args[0]), true
		}
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

func (g *generator) equalityArg(expr ir.Expr, other checker.Type) string {
	if _, ok := parseNullableType(string(other)); ok {
		if _, selfNullable := parseNullableType(string(expr.ResultType())); !selfNullable {
			return g.exprAs(expr, other)
		}
	}
	return g.expr(expr)
}

func (g *generator) symbolModuleCall(fn *stdlib.Function, args []string, resultType checker.Type) string {
	if len(args) != 1 {
		return zeroValue(resultType)
	}
	switch fn.Intrinsic {
	case "symbol.create", "symbol.unique":
		return quoteString("l:") + " + " + args[0]
	case "symbol.for":
		return quoteString("g:") + " + " + args[0]
	case "symbol.keyFor":
		return fmt.Sprintf("if %s.has_prefix(%s) { Some(%s[2:%s.length()].to_owned()) } else { None }", args[0], quoteString("g:"), args[0], args[0])
	case "symbol.description":
		return fmt.Sprintf("Some(%s[2:%s.length()].to_owned())", args[0], args[0])
	case "symbol.toString":
		return fmt.Sprintf("%s + %s[2:%s.length()].to_owned() + %s", quoteString("Symbol("), args[0], args[0], quoteString(")"))
	default:
		return runtimeTrap(fn.Intrinsic)
	}
}

func stdlibCallModuleName(call *ir.CallExpr) (string, bool) {
	sel, ok := call.Callee.(*ir.SelectorExpr)
	if !ok {
		return "", false
	}
	if at, ok := sel.Receiver.(*ir.AtExpr); ok && at.Name != "" {
		return at.Name, true
	}
	return checker.ModuleNamespaceName(sel.Receiver.ResultType())
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
	if !ok {
		return "", false
	}
	receiver := g.expr(sel.Receiver)
	args := g.intrinsicArgs(call.Args)
	if fn.Intrinsic == "" {
		if fn.Body != nil && fn.Name == "isEmpty" && len(args) == 0 {
			return receiver + ".length() == 0", true
		}
		return "", false
	}
	switch {
	case strings.HasPrefix(fn.Intrinsic, "array."):
		return g.arrayIntrinsicCall(fn, receiver, args, call.Args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "string."):
		return g.stringIntrinsicCall(fn, receiver, args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "int."), strings.HasPrefix(fn.Intrinsic, "char."), strings.HasPrefix(fn.Intrinsic, "bool."):
		return g.primitiveIntrinsicCall(fn, sel.Receiver, receiver, args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "bytes."):
		return g.bytesIntrinsicCall(fn, receiver, args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "buffer."):
		return g.bufferIntrinsicCall(fn, receiver, args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "reader."):
		return g.readerIntrinsicCall(fn, receiver, args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "writer."):
		return g.writerIntrinsicCall(fn, receiver, args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "map."), strings.HasPrefix(fn.Intrinsic, "weakMap."):
		return g.mapIntrinsicCall(fn, receiver, args, call.Args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "set."), strings.HasPrefix(fn.Intrinsic, "weakSet."):
		return g.setIntrinsicCall(fn, receiver, args, call.Args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "stringbuffer."):
		return g.stringBufferIntrinsicCall(fn, receiver, args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "regex."):
		return g.regexIntrinsicCall(fn, receiver, args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "iter."):
		return g.iterIntrinsicCall(call, fn, receiver, args), true
	default:
		return runtimeTrap(fn.Intrinsic), true
	}
}

func (g *generator) primitiveIntrinsicCall(fn *stdlib.Function, rawReceiver ir.Expr, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "int.toString", "char.toString", "bool.toString":
		return fmt.Sprintf("(%s).to_string()", receiver)
	case "bool.not":
		if rawReceiver != nil {
			return g.mbtNotExpr(rawReceiver)
		}
		return fmt.Sprintf("!(%s)", receiver)
	case "bool.xor":
		if len(args) == 1 {
			return fmt.Sprintf("(%s) != (%s)", receiver, args[0])
		}
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

func (g *generator) arrayIntrinsicCall(fn *stdlib.Function, receiver string, args []string, rawArgs []ir.Expr, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "array.len":
		return receiver + ".length()"
	case "array.push":
		if len(args) == 1 {
			target := g.nextTemp("__array")
			return fmt.Sprintf("{ let %s = %s; %s.push(%s); %s.length() }", target, receiver, target, args[0], target)
		}
	case "array.set":
		if len(args) == 2 {
			target := g.nextTemp("__array")
			index := g.nextTemp("__index")
			value := g.nextTemp("__value")
			return fmt.Sprintf("{ let %s = %s; let %s = %s; let %s = %s; %s[%s] = %s; %s }", target, receiver, index, args[0], value, args[1], target, index, value, value)
		}
	case "array.pop":
		return receiver + ".pop().unwrap()"
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
	case "array.each":
		return g.arrayEachExpr(receiver, args, rawArgs, resultType)
	case "array.map":
		return g.arrayMapExpr(receiver, args, rawArgs, resultType)
	case "array.reduce", "array.foldl":
		return g.arrayReduceExpr(receiver, args, rawArgs, resultType, false)
	case "array.foldr":
		return g.arrayReduceExpr(receiver, args, rawArgs, resultType, true)
	}
	return runtimeTrap(fn.Intrinsic)
}

func (g *generator) arrayEachExpr(receiver string, args []string, rawArgs []ir.Expr, resultType checker.Type) string {
	if len(args) != 1 || len(rawArgs) != 1 {
		return zeroValue(resultType)
	}
	callback := args[0]
	arity := callbackArity(rawArgs[0], 3)
	target := g.nextTemp("__array")
	index := g.nextTemp("__index")
	value := g.nextTemp("__value")
	call := callbackCall(callback, arity, value, index, target)
	return fmt.Sprintf("{ let %s = %s; let mut %s = 0; while %s < %s.length() { let %s = %s[%s]; ignore(%s); %s = %s + 1 }; () }", target, receiver, index, index, target, value, target, index, call, index, index)
}

func (g *generator) arrayMapExpr(receiver string, args []string, rawArgs []ir.Expr, resultType checker.Type) string {
	if len(args) != 1 || len(rawArgs) != 1 {
		return zeroValue(resultType)
	}
	elemType := checker.Unknown
	if elem, ok := checker.ArrayElement(resultType); ok {
		elemType = elem
	}
	callback := args[0]
	arity := callbackArity(rawArgs[0], 3)
	target := g.nextTemp("__array")
	result := g.nextTemp("__result")
	index := g.nextTemp("__index")
	value := g.nextTemp("__value")
	call := callbackCall(callback, arity, value, index, target)
	return fmt.Sprintf("{ let %s = %s; let %s : Array[%s] = []; let mut %s = 0; while %s < %s.length() { let %s = %s[%s]; %s.push(%s); %s = %s + 1 }; %s }", target, receiver, result, g.mbtType(elemType), index, index, target, value, target, index, result, call, index, index, result)
}

func (g *generator) arrayReduceExpr(receiver string, args []string, rawArgs []ir.Expr, resultType checker.Type, reverse bool) string {
	if len(args) != 2 || len(rawArgs) != 2 {
		return zeroValue(resultType)
	}
	callback := args[1]
	arity := callbackArity(rawArgs[1], 4)
	target := g.nextTemp("__array")
	result := g.nextTemp("__result")
	index := g.nextTemp("__index")
	value := g.nextTemp("__value")
	call := callbackCall(callback, arity, result, value, index, target)
	if reverse {
		return fmt.Sprintf("{ let %s = %s; let mut %s = %s; let mut %s = %s.length() - 1; while %s >= 0 { let %s = %s[%s]; %s = %s; %s = %s - 1 }; %s }", target, receiver, result, args[0], index, target, index, value, target, index, result, call, index, index, result)
	}
	return fmt.Sprintf("{ let %s = %s; let mut %s = %s; let mut %s = 0; while %s < %s.length() { let %s = %s[%s]; %s = %s; %s = %s + 1 }; %s }", target, receiver, result, args[0], index, index, target, value, target, index, result, call, index, index, result)
}

func callbackArity(expr ir.Expr, fallback int) int {
	if lambda, ok := expr.(*ir.LambdaExpr); ok {
		return len(lambda.Params)
	}
	params, _, ok := parseFuncType(string(expr.ResultType()))
	if !ok {
		return fallback
	}
	return len(params)
}

func callbackCall(callback string, arity int, args ...string) string {
	if arity < len(args) {
		args = args[:arity]
	}
	return fmt.Sprintf("(%s)(%s)", callback, strings.Join(args, ", "))
}

func (g *generator) stringIntrinsicCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	g.useString = true
	switch fn.Intrinsic {
	case "string.length":
		return fmt.Sprintf("rune_string_length(%s)", receiver)
	case "string.isEmpty":
		return receiver + ".is_empty()"
	case "string.toString":
		return receiver
	case "string.at":
		if len(args) == 1 {
			if resultType == checker.Char {
				return fmt.Sprintf("rune_string_at(%s, %s)", receiver, args[0])
			}
			return fmt.Sprintf("rune_string_at(%s, %s).to_string()", receiver, args[0])
		}
	case "string.slice":
		if len(args) == 2 {
			return fmt.Sprintf("rune_string_slice(%s, %s, %s)", receiver, args[0], args[1])
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

func (g *generator) regexModuleCall(fn *stdlib.Function, args []string, resultType checker.Type) string {
	g.useRegex = true
	switch fn.Intrinsic {
	case "regex.new":
		if len(args) == 2 {
			return fmt.Sprintf("rune_regex_new(%s, %s)", args[0], args[1])
		}
	case "regex.escape":
		if len(args) == 1 {
			return fmt.Sprintf("rune_regex_escape(%s)", args[0])
		}
	}
	return zeroValue(resultType)
}

func (g *generator) fsModuleCall(fn *stdlib.Function, args []string, resultType checker.Type) string {
	g.useFS = true
	switch fn.Intrinsic {
	case "fs.readFile":
		if len(args) == 1 {
			return fmt.Sprintf("rune_fs_read_file(%s)", args[0])
		}
	case "fs.readFileText":
		if len(args) == 1 {
			return fmt.Sprintf("rune_fs_read_file_text(%s)", args[0])
		}
	case "fs.writeFile":
		if len(args) == 2 {
			return fmt.Sprintf("rune_fs_write_file(%s, %s)", args[0], args[1])
		}
	case "fs.writeFileText":
		if len(args) == 2 {
			return fmt.Sprintf("rune_fs_write_file_text(%s, %s)", args[0], args[1])
		}
	case "fs.exists":
		if len(args) == 1 {
			return fmt.Sprintf("rune_fs_exists(%s)", args[0])
		}
	case "fs.readdir":
		if len(args) == 1 {
			return fmt.Sprintf("rune_fs_readdir(%s)", args[0])
		}
	case "fs.mkdir":
		if len(args) == 1 {
			return fmt.Sprintf("rune_fs_mkdir(%s)", args[0])
		}
	case "fs.remove":
		if len(args) == 1 {
			return fmt.Sprintf("rune_fs_remove(%s)", args[0])
		}
	case "fs.stat":
		if len(args) == 1 {
			return fmt.Sprintf("rune_fs_stat(%s)", args[0])
		}
	}
	return zeroValue(resultType)
}

func (g *generator) compressModuleCall(fn *stdlib.Function, args []string, resultType checker.Type) string {
	g.useCompress = true
	if len(args) != 1 {
		return zeroValue(resultType)
	}
	switch fn.Intrinsic {
	case "compress.gzip":
		return fmt.Sprintf("rune_compress_gzip(%s)", args[0])
	case "compress.gunzip":
		return fmt.Sprintf("rune_compress_gunzip(%s)", args[0])
	case "compress.deflate":
		return fmt.Sprintf("rune_compress_deflate(%s)", args[0])
	case "compress.inflate":
		return fmt.Sprintf("rune_compress_inflate(%s)", args[0])
	case "compress.brotli":
		return fmt.Sprintf("rune_compress_brotli(%s)", args[0])
	case "compress.unbrotli":
		return fmt.Sprintf("rune_compress_unbrotli(%s)", args[0])
	case "compress.zstd":
		return fmt.Sprintf("rune_compress_zstd(%s)", args[0])
	case "compress.unzstd":
		return fmt.Sprintf("rune_compress_unzstd(%s)", args[0])
	case "compress.gzipText":
		return fmt.Sprintf("rune_compress_gzip_text(%s)", args[0])
	case "compress.gunzipText":
		return fmt.Sprintf("rune_compress_gunzip_text(%s)", args[0])
	case "compress.brotliText":
		return fmt.Sprintf("rune_compress_brotli_text(%s)", args[0])
	case "compress.unbrotliText":
		return fmt.Sprintf("rune_compress_unbrotli_text(%s)", args[0])
	case "compress.zstdText":
		return fmt.Sprintf("rune_compress_zstd_text(%s)", args[0])
	case "compress.unzstdText":
		return fmt.Sprintf("rune_compress_unzstd_text(%s)", args[0])
	default:
		return runtimeTrap(fn.Intrinsic)
	}
}

func (g *generator) regexIntrinsicCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	g.useRegex = true
	switch fn.Intrinsic {
	case "regex.exec":
		if len(args) == 1 {
			return fmt.Sprintf("rune_regex_exec(%s, %s)", receiver, args[0])
		}
	case "regex.match":
		if len(args) == 1 {
			return fmt.Sprintf("rune_regex_match(%s, %s)", receiver, args[0])
		}
	case "regex.matchAll":
		if len(args) == 1 {
			return fmt.Sprintf("rune_regex_match_all(%s, %s)", receiver, args[0])
		}
	case "regex.test":
		if len(args) == 1 {
			return fmt.Sprintf("rune_regex_test(%s, %s)", receiver, args[0])
		}
	case "regex.replace":
		if len(args) == 2 {
			return fmt.Sprintf("rune_regex_replace(%s, %s, %s, false)", receiver, args[0], args[1])
		}
	case "regex.replaceAll":
		if len(args) == 2 {
			return fmt.Sprintf("rune_regex_replace(%s, %s, %s, true)", receiver, args[0], args[1])
		}
	case "regex.search":
		if len(args) == 1 {
			return fmt.Sprintf("rune_regex_search(%s, %s)", receiver, args[0])
		}
	case "regex.split":
		if len(args) == 1 {
			return fmt.Sprintf("rune_regex_split(%s, %s)", receiver, args[0])
		}
	case "regex.source":
		return receiver + ".source"
	case "regex.flags":
		return receiver + ".flags"
	case "regex.global":
		return fmt.Sprintf("rune_regex_has_flag(%s, \"g\")", receiver)
	case "regex.ignoreCase":
		return fmt.Sprintf("rune_regex_has_flag(%s, \"i\")", receiver)
	case "regex.multiline":
		return fmt.Sprintf("rune_regex_has_flag(%s, \"m\")", receiver)
	case "regex.dotAll":
		return fmt.Sprintf("rune_regex_has_flag(%s, \"s\")", receiver)
	case "regex.unicode":
		return fmt.Sprintf("rune_regex_has_flag(%s, \"u\")", receiver)
	case "regex.unicodeSets":
		return fmt.Sprintf("rune_regex_has_flag(%s, \"v\")", receiver)
	case "regex.sticky":
		return fmt.Sprintf("rune_regex_has_flag(%s, \"y\")", receiver)
	case "regex.hasIndices":
		return fmt.Sprintf("rune_regex_has_flag(%s, \"d\")", receiver)
	case "regex.lastIndex":
		return receiver + ".last_index"
	case "regex.setLastIndex":
		if len(args) == 1 {
			return fmt.Sprintf("{ %s.last_index = %s; %s }", receiver, args[0], args[0])
		}
	}
	_ = resultType
	return runtimeTrap(fn.Intrinsic)
}

func (g *generator) bytesModuleCall(fn *stdlib.Function, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "bytes.new":
		if len(args) == 1 {
			return fmt.Sprintf("Array::make(%s, 0)", args[0])
		}
	case "bytes.fromInts":
		if len(args) == 1 {
			return args[0] + ".copy()"
		}
	}
	return zeroValue(resultType)
}

func (g *generator) bytesIntrinsicCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "bytes.length", "bytes.byteLength":
		return receiver + ".length()"
	case "bytes.clone":
		return receiver + ".copy()"
	case "bytes.slice":
		if len(args) == 2 {
			return fmt.Sprintf("%s[%s:%s].to_owned()", receiver, args[0], args[1])
		}
	case "bytes.toInts":
		return receiver + ".copy()"
	case "bytes.getInt4":
		if len(args) == 1 {
			return fmt.Sprintf("{ let __nibble = if %s %% 2 == 0 { %s[%s / 2] / 16 } else { %s[%s / 2] %% 16 }; if __nibble >= 8 { __nibble - 16 } else { __nibble } }", args[0], receiver, args[0], receiver, args[0])
		}
	case "bytes.getInt8":
		if len(args) == 1 {
			return fmt.Sprintf("{ let __value = %s[%s]; if __value >= 128 { __value - 256 } else { __value } }", receiver, args[0])
		}
	case "bytes.getUInt8":
		if len(args) == 1 {
			return fmt.Sprintf("%s[%s]", receiver, args[0])
		}
	case "bytes.getInt16":
		if len(args) == 2 {
			return g.bytesReadNumberExpr(receiver, args[0], 2, args[1], true)
		}
	case "bytes.getUInt16":
		if len(args) == 2 {
			return g.bytesReadNumberExpr(receiver, args[0], 2, args[1], false)
		}
	case "bytes.getInt":
		if len(args) == 2 {
			return g.bytesReadNumberExpr(receiver, args[0], 4, args[1], true)
		}
	case "bytes.getUInt":
		if len(args) == 2 {
			return g.bytesReadNumberExpr(receiver, args[0], 4, args[1], false)
		}
	case "bytes.getInt64":
		if len(args) == 2 {
			return g.bytesReadNumberExpr(receiver, args[0], 8, args[1], true)
		}
	case "bytes.getUInt64":
		if len(args) == 2 {
			return g.bytesReadNumberExpr(receiver, args[0], 8, args[1], false)
		}
	case "bytes.getFloat":
		if len(args) == 2 {
			return g.bytesReadFloatExpr(receiver, args[0], 4, args[1])
		}
	case "bytes.getDouble":
		if len(args) == 2 {
			return g.bytesReadFloatExpr(receiver, args[0], 8, args[1])
		}
	case "bytes.setInt4":
		if len(args) == 2 {
			return fmt.Sprintf("{ let __byte_index = %s / 2; let __old = %s[__byte_index]; let __nibble = (%s) & 15; %s[__byte_index] = if %s %% 2 == 0 { (__old & 15) | (__nibble << 4) } else { (__old & 240) | __nibble }; %s }", args[0], receiver, args[1], receiver, args[0], args[1])
		}
	case "bytes.setInt8", "bytes.setUInt8":
		if len(args) == 2 {
			return fmt.Sprintf("{ %s[%s] = (%s) & 255; %s }", receiver, args[0], args[1], args[1])
		}
	case "bytes.setInt16", "bytes.setUInt16":
		if len(args) == 3 {
			return g.bytesWriteNumberExpr(receiver, args[0], args[1], 2, args[2])
		}
	case "bytes.setInt", "bytes.setUInt":
		if len(args) == 3 {
			return g.bytesWriteNumberExpr(receiver, args[0], args[1], 4, args[2])
		}
	case "bytes.setInt64", "bytes.setUInt64":
		if len(args) == 3 {
			return g.bytesWriteNumberExpr(receiver, args[0], args[1], 8, args[2])
		}
	case "bytes.setFloat":
		if len(args) == 3 {
			return g.bytesWriteFloatExpr(receiver, args[0], args[1], 4, args[2])
		}
	case "bytes.setDouble":
		if len(args) == 3 {
			return g.bytesWriteFloatExpr(receiver, args[0], args[1], 8, args[2])
		}
	}
	_ = resultType
	return runtimeTrap(fn.Intrinsic)
}

func (g *generator) bytesWriteNumberExpr(receiver string, index string, value string, size int, littleEndian string) string {
	return g.bytesWriteNumberResultExpr(receiver, index, value, size, littleEndian, value)
}

func (g *generator) bytesWriteNumberResultExpr(receiver string, index string, value string, size int, littleEndian string, result string) string {
	target := g.nextTemp("__bytes")
	offset := g.nextTemp("__offset")
	little := make([]string, 0, size)
	big := make([]string, 0, size)
	for i := 0; i < size; i++ {
		little = append(little, fmt.Sprintf("%s[%s + %d] = %s", target, offset, i, shiftedByteExpr(value, i*8)))
		big = append(big, fmt.Sprintf("%s[%s + %d] = %s", target, offset, i, shiftedByteExpr(value, (size-1-i)*8)))
	}
	return fmt.Sprintf("{ let %s = %s; let %s = %s; if %s { %s } else { %s }; %s }", target, receiver, offset, index, littleEndian, strings.Join(little, "; "), strings.Join(big, "; "), result)
}

func (g *generator) bytesReadNumberExpr(receiver string, index string, size int, littleEndian string, signed bool) string {
	target := g.nextTemp("__bytes")
	offset := g.nextTemp("__offset")
	value := g.nextTemp("__value")
	little := numberDecodeMaybeSignedAtExpr(target, offset, size, true, signed)
	big := numberDecodeMaybeSignedAtExpr(target, offset, size, false, signed)
	return fmt.Sprintf("{ let %s = %s; let %s = %s; let %s = if %s { %s } else { %s }; %s }", target, receiver, offset, index, value, littleEndian, little, big, value)
}

func (g *generator) bytesWriteFloatExpr(receiver string, index string, value string, size int, littleEndian string) string {
	if size == 4 {
		return g.bytesWriteNumberResultExpr(receiver, index, fmt.Sprintf("(%s).reinterpret_as_uint().reinterpret_as_int()", value), 4, littleEndian, value)
	}
	return g.bytesWriteUInt64BitsExpr(receiver, index, fmt.Sprintf("(%s).reinterpret_as_uint64()", value), littleEndian, value)
}

func (g *generator) bytesReadFloatExpr(receiver string, index string, size int, littleEndian string) string {
	target := g.nextTemp("__bytes")
	offset := g.nextTemp("__offset")
	bits := g.nextTemp("__bits")
	if size == 4 {
		little := numberDecodeAtExpr(target, offset, 4, true)
		big := numberDecodeAtExpr(target, offset, 4, false)
		return fmt.Sprintf("{ let %s = %s; let %s = %s; let %s = if %s { %s } else { %s }; Float::reinterpret_from_uint(%s.reinterpret_as_uint()) }", target, receiver, offset, index, bits, littleEndian, little, big, bits)
	}
	little := uint64BitsDecodeAtExpr(target, offset, true)
	big := uint64BitsDecodeAtExpr(target, offset, false)
	return fmt.Sprintf("{ let %s = %s; let %s = %s; let %s = if %s { %s } else { %s }; %s.reinterpret_as_double() }", target, receiver, offset, index, bits, littleEndian, little, big, bits)
}

func (g *generator) bytesWriteUInt64BitsExpr(receiver string, index string, bits string, littleEndian string, result string) string {
	target := g.nextTemp("__bytes")
	offset := g.nextTemp("__offset")
	bitsName := g.nextTemp("__bits")
	little := make([]string, 0, 8)
	big := make([]string, 0, 8)
	for i := 0; i < 8; i++ {
		little = append(little, fmt.Sprintf("%s[%s + %d] = %s", target, offset, i, shiftedUInt64ByteExpr(bitsName, i*8)))
		big = append(big, fmt.Sprintf("%s[%s + %d] = %s", target, offset, i, shiftedUInt64ByteExpr(bitsName, (7-i)*8)))
	}
	return fmt.Sprintf("{ let %s = %s; let %s = %s; let %s = %s; if %s { %s } else { %s }; %s }", target, receiver, offset, index, bitsName, bits, littleEndian, strings.Join(little, "; "), strings.Join(big, "; "), result)
}

func (g *generator) streamModuleCall(fn *stdlib.Function, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "buffer.new", "writer.new", "writer.withCapacity":
		return "([] : Array[Int])"
	case "buffer.fromBytes":
		if len(args) == 1 {
			return args[0] + ".copy()"
		}
	case "reader.new":
		if len(args) == 1 {
			g.useReader = true
			return "RuneReader::{ data: " + args[0] + ".copy(), position: 0, nibble: 0 }"
		}
	}
	return zeroValue(resultType)
}

func (g *generator) bufferIntrinsicCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "buffer.length":
		return receiver + ".length()"
	case "buffer.isEmpty":
		return receiver + ".length() == 0"
	case "buffer.clear":
		return fmt.Sprintf("{ %s.clear(); () }", receiver)
	case "buffer.clone":
		return receiver + ".copy()"
	case "buffer.toBytes", "buffer.toInts":
		return receiver + ".copy()"
	case "buffer.append", "buffer.appendInt":
		if len(args) == 1 {
			target := g.nextTemp("__buffer")
			return fmt.Sprintf("{ let %s = %s; %s.push((%s) & 255); %s }", target, receiver, target, args[0], target)
		}
	case "buffer.appendBytes":
		if len(args) == 1 {
			target := g.nextTemp("__buffer")
			return fmt.Sprintf("{ let %s = %s; %s.iter().each(fn(__value) { %s.push(__value) }); %s }", target, receiver, args[0], target, target)
		}
	case "buffer.reader", "buffer.writer":
		if fn.Intrinsic == "buffer.reader" {
			g.useReader = true
			return "RuneReader::{ data: " + receiver + ".copy(), position: 0, nibble: 0 }"
		}
		return receiver
	}
	_ = resultType
	return runtimeTrap(fn.Intrinsic)
}

func (g *generator) readerIntrinsicCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	g.useReader = true
	switch fn.Intrinsic {
	case "reader.length":
		return receiver + ".data.length()"
	case "reader.position":
		return receiver + ".position"
	case "reader.remaining":
		return fmt.Sprintf("%s.data.length() - %s.position", receiver, receiver)
	case "reader.isEmpty":
		return fmt.Sprintf("%s.position >= %s.data.length()", receiver, receiver)
	case "reader.seek":
		if len(args) == 1 {
			return fmt.Sprintf("{ let __reader = %s; __reader.position = %s; __reader.nibble = 0; __reader.position }", receiver, args[0])
		}
	case "reader.skip":
		if len(args) == 1 {
			return fmt.Sprintf("{ let __reader = %s; __reader.position = __reader.position + %s; __reader.nibble = 0; __reader.position }", receiver, args[0])
		}
	case "reader.readBytes":
		if len(args) == 1 {
			return fmt.Sprintf("{ let __reader = %s; if __reader.nibble == 1 { __reader.position = __reader.position + 1; __reader.nibble = 0 }; let __start = __reader.position; __reader.position = __reader.position + %s; __reader.data[__start:__reader.position].to_owned() }", receiver, args[0])
		}
	case "reader.readInt4":
		return g.readerReadInt4Expr(receiver)
	case "reader.readInt8":
		return g.readerReadInt8Expr(receiver)
	case "reader.readUInt8":
		return g.readerReadUInt8Expr(receiver)
	case "reader.readInt16":
		if len(args) == 1 {
			return g.readerReadNumberExpr(receiver, 2, args[0], true)
		}
	case "reader.readUInt16":
		if len(args) == 1 {
			return g.readerReadNumberExpr(receiver, 2, args[0], false)
		}
	case "reader.readInt":
		if len(args) == 1 {
			return g.readerReadNumberExpr(receiver, 4, args[0], true)
		}
	case "reader.readUInt":
		if len(args) == 1 {
			return g.readerReadNumberExpr(receiver, 4, args[0], false)
		}
	case "reader.readInt64":
		if len(args) == 1 {
			return g.readerReadNumberExpr(receiver, 8, args[0], true)
		}
	case "reader.readUInt64":
		if len(args) == 1 {
			return g.readerReadNumberExpr(receiver, 8, args[0], false)
		}
	case "reader.readFloat":
		if len(args) == 1 {
			return g.readerReadFloatExpr(receiver, 4, args[0])
		}
	case "reader.readDouble":
		if len(args) == 1 {
			return g.readerReadFloatExpr(receiver, 8, args[0])
		}
	}
	_ = resultType
	return runtimeTrap(fn.Intrinsic)
}

func (g *generator) writerIntrinsicCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "writer.length", "writer.position":
		return receiver + ".length()"
	case "writer.clear":
		return fmt.Sprintf("{ %s.clear(); () }", receiver)
	case "writer.toBytes", "writer.toInts":
		return receiver + ".copy()"
	case "writer.writeBytes":
		if len(args) == 1 {
			target := g.nextTemp("__writer")
			return fmt.Sprintf("{ let %s = %s; %s.iter().each(fn(__value) { %s.push(__value) }); %s }", target, receiver, args[0], target, target)
		}
	case "writer.writeInt4":
		if len(args) == 1 {
			target := g.nextTemp("__writer")
			return fmt.Sprintf("{ let %s = %s; let __nibble = (%s) & 15; if %s.length() > 0 && (%s[%s.length() - 1] & 15) == 0 { %s[%s.length() - 1] = (%s[%s.length() - 1] & 240) | __nibble } else { %s.push(__nibble << 4) }; %s }", target, receiver, args[0], target, target, target, target, target, target, target, target, target)
		}
	case "writer.writeInt8", "writer.writeUInt8":
		if len(args) == 1 {
			target := g.nextTemp("__writer")
			return fmt.Sprintf("{ let %s = %s; %s.push((%s) & 255); %s }", target, receiver, target, args[0], target)
		}
	case "writer.writeInt16", "writer.writeUInt16":
		if len(args) == 2 {
			return g.writerWriteNumberExpr(receiver, args[0], 2, args[1])
		}
	case "writer.writeInt", "writer.writeUInt":
		if len(args) == 2 {
			return g.writerWriteNumberExpr(receiver, args[0], 4, args[1])
		}
	case "writer.writeInt64", "writer.writeUInt64":
		if len(args) == 2 {
			return g.writerWriteNumberExpr(receiver, args[0], 8, args[1])
		}
	case "writer.writeFloat":
		if len(args) == 2 {
			return g.writerWriteFloatExpr(receiver, args[0], 4, args[1])
		}
	case "writer.writeDouble":
		if len(args) == 2 {
			return g.writerWriteFloatExpr(receiver, args[0], 8, args[1])
		}
	}
	_ = resultType
	return runtimeTrap(fn.Intrinsic)
}

func (g *generator) writerWriteNumberExpr(receiver string, value string, size int, littleEndian string) string {
	target := g.nextTemp("__writer")
	little := make([]string, 0, size)
	big := make([]string, 0, size)
	for i := 0; i < size; i++ {
		little = append(little, fmt.Sprintf("%s.push(%s)", target, shiftedByteExpr(value, i*8)))
		big = append(big, fmt.Sprintf("%s.push(%s)", target, shiftedByteExpr(value, (size-1-i)*8)))
	}
	return fmt.Sprintf("{ let %s = %s; if %s { %s } else { %s }; %s }", target, receiver, littleEndian, strings.Join(little, "; "), strings.Join(big, "; "), target)
}

func shiftedByteExpr(value string, shift int) string {
	if shift >= 32 {
		return "0"
	}
	return fmt.Sprintf("((%s) >> %d) & 255", value, shift)
}

func shiftedUInt64ByteExpr(value string, shift int) string {
	return fmt.Sprintf("(((%s) >> %d).to_uint().reinterpret_as_int()) & 255", value, shift)
}

func (g *generator) writerWriteFloatExpr(receiver string, value string, size int, littleEndian string) string {
	if size == 4 {
		return g.writerWriteNumberExpr(receiver, fmt.Sprintf("(%s).reinterpret_as_uint().reinterpret_as_int()", value), 4, littleEndian)
	}
	return g.writerWriteUInt64BitsExpr(receiver, fmt.Sprintf("(%s).reinterpret_as_uint64()", value), littleEndian)
}

func (g *generator) writerWriteUInt64BitsExpr(receiver string, bits string, littleEndian string) string {
	target := g.nextTemp("__writer")
	bitsName := g.nextTemp("__bits")
	little := make([]string, 0, 8)
	big := make([]string, 0, 8)
	for i := 0; i < 8; i++ {
		little = append(little, fmt.Sprintf("%s.push(%s)", target, shiftedUInt64ByteExpr(bitsName, i*8)))
		big = append(big, fmt.Sprintf("%s.push(%s)", target, shiftedUInt64ByteExpr(bitsName, (7-i)*8)))
	}
	return fmt.Sprintf("{ let %s = %s; let %s = %s; if %s { %s } else { %s }; %s }", target, receiver, bitsName, bits, littleEndian, strings.Join(little, "; "), strings.Join(big, "; "), target)
}

func (g *generator) readerReadNumberExpr(receiver string, size int, littleEndian string, signed bool) string {
	target := g.nextTemp("__reader")
	value := g.nextTemp("__value")
	offset := g.nextTemp("__offset")
	little := numberDecodeMaybeSignedAtExpr(target+".data", offset, size, true, signed)
	big := numberDecodeMaybeSignedAtExpr(target+".data", offset, size, false, signed)
	return fmt.Sprintf("{ let %s = %s; if %s.nibble == 1 { %s.position = %s.position + 1; %s.nibble = 0 }; let %s = %s.position; let %s = if %s { %s } else { %s }; %s.position = %s.position + %d; %s }", target, receiver, target, target, target, target, offset, target, value, littleEndian, little, big, target, target, size, value)
}

func (g *generator) readerReadFloatExpr(receiver string, size int, littleEndian string) string {
	target := g.nextTemp("__reader")
	offset := g.nextTemp("__offset")
	bits := g.nextTemp("__bits")
	if size == 4 {
		little := numberDecodeAtExpr(target+".data", offset, 4, true)
		big := numberDecodeAtExpr(target+".data", offset, 4, false)
		return fmt.Sprintf("{ let %s = %s; if %s.nibble == 1 { %s.position = %s.position + 1; %s.nibble = 0 }; let %s = %s.position; let %s = if %s { %s } else { %s }; %s.position = %s.position + 4; Float::reinterpret_from_uint(%s.reinterpret_as_uint()) }", target, receiver, target, target, target, target, offset, target, bits, littleEndian, little, big, target, target, bits)
	}
	little := uint64BitsDecodeAtExpr(target+".data", offset, true)
	big := uint64BitsDecodeAtExpr(target+".data", offset, false)
	return fmt.Sprintf("{ let %s = %s; if %s.nibble == 1 { %s.position = %s.position + 1; %s.nibble = 0 }; let %s = %s.position; let %s = if %s { %s } else { %s }; %s.position = %s.position + 8; %s.reinterpret_as_double() }", target, receiver, target, target, target, target, offset, target, bits, littleEndian, little, big, target, target, bits)
}

func (g *generator) readerReadUInt8Expr(receiver string) string {
	target := g.nextTemp("__reader")
	return fmt.Sprintf("{ let %s = %s; if %s.nibble == 1 { %s.position = %s.position + 1; %s.nibble = 0 }; let __value = %s.data[%s.position]; %s.position = %s.position + 1; __value }", target, receiver, target, target, target, target, target, target, target, target)
}

func (g *generator) readerReadInt8Expr(receiver string) string {
	value := g.readerReadUInt8Expr(receiver)
	return fmt.Sprintf("{ let __value = %s; if __value >= 128 { __value - 256 } else { __value } }", value)
}

func (g *generator) readerReadInt4Expr(receiver string) string {
	target := g.nextTemp("__reader")
	return fmt.Sprintf("{ let %s = %s; let __byte = %s.data[%s.position]; let __nibble = if %s.nibble == 0 { %s.nibble = 1; __byte / 16 } else { %s.nibble = 0; %s.position = %s.position + 1; __byte & 15 }; if __nibble >= 8 { __nibble - 16 } else { __nibble } }", target, receiver, target, target, target, target, target, target, target)
}

func numberDecodeExpr(target string, size int, little bool) string {
	return numberDecodeAtExpr(target, "0", size, little)
}

func numberDecodeMaybeSignedExpr(target string, size int, little bool, signed bool) string {
	return numberDecodeMaybeSignedAtExpr(target, "0", size, little, signed)
}

func numberDecodeMaybeSignedAtExpr(target string, offset string, size int, little bool, signed bool) string {
	unsigned := numberDecodeAtExpr(target, offset, size, little)
	if !signed {
		return unsigned
	}
	signIndex := 0
	if little {
		signIndex = size - 1
	}
	signByte := fmt.Sprintf("%s[%s + %d]", target, offset, signIndex)
	magnitude := numberDecodeComplementAtExpr(target, offset, size, little)
	return fmt.Sprintf("if %s >= 128 { 0 - (%s + 1) } else { %s }", signByte, magnitude, unsigned)
}

func numberDecodeAtExpr(target string, offset string, size int, little bool) string {
	parts := make([]string, 0, size)
	for i := 0; i < size; i++ {
		index := i
		shift := i * 8
		if !little {
			shift = (size - 1 - i) * 8
		}
		if shift >= 32 {
			continue
		}
		parts = append(parts, fmt.Sprintf("(%s[%s + %d] << %d)", target, offset, index, shift))
	}
	if len(parts) == 0 {
		return "0"
	}
	return strings.Join(parts, " + ")
}

func numberDecodeComplementAtExpr(target string, offset string, size int, little bool) string {
	parts := make([]string, 0, size)
	for i := 0; i < size; i++ {
		index := i
		shift := i * 8
		if !little {
			shift = (size - 1 - i) * 8
		}
		if shift >= 32 {
			continue
		}
		parts = append(parts, fmt.Sprintf("((255 - %s[%s + %d]) << %d)", target, offset, index, shift))
	}
	if len(parts) == 0 {
		return "0"
	}
	return strings.Join(parts, " + ")
}

func uint64BitsDecodeExpr(target string, little bool) string {
	return uint64BitsDecodeAtExpr(target, "0", little)
}

func uint64BitsDecodeAtExpr(target string, offset string, little bool) string {
	low := numberDecodeAtExpr(target, offset, 4, little)
	highOffset := fmt.Sprintf("(%s + 4)", offset)
	if !little {
		low = numberDecodeAtExpr(target, highOffset, 4, false)
		highOffset = offset
	}
	high := numberDecodeAtExpr(target, highOffset, 4, little)
	return fmt.Sprintf("((%s).reinterpret_as_uint().to_uint64() | ((%s).reinterpret_as_uint().to_uint64() << 32))", low, high)
}

func (g *generator) mapIntrinsicCall(fn *stdlib.Function, receiver string, args []string, rawArgs []ir.Expr, resultType checker.Type) string {
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
			arity := callbackArity(rawArgs[0], 3)
			return fmt.Sprintf("%s.each((k, v) => { ignore(%s) })", receiver, callbackCall(args[0], arity, "v", "k", receiver))
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

func (g *generator) setIntrinsicCall(fn *stdlib.Function, receiver string, args []string, rawArgs []ir.Expr, resultType checker.Type) string {
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
			arity := callbackArity(rawArgs[0], 3)
			return fmt.Sprintf("%s.each((v) => { ignore(%s) })", receiver, callbackCall(args[0], arity, "v", "v", receiver))
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
	case "stringbuffer.isEmpty":
		return receiver + ".to_string().length() == 0"
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
		if len(rawArgs) == 1 {
			if sel, ok := rawArgs[0].(*ir.SelectorExpr); ok && sel.Name == "next" {
				if _, ok := checker.IterValue(sel.Receiver.ResultType()); ok {
					return g.expr(sel.Receiver)
				}
			}
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
		return fmt.Sprintf("%s.eachi(fn(index, value) { ignore(%s) })", receiver, callbackCall(callback, 2, "value", "index"))
	default:
		iter := g.nextTemp("__iter")
		return fmt.Sprintf("{ let %s = %s; %s.eachi(fn(index, value) { ignore(%s) }) }", iter, receiver, iter, callbackCall(callback, arity, "value", "index", iter))
	}
}

func (g *generator) iterMapExpr(receiver string, callback string, arity int) string {
	switch arity {
	case 0:
		return fmt.Sprintf("%s.map(fn(_) { %s() }).to_array()", receiver, callback)
	case 1:
		return fmt.Sprintf("%s.map(%s).to_array()", receiver, callback)
	case 2:
		return fmt.Sprintf("%s.mapi(fn(index, value) { %s }).to_array()", receiver, callbackCall(callback, 2, "value", "index"))
	default:
		iter := g.nextTemp("__iter")
		return fmt.Sprintf("{ let %s = %s; %s.mapi(fn(index, value) { %s }).to_array() }", iter, receiver, iter, callbackCall(callback, arity, "value", "index", iter))
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
