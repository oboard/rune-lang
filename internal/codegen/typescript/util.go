package tscodegen

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
)

type generator struct {
	buf       bytes.Buffer
	file      *ir.File
	indent    int
	temp      int
	thisNames []string
	signals   []map[string]checker.Type
	reactives []map[string]checker.Type
}

func (g *generator) line(s string) {
	for i := 0; i < g.indent; i++ {
		g.buf.WriteString("  ")
	}
	g.buf.WriteString(s)
	g.buf.WriteByte('\n')
}

func (g *generator) linef(format string, args ...any) {
	g.line(fmt.Sprintf(format, args...))
}

func (g *generator) lineExpr(prefix string, expr string, suffix string) {
	lines := strings.Split(expr, "\n")
	if len(lines) == 0 {
		g.line(prefix + suffix)
		return
	}
	if len(lines) == 1 {
		g.line(prefix + lines[0] + suffix)
		return
	}
	g.line(prefix + lines[0])
	for _, line := range lines[1 : len(lines)-1] {
		g.line(line)
	}
	g.line(lines[len(lines)-1] + suffix)
}

func (g *generator) nextTemp(prefix string) string {
	g.temp++
	return fmt.Sprintf("%s%d", prefix, g.temp)
}

func tsType(typ checker.Type) string {
	if inner, ok := parseTSNullableType(string(typ)); ok {
		return tsType(checker.Type(inner)) + " | null"
	}
	if base, args, ok := parseTSGenericType(string(typ)); ok {
		switch base {
		case "ReadonlyArray":
			return "ReadonlyArray<" + tsType(checker.Type(args[0])) + ">"
		case "Tuple":
			return "[" + tsTypeList(args) + "]"
		case "ReadonlyTuple":
			return "readonly [" + tsTypeList(args) + "]"
		case "Map":
			return "Map<" + tsTypeList(args) + ">"
		case "Set":
			return "Set<" + tsType(checker.Type(args[0])) + ">"
		case "WeakMap":
			return "WeakMap<" + tsTypeList(args) + ">"
		case "WeakSet":
			return "WeakSet<" + tsType(checker.Type(args[0])) + ">"
		case "Record":
			return "Record<" + tsTypeList(args) + ">"
		}
	}
	if elem, ok := checker.ArrayElement(typ); ok {
		return tsType(elem) + "[]"
	}
	if fields, ok := parseTSObjectType(string(typ)); ok {
		parts := make([]string, 0, len(fields))
		for _, field := range fields {
			parts = append(parts, fmt.Sprintf("%s: %s", tsPropertyName(field.name), tsType(checker.Type(field.typ))))
		}
		return "{ " + strings.Join(parts, "; ") + " }"
	}
	if params, ret, ok := parseTSFuncType(string(typ)); ok {
		tsParams := make([]string, 0, len(params))
		for i, param := range params {
			tsParams = append(tsParams, fmt.Sprintf("arg%d: %s", i, tsType(checker.Type(param))))
		}
		return "(" + strings.Join(tsParams, ", ") + ") => " + tsType(checker.Type(ret))
	}
	switch typ {
	case checker.Int:
		return "number"
	case checker.Double:
		return "number"
	case checker.BigInt:
		return "bigint"
	case checker.String:
		return "string"
	case checker.Bool:
		return "boolean"
	case checker.Null:
		return "null"
	case checker.Object:
		return "object"
	case checker.Never:
		return "never"
	case checker.Symbol:
		return "symbol"
	case checker.Void:
		return "void"
	case checker.HTMLElement:
		return "HTMLElement"
	case checker.Unknown:
		return "any"
	default:
		return mangleIdent(string(typ))
	}
}

func tsTypeList(args []string) string {
	out := make([]string, 0, len(args))
	for _, arg := range args {
		out = append(out, tsType(checker.Type(arg)))
	}
	return strings.Join(out, ", ")
}

func parseTSGenericType(name string) (string, []string, bool) {
	idx := strings.IndexByte(name, '[')
	if idx <= 0 || !strings.HasSuffix(name, "]") {
		return "", nil, false
	}
	base := name[:idx]
	args := splitTSTypeList(strings.TrimSuffix(name[idx+1:], "]"))
	switch base {
	case "ReadonlyArray", "Set", "WeakSet":
		return base, args, len(args) == 1
	case "Tuple", "ReadonlyTuple":
		return base, args, len(args) > 0
	case "Map", "WeakMap", "Record":
		return base, args, len(args) == 2
	default:
		return "", nil, false
	}
}

