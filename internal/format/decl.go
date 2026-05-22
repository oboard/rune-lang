package format

import (
	"fmt"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
)

func (f *formatter) structType(typ *ast.StructType) {
	f.linef("%s%s: {", typ.Name, formatGenerics(typ.Generics))
	f.indent++
	for _, field := range typ.Fields {
		f.linef("%s: %s", field.Name, field.Type)
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

func (f *formatter) function(fn *ast.Function) {
	for _, ann := range fn.Annotations {
		if ann.Value == "" {
			f.linef("@%s", ann.Name)
		} else {
			f.linef("@%s(%q)", ann.Name, ann.Value)
		}
	}
	var params []string
	for _, param := range fn.Params {
		params = append(params, fmt.Sprintf("%s: %s", param.Name, param.Type))
	}
	ret := ""
	if fn.ReturnType != "" {
		ret = " -> " + fn.ReturnType
	}
	switch body := fn.Body.(type) {
	case *ast.PatternBlock:
		f.linef("%s%s(%s)%s => {", fn.Name, formatGenerics(fn.Generics), strings.Join(params, ", "), ret)
		f.indent++
		for _, branch := range body.Branches {
			f.linef("%s => %s", f.pattern(branch.Pattern), f.expr(branch.Expr))
		}
		f.indent--
		f.line("}")
	case *ast.BlockExpr:
		f.linef("%s%s(%s)%s => {", fn.Name, formatGenerics(fn.Generics), strings.Join(params, ", "), ret)
		f.indent++
		for i, stmt := range body.Statements {
			formatted := f.stmt(stmt)
			f.line(formatted)
			if i < len(body.Statements)-1 && strings.Contains(formatted, "\n") {
				f.line("")
			}
		}
		f.indent--
		f.line("}")
	default:
		f.linef("%s%s(%s)%s => %s", fn.Name, formatGenerics(fn.Generics), strings.Join(params, ", "), ret, f.expr(fn.Body))
	}
}

func formatGenerics(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return "[" + strings.Join(names, ", ") + "]"
}
