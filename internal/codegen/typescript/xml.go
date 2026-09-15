package tscodegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
)

type xmlEmitter struct {
	g      *generator
	b      strings.Builder
	indent int
}

func (g *generator) xmlExpr(elem *ir.XMLElement) string {
	emitter := &xmlEmitter{g: g}
	emitter.line("(() => {")
	emitter.indent++
	root := emitter.element(elem)
	emitter.linef("return %s;", root)
	emitter.indent--
	emitter.line("})()")
	return strings.TrimRight(emitter.b.String(), "\n")
}

func (e *xmlEmitter) line(s string) {
	for i := 0; i < e.indent; i++ {
		e.b.WriteString("  ")
	}
	e.b.WriteString(s)
	e.b.WriteByte('\n')
}

func (e *xmlEmitter) linef(format string, args ...any) {
	e.line(fmt.Sprintf(format, args...))
}

func (e *xmlEmitter) lineExpr(prefix string, expr string, suffix string) {
	lines := strings.Split(expr, "\n")
	if len(lines) == 0 {
		e.line(prefix + suffix)
		return
	}
	if len(lines) == 1 {
		e.line(prefix + lines[0] + suffix)
		return
	}
	e.line(prefix + lines[0])
	for _, line := range lines[1 : len(lines)-1] {
		e.line(line)
	}
	e.line(lines[len(lines)-1] + suffix)
}

func (e *xmlEmitter) element(elem *ir.XMLElement) string {
	name := e.g.nextTemp("__el")
	if component := e.g.lookupWebComponentFunction(elem.Tag); component != nil {
		factory := e.webComponentFactory(component, elem)
		e.linef("const %s = document.createElement(runeDefineWebComponent(%s, %s));", name, strconv.Quote(elem.Tag), factory)
		for _, attr := range elem.Attrs {
			e.attr(name, elem.Tag, attr)
		}
	} else if component := e.g.lookupHTMLElementComponentFunction(elem.Tag); component != nil {
		e.linef("const %s = %s;", name, e.componentCall(component, elem))
		for _, attr := range elem.Attrs {
			if attr.Event || !functionHasParam(component, attr.Name) {
				e.attr(name, elem.Tag, attr)
			}
		}
	} else {
		e.linef("const %s = document.createElement(%s);", name, strconv.Quote(elem.Tag))
		for _, attr := range elem.Attrs {
			e.attr(name, elem.Tag, attr)
		}
	}
	for _, child := range elem.Children {
		e.child(name, child)
	}
	return name
}

func (e *xmlEmitter) webComponentFactory(fn *ir.Function, elem *ir.XMLElement) string {
	name := FunctionSymbolName(fn)
	if len(fn.Params) == 0 {
		return name
	}
	return fmt.Sprintf("() => %s", e.componentCall(fn, elem))
}

func (e *xmlEmitter) componentCall(fn *ir.Function, elem *ir.XMLElement) string {
	name := FunctionSymbolName(fn)
	if len(fn.Params) == 0 {
		return name + "()"
	}
	attrByName := map[string]ir.XMLAttr{}
	for _, attr := range elem.Attrs {
		if !attr.Event {
			attrByName[attr.Name] = attr
		}
	}
	args := make([]string, 0, len(fn.Params))
	for _, param := range fn.Params {
		if attr, ok := attrByName[param.Name]; ok && attr.Value != nil {
			args = append(args, e.g.expr(attr.Value))
			continue
		}
		args = append(args, e.g.zeroValue(param.Type))
	}
	return fmt.Sprintf("%s(%s)", name, strings.Join(args, ", "))
}

func functionHasParam(fn *ir.Function, name string) bool {
	for _, param := range fn.Params {
		if param.Name == name {
			return true
		}
	}
	return false
}

// eventListener keeps the listener parameter compatible with DOM's
// addEventListener overload while invoking the Rune handler through a narrowed
// event view. A callback that directly accepts a narrower target is rejected by
// TypeScript's contravariant function-parameter checking.
func (e *xmlEmitter) eventListener(lambda *ir.LambdaExpr, tag, eventName string) string {
	if len(lambda.Params) == 0 {
		return e.g.expr(lambda)
	}
	narrowedType := e.eventViewType(tag, eventName)
	handler := e.eventLambda(lambda, narrowedType)
	return "(e) => { (" + handler + ")(e as " + narrowedType + "); }"
}

// eventLambda renders a Rune event lambda with its first argument narrowed to
// the tag's lib.dom.d.ts element type.
func (e *xmlEmitter) preventingEventListener(listener string) string {
	return "(e) => { e.preventDefault(); (" + listener + ")(e); }"
}

