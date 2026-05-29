package compiler

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/kdy1/go-typescript-eslint/pkg/typescriptestree"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

func parseTypeScriptImport(path string, src string, importPos lexer.Position) (ast.TSImport, []Diagnostic) {
	imp, diags := parseTypeScriptImportAST(path, src, importPos)
	if len(diags) == 0 {
		return imp, nil
	}
	fallback := scanTypeScriptExportFunctions(path, src, importPos)
	if len(fallback.Functions) > 0 || len(fallback.Values) > 0 {
		return fallback, nil
	}
	return imp, diags
}

func parseTypeScriptImportAST(path string, src string, importPos lexer.Position) (ast.TSImport, []Diagnostic) {
	imp := ast.TSImport{Path: path, Pos: importPos}
	opts := typescriptestree.NewBuilder().
		WithSourceType(typescriptestree.SourceTypeModule).
		WithRange(true).
		MustBuild()
	result, err := typescriptestree.Parse(src, opts)
	if err != nil {
		return imp, []Diagnostic{{Message: fmt.Sprintf("parse TypeScript import: %v", err), Path: path}}
	}
	data, err := json.Marshal(result.AST.Body)
	if err != nil {
		return imp, []Diagnostic{{Message: fmt.Sprintf("read TypeScript AST: %v", err), Path: path}}
	}
	var statements []map[string]any
	if err := json.Unmarshal(data, &statements); err != nil {
		return imp, []Diagnostic{{Message: fmt.Sprintf("read TypeScript AST: %v", err), Path: path}}
	}
	var diags []Diagnostic
	for _, stmt := range statements {
		fn, ok, fnDiags := exportedTypeScriptFunction(path, src, stmt)
		diags = append(diags, fnDiags...)
		if ok {
			imp.Functions = append(imp.Functions, fn)
		}
		values, valueDiags := exportedTypeScriptValues(path, src, stmt)
		diags = append(diags, valueDiags...)
		imp.Values = append(imp.Values, values...)
	}
	return imp, diags
}

var typeScriptExportFunctionRE = regexp.MustCompile(`(?m)\bexport\s+(async\s+)?function\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*\(([^)]*)\)\s*(?::\s*([^{;\n]+))?`)
var typeScriptExportValueRE = regexp.MustCompile(`(?m)\bexport\s+(?:const|let|var)\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*(?::\s*([^=;\n]+))?`)

func scanTypeScriptExportFunctions(path string, src string, importPos lexer.Position) ast.TSImport {
	imp := ast.TSImport{Path: path, Pos: importPos}
	matches := typeScriptExportFunctionRE.FindAllStringSubmatchIndex(src, -1)
	for _, match := range matches {
		if len(match) < 10 {
			continue
		}
		name := src[match[4]:match[5]]
		fn := ast.TSFunction{
			Name:       name,
			Routine:    match[2] >= 0,
			ReturnType: ast.RawType(typeScriptTextType(src, match[8], match[9], "Dynamic")),
			Pos:        positionFromOffset(src, match[0]),
			NamePos:    positionFromOffset(src, match[4]),
			SourcePath: path,
		}
		for _, param := range splitTypeScriptParams(src[match[6]:match[7]]) {
			name, typ := parseTypeScriptParamText(param)
			if name == "" {
				continue
			}
			paramOffset := strings.Index(src[match[6]:match[7]], param)
			pos := positionFromOffset(src, match[6]+max(paramOffset, 0))
			fn.Params = append(fn.Params, ast.Param{Name: name, Type: ast.RawType(typ), Pos: pos})
		}
		imp.Functions = append(imp.Functions, fn)
	}
	valueMatches := typeScriptExportValueRE.FindAllStringSubmatchIndex(src, -1)
	for _, match := range valueMatches {
		if len(match) < 6 {
			continue
		}
		name := src[match[2]:match[3]]
		imp.Values = append(imp.Values, ast.TSValue{
			Name:       name,
			Type:       ast.RawType(typeScriptTextType(src, match[4], match[5], "Dynamic")),
			Pos:        positionFromOffset(src, match[0]),
			NamePos:    positionFromOffset(src, match[2]),
			SourcePath: path,
		})
	}
	return imp
}

