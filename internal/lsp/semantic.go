package lsp

import (
	"sort"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/compiler"
	"github.com/oboard/rune-lang/internal/lexer"
)

func (s *server) semanticTokens(uri string) any {
	prog, _ := s.analyze(uri)
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
	for _, enum := range prog.File.Enums {
		tokens = append(tokens, semanticTokenFor(enum.NamePos, enum.Name, semanticTokenTypeEnum, 0))
		for _, member := range enum.Members {
			tokens = append(tokens, semanticTokenFor(member.Pos, member.Name, semanticTokenTypeEnumMember, 0))
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
		case *ast.ObjectDestructureStmt:
			for _, field := range stmt.Fields {
				if _, ok := signals[field.Name]; !ok {
					continue
				}
				tokens = append(tokens, semanticToken{
					line:      max(field.NamePos.Line-1, 0),
					character: max(field.NamePos.Column-1, 0),
					length:    len(field.Name),
					tokenType: semanticTokenTypeVariable,
					modifiers: semanticTokenModifierModification,
				})
			}
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
		case *ast.SelectorExpr:
			tokens = append(tokens, enumSelectorSemanticTokens(prog, expr)...)
		}
	})
	tokens = append(tokens, templateExpressionSemanticTokens(prog, signals)...)
	sort.Slice(tokens, func(i, j int) bool {
		if tokens[i].line != tokens[j].line {
			return tokens[i].line < tokens[j].line
		}
		return tokens[i].character < tokens[j].character
	})
	return map[string]any{"data": encodeSemanticTokens(tokens)}
}

func templateExpressionSemanticTokens(prog *compiler.Program, signals map[string][]string) []semanticToken {
	var tokens []semanticToken
	walkTemplateExprs(prog.File, func(expr ast.Expr) {
		switch expr := expr.(type) {
		case *ast.CallExpr:
			if token, ok := templateCallSemanticToken(prog, expr); ok {
				tokens = append(tokens, token)
			}
		case *ast.Identifier:
			if token, ok := templateIdentifierSemanticToken(prog, signals, expr); ok {
				tokens = append(tokens, token)
			}
		}
	})
	return tokens
}

func templateCallSemanticToken(prog *compiler.Program, call *ast.CallExpr) (semanticToken, bool) {
	if _, ok := asyncCallSemanticToken(prog, call); ok {
		return semanticToken{}, false
	}
	switch callee := call.Callee.(type) {
	case *ast.Identifier:
		return semanticToken{}, false
	case *ast.SelectorExpr:
		if !selectorCallResolves(prog, callee) {
			return semanticToken{}, false
		}
		return semanticTokenFor(callee.NamePos, callee.Name, semanticTokenTypeFunction, 0), true
	default:
		return semanticToken{}, false
	}
}

func selectorCallResolves(prog *compiler.Program, sel *ast.SelectorExpr) bool {
	if at, ok := sel.Receiver.(*ast.AtExpr); ok && at.Name != "" {
		_, ok := prog.Info.Stdlib.Function(at.Name, sel.Name)
		return ok
	}
	receiver := prog.Info.ExprTypes[sel.Receiver]
	if moduleName, ok := checker.ModuleNamespaceName(receiver); ok {
		_, ok := prog.Info.Stdlib.Function(moduleName, sel.Name)
		return ok
	}
	if prog.Info.ResolvedSelectorFunctions[sel] != nil {
		return true
	}
	if moduleName, receiverName, ok := stdlibReceiverModule(receiver); ok {
		_, ok := prog.Info.Stdlib.ReceiverFunction(moduleName, receiverName, sel.Name)
		return ok
	}
	structInfo := prog.Info.Types[baseType(receiver)]
	return structInfo != nil && structInfo.Methods[sel.Name] != nil
}

