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

func (e *xmlEmitter) element(elem *ir.XMLElement) string {
	name := e.g.nextTemp("__el")
	e.linef("const %s = document.createElement(%s);", name, strconv.Quote(elem.Tag))
	for _, attr := range elem.Attrs {
		e.attr(name, attr)
	}
	for _, child := range elem.Children {
		e.child(name, child)
	}
	return name
}

func (e *xmlEmitter) attr(elemName string, attr ir.XMLAttr) {
	if attr.Event {
		if attr.Value == nil {
			return
		}
		if _, ok := attr.Value.(*ir.LambdaExpr); ok {
			e.linef("%s.addEventListener(%s, %s);", elemName, strconv.Quote(attr.Name), e.g.expr(attr.Value))
			return
		}
		e.linef("%s.addEventListener(%s, () => { %s; });", elemName, strconv.Quote(attr.Name), e.g.stmtExpr(attr.Value))
		return
	}
	if attr.Value == nil {
		e.linef("%s.setAttribute(%s, \"\");", elemName, strconv.Quote(attr.Name))
		return
	}
	value := e.g.expr(attr.Value)
	e.linef("%s.setAttribute(%s, String(%s));", elemName, strconv.Quote(attr.Name), value)
	for _, dep := range e.g.exprSignalDeps(attr.Value) {
		e.linef("%s.watch(() => { %s.setAttribute(%s, String(%s)); });", mangleIdent(dep), elemName, strconv.Quote(attr.Name), e.g.expr(attr.Value))
	}
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
	textName := e.g.nextTemp("__text")
	e.linef("const %s = document.createTextNode(String(%s));", textName, e.g.expr(child.Expr))
	for _, dep := range e.g.exprSignalDeps(child.Expr) {
		e.linef("%s.watch(() => { %s.data = String(%s); });", mangleIdent(dep), textName, e.g.expr(child.Expr))
	}
	e.linef("%s.appendChild(%s);", parent, textName)
}
