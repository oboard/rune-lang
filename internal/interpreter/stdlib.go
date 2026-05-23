package interpreter

import (
	"fmt"
	"math/big"
	"reflect"
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
	case "assert.eq":
		if len(values) != 2 {
			return nil, fmt.Errorf("@assert.eq expects 2 args, got %d", len(values))
		}
		if valuesEqual(values[0], values[1]) {
			return nil, nil
		}
		return nil, fmt.Errorf("assert.eq failed: actual %s, expected %s", Format(values[0]), Format(values[1]))
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
	case fn.Intrinsic == "array.each" || fn.Intrinsic == "array.forEach":
		if len(args) != 1 {
			return nil, fmt.Errorf("array.each expects 1 args, got %d", len(args))
		}
		closure, err := i.evalLambdaArg(args[0], env)
		if err != nil {
			return nil, err
		}
		for _, elem := range array.Elements {
			if _, err := i.callClosure(closure, []Value{elem}); err != nil {
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
		for _, elem := range array.Elements {
			value, err := i.callClosure(closure, []Value{elem})
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
