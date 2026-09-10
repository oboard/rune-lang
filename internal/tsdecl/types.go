package tsdecl

import (
	"strings"
)

// Void marks absent method return types — mapped to Rune Void.
var Void = Primitive{Name: "void"}

// parseMembers extracts interface members from a raw interface body. Entries
// are split at top-level `;` boundaries; every entry is then classified as a
// method, property, or skipped (index signatures, call signatures, new).
func parseMembers(body string) []Member {
	var out []Member
	for _, entry := range splitStatements(body) {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		mem := parseMember(entry)
		if mem.Name != "" {
			out = append(out, mem)
		}
	}
	return out
}

// splitStatements breaks an interface body into member texts at `;`.
// Generic-parameter angles (`<K extends X>`) may legitimately span `;` in
// declaration files, so angle depth doesn't gate statement splitting; only
// paren/bracket/brace depth matters there.
func splitStatements(body string) []string {
	var out []string
	depthRound, depthSquare, depthCurly := 0, 0, 0
	start := 0
	for i := 0; i < len(body); i++ {
		c := body[i]
		switch c {
		case '(':
			depthRound++
		case ')':
			depthRound--
		case '[':
			depthSquare++
		case ']':
			depthSquare--
		case '{':
			depthCurly++
		case '}':
			depthCurly--
		case '"', '\'':
			i = skipString(body, i, c)
		case ';':
			if depthRound == 0 && depthSquare == 0 && depthCurly == 0 {
				out = append(out, body[start:i])
				start = i + 1
			}
		}
	}
	if last := strings.TrimSpace(body[start:]); last != "" {
		out = append(out, last)
	}
	return out
}

// parseMember classifies one interface member entry.
func parseMember(text string) Member {
	// Index signatures have no name — skip entirely.
	if strings.HasPrefix(text, "[") {
		return Member{}
	}
	// Call signatures and constructors — not meaningful outside `new`-less
	// interfaces, skip.
	if strings.HasPrefix(text, "(") || strings.HasPrefix(text, "new ") || strings.HasPrefix(text, "new(") {
		return Member{}
	}
	mem := Member{Signature: text}
	// Strip leading modifiers; they don't change member semantics here.
	for {
		stripped := false
		for _, prefix := range []string{"readonly ", "public ", "protected ", "private ", "static ", "abstract ", "override "} {
			if strings.HasPrefix(text, prefix) {
				text = strings.TrimSpace(strings.TrimPrefix(text, prefix))
				stripped = true
			}
		}
		if !stripped {
			break
		}
	}
	// Accessors map to plain properties — the getter's return type IS the
	// property type. Setters are modeled as methods named after the property
	// (write-only escape hatch for `.classList = "..."`).
	if strings.HasPrefix(text, "get ") {
		inner := text[4:]
		mem.Name = firstWord(inner)
		if idx := strings.Index(inner, ":"); idx >= 0 {
			mem.Type = parseTypeText(strings.TrimSpace(inner[idx+1:]))
		} else {
			mem.Type = Any{}
		}
		return mem
	}
	if strings.HasPrefix(text, "set ") {
		inner := text[4:]
		mem.Name = firstWord(inner)
		mem.IsMethod = true
		if open := strings.Index(inner, "("); open >= 0 {
			innerParens := inner[open:]
			paramsText, _ := splitParenOnce(innerParens)
			mem.Params = parseParams(paramsText)
		}
		mem.Type = Void
		return mem
	}
	// Optional marker.
	if idx := strings.Index(text, "?:"); idx >= 0 && !strings.Contains(text[:idx], ":") {
		mem.Optional = true
	}
	// Find the member name — runs until `(`, `?`, `:` or `<`.
	name, rest := splitMemberName(text)
	if name == "" {
		return Member{}
	}
	mem.Name = strings.TrimSuffix(name, "?")
	mem.Optional = mem.Optional || strings.HasSuffix(name, "?")
	rest = strings.TrimSpace(rest)
	switch {
	case rest == "":
		mem.Type = Any{}
	case rest[0] == '(':
		mem.IsMethod = true
		paramsText, tail := splitParenOnce(rest)
		mem.Params = parseParams(paramsText)
		tail = strings.TrimSpace(tail)
		if strings.HasPrefix(tail, ":") {
			mem.Type = parseTypeText(strings.TrimSpace(tail[1:]))
		} else {
			mem.Type = Void
		}
	case rest[0] == '<':
		// Generic method: swallow the type-parameter list, then parse as
		// method.
		mem.IsMethod = true
		end := skipBalanced(text, strings.Index(text, "<"), '<', '>')
		after := strings.TrimSpace(text[end:])
		if !strings.HasPrefix(after, "(") {
			mem.Type = Any{}
			break
		}
		paramsText, tail := splitParenOnce(after)
		// Remember the declared parameter names.
		mem.Params = parseParams(paramsText)
		tail = strings.TrimSpace(tail)
		if strings.HasPrefix(tail, ":") {
			mem.Type = parseTypeText(strings.TrimSpace(tail[1:]))
		} else {
			mem.Type = Void
		}
	case rest[0] == ':':
		mem.Type = parseTypeText(strings.TrimSpace(rest[1:]))
	default:
		mem.Type = Any{}
	}
	return mem
}

