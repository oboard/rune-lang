package interpreter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"sort"
	"strings"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
)

func (i *Interpreter) jsonStringify(value Value) (string, error) {
	return i.jsonValue(value, false)
}

func (i *Interpreter) jsonParse(text string, targetType checker.Type) (Value, error) {
	decoder := json.NewDecoder(bytes.NewBufferString(text))
	decoder.UseNumber()
	var raw any
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("parse JSON: multiple values")
		}
		return nil, fmt.Errorf("parse JSON: %w", err)
	}
	return i.jsonDecodedValue(raw, i.jsonZeroValue(targetType), targetType)
}

func (i *Interpreter) jsonDecodedValue(raw any, seed Value, typ checker.Type) (Value, error) {
	if inner, ok := jsonNullableType(typ); ok {
		if raw == nil {
			return NullValue, nil
		}
		return i.jsonDecodedValue(raw, seed, inner)
	}
	if elem, ok := checker.ArrayElement(typ); ok {
		values, ok := raw.([]any)
		if !ok {
			return nil, jsonTypeError(typ, raw)
		}
		var seedValues []Value
		if array, ok := seed.(*Array); ok {
			seedValues = array.Elements
		}
		out := &Array{Elements: make([]Value, 0, len(values))}
		for idx, value := range values {
			var itemSeed Value
			if idx < len(seedValues) {
				itemSeed = seedValues[idx]
			}
			decoded, err := i.jsonDecodedValue(value, itemSeed, elem)
			if err != nil {
				return nil, fmt.Errorf("JSON array item %d: %w", idx, err)
			}
			out.Elements = append(out.Elements, decoded)
		}
		return out, nil
	}
	if definition := i.types[jsonBaseTypeName(typ)]; definition != nil {
		object, ok := raw.(map[string]any)
		if !ok {
			return nil, jsonTypeError(typ, raw)
		}
		out := &Struct{TypeName: string(typ), Fields: map[string]Value{}, Order: make([]string, 0, len(definition.Fields))}
		if existing, ok := seed.(*Struct); ok {
			for name, value := range existing.Fields {
				out.Fields[name] = value
			}
		}
		for _, field := range definition.Fields {
			out.Order = append(out.Order, field.Name)
			if definition.JSONObject && field.JSONIgnore {
				continue
			}
			jsonName := field.Name
			if definition.JSONObject {
				jsonName = field.JSONName
			}
			value, exists := object[jsonName]
			if !exists {
				continue
			}
			decoded, err := i.jsonDecodedValue(value, out.Fields[field.Name], field.Type)
			if err != nil {
				return nil, fmt.Errorf("JSON field %q: %w", jsonName, err)
			}
			out.Fields[field.Name] = decoded
		}
		return out, nil
	}
	switch typ {
	case checker.String:
		value, ok := raw.(string)
		if !ok {
			return nil, jsonTypeError(typ, raw)
		}
		return value, nil
	case checker.Char:
		value, ok := raw.(string)
		if !ok || len([]rune(value)) != 1 {
			return nil, jsonTypeError(typ, raw)
		}
		return Char([]rune(value)[0]), nil
	case checker.Bool:
		value, ok := raw.(bool)
		if !ok {
			return nil, jsonTypeError(typ, raw)
		}
		return value, nil
	case checker.Int, checker.Int4, checker.Int8, checker.Int16, checker.Int64,
		checker.UInt, checker.UInt8, checker.UInt16, checker.UInt64:
		value, ok := raw.(json.Number)
		if !ok {
			return nil, jsonTypeError(typ, raw)
		}
		number, err := value.Int64()
		if err != nil {
			return nil, jsonTypeError(typ, raw)
		}
		return int(number), nil
	case checker.Double, checker.Float:
		value, ok := raw.(json.Number)
		if !ok {
			return nil, jsonTypeError(typ, raw)
		}
		number, err := value.Float64()
		if err != nil {
			return nil, jsonTypeError(typ, raw)
		}
		return number, nil
	case checker.BigInt:
		value, ok := raw.(json.Number)
		if !ok {
			return nil, jsonTypeError(typ, raw)
		}
		number, ok := new(big.Int).SetString(string(value), 10)
		if !ok {
			return nil, jsonTypeError(typ, raw)
		}
		return number, nil
	case checker.Object, checker.Unknown:
		return i.jsonDynamicValue(raw)
	}
	if enum := i.enums[jsonBaseTypeName(typ)]; enum != nil {
		value, ok := raw.(json.Number)
		if !ok {
			return nil, jsonTypeError(typ, raw)
		}
		number, err := value.Int64()
		if err != nil {
			return nil, jsonTypeError(typ, raw)
		}
		for _, member := range enum.Members {
			if member.HasValue && member.Value == int(number) {
				return EnumValue{TypeName: enum.Name, Name: member.Name, Value: member.Value}, nil
			}
		}
		return nil, fmt.Errorf("JSON value %d is not a member of %s", number, typ)
	}
	return nil, fmt.Errorf("cannot parse JSON into %s", typ)
}

