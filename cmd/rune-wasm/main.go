//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"syscall/js"
	"time"

	"github.com/oboard/rune-lang/internal/compiler"
	runefmt "github.com/oboard/rune-lang/internal/format"
	"github.com/oboard/rune-lang/internal/parser"
	"github.com/oboard/rune-lang/internal/stdlib"
)

type bridge struct {
	registry *stdlib.Registry
}

type response struct {
	OK          bool         `json:"ok"`
	Error       string       `json:"error,omitempty"`
	Diagnostics []diagnostic `json:"diagnostics,omitempty"`
	TypeScript  string       `json:"typescript,omitempty"`
	Formatted   string       `json:"formatted,omitempty"`
	ElapsedMS   float64      `json:"elapsedMs,omitempty"`
}

type diagnostic struct {
	Message string `json:"message"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
}

func main() {
	b := &bridge{}
	js.Global().Set("runeInitCompiler", js.FuncOf(b.initCompiler))
	js.Global().Set("runeCompile", js.FuncOf(b.compile))
	js.Global().Set("runeFormat", js.FuncOf(b.format))
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
		b.registry = reg
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
