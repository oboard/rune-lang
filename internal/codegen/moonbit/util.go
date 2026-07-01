package moonbitcodegen

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/lexer"
)

var reserved = map[string]bool{
	"_": true, "abstract": true, "alias": true, "and": true, "anyframe": true,
	"anytype": true, "as": true, "asm": true, "assert": true, "assume": true,
	"async": true, "atomic": true, "await": true, "break": true, "catch": true,
	"comptime": true, "const": true, "constructor": true, "continue": true,
	"declare": true, "define": true, "defer": true, "derive": true, "do": true,
	"downcast": true, "dyn": true, "dynclass": true, "dynobj": true,
	"dynrec": true, "else": true, "enum": true, "enumview": true,
	"errdefer": true, "export": true, "extern": true, "extenum": true,
	"false": true, "final": true, "finally": true, "fn": true, "fnalias": true,
	"for": true, "guard": true, "if": true, "impl": true, "import": true,
	"in": true, "include": true, "inherit": true, "is": true, "isnot": true,
	"lazy": true, "let": true, "letrec": true, "lexmatch": true, "local": true,
	"loop": true, "macro": true, "match": true, "member": true, "method": true,
	"mixin": true, "module": true, "move": true, "mut": true, "namespace": true,
	"noasync": true, "nobreak": true, "noraise": true, "opaque": true,
	"orelse": true, "override": true, "package": true, "priv": true,
	"proof_assert": true, "proof_let": true, "protected": true, "pub": true,
	"raise": true, "readonly": true, "recur": true, "ref": true, "resume": true,
	"return": true, "sealed": true, "static": true, "struct": true,
	"suberror": true, "super": true, "test": true, "threadlocal": true,
	"throw": true, "trait": true, "traitalias": true, "true": true,
	"try": true, "type": true, "typealias": true, "typeof": true,
	"unsafe": true, "unreachable": true, "upcast": true, "use": true,
	"using": true, "var": true, "virtual": true, "void": true, "volatile": true,
	"where": true, "while": true, "with": true, "yield": true,
}

var identRE = regexp.MustCompile(`[^A-Za-z0-9_]`)

func mangleIdent(name string) string {
	if name == "" {
		return "_"
	}
	name = identRE.ReplaceAllString(name, "_")
	if name == "" {
		name = "_"
	}
	first, _ := utf8FirstRune(name)
	if unicode.IsDigit(first) || unicode.IsUpper(first) {
		name = "rune_" + name
	}
	if reserved[name] {
		return name + "_"
	}
	return name
}

func mangleType(name string) string {
	if name == "" {
		return "RuneValue"
	}
	name = identRE.ReplaceAllString(name, "_")
	if name == "" {
		return "RuneValue"
	}
	first, _ := utf8FirstRune(name)
	if !unicode.IsUpper(first) {
		name = "Rune_" + name
	}
	return name
}

func mangleMethod(typeName string, method string) string {
	return mangleType(typeName) + "::" + mangleIdent(method)
}

func utf8FirstRune(s string) (rune, int) {
	for i, ch := range s {
		return ch, i
	}
	return '_', 0
}

