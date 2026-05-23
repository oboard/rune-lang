package gocodegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
)

type jsonObjectField struct {
	name string
	typ  checker.Type
}

func (g *generator) jsonStringifyCall(call *ir.CallExpr) (string, bool) {
	fn, ok := g.stdlibFunctionFromCall(call)
	if !ok || fn.Intrinsic != "json.stringify" {
		return "", false
	}
	if len(call.Args) != 1 {
		return "/* invalid @json.stringify */", true
	}
	value := g.jsonValueExpr(call.Args[0])
	return fmt.Sprintf("func() string { %s, _ := json.Marshal(%s); return string(%s) }()", mangleIdent("rune_json_bytes"), value, mangleIdent("rune_json_bytes")), true
}

func (g *generator) jsonValueExpr(expr ir.Expr) string {
	switch e := expr.(type) {
	case *ir.AnonymousObjectLiteral:
		return g.jsonObjectLiteral(e.Fields, e.ResultType())
	case *ir.StructLiteral:
		return g.jsonObjectLiteral(e.Fields, e.ResultType())
	case *ir.ArrayLiteral:
		if elem, ok := checker.ArrayElement(e.ResultType()); ok && g.jsonNeedsConversion(elem) && !jsonArrayLiteralHasSpread(e) {
			elems := make([]string, 0, len(e.Elements))
			for _, elem := range e.Elements {
				elems = append(elems, g.jsonValueExpr(elem))
			}
			return g.jsonGoType(e.ResultType()) + "{" + strings.Join(elems, ", ") + "}"
		}
	}
	return g.jsonValueFromSource(g.expr(expr), expr.ResultType())
}

func (g *generator) jsonValueFromSource(source string, typ checker.Type) string {
	if fields, ok := g.jsonObjectFields(typ); ok {
		jsonType := g.jsonGoType(typ)
		value := mangleIdent("rune_json_value")
		return fmt.Sprintf("func() %s { %s := %s; return %s{%s} }()", jsonType, value, source, jsonType, g.jsonObjectValues(value, fields))
	}
	if elem, ok := checker.ArrayElement(typ); ok {
		if !g.jsonNeedsConversion(elem) {
			return source
		}
		jsonType := g.jsonGoType(typ)
		value := mangleIdent("rune_json_value")
		out := mangleIdent("rune_json_out")
		item := mangleIdent("rune_json_item")
		return fmt.Sprintf("func() %s { %s := %s; %s := make(%s, 0, len(%s)); for _, %s := range %s { %s = append(%s, %s) }; return %s }()", jsonType, value, source, out, jsonType, value, item, value, out, out, g.jsonValueFromSource(item, elem), out)
	}
	return source
}

func (g *generator) jsonGoType(typ checker.Type) string {
	if fields, ok := g.jsonObjectFields(typ); ok {
		parts := make([]string, 0, len(fields))
		for idx, field := range fields {
			parts = append(parts, fmt.Sprintf("F%d %s `json:%s`", idx, g.jsonGoType(field.typ), strconv.Quote(field.name)))
		}
		return "struct{" + strings.Join(parts, "; ") + "}"
	}
	if elem, ok := checker.ArrayElement(typ); ok {
		return "[]" + g.jsonGoType(elem)
	}
	return goType(typ)
}

func (g *generator) jsonObjectValues(source string, fields []jsonObjectField) string {
	values := make([]string, 0, len(fields))
	for idx, field := range fields {
		fieldSource := source + "." + mangleIdent(field.name)
		values = append(values, fmt.Sprintf("F%d: %s", idx, g.jsonValueFromSource(fieldSource, field.typ)))
	}
	return strings.Join(values, ", ")
}

func (g *generator) jsonObjectLiteral(fields []ir.FieldValue, typ checker.Type) string {
	jsonType := g.jsonGoType(typ)
	return fmt.Sprintf("%s{%s}", jsonType, g.jsonObjectLiteralValues(fields, typ))
}

func (g *generator) jsonObjectLiteralValues(fields []ir.FieldValue, typ checker.Type) string {
	objectFields, ok := g.jsonObjectFields(typ)
	if !ok {
		return ""
	}
	byName := map[string]ir.Expr{}
	for _, field := range fields {
		byName[field.Name] = field.Value
	}
	values := make([]string, 0, len(objectFields))
	for idx, field := range objectFields {
		value := byName[field.name]
		if value == nil {
			continue
		}
		values = append(values, fmt.Sprintf("F%d: %s", idx, g.jsonValueExpr(value)))
	}
	return strings.Join(values, ", ")
}

func (g *generator) jsonNeedsConversion(typ checker.Type) bool {
	if _, ok := g.jsonObjectFields(typ); ok {
		return true
	}
	if elem, ok := checker.ArrayElement(typ); ok {
		return g.jsonNeedsConversion(elem)
	}
	return false
}

func (g *generator) jsonObjectFields(typ checker.Type) ([]jsonObjectField, bool) {
	fields, ok := parseGoObjectType(string(typ))
	if ok {
		out := make([]jsonObjectField, 0, len(fields))
		for _, field := range fields {
			fieldType := checker.Type(field.typ)
			if jsonOmitType(fieldType) {
				continue
			}
			out = append(out, jsonObjectField{name: field.name, typ: fieldType})
		}
		return out, true
	}
	for _, candidate := range g.file.Types {
		if candidate.Name != string(typ) {
			continue
		}
		out := make([]jsonObjectField, 0, len(candidate.Fields))
		for _, field := range candidate.Fields {
			if jsonOmitType(field.Type) {
				continue
			}
			out = append(out, jsonObjectField{name: field.Name, typ: field.Type})
		}
		return out, true
	}
	return nil, false
}

func jsonOmitType(typ checker.Type) bool {
	if typ == checker.Void || typ == checker.Symbol {
		return true
	}
	_, _, ok := parseGoFuncType(string(typ))
	return ok
}

func jsonArrayLiteralHasSpread(lit *ir.ArrayLiteral) bool {
	for _, elem := range lit.Elements {
		if _, ok := elem.(*ir.SpreadExpr); ok {
			return true
		}
	}
	return false
}
