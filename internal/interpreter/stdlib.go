package interpreter

import (
	"fmt"

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
	case "go.stmt", "go.expr", "go.import":
		return nil, fmt.Errorf("@%s.%s is only supported by the Go backend", module, name)
	default:
		return nil, fmt.Errorf("@%s.%s is not supported by the interpreter", module, name)
	}
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
	case fn.Intrinsic == "array.get":
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
	case fn.Intrinsic == "array.each":
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
