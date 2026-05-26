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
	case "int.toString", "int4.fromInt", "int8.fromInt", "int16.fromInt", "int64.fromInt",
		"uint.fromInt", "uint8.fromInt", "uint16.fromInt", "uint64.fromInt",
		"float.fromDouble", "int4.toInt", "int8.toInt", "int16.toInt", "int64.toInt",
		"uint.toInt", "uint8.toInt", "uint16.toInt", "uint64.toInt", "float.toDouble":
		return g.numericIntrinsicCall(fn, args, call.ResultType()), true
	case "binary.new", "binary.fromInts":
		return g.binaryModuleCall(fn, args, call.ResultType()), true
	case "buffer.new", "buffer.fromBinary", "reader.new", "writer.new", "writer.withCapacity":
		return g.streamModuleCall(fn, args, call.ResultType()), true
	case "fs.readFile":
		if len(args) != 1 {
			return g.zeroValue(call.ResultType()), true
		}
		return fmt.Sprintf("runeReadFile(%s)", args[0]), true
	case "fs.readFileText", "fs.writeFile", "fs.writeFileText", "fs.exists", "fs.readdir", "fs.mkdir", "fs.remove", "fs.stat":
		return g.fsModuleCall(fn, args, call.ResultType()), true
	case "path.basename", "path.dirname", "path.extname", "path.join", "path.normalize", "path.resolve", "path.relative", "path.isAbsolute":
		return g.pathModuleCall(fn, args, call.ResultType()), true
	case "process.argv", "process.cwd", "process.env", "process.exit", "process.platform":
		return g.processModuleCall(fn, args, call.ResultType()), true
	case "stringbuffer.new", "stringbuffer.from":
		return g.stringBufferModuleCall(fn, args, call.ResultType()), true
	case "iter.range", "iter.rangeStep", "iter.repeat", "iter.empty":
		return g.iterModuleCall(fn, args, call.ResultType()), true
	case "compress.gzip", "compress.gunzip", "compress.deflate", "compress.inflate",
		"compress.brotli", "compress.unbrotli", "compress.zstd", "compress.unzstd",
		"compress.gzipText", "compress.gunzipText", "compress.brotliText", "compress.unbrotliText",
		"compress.zstdText", "compress.unzstdText":
		return g.compressModuleCall(fn, args, call.ResultType()), true
	case "net.connect", "net.listen":
		return g.netModuleCall(fn, args, call.ResultType()), true
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
	case strings.HasPrefix(fn.Intrinsic, "int."):
		if fn.Intrinsic == "int.toString" {
			return receiver + ".toString()", true
		}
		return g.unsupportedIntrinsic(fn, call.ResultType()), true
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
	case strings.HasPrefix(fn.Intrinsic, "binary."):
		return g.binaryReceiverCall(fn, receiver, args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "buffer."):
		return g.bufferReceiverCall(fn, receiver, args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "reader."):
		return g.readerReceiverCall(fn, receiver, args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "writer."):
		return g.writerReceiverCall(fn, receiver, args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "stringbuffer."):
		return g.stringBufferReceiverCall(fn, receiver, args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "netConnection."):
		return g.netConnectionReceiverCall(fn, receiver, args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "netListener."):
		return g.netListenerReceiverCall(fn, receiver, args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "map."), strings.HasPrefix(fn.Intrinsic, "weakMap."):
		return g.mapIntrinsicCall(fn, receiver, args, call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "set."), strings.HasPrefix(fn.Intrinsic, "weakSet."):
		return g.setIntrinsicCall(fn, receiver, args, call.ResultType()), true
	default:
		return g.unsupportedIntrinsic(fn, call.ResultType()), true
	}
}

func (g *generator) fsModuleCall(fn *stdlib.Function, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "fs.readFileText":
		return fmt.Sprintf("runeFsReadFileText(%s)", args[0])
	case "fs.writeFile":
		return fmt.Sprintf("runeFsWriteFile(%s, %s)", args[0], args[1])
	case "fs.writeFileText":
		return fmt.Sprintf("runeFsWriteFileText(%s, %s)", args[0], args[1])
	case "fs.exists":
		return fmt.Sprintf("runeFsExists(%s)", args[0])
	case "fs.readdir":
		return fmt.Sprintf("runeFsReaddir(%s)", args[0])
	case "fs.mkdir":
		return fmt.Sprintf("runeFsMkdir(%s)", args[0])
	case "fs.remove":
		return fmt.Sprintf("runeFsRemove(%s)", args[0])
	case "fs.stat":
		return fmt.Sprintf("runeFsStat(%s)", args[0])
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}

