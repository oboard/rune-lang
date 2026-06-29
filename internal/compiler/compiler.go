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
	moonbitcodegen "github.com/oboard/rune-lang/internal/codegen/moonbit"
	tscodegen "github.com/oboard/rune-lang/internal/codegen/typescript"
	"github.com/oboard/rune-lang/internal/ir"
	"github.com/oboard/rune-lang/internal/lexer"
	"github.com/oboard/rune-lang/internal/macro"
	"github.com/oboard/rune-lang/internal/parser"
	"github.com/oboard/rune-lang/internal/stdlib"
)

type Diagnostic struct {
	Message  string
	Pos      lexer.Position
	Path     string
	Severity checker.DiagnosticSeverity
	Code     string
	Kind     string
}

type Program struct {
	Path   string
	Source string
	File   *ast.File
	Info   *checker.Info
	IR     *ir.File
	Macros []macro.Invocation
}

func AnalyzeFile(path string) (*Program, []Diagnostic) {
	return analyzeFile(path, false)
}

func AnalyzeFileWithWarnings(path string) (*Program, []Diagnostic) {
	return analyzeFile(path, true)
}

func analyzeFile(path string, includeWarnings bool) (*Program, []Diagnostic) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, []Diagnostic{{Message: err.Error(), Path: path}}
	}
	file, src, parseErrs, loadDiags := loadImportGraph(path, string(data), true)
	reg, err := stdlib.LoadDefault()
	info, compileTimeFunctions, checkDiags := checkAndExpand(file, reg, path, parseErrs)
	if err != nil {
		checkDiags = append([]checker.Diagnostic{{Message: err.Error()}}, checkDiags...)
	}
	return analyzedProgram(path, src, file, info, compileTimeFunctions, parseErrs, checkDiags, loadDiags, includeWarnings)
}

func AnalyzeSource(path string, src string) (*Program, []Diagnostic) {
	return analyzeSource(path, src, nil, false)
}

func AnalyzeSourceWithWarnings(path string, src string) (*Program, []Diagnostic) {
	return analyzeSource(path, src, nil, true)
}

func analyzeSource(path string, src string, reg *stdlib.Registry, includeWarnings bool) (*Program, []Diagnostic) {
	file, src, parseErrs, loadDiags := loadImportGraph(path, src, true)
	var err error
	if reg == nil {
		reg, err = stdlib.LoadDefault()
	}
	info, compileTimeFunctions, checkDiags := checkAndExpand(file, reg, path, parseErrs)
	if err != nil {
		checkDiags = append([]checker.Diagnostic{{Message: err.Error()}}, checkDiags...)
	}
	return analyzedProgram(path, src, file, info, compileTimeFunctions, parseErrs, checkDiags, loadDiags, includeWarnings)
}

func AnalyzeSourceWithStdlib(path string, src string, reg *stdlib.Registry) (*Program, []Diagnostic) {
	return analyzeSource(path, src, reg, false)
}

func checkAndExpand(file *ast.File, reg *stdlib.Registry, path string, parseErrs []parser.Error) (*checker.Info, map[*ast.Function]bool, []checker.Diagnostic) {
	info, checkDiags := checker.CheckWithStdlibForPath(file, reg, path)
	if len(parseErrs) > 0 || len(checkDiags) > 0 {
		return info, nil, checkDiags
	}
	changed, macroDiags := macro.Expand(file, info)
	checkDiags = append(checkDiags, macroDiags...)
	if len(macroDiags) > 0 {
		return info, nil, checkDiags
	}
	if changed {
		info, checkDiags = checker.CheckWithStdlibForPath(file, reg, path)
		if len(checkDiags) > 0 {
			return info, nil, checkDiags
		}
	}
	constChanged, compileTimeFunctions, constDiags := expandCompileTimeExprs(file, info)
	checkDiags = append(checkDiags, constDiags...)
	if len(constDiags) > 0 || !constChanged {
		return info, compileTimeFunctions, checkDiags
	}
	info, checkDiags = checker.CheckWithStdlibForPath(file, reg, path)
	return info, compileTimeFunctions, checkDiags
}