func mbtType(typ checker.Type) string {
	name := string(typ)
	if strings.HasPrefix(name, "&") {
		return "Unit"
	}
	if inner, ok := parseNullableType(name); ok {
		return mbtType(checker.Type(inner)) + "?"
	}
	if elem, ok := checker.ArrayElement(typ); ok {
		return "Array[" + mbtType(elem) + "]"
	}
	if base, args, ok := parseGenericType(name); ok {
		switch base {
		case "ReadonlyArray":
			return "Array[" + mbtType(checker.Type(args[0])) + "]"
		case "Tuple", "ReadonlyTuple":
			return "(" + mbtTypeList(args) + ")"
		case "Map", "WeakMap", "Record":
			return "Map[" + mbtTypeList(args) + "]"
		case "Set", "WeakSet":
			return "Set[" + mbtType(checker.Type(args[0])) + "]"
		case "Result":
			return "Result[" + mbtTypeList(args) + "]"
		case "Task":
			return "Unit"
		case "Iter":
			return "Iter[" + mbtType(checker.Type(args[0])) + "]"
		case "Func", "AsyncFunc":
			if len(args) == 0 {
				return "() -> Unit"
			}
			params := make([]string, 0, len(args)-1)
			for _, arg := range args[:len(args)-1] {
				params = append(params, mbtType(checker.Type(arg)))
			}
			return "(" + strings.Join(params, ", ") + ") -> " + mbtType(checker.Type(args[len(args)-1]))
		}
	}
	if params, ret, ok := parseFuncType(name); ok {
		parts := make([]string, 0, len(params))
		for _, param := range params {
			parts = append(parts, mbtType(checker.Type(param)))
		}
		return "(" + strings.Join(parts, ", ") + ") -> " + mbtType(checker.Type(ret))
	}
	switch typ {
	case checker.Int, checker.Int4, checker.Int8, checker.Int16:
		return "Int"
	case checker.Int64:
		return "Int64"
	case checker.UInt, checker.UInt8, checker.UInt16:
		return "UInt"
	case checker.UInt64:
		return "UInt64"
	case checker.Double:
		return "Double"
	case checker.Float:
		return "Float"
	case checker.String:
		return "String"
	case checker.Char:
		return "Char"
	case checker.Bool:
		return "Bool"
	case checker.Void, checker.Null, checker.Never, checker.Unknown:
		return "Unit"
	case checker.BigInt:
		return "BigInt"
	case checker.Bytes, checker.Data, checker.Buffer, checker.Writer:
		return "Array[Int]"
	case checker.Reader:
		return "RuneReader"
	case checker.Error:
		return "String"
	case checker.StringBuffer:
		return "StringBuilder"
	case checker.Symbol:
		return "String"
	case checker.Regex:
		return "RuneRegex"
	case checker.FileStat:
		return "RuneFileStat"
	case checker.Object:
		return "Json"
	case checker.HTMLElement, checker.WebComponent,
		checker.TCPConnection, checker.TCPListener:
		return "Unit"
	default:
		if mbtJSONValueType(typ) {
			return "Json"
		}
		return mangleType(name)
	}
}

func (g *generator) mbtType(typ checker.Type) string {
	name := string(typ)
	if fields, ok := parseObjectType(name); ok {
		if len(fields) == 0 {
			return "Unit"
		}
		return g.anonymousTypeName(typ)
	}
	if strings.HasPrefix(name, "&") {
		return "Unit"
	}
	if inner, ok := parseNullableType(name); ok {
		return g.mbtType(checker.Type(inner)) + "?"
	}
	if elem, ok := checker.ArrayElement(typ); ok {
		return "Array[" + g.mbtType(elem) + "]"
	}
	if base, args, ok := parseGenericType(name); ok {
		switch base {
		case "ReadonlyArray":
			return "Array[" + g.mbtType(checker.Type(args[0])) + "]"
		case "Tuple", "ReadonlyTuple":
			return "(" + g.mbtTypeList(args) + ")"
		case "Map", "WeakMap", "Record":
			return "Map[" + g.mbtTypeList(args) + "]"
		case "Set", "WeakSet":
			return "Set[" + g.mbtType(checker.Type(args[0])) + "]"
		case "Result":
			return "Result[" + g.mbtTypeList(args) + "]"
		case "Task":
			return "Unit"
		case "Iter":
			return "Iter[" + g.mbtType(checker.Type(args[0])) + "]"
		case "Func", "AsyncFunc":
			if len(args) == 0 {
				return "() -> Unit"
			}
			params := make([]string, 0, len(args)-1)
			for _, arg := range args[:len(args)-1] {
				params = append(params, g.mbtType(checker.Type(arg)))
			}
			return "(" + strings.Join(params, ", ") + ") -> " + g.mbtType(checker.Type(args[len(args)-1]))
		}
	}
	if params, ret, ok := parseFuncType(name); ok {
		parts := make([]string, 0, len(params))
		for _, param := range params {
			parts = append(parts, g.mbtType(checker.Type(param)))
		}
		return "(" + strings.Join(parts, ", ") + ") -> " + g.mbtType(checker.Type(ret))
	}
	return mbtType(typ)
}

