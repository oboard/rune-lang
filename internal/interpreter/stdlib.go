package interpreter

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"github.com/oboard/rune-lang/internal/ir"
)

func (i *Interpreter) callModuleFunction(module string, name string, args []ir.Expr, env *Env) (Value, error) {
	if i.file.Stdlib == nil {
		return nil, fmt.Errorf("stdlib is not loaded")
	}
	fn, ok := i.file.Stdlib.Function(module, name)
	if !ok {
		return nil, fmt.Errorf("unknown module function @%s.%s", module, name)
	}
	values, err := i.evalArgs(args, env)
	if err != nil {
		return nil, err
	}
	if i.compileTime && fn.Body == nil {
		return nil, fmt.Errorf("compile-time macro cannot call @%s.%s", module, name)
	}
	if fn.Body != nil {
		local := NewEnv(i.globals)
		for idx, param := range fn.ParamNames {
			if idx < len(values) {
				local.Define(param, values[idx])
			}
		}
		return i.eval(ir.LowerExpr(fn.Body, nil), local)
	}
	if fn.Go != nil {
		return i.callGoBackedFunction(fn.Go.Symbol, values)
	}
	switch fn.Intrinsic {
	case "io.scan":
		if len(values) != 0 {
			return nil, fmt.Errorf("@io.scan expects 0 args, got %d", len(values))
		}
		var token string
		if _, err := fmt.Fscan(i.in, &token); err != nil {
			if err == io.EOF {
				return NullValue, nil
			}
			return nil, err
		}
		return token, nil
	case "io.scanLine":
		if len(values) != 0 {
			return nil, fmt.Errorf("@io.scanLine expects 0 args, got %d", len(values))
		}
		line, err := i.in.ReadString('\n')
		if err != nil && len(line) == 0 {
			if err == io.EOF {
				return NullValue, nil
			}
			return nil, err
		}
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
		return line, nil
	case "io.readAll":
		if len(values) != 0 {
			return nil, fmt.Errorf("@io.readAll expects 0 args, got %d", len(values))
		}
		data, err := io.ReadAll(i.in)
		if err != nil {
			return nil, err
		}
		return string(data), nil
	case "int4.fromInt", "int8.fromInt", "int16.fromInt", "int64.fromInt",
		"uint.fromInt", "uint8.fromInt", "uint16.fromInt", "uint64.fromInt",
		"float.fromDouble", "int4.toInt", "int8.toInt", "int16.toInt", "int64.toInt",
		"uint.toInt", "uint8.toInt", "uint16.toInt", "uint64.toInt", "float.toDouble":
		return callNumericIntrinsic(fn.Intrinsic, values)
	case "bytes.new":
		if len(values) != 1 {
			return nil, fmt.Errorf("@bytes.new expects 1 arg, got %d", len(values))
		}
		length, ok := values[0].(int)
		if !ok {
			return nil, fmt.Errorf("@bytes.new length expects Int")
		}
		if length < 0 {
			return nil, fmt.Errorf("@bytes.new length out of range")
		}
		return &Bytes{Data: make([]byte, length)}, nil
	case "bytes.fromInts":
		if len(values) != 1 {
			return nil, fmt.Errorf("@bytes.fromInts expects 1 arg, got %d", len(values))
		}
		array, ok := values[0].(*Array)
		if !ok {
			return nil, fmt.Errorf("@bytes.fromInts expects Array[Int]")
		}
		out := make([]byte, len(array.Elements))
		for idx, elem := range array.Elements {
			value, ok := elem.(int)
			if !ok {
				return nil, fmt.Errorf("@bytes.fromInts element %d expects Int", idx)
			}
			out[idx] = byte(value)
		}
		return &Bytes{Data: out}, nil
	case "buffer.new":
		if len(values) != 0 {
			return nil, fmt.Errorf("@buffer.new expects 0 args, got %d", len(values))
		}
		return &Buffer{}, nil
	case "buffer.fromBytes":
		if len(values) != 1 {
			return nil, fmt.Errorf("@buffer.fromBytes expects 1 arg, got %d", len(values))
		}
		value, ok := values[0].(*Bytes)
		if !ok {
			return nil, fmt.Errorf("@buffer.fromBytes expects Bytes")
		}
		return &Buffer{Data: append([]byte(nil), value.Data...)}, nil
	case "reader.new":
		if len(values) != 1 {
			return nil, fmt.Errorf("@reader.new expects 1 arg, got %d", len(values))
		}
		value, ok := values[0].(*Bytes)
		if !ok {
			return nil, fmt.Errorf("@reader.new expects Bytes")
		}
		return &Reader{Data: append([]byte(nil), value.Data...)}, nil
	case "writer.new":
		if len(values) != 0 {
			return nil, fmt.Errorf("@writer.new expects 0 args, got %d", len(values))
		}
		return &Writer{}, nil
	case "writer.withCapacity":
		if len(values) != 1 {
			return nil, fmt.Errorf("@writer.withCapacity expects 1 arg, got %d", len(values))
		}
		capacity, ok := values[0].(int)
		if !ok {
			return nil, fmt.Errorf("@writer.withCapacity expects Int")
		}
		if capacity < 0 {
			return nil, fmt.Errorf("@writer.withCapacity capacity out of range")
		}
		return &Writer{Data: make([]byte, 0, capacity)}, nil
	case "stringbuffer.new":
		if len(values) != 0 {
			return nil, fmt.Errorf("@stringbuffer.new expects 0 args, got %d", len(values))
		}
		return &StringBuffer{}, nil
	case "stringbuffer.from":
		if len(values) != 1 {
			return nil, fmt.Errorf("@stringbuffer.from expects 1 arg, got %d", len(values))
		}
		value, ok := values[0].(string)
		if !ok {
			return nil, fmt.Errorf("@stringbuffer.from expects String")
		}
		return &StringBuffer{Parts: []string{value}}, nil
	case "path.basename":
		return pathStringUnary(values, "@path.basename", filepath.Base)
	case "path.dirname":
		return pathStringUnary(values, "@path.dirname", filepath.Dir)
	case "path.extname":
		return pathStringUnary(values, "@path.extname", filepath.Ext)
	case "path.join":
		parts, err := stringArrayArg(values, "@path.join")
		if err != nil {
			return nil, err
		}
		return filepath.Join(parts...), nil
	case "path.normalize":
		return pathStringUnary(values, "@path.normalize", filepath.Clean)
	case "path.resolve":
		parts, err := stringArrayArg(values, "@path.resolve")
		if err != nil {
			return nil, err
		}
		joined := filepath.Join(parts...)
		if abs, err := filepath.Abs(joined); err == nil {
			return abs, nil
		}
		return joined, nil
	case "path.relative":
		if len(values) != 2 {
			return nil, fmt.Errorf("@path.relative expects 2 args, got %d", len(values))
		}
		from, ok := values[0].(string)
		if !ok {
			return nil, fmt.Errorf("@path.relative from expects String")
		}
		to, ok := values[1].(string)
		if !ok {
			return nil, fmt.Errorf("@path.relative to expects String")
		}
		rel, err := filepath.Rel(from, to)
		if err != nil {
			return nil, err
		}
		return rel, nil
	case "path.isAbsolute":
		if len(values) != 1 {
			return nil, fmt.Errorf("@path.isAbsolute expects 1 arg, got %d", len(values))
		}
		path, ok := values[0].(string)
		if !ok {
			return nil, fmt.Errorf("@path.isAbsolute expects String")
		}
		return filepath.IsAbs(path), nil
	case "process.argv":
		if len(values) != 0 {
			return nil, fmt.Errorf("@process.argv expects 0 args, got %d", len(values))
		}
		out := &Array{Elements: make([]Value, 0, len(os.Args))}
		for _, arg := range os.Args {
			out.Elements = append(out.Elements, arg)
		}
		return out, nil
	case "process.cwd":
		if len(values) != 0 {
			return nil, fmt.Errorf("@process.cwd expects 0 args, got %d", len(values))
		}
		return os.Getwd()
	case "process.env":
		if len(values) != 1 {
			return nil, fmt.Errorf("@process.env expects 1 arg, got %d", len(values))
		}
		name, ok := values[0].(string)
		if !ok {
			return nil, fmt.Errorf("@process.env expects String")
		}
		if value, ok := os.LookupEnv(name); ok {
			return value, nil
		}
		return NullValue, nil
	case "process.exit":
		if len(values) != 1 {
			return nil, fmt.Errorf("@process.exit expects 1 arg, got %d", len(values))
		}
		code, ok := values[0].(int)
		if !ok {
			return nil, fmt.Errorf("@process.exit expects Int")
		}
		os.Exit(code)
		return nil, nil
	case "process.platform":
		if len(values) != 0 {
			return nil, fmt.Errorf("@process.platform expects 0 args, got %d", len(values))
		}
		return runtime.GOOS, nil
	case "map.new":
		if len(values) != 2 {
			return nil, fmt.Errorf("@map.new expects 2 args, got %d", len(values))
		}
		return &Map{Entries: map[string]mapEntry{}}, nil
	case "set.new":
		if len(values) != 1 {
			return nil, fmt.Errorf("@set.new expects 1 arg, got %d", len(values))
		}
		return &Set{Entries: map[string]Value{}}, nil
	case "assert.eq":
		if len(values) != 2 {
			return nil, fmt.Errorf("@assert.eq expects 2 args, got %d", len(values))
		}
		if valuesEqual(values[0], values[1]) {
			return nil, nil
		}
		return nil, fmt.Errorf("assert.eq failed: actual %s, expected %s", Format(values[0]), Format(values[1]))
	case "json.stringify":
		if len(values) != 1 {
			return nil, fmt.Errorf("@json.stringify expects 1 arg, got %d", len(values))
		}
		return jsonStringify(values[0])
	case "regex.new":
		if len(values) != 2 {
			return nil, fmt.Errorf("@regex.new expects 2 args, got %d", len(values))
		}
		pattern, ok := values[0].(string)
		if !ok {
			return nil, fmt.Errorf("@regex.new pattern expects String")
		}
		flags, ok := values[1].(string)
		if !ok {
			return nil, fmt.Errorf("@regex.new flags expects String")
		}
		return newRegex(pattern, flags)
	case "regex.escape":
		if len(values) != 1 {
			return nil, fmt.Errorf("@regex.escape expects 1 arg, got %d", len(values))
		}
		value, ok := values[0].(string)
		if !ok {
			return nil, fmt.Errorf("@regex.escape expects String")
		}
		return regexp.QuoteMeta(value), nil
	case "go.stmt", "go.expr", "go.import":
		return nil, fmt.Errorf("@%s.%s is only supported by the Go backend", module, name)
	default:
		return nil, fmt.Errorf("@%s.%s is not supported by the interpreter", module, name)
	}
}

