package checker

import (
	"testing"

	"github.com/oboard/rune-lang/internal/ast"
)

func TestEventModifierInfersReferencedHandlerParam(t *testing.T) {
	file := parseSource(t, `
main() => {
  addTodo := (e) => e.type
  <form @submit.prevent={addTodo} />
}
`)
	info, diags := Check(file)
	if len(diags) != 0 {
		t.Fatalf("diagnostics: %#v", diags)
	}
	block := file.Functions[0].Body.(*ast.BlockExpr)
	lambda := block.Statements[0].(*ast.LetStmt).Value.(*ast.LambdaExpr)
	ident := lambda.Body.(*ast.SelectorExpr).Receiver.(*ast.Identifier)
	if got := info.ExprTypes[ident]; got != Type("SubmitEvent") {
		t.Fatalf("addTodo event type = %s, want SubmitEvent", got)
	}
}
