package interpreter

import (
	"fmt"
	"reflect"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
	"github.com/oboard/rune-lang/internal/lexer"
)

func (i *Interpreter) eval(expr ir.Expr, env *Env) (Value, error) {
	switch e := expr.(type) {
	case *ir.Identifier:
		if value, ok := env.Get(e.Name); ok {
			return value, nil
		}
		if fn := i.functions[e.Name]; fn != nil {
			return fn, nil
		}
		return nil, fmt.Errorf("undefined name %q", e.Name)
	case *ir.AtExpr:
		return e, nil
	case *ir.ThisExpr:
		if value, ok := env.Get("this"); ok {
			return value, nil
		}
		return nil, fmt.Errorf("this is not defined")
	case *ir.IntegerLiteral:
		return e.Value, nil
	case *ir.StringLiteral:
		return e.Value, nil
	case *ir.BoolLiteral:
		return e.Value, nil
	case *ir.UnaryExpr:
		return i.evalUnary(e, env)
	case *ir.BinaryExpr:
		return i.evalBinary(e, env)
	case *ir.CallExpr:
		return i.evalCall(e, env)
	case *ir.LambdaExpr:
		return &Closure{Params: e.Params, Body: e.Body, Env: env}, nil
	case *ir.SelectorExpr:
		return i.evalSelector(e, env)
	case *ir.IndexExpr:
		receiver, err := i.eval(e.Receiver, env)
		if err != nil {
			return nil, err
		}
		index, err := i.eval(e.Index, env)
		if err != nil {
			return nil, err
		}
		return indexValue(receiver, index)
	case *ir.ArrayLiteral:
		array := &Array{Elements: make([]Value, 0, len(e.Elements))}
		for _, elem := range e.Elements {
			value, err := i.eval(elem, env)
			if err != nil {
				return nil, err
			}
			array.Elements = append(array.Elements, value)
		}
		return array, nil
	case *ir.StructLiteral:
		fields := map[string]Value{}
		for _, field := range e.Fields {
			value, err := i.eval(field.Value, env)
			if err != nil {
				return nil, err
			}
			fields[field.Name] = value
		}
		return &Struct{TypeName: e.TypeName, Fields: fields}, nil
	case *ir.AnonymousObjectLiteral:
		fields := map[string]Value{}
		obj := &Struct{TypeName: string(e.ResultType()), Fields: fields}
		objectEnv := NewEnv(env)
		objectEnv.Define("this", obj)
		for _, field := range e.Fields {
			value, err := i.eval(field.Value, objectEnv)
			if err != nil {
				return nil, err
			}
			fields[field.Name] = value
		}
		return obj, nil
	case *ir.BlockExpr:
		return i.evalBlock(e, env)
	case *ir.PatternBlock:
		return nil, fmt.Errorf("pattern block cannot be evaluated without a subject")
	case *ir.MatchExpr:
		subject, err := i.eval(e.Subject, env)
		if err != nil {
			return nil, err
		}
		return i.evalPatternBlock(&ir.PatternBlock{ExprBase: e.ExprBase, Branches: e.Branches}, subject, env)
	default:
		return nil, fmt.Errorf("unsupported expression %T", expr)
	}
}

func (i *Interpreter) evalUnary(expr *ir.UnaryExpr, env *Env) (Value, error) {
	value, err := i.eval(expr.Expr, env)
	if err != nil {
		return nil, err
	}
	switch expr.Op {
	case lexer.Minus:
		n, ok := value.(int)
		if !ok {
			return nil, fmt.Errorf("operator '-' expects Int")
		}
		return -n, nil
	case lexer.Bang:
		b, ok := value.(bool)
		if !ok {
			return nil, fmt.Errorf("operator '!' expects Bool")
		}
		return !b, nil
	default:
		return nil, fmt.Errorf("unsupported unary operator %s", expr.Op)
	}
}