func callNumericIntrinsic(intrinsic string, values []Value) (Value, error) {
	if len(values) != 1 {
		return nil, fmt.Errorf("%s expects 1 arg, got %d", intrinsic, len(values))
	}
	value := values[0]
	switch intrinsic {
	case "int4.fromInt":
		n, ok := value.(int)
		if !ok {
			return nil, fmt.Errorf("%s expects Int", intrinsic)
		}
		return int4(n), nil
	case "int8.fromInt":
		n, ok := value.(int)
		if !ok {
			return nil, fmt.Errorf("%s expects Int", intrinsic)
		}
		return int8(n), nil
	case "int16.fromInt":
		n, ok := value.(int)
		if !ok {
			return nil, fmt.Errorf("%s expects Int", intrinsic)
		}
		return int16(n), nil
	case "int64.fromInt":
		n, ok := value.(int)
		if !ok {
			return nil, fmt.Errorf("%s expects Int", intrinsic)
		}
		return int64(n), nil
	case "uint.fromInt":
		n, ok := value.(int)
		if !ok {
			return nil, fmt.Errorf("%s expects Int", intrinsic)
		}
		return uint(n), nil
	case "uint8.fromInt":
		n, ok := value.(int)
		if !ok {
			return nil, fmt.Errorf("%s expects Int", intrinsic)
		}
		return uint8(n), nil
	case "uint16.fromInt":
		n, ok := value.(int)
		if !ok {
			return nil, fmt.Errorf("%s expects Int", intrinsic)
		}
		return uint16(n), nil
	case "uint64.fromInt":
		n, ok := value.(int)
		if !ok {
			return nil, fmt.Errorf("%s expects Int", intrinsic)
		}
		return uint64(n), nil
	case "float.fromDouble":
		n, ok := value.(float64)
		if !ok {
			return nil, fmt.Errorf("%s expects Double", intrinsic)
		}
		return float32(n), nil
	case "int4.toInt", "int8.toInt":
		n, ok := value.(int8)
		if !ok {
			return nil, fmt.Errorf("%s expects Int8", intrinsic)
		}
		return int(n), nil
	case "int16.toInt":
		n, ok := value.(int16)
		if !ok {
			return nil, fmt.Errorf("%s expects Int16", intrinsic)
		}
		return int(n), nil
	case "int64.toInt":
		n, ok := value.(int64)
		if !ok {
			return nil, fmt.Errorf("%s expects Int64", intrinsic)
		}
		return int(n), nil
	case "uint.toInt":
		n, ok := value.(uint)
		if !ok {
			return nil, fmt.Errorf("%s expects UInt", intrinsic)
		}
		return int(n), nil
	case "uint8.toInt":
		n, ok := value.(uint8)
		if !ok {
			return nil, fmt.Errorf("%s expects UInt8", intrinsic)
		}
		return int(n), nil
	case "uint16.toInt":
		n, ok := value.(uint16)
		if !ok {
			return nil, fmt.Errorf("%s expects UInt16", intrinsic)
		}
		return int(n), nil
	case "uint64.toInt":
		n, ok := value.(uint64)
		if !ok {
			return nil, fmt.Errorf("%s expects UInt64", intrinsic)
		}
		return int(n), nil
	case "float.toDouble":
		n, ok := value.(float32)
		if !ok {
			return nil, fmt.Errorf("%s expects Float", intrinsic)
		}
		return float64(n), nil
	default:
		return nil, fmt.Errorf("%s is not supported by the interpreter", intrinsic)
	}
}

func int4(value int) int8 {
	n := value & 0xf
	if n >= 8 {
		return int8(n - 16)
	}
	return int8(n)
}

func pathStringUnary(values []Value, name string, fn func(string) string) (Value, error) {
	if len(values) != 1 {
		return nil, fmt.Errorf("%s expects 1 arg, got %d", name, len(values))
	}
	value, ok := values[0].(string)
	if !ok {
		return nil, fmt.Errorf("%s expects String", name)
	}
	return fn(value), nil
}

func stringArrayArg(values []Value, name string) ([]string, error) {
	if len(values) != 1 {
		return nil, fmt.Errorf("%s expects 1 arg, got %d", name, len(values))
	}
	array, ok := values[0].(*Array)
	if !ok {
		return nil, fmt.Errorf("%s expects Array[String]", name)
	}
	out := make([]string, 0, len(array.Elements))
	for idx, elem := range array.Elements {
		value, ok := elem.(string)
		if !ok {
			return nil, fmt.Errorf("%s element %d expects String", name, idx)
		}
		out = append(out, value)
	}
	return out, nil
}

func (i *Interpreter) callStringBufferMethod(value *StringBuffer, name string, args []ir.Expr, env *Env) (Value, error) {
	if i.file.Stdlib == nil {
		return nil, fmt.Errorf("stdlib is not loaded")
	}
	fn, ok := i.file.Stdlib.ReceiverFunction("stringbuffer", "StringBuffer", name)
	if !ok {
		return nil, fmt.Errorf("type StringBuffer has no method %q", name)
	}
	values, err := i.evalArgs(args, env)
	if err != nil {
		return nil, err
	}
	switch fn.Intrinsic {
	case "stringbuffer.length":
		return len([]rune(strings.Join(value.Parts, ""))), nil
	case "stringbuffer.clear":
		value.Parts = nil
		return nil, nil
	case "stringbuffer.append":
		if len(values) != 1 {
			return nil, fmt.Errorf("stringbuffer.append expects 1 arg, got %d", len(values))
		}
		text, ok := values[0].(string)
		if !ok {
			return nil, fmt.Errorf("stringbuffer.append expects String")
		}
		value.Parts = append(value.Parts, text)
		return value, nil
	case "stringbuffer.appendLine":
		if len(values) != 1 {
			return nil, fmt.Errorf("stringbuffer.appendLine expects 1 arg, got %d", len(values))
		}
		text, ok := values[0].(string)
		if !ok {
			return nil, fmt.Errorf("stringbuffer.appendLine expects String")
		}
		value.Parts = append(value.Parts, text, "\n")
		return value, nil
	case "stringbuffer.toString":
		return strings.Join(value.Parts, ""), nil
	default:
		if fn.Body == nil {
			return nil, fmt.Errorf("stringbuffer.%s is not supported by the interpreter", name)
		}
		local := NewEnv(env)
		local.Define("this", value)
		for idx, param := range fn.ParamNames {
			if idx < len(values) {
				local.Define(param, values[idx])
			}
		}
		return i.eval(ir.LowerExpr(fn.Body, nil), local)
	}
}

