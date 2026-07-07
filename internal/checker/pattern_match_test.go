package checker

import (
	"strings"
	"testing"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/parser"
)

func TestArrayAndAliasPatternBindings(t *testing.T) {
	src := `score(values: Array[Int]) -> Int => values {
  [head, ..rest, tail] @ whole => head + tail + rest.length() + whole.length()
  _ => 0
}
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
	match := file.Functions[0].Body.(*ast.MatchExpr)
	asPattern := match.Branches[0].Pattern.(*ast.AsPattern)
	if got, want := asPattern.Type, "Array[Int]"; got != want {
		t.Fatalf("alias binding type = %q, want %q", got, want)
	}
	arrayPattern := asPattern.Pattern.(*ast.ArrayPattern)
	if got, want := arrayPattern.RestType, "Array[Int]"; got != want {
		t.Fatalf("rest type = %q, want %q", got, want)
	}
	if got, want := info.Functions["score"].Return, Int; got != want {
		t.Fatalf("return = %s, want %s", got, want)
	}
}

func TestStringArrayPatternBindings(t *testing.T) {
	src := `middle(text: String) -> String => text {
  [first, ..rest, last] => rest + first.toString() + last.toString()
  _ => ""
}
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	_, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
	match := file.Functions[0].Body.(*ast.MatchExpr)
	arrayPattern := match.Branches[0].Pattern.(*ast.ArrayPattern)
	if got, want := arrayPattern.RestType, "String"; got != want {
		t.Fatalf("rest type = %q, want %q", got, want)
	}
	for _, elem := range arrayPattern.Elements {
		binding, ok := elem.(*ast.BindingPattern)
		if !ok || binding.Type != "Char" {
			t.Fatalf("element pattern = %#v, want Char binding", elem)
		}
	}
}

func TestPatternPredicateExpressionInfersEnumMembers(t *testing.T) {
	src := `TokenKind: {
  Ident = 0
  Int = 1
  RParen = 2
}

canEndValueToken(kind: TokenKind) -> Bool => kind ~ (Ident | Int | RParen)
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
	if got := info.Functions["canEndValueToken"].Return; got != Bool {
		t.Fatalf("return = %s, want Bool", got)
	}
	match := file.Functions[0].Body.(*ast.MatchExpr)
	orPattern := match.Branches[0].Pattern.(*ast.OrPattern)
	for _, alternative := range orPattern.Alternatives {
		binding, ok := alternative.(*ast.BindingPattern)
		if !ok || !binding.Constant || binding.Type != "TokenKind" {
			t.Fatalf("alternative = %#v, want inferred enum member binding", alternative)
		}
	}
}

func TestConstructorArgPatternsAndOrBindings(t *testing.T) {
	src := `Expr: {
  Lit(value: Int)
  Add(left: Expr, right: Expr)
  Mul(left: Expr, right: Expr)
}

eval(expr: Expr) -> Int => expr {
  Add(Lit(left), Lit(0)) => left
  Add(left, right) | Mul(left, right) => 1
  _ => 0
}
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	info, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
	match := file.Functions[0].Body.(*ast.MatchExpr)
	leftExpr := match.Branches[0].Expr.(*ast.Identifier)
	if got, want := info.ExprTypes[leftExpr], Int; got != want {
		t.Fatalf("nested constructor binding type = %s, want %s", got, want)
	}
}

func TestOrPatternRequiresSameBindings(t *testing.T) {
	src := `Expr: {
  Lit(value: Int)
  Add(left: Expr, right: Expr)
}

bad(expr: Expr) -> Int => expr {
  Add(left, right) | Lit(left) => 1
  _ => 0
}
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	_, diags := Check(file)
	if len(diags) == 0 {
		t.Fatalf("Check() succeeded, want or-pattern binding diagnostic")
	}
}

func TestMapPatternAllowsConstKeys(t *testing.T) {
	src := `const KeyA = "a"

score(values: Map[String, Int]) -> Int => values {
  { KeyA: value, .. } => value
  _ => 0
}
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	_, diags := Check(file)
	if len(diags) > 0 {
		t.Fatalf("check diagnostics: %v", diags)
	}
}

func TestMapPatternRejectsNonConstKeys(t *testing.T) {
	src := `score(values: Map[String, Int], KeyA: String) -> Int => values {
  { KeyA: value, .. } => value
  _ => 0
}
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	_, diags := Check(file)
	for _, diag := range diags {
		if strings.Contains(diag.Message, "map pattern key must be a literal or const") {
			return
		}
	}
	t.Fatalf("diagnostics = %#v, want map pattern key diagnostic", diags)
}

func TestRangePatternRejectsNonConstBounds(t *testing.T) {
	src := `score(value: Int, Limit: Int) -> Int => value {
  0..<Limit => 1
  _ => 0
}
`
	file, parseErrs := parser.Parse(src)
	if len(parseErrs) > 0 {
		t.Fatalf("parse errors: %v", parseErrs)
	}
	_, diags := Check(file)
	for _, diag := range diags {
		if strings.Contains(diag.Message, "range pattern bound must be a literal, const, or '_'") {
			return
		}
	}
	t.Fatalf("diagnostics = %#v, want range pattern bound diagnostic", diags)
}