func (i *Interpreter) evalBinary(expr *ir.BinaryExpr, env *Env) (Value, error) {
	left, err := i.eval(expr.Left, env)
	if err != nil {
		return nil, err
	}
	right, err := i.eval(expr.Right, env)
	if err != nil {
		return nil, err
	}
	switch expr.Op {
	case lexer.Plus:
		if leftString, ok := left.(string); ok {
			rightString, ok := right.(string)
			if !ok {
				return nil, fmt.Errorf("string concatenation expects String")
			}
			return leftString + rightString, nil
		}
		if _, ok := right.(string); ok {
			return nil, fmt.Errorf("string concatenation expects String")
		}
		l, ok := left.(int)
		if !ok {
			return nil, fmt.Errorf("arithmetic expects Int")
		}
		r, ok := right.(int)
		if !ok {
			return nil, fmt.Errorf("arithmetic expects Int")
		}
		return l + r, nil
	case lexer.Minus, lexer.Star, lexer.Slash, lexer.Percent:
		l, ok := left.(int)
		if !ok {
			return nil, fmt.Errorf("arithmetic expects Int")
		}
		r, ok := right.(int)
		if !ok {
			return nil, fmt.Errorf("arithmetic expects Int")
		}
		switch expr.Op {
		case lexer.Minus:
			return l - r, nil
		case lexer.Star:
			return l * r, nil
		case lexer.Slash:
			return l / r, nil
		case lexer.Percent:
			return l % r, nil
		}
	case lexer.EqualEqual:
		return reflect.DeepEqual(left, right), nil
	case lexer.BangEqual:
		return !reflect.DeepEqual(left, right), nil
	case lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
		l, ok := left.(int)
		if !ok {
			return nil, fmt.Errorf("comparison expects Int")
		}
		r, ok := right.(int)
		if !ok {
			return nil, fmt.Errorf("comparison expects Int")
		}
		switch expr.Op {
		case lexer.Less:
			return l < r, nil
		case lexer.LessEqual:
			return l <= r, nil
		case lexer.Greater:
			return l > r, nil
		case lexer.GreaterEqual:
			return l >= r, nil
		}
	}
	return nil, fmt.Errorf("unsupported binary operator %s", expr.Op)
}

func (i *Interpreter) evalBlock(block *ir.BlockExpr, env *Env) (Value, error) {
	local := NewEnv(env)
	var result Value
	for _, stmt := range block.Statements {
		value, _, err := i.exec(stmt, local)
		if err != nil {
			return nil, err
		}
		result = value
	}
	return result, nil
}

func (i *Interpreter) evalSelector(expr *ir.SelectorExpr, env *Env) (Value, error) {
	receiver, err := i.eval(expr.Receiver, env)
	if err != nil {
		return nil, err
	}
	switch value := receiver.(type) {
	case *Struct:
		field, ok := value.Fields[expr.Name]
		if !ok {
			return nil, fmt.Errorf("type %s has no field %q", value.TypeName, expr.Name)
		}
		return field, nil
	default:
		return nil, fmt.Errorf("cannot select %q from %s", expr.Name, typeName(value))
	}
}

func indexValue(receiver Value, index Value) (Value, error) {
	array, ok := receiver.(*Array)
	if !ok {
		return nil, fmt.Errorf("%s is not indexable", typeName(receiver))
	}
	i, ok := index.(int)
	if !ok {
		return nil, fmt.Errorf("array index expects Int")
	}
	if i < 0 || i >= len(array.Elements) {
		return nil, fmt.Errorf("array index %d out of range", i)
	}
	return array.Elements[i], nil
}

func typeName(value Value) string {
	switch v := value.(type) {
	case nil:
		return "Void"
	case int:
		return string(checker.Int)
	case string:
		return string(checker.String)
	case bool:
		return string(checker.Bool)
	case *Array:
		return "Array"
	case *Struct:
		return v.TypeName
	default:
		return fmt.Sprintf("%T", value)
	}
}
