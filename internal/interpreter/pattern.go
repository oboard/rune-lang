package interpreter

import (
	"fmt"
	"reflect"

	"github.com/oboard/rune-lang/internal/ir"
	"github.com/oboard/rune-lang/internal/lexer"
)

func (i *Interpreter) evalPatternBlock(block *ir.PatternBlock, subject Value, env *Env) (Value, error) {
	for _, branch := range block.Branches {
		matched, err := i.matchPattern(branch.Pattern, subject, env)
		if err != nil {
			return nil, err
		}
		if matched {
			return i.eval(branch.Expr, env)
		}
	}
	return nil, nil
}

func (i *Interpreter) matchPattern(pattern ir.Pattern, subject Value, env *Env) (bool, error) {
	switch p := pattern.(type) {
	case *ir.WildcardPattern:
		return true, nil
	case *ir.LiteralPattern:
		value, err := i.eval(p.Value, env)
		if err != nil {
			return false, err
		}
		return reflect.DeepEqual(subject, value), nil
	case *ir.ComparePattern:
		value, err := i.eval(p.Value, env)
		if err != nil {
			return false, err
		}
		left, ok := subject.(int)
		if !ok {
			return false, fmt.Errorf("comparison pattern expects Int subject")
		}
		right, ok := value.(int)
		if !ok {
			return false, fmt.Errorf("comparison pattern expects Int value")
		}
		switch p.Op {
		case lexer.Less:
			return left < right, nil
		case lexer.LessEqual:
			return left <= right, nil
		case lexer.Greater:
			return left > right, nil
		case lexer.GreaterEqual:
			return left >= right, nil
		case lexer.EqualEqual:
			return left == right, nil
		case lexer.BangEqual:
			return left != right, nil
		default:
			return false, fmt.Errorf("unsupported comparison pattern %s", p.Op)
		}
	case *ir.TuplePattern:
		array, ok := subject.(*Array)
		if !ok || len(array.Elements) != len(p.Elements) {
			return false, nil
		}
		for idx, elem := range p.Elements {
			matched, err := i.matchPattern(elem, array.Elements[idx], env)
			if err != nil || !matched {
				return matched, err
			}
		}
		return true, nil
	default:
		return false, fmt.Errorf("unsupported pattern %T", pattern)
	}
}