func parseTSNullableType(name string) (string, bool) {
	if !strings.HasSuffix(name, "?") || name == "?" {
		return "", false
	}
	return strings.TrimSuffix(name, "?"), true
}

type tsObjectField struct {
	name string
	typ  string
}

func parseTSObjectType(name string) ([]tsObjectField, bool) {
	if !strings.HasPrefix(name, "{") || !strings.HasSuffix(name, "}") {
		return nil, false
	}
	inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(name, "{"), "}"))
	if inner == "" {
		return nil, true
	}
	parts := splitTSTypeList(inner)
	fields := make([]tsObjectField, 0, len(parts))
	for _, part := range parts {
		fieldName, fieldType, ok := splitTSObjectField(part)
		if !ok {
			return nil, false
		}
		fields = append(fields, tsObjectField{name: fieldName, typ: fieldType})
	}
	return fields, true
}

func splitTSObjectField(src string) (string, string, bool) {
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

func parseTSFuncType(name string) ([]string, string, bool) {
	if !strings.HasPrefix(name, "Func[") || !strings.HasSuffix(name, "]") {
		return nil, "", false
	}
	parts := splitTSTypeList(strings.TrimSuffix(strings.TrimPrefix(name, "Func["), "]"))
	if len(parts) == 0 {
		return nil, "", false
	}
	return parts[:len(parts)-1], parts[len(parts)-1], true
}

func splitTSTypeList(src string) []string {
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
	if _, ok := parseTSNullableType(string(typ)); ok {
		return "null"
	}
	if _, ok := checker.ArrayElement(typ); ok {
		return "[]"
	}
	if _, _, ok := parseTSFuncType(string(typ)); ok {
		return "undefined as any"
	}
	switch typ {
	case checker.Int:
		return "0"
	case checker.Double:
		return "0"
	case checker.BigInt:
		return "0n"
	case checker.String:
		return `""`
	case checker.Bool:
		return "false"
	case checker.Null:
		return "null"
	case checker.Never:
		return "undefined as never"
	case checker.Void:
		return "undefined"
	case checker.HTMLElement:
		return "document.createElement(\"div\")"
	default:
		return "{} as " + tsType(typ)
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

var unsafeIdentChars = regexp.MustCompile(`[^A-Za-z0-9_$]`)

func mangleIdent(name string) string {
	if name == "" {
		return "__"
	}
	name = unsafeIdentChars.ReplaceAllString(name, "_")
	return "__" + name
}

var tsSafePropertyIdent = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*$`)

var tsReservedPropertyNames = map[string]bool{
	"await":      true,
	"break":      true,
	"case":       true,
	"catch":      true,
	"class":      true,
	"const":      true,
	"continue":   true,
	"debugger":   true,
	"default":    true,
	"delete":     true,
	"do":         true,
	"else":       true,
	"enum":       true,
	"export":     true,
	"extends":    true,
	"false":      true,
	"finally":    true,
	"for":        true,
	"function":   true,
	"if":         true,
	"implements": true,
	"import":     true,
	"in":         true,
	"instanceof": true,
	"interface":  true,
	"let":        true,
	"new":        true,
	"null":       true,
	"package":    true,
	"private":    true,
	"protected":  true,
	"public":     true,
	"return":     true,
	"static":     true,
	"super":      true,
	"switch":     true,
	"this":       true,
	"throw":      true,
	"true":       true,
	"try":        true,
	"typeof":     true,
	"var":        true,
	"void":       true,
	"while":      true,
	"with":       true,
	"yield":      true,
}

func tsPropertyName(name string) string {
	if tsCanUseBareProperty(name) {
		return name
	}
	return strconv.Quote(name)
}

func tsPropertyAccess(receiver string, name string) string {
	if tsCanUseBareProperty(name) {
		return receiver + "." + name
	}
	return receiver + "[" + strconv.Quote(name) + "]"
}

func tsCanUseBareProperty(name string) bool {
	return tsSafePropertyIdent.MatchString(name) && !tsReservedPropertyNames[name]
}

func mangleMethod(typeName string, methodName string) string {
	return mangleIdent(typeName + "_" + methodName)
}

func baseTypeName(typ checker.Type) string {
	name := string(typ)
	if strings.HasPrefix(name, "{") && strings.HasSuffix(name, "}") {
		return name
	}
	if i := strings.IndexByte(name, '['); i >= 0 {
		return name[:i]
	}
	return name
}
