package moonbitcodegen

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
	ignore     bool
}

func (g *generator) jsonStringifyCall(call *ir.CallExpr) (string, bool) {
	if len(call.Args) != 1 {
		return quoteString("null"), true
	}
	return g.jsonValueExpr(call.Args[0]) + ".stringify()", true
}

func (g *generator) jsonParseCall(call *ir.CallExpr) (string, bool) {
	if len(call.Args) != 1 {
		return zeroValue(call.ResultType()), true
	}
	raw := g.nextTemp("__json")
	target := call.ResultType()
	return fmt.Sprintf(
		"try { let %s = @json.parse(%s); %s } catch { _ => %s }",
		raw,
		g.expr(call.Args[0]),
		g.jsonDecodedValue(raw, g.jsonDecodeZeroValue(target), target),
		g.jsonDecodeZeroValue(target),
	), true
}

func (g *generator) jsonValueExpr(expr ir.Expr) string {
	switch e := expr.(type) {
	case *ir.AnonymousObjectLiteral:
		return g.jsonObjectLiteral(e.Fields, e.ResultType())
	case *ir.StructLiteral:
		return g.jsonObjectLiteral(e.Fields, e.ResultType())
	case *ir.ArrayLiteral:
		if jsonArrayLiteralHasSpread(e) {
			return g.jsonValueFromSource(g.expr(expr), expr.ResultType())
		}
		elems := make([]string, 0, len(e.Elements))
		for _, elem := range e.Elements {
			elems = append(elems, g.jsonValueExpr(elem))
		}
		return "Json::array([" + strings.Join(elems, ", ") + "])"
	}
	return g.jsonValueFromSource(g.expr(expr), expr.ResultType())
}

func (g *generator) jsonValueFromSource(source string, typ checker.Type) string {
	if mbtJSONValueType(typ) {
		return source
	}
	if inner, ok := parseNullableType(string(typ)); ok {
		value := g.nextTemp("__json_value")
		return fmt.Sprintf("match %s { Some(%s) => %s; None => Json::null() }", source, value, g.jsonValueFromSource(value, checker.Type(inner)))
	}
	if fields, ok := g.jsonEncodeObjectFields(typ); ok {
		value := g.nextTemp("__json_value")
		return fmt.Sprintf("{ let %s = %s; Json::object({ %s }) }", value, source, g.jsonObjectValues(value, fields))
	}
	if elem, ok := checker.ArrayElement(typ); ok {
		item := g.nextTemp("__json_item")
		return fmt.Sprintf("Json::array(%s.map(fn(%s) { %s }))", source, item, g.jsonValueFromSource(item, elem))
	}
	switch typ {
	case checker.String:
		return fmt.Sprintf("Json::string(%s)", source)
	case checker.Char:
		return fmt.Sprintf("Json::string((%s).to_string())", source)
	case checker.Int, checker.Int4, checker.Int8, checker.Int16, checker.UInt, checker.UInt8, checker.UInt16:
		return fmt.Sprintf("Json::number((%s).to_double())", source)
	case checker.Int64, checker.UInt64, checker.BigInt:
		return fmt.Sprintf("Json::string((%s).to_string())", source)
	case checker.Double:
		return fmt.Sprintf("Json::number(%s)", source)
	case checker.Float:
		return fmt.Sprintf("Json::number((%s).to_double())", source)
	case checker.Bool:
		return fmt.Sprintf("Json::boolean(%s)", source)
	case checker.Null, checker.Void:
		return "Json::null()"
	default:
		return "Json::null()"
	}
}

func (g *generator) jsonObjectLiteral(fields []ir.FieldValue, typ checker.Type) string {
	return "Json::object({ " + g.jsonObjectLiteralValues(fields, typ) + " })"
}

func (g *generator) jsonObjectLiteralValues(fields []ir.FieldValue, typ checker.Type) string {
	objectFields, ok := g.jsonEncodeObjectFields(typ)
	if !ok {
		objectFields = make([]jsonObjectField, 0, len(fields))
		for _, field := range fields {
			if _, ok := field.Value.(*ir.LambdaExpr); ok {
				continue
			}
			if jsonOmitType(field.Value.ResultType()) {
				continue
			}
			objectFields = append(objectFields, jsonObjectField{
				sourceName: field.Name,
				jsonName:   field.Name,
				typ:        field.Value.ResultType(),
			})
		}
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
		values = append(values, fmt.Sprintf("%s: %s", quoteString(field.jsonName), g.jsonValueExpr(value)))
	}
	return strings.Join(values, ", ")
}

func (g *generator) jsonObjectValues(source string, fields []jsonObjectField) string {
	values := make([]string, 0, len(fields))
	for _, field := range fields {
		fieldSource := source + "." + mangleIdent(field.sourceName)
		values = append(values, fmt.Sprintf("%s: %s", quoteString(field.jsonName), g.jsonValueFromSource(fieldSource, field.typ)))
	}
	return strings.Join(values, ", ")
}