func (g *generator) mbtTypeList(args []string) string {
	out := make([]string, 0, len(args))
	for _, arg := range args {
		out = append(out, g.mbtType(checker.Type(arg)))
	}
	return strings.Join(out, ", ")
}

func mbtTypeList(args []string) string {
	out := make([]string, 0, len(args))
	for _, arg := range args {
		out = append(out, mbtType(checker.Type(arg)))
	}
	return strings.Join(out, ", ")
}

func mbtGenerics(names []string) string {
	if len(names) == 0 {
		return ""
	}
	parts := make([]string, 0, len(names))
	for _, name := range names {
		parts = append(parts, mangleType(name))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func mbtFnPrefix(generics []string) string {
	return "fn" + mbtGenerics(generics)
}

func parseGenericType(name string) (string, []string, bool) {
	idx := strings.IndexByte(name, '[')
	if idx <= 0 || !strings.HasSuffix(name, "]") {
		return "", nil, false
	}
	base := name[:idx]
	args := splitTypeList(strings.TrimSuffix(name[idx+1:], "]"))
	switch base {
	case "ReadonlyArray", "Set", "WeakSet", "Task", "Iter":
		return base, args, len(args) == 1
	case "Tuple", "ReadonlyTuple":
		return base, args, len(args) > 0
	case "Map", "WeakMap", "Record", "Result":
		return base, args, len(args) == 2
	case "Func", "AsyncFunc":
		return base, args, len(args) > 0
	default:
		return "", nil, false
	}
}

func isResultType(typ checker.Type) bool {
	base, _, ok := parseGenericType(string(typ))
	return ok && base == "Result"
}

func parseNullableType(name string) (string, bool) {
	if !strings.HasSuffix(name, "?") || name == "?" {
		return "", false
	}
	return strings.TrimSuffix(name, "?"), true
}

func parseFuncType(name string) ([]string, string, bool) {
	if !strings.HasPrefix(name, "(") {
		return nil, "", false
	}
	idx := strings.LastIndex(name, ")->")
	if idx < 0 {
		idx = strings.LastIndex(name, ") ->")
	}
	if idx < 0 {
		return nil, "", false
	}
	paramsEnd := strings.Index(name, ")")
	if paramsEnd < 0 {
		return nil, "", false
	}
	params := strings.TrimSpace(name[1:paramsEnd])
	ret := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(name[paramsEnd+1:]), "->"))
	if params == "" {
		return nil, ret, true
	}
	return splitTypeList(params), ret, ret != ""
}

type objectField struct {
	name string
	typ  string
}

func parseObjectType(name string) ([]objectField, bool) {
	if !strings.HasPrefix(name, "{") || !strings.HasSuffix(name, "}") {
		return nil, false
	}
	inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(name, "{"), "}"))
	if inner == "" {
		return nil, true
	}
	parts := splitTypeList(inner)
	fields := make([]objectField, 0, len(parts))
	for _, part := range parts {
		fieldName, fieldType, ok := splitObjectField(part)
		if !ok {
			return nil, false
		}
		fields = append(fields, objectField{name: fieldName, typ: fieldType})
	}
	return fields, true
}

