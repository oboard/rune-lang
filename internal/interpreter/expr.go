package interpreter

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"

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
	case *ir.TemplateLiteral:
		return i.evalTemplateLiteral(e, env)
	case *ir.CharLiteral:
		return Char(e.Value), nil
	case *ir.RegexLiteral:
		return newRegex(e.Pattern, e.Flags)
	case *ir.BoolLiteral:
		return e.Value, nil
	case *ir.NullLiteral:
		return NullValue, nil
	case *ir.UnaryExpr:
		return i.evalUnary(e, env)
	case *ir.PostfixExpr:
		return i.evalPostfix(e, env)
	case *ir.ResultUnwrapExpr:
		return nil, fmt.Errorf("result unwrap is only supported by generated backends")
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
		if e.Alternative == nil {
			return nil, nil
		}
		return i.eval(e.Alternative, env)
	case *ir.AssignExpr:
		if target, ok := e.Target.(*ir.IndexExpr); ok {
			return i.assignIndex(target, e.Value, env)
		}
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
	case *ir.TupleLiteral:
		tuple := &Tuple{Elements: make([]Value, 0, len(e.Elements))}
		for _, elem := range e.Elements {
			value, err := i.eval(elem, env)
			if err != nil {
				return nil, err
			}
			tuple.Elements = append(tuple.Elements, value)
		}
		return tuple, nil
	case *ir.SpreadExpr:
		return nil, fmt.Errorf("spread is only supported inside array literals")
	case *ir.MapLiteral:
		out := &Map{Entries: make(map[string]mapEntry, len(e.Entries))}
		for _, entry := range e.Entries {
			key, err := i.eval(entry.Key, env)
			if err != nil {
				return nil, err
			}
			value, err := i.eval(entry.Value, env)
			if err != nil {
				return nil, err
			}
			out.Entries[valueKey(key)] = mapEntry{Key: key, Value: value}
		}
		return out, nil
	case *ir.ReactiveLiteral:
		return i.eval(e.Value, env)
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
	case *ir.XMLElement:
		return nil, fmt.Errorf("XML is only supported by the TypeScript backend")
	case *ir.WatchExpr:
		return nil, fmt.Errorf("watch expressions are only supported by generated backends")
	default:
		return nil, fmt.Errorf("unsupported expression %T", expr)
	}
}

func (i *Interpreter) evalTemplateLiteral(lit *ir.TemplateLiteral, env *Env) (Value, error) {
	var b strings.Builder
	for _, part := range lit.Parts {
		b.WriteString(part.Text)
		if part.Expr == nil {
			continue
		}
		value, err := i.eval(part.Expr, env)
		if err != nil {
			return nil, err
		}
		b.WriteString(printValue(value))
	}
	return b.String(), nil
}

func (i *Interpreter) assignIndex(target *ir.IndexExpr, valueExpr ir.Expr, env *Env) (Value, error) {
	receiver, err := i.eval(target.Receiver, env)
	if err != nil {
		return nil, err
	}
	index, err := i.eval(target.Index, env)
	if err != nil {
		return nil, err
	}
	value, err := i.eval(valueExpr, env)
	if err != nil {
		return nil, err
	}
	switch target := receiver.(type) {
	case *Map:
		target.Entries[valueKey(index)] = mapEntry{Key: index, Value: value}
		return nil, nil
	default:
		return nil, fmt.Errorf("%s is not assignable by index", typeName(receiver))
	}
}

func (i *Interpreter) evalPostfix(expr *ir.PostfixExpr, env *Env) (Value, error) {
	if expr.Op != lexer.PlusPlus {
		return nil, fmt.Errorf("unsupported postfix operator %s", expr.Op)
	}
	target, ok := expr.Expr.(*ir.Identifier)
	if !ok {
		return nil, fmt.Errorf("operator '++' expects an assignable name")
	}
	value, ok := env.Get(target.Name)
	if !ok {
		return nil, fmt.Errorf("undefined name %q", target.Name)
	}
	next, err := incrementValue(value)
	if err != nil {
		return nil, err
	}
	if err := env.Assign(target.Name, next); err != nil {
		return nil, err
	}
	return value, nil
}