func (i *Interpreter) callBytesMethod(value *Bytes, name string, args []ir.Expr, env *Env) (Value, error) {
	if i.file.Stdlib == nil {
		return nil, fmt.Errorf("stdlib is not loaded")
	}
	fn, ok := i.file.Stdlib.ReceiverFunction("bytes", "Bytes", name)
	if !ok {
		return nil, fmt.Errorf("type Bytes has no method %q", name)
	}
	values, err := i.evalArgs(args, env)
	if err != nil {
		return nil, err
	}
	switch fn.Intrinsic {
	case "bytes.length":
		if len(values) != 0 {
			return nil, fmt.Errorf("bytes.%s expects 0 args, got %d", name, len(values))
		}
		return len(value.Data), nil
	case "bytes.clone":
		if len(values) != 0 {
			return nil, fmt.Errorf("bytes.clone expects 0 args, got %d", len(values))
		}
		return &Bytes{Data: append([]byte(nil), value.Data...)}, nil
	case "bytes.slice":
		if len(values) != 2 {
			return nil, fmt.Errorf("bytes.slice expects 2 args, got %d", len(values))
		}
		start, ok := values[0].(int)
		if !ok {
			return nil, fmt.Errorf("bytes.slice start expects Int")
		}
		end, ok := values[1].(int)
		if !ok {
			return nil, fmt.Errorf("bytes.slice end expects Int")
		}
		if err := checkBytesRange(value, start, end-start); err != nil {
			return nil, err
		}
		return &Bytes{Data: append([]byte(nil), value.Data[start:end]...)}, nil
	case "bytes.toInts":
		if len(values) != 0 {
			return nil, fmt.Errorf("bytes.toInts expects 0 args, got %d", len(values))
		}
		out := &Array{Elements: make([]Value, 0, len(value.Data))}
		for _, b := range value.Data {
			out.Elements = append(out.Elements, int(b))
		}
		return out, nil
	case "bytes.getInt4":
		offset, err := bytesOffset(values, 1, name)
		if err != nil {
			return nil, err
		}
		return bytesGetInt4(value, offset)
	case "bytes.setInt4":
		if len(values) != 2 {
			return nil, fmt.Errorf("bytes.%s expects 2 args, got %d", name, len(values))
		}
		offset, ok := values[0].(int)
		if !ok {
			return nil, fmt.Errorf("bytes.%s offset expects Int", name)
		}
		n, ok := values[1].(int8)
		if !ok {
			return nil, fmt.Errorf("bytes.%s value expects Int4", name)
		}
		return bytesSetInt4(value, offset, n)
	}
	if result, ok, err := callBytesDataViewMethod(value, name, fn.Intrinsic, values); ok || err != nil {
		return result, err
	}
	return nil, fmt.Errorf("bytes.%s is not supported by the interpreter", name)
}

func callBytesDataViewMethod(value *Bytes, name string, intrinsic string, values []Value) (Value, bool, error) {
	switch intrinsic {
	case "bytes.getInt8":
		offset, err := bytesOffset(values, 1, name)
		if err != nil {
			return nil, true, err
		}
		if err := checkBytesRange(value, offset, 1); err != nil {
			return nil, true, err
		}
		return int8(value.Data[offset]), true, nil
	case "bytes.setInt8":
		offset, err := bytesOffset(values, 2, name)
		if err != nil {
			return nil, true, err
		}
		n, ok := values[1].(int8)
		if !ok {
			return nil, true, fmt.Errorf("bytes.%s value expects Int8", name)
		}
		if err := checkBytesRange(value, offset, 1); err != nil {
			return nil, true, err
		}
		value.Data[offset] = byte(n)
		return n, true, nil
	case "bytes.getUInt8":
		offset, err := bytesOffset(values, 1, name)
		if err != nil {
			return nil, true, err
		}
		if err := checkBytesRange(value, offset, 1); err != nil {
			return nil, true, err
		}
		return value.Data[offset], true, nil
	case "bytes.setUInt8":
		offset, err := bytesOffset(values, 2, name)
		if err != nil {
			return nil, true, err
		}
		n, ok := values[1].(uint8)
		if !ok {
			return nil, true, fmt.Errorf("bytes.%s value expects UInt8", name)
		}
		if err := checkBytesRange(value, offset, 1); err != nil {
			return nil, true, err
		}
		value.Data[offset] = n
		return n, true, nil
	case "bytes.getInt16":
		offset, order, err := bytesOffsetOrder(value, values, 2, name, 2)
		if err != nil {
			return nil, true, err
		}
		return int16(order.Uint16(value.Data[offset:])), true, nil
	case "bytes.setInt16":
		offset, order, err := bytesOffsetOrder(value, values, 3, name, 2)
		if err != nil {
			return nil, true, err
		}
		n, ok := values[1].(int16)
		if !ok {
			return nil, true, fmt.Errorf("bytes.%s value expects Int16", name)
		}
		order.PutUint16(value.Data[offset:], uint16(n))
		return n, true, nil
	case "bytes.getUInt16":
		offset, order, err := bytesOffsetOrder(value, values, 2, name, 2)
		if err != nil {
			return nil, true, err
		}
		return order.Uint16(value.Data[offset:]), true, nil
	case "bytes.setUInt16":
		offset, order, err := bytesOffsetOrder(value, values, 3, name, 2)
		if err != nil {
			return nil, true, err
		}
		n, ok := values[1].(uint16)
		if !ok {
			return nil, true, fmt.Errorf("bytes.%s value expects UInt16", name)
		}
		order.PutUint16(value.Data[offset:], n)
		return n, true, nil
	case "bytes.getInt":
		offset, order, err := bytesOffsetOrder(value, values, 2, name, 4)
		if err != nil {
			return nil, true, err
		}
		return int(int32(order.Uint32(value.Data[offset:]))), true, nil
	case "bytes.setInt":
		offset, order, err := bytesOffsetOrder(value, values, 3, name, 4)
		if err != nil {
			return nil, true, err
		}
		n, ok := values[1].(int)
		if !ok {
			return nil, true, fmt.Errorf("bytes.%s value expects Int", name)
		}
		order.PutUint32(value.Data[offset:], uint32(int32(n)))
		return n, true, nil
	case "bytes.getUInt":
		offset, order, err := bytesOffsetOrder(value, values, 2, name, 4)
		if err != nil {
			return nil, true, err
		}
		return uint(order.Uint32(value.Data[offset:])), true, nil
	case "bytes.setUInt":
		offset, order, err := bytesOffsetOrder(value, values, 3, name, 4)
		if err != nil {
			return nil, true, err
		}
		n, ok := values[1].(uint)
		if !ok {
			return nil, true, fmt.Errorf("bytes.%s value expects UInt", name)
		}
		order.PutUint32(value.Data[offset:], uint32(n))
		return n, true, nil
	case "bytes.getInt64":
		offset, order, err := bytesOffsetOrder(value, values, 2, name, 8)
		if err != nil {
			return nil, true, err
		}
		return int64(order.Uint64(value.Data[offset:])), true, nil
	case "bytes.setInt64":
		offset, order, err := bytesOffsetOrder(value, values, 3, name, 8)
		if err != nil {
			return nil, true, err
		}
		n, ok := values[1].(int64)
		if !ok {
			return nil, true, fmt.Errorf("bytes.%s value expects Int64", name)
		}
		order.PutUint64(value.Data[offset:], uint64(n))
		return n, true, nil
	case "bytes.getUInt64":
		offset, order, err := bytesOffsetOrder(value, values, 2, name, 8)
		if err != nil {
			return nil, true, err
		}
		return order.Uint64(value.Data[offset:]), true, nil
	case "bytes.setUInt64":
		offset, order, err := bytesOffsetOrder(value, values, 3, name, 8)
		if err != nil {
			return nil, true, err
		}
		n, ok := values[1].(uint64)
		if !ok {
			return nil, true, fmt.Errorf("bytes.%s value expects UInt64", name)
		}
		order.PutUint64(value.Data[offset:], n)
		return n, true, nil
	case "bytes.getFloat":
		offset, order, err := bytesOffsetOrder(value, values, 2, name, 4)
		if err != nil {
			return nil, true, err
		}
		return math.Float32frombits(order.Uint32(value.Data[offset:])), true, nil
	case "bytes.setFloat":
		offset, order, err := bytesOffsetOrder(value, values, 3, name, 4)
		if err != nil {
			return nil, true, err
		}
		n, ok := values[1].(float32)
		if !ok {
			return nil, true, fmt.Errorf("bytes.%s value expects Float", name)
		}
		order.PutUint32(value.Data[offset:], math.Float32bits(n))
		return n, true, nil
	case "bytes.getDouble":
		offset, order, err := bytesOffsetOrder(value, values, 2, name, 8)
		if err != nil {
			return nil, true, err
		}
		return math.Float64frombits(order.Uint64(value.Data[offset:])), true, nil
	case "bytes.setDouble":
		offset, order, err := bytesOffsetOrder(value, values, 3, name, 8)
		if err != nil {
			return nil, true, err
		}
		n, ok := values[1].(float64)
		if !ok {
			return nil, true, fmt.Errorf("bytes.%s value expects Double", name)
		}
		order.PutUint64(value.Data[offset:], math.Float64bits(n))
		return n, true, nil
	default:
		return nil, false, nil
	}
}

func bytesOffset(values []Value, expected int, name string) (int, error) {
	if len(values) != expected {
		return 0, fmt.Errorf("bytes.%s expects %d args, got %d", name, expected, len(values))
	}
	offset, ok := values[0].(int)
	if !ok {
		return 0, fmt.Errorf("bytes.%s offset expects Int", name)
	}
	return offset, nil
}

