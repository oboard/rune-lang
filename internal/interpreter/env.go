package interpreter

import "fmt"

type Env struct {
	parent *Env
	values map[string]Value
}

func NewEnv(parent *Env) *Env {
	return &Env{parent: parent, values: map[string]Value{}}
}

func (e *Env) Define(name string, value Value) {
	e.values[name] = value
}

func (e *Env) Assign(name string, value Value) error {
	if _, ok := e.values[name]; ok {
		e.values[name] = value
		return nil
	}
	if e.parent != nil {
		return e.parent.Assign(name, value)
	}
	return fmt.Errorf("undefined name %q", name)
}

func (e *Env) Get(name string) (Value, bool) {
	if value, ok := e.values[name]; ok {
		return value, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}
