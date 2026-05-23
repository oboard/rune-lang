package format

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
)

const maxLineLength = 40

func (f *formatter) expr(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Identifier:
		return e.Name
	case *ast.AtExpr:
		return "@" + e.Name
	case *ast.IntegerLiteral:
		return strconv.Itoa(e.Value)
	case *ast.DoubleLiteral:
		if e.Raw != "" {
			return e.Raw
		}
		return strconv.FormatFloat(e.Value, 'f', -1, 64)
	case *ast.BigIntLiteral:
		return e.Value + "n"
	case *ast.StringLiteral:
		return strconv.Quote(e.Value)
	case *ast.BoolLiteral:
		if e.Value {
			return "true"
		}
		return "false"
	case *ast.NullLiteral:
		return "null"
	case *ast.UnaryExpr:
		return e.Op.String() + f.expr(e.Expr)
	case *ast.PostfixExpr:
		return f.expr(e.Expr) + e.Op.String()
	case *ast.BinaryExpr:
		return fmt.Sprintf("%s %s %s", f.exprWithParens(e.Left), e.Op, f.exprWithParens(e.Right))
	case *ast.AssignExpr:
		return fmt.Sprintf("%s = %s", e.Name, f.expr(e.Value))
	case *ast.CallExpr:
		if formatted, ok := f.chainCallExpr(e); ok {
			return formatted
		}
		return f.expr(e.Callee) + f.formatCallArgs(e.Args)
	case *ast.LambdaExpr:
		return f.lambdaExpr(e)
	case *ast.IndexExpr:
		return fmt.Sprintf("%s[%s]", f.expr(e.Receiver), f.expr(e.Index))
	case *ast.SelectorExpr:
		if _, ok := e.Receiver.(*ast.ThisExpr); ok {
			return "." + e.Name
		}
		return f.expr(e.Receiver) + "." + e.Name
	case *ast.ArrayLiteral:
		elems := make([]string, 0, len(e.Elements))
		for _, elem := range e.Elements {
			elems = append(elems, f.expr(elem))
		}
		return "[" + strings.Join(elems, ", ") + "]"
	case *ast.SpreadExpr:
		return "..." + f.expr(e.Expr)
	case *ast.ReactiveLiteral:
		return "$" + f.expr(e.Value)
	case *ast.StructLiteral:
		return f.structLiteral(e)
	case *ast.AnonymousObjectLiteral:
		return f.anonymousObjectLiteral(e)
	case *ast.XMLElement:
		return f.xmlElement(e)
	case *ast.BlockExpr:
		return f.blockExpr(e)
	case *ast.PatternBlock:
		return "{ ... }"
	case *ast.MatchExpr:
		return f.matchExpr(e)
	case *ast.WatchExpr:
		return f.watchExpr(e)
	case *ast.ThisExpr:
		return "this"
	default:
		return ""
	}
}

func (f *formatter) chainCallExpr(call *ast.CallExpr) (string, bool) {
	parts := f.callChainParts(call)
	if len(parts) < 2 {
		return "", false
	}
	flat := strings.Join(parts, "")
	if len(flat) <= maxLineLength || len(parts) < 3 || strings.Contains(flat, "\n") {
		return flat, true
	}
	continuationIndent := indentString(f.indent) + strings.Repeat(" ", len(parts[0])+2)
	var b strings.Builder
	b.WriteString(parts[0])
	b.WriteString(parts[1])
	for _, part := range parts[2:] {
		b.WriteByte('\n')
		b.WriteString(continuationIndent)
		b.WriteString(part)
	}
	return b.String(), true
}

func (f *formatter) callChainParts(expr ast.Expr) []string {
	if call, ok := expr.(*ast.CallExpr); ok {
		if sel, ok := call.Callee.(*ast.SelectorExpr); ok {
			parts := f.callChainParts(sel.Receiver)
			if len(parts) == 0 {
				parts = []string{f.chainReceiverExpr(sel.Receiver)}
			}
			parts = append(parts, "."+sel.Name+f.formatCallArgs(call.Args))
			return parts
		}
	}
	return nil
}

func (f *formatter) formatCallArgs(args []ast.Expr) string {
	formatted := make([]string, 0, len(args))
	for _, arg := range args {
		formatted = append(formatted, f.expr(arg))
	}
	return "(" + strings.Join(formatted, ", ") + ")"
}

func (f *formatter) chainReceiverExpr(expr ast.Expr) string {
	if _, ok := expr.(*ast.ThisExpr); ok {
		return ""
	}
	return f.expr(expr)
}

func (f *formatter) xmlElement(elem *ast.XMLElement) string {
	return f.xmlElementWithIndent(elem, f.indent, false)
}

