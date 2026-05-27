package interpreter

import (
	"fmt"
	"io"
	"os"

	"github.com/oboard/rune-lang/internal/ir"
)

type Interpreter struct {
	file      *ir.File
	functions map[string]*ir.Function
	types     map[string]*ir.StructType
	enums     map[string]*ir.EnumType
	globals   *Env
	out       io.Writer
}

func New(file *ir.File, opts ...Option) *Interpreter {
	i := &Interpreter{
		functions: map[string]*ir.Function{},
		types:     map[string]*ir.StructType{},
		enums:     map[string]*ir.EnumType{},
		globals:   NewEnv(nil),
		out:       os.Stdout,
	}
	for _, opt := range opts {
		opt(i)
	}
	i.Load(file)
	return i
}

type Option func(*Interpreter)

func WithOutput(out io.Writer) Option {
	return func(i *Interpreter) {
		i.out = out
	}
}

func (i *Interpreter) Load(file *ir.File) {
	i.file = file
	for _, typ := range file.Types {
		i.types[typ.Name] = typ
	}
	for _, enum := range file.Enums {
		i.enums[enum.Name] = enum
	}
	for _, fn := range file.Functions {
		i.functions[fn.Name] = fn
	}
}

func (i *Interpreter) RunMain() error {
	fn := i.functions["main"]
	if fn == nil {
		return fmt.Errorf("main is not defined")
	}
	_, err := i.callFunctionValue(fn, nil)
	return err
}

func (i *Interpreter) RunTest(test *ir.Test) error {
	if test == nil || test.Body == nil {
		return fmt.Errorf("test body is not defined")
	}
	_, err := i.eval(test.Body, NewEnv(i.globals))
	return err
}

func (i *Interpreter) Eval(expr ir.Expr) (Value, error) {
	return i.eval(expr, i.globals)
}

func (i *Interpreter) Exec(stmt ir.Stmt) (Value, bool, error) {
	return i.exec(stmt, i.globals)
}

func (i *Interpreter) exec(stmt ir.Stmt, env *Env) (Value, bool, error) {
	switch s := stmt.(type) {
	case *ir.LetStmt:
		value, err := i.eval(s.Value, env)
		if err != nil {
			return nil, false, err
		}
		env.Define(s.Name, value)
		return nil, false, nil
	case *ir.ObjectDestructureStmt:
		value, err := i.eval(s.Value, env)
		if err != nil {
			return nil, false, err
		}
		object, ok := value.(*Struct)
		if !ok {
			return nil, false, fmt.Errorf("cannot destructure %s", typeName(value))
		}
		for _, field := range s.Fields {
			fieldValue, ok := object.Fields[field.Field]
			if !ok {
				return nil, false, fmt.Errorf("type %s has no field %q", object.TypeName, field.Field)
			}
			env.Define(field.Name, fieldValue)
		}
		return nil, false, nil
	case *ir.AssignStmt:
		value, err := i.eval(s.Value, env)
		if err != nil {
			return nil, false, err
		}
		if err := env.Assign(s.Name, value); err != nil {
			return nil, false, err
		}
		return nil, false, nil
	case *ir.ExprStmt:
		value, err := i.eval(s.Expr, env)
		return value, s.Expr.ResultType() != "Void", err
	default:
		return nil, false, fmt.Errorf("unsupported statement %T", stmt)
	}
}
