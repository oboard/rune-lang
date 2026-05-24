package format

import (
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
)

func (f *formatter) structType(typ *ast.StructType) {
	f.linef("%s%s: {", typ.Name, formatGenerics(typ.Generics))
	f.indent++
	for _, field := range typ.Fields {
		f.linef("%s: %s", field.Name, formatType(field.Type, field.TypeDisplay))
	}
	if len(typ.Fields) > 0 && len(typ.Methods) > 0 {
		f.line("")
	}
	for i, method := range typ.Methods {
		if i > 0 {
			f.line("")
		}
		f.function(method)
	}
	f.indent--
	f.line("}")
}

func (f *formatter) enumType(enum *ast.EnumType) {
	f.linef("%s: {", enum.Name)
	f.indent++
	for _, member := range enum.Members {
		f.linef("%s = %d", member.Name, member.Value)
	}
	f.indent--
	f.line("}")
}

func (f *formatter) function(fn *ast.Function) {
	for _, ann := range fn.Annotations {
		if ann.Value == "" {
			f.linef("@%s", ann.Name)
		} else {
			f.linef("@%s(%q)", ann.Name, ann.Value)
		}
	}
	signature := f.functionSignature(fn)
	switch body := fn.Body.(type) {
	case *ast.PatternBlock:
		f.lineSignature(signature, " => {")
		f.indent++
		for _, branch := range body.Branches {
			f.linef("%s => %s", f.pattern(branch.Pattern), f.expr(branch.Expr))
		}
		f.indent--
		f.line("}")
	case *ast.BlockExpr:
		f.lineSignature(signature, " => {")
		f.indent++
		for i, stmt := range body.Statements {
			formatted := f.stmt(stmt)
			f.line(formatted)
			if i < len(body.Statements)-1 && separatesFollowingStatement(stmt, body.Statements[i+1], formatted) {
				f.line("")
			}
		}
		f.indent--
		f.line("}")
	default:
		bodyText := f.expr(fn.Body)
		if _, ok := fn.Body.(*ast.AnonymousObjectLiteral); ok {
			bodyText = "(" + bodyText + ")"
		}
		f.lineSignature(signature, " => "+bodyText)
	}
}

func (f *formatter) functionSignature(fn *ast.Function) []string {
	params := make([]string, 0, len(fn.Params))
	for _, param := range fn.Params {
		params = append(params, formatParam(param))
	}
	ret := formatReturnType(fn.ReturnType, fn.ReturnDisplay)
	single := fn.Name + formatGenerics(fn.Generics) + "(" + strings.Join(params, ", ") + ")" + ret
	if len(indentString(f.indent)+single) <= maxLineLength {
		return []string{single}
	}

	lines := []string{fn.Name + formatGenerics(fn.Generics) + "("}
	for i, param := range fn.Params {
		paramLines := f.formatParamLines(param)
		if i < len(fn.Params)-1 {
			paramLines[len(paramLines)-1] += ","
		}
		for _, line := range paramLines {
			lines = append(lines, "  "+line)
		}
	}
	lines = append(lines, ")"+ret)
	return lines
}

func (f *formatter) lineSignature(lines []string, suffix string) {
	for _, line := range lines[:len(lines)-1] {
		f.line(line)
	}
	f.line(lines[len(lines)-1] + suffix)
}

func formatParam(param ast.Param) string {
	if param.Type == "" {
		return param.Name
	}
	return param.Name + ": " + formatType(param.Type, param.TypeDisplay)
}

func (f *formatter) formatParamLines(param ast.Param) []string {
	if param.Type == "" {
		return []string{param.Name}
	}
	typ := formatType(param.Type, param.TypeDisplay)
	single := param.Name + ": " + typ
	if len(indentString(f.indent+1)+single) <= maxLineLength {
		return []string{single}
	}
	if params, ret, ok := splitFunctionType(typ); ok {
		lines := []string{param.Name + ": ("}
		for i, nested := range params {
			line := "  " + nested
			if i < len(params)-1 {
				line += ","
			}
			lines = append(lines, line)
		}
		lines = append(lines, ") -> "+ret)
		return lines
	}
	return []string{single}
}

func formatReturnType(canonical string, display string) string {
	if canonical == "" {
		return ""
	}
	return " -> " + formatType(canonical, display)
}

func splitFunctionType(typ string) ([]string, string, bool) {
	if !strings.HasPrefix(typ, "(") {
		return nil, "", false
	}
	close := matchingCloseParen(typ)
	if close < 0 {
		return nil, "", false
	}
	rest := strings.TrimSpace(typ[close+1:])
	if !strings.HasPrefix(rest, "->") {
		return nil, "", false
	}
	params := splitTopLevelComma(typ[1:close])
	ret := strings.TrimSpace(strings.TrimPrefix(rest, "->"))
	return params, ret, ret != ""
}

func matchingCloseParen(s string) int {
	depth := 0
	bracketDepth := 0
	for i, ch := range s {
		switch ch {
		case '[':
			if depth > 0 {
				bracketDepth++
			}
		case ']':
			if depth > 0 && bracketDepth > 0 {
				bracketDepth--
			}
		case '(':
			if bracketDepth == 0 {
				depth++
			}
		case ')':
			if bracketDepth == 0 {
				depth--
				if depth == 0 {
					return i
				}
			}
		}
	}
	return -1
}

func splitTopLevelComma(s string) []string {
	var parts []string
	start := 0
	parenDepth := 0
	bracketDepth := 0
	for i, ch := range s {
		switch ch {
		case '(':
			parenDepth++
		case ')':
			if parenDepth > 0 {
				parenDepth--
			}
		case '[':
			bracketDepth++
		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
		case ',':
			if parenDepth == 0 && bracketDepth == 0 {
				parts = append(parts, strings.TrimSpace(s[start:i]))
				start = i + 1
			}
		}
	}
	if tail := strings.TrimSpace(s[start:]); tail != "" {
		parts = append(parts, tail)
	}
	return parts
}

func separatesFollowingStatement(stmt ast.Stmt, next ast.Stmt, formatted string) bool {
	if stmtIsXMLExpr(next) {
		return true
	}
	let, ok := stmt.(*ast.LetStmt)
	if !ok {
		return false
	}
	_, ok = let.Value.(*ast.AnonymousObjectLiteral)
	return ok && strings.Contains(formatted, "\n")
}

func stmtIsXMLExpr(stmt ast.Stmt) bool {
	expr, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return false
	}
	_, ok = expr.Expr.(*ast.XMLElement)
	return ok
}

func formatGenerics(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return "[" + strings.Join(names, ", ") + "]"
}

func formatType(canonical string, display string) string {
	if display != "" {
		return display
	}
	return canonical
}
