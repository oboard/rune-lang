package interpreter

import (
	"fmt"
	"math/big"
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
	case *ir.DoubleLiteral:
		return e.Value, nil
	case *ir.BigIntLiteral:
		value, ok := new(big.Int).SetString(e.Value, 10)
		if !ok {
			return nil, fmt.Errorf("invalid BigInt literal %q", e.Value)
		}
		return value, nil
	case *ir.StringLiteral:
		return e.Value, nil
	case *ir.RegexLiteral:
		return newRegex(e.Pattern, e.Flags)
	case *ir.BoolLiteral:
		return e.Value, nil
	case *ir.NullLiteral:
		return NullValue, nil
	case *ir.UnaryExpr:
		return i.evalUnary(e, env)
	case *ir.BinaryExpr:
		return i.evalBinary(e, env)
	case *ir.TernaryExpr:
		condition, err := i.eval(e.Condition, env)
		if err != nil {
			return nil, err
		}
		value, ok := condition.(bool)
		if !ok {
			return nil, fmt.Errorf("ternary condition expects Bool")
		}
		if value {
			return i.eval(e.Consequence, env)
		}
		return i.eval(e.Alternative, env)
	case *ir.AssignExpr:
		value, err := i.eval(e.Value, env)
		if err != nil {
			return nil, err
		}
		if err := env.Assign(e.Name, value); err != nil {
			return nil, err
		}
		return nil, nil
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
			if spread, ok := elem.(*ir.SpreadExpr); ok {
				value, err := i.eval(spread.Expr, env)
				if err != nil {
					return nil, err
				}
				spreadArray, ok := value.(*Array)
				if !ok {
					return nil, fmt.Errorf("spread expects Array")
				}
				array.Elements = append(array.Elements, spreadArray.Elements...)
				continue
			}
			value, err := i.eval(elem, env)
			if err != nil {
				return nil, err
			}
			array.Elements = append(array.Elements, value)
		}
		return array, nil
	case *ir.StructLiteral:
		fields := map[string]Value{}
		order := make([]string, 0, len(e.Fields))
		for _, field := range e.Fields {
			value, err := i.eval(field.Value, env)
			if err != nil {
				return nil, err
			}
			fields[field.Name] = value
			order = append(order, field.Name)
		}
		return &Struct{TypeName: e.TypeName, Fields: fields, Order: order}, nil
	case *ir.AnonymousObjectLiteral:
		fields := map[string]Value{}
		order := make([]string, 0, len(e.Fields))
		obj := &Struct{TypeName: string(e.ResultType()), Fields: fields, Order: order}
		objectEnv := NewEnv(env)
		objectEnv.Define("this", obj)
		for _, field := range e.Fields {
			value, err := i.eval(field.Value, objectEnv)
			if err != nil {
				return nil, err
			}
			fields[field.Name] = value
			obj.Order = append(obj.Order, field.Name)
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
		switch n := value.(type) {
		case int:
			return -n, nil
		case float64:
			return -n, nil
		case *big.Int:
			return new(big.Int).Neg(n), nil
		default:
			return nil, fmt.Errorf("operator '-' expects Int, Double, or BigInt")
		}
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
	switch expr.Op {
	case lexer.AndAnd:
		l, ok := left.(bool)
		if !ok {
			return nil, fmt.Errorf("operator '&&' expects Bool")
		}
		if !l {
			return false, nil
		}
		right, err := i.eval(expr.Right, env)
		if err != nil {
			return nil, err
		}
		r, ok := right.(bool)
		if !ok {
			return nil, fmt.Errorf("operator '&&' expects Bool")
		}
		return r, nil
	case lexer.OrOr:
		l, ok := left.(bool)
		if !ok {
			return nil, fmt.Errorf("operator '||' expects Bool")
		}
		if l {
			return true, nil
		}
		right, err := i.eval(expr.Right, env)
		if err != nil {
			return nil, err
		}
		r, ok := right.(bool)
		if !ok {
			return nil, fmt.Errorf("operator '||' expects Bool")
		}
		return r, nil
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
		return evalNumericBinary(expr.Op, left, right)
	case lexer.Minus, lexer.Star, lexer.Slash, lexer.Percent:
		return evalNumericBinary(expr.Op, left, right)
	case lexer.EqualEqual:
		return reflect.DeepEqual(left, right), nil
	case lexer.BangEqual:
		return !reflect.DeepEqual(left, right), nil
	case lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
		return evalOrderedComparison(expr.Op, left, right)
	}
	return nil, fmt.Errorf("unsupported binary operator %s", expr.Op)
}