func (f *formatter) xmlElementWithIndent(elem *ast.XMLElement, indent int, leadingIndent bool) string {
	if f.xmlElementCanInline(elem) {
		prefix := ""
		if leadingIndent {
			prefix = indentString(indent)
		}
		return prefix + f.xmlElementInline(elem)
	}

	var b strings.Builder
	if leadingIndent {
		b.WriteString(indentString(indent))
	}
	b.WriteByte('<')
	b.WriteString(elem.Tag)
	for _, attr := range elem.Attrs {
		b.WriteByte(' ')
		b.WriteString(f.xmlAttr(attr))
	}
	b.WriteByte('>')
	b.WriteByte('\n')
	for _, child := range elem.Children {
		if child.Text != "" {
			b.WriteString(indentString(indent + 1))
			b.WriteString(child.Text)
			b.WriteByte('\n')
			continue
		}
		if child.Expr == nil {
			continue
		}
		if nested, ok := child.Expr.(*ast.XMLElement); ok {
			b.WriteString(f.xmlElementWithIndent(nested, indent+1, true))
			b.WriteByte('\n')
			continue
		}
		b.WriteString(f.xmlChildExpr(child.Expr, indent+1))
		b.WriteByte('\n')
	}
	b.WriteString(indentString(indent))
	b.WriteString("</")
	b.WriteString(elem.Tag)
	b.WriteByte('>')
	return b.String()
}

func (f *formatter) xmlElementCanInline(elem *ast.XMLElement) bool {
	if len(elem.Children) == 0 {
		return true
	}
	for _, child := range elem.Children {
		if child.Expr == nil {
			continue
		}
		if _, ok := child.Expr.(*ast.XMLElement); ok {
			return false
		}
		if f.exprNeedsMultiline(child.Expr) {
			return false
		}
	}
	return true
}

func (f *formatter) xmlElementInline(elem *ast.XMLElement) string {
	var b strings.Builder
	b.WriteByte('<')
	b.WriteString(elem.Tag)
	for _, attr := range elem.Attrs {
		b.WriteByte(' ')
		b.WriteString(f.xmlAttr(attr))
	}
	if len(elem.Children) == 0 {
		b.WriteString(" />")
		return b.String()
	}
	b.WriteByte('>')
	for _, child := range elem.Children {
		if child.Text != "" {
			b.WriteString(child.Text)
			continue
		}
		if child.Expr == nil {
			continue
		}
		if nested, ok := child.Expr.(*ast.XMLElement); ok {
			b.WriteString(f.xmlElementInline(nested))
			continue
		}
		b.WriteByte('{')
		b.WriteString(f.expr(child.Expr))
		b.WriteByte('}')
	}
	b.WriteString("</")
	b.WriteString(elem.Tag)
	b.WriteByte('>')
	return b.String()
}

func (f *formatter) xmlAttr(attr ast.XMLAttr) string {
	var b strings.Builder
	if attr.Event {
		b.WriteByte('@')
	}
	b.WriteString(attr.Name)
	if attr.Value != nil {
		b.WriteString("={")
		b.WriteString(f.expr(attr.Value))
		b.WriteByte('}')
	}
	return b.String()
}

func (f *formatter) xmlChildExpr(expr ast.Expr, indent int) string {
	previous := f.indent
	f.indent = indent
	formatted := f.expr(expr)
	f.indent = previous
	lines := strings.Split(formatted, "\n")
	if len(lines) == 1 {
		return indentString(indent) + "{" + lines[0] + "}"
	}
	var b strings.Builder
	b.WriteString(indentString(indent))
	b.WriteByte('{')
	b.WriteString(lines[0])
	b.WriteByte('\n')
	for _, line := range lines[1 : len(lines)-1] {
		b.WriteString(line)
		b.WriteByte('\n')
	}
	b.WriteString(lines[len(lines)-1])
	b.WriteByte('}')
	return b.String()
}

func (f *formatter) lambdaExpr(lambda *ast.LambdaExpr) string {
	params := f.lambdaParams(lambda)
	if lambda.Implicit {
		return f.expr(lambda.Body)
	}
	if xml, ok := lambda.Body.(*ast.XMLElement); ok {
		return "(" + strings.Join(params, ", ") + ") => (\n" +
			f.xmlElementWithIndent(xml, f.indent+2, true) + "\n" +
			indentString(f.indent) + ")"
	}
	return "(" + strings.Join(params, ", ") + ") => " + f.expr(lambda.Body)
}

func (f *formatter) lambdaParams(lambda *ast.LambdaExpr) []string {
	params := make([]string, 0, len(lambda.Params))
	for i, param := range lambda.Params {
		if i < len(lambda.ParamTypes) && lambda.ParamTypes[i] != "" {
			params = append(params, fmt.Sprintf("%s: %s", param, f.lambdaParamType(lambda, i)))
		} else {
			params = append(params, param)
		}
	}
	return params
}

func (f *formatter) lambdaParamType(lambda *ast.LambdaExpr, index int) string {
	display := ""
	if index < len(lambda.ParamTypeDisplays) {
		display = lambda.ParamTypeDisplays[index]
	}
	return formatType(lambda.ParamTypes[index], display)
}

func (f *formatter) watchExpr(watch *ast.WatchExpr) string {
	lambda, ok := watch.Handler.(*ast.LambdaExpr)
	if !ok {
		return fmt.Sprintf("%s -> %s", f.expr(watch.Target), f.expr(watch.Handler))
	}
	if lambda.Implicit {
		return fmt.Sprintf("%s -> %s", f.expr(watch.Target), f.expr(lambda.Body))
	}
	return fmt.Sprintf("%s -> (%s) => %s", f.expr(watch.Target), strings.Join(f.lambdaParams(lambda), ", "), f.expr(lambda.Body))
}

