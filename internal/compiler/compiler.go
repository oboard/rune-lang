package compiler

import (
	"fmt"
	"os"
	"path/filepath"

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
		return nil, []Diagnostic{{Message: err.Error()}}
	}
	return AnalyzeSource(path, string(data))
}

func AnalyzeSource(path string, src string) (*Program, []Diagnostic) {
	file, parseErrs := parser.Parse(src)
	info, checkDiags := checker.Check(file)
	return analyzedProgram(path, src, file, info, parseErrs, checkDiags)
}

func AnalyzeSourceWithStdlib(path string, src string, reg *stdlib.Registry) (*Program, []Diagnostic) {
	file, parseErrs := parser.Parse(src)
	info, checkDiags := checker.CheckWithStdlib(file, reg)
	return analyzedProgram(path, src, file, info, parseErrs, checkDiags)
}

func analyzedProgram(path string, src string, file *ast.File, info *checker.Info, parseErrs []parser.Error, checkDiags []checker.Diagnostic) (*Program, []Diagnostic) {
	var diags []Diagnostic
	for _, err := range parseErrs {
		diags = append(diags, Diagnostic{Message: err.Message, Pos: err.Pos})
	}
	for _, diag := range checkDiags {
		diags = append(diags, Diagnostic{Message: diag.Message, Pos: diag.Pos})
	}
	return &Program{Path: path, Source: src, File: file, Info: info, IR: ir.LowerFile(file, info)}, diags
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
