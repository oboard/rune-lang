package format

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
)

func (f *formatter) expr(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Identifier:
		return e.Name
	case *ast.AtExpr:
		return "@" + e.Name
	case *ast.IntegerLiteral:
		return strconv.Itoa(e.Value)
	case *ast.StringLiteral:
		return strconv.Quote(e.Value)
	case *ast.BoolLiteral:
		if e.Value {
			return "true"
		}
		return "false"
	case *ast.UnaryExpr:
		return e.Op.String() + f.expr(e.Expr)
	case *ast.BinaryExpr:
		return fmt.Sprintf("%s %s %s", f.exprWithParens(e.Left), e.Op, f.exprWithParens(e.Right))
	case *ast.CallExpr:
		args := make([]string, 0, len(e.Args))
		for _, arg := range e.Args {
			args = append(args, f.expr(arg))
		}
		return fmt.Sprintf("%s(%s)", f.expr(e.Callee), strings.Join(args, ", "))
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
	case *ast.StructLiteral:
		return f.structLiteral(e)
	case *ast.AnonymousObjectLiteral:
		return f.anonymousObjectLiteral(e)
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

func (f *formatter) lambdaExpr(lambda *ast.LambdaExpr) string {
	params := f.lambdaParams(lambda)
	if lambda.Implicit {
		return f.expr(lambda.Body)
	}
	return "(" + strings.Join(params, ", ") + ") => " + f.expr(lambda.Body)
}

func (f *formatter) lambdaParams(lambda *ast.LambdaExpr) []string {
	params := make([]string, 0, len(lambda.Params))
	for i, param := range lambda.Params {
		if i < len(lambda.ParamTypes) && lambda.ParamTypes[i] != "" {
			params = append(params, fmt.Sprintf("%s: %s", param, lambda.ParamTypes[i]))
		} else {
			params = append(params, param)
		}
	}
	return params
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
		if i < len(block.Statements)-1 && strings.Contains(formatted, "\n") && separatesFollowingStatement(stmt) {
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
			params = append(params, fmt.Sprintf("%s: %s", param, lambda.ParamTypes[i]))
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
