package lsp

import (
	"fmt"
	"net/url"
	"regexp"
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
	if _, ok := checker.ArrayElement(receiver); ok {
		return stdlibHover(prog.Info.Stdlib, "array", sel.Name, sel.NamePos)
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
	walkFileStatements(prog.File, func(stmt ast.Stmt) {
		if found != nil {
			return
		}
		if let, ok := stmt.(*ast.LetStmt); ok && containsSymbol(pos, let.Pos, let.Name) {
			if typ := prog.Info.ExprTypes[let.Value]; typ != "" && typ != checker.Unknown {
				found = hoverResult(fmt.Sprintf("%s: %s", let.Name, displayCheckerType(prog.Info, typ)), let.Pos, let.Name)
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
				found = hoverResult(fmt.Sprintf("%s: %s", e.Name, displayCheckerType(prog.Info, typ)), e.Pos, e.Name)
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
	if target := s.methodTarget(uri, prog, pos); target != nil {
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
	prog, _ := compiler.AnalyzeSource(uri, text)
	if prog != nil {
		if target := s.methodTarget(uri, prog, pos); target != nil {
			if target.structName == "" {
				return nil
			}
			return map[string]any{
				"changes": map[string]any{uri: methodRenameEdits(prog, target, newName)},
			}
		}
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
	if _, ok := checker.ArrayElement(receiver); ok {
		return stdlibTarget(prog.Info.Stdlib, "array", sel.Name)
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
		return fmt.Sprintf("(%s) -> %s", strings.Join(params, ", "), displayCheckerTypeIndent(info, checker.Type(ret), indent))
	}
	return name
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

func walkFileExprs(file *ast.File, visit func(ast.Expr)) {
	for _, typ := range file.Types {
		for _, method := range typ.Methods {
			ast.WalkExpr(method.Body, visit)
		}
	}
	for _, fn := range file.Functions {
		ast.WalkExpr(fn.Body, visit)
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

func fileURI(path string) string {
	return (&url.URL{Scheme: "file", Path: path}).String()
}