func bytesOffsetOrder(value *Bytes, values []Value, expected int, name string, size int) (int, binary.ByteOrder, error) {
	offset, err := bytesOffset(values, expected, name)
	if err != nil {
		return 0, nil, err
	}
	littleEndian, ok := values[expected-1].(bool)
	if !ok {
		return 0, nil, fmt.Errorf("bytes.%s littleEndian expects Bool", name)
	}
	if err := checkBytesRange(value, offset, size); err != nil {
		return 0, nil, err
	}
	if littleEndian {
		return offset, binary.LittleEndian, nil
	}
	return offset, binary.BigEndian, nil
}

func checkBytesRange(value *Bytes, offset int, size int) error {
	if offset < 0 || size < 0 || offset+size > len(value.Data) {
		return fmt.Errorf("bytes offset out of range")
	}
	return nil
}

func (b *Bytes) GetInt8(offset int) int8 {
	_ = checkBytesRange(b, offset, 1)
	return int8(b.Data[offset])
}

func (b *Bytes) SetInt8(offset int, value int8) int8 {
	_ = checkBytesRange(b, offset, 1)
	b.Data[offset] = byte(value)
	return value
}

func (b *Bytes) GetUInt8(offset int) uint8 {
	_ = checkBytesRange(b, offset, 1)
	return b.Data[offset]
}

func (b *Bytes) SetUInt8(offset int, value uint8) uint8 {
	_ = checkBytesRange(b, offset, 1)
	b.Data[offset] = value
	return value
}

func (b *Bytes) GetInt16(offset int, littleEndian bool) int16 {
	_ = checkBytesRange(b, offset, 2)
	return int16(bytesOrder(littleEndian).Uint16(b.Data[offset:]))
}

func (b *Bytes) SetInt16(offset int, value int16, littleEndian bool) int16 {
	_ = checkBytesRange(b, offset, 2)
	bytesOrder(littleEndian).PutUint16(b.Data[offset:], uint16(value))
	return value
}

func (b *Bytes) GetUInt16(offset int, littleEndian bool) uint16 {
	_ = checkBytesRange(b, offset, 2)
	return bytesOrder(littleEndian).Uint16(b.Data[offset:])
}

func (b *Bytes) SetUInt16(offset int, value uint16, littleEndian bool) uint16 {
	_ = checkBytesRange(b, offset, 2)
	bytesOrder(littleEndian).PutUint16(b.Data[offset:], value)
	return value
}

func (b *Bytes) GetInt(offset int, littleEndian bool) int {
	_ = checkBytesRange(b, offset, 4)
	return int(int32(bytesOrder(littleEndian).Uint32(b.Data[offset:])))
}

func (b *Bytes) SetInt(offset int, value int, littleEndian bool) int {
	_ = checkBytesRange(b, offset, 4)
	bytesOrder(littleEndian).PutUint32(b.Data[offset:], uint32(int32(value)))
	return value
}

func (b *Bytes) GetUInt(offset int, littleEndian bool) uint {
	_ = checkBytesRange(b, offset, 4)
	return uint(bytesOrder(littleEndian).Uint32(b.Data[offset:]))
}

func (b *Bytes) SetUInt(offset int, value uint, littleEndian bool) uint {
	_ = checkBytesRange(b, offset, 4)
	bytesOrder(littleEndian).PutUint32(b.Data[offset:], uint32(value))
	return value
}

func (b *Bytes) GetInt64(offset int, littleEndian bool) int64 {
	_ = checkBytesRange(b, offset, 8)
	return int64(bytesOrder(littleEndian).Uint64(b.Data[offset:]))
}

func (b *Bytes) SetInt64(offset int, value int64, littleEndian bool) int64 {
	_ = checkBytesRange(b, offset, 8)
	bytesOrder(littleEndian).PutUint64(b.Data[offset:], uint64(value))
	return value
}

func (b *Bytes) GetUInt64(offset int, littleEndian bool) uint64 {
	_ = checkBytesRange(b, offset, 8)
	return bytesOrder(littleEndian).Uint64(b.Data[offset:])
}

func (b *Bytes) SetUInt64(offset int, value uint64, littleEndian bool) uint64 {
	_ = checkBytesRange(b, offset, 8)
	bytesOrder(littleEndian).PutUint64(b.Data[offset:], value)
	return value
}

func (b *Bytes) GetFloat(offset int, littleEndian bool) float32 {
	_ = checkBytesRange(b, offset, 4)
	return math.Float32frombits(bytesOrder(littleEndian).Uint32(b.Data[offset:]))
}

func (b *Bytes) SetFloat(offset int, value float32, littleEndian bool) float32 {
	_ = checkBytesRange(b, offset, 4)
	bytesOrder(littleEndian).PutUint32(b.Data[offset:], math.Float32bits(value))
	return value
}

func (b *Bytes) GetDouble(offset int, littleEndian bool) float64 {
	_ = checkBytesRange(b, offset, 8)
	return math.Float64frombits(bytesOrder(littleEndian).Uint64(b.Data[offset:]))
}

func (b *Bytes) SetDouble(offset int, value float64, littleEndian bool) float64 {
	_ = checkBytesRange(b, offset, 8)
	bytesOrder(littleEndian).PutUint64(b.Data[offset:], math.Float64bits(value))
	return value
}

func bytesOrder(littleEndian bool) binary.ByteOrder {
	if littleEndian {
		return binary.LittleEndian
	}
	return binary.BigEndian
}

func bytesGetInt4(value *Bytes, index int) (int8, error) {
	if index < 0 {
		return 0, fmt.Errorf("bytes offset out of range")
	}
	byteIndex := index / 2
	if err := checkBytesRange(value, byteIndex, 1); err != nil {
		return 0, err
	}
	b := value.Data[byteIndex]
	var nibble byte
	if index%2 == 0 {
		nibble = b >> 4
	} else {
		nibble = b & 0x0f
	}
	if nibble >= 8 {
		return int8(nibble) - 16, nil
	}
	return int8(nibble), nil
}

func bytesSetInt4(value *Bytes, index int, n int8) (int8, error) {
	if index < 0 {
		return 0, fmt.Errorf("bytes offset out of range")
	}
	byteIndex := index / 2
	if err := checkBytesRange(value, byteIndex, 1); err != nil {
		return 0, err
	}
	nibble := byte(n) & 0x0f
	if index%2 == 0 {
		value.Data[byteIndex] = (value.Data[byteIndex] & 0x0f) | (nibble << 4)
	} else {
		value.Data[byteIndex] = (value.Data[byteIndex] & 0xf0) | nibble
	}
	return n, nil
}

func (i *Interpreter) callBufferMethod(value *Buffer, name string, args []ir.Expr, env *Env) (Value, error) {
	if i.file.Stdlib == nil {
		return nil, fmt.Errorf("stdlib is not loaded")
	}
	fn, ok := i.file.Stdlib.ReceiverFunction("buffer", "Buffer", name)
	if !ok {
		return nil, fmt.Errorf("type Buffer has no method %q", name)
	}
	values, err := i.evalArgs(args, env)
	if err != nil {
		return nil, err
	}
	switch fn.Intrinsic {
	case "buffer.length":
		if len(values) != 0 {
			return nil, fmt.Errorf("buffer.%s expects 0 args, got %d", name, len(values))
		}
		return len(value.Data), nil
	case "buffer.clear":
		if len(values) != 0 {
			return nil, fmt.Errorf("buffer.clear expects 0 args, got %d", len(values))
		}
		value.Data = nil
		return nil, nil
	case "buffer.clone":
		if len(values) != 0 {
			return nil, fmt.Errorf("buffer.clone expects 0 args, got %d", len(values))
		}
		return &Buffer{Data: append([]byte(nil), value.Data...)}, nil
	case "buffer.toBytes":
		if len(values) != 0 {
			return nil, fmt.Errorf("buffer.toBytes expects 0 args, got %d", len(values))
		}
		return &Bytes{Data: append([]byte(nil), value.Data...)}, nil
	case "buffer.toInts":
		if len(values) != 0 {
			return nil, fmt.Errorf("buffer.toInts expects 0 args, got %d", len(values))
		}
		return intsFromBytes(value.Data), nil
	case "buffer.append":
		if len(values) != 1 {
			return nil, fmt.Errorf("buffer.append expects 1 arg, got %d", len(values))
		}
		n, ok := values[0].(uint8)
		if !ok {
			return nil, fmt.Errorf("buffer.append expects UInt8")
		}
		value.Data = append(value.Data, n)
		return value, nil
	case "buffer.appendInt":
		if len(values) != 1 {
			return nil, fmt.Errorf("buffer.appendInt expects 1 arg, got %d", len(values))
		}
		n, ok := values[0].(int)
		if !ok {
			return nil, fmt.Errorf("buffer.appendInt expects Int")
		}
		value.Data = append(value.Data, byte(n))
		return value, nil
	case "buffer.appendBytes":
		if len(values) != 1 {
			return nil, fmt.Errorf("buffer.appendBytes expects 1 arg, got %d", len(values))
		}
		bytesValue, ok := values[0].(*Bytes)
		if !ok {
			return nil, fmt.Errorf("buffer.appendBytes expects Bytes")
		}
		value.Data = append(value.Data, bytesValue.Data...)
		return value, nil
	case "buffer.reader":
		if len(values) != 0 {
			return nil, fmt.Errorf("buffer.reader expects 0 args, got %d", len(values))
		}
		return &Reader{Data: append([]byte(nil), value.Data...)}, nil
	case "buffer.writer":
		if len(values) != 0 {
			return nil, fmt.Errorf("buffer.writer expects 0 args, got %d", len(values))
		}
		return &Writer{Data: append([]byte(nil), value.Data...)}, nil
	default:
		if fn.Body == nil {
			return nil, fmt.Errorf("buffer.%s is not supported by the interpreter", name)
		}
		local := NewEnv(env)
		local.Define("this", value)
		for idx, param := range fn.ParamNames {
			if idx < len(values) {
				local.Define(param, values[idx])
			}
		}
		return i.eval(ir.LowerExpr(fn.Body, nil), local)
	}
}

