package lsp

import (
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/format"
)

func (s *server) expandedMacro(uri string) any {
	prog, diags := s.analyze(uri)
	if prog == nil {
		return map[string]any{"error": "macro expansion is unavailable"}
	}
	if len(diags) > 0 {
		messages := make([]string, 0, len(diags))
		for _, diag := range diags {
			messages = append(messages, diag.Message)
		}
		return map[string]any{"error": strings.Join(messages, "\n")}
	}
	return map[string]any{
		"source": format.File(expandedDocumentFile(uri, prog.File)),
	}
}

func expandedDocumentFile(uri string, file *ast.File) *ast.File {
	out := &ast.File{
		GoImports: append([]ast.GoImport(nil), file.GoImports...),
		TSImports: append([]ast.TSImport(nil), file.TSImports...),
	}
	for _, typ := range file.Types {
		if !sourceMatchesDocument(uri, typ.SourcePath) {
			continue
		}
		clone := *typ
		clone.Annotations = nil
		clone.Fields = append([]ast.Field(nil), typ.Fields...)
		for i := range clone.Fields {
			clone.Fields[i].Annotations = nil
		}
		clone.Methods = cloneExpandedFunctions(typ.Methods)
		out.Types = append(out.Types, &clone)
	}
	for _, enum := range file.Enums {
		if !sourceMatchesDocument(uri, enum.SourcePath) {
			continue
		}
		clone := *enum
		clone.Annotations = nil
		clone.Members = append([]ast.EnumMember(nil), enum.Members...)
		for i := range clone.Members {
			clone.Members[i].Annotations = nil
		}
		out.Enums = append(out.Enums, &clone)
	}
	for _, fn := range file.Functions {
		if fn.Macro || !sourceMatchesDocument(uri, fn.SourcePath) {
			continue
		}
		clone := *fn
		clone.Annotations = nil
		out.Functions = append(out.Functions, &clone)
	}
	for _, test := range file.Tests {
		if sourceMatchesDocument(uri, test.SourcePath) {
			out.Tests = append(out.Tests, test)
		}
	}
	return out
}

func cloneExpandedFunctions(functions []*ast.Function) []*ast.Function {
	out := make([]*ast.Function, 0, len(functions))
	for _, fn := range functions {
		if fn.Macro {
			continue
		}
		clone := *fn
		clone.Annotations = nil
		out = append(out, &clone)
	}
	return out
}
