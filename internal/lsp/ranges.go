package lsp

import (
	"strings"
	"unicode/utf16"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

func lspRange(pos lexer.Position) map[string]any {
	return symbolRange(pos, 1)
}

func symbolRange(pos lexer.Position, length int) map[string]any {
	line := max(pos.Line-1, 0)
	char := max(pos.Column-1, 0)
	return map[string]any{
		"start": position{Line: line, Character: char},
		"end":   position{Line: line, Character: char + max(length, 1)},
	}
}

func functionRange(fn *ast.Function) map[string]any {
	start := fn.NamePos
	end := fn.Body.Position()
	return map[string]any{
		"start": position{Line: max(start.Line-1, 0), Character: max(start.Column-1, 0)},
		"end":   position{Line: max(end.Line-1, 0), Character: max(end.Column, 1)},
	}
}

func fatArrowPosition(text string, fn *ast.Function) (position, bool) {
	return fatArrowPositionFromOffset(text, fn.NamePos.Offset)
}

func fatArrowPositionFromOffset(text string, offset int) (position, bool) {
	start := max(offset, 0)
	if start >= len(text) {
		start = 0
	}
	idx := strings.Index(text[start:], "=>")
	if idx < 0 {
		return position{}, false
	}
	return positionAtOffset(text, start+idx), true
}

func positionAtOffset(text string, byteOffset int) position {
	line := 0
	lineStart := 0
	for i, ch := range text {
		if i >= byteOffset {
			break
		}
		if ch == '\n' {
			line++
			lineStart = i + 1
		}
	}
	lineText := text[lineStart:]
	if idx := strings.Index(lineText, "\n"); idx >= 0 {
		lineText = lineText[:idx]
	}
	runes := []rune(lineText[:byteOffset-lineStart])
	charOffset := len(utf16.Encode(runes))
	return position{Line: line, Character: charOffset}
}

func wordAt(text string, pos position) string {
	lines := strings.Split(text, "\n")
	if pos.Line < 0 || pos.Line >= len(lines) {
		return ""
	}
	line := lines[pos.Line]
	utf16Chars := utf16.Encode([]rune(line))
	if len(utf16Chars) == 0 {
		return ""
	}
	idx := min(max(pos.Character, 0), len(utf16Chars)-1)
	runes := utf16.Decode(utf16Chars)
	if len(runes) == 0 {
		return ""
	}
	if !isWord(runes[idx]) && idx > 0 {
		idx--
	}
	if idx >= len(runes) || !isWord(runes[idx]) {
		return ""
	}
	start := idx
	for start > 0 && isWord(runes[start-1]) {
		start--
	}
	end := idx + 1
	for end < len(runes) && isWord(runes[end]) {
		end++
	}
	return string(runes[start:end])
}

func isWord(ch rune) bool {
	return ch == '_' || ch >= '0' && ch <= '9' || ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z'
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

