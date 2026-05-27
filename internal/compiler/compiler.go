package compiler

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	gocodegen "github.com/oboard/rune-lang/internal/codegen/go"
	tscodegen "github.com/oboard/rune-lang/internal/codegen/typescript"
	"github.com/oboard/rune-lang/internal/ir"
	"github.com/oboard/rune-lang/internal/lexer"
	"github.com/oboard/rune-lang/internal/parser"
	"github.com/oboard/rune-lang/internal/stdlib"
)

type Diagnostic struct {
	Message string
	Pos     lexer.Position
	Path    string
}

type Program struct {
	Path   string
	Source string
	File   *ast.File
	Info   *checker.Info
	IR     *ir.File
}

func AnalyzeFile(path string) (*Program, []Diagnostic) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, []Diagnostic{{Message: err.Error(), Path: path}}
	}
	file, src, parseErrs, loadDiags := loadImportGraph(path, string(data), true)
	reg, err := stdlib.LoadDefault()
	info, checkDiags := checker.CheckWithStdlibForPath(file, reg, path)
	if err != nil {
		checkDiags = append([]checker.Diagnostic{{Message: err.Error()}}, checkDiags...)
	}
	return analyzedProgram(path, src, file, info, parseErrs, checkDiags, loadDiags)
}

func AnalyzeSource(path string, src string) (*Program, []Diagnostic) {
	file, src, parseErrs, loadDiags := loadImportGraph(path, src, true)
	reg, err := stdlib.LoadDefault()
	info, checkDiags := checker.CheckWithStdlibForPath(file, reg, path)
	if err != nil {
		checkDiags = append([]checker.Diagnostic{{Message: err.Error()}}, checkDiags...)
	}
	return analyzedProgram(path, src, file, info, parseErrs, checkDiags, loadDiags)
}

func AnalyzeSourceWithStdlib(path string, src string, reg *stdlib.Registry) (*Program, []Diagnostic) {
	file, src, parseErrs, loadDiags := loadImportGraph(path, src, true)
	info, checkDiags := checker.CheckWithStdlibForPath(file, reg, path)
	return analyzedProgram(path, src, file, info, parseErrs, checkDiags, loadDiags)
}

func analyzedProgram(path string, src string, file *ast.File, info *checker.Info, parseErrs []parser.Error, checkDiags []checker.Diagnostic, diags []Diagnostic) (*Program, []Diagnostic) {
	for _, err := range parseErrs {
		diags = append(diags, Diagnostic{Message: err.Message, Pos: err.Pos, Path: path})
	}
	for _, diag := range checkDiags {
		diags = append(diags, Diagnostic{Message: diag.Message, Pos: diag.Pos, Path: path})
	}
	return &Program{Path: path, Source: src, File: file, Info: info, IR: ir.LowerFile(file, info)}, diags
}

func loadImportGraph(entryPath string, entrySource string, hasEntrySource bool) (*ast.File, string, []parser.Error, []Diagnostic) {
	loader := importLoader{
		files:    map[string]bool{},
		tsFiles:  map[string]bool{},
		visiting: map[string]bool{},
		sources:  map[string]string{},
	}
	if hasEntrySource {
		if normalized, ok := normalizeImportPath(entryPath); ok {
			loader.sources[normalized] = entrySource
		} else {
			loader.sources[entryPath] = entrySource
		}
	}
	file, parseErrs, diags := loader.load(entryPath)
	if file == nil {
		file = &ast.File{}
	}
	return file, entrySource, parseErrs, diags
}

type importLoader struct {
	files    map[string]bool
	tsFiles  map[string]bool
	visiting map[string]bool
	sources  map[string]string
}

func (l *importLoader) load(path string) (*ast.File, []parser.Error, []Diagnostic) {
	normalized, ok := normalizeImportPath(path)
	if !ok {
		file, errs := parser.Parse(l.sources[path])
		annotateSourcePath(file, path)
		return file, errs, nil
	}
	if l.files[normalized] {
		return &ast.File{}, nil, nil
	}
	if l.visiting[normalized] {
		return &ast.File{}, nil, []Diagnostic{{Message: fmt.Sprintf("cyclic import %q", path), Path: path}}
	}
	l.visiting[normalized] = true
	defer delete(l.visiting, normalized)

	src, ok := l.sources[normalized]
	if !ok {
		data, err := os.ReadFile(normalized)
		if err != nil {
			return &ast.File{}, nil, []Diagnostic{{Message: err.Error(), Path: normalized}}
		}
		src = string(data)
		l.sources[normalized] = src
	}
	file, parseErrs := parser.Parse(src)
	annotateSourcePath(file, normalized)
	var diags []Diagnostic
	merged := &ast.File{}
	for _, imp := range file.Imports {
		importPath, err := ResolveRuneImport(normalized, imp.Path)
		if err != nil {
			diags = append(diags, Diagnostic{Message: err.Error(), Pos: imp.Pos, Path: normalized})
			continue
		}
		if filepath.Ext(importPath) == ".ts" {
			tsImport, importedDiags := l.loadTypeScript(importPath, imp.Pos)
			diags = append(diags, importedDiags...)
			if tsImport != nil {
				merged.TSImports = append(merged.TSImports, *tsImport)
			}
			continue
		}
		imported, importedParseErrs, importedDiags := l.load(importPath)
		parseErrs = append(parseErrs, importedParseErrs...)
		diags = append(diags, importedDiags...)
		mergeFile(merged, imported, false)
	}
	mergeFile(merged, file, true)
	l.files[normalized] = true
	return merged, parseErrs, diags
}

