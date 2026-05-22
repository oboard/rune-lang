package format

import "strings"

func preserveLineComments(source string, formatted string) string {
	inlineComments := map[string][]string{}
	leadingComments := map[string][][]string{}
	var pendingLeading []string

	for _, line := range strings.Split(source, "\n") {
		commentStart := findLineComment(line)
		if commentStart < 0 {
			if strings.TrimSpace(line) != "" {
				key := canonicalLineKey(line)
				if len(pendingLeading) > 0 {
					leadingComments[key] = append(leadingComments[key], pendingLeading)
					pendingLeading = nil
				}
			}
			continue
		}

		code := line[:commentStart]
		comment := strings.TrimSpace(line[commentStart:])
		if strings.TrimSpace(code) == "" {
			pendingLeading = append(pendingLeading, comment)
			continue
		}

		key := canonicalLineKey(code)
		if len(pendingLeading) > 0 {
			leadingComments[key] = append(leadingComments[key], pendingLeading)
			pendingLeading = nil
		}
		inlineComments[key] = append(inlineComments[key], comment)
	}

	hadTrailingNewline := strings.HasSuffix(formatted, "\n")
	formatted = strings.TrimSuffix(formatted, "\n")
	var out []string
	if formatted != "" {
		for _, line := range strings.Split(formatted, "\n") {
			key := canonicalLineKey(line)
			if groups := leadingComments[key]; len(groups) > 0 {
				for _, comment := range groups[0] {
					out = append(out, leadingWhitespace(line)+comment)
				}
				leadingComments[key] = groups[1:]
			}
			if comments := inlineComments[key]; len(comments) > 0 {
				line += " " + comments[0]
				inlineComments[key] = comments[1:]
			}
			out = append(out, line)
		}
	}

	for _, groups := range leadingComments {
		for _, group := range groups {
			out = appendPendingComments(out, group)
		}
	}
	out = appendPendingComments(out, pendingLeading)
	for _, comments := range inlineComments {
		out = appendPendingComments(out, comments)
	}

	result := strings.Join(out, "\n")
	if hadTrailingNewline {
		result += "\n"
	}
	return result
}

func appendPendingComments(out []string, comments []string) []string {
	for _, comment := range comments {
		if comment != "" {
			out = append(out, comment)
		}
	}
	return out
}

func findLineComment(line string) int {
	inString := false
	escaped := false
	for i := 0; i < len(line)-1; i++ {
		ch := line[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}
		if ch == '"' {
			inString = true
			continue
		}
		if ch == '/' && line[i+1] == '/' {
			return i
		}
	}
	return -1
}

func canonicalLineKey(line string) string {
	key := removeWhitespaceOutsideStrings(strings.TrimSpace(line))
	return strings.TrimSuffix(key, ",")
}

func removeWhitespaceOutsideStrings(line string) string {
	var b strings.Builder
	inString := false
	escaped := false
	for _, ch := range line {
		if inString {
			b.WriteRune(ch)
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}
		if ch == '"' {
			inString = true
			b.WriteRune(ch)
			continue
		}
		if ch == ' ' || ch == '\t' || ch == '\r' {
			continue
		}
		b.WriteRune(ch)
	}
	return b.String()
}

func leadingWhitespace(line string) string {
	return line[:len(line)-len(strings.TrimLeft(line, " \t"))]
}