func (f *formatter) blockExpr(block *ast.BlockExpr) string {
	var b strings.Builder
	bodyIndent := indentString(f.indent + 1)
	closeIndent := indentString(f.indent)
	b.WriteString("{\n")
	previous := f.indent
	f.indent++
	for i, stmt := range block.Statements {
		formatted := f.stmt(stmt)
		for j, line := range strings.Split(formatted, "\n") {
			if j == 0 {
				b.WriteString(bodyIndent)
			} else if line != "" {
				b.WriteString(indentString(f.indent))
			}
			b.WriteString(line)
			b.WriteByte('\n')
		}
		if i < len(block.Statements)-1 && separatesFollowingStatement(stmt, block.Statements[i+1], formatted) {
			b.WriteByte('\n')
		}
	}
	f.indent = previous
	b.WriteString(closeIndent)
	b.WriteString("}")
	return b.String()
}

func (f *formatter) structLiteral(lit *ast.StructLiteral) string {
	if len(lit.Fields) == 0 {
		return lit.TypeName + " {}"
	}
	var b strings.Builder
	fieldIndent := indentString(f.indent + 1)
	closeIndent := indentString(f.indent)
	b.WriteString(lit.TypeName)
	b.WriteString(" {\n")
	for _, field := range lit.Fields {
		b.WriteString(fieldIndent)
		b.WriteString(field.Name)
		b.WriteString(": ")
		b.WriteString(f.exprWithIndent(field.Value, f.indent+1))
		b.WriteByte('\n')
	}
	b.WriteString(closeIndent)
	b.WriteString("}")
	return b.String()
}

func (f *formatter) matchExpr(match *ast.MatchExpr) string {
	var b strings.Builder
	b.WriteString(f.expr(match.Subject))
	b.WriteString(" {\n")
	branchIndent := indentString(f.indent + 1)
	closeIndent := indentString(f.indent)
	for _, branch := range match.Branches {
		b.WriteString(branchIndent)
		b.WriteString(f.pattern(branch.Pattern))
		b.WriteString(" => ")
		b.WriteString(f.exprWithIndent(branch.Expr, f.indent+1))
		b.WriteByte('\n')
	}
	b.WriteString(closeIndent)
	b.WriteString("}")
	return b.String()
}

func (f *formatter) anonymousObjectLiteral(obj *ast.AnonymousObjectLiteral) string {
	if len(obj.Fields) == 0 {
		return "{}"
	}
	var b strings.Builder
	fieldIndent := indentString(f.indent + 1)
	closeIndent := indentString(f.indent)
	b.WriteString("{\n")
	seenMethod := false
	for i, field := range obj.Fields {
		lambda, isMethod := field.Value.(*ast.LambdaExpr)
		if isMethod && !seenMethod && i > 0 {
			b.WriteByte('\n')
		}
		seenMethod = seenMethod || isMethod
		b.WriteString(fieldIndent)
		if isMethod {
			b.WriteString(f.anonymousObjectMethod(field.Name, lambda))
		} else {
			b.WriteString(field.Name)
			b.WriteString(": ")
			b.WriteString(f.exprWithIndent(field.Value, f.indent+1))
			b.WriteString(",")
		}
		b.WriteByte('\n')
	}
	b.WriteString(closeIndent)
	b.WriteString("}")
	return b.String()
}

func (f *formatter) anonymousObjectMethod(name string, lambda *ast.LambdaExpr) string {
	params := make([]string, 0, len(lambda.Params))
	for i, param := range lambda.Params {
		if i < len(lambda.ParamTypes) && lambda.ParamTypes[i] != "" {
			params = append(params, fmt.Sprintf("%s: %s", param, f.lambdaParamType(lambda, i)))
		} else {
			params = append(params, param)
		}
	}
	return fmt.Sprintf("%s(%s) => %s", name, strings.Join(params, ", "), f.exprWithIndent(lambda.Body, f.indent+1))
}

func (f *formatter) exprWithIndent(expr ast.Expr, indent int) string {
	previous := f.indent
	f.indent = indent
	out := f.expr(expr)
	f.indent = previous
	return out
}

func (f *formatter) exprWithParens(expr ast.Expr) string {
	if _, ok := expr.(*ast.BinaryExpr); ok {
		return "(" + f.expr(expr) + ")"
	}
	return f.expr(expr)
}

func (f *formatter) exprNeedsMultiline(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.XMLElement:
		return !f.xmlElementCanInline(e)
	case *ast.LambdaExpr:
		_, ok := e.Body.(*ast.XMLElement)
		return ok || f.exprNeedsMultiline(e.Body)
	case *ast.CallExpr:
		if f.exprNeedsMultiline(e.Callee) {
			return true
		}
		for _, arg := range e.Args {
			if f.exprNeedsMultiline(arg) {
				return true
			}
		}
		return false
	default:
		return false
	}
}
