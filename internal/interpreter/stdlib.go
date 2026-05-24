package interpreter

import (
	"fmt"
	"math/big"
	"reflect"
	"regexp"
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
	if fn.Go != nil {
		return i.callGoBackedFunction(fn.Go.Symbol, values)
	}
	switch fn.Intrinsic {
	case "map.new":
		if len(values) != 2 {
			return nil, fmt.Errorf("@map.newMap expects 2 args, got %d", len(values))
		}
		return &Map{Entries: map[string]mapEntry{}}, nil
	case "set.new":
		if len(values) != 1 {
			return nil, fmt.Errorf("@map.newSet expects 1 arg, got %d", len(values))
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
	case fn.Intrinsic == "array.each" || fn.Intrinsic == "array.forEach":
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
		return string(runes[index]), nil
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
	case "map.forEach":
		if len(args) != 1 {
			return nil, fmt.Errorf("map.forEach expects 1 arg, got %d", len(args))
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
	fn, ok := i.file.Stdlib.ReceiverFunction("map", "Set", name)
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
	case "set.forEach":
		if len(args) != 1 {
			return nil, fmt.Errorf("set.forEach expects 1 arg, got %d", len(args))
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
