package format

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
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
	case *ast.TemplateLiteral:
		return f.templateLiteral(e)
	case *ast.CharLiteral:
		return strconv.QuoteRune(e.Value)
	case *ast.RegexLiteral:
		return e.Raw
	case *ast.BoolLiteral:
		if e.Value {
			return "true"
		}
		return "false"
	case *ast.NullLiteral:
		return "null"
	case *ast.UnaryExpr:
		return e.Op.String() + f.unaryOperandExpr(e.Expr)
	case *ast.PostfixExpr:
		return f.postfixReceiverExpr(e.Expr) + e.Op.String()
	case *ast.ResultUnwrapExpr:
		return f.postfixReceiverExpr(e.Expr) + "?"
	case *ast.BinaryExpr:
		if f.isPatternPredicateBitOr(e) {
			return strings.Join(f.patternPredicateBitOrParts(e), " | ")
		}
		if e.Op == lexer.DotDotEqual {
			return fmt.Sprintf("%s..=%s", f.exprWithParens(e.Left), f.exprWithParens(e.Right))
		}
		return fmt.Sprintf("%s %s %s", f.exprWithParens(e.Left), e.Op, f.exprWithParens(e.Right))
	case *ast.TernaryExpr:
		return f.ternaryExpr(e)
	case *ast.AssignExpr:
		target := e.Name
		if e.Target != nil {
			target = f.expr(e.Target)
		}
		return fmt.Sprintf("%s = %s", target, f.expr(e.Value))
	case *ast.CallExpr:
		if formatted, ok := f.chainCallExpr(e); ok {
			return formatted
		}
		return f.postfixReceiverExpr(e.Callee) + f.formatCallArgs(e.Args)
	case *ast.LambdaExpr:
		return f.lambdaExpr(e)
	case *ast.IndexExpr:
		return fmt.Sprintf("%s[%s]", f.postfixReceiverExpr(e.Receiver), f.expr(e.Index))
	case *ast.SelectorExpr:
		if _, ok := e.Receiver.(*ast.ThisExpr); ok {
			return "." + e.Name
		}
		return f.postfixReceiverExpr(e.Receiver) + "." + e.Name
	case *ast.ArrayLiteral:
		elems := make([]string, 0, len(e.Elements))
		for _, elem := range e.Elements {
			elems = append(elems, f.expr(elem))
		}
		return "[" + strings.Join(elems, ", ") + "]"
	case *ast.TupleLiteral:
		elems := make([]string, 0, len(e.Elements))
		for _, elem := range e.Elements {
			elems = append(elems, f.expr(elem))
		}
		return "(" + strings.Join(elems, ", ") + ")"
	case *ast.SpreadExpr:
		return "..." + f.expr(e.Expr)
	case *ast.ReactiveLiteral:
		return "$" + f.expr(e.Value)
	case *ast.MapLiteral:
		return f.mapLiteral(e)
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

func (f *formatter) templateLiteral(lit *ast.TemplateLiteral) string {
	var b strings.Builder
	b.WriteByte('`')
	for _, part := range lit.Parts {
		if part.Text != "" {
			b.WriteString(escapeTemplateText(part.Text))
		}
		if part.Expr != nil {
			b.WriteString("${")
			b.WriteString(f.expr(part.Expr))
			b.WriteByte('}')
		}
	}
	b.WriteByte('`')
	return b.String()
}