func splitTypeScriptParams(src string) []string {
	var out []string
	depth := 0
	start := 0
	for i, ch := range src {
		switch ch {
		case '(', '[', '{', '<':
			depth++
		case ')', ']', '}', '>':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				out = append(out, strings.TrimSpace(src[start:i]))
				start = i + 1
			}
		}
	}
	last := strings.TrimSpace(src[start:])
	if last != "" {
		out = append(out, last)
	}
	return out
}

func parseTypeScriptParamText(src string) (string, string) {
	src = strings.TrimSpace(src)
	if src == "" {
		return "", "Dynamic"
	}
	colon := strings.Index(src, ":")
	if colon < 0 {
		return strings.TrimSpace(strings.TrimPrefix(src, "...")), "Dynamic"
	}
	name := strings.TrimSpace(strings.TrimPrefix(src[:colon], "..."))
	name = strings.TrimSuffix(name, "?")
	return name, typeScriptTextType(src, colon+1, len(src), "Dynamic")
}

func typeScriptTextType(src string, start int, end int, fallback string) string {
	if start < 0 || end < 0 || start > end || end > len(src) {
		return fallback
	}
	switch strings.TrimSpace(src[start:end]) {
	case "string":
		return "String"
	case "boolean":
		return "Bool"
	case "bigint":
		return "BigInt"
	case "number":
		return "Double"
	case "void", "undefined":
		return "Void"
	default:
		return fallback
	}
}

func exportedTypeScriptFunction(path string, src string, stmt map[string]any) (ast.TSFunction, bool, []Diagnostic) {
	if stringField(stmt, "type") != "ExportNamedDeclaration" {
		return ast.TSFunction{}, false, nil
	}
	if kind, ok := stmt["exportKind"].(string); ok && kind == "type" {
		return ast.TSFunction{}, false, nil
	}
	decl, ok := mapField(stmt, "declaration")
	if !ok || stringField(decl, "type") != "FunctionDeclaration" {
		return ast.TSFunction{}, false, nil
	}
	id, ok := mapField(decl, "id")
	if !ok {
		pos := positionFromOffset(src, rangeStart(decl))
		return ast.TSFunction{}, false, []Diagnostic{{Message: "unsupported anonymous TypeScript export", Pos: pos, Path: path}}
	}
	name := stringField(id, "name")
	if name == "" {
		pos := positionFromOffset(src, rangeStart(id))
		return ast.TSFunction{}, false, []Diagnostic{{Message: "unsupported anonymous TypeScript export", Pos: pos, Path: path}}
	}
	fn := ast.TSFunction{
		Name:       name,
		Routine:    boolField(decl, "async"),
		ReturnType: ast.RawType(typeScriptReturnType(decl)),
		Pos:        positionFromOffset(src, rangeStart(decl)),
		NamePos:    positionFromOffset(src, rangeStart(id)),
		SourcePath: path,
	}
	params, ok, diags := typeScriptParams(path, src, name, decl)
	if !ok {
		return ast.TSFunction{}, false, diags
	}
	fn.Params = params
	return fn, true, nil
}

func exportedTypeScriptValues(path string, src string, stmt map[string]any) ([]ast.TSValue, []Diagnostic) {
	if stringField(stmt, "type") != "ExportNamedDeclaration" {
		return nil, nil
	}
	if kind, ok := stmt["exportKind"].(string); ok && kind == "type" {
		return nil, nil
	}
	decl, ok := mapField(stmt, "declaration")
	if !ok || stringField(decl, "type") != "VariableDeclaration" {
		return nil, nil
	}
	declarations, ok := arrayField(decl, "declarations")
	if !ok {
		return nil, nil
	}
	values := make([]ast.TSValue, 0, len(declarations))
	var diags []Diagnostic
	for _, item := range declarations {
		declarator, ok := item.(map[string]any)
		if !ok {
			continue
		}
		id, ok := mapField(declarator, "id")
		if !ok || stringField(id, "type") != "Identifier" {
			pos := positionFromOffset(src, rangeStart(declarator))
			diags = append(diags, Diagnostic{Message: "unsupported exported TypeScript variable pattern", Pos: pos, Path: path})
			continue
		}
		name := stringField(id, "name")
		if name == "" {
			continue
		}
		values = append(values, ast.TSValue{
			Name:       name,
			Type:       ast.RawType(typeScriptAnnotatedType(id, "Dynamic")),
			Pos:        positionFromOffset(src, rangeStart(declarator)),
			NamePos:    positionFromOffset(src, rangeStart(id)),
			SourcePath: path,
		})
	}
	return values, diags
}

