package parser

import (
	"testing"

	"github.com/oboard/rune-lang/internal/ast"
)

func TestGenericTypeDeclRequiresColon(t *testing.T) {
	file, errs := Parse(`Array[T] {
    length[T]() -> Int => "%array.len"
}
`)
	if len(errs) == 0 {
		t.Fatalf("Parse() accepted generic type declaration without ':'")
	}
	if len(file.Types) != 0 {
		t.Fatalf("Parse() produced %d types for invalid declaration", len(file.Types))
	}
}

func TestGenericTypeDeclWithColon(t *testing.T) {
	file, errs := Parse(`Array[T]: {
    length[T]() -> Int => "%array.len"
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if len(file.Types) != 1 {
		t.Fatalf("Parse() produced %d types, want 1", len(file.Types))
	}
	if got := file.Types[0].Generics; len(got) != 1 || got[0] != "T" {
		t.Fatalf("type generics = %v, want [T]", got)
	}
}

func TestFunctionDeclRequiresFatArrow(t *testing.T) {
	file, errs := Parse(`main() {
}
`)
	if len(errs) == 0 {
		t.Fatalf("Parse() accepted function declaration without '=>'")
	}
	if len(file.Functions) != 0 {
		t.Fatalf("Parse() produced %d functions for invalid declaration", len(file.Functions))
	}
}

func TestLambdaRequiresParenthesizedParams(t *testing.T) {
	_, errs := Parse(`main() => {
    values.map(value => value + 1)
}
`)
	if len(errs) == 0 {
		t.Fatalf("Parse() accepted lambda without parenthesized params")
	}
}

func TestLambdaWithParenthesizedParams(t *testing.T) {
	_, errs := Parse(`main() => {
    values.map((value) => value + 1)
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
}

func TestLambdaWithAnnotatedParams(t *testing.T) {
	file, errs := Parse(`Input: {
    a: Int
}

main() => {
    f := (value: Input) => value.a + 1
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	block, ok := file.Functions[0].Body.(*ast.BlockExpr)
	if !ok || len(block.Statements) != 1 {
		t.Fatalf("main body = %#v, want one statement", file.Functions[0].Body)
	}
	let, ok := block.Statements[0].(*ast.LetStmt)
	if !ok {
		t.Fatalf("statement = %T, want LetStmt", block.Statements[0])
	}
	lambda, ok := let.Value.(*ast.LambdaExpr)
	if !ok {
		t.Fatalf("let value = %T, want LambdaExpr", let.Value)
	}
	if len(lambda.ParamTypes) != 1 || lambda.ParamTypes[0] != "Input" {
		t.Fatalf("lambda param types = %v, want [Input]", lambda.ParamTypes)
	}
}

func TestBlockStartingWithCallIsNotObjectLiteral(t *testing.T) {
	file, errs := Parse(`nestedMatch() => {
}

main() => {
    nestedMatch()
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if len(file.Functions) != 2 {
		t.Fatalf("functions = %d, want 2", len(file.Functions))
	}
	block, ok := file.Functions[1].Body.(*ast.BlockExpr)
	if !ok || len(block.Statements) != 1 {
		t.Fatalf("main body = %#v, want one block statement", file.Functions[1].Body)
	}
	stmt, ok := block.Statements[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("statement = %T, want ExprStmt", block.Statements[0])
	}
	if _, ok := stmt.Expr.(*ast.CallExpr); !ok {
		t.Fatalf("expression = %T, want CallExpr", stmt.Expr)
	}
}

func TestObjectMethodDependsOnObjectContext(t *testing.T) {
	file, errs := Parse(`main() => {
    obj := {
        greet() => @io.println("hi")
    }
    obj.greet()
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	block, ok := file.Functions[0].Body.(*ast.BlockExpr)
	if !ok || len(block.Statements) != 2 {
		t.Fatalf("main body = %#v, want two block statements", file.Functions[0].Body)
	}
	let, ok := block.Statements[0].(*ast.LetStmt)
	if !ok {
		t.Fatalf("first statement = %T, want LetStmt", block.Statements[0])
	}
	obj, ok := let.Value.(*ast.AnonymousObjectLiteral)
	if !ok || len(obj.Fields) != 1 {
		t.Fatalf("let value = %#v, want object with one method", let.Value)
	}
	if _, ok := obj.Fields[0].Value.(*ast.LambdaExpr); !ok {
		t.Fatalf("object field = %T, want LambdaExpr method", obj.Fields[0].Value)
	}
}

func TestAnonymousObjectMethodMembers(t *testing.T) {
	file, errs := Parse(`main() => {
    obj := {
        name: "Alice"
        nextAge() => .age + 1
    }
}
`)
	if len(errs) > 0 {
		t.Fatalf("Parse() errors = %v", errs)
	}
	if len(file.Functions) != 1 {
		t.Fatalf("Parse() functions = %d, want 1", len(file.Functions))
	}
	block, ok := file.Functions[0].Body.(*ast.BlockExpr)
	if !ok || len(block.Statements) != 1 {
		t.Fatalf("main body = %#v, want one block statement", file.Functions[0].Body)
	}
	let, ok := block.Statements[0].(*ast.LetStmt)
	if !ok {
		t.Fatalf("statement = %T, want LetStmt", block.Statements[0])
	}
	obj, ok := let.Value.(*ast.AnonymousObjectLiteral)
	if !ok {
		t.Fatalf("let value = %T, want AnonymousObjectLiteral", let.Value)
	}
	if len(obj.Fields) != 2 {
		t.Fatalf("object fields = %d, want 2", len(obj.Fields))
	}
	method, ok := obj.Fields[1].Value.(*ast.LambdaExpr)
	if !ok {
		t.Fatalf("method field = %T, want LambdaExpr", obj.Fields[1].Value)
	}
	if obj.Fields[1].Name != "nextAge" || len(method.Params) != 0 {
		t.Fatalf("method field = %s params=%v, want nextAge()", obj.Fields[1].Name, method.Params)
	}
}
