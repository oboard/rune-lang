// Package astprint renders the Rune AST back to source. It is the printer used
// by tsfront's Rune output path and by formatter/roundtrip tests.
package astprint

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

// File renders a complete Rune source file.
func File(file *ast.File) string {
	var p printer
	for _, imp := range file.Imports {
		p.line("@" + strconv.Quote(imp.Path))
	}
	if len(file.Imports) > 0 {
		p.newline()
	}
	for _, typ := range file.Types {
		p.structType(typ)
		p.newline()
	}
	for _, enum := range file.Enums {
		p.enumType(enum)
		p.newline()
	}
	for _, constant := range file.Constants {
		p.constant(constant)
		p.newline()
	}
	for _, fn := range file.Functions {
		p.function(fn)
		p.newline()
	}
	for _, test := range file.Tests {
		p.line("? " + strconv.Quote(test.Name) + " ")
		p.expr(test.Body)
		p.newline()
	}
	return strings.TrimRight(p.b.String(), "\n") + "\n"
}

func (p *printer) structType(typ *ast.StructType) {
	prefix := "+ "
	if typ.Private {
		prefix = ""
	}
	p.line(prefix + typ.Name + ": {")
	p.indent++
	for _, field := range typ.Fields {
		fieldPrefix := ""
		if field.Private {
			fieldPrefix = "- "
		}
		p.line(fieldPrefix + field.Name + ": " + field.Type.Display())
	}
	for _, method := range typ.Methods {
		p.method(method)
	}
	p.indent--
	p.line("}")
}

func (p *printer) enumType(enum *ast.EnumType) {
	prefix := "+ "
	if enum.Private {
		prefix = ""
	}
	p.line(prefix + enum.Name + ": {")
	p.indent++
	for _, member := range enum.Members {
		line := member.Name
		if len(member.Params) > 0 {
			parts := make([]string, 0, len(member.Params))
			for _, param := range member.Params {
				if param.Name != "" {
					parts = append(parts, param.Name+": "+param.Type.Display())
				} else {
					parts = append(parts, param.Type.Display())
				}
			}
			line += "(" + strings.Join(parts, ", ") + ")"
		}
		p.line(line)
	}
	for _, method := range enum.Methods {
		p.method(method)
	}
	p.indent--
	p.line("}")
}

// method renders a declaration that lives inside a struct/enum body.
func (p *printer) method(fn *ast.Function) {
	prefix := ""
	if fn.Private {
		prefix = "- "
	}
	if fn.Static {
		prefix += "::"
	}
	var b strings.Builder
	b.WriteString(prefix)
	b.WriteString(fn.Name)
	b.WriteString("(")
	for index, param := range fn.Params {
		if index > 0 {
			b.WriteString(", ")
		}
		b.WriteString(param.Name)
		if !param.Type.IsZero() {
			b.WriteString(": ")
			b.WriteString(param.Type.Display())
		}
	}
	b.WriteString(")")
	if !fn.ReturnType.IsZero() {
		b.WriteString(" -> ")
		b.WriteString(fn.ReturnType.Display())
	}
	b.WriteString(" => ")
	p.b.WriteString(strings.Repeat("  ", p.indent))
	p.b.WriteString(b.String())
	p.expr(fn.Body)
	p.b.WriteByte('\n')
}

// Expr renders a single expression.
func Expr(expr ast.Expr) string {
	var p printer
	p.expr(expr)
	return p.b.String()
}

type printer struct {
	b      strings.Builder
	indent int
}

func (p *printer) line(text string) {
	p.b.WriteString(strings.Repeat("  ", p.indent))
	p.b.WriteString(text)
	p.b.WriteByte('\n')
}

func (p *printer) newline() {
	p.b.WriteByte('\n')
}

func (p *printer) function(fn *ast.Function) {
	var b strings.Builder
	if !fn.Private {
		b.WriteString("+ ")
	}
	if fn.Routine {
		b.WriteString("~ ")
	}
	b.WriteString(fn.Name)
	if len(fn.Generics) > 0 {
		b.WriteString("[")
		b.WriteString(strings.Join(fn.Generics, ", "))
		b.WriteString("]")
	}
	b.WriteString("(")
	for index, param := range fn.Params {
		if index > 0 {
			b.WriteString(", ")
		}
		b.WriteString(param.Name)
		if !param.Type.IsZero() {
			b.WriteString(": ")
			b.WriteString(param.Type.Display())
		}
	}
	b.WriteString(")")
	if !fn.ReturnType.IsZero() {
		b.WriteString(" -> ")
		b.WriteString(fn.ReturnType.Display())
	}
	b.WriteString(" => ")
	p.b.WriteString(strings.Repeat("  ", p.indent))
	p.b.WriteString(b.String())
	p.expr(fn.Body)
	p.b.WriteByte('\n')
}

