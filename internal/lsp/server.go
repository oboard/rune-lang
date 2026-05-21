package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/compiler"
	"github.com/oboard/rune-lang/internal/lexer"
)

func Serve(in io.Reader, out io.Writer) error {
	s := &server{
		reader: bufio.NewReader(in),
		out:    out,
		docs:   map[string]string{},
	}
	for {
		msg, err := s.readMessage()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if err := s.handle(msg); err != nil {
			return err
		}
	}
}

type server struct {
	reader *bufio.Reader
	out    io.Writer
	docs   map[string]string
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

func (s *server) handle(req request) error {
	switch req.Method {
	case "initialize":
		return s.respond(req.ID, map[string]any{
			"capabilities": map[string]any{
				"textDocumentSync":       1,
				"hoverProvider":          true,
				"completionProvider":     map[string]any{"triggerCharacters": []string{"@", "."}},
				"definitionProvider":     true,
				"documentSymbolProvider": true,
				"renameProvider":         true,
			},
			"serverInfo": map[string]any{
				"name":    "rune-lsp",
				"version": "0.1.0",
			},
		})
	case "shutdown":
		return s.respond(req.ID, nil)
	case "exit":
		return io.EOF
	case "textDocument/didOpen":
		var params didOpenParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return err
		}
		s.docs[params.TextDocument.URI] = params.TextDocument.Text
		return s.publishDiagnostics(params.TextDocument.URI)
	case "textDocument/didChange":
		var params didChangeParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return err
		}
		if len(params.ContentChanges) > 0 {
			s.docs[params.TextDocument.URI] = params.ContentChanges[len(params.ContentChanges)-1].Text
		}
		return s.publishDiagnostics(params.TextDocument.URI)
	case "textDocument/hover":
		var params textPositionParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return err
		}
		return s.respond(req.ID, s.hover(params.TextDocument.URI, params.Position))
	case "textDocument/completion":
		var params textPositionParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return err
		}
		return s.respond(req.ID, s.completion(params.TextDocument.URI))
	case "textDocument/definition":
		var params textPositionParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return err
		}
		return s.respond(req.ID, s.definition(params.TextDocument.URI, params.Position))
	case "textDocument/documentSymbol":
		var params documentSymbolParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return err
		}
		return s.respond(req.ID, s.documentSymbols(params.TextDocument.URI))
	case "textDocument/rename":
		var params renameParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return err
		}
		return s.respond(req.ID, s.rename(params.TextDocument.URI, params.Position, params.NewName))
	default:
		if len(req.ID) > 0 {
			return s.respond(req.ID, nil)
		}
		return nil
	}
}

func (s *server) publishDiagnostics(uri string) error {
	_, diags := compiler.AnalyzeSource(uri, s.docs[uri])
	items := make([]map[string]any, 0, len(diags))
	for _, diag := range diags {
		items = append(items, map[string]any{
			"range":    lspRange(diag.Pos),
			"severity": 1,
			"source":   "rune",
			"message":  diag.Message,
		})
	}
	return s.notify("textDocument/publishDiagnostics", map[string]any{
		"uri":         uri,
		"diagnostics": items,
	})
}

func (s *server) hover(uri string, pos position) any {
	prog, _ := compiler.AnalyzeSource(uri, s.docs[uri])
	word := wordAt(s.docs[uri], pos)
	if word == "" || prog == nil {
		return nil
	}
	for _, fn := range prog.File.Functions {
		if fn.Name == word {
			ret := checker.Void
			if info := prog.Info.Functions[fn.Name]; info != nil {
				ret = info.Return
			}
			return map[string]any{
				"contents": map[string]any{
					"kind":  "markdown",
					"value": fmt.Sprintf("```rune\n%s -> %s\n```", fn.Signature(), ret),
				},
				"range": symbolRange(fn.NamePos, len(fn.Name)),
			}
		}
		for _, param := range fn.Params {
			if param.Name == word {
				return map[string]any{
					"contents": map[string]any{
						"kind":  "markdown",
						"value": fmt.Sprintf("```rune\n%s: %s\n```", param.Name, param.Type),
					},
					"range": symbolRange(param.Pos, len(param.Name)),
				}
			}
		}
	}
	return nil
}

func (s *server) completion(uri string) any {
	prog, _ := compiler.AnalyzeSource(uri, s.docs[uri])
	var items []map[string]any
	if prog == nil {
		return items
	}
	if prog.Info.Stdlib != nil {
		for _, moduleName := range prog.Info.Stdlib.ModuleNames() {
			module := prog.Info.Stdlib.Modules[moduleName]
			for _, fn := range module.Functions {
				items = append(items, map[string]any{
					"label":  "@" + moduleName + "." + fn.Name,
					"kind":   3,
					"detail": "core/" + moduleName,
				})
			}
		}
	}
	for _, fn := range prog.File.Functions {
		items = append(items, map[string]any{
			"label":  fn.Name,
			"kind":   3,
			"detail": fn.Signature(),
		})
	}
	return items
}