func evalNumericBinary(op lexer.Kind, left Value, right Value) (Value, error) {
	switch l := left.(type) {
	case int:
		r, ok := right.(int)
		if !ok {
			return nil, fmt.Errorf("arithmetic operands must have matching numeric types")
		}
		switch op {
		case lexer.Plus:
			return l + r, nil
		case lexer.Minus:
			return l - r, nil
		case lexer.Star:
			return l * r, nil
		case lexer.Slash:
			return l / r, nil
		case lexer.Percent:
			return l % r, nil
		}
	case float64:
		r, ok := right.(float64)
		if !ok {
			return nil, fmt.Errorf("arithmetic operands must have matching numeric types")
		}
		switch op {
		case lexer.Plus:
			return l + r, nil
		case lexer.Minus:
			return l - r, nil
		case lexer.Star:
			return l * r, nil
		case lexer.Slash:
			return l / r, nil
		case lexer.Percent:
			return nil, fmt.Errorf("operator '%%' expects Int or BigInt")
		}
	case *big.Int:
		r, ok := right.(*big.Int)
		if !ok {
			return nil, fmt.Errorf("arithmetic operands must have matching numeric types")
		}
		out := new(big.Int)
		switch op {
		case lexer.Plus:
			return out.Add(l, r), nil
		case lexer.Minus:
			return out.Sub(l, r), nil
		case lexer.Star:
			return out.Mul(l, r), nil
		case lexer.Slash:
			return out.Quo(l, r), nil
		case lexer.Percent:
			return out.Rem(l, r), nil
		}
	}
	return nil, fmt.Errorf("arithmetic expects Int, Double, or BigInt")
}

func evalOrderedComparison(op lexer.Kind, left Value, right Value) (Value, error) {
	cmp, err := compareOrdered(left, right)
	if err != nil {
		return nil, err
	}
	switch op {
	case lexer.Less:
		return cmp < 0, nil
	case lexer.LessEqual:
		return cmp <= 0, nil
	case lexer.Greater:
		return cmp > 0, nil
	case lexer.GreaterEqual:
		return cmp >= 0, nil
	default:
		return nil, fmt.Errorf("unsupported comparison operator %s", op)
	}
}

func compareOrdered(left Value, right Value) (int, error) {
	switch l := left.(type) {
	case int:
		r, ok := right.(int)
		if !ok {
			return 0, fmt.Errorf("comparison operands must have matching ordered types")
		}
		switch {
		case l < r:
			return -1, nil
		case l > r:
			return 1, nil
		default:
			return 0, nil
		}
	case float64:
		r, ok := right.(float64)
		if !ok {
			return 0, fmt.Errorf("comparison operands must have matching ordered types")
		}
		switch {
		case l < r:
			return -1, nil
		case l > r:
			return 1, nil
		default:
			return 0, nil
		}
	case *big.Int:
		r, ok := right.(*big.Int)
		if !ok {
			return 0, fmt.Errorf("comparison operands must have matching ordered types")
		}
		return l.Cmp(r), nil
	case string:
		r, ok := right.(string)
		if !ok {
			return 0, fmt.Errorf("comparison operands must have matching ordered types")
		}
		switch {
		case l < r:
			return -1, nil
		case l > r:
			return 1, nil
		default:
			return 0, nil
		}
	default:
		return 0, fmt.Errorf("comparison expects Int, Double, BigInt, or String")
	}
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
	if ident, ok := expr.Receiver.(*ir.Identifier); ok {
		if _, exists := env.Get(ident.Name); !exists {
			if enum := i.enums[ident.Name]; enum != nil {
				for _, member := range enum.Members {
					if member.Name == expr.Name {
						return EnumValue{TypeName: enum.Name, Name: member.Name, Value: member.Value}, nil
					}
				}
				return nil, fmt.Errorf("enum %s has no member %q", enum.Name, expr.Name)
			}
		}
	}
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
	case float64:
		return string(checker.Double)
	case *big.Int:
		return string(checker.BigInt)
	case nullValue:
		return string(checker.Null)
	case string:
		return string(checker.String)
	case bool:
		return string(checker.Bool)
	case *Array:
		return "Array"
	case *Map:
		return "Map"
	case *Set:
		return "Set"
	case *Regex:
		return string(checker.Regex)
	case EnumValue:
		return v.TypeName
	case *Struct:
		return v.TypeName
	default:
		return fmt.Sprintf("%T", value)
	}
}
