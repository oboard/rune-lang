package lsp

import (
	"regexp"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/compiler"
)

func (s *server) rename(uri string, pos position, newName string) any {
	text := s.docs[uri]
	word := wordAt(text, pos)
	if word == "" {
		return nil
	}
	prog, _ := s.analyze(uri)
	if prog != nil {
		if target := typeTarget(uri, prog, pos); target != nil {
			return map[string]any{
				"changes": map[string]any{uri: wordRenameEdits(text, target.name, newName)},
			}
		}
		if target := s.methodTarget(uri, prog, pos); target != nil {
			if target.structName == "" {
				return nil
			}
			return map[string]any{
				"changes": map[string]any{uri: methodRenameEdits(prog, target, newName)},
			}
		}
		if target := functionTarget(uri, prog, pos); target != nil && target.external {
			return nil
		}
		if target := externalValueTarget(uri, prog, pos); target != nil {
			return nil
		}
	}
	return map[string]any{
		"changes": map[string]any{uri: wordRenameEdits(text, word, newName)},
	}
}

func wordRenameEdits(text string, oldName string, newName string) []map[string]any {
	ident := regexp.MustCompile(`\b` + regexp.QuoteMeta(oldName) + `\b`)
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
	return edits
}

func methodRenameEdits(prog *compiler.Program, target *methodTarget, newName string) []map[string]any {
	var edits []map[string]any
	for _, typ := range prog.File.Types {
		if typ.Name != target.structName {
			continue
		}
		for _, method := range typ.Methods {
			if method.Name == target.name {
				edits = append(edits, textEdit(method.NamePos, method.Name, newName))
			}
		}
	}
	walkFileSelectors(prog.File, func(sel *ast.SelectorExpr) {
		if sel.Name != target.name {
			return
		}
		if baseType(prog.Info.ExprTypes[sel.Receiver]) != target.structName {
			return
		}
		edits = append(edits, textEdit(sel.NamePos, sel.Name, newName))
	})
	return edits
}
