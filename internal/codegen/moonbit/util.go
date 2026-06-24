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
	"as": true, "async": true, "break": true, "catch": true, "const": true,
	"continue": true, "derive": true, "else": true, "enum": true, "fn": true,
	"for": true, "guard": true, "if": true, "impl": true, "in": true,
	"is": true, "let": true, "loop": true, "match": true, "mut": true,
	"priv": true, "pub": true, "raise": true, "return": true, "struct": true,
	"test": true, "trait": true, "try": true, "type": true, "while": true,
	"with": true,
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
	case checker.Bytes, checker.Data:
		return "Bytes"
	case checker.Buffer:
		return "Buffer"
	case checker.Error:
		return "String"
	case checker.StringBuffer:
		return "StringBuilder"
	case checker.Regex, checker.Object, checker.Symbol, checker.HTMLElement, checker.WebComponent,
		checker.Reader, checker.Writer, checker.FileStat, checker.TCPConnection,
		checker.TCPListener:
		return "Unit"
	default:
		return mangleType(name)
	}
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
	default:
		return "", nil, false
	}
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
	case checker.Char:
		return quoteChar(0)
	case checker.Bool:
		return "false"
	case checker.StringBuffer:
		return "StringBuilder()"
	default:
		return "()"
	}
}

func unsupportedError(backendName string, feature string) error {
	return fmt.Errorf("MoonBit backend does not support %s", feature)
}
