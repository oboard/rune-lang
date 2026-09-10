package tsdecl

import (
	"strings"
)

// parseTypeText maps a TS type expression to the model. Anything
// unrepresentable degrades to Any (Rune Dynamic) rather than failing.
func parseTypeText(text string) Type {
	text = strings.TrimSpace(text)
	if text == "" {
		return Any{}
	}
	// 1. Top-level union / intersection.
	if findTopLevel(text, '|') >= 0 {
		return unionOf(text, '|')
	}
	if findTopLevel(text, '&') >= 0 {
		return unionOf(text, '&')
	}
	// 2. Group-wrapped — callback shape, array suffix, or group-only.
	if strings.HasPrefix(text, "(") {
		inner, tail := splitParenOnce(text)
		if cb, ok := parseCallbackTypeFrom(text); ok {
			return cb
		}
		if tail == "[]" {
			return Array{Elem: parseTypeText(strings.TrimSpace(inner))}
		}
		if tail == "" {
			return parseTypeText(strings.TrimSpace(inner))
		}
		return Any{}
	}
	// 3. Unwrapped callback shape `(params) => ret`.
	if strings.Contains(text, "=>") {
		if cb, ok := parseCallbackType(text); ok {
			return cb
		}
		return Any{}
	}
	// 4. Bare array suffix.
	if strings.HasSuffix(text, "[]") {
		return Array{Elem: parseTypeText(text[:len(text)-2])}
	}
	// 5. Keywords.
	if kw := parsePrimitive(text); kw != nil {
		return kw
	}
	// 6. `typeof X`.
	if strings.HasPrefix(text, "typeof ") {
		return Ref{Name: text[7:]}
	}
	// 7. Quoted literals.
	if strings.HasPrefix(text, `"`) || strings.HasPrefix(text, `'`) || strings.HasPrefix(text, "`") {
		return Literal{Value: strings.Trim(text, "\"'`")}
	}
	// 8. Generics `Base<Args>` — kept for `Map<K,V>`/`Set<T>`, array is a
	// special-cased.
	if open := strings.Index(text, "<"); open > 0 && strings.HasSuffix(text, ">") {
		base := strings.TrimSpace(text[:open])
		if isBalancedWrapped(text[open:]) {
			if base == "Array" || base == "ReadonlyArray" {
				return Array{Elem: parseTypeText(text[open+1 : len(text)-1])}
			}
			args := splitTopLevel(text[open+1:len(text)-1], ',')
			refs := make([]Type, 0, len(args))
			for _, a := range args {
				refs = append(refs, parseTypeText(a))
			}
			return Ref{Name: base, Args: refs}
		}
	}
	// 9. Anything containing brackets/quotes/equal-signs isn't a plain ref.
	if strings.ContainsAny(text, "[]{}()\"'&|<>=") {
		return Any{}
	}
	// 10. Plain identifier reference.
	return Ref{Name: text}
}

func unionOf(text string, sep byte) Type {
	parts := splitTopLevel(text, sep)
	opts := make([]Type, 0, len(parts))
	for _, p := range parts {
		opts = append(opts, parseTypeText(p))
	}
	if sep == '&' {
		return Intersect{Options: opts}
	}
	return Union{Options: opts}
}

// parsePrimitive matches TS keyword leaves to the model.
func parsePrimitive(text string) Type {
	switch text {
	case "string":
		return Primitive{Name: "String"}
	case "boolean":
		return Primitive{Name: "Bool"}
	case "number":
		return Primitive{Name: "Double"}
	case "bigint":
		return Primitive{Name: "BigInt"}
	case "symbol":
		return Primitive{Name: "Symbol"}
	case "null":
		return Primitive{Name: "Null"}
	case "undefined", "void":
		return Void
	case "object", "Object":
		return Primitive{Name: "Object"}
	case "any", "unknown":
		return Any{}
	case "true", "false":
		return Literal{Value: text}
	}
	if isNumericLiteral(text) {
		return Literal{Value: text}
	}
	return nil
}

