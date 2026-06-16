package tscodegen

import (
	"fmt"
	"strings"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
)

type jsonObjectField struct {
	sourceName string
	jsonName   string
	typ        checker.Type
}

func (g *generator) jsonParseCall(call *ir.CallExpr) (string, bool) {
	fn, ok := g.stdlibFunctionFromCall(call)
	if !ok || fn.Intrinsic != "json.parse" {
		return "", false
	}
	if len(call.Args) != 1 {
		return "undefined", true
	}
	raw := mangleIdent("rune_json_raw")
	return fmt.Sprintf(
		"((%s: any): %s => %s)(JSON.parse(%s))",
		raw,
		tsType(call.ResultType()),
		g.jsonDecodedValue(raw, g.jsonDecodeZeroValue(call.ResultType()), call.ResultType()),
		g.expr(call.Args[0]),
	), true
}

func (g *generator) jsonDecodeZeroValue(typ checker.Type) string {
	if _, ok := parseTSNullableType(string(typ)); ok {
		return "null"
	}
	if _, ok := checker.ArrayElement(typ); ok {
		return "[]"
	}
	for _, candidate := range g.file.Types {
		if candidate.Name != string(typ) {
			continue
		}
		values := make([]string, 0, len(candidate.Fields))
		for _, field := range candidate.Fields {
			values = append(values, fmt.Sprintf(
				"%s: %s",
				tsPropertyName(field.Name),
				g.jsonDecodeZeroValue(field.Type),
			))
		}
		return "({ " + strings.Join(values, ", ") + " })"
	}
	return g.zeroValue(typ)
}

func (g *generator) jsonDecodedValue(source string, seed string, typ checker.Type) string {
	if inner, ok := parseTSNullableType(string(typ)); ok {
		innerType := checker.Type(inner)
		return fmt.Sprintf(
			"(%s === null ? null : %s)",
			source,
			g.jsonDecodedValue(source, g.zeroValue(innerType), innerType),
		)
	}
	if fields, ok := g.jsonObjectFields(typ); ok {
		values := make([]string, 0, len(fields)+1)
		values = append(values, "..."+seed)
		for _, field := range fields {
			values = append(values, fmt.Sprintf(
				"%s: %s",
				tsPropertyName(field.sourceName),
				g.jsonDecodedValue(
					fmt.Sprintf("%s[%q]", source, field.jsonName),
					tsPropertyAccess(seed, field.sourceName),
					field.typ,
				),
			))
		}
		return "({ " + strings.Join(values, ", ") + " })"
	}
	if elem, ok := checker.ArrayElement(typ); ok {
		item := mangleIdent("rune_json_item")
		index := mangleIdent("rune_json_index")
		itemSeed := fmt.Sprintf("(%s[%s] ?? %s)", seed, index, g.zeroValue(elem))
		return fmt.Sprintf(
			"%s.map((%s: any, %s: number) => %s)",
			source,
			item,
			index,
			g.jsonDecodedValue(item, itemSeed, elem),
		)
	}
	return source
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
			return "[" + strings.Join(elems, ", ") + "]"
		}
	}
	return g.jsonValueFromSource(g.expr(expr), expr.ResultType())
}

func (g *generator) jsonValueFromSource(source string, typ checker.Type) string {
	if fields, ok := g.jsonObjectFields(typ); ok {
		value := mangleIdent("rune_json_value")
		return fmt.Sprintf("((%s) => ({ %s }))(%s)", value, g.jsonObjectValues(value, fields), source)
	}
	if elem, ok := checker.ArrayElement(typ); ok {
		if !g.jsonNeedsConversion(elem) {
			return source
		}
		item := mangleIdent("rune_json_item")
		return fmt.Sprintf("%s.map((%s) => %s)", source, item, g.jsonValueFromSource(item, elem))
	}
	return source
}

func (g *generator) jsonObjectValues(source string, fields []jsonObjectField) string {
	values := make([]string, 0, len(fields))
	for _, field := range fields {
		fieldSource := tsPropertyAccess(source, field.sourceName)
		values = append(values, fmt.Sprintf("%s: %s", tsPropertyName(field.jsonName), g.jsonValueFromSource(fieldSource, field.typ)))
	}
	return strings.Join(values, ", ")
}

func (g *generator) jsonObjectLiteral(fields []ir.FieldValue, typ checker.Type) string {
	return "({ " + g.jsonObjectLiteralValues(fields, typ) + " })"
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
	for _, field := range objectFields {
		value := byName[field.sourceName]
		if value == nil {
			continue
		}
		values = append(values, fmt.Sprintf("%s: %s", tsPropertyName(field.jsonName), g.jsonValueExpr(value)))
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
	fields, ok := parseTSObjectType(string(typ))
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
	_, _, ok := parseTSFuncType(string(typ))
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
