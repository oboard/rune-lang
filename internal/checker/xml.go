package checker

import (
	"github.com/oboard/rune-lang/internal/ast"
)

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
			eventType := c.domEventType(attr.Name)
			c.applyExpectedType(attr.Value, FuncOfTypes([]Type{eventType}, Void))
			// Narrow e.target / e.currentTarget to the handler element for
			// form-control tags (React ChangeEvent<T> parity).
			c.applyDOMEventTargetNarrowing(attr.Value, elem.Tag, eventType)
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

// applyDOMEventTargetNarrowing walks a handler lambda and registers per-selector
// overrides so `e.target`, `e.currentTarget`, etc. resolve to the concrete element
// interface for form-control tags. The fields to narrow are derived from the
// event interface's members whose declared type matches the un-narrowed target
// type (the fallback from domTargetType) rather than hardcoding field names.
func (c *checker) applyDOMEventTargetNarrowing(handler ast.Expr, tag string, eventType Type) {
	lambda, ok := handler.(*ast.LambdaExpr)
	if !ok || len(lambda.Params) == 0 {
		return
	}
	paramName := lambda.Params[0]
	targetType := c.domTargetType(tag)
	// Determine the un-narrowed (vanilla) target type from the Event interface's
	// `target` field rather than hardcoding "EventTarget".
	eventInfo := c.info.Types[string(eventType)]
	if eventInfo == nil {
		return
	}
	vanillaTargetField, hasVanillaTarget := eventInfo.ByName["target"]
	vanillaTargetType := Type("EventTarget") // fallback if target field not found
	if hasVanillaTarget && vanillaTargetField.Type != Unknown {
		vanillaTargetType = vanillaTargetField.Type
	}
	// Only narrow if the concrete element type differs from vanilla.
	if targetType == vanillaTargetType || targetType == Unknown {
		return
	}
	// Find which fields of the event interface are typed as the vanilla target
	// type — those are the ones to narrow per-handler.
	narrowFields := map[string]bool{}
	for name, field := range eventInfo.ByName {
		if field.Type == vanillaTargetType {
			narrowFields[name] = true
		}
	}
	if len(narrowFields) == 0 {
		return
	}
	ast.WalkExpr(lambda.Body, func(expr ast.Expr) {
		sel, ok := expr.(*ast.SelectorExpr)
		if !ok || sel.Static {
			return
		}
		if !narrowFields[sel.Name] {
			return
		}
		ident, ok := sel.Receiver.(*ast.Identifier)
		if !ok || ident.Name != paramName {
			return
		}
		c.domTargetOverrides[sel] = targetType
	})
}