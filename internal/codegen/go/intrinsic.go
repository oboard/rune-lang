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
	if !ok {
		return "", false
	}
	if sel, ok := call.Callee.(*ir.SelectorExpr); ok {
		if at, ok := sel.Receiver.(*ir.AtExpr); ok && at.Name == "iter" && fn.Intrinsic == "" {
			return g.iterModuleCall(fn, g.intrinsicArgs(call.Args), call.ResultType()), true
		}
	}
	if fn.Intrinsic == "" && fn.Go == nil {
		return "", false
	}
	if fn.Go != nil && fn.Go.Symbol != "" {
		return fmt.Sprintf("%s(%s)", fn.Go.Symbol, strings.Join(g.intrinsicArgs(call.Args), ", ")), true
	}
	args := g.intrinsicArgs(call.Args)
	switch fn.Intrinsic {
	case "int.toString", "int4.fromInt", "int8.fromInt", "int16.fromInt", "int64.fromInt",
		"uint.fromInt", "uint8.fromInt", "uint16.fromInt", "uint64.fromInt",
		"float.fromDouble", "int4.toInt", "int8.toInt", "int16.toInt", "int64.toInt",
		"uint.toInt", "uint8.toInt", "uint16.toInt", "uint64.toInt", "float.toDouble":
		return g.numericIntrinsicCall(fn, args, call.ResultType()), true
	case "bytes.new", "bytes.fromInts":
		return g.bytesModuleCall(fn, args, call.ResultType()), true
	case "buffer.new", "buffer.fromBytes", "reader.new", "writer.new", "writer.withCapacity":
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
	case "compress.gzip", "compress.gunzip", "compress.deflate", "compress.inflate",
		"compress.brotli", "compress.unbrotli", "compress.zstd", "compress.unzstd",
		"compress.gzipText", "compress.gunzipText", "compress.brotliText", "compress.unbrotliText",
		"compress.zstdText", "compress.unzstdText":
		return g.compressModuleCall(fn, args, call.ResultType()), true
	case "net.connect", "net.listen":
		return g.netModuleCall(fn, args, call.ResultType()), true
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
	case strings.HasPrefix(fn.Intrinsic, "int."), strings.HasPrefix(fn.Intrinsic, "string."), strings.HasPrefix(fn.Intrinsic, "char."), strings.HasPrefix(fn.Intrinsic, "bool."), strings.HasPrefix(fn.Intrinsic, "regex."):
		return g.primitiveIntrinsicCall(fn, g.expr(sel.Receiver), g.intrinsicArgs(call.Args), call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "bytes."):
		return g.bytesReceiverCall(fn, g.expr(sel.Receiver), g.intrinsicArgs(call.Args), call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "buffer."):
		return g.bufferReceiverCall(fn, g.expr(sel.Receiver), g.intrinsicArgs(call.Args), call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "reader."):
		return g.readerReceiverCall(fn, g.expr(sel.Receiver), g.intrinsicArgs(call.Args), call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "writer."):
		return g.writerReceiverCall(fn, g.expr(sel.Receiver), g.intrinsicArgs(call.Args), call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "stringbuffer."):
		return g.stringBufferReceiverCall(fn, g.expr(sel.Receiver), g.intrinsicArgs(call.Args), call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "netConnection."):
		return g.netConnectionReceiverCall(fn, g.expr(sel.Receiver), g.intrinsicArgs(call.Args), call.ResultType()), true
	case strings.HasPrefix(fn.Intrinsic, "netListener."):
		return g.netListenerReceiverCall(fn, g.expr(sel.Receiver), g.intrinsicArgs(call.Args), call.ResultType()), true
	default:
		return g.unsupportedIntrinsic(fn, call.ResultType()), true
	}
}

