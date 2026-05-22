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
	if elem, ok := checker.ArrayElement(typ); ok {
		return "[]" + goType(elem)
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
	case checker.String:
		return "string"
	case checker.Bool:
		return "bool"
	case checker.Unknown:
		return "any"
	default:
		return mangleIdent(string(typ))
	}
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
		case '[':
			depth++
		case ']':
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
	if _, ok := checker.ArrayElement(typ); ok {
		return "nil"
	}
	switch typ {
	case checker.Int:
		return "0"
	case checker.String:
		return `""`
	case checker.Bool:
		return "false"
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