func splitObjectField(src string) (string, string, bool) {
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

func splitTypeList(src string) []string {
	var out []string
	depth := 0
	start := 0
	for i, ch := range src {
		switch ch {
		case '[', '(', '{':
			depth++
		case ']', ')', '}':
			depth--
		case ',':
			if depth == 0 {
				out = append(out, strings.TrimSpace(src[start:i]))
				start = i + 1
			}
		}
	}
	if tail := strings.TrimSpace(src[start:]); tail != "" {
		out = append(out, tail)
	}
	return out
}

func mbtBinaryOp(op lexer.Kind) string {
	switch op {
	case lexer.AndAnd:
		return "&&"
	case lexer.OrOr:
		return "||"
	case lexer.QuestionQuestion:
		return "??"
	case lexer.EqualEqual:
		return "=="
	case lexer.BangEqual:
		return "!="
	default:
		return op.String()
	}
}

func mbtPrecedence(op lexer.Kind) int {
	switch op {
	case lexer.OrOr:
		return 1
	case lexer.AndAnd:
		return 2
	case lexer.EqualEqual, lexer.BangEqual, lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
		return 3
	case lexer.Plus, lexer.Minus:
		return 4
	case lexer.Star, lexer.Slash, lexer.Percent:
		return 5
	default:
		return 0
	}
}

func quoteString(value string) string {
	return strconv.Quote(value)
}

func quoteChar(value rune) string {
	switch value {
	case '\n':
		return "'\\n'"
	case '\r':
		return "'\\r'"
	case '\t':
		return "'\\t'"
	case '\'':
		return "'\\''"
	case '\\':
		return "'\\\\'"
	}
	if value < 0x20 || value == 0x7f {
		return fmt.Sprintf("'\\u%04x'", value)
	}
	return strconv.QuoteRune(value)
}

func zeroValue(typ checker.Type) string {
	if inner, ok := parseNullableType(string(typ)); ok && inner != "" {
		return "None"
	}
	if base, args, ok := parseGenericType(string(typ)); ok {
		switch base {
		case "Iter":
			return "Iter::new(fn() { None })"
		case "ReadonlyArray":
			return "[]"
		case "Map", "WeakMap", "Record":
			return "{}"
		case "Set", "WeakSet":
			return "Set::new()"
		}
		_ = args
	}
	switch typ {
	case checker.Int, checker.Int4, checker.Int8, checker.Int16, checker.UInt, checker.UInt8, checker.UInt16:
		return "0"
	case checker.Int64:
		return "0L"
	case checker.UInt64:
		return "0UL"
	case checker.Double, checker.Float:
		return "0.0"
	case checker.String:
		return "\"\""
	case checker.Bytes, checker.Data, checker.Buffer, checker.Writer:
		return "[]"
	case checker.Reader:
		return "RuneReader::{ data: [], position: 0, nibble: 0 }"
	case checker.Symbol:
		return "\"l:\""
	case checker.Regex:
		return "rune_regex_new(\"\", \"\")"
	case checker.Char:
		return quoteChar(0)
	case checker.Bool:
		return "false"
	case checker.StringBuffer:
		return "StringBuilder()"
	case checker.FileStat:
		return "RuneFileStat::{ size: 0, isFile: false, isDirectory: false }"
	default:
		if mbtJSONValueType(typ) {
			return "Json::null()"
		}
		return "()"
	}
}

func zeroFunctionValue(typ checker.Type) (string, bool) {
	name := string(typ)
	if base, args, ok := parseGenericType(name); ok && (base == "Func" || base == "AsyncFunc") {
		return zeroFunctionValueFromParts(args[:len(args)-1], checker.Type(args[len(args)-1])), true
	}
	if params, ret, ok := parseFuncType(name); ok {
		return zeroFunctionValueFromParts(params, checker.Type(ret)), true
	}
	return "", false
}

func zeroFunctionValueFromParts(params []string, ret checker.Type) string {
	names := make([]string, 0, len(params))
	for range params {
		names = append(names, "_")
	}
	return "(" + strings.Join(names, ", ") + ") => " + zeroValue(ret)
}

func mbtJSONValueType(typ checker.Type) bool {
	name := string(typ)
	return typ == checker.Object || name == "Dynamic"
}

func unsupportedError(backendName string, feature string) error {
	return fmt.Errorf("MoonBit backend does not support %s", feature)
}
