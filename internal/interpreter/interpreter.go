package interpreter

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/oboard/rune-lang/internal/ir"
)

type Interpreter struct {
	file                  *ir.File
	functions             map[string]*ir.Function
	types                 map[string]*ir.StructType
	enums                 map[string]*ir.EnumType
	globals               *Env
	in                    *bufio.Reader
	out                   io.Writer
	compileTime           bool
	compileTimeLocalNames map[string]bool
}

func New(file *ir.File, opts ...Option) *Interpreter {
	i := &Interpreter{
		functions: map[string]*ir.Function{},
		types:     map[string]*ir.StructType{},
		enums:     map[string]*ir.EnumType{},
		globals:   NewEnv(nil),
		in:        bufio.NewReader(os.Stdin),
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

func WithInput(in io.Reader) Option {
	return func(i *Interpreter) {
		i.in = bufio.NewReader(in)
	}
}

func WithCompileTime() Option {
	return func(i *Interpreter) {
		i.compileTime = true
	}
}

// EnableCompileTimeLocalName allows a bare-identifier call inside a compile-time
// macro body to resolve to the same-named stdlib function with a Rune body.
func (i *Interpreter) EnableCompileTimeLocalName(name string) {
	if i.compileTimeLocalNames == nil {
		i.compileTimeLocalNames = map[string]bool{}
	}
	i.compileTimeLocalNames[name] = true
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
	for _, constant := range file.Constants {
		value, err := i.eval(constant.Value, i.globals)
		if err == nil {
			i.globals.Define(constant.Name, value)
		}
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

func (i *Interpreter) EvalWithBindings(expr ir.Expr, bindings map[string]Value) (Value, error) {
	env := NewEnv(i.globals)
	for name, value := range bindings {
		env.Define(name, value)
	}
	return i.eval(expr, env)
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
	case *ir.ReturnStmt:
		var value Value
		if s.Value != nil {
			evaluated, err := i.eval(s.Value, env)
			if err != nil {
				return nil, false, err
			}
			value = evaluated
		}
		return nil, false, &returnSignal{value: value}
	case *ir.BreakStmt:
		return nil, false, &breakSignal{}
	case *ir.ContinueStmt:
		return nil, false, &continueSignal{}
	case *ir.ThrowStmt:
		value, err := i.eval(s.Value, env)
		if err != nil {
			return nil, false, err
		}
		return nil, false, &throwSignal{value: value}
	case *ir.TryStmt:
		return nil, false, i.execTry(s, env)
	case *ir.IfStmt:
		condition, err := i.eval(s.Cond, env)
		if err != nil {
			return nil, false, err
		}
		keepGoing, ok := condition.(bool)
		if !ok {
			return nil, false, fmt.Errorf("if condition expects Bool")
		}
		statements := s.Then
		if !keepGoing {
			statements = s.Else
		}
		for _, statement := range statements {
			if _, _, err := i.exec(statement, env); err != nil {
				return nil, false, err
			}
		}
		return nil, false, nil
	default:
		return nil, false, fmt.Errorf("unsupported statement %T", stmt)
	}
}

type returnSignal struct {
	value Value
}

func (e *returnSignal) Error() string {
	return "return"
}

type breakSignal struct{}

func (*breakSignal) Error() string {
	return "break"
}

type continueSignal struct{}

func (*continueSignal) Error() string {
	return "continue"
}

type throwSignal struct {
	value Value
}

func (e *throwSignal) Error() string {
	return "uncaught throw"
}

func (i *Interpreter) execTry(stmt *ir.TryStmt, env *Env) error {
	var runErr error
	func() {
		defer func() {
			if stmt.Finally == nil {
				return
			}
			if _, err := i.evalBlock(&ir.BlockExpr{Statements: stmt.Finally}, env); err != nil {
				runErr = err
			}
		}()
		if _, err := i.evalBlock(&ir.BlockExpr{Statements: stmt.Body}, env); err != nil {
			thrown, ok := err.(*throwSignal)
			if !ok || stmt.Catch == nil {
				runErr = err
				return
			}
			catchEnv := NewEnv(env)
			if stmt.CatchName != "" {
				catchEnv.Define(stmt.CatchName, thrown.value)
			}
			if _, catchErr := i.evalBlock(&ir.BlockExpr{Statements: stmt.Catch}, catchEnv); catchErr != nil {
				runErr = catchErr
			}
		}
	}()
	return runErr
}
