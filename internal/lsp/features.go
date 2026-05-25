package lsp

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/compiler"
	"github.com/oboard/rune-lang/internal/lexer"
	"github.com/oboard/rune-lang/internal/stdlib"
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

func (s *server) clearDiagnostics(uri string) error {
	return s.notify("textDocument/publishDiagnostics", map[string]any{
		"uri":         uri,
		"diagnostics": []map[string]any{},
	})
}

func (s *server) hover(uri string, pos position) any {
	prog, _ := compiler.AnalyzeSource(uri, s.docs[uri])
	word := wordAt(s.docs[uri], pos)
	if word == "" || prog == nil {
		return nil
	}
	if h := s.methodHover(uri, prog, pos); h != nil {
		return h
	}
	if h := s.exprHover(prog, pos); h != nil {
		return h
	}
	if h := typeHover(prog, pos); h != nil {
		return h
	}
	for _, fn := range prog.File.Functions {
		if fn.Name == word {
			return map[string]any{
				"contents": map[string]any{
					"kind":  "markdown",
					"value": fmt.Sprintf("```rune\n%s\n```", functionSignature(prog.Info, fn)),
				},
				"range": symbolRange(fn.NamePos, len(fn.Name)),
			}
		}
		fnInfo := prog.Info.Functions[fn.Name]
		for i, param := range fn.Params {
			if param.Name == word {
				typ := param.Type
				if fnInfo != nil && i < len(fnInfo.Params) && fnInfo.Params[i].Type != "" && fnInfo.Params[i].Type != checker.Unknown {
					typ = displayCheckerType(prog.Info, fnInfo.Params[i].Type)
				}
				if typ == "" {
					typ = string(checker.Unknown)
				}
				return map[string]any{
					"contents": map[string]any{
						"kind":  "markdown",
						"value": fmt.Sprintf("```rune\n%s: %s\n```", param.Name, typ),
					},
					"range": symbolRange(param.Pos, len(param.Name)),
				}
			}
		}
	}
	return nil
}

func (s *server) methodHover(uri string, prog *compiler.Program, pos position) any {
	if hover := methodDeclHover(uri, prog, pos); hover != nil {
		return hover
	}
	sel := selectorAt(prog.File, pos)
	if sel == nil {
		return nil
	}
	if at, ok := sel.Receiver.(*ast.AtExpr); ok {
		return stdlibHover(prog.Info.Stdlib, at.Name, sel.Name, sel.NamePos)
	}
	receiver := prog.Info.ExprTypes[sel.Receiver]
	if moduleName, receiverName, ok := stdlibReceiverModule(receiver); ok {
		return stdlibReceiverHover(prog.Info.Stdlib, moduleName, receiverName, sel.Name, sel.NamePos)
	}
	structInfo := prog.Info.Types[baseType(receiver)]
	if structInfo == nil {
		return nil
	}
	if field, ok := structInfo.ByName[sel.Name]; ok {
		return hoverResult(fmt.Sprintf("%s: %s", sel.Name, displayCheckerType(prog.Info, field.Type)), sel.NamePos, sel.Name)
	}
	method := structInfo.Methods[sel.Name]
	if method == nil || method.Node == nil {
		return nil
	}
	return hoverResult(classMethodSignature(structInfo.Name, method), sel.NamePos, sel.Name)
}

func (s *server) exprHover(prog *compiler.Program, pos position) any {
	var found any
	signals := signalGraph(prog.File)
	walkFileStatements(prog.File, func(stmt ast.Stmt) {
		if found != nil {
			return
		}
		if let, ok := stmt.(*ast.LetStmt); ok && containsSymbol(pos, let.Pos, let.Name) {
			if typ := prog.Info.ExprTypes[let.Value]; typ != "" && typ != checker.Unknown {
				found = hoverResult(localHoverText(prog.Info, let.Name, typ, signals), let.Pos, let.Name)
			}
		}
	})
	if found != nil {
		return found
	}
	walkFileExprs(prog.File, func(expr ast.Expr) {
		if found != nil {
			return
		}
		switch e := expr.(type) {
		case *ast.Identifier:
			if !containsSymbol(pos, e.Pos, e.Name) {
				return
			}
			if typ := prog.Info.ExprTypes[e]; typ != "" && typ != checker.Unknown {
				found = hoverResult(localHoverText(prog.Info, e.Name, typ, signals), e.Pos, e.Name)
			}
		case *ast.ThisExpr:
			if !containsSymbol(pos, e.Pos, "this") {
				return
			}
			if typ := prog.Info.ExprTypes[e]; typ != "" && typ != checker.Unknown {
				found = hoverResult(fmt.Sprintf("this: %s", displayCheckerType(prog.Info, typ)), e.Pos, "this")
			}
		}
	})
	return found
}

func localHoverText(info *checker.Info, name string, typ checker.Type, signals map[string][]string) string {
	text := fmt.Sprintf("%s: %s", name, displayCheckerType(info, typ))
	deps, ok := signals[name]
	if !ok {
		return text
	}
	kind := "signal"
	if len(deps) > 0 {
		kind = "computed"
	}
	text += "\n" + kind
	if chain := dependencyChain(name, signals); chain != "" {
		text += "\ndeps: " + chain
	}
	return text
}

