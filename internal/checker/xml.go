package checker

import "github.com/oboard/rune-lang/internal/ast"

func (c *checker) inferXMLElement(elem *ast.XMLElement, env map[string]Type) Type {
	if fn, ok := c.resolveFunction(elem.Tag, elem.Pos); ok {
		c.info.XMLResolvedFunctions[elem] = fn
	}
	for _, attr := range elem.Attrs {
		if attr.Value != nil {
			c.inferExpr(attr.Value, env)
		}
	}
	for _, child := range elem.Children {
		if child.Expr != nil {
			c.inferExpr(child.Expr, env)
		}
	}
	return HTMLElement
}