func (g *generator) jsonDecodeZeroValue(typ checker.Type) string {
	if _, ok := parseNullableType(string(typ)); ok {
		return "None"
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
			values = append(values, fmt.Sprintf("%s: %s", mangleIdent(field.Name), g.jsonDecodeZeroValue(field.Type)))
		}
		return fmt.Sprintf("%s::{ %s }", mangleType(candidate.Name), strings.Join(values, ", "))
	}
	return zeroValue(typ)
}

func (g *generator) jsonDecodedValue(source string, seed string, typ checker.Type) string {
	if mbtJSONValueType(typ) {
		return source
	}
	if inner, ok := parseNullableType(string(typ)); ok {
		innerType := checker.Type(inner)
		return fmt.Sprintf("match %s { Null => None; _ => Some(%s) }", source, g.jsonDecodedValue(source, g.jsonDecodeZeroValue(innerType), innerType))
	}
	if fields, ok := g.jsonDecodeObjectFields(typ); ok {
		obj := g.nextTemp("__json_obj")
		values := make([]string, 0, len(fields))
		for _, field := range fields {
			seedField := fmt.Sprintf("%s.%s", seed, mangleIdent(field.sourceName))
			if field.ignore {
				values = append(values, fmt.Sprintf("%s: %s", mangleIdent(field.sourceName), seedField))
				continue
			}
			values = append(values, fmt.Sprintf(
				"%s: %s",
				mangleIdent(field.sourceName),
				g.jsonDecodedValue(
					fmt.Sprintf("%s.get(%s).unwrap_or(Json::null())", obj, quoteString(field.jsonName)),
					seedField,
					field.typ,
				),
			))
		}
		return fmt.Sprintf("match %s { Object(%s) => %s::{ %s }; _ => %s }", source, obj, mangleType(string(typ)), strings.Join(values, ", "), seed)
	}
	if elem, ok := checker.ArrayElement(typ); ok {
		items := g.nextTemp("__json_items")
		item := g.nextTemp("__json_item")
		return fmt.Sprintf("match %s { Array(%s) => %s.map(fn(%s) { %s }); _ => %s }", source, items, items, item, g.jsonDecodedValue(item, g.jsonDecodeZeroValue(elem), elem), seed)
	}
	switch typ {
	case checker.String:
		return fmt.Sprintf("match %s { String(value) => value; _ => %s }", source, seed)
	case checker.Char:
		return fmt.Sprintf("match %s { String(value) if value.length() > 0 => value[0].unsafe_to_char(); _ => %s }", source, seed)
	case checker.Int, checker.Int4, checker.Int8, checker.Int16:
		return fmt.Sprintf("match %s { Number(value, ..) => value.to_int(); _ => %s }", source, seed)
	case checker.UInt, checker.UInt8, checker.UInt16:
		return fmt.Sprintf("match %s { Number(value, ..) => value.to_int().to_uint(); _ => %s }", source, seed)
	case checker.Double:
		return fmt.Sprintf("match %s { Number(value, ..) => value; _ => %s }", source, seed)
	case checker.Float:
		return fmt.Sprintf("match %s { Number(value, ..) => Float::from_double(value); _ => %s }", source, seed)
	case checker.Bool:
		return fmt.Sprintf("match %s { True => true; False => false; _ => %s }", source, seed)
	default:
		return seed
	}
}

func (g *generator) jsonEncodeObjectFields(typ checker.Type) ([]jsonObjectField, bool) {
	if fields, ok := parseObjectType(string(typ)); ok {
		out := make([]jsonObjectField, 0, len(fields))
		for _, field := range fields {
			if jsonOmitType(checker.Type(field.typ)) {
				continue
			}
			out = append(out, jsonObjectField{sourceName: field.name, jsonName: field.name, typ: checker.Type(field.typ)})
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

func (g *generator) jsonDecodeObjectFields(typ checker.Type) ([]jsonObjectField, bool) {
	for _, candidate := range g.file.Types {
		if candidate.Name != string(typ) {
			continue
		}
		out := make([]jsonObjectField, 0, len(candidate.Fields))
		for _, field := range candidate.Fields {
			if jsonOmitType(field.Type) {
				continue
			}
			jsonName := field.Name
			if candidate.JSONObject {
				jsonName = field.JSONName
			}
			out = append(out, jsonObjectField{
				sourceName: field.Name,
				jsonName:   jsonName,
				typ:        field.Type,
				ignore:     candidate.JSONObject && field.JSONIgnore,
			})
		}
		return out, true
	}
	return nil, false
}

func jsonOmitType(typ checker.Type) bool {
	if typ == checker.Void || typ == checker.Symbol {
		return true
	}
	_, _, ok := parseFuncType(string(typ))
	if ok {
		return true
	}
	base, _, ok := parseGenericType(string(typ))
	return ok && (base == "Func" || base == "AsyncFunc")
}

func jsonArrayLiteralHasSpread(lit *ir.ArrayLiteral) bool {
	for _, elem := range lit.Elements {
		if _, ok := elem.(*ir.SpreadExpr); ok {
			return true
		}
	}
	return false
}
