package format

import "strings"

func preserveLineComments(source string, formatted string) string {
	inlineComments := map[string][]string{}
	leadingComments := map[string][][]string{}
	var leadingKeys []string
	var inlineKeys []string
	var pendingLeading []string

	for _, line := range strings.Split(source, "\n") {
		commentStart := findLineComment(line)
		if commentStart < 0 {
			if strings.TrimSpace(line) != "" {
				key := canonicalLineKey(line)
				if len(pendingLeading) > 0 {
					leadingComments[key] = append(leadingComments[key], pendingLeading)
					if !containsKey(leadingKeys, key) {
						leadingKeys = append(leadingKeys, key)
					}
					pendingLeading = nil
				}
			}
			continue
		}

		code := line[:commentStart]
		comment := strings.TrimSpace(line[commentStart:])
		if strings.TrimSpace(code) == "" {
			pendingLeading = append(pendingLeading, leadingWhitespace(line)+comment)
			continue
		}

		key := canonicalLineKey(code)
		if len(pendingLeading) > 0 {
			leadingComments[key] = append(leadingComments[key], pendingLeading)
			if !containsKey(leadingKeys, key) {
				leadingKeys = append(leadingKeys, key)
			}
			pendingLeading = nil
		}
		inlineComments[key] = append(inlineComments[key], comment)
		if !containsKey(inlineKeys, key) {
			inlineKeys = append(inlineKeys, key)
		}
	}

	hadTrailingNewline := strings.HasSuffix(formatted, "\n")
	formatted = strings.TrimSuffix(formatted, "\n")
	var out []string
	var closeComments []pendingCloseComment
	if formatted != "" {
		for _, line := range strings.Split(formatted, "\n") {
			key := canonicalLineKey(line)
			if groups, matchedKey := takeLeadingComments(leadingComments, leadingKeys, key); len(groups) > 0 {
				for _, comment := range groups[0] {
					out = append(out, formatLeadingComment(line, comment))
				}
				leadingComments[matchedKey] = groups[1:]
			}
			if comments := inlineComments[key]; len(comments) > 0 {
				line += " " + comments[0]
				inlineComments[key] = comments[1:]
			} else if closeComment, ok := takeCloseComment(closeComments, line); ok {
				line += " " + closeComment
				closeComments = closeComments[:len(closeComments)-1]
			} else if comment, matchedKey, ok := takeChainContinuationInlineComment(inlineComments, inlineKeys, key); ok {
				line += " " + comment
				inlineComments[matchedKey] = inlineComments[matchedKey][1:]
			} else if comment, matchedKey, ok := takeExpandedInlineComment(inlineComments, inlineKeys, key); ok {
				inlineComments[matchedKey] = inlineComments[matchedKey][1:]
				closeComments = append(closeComments, pendingCloseComment{
					indent:  leadingWhitespace(line),
					comment: comment,
				})
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

type pendingCloseComment struct {
	indent  string
	comment string
}

func formatLeadingComment(line string, comment string) string {
	originalIndent := leadingWhitespace(comment)
	if strings.TrimSpace(line) == "}" && originalIndent != "" {
		return comment
	}
	return leadingWhitespace(line) + strings.TrimLeft(comment, " \t")
}

func takeLeadingComments(comments map[string][][]string, keys []string, formattedKey string) ([][]string, string) {
	if groups := comments[formattedKey]; len(groups) > 0 {
		return groups, formattedKey
	}
	for _, key := range keys {
		if key == formattedKey {
			continue
		}
		if isExpandedLineKey(key, formattedKey) {
			if groups := comments[key]; len(groups) > 0 {
				return groups, key
			}
		}
	}
	if anchor, ok := declarationAnchor(formattedKey); ok {
		for _, key := range keys {
			if key == formattedKey {
				continue
			}
			keyAnchor, keyOK := declarationAnchor(key)
			if !keyOK || keyAnchor != anchor {
				continue
			}
			if groups := comments[key]; len(groups) > 0 {
				return groups, key
			}
		}
	}
	return nil, ""
}

func takeExpandedInlineComment(comments map[string][]string, keys []string, formattedKey string) (string, string, bool) {
	for _, key := range keys {
		if isExpandedLineKey(key, formattedKey) {
			if values := comments[key]; len(values) > 0 {
				return values[0], key, true
			}
		}
	}
	return "", "", false
}

func takeChainContinuationInlineComment(comments map[string][]string, keys []string, formattedKey string) (string, string, bool) {
	if !strings.HasPrefix(formattedKey, ".") {
		return "", "", false
	}
	for _, key := range keys {
		if key == formattedKey || !strings.HasSuffix(key, formattedKey) {
			continue
		}
		if values := comments[key]; len(values) > 0 {
			return values[0], key, true
		}
	}
	return "", "", false
}

func takeCloseComment(comments []pendingCloseComment, line string) (string, bool) {
	if len(comments) == 0 || strings.TrimSpace(line) != "}" {
		return "", false
	}
	last := comments[len(comments)-1]
	if leadingWhitespace(line) != last.indent {
		return "", false
	}
	return last.comment, true
}

func isExpandedLineKey(originalKey string, formattedKey string) bool {
	return strings.HasSuffix(formattedKey, "{") &&
		strings.HasPrefix(originalKey, formattedKey) &&
		len(originalKey) > len(formattedKey)
}

func declarationAnchor(key string) (string, bool) {
	open := strings.IndexByte(key, '(')
	if open <= 0 {
		return "", false
	}
	head := key[:open]
	if !validDeclarationAnchorHead(head) {
		return "", false
	}
	tail := key[open:]
	if strings.HasSuffix(key, "(") || strings.Contains(tail, ")=>") || strings.Contains(tail, ")->") {
		return head, true
	}
	return "", false
}

func validDeclarationAnchorHead(head string) bool {
	for i, ch := range head {
		if i == 0 && !isIdentStart(ch) {
			return false
		}
		if isIdentContinue(ch) || ch == '[' || ch == ']' || ch == ',' {
			continue
		}
		return false
	}
	return true
}

func isIdentStart(ch rune) bool {
	return ch == '_' || ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z')
}

func isIdentContinue(ch rune) bool {
	return isIdentStart(ch) || ('0' <= ch && ch <= '9')
}

func containsKey(keys []string, key string) bool {
	for _, existing := range keys {
		if existing == key {
			return true
		}
	}
	return false
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
