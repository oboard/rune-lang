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
	walkFileAnnotations(file, func(annotation *ast.Annotation) {
		for _, arg := range annotation.Args {
			ast.WalkExpr(arg, visit)
		}
	})
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

func walkFileAnnotations(file *ast.File, visit func(*ast.Annotation)) {
	for _, typ := range file.Types {
		walkAnnotations(typ.Annotations, visit)
		for i := range typ.Fields {
			walkAnnotations(typ.Fields[i].Annotations, visit)
		}
		for _, method := range typ.Methods {
			walkAnnotations(method.Annotations, visit)
		}
	}
	for _, enum := range file.Enums {
		walkAnnotations(enum.Annotations, visit)
		for i := range enum.Members {
			walkAnnotations(enum.Members[i].Annotations, visit)
		}
	}
	for _, fn := range file.Functions {
		walkAnnotations(fn.Annotations, visit)
	}
}

func walkDocumentAnnotations(uri string, file *ast.File, visit func(*ast.Annotation)) {
	for _, typ := range file.Types {
		if !sourceMatchesDocument(uri, typ.SourcePath) {
			continue
		}
		walkAnnotations(typ.Annotations, visit)
		for i := range typ.Fields {
			walkAnnotations(typ.Fields[i].Annotations, visit)
		}
		for _, method := range typ.Methods {
			walkAnnotations(method.Annotations, visit)
		}
	}
	for _, enum := range file.Enums {
		if !sourceMatchesDocument(uri, enum.SourcePath) {
			continue
		}
		walkAnnotations(enum.Annotations, visit)
		for i := range enum.Members {
			walkAnnotations(enum.Members[i].Annotations, visit)
		}
	}
	for _, fn := range file.Functions {
		if sourceMatchesDocument(uri, fn.SourcePath) {
			walkAnnotations(fn.Annotations, visit)
		}
	}
}

func walkAnnotations(annotations []ast.Annotation, visit func(*ast.Annotation)) {
	for i := range annotations {
		visit(&annotations[i])
	}
}

func annotationAt(file *ast.File, pos position) *ast.Annotation {
	var found *ast.Annotation
	walkFileAnnotations(file, func(annotation *ast.Annotation) {
		if found == nil && containsSymbol(pos, annotation.NamePos, annotation.Name) {
			found = annotation
		}
	})
	return found
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
	walkDocumentAnnotations(uri, file, func(annotation *ast.Annotation) {
		for _, arg := range annotation.Args {
			ast.WalkExpr(arg, visit)
		}
	})
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