func (i *Interpreter) callReaderMethod(value *Reader, name string, args []ir.Expr, env *Env) (Value, error) {
	if i.file.Stdlib == nil {
		return nil, fmt.Errorf("stdlib is not loaded")
	}
	fn, ok := i.file.Stdlib.ReceiverFunction("reader", "Reader", name)
	if !ok {
		return nil, fmt.Errorf("type Reader has no method %q", name)
	}
	values, err := i.evalArgs(args, env)
	if err != nil {
		return nil, err
	}
	switch fn.Intrinsic {
	case "reader.length":
		return len(value.Data), nil
	case "reader.position":
		return value.Offset, nil
	case "reader.remaining":
		return len(value.Data) - value.Offset, nil
	case "reader.seek":
		if len(values) != 1 {
			return nil, fmt.Errorf("reader.seek expects 1 arg, got %d", len(values))
		}
		offset, ok := values[0].(int)
		if !ok {
			return nil, fmt.Errorf("reader.seek expects Int")
		}
		if offset < 0 || offset > len(value.Data) {
			return nil, fmt.Errorf("reader offset out of range")
		}
		value.Offset = offset
		value.Nibble = 0
		return value.Offset, nil
	case "reader.skip":
		if len(values) != 1 {
			return nil, fmt.Errorf("reader.skip expects 1 arg, got %d", len(values))
		}
		count, ok := values[0].(int)
		if !ok {
			return nil, fmt.Errorf("reader.skip expects Int")
		}
		if count < 0 || value.Offset+count > len(value.Data) {
			return nil, fmt.Errorf("reader offset out of range")
		}
		value.Offset += count
		value.Nibble = 0
		return value.Offset, nil
	case "reader.readBytes":
		if len(values) != 1 {
			return nil, fmt.Errorf("reader.%s expects 1 arg, got %d", name, len(values))
		}
		length, ok := values[0].(int)
		if !ok {
			return nil, fmt.Errorf("reader.%s expects Int", name)
		}
		if err := readerAlign(value); err != nil {
			return nil, err
		}
		if length < 0 || value.Offset+length > len(value.Data) {
			return nil, fmt.Errorf("reader offset out of range")
		}
		out := &Bytes{Data: append([]byte(nil), value.Data[value.Offset:value.Offset+length]...)}
		value.Offset += length
		return out, nil
	case "reader.readInt4":
		if len(values) != 0 {
			return nil, fmt.Errorf("reader.readInt4 expects 0 args, got %d", len(values))
		}
		return readerReadInt4(value)
	}
	if result, ok, err := callReaderNumericMethod(value, name, fn.Intrinsic, values); ok || err != nil {
		return result, err
	}
	if fn.Body == nil {
		return nil, fmt.Errorf("reader.%s is not supported by the interpreter", name)
	}
	local := NewEnv(env)
	local.Define("this", value)
	for idx, param := range fn.ParamNames {
		if idx < len(values) {
			local.Define(param, values[idx])
		}
	}
	return i.eval(ir.LowerExpr(fn.Body, nil), local)
}

func (i *Interpreter) callWriterMethod(value *Writer, name string, args []ir.Expr, env *Env) (Value, error) {
	if i.file.Stdlib == nil {
		return nil, fmt.Errorf("stdlib is not loaded")
	}
	fn, ok := i.file.Stdlib.ReceiverFunction("writer", "Writer", name)
	if !ok {
		return nil, fmt.Errorf("type Writer has no method %q", name)
	}
	values, err := i.evalArgs(args, env)
	if err != nil {
		return nil, err
	}
	switch fn.Intrinsic {
	case "writer.length", "writer.position":
		return len(value.Data), nil
	case "writer.clear":
		if len(values) != 0 {
			return nil, fmt.Errorf("writer.clear expects 0 args, got %d", len(values))
		}
		value.Data = nil
		value.Nibble = 0
		return nil, nil
	case "writer.toBytes":
		if len(values) != 0 {
			return nil, fmt.Errorf("writer.toBytes expects 0 args, got %d", len(values))
		}
		return &Bytes{Data: append([]byte(nil), value.Data...)}, nil
	case "writer.toInts":
		if len(values) != 0 {
			return nil, fmt.Errorf("writer.toInts expects 0 args, got %d", len(values))
		}
		return intsFromBytes(value.Data), nil
	case "writer.writeBytes":
		if len(values) != 1 {
			return nil, fmt.Errorf("writer.%s expects 1 arg, got %d", name, len(values))
		}
		bytesValue, ok := values[0].(*Bytes)
		if !ok {
			return nil, fmt.Errorf("writer.%s expects Bytes", name)
		}
		writerAlign(value)
		value.Data = append(value.Data, bytesValue.Data...)
		return value, nil
	case "writer.writeInt4":
		if len(values) != 1 {
			return nil, fmt.Errorf("writer.writeInt4 expects 1 arg, got %d", len(values))
		}
		n, ok := values[0].(int8)
		if !ok {
			return nil, fmt.Errorf("writer.writeInt4 expects Int4")
		}
		writerWriteInt4(value, n)
		return value, nil
	}
	if ok, err := callWriterNumericMethod(value, name, fn.Intrinsic, values); ok || err != nil {
		return value, err
	}
	if fn.Body == nil {
		return nil, fmt.Errorf("writer.%s is not supported by the interpreter", name)
	}
	local := NewEnv(env)
	local.Define("this", value)
	for idx, param := range fn.ParamNames {
		if idx < len(values) {
			local.Define(param, values[idx])
		}
	}
	return i.eval(ir.LowerExpr(fn.Body, nil), local)
}

func intsFromBytes(bytes []byte) *Array {
	out := &Array{Elements: make([]Value, 0, len(bytes))}
	for _, b := range bytes {
		out.Elements = append(out.Elements, int(b))
	}
	return out
}

func readerAlign(reader *Reader) error {
	if reader.Nibble == 1 {
		reader.Offset++
		reader.Nibble = 0
	}
	if reader.Offset > len(reader.Data) {
		return fmt.Errorf("reader offset out of range")
	}
	return nil
}

func writerAlign(writer *Writer) {
	writer.Nibble = 0
}

func readerBytes(reader *Reader, size int) (*Bytes, error) {
	if err := readerAlign(reader); err != nil {
		return nil, err
	}
	if err := checkBytesRange(&Bytes{Data: reader.Data}, reader.Offset, size); err != nil {
		return nil, fmt.Errorf("reader offset out of range")
	}
	out := &Bytes{Data: reader.Data}
	out = &Bytes{Data: out.Data}
	offset := reader.Offset
	reader.Offset += size
	return &Bytes{Data: out.Data[offset:reader.Offset]}, nil
}

func readerReadInt4(reader *Reader) (int8, error) {
	if reader.Offset < 0 || reader.Offset >= len(reader.Data) {
		return 0, fmt.Errorf("reader offset out of range")
	}
	b := &Bytes{Data: reader.Data}
	n, err := bytesGetInt4(b, reader.Offset*2+reader.Nibble)
	if err != nil {
		return 0, err
	}
	if reader.Nibble == 0 {
		reader.Nibble = 1
	} else {
		reader.Nibble = 0
		reader.Offset++
	}
	return n, nil
}

func writerWriteInt4(writer *Writer, value int8) {
	if writer.Nibble == 0 {
		writer.Data = append(writer.Data, byte(value&0x0f)<<4)
		writer.Nibble = 1
		return
	}
	writer.Data[len(writer.Data)-1] = (writer.Data[len(writer.Data)-1] & 0xf0) | (byte(value) & 0x0f)
	writer.Nibble = 0
}