func analyzedProgram(path string, src string, file *ast.File, info *checker.Info, compileTimeFunctions map[*ast.Function]bool, parseErrs []parser.Error, checkDiags []checker.Diagnostic, diags []Diagnostic, includeWarnings bool) (*Program, []Diagnostic) {
	for _, err := range parseErrs {
		diags = append(diags, Diagnostic{Message: err.Message, Pos: err.Pos, Path: path})
	}
	for _, diag := range checkDiags {
		diags = append(diags, Diagnostic{Message: diag.Message, Pos: diag.Pos, Path: path, Severity: diag.Severity, Code: diag.Code, Kind: diag.Kind})
	}
	if includeWarnings && len(parseErrs) == 0 && !hasErrorDiagnostics(checkDiags) {
		for _, diag := range checker.Lint(file, info) {
			diags = append(diags, Diagnostic{Message: diag.Message, Pos: diag.Pos, Path: path, Severity: diag.Severity, Code: diag.Code, Kind: diag.Kind})
		}
	}
	lowered := ir.LowerFile(file, info)
	pruneCompileTimeOnlyFunctions(file, info, lowered, compileTimeFunctions)
	return &Program{
		Path:   path,
		Source: src,
		File:   file,
		Info:   info,
		IR:     lowered,
		Macros: macro.Plan(file, info),
	}, diags
}

func hasErrorDiagnostics(diags []checker.Diagnostic) bool {
	for _, diag := range diags {
		if diag.Severity != checker.SeverityWarning {
			return true
		}
	}
	return false
}

func pruneCompileTimeOnlyFunctions(file *ast.File, info *checker.Info, lowered *ir.File, compileTimeFunctions map[*ast.Function]bool) {
	if file == nil || info == nil || lowered == nil || len(compileTimeFunctions) == 0 {
		return
	}
	runtimeFunctions := map[*ast.Function]bool{}
	markRuntimeFunction := func(fn *ast.Function) {}
	markRuntimeExpr := func(expr ast.Expr) {}
	markRuntimeFunction = func(fn *ast.Function) {
		if fn == nil || runtimeFunctions[fn] {
			return
		}
		runtimeFunctions[fn] = true
		markRuntimeExpr(fn.Body)
	}
	markRuntimeExpr = func(expr ast.Expr) {
		ast.WalkExpr(expr, func(expr ast.Expr) {
			switch e := expr.(type) {
			case *ast.Identifier:
				if fn := info.ResolvedFunctions[e]; fn != nil {
					markRuntimeFunction(fn.Node)
				}
			case *ast.SelectorExpr:
				if fn := info.ResolvedSelectorFunctions[e]; fn != nil {
					markRuntimeFunction(fn.Node)
				}
			}
		})
	}
	for _, fn := range file.Functions {
		if !compileTimeFunctions[fn] {
			markRuntimeFunction(fn)
		}
	}
	for _, typ := range file.Types {
		for _, method := range typ.Methods {
			if !compileTimeFunctions[method] {
				markRuntimeFunction(method)
			}
		}
	}
	for _, test := range file.Tests {
		markRuntimeExpr(test.Body)
	}
	lowered.Functions = pruneLoweredFunctions(file.Functions, lowered.Functions, compileTimeFunctions, runtimeFunctions)
	for idx := range file.Types {
		if idx >= len(lowered.Types) {
			break
		}
		lowered.Types[idx].Methods = pruneLoweredFunctions(file.Types[idx].Methods, lowered.Types[idx].Methods, compileTimeFunctions, runtimeFunctions)
	}
}

