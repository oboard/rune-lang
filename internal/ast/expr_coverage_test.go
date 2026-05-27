package ast

import (
	goast "go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
)

func TestExpressionNodesHavePipelineCoverage(t *testing.T) {
	root := repoRoot(t)
	astExprs := exprNodeTypes(t,
		filepath.Join(root, "internal", "ast", "expr.go"),
		filepath.Join(root, "internal", "ast", "stmt.go"),
	)
	irExprs := exprNodeTypes(t, filepath.Join(root, "internal", "ir", "expr.go"))

	assertTypeSwitchCovers(t, astExprs, filepath.Join(root, "internal", "ast", "walk.go"), "WalkExpr")
	assertTypeSwitchCovers(t, astExprs, filepath.Join(root, "internal", "checker", "infer.go"), "inferExprType")
	assertTypeSwitchCovers(t, astExprs, filepath.Join(root, "internal", "ir", "lower.go"), "expr")
	assertTypeSwitchCovers(t, astExprs, filepath.Join(root, "internal", "format", "expr.go"), "expr")

	assertTypeSwitchCovers(t, irExprs, filepath.Join(root, "internal", "ir", "walk.go"), "WalkExpr")
	assertTypeSwitchCovers(t, irExprs, filepath.Join(root, "internal", "interpreter", "expr.go"), "eval")
	assertTypeSwitchCovers(t, irExprs, filepath.Join(root, "internal", "codegen", "go", "expr.go"), "exprPrec")
	assertTypeSwitchCovers(t, irExprs, filepath.Join(root, "internal", "codegen", "typescript", "expr.go"), "exprPrec")
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func exprNodeTypes(t *testing.T, paths ...string) []string {
	t.Helper()
	seen := map[string]bool{}
	for _, path := range paths {
		file := parseGoFile(t, path)
		for _, decl := range file.Decls {
			fn, ok := decl.(*goast.FuncDecl)
			if !ok || fn.Name.Name != "exprNode" || fn.Recv == nil || len(fn.Recv.List) == 0 {
				continue
			}
			if name := receiverTypeName(fn.Recv.List[0].Type); name != "" {
				seen[name] = true
			}
		}
	}
	return sortedKeys(seen)
}

func assertTypeSwitchCovers(t *testing.T, want []string, path string, funcName string) {
	t.Helper()
	cases := typeSwitchCasesInFunc(t, path, funcName)
	var missing []string
	for _, name := range want {
		if !cases[name] {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("%s:%s is missing expression cases: %v", filepath.ToSlash(path), funcName, missing)
	}
}

func typeSwitchCasesInFunc(t *testing.T, path string, funcName string) map[string]bool {
	t.Helper()
	file := parseGoFile(t, path)
	cases := map[string]bool{}
	for _, decl := range file.Decls {
		fn, ok := decl.(*goast.FuncDecl)
		if !ok || fn.Name.Name != funcName {
			continue
		}
		goast.Inspect(fn.Body, func(node goast.Node) bool {
			stmt, ok := node.(*goast.TypeSwitchStmt)
			if !ok {
				return true
			}
			for _, bodyStmt := range stmt.Body.List {
				clause, ok := bodyStmt.(*goast.CaseClause)
				if !ok {
					continue
				}
				for _, expr := range clause.List {
					if name := caseTypeName(expr); name != "" {
						cases[name] = true
					}
				}
			}
			return true
		})
	}
	if len(cases) == 0 {
		t.Fatalf("%s:%s has no type switch cases", filepath.ToSlash(path), funcName)
	}
	return cases
}

func parseGoFile(t *testing.T, path string) *goast.File {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", filepath.ToSlash(path), err)
	}
	return file
}

func receiverTypeName(expr goast.Expr) string {
	if star, ok := expr.(*goast.StarExpr); ok {
		return receiverTypeName(star.X)
	}
	if ident, ok := expr.(*goast.Ident); ok {
		return ident.Name
	}
	return ""
}

func caseTypeName(expr goast.Expr) string {
	if star, ok := expr.(*goast.StarExpr); ok {
		return caseTypeName(star.X)
	}
	switch e := expr.(type) {
	case *goast.Ident:
		return e.Name
	case *goast.SelectorExpr:
		return e.Sel.Name
	default:
		return ""
	}
}

func sortedKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
