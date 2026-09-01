package tscodegen

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
	"github.com/oboard/rune-lang/internal/lexer"
)

type generator struct {
	buf        bytes.Buffer
	file       *ir.File
	indent     int
	temp       int
	errors     []error
	thisNames  []string
	mapGetters map[string]string
	signals    []map[string]checker.Type
	signalDeps []map[string][]string
	reactives  []map[string]checker.Type
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
	if strings.HasPrefix(string(typ), "&") {
		return "any"
	}
	if inner, ok := parseTSNullableType(string(typ)); ok {
		return tsType(checker.Type(inner)) + " | null"
	}
	if base, args, ok := parseTSGenericType(string(typ)); ok {
		switch base {
		case "Result":
			return "RuneResult<" + tsType(checker.Type(args[0])) + ", " + tsType(checker.Type(args[1])) + ">"
		case "Task":
			return "Promise<" + tsType(checker.Type(args[0])) + ">"
		case "Iter":
			return "RuneIter<" + tsType(checker.Type(args[0])) + ">"
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
	case checker.Int, checker.Int4, checker.Int8, checker.Int16, checker.UInt, checker.UInt8, checker.UInt16:
		return "number"
	case checker.Double, checker.Float:
		return "number"
	case checker.BigInt, checker.Int64, checker.UInt64:
		return "bigint"
	case checker.String:
		return "string"
	case checker.Char:
		return "string"
	case checker.Bool:
		return "boolean"
	case checker.Null:
		return "null"
	case checker.Object:
		return "object"
	case checker.Bytes:
		return "DataView"
	case checker.Buffer:
		return "RuneBuffer"
	case checker.Reader:
		return "RuneReader"
	case checker.Writer:
		return "RuneWriter"
	case checker.StringBuffer:
		return "RuneStringBuffer"
	case checker.FileStat:
		return "RuneFileStat"
	case checker.TCPConnection:
		return "RuneTCPConnection"
	case checker.TCPListener:
		return "RuneTCPListener"
	case checker.Data:
		return "Uint8Array"
	case checker.Error:
		return "RuneError"
	case checker.Never:
		return "never"
	case checker.Symbol:
		return "symbol"
	case checker.Regex:
		return "RegExp"
	case checker.Void:
		return "void"
	case checker.HTMLElement:
		return "HTMLElement"
	case checker.WebComponent:
		return "CustomElementConstructor"
	case checker.Unknown:
		return "any"
	default:
		return mangleIdent(string(typ))
	}
}

func tsGenerics(names []string, constraints map[string]string) string {
	if len(names) == 0 {
		return ""
	}
	parts := make([]string, 0, len(names))
	for _, name := range names {
		parts = append(parts, mangleIdent(name)+" extends "+tsGenericConstraint(constraints[name]))
	}
	return "<" + strings.Join(parts, ", ") + ">"
}

func tsGenericConstraint(name string) string {
	switch name {
	case "Add", "Sub", "Mul", "Div", "Number":
		return "number"
	default:
		return "unknown"
	}
}

func isTSGenericResultType(typ checker.Type) bool {
	name := string(typ)
	if name == "" || strings.HasPrefix(name, "&") || strings.ContainsAny(name, "[]{}(),?") {
		return false
	}
	switch typ {
	case checker.Int, checker.Int4, checker.Int8, checker.Int16, checker.UInt, checker.UInt8, checker.UInt16,
		checker.Double, checker.Float, checker.BigInt, checker.Int64, checker.UInt64, checker.String, checker.Char,
		checker.Bool, checker.Null, checker.Object, checker.Bytes, checker.Buffer, checker.Reader, checker.Writer,
		checker.StringBuffer, checker.FileStat, checker.TCPConnection, checker.TCPListener, checker.Data,
		checker.Error, checker.Never, checker.Symbol, checker.Regex, checker.Void, checker.HTMLElement,
		checker.WebComponent, checker.Unknown:
		return false
	default:
		return true
	}
}

func tsArithmeticOp(op lexer.Kind) bool {
	switch op {
	case lexer.Plus, lexer.Minus, lexer.Star, lexer.Slash, lexer.Percent:
		return true
	default:
		return false
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
	if base, _, ok := parseTSGenericType(string(typ)); ok {
		switch base {
		case "Result":
			return "{ ok: false, error: undefined as any }"
		case "Task":
			return "Promise.resolve(undefined as any)"
		}
	}
	switch typ {
	case checker.Int, checker.Int4, checker.Int8, checker.Int16, checker.UInt, checker.UInt8, checker.UInt16:
		return "0"
	case checker.Double, checker.Float:
		return "0"
	case checker.BigInt, checker.Int64, checker.UInt64:
		return "0n"
	case checker.String:
		return `""`
	case checker.Char:
		return `"\0"`
	case checker.Bool:
		return "false"
	case checker.Regex:
		return `/(?:)/`
	case checker.Symbol:
		return "Symbol()"
	case checker.Bytes:
		return "new DataView(new ArrayBuffer(0))"
	case checker.Buffer:
		return "new RuneBuffer()"
	case checker.Reader:
		return "new RuneReader(new DataView(new ArrayBuffer(0)))"
	case checker.Writer:
		return "new RuneWriter()"
	case checker.StringBuffer:
		return "new RuneStringBuffer()"
	case checker.FileStat:
		return "{ size: 0, isFile: false, isDirectory: false }"
	case checker.TCPConnection, checker.TCPListener:
		return "undefined as any"
	case checker.Data:
		return "new Uint8Array()"
	case checker.Error:
		return "{ code: 0, message: \"\", cause: null }"
	case checker.Null:
		return "null"
	case checker.Never:
		return "undefined as never"
	case checker.Void:
		return "undefined"
	case checker.HTMLElement:
		return "document.createElement(\"div\")"
	case checker.WebComponent:
		return "class extends HTMLElement {}"
	default:
		return "{} as " + tsType(typ)
	}
}

func (g *generator) zeroValue(typ checker.Type) string {
	if g.hasEnumType(typ) {
		return "0 as " + tsType(typ)
	}
	return zeroValue(typ)
}

func (g *generator) hasEnumType(typ checker.Type) bool {
	for _, enum := range g.file.Enums {
		if checker.Type(enum.Name) == typ {
			return true
		}
	}
	return false
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

func FunctionSymbolName(fn *ir.Function) string {
	name := fn.SourceName
	if name == "" {
		name = fn.Name
	}
	return mangleIdent(name)
}

func isSafeMangledIdentRune(ch rune) bool {
	return ch == '_' || ch == '$' ||
		('a' <= ch && ch <= 'z') ||
		('A' <= ch && ch <= 'Z') ||
		('0' <= ch && ch <= '9')
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

func tsExportName(name string) string {
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
