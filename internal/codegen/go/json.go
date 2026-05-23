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

type jsonGoEmitter struct {
	g     *generator
	names map[string]string
	decls []string
	next  int
}

func (g *generator) jsonStringifyCall(call *ir.CallExpr) (string, bool) {
	fn, ok := g.stdlibFunctionFromCall(call)
	if !ok || fn.Intrinsic != "json.stringify" {
		return "", false
	}
	if len(call.Args) != 1 {
		return "/* invalid @json.stringify */", true
	}
	emitter := newJSONGoEmitter(g)
	value := emitter.valueExpr(call.Args[0])
	return fmt.Sprintf("func() string { %sb, _ := json.Marshal(%s); return string(b) }()", emitter.declarationText(), value), true
}

func newJSONGoEmitter(g *generator) *jsonGoEmitter {
	return &jsonGoEmitter{
		g:     g,
		names: map[string]string{},
	}
}

func (e *jsonGoEmitter) declarationText() string {
	if len(e.decls) == 0 {
		return ""
	}
	return strings.Join(e.decls, "; ") + "; "
}

func (e *jsonGoEmitter) valueExpr(expr ir.Expr) string {
	switch value := expr.(type) {
	case *ir.AnonymousObjectLiteral:
		return e.jsonObjectLiteral(value.Fields, value.ResultType())
	case *ir.StructLiteral:
		return e.jsonObjectLiteral(value.Fields, value.ResultType())
	case *ir.ArrayLiteral:
		if elem, ok := checker.ArrayElement(value.ResultType()); ok && e.jsonNeedsConversion(elem) && !jsonArrayLiteralHasSpread(value) {
			elems := make([]string, 0, len(value.Elements))
			for _, elem := range value.Elements {
				elems = append(elems, e.valueExpr(elem))
			}
			return e.jsonGoType(value.ResultType()) + "{" + strings.Join(elems, ", ") + "}"
		}
	}
	return e.jsonValueFromSource(e.g.expr(expr), expr.ResultType())
}

func (e *jsonGoEmitter) jsonValueFromSource(source string, typ checker.Type) string {
	if fields, ok := e.g.jsonObjectFields(typ); ok {
		jsonType := e.jsonGoType(typ)
		value := "v"
		return fmt.Sprintf("func() %s { %s := %s; return %s{%s} }()", jsonType, value, source, jsonType, e.jsonObjectValues(value, fields))
	}
	if elem, ok := checker.ArrayElement(typ); ok {
		if !e.jsonNeedsConversion(elem) {
			return source
		}
		jsonType := e.jsonGoType(typ)
		value := "v"
		out := "out"
		item := "item"
		return fmt.Sprintf("func() %s { %s := %s; %s := make(%s, 0, len(%s)); for _, %s := range %s { %s = append(%s, %s) }; return %s }()", jsonType, value, source, out, jsonType, value, item, value, out, out, e.jsonValueFromSource(item, elem), out)
	}
	return source
}

func (e *jsonGoEmitter) jsonGoType(typ checker.Type) string {
	if fields, ok := e.g.jsonObjectFields(typ); ok {
		key := string(typ)
		if name, ok := e.names[key]; ok {
			return name
		}
		name := fmt.Sprintf("json%d", e.next)
		e.next++
		e.names[key] = name
		parts := make([]string, 0, len(fields))
		for idx, field := range fields {
			parts = append(parts, fmt.Sprintf("F%d %s `json:%s`", idx, e.jsonGoType(field.typ), strconv.Quote(field.name)))
		}
		e.decls = append(e.decls, fmt.Sprintf("type %s struct{%s}", name, strings.Join(parts, "; ")))
		return name
	}
	if elem, ok := checker.ArrayElement(typ); ok {
		return "[]" + e.jsonGoType(elem)
	}
	return goType(typ)
}

func (e *jsonGoEmitter) jsonObjectValues(source string, fields []jsonObjectField) string {
	values := make([]string, 0, len(fields))
	for idx, field := range fields {
		fieldSource := source + "." + mangleIdent(field.name)
		values = append(values, fmt.Sprintf("F%d: %s", idx, e.jsonValueFromSource(fieldSource, field.typ)))
	}
	return strings.Join(values, ", ")
}

func (e *jsonGoEmitter) jsonObjectLiteral(fields []ir.FieldValue, typ checker.Type) string {
	jsonType := e.jsonGoType(typ)
	return fmt.Sprintf("%s{%s}", jsonType, e.jsonObjectLiteralValues(fields, typ))
}

func (e *jsonGoEmitter) jsonObjectLiteralValues(fields []ir.FieldValue, typ checker.Type) string {
	objectFields, ok := e.g.jsonObjectFields(typ)
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
		values = append(values, fmt.Sprintf("F%d: %s", idx, e.valueExpr(value)))
	}
	return strings.Join(values, ", ")
}

func (e *jsonGoEmitter) jsonNeedsConversion(typ checker.Type) bool {
	if _, ok := e.g.jsonObjectFields(typ); ok {
		return true
	}
	if elem, ok := checker.ArrayElement(typ); ok {
		return e.jsonNeedsConversion(elem)
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
