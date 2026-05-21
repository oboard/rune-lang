package interpreter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/ir"
)

type Value any

type Array struct {
	Elements []Value
}

type Struct struct {
	TypeName string
	Fields   map[string]Value
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
	case string:
		return strconv.Quote(v)
	case *Array:
		parts := make([]string, 0, len(v.Elements))
		for _, elem := range v.Elements {
			parts = append(parts, Format(elem))
		}
		return "[" + strings.Join(parts, ", ") + "]"
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