func (g *generator) pathModuleCall(fn *stdlib.Function, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "path.basename":
		return fmt.Sprintf("runePathBasename(%s)", args[0])
	case "path.dirname":
		return fmt.Sprintf("runePathDirname(%s)", args[0])
	case "path.extname":
		return fmt.Sprintf("runePathExtname(%s)", args[0])
	case "path.join":
		return fmt.Sprintf("runePathJoin(%s)", args[0])
	case "path.normalize":
		return fmt.Sprintf("runePathNormalize(%s)", args[0])
	case "path.resolve":
		return fmt.Sprintf("runePathResolve(%s)", args[0])
	case "path.relative":
		return fmt.Sprintf("runePathRelative(%s, %s)", args[0], args[1])
	case "path.isAbsolute":
		return fmt.Sprintf("runePathIsAbsolute(%s)", args[0])
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}

func (g *generator) processModuleCall(fn *stdlib.Function, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "process.argv":
		return "runeProcessArgv()"
	case "process.cwd":
		return "runeProcessCwd()"
	case "process.env":
		return fmt.Sprintf("runeProcessEnv(%s)", args[0])
	case "process.exit":
		return fmt.Sprintf("runeProcessExit(%s)", args[0])
	case "process.platform":
		return "runeProcessPlatform()"
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}

func (g *generator) stringBufferModuleCall(fn *stdlib.Function, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "stringbuffer.new":
		return "new RuneStringBuffer()"
	case "stringbuffer.from":
		return fmt.Sprintf("new RuneStringBuffer(%s)", args[0])
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}

func (g *generator) iterModuleCall(fn *stdlib.Function, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "iter.range":
		return fmt.Sprintf("runeIterRange(%s, %s)", args[0], args[1])
	case "iter.rangeStep":
		return fmt.Sprintf("runeIterRangeStep(%s, %s, %s)", args[0], args[1], args[2])
	case "iter.repeat":
		return fmt.Sprintf("Array.from({ length: %s }, () => %s)", args[1], args[0])
	case "iter.empty":
		return "[]"
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}

func (g *generator) compressModuleCall(fn *stdlib.Function, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "compress.gzip":
		return fmt.Sprintf("runeCompressGzip(%s)", args[0])
	case "compress.gunzip":
		return fmt.Sprintf("runeCompressGunzip(%s)", args[0])
	case "compress.deflate":
		return fmt.Sprintf("runeCompressDeflate(%s)", args[0])
	case "compress.inflate":
		return fmt.Sprintf("runeCompressInflate(%s)", args[0])
	case "compress.brotli":
		return fmt.Sprintf("runeCompressBrotli(%s)", args[0])
	case "compress.unbrotli":
		return fmt.Sprintf("runeCompressUnbrotli(%s)", args[0])
	case "compress.zstd":
		return fmt.Sprintf("runeCompressZstd(%s)", args[0])
	case "compress.unzstd":
		return fmt.Sprintf("runeCompressUnzstd(%s)", args[0])
	case "compress.gzipText":
		return fmt.Sprintf("runeCompressGzipText(%s)", args[0])
	case "compress.gunzipText":
		return fmt.Sprintf("runeCompressGunzipText(%s)", args[0])
	case "compress.brotliText":
		return fmt.Sprintf("runeCompressBrotliText(%s)", args[0])
	case "compress.unbrotliText":
		return fmt.Sprintf("runeCompressUnbrotliText(%s)", args[0])
	case "compress.zstdText":
		return fmt.Sprintf("runeCompressZstdText(%s)", args[0])
	case "compress.unzstdText":
		return fmt.Sprintf("runeCompressUnzstdText(%s)", args[0])
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}

func (g *generator) netModuleCall(fn *stdlib.Function, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "net.connect":
		return fmt.Sprintf("runeNetConnect(%s)", args[0])
	case "net.listen":
		return fmt.Sprintf("runeNetListen(%s)", args[0])
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}

