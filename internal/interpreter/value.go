package interpreter

import (
	"fmt"
	"math/big"
	"regexp"
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

type Binary struct {
	Data []byte
}

type Buffer struct {
	Data []byte
}

type Reader struct {
	Data   []byte
	Offset int
	Nibble int
}

type Writer struct {
	Data   []byte
	Nibble int
}

type StringBuffer struct {
	Parts []string
}

type Regex struct {
	Source    string
	Flags     string
	LastIndex int
	expr      *regexp.Regexp
}

type EnumValue struct {
	TypeName string
	Name     string
	Value    int
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
	case *Binary:
		parts := make([]string, 0, len(v.Data))
		for _, value := range v.Data {
			parts = append(parts, strconv.Itoa(int(value)))
		}
		return "Binary [" + strings.Join(parts, ", ") + "]"
	case *Buffer:
		parts := make([]string, 0, len(v.Data))
		for _, value := range v.Data {
			parts = append(parts, strconv.Itoa(int(value)))
		}
		return "Buffer [" + strings.Join(parts, ", ") + "]"
	case *Reader:
		return fmt.Sprintf("Reader { position: %d, remaining: %d }", v.Offset, len(v.Data)-v.Offset)
	case *Writer:
		parts := make([]string, 0, len(v.Data))
		for _, value := range v.Data {
			parts = append(parts, strconv.Itoa(int(value)))
		}
		return "Writer [" + strings.Join(parts, ", ") + "]"
	case *StringBuffer:
		return strconv.Quote(strings.Join(v.Parts, ""))
	case *Regex:
		return "/" + v.Source + "/" + v.Flags
	case EnumValue:
		return v.TypeName + "." + v.Name
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
	case EnumValue:
		return Format(v)
	default:
		return Format(v)
	}
}
