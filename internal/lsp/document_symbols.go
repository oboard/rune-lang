package lsp

func (s *server) documentSymbols(uri string) any {
	prog, _ := s.analyze(uri)
	if prog == nil {
		return []any{}
	}
	items := make([]map[string]any, 0, len(prog.File.Functions))
	for _, fn := range prog.File.Functions {
		if !sourceMatchesDocument(uri, fn.SourcePath) {
			continue
		}
		rng := functionRange(fn)
		items = append(items, map[string]any{
			"name":           fn.Name,
			"kind":           12,
			"range":          rng,
			"selectionRange": symbolRange(fn.NamePos, len(fn.Name)),
			"detail":         functionSignature(prog.Info, fn),
		})
	}
	return items
}