func typeHover(prog *compiler.Program, pos position) any {
	if name, tokPos, ok := structLiteralTypeAt(prog.File, pos); ok {
		if typ := structTypeByName(prog.File, name); typ != nil {
			return hoverResult(structTypeSignature(prog.Info, typ), tokPos, name)
		}
	}
	for _, typ := range prog.File.Types {
		if containsSymbol(pos, typ.NamePos, typ.Name) {
			return hoverResult(structTypeSignature(prog.Info, typ), typ.NamePos, typ.Name)
		}
	}
	for _, enum := range prog.File.Enums {
		if containsSymbol(pos, enum.NamePos, enum.Name) {
			return hoverResult(enumTypeSignature(enum), enum.NamePos, enum.Name)
		}
	}
	return nil
}

func structLiteralTypeAt(file *ast.File, pos position) (string, lexer.Position, bool) {
	var name string
	var tokPos lexer.Position
	walkFileExprs(file, func(expr ast.Expr) {
		if name != "" {
			return
		}
		lit, ok := expr.(*ast.StructLiteral)
		if !ok || !containsSymbol(pos, lit.Pos, lit.TypeName) {
			return
		}
		name = lit.TypeName
		tokPos = lit.Pos
	})
	return name, tokPos, name != ""
}

func structTypeByName(file *ast.File, name string) *ast.StructType {
	for _, typ := range file.Types {
		if typ.Name == name {
			return typ
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
			"detail": functionSignature(prog.Info, fn),
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
	if target := s.methodTarget(uri, prog, pos); target != nil {
		return target.location()
	}
	if target := fieldTarget(uri, prog, pos); target != nil {
		return target.location()
	}
	if target := typeTarget(uri, prog, pos); target != nil {
		return target.location()
	}
	if target := localTarget(uri, prog, pos); target != nil {
		return target.location()
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

func (s *server) references(uri string, pos position, includeDeclaration bool) any {
	prog, _ := compiler.AnalyzeSource(uri, s.docs[uri])
	if prog == nil {
		return []map[string]any{}
	}
	if target := functionTarget(uri, prog, pos); target != nil {
		return functionReferences(prog, target, includeDeclaration)
	}
	if target := s.methodTarget(uri, prog, pos); target != nil && target.structName != "" {
		return methodReferences(prog, target, includeDeclaration)
	}
	return []map[string]any{}
}

func functionTarget(uri string, prog *compiler.Program, pos position) *methodTarget {
	for _, fn := range prog.File.Functions {
		if containsSymbol(pos, fn.NamePos, fn.Name) {
			return &methodTarget{uri: uri, name: fn.Name, pos: fn.NamePos}
		}
	}
	call := functionCallAt(prog.File, pos)
	if call == nil {
		return nil
	}
	for _, fn := range prog.File.Functions {
		if fn.Name == call.Name {
			return &methodTarget{uri: uri, name: fn.Name, pos: fn.NamePos}
		}
	}
	return nil
}

func functionCallAt(file *ast.File, pos position) *ast.Identifier {
	var found *ast.Identifier
	walkFileCalls(file, func(call *ast.CallExpr) {
		if found != nil {
			return
		}
		ident, ok := call.Callee.(*ast.Identifier)
		if ok && containsSymbol(pos, ident.Pos, ident.Name) {
			found = ident
		}
	})
	return found
}

func functionReferences(prog *compiler.Program, target *methodTarget, includeDeclaration bool) []map[string]any {
	refs := []map[string]any{}
	if includeDeclaration {
		refs = append(refs, target.location())
	}
	walkFileCalls(prog.File, func(call *ast.CallExpr) {
		ident, ok := call.Callee.(*ast.Identifier)
		if !ok || ident.Name != target.name {
			return
		}
		refs = append(refs, referenceLocation(target.uri, ident.Pos, ident.Name))
	})
	return refs
}

func methodReferences(prog *compiler.Program, target *methodTarget, includeDeclaration bool) []map[string]any {
	refs := []map[string]any{}
	if includeDeclaration {
		refs = append(refs, target.location())
	}
	walkFileCalls(prog.File, func(call *ast.CallExpr) {
		sel, ok := call.Callee.(*ast.SelectorExpr)
		if !ok || sel.Name != target.name {
			return
		}
		if baseType(prog.Info.ExprTypes[sel.Receiver]) != target.structName {
			return
		}
		refs = append(refs, referenceLocation(target.uri, sel.NamePos, sel.Name))
	})
	return refs
}

func referenceLocation(uri string, pos lexer.Position, name string) map[string]any {
	return map[string]any{
		"uri":   uri,
		"range": symbolRange(pos, len(name)),
	}
}

func localTarget(uri string, prog *compiler.Program, pos position) *methodTarget {
	name := identifierNameAt(prog.File, pos)
	if name == "" {
		name = letNameAt(prog.File, pos)
	}
	if name == "" {
		return nil
	}
	if target := letTarget(uri, prog.File, name); target != nil {
		return target
	}
	if target := paramTarget(uri, prog.File, name); target != nil {
		return target
	}
	return nil
}

func typeTarget(uri string, prog *compiler.Program, pos position) *methodTarget {
	name := wordAt(prog.Source, pos)
	if name == "" {
		return nil
	}
	for _, typ := range prog.File.Types {
		if typ.Name == name {
			return &methodTarget{uri: uri, name: typ.Name, pos: typ.NamePos}
		}
	}
	return nil
}

func identifierNameAt(file *ast.File, pos position) string {
	var found string
	walkFileExprs(file, func(expr ast.Expr) {
		if found != "" {
			return
		}
		if ident, ok := expr.(*ast.Identifier); ok && containsSymbol(pos, ident.Pos, ident.Name) {
			found = ident.Name
		}
	})
	return found
}

func letNameAt(file *ast.File, pos position) string {
	var found string
	walkFileStatements(file, func(stmt ast.Stmt) {
		if found != "" {
			return
		}
		if let, ok := stmt.(*ast.LetStmt); ok && containsSymbol(pos, let.Pos, let.Name) {
			found = let.Name
		}
	})
	return found
}

func letTarget(uri string, file *ast.File, name string) *methodTarget {
	var found *methodTarget
	walkFileStatements(file, func(stmt ast.Stmt) {
		if found != nil {
			return
		}
		if let, ok := stmt.(*ast.LetStmt); ok && let.Name == name {
			found = &methodTarget{uri: uri, name: let.Name, pos: let.Pos}
		}
	})
	return found
}

func paramTarget(uri string, file *ast.File, name string) *methodTarget {
	for _, typ := range file.Types {
		for _, method := range typ.Methods {
			for _, param := range method.Params {
				if param.Name == name {
					return &methodTarget{uri: uri, name: param.Name, pos: param.Pos}
				}
			}
		}
	}
	for _, fn := range file.Functions {
		for _, param := range fn.Params {
			if param.Name == name {
				return &methodTarget{uri: uri, name: param.Name, pos: param.Pos}
			}
		}
	}
	return nil
}

func fieldTarget(uri string, prog *compiler.Program, pos position) *methodTarget {
	sel := selectorAt(prog.File, pos)
	if sel == nil {
		return nil
	}
	receiver := prog.Info.ExprTypes[sel.Receiver]
	structInfo := prog.Info.Types[baseType(receiver)]
	if structInfo == nil {
		return nil
	}
	if structInfo.Node != nil {
		for _, field := range structInfo.Node.Fields {
			if field.Name == sel.Name {
				return &methodTarget{uri: uri, name: field.Name, pos: field.Pos, structName: structInfo.Name}
			}
		}
	}
	return anonymousFieldTarget(uri, prog, structInfo.Name, sel.Name)
}

func anonymousFieldTarget(uri string, prog *compiler.Program, typeName string, fieldName string) *methodTarget {
	var found *methodTarget
	walkFileExprs(prog.File, func(expr ast.Expr) {
		if found != nil {
			return
		}
		obj, ok := expr.(*ast.AnonymousObjectLiteral)
		if !ok {
			return
		}
		if baseType(prog.Info.ExprTypes[obj]) != typeName {
			objTypeName := baseType(prog.Info.ExprTypes[obj])
			if !anonymousObjectTypeCanSatisfyField(prog.Info, objTypeName, typeName, fieldName) {
				return
			}
		}
		for _, field := range obj.Fields {
			if field.Name == fieldName {
				found = &methodTarget{uri: uri, name: field.Name, pos: field.Pos, structName: typeName}
				return
			}
		}
	})
	return found
}

func anonymousObjectTypeCanSatisfyField(info *checker.Info, objectType string, targetType string, fieldName string) bool {
	if info == nil {
		return false
	}
	objectInfo := info.Types[objectType]
	targetInfo := info.Types[targetType]
	if objectInfo == nil || targetInfo == nil {
		return false
	}
	if _, ok := targetInfo.ByName[fieldName]; !ok {
		return false
	}
	for _, targetField := range targetInfo.Fields {
		if _, ok := objectInfo.ByName[targetField.Name]; !ok {
			return false
		}
	}
	return true
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
			"detail":         functionSignature(prog.Info, fn),
		})
	}
	return items
}

func (s *server) inlayHints(uri string) any {
	text := s.docs[uri]
	prog, _ := compiler.AnalyzeSource(uri, text)
	if prog == nil {
		return []any{}
	}
	var hints []map[string]any
	for _, fn := range prog.File.Functions {
		fnInfo := prog.Info.Functions[fn.Name]
		if fnInfo == nil {
			continue
		}
		for i, param := range fn.Params {
			if param.Type != "" || i >= len(fnInfo.Params) {
				continue
			}
			typ := fnInfo.Params[i].Type
			if typ == "" || typ == checker.Unknown {
				continue
			}
			pos := position{
				Line:      max(param.Pos.Line-1, 0),
				Character: max(param.Pos.Column-1, 0) + len(param.Name),
			}
			hints = append(hints, map[string]any{
				"position": pos,
				"label":    ": " + displayCheckerTypeOneLine(prog.Info, typ),
				"kind":     1,
				"tooltip":  functionSignature(prog.Info, fn),
			})
		}
		if fn.ReturnType != "" || fnInfo.Return == "" || fnInfo.Return == checker.Unknown {
			continue
		}
		if pos, ok := fatArrowPosition(text, fn); ok {
			hints = append(hints, map[string]any{
				"position": pos,
				"label":    "-> " + displayCheckerTypeOneLine(prog.Info, fnInfo.Return) + " ",
				"kind":     1,
				"tooltip":  functionSignature(prog.Info, fn),
			})
		}
	}
	walkFileExprs(prog.File, func(expr ast.Expr) {
		lambda, ok := expr.(*ast.LambdaExpr)
		if !ok {
			return
		}
		params, ret, ok := parseRawDisplayFuncType(string(prog.Info.ExprTypes[lambda]))
		if !ok {
			return
		}
		for i, name := range lambda.Params {
			if i >= len(lambda.ParamPos) || i >= len(params) {
				continue
			}
			if i < len(lambda.ParamTypes) && lambda.ParamTypes[i] != "" {
				continue
			}
			typ := checker.Type(params[i])
			if typ == "" || typ == checker.Unknown {
				continue
			}
			pos := position{
				Line:      max(lambda.ParamPos[i].Line-1, 0),
				Character: max(lambda.ParamPos[i].Column-1, 0) + len(name),
			}
			hints = append(hints, map[string]any{
				"position": pos,
				"label":    ": " + displayCheckerTypeOneLine(prog.Info, typ),
				"kind":     1,
				"tooltip":  displayCheckerType(prog.Info, prog.Info.ExprTypes[lambda]),
			})
		}
		if lambda.Implicit || lambda.ReturnType != "" || ret == "" || ret == string(checker.Unknown) {
			return
		}
		if pos, ok := fatArrowPositionFromOffset(text, lambda.Pos.Offset); ok {
			hints = append(hints, map[string]any{
				"position": pos,
				"label":    "-> " + displayCheckerTypeOneLine(prog.Info, checker.Type(ret)) + " ",
				"kind":     1,
				"tooltip":  displayCheckerType(prog.Info, prog.Info.ExprTypes[lambda]),
			})
		}
	})
	return hints
}

func (s *server) semanticTokens(uri string) any {
	prog, _ := compiler.AnalyzeSource(uri, s.docs[uri])
	if prog == nil {
		return map[string]any{"data": []int{}}
	}
	signals := signalGraph(prog.File)
	var tokens []semanticToken
	for _, typ := range prog.File.Types {
		tokens = append(tokens, semanticToken{
			line:      max(typ.NamePos.Line-1, 0),
			character: max(typ.NamePos.Column-1, 0),
			length:    len(typ.Name),
			tokenType: semanticTokenTypeType,
		})
		for _, method := range typ.Methods {
			if method.Routine {
				tokens = append(tokens, semanticToken{
					line:      max(method.NamePos.Line-1, 0),
					character: max(method.NamePos.Column-1, 0),
					length:    len(method.Name),
					tokenType: semanticTokenTypeFunction,
					modifiers: semanticTokenModifierAsync,
				})
			}
		}
	}
	for _, fn := range prog.File.Functions {
		if !fn.Routine {
			continue
		}
		tokens = append(tokens, semanticToken{
			line:      max(fn.NamePos.Line-1, 0),
			character: max(fn.NamePos.Column-1, 0),
			length:    len(fn.Name),
			tokenType: semanticTokenTypeFunction,
			modifiers: semanticTokenModifierAsync,
		})
	}
	walkFileStatements(prog.File, func(stmt ast.Stmt) {
		switch stmt := stmt.(type) {
		case *ast.LetStmt:
			if _, ok := signals[stmt.Name]; !ok {
				return
			}
			tokens = append(tokens, semanticToken{
				line:      max(stmt.Pos.Line-1, 0),
				character: max(stmt.Pos.Column-1, 0),
				length:    len(stmt.Name),
				tokenType: semanticTokenTypeVariable,
				modifiers: semanticTokenModifierModification,
			})
		case *ast.AssignStmt:
			if _, ok := signals[stmt.Name]; !ok {
				return
			}
			tokens = append(tokens, semanticToken{
				line:      max(stmt.Pos.Line-1, 0),
				character: max(stmt.Pos.Column-1, 0),
				length:    len(stmt.Name),
				tokenType: semanticTokenTypeVariable,
				modifiers: semanticTokenModifierModification,
			})
		}
	})
	walkFileExprs(prog.File, func(expr ast.Expr) {
		switch expr := expr.(type) {
		case *ast.CallExpr:
			if token, ok := asyncCallSemanticToken(prog, expr); ok {
				tokens = append(tokens, token)
			}
		case *ast.Identifier:
			if _, signal := signals[expr.Name]; !signal {
				return
			}
			tokens = append(tokens, semanticToken{
				line:      max(expr.Pos.Line-1, 0),
				character: max(expr.Pos.Column-1, 0),
				length:    len(expr.Name),
				tokenType: semanticTokenTypeVariable,
				modifiers: semanticTokenModifierModification,
			})
		case *ast.StructLiteral:
			if structTypeByName(prog.File, expr.TypeName) == nil {
				return
			}
			tokens = append(tokens, semanticToken{
				line:      max(expr.Pos.Line-1, 0),
				character: max(expr.Pos.Column-1, 0),
				length:    len(expr.TypeName),
				tokenType: semanticTokenTypeType,
			})
		}
	})
	sort.Slice(tokens, func(i, j int) bool {
		if tokens[i].line != tokens[j].line {
			return tokens[i].line < tokens[j].line
		}
		return tokens[i].character < tokens[j].character
	})
	return map[string]any{"data": encodeSemanticTokens(tokens)}
}

func signalGraph(file *ast.File) map[string][]string {
	signals := map[string][]string{}
	walkFileStatements(file, func(stmt ast.Stmt) {
		let, ok := stmt.(*ast.LetStmt)
		if !ok {
			return
		}
		deps := exprSignalDeps(let.Value, signals)
		if let.Signal || len(deps) > 0 {
			signals[let.Name] = deps
		}
	})
	return signals
}

func asyncCallSemanticToken(prog *compiler.Program, call *ast.CallExpr) (semanticToken, bool) {
	switch callee := call.Callee.(type) {
	case *ast.Identifier:
		fn := prog.Info.Functions[callee.Name]
		if fn == nil || !fn.Routine {
			return semanticToken{}, false
		}
		return semanticToken{
			line:      max(callee.Pos.Line-1, 0),
			character: max(callee.Pos.Column-1, 0),
			length:    len(callee.Name),
			tokenType: semanticTokenTypeFunction,
			modifiers: semanticTokenModifierAsync,
		}, true
	case *ast.SelectorExpr:
		if asyncSelectorCall(prog, callee) {
			return semanticToken{
				line:      max(callee.NamePos.Line-1, 0),
				character: max(callee.NamePos.Column-1, 0),
				length:    len(callee.Name),
				tokenType: semanticTokenTypeFunction,
				modifiers: semanticTokenModifierAsync,
			}, true
		}
	}
	return semanticToken{}, false
}

func asyncSelectorCall(prog *compiler.Program, sel *ast.SelectorExpr) bool {
	if at, ok := sel.Receiver.(*ast.AtExpr); ok {
		fn, ok := prog.Info.Stdlib.Function(at.Name, sel.Name)
		return ok && fn.Routine
	}
	receiver := prog.Info.ExprTypes[sel.Receiver]
	if moduleName, receiverName, ok := stdlibReceiverModule(receiver); ok {
		fn, ok := prog.Info.Stdlib.ReceiverFunction(moduleName, receiverName, sel.Name)
		return ok && fn.Routine
	}
	structInfo := prog.Info.Types[baseType(receiver)]
	if structInfo == nil {
		return false
	}
	method := structInfo.Methods[sel.Name]
	return method != nil && method.Routine
}

func exprSignalDeps(expr ast.Expr, signals map[string][]string) []string {
	seen := map[string]bool{}
	var deps []string
	ast.WalkExpr(expr, func(expr ast.Expr) {
		if ident, ok := expr.(*ast.Identifier); ok {
			if _, signal := signals[ident.Name]; signal && !seen[ident.Name] {
				seen[ident.Name] = true
				deps = append(deps, ident.Name)
			}
		}
	})
	return deps
}

func dependencyChain(name string, signals map[string][]string) string {
	return dependencyChainPath(name, signals, map[string]bool{})
}

func dependencyChainPath(name string, signals map[string][]string, path map[string]bool) string {
	deps, ok := signals[name]
	if !ok || len(deps) == 0 {
		return ""
	}
	path[name] = true
	defer delete(path, name)
	chains := make([]string, 0, len(deps))
	for _, dep := range deps {
		if path[dep] {
			chains = append(chains, name+" -> "+dep+" (cycle)")
			continue
		}
		if chain := dependencyChainPath(dep, signals, path); chain != "" {
			chains = append(chains, name+" -> "+chain)
		} else {
			chains = append(chains, name+" -> "+dep)
		}
	}
	return strings.Join(chains, ", ")
}

type semanticToken struct {
	line      int
	character int
	length    int
	tokenType int
	modifiers int
}

const (
	semanticTokenTypeVariable = 0
	semanticTokenTypeType     = 1
	semanticTokenTypeFunction = 2

	semanticTokenModifierModification = 1 << 0
	semanticTokenModifierAsync        = 1 << 1
)

func encodeSemanticTokens(tokens []semanticToken) []int {
	data := make([]int, 0, len(tokens)*5)
	prevLine := 0
	prevChar := 0
	for i, token := range tokens {
		deltaLine := token.line - prevLine
		deltaChar := token.character
		if i > 0 && deltaLine == 0 {
			deltaChar = token.character - prevChar
		}
		data = append(data,
			deltaLine,
			deltaChar,
			token.length,
			token.tokenType,
			token.modifiers,
		)
		prevLine = token.line
		prevChar = token.character
	}
	return data
}

func (s *server) rename(uri string, pos position, newName string) any {
	text := s.docs[uri]
	word := wordAt(text, pos)
	if word == "" {
		return nil
	}
	prog, _ := compiler.AnalyzeSource(uri, text)
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

type methodTarget struct {
	uri        string
	name       string
	pos        lexer.Position
	structName string
}

func (t *methodTarget) location() map[string]any {
	return map[string]any{
		"uri":   t.uri,
		"range": symbolRange(t.pos, len(t.name)),
	}
}

func (s *server) methodTarget(uri string, prog *compiler.Program, pos position) *methodTarget {
	if target := methodDeclTarget(uri, prog.File, pos); target != nil {
		return target
	}
	sel := selectorAt(prog.File, pos)
	if sel == nil {
		return nil
	}
	if at, ok := sel.Receiver.(*ast.AtExpr); ok {
		return stdlibTarget(prog.Info.Stdlib, at.Name, sel.Name)
	}
	receiver := prog.Info.ExprTypes[sel.Receiver]
	if moduleName, receiverName, ok := stdlibReceiverModule(receiver); ok {
		return stdlibReceiverTarget(prog.Info.Stdlib, moduleName, receiverName, sel.Name)
	}
	structInfo := prog.Info.Types[baseType(receiver)]
	if structInfo == nil {
		return nil
	}
	method := structInfo.Methods[sel.Name]
	if method == nil || method.Node == nil {
		return nil
	}
	return &methodTarget{
		uri:        uri,
		name:       method.Node.Name,
		pos:        method.Node.NamePos,
		structName: structInfo.Name,
	}
}

func methodDeclTarget(uri string, file *ast.File, pos position) *methodTarget {
	for _, typ := range file.Types {
		for _, method := range typ.Methods {
			if containsSymbol(pos, method.NamePos, method.Name) {
				return &methodTarget{
					uri:        uri,
					name:       method.Name,
					pos:        method.NamePos,
					structName: typ.Name,
				}
			}
		}
	}
	return nil
}

func stdlibTarget(reg *stdlib.Registry, moduleName string, functionName string) *methodTarget {
	if reg == nil {
		return nil
	}
	fn, ok := reg.Function(moduleName, functionName)
	if !ok || fn.SourcePath == "" {
		return nil
	}
	return &methodTarget{
		uri:  fileURI(fn.SourcePath),
		name: fn.Name,
		pos:  fn.Pos,
	}
}

func stdlibReceiverTarget(reg *stdlib.Registry, moduleName string, receiverName string, functionName string) *methodTarget {
	if reg == nil {
		return nil
	}
	fn, ok := reg.ReceiverFunction(moduleName, receiverName, functionName)
	if !ok || fn.SourcePath == "" {
		return nil
	}
	return &methodTarget{
		uri:  fileURI(fn.SourcePath),
		name: fn.Name,
		pos:  fn.Pos,
	}
}

func methodDeclHover(uri string, prog *compiler.Program, pos position) any {
	for _, typ := range prog.File.Types {
		structInfo := prog.Info.Types[typ.Name]
		if structInfo == nil {
			continue
		}
		for _, method := range typ.Methods {
			if !containsSymbol(pos, method.NamePos, method.Name) {
				continue
			}
			info := structInfo.Methods[method.Name]
			if info == nil {
				continue
			}
			return hoverResult(classMethodSignature(typ.Name, info), method.NamePos, method.Name)
		}
	}
	return nil
}

func stdlibHover(reg *stdlib.Registry, moduleName string, functionName string, pos lexer.Position) any {
	if reg == nil {
		return nil
	}
	fn, ok := reg.Function(moduleName, functionName)
	if !ok {
		return nil
	}
	return stdlibFunctionHover(moduleName, fn, pos)
}

func stdlibReceiverHover(reg *stdlib.Registry, moduleName string, receiverName string, functionName string, pos lexer.Position) any {
	if reg == nil {
		return nil
	}
	fn, ok := reg.ReceiverFunction(moduleName, receiverName, functionName)
	if !ok {
		return nil
	}
	return stdlibFunctionHover(moduleName, fn, pos)
}

func stdlibFunctionHover(moduleName string, fn *stdlib.Function, pos lexer.Position) any {
	value := fmt.Sprintf("```rune\n%s\n```", stdlibSignature(moduleName, fn))
	if fn.SourcePath != "" {
		value += "\n" + fn.SourcePath
	}
	return map[string]any{
		"contents": map[string]any{
			"kind":  "markdown",
			"value": value,
		},
		"range": symbolRange(pos, len(fn.Name)),
	}
}

func hoverResult(signature string, pos lexer.Position, name string) any {
	return map[string]any{
		"contents": map[string]any{
			"kind":  "markdown",
			"value": fmt.Sprintf("```rune\n%s\n```", signature),
		},
		"range": symbolRange(pos, len(name)),
	}
}

func classMethodSignature(typeName string, fn *checker.FuncInfo) string {
	ret := fn.Return
	if ret == "" {
		ret = checker.Void
	}
	sig := fn.Node.Signature()
	return fmt.Sprintf("%s.%s -> %s", typeName, sig, ret)
}

func structTypeSignature(info *checker.Info, typ *ast.StructType) string {
	lines := []string{fmt.Sprintf("%s%s: {", typ.Name, formatSignatureGenerics(typ.Generics))}
	var structInfo *checker.StructInfo
	if info != nil {
		structInfo = info.Types[typ.Name]
	}
	for _, field := range typ.Fields {
		fieldType := field.Type
		if structInfo != nil {
			if inferred, ok := structInfo.ByName[field.Name]; ok && inferred.Type != "" && inferred.Type != checker.Unknown {
				fieldType = displayCheckerType(info, inferred.Type)
			}
		}
		if fieldType == "" {
			fieldType = string(checker.Unknown)
		}
		lines = append(lines, fmt.Sprintf("  %s: %s", field.Name, fieldType))
	}
	for _, method := range typ.Methods {
		lines = append(lines, "  "+methodSignature(info, typ.Name, method))
	}
	lines = append(lines, "}")
	return strings.Join(lines, "\n")
}

func methodSignature(info *checker.Info, typeName string, fn *ast.Function) string {
	params := make([]string, 0, len(fn.Params))
	var methodInfo *checker.FuncInfo
	if info != nil {
		if structInfo := info.Types[typeName]; structInfo != nil {
			methodInfo = structInfo.Methods[fn.Name]
		}
	}
	for i, param := range fn.Params {
		typ := param.Type
		if methodInfo != nil && i < len(methodInfo.Params) && methodInfo.Params[i].Type != "" && methodInfo.Params[i].Type != checker.Unknown {
			typ = displayCheckerType(info, methodInfo.Params[i].Type)
		}
		if typ == "" {
			typ = string(checker.Unknown)
		}
		params = append(params, fmt.Sprintf("%s: %s", param.Name, typ))
	}
	ret := fn.ReturnType
	if methodInfo != nil && methodInfo.Return != "" && methodInfo.Return != checker.Unknown {
		ret = displayCheckerType(info, methodInfo.Return)
	}
	if ret == "" {
		ret = string(checker.Void)
	}
	return fmt.Sprintf("%s%s(%s) -> %s", fn.Name, formatSignatureGenerics(fn.Generics), strings.Join(params, ", "), ret)
}

func enumTypeSignature(enum *ast.EnumType) string {
	if len(enum.Members) == 0 {
		return fmt.Sprintf("%s: {}", enum.Name)
	}
	lines := []string{fmt.Sprintf("%s: {", enum.Name)}
	for _, member := range enum.Members {
		lines = append(lines, fmt.Sprintf("  %s = %d", member.Name, member.Value))
	}
	lines = append(lines, "}")
	return strings.Join(lines, "\n")
}

func functionSignature(info *checker.Info, fn *ast.Function) string {
	fnInfo := info.Functions[fn.Name]
	if fnInfo == nil {
		return fn.Signature()
	}
	params := make([]string, 0, len(fn.Params))
	for i, param := range fn.Params {
		typ := param.Type
		if i < len(fnInfo.Params) && fnInfo.Params[i].Type != "" && fnInfo.Params[i].Type != checker.Unknown {
			typ = displayCheckerType(info, fnInfo.Params[i].Type)
		}
		if typ == "" {
			typ = string(checker.Unknown)
		}
		params = append(params, fmt.Sprintf("%s: %s", param.Name, typ))
	}
	ret := fn.ReturnType
	if fnInfo.Return != "" && fnInfo.Return != checker.Unknown {
		ret = displayCheckerType(info, fnInfo.Return)
	}
	if ret == "" {
		ret = string(checker.Void)
	}
	return fmt.Sprintf("%s%s(%s) -> %s", fn.Name, formatSignatureGenerics(fn.Generics), strings.Join(params, ", "), ret)
}

func formatSignatureGenerics(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return "[" + strings.Join(names, ", ") + "]"
}

func stdlibSignature(moduleName string, fn *stdlib.Function) string {
	owner := "@" + moduleName
	if fn.Receiver != "" {
		owner = fn.Receiver
	}
	generics := ""
	if len(fn.Generics) > 0 {
		generics = "[" + strings.Join(fn.Generics, ", ") + "]"
	}
	params := make([]string, 0, len(fn.Params))
	for i, typ := range fn.Params {
		name := fmt.Sprintf("arg%d", i+1)
		if i < len(fn.ParamNames) && fn.ParamNames[i] != "" {
			name = fn.ParamNames[i]
		}
		params = append(params, fmt.Sprintf("%s: %s", name, displayType(typ)))
	}
	ret := displayType(fn.Return)
	if ret == "" {
		ret = string(checker.Void)
	}
	return fmt.Sprintf("%s.%s%s(%s) -> %s", owner, fn.Name, generics, strings.Join(params, ", "), ret)
}

func displayType(name string) string {
	params, ret, ok := parseDisplayFuncType(name)
	if !ok {
		return name
	}
	return fmt.Sprintf("(%s) -> %s", strings.Join(params, ", "), ret)
}

func displayCheckerType(info *checker.Info, typ checker.Type) string {
	return displayCheckerTypeIndent(info, typ, 0)
}

func displayCheckerTypeOneLine(info *checker.Info, typ checker.Type) string {
	return strings.Join(strings.Fields(displayCheckerType(info, typ)), " ")
}

func displayCheckerTypeIndent(info *checker.Info, typ checker.Type, indent int) string {
	name := string(typ)
	if strings.HasPrefix(name, "{") && strings.HasSuffix(name, "}") {
		if info != nil {
			if structInfo := info.Types[name]; structInfo != nil {
				if len(structInfo.Fields) == 0 {
					return "{}"
				}
				currentIndent := strings.Repeat("  ", indent)
				fieldIndent := strings.Repeat("  ", indent+1)
				lines := []string{"{"}
				for _, field := range structInfo.Fields {
					lines = append(lines, fmt.Sprintf("%s%s: %s", fieldIndent, field.Name, displayCheckerTypeIndent(info, field.Type, indent+1)))
				}
				lines = append(lines, currentIndent+"}")
				return strings.Join(lines, "\n")
			}
		}
	}
	if params, ret, ok := parseRawDisplayFuncType(name); ok {
		for i, param := range params {
			params[i] = displayCheckerTypeIndent(info, checker.Type(param), indent)
		}
		return displayFuncType(params, displayCheckerTypeIndent(info, checker.Type(ret), indent))
	}
	return name
}

func displayFuncType(params []string, ret string) string {
	switch len(params) {
	case 0:
		return "() -> " + ret
	case 1:
		return params[0] + " -> " + ret
	default:
		return fmt.Sprintf("(%s) -> %s", strings.Join(params, ", "), ret)
	}
}

func parseDisplayFuncType(name string) ([]string, string, bool) {
	params, ret, ok := parseRawDisplayFuncType(name)
	if !ok {
		return nil, "", false
	}
	for i, param := range params {
		params[i] = displayType(param)
	}
	return params, displayType(ret), true
}

func parseRawDisplayFuncType(name string) ([]string, string, bool) {
	if !strings.HasPrefix(name, "Func[") || !strings.HasSuffix(name, "]") {
		return nil, "", false
	}
	parts := splitDisplayTypeList(strings.TrimSuffix(strings.TrimPrefix(name, "Func["), "]"))
	if len(parts) == 0 {
		return nil, "", false
	}
	return parts[:len(parts)-1], parts[len(parts)-1], true
}

func splitDisplayTypeList(src string) []string {
	var out []string
	depth := 0
	start := 0
	for i, ch := range src {
		switch ch {
		case '[', '{', '(':
			depth++
		case ']', '}', ')':
			depth--
		case ',':
			if depth == 0 {
				out = append(out, strings.TrimSpace(src[start:i]))
				start = i + 1
			}
		}
	}
	out = append(out, strings.TrimSpace(src[start:]))
	return out
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

func selectorAt(file *ast.File, pos position) *ast.SelectorExpr {
	var found *ast.SelectorExpr
	walkFileSelectors(file, func(sel *ast.SelectorExpr) {
		if found == nil && containsSymbol(pos, sel.NamePos, sel.Name) {
			found = sel
		}
	})
	return found
}

func walkFileSelectors(file *ast.File, visit func(*ast.SelectorExpr)) {
	walkFileExprs(file, func(expr ast.Expr) {
		if sel, ok := expr.(*ast.SelectorExpr); ok {
			visit(sel)
		}
	})
}

func walkFileCalls(file *ast.File, visit func(*ast.CallExpr)) {
	walkFileExprs(file, func(expr ast.Expr) {
		if call, ok := expr.(*ast.CallExpr); ok {
			visit(call)
		}
	})
}

func walkFileExprs(file *ast.File, visit func(ast.Expr)) {
	for _, typ := range file.Types {
		for _, method := range typ.Methods {
			ast.WalkExpr(method.Body, visit)
		}
	}
	for _, fn := range file.Functions {
		ast.WalkExpr(fn.Body, visit)
	}
	for _, test := range file.Tests {
		ast.WalkExpr(test.Body, visit)
	}
}

func walkFileStatements(file *ast.File, visit func(ast.Stmt)) {
	walkFileExprs(file, func(expr ast.Expr) {
		block, ok := expr.(*ast.BlockExpr)
		if !ok {
			return
		}
		for _, stmt := range block.Statements {
			visit(stmt)
		}
	})
}

func containsSymbol(pos position, start lexer.Position, name string) bool {
	line := start.Line - 1
	char := start.Column - 1
	return pos.Line == line && pos.Character >= char && pos.Character <= char+len(name)
}

func textEdit(pos lexer.Position, oldName string, newName string) map[string]any {
	return map[string]any{
		"range":   symbolRange(pos, len(oldName)),
		"newText": newName,
	}
}

func baseType(typ checker.Type) string {
	name := string(typ)
	if strings.HasPrefix(name, "{") && strings.HasSuffix(name, "}") {
		return name
	}
	if i := strings.IndexByte(name, '['); i >= 0 {
		return name[:i]
	}
	return name
}

func stdlibReceiverModule(receiver checker.Type) (string, string, bool) {
	if _, ok := checker.ArrayElement(receiver); ok {
		return "array", "Array", true
	}
	return checker.StdlibReceiverModule(receiver)
}

func fileURI(path string) string {
	return (&url.URL{Scheme: "file", Path: path}).String()
}