func callReaderNumericMethod(reader *Reader, name string, intrinsic string, values []Value) (Value, bool, error) {
	sizeByIntrinsic := map[string]int{
		"reader.readInt8":   1,
		"reader.readUInt8":  1,
		"reader.readInt16":  2,
		"reader.readUInt16": 2,
		"reader.readInt":    4,
		"reader.readUInt":   4,
		"reader.readInt64":  8,
		"reader.readUInt64": 8,
		"reader.readFloat":  4,
		"reader.readDouble": 8,
	}
	size, ok := sizeByIntrinsic[intrinsic]
	if !ok {
		return nil, false, nil
	}
	if size == 1 {
		if len(values) != 0 {
			return nil, true, fmt.Errorf("reader.%s expects 0 args, got %d", name, len(values))
		}
	} else if len(values) != 1 {
		return nil, true, fmt.Errorf("reader.%s expects 1 arg, got %d", name, len(values))
	}
	littleEndian := false
	if size > 1 {
		var ok bool
		littleEndian, ok = values[0].(bool)
		if !ok {
			return nil, true, fmt.Errorf("reader.%s littleEndian expects Bool", name)
		}
	}
	chunk, err := readerBytes(reader, size)
	if err != nil {
		return nil, true, err
	}
	switch intrinsic {
	case "reader.readInt8":
		return chunk.GetInt8(0), true, nil
	case "reader.readUInt8":
		return chunk.GetUInt8(0), true, nil
	case "reader.readInt16":
		return chunk.GetInt16(0, littleEndian), true, nil
	case "reader.readUInt16":
		return chunk.GetUInt16(0, littleEndian), true, nil
	case "reader.readInt":
		return chunk.GetInt(0, littleEndian), true, nil
	case "reader.readUInt":
		return chunk.GetUInt(0, littleEndian), true, nil
	case "reader.readInt64":
		return chunk.GetInt64(0, littleEndian), true, nil
	case "reader.readUInt64":
		return chunk.GetUInt64(0, littleEndian), true, nil
	case "reader.readFloat":
		return chunk.GetFloat(0, littleEndian), true, nil
	case "reader.readDouble":
		return chunk.GetDouble(0, littleEndian), true, nil
	default:
		return nil, false, nil
	}
}

func callWriterNumericMethod(writer *Writer, name string, intrinsic string, values []Value) (bool, error) {
	sizeByIntrinsic := map[string]int{
		"writer.writeInt8":   1,
		"writer.writeUInt8":  1,
		"writer.writeInt16":  2,
		"writer.writeUInt16": 2,
		"writer.writeInt":    4,
		"writer.writeUInt":   4,
		"writer.writeInt64":  8,
		"writer.writeUInt64": 8,
		"writer.writeFloat":  4,
		"writer.writeDouble": 8,
	}
	size, ok := sizeByIntrinsic[intrinsic]
	if !ok {
		return false, nil
	}
	if size == 1 {
		if len(values) != 1 {
			return true, fmt.Errorf("writer.%s expects 1 arg, got %d", name, len(values))
		}
	} else if len(values) != 2 {
		return true, fmt.Errorf("writer.%s expects 2 args, got %d", name, len(values))
	}
	littleEndian := false
	if size > 1 {
		var ok bool
		littleEndian, ok = values[1].(bool)
		if !ok {
			return true, fmt.Errorf("writer.%s littleEndian expects Bool", name)
		}
	}
	writerAlign(writer)
	chunk := &Bytes{Data: make([]byte, size)}
	switch intrinsic {
	case "writer.writeInt8":
		n, ok := values[0].(int8)
		if !ok {
			return true, fmt.Errorf("writer.%s expects Int8", name)
		}
		chunk.SetInt8(0, n)
	case "writer.writeUInt8":
		n, ok := values[0].(uint8)
		if !ok {
			return true, fmt.Errorf("writer.%s expects UInt8", name)
		}
		chunk.SetUInt8(0, n)
	case "writer.writeInt16":
		n, ok := values[0].(int16)
		if !ok {
			return true, fmt.Errorf("writer.%s expects Int16", name)
		}
		chunk.SetInt16(0, n, littleEndian)
	case "writer.writeUInt16":
		n, ok := values[0].(uint16)
		if !ok {
			return true, fmt.Errorf("writer.%s expects UInt16", name)
		}
		chunk.SetUInt16(0, n, littleEndian)
	case "writer.writeInt":
		n, ok := values[0].(int)
		if !ok {
			return true, fmt.Errorf("writer.%s expects Int", name)
		}
		chunk.SetInt(0, n, littleEndian)
	case "writer.writeUInt":
		n, ok := values[0].(uint)
		if !ok {
			return true, fmt.Errorf("writer.%s expects UInt", name)
		}
		chunk.SetUInt(0, n, littleEndian)
	case "writer.writeInt64":
		n, ok := values[0].(int64)
		if !ok {
			return true, fmt.Errorf("writer.%s expects Int64", name)
		}
		chunk.SetInt64(0, n, littleEndian)
	case "writer.writeUInt64":
		n, ok := values[0].(uint64)
		if !ok {
			return true, fmt.Errorf("writer.%s expects UInt64", name)
		}
		chunk.SetUInt64(0, n, littleEndian)
	case "writer.writeFloat":
		n, ok := values[0].(float32)
		if !ok {
			return true, fmt.Errorf("writer.%s expects Float", name)
		}
		chunk.SetFloat(0, n, littleEndian)
	case "writer.writeDouble":
		n, ok := values[0].(float64)
		if !ok {
			return true, fmt.Errorf("writer.%s expects Double", name)
		}
		chunk.SetDouble(0, n, littleEndian)
	}
	writer.Data = append(writer.Data, chunk.Data...)
	return true, nil
}

func valuesEqual(left Value, right Value) bool {
	if l, ok := left.(*big.Int); ok {
		r, ok := right.(*big.Int)
		return ok && l.Cmp(r) == 0
	}
	return reflect.DeepEqual(left, right)
}

func (i *Interpreter) callGoBackedFunction(symbol string, args []Value) (Value, error) {
	switch symbol {
	case "fmt.Print":
		for _, arg := range args {
			fmt.Fprint(i.out, printValue(arg))
		}
		return nil, nil
	case "fmt.Println":
		for _, arg := range args {
			fmt.Fprint(i.out, printValue(arg))
		}
		fmt.Fprintln(i.out)
		return nil, nil
	case "fmt.Printf":
		if len(args) == 0 {
			return nil, fmt.Errorf("fmt.Printf expects a format string")
		}
		format, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("fmt.Printf expects a format string")
		}
		fmt.Fprintf(i.out, format, valuesAsAny(args[1:])...)
		return nil, nil
	default:
		return nil, fmt.Errorf("Go-backed function %s is not supported by the interpreter", symbol)
	}
}

func valuesAsAny(values []Value) []any {
	out := make([]any, len(values))
	for idx, value := range values {
		out[idx] = value
	}
	return out
}

