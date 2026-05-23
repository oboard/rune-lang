package lsp

import (
	"encoding/json"
	"io"
)

func (s *server) handle(req request) error {
	switch req.Method {
	case "initialize":
		return s.respond(req.ID, map[string]any{
			"capabilities": map[string]any{
				"textDocumentSync":           1,
				"hoverProvider":              true,
				"completionProvider":         map[string]any{"triggerCharacters": []string{"@", "."}},
				"definitionProvider":         true,
				"codeLensProvider":           map[string]any{"resolveProvider": false},
				"documentSymbolProvider":     true,
				"documentFormattingProvider": true,
				"inlayHintProvider":          true,
				"semanticTokensProvider": map[string]any{
					"legend": map[string]any{
						"tokenTypes":     []string{"variable"},
						"tokenModifiers": []string{"modification"},
					},
					"full": true,
				},
				"renameProvider": true,
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
	case "textDocument/codeLens":
		var params codeLensParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return err
		}
		return s.respond(req.ID, s.codeLenses(params.TextDocument.URI))
	case "textDocument/documentSymbol":
		var params documentSymbolParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return err
		}
		return s.respond(req.ID, s.documentSymbols(params.TextDocument.URI))
	case "textDocument/formatting":
		var params formattingParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return err
		}
		return s.respond(req.ID, s.formatting(params.TextDocument.URI))
	case "textDocument/inlayHint":
		var params inlayHintParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return err
		}
		return s.respond(req.ID, s.inlayHints(params.TextDocument.URI))
	case "textDocument/semanticTokens/full":
		var params semanticTokensParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return err
		}
		return s.respond(req.ID, s.semanticTokens(params.TextDocument.URI))
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
