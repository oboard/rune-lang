package gocodegen

import (
	"strings"
	"unicode/utf8"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
	"github.com/oboard/rune-lang/internal/stdlib"
	"github.com/oboard/rune-lang/internal/syntax"
)

func (g *generator) collectExprImports(expr ir.Expr) {
	if at, ok := expr.(*ir.AtExpr); ok {
		if path, ok := checker.GoPackageImportPath(at.Path); ok {
			g.imports[path] = true
		}
	}
	if fn, ok := g.stdlibFunctionFromExpr(expr); ok && fn.Go != nil && fn.Go.Import != "" {
		g.imports[fn.Go.Import] = true
	}
	if fn, ok := g.stdlibFunctionFromExpr(expr); ok && (fn.Intrinsic == "json.stringify" || fn.Intrinsic == "json.parse") {
		g.imports["encoding/json"] = true
	}
	if fn, ok := g.stdlibFunctionFromExpr(expr); ok && (fn.Intrinsic == "regex.new" || fn.Intrinsic == "regex.escape") {
		g.imports["regexp"] = true
	}
	if fn, ok := g.stdlibFunctionFromExpr(expr); ok && fn.Intrinsic == "fs.readFile" {
		g.imports["os"] = true
	}
	if fn, ok := g.stdlibFunctionFromExpr(expr); ok && fn.Intrinsic == "int.toString" {
		g.imports["strconv"] = true
	}
	if fn, ok := g.stdlibFunctionFromExpr(expr); ok {
		switch fn.Intrinsic {
		case "int.toBigInt", "bigint.fromInt", "bigint.toDouble":
			g.imports["math/big"] = true
		case "double.trunc", "double.floor", "double.ceil", "double.round":
			g.imports["math"] = true
		}
	}
	if call, ok := expr.(*ir.CallExpr); ok {
		if sel, ok := call.Callee.(*ir.SelectorExpr); ok {
			switch sel.Receiver.ResultType() {
			case checker.Int:
				if sel.Name == "toString" {
					g.imports["strconv"] = true
				}
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
	if g.file.Stdlib == nil {
		return nil, false
	}
	if at, ok := sel.Receiver.(*ir.AtExpr); ok && at.Name != "" {
		return g.file.Stdlib.Function(at.Name, sel.Name)
	}
	if moduleName, ok := checker.ModuleNamespaceName(sel.Receiver.ResultType()); ok {
		return g.file.Stdlib.Function(moduleName, sel.Name)
	}
	return nil, false
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
		ch, size := utf8.DecodeRuneInString(src[j:])
		if ch == utf8.RuneError && size == 0 || !syntax.IsIdentStart(ch) {
			b.WriteByte('$')
			i++
			continue
		}
		j += size
		for j < len(src) {
			ch, size = utf8.DecodeRuneInString(src[j:])
			if ch == utf8.RuneError && size == 0 || !syntax.IsIdentContinue(ch) {
				break
			}
			j += size
		}
		b.WriteString(mangleIdent(src[i+1 : j]))
		i = j
	}
	return b.String()
}