func (i *Interpreter) callArrayMethod(array *Array, name string, args []ir.Expr, env *Env) (Value, error) {
	if i.file.Stdlib == nil {
		return nil, fmt.Errorf("stdlib is not loaded")
	}
	fn, ok := i.file.Stdlib.Function("array", name)
	if !ok {
		return nil, fmt.Errorf("type Array has no method %q", name)
	}
	switch {
	case fn.Intrinsic == "array.len":
		return len(array.Elements), nil
	case fn.Intrinsic == "array.get" || fn.Intrinsic == "array.at":
		if len(args) != 1 {
			return nil, fmt.Errorf("array.%s expects 1 args, got %d", name, len(args))
		}
		index, err := i.eval(args[0], env)
		if err != nil {
			return nil, err
		}
		return indexValue(array, index)
	case fn.Intrinsic == "array.push":
		if len(args) != 1 {
			return nil, fmt.Errorf("array.push expects 1 args, got %d", len(args))
		}
		value, err := i.eval(args[0], env)
		if err != nil {
			return nil, err
		}
		array.Elements = append(array.Elements, value)
		return len(array.Elements), nil
	case fn.Intrinsic == "array.set":
		if len(args) != 2 {
			return nil, fmt.Errorf("array.set expects 2 args, got %d", len(args))
		}
		index, err := i.eval(args[0], env)
		if err != nil {
			return nil, err
		}
		value, err := i.eval(args[1], env)
		if err != nil {
			return nil, err
		}
		n, ok := index.(int)
		if !ok {
			return nil, fmt.Errorf("array.set index expects Int")
		}
		if n < 0 || n >= len(array.Elements) {
			return nil, fmt.Errorf("array index %d out of range", n)
		}
		array.Elements[n] = value
		return value, nil
	case fn.Intrinsic == "array.pop":
		if len(args) != 0 {
			return nil, fmt.Errorf("array.pop expects 0 args, got %d", len(args))
		}
		if len(array.Elements) == 0 {
			return nil, fmt.Errorf("array.pop on empty array")
		}
		last := array.Elements[len(array.Elements)-1]
		array.Elements = array.Elements[:len(array.Elements)-1]
		return last, nil
	case fn.Intrinsic == "array.first":
		if len(array.Elements) == 0 {
			return nil, fmt.Errorf("array.first on empty array")
		}
		return array.Elements[0], nil
	case fn.Intrinsic == "array.last":
		if len(array.Elements) == 0 {
			return nil, fmt.Errorf("array.last on empty array")
		}
		return array.Elements[len(array.Elements)-1], nil
	case fn.Intrinsic == "array.slice":
		if len(args) != 2 {
			return nil, fmt.Errorf("array.slice expects 2 args, got %d", len(args))
		}
		start, err := i.eval(args[0], env)
		if err != nil {
			return nil, err
		}
		end, err := i.eval(args[1], env)
		if err != nil {
			return nil, err
		}
		s, ok := start.(int)
		if !ok {
			return nil, fmt.Errorf("array.slice start expects Int")
		}
		e, ok := end.(int)
		if !ok {
			return nil, fmt.Errorf("array.slice end expects Int")
		}
		if s < 0 || e < s || e > len(array.Elements) {
			return nil, fmt.Errorf("array.slice range out of bounds")
		}
		return &Array{Elements: append([]Value{}, array.Elements[s:e]...)}, nil
	case fn.Intrinsic == "array.clone":
		return &Array{Elements: append([]Value{}, array.Elements...)}, nil
	case fn.Intrinsic == "array.reverse":
		out := append([]Value{}, array.Elements...)
		for left, right := 0, len(out)-1; left < right; left, right = left+1, right-1 {
			out[left], out[right] = out[right], out[left]
		}
		return &Array{Elements: out}, nil
	case fn.Intrinsic == "array.contains":
		if len(args) != 1 {
			return nil, fmt.Errorf("array.contains expects 1 args, got %d", len(args))
		}
		value, err := i.eval(args[0], env)
		if err != nil {
			return nil, err
		}
		for _, elem := range array.Elements {
			if valuesEqual(elem, value) {
				return true, nil
			}
		}
		return false, nil
	case fn.Intrinsic == "array.each":
		if len(args) != 1 {
			return nil, fmt.Errorf("array.each expects 1 args, got %d", len(args))
		}
		closure, err := i.evalLambdaArg(args[0], env)
		if err != nil {
			return nil, err
		}
		for idx, elem := range array.Elements {
			if _, err := i.callClosure(closure, arrayCallbackArgs(closure, elem, idx, array)); err != nil {
				return nil, err
			}
		}
		return nil, nil
	case fn.Intrinsic == "array.map":
		if len(args) != 1 {
			return nil, fmt.Errorf("array.map expects 1 args, got %d", len(args))
		}
		closure, err := i.evalLambdaArg(args[0], env)
		if err != nil {
			return nil, err
		}
		result := &Array{Elements: make([]Value, 0, len(array.Elements))}
		for idx, elem := range array.Elements {
			value, err := i.callClosure(closure, arrayCallbackArgs(closure, elem, idx, array))
			if err != nil {
				return nil, err
			}
			result.Elements = append(result.Elements, value)
		}
		return result, nil
	case fn.Intrinsic == "array.reduce" || fn.Intrinsic == "array.foldl":
		return i.callArrayReduce(array, name, args, env, false)
	case fn.Intrinsic == "array.foldr":
		return i.callArrayReduce(array, name, args, env, true)
	default:
		if fn.Body == nil {
			return nil, fmt.Errorf("array.%s is not supported by the interpreter", name)
		}
		values, err := i.evalArgs(args, env)
		if err != nil {
			return nil, err
		}
		if len(values) != len(fn.ParamNames) {
			return nil, fmt.Errorf("array.%s expects %d args, got %d", name, len(fn.ParamNames), len(values))
		}
		local := NewEnv(env)
		local.Define("this", array)
		for idx, param := range fn.ParamNames {
			local.Define(param, values[idx])
		}
		return i.eval(ir.LowerExpr(fn.Body, nil), local)
	}
}

func (i *Interpreter) callArrayReduce(array *Array, name string, args []ir.Expr, env *Env, reverse bool) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("array.%s expects 2 args, got %d", name, len(args))
	}
	accumulator, err := i.eval(args[0], env)
	if err != nil {
		return nil, err
	}
	closure, err := i.evalLambdaArg(args[1], env)
	if err != nil {
		return nil, err
	}
	if reverse {
		for idx := len(array.Elements) - 1; idx >= 0; idx-- {
			accumulator, err = i.callClosure(closure, arrayReducerArgs(closure, accumulator, array.Elements[idx], idx, array))
			if err != nil {
				return nil, err
			}
		}
		return accumulator, nil
	}
	for idx, elem := range array.Elements {
		accumulator, err = i.callClosure(closure, arrayReducerArgs(closure, accumulator, elem, idx, array))
		if err != nil {
			return nil, err
		}
	}
	return accumulator, nil
}

func arrayCallbackArgs(closure *Closure, elem Value, idx int, array *Array) []Value {
	values := []Value{elem}
	if len(closure.Params) >= 2 {
		values = append(values, idx)
	}
	if len(closure.Params) >= 3 {
		values = append(values, array)
	}
	return values
}

func arrayReducerArgs(closure *Closure, accumulator Value, elem Value, idx int, array *Array) []Value {
	values := []Value{accumulator}
	if len(closure.Params) >= 2 {
		values = append(values, elem)
	}
	if len(closure.Params) >= 3 {
		values = append(values, idx)
	}
	if len(closure.Params) >= 4 {
		values = append(values, array)
	}
	return values
}

func (i *Interpreter) callStringMethod(value string, name string, args []ir.Expr, env *Env) (Value, error) {
	if i.file.Stdlib == nil {
		return nil, fmt.Errorf("stdlib is not loaded")
	}
	fn, ok := i.file.Stdlib.ReceiverFunction("string", "String", name)
	if !ok {
		return nil, fmt.Errorf("type String has no method %q", name)
	}
	values, err := i.evalArgs(args, env)
	if err != nil {
		return nil, err
	}
	stringArg := func(index int) (string, error) {
		if index >= len(values) {
			return "", fmt.Errorf("string.%s expects more args", name)
		}
		arg, ok := values[index].(string)
		if !ok {
			return "", fmt.Errorf("string.%s argument %d expects String", name, index+1)
		}
		return arg, nil
	}
	switch fn.Intrinsic {
	case "string.length":
		return len([]rune(value)), nil
	case "string.toString":
		return value, nil
	case "string.at":
		if len(values) != 1 {
			return nil, fmt.Errorf("string.%s expects 1 arg, got %d", name, len(values))
		}
		index, ok := values[0].(int)
		if !ok {
			return nil, fmt.Errorf("string.%s expects Int", name)
		}
		runes := []rune(value)
		if index < 0 || index >= len(runes) {
			return nil, fmt.Errorf("string index %d out of range", index)
		}
		return Char(runes[index]), nil
	case "string.slice":
		if len(values) != 2 {
			return nil, fmt.Errorf("string.slice expects 2 args, got %d", len(values))
		}
		start, ok := values[0].(int)
		if !ok {
			return nil, fmt.Errorf("string.slice start expects Int")
		}
		end, ok := values[1].(int)
		if !ok {
			return nil, fmt.Errorf("string.slice end expects Int")
		}
		runes := []rune(value)
		if start < 0 || end < start || end > len(runes) {
			return nil, fmt.Errorf("string.slice range out of bounds")
		}
		return string(runes[start:end]), nil
	case "string.concat":
		arg, err := stringArg(0)
		if err != nil {
			return nil, err
		}
		return value + arg, nil
	case "string.includes":
		arg, err := stringArg(0)
		if err != nil {
			return nil, err
		}
		return strings.Contains(value, arg), nil
	case "string.startsWith":
		arg, err := stringArg(0)
		if err != nil {
			return nil, err
		}
		return strings.HasPrefix(value, arg), nil
	case "string.endsWith":
		arg, err := stringArg(0)
		if err != nil {
			return nil, err
		}
		return strings.HasSuffix(value, arg), nil
	case "string.indexOf":
		arg, err := stringArg(0)
		if err != nil {
			return nil, err
		}
		return strings.Index(value, arg), nil
	case "string.lastIndexOf":
		arg, err := stringArg(0)
		if err != nil {
			return nil, err
		}
		return strings.LastIndex(value, arg), nil
	case "string.toLowerCase":
		return strings.ToLower(value), nil
	case "string.toUpperCase":
		return strings.ToUpper(value), nil
	case "string.trim":
		return strings.TrimSpace(value), nil
	case "string.trimStart":
		return strings.TrimLeftFunc(value, func(r rune) bool { return strings.TrimSpace(string(r)) == "" }), nil
	case "string.trimEnd":
		return strings.TrimRightFunc(value, func(r rune) bool { return strings.TrimSpace(string(r)) == "" }), nil
	case "string.repeat":
		if len(values) != 1 {
			return nil, fmt.Errorf("string.repeat expects 1 arg, got %d", len(values))
		}
		count, ok := values[0].(int)
		if !ok {
			return nil, fmt.Errorf("string.repeat expects Int")
		}
		return strings.Repeat(value, count), nil
	case "string.replace":
		search, err := stringArg(0)
		if err != nil {
			return nil, err
		}
		replacement, err := stringArg(1)
		if err != nil {
			return nil, err
		}
		return strings.Replace(value, search, replacement, 1), nil
	case "string.replaceAll":
		search, err := stringArg(0)
		if err != nil {
			return nil, err
		}
		replacement, err := stringArg(1)
		if err != nil {
			return nil, err
		}
		return strings.ReplaceAll(value, search, replacement), nil
	case "string.split":
		separator, err := stringArg(0)
		if err != nil {
			return nil, err
		}
		parts := strings.Split(value, separator)
		out := &Array{Elements: make([]Value, 0, len(parts))}
		for _, part := range parts {
			out.Elements = append(out.Elements, part)
		}
		return out, nil
	default:
		if fn.Body == nil {
			return nil, fmt.Errorf("string.%s is not supported by the interpreter", name)
		}
		local := NewEnv(env)
		local.Define("this", value)
		for idx, param := range fn.ParamNames {
			if idx < len(values) {
				local.Define(param, values[idx])
			}
		}
		return i.eval(ir.LowerExpr(fn.Body, nil), local)
	}
}

