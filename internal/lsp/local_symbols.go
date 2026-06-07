package lsp

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/compiler"
	"github.com/oboard/rune-lang/internal/lexer"
)

func localTarget(uri string, prog *compiler.Program, pos position) *methodTarget {
	if target := localDeclarationTarget(uri, prog.File, pos); target != nil {
		return target
	}
	if ident := identifierAt(prog.File, pos); ident != nil {
		return localTargetForIdentifier(uri, prog.File, ident)
	}
	name := letNameAt(prog.File, pos)
	if name == "" {
		return nil
	}
	if target := letTarget(uri, prog.File, name); target != nil {
		return target
	}
	return nil
}

type localParam struct {
	name string
	pos  lexer.Position
}

type localScope struct {
	uri    string
	body   ast.Expr
	params []localParam
	parent *localScope
}

func localDeclarationTarget(uri string, file *ast.File, pos position) *methodTarget {
	if target := paramDeclarationTarget(uri, file, pos); target != nil {
		return target
	}
	return nil
}

func paramDeclarationTarget(uri string, file *ast.File, pos position) *methodTarget {
	for _, typ := range file.Types {
		for _, method := range typ.Methods {
			scopeURI := sourceURI(uri, method.SourcePath)
			for _, param := range method.Params {
				if containsSymbol(pos, param.Pos, param.Name) {
					return &methodTarget{uri: scopeURI, name: param.Name, pos: param.Pos, scope: method.Body}
				}
			}
		}
	}
	for _, fn := range file.Functions {
		scopeURI := sourceURI(uri, fn.SourcePath)
		for _, param := range fn.Params {
			if containsSymbol(pos, param.Pos, param.Name) {
				return &methodTarget{uri: scopeURI, name: param.Name, pos: param.Pos, scope: fn.Body}
			}
		}
	}
	var found *methodTarget
	walkFileExprs(file, func(expr ast.Expr) {
		if found != nil {
			return
		}
		lambda, ok := expr.(*ast.LambdaExpr)
		if !ok {
			return
		}
		for i, name := range lambda.Params {
			if i >= len(lambda.ParamPos) || !containsSymbol(pos, lambda.ParamPos[i], name) {
				continue
			}
			found = &methodTarget{uri: uri, name: name, pos: lambda.ParamPos[i], scope: lambda.Body}
			return
		}
	})
	return found
}

func localTargetForIdentifier(uri string, file *ast.File, ident *ast.Identifier) *methodTarget {
	for scope := localScopeForExpr(uri, file, ident); scope != nil; scope = scope.parent {
		if target := letTargetInScope(scope, ident.Name, ident.Pos); target != nil {
			return target
		}
		if target := paramTargetInScope(scope, ident.Name); target != nil {
			return target
		}
	}
	return nil
}

func localScopeForExpr(uri string, file *ast.File, target ast.Expr) *localScope {
	for _, typ := range file.Types {
		for _, method := range typ.Methods {
			if !exprTreeContains(method.Body, target) {
				continue
			}
			scope := &localScope{
				uri:    sourceURI(uri, method.SourcePath),
				body:   method.Body,
				params: functionLocalParams(method.Params),
			}
			return innermostLocalScope(scope, target)
		}
	}
	for _, fn := range file.Functions {
		if !exprTreeContains(fn.Body, target) {
			continue
		}
		scope := &localScope{
			uri:    sourceURI(uri, fn.SourcePath),
			body:   fn.Body,
			params: functionLocalParams(fn.Params),
		}
		return innermostLocalScope(scope, target)
	}
	for _, test := range file.Tests {
		if !exprTreeContains(test.Body, target) {
			continue
		}
		scope := &localScope{
			uri:  sourceURI(uri, test.SourcePath),
			body: test.Body,
		}
		return innermostLocalScope(scope, target)
	}
	return nil
}

func innermostLocalScope(scope *localScope, target ast.Expr) *localScope {
	var child *localScope
	ast.WalkExpr(scope.body, func(expr ast.Expr) {
		if child != nil {
			return
		}
		lambda, ok := expr.(*ast.LambdaExpr)
		if !ok || !exprTreeContains(lambda.Body, target) {
			return
		}
		child = &localScope{
			uri:    scope.uri,
			body:   lambda.Body,
			params: lambdaLocalParams(lambda),
			parent: scope,
		}
	})
	if child != nil {
		return innermostLocalScope(child, target)
	}
	return scope
}

