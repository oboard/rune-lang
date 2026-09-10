package tsdecl

import (
	"regexp"
	"strings"
)

var (
	interfaceRE  = regexp.MustCompile(`(?m)^interface\s+([A-Za-z_$][A-Za-z0-9_$]*)`)
	typeAliasRE  = regexp.MustCompile(`(?m)^type\s+([A-Za-z_$][A-Za-z0-9_$]*)`)
	valueDeclRE  = regexp.MustCompile(`(?m)^declare\s+(?:var|let|const)\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*:`)
	funcDeclRE   = regexp.MustCompile(`(?m)^declare\s+(?:async\s+)?function\s+([A-Za-z_$][A-Za-z0-9_$]*)`)
)

// Parse processes a TypeScript declaration source, specialising in
// interfaces, type aliases, declare-var globals, and declare-function
// statements. The source may contain anything; unrecognised constructs are
// skipped.
func Parse(src string) *DOM {
	dom := &DOM{
		Interfaces: map[string]*Interface{},
		Aliases:    map[string]*Alias{},
		Globals:    map[string]*Value{},
		Functions:  []*Function{},
	}
	cleaned := cleanSource(src)
	parseInterfaces(cleaned, dom)
	parseTypeAliases(cleaned, dom)
	parseGlobals(cleaned, dom)
	parseFunctions(cleaned, dom)
	return dom
}

// cleanSource strips comments but preserves line structure so statement
// boundaries remain stable.
func cleanSource(src string) string {
	out := make([]byte, 0, len(src))
	inBlock := false
	i := 0
	for i < len(src) {
		c := src[i]
		switch {
		case inBlock:
			if c == '*' && i+1 < len(src) && src[i+1] == '/' {
				inBlock = false
				i += 2
				continue
			}
			if c == '\n' || c == '\r' {
				out = append(out, c)
			} else {
				out = append(out, ' ')
			}
			i++
		case c == '/' && i+1 < len(src) && src[i+1] == '*':
			inBlock = true
			out = append(out, ' ', ' ')
			i += 2
		case c == '/' && i+1 < len(src) && src[i+1] == '/':
			for i < len(src) && src[i] != '\n' && src[i] != '\r' {
				i++
			}
		case c == '"' || c == '\'':
			// Copy string literal verbatim.
			quote := c
			out = append(out, c)
			i++
			for i < len(src) {
				out = append(out, src[i])
				if src[i] == '\\' && i+1 < len(src) {
					i++
					out = append(out, src[i])
				} else if src[i] == quote {
					i++
					break
				}
				i++
			}
		default:
			out = append(out, c)
			i++
		}
	}
	return string(out)
}

// parseInterfaces locates `interface Name<...> extends Bases { body }`
// blocks with a line-anchored regex, then balances braces to capture the
// body.
func parseInterfaces(src string, dom *DOM) {
	for _, loc := range interfaceRE.FindAllStringSubmatchIndex(src, -1) {
		name := src[loc[2]:loc[3]]
		headerEnd := findBraceOpening(src, loc[1])
		if headerEnd < 0 {
			continue
		}
		bodyEnd := skipBalanced(src, headerEnd, '{', '}')
		if bodyEnd <= headerEnd {
			continue
		}
		body := src[headerEnd+1 : bodyEnd-1]
		bases := parseHeritageBases(src[loc[1]:headerEnd])
		mergeInterface(dom, name, bases, parseMembers(body))
		dom.Order = append(dom.Order, name)
	}
}

func mergeInterface(dom *DOM, name string, bases []string, members []Member) {
	if existing := dom.Interfaces[name]; existing != nil {
		supplement(existing, bases, members)
		return
	}
	dom.Interfaces[name] = &Interface{Name: name, Bases: bases, Members: members}
}

func supplement(iface *Interface, bases []string, members []Member) {
	seenBases := map[string]bool{}
	for _, b := range iface.Bases {
		seenBases[b] = true
	}
	for _, b := range bases {
		if !seenBases[b] {
			iface.Bases = append(iface.Bases, b)
			seenBases[b] = true
		}
	}
	seenMembers := map[string]bool{}
	for _, m := range iface.Members {
		seenMembers[m.Name+"\x00"+memberKind(m)] = true
	}
	for _, m := range members {
		key := m.Name + "\x00" + memberKind(m)
		if !seenMembers[key] {
			iface.Members = append(iface.Members, m)
			seenMembers[key] = true
		}
	}
}

func memberKind(m Member) string {
	if m.IsMethod {
		return "m"
	}
	return "f"
}

// parseHeritageBases extracts the `extends A, B, C` clause from an interface
// header.
func parseHeritageBases(header string) []string {
	idx := strings.LastIndex(header, "extends")
	if idx < 0 {
		return nil
	}
	clause := strings.TrimSpace(header[idx+len("extends"):])
	if clause == "" {
		return nil
	}
	return splitTopLevel(clause, ',')
}

