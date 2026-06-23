package lsp

import (
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/compiler"
	"github.com/oboard/rune-lang/internal/lexer"
)

func (s *server) completion(uri string, pos position) any {
	prog, _ := s.analyze(uri)
	if prog == nil {
		return []map[string]any{}
	}
	if moduleName, ok := annotationCompletionModule(s.docs[uri], pos); ok {
		return stdlibMacroCompletion(prog.Info, moduleName)
	}
	if looksLikeAnnotationCompletion(s.docs[uri], pos) {
		return macroCompletion(prog)
	}
	if items, ok := s.memberCompletion(uri, prog, pos); ok {
		return items
	}
	return globalCompletion(prog)
}

func looksLikeAnnotationCompletion(text string, pos position) bool {
	offset, ok := offsetFromPosition(text, pos)
	if !ok {
		return false
	}
	lineStart := strings.LastIndexByte(text[:offset], '\n') + 1
	prefix := strings.TrimSpace(text[lineStart:offset])
	if !strings.HasPrefix(prefix, "#") || strings.Contains(prefix, ".") {
		return false
	}
	for i := 1; i < len(prefix); i++ {
		if !isIdentByte(prefix[i]) {
			return false
		}
	}
	return true
}

func annotationCompletionModule(text string, pos position) (string, bool) {
	offset, ok := offsetFromPosition(text, pos)
	if !ok {
		return "", false
	}
	lineStart := strings.LastIndexByte(text[:offset], '\n') + 1
	prefix := strings.TrimSpace(text[lineStart:offset])
	if !strings.HasPrefix(prefix, "#") {
		return "", false
	}
	dot := strings.IndexByte(prefix, '.')
	if dot <= 1 {
		return "", false
	}
	moduleName := prefix[1:dot]
	for i := range moduleName {
		if !isIdentByte(moduleName[i]) {
			return "", false
		}
	}
	return moduleName, true
}

func globalCompletion(prog *compiler.Program) []map[string]any {
	var items []map[string]any
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
	for _, trait := range prog.File.Traits {
		items = append(items, map[string]any{
			"label":  trait.Name,
			"kind":   8,
			"detail": traitTypeSignature(prog.Info, trait),
		})
	}
	for _, fn := range prog.File.Functions {
		if fn.Macro {
			continue
		}
		items = append(items, map[string]any{
			"label":  fn.Name,
			"kind":   3,
			"detail": functionSignature(prog.Info, fn),
		})
	}
	for _, fn := range prog.Info.ExternalFunctions {
		items = append(items, map[string]any{
			"label":  fn.Name,
			"kind":   3,
			"detail": functionValueSignature(prog.Info, fn),
		})
	}
	for _, value := range prog.Info.ExternalValues {
		items = append(items, map[string]any{
			"label":  value.Name,
			"kind":   6,
			"detail": fmt.Sprintf("%s: %s", value.Name, displayCheckerType(prog.Info, value.Type)),
		})
	}
	return items
}

func (s *server) memberCompletion(uri string, prog *compiler.Program, pos position) ([]map[string]any, bool) {
	if !looksLikeMemberCompletion(s.docs[uri], pos) {
		return nil, false
	}
	receiver, ok := memberCompletionReceiverType(s.docs[uri], prog, pos)
	if !ok || receiver == "" || receiver == checker.Unknown {
		return nil, false
	}
	if moduleName, ok := checker.ModuleNamespaceName(receiver); ok {
		return stdlibModuleCompletion(prog.Info, moduleName), true
	}
	if importPath, ok := checker.ImportNamespacePath(receiver); ok {
		return importNamespaceCompletion(prog.Info, importPath), true
	}
	if moduleName, receiverName, ok := stdlibReceiverModule(receiver); ok {
		return stdlibMemberCompletion(prog.Info, receiver, moduleName, receiverName), true
	}
	if traitInfo := traitInfoForType(prog.Info, receiver); traitInfo != nil {
		return traitMemberCompletion(prog.Info, traitInfo), true
	}
	if structInfo := prog.Info.Types[baseType(receiver)]; structInfo != nil {
		return structMemberCompletion(prog.Info, structInfo), true
	}
	return nil, false
}

func looksLikeMemberCompletion(text string, pos position) bool {
	offset, ok := offsetFromPosition(text, pos)
	if !ok {
		return false
	}
	lineStart := strings.LastIndexByte(text[:offset], '\n') + 1
	prefixStart := offset
	for prefixStart > lineStart && isIdentByte(text[prefixStart-1]) {
		prefixStart--
	}
	return prefixStart > lineStart && text[prefixStart-1] == '.'
}

