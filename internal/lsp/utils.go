package lsp

import (
	"net/url"
	"path/filepath"
	"strings"
	"unicode/utf16"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/lexer"
)

func containsSymbol(pos position, start lexer.Position, name string) bool {
	line := start.Line - 1
	char := start.Column - 1
	return pos.Line == line && pos.Character >= char && pos.Character <= char+len(name)
}

func containsToken(pos position, tok lexer.Token) bool {
	line := tok.Pos.Line - 1
	char := tok.Pos.Column - 1
	return pos.Line == line && pos.Character >= char && pos.Character <= char+len([]rune(tok.Lexeme))
}

func offsetFromPosition(text string, pos position) (int, bool) {
	if pos.Line < 0 || pos.Character < 0 {
		return 0, false
	}
	line := 0
	lineStart := 0
	for i, ch := range text {
		if line == pos.Line {
			lineBytes := text[lineStart:]
			if idx := strings.Index(lineBytes, "\n"); idx >= 0 {
				lineBytes = lineBytes[:idx]
			}
			runes := []rune(lineBytes)
			maxUTF16Chars := len(utf16.Encode(runes))
			if pos.Character >= maxUTF16Chars {
				return lineStart + len(lineBytes), true
			}
			var byteOffset int
			utf16Count := 0
			for _, r := range runes {
				if utf16Count >= pos.Character {
					break
				}
				encoded := utf16.Encode([]rune{r})
				utf16Count += len(encoded)
				byteOffset += len(string([]rune{r}))
			}
			if lineStart+byteOffset > len(text) {
				return len(text), true
			}
			return lineStart + byteOffset, true
		}
		if ch == '\n' {
			line++
			lineStart = i + 1
		}
	}
	if line == pos.Line {
		lineBytes := text[lineStart:]
		runes := []rune(lineBytes)
		maxUTF16Chars := len(utf16.Encode(runes))
		if pos.Character >= maxUTF16Chars {
			return lineStart + len(lineBytes), true
		}
		var byteOffset int
		utf16Count := 0
		for _, r := range runes {
			if utf16Count >= pos.Character {
				break
			}
			encoded := utf16.Encode([]rune{r})
			utf16Count += len(encoded)
			byteOffset += len(string([]rune{r}))
		}
		if lineStart+byteOffset > len(text) {
			return len(text), true
		}
		return lineStart + byteOffset, true
	}
	return len(text), false
}

func isIdentByte(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_'
}

func textEdit(pos lexer.Position, oldName string, newName string) map[string]any {
	return map[string]any{
		"range":   symbolRange(pos, len(oldName)),
		"newText": newName,
	}
}

func baseType(typ checker.Type) string {
	name := string(typ)
	if strings.HasPrefix(name, "{") && strings.HasSuffix(name, "}") {
		return name
	}
	if i := strings.IndexByte(name, '['); i >= 0 {
		return name[:i]
	}
	return name
}

func stdlibReceiverModule(receiver checker.Type) (string, string, bool) {
	if _, ok := checker.ArrayElement(receiver); ok {
		return "array", "Array", true
	}
	return checker.StdlibReceiverModule(receiver)
}

func fileURI(path string) string {
	return (&url.URL{Scheme: "file", Path: path}).String()
}

func filePathFromURI(uri string) string {
	parsed, err := url.Parse(uri)
	if err != nil || parsed.Scheme != "file" {
		return ""
	}
	return parsed.Path
}

func sourceURI(defaultURI string, sourcePath string) string {
	if sourcePath == "" {
		return defaultURI
	}
	return fileURI(sourcePath)
}

func sourceMatchesDocument(uri string, sourcePath string) bool {
	if sourcePath == "" {
		return true
	}
	docPath := filePathFromURI(uri)
	if docPath == "" {
		return false
	}
	docPath = filepath.Clean(docPath)
	sourcePath = filepath.Clean(sourcePath)
	if abs, err := filepath.Abs(docPath); err == nil {
		docPath = abs
	}
	if abs, err := filepath.Abs(sourcePath); err == nil {
		sourcePath = abs
	}
	return docPath == sourcePath
}