// parseTypeAliases extracts `type Name = <text>;` statements.
func parseTypeAliases(src string, dom *DOM) {
	for _, loc := range typeAliasRE.FindAllStringSubmatchIndex(src, -1) {
		name := src[loc[2]:loc[3]]
		stmtEnd := findStatementEnd(src, loc[1])
		stmt := strings.TrimSpace(src[loc[1]:stmtEnd])
		body := strings.TrimPrefix(stmt, "type")
		body = strings.TrimPrefix(strings.TrimSpace(body), name)
		body = strings.TrimSpace(body)
		if !strings.HasPrefix(body, "=") {
			continue
		}
		body = strings.TrimSpace(body[1:])
		alias := &Alias{Name: name, Pos: Position{Offset: loc[2]}}
		if strings.Contains(body, "=>") {
			if cb, ok := parseCallbackType(body); ok {
				alias.Node = cb
			}
		} else if !strings.ContainsAny(body, "=&<>()[],:{}") {
			alias.Node = parseTypeText(body)
		}
		dom.Aliases[name] = alias
	}
}

// parseGlobals captures `declare var/let/const name: Type;` values.
func parseGlobals(src string, dom *DOM) {
	for _, loc := range valueDeclRE.FindAllStringSubmatchIndex(src, -1) {
		name := src[loc[2]:loc[3]]
		// Full match includes the colon; the type text runs from the match end
		// (just past `:`) to the statement terminator.
		stmtEnd := findStatementEnd(src, loc[1])
		typeText := strings.TrimSpace(src[loc[1]:stmtEnd])
		// Strip a trailing default value (`document.all.__context__ = ...` in
		// some libs is handled as a no-value event).
		if eq := findTopLevel(typeText, '='); eq >= 0 {
			typeText = strings.TrimSpace(typeText[:eq])
		}
		if strings.HasPrefix(typeText, "{") || typeText == "" {
			continue
		}
		node := parseTypeText(typeText)
		if _, isAny := node.(Any); isAny {
			continue
		}
		dom.Globals[name] = &Value{Name: name, Type: node, Pos: Position{Offset: loc[2]}}
	}
}

// parseFunctions captures `declare [async] function name(params): Return;`.
func parseFunctions(src string, dom *DOM) {
	for _, loc := range funcDeclRE.FindAllStringSubmatchIndex(src, -1) {
		name := src[loc[2]:loc[3]]
		stmtEnd := findStatementEnd(src, loc[1])
		stmt := strings.TrimSpace(src[loc[1]:stmtEnd])
		open := strings.Index(stmt, "(")
		if open < 0 {
			continue
		}
		close := skipBalanced(stmt, open, '(', ')')
		if close > len(stmt) {
			continue
		}
		paramsText := stmt[open+1 : close-1]
		tail := strings.TrimSpace(stmt[close:])
		ret := Type(Any{})
		if strings.HasPrefix(tail, ":") {
			ret = parseTypeText(strings.TrimSpace(tail[1:]))
		}
		dom.Functions = append(dom.Functions, &Function{
			Name:   name,
			Params: parseParams(paramsText),
			Return: ret,
			Pos:    Position{Offset: loc[2]},
		})
	}
}

// findBraceOpening locates the next top-level `{` after an offset.
func findBraceOpening(src string, from int) int {
	i := from
	depthPar, depthAng := 0, 0
	for i < len(src) {
		c := src[i]
		switch c {
		case '(':
			depthPar++
		case ')':
			depthPar--
		case '<':
			depthAng++
		case '>':
			depthAng--
		case '"', '\'':
			i = skipString(src, i, c)
		case '{':
			if depthPar == 0 && depthAng == 0 {
				return i
			}
		}
		i++
	}
	return -1
}

// skipBalanced returns the offset just past the matching close bracket.
func skipBalanced(src string, start int, open, close byte) int {
	depth := 0
	for i := start; i < len(src); i++ {
		c := src[i]
		switch c {
		case '"', '\'':
			i = skipString(src, i, c)
		default:
			if c == open {
				depth++
			} else if c == close {
				depth--
				if depth == 0 {
					return i + 1
				}
			}
		}
	}
	return len(src)
}

// skipString returns the offset just past the closing quote.
func skipString(src string, start int, quote byte) int {
	i := start + 1
	for i < len(src) {
		if src[i] == '\\' {
			i += 2
			continue
		}
		if src[i] == quote {
			return i + 1
		}
		i++
	}
	return len(src)
}

// findStatementEnd returns the offset of the next top-level `;` or end of
// source.
func findStatementEnd(src string, start int) int {
	dR, dS, dC, dA := 0, 0, 0, 0
	for i := start; i < len(src); i++ {
		c := src[i]
		switch c {
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
			i = skipString(src, i, c)
		case ';':
			if dR == 0 && dS == 0 && dC == 0 && dA == 0 {
				return i
			}
		}
	}
	return len(src)
}

// splitTopLevel splits text on the separator at zero nesting depth.
// `=>` arrows are treated as a single token so callback types don't break
// angle balance.
func splitTopLevel(src string, sep byte) []string {
	var out []string
	dR, dS, dC, dA := 0, 0, 0, 0
	start := 0
	for i := 0; i < len(src); i++ {
		c := src[i]
		switch c {
		case '=':
			if i+1 < len(src) && src[i+1] == '>' {
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
			i = skipString(src, i, c)
		case sep:
			if dR == 0 && dS == 0 && dC == 0 && dA == 0 {
				out = append(out, strings.TrimSpace(src[start:i]))
				start = i + 1
			}
		}
	}
	if last := strings.TrimSpace(src[start:]); last != "" {
		out = append(out, last)
	}
	return out
}