// findTopLevel returns the offset of the first occurrence of sep at zero
// nesting depth. `=>` arrows are treated as a single token so callback types
// don't interfere with generic-balance tracking.
func findTopLevel(text string, sep byte) int {
	dR, dS, dC, dA := 0, 0, 0, 0
	for i := 0; i < len(text); i++ {
		c := text[i]
		switch c {
		case '=':
			if i+1 < len(text) && text[i+1] == '>' {
				i++
				continue
			}
		case '(':
			dR++
		case ')':
			dR--
		case '[':
			dS++
		case ']':
			dS--
		case '{':
			dC++
		case '}':
			dC--
		case '<':
			dA++
		case '>':
			dA--
		case '"', '\'':
			i = skipString(text, i, c)
		case sep:
			if dR == 0 && dS == 0 && dC == 0 && dA == 0 {
				return i
			}
		}
	}
	return -1
}

// isBalancedWrapped checks if the whole text is one balanced outer pair.
func isBalancedWrapped(text string) bool {
	if len(text) < 2 {
		return false
	}
	open := text[0]
	var close byte
	switch open {
	case '(':
		close = ')'
	case '<':
		close = '>'
	case '[':
		close = ']'
	case '{':
		close = '}'
	default:
		return false
	}
	depth := 0
	for i := 0; i < len(text); i++ {
		c := text[i]
		if c == '"' || c == '\'' {
			i = skipString(text, i, c)
			continue
		}
		if c == open {
			depth++
		} else if c == close {
			depth--
			if depth == 0 {
				return i == len(text)-1
			}
		}
	}
	return false
}

// parseCallbackTypeFrom tries `(...) => R`.
func parseCallbackTypeFrom(text string) (Callback, bool) {
	if !strings.HasPrefix(text, "(") {
		return Callback{}, false
	}
	paramsText, tail := splitParenOnce(text)
	tail = strings.TrimSpace(tail)
	if !strings.HasPrefix(tail, "=>") {
		return Callback{}, false
	}
	ret := parseTypeText(strings.TrimSpace(tail[2:]))
	return Callback{Params: parseParams(paramsText), Return: ret}, true
}

// TypeName renders a parsed node back to a stable string.
func TypeName(t Type) string {
	switch t := t.(type) {
	case nil:
		return "any"
	case Any:
		return "any"
	case Primitive:
		return t.Name
	case Literal:
		return `"` + t.Value + `"`
	case Ref:
		if len(t.Args) > 0 {
			parts := make([]string, 0, len(t.Args))
			for _, a := range t.Args {
				parts = append(parts, TypeName(a))
			}
			return t.Name + "<" + joinStrings(parts, ", ") + ">"
		}
		return t.Name
	case Array:
		return "Array<" + TypeName(t.Elem) + ">"
	case Union:
		parts := make([]string, 0, len(t.Options))
		for _, o := range t.Options {
			parts = append(parts, TypeName(o))
		}
		return joinStrings(parts, " | ")
	case Intersect:
		parts := make([]string, 0, len(t.Options))
		for _, o := range t.Options {
			parts = append(parts, TypeName(o))
		}
		return joinStrings(parts, " & ")
	case Callback:
		parts := make([]string, 0, len(t.Params))
		for _, p := range t.Params {
			part := p.Name
			if p.Type != nil {
				part += ": " + TypeName(p.Type)
			}
			parts = append(parts, part)
		}
		ret := "void"
		if t.Return != nil {
			ret = TypeName(t.Return)
		}
		return "(" + joinStrings(parts, ", ") + ") => " + ret
	}
	return "any"
}

func joinStrings(items []string, sep string) string {
	out := ""
	for i, item := range items {
		if i > 0 {
			out += sep
		}
		out += item
	}
	return out
}

var digitBytes = func(c byte) bool { return c >= '0' && c <= '9' }

func isNumericLiteral(text string) bool {
	if text == "" {
		return false
	}
	if text[0] == '-' {
		if len(text) < 2 {
			return false
		}
		text = text[1:]
	}
	if text == "" {
		return false
	}
	if text[0] < '0' || text[0] > '9' {
		return false
	}
	for i := 1; i < len(text); i++ {
		c := text[i]
		if c == '.' {
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
