package compiler

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	gocodegen "github.com/oboard/rune-lang/internal/codegen/go"
	"github.com/oboard/rune-lang/internal/ir"
	"github.com/oboard/rune-lang/internal/lexer"
	"github.com/oboard/rune-lang/internal/parser"
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
	var diags []Diagnostic
	for _, err := range parseErrs {
		diags = append(diags, Diagnostic{Message: err.Message, Pos: err.Pos})
	}
	info, checkDiags := checker.Check(file)
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

func GeneratedPath(inputPath string) string {
	base := filepath.Base(inputPath)
	ext := filepath.Ext(base)
	if ext != "" {
		base = base[:len(base)-len(ext)]
	}
	return fmt.Sprintf("%s.go", base)
}
