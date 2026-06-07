package lsp

import (
	"github.com/oboard/rune-lang/internal/ast"
)

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

func walkTemplateExprs(file *ast.File, visit func(ast.Expr)) {
	walkFileExprs(file, func(expr ast.Expr) {
		lit, ok := expr.(*ast.TemplateLiteral)
		if !ok {
			return
		}
		for _, part := range lit.Parts {
			ast.WalkExpr(part.Expr, visit)
		}
	})
}

func walkDocumentExprs(uri string, file *ast.File, visit func(ast.Expr)) {
	for _, typ := range file.Types {
		if !sourceMatchesDocument(uri, typ.SourcePath) {
			continue
		}
		for _, method := range typ.Methods {
			ast.WalkExpr(method.Body, visit)
		}
	}
	for _, fn := range file.Functions {
		if !sourceMatchesDocument(uri, fn.SourcePath) {
			continue
		}
		ast.WalkExpr(fn.Body, visit)
	}
	for _, test := range file.Tests {
		if !sourceMatchesDocument(uri, test.SourcePath) {
			continue
		}
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
