package interpreter

import (
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/oboard/rune-lang/internal/ir"
)

func jsonStringify(value Value) (string, error) {
	return jsonValue(value, false)
}

func jsonValue(value Value, inArray bool) (string, error) {
	switch v := value.(type) {
	case nil:
		return "null", nil
	case nullValue:
		return "null", nil
	case string:
		data, err := json.Marshal(v)
		return string(data), err
	case Char:
		data, err := json.Marshal(string(rune(v)))
		return string(data), err
	case int, float64, bool:
		data, err := json.Marshal(v)
		return string(data), err
	case EnumValue:
		data, err := json.Marshal(v.Value)
		return string(data), err
	case *big.Int:
		return v.String(), nil
	case *Array:
		parts := make([]string, 0, len(v.Elements))
		for _, elem := range v.Elements {
			part, err := jsonValue(elem, true)
			if err != nil {
				return "", err
			}
			parts = append(parts, part)
		}
		return "[" + strings.Join(parts, ",") + "]", nil
	case *Struct:
		parts := make([]string, 0, len(v.Fields))
		for _, name := range jsonFieldOrder(v) {
			field, ok := v.Fields[name]
			if !ok || jsonOmitValue(field) {
				continue
			}
			key, err := json.Marshal(name)
			if err != nil {
				return "", err
			}
			part, err := jsonValue(field, false)
			if err != nil {
				return "", err
			}
			parts = append(parts, string(key)+":"+part)
		}
		return "{" + strings.Join(parts, ",") + "}", nil
	case *Closure, *ir.Function:
		if inArray {
			return "null", nil
		}
		return "", fmt.Errorf("cannot stringify function value")
	default:
		return "", fmt.Errorf("cannot stringify %s", typeName(value))
	}
}

func jsonFieldOrder(value *Struct) []string {
	if len(value.Order) > 0 {
		seen := map[string]bool{}
		out := make([]string, 0, len(value.Fields))
		for _, name := range value.Order {
			if _, ok := value.Fields[name]; ok && !seen[name] {
				out = append(out, name)
				seen[name] = true
			}
		}
		if len(out) == len(value.Fields) {
			return out
		}
		for name := range value.Fields {
			if !seen[name] {
				out = append(out, name)
			}
		}
		return out
	}
	out := make([]string, 0, len(value.Fields))
	for name := range value.Fields {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func jsonOmitValue(value Value) bool {
	switch value.(type) {
	case *Closure, *ir.Function:
		return true
	default:
		return false
	}
}