func (s *server) definition(uri string, pos position) any {
	prog, _ := compiler.AnalyzeSource(uri, s.docs[uri])
	word := wordAt(s.docs[uri], pos)
	if word == "" || prog == nil {
		return nil
	}
	for _, fn := range prog.File.Functions {
		if fn.Name == word {
			return map[string]any{
				"uri":   uri,
				"range": symbolRange(fn.NamePos, len(fn.Name)),
			}
		}
	}
	return nil
}

func (s *server) documentSymbols(uri string) any {
	prog, _ := compiler.AnalyzeSource(uri, s.docs[uri])
	if prog == nil {
		return []any{}
	}
	items := make([]map[string]any, 0, len(prog.File.Functions))
	for _, fn := range prog.File.Functions {
		rng := functionRange(fn)
		items = append(items, map[string]any{
			"name":           fn.Name,
			"kind":           12,
			"range":          rng,
			"selectionRange": symbolRange(fn.NamePos, len(fn.Name)),
			"detail":         fn.Signature(),
		})
	}
	return items
}

func (s *server) rename(uri string, pos position, newName string) any {
	text := s.docs[uri]
	word := wordAt(text, pos)
	if word == "" {
		return nil
	}
	ident := regexp.MustCompile(`\b` + regexp.QuoteMeta(word) + `\b`)
	var edits []map[string]any
	lines := strings.Split(text, "\n")
	for lineNo, line := range lines {
		for _, loc := range ident.FindAllStringIndex(line, -1) {
			edits = append(edits, map[string]any{
				"range": map[string]any{
					"start": position{Line: lineNo, Character: loc[0]},
					"end":   position{Line: lineNo, Character: loc[1]},
				},
				"newText": newName,
			})
		}
	}
	return map[string]any{
		"changes": map[string]any{uri: edits},
	}
}

func (s *server) readMessage() (request, error) {
	contentLength := -1
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return request{}, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(parts[0]), "Content-Length") {
			n, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return request{}, err
			}
			contentLength = n
		}
	}
	if contentLength < 0 {
		return request{}, fmt.Errorf("missing Content-Length header")
	}
	body := make([]byte, contentLength)
	if _, err := io.ReadFull(s.reader, body); err != nil {
		return request{}, err
	}
	var req request
	if err := json.Unmarshal(body, &req); err != nil {
		return request{}, err
	}
	return req, nil
}

func (s *server) respond(id json.RawMessage, result any) error {
	if len(id) == 0 {
		return nil
	}
	payload := map[string]any{
		"jsonrpc": "2.0",
		"id":      json.RawMessage(id),
		"result":  result,
	}
	return s.write(payload)
}

func (s *server) notify(method string, params any) error {
	return s.write(map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	})
}

func (s *server) write(payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(s.out, "Content-Length: %d\r\n\r\n%s", len(data), data)
	return err
}

type textDocumentIdentifier struct {
	URI string `json:"uri"`
}

type textDocumentItem struct {
	URI  string `json:"uri"`
	Text string `json:"text"`
}

type didOpenParams struct {
	TextDocument textDocumentItem `json:"textDocument"`
}

type didChangeParams struct {
	TextDocument   textDocumentIdentifier `json:"textDocument"`
	ContentChanges []struct {
		Text string `json:"text"`
	} `json:"contentChanges"`
}

type textPositionParams struct {
	TextDocument textDocumentIdentifier `json:"textDocument"`
	Position     position               `json:"position"`
}

type documentSymbolParams struct {
	TextDocument textDocumentIdentifier `json:"textDocument"`
}

type renameParams struct {
	TextDocument textDocumentIdentifier `json:"textDocument"`
	Position     position               `json:"position"`
	NewName      string                 `json:"newName"`
}

type position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

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

func wordAt(text string, pos position) string {
	lines := strings.Split(text, "\n")
	if pos.Line < 0 || pos.Line >= len(lines) {
		return ""
	}
	line := []rune(lines[pos.Line])
	if len(line) == 0 {
		return ""
	}
	idx := min(max(pos.Character, 0), len(line)-1)
	if !isWord(line[idx]) && idx > 0 {
		idx--
	}
	if !isWord(line[idx]) {
		return ""
	}
	start := idx
	for start > 0 && isWord(line[start-1]) {
		start--
	}
	end := idx + 1
	for end < len(line) && isWord(line[end]) {
		end++
	}
	return string(line[start:end])
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
