package interpreter

import (
	"fmt"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
)

func (i *Interpreter) evalCall(call *ir.CallExpr, env *Env) (Value, error) {
	if ident, ok := call.Callee.(*ir.Identifier); ok {
		if fn := i.functions[ident.Name]; fn != nil {
			return i.callFunction(fn, call.Args, env)
		}
	}
	if sel, ok := call.Callee.(*ir.SelectorExpr); ok {
		if at, ok := sel.Receiver.(*ir.AtExpr); ok {
			return i.callModuleFunction(at.Name, sel.Name, call.Args, env)
		}
		receiver, err := i.eval(sel.Receiver, env)
		if err != nil {
			return nil, err
		}
		switch value := receiver.(type) {
		case *Array:
			return i.callArrayMethod(value, sel.Name, call.Args, env)
		case *Map:
			return i.callMapMethod(value, sel.Name, call.Args, env)
		case *Set:
			return i.callSetMethod(value, sel.Name, call.Args, env)
		case string:
			return i.callStringMethod(value, sel.Name, call.Args, env)
		case bool:
			return i.callBoolMethod(value, sel.Name, call.Args, env)
		case *Struct:
			if field, ok := value.Fields[sel.Name]; ok {
				return i.callCallable(field, call.Args, env)
			}
			return i.callMethod(value, sel.Name, call.Args, env)
		default:
			return nil, fmt.Errorf("type %s has no method %q", typeName(value), sel.Name)
		}
	}
	callee, err := i.eval(call.Callee, env)
	if err != nil {
		return nil, err
	}
	switch fn := callee.(type) {
	case *ir.Function:
		args, err := i.evalArgs(call.Args, env)
		if err != nil {
			return nil, err
		}
		return i.callFunctionValue(fn, args)
	default:
		return i.callCallable(fn, call.Args, env)
	}
}

func (i *Interpreter) callCallable(callee Value, args []ir.Expr, env *Env) (Value, error) {
	values, err := i.evalArgs(args, env)
	if err != nil {
		return nil, err
	}
	switch fn := callee.(type) {
	case *Closure:
		return i.callClosure(fn, values)
	case *ir.Function:
		return i.callFunctionValue(fn, values)
	default:
		return nil, fmt.Errorf("%s is not callable", typeName(callee))
	}
}

func (i *Interpreter) callFunction(fn *ir.Function, args []ir.Expr, env *Env) (Value, error) {
	values, err := i.evalArgs(args, env)
	if err != nil {
		return nil, err
	}
	return i.callFunctionValue(fn, values)
}

func (i *Interpreter) callFunctionValue(fn *ir.Function, args []Value) (Value, error) {
	if len(args) != len(fn.Params) {
		return nil, fmt.Errorf("function %q expects %d args, got %d", fn.Name, len(fn.Params), len(args))
	}
	env := NewEnv(i.globals)
	for idx, param := range fn.Params {
		env.Define(param.Name, args[idx])
	}
	if block, ok := fn.Body.(*ir.PatternBlock); ok {
		if len(args) == 0 {
			return nil, fmt.Errorf("pattern function %q expects a subject", fn.Name)
		}
		return i.evalPatternBlock(block, args[0], env)
	}
	return i.evalFunctionBody(fn.Body, fn.Return, env)
}

func (i *Interpreter) callMethod(receiver *Struct, name string, args []ir.Expr, env *Env) (Value, error) {
	typ := i.types[receiver.TypeName]
	if typ == nil {
		return nil, fmt.Errorf("unknown type %q", receiver.TypeName)
	}
	var method *ir.Function
	for _, candidate := range typ.Methods {
		if candidate.Name == name {
			method = candidate
			break
		}
	}
	if method == nil {
		return nil, fmt.Errorf("type %s has no method %q", receiver.TypeName, name)
	}
	values, err := i.evalArgs(args, env)
	if err != nil {
		return nil, err
	}
	if len(values) != len(method.Params) {
		return nil, fmt.Errorf("method %s.%s expects %d args, got %d", receiver.TypeName, name, len(method.Params), len(values))
	}
	local := NewEnv(i.globals)
	local.Define("this", receiver)
	for idx, param := range method.Params {
		local.Define(param.Name, values[idx])
	}
	return i.evalFunctionBody(method.Body, method.Return, local)
}

func (i *Interpreter) callClosure(fn *Closure, args []Value) (Value, error) {
	if len(args) != len(fn.Params) {
		return nil, fmt.Errorf("lambda expects %d args, got %d", len(fn.Params), len(args))
	}
	env := NewEnv(fn.Env)
	for idx, param := range fn.Params {
		env.Define(param, args[idx])
	}
	return i.eval(fn.Body, env)
}

func (i *Interpreter) evalFunctionBody(body ir.Expr, ret checker.Type, env *Env) (Value, error) {
	value, err := i.eval(body, env)
	if err != nil {
		return nil, err
	}
	if ret == checker.Void {
		return nil, nil
	}
	return value, nil
}

func (i *Interpreter) evalArgs(args []ir.Expr, env *Env) ([]Value, error) {
	values := make([]Value, 0, len(args))
	for _, arg := range args {
		value, err := i.eval(arg, env)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}
