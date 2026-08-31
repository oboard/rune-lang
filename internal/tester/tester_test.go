package tester

import (
	"testing"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
)

func TestTypeScriptTestRuntimeSourceUsesSelfhostForEligibleMainOnly(t *testing.T) {
	cleanup := registerSelfhostCompilers(
		func(files []SourceFile) CompileResult {
			return CompileResult{Ok: true, Output: "// selfhost-ts"}
		},
		nil,
	)
	defer cleanup()

	src, err := typeScriptTestRuntimeSource(testEligibleMainFile())
	if err != nil {
		t.Fatalf("typeScriptTestRuntimeSource() error = %v", err)
	}
	if src != "// selfhost-ts" {
		t.Fatalf("typeScriptTestRuntimeSource() = %q, want selfhost output", src)
	}
}

func TestTypeScriptTestRuntimeSourceFallsBackWhenIneligible(t *testing.T) {
	cleanup := registerSelfhostCompilers(
		func(files []SourceFile) CompileResult {
			return CompileResult{Ok: true, Output: "// selfhost-ts"}
		},
		nil,
	)
	defer cleanup()

	src, err := typeScriptTestRuntimeSource(testIneligibleFunctionsFile())
	if err != nil {
		t.Fatalf("typeScriptTestRuntimeSource() error = %v", err)
	}
	if src == "// selfhost-ts" {
		t.Fatal("typeScriptTestRuntimeSource() used selfhost output for ineligible file")
	}
}

func TestMoonBitTestRuntimeSourceUsesSelfhostForEligibleMainOnly(t *testing.T) {
	cleanup := registerSelfhostCompilers(
		nil,
		func(files []SourceFile) CompileResult {
			return CompileResult{Ok: true, Output: "// selfhost-mbt"}
		},
	)
	defer cleanup()

	src, err := moonBitTestRuntimeSource(testEligibleMainFile())
	if err != nil {
		t.Fatalf("moonBitTestRuntimeSource() error = %v", err)
	}
	if src != "// selfhost-mbt" {
		t.Fatalf("moonBitTestRuntimeSource() = %q, want selfhost output", src)
	}
}

func TestMoonBitTestRuntimeSourceFallsBackWhenIneligible(t *testing.T) {
	cleanup := registerSelfhostCompilers(
		nil,
		func(files []SourceFile) CompileResult {
			return CompileResult{Ok: true, Output: "// selfhost-mbt"}
		},
	)
	defer cleanup()

	src, err := moonBitTestRuntimeSource(testIneligibleFunctionsFile())
	if err != nil {
		t.Fatalf("moonBitTestRuntimeSource() error = %v", err)
	}
	if src == "// selfhost-mbt" {
		t.Fatal("moonBitTestRuntimeSource() used selfhost output for ineligible file")
	}
}

func TestTypeScriptTestRuntimeSourceUsesSelfhostForDuplicateMainWrapper(t *testing.T) {
	cleanup := registerSelfhostCompilers(
		func(files []SourceFile) CompileResult {
			return CompileResult{Ok: true, Output: "// selfhost-dup-main-ts"}
		},
		nil,
	)
	defer cleanup()

	src, err := typeScriptTestRuntimeSource(testDuplicateMainFile())
	if err != nil {
		t.Fatalf("typeScriptTestRuntimeSource() error = %v", err)
	}
	if src != "// selfhost-dup-main-ts" {
		t.Fatalf("typeScriptTestRuntimeSource() = %q, want duplicate-main selfhost output", src)
	}
}

func TestMoonBitTestRuntimeSourceUsesSelfhostForDuplicateMainWrapper(t *testing.T) {
	cleanup := registerSelfhostCompilers(
		nil,
		func(files []SourceFile) CompileResult {
			return CompileResult{Ok: true, Output: "// selfhost-dup-main-mbt"}
		},
	)
	defer cleanup()

	src, err := moonBitTestRuntimeSource(testDuplicateMainFile())
	if err != nil {
		t.Fatalf("moonBitTestRuntimeSource() error = %v", err)
	}
	if src != "// selfhost-dup-main-mbt" {
		t.Fatalf("moonBitTestRuntimeSource() = %q, want duplicate-main selfhost output", src)
	}
}

func testEligibleMainFile() *ir.File {
	return &ir.File{
		Functions: []*ir.Function{{
			Name:       "main",
			SourceName: "main",
			Return:     checker.Void,
			Body:       &ir.BlockExpr{ExprBase: ir.ExprBase{Type: checker.Void}},
		}},
	}
}

func testIneligibleFunctionsFile() *ir.File {
	return &ir.File{
		Functions: []*ir.Function{
			{
				Name:       "helper",
				SourceName: "helper",
				Return:     checker.Int,
				Body:       &ir.IntegerLiteral{ExprBase: ir.ExprBase{Type: checker.Int}, Value: 1},
			},
			{
				Name:       "main",
				SourceName: "main",
				Return:     checker.Void,
				Body:       &ir.BlockExpr{ExprBase: ir.ExprBase{Type: checker.Void}},
			},
		},
	}
}

func testDuplicateMainFile() *ir.File {
	return &ir.File{
		Functions: []*ir.Function{
			{
				Name:       "main",
				SourceName: "main",
				Return:     checker.Int,
				Body:       &ir.IntegerLiteral{ExprBase: ir.ExprBase{Type: checker.Int}, Value: 1},
			},
			{
				Name:       "main",
				SourceName: "main",
				Return:     checker.Void,
				Body:       &ir.BlockExpr{ExprBase: ir.ExprBase{Type: checker.Void}},
			},
		},
	}
}
