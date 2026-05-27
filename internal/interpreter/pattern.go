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
		cmp, err := compareOrdered(subject, value)
		if err != nil {
			return false, fmt.Errorf("comparison pattern expects matching ordered values: %w", err)
		}
		switch p.Op {
		case lexer.Less:
			return cmp < 0, nil
		case lexer.LessEqual:
			return cmp <= 0, nil
		case lexer.Greater:
			return cmp > 0, nil
		case lexer.GreaterEqual:
			return cmp >= 0, nil
		case lexer.EqualEqual:
			return cmp == 0, nil
		case lexer.BangEqual:
			return cmp != 0, nil
		default:
			return false, fmt.Errorf("unsupported comparison pattern %s", p.Op)
		}
	case *ir.RangePattern:
		start, err := i.eval(p.Start, env)
		if err != nil {
			return false, err
		}
		end, err := i.eval(p.End, env)
		if err != nil {
			return false, err
		}
		lower, err := compareOrdered(subject, start)
		if err != nil {
			return false, fmt.Errorf("range pattern expects matching ordered values: %w", err)
		}
		upper, err := compareOrdered(subject, end)
		if err != nil {
			return false, fmt.Errorf("range pattern expects matching ordered values: %w", err)
		}
		return lower >= 0 && upper <= 0, nil
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
