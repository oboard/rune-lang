package checker

import (
	"testing"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/parser"
)

func TestPatternPredicateInfersParamAndReturn(t *testing.T) {
	src := `isSpace(ch) => ' ' | '\t' | '\r'
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
	fn := info.Functions["isSpace"]
	if fn == nil {
		t.Fatalf("isSpace info missing")
	}
	if got := fn.Params[0].Type; got != Char {
		t.Fatalf("isSpace param type = %s, want Char", got)
	}
	if got := fn.Return; got != Bool {
		t.Fatalf("isSpace return = %s, want Bool", got)
	}
	if _, ok := file.Functions[0].Body.(*ast.PatternBlock); !ok {
		t.Fatalf("body = %T, want PatternBlock", file.Functions[0].Body)
	}
}

func TestPatternPredicateRangeInfersParamAndReturn(t *testing.T) {
	src := `isDigit(ch) => ('0'..='9')
isAlpha(ch: Char) -> Bool => ('a'..='z') | ('A'..='Z')
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
	digit := info.Functions["isDigit"]
	if digit == nil {
		t.Fatalf("isDigit info missing")
	}
	if got := digit.Params[0].Type; got != Char {
		t.Fatalf("isDigit param type = %s, want Char", got)
	}
	if got := digit.Return; got != Bool {
		t.Fatalf("isDigit return = %s, want Bool", got)
	}
	block, ok := file.Functions[0].Body.(*ast.PatternBlock)
	if !ok || len(block.Branches) != 2 {
		t.Fatalf("isDigit body = %T, want range PatternBlock", file.Functions[0].Body)
	}
	if _, ok := block.Branches[0].Pattern.(*ast.RangePattern); !ok {
		t.Fatalf("isDigit first pattern = %T, want RangePattern", block.Branches[0].Pattern)
	}
	alpha := info.Functions["isAlpha"]
	if alpha == nil {
		t.Fatalf("isAlpha info missing")
	}
	if got := alpha.Return; got != Bool {
		t.Fatalf("isAlpha return = %s, want Bool", got)
	}
	alphaBlock, ok := file.Functions[1].Body.(*ast.PatternBlock)
	if !ok || len(alphaBlock.Branches) != 3 {
		t.Fatalf("isAlpha body = %T, want two ranges plus wildcard", file.Functions[1].Body)
	}
}

func TestPatternPredicateDoesNotStealValidBitwise(t *testing.T) {
	src := `mask() => 1 | 2
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
	if got := info.Functions["mask"].Return; got != Int {
		t.Fatalf("mask return = %s, want Int", got)
	}
	if _, ok := file.Functions[0].Body.(*ast.BinaryExpr); !ok {
		t.Fatalf("body = %T, want BinaryExpr", file.Functions[0].Body)
	}
}
