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