func incrementValue(value Value) (Value, error) {
	switch n := value.(type) {
	case int:
		return n + 1, nil
	case int8:
		return n + 1, nil
	case int16:
		return n + 1, nil
	case int64:
		return n + 1, nil
	case uint:
		return n + 1, nil
	case uint8:
		return n + 1, nil
	case uint16:
		return n + 1, nil
	case uint64:
		return n + 1, nil
	case float32:
		return n + 1, nil
	case float64:
		return n + 1, nil
	case *big.Int:
		return new(big.Int).Add(n, big.NewInt(1)), nil
	default:
		return nil, fmt.Errorf("operator '++' expects a numeric type")
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
		case int8:
			return -n, nil
		case int16:
			return -n, nil
		case int64:
			return -n, nil
		case float64:
			return -n, nil
		case float32:
			return -n, nil
		case *big.Int:
			return new(big.Int).Neg(n), nil
		default:
			return nil, fmt.Errorf("operator '-' expects a numeric type")
		}
	case lexer.Tilde:
		switch n := value.(type) {
		case int:
			return ^n, nil
		case int8:
			return ^n, nil
		case int16:
			return ^n, nil
		case int64:
			return ^n, nil
		case uint:
			return ^n, nil
		case uint8:
			return ^n, nil
		case uint16:
			return ^n, nil
		case uint64:
			return ^n, nil
		case *big.Int:
			return new(big.Int).Not(n), nil
		default:
			return nil, fmt.Errorf("operator '~' expects an integer type")
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
	case lexer.QuestionQuestion:
		if !isNullValue(left) {
			return left, nil
		}
		return i.eval(expr.Right, env)
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
		return evalNumericBytes(expr.Op, left, right)
	case lexer.Minus, lexer.Star, lexer.Slash, lexer.Percent:
		return evalNumericBytes(expr.Op, left, right)
	case lexer.BitAnd, lexer.BitOr, lexer.BitXor, lexer.ShiftLeft, lexer.ShiftRight, lexer.UnsignedShiftRight:
		return evalBitwiseBytes(expr.Op, left, right)
	case lexer.EqualEqual:
		return reflect.DeepEqual(left, right), nil
	case lexer.BangEqual:
		return !reflect.DeepEqual(left, right), nil
	case lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
		return evalOrderedComparison(expr.Op, left, right)
	}
	return nil, fmt.Errorf("unsupported binary operator %s", expr.Op)
}