func memberCompletionReceiverType(text string, prog *compiler.Program, pos position) (checker.Type, bool) {
	if sel := selectorAtCompletion(prog.File, pos); sel != nil {
		if typ := prog.Info.ExprTypes[sel.Receiver]; typ != "" {
			return typ, true
		}
	}
	name := memberCompletionReceiverName(text, pos)
	if name == "" {
		return "", false
	}
	return localValueTypeBeforePosition(prog, name, pos)
}

func selectorAtCompletion(file *ast.File, pos position) *ast.SelectorExpr {
	var found *ast.SelectorExpr
	walkFileSelectors(file, func(sel *ast.SelectorExpr) {
		if found != nil {
			return
		}
		line := sel.NamePos.Line - 1
		char := sel.NamePos.Column - 1
		if pos.Line == line && pos.Character >= char && pos.Character <= char+max(len(sel.Name), 0) {
			found = sel
		}
	})
	return found
}

func memberCompletionReceiverName(text string, pos position) string {
	offset, ok := offsetFromPosition(text, pos)
	if !ok {
		return ""
	}
	lineStart := strings.LastIndexByte(text[:offset], '\n') + 1
	prefixStart := offset
	for prefixStart > lineStart && isIdentByte(text[prefixStart-1]) {
		prefixStart--
	}
	if prefixStart <= lineStart || text[prefixStart-1] != '.' {
		return ""
	}
	end := prefixStart - 1
	start := end
	for start > lineStart && isIdentByte(text[start-1]) {
		start--
	}
	if start == end {
		return ""
	}
	return text[start:end]
}

func localValueTypeBeforePosition(prog *compiler.Program, name string, pos position) (checker.Type, bool) {
	var best checker.Type
	bestOffset := -1
	consider := func(tok lexer.Position, typ checker.Type) {
		if typ == "" || typ == checker.Unknown || !tokenBeforePosition(tok, pos) {
			return
		}
		offset := tok.Offset
		if offset == 0 {
			offset = (tok.Line * 1_000_000) + tok.Column
		}
		if offset > bestOffset {
			bestOffset = offset
			best = typ
		}
	}
	for _, typ := range prog.File.Types {
		structInfo := prog.Info.Types[typ.Name]
		for _, method := range typ.Methods {
			methodInfo := (*checker.FuncInfo)(nil)
			if structInfo != nil {
				methodInfo = structInfo.Methods[method.Name]
			}
			for i, param := range method.Params {
				if param.Name != name {
					continue
				}
				typ := checker.Type(param.Type.Canonical())
				if methodInfo != nil && i < len(methodInfo.Params) {
					typ = methodInfo.Params[i].Type
				}
				consider(param.Pos, typ)
			}
		}
	}
	for _, fn := range prog.File.Functions {
		fnInfo := prog.Info.FunctionDecls[fn]
		for i, param := range fn.Params {
			if param.Name != name {
				continue
			}
			typ := checker.Type(param.Type.Canonical())
			if fnInfo != nil && i < len(fnInfo.Params) {
				typ = fnInfo.Params[i].Type
			}
			consider(param.Pos, typ)
		}
	}
	walkFileStatements(prog.File, func(stmt ast.Stmt) {
		switch stmt := stmt.(type) {
		case *ast.LetStmt:
			if stmt.Name == name {
				consider(stmt.Pos, prog.Info.ExprTypes[stmt.Value])
			}
		case *ast.ObjectDestructureStmt:
			valueType := prog.Info.ExprTypes[stmt.Value]
			for _, field := range stmt.Fields {
				if field.Name != name {
					continue
				}
				if typ, ok := checker.FieldType(prog.Info, valueType, field.Field); ok {
					consider(field.NamePos, typ)
				}
			}
		}
	})
	return best, best != ""
}

func tokenBeforePosition(tok lexer.Position, pos position) bool {
	line := tok.Line - 1
	char := tok.Column - 1
	return line < pos.Line || (line == pos.Line && char <= pos.Character)
}

func stdlibModuleCompletion(info *checker.Info, moduleName string) []map[string]any {
	if info == nil || info.Stdlib == nil {
		return nil
	}
	module := info.Stdlib.Modules[moduleName]
	if module == nil {
		return nil
	}
	items := make([]map[string]any, 0)
	for i := range module.Functions {
		fn := &module.Functions[i]
		if fn.Macro || fn.Receiver != "" || fn.TopLevelOnly {
			continue
		}
		items = append(items, map[string]any{
			"label":  fn.Name,
			"kind":   3,
			"detail": stdlibSignature(moduleName, fn),
		})
	}
	return items
}

func stdlibMacroCompletion(info *checker.Info, moduleName string) []map[string]any {
	if info == nil || info.Stdlib == nil {
		return nil
	}
	module := info.Stdlib.Modules[moduleName]
	if module == nil {
		return nil
	}
	items := make([]map[string]any, 0)
	for i := range module.Functions {
		fn := &module.Functions[i]
		if !fn.Macro || fn.Receiver != "" {
			continue
		}
		items = append(items, map[string]any{
			"label":  fn.Name,
			"kind":   3,
			"detail": stdlibSignature(moduleName, fn),
		})
	}
	return items
}

