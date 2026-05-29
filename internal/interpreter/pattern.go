package interpreter

import (
	"fmt"
	"reflect"

	"github.com/oboard/rune-lang/internal/ir"
	"github.com/oboard/rune-lang/internal/lexer"
)

func (i *Interpreter) evalPatternBlock(block *ir.PatternBlock, subject Value, env *Env) (Value, error) {
	for _, branch := range block.Branches {
		branchEnv := NewEnv(env)
		matched, err := i.matchPattern(branch.Pattern, subject, branchEnv)
		if err != nil {
			return nil, err
		}
		if matched {
			return i.eval(branch.Expr, branchEnv)
		}
	}
	return nil, nil
}

func (i *Interpreter) matchPattern(pattern ir.Pattern, subject Value, env *Env) (bool, error) {
	switch p := pattern.(type) {
	case *ir.WildcardPattern:
		return true, nil
	case *ir.BindingPattern:
		env.Define(p.Name, subject)
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
	case *ir.OrPattern:
		for _, alternative := range p.Alternatives {
			branchEnv := NewEnv(env)
			matched, err := i.matchPattern(alternative, subject, branchEnv)
			if err != nil || matched {
				return matched, err
			}
		}
		return false, nil
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
	case *ir.MapPattern:
		return i.matchMapPattern(p, subject, env)
	case *ir.ObjectPattern:
		return i.matchObjectPattern(p, subject, env)
	default:
		return false, fmt.Errorf("unsupported pattern %T", pattern)
	}
}

func (i *Interpreter) matchMapPattern(pattern *ir.MapPattern, subject Value, env *Env) (bool, error) {
	value, ok := subject.(*Map)
	if !ok {
		return false, nil
	}
	for _, entry := range pattern.Entries {
		key, err := i.eval(entry.Key, env)
		if err != nil {
			return false, err
		}
		mapEntry, exists := value.Entries[valueKey(key)]
		if !exists {
			if entry.Optional {
				continue
			}
			return false, nil
		}
		matched, err := i.matchPattern(entry.Pattern, mapEntry.Value, env)
		if err != nil || !matched {
			return matched, err
		}
	}
	return true, nil
}

func (i *Interpreter) matchObjectPattern(pattern *ir.ObjectPattern, subject Value, env *Env) (bool, error) {
	value, ok := subject.(*Struct)
	if !ok {
		return false, nil
	}
	for _, field := range pattern.Fields {
		fieldValue, exists := value.Fields[field.Name]
		if !exists {
			if field.Optional {
				continue
			}
			return false, nil
		}
		matched, err := i.matchPattern(field.Pattern, fieldValue, env)
		if err != nil || !matched {
			return matched, err
		}
	}
	return true, nil
}
