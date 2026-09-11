package checker

import "github.com/oboard/rune-lang/internal/ast"

func (c *checker) inferXMLElement(elem *ast.XMLElement, env map[string]Type) Type {
	if fn, ok := c.resolveFunction(elem.Tag, elem.Pos); ok {
		c.info.XMLResolvedFunctions[elem] = fn
	}
	// XML attributes are unconstrained — clear expectedType so anonymous
	// object literals aren't coerced into the surrounding HTMLElement shape
	// (they're style/aria dictionaries, not the element). Event attributes are
	// the exception: their value must be a `(Event) => ...` handler, so we seed
	// that expected shape to type the callback parameter. The event interface
	// is resolved from lib.dom.d.ts's GlobalEventHandlersEventMap chain rather
	// than a hand-maintained table.
	prev := c.expectedType
	c.expectedType = Unknown
	for _, attr := range elem.Attrs {
		if attr.Value == nil {
			continue
		}
		if attr.Event {
			// Seed the handler's parameter with the DOM event interface before
			// inference so `(e) => e.key` resolves `e` to KeyboardEvent (etc.)
			// instead of a structurally-inferred anonymous field type.
			c.applyExpectedType(attr.Value, FuncOfTypes([]Type{c.domEventType(attr.Name)}, Void))
			c.inferExpr(attr.Value, env)
			continue
		}
		c.inferExpr(attr.Value, env)
	}
	for _, child := range elem.Children {
		if child.Expr != nil {
			c.inferExpr(child.Expr, env)
		}
	}
	c.expectedType = prev
	return HTMLElement
}