func (g *generator) fsModuleCall(fn *stdlib.Function, args []string, resultType checker.Type) string {
	if len(args) == 0 {
		return g.zeroValue(resultType)
	}
	switch fn.Intrinsic {
	case "fs.readFileText":
		return fmt.Sprintf("runeFsReadFileText(%s)", args[0])
	case "fs.writeFile":
		if len(args) != 2 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("runeFsWriteFile(%s, %s)", args[0], args[1])
	case "fs.writeFileText":
		if len(args) != 2 {
			return g.zeroValue(resultType)
		}
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
		return "newRuneStringBuffer()"
	case "stringbuffer.from":
		return fmt.Sprintf("newRuneStringBufferFromString(%s)", args[0])
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}

func (g *generator) iterModuleCall(fn *stdlib.Function, args []string, resultType checker.Type) string {
	elem, ok := checker.IterValue(resultType)
	if !ok {
		elem = checker.Unknown
	}
	elemType := goType(elem)
	nextType := fmt.Sprintf("struct{ F0 %s; F1 bool }", elemType)
	iterType := goType(resultType)
	switch fn.Name {
	case "new":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("%s{__next: %s}", iterType, args[0])
	case "fromArray":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("func() %s { values := %s; index := -1; return %s{__next: func() %s { index++; if index >= len(values) { var zero %s; return %s{F0: zero, F1: false} }; return %s{F0: values[index], F1: true} }} }()", iterType, args[0], iterType, nextType, elemType, nextType, nextType)
	case "range":
		if len(args) != 2 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("func() %s { start := %s; end := %s; step := 1; value := start - step; return %s{__next: func() %s { value += step; return %s{F0: value, F1: value < end} }} }()", iterType, args[0], args[1], iterType, nextType, nextType)
	case "rangeStep":
		if len(args) != 3 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("func() %s { start := %s; end := %s; step := %s; value := start - step; return %s{__next: func() %s { value += step; if step == 0 { return %s{F0: 0, F1: false} }; if step > 0 { return %s{F0: value, F1: value < end} }; return %s{F0: value, F1: value > end} }} }()", iterType, args[0], args[1], args[2], iterType, nextType, nextType, nextType, nextType)
	case "repeat":
		if len(args) != 2 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("func() %s { value := %s; count := %s; index := -1; return %s{__next: func() %s { index++; return %s{F0: value, F1: index < count} }} }()", iterType, args[0], args[1], iterType, nextType, nextType)
	case "empty":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("func() %s { value := %s; return %s{__next: func() %s { return %s{F0: value, F1: false} }} }()", iterType, args[0], iterType, nextType, nextType)
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
		return "newRuneBuffer()"
	case "buffer.fromBytes":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("newRuneBufferFromBytes(%s)", args[0])
	case "reader.new":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("newRuneReader(%s)", args[0])
	case "writer.new":
		if len(args) != 0 {
			return g.zeroValue(resultType)
		}
		return "newRuneWriter()"
	case "writer.withCapacity":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("newRuneWriterWithCapacity(%s)", args[0])
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
		return fmt.Sprintf("strconv.Itoa(%s)", value)
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

func (g *generator) bytesModuleCall(fn *stdlib.Function, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "bytes.new":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("newRuneBytes(%s)", args[0])
	case "bytes.fromInts":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("runeBytesFromInts(%s)", args[0])
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}

func (g *generator) bytesReceiverCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "bytes.length":
		return fmt.Sprintf("%s.ByteLength()", receiver)
	case "bytes.clone":
		return fmt.Sprintf("%s.Clone()", receiver)
	case "bytes.slice":
		if len(args) != 2 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("%s.Slice(%s, %s)", receiver, args[0], args[1])
	case "bytes.toInts":
		return fmt.Sprintf("%s.ToInts()", receiver)
	case "bytes.getInt4":
		return fmt.Sprintf("%s.GetInt4(%s)", receiver, args[0])
	case "bytes.setInt4":
		return fmt.Sprintf("%s.SetInt4(%s, %s)", receiver, args[0], args[1])
	}
	methods := map[string]string{
		"bytes.getInt8":   "GetInt8",
		"bytes.getUInt8":  "GetUInt8",
		"bytes.getInt16":  "GetInt16",
		"bytes.getUInt16": "GetUInt16",
		"bytes.getInt":    "GetInt",
		"bytes.getUInt":   "GetUInt",
		"bytes.getInt64":  "GetInt64",
		"bytes.getUInt64": "GetUInt64",
		"bytes.getFloat":  "GetFloat",
		"bytes.getDouble": "GetDouble",
		"bytes.setInt8":   "SetInt8",
		"bytes.setUInt8":  "SetUInt8",
		"bytes.setInt16":  "SetInt16",
		"bytes.setUInt16": "SetUInt16",
		"bytes.setInt":    "SetInt",
		"bytes.setUInt":   "SetUInt",
		"bytes.setInt64":  "SetInt64",
		"bytes.setUInt64": "SetUInt64",
		"bytes.setFloat":  "SetFloat",
		"bytes.setDouble": "SetDouble",
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

func (g *generator) bufferReceiverCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "buffer.length":
		return fmt.Sprintf("%s.ByteLength()", receiver)
	case "buffer.clear":
		return fmt.Sprintf("%s.Clear()", receiver)
	case "buffer.clone":
		return fmt.Sprintf("%s.Clone()", receiver)
	case "buffer.toBytes":
		return fmt.Sprintf("%s.ToBytes()", receiver)
	case "buffer.toInts":
		return fmt.Sprintf("%s.ToInts()", receiver)
	case "buffer.append":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("%s.Append(%s)", receiver, args[0])
	case "buffer.appendInt":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("%s.AppendInt(%s)", receiver, args[0])
	case "buffer.appendBytes":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("%s.AppendBytes(%s)", receiver, args[0])
	case "buffer.reader":
		return fmt.Sprintf("%s.Reader()", receiver)
	case "buffer.writer":
		return fmt.Sprintf("%s.Writer()", receiver)
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}

func (g *generator) readerReceiverCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "reader.length":
		return fmt.Sprintf("%s.ByteLength()", receiver)
	case "reader.position":
		return fmt.Sprintf("%s.Position()", receiver)
	case "reader.remaining":
		return fmt.Sprintf("%s.Remaining()", receiver)
	case "reader.seek":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("%s.Seek(%s)", receiver, args[0])
	case "reader.skip":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("%s.Skip(%s)", receiver, args[0])
	case "reader.readBytes":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("%s.ReadBytes(%s)", receiver, args[0])
	case "reader.readInt4":
		return fmt.Sprintf("%s.ReadInt4()", receiver)
	}
	methods := map[string]string{
		"reader.readInt8":   "ReadInt8",
		"reader.readUInt8":  "ReadUInt8",
		"reader.readInt16":  "ReadInt16",
		"reader.readUInt16": "ReadUInt16",
		"reader.readInt":    "ReadInt",
		"reader.readUInt":   "ReadUInt",
		"reader.readInt64":  "ReadInt64",
		"reader.readUInt64": "ReadUInt64",
		"reader.readFloat":  "ReadFloat",
		"reader.readDouble": "ReadDouble",
	}
	method, ok := methods[fn.Intrinsic]
	if !ok {
		return g.unsupportedIntrinsic(fn, resultType)
	}
	if strings.HasSuffix(method, "8") {
		return fmt.Sprintf("%s.%s()", receiver, method)
	}
	return fmt.Sprintf("%s.%s(%s)", receiver, method, strings.Join(args, ", "))
}

