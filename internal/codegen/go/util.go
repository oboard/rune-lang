package gocodegen

import (
	"fmt"
	"strings"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
)

func (g *generator) line(s string) {
	for i := 0; i < g.indent; i++ {
		g.buf.WriteByte('\t')
	}
	g.buf.WriteString(s)
	g.buf.WriteByte('\n')
}

func (g *generator) linef(format string, args ...any) {
	g.line(fmt.Sprintf(format, args...))
}

func goType(typ checker.Type) string {
	if _, ok := parseGoNullableType(string(typ)); ok {
		return "any"
	}
	if elem, ok := checker.ArrayElement(typ); ok {
		return "[]" + goType(elem)
	}
	if fields, ok := parseGoObjectType(string(typ)); ok {
		parts := make([]string, 0, len(fields))
		for _, field := range fields {
			parts = append(parts, fmt.Sprintf("%s %s", mangleIdent(field.name), goType(checker.Type(field.typ))))
		}
		return "struct{" + strings.Join(parts, "; ") + "}"
	}
	if params, ret, ok := parseGoFuncType(string(typ)); ok {
		goParams := make([]string, 0, len(params))
		for _, param := range params {
			goParams = append(goParams, goType(checker.Type(param)))
		}
		if ret == string(checker.Void) {
			return "func(" + strings.Join(goParams, ", ") + ")"
		}
		return "func(" + strings.Join(goParams, ", ") + ") " + goType(checker.Type(ret))
	}
	switch typ {
	case checker.Int:
		return "int"
	case checker.Double:
		return "float64"
	case checker.BigInt:
		return "*big.Int"
	case checker.String:
		return "string"
	case checker.Bool:
		return "bool"
	case checker.Null:
		return "any"
	case checker.Object:
		return "any"
	case checker.Never:
		return "struct{}"
	case checker.Symbol:
		return "runeSymbol"
	case checker.HTMLElement:
		return "any"
	case checker.Unknown:
		return "any"
	default:
		return mangleIdent(string(typ))
	}
}

func parseGoNullableType(name string) (string, bool) {
	if !strings.HasSuffix(name, "?") || name == "?" {
		return "", false
	}
	return strings.TrimSuffix(name, "?"), true
}

type goObjectField struct {
	name string
	typ  string
}

func parseGoObjectType(name string) ([]goObjectField, bool) {
	if !strings.HasPrefix(name, "{") || !strings.HasSuffix(name, "}") {
		return nil, false
	}
	inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(name, "{"), "}"))
	if inner == "" {
		return nil, true
	}
	parts := splitGoTypeList(inner)
	fields := make([]goObjectField, 0, len(parts))
	for _, part := range parts {
		fieldName, fieldType, ok := splitGoObjectField(part)
		if !ok {
			return nil, false
		}
		fields = append(fields, goObjectField{name: fieldName, typ: fieldType})
	}
	return fields, true
}

func splitGoObjectField(src string) (string, string, bool) {
	depth := 0
	for i, ch := range src {
		switch ch {
		case '[', '{', '(':
			depth++
		case ']', '}', ')':
			depth--
		case ':':
			if depth == 0 {
				return strings.TrimSpace(src[:i]), strings.TrimSpace(src[i+1:]), true
			}
		}
	}
	return "", "", false
}

func parseGoFuncType(name string) ([]string, string, bool) {
	if !strings.HasPrefix(name, "Func[") || !strings.HasSuffix(name, "]") {
		return nil, "", false
	}
	parts := splitGoTypeList(strings.TrimSuffix(strings.TrimPrefix(name, "Func["), "]"))
	if len(parts) == 0 {
		return nil, "", false
	}
	return parts[:len(parts)-1], parts[len(parts)-1], true
}

func splitGoTypeList(src string) []string {
	var out []string
	depth := 0
	start := 0
	for i, ch := range src {
		switch ch {
		case '[', '{', '(':
			depth++
		case ']', '}', ')':
			depth--
		case ',':
			if depth == 0 {
				out = append(out, strings.TrimSpace(src[start:i]))
				start = i + 1
			}
		}
	}
	out = append(out, strings.TrimSpace(src[start:]))
	return out
}

func zeroValue(typ checker.Type) string {
	if _, ok := parseGoNullableType(string(typ)); ok {
		return "any(nil)"
	}
	if _, ok := checker.ArrayElement(typ); ok {
		return "nil"
	}
	if _, _, ok := parseGoFuncType(string(typ)); ok {
		return "nil"
	}
	switch typ {
	case checker.Int:
		return "0"
	case checker.Double:
		return "0"
	case checker.BigInt:
		return "runeBigInt(\"0\")"
	case checker.String:
		return `""`
	case checker.Bool:
		return "false"
	case checker.Null:
		return "any(nil)"
	case checker.HTMLElement:
		return "nil"
	default:
		return fmt.Sprintf("%s{}", goType(typ))
	}
}

func hasMain(file *ir.File) bool {
	for _, fn := range file.Functions {
		if fn.Name == "main" {
			return true
		}
	}
	return false
}

func mangleIdent(name string) string {
	return "__" + name
}
