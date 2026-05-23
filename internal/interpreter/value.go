package interpreter

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/ir"
)

type Value any

type nullValue struct{}

var NullValue Value = nullValue{}

type Array struct {
	Elements []Value
}

type Map struct {
	Entries map[string]mapEntry
}

type mapEntry struct {
	Key   Value
	Value Value
}

type Set struct {
	Entries map[string]Value
}

type Struct struct {
	TypeName string
	Fields   map[string]Value
	Order    []string
}

type Closure struct {
	Params []string
	Body   ir.Expr
	Env    *Env
}

func Format(value Value) string {
	switch v := value.(type) {
	case nil:
		return "void"
	case nullValue:
		return "null"
	case string:
		return strconv.Quote(v)
	case *big.Int:
		return v.String() + "n"
	case *Array:
		parts := make([]string, 0, len(v.Elements))
		for _, elem := range v.Elements {
			parts = append(parts, Format(elem))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case *Map:
		parts := make([]string, 0, len(v.Entries))
		for _, entry := range v.Entries {
			parts = append(parts, Format(entry.Key)+": "+Format(entry.Value))
		}
		return "Map { " + strings.Join(parts, ", ") + " }"
	case *Set:
		parts := make([]string, 0, len(v.Entries))
		for _, value := range v.Entries {
			parts = append(parts, Format(value))
		}
		return "Set { " + strings.Join(parts, ", ") + " }"
	case *Struct:
		parts := make([]string, 0, len(v.Fields))
		for name, value := range v.Fields {
			parts = append(parts, name+": "+Format(value))
		}
		return v.TypeName + " { " + strings.Join(parts, ", ") + " }"
	default:
		return fmt.Sprint(value)
	}
}

func printValue(value Value) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	default:
		return Format(v)
	}
}