func (p *printer) constant(constant *ast.ConstDecl) {
	prefix := ""
	if !constant.Private {
		prefix = "+ "
	}
	p.b.WriteString(prefix + constant.Name + " := ")
	p.expr(constant.Value)
	p.b.WriteByte('\n')
}

func (p *printer) stmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.LetStmt:
		prefix := ""
		if s.Signal {
			prefix = "$"
		}
		op := ":="
		if s.Mutable {
			op = ":=:"
		}
		p.b.WriteString(strings.Repeat("  ", p.indent) + prefix + s.Name + " " + op + " ")
		p.expr(s.Value)
		if !s.Type.IsZero() {
			p.b.WriteString(" : " + s.Type.Display())
		}
		p.b.WriteByte('\n')
	case *ast.AssignStmt:
		p.b.WriteString(strings.Repeat("  ", p.indent) + s.Name + " = ")
		p.expr(s.Value)
		p.b.WriteByte('\n')
	case *ast.ReturnStmt:
		p.b.WriteString(strings.Repeat("  ", p.indent) + "return")
		if s.Value != nil {
			p.b.WriteByte(' ')
			p.expr(s.Value)
		}
		p.b.WriteByte('\n')
	case *ast.BreakStmt:
		p.line("^")
	case *ast.ContinueStmt:
		p.line("^>")
	case *ast.ThrowStmt:
		p.b.WriteString(strings.Repeat("  ", p.indent) + "!! ")
		p.expr(s.Value)
		p.b.WriteByte('\n')
	case *ast.TryStmt:
		p.line("{")
		p.indent++
		for _, stmt := range s.Body.Statements {
			p.stmt(stmt)
		}
		p.indent--
		if s.Catch != nil {
			if s.CatchName != "" {
				p.line("} !? {")
			} else {
				p.line("} ?! {")
			}
			p.indent++
			if s.CatchName != "" {
				p.line(s.CatchName + " => {")
				p.indent++
			}
			for _, stmt := range s.Catch.Statements {
				p.stmt(stmt)
			}
			if s.CatchName != "" {
				p.indent--
				p.line("}")
			}
			p.indent--
		}
		if s.Finally != nil {
			if s.Catch != nil {
				p.line("} -- {")
			} else {
				p.line("} -- {")
			}
			p.indent++
			for _, stmt := range s.Finally.Statements {
				p.stmt(stmt)
			}
			p.indent--
		}
		p.line("}")
	case *ast.ExprStmt:
		p.b.WriteString(strings.Repeat("  ", p.indent))
		p.expr(s.Expr)
		p.b.WriteByte('\n')
	}
}

