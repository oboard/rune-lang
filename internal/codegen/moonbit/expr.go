package moonbitcodegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
	"github.com/oboard/rune-lang/internal/lexer"
)

func (g *generator) expr(expr ir.Expr) string {
	return g.exprPrec(expr, 0)
}

func (g *generator) exprPrec(expr ir.Expr, parentPrec int) string {
	switch e := expr.(type) {
	case *ir.Identifier:
		return mangleIdent(e.Name)
	case *ir.AtExpr:
		return "()"
	case *ir.ThisExpr:
		if len(g.thisNames) > 0 {
			return g.thisNames[len(g.thisNames)-1]
		}
		return "self"
	case *ir.IntegerLiteral:
		return strconv.Itoa(e.Value)
	case *ir.DoubleLiteral:
		if e.Raw != "" {
			return e.Raw
		}
		return strconv.FormatFloat(e.Value, 'f', -1, 64)
	case *ir.BigIntLiteral:
		return e.Value
	case *ir.StringLiteral:
		return quoteString(e.Value)
	case *ir.TemplateLiteral:
		return g.templateLiteral(e)
	case *ir.CharLiteral:
		return strconv.QuoteRune(e.Value)
	case *ir.RegexLiteral:
		g.addError(fmt.Errorf("MoonBit backend does not support regex literals"))
		return "()"
	case *ir.BoolLiteral:
		if e.Value {
			return "true"
		}
		return "false"
	case *ir.NullLiteral:
		return "None"
	case *ir.UnaryExpr:
		op := e.Op.String()
		if e.Op == lexer.Bang {
			op = "!"
		}
		out := op + g.exprPrec(e.Expr, 6)
		if 6 < parentPrec {
			return "(" + out + ")"
		}
		return out
	case *ir.PostfixExpr:
		switch e.Op {
		case lexer.PlusPlus:
			return g.expr(e.Expr) + " += 1"
		default:
			g.addError(fmt.Errorf("MoonBit backend does not support postfix operator %s", e.Op))
			return g.expr(e.Expr)
		}
	case *ir.ResultUnwrapExpr:
		g.addError(fmt.Errorf("MoonBit backend does not support result unwrap yet"))
		return g.expr(e.Expr)
	case *ir.BinaryExpr:
		if e.Op == lexer.QuestionQuestion {
			return fmt.Sprintf("%s.unwrap_or(%s)", g.exprPrec(e.Left, 6), g.expr(e.Right))
		}
		prec := mbtPrecedence(e.Op)
		out := fmt.Sprintf("%s %s %s", g.exprPrec(e.Left, prec), mbtBinaryOp(e.Op), g.exprPrec(e.Right, prec+1))
		if prec < parentPrec {
			return "(" + out + ")"
		}
		return out
	case *ir.TernaryExpr:
		alt := "()"
		if e.Alternative != nil {
			alt = g.expr(e.Alternative)
		}
		return fmt.Sprintf("if %s { %s } else { %s }", g.expr(e.Condition), g.expr(e.Consequence), alt)
	case *ir.AssignExpr:
		if target, ok := e.Target.(*ir.IndexExpr); ok {
			return fmt.Sprintf("%s[%s] = %s", g.expr(target.Receiver), g.expr(target.Index), g.expr(e.Value))
		}
		if e.Target != nil && e.Name == "" {
			return fmt.Sprintf("%s = %s", g.expr(e.Target), g.expr(e.Value))
		}
		return fmt.Sprintf("%s = %s", mangleIdent(e.Name), g.expr(e.Value))
	case *ir.CallExpr:
		return g.callExpr(e)
	case *ir.LambdaExpr:
		return g.lambda(e)
	case *ir.SelectorExpr:
		if member, ok := g.enumMemberSelector(e); ok {
			return member
		}
		if e.Static {
			if ident, ok := e.Receiver.(*ir.Identifier); ok {
				return mangleMethod(ident.Name, e.Name)
			}
		}
		if at, ok := e.Receiver.(*ir.AtExpr); ok {
			return "@" + at.Name + "." + e.Name
		}
		return g.expr(e.Receiver) + "." + mangleIdent(e.Name)
	case *ir.IndexExpr:
		if _, _, ok := checker.MapKeyValue(e.Receiver.ResultType()); ok {
			return fmt.Sprintf("%s.get(%s)", g.expr(e.Receiver), g.expr(e.Index))
		}
		if e.Receiver.ResultType() == checker.String && e.ResultType() == checker.Char {
			return fmt.Sprintf("%s[%s].unsafe_to_char()", g.expr(e.Receiver), g.expr(e.Index))
		}
		return fmt.Sprintf("%s[%s]", g.expr(e.Receiver), g.expr(e.Index))
	case *ir.ArrayLiteral:
		elems := make([]string, 0, len(e.Elements))
		for _, elem := range e.Elements {
			elems = append(elems, g.expr(elem))
		}
		return "[" + strings.Join(elems, ", ") + "]"
	case *ir.TupleLiteral:
		elems := make([]string, 0, len(e.Elements))
		for _, elem := range e.Elements {
			elems = append(elems, g.expr(elem))
		}
		return "(" + strings.Join(elems, ", ") + ")"
	case *ir.MapLiteral:
		if len(e.Entries) == 0 {
			return "{}"
		}
		parts := make([]string, 0, len(e.Entries))
		for _, entry := range e.Entries {
			parts = append(parts, fmt.Sprintf("%s: %s", g.expr(entry.Key), g.expr(entry.Value)))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case *ir.SpreadExpr:
		g.addError(fmt.Errorf("MoonBit backend does not support spread expressions yet"))
		return g.expr(e.Expr)
	case *ir.ReactiveLiteral:
		g.addError(fmt.Errorf("MoonBit backend does not support reactive literals"))
		return g.expr(e.Value)
	case *ir.StructLiteral:
		fields := make([]string, 0, len(e.Fields))
		for _, field := range e.Fields {
			fields = append(fields, fmt.Sprintf("%s: %s", mangleIdent(field.Name), g.expr(field.Value)))
		}
		return fmt.Sprintf("%s::{ %s }", mangleType(e.TypeName), strings.Join(fields, ", "))
	case *ir.AnonymousObjectLiteral:
		g.addError(fmt.Errorf("MoonBit backend does not support anonymous object literals yet"))
		return "()"
	case *ir.BlockExpr:
		return "{ " + g.blockInline(e, e.ResultType()) + " }"
	case *ir.PatternBlock:
		g.addError(fmt.Errorf("MoonBit backend only supports pattern blocks as function bodies"))
		return "()"
	case *ir.MatchExpr:
		return g.matchExpr(e)
	case *ir.WatchExpr:
		g.addError(fmt.Errorf("MoonBit backend does not support watch expressions"))
		return "()"
	case *ir.XMLElement:
		g.addError(fmt.Errorf("MoonBit backend does not support XML expressions"))
		return "()"
	default:
		return zeroValue(expr.ResultType())
	}
}

func (g *generator) templateLiteral(lit *ir.TemplateLiteral) string {
	parts := make([]string, 0, len(lit.Parts))
	for _, part := range lit.Parts {
		if part.Text != "" {
			parts = append(parts, quoteString(part.Text))
		}
		if part.Expr != nil {
			parts = append(parts, g.showExpr(part.Expr))
		}
	}
	if len(parts) == 0 {
		return "\"\""
	}
	return strings.Join(parts, " + ")
}

func (g *generator) showExpr(expr ir.Expr) string {
	if expr.ResultType() == checker.String {
		return g.expr(expr)
	}
	return fmt.Sprintf("%s.to_string()", g.expr(expr))
}

func (g *generator) callExpr(call *ir.CallExpr) string {
	if out, ok := g.moduleIntrinsicCall(call); ok {
		return out
	}
	if out, ok := g.receiverIntrinsicCall(call); ok {
		return out
	}
	args := make([]string, 0, len(call.Args))
	for _, arg := range call.Args {
		args = append(args, g.expr(arg))
	}
	if sel, ok := call.Callee.(*ir.SelectorExpr); ok {
		if sel.Static {
			return g.expr(sel) + "(" + strings.Join(args, ", ") + ")"
		}
		if typ := g.structTypeFromReceiver(sel.Receiver); typ != nil {
			args = append([]string{g.expr(sel.Receiver)}, args...)
			return mangleMethod(typ.Name, sel.Name) + "(" + strings.Join(args, ", ") + ")"
		}
		return g.expr(sel.Receiver) + "." + mangleIdent(sel.Name) + "(" + strings.Join(args, ", ") + ")"
	}
	return g.expr(call.Callee) + "(" + strings.Join(args, ", ") + ")"
}

func (g *generator) structTypeFromReceiver(receiver ir.Expr) *ir.StructType {
	name := string(receiver.ResultType())
	for _, typ := range g.file.Types {
		if typ.Name == name {
			return typ
		}
	}
	return nil
}

func (g *generator) enumMemberSelector(sel *ir.SelectorExpr) (string, bool) {
	ident, ok := sel.Receiver.(*ir.Identifier)
	if !ok {
		return "", false
	}
	for _, enum := range g.file.Enums {
		enumType := checker.Type(enum.Name)
		if enum.Name != ident.Name || sel.ResultType() != enumType || ident.ResultType() != enumType {
			continue
		}
		for _, member := range enum.Members {
			if member.Name != sel.Name {
				continue
			}
			if enumHasValueMembers(enum) {
				return mangleIdent(enum.Name + "_" + member.Name), true
			}
			return mangleType(enum.Name) + "::" + mangleType(member.Name), true
		}
	}
	return "", false
}

func (g *generator) lambda(lambda *ir.LambdaExpr) string {
	params := make([]string, 0, len(lambda.Params))
	for _, param := range lambda.Params {
		params = append(params, mangleIdent(param))
	}
	if block, ok := lambda.Body.(*ir.BlockExpr); ok {
		return "(" + strings.Join(params, ", ") + ") => { " + g.blockInline(block, block.ResultType()) + " }"
	}
	return "(" + strings.Join(params, ", ") + ") => " + g.expr(lambda.Body)
}

func (g *generator) blockInline(block *ir.BlockExpr, ret checker.Type) string {
	parts := make([]string, 0, len(block.Statements))
	for i, stmt := range block.Statements {
		last := i == len(block.Statements)-1
		switch s := stmt.(type) {
		case *ir.LetStmt:
			parts = append(parts, fmt.Sprintf("let %s = %s", mangleIdent(s.Name), g.expr(s.Value)))
		case *ir.AssignStmt:
			parts = append(parts, fmt.Sprintf("%s = %s", mangleIdent(s.Name), g.expr(s.Value)))
		case *ir.ExprStmt:
			if last && ret != checker.Void {
				parts = append(parts, g.expr(s.Expr))
			} else {
				parts = append(parts, g.expr(s.Expr))
			}
		}
	}
	return strings.Join(parts, "; ")
}

func (g *generator) matchExpr(match *ir.MatchExpr) string {
	subject := g.nextTemp("__match")
	branches := make([]string, 0, len(match.Branches))
	for _, branch := range match.Branches {
		branches = append(branches, fmt.Sprintf("%s => %s", g.patternFor(subject, branch.Pattern), g.expr(branch.Expr)))
	}
	return fmt.Sprintf("{ let %s = %s; match %s { %s } }", subject, g.expr(match.Subject), subject, strings.Join(branches, "; "))
}

func (g *generator) patternBlock(fn *ir.Function, block *ir.PatternBlock, ret checker.Type) {
	if len(fn.Params) != 1 {
		g.addError(fmt.Errorf("%s: pattern blocks currently require exactly one parameter", block.Pos))
		g.line(zeroValue(ret))
		return
	}
	subject := mangleIdent(fn.Params[0].Name)
	g.linef("match %s {", subject)
	g.indent++
	for _, branch := range block.Branches {
		g.linef("%s => %s", g.patternFor(subject, branch.Pattern), g.expr(branch.Expr))
	}
	g.indent--
	g.line("}")
}

func (g *generator) pattern(pattern ir.Pattern) string {
	return g.patternFor("_", pattern)
}

func (g *generator) patternFor(subject string, pattern ir.Pattern) string {
	switch p := pattern.(type) {
	case *ir.WildcardPattern:
		return "_"
	case *ir.BindingPattern:
		return mangleIdent(p.Name)
	case *ir.LiteralPattern:
		return g.expr(p.Value)
	case *ir.ComparePattern:
		return subjectGuard(subject, fmt.Sprintf("%s %s %s", subject, mbtBinaryOp(p.Op), g.expr(p.Value)))
	case *ir.RangePattern:
		return subjectGuard(subject, fmt.Sprintf("%s >= %s && %s <= %s", subject, g.expr(p.Start), subject, g.expr(p.End)))
	case *ir.OrPattern:
		conditions := make([]string, 0, len(p.Alternatives))
		for _, alt := range p.Alternatives {
			conditions = append(conditions, g.patternCondition(subject, alt))
		}
		return subjectGuard(subject, strings.Join(conditions, " || "))
	case *ir.TuplePattern:
		parts := make([]string, 0, len(p.Elements))
		for _, elem := range p.Elements {
			parts = append(parts, g.patternFor(subject, elem))
		}
		return "(" + strings.Join(parts, ", ") + ")"
	case *ir.ConstructorPattern:
		if p.Binding == "" {
			return p.Name
		}
		return fmt.Sprintf("%s(%s)", p.Name, mangleIdent(p.Binding))
	default:
		g.addError(fmt.Errorf("MoonBit backend does not support complex pattern %T yet", pattern))
		return "_"
	}
}

func subjectGuard(subject string, condition string) string {
	if subject == "_" {
		return "_"
	}
	return "_ if " + condition
}

func (g *generator) patternCondition(subject string, pattern ir.Pattern) string {
	switch p := pattern.(type) {
	case *ir.WildcardPattern, *ir.BindingPattern:
		return "true"
	case *ir.LiteralPattern:
		return fmt.Sprintf("%s == %s", subject, g.expr(p.Value))
	case *ir.ComparePattern:
		return fmt.Sprintf("%s %s %s", subject, mbtBinaryOp(p.Op), g.expr(p.Value))
	case *ir.RangePattern:
		return fmt.Sprintf("%s >= %s && %s <= %s", subject, g.expr(p.Start), subject, g.expr(p.End))
	case *ir.OrPattern:
		conditions := make([]string, 0, len(p.Alternatives))
		for _, alt := range p.Alternatives {
			conditions = append(conditions, g.patternCondition(subject, alt))
		}
		return "(" + strings.Join(conditions, " || ") + ")"
	default:
		g.addError(fmt.Errorf("MoonBit backend does not support guarded pattern %T yet", pattern))
		return "false"
	}
}
