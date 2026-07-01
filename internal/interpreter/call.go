package interpreter

import (
	"fmt"
	"strings"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
)

func (i *Interpreter) evalCall(call *ir.CallExpr, env *Env) (Value, error) {
	if ident, ok := call.Callee.(*ir.Identifier); ok {
		if value, ok, err := i.callEnumConstructor(ident.Name, call.Args, env); ok || err != nil {
			return value, err
		}
		if fn := i.functions[ident.Name]; fn != nil {
			return i.callFunction(fn, call.Args, env)
		}
	}
	if sel, ok := call.Callee.(*ir.SelectorExpr); ok {
		if at, ok := sel.Receiver.(*ir.AtExpr); ok {
			return i.callAtSelector(at, sel, call.Args, call.ResultType(), env)
		}
		if value, ok, err := i.callQualifiedEnumConstructor(sel, call.Args, env); ok || err != nil {
			return value, err
		}
		if sel.Static {
			ident, ok := sel.Receiver.(*ir.Identifier)
			if !ok {
				return nil, fmt.Errorf("static selector receiver must be a type")
			}
			if _, shadowed := env.Get(ident.Name); !shadowed {
				if typ := i.types[ident.Name]; typ != nil {
					return i.callStaticMethod(typ, sel.Name, call.Args, env)
				}
				if enum := i.enums[ident.Name]; enum != nil {
					return i.callEnumStaticMethod(enum, sel.Name, call.Args, env)
				}
			}
			return nil, fmt.Errorf("unknown type %q", ident.Name)
		}
		receiver, err := i.eval(sel.Receiver, env)
		if err != nil {
			return nil, err
		}
		switch value := receiver.(type) {
		case *ir.AtExpr:
			return i.callAtSelector(value, sel, call.Args, call.ResultType(), env)
		case *Array:
			return i.callArrayMethod(value, sel.Name, call.Args, env)
		case *Map:
			return i.callMapMethod(value, sel.Name, call.Args, env)
		case *Set:
			return i.callSetMethod(value, sel.Name, call.Args, env)
		case *Bytes:
			return i.callBytesMethod(value, sel.Name, call.Args, env)
		case *Buffer:
			return i.callBufferMethod(value, sel.Name, call.Args, env)
		case *Reader:
			return i.callReaderMethod(value, sel.Name, call.Args, env)
		case *Writer:
			return i.callWriterMethod(value, sel.Name, call.Args, env)
		case *StringBuffer:
			return i.callStringBufferMethod(value, sel.Name, call.Args, env)
		case *Regex:
			return i.callRegexMethod(value, sel.Name, call.Args, env)
		case string:
			return i.callStringMethod(value, sel.Name, call.Args, env)
		case Char:
			return i.callCharMethod(value, sel.Name, call.Args, env)
		case bool:
			return i.callBoolMethod(value, sel.Name, call.Args, env)
		case *Struct:
			if field, ok := value.Fields[sel.Name]; ok {
				return i.callCallable(field, call.Args, env)
			}
			return i.callMethod(value, sel.Name, call.Args, env)
		case EnumValue:
			return i.callEnumMethod(value, sel.Name, call.Args, env)
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

func (i *Interpreter) callAtSelector(at *ir.AtExpr, sel *ir.SelectorExpr, args []ir.Expr, resultType checker.Type, env *Env) (Value, error) {
	if at.Name != "" {
		return i.callModuleFunction(at.Name, sel.Name, args, resultType, env)
	}
	name := selectorResolvedName(sel)
	if fn := i.functions[name]; fn != nil {
		return i.callFunction(fn, args, env)
	}
	return nil, fmt.Errorf("import has no function %q", sel.Name)
}

func selectorResolvedName(sel *ir.SelectorExpr) string {
	if sel.ResolvedName != "" {
		return sel.ResolvedName
	}
	return sel.Name
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
	case stdlibFunctionValue:
		return i.callModuleFunctionValue(fn.Module, fn.Name, values, fn.ResultType)
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

func (i *Interpreter) callQualifiedEnumConstructor(sel *ir.SelectorExpr, args []ir.Expr, env *Env) (Value, bool, error) {
	ident, ok := sel.Receiver.(*ir.Identifier)
	if !ok {
		return nil, false, nil
	}
	enum := i.enums[ident.Name]
	if enum == nil {
		return nil, false, nil
	}
	for idx := range enum.Members {
		member := &enum.Members[idx]
		if member.Name != sel.Name || member.HasValue {
			continue
		}
		if len(args) != len(member.Params) {
			return nil, true, fmt.Errorf("constructor %q expects %d args, got %d", sel.Name, len(member.Params), len(args))
		}
		values, err := i.evalArgs(args, env)
		if err != nil {
			return nil, true, err
		}
		return EnumValue{TypeName: enum.Name, Name: member.Name, Value: member.Value, Payload: values}, true, nil
	}
	return nil, false, nil
}

func (i *Interpreter) callEnumConstructor(name string, args []ir.Expr, env *Env) (Value, bool, error) {
	var enumType *ir.EnumType
	var enumMember *ir.EnumMember
	for _, enum := range i.enums {
		for idx := range enum.Members {
			member := &enum.Members[idx]
			if member.Name != name || member.HasValue {
				continue
			}
			if enumMember != nil {
				return nil, true, fmt.Errorf("ambiguous enum constructor %q", name)
			}
			enumType = enum
			enumMember = member
		}
	}
	if enumMember == nil {
		if value, ok, err := i.callSyntaxExprConstructor(name, args, env); ok || err != nil {
			return value, ok, err
		}
		return i.callStdlibConstructor(name, args, env)
	}
	if len(args) != len(enumMember.Params) {
		return nil, true, fmt.Errorf("constructor %q expects %d args, got %d", name, len(enumMember.Params), len(args))
	}
	values, err := i.evalArgs(args, env)
	if err != nil {
		return nil, true, err
	}
	return EnumValue{TypeName: enumType.Name, Name: enumMember.Name, Value: enumMember.Value, Payload: values}, true, nil
}

func (i *Interpreter) callSyntaxExprConstructor(name string, args []ir.Expr, env *Env) (Value, bool, error) {
	paramCount, ok := syntaxExprConstructorParamCount(name)
	if !ok {
		return nil, false, nil
	}
	if len(args) != paramCount {
		return nil, true, fmt.Errorf("constructor %q expects %d args, got %d", name, paramCount, len(args))
	}
	values, err := i.evalArgs(args, env)
	if err != nil {
		return nil, true, err
	}
	return EnumValue{TypeName: "SyntaxExpr", Name: name, Payload: values}, true, nil
}

func syntaxExprConstructorParamCount(name string) (int, bool) {
	switch name {
	case "IdentifierExpr", "ModuleExpr", "StringExpr", "BoolExpr":
		return 1, true
	case "NullExpr":
		return 0, true
	case "SelectorExpr", "StaticSelectorExpr", "CallExpr", "StructExpr":
		return 2, true
	case "BlockExpr":
		return 1, true
	default:
		return 0, false
	}
}

func (i *Interpreter) callStdlibConstructor(name string, args []ir.Expr, env *Env) (Value, bool, error) {
	if i.file == nil || i.file.Stdlib == nil {
		return nil, false, nil
	}
	typeName := ""
	paramCount := 0
	found := false
	for _, typ := range i.file.Stdlib.Types {
		for _, constructor := range typ.Constructors {
			if constructor.Name != name {
				continue
			}
			if found {
				return nil, true, fmt.Errorf("ambiguous constructor %q", name)
			}
			typeName = typ.Name
			paramCount = len(constructor.Params)
			found = true
		}
	}
	if !found {
		return nil, false, nil
	}
	if len(args) != paramCount {
		return nil, true, fmt.Errorf("constructor %q expects %d args, got %d", name, paramCount, len(args))
	}
	values, err := i.evalArgs(args, env)
	if err != nil {
		return nil, true, err
	}
	return EnumValue{TypeName: typeName, Name: name, Payload: values}, true, nil
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
		if value, ok, err := i.callStdlibStructMethod(receiver, name, args, env); ok || err != nil {
			return value, err
		}
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

func (i *Interpreter) callStructMethodValues(receiver *Struct, name string, values []Value) (Value, error) {
	typ := i.types[receiver.TypeName]
	if typ == nil {
		return nil, fmt.Errorf("unknown type %q", receiver.TypeName)
	}
	var method *ir.Function
	for _, candidate := range typ.Methods {
		if candidate.Name == name && !candidate.Static {
			method = candidate
			break
		}
	}
	if method == nil {
		return nil, fmt.Errorf("type %s has no method %q", receiver.TypeName, name)
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

func (i *Interpreter) callEnumMethod(receiver EnumValue, name string, args []ir.Expr, env *Env) (Value, error) {
	enum := i.enums[receiver.TypeName]
	if enum == nil {
		return nil, fmt.Errorf("unknown type %q", receiver.TypeName)
	}
	var method *ir.Function
	for _, candidate := range enum.Methods {
		if candidate.Name == name && !candidate.Static {
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

func (i *Interpreter) callStaticMethod(typ *ir.StructType, name string, args []ir.Expr, env *Env) (Value, error) {
	var method *ir.Function
	for _, candidate := range typ.Methods {
		if candidate.Name == name && candidate.Static {
			method = candidate
			break
		}
	}
	if method == nil {
		return nil, fmt.Errorf("type %s has no static method %q", typ.Name, name)
	}
	values, err := i.evalArgs(args, env)
	if err != nil {
		return nil, err
	}
	if len(values) != len(method.Params) {
		return nil, fmt.Errorf("static method %s::%s expects %d args, got %d", typ.Name, name, len(method.Params), len(values))
	}
	local := NewEnv(i.globals)
	for idx, param := range method.Params {
		local.Define(param.Name, values[idx])
	}
	return i.evalFunctionBody(method.Body, method.Return, local)
}

func (i *Interpreter) callEnumStaticMethod(enum *ir.EnumType, name string, args []ir.Expr, env *Env) (Value, error) {
	var method *ir.Function
	for _, candidate := range enum.Methods {
		if candidate.Name == name && candidate.Static {
			method = candidate
			break
		}
	}
	if method == nil {
		return nil, fmt.Errorf("type %s has no static method %q", enum.Name, name)
	}
	values, err := i.evalArgs(args, env)
	if err != nil {
		return nil, err
	}
	if len(values) != len(method.Params) {
		return nil, fmt.Errorf("static method %s::%s expects %d args, got %d", enum.Name, name, len(method.Params), len(values))
	}
	local := NewEnv(i.globals)
	for idx, param := range method.Params {
		local.Define(param.Name, values[idx])
	}
	return i.evalFunctionBody(method.Body, method.Return, local)
}

func (i *Interpreter) callStdlibStructMethod(receiver *Struct, name string, args []ir.Expr, env *Env) (Value, bool, error) {
	if i.file.Stdlib == nil {
		return nil, false, nil
	}
	fn, ok := i.file.Stdlib.ReceiverFunction("iter", structBaseTypeName(receiver.TypeName), name)
	if !ok || fn.Body == nil {
		return nil, false, nil
	}
	values, err := i.evalArgs(args, env)
	if err != nil {
		return nil, true, err
	}
	if len(values) != len(fn.ParamNames) {
		return nil, true, fmt.Errorf("%s.%s expects %d args, got %d", receiver.TypeName, name, len(fn.ParamNames), len(values))
	}
	local := NewEnv(i.globals)
	local.Define("this", receiver)
	for idx, param := range fn.ParamNames {
		local.Define(param, values[idx])
	}
	value, err := i.eval(ir.LowerExprExpected(fn.Body, nil, checker.Type(fn.Return)), local)
	return value, true, err
}

func structBaseTypeName(name string) string {
	if idx := strings.Index(name, "["); idx >= 0 {
		return name[:idx]
	}
	return name
}

func (i *Interpreter) callClosure(fn *Closure, args []Value) (Value, error) {
	if len(args) < len(fn.Params) {
		return nil, fmt.Errorf("lambda expects %d args, got %d", len(fn.Params), len(args))
	}
	if len(args) > len(fn.Params) {
		args = args[:len(fn.Params)]
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