func templateIdentifierSemanticToken(prog *compiler.Program, signals map[string][]string, ident *ast.Identifier) (semanticToken, bool) {
	if ident.Name == "" || ident.Name == "<error>" {
		return semanticToken{}, false
	}
	if _, signal := signals[ident.Name]; signal {
		return semanticToken{}, false
	}
	if fn := prog.Info.ResolvedFunctions[ident]; fn != nil {
		if fn.Routine {
			return semanticToken{}, false
		}
		return semanticTokenFor(ident.Pos, ident.Name, semanticTokenTypeFunction, 0), true
	}
	if _, ok := prog.Info.Types[ident.Name]; ok {
		return semanticTokenFor(ident.Pos, ident.Name, semanticTokenTypeType, 0), true
	}
	if _, ok := prog.Info.Enums[ident.Name]; ok {
		return semanticTokenFor(ident.Pos, ident.Name, semanticTokenTypeEnum, 0), true
	}
	if prog.Info.ResolvedValues[ident] != nil {
		return semanticTokenFor(ident.Pos, ident.Name, semanticTokenTypeVariable, 0), true
	}
	typ := prog.Info.ExprTypes[ident]
	if typ == "" || typ == checker.Unknown {
		return semanticToken{}, false
	}
	return semanticTokenFor(ident.Pos, ident.Name, semanticTokenTypeVariable, 0), true
}

func enumSelectorSemanticTokens(prog *compiler.Program, sel *ast.SelectorExpr) []semanticToken {
	ident, ok := sel.Receiver.(*ast.Identifier)
	if !ok {
		return nil
	}
	enum := prog.Info.Enums[ident.Name]
	if enum == nil {
		return nil
	}
	if _, ok := enum.ByName[sel.Name]; !ok {
		return nil
	}
	return []semanticToken{
		semanticTokenFor(ident.Pos, ident.Name, semanticTokenTypeEnum, 0),
		semanticTokenFor(sel.NamePos, sel.Name, semanticTokenTypeEnumMember, 0),
	}
}

func semanticTokenFor(pos lexer.Position, name string, tokenType int, modifiers int) semanticToken {
	return semanticToken{
		line:      max(pos.Line-1, 0),
		character: max(pos.Column-1, 0),
		length:    len(name),
		tokenType: tokenType,
		modifiers: modifiers,
	}
}

func signalGraph(file *ast.File) map[string][]string {
	signals := map[string][]string{}
	walkFileStatements(file, func(stmt ast.Stmt) {
		switch stmt := stmt.(type) {
		case *ast.LetStmt:
			deps := exprSignalDeps(stmt.Value, signals)
			if stmt.Signal || len(deps) > 0 {
				signals[stmt.Name] = deps
			}
		case *ast.ObjectDestructureStmt:
			deps := exprSignalDeps(stmt.Value, signals)
			if stmt.Signal || len(deps) > 0 {
				for _, field := range stmt.Fields {
					signals[field.Name] = deps
				}
			}
		}
	})
	return signals
}

func asyncCallSemanticToken(prog *compiler.Program, call *ast.CallExpr) (semanticToken, bool) {
	switch callee := call.Callee.(type) {
	case *ast.Identifier:
		fn := prog.Info.ResolvedFunctions[callee]
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
	if at, ok := sel.Receiver.(*ast.AtExpr); ok && at.Name != "" {
		fn, ok := prog.Info.Stdlib.Function(at.Name, sel.Name)
		return ok && fn.Routine
	}
	receiver := prog.Info.ExprTypes[sel.Receiver]
	if moduleName, ok := checker.ModuleNamespaceName(receiver); ok {
		fn, ok := prog.Info.Stdlib.Function(moduleName, sel.Name)
		return ok && fn.Routine
	}
	if fn := prog.Info.ResolvedSelectorFunctions[sel]; fn != nil {
		return fn.Routine
	}
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
	semanticTokenTypeVariable   = 0
	semanticTokenTypeType       = 1
	semanticTokenTypeFunction   = 2
	semanticTokenTypeEnum       = 3
	semanticTokenTypeEnumMember = 4

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