func (l *importLoader) loadTypeScript(path string, pos lexer.Position) (*ast.TSImport, []Diagnostic) {
	normalized, ok := normalizeImportPath(path)
	if !ok {
		return nil, []Diagnostic{{Message: fmt.Sprintf("cannot resolve TypeScript import %q", path), Pos: pos, Path: path}}
	}
	if l.tsFiles[normalized] {
		return nil, nil
	}
	src, ok := l.sources[normalized]
	if !ok {
		data, err := os.ReadFile(normalized)
		if err != nil {
			return nil, []Diagnostic{{Message: err.Error(), Pos: pos, Path: normalized}}
		}
		src = string(data)
		l.sources[normalized] = src
	}
	imp, diags := parseTypeScriptImport(normalized, src, pos)
	l.tsFiles[normalized] = true
	return &imp, diags
}

func mergeFile(dst *ast.File, src *ast.File, includeTests bool) {
	if src == nil {
		return
	}
	dst.GoImports = append(dst.GoImports, src.GoImports...)
	dst.TSImports = append(dst.TSImports, src.TSImports...)
	dst.Types = append(dst.Types, src.Types...)
	dst.Enums = append(dst.Enums, src.Enums...)
	dst.Functions = append(dst.Functions, src.Functions...)
	if includeTests {
		dst.Tests = append(dst.Tests, src.Tests...)
	}
}

func annotateSourcePath(file *ast.File, sourcePath string) {
	if file == nil {
		return
	}
	for _, typ := range file.Types {
		typ.SourcePath = sourcePath
		for _, method := range typ.Methods {
			method.SourcePath = sourcePath
		}
	}
	for _, enum := range file.Enums {
		enum.SourcePath = sourcePath
	}
	for _, fn := range file.Functions {
		fn.SourcePath = sourcePath
	}
	for _, test := range file.Tests {
		test.SourcePath = sourcePath
	}
}

func ResolveRuneImport(fromPath string, spec string) (string, error) {
	if spec == "" {
		return "", fmt.Errorf("empty import path")
	}
	if ext := filepath.Ext(spec); ext == "" || ext == "." {
		return "", fmt.Errorf("import path %q must include a file extension", spec)
	}
	candidate := spec
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(filepath.Dir(fromPath), candidate)
	}
	candidate = filepath.Clean(candidate)
	if _, err := os.Stat(candidate); err == nil {
		return filepath.Abs(candidate)
	}
	return "", fmt.Errorf("cannot resolve import %q from %s", spec, fromPath)
}

func normalizeImportPath(path string) (string, bool) {
	if path == "" || strings.HasPrefix(path, "<") {
		return path, false
	}
	if strings.HasPrefix(path, "file://") {
		normalized := checkerPathFromURI(path)
		if normalized != "" {
			path = normalized
		}
	}
	abs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return filepath.Clean(path), true
	}
	return abs, true
}

func checkerPathFromURI(uri string) string {
	parsed, err := url.Parse(uri)
	if err != nil || parsed.Scheme != "file" {
		return ""
	}
	return parsed.Path
}

func GenerateGoFile(path string) (string, []Diagnostic) {
	prog, diags := AnalyzeFile(path)
	if len(diags) > 0 {
		return "", diags
	}
	src, err := gocodegen.GenerateIR(prog.IR)
	if err != nil {
		return "", []Diagnostic{{Message: err.Error()}}
	}
	return src, nil
}

func GenerateTypeScriptFile(path string) (string, []Diagnostic) {
	prog, diags := AnalyzeFile(path)
	if len(diags) > 0 {
		return "", diags
	}
	src, err := tscodegen.GenerateIR(prog.IR)
	if err != nil {
		return "", []Diagnostic{{Message: err.Error()}}
	}
	return src, nil
}

func GenerateTypeScriptSource(path string, src string, reg *stdlib.Registry) (string, []Diagnostic) {
	prog, diags := AnalyzeSourceWithStdlib(path, src, reg)
	if len(diags) > 0 {
		return "", diags
	}
	out, err := tscodegen.GenerateIR(prog.IR)
	if err != nil {
		return "", []Diagnostic{{Message: err.Error()}}
	}
	return out, nil
}

func GeneratedPath(inputPath string) string {
	base := filepath.Base(inputPath)
	ext := filepath.Ext(base)
	if ext != "" {
		base = base[:len(base)-len(ext)]
	}
	return fmt.Sprintf("%s.go", base)
}