func (i *Interpreter) jsonZeroValue(typ checker.Type) Value {
	if _, ok := jsonNullableType(typ); ok {
		return NullValue
	}
	if _, ok := checker.ArrayElement(typ); ok {
		return &Array{Elements: []Value{}}
	}
	if definition := i.types[jsonBaseTypeName(typ)]; definition != nil {
		out := &Struct{
			TypeName: string(typ),
			Fields:   make(map[string]Value, len(definition.Fields)),
			Order:    make([]string, 0, len(definition.Fields)),
		}
		for _, field := range definition.Fields {
			out.Order = append(out.Order, field.Name)
			out.Fields[field.Name] = i.jsonZeroValue(field.Type)
		}
		return out
	}
	if enum := i.enums[jsonBaseTypeName(typ)]; enum != nil {
		for _, member := range enum.Members {
			if member.HasValue && member.Value == 0 {
				return EnumValue{TypeName: enum.Name, Name: member.Name, Value: member.Value}
			}
		}
		if len(enum.Members) > 0 {
			member := enum.Members[0]
			return EnumValue{TypeName: enum.Name, Name: member.Name, Value: member.Value}
		}
	}
	switch typ {
	case checker.String:
		return ""
	case checker.Char:
		return Char(0)
	case checker.Bool:
		return false
	case checker.Int, checker.Int4, checker.Int8, checker.Int16, checker.Int64,
		checker.UInt, checker.UInt8, checker.UInt16, checker.UInt64:
		return 0
	case checker.Double, checker.Float:
		return float64(0)
	case checker.BigInt:
		return big.NewInt(0)
	case checker.Null:
		return NullValue
	case checker.Object, checker.Unknown:
		return &Struct{TypeName: "Object", Fields: map[string]Value{}}
	default:
		return nil
	}
}

func (i *Interpreter) jsonDynamicValue(raw any) (Value, error) {
	switch value := raw.(type) {
	case nil:
		return NullValue, nil
	case string, bool:
		return value, nil
	case json.Number:
		if integer, err := value.Int64(); err == nil {
			return int(integer), nil
		}
		number, err := value.Float64()
		if err != nil {
			return nil, err
		}
		return number, nil
	case []any:
		out := &Array{Elements: make([]Value, 0, len(value))}
		for _, item := range value {
			decoded, err := i.jsonDynamicValue(item)
			if err != nil {
				return nil, err
			}
			out.Elements = append(out.Elements, decoded)
		}
		return out, nil
	case map[string]any:
		out := &Struct{TypeName: "Object", Fields: map[string]Value{}}
		for name, item := range value {
			decoded, err := i.jsonDynamicValue(item)
			if err != nil {
				return nil, err
			}
			out.Fields[name] = decoded
			out.Order = append(out.Order, name)
		}
		sort.Strings(out.Order)
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported JSON value %T", raw)
	}
}

func jsonNullableType(typ checker.Type) (checker.Type, bool) {
	name := string(typ)
	if !strings.HasSuffix(name, "?") {
		return "", false
	}
	return checker.Type(strings.TrimSuffix(name, "?")), true
}

func jsonBaseTypeName(typ checker.Type) string {
	name := string(typ)
	if idx := strings.IndexByte(name, '['); idx >= 0 {
		return name[:idx]
	}
	return name
}

func jsonTypeError(typ checker.Type, raw any) error {
	return fmt.Errorf("expected %s, got %T", typ, raw)
}

func (i *Interpreter) jsonValue(value Value, inArray bool) (string, error) {
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
			part, err := i.jsonValue(elem, true)
			if err != nil {
				return "", err
			}
			parts = append(parts, part)
		}
		return "[" + strings.Join(parts, ",") + "]", nil
	case *Struct:
		parts := make([]string, 0, len(v.Fields))
		for _, fieldInfo := range i.jsonFields(v) {
			field, ok := v.Fields[fieldInfo.sourceName]
			if !ok || jsonOmitValue(field) {
				continue
			}
			key, err := json.Marshal(fieldInfo.jsonName)
			if err != nil {
				return "", err
			}
			part, err := i.jsonValue(field, false)
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

type interpreterJSONField struct {
	sourceName string
	jsonName   string
}

func (i *Interpreter) jsonFields(value *Struct) []interpreterJSONField {
	typ := i.types[value.TypeName]
	if typ == nil || !typ.JSONObject {
		names := jsonFieldOrder(value)
		out := make([]interpreterJSONField, 0, len(names))
		for _, name := range names {
			out = append(out, interpreterJSONField{sourceName: name, jsonName: name})
		}
		return out
	}
	out := make([]interpreterJSONField, 0, len(typ.Fields))
	for _, field := range typ.Fields {
		if field.JSONIgnore {
			continue
		}
		out = append(out, interpreterJSONField{sourceName: field.Name, jsonName: field.JSONName})
	}
	return out
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