func (g *generator) streamModuleCall(fn *stdlib.Function, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "buffer.new":
		if len(args) != 0 {
			return g.zeroValue(resultType)
		}
		return "new RuneBuffer()"
	case "buffer.fromBinary":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("RuneBuffer.fromBinary(%s)", args[0])
	case "reader.new":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("new RuneReader(%s)", args[0])
	case "writer.new":
		if len(args) != 0 {
			return g.zeroValue(resultType)
		}
		return "new RuneWriter()"
	case "writer.withCapacity":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("new RuneWriter(%s)", args[0])
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}

func (g *generator) numericIntrinsicCall(fn *stdlib.Function, args []string, resultType checker.Type) string {
	if len(args) != 1 {
		return g.zeroValue(resultType)
	}
	value := args[0]
	switch fn.Intrinsic {
	case "int.toString":
		return fmt.Sprintf("(%s).toString()", value)
	case "int4.fromInt":
		return fmt.Sprintf("((__value: number): number => { const __n = __value & 0xf; return __n >= 8 ? __n - 16 : __n; })(%s)", value)
	case "int8.fromInt":
		return fmt.Sprintf("((__value: number): number => (__value << 24) >> 24)(%s)", value)
	case "int16.fromInt":
		return fmt.Sprintf("((__value: number): number => (__value << 16) >> 16)(%s)", value)
	case "int64.fromInt":
		return fmt.Sprintf("BigInt(%s)", value)
	case "uint.fromInt":
		return fmt.Sprintf("(%s >>> 0)", value)
	case "uint8.fromInt":
		return fmt.Sprintf("(%s & 0xff)", value)
	case "uint16.fromInt":
		return fmt.Sprintf("(%s & 0xffff)", value)
	case "uint64.fromInt":
		return fmt.Sprintf("BigInt.asUintN(64, BigInt(%s))", value)
	case "float.fromDouble":
		return fmt.Sprintf("Math.fround(%s)", value)
	case "int64.toInt", "uint64.toInt":
		return fmt.Sprintf("Number(%s)", value)
	default:
		return value
	}
}

func (g *generator) binaryModuleCall(fn *stdlib.Function, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "binary.new":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("new DataView(new ArrayBuffer(%s))", args[0])
	case "binary.fromInts":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("((__bytes: number[]): DataView => { const __array = new Uint8Array(__bytes.map((__value) => __value & 0xff)); return new DataView(__array.buffer.slice(__array.byteOffset, __array.byteOffset + __array.byteLength)); })(%s)", args[0])
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}

func (g *generator) binaryReceiverCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "binary.length":
		return receiver + ".byteLength"
	case "binary.clone":
		return fmt.Sprintf("((__view: DataView): DataView => new DataView(__view.buffer.slice(__view.byteOffset, __view.byteOffset + __view.byteLength)))(%s)", receiver)
	case "binary.slice":
		if len(args) != 2 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("((__view: DataView): DataView => new DataView(__view.buffer.slice(__view.byteOffset + %s, __view.byteOffset + %s)))(%s)", args[0], args[1], receiver)
	case "binary.toInts":
		return fmt.Sprintf("Array.from(new Uint8Array(%s.buffer, %s.byteOffset, %s.byteLength))", receiver, receiver, receiver)
	case "binary.getInt4":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("((__index: number): number => { const __byte = %s.getUint8(Math.trunc(__index / 2)); const __nibble = __index %% 2 === 0 ? (__byte >> 4) : (__byte & 0xf); return __nibble >= 8 ? __nibble - 16 : __nibble; })(%s)", receiver, args[0])
	case "binary.setInt4":
		if len(args) != 2 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("((__index: number, __value: number): number => { const __byteIndex = Math.trunc(__index / 2); const __old = %s.getUint8(__byteIndex); const __nibble = __value & 0xf; %s.setUint8(__byteIndex, __index %% 2 === 0 ? ((__old & 0x0f) | (__nibble << 4)) : ((__old & 0xf0) | __nibble)); return __value; })(%s, %s)", receiver, receiver, args[0], args[1])
	}
	if expr, ok := g.binaryDataViewCall(fn.Intrinsic, receiver, args, resultType); ok {
		return expr
	}
	return g.unsupportedIntrinsic(fn, resultType)
}

