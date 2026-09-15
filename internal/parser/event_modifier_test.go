package parser

import (
	"testing"

	"github.com/oboard/rune-lang/internal/ast"
)

func TestParseEventModifierAsFullName(t *testing.T) {
	file, errs := Parse(`main() => <form @submit.prevent={addTodo} />`)
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	elem := file.Functions[0].Body.(*ast.XMLElement)
	attr := elem.Attrs[0]
	if !attr.Event || attr.Name != "submit.prevent" {
		t.Fatalf("attribute = %#v, want event name submit.prevent", attr)
	}
}