func macroCompletion(prog *compiler.Program) []map[string]any {
	items := make([]map[string]any, 0)
	for _, fn := range prog.File.Functions {
		if !fn.Macro {
			continue
		}
		items = append(items, map[string]any{
			"label":  fn.Name,
			"kind":   3,
			"detail": functionSignature(prog.Info, fn),
		})
	}
	if prog.Info.Stdlib == nil {
		return items
	}
	for _, moduleName := range prog.Info.Stdlib.ModuleNames() {
		module := prog.Info.Stdlib.Modules[moduleName]
		for i := range module.Functions {
			fn := &module.Functions[i]
			if !fn.Macro || fn.Receiver != "" {
				continue
			}
			items = append(items, map[string]any{
				"label":  moduleName + "." + fn.Name,
				"kind":   3,
				"detail": stdlibSignature(moduleName, fn),
			})
		}
	}
	return items
}

func importNamespaceCompletion(info *checker.Info, sourcePath string) []map[string]any {
	if info == nil {
		return nil
	}
	items := make([]map[string]any, 0)
	for _, fn := range info.Functions {
		if fn.Private || !sameLSPSourcePath(fn.SourcePath, sourcePath) {
			continue
		}
		items = append(items, map[string]any{
			"label":  fn.Name,
			"kind":   3,
			"detail": functionValueSignature(info, fn),
		})
	}
	for _, value := range info.ExternalValues {
		if !sameLSPSourcePath(value.SourcePath, sourcePath) {
			continue
		}
		items = append(items, map[string]any{
			"label":  value.Name,
			"kind":   6,
			"detail": fmt.Sprintf("%s: %s", value.Name, displayCheckerType(info, value.Type)),
		})
	}
	return items
}

func sameLSPSourcePath(a string, b string) bool {
	return cleanLSPSourcePath(a) == cleanLSPSourcePath(b)
}

func cleanLSPSourcePath(path string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "file://") {
		if uri, err := url.Parse(path); err == nil {
			path = uri.Path
		}
	}
	clean := filepath.Clean(path)
	if abs, err := filepath.Abs(clean); err == nil {
		return abs
	}
	return clean
}

func stdlibMemberCompletion(info *checker.Info, receiver checker.Type, moduleName string, receiverName string) []map[string]any {
	if info == nil || info.Stdlib == nil {
		return nil
	}
	module := info.Stdlib.Modules[moduleName]
	if module == nil {
		return nil
	}
	items := make([]map[string]any, 0)
	for _, fn := range module.Functions {
		if fn.Receiver != receiverName {
			continue
		}
		items = append(items, map[string]any{
			"label":  fn.Name,
			"kind":   2,
			"detail": stdlibMemberSignature(info, receiver, fn),
		})
	}
	return items
}

func structMemberCompletion(info *checker.Info, structInfo *checker.StructInfo) []map[string]any {
	items := make([]map[string]any, 0, len(structInfo.Fields)+len(structInfo.Methods))
	for _, field := range structInfo.Fields {
		items = append(items, map[string]any{
			"label":  field.Name,
			"kind":   5,
			"detail": fmt.Sprintf("%s: %s", field.Name, displayCheckerType(info, field.Type)),
		})
	}
	methodNames := make([]string, 0, len(structInfo.Methods))
	for name := range structInfo.Methods {
		methodNames = append(methodNames, name)
	}
	sort.Strings(methodNames)
	for _, name := range methodNames {
		method := structInfo.Methods[name]
		items = append(items, map[string]any{
			"label":  name,
			"kind":   2,
			"detail": classMethodSignature(structInfo.Name, method),
		})
	}
	return items
}

func traitMemberCompletion(info *checker.Info, traitInfo *checker.TraitInfo) []map[string]any {
	items := make([]map[string]any, 0, len(traitInfo.Fields)+len(traitInfo.Methods))
	for _, field := range traitInfo.Fields {
		items = append(items, map[string]any{
			"label":  field.Name,
			"kind":   5,
			"detail": fmt.Sprintf("%s: %s", field.Name, displayCheckerType(info, field.Type)),
		})
	}
	methodNames := make([]string, 0, len(traitInfo.Methods))
	for name := range traitInfo.Methods {
		methodNames = append(methodNames, name)
	}
	sort.Strings(methodNames)
	for _, name := range methodNames {
		method := traitInfo.Methods[name]
		items = append(items, map[string]any{
			"label":  name,
			"kind":   2,
			"detail": traitMemberSignature(info, traitInfo.Name, method),
		})
	}
	return items
}