func (g *generator) binaryDataViewCall(intrinsic string, receiver string, args []string, resultType checker.Type) (string, bool) {
	methods := map[string]string{
		"binary.getInt8":   "getInt8",
		"binary.getUInt8":  "getUint8",
		"binary.getInt16":  "getInt16",
		"binary.getUInt16": "getUint16",
		"binary.getInt":    "getInt32",
		"binary.getUInt":   "getUint32",
		"binary.getInt64":  "getBigInt64",
		"binary.getUInt64": "getBigUint64",
		"binary.getFloat":  "getFloat32",
		"binary.getDouble": "getFloat64",
		"binary.setInt8":   "setInt8",
		"binary.setUInt8":  "setUint8",
		"binary.setInt16":  "setInt16",
		"binary.setUInt16": "setUint16",
		"binary.setInt":    "setInt32",
		"binary.setUInt":   "setUint32",
		"binary.setInt64":  "setBigInt64",
		"binary.setUInt64": "setBigUint64",
		"binary.setFloat":  "setFloat32",
		"binary.setDouble": "setFloat64",
	}
	method, ok := methods[intrinsic]
	if !ok {
		return "", false
	}
	if strings.HasPrefix(method, "get") {
		if method == "getInt8" || method == "getUint8" {
			if len(args) != 1 {
				return g.zeroValue(resultType), true
			}
			return fmt.Sprintf("%s.%s(%s)", receiver, method, args[0]), true
		}
		if len(args) != 2 {
			return g.zeroValue(resultType), true
		}
		return fmt.Sprintf("%s.%s(%s, %s)", receiver, method, args[0], args[1]), true
	}
	if method == "setInt8" || method == "setUint8" {
		if len(args) != 2 {
			return g.zeroValue(resultType), true
		}
		return fmt.Sprintf("((__value: %s): %s => { %s.%s(%s, __value); return __value; })(%s)", tsType(resultType), tsType(resultType), receiver, method, args[0], args[1]), true
	}
	if len(args) != 3 {
		return g.zeroValue(resultType), true
	}
	return fmt.Sprintf("((__value: %s): %s => { %s.%s(%s, __value, %s); return __value; })(%s)", tsType(resultType), tsType(resultType), receiver, method, args[0], args[2], args[1]), true
}

func (g *generator) bufferReceiverCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "buffer.length":
		return receiver + ".byteLength"
	case "buffer.clear":
		return receiver + ".clear()"
	case "buffer.clone":
		return receiver + ".clone()"
	case "buffer.toBinary":
		return receiver + ".toBinary()"
	case "buffer.toInts":
		return receiver + ".toInts()"
	case "buffer.append":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("%s.append(%s)", receiver, args[0])
	case "buffer.appendInt":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("%s.appendInt(%s)", receiver, args[0])
	case "buffer.appendBinary":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("%s.appendBinary(%s)", receiver, args[0])
	case "buffer.reader":
		return receiver + ".reader()"
	case "buffer.writer":
		return receiver + ".writer()"
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}

func (g *generator) readerReceiverCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "reader.length":
		return receiver + ".byteLength"
	case "reader.position":
		return receiver + ".position()"
	case "reader.remaining":
		return receiver + ".remaining()"
	case "reader.seek":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("%s.seek(%s)", receiver, args[0])
	case "reader.skip":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("%s.skip(%s)", receiver, args[0])
	case "reader.readBinary":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("%s.readBinary(%s)", receiver, args[0])
	case "reader.readInt4":
		return receiver + ".readInt4()"
	}
	if expr, ok := g.readerNumericCall(fn.Intrinsic, receiver, args, resultType); ok {
		return expr
	}
	return g.unsupportedIntrinsic(fn, resultType)
}

func (g *generator) readerNumericCall(intrinsic string, receiver string, args []string, resultType checker.Type) (string, bool) {
	methods := map[string]string{
		"reader.readInt8":   "readInt8",
		"reader.readUInt8":  "readUInt8",
		"reader.readInt16":  "readInt16",
		"reader.readUInt16": "readUInt16",
		"reader.readInt":    "readInt",
		"reader.readUInt":   "readUInt",
		"reader.readInt64":  "readInt64",
		"reader.readUInt64": "readUInt64",
		"reader.readFloat":  "readFloat",
		"reader.readDouble": "readDouble",
	}
	method, ok := methods[intrinsic]
	if !ok {
		return "", false
	}
	if method == "readInt8" || method == "readUInt8" {
		if len(args) != 0 {
			return g.zeroValue(resultType), true
		}
		return fmt.Sprintf("%s.%s()", receiver, method), true
	}
	if len(args) != 1 {
		return g.zeroValue(resultType), true
	}
	return fmt.Sprintf("%s.%s(%s)", receiver, method, args[0]), true
}