func (e *xmlEmitter) eventLambda(lambda *ir.LambdaExpr, eventType string) string {
	params, ret, ok := parseTSFuncType(string(lambda.ResultType()))
	if !ok {
		params = make([]string, len(lambda.Params))
		ret = string(checker.Void)
	}
	tsParams := make([]string, 0, len(lambda.Params))
	for i, name := range lambda.Params {
		typ := checker.Unknown
		if i < len(params) {
			typ = checker.Type(params[i])
		}
		if i == 0 {
			tsParams = append(tsParams, fmt.Sprintf("%s: %s", mangleIdent(name), eventType))
			continue
		}
		tsParams = append(tsParams, fmt.Sprintf("%s: %s", mangleIdent(name), tsType(typ)))
	}
	if block, ok := lambda.Body.(*ir.BlockExpr); ok {
		return "(" + strings.Join(tsParams, ", ") + "): " + tsType(checker.Type(ret)) + " => { " + e.g.blockInline(block, checker.Type(ret)) + " }"
	}
	if checker.Type(ret) == checker.Void {
		return "(" + strings.Join(tsParams, ", ") + ") => { " + e.g.stmtExpr(lambda.Body) + "; }"
	}
	return "(" + strings.Join(tsParams, ", ") + "): " + tsType(checker.Type(ret)) + " => " + e.g.arrowExprBody(lambda.Body)
}

func (e *xmlEmitter) eventViewType(tag, eventName string) string {
	eventType := tsType(checker.DOMEventType(eventName))
	targetType := tsType(checker.DOMTargetType(tag))
	if targetType == "EventTarget" {
		return eventType
	}
	return eventType + " & { target: " + targetType + "; currentTarget: " + targetType + " }"
}

func (e *xmlEmitter) attrValue(attr ir.XMLAttr) string {
	if attr.Name == "style" && attr.Value != nil {
		if _, ok := attr.Value.(*ir.AnonymousObjectLiteral); ok {
			return fmt.Sprintf("runeStyleText(%s)", e.g.expr(attr.Value))
		}
	}
	return e.g.expr(attr.Value)
}

func (e *xmlEmitter) attr(elemName, tag string, attr ir.XMLAttr) {
	if attr.Spread && attr.Value != nil {
		o := e.g.expr(attr.Value)
		// Apply spread properties as attributes
		e.linef("for (const [key, value] of Object.entries(%s)) {", o)
		e.linef("  %s.setAttribute(key, String(value));", elemName)
		e.linef("}")
		return
	}
	if attr.Event {
		if attr.Value == nil {
			return
		}
		eventName := checker.DOMEventBaseName(attr.Name)
		prevent := checker.DOMEventHasModifier(attr.Name, "prevent")
		if lambda, ok := attr.Value.(*ir.LambdaExpr); ok {
			listener := e.eventListener(lambda, tag, attr.Name)
			if prevent {
				listener = e.preventingEventListener(listener)
			}
			e.linef("%s.addEventListener(%s, %s);", elemName, strconv.Quote(eventName), listener)
			return
		}
		handler := e.g.stmtExpr(attr.Value)
		prefix := ""
		if prevent {
			prefix = "e.preventDefault(); "
		}
		if _, ok := attr.Value.(*ir.Identifier); ok {
			e.linef("%s.addEventListener(%s, (e) => { %s%s(e); });", elemName, strconv.Quote(eventName), prefix, handler)
			return
		}
		e.linef("%s.addEventListener(%s, (e) => { %s%s; });", elemName, strconv.Quote(eventName), prefix, handler)
		return
	}
	if attr.Value == nil {
		e.linef("%s.setAttribute(%s, \"\");", elemName, strconv.Quote(attr.Name))
		return
	}
	value := e.attrValue(attr)
	if booleanDOMProperty(tag, attr.Name) {
		e.linef("%s.%s = Boolean(%s);", elemName, attr.Name, value)
		for _, dep := range e.g.exprSignalDeps(attr.Value) {
			e.linef("runeWatch(%s, () => { %s.%s = Boolean(%s); });", mangleIdent(dep), elemName, attr.Name, e.attrValue(attr))
		}
		return
	}
	if stringDOMProperty(tag, attr.Name) {
		e.linef("%s.%s = String(%s);", elemName, attr.Name, value)
		for _, dep := range e.g.exprSignalDeps(attr.Value) {
			e.linef("runeWatch(%s, () => { %s.%s = String(%s); });", mangleIdent(dep), elemName, attr.Name, e.attrValue(attr))
		}
		return
	}
	e.linef("%s.setAttribute(%s, String(%s));", elemName, strconv.Quote(attr.Name), value)
	for _, dep := range e.g.exprSignalDeps(attr.Value) {
		e.linef("runeWatch(%s, () => { %s.setAttribute(%s, String(%s)); });", mangleIdent(dep), elemName, strconv.Quote(attr.Name), e.attrValue(attr))
	}
}