func typeScriptParams(path string, src string, fnName string, decl map[string]any) ([]ast.Param, bool, []Diagnostic) {
	values, ok := arrayField(decl, "params")
	if !ok {
		return nil, true, nil
	}
	params := make([]ast.Param, 0, len(values))
	var diags []Diagnostic
	for _, value := range values {
		param, ok := value.(map[string]any)
		if !ok || stringField(param, "type") != "Identifier" {
			pos := positionFromOffset(src, rangeStart(param))
			diags = append(diags, Diagnostic{
				Message: fmt.Sprintf("unsupported parameter in exported TypeScript function %q", fnName),
				Pos:     pos,
				Path:    path,
			})
			continue
		}
		name := stringField(param, "name")
		if name == "" {
			name = "_"
		}
		params = append(params, ast.Param{
			Name: name,
			Type: ast.RawType(typeScriptAnnotatedType(param, "Dynamic")),
			Pos:  positionFromOffset(src, rangeStart(param)),
		})
	}
	if len(diags) > 0 {
		return nil, false, diags
	}
	return params, true, nil
}

func typeScriptReturnType(decl map[string]any) string {
	return typeScriptAnnotatedType(decl, "Dynamic")
}

func typeScriptAnnotatedType(node map[string]any, fallback string) string {
	annotation, ok := mapField(node, "typeAnnotation")
	if !ok {
		return fallback
	}
	typ, ok := mapField(annotation, "typeAnnotation")
	if !ok {
		return fallback
	}
	return typeScriptTypeNode(typ)
}

func typeScriptTypeNode(node map[string]any) string {
	switch stringField(node, "type") {
	case "TSStringKeyword":
		return "String"
	case "TSBooleanKeyword":
		return "Bool"
	case "TSBigIntKeyword":
		return "BigInt"
	case "TSNumberKeyword":
		return "Double"
	case "TSVoidKeyword", "TSUndefinedKeyword":
		return "Void"
	default:
		return "Dynamic"
	}
}

func mapField(node map[string]any, name string) (map[string]any, bool) {
	value, ok := node[name]
	if !ok || value == nil {
		return nil, false
	}
	out, ok := value.(map[string]any)
	return out, ok
}

func arrayField(node map[string]any, name string) ([]any, bool) {
	value, ok := node[name]
	if !ok || value == nil {
		return nil, false
	}
	out, ok := value.([]any)
	return out, ok
}

func stringField(node map[string]any, name string) string {
	value, ok := node[name].(string)
	if !ok {
		return ""
	}
	return value
}

func boolField(node map[string]any, name string) bool {
	value, ok := node[name].(bool)
	return ok && value
}

func rangeStart(node map[string]any) int {
	value, ok := node["range"].([]any)
	if !ok || len(value) == 0 {
		return 0
	}
	start, ok := value[0].(float64)
	if !ok {
		return 0
	}
	return int(start)
}

func positionFromOffset(src string, offset int) lexer.Position {
	if offset < 0 {
		offset = 0
	}
	pos := lexer.Position{Offset: offset, Line: 1, Column: 1}
	for i, ch := range src {
		if i >= offset {
			break
		}
		if ch == '\n' {
			pos.Line++
			pos.Column = 1
			continue
		}
		pos.Column++
	}
	return pos
}