func pruneLoweredFunctions(astFunctions []*ast.Function, loweredFunctions []*ir.Function, compileTimeFunctions map[*ast.Function]bool, runtimeFunctions map[*ast.Function]bool) []*ir.Function {
	out := make([]*ir.Function, 0, len(loweredFunctions))
	loweredIdx := 0
	for _, fn := range astFunctions {
		if fn.Macro {
			continue
		}
		if loweredIdx >= len(loweredFunctions) {
			break
		}
		lowered := loweredFunctions[loweredIdx]
		loweredIdx++
		if compileTimeFunctions[fn] && !runtimeFunctions[fn] && fn.Name != "main" {
			continue
		}
		out = append(out, lowered)
	}
	if loweredIdx < len(loweredFunctions) {
		out = append(out, loweredFunctions[loweredIdx:]...)
	}
	return out
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
	for _, imp := range fileImportRefs(file) {
		importPath, err := resolveRuneImportRef(normalized, imp)
		if err != nil {
			diags = append(diags, Diagnostic{Message: err.Error(), Pos: imp.pos, Path: normalized})
			continue
		}
		if imp.expr != nil {
			imp.expr.SourcePath = importPath
		}
		if sameImportPath(importPath, normalized) {
			continue
		}
		if filepath.Ext(importPath) == ".ts" {
			tsImport, importedDiags := l.loadTypeScript(importPath, imp.path, imp.pos)
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

func sameImportPath(left string, right string) bool {
	leftAbs, leftErr := filepath.Abs(filepath.Clean(left))
	rightAbs, rightErr := filepath.Abs(filepath.Clean(right))
	if leftErr == nil && rightErr == nil {
		return leftAbs == rightAbs
	}
	return filepath.Clean(left) == filepath.Clean(right)
}

type importRef struct {
	path   string
	module bool
	pos    lexer.Position
	expr   *ast.AtExpr
}

func fileImportRefs(file *ast.File) []importRef {
	if file == nil {
		return nil
	}
	refs := make([]importRef, 0, len(file.Imports))
	for _, imp := range file.Imports {
		refs = append(refs, importRef{path: imp.Path, module: imp.Module, pos: imp.Pos})
	}
	for _, expr := range importExpressionRefs(file) {
		refs = append(refs, importRef{path: expr.Path, pos: expr.Pos, expr: expr})
	}
	return refs
}

func resolveRuneImportRef(fromPath string, imp importRef) (string, error) {
	if imp.module {
		return ResolveRuneModuleImport(fromPath, imp.path)
	}
	return ResolveRuneImport(fromPath, imp.path)
}

func importExpressionRefs(file *ast.File) []*ast.AtExpr {
	var refs []*ast.AtExpr
	visit := func(expr ast.Expr) {
		if at, ok := expr.(*ast.AtExpr); ok && at.Path != "" {
			refs = append(refs, at)
		}
	}
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
	return refs
}

func (l *importLoader) loadTypeScript(path string, specifier string, pos lexer.Position) (*ast.TSImport, []Diagnostic) {
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
	imp.Specifier = specifier
	l.tsFiles[normalized] = true
	return &imp, diags
}

func mergeFile(dst *ast.File, src *ast.File, includeTests bool) {
	if src == nil {
		return
	}
	dst.Imports = append(dst.Imports, src.Imports...)
	dst.GoImports = append(dst.GoImports, src.GoImports...)
	dst.TSImports = append(dst.TSImports, src.TSImports...)
	dst.Traits = append(dst.Traits, src.Traits...)
	dst.Types = append(dst.Types, src.Types...)
	dst.Enums = append(dst.Enums, src.Enums...)
	dst.Constants = append(dst.Constants, src.Constants...)
	dst.Functions = append(dst.Functions, src.Functions...)
	if includeTests {
		dst.Tests = append(dst.Tests, src.Tests...)
	}
}

func annotateSourcePath(file *ast.File, sourcePath string) {
	if file == nil {
		return
	}
	for _, trait := range file.Traits {
		trait.SourcePath = sourcePath
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
	for _, constant := range file.Constants {
		constant.SourcePath = sourcePath
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

func ResolveRuneModuleImport(fromPath string, module string) (string, error) {
	if module == "" {
		return "", fmt.Errorf("empty import path")
	}
	spec := filepath.Join(module, module+".rn")
	if path, err := ResolveRuneImport(fromPath, spec); err == nil {
		return path, nil
	}
	if filepath.Base(filepath.Dir(fromPath)) == strings.TrimSuffix(filepath.Base(fromPath), filepath.Ext(fromPath)) {
		if path, err := ResolveRuneImport(fromPath, filepath.Join("..", spec)); err == nil {
			return path, nil
		}
	}
	if path, ok := resolveCoreModuleImport(fromPath, spec); ok {
		return path, nil
	}
	return ResolveRuneImport(fromPath, spec)
}

func resolveCoreModuleImport(fromPath string, spec string) (string, bool) {
	for _, start := range []string{filepath.Dir(fromPath), "."} {
		dir, err := filepath.Abs(start)
		if err != nil {
			continue
		}
		for {
			candidate := filepath.Join(dir, "core", spec)
			if _, err := os.Stat(candidate); err == nil {
				if abs, err := filepath.Abs(candidate); err == nil {
					return abs, true
				}
				return candidate, true
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return "", false
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

func GenerateMoonBitFile(path string) (string, []Diagnostic) {
	prog, diags := AnalyzeFile(path)
	if len(diags) > 0 {
		return "", diags
	}
	src, err := moonbitcodegen.GenerateIR(prog.IR)
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