func evalNumericBytes(op lexer.Kind, left Value, right Value) (Value, error) {
	switch l := left.(type) {
	case int:
		r, ok := right.(int)
		if !ok {
			return nil, fmt.Errorf("arithmetic operands must have matching numeric types")
		}
		return evalSignedNumericBytes(op, l, r)
	case int8:
		r, ok := right.(int8)
		if !ok {
			return nil, fmt.Errorf("arithmetic operands must have matching numeric types")
		}
		return evalSignedNumericBytes(op, l, r)
	case int16:
		r, ok := right.(int16)
		if !ok {
			return nil, fmt.Errorf("arithmetic operands must have matching numeric types")
		}
		return evalSignedNumericBytes(op, l, r)
	case int64:
		r, ok := right.(int64)
		if !ok {
			return nil, fmt.Errorf("arithmetic operands must have matching numeric types")
		}
		return evalSignedNumericBytes(op, l, r)
	case uint:
		r, ok := right.(uint)
		if !ok {
			return nil, fmt.Errorf("arithmetic operands must have matching numeric types")
		}
		return evalUnsignedNumericBytes(op, l, r)
	case uint8:
		r, ok := right.(uint8)
		if !ok {
			return nil, fmt.Errorf("arithmetic operands must have matching numeric types")
		}
		return evalUnsignedNumericBytes(op, l, r)
	case uint16:
		r, ok := right.(uint16)
		if !ok {
			return nil, fmt.Errorf("arithmetic operands must have matching numeric types")
		}
		return evalUnsignedNumericBytes(op, l, r)
	case uint64:
		r, ok := right.(uint64)
		if !ok {
			return nil, fmt.Errorf("arithmetic operands must have matching numeric types")
		}
		return evalUnsignedNumericBytes(op, l, r)
	case float32:
		r, ok := right.(float32)
		if !ok {
			return nil, fmt.Errorf("arithmetic operands must have matching numeric types")
		}
		return evalFloatNumericBytes(op, l, r)
	case float64:
		r, ok := right.(float64)
		if !ok {
			return nil, fmt.Errorf("arithmetic operands must have matching numeric types")
		}
		return evalFloatNumericBytes(op, l, r)
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
	return nil, fmt.Errorf("arithmetic expects numeric operands")
}

type signedNumber interface {
	~int | ~int8 | ~int16 | ~int64
}

type unsignedNumber interface {
	~uint | ~uint8 | ~uint16 | ~uint64
}

type floatNumber interface {
	~float32 | ~float64
}

func evalSignedNumericBytes[T signedNumber](op lexer.Kind, left T, right T) (Value, error) {
	switch op {
	case lexer.Plus:
		return left + right, nil
	case lexer.Minus:
		return left - right, nil
	case lexer.Star:
		return left * right, nil
	case lexer.Slash:
		return left / right, nil
	case lexer.Percent:
		return left % right, nil
	default:
		return nil, fmt.Errorf("unsupported arithmetic operator %s", op)
	}
}

func evalUnsignedNumericBytes[T unsignedNumber](op lexer.Kind, left T, right T) (Value, error) {
	switch op {
	case lexer.Plus:
		return left + right, nil
	case lexer.Minus:
		return left - right, nil
	case lexer.Star:
		return left * right, nil
	case lexer.Slash:
		return left / right, nil
	case lexer.Percent:
		return left % right, nil
	default:
		return nil, fmt.Errorf("unsupported arithmetic operator %s", op)
	}
}

func evalFloatNumericBytes[T floatNumber](op lexer.Kind, left T, right T) (Value, error) {
	switch op {
	case lexer.Plus:
		return left + right, nil
	case lexer.Minus:
		return left - right, nil
	case lexer.Star:
		return left * right, nil
	case lexer.Slash:
		return left / right, nil
	case lexer.Percent:
		return nil, fmt.Errorf("operator '%%' expects integer operands")
	default:
		return nil, fmt.Errorf("unsupported arithmetic operator %s", op)
	}
}

func evalBitwiseBytes(op lexer.Kind, left Value, right Value) (Value, error) {
	switch l := left.(type) {
	case int:
		r, ok := right.(int)
		if !ok {
			return nil, fmt.Errorf("bitwise operands must have matching integer types")
		}
		return evalSignedBitwiseBytes(op, l, r)
	case int8:
		r, ok := right.(int8)
		if !ok {
			return nil, fmt.Errorf("bitwise operands must have matching integer types")
		}
		return evalSignedBitwiseBytes(op, l, r)
	case int16:
		r, ok := right.(int16)
		if !ok {
			return nil, fmt.Errorf("bitwise operands must have matching integer types")
		}
		return evalSignedBitwiseBytes(op, l, r)
	case int64:
		r, ok := right.(int64)
		if !ok {
			return nil, fmt.Errorf("bitwise operands must have matching integer types")
		}
		return evalSignedBitwiseBytes(op, l, r)
	case uint:
		r, ok := right.(uint)
		if !ok {
			return nil, fmt.Errorf("bitwise operands must have matching integer types")
		}
		return evalUnsignedBitwiseBytes(op, l, r)
	case uint8:
		r, ok := right.(uint8)
		if !ok {
			return nil, fmt.Errorf("bitwise operands must have matching integer types")
		}
		return evalUnsignedBitwiseBytes(op, l, r)
	case uint16:
		r, ok := right.(uint16)
		if !ok {
			return nil, fmt.Errorf("bitwise operands must have matching integer types")
		}
		return evalUnsignedBitwiseBytes(op, l, r)
	case uint64:
		r, ok := right.(uint64)
		if !ok {
			return nil, fmt.Errorf("bitwise operands must have matching integer types")
		}
		return evalUnsignedBitwiseBytes(op, l, r)
	case *big.Int:
		r, ok := right.(*big.Int)
		if !ok {
			return nil, fmt.Errorf("bitwise operands must have matching integer types")
		}
		return evalBigIntBitwiseBytes(op, l, r)
	default:
		return nil, fmt.Errorf("bitwise operator expects integer operands")
	}
}

func evalSignedBitwiseBytes[T signedNumber](op lexer.Kind, left T, right T) (Value, error) {
	switch op {
	case lexer.BitAnd:
		return left & right, nil
	case lexer.BitOr:
		return left | right, nil
	case lexer.BitXor:
		return left ^ right, nil
	case lexer.ShiftLeft:
		if right < 0 {
			return nil, fmt.Errorf("shift count must be non-negative")
		}
		return left << uint(right), nil
	case lexer.ShiftRight:
		if right < 0 {
			return nil, fmt.Errorf("shift count must be non-negative")
		}
		return left >> uint(right), nil
	case lexer.UnsignedShiftRight:
		return nil, fmt.Errorf("operator '>>>' expects an unsigned integer left operand")
	default:
		return nil, fmt.Errorf("unsupported bitwise operator %s", op)
	}
}

func evalUnsignedBitwiseBytes[T unsignedNumber](op lexer.Kind, left T, right T) (Value, error) {
	switch op {
	case lexer.BitAnd:
		return left & right, nil
	case lexer.BitOr:
		return left | right, nil
	case lexer.BitXor:
		return left ^ right, nil
	case lexer.ShiftLeft:
		return left << uint(right), nil
	case lexer.ShiftRight, lexer.UnsignedShiftRight:
		return left >> uint(right), nil
	default:
		return nil, fmt.Errorf("unsupported bitwise operator %s", op)
	}
}

func evalBigIntBitwiseBytes(op lexer.Kind, left *big.Int, right *big.Int) (Value, error) {
	out := new(big.Int)
	switch op {
	case lexer.BitAnd:
		return out.And(left, right), nil
	case lexer.BitOr:
		return out.Or(left, right), nil
	case lexer.BitXor:
		return out.Xor(left, right), nil
	case lexer.ShiftLeft:
		if right.Sign() < 0 {
			return nil, fmt.Errorf("shift count must be non-negative")
		}
		return out.Lsh(left, uint(right.Int64())), nil
	case lexer.ShiftRight:
		if right.Sign() < 0 {
			return nil, fmt.Errorf("shift count must be non-negative")
		}
		return out.Rsh(left, uint(right.Int64())), nil
	case lexer.UnsignedShiftRight:
		return nil, fmt.Errorf("operator '>>>' expects an unsigned integer left operand")
	default:
		return nil, fmt.Errorf("unsupported bitwise operator %s", op)
	}
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
	case int8:
		r, ok := right.(int8)
		if !ok {
			return 0, fmt.Errorf("comparison operands must have matching ordered types")
		}
		return compareOrderedValues(l, r), nil
	case int16:
		r, ok := right.(int16)
		if !ok {
			return 0, fmt.Errorf("comparison operands must have matching ordered types")
		}
		return compareOrderedValues(l, r), nil
	case int64:
		r, ok := right.(int64)
		if !ok {
			return 0, fmt.Errorf("comparison operands must have matching ordered types")
		}
		return compareOrderedValues(l, r), nil
	case uint:
		r, ok := right.(uint)
		if !ok {
			return 0, fmt.Errorf("comparison operands must have matching ordered types")
		}
		return compareOrderedValues(l, r), nil
	case uint8:
		r, ok := right.(uint8)
		if !ok {
			return 0, fmt.Errorf("comparison operands must have matching ordered types")
		}
		return compareOrderedValues(l, r), nil
	case uint16:
		r, ok := right.(uint16)
		if !ok {
			return 0, fmt.Errorf("comparison operands must have matching ordered types")
		}
		return compareOrderedValues(l, r), nil
	case uint64:
		r, ok := right.(uint64)
		if !ok {
			return 0, fmt.Errorf("comparison operands must have matching ordered types")
		}
		return compareOrderedValues(l, r), nil
	case float32:
		r, ok := right.(float32)
		if !ok {
			return 0, fmt.Errorf("comparison operands must have matching ordered types")
		}
		return compareOrderedValues(l, r), nil
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
	case Char:
		r, ok := right.(Char)
		if !ok {
			return 0, fmt.Errorf("comparison operands must have matching ordered types")
		}
		return compareOrderedValues(l, r), nil
	default:
		return 0, fmt.Errorf("comparison expects a numeric type, String, or Char")
	}
}