// splitMemberName splits `name<...` / `name(...)` / `name: Type` into name and
// remainder.
func splitMemberName(text string) (string, string) {
	for i := 0; i < len(text); i++ {
		c := text[i]
		if c == '(' || c == ':' || c == '<' || c == '[' {
			return strings.TrimSpace(text[:i]), text[i:]
		}
		if c == '?' && i+1 < len(text) && (text[i+1] == ':' || text[i+1] == '(') {
			return strings.TrimSpace(text[:i+1]), text[i+1:]
		}
	}
	return strings.TrimSpace(text), ""
}

// splitParenOnce cuts the leading balanced `( ... )` group off text.
func splitParenOnce(text string) (inner, tail string) {
	if !strings.HasPrefix(text, "(") {
		return "", text
	}
	end := skipBalanced(text, 0, '(', ')')
	if end > 0 && end <= len(text) {
		inner = text[1:min(end-1, len(text))]
	}
	if end < len(text) {
		tail = text[end:]
	}
	return inner, tail
}

// isIdentByte reports whether a byte may appear in a TypeScript identifier.
func isIdentByte(c byte) bool {
	return c == '_' || c == '$' ||
		(c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// firstWord returns the first identifier-shaped token of text.
func firstWord(text string) string {
	text = strings.TrimSpace(text)
	for i := 0; i < len(text); i++ {
		if !isIdentByte(text[i]) {
			return text[:i]
		}
	}
	return text
}

// parseParams parses a `a: T, b?: U` parameter list.
func parseParams(text string) []Param {
	var out []Param
	for _, part := range splitTopLevel(text, ',') {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		// Varargs `...tokens: T[]` — mark the param as a rest param so callers
		// can emit zero-or-more positional arguments whose types match T.
		rest := strings.HasPrefix(part, "...")
		part = strings.TrimPrefix(part, "...")
		if strings.HasPrefix(part, "this:") {
			continue
		}
		p := Param{Rest: rest}
		if idx := strings.Index(part, ":"); idx >= 0 {
			namePart := strings.TrimSpace(part[:idx])
			if strings.HasSuffix(namePart, "?") {
				p.Optional = true
			}
			p.Name = strings.TrimSuffix(namePart, "?")
			typ := parseTypeText(strings.TrimSpace(part[idx+1:]))
			// Rest params type as `T[]` in TS but should read as `T` for
			// callers in a variadic model — unwrap the array wrapper.
			if rest {
				if arr, ok := typ.(Array); ok {
					typ = arr.Elem
				}
			}
			p.Type = typ
		} else {
			if strings.HasSuffix(part, "?") {
				p.Optional = true
			}
			p.Name = strings.TrimSuffix(part, "?")
			p.Type = Any{}
		}
		if p.Name != "" {
			out = append(out, p)
		}
	}
	return out
}

// parseCallbackType parses `(...args): ret => Return` or `(args: T) => R` text.
func parseCallbackType(text string) (Callback, bool) {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "(") {
		return parseCallbackTypeFrom(text)
	}
	// Unwrapped shape: `args => ret`.
	if findTopLevelArrow(text) < 0 {
		return Callback{}, false
	}
	left := text[:findTopLevelArrow(text)]
	right := strings.TrimSpace(text[findTopLevelArrow(text)+2:])
	if strings.HasPrefix(left, "(") && isBalancedWrapped(left) {
		return Callback{Params: parseParams(strings.TrimSpace(left[1 : len(left)-1])), Return: parseTypeText(right)}, true
	}
	return Callback{}, false
}

func findTopLevelArrow(text string) int {
	depth := 0
	for i := 0; i+1 < len(text); i++ {
		c := text[i]
		switch c {
		case '(':
			depth++
		case ')':
			depth--
		case '"', '\'':
			i = skipString(text, i, c)
		case '=':
			if text[i+1] == '>' && depth == 0 {
				return i
			}
		}
	}
	return -1
}
