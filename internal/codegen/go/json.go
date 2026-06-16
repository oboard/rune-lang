package gocodegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
)

type jsonObjectField struct {
	sourceName string
	jsonName   string
	typ        checker.Type
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

func (g *generator) jsonParseCall(call *ir.CallExpr) (string, bool) {
	fn, ok := g.stdlibFunctionFromCall(call)
	if !ok || fn.Intrinsic != "json.parse" {
		return "", false
	}
	if len(call.Args) != 1 {
		return "/* invalid @json.parse */", true
	}
	emitter := newJSONGoEmitter(g)
	targetType := call.ResultType()
	wireType := emitter.decodeGoType(targetType)
	raw := "raw"
	seed := g.zeroValue(targetType)
	value := emitter.decodedValue(raw, seed, targetType)
	return fmt.Sprintf(
		"func() %s { %stype %sAlias = %s; var %s %sAlias; if err := json.Unmarshal([]byte(%s), &%s); err != nil { panic(err) }; return %s }()",
		goType(targetType),
		emitter.declarationText(),
		wireType,
		wireType,
		raw,
		wireType,
		g.expr(call.Args[0]),
		raw,
		value,
	), true
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
			parts = append(parts, fmt.Sprintf("F%d %s `json:%s`", idx, e.jsonGoType(field.typ), strconv.Quote(field.jsonName)))
		}
		e.decls = append(e.decls, fmt.Sprintf("type %s struct{%s}", name, strings.Join(parts, "; ")))
		return name
	}
	if elem, ok := checker.ArrayElement(typ); ok {
		return "[]" + e.jsonGoType(elem)
	}
	return goType(typ)
}

func (e *jsonGoEmitter) decodeGoType(typ checker.Type) string {
	if inner, ok := parseGoNullableType(string(typ)); ok {
		return "*" + e.decodeGoType(checker.Type(inner))
	}
	if fields, ok := e.g.jsonObjectFields(typ); ok {
		key := "decode:" + string(typ)
		if name, ok := e.names[key]; ok {
			return name
		}
		name := fmt.Sprintf("json%d", e.next)
		e.next++
		e.names[key] = name
		parts := make([]string, 0, len(fields))
		for idx, field := range fields {
			parts = append(parts, fmt.Sprintf("F%d %s `json:%s`", idx, e.decodeGoType(field.typ), strconv.Quote(field.jsonName)))
		}
		e.decls = append(e.decls, fmt.Sprintf("type %s struct{%s}", name, strings.Join(parts, "; ")))
		return name
	}
	if elem, ok := checker.ArrayElement(typ); ok {
		return "[]" + e.decodeGoType(elem)
	}
	if typ == checker.Char {
		return "string"
	}
	return goType(typ)
}

func (e *jsonGoEmitter) decodedValue(source string, seed string, typ checker.Type) string {
	if inner, ok := parseGoNullableType(string(typ)); ok {
		innerType := checker.Type(inner)
		return fmt.Sprintf(
			"func() any { if %s == nil { return nil }; return %s }()",
			source,
			e.decodedValue("*"+source, e.g.zeroValue(innerType), innerType),
		)
	}
	if fields, ok := e.g.jsonObjectFields(typ); ok {
		out := "out"
		assignments := make([]string, 0, len(fields))
		for idx, field := range fields {
			fieldName := mangleIdent(field.sourceName)
			assignments = append(assignments, fmt.Sprintf(
				"%s.%s = %s",
				out,
				fieldName,
				e.decodedValue(
					fmt.Sprintf("%s.F%d", source, idx),
					seed+"."+fieldName,
					field.typ,
				),
			))
		}
		body := ""
		if len(assignments) > 0 {
			body = strings.Join(assignments, "; ") + "; "
		}
		return fmt.Sprintf(
			"func() %s { %s := %s; %sreturn %s }()",
			goType(typ),
			out,
			seed,
			body,
			out,
		)
	}
	if elem, ok := checker.ArrayElement(typ); ok {
		index := "idx"
		item := "item"
		out := "out"
		if !e.decodeUsesSeed(elem) {
			return fmt.Sprintf(
				"func() %s { %s := make(%s, len(%s)); for %s, %s := range %s { %s[%s] = %s }; return %s }()",
				goType(typ),
				out,
				goType(typ),
				source,
				index,
				item,
				source,
				out,
				index,
				e.decodedValue(item, e.g.zeroValue(elem), elem),
				out,
			)
		}
		itemSeed := "itemSeed"
		return fmt.Sprintf(
			"func() %s { %s := make(%s, len(%s)); for %s, %s := range %s { var %s %s; if %s < len(%s) { %s = %s[%s] }; %s[%s] = %s }; return %s }()",
			goType(typ),
			out,
			goType(typ),
			source,
			index,
			item,
			source,
			itemSeed,
			goType(elem),
			index,
			seed,
			itemSeed,
			seed,
			index,
			out,
			index,
			e.decodedValue(item, itemSeed, elem),
			out,
		)
	}
	if typ == checker.Char {
		return "[]rune(" + source + ")[0]"
	}
	return source
}

func (e *jsonGoEmitter) decodeUsesSeed(typ checker.Type) bool {
	if inner, ok := parseGoNullableType(string(typ)); ok {
		return e.decodeUsesSeed(checker.Type(inner))
	}
	if _, ok := e.g.jsonObjectFields(typ); ok {
		return true
	}
	if elem, ok := checker.ArrayElement(typ); ok {
		return e.decodeUsesSeed(elem)
	}
	return false
}

func (e *jsonGoEmitter) jsonObjectValues(source string, fields []jsonObjectField) string {
	values := make([]string, 0, len(fields))
	for idx, field := range fields {
		fieldSource := source + "." + mangleIdent(field.sourceName)
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
		value := byName[field.sourceName]
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
			out = append(out, jsonObjectField{sourceName: field.name, jsonName: field.name, typ: fieldType})
		}
		return out, true
	}
	for _, candidate := range g.file.Types {
		if candidate.Name != string(typ) {
			continue
		}
		out := make([]jsonObjectField, 0, len(candidate.Fields))
		for _, field := range candidate.Fields {
			if jsonOmitType(field.Type) || (candidate.JSONObject && field.JSONIgnore) {
				continue
			}
			jsonName := field.Name
			if candidate.JSONObject {
				jsonName = field.JSONName
			}
			out = append(out, jsonObjectField{sourceName: field.Name, jsonName: jsonName, typ: field.Type})
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