func exprTreeContains(root ast.Expr, target ast.Expr) bool {
	found := false
	ast.WalkExpr(root, func(expr ast.Expr) {
		if expr == target {
			found = true
		}
	})
	return found
}

func functionLocalParams(params []ast.Param) []localParam {
	out := make([]localParam, 0, len(params))
	for _, param := range params {
		out = append(out, localParam{name: param.Name, pos: param.Pos})
	}
	return out
}

func lambdaLocalParams(lambda *ast.LambdaExpr) []localParam {
	out := make([]localParam, 0, len(lambda.Params))
	for i, name := range lambda.Params {
		if i >= len(lambda.ParamPos) {
			continue
		}
		out = append(out, localParam{name: name, pos: lambda.ParamPos[i]})
	}
	return out
}

func letTargetInScope(scope *localScope, name string, before lexer.Position) *methodTarget {
	var found *methodTarget
	bestOffset := -1
	walkExprStatements(scope.body, func(stmt ast.Stmt) {
		switch stmt := stmt.(type) {
		case *ast.LetStmt:
			if stmt.Name != name || !lexerPositionBeforeOrEqual(stmt.Pos, before) {
				return
			}
			if offset := lexerPositionOffset(stmt.Pos); offset > bestOffset {
				bestOffset = offset
				found = &methodTarget{uri: scope.uri, name: stmt.Name, pos: stmt.Pos, scope: scope.body}
			}
		case *ast.ObjectDestructureStmt:
			for _, field := range stmt.Fields {
				if field.Name != name || !lexerPositionBeforeOrEqual(field.NamePos, before) {
					continue
				}
				if offset := lexerPositionOffset(field.NamePos); offset > bestOffset {
					bestOffset = offset
					found = &methodTarget{uri: scope.uri, name: field.Name, pos: field.NamePos, scope: scope.body}
				}
			}
		}
	})
	return found
}

func walkExprStatements(expr ast.Expr, visit func(ast.Stmt)) {
	ast.WalkExpr(expr, func(candidate ast.Expr) {
		block, ok := candidate.(*ast.BlockExpr)
		if !ok {
			return
		}
		if candidate != expr && nestedLambdaContainsExpr(expr, candidate) {
			return
		}
		for _, stmt := range block.Statements {
			visit(stmt)
		}
	})
}

func nestedLambdaContainsExpr(root ast.Expr, target ast.Expr) bool {
	found := false
	ast.WalkExpr(root, func(expr ast.Expr) {
		if found {
			return
		}
		lambda, ok := expr.(*ast.LambdaExpr)
		if !ok {
			return
		}
		found = exprTreeContains(lambda.Body, target)
	})
	return found
}

func lexerPositionBeforeOrEqual(a lexer.Position, b lexer.Position) bool {
	if a.Offset > 0 && b.Offset > 0 {
		return a.Offset <= b.Offset
	}
	if a.Line != b.Line {
		return a.Line < b.Line
	}
	return a.Column <= b.Column
}

func lexerPositionOffset(pos lexer.Position) int {
	if pos.Offset > 0 {
		return pos.Offset
	}
	return (pos.Line * 1_000_000) + pos.Column
}

func paramTargetInScope(scope *localScope, name string) *methodTarget {
	for _, param := range scope.params {
		if param.name == name {
			return &methodTarget{uri: scope.uri, name: param.name, pos: param.pos, scope: scope.body}
		}
	}
	return nil
}

func identifierAt(file *ast.File, pos position) *ast.Identifier {
	var found string
	var foundIdent *ast.Identifier
	walkFileExprs(file, func(expr ast.Expr) {
		if found != "" {
			return
		}
		if ident, ok := expr.(*ast.Identifier); ok && containsSymbol(pos, ident.Pos, ident.Name) {
			found = ident.Name
			foundIdent = ident
		}
	})
	return foundIdent
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
		if destructure, ok := stmt.(*ast.ObjectDestructureStmt); ok {
			for _, field := range destructure.Fields {
				if containsSymbol(pos, field.NamePos, field.Name) {
					found = field.Name
					return
				}
			}
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
		if destructure, ok := stmt.(*ast.ObjectDestructureStmt); ok {
			for _, field := range destructure.Fields {
				if field.Name == name {
					found = &methodTarget{uri: uri, name: field.Name, pos: field.NamePos}
					return
				}
			}
		}
	})
	return found
}