func (g *generator) writerReceiverCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "writer.length", "writer.position":
		return receiver + ".position()"
	case "writer.clear":
		return receiver + ".clear()"
	case "writer.toBinary":
		return receiver + ".toBinary()"
	case "writer.toInts":
		return receiver + ".toInts()"
	case "writer.writeBinary":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("%s.writeBinary(%s)", receiver, args[0])
	case "writer.writeInt4":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("%s.writeInt4(%s)", receiver, args[0])
	}
	if expr, ok := g.writerNumericCall(fn.Intrinsic, receiver, args, resultType); ok {
		return expr
	}
	return g.unsupportedIntrinsic(fn, resultType)
}

func (g *generator) writerNumericCall(intrinsic string, receiver string, args []string, resultType checker.Type) (string, bool) {
	methods := map[string]string{
		"writer.writeInt8":   "writeInt8",
		"writer.writeUInt8":  "writeUInt8",
		"writer.writeInt16":  "writeInt16",
		"writer.writeUInt16": "writeUInt16",
		"writer.writeInt":    "writeInt",
		"writer.writeUInt":   "writeUInt",
		"writer.writeInt64":  "writeInt64",
		"writer.writeUInt64": "writeUInt64",
		"writer.writeFloat":  "writeFloat",
		"writer.writeDouble": "writeDouble",
	}
	method, ok := methods[intrinsic]
	if !ok {
		return "", false
	}
	if method == "writeInt8" || method == "writeUInt8" {
		if len(args) != 1 {
			return g.zeroValue(resultType), true
		}
		return fmt.Sprintf("%s.%s(%s)", receiver, method, args[0]), true
	}
	if len(args) != 2 {
		return g.zeroValue(resultType), true
	}
	return fmt.Sprintf("%s.%s(%s, %s)", receiver, method, args[0], args[1]), true
}

func (g *generator) stringBufferReceiverCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "stringbuffer.length":
		return receiver + ".length()"
	case "stringbuffer.clear":
		return receiver + ".clear()"
	case "stringbuffer.append":
		return fmt.Sprintf("%s.append(%s)", receiver, args[0])
	case "stringbuffer.appendLine":
		return fmt.Sprintf("%s.appendLine(%s)", receiver, args[0])
	case "stringbuffer.toString":
		return receiver + ".toString()"
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}

func (g *generator) netConnectionReceiverCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "netConnection.read":
		return fmt.Sprintf("runeNetConnectionRead(%s, %s)", receiver, args[0])
	case "netConnection.write":
		return fmt.Sprintf("runeNetConnectionWrite(%s, %s)", receiver, args[0])
	case "netConnection.close":
		return fmt.Sprintf("runeNetConnectionClose(%s)", receiver)
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}

func (g *generator) netListenerReceiverCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "netListener.address":
		return fmt.Sprintf("runeNetListenerAddress(%s)", receiver)
	case "netListener.accept":
		return fmt.Sprintf("runeNetListenerAccept(%s)", receiver)
	case "netListener.close":
		return fmt.Sprintf("runeNetListenerClose(%s)", receiver)
	default:
		return g.unsupportedIntrinsic(fn, resultType)
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
	case "array.each":
		if len(args) != 1 {
			return "undefined"
		}
		return fmt.Sprintf("(() => { for (const [__index, __value] of %s.entries()) { (%s)(__value, __index, %s); } })()", receiver, args[0], receiver)
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
	case "bool.not":
		return "!" + receiver
	case "bool.xor":
		if len(args) != 1 {
			return "undefined"
		}
		return fmt.Sprintf("%s !== %s", receiver, args[0])
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
	case "map.each":
		if len(args) != 1 {
			return "undefined"
		}
		return fmt.Sprintf("(() => { for (const [__key, __value] of %s.entries()) { (%s)(__value, __key, %s); } })()", receiver, args[0], receiver)
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
	case "set.each":
		if len(args) != 1 {
			return "undefined"
		}
		return fmt.Sprintf("(() => { for (const __value of %s.values()) { (%s)(__value, __value, %s); } })()", receiver, args[0], receiver)
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}
