package parser

import (
	"testing"

	"github.com/oboard/rune-lang/internal/ast"
)

func TestParseOptionalSelectorInXMLExpression(t *testing.T) {
	file, errs := Parse(`main() => <input @change={(e) => e.target?.value ?? ""} />`)
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	xml := file.Functions[0].Body.(*ast.XMLElement)
	lambda := xml.Attrs[0].Value.(*ast.LambdaExpr)
	binary := lambda.Body.(*ast.BinaryExpr)
	selector, ok := binary.Left.(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("left = %T, want *ast.SelectorExpr", binary.Left)
	}
	if !selector.Optional || selector.Name != "value" {
		t.Fatalf("selector = %#v, want optional .value", selector)
	}
}
