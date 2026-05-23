package gocodegen

import (
	"strings"
	"unicode"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func (g *generator) collectExprImports(expr ir.Expr) {
	if fn, ok := g.stdlibFunctionFromExpr(expr); ok && fn.Go != nil && fn.Go.Import != "" {
		g.imports[fn.Go.Import] = true
	}
	if call, ok := expr.(*ir.CallExpr); ok {
		if sel, ok := call.Callee.(*ir.SelectorExpr); ok {
			switch sel.Receiver.ResultType() {
			case checker.String:
				switch sel.Name {
				case "includes", "startsWith", "endsWith", "indexOf", "lastIndexOf", "toLowerCase", "toUpperCase", "trim", "trimStart", "trimEnd", "repeat", "replace", "replaceAll", "split":
					g.imports["strings"] = true
				}
				switch sel.Name {
				case "trimStart", "trimEnd":
					g.imports["unicode"] = true
				}
			case checker.Bool:
				if sel.Name == "toString" {
					g.imports["strconv"] = true
				}
			}
		}
	}
}

func (g *generator) goFFICall(call *ir.CallExpr) (string, bool) {
	fn, ok := g.stdlibFunctionFromCall(call)
	if !ok || fn.Intrinsic == "" {
		return "", false
	}
	switch fn.Intrinsic {
	case "go.import":
		return "/* @go.import must be top-level */", true
	case "go.stmt":
		if len(call.Args) != 1 {
			return "/* invalid @go.stmt */", true
		}
		lit, ok := call.Args[0].(*ir.StringLiteral)
		if !ok {
			return "/* invalid @go.stmt */", true
		}
		return rewriteGoFFI(lit.Value), true
	case "go.expr":
		if len(call.Args) != 1 {
			return "/* invalid @go.expr */", true
		}
		lit, ok := call.Args[0].(*ir.StringLiteral)
		if !ok {
			return "/* invalid @go.expr */", true
		}
		return rewriteGoFFI(lit.Value), true
	default:
		return "", false
	}
}

func (g *generator) stdlibFunctionFromExpr(expr ir.Expr) (*stdlib.Function, bool) {
	call, ok := expr.(*ir.CallExpr)
	if !ok {
		return nil, false
	}
	return g.stdlibFunctionFromCall(call)
}

func (g *generator) stdlibFunctionFromCall(call *ir.CallExpr) (*stdlib.Function, bool) {
	sel, ok := call.Callee.(*ir.SelectorExpr)
	if !ok {
		return nil, false
	}
	at, ok := sel.Receiver.(*ir.AtExpr)
	if !ok || g.file.Stdlib == nil {
		return nil, false
	}
	return g.file.Stdlib.Function(at.Name, sel.Name)
}

func rewriteGoFFI(src string) string {
	var b strings.Builder
	for i := 0; i < len(src); {
		if src[i] != '$' {
			b.WriteByte(src[i])
			i++
			continue
		}
		if i+1 < len(src) && src[i+1] == '$' {
			b.WriteByte('$')
			i += 2
			continue
		}
		j := i + 1
		if j >= len(src) || !isGoFFIIdentStart(rune(src[j])) {
			b.WriteByte('$')
			i++
			continue
		}
		j++
		for j < len(src) && isGoFFIIdentContinue(rune(src[j])) {
			j++
		}
		b.WriteString(mangleIdent(src[i+1 : j]))
		i = j
	}
	return b.String()
}

func isGoFFIIdentStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isGoFFIIdentContinue(r rune) bool {
	return isGoFFIIdentStart(r) || unicode.IsDigit(r)
}