func (g *generator) writerReceiverCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "writer.length", "writer.position":
		return fmt.Sprintf("%s.Position()", receiver)
	case "writer.clear":
		return fmt.Sprintf("%s.Clear()", receiver)
	case "writer.toBytes":
		return fmt.Sprintf("%s.ToBytes()", receiver)
	case "writer.toInts":
		return fmt.Sprintf("%s.ToInts()", receiver)
	case "writer.writeBytes":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("%s.WriteBytes(%s)", receiver, args[0])
	case "writer.writeInt4":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("%s.WriteInt4(%s)", receiver, args[0])
	}
	methods := map[string]string{
		"writer.writeInt8":   "WriteInt8",
		"writer.writeUInt8":  "WriteUInt8",
		"writer.writeInt16":  "WriteInt16",
		"writer.writeUInt16": "WriteUInt16",
		"writer.writeInt":    "WriteInt",
		"writer.writeUInt":   "WriteUInt",
		"writer.writeInt64":  "WriteInt64",
		"writer.writeUInt64": "WriteUInt64",
		"writer.writeFloat":  "WriteFloat",
		"writer.writeDouble": "WriteDouble",
	}
	method, ok := methods[fn.Intrinsic]
	if !ok {
		return g.unsupportedIntrinsic(fn, resultType)
	}
	return fmt.Sprintf("%s.%s(%s)", receiver, method, strings.Join(args, ", "))
}

func (g *generator) stringBufferReceiverCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "stringbuffer.length":
		return fmt.Sprintf("%s.Length()", receiver)
	case "stringbuffer.clear":
		return fmt.Sprintf("%s.Clear()", receiver)
	case "stringbuffer.append":
		return fmt.Sprintf("%s.Append(%s)", receiver, args[0])
	case "stringbuffer.appendLine":
		return fmt.Sprintf("%s.AppendLine(%s)", receiver, args[0])
	case "stringbuffer.toString":
		return fmt.Sprintf("%s.ToString()", receiver)
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}

func (g *generator) netConnectionReceiverCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "netConnection.read":
		return fmt.Sprintf("%s.Read(%s)", receiver, args[0])
	case "netConnection.write":
		return fmt.Sprintf("%s.Write(%s)", receiver, args[0])
	case "netConnection.close":
		return fmt.Sprintf("%s.Close()", receiver)
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}

func (g *generator) netListenerReceiverCall(fn *stdlib.Function, receiver string, args []string, resultType checker.Type) string {
	switch fn.Intrinsic {
	case "netListener.address":
		return fmt.Sprintf("%s.Address()", receiver)
	case "netListener.accept":
		return fmt.Sprintf("%s.Accept()", receiver)
	case "netListener.close":
		return fmt.Sprintf("%s.Close()", receiver)
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
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
	case "int.toString":
		return fmt.Sprintf("strconv.Itoa(%s)", receiver)
	case "string.length":
		return fmt.Sprintf("len([]rune(%s))", receiver)
	case "string.toString":
		return receiver
	case "string.at":
		if len(args) != 1 {
			return "/* invalid string.at */"
		}
		return fmt.Sprintf("[]rune(%s)[%s]", receiver, args[0])
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
	case "char.toString":
		return fmt.Sprintf("string(%s)", receiver)
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
