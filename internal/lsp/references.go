package lsp

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/compiler"
	"github.com/oboard/rune-lang/internal/lexer"
)

func (s *server) references(uri string, pos position, includeDeclaration bool) any {
	prog, _ := s.analyze(uri)
	if prog == nil {
		return []map[string]any{}
	}
	if target := functionTarget(uri, prog, pos); target != nil {
		return functionReferences(prog, target, includeDeclaration)
	}
	if target := s.methodTarget(uri, prog, pos); target != nil && target.structName != "" {
		return methodReferences(prog, target, includeDeclaration)
	}
	if target := localTarget(uri, prog, pos); target != nil && target.scope != nil {
		return localReferences(uri, prog, target, includeDeclaration)
	}
	return []map[string]any{}
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

func localReferences(uri string, prog *compiler.Program, target *methodTarget, includeDeclaration bool) []map[string]any {
	refs := []map[string]any{}
	if includeDeclaration {
		refs = append(refs, target.location())
	}
	ast.WalkExpr(target.scope, func(expr ast.Expr) {
		ident, ok := expr.(*ast.Identifier)
		if !ok || ident.Name != target.name {
			return
		}
		resolved := localTarget(uri, prog, positionFromLexer(ident.Pos))
		if !sameLocalTarget(resolved, target) {
			return
		}
		refs = append(refs, referenceLocation(target.uri, ident.Pos, ident.Name))
	})
	return refs
}

func sameLocalTarget(a *methodTarget, b *methodTarget) bool {
	if a == nil || b == nil || a.scope == nil || b.scope == nil {
		return false
	}
	return a.uri == b.uri && a.name == b.name && sameLexerPosition(a.pos, b.pos)
}

func sameLexerPosition(a lexer.Position, b lexer.Position) bool {
	return a.Offset == b.Offset && a.Line == b.Line && a.Column == b.Column
}

func positionFromLexer(pos lexer.Position) position {
	return position{
		Line:      max(pos.Line-1, 0),
		Character: max(pos.Column-1, 0),
	}
}

func referenceLocation(uri string, pos lexer.Position, name string) map[string]any {
	return map[string]any{
		"uri":   uri,
		"range": symbolRange(pos, len(name)),
	}
}
