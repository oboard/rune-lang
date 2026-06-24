package lsp

func (s *server) publishDiagnostics(uri string) error {
	_, diags := s.analyze(uri)
	items := make([]map[string]any, 0, len(diags))
	for _, diag := range diags {
		severity := 1 // Error
		if diag.Severity == "warning" {
			severity = 2 // Warning
		}
		items = append(items, map[string]any{
			"range":    lspRange(diag.Pos),
			"severity": severity,
			"source":   "rune",
			"message":  diag.Message,
		})
	}
	return s.notify("textDocument/publishDiagnostics", map[string]any{
		"uri":         uri,
		"diagnostics": items,
	})
}

func (s *server) clearDiagnostics(uri string) error {
	return s.notify("textDocument/publishDiagnostics", map[string]any{
		"uri":         uri,
		"diagnostics": []map[string]any{},
	})
}