type orderedValue interface {
	~int | ~int8 | ~int16 | ~int64 | ~int32 | ~uint | ~uint8 | ~uint16 | ~uint64 | ~float32 | ~float64
}

func compareOrderedValues[T orderedValue](left T, right T) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
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
	if expr.Static {
		ident, ok := expr.Receiver.(*ir.Identifier)
		if !ok {
			return nil, fmt.Errorf("static selector receiver must be a type")
		}
		if _, shadowed := env.Get(ident.Name); shadowed {
			return nil, fmt.Errorf("static selector receiver %q is a value, not a type", ident.Name)
		}
		typ := i.types[ident.Name]
		if typ == nil {
			return nil, fmt.Errorf("unknown type %q", ident.Name)
		}
		for _, method := range typ.Methods {
			if method.Name == expr.Name && method.Static {
				return method, nil
			}
		}
		return nil, fmt.Errorf("type %s has no static method %q", typ.Name, expr.Name)
	}
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
	case *ir.AtExpr:
		if value.Name != "" {
			return nil, fmt.Errorf("cannot select %q from module @%s", expr.Name, value.Name)
		}
		if fn := i.functions[selectorResolvedName(expr)]; fn != nil {
			return fn, nil
		}
		return nil, fmt.Errorf("import has no member %q", expr.Name)
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
	switch value := receiver.(type) {
	case *Array:
		i, ok := index.(int)
		if !ok {
			return nil, fmt.Errorf("array index expects Int")
		}
		if i < 0 || i >= len(value.Elements) {
			return nil, fmt.Errorf("array index %d out of range", i)
		}
		return value.Elements[i], nil
	case *Tuple:
		i, ok := index.(int)
		if !ok {
			return nil, fmt.Errorf("tuple index expects Int")
		}
		if i < 0 || i >= len(value.Elements) {
			return nil, fmt.Errorf("tuple index %d out of range", i)
		}
		return value.Elements[i], nil
	case *Map:
		entry, ok := value.Entries[valueKey(index)]
		if !ok {
			return NullValue, nil
		}
		return entry.Value, nil
	default:
		return nil, fmt.Errorf("%s is not indexable", typeName(receiver))
	}
}

func isNullValue(value Value) bool {
	_, ok := value.(nullValue)
	return ok
}

func typeName(value Value) string {
	switch v := value.(type) {
	case nil:
		return "Void"
	case int:
		return string(checker.Int)
	case int8:
		return string(checker.Int8)
	case int16:
		return string(checker.Int16)
	case int64:
		return string(checker.Int64)
	case float64:
		return string(checker.Double)
	case float32:
		return string(checker.Float)
	case *big.Int:
		return string(checker.BigInt)
	case uint:
		return string(checker.UInt)
	case uint8:
		return string(checker.UInt8)
	case uint16:
		return string(checker.UInt16)
	case uint64:
		return string(checker.UInt64)
	case nullValue:
		return string(checker.Null)
	case string:
		return string(checker.String)
	case Char:
		return string(checker.Char)
	case bool:
		return string(checker.Bool)
	case *Array:
		return "Array"
	case *Tuple:
		return "Tuple"
	case *Map:
		return "Map"
	case *Set:
		return "Set"
	case *Bytes:
		return string(checker.Bytes)
	case *Buffer:
		return string(checker.Buffer)
	case *Reader:
		return string(checker.Reader)
	case *Writer:
		return string(checker.Writer)
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