func (p *printer) expr(expr ast.Expr) {
	switch e := expr.(type) {
	case nil:
		return
	case *ast.Identifier:
		p.b.WriteString(e.Name)
	case *ast.AtExpr:
		if e.Path != "" {
			p.b.WriteString("@" + strconv.Quote(e.Path))
		} else {
			p.b.WriteString("@" + e.Name)
		}
	case *ast.ThisExpr:
		p.b.WriteString("this")
	case *ast.IntegerLiteral:
		p.b.WriteString(strconv.Itoa(e.Value))
	case *ast.DoubleLiteral:
		raw := e.Raw
		if raw == "" {
			raw = strconv.FormatFloat(e.Value, 'f', -1, 64)
		}
		if !strings.ContainsAny(raw, ".eE") {
			raw += ".0"
		}
		p.b.WriteString(raw)
	case *ast.BigIntLiteral:
		p.b.WriteString(e.Value + "n")
	case *ast.StringLiteral:
		p.b.WriteString(strconv.Quote(e.Value))
	case *ast.TemplateLiteral:
		p.b.WriteString("`")
		for _, part := range e.Parts {
			p.b.WriteString(strings.ReplaceAll(part.Text, "`", "\\`"))
			if part.Expr != nil {
				p.b.WriteString("\\(")
				p.expr(part.Expr)
				p.b.WriteString(")")
			}
		}
		p.b.WriteString("`")
	case *ast.CharLiteral:
		p.b.WriteString(strconv.QuoteRune(e.Value))
	case *ast.RegexLiteral:
		if e.Raw != "" {
			p.b.WriteString(e.Raw)
		} else {
			p.b.WriteString("/" + e.Pattern + "/" + e.Flags)
		}
	case *ast.BoolLiteral:
		if e.Value {
			p.b.WriteString("true")
		} else {
			p.b.WriteString("false")
		}
	case *ast.NullLiteral:
		p.b.WriteString("null")
	case *ast.UndefinedLiteral:
		p.b.WriteString("undefined")
	case *ast.UnaryExpr:
		p.b.WriteString(e.Op.String())
		p.expr(e.Expr)
	case *ast.TypeOfExpr:
		p.b.WriteString("typeof ")
		p.expr(e.Argument)
	case *ast.BinaryExpr:
		p.b.WriteString("(")
		p.expr(e.Left)
		p.b.WriteString(" " + e.Op.String() + " ")
		p.expr(e.Right)
		p.b.WriteString(")")
	case *ast.TernaryExpr:
		p.b.WriteString("(")
		p.condition(e.Condition)
		p.b.WriteString(") ? ")
		p.expr(e.Consequence)
		if e.Alternative != nil {
			p.b.WriteString(" : ")
			p.expr(e.Alternative)
		}
	case *ast.IfExpr:
		p.b.WriteString("(")
		p.condition(e.Cond)
		p.b.WriteString(") ? ")
		p.block(e.Then)
		if e.Else != nil {
			p.b.WriteString(" : ")
			p.block(e.Else)
		}
	case *ast.WhileExpr:
		p.b.WriteString("* (")
		p.condition(e.Cond)
		p.b.WriteString(") ")
		p.block(e.Body)
	case *ast.ForOfExpr:
		p.b.WriteString("for (const " + e.Name + " of ")
		p.expr(e.Iterable)
		p.b.WriteString(") ")
		p.block(e.Body)
	case *ast.AssignExpr:
		if e.Target != nil {
			p.expr(e.Target)
		} else {
			p.b.WriteString(e.Name)
		}
		p.b.WriteString(" = ")
		p.expr(e.Value)
	case *ast.CallExpr:
		p.expr(e.Callee)
		p.b.WriteString("(")
		for index, arg := range e.Args {
			if index > 0 {
				p.b.WriteString(", ")
			}
			p.expr(arg)
		}
		p.b.WriteString(")")
	case *ast.SelectorExpr:
		if _, ok := e.Receiver.(*ast.ThisExpr); ok {
			p.b.WriteString("." + e.Name)
			return
		}
		p.expr(e.Receiver)
		if e.Static {
			p.b.WriteString("::")
		} else {
			p.b.WriteString(".")
		}
		p.b.WriteString(e.Name)
	case *ast.IndexExpr:
		p.expr(e.Receiver)
		p.b.WriteString("[")
		p.expr(e.Index)
		p.b.WriteString("]")
	case *ast.ArrayLiteral:
		p.b.WriteString("[")
		for index, elem := range e.Elements {
			if index > 0 {
				p.b.WriteString(", ")
			}
			p.expr(elem)
		}
		p.b.WriteString("]")
	case *ast.MapLiteral:
		p.b.WriteString("{")
		for index, entry := range e.Entries {
			if index > 0 {
				p.b.WriteString(", ")
			}
			p.expr(entry.Key)
			p.b.WriteString(": ")
			p.expr(entry.Value)
		}
		p.b.WriteString("}")
	case *ast.AnonymousObjectLiteral:
		p.b.WriteString("{")
		for index, field := range e.Fields {
			if index > 0 {
				p.b.WriteString(", ")
			}
			p.b.WriteString(field.Name + ": ")
			p.expr(field.Value)
		}
		p.b.WriteString("}")
	case *ast.StructLiteral:
		p.b.WriteString(e.TypeName + " {")
		for index, field := range e.Fields {
			if index > 0 {
				p.b.WriteString(", ")
			}
			p.b.WriteString(field.Name + ": ")
			p.expr(field.Value)
		}
		p.b.WriteString("}")
	case *ast.BlockExpr:
		p.block(e)
	case *ast.LambdaExpr:
		p.b.WriteString("(")
		for index, param := range e.Params {
			if index > 0 {
				p.b.WriteString(", ")
			}
			p.b.WriteString(param)
			if index < len(e.ParamTypes) && !e.ParamTypes[index].IsZero() {
				p.b.WriteString(": " + e.ParamTypes[index].Display())
			}
		}
		p.b.WriteString(") => ")
		p.expr(e.Body)
	case *ast.MatchExpr:
		p.expr(e.Subject)
		p.b.WriteString(" {\n")
		p.indent++
		for _, branch := range e.Branches {
			p.b.WriteString(strings.Repeat("  ", p.indent))
			p.pattern(branch.Pattern)
			p.b.WriteString(" => ")
			p.expr(branch.Expr)
			p.b.WriteByte('\n')
		}
		p.indent--
		p.b.WriteString(strings.Repeat("  ", p.indent) + "}")
	case *ast.ResultUnwrapExpr:
		p.expr(e.Expr)
		p.b.WriteString("?")
	case *ast.PostfixExpr:
		p.expr(e.Expr)
		p.b.WriteString(e.Op.String())
	case *ast.SpreadExpr:
		p.b.WriteString("..")
		p.expr(e.Expr)
	case *ast.ReactiveLiteral:
		p.b.WriteString("$")
		p.expr(e.Value)
	case *ast.CompileTimeExpr:
		p.b.WriteString("'")
		p.expr(e.Expr)
	default:
		p.b.WriteString("/* unsupported " + fmt.Sprintf("%T", expr) + " */")
	}
}

