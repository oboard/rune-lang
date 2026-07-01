package interpreter

import (
	"fmt"
	"math/big"
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
		if p.Constant {
			if enumValue, ok := subject.(EnumValue); ok && p.Type != "" {
				return enumValue.Name == p.Name, nil
			}
			value, ok := env.Get(p.Name)
			if !ok {
				return false, fmt.Errorf("undefined constant %q", p.Name)
			}
			return reflect.DeepEqual(subject, value) || numericPatternEqual(subject, value), nil
		}
		env.Define(p.Name, subject)
		return true, nil
	case *ir.LiteralPattern:
		value, err := i.eval(p.Value, env)
		if err != nil {
			return false, err
		}
		return reflect.DeepEqual(subject, value) || numericPatternEqual(subject, value), nil
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
		if p.Start != nil {
			start, err := i.eval(p.Start, env)
			if err != nil {
				return false, err
			}
			lower, err := compareOrdered(subject, start)
			if err != nil {
				return false, fmt.Errorf("range pattern expects matching ordered values: %w", err)
			}
			if lower < 0 {
				return false, nil
			}
		}
		if p.End != nil {
			end, err := i.eval(p.End, env)
			if err != nil {
				return false, err
			}
			upper, err := compareOrdered(subject, end)
			if err != nil {
				return false, fmt.Errorf("range pattern expects matching ordered values: %w", err)
			}
			if p.Inclusive {
				return upper <= 0, nil
			}
			return upper < 0, nil
		}
		return true, nil
	case *ir.OrPattern:
		for _, alternative := range p.Alternatives {
			branchEnv := NewEnv(env)
			matched, err := i.matchPattern(alternative, subject, branchEnv)
			if err != nil {
				return false, err
			}
			if matched {
				for name, value := range branchEnv.values {
					env.Define(name, value)
				}
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
	case *ir.ArrayPattern:
		return i.matchArrayPattern(p, subject, env)
	case *ir.AsPattern:
		matched, err := i.matchPattern(p.Pattern, subject, env)
		if err != nil || !matched {
			return matched, err
		}
		env.Define(p.Name, subject)
		return true, nil
	case *ir.MapPattern:
		return i.matchMapPattern(p, subject, env)
	case *ir.ObjectPattern:
		return i.matchObjectPattern(p, subject, env)
	case *ir.ConstructorPattern:
		value, ok := subject.(EnumValue)
		if ok {
			if value.Name != p.Name {
				return false, nil
			}
			if p.Rest {
				if len(value.Payload) < len(p.Args) {
					return false, nil
				}
			} else if len(value.Payload) != len(p.Args) {
				return false, nil
			}
			for idx, arg := range p.Args {
				matched, err := i.matchPattern(arg, value.Payload[idx], env)
				if err != nil || !matched {
					return matched, err
				}
			}
			return true, nil
		}
		if matched, handled, err := i.matchJSONConstructorPattern(p, subject, env); handled {
			return matched, err
		}
		return false, nil
	default:
		return false, fmt.Errorf("unsupported pattern %T", pattern)
	}
}

func (i *Interpreter) matchJSONConstructorPattern(pattern *ir.ConstructorPattern, subject Value, env *Env) (bool, bool, error) {
	var payload Value
	switch pattern.Name {
	case "Array":
		if _, ok := subject.(*Array); !ok {
			return false, true, nil
		}
		payload = subject
	case "Object":
		if _, ok := subject.(*Struct); !ok {
			return false, true, nil
		}
		payload = subject
	case "String":
		value, ok := subject.(string)
		if !ok {
			return false, true, nil
		}
		payload = value
	case "Bool":
		value, ok := subject.(bool)
		if !ok {
			return false, true, nil
		}
		payload = value
	case "Number":
		if !isJSONNumberValue(subject) {
			return false, true, nil
		}
		payload = subject
	case "Null":
		return isNullValue(subject), true, nil
	default:
		return false, false, nil
	}
	if pattern.Rest {
		if len(pattern.Args) > 1 {
			return false, true, nil
		}
	} else if len(pattern.Args) != 1 {
		return false, true, nil
	}
	if len(pattern.Args) == 0 {
		return true, true, nil
	}
	matched, err := i.matchPattern(pattern.Args[0], payload, env)
	return matched, true, err
}

func isJSONNumberValue(value Value) bool {
	switch value.(type) {
	case int, int8, int16, int64, uint, uint8, uint16, uint64, float32, float64, *big.Int:
		return true
	default:
		return false
	}
}

func (i *Interpreter) matchArrayPattern(pattern *ir.ArrayPattern, subject Value, env *Env) (bool, error) {
	if irArrayPatternHasBits(pattern) {
		return i.matchBitArrayPattern(pattern, subject, env)
	}
	elems, rest, ok := sequencePatternValues(subject)
	if !ok {
		return false, nil
	}
	parts, restIndex, err := i.expandArrayPatternParts(pattern, env)
	if err != nil {
		return false, err
	}
	required := len(parts)
	if pattern.RestIndex >= 0 {
		if len(elems) < required {
			return false, nil
		}
	} else if len(elems) != required {
		return false, nil
	}
	for idx, part := range parts {
		valueIndex := idx
		if restIndex >= 0 && idx >= restIndex {
			valueIndex = len(elems) - (len(parts) - idx)
		}
		if part.Pattern != nil {
			matched, err := i.matchPattern(part.Pattern, elems[valueIndex], env)
			if err != nil || !matched {
				return matched, err
			}
			continue
		}
		if !reflect.DeepEqual(elems[valueIndex], part.Value) && !numericPatternEqual(elems[valueIndex], part.Value) {
			return false, nil
		}
	}
	if pattern.RestBinding != "" {
		start := restIndex
		end := len(elems) - (len(parts) - restIndex)
		env.Define(pattern.RestBinding, rest(start, end))
	}
	return true, nil
}

type arrayPatternPart struct {
	Pattern ir.Pattern
	Value   Value
}

func (i *Interpreter) expandArrayPatternParts(pattern *ir.ArrayPattern, env *Env) ([]arrayPatternPart, int, error) {
	parts := []arrayPatternPart{}
	restIndex := -1
	for idx, elem := range pattern.Elements {
		if pattern.RestIndex == idx {
			restIndex = len(parts)
		}
		if spread, ok := elem.(*ir.SequenceSpreadPattern); ok {
			value, err := i.eval(spread.Value, env)
			if err != nil {
				return nil, -1, err
			}
			values, _, ok := sequencePatternValues(value)
			if !ok {
				return nil, -1, fmt.Errorf("array pattern spread expects Array, String, or Bytes")
			}
			for _, value := range values {
				parts = append(parts, arrayPatternPart{Value: value})
			}
			continue
		}
		parts = append(parts, arrayPatternPart{Pattern: elem})
	}
	if pattern.RestIndex == len(pattern.Elements) {
		restIndex = len(parts)
	}
	return parts, restIndex, nil
}

func irArrayPatternHasBits(pattern *ir.ArrayPattern) bool {
	for _, elem := range pattern.Elements {
		if _, ok := elem.(*ir.BitPattern); ok {
			return true
		}
	}
	return false
}

func (i *Interpreter) matchBitArrayPattern(pattern *ir.ArrayPattern, subject Value, env *Env) (bool, error) {
	data, rest, ok := bitPatternBytes(subject)
	if !ok {
		return false, nil
	}
	requiredBits := 0
	for idx, elem := range pattern.Elements {
		bit, ok := elem.(*ir.BitPattern)
		if !ok {
			return false, nil
		}
		if pattern.RestIndex < 0 || idx < pattern.RestIndex {
			requiredBits += bit.Width
		}
	}
	if pattern.RestIndex >= 0 {
		if len(data)*8 < requiredBits {
			return false, nil
		}
	} else if len(data)*8 != requiredBits {
		return false, nil
	}
	bitOffset := 0
	for idx, elem := range pattern.Elements {
		bit := elem.(*ir.BitPattern)
		if pattern.RestIndex >= 0 && idx >= pattern.RestIndex {
			bitOffset = len(data)*8 - bitTailWidth(pattern, idx)
		}
		raw := readBits(data, bitOffset, bit.Width, bit.Endian == "le")
		value := bitPatternValue(raw, bit.Width, bit.Signed)
		matched, err := i.matchPattern(bit.Value, value, env)
		if err != nil || !matched {
			return matched, err
		}
		bitOffset += bit.Width
	}
	if pattern.RestBinding != "" {
		if requiredBits%8 != 0 {
			return false, nil
		}
		env.Define(pattern.RestBinding, rest(requiredBits/8, len(data)))
	}
	return true, nil
}

func bitTailWidth(pattern *ir.ArrayPattern, start int) int {
	width := 0
	for idx := start; idx < len(pattern.Elements); idx++ {
		if bit, ok := pattern.Elements[idx].(*ir.BitPattern); ok {
			width += bit.Width
		}
	}
	return width
}

func bitPatternBytes(subject Value) ([]byte, func(int, int) Value, bool) {
	switch v := subject.(type) {
	case *Bytes:
		data := append([]byte(nil), v.Data...)
		return data, func(start int, end int) Value {
			return &Bytes{Data: append([]byte(nil), v.Data[start:end]...)}
		}, true
	case *Array:
		data := make([]byte, 0, len(v.Elements))
		for _, elem := range v.Elements {
			switch n := elem.(type) {
			case uint8:
				data = append(data, byte(n))
			case int:
				if n < 0 || n > 255 {
					return nil, nil, false
				}
				data = append(data, byte(n))
			default:
				return nil, nil, false
			}
		}
		return data, func(start int, end int) Value {
			return &Array{Elements: append([]Value(nil), v.Elements[start:end]...)}
		}, true
	default:
		return nil, nil, false
	}
}

func readBits(data []byte, offset int, width int, littleEndian bool) uint64 {
	if littleEndian && offset%8 == 0 && width%8 == 0 {
		var out uint64
		start := offset / 8
		for idx := 0; idx < width/8; idx++ {
			out |= uint64(data[start+idx]) << uint(8*idx)
		}
		return out
	}
	var out uint64
	for idx := 0; idx < width; idx++ {
		bitIndex := offset + idx
		bit := (data[bitIndex/8] >> uint(7-bitIndex%8)) & 1
		out = (out << 1) | uint64(bit)
	}
	return out
}

func bitPatternValue(raw uint64, width int, signed bool) Value {
	if signed {
		if width == 64 {
			return int64(raw)
		}
		sign := uint64(1) << uint(width-1)
		if raw&sign != 0 {
			raw |= ^uint64(0) << uint(width)
		}
		if width > 32 {
			return int64(raw)
		}
		return int(int64(raw))
	}
	if width > 32 {
		return raw
	}
	return uint(raw)
}

func numericPatternEqual(subject Value, value Value) bool {
	s, ok := numericPatternUint64(subject)
	if !ok {
		return false
	}
	v, ok := numericPatternUint64(value)
	return ok && s == v
}

func numericPatternUint64(value Value) (uint64, bool) {
	switch n := value.(type) {
	case int:
		if n < 0 {
			return uint64(n), true
		}
		return uint64(n), true
	case int64:
		return uint64(n), true
	case uint:
		return uint64(n), true
	case uint8:
		return uint64(n), true
	case uint16:
		return uint64(n), true
	case uint64:
		return n, true
	default:
		return 0, false
	}
}

func sequencePatternValues(subject Value) ([]Value, func(int, int) Value, bool) {
	switch v := subject.(type) {
	case *Array:
		elems := append([]Value(nil), v.Elements...)
		return elems, func(start int, end int) Value {
			return &Array{Elements: append([]Value(nil), v.Elements[start:end]...)}
		}, true
	case string:
		runes := []rune(v)
		elems := make([]Value, 0, len(runes))
		for _, ch := range runes {
			elems = append(elems, Char(ch))
		}
		return elems, func(start int, end int) Value {
			return string(runes[start:end])
		}, true
	case *Bytes:
		elems := make([]Value, 0, len(v.Data))
		for _, b := range v.Data {
			elems = append(elems, uint8(b))
		}
		return elems, func(start int, end int) Value {
			return &Bytes{Data: append([]byte(nil), v.Data[start:end]...)}
		}, true
	default:
		return nil, nil, false
	}
}

func (i *Interpreter) matchMapPattern(pattern *ir.MapPattern, subject Value, env *Env) (bool, error) {
	if pattern.Access == "get" {
		return i.matchMapLikePattern(pattern, subject, env)
	}
	value, ok := subject.(*Map)
	if !ok {
		if object, ok := subject.(*Struct); ok {
			return i.matchObjectMapPattern(pattern, object, env)
		}
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
				matched, err := i.matchPattern(entry.Pattern, NullValue, env)
				if err != nil || !matched {
					return matched, err
				}
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

func (i *Interpreter) matchMapLikePattern(pattern *ir.MapPattern, subject Value, env *Env) (bool, error) {
	receiver, ok := subject.(*Struct)
	if !ok {
		return false, nil
	}
	for _, entry := range pattern.Entries {
		key, err := i.eval(entry.Key, env)
		if err != nil {
			return false, err
		}
		value, err := i.callStructMethodValues(receiver, "get", []Value{key})
		if err != nil {
			return false, err
		}
		if isNullValue(value) && !entry.Optional {
			return false, nil
		}
		matched, err := i.matchPattern(entry.Pattern, value, env)
		if err != nil || !matched {
			return matched, err
		}
	}
	return true, nil
}

func (i *Interpreter) matchObjectMapPattern(pattern *ir.MapPattern, object *Struct, env *Env) (bool, error) {
	for _, entry := range pattern.Entries {
		key, err := i.eval(entry.Key, env)
		if err != nil {
			return false, err
		}
		name, ok := key.(string)
		if !ok {
			return false, nil
		}
		fieldValue, exists := object.Fields[name]
		if !exists {
			if entry.Optional {
				matched, err := i.matchPattern(entry.Pattern, NullValue, env)
				if err != nil || !matched {
					return matched, err
				}
				continue
			}
			return false, nil
		}
		matched, err := i.matchPattern(entry.Pattern, fieldValue, env)
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
