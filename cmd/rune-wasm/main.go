//go:build js && wasm

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"syscall/js"
	"time"

	"github.com/oboard/rune-lang/internal/compiler"
	runefmt "github.com/oboard/rune-lang/internal/format"
	"github.com/oboard/rune-lang/internal/interpreter"
	"github.com/oboard/rune-lang/internal/lsp"
	"github.com/oboard/rune-lang/internal/parser"
	"github.com/oboard/rune-lang/internal/stdlib"
)

type bridge struct {
	registry *stdlib.Registry
	lsp      *lsp.Session
}

type response struct {
	OK          bool         `json:"ok"`
	Error       string       `json:"error,omitempty"`
	Diagnostics []diagnostic `json:"diagnostics,omitempty"`
	TypeScript  string       `json:"typescript,omitempty"`
	Formatted   string       `json:"formatted,omitempty"`
	Tests       []testResult `json:"tests,omitempty"`
	LSP         any          `json:"lsp,omitempty"`
	ElapsedMS   float64      `json:"elapsedMs,omitempty"`
}

type diagnostic struct {
	Message string `json:"message"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
}

type lspRequest struct {
	Method             string `json:"method"`
	URI                string `json:"uri"`
	Source             string `json:"source"`
	Line               int    `json:"line"`
	Character          int    `json:"character"`
	NewName            string `json:"newName"`
	IncludeDeclaration bool   `json:"includeDeclaration"`
	TabSize            int    `json:"tabSize"`
	InsertSpaces       bool   `json:"insertSpaces"`
}

type testRequest struct {
	URI    string `json:"uri"`
	Source string `json:"source"`
	Name   string `json:"name"`
}

type testResult struct {
	Name      string  `json:"name"`
	Passed    bool    `json:"passed"`
	Error     string  `json:"error,omitempty"`
	Output    string  `json:"output,omitempty"`
	ElapsedMS float64 `json:"elapsedMs,omitempty"`
}

func main() {
	b := &bridge{}
	js.Global().Set("runeInitCompiler", js.FuncOf(b.initCompiler))
	js.Global().Set("runeCompile", js.FuncOf(b.compile))
	js.Global().Set("runeFormat", js.FuncOf(b.format))
	js.Global().Set("runeLSP", js.FuncOf(b.handleLSP))
	js.Global().Set("runeTest", js.FuncOf(b.runTest))
	select {}
}

func (b *bridge) initCompiler(_ js.Value, args []js.Value) any {
	return safeJSON(func() response {
		if len(args) != 1 || args[0].Type() != js.TypeString {
			return response{OK: false, Error: "runeInitCompiler expects a JSON object of core sources"}
		}
		var sources map[string]string
		if err := json.Unmarshal([]byte(args[0].String()), &sources); err != nil {
			return response{OK: false, Error: err.Error()}
		}
		reg, err := stdlib.LoadSources(sources)
		if err != nil {
			return response{OK: false, Error: err.Error()}
		}
		stdlib.SetDefault(reg)
		b.registry = reg
		b.lsp = lsp.NewSession()
		return response{OK: true}
	})
}

func (b *bridge) compile(_ js.Value, args []js.Value) any {
	return safeJSON(func() response {
		start := time.Now()
		if b.registry == nil {
			return response{OK: false, Error: "Rune compiler has not been initialized"}
		}
		if len(args) != 1 || args[0].Type() != js.TypeString {
			return response{OK: false, Error: "runeCompile expects source text"}
		}
		src := args[0].String()
		ts, diags := compiler.GenerateTypeScriptSource("playground.rn", src, b.registry)
		out := response{
			OK:          len(diags) == 0,
			Diagnostics: convertDiagnostics(diags),
			TypeScript:  ts,
			ElapsedMS:   float64(time.Since(start).Microseconds()) / 1000,
		}
		if len(diags) > 0 {
			out.Error = "compile failed"
		}
		return out
	})
}

func (b *bridge) format(_ js.Value, args []js.Value) any {
	return safeJSON(func() response {
		start := time.Now()
		if len(args) != 1 || args[0].Type() != js.TypeString {
			return response{OK: false, Error: "runeFormat expects source text"}
		}
		src := args[0].String()
		file, errs := parser.Parse(src)
		if len(errs) > 0 {
			diags := make([]compiler.Diagnostic, 0, len(errs))
			for _, err := range errs {
				diags = append(diags, compiler.Diagnostic{Message: err.Message, Pos: err.Pos})
			}
			return response{
				OK:          false,
				Error:       "format failed",
				Diagnostics: convertDiagnostics(diags),
				ElapsedMS:   float64(time.Since(start).Microseconds()) / 1000,
			}
		}
		return response{
			OK:        true,
			Formatted: runefmt.Source(file, src),
			ElapsedMS: float64(time.Since(start).Microseconds()) / 1000,
		}
	})
}

func (b *bridge) runTest(_ js.Value, args []js.Value) any {
	return safeJSON(func() response {
		start := time.Now()
		if b.registry == nil {
			return response{OK: false, Error: "Rune compiler has not been initialized"}
		}
		if len(args) != 1 || args[0].Type() != js.TypeString {
			return response{OK: false, Error: "runeTest expects a JSON request"}
		}
		var req testRequest
		if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
			return response{OK: false, Error: err.Error()}
		}
		uri := req.URI
		if uri == "" {
			uri = "playground.rn"
		}
		prog, diags := compiler.AnalyzeSourceWithStdlib(uri, req.Source, b.registry)
		out := response{
			OK:          len(diags) == 0,
			Diagnostics: convertDiagnostics(diags),
			ElapsedMS:   float64(time.Since(start).Microseconds()) / 1000,
		}
		if len(diags) > 0 {
			out.Error = "compile failed"
			return out
		}
		for _, test := range prog.IR.Tests {
			if req.Name != "" && test.Name != req.Name {
				continue
			}
			var testOutput bytes.Buffer
			testStart := time.Now()
			err := interpreter.New(prog.IR, interpreter.WithOutput(&testOutput)).RunTest(test)
			result := testResult{
				Name:      test.Name,
				Passed:    err == nil,
				Output:    testOutput.String(),
				ElapsedMS: float64(time.Since(testStart).Microseconds()) / 1000,
			}
			if err != nil {
				result.Error = err.Error()
				out.OK = false
			}
			out.Tests = append(out.Tests, result)
		}
		if len(out.Tests) == 0 {
			out.OK = false
			if req.Name == "" {
				out.Error = "no tests found"
			} else {
				out.Error = "test not found: " + req.Name
			}
		}
		out.ElapsedMS = float64(time.Since(start).Microseconds()) / 1000
		return out
	})
}

func (b *bridge) handleLSP(_ js.Value, args []js.Value) any {
	return safeJSON(func() response {
		start := time.Now()
		if b.lsp == nil {
			return response{OK: false, Error: "Rune LSP has not been initialized"}
		}
		if len(args) != 1 || args[0].Type() != js.TypeString {
			return response{OK: false, Error: "runeLSP expects a JSON request"}
		}
		var req lspRequest
		if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
			return response{OK: false, Error: err.Error()}
		}
		uri := req.URI
		if uri == "" {
			uri = "file:///playground.rn"
		}
		b.lsp.SetDocument(uri, req.Source)
		out := response{OK: true, ElapsedMS: float64(time.Since(start).Microseconds()) / 1000}
		switch req.Method {
		case "diagnostics":
			out.LSP = b.lsp.Diagnostics(uri)
		case "hover":
			out.LSP = b.lsp.Hover(uri, req.Line, req.Character)
		case "completion":
			out.LSP = b.lsp.Completion(uri, req.Line, req.Character)
		case "definition":
			out.LSP = b.lsp.Definition(uri, req.Line, req.Character)
		case "references":
			out.LSP = b.lsp.References(uri, req.Line, req.Character, req.IncludeDeclaration)
		case "codeLens":
			out.LSP = b.lsp.CodeLenses(uri)
		case "documentSymbol":
			out.LSP = b.lsp.DocumentSymbols(uri)
		case "formatting":
			tabSize := req.TabSize
			if tabSize <= 0 {
				tabSize = 2
			}
			out.LSP = b.lsp.Formatting(uri, tabSize, req.InsertSpaces)
		case "inlayHint":
			out.LSP = b.lsp.InlayHints(uri)
		case "semanticTokens":
			out.LSP = b.lsp.SemanticTokens(uri)
		case "rename":
			out.LSP = b.lsp.Rename(uri, req.Line, req.Character, req.NewName)
		default:
			out.OK = false
			out.Error = "unknown LSP method: " + req.Method
		}
		out.ElapsedMS = float64(time.Since(start).Microseconds()) / 1000
		return out
	})
}

func safeJSON(fn func() response) (out string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			out = marshal(response{
				OK:    false,
				Error: fmt.Sprintf("%v\n%s", recovered, debug.Stack()),
			})
		}
	}()
	return marshal(fn())
}

func marshal(resp response) string {
	data, err := json.Marshal(resp)
	if err != nil {
		fallback, _ := json.Marshal(response{OK: false, Error: err.Error()})
		return string(fallback)
	}
	return string(data)
}

func convertDiagnostics(diags []compiler.Diagnostic) []diagnostic {
	if len(diags) == 0 {
		return nil
	}
	out := make([]diagnostic, 0, len(diags))
	for _, diag := range diags {
		out = append(out, diagnostic{
			Message: diag.Message,
			Line:    diag.Pos.Line,
			Column:  diag.Pos.Column,
		})
	}
	return out
}
