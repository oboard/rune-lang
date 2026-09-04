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
	if strings.HasPrefix(string(typ), "&") {
		return "any"
	}
	if _, ok := parseGoNullableType(string(typ)); ok {
		return "any"
	}
	if base, args, ok := parseGoGenericType(string(typ)); ok {
		switch base {
		case "Result":
			return "runeResult[" + goType(checker.Type(args[0])) + ", " + goType(checker.Type(args[1])) + "]"
		case "Task":
			return "runeTask[" + goType(checker.Type(args[0])) + "]"
		case "Iter":
			return "runeIter[" + goType(checker.Type(args[0])) + "]"
		case "ReadonlyArray":
			return "[]" + goType(checker.Type(args[0]))
		case "Tuple", "ReadonlyTuple":
			fields := make([]string, 0, len(args))
			for idx, arg := range args {
				fields = append(fields, fmt.Sprintf("F%d %s", idx, goType(checker.Type(arg))))
			}
			return "struct{" + strings.Join(fields, "; ") + "}"
		case "Map", "WeakMap", "Record":
			return "map[" + goType(checker.Type(args[0])) + "]" + goType(checker.Type(args[1]))
		case "Set", "WeakSet":
			return "map[" + goType(checker.Type(args[0])) + "]struct{}"
		}
	}
	if elem, ok := checker.ArrayElement(typ); ok {
		return "[]" + goType(elem)
	}
		if fields, ok := parseGoObjectType(string(typ)); ok {
			fields = sortedGoObjectFields(fields)
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
	case checker.Int4, checker.Int8:
		return "int8"
	case checker.Int16:
		return "int16"
	case checker.Int64:
		return "int64"
	case checker.Double:
		return "float64"
	case checker.Float:
		return "float32"
	case checker.BigInt:
		return "*big.Int"
	case checker.UInt:
		return "uint"
	case checker.UInt8:
		return "uint8"
	case checker.UInt16:
		return "uint16"
	case checker.UInt64:
		return "uint64"
	case checker.String:
		return "string"
	case checker.Char:
		return "rune"
	case checker.Bool:
		return "bool"
	case checker.Null:
		return "any"
	case checker.Object:
		return "any"
	case checker.Bytes:
		return "*runeBytes"
	case checker.Buffer:
		return "*runeBuffer"
	case checker.Reader:
		return "*runeReader"
	case checker.Writer:
		return "*runeWriter"
	case checker.StringBuffer:
		return "*runeStringBuffer"
	case checker.FileStat:
		return "*runeFileStat"
	case checker.TCPConnection:
		return "*runeTCPConnection"
	case checker.TCPListener:
		return "*runeTCPListener"
	case checker.Data:
		return "[]byte"
	case checker.Error:
		return "*runeError"
	case checker.Never:
		return "struct{}"
	case checker.Symbol:
		return "runeSymbol"
	case checker.Regex:
		return "*runeRegex"
	case checker.Void:
		return "struct{}"
	case checker.HTMLElement:
		return "any"
	case checker.WebComponent:
		return "any"
	case checker.Unknown:
		return "any"
	default:
		return mangleIdent(string(typ))
	}
}

func goGenerics(names []string, constraints map[string]string) string {
	if len(names) == 0 {
		return ""
	}
	parts := make([]string, 0, len(names))
	for _, name := range names {
		parts = append(parts, mangleIdent(name)+" "+goGenericConstraint(constraints[name]))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func goGenericConstraint(name string) string {
	switch name {
	case "Add", "Sub", "Mul", "Div", "Number":
		return "runeNumber"
	default:
		return "any"
	}
}

func parseGoGenericType(name string) (string, []string, bool) {
	idx := strings.IndexByte(name, '[')
	if idx <= 0 || !strings.HasSuffix(name, "]") {
		return "", nil, false
	}
	base := name[:idx]
	args := splitGoTypeList(strings.TrimSuffix(name[idx+1:], "]"))
	switch base {
	case "ReadonlyArray", "Set", "WeakSet", "Task", "Iter":
		return base, args, len(args) == 1
	case "Tuple", "ReadonlyTuple":
		return base, args, len(args) > 0
	case "Map", "WeakMap", "Record", "Result":
		return base, args, len(args) == 2
	default:
		return "", nil, false
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
	if base, args, ok := parseGoGenericType(string(typ)); ok {
		switch base {
		case "Result":
			return fmt.Sprintf("%s{}", goType(typ))
		case "Task":
			return "nil"
		default:
			_ = args
		}
	}
	switch typ {
	case checker.Int, checker.Int4, checker.Int8, checker.Int16, checker.Int64, checker.UInt, checker.UInt8, checker.UInt16, checker.UInt64:
		return "0"
	case checker.Double, checker.Float:
		return "0"
	case checker.BigInt:
		return "runeBigInt(\"0\")"
	case checker.String:
		return `""`
	case checker.Char:
		return "rune(0)"
	case checker.Bool:
		return "false"
	case checker.Regex:
		return `newRuneRegex("", "")`
	case checker.Symbol:
		return "runeSymbol{}"
	case checker.Bytes:
		return "newRuneBytes(0)"
	case checker.Buffer:
		return "newRuneBuffer()"
	case checker.Reader:
		return "newRuneReader(newRuneBytes(0))"
	case checker.Writer:
		return "newRuneWriter()"
	case checker.StringBuffer:
		return "newRuneStringBuffer()"
	case checker.FileStat:
		return "&runeFileStat{}"
	case checker.TCPConnection:
		return "nil"
	case checker.TCPListener:
		return "nil"
	case checker.Data:
		return "nil"
	case checker.Error:
		return "nil"
	case checker.Null:
		return "any(nil)"
	case checker.Void:
		return "struct{}{}"
	case checker.HTMLElement:
		return "nil"
	case checker.WebComponent:
		return "nil"
	default:
		return fmt.Sprintf("%s{}", goType(typ))
	}
}

func (g *generator) zeroValue(typ checker.Type) string {
	if enum := g.enumForType(typ); enum != nil && !enumHasPayload(enum) {
		return fmt.Sprintf("%s(0)", goType(typ))
	}
	return zeroValue(typ)
}

func hasMain(file *ir.File) bool {
	return mainFunction(file) != nil
}

func mainFunction(file *ir.File) *ir.Function {
	for _, fn := range file.Functions {
		if fn.Name == "main" {
			return fn
		}
	}
	return nil
}

func mangleIdent(name string) string {
	var b strings.Builder
	b.WriteString("__")
	for _, ch := range name {
		if isSafeMangledIdentRune(ch) {
			b.WriteRune(ch)
			continue
		}
		fmt.Fprintf(&b, "_u%X_", ch)
	}
	return b.String()
}

func mangleEnumMember(enumName string, memberName string) string {
	return mangleIdent(enumName + "_" + memberName)
}

func isSafeMangledIdentRune(ch rune) bool {
	return ch == '_' ||
		('a' <= ch && ch <= 'z') ||
		('A' <= ch && ch <= 'Z') ||
		('0' <= ch && ch <= '9')
}
