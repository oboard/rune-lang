package lsp

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/compiler"
)

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