func (i *Interpreter) callCharMethod(value Char, name string, args []ir.Expr, env *Env) (Value, error) {
	if i.file.Stdlib == nil {
		return nil, fmt.Errorf("stdlib is not loaded")
	}
	if _, ok := i.file.Stdlib.ReceiverFunction("char", "Char", name); !ok {
		return nil, fmt.Errorf("type Char has no method %q", name)
	}
	values, err := i.evalArgs(args, env)
	if err != nil {
		return nil, err
	}
	switch name {
	case "toString":
		if len(values) != 0 {
			return nil, fmt.Errorf("char.toString expects 0 args, got %d", len(values))
		}
		return string(rune(value)), nil
	default:
		return nil, fmt.Errorf("char.%s is not supported by the interpreter", name)
	}
}

func (i *Interpreter) callBoolMethod(value bool, name string, args []ir.Expr, env *Env) (Value, error) {
	if i.file.Stdlib == nil {
		return nil, fmt.Errorf("stdlib is not loaded")
	}
	if _, ok := i.file.Stdlib.ReceiverFunction("bool", "Bool", name); !ok {
		return nil, fmt.Errorf("type Bool has no method %q", name)
	}
	values, err := i.evalArgs(args, env)
	if err != nil {
		return nil, err
	}
	switch name {
	case "not":
		if len(values) != 0 {
			return nil, fmt.Errorf("bool.not expects 0 args, got %d", len(values))
		}
		return !value, nil
	case "xor":
		if len(values) != 1 {
			return nil, fmt.Errorf("bool.xor expects 1 arg, got %d", len(values))
		}
		other, ok := values[0].(bool)
		if !ok {
			return nil, fmt.Errorf("bool.xor expects Bool")
		}
		return value != other, nil
	case "toString":
		if len(values) != 0 {
			return nil, fmt.Errorf("bool.toString expects 0 args, got %d", len(values))
		}
		if value {
			return "true", nil
		}
		return "false", nil
	default:
		return nil, fmt.Errorf("bool.%s is not supported by the interpreter", name)
	}
}

func (i *Interpreter) callMapMethod(value *Map, name string, args []ir.Expr, env *Env) (Value, error) {
	if i.file.Stdlib == nil {
		return nil, fmt.Errorf("stdlib is not loaded")
	}
	fn, ok := i.file.Stdlib.ReceiverFunction("map", "Map", name)
	if !ok {
		return nil, fmt.Errorf("type Map has no method %q", name)
	}
	values, err := i.evalArgs(args, env)
	if err != nil {
		return nil, err
	}
	switch fn.Intrinsic {
	case "map.size":
		return len(value.Entries), nil
	case "map.has":
		if len(values) != 1 {
			return nil, fmt.Errorf("map.has expects 1 arg, got %d", len(values))
		}
		_, ok := value.Entries[valueKey(values[0])]
		return ok, nil
	case "map.getOr":
		if len(values) != 2 {
			return nil, fmt.Errorf("map.getOr expects 2 args, got %d", len(values))
		}
		if entry, ok := value.Entries[valueKey(values[0])]; ok {
			return entry.Value, nil
		}
		return values[1], nil
	case "map.set":
		if len(values) != 2 {
			return nil, fmt.Errorf("map.set expects 2 args, got %d", len(values))
		}
		value.Entries[valueKey(values[0])] = mapEntry{Key: values[0], Value: values[1]}
		return value, nil
	case "map.delete":
		if len(values) != 1 {
			return nil, fmt.Errorf("map.delete expects 1 arg, got %d", len(values))
		}
		key := valueKey(values[0])
		_, existed := value.Entries[key]
		delete(value.Entries, key)
		return existed, nil
	case "map.clear":
		for key := range value.Entries {
			delete(value.Entries, key)
		}
		return nil, nil
	case "map.keys":
		out := &Array{Elements: make([]Value, 0, len(value.Entries))}
		for _, entry := range value.Entries {
			out.Elements = append(out.Elements, entry.Key)
		}
		return out, nil
	case "map.values":
		out := &Array{Elements: make([]Value, 0, len(value.Entries))}
		for _, entry := range value.Entries {
			out.Elements = append(out.Elements, entry.Value)
		}
		return out, nil
	case "map.each":
		if len(args) != 1 {
			return nil, fmt.Errorf("map.each expects 1 arg, got %d", len(args))
		}
		closure, err := i.evalLambdaArg(args[0], env)
		if err != nil {
			return nil, err
		}
		for _, entry := range value.Entries {
			callbackArgs := []Value{entry.Value}
			if len(closure.Params) >= 2 {
				callbackArgs = append(callbackArgs, entry.Key)
			}
			if len(closure.Params) >= 3 {
				callbackArgs = append(callbackArgs, value)
			}
			if _, err := i.callClosure(closure, callbackArgs); err != nil {
				return nil, err
			}
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("map.%s is not supported by the interpreter", name)
	}
}

func (i *Interpreter) callSetMethod(value *Set, name string, args []ir.Expr, env *Env) (Value, error) {
	if i.file.Stdlib == nil {
		return nil, fmt.Errorf("stdlib is not loaded")
	}
	fn, ok := i.file.Stdlib.ReceiverFunction("set", "Set", name)
	if !ok {
		return nil, fmt.Errorf("type Set has no method %q", name)
	}
	values, err := i.evalArgs(args, env)
	if err != nil {
		return nil, err
	}
	switch fn.Intrinsic {
	case "set.size":
		return len(value.Entries), nil
	case "set.has":
		if len(values) != 1 {
			return nil, fmt.Errorf("set.has expects 1 arg, got %d", len(values))
		}
		_, ok := value.Entries[valueKey(values[0])]
		return ok, nil
	case "set.add":
		if len(values) != 1 {
			return nil, fmt.Errorf("set.add expects 1 arg, got %d", len(values))
		}
		value.Entries[valueKey(values[0])] = values[0]
		return value, nil
	case "set.delete":
		if len(values) != 1 {
			return nil, fmt.Errorf("set.delete expects 1 arg, got %d", len(values))
		}
		key := valueKey(values[0])
		_, existed := value.Entries[key]
		delete(value.Entries, key)
		return existed, nil
	case "set.clear":
		for key := range value.Entries {
			delete(value.Entries, key)
		}
		return nil, nil
	case "set.values":
		out := &Array{Elements: make([]Value, 0, len(value.Entries))}
		for _, entry := range value.Entries {
			out.Elements = append(out.Elements, entry)
		}
		return out, nil
	case "set.each":
		if len(args) != 1 {
			return nil, fmt.Errorf("set.each expects 1 arg, got %d", len(args))
		}
		closure, err := i.evalLambdaArg(args[0], env)
		if err != nil {
			return nil, err
		}
		for _, entry := range value.Entries {
			callbackArgs := []Value{entry}
			if len(closure.Params) >= 2 {
				callbackArgs = append(callbackArgs, entry)
			}
			if len(closure.Params) >= 3 {
				callbackArgs = append(callbackArgs, value)
			}
			if _, err := i.callClosure(closure, callbackArgs); err != nil {
				return nil, err
			}
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("set.%s is not supported by the interpreter", name)
	}
}

func valueKey(value Value) string {
	return typeName(value) + ":" + Format(value)
}

func (i *Interpreter) evalLambdaArg(expr ir.Expr, env *Env) (*Closure, error) {
	value, err := i.eval(expr, env)
	if err != nil {
		return nil, err
	}
	closure, ok := value.(*Closure)
	if !ok {
		return nil, fmt.Errorf("expected lambda")
	}
	return closure, nil
}