func booleanDOMProperty(tag, name string) bool {
	switch name {
	case "checked", "disabled", "selected", "multiple", "readOnly", "required":
		return true
	default:
		return false
	}
}

func stringDOMProperty(tag, name string) bool {
	return name == "value" && (tag == "input" || tag == "textarea" || tag == "select")
}

func (g *generator) lookupWebComponentFunction(tag string) *ir.Function {
	return g.lookupComponentFunction(tag, checker.WebComponent)
}

func (g *generator) lookupHTMLElementComponentFunction(tag string) *ir.Function {
	return g.lookupComponentFunction(tag, checker.HTMLElement)
}

func (g *generator) lookupComponentFunction(tag string, returnType checker.Type) *ir.Function {
	if g.file == nil {
		return nil
	}
	for _, fn := range g.file.Functions {
		if (fn.SourceName == tag || fn.Name == tag) && fn.Return == returnType {
			return fn
		}
	}
	return nil
}

func (e *xmlEmitter) child(parent string, child ir.XMLChild) {
	if child.Text != "" {
		e.linef("%s.appendChild(document.createTextNode(%s));", parent, strconv.Quote(child.Text))
		return
	}
	if child.Expr == nil {
		return
	}
	if elem, ok := child.Expr.(*ir.XMLElement); ok {
		childName := e.element(elem)
		e.linef("%s.appendChild(%s);", parent, childName)
		return
	}
	if child.Expr.ResultType() == checker.HTMLElement {
		e.linef("%s.appendChild(%s);", parent, e.g.expr(child.Expr))
		return
	}
	if elem, ok := checker.ArrayElement(child.Expr.ResultType()); ok && elem == checker.HTMLElement {
		deps := e.g.exprSignalDeps(child.Expr)
		if len(deps) > 0 {
			e.reactiveElementArray(parent, child.Expr, deps)
			return
		}
		children := e.g.nextTemp("__children")
		childName := e.g.nextTemp("__child")
		e.lineExpr("const "+children+" = ", e.g.expr(child.Expr), ";")
		e.linef("for (const %s of %s) {", childName, children)
		e.indent++
		e.linef("%s.appendChild(%s);", parent, childName)
		e.indent--
		e.line("}")
		return
	}
	textName := e.g.nextTemp("__text")
	e.linef("const %s = document.createTextNode(String(%s));", textName, e.g.expr(child.Expr))
	for _, dep := range e.g.exprSignalDeps(child.Expr) {
		e.linef("runeWatch(%s, () => { %s.data = String(%s); });", mangleIdent(dep), textName, e.g.expr(child.Expr))
	}
	e.linef("%s.appendChild(%s);", parent, textName)
}

func (e *xmlEmitter) reactiveElementArray(parent string, expr ir.Expr, deps []string) {
	start := e.g.nextTemp("__start")
	end := e.g.nextTemp("__end")
	render := e.g.nextTemp("__render")
	node := e.g.nextTemp("__node")
	next := e.g.nextTemp("__next")
	children := e.g.nextTemp("__children")
	child := e.g.nextTemp("__child")
	e.linef("const %s = document.createComment(\"rune\");", start)
	e.linef("const %s = document.createComment(\"rune\");", end)
	e.linef("%s.appendChild(%s);", parent, start)
	e.linef("%s.appendChild(%s);", parent, end)
	e.linef("const %s = () => {", render)
	e.indent++
	e.linef("let %s = %s.nextSibling;", node, start)
	e.linef("while (%s && %s !== %s) {", node, node, end)
	e.indent++
	e.linef("const %s = %s.nextSibling;", next, node)
	e.linef("%s.remove();", node)
	e.linef("%s = %s;", node, next)
	e.indent--
	e.line("}")
	e.lineExpr("const "+children+" = ", e.g.expr(expr), ";")
	e.linef("for (const %s of %s) {", child, children)
	e.indent++
	e.linef("%s.insertBefore(%s, %s);", parent, child, end)
	e.indent--
	e.line("}")
	e.indent--
	e.line("};")
	e.linef("%s();", render)
	for _, dep := range deps {
		e.linef("runeWatch(%s, %s);", mangleIdent(dep), render)
	}
}