func escapeTemplateText(text string) string {
	var b strings.Builder
	for i, ch := range text {
		switch ch {
		case '\\':
			b.WriteString(`\\`)
		case '`':
			b.WriteString("\\`")
		case '\r':
			b.WriteString(`\r`)
		case '$':
			if i+1 < len(text) && text[i+1] == '{' {
				b.WriteString(`\$`)
			} else {
				b.WriteRune(ch)
			}
		default:
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func (f *formatter) chainCallExpr(call *ast.CallExpr) (string, bool) {
	parts := f.callChainParts(call)
	if len(parts) < 2 {
		return "", false
	}
	return strings.Join(parts, ""), true
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
	return f.postfixReceiverExpr(expr)
}

func (f *formatter) postfixReceiverExpr(expr ast.Expr) string {
	formatted := f.expr(expr)
	switch expr.(type) {
	case *ast.AssignExpr, *ast.BinaryExpr, *ast.LambdaExpr, *ast.WatchExpr:
		return "(" + formatted + ")"
	default:
		return formatted
	}
}

func (f *formatter) unaryOperandExpr(expr ast.Expr) string {
	formatted := f.expr(expr)
	switch expr.(type) {
	case *ast.AssignExpr, *ast.BinaryExpr, *ast.WatchExpr:
		return "(" + formatted + ")"
	default:
		return formatted
	}
}

func (f *formatter) xmlElement(elem *ast.XMLElement) string {
	return f.xmlElementWithIndent(elem, f.indent, false)
}

func (f *formatter) xmlElementWithIndent(elem *ast.XMLElement, indent int, leadingIndent bool) string {
	if f.xmlElementCanInline(elem) {
		prefix := ""
		if leadingIndent {
			prefix = f.indentString(indent)
		}
		return prefix + f.xmlElementInline(elem)
	}

	var b strings.Builder
	if leadingIndent {
		b.WriteString(f.indentString(indent))
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
			b.WriteString(f.indentString(indent + 1))
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
	b.WriteString(f.indentString(indent))
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
		return f.indentString(indent) + "{" + lines[0] + "}"
	}
	var b strings.Builder
	b.WriteString(f.indentString(indent))
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
			f.indentString(f.indent) + ")"
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
	bodyIndent := f.indentString(f.indent + 1)
	closeIndent := f.indentString(f.indent)
	b.WriteString("{\n")
	previous := f.indent
	f.indent++
	for i, stmt := range block.Statements {
		formatted := f.stmt(stmt)
		for j, line := range strings.Split(formatted, "\n") {
			if j == 0 {
				b.WriteString(bodyIndent)
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
	fieldIndent := f.indentString(f.indent + 1)
	closeIndent := f.indentString(f.indent)
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

func (f *formatter) mapLiteral(lit *ast.MapLiteral) string {
	if len(lit.Entries) == 0 {
		return "{}"
	}
	var b strings.Builder
	entryIndent := f.indentString(f.indent + 1)
	closeIndent := f.indentString(f.indent)
	b.WriteString("{\n")
	for _, entry := range lit.Entries {
		b.WriteString(entryIndent)
		b.WriteString(f.mapKeyExpr(entry.Key))
		b.WriteString(": ")
		b.WriteString(f.exprWithIndent(entry.Value, f.indent+1))
		b.WriteString(",\n")
	}
	b.WriteString(closeIndent)
	b.WriteString("}")
	return b.String()
}

func (f *formatter) mapKeyExpr(expr ast.Expr) string {
	formatted := f.expr(expr)
	if mapKeyNeedsParens(expr) {
		return "(" + formatted + ")"
	}
	return formatted
}

func mapKeyNeedsParens(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Identifier, *ast.SelectorExpr, *ast.CallExpr, *ast.IndexExpr, *ast.PostfixExpr:
		return true
	case *ast.BinaryExpr:
		return mapKeyNeedsParens(e.Left)
	case *ast.TernaryExpr:
		return mapKeyNeedsParens(e.Condition)
	default:
		return false
	}
}

func (f *formatter) matchExpr(match *ast.MatchExpr) string {
	var b strings.Builder
	b.WriteString(f.expr(match.Subject))
	b.WriteString(" {\n")
	branchIndent := f.indentString(f.indent + 1)
	closeIndent := f.indentString(f.indent)
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
	fieldIndent := f.indentString(f.indent + 1)
	closeIndent := f.indentString(f.indent)
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
			b.WriteString(privatePrefix(field.Private))
			b.WriteString(f.anonymousObjectMethod(field.Name, lambda))
		} else {
			b.WriteString(privatePrefix(field.Private))
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
	ret := formatReturnType(lambda.ReturnType, lambda.ReturnDisplay)
	return fmt.Sprintf("%s(%s)%s => %s", name, strings.Join(params, ", "), ret, f.exprWithIndent(lambda.Body, f.indent+1))
}

func (f *formatter) exprWithIndent(expr ast.Expr, indent int) string {
	previous := f.indent
	f.indent = indent
	out := f.expr(expr)
	f.indent = previous
	return out
}

func (f *formatter) ternaryExpr(expr *ast.TernaryExpr) string {
	condition := f.exprWithIndent(expr.Condition, f.indent+1)
	consequence := f.exprWithIndent(expr.Consequence, f.indent+1)

	var b strings.Builder
	b.WriteString("(\n")
	b.WriteString(f.indentString(f.indent + 1))
	b.WriteString(appendToLastLine(condition, " ? "+consequence))
	b.WriteByte('\n')
	if expr.Alternative != nil {
		alternative := f.exprWithIndent(expr.Alternative, f.indent+2)
		b.WriteString(f.indentString(f.indent + 2))
		b.WriteString(": ")
		b.WriteString(alternative)
		b.WriteByte('\n')
	}
	b.WriteString(f.indentString(f.indent))
	b.WriteString(")")
	return b.String()
}

func (f *formatter) exprWithParens(expr ast.Expr) string {
	switch expr.(type) {
	case *ast.BinaryExpr:
		return "(" + f.expr(expr) + ")"
	}
	return f.expr(expr)
}

func appendToLastLine(text string, suffix string) string {
	if !strings.HasSuffix(text, "\n") {
		return text + suffix
	}
	return strings.TrimSuffix(text, "\n") + suffix + "\n"
}

func (f *formatter) isPatternPredicateBitOr(expr *ast.BinaryExpr) bool {
	if expr == nil || expr.Op != lexer.BitOr {
		return false
	}
	leaves := f.patternPredicateBitOrLeaves(expr)
	if len(leaves) < 2 {
		return false
	}
	allBitwise := true
	for _, leaf := range leaves {
		if !patternPredicateLeaf(leaf) {
			return false
		}
		if !bitwiseLiteralLeaf(leaf) {
			allBitwise = false
		}
	}
	return !allBitwise
}

func (f *formatter) patternPredicateBitOrParts(expr ast.Expr) []string {
	leaves := f.patternPredicateBitOrLeaves(expr)
	parts := make([]string, 0, len(leaves))
	for _, leaf := range leaves {
		part := f.expr(leaf)
		if patternPredicateRangeLeaf(leaf) {
			part = "(" + part + ")"
		}
		parts = append(parts, part)
	}
	return parts
}

func (f *formatter) patternPredicateBitOrLeaves(expr ast.Expr) []ast.Expr {
	if binary, ok := expr.(*ast.BinaryExpr); ok && binary.Op == lexer.BitOr {
		out := f.patternPredicateBitOrLeaves(binary.Left)
		out = append(out, f.patternPredicateBitOrLeaves(binary.Right)...)
		return out
	}
	return []ast.Expr{expr}
}

func patternPredicateLeaf(expr ast.Expr) bool {
	if patternPredicateRangeLeaf(expr) {
		return true
	}
	switch expr.(type) {
	case *ast.BoolLiteral, *ast.IntegerLiteral, *ast.BigIntLiteral, *ast.StringLiteral, *ast.CharLiteral, *ast.NullLiteral, *ast.SelectorExpr:
		return true
	default:
		return false
	}
}

func patternPredicateRangeLeaf(expr ast.Expr) bool {
	binary, ok := expr.(*ast.BinaryExpr)
	return ok && binary.Op == lexer.DotDotEqual
}

func bitwiseLiteralLeaf(expr ast.Expr) bool {
	switch expr.(type) {
	case *ast.IntegerLiteral, *ast.BigIntLiteral:
		return true
	default:
		return false
	}
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
	case *ast.TernaryExpr:
		return true
	default:
		return false
	}
}