func (p *printer) condition(expr ast.Expr) {
	if binary, ok := expr.(*ast.BinaryExpr); ok {
		p.expr(binary.Left)
		p.b.WriteString(" " + binary.Op.String() + " ")
		p.expr(binary.Right)
		return
	}
	p.expr(expr)
}

func (p *printer) block(expr ast.Expr) {
	block, ok := expr.(*ast.BlockExpr)
	if !ok {
		p.b.WriteString("{\n")
		p.indent++
		p.b.WriteString(strings.Repeat("  ", p.indent))
		p.expr(expr)
		p.b.WriteByte('\n')
		p.indent--
		p.b.WriteString(strings.Repeat("  ", p.indent) + "}")
		return
	}
	p.b.WriteString("{\n")
	p.indent++
	for _, stmt := range block.Statements {
		p.stmt(stmt)
	}
	p.indent--
	p.b.WriteString(strings.Repeat("  ", p.indent) + "}")
}

func (p *printer) pattern(pattern ast.Pattern) {
	switch pat := pattern.(type) {
	case *ast.WildcardPattern:
		p.b.WriteString("_")
	case *ast.BindingPattern:
		p.b.WriteString(pat.Name)
	case *ast.LiteralPattern:
		p.expr(pat.Value)
	case *ast.ComparePattern:
		p.b.WriteString(pat.Op.String())
		p.b.WriteString(" ")
		p.expr(pat.Value)
	case *ast.OrPattern:
		for index, alternative := range pat.Alternatives {
			if index > 0 {
				p.b.WriteString(" | ")
			}
			p.pattern(alternative)
		}
	case *ast.TuplePattern:
		p.b.WriteString("(")
		for index, elem := range pat.Elements {
			if index > 0 {
				p.b.WriteString(", ")
			}
			p.pattern(elem)
		}
		p.b.WriteString(")")
	case *ast.ArrayPattern:
		p.b.WriteString("[")
		for index, elem := range pat.Elements {
			if index > 0 {
				p.b.WriteString(", ")
			}
			p.pattern(elem)
		}
		if pat.RestBinding != "" {
			p.b.WriteString(", .." + pat.RestBinding)
		}
		p.b.WriteString("]")
	case *ast.ConstructorPattern:
		p.b.WriteString(pat.Name + "(")
		for index, elem := range pat.Args {
			if index > 0 {
				p.b.WriteString(", ")
			}
			p.pattern(elem)
		}
		if pat.Rest {
			p.b.WriteString(", ..")
		}
		p.b.WriteString(")")
	default:
		p.b.WriteString("_")
	}
}

// Ensure lexer is referenced when this package is built without expressions
// that use Kind.String directly (kept for future pattern support).
var _ = lexer.Illegal
