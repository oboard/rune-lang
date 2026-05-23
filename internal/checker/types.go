package checker

import (
	"fmt"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

func (c *checker) resolveType(name string) Type {
	return c.resolveTypeWithGenerics(name, nil)
}

func (c *checker) resolveTypeWithGenerics(name string, generics map[string]bool) Type {
	switch name {
	case "Int":
		return Int
	case "Double":
		return Double
	case "BigInt":
		return BigInt
	case "String":
		return String
	case "Bool":
		return Bool
	case "Null":
		return Null
	case "Object":
		return Object
	case "Never":
		return Never
	case "Symbol":
		return Symbol
	case "Void":
		return Void
	case "HTMLElement":
		return HTMLElement
	default:
		if isDynamicTypeName(name) {
			return Unknown
		}
		if generics[name] {
			return Type(name)
		}
		if innerName, ok := parseNullableType(name); ok {
			inner := c.resolveTypeWithGenerics(innerName, generics)
			if inner == Unknown {
				return Unknown
			}
			return NullableOf(inner)
		}
		if elemName, ok := parseArrayType(name); ok {
			elem := c.resolveTypeWithGenerics(elemName, generics)
			if elem == Unknown {
				return Unknown
			}
			return ArrayOf(elem)
		}
		if base, args, ok := parseGenericType(name); ok && isBuiltinGenericType(base) {
			resolved := make([]Type, 0, len(args))
			for _, arg := range args {
				typ := c.resolveTypeWithGenerics(arg, generics)
				if typ == Unknown && !isDynamicTypeName(arg) {
					return Unknown
				}
				resolved = append(resolved, typ)
			}
			return genericTypeOf(base, resolved)
		}
		if params, ret, ok := parseFuncType(name); ok {
			types := make([]Type, 0, len(params)+1)
			for _, param := range params {
				typ := c.resolveTypeWithGenerics(param, generics)
				if typ == Unknown && !isDynamicTypeName(param) {
					return Unknown
				}
				types = append(types, typ)
			}
			retType := c.resolveTypeWithGenerics(ret, generics)
			if retType == Unknown && !isDynamicTypeName(ret) {
				return Unknown
			}
			return funcTypeOf(types, retType)
		}
		if _, ok := c.info.Types[name]; ok {
			return Type(name)
		}
		return Unknown
	}
}

func (c *checker) resolveDeclaredReturn(name string) Type {
	if isDynamicTypeName(name) {
		return Unknown
	}
	return c.resolveType(name)
}

func ArrayOf(elem Type) Type {
	return Type("Array[" + string(elem) + "]")
}

func genericTypeOf(base string, args []Type) Type {
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		parts = append(parts, string(arg))
	}
	return Type(base + "[" + strings.Join(parts, ",") + "]")
}

func NullableOf(elem Type) Type {
	if elem == Null {
		return Null
	}
	return Type(string(elem) + "?")
}

func FuncOf(arg Type, ret Type) Type {
	return Type("Func[" + string(arg) + "," + string(ret) + "]")
}

func FuncOfTypes(params []Type, ret Type) Type {
	parts := make([]string, 0, len(params)+1)
	for _, param := range params {
		parts = append(parts, string(param))
	}
	parts = append(parts, string(ret))
	return Type("Func[" + strings.Join(parts, ",") + "]")
}

func DisplayType(typ Type) string {
	return displayTypeName(string(typ))
}

func displayTypeName(name string) string {
	if params, ret, ok := parseFuncType(name); ok {
		for i, param := range params {
			params[i] = displayTypeName(param)
		}
		return displayFuncType(params, displayTypeName(ret))
	}
	if strings.HasPrefix(name, "{") && strings.HasSuffix(name, "}") {
		return displayObjectTypeName(name)
	}
	if inner, ok := parseNullableType(name); ok {
		return displayTypeName(inner) + "?"
	}
	return name
}

func displayFuncType(params []string, ret string) string {
	switch len(params) {
	case 0:
		return "() -> " + ret
	case 1:
		return params[0] + " -> " + ret
	default:
		return "(" + strings.Join(params, ", ") + ") -> " + ret
	}
}

func displayObjectTypeName(name string) string {
	inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(name, "{"), "}"))
	if inner == "" {
		return "{}"
	}
	fields := make([]string, 0)
	for _, part := range splitTypeList(inner) {
		idx := strings.Index(part, ":")
		if idx < 0 {
			fields = append(fields, strings.TrimSpace(part))
			continue
		}
		fieldName := strings.TrimSpace(part[:idx])
		fieldType := displayTypeName(strings.TrimSpace(part[idx+1:]))
		fields = append(fields, fieldName+": "+fieldType)
	}
	return "{" + strings.Join(fields, ", ") + "}"
}

func ArrayElement(typ Type) (Type, bool) {
	name := string(typ)
	elem, ok := parseArrayType(name)
	if !ok {
		return Unknown, false
	}
	return Type(elem), true
}

func receiverType(typ *ast.StructType) Type {
	if len(typ.Generics) == 0 {
		return Type(typ.Name)
	}
	return Type(typ.Name + "[" + strings.Join(typ.Generics, ",") + "]")
}

func ObjectOf(fields []FieldInfo) Type {
	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		parts = append(parts, field.Name+": "+string(field.Type))
	}
	return Type("{" + strings.Join(parts, ", ") + "}")
}

func isObjectType(typ Type) bool {
	name := string(typ)
	return strings.HasPrefix(name, "{") && strings.HasSuffix(name, "}")
}

func baseTypeName(typ Type) string {
	if isObjectType(typ) {
		return string(typ)
	}
	name := string(typ)
	if i := strings.IndexByte(name, '['); i >= 0 {
		return name[:i]
	}
	return name
}

func genericSet(names ...string) map[string]bool {
	out := make(map[string]bool, len(names))
	for _, name := range names {
		if name != "" {
			out[name] = true
		}
	}
	return out
}

func isDynamicTypeName(name string) bool {
	return name == "Dynamic" || name == "Any"
}

func typesCompatible(expected Type, actual Type, generics []string) bool {
	return typesCompatibleWithSet(expected, actual, genericSet(generics...))
}

func typesCompatibleWithSet(expected Type, actual Type, generics map[string]bool) bool {
	if expected == Unknown || actual == Unknown || expected == actual {
		return true
	}
	if actual == Never {
		return true
	}
	if expectedInner, ok := parseNullableType(string(expected)); ok {
		if actual == Null {
			return true
		}
		if actualInner, actualNullable := parseNullableType(string(actual)); actualNullable {
			return typesCompatibleWithSet(Type(expectedInner), Type(actualInner), generics)
		}
		return typesCompatibleWithSet(Type(expectedInner), actual, generics)
	}
	if _, ok := parseNullableType(string(actual)); ok {
		return false
	}
	if expected == Object {
		return isObjectLike(actual)
	}
	if generics[string(expected)] || generics[string(actual)] {
		return true
	}
	if expectedElem, ok := parseArrayType(string(expected)); ok {
		actualElem, actualArray := parseArrayType(string(actual))
		if !actualArray {
			return false
		}
		return typesCompatibleWithSet(Type(expectedElem), Type(actualElem), generics)
	}
	if expectedBase, expectedArgs, ok := parseGenericType(string(expected)); ok && isBuiltinGenericType(expectedBase) {
		actualBase, actualArgs, actualGeneric := parseGenericType(string(actual))
		if !actualGeneric || expectedBase != actualBase || len(expectedArgs) != len(actualArgs) {
			return false
		}
		for i, expectedArg := range expectedArgs {
			if !typesCompatibleWithSet(Type(expectedArg), Type(actualArgs[i]), generics) {
				return false
			}
		}
		return true
	}
	if expectedParams, expectedRet, ok := parseFuncType(string(expected)); ok {
		actualParams, actualRet, actualFunc := parseFuncType(string(actual))
		if !actualFunc || len(expectedParams) != len(actualParams) {
			return false
		}
		for i, expectedParam := range expectedParams {
			if !typesCompatibleWithSet(Type(expectedParam), Type(actualParams[i]), generics) {
				return false
			}
		}
		return typesCompatibleWithSet(Type(expectedRet), Type(actualRet), generics)
	}
	if isObjectType(expected) && isObjectType(actual) {
		return objectTypesCompatible(expected, actual)
	}
	return false
}

func isObjectLike(typ Type) bool {
	if isObjectType(typ) {
		return true
	}
	switch baseTypeName(typ) {
	case "Array", "Map", "Set", "WeakMap", "WeakSet", "Record", "Tuple", "ReadonlyArray", "ReadonlyTuple":
		return true
	case string(Int), string(Double), string(BigInt), string(String), string(Bool), string(Null), string(Void), string(Symbol):
		return false
	default:
		return typ != Unknown && typ != Never
	}
}

func objectTypesCompatible(expected Type, actual Type) bool {
	expectedFields := parseObjectFields(string(expected))
	actualFields := parseObjectFields(string(actual))
	if len(expectedFields) == 0 || len(actualFields) == 0 {
		return false
	}
	if len(expectedFields) != len(actualFields) {
		return false
	}
	for name, expectedType := range expectedFields {
		actualType, ok := actualFields[name]
		if !ok {
			return false
		}
		if !typesCompatible(expectedType, actualType, nil) {
			return false
		}
	}
	return true
}

func (c *checker) unifyTypes(left Type, right Type) (Type, bool) {
	if left == Unknown {
		return right, true
	}
	if right == Unknown || left == right {
		return left, true
	}
	if left == Null {
		if right == Void {
			return Unknown, false
		}
		return NullableOf(right), true
	}
	if right == Null {
		if left == Void {
			return Unknown, false
		}
		return NullableOf(left), true
	}
	if leftInner, ok := parseNullableType(string(left)); ok {
		if right == Null {
			return left, true
		}
		if rightInner, rightNullable := parseNullableType(string(right)); rightNullable {
			elem, ok := c.unifyTypes(Type(leftInner), Type(rightInner))
			if !ok {
				return Unknown, false
			}
			return NullableOf(elem), true
		}
		elem, ok := c.unifyTypes(Type(leftInner), right)
		if !ok {
			return Unknown, false
		}
		return NullableOf(elem), true
	}
	if rightInner, ok := parseNullableType(string(right)); ok {
		if left == Null {
			return right, true
		}
		elem, ok := c.unifyTypes(left, Type(rightInner))
		if !ok {
			return Unknown, false
		}
		return NullableOf(elem), true
	}
	if leftElem, ok := parseArrayType(string(left)); ok {
		rightElem, rightArray := parseArrayType(string(right))
		if !rightArray {
			return Unknown, false
		}
		elem, ok := c.unifyTypes(Type(leftElem), Type(rightElem))
		if !ok {
			return Unknown, false
		}
		return ArrayOf(elem), true
	}
	if leftBase, leftArgs, ok := parseGenericType(string(left)); ok && isBuiltinGenericType(leftBase) {
		rightBase, rightArgs, rightGeneric := parseGenericType(string(right))
		if !rightGeneric || leftBase != rightBase || len(leftArgs) != len(rightArgs) {
			return Unknown, false
		}
		args := make([]Type, 0, len(leftArgs))
		for i := range leftArgs {
			arg, ok := c.unifyTypes(Type(leftArgs[i]), Type(rightArgs[i]))
			if !ok {
				return Unknown, false
			}
			args = append(args, arg)
		}
		return genericTypeOf(leftBase, args), true
	}
	if leftParams, leftRet, ok := parseFuncType(string(left)); ok {
		rightParams, rightRet, rightFunc := parseFuncType(string(right))
		if !rightFunc || len(leftParams) != len(rightParams) {
			return Unknown, false
		}
		params := make([]Type, 0, len(leftParams))
		for i := range leftParams {
			param, ok := c.unifyTypes(Type(leftParams[i]), Type(rightParams[i]))
			if !ok {
				return Unknown, false
			}
			params = append(params, param)
		}
		ret, ok := c.unifyTypes(Type(leftRet), Type(rightRet))
		if !ok {
			return Unknown, false
		}
		return FuncOfTypes(params, ret), true
	}
	if isObjectType(left) && isObjectType(right) {
		return c.unifyObjectTypes(left, right)
	}
	return Unknown, false
}

func (c *checker) unifyObjectTypes(left Type, right Type) (Type, bool) {
	leftInfo := c.info.Types[baseTypeName(left)]
	rightInfo := c.info.Types[baseTypeName(right)]
	if leftInfo == nil || rightInfo == nil || len(leftInfo.Fields) != len(rightInfo.Fields) {
		return Unknown, false
	}
	fields := make([]FieldInfo, 0, len(leftInfo.Fields))
	for _, leftField := range leftInfo.Fields {
		rightField, ok := rightInfo.ByName[leftField.Name]
		if !ok {
			return Unknown, false
		}
		fieldType, ok := c.unifyTypes(leftField.Type, rightField.Type)
		if !ok {
			return Unknown, false
		}
		fields = append(fields, FieldInfo{Name: leftField.Name, Type: fieldType})
	}
	typ := ObjectOf(fields)
	byName := map[string]FieldInfo{}
	for _, field := range fields {
		byName[field.Name] = field
	}
	c.registerAnonymousObjectType(typ, fields, byName)
	return typ, true
}

func (c *checker) objectHasFields(actual Type, expected Type) bool {
	actualInfo := c.info.Types[baseTypeName(actual)]
	expectedInfo := c.info.Types[baseTypeName(expected)]
	if actualInfo == nil || expectedInfo == nil {
		return false
	}
	for _, expectedField := range expectedInfo.Fields {
		actualField, ok := actualInfo.ByName[expectedField.Name]
		if !ok {
			return false
		}
		if _, ok := c.unifyTypes(expectedField.Type, actualField.Type); !ok {
			return false
		}
	}
	return true
}

func parseObjectFields(name string) map[string]Type {
	out := map[string]Type{}
	if !strings.HasPrefix(name, "{") || !strings.HasSuffix(name, "}") {
		return out
	}
	inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(name, "{"), "}"))
	if inner == "" {
		return out
	}
	for _, part := range splitTypeList(inner) {
		idx := strings.Index(part, ":")
		if idx < 0 {
			continue
		}
		out[strings.TrimSpace(part[:idx])] = Type(strings.TrimSpace(part[idx+1:]))
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func cloneEnv(env map[string]Type) map[string]Type {
	out := make(map[string]Type, len(env))
	for k, v := range env {
		out[k] = v
	}
	return out
}

func (c *checker) errorf(pos lexer.Position, format string, args ...any) {
	c.diags = append(c.diags, Diagnostic{Message: fmt.Sprintf(format, displayArgs(args)...), Pos: pos})
}

func displayArgs(args []any) []any {
	out := make([]any, 0, len(args))
	for _, arg := range args {
		if typ, ok := arg.(Type); ok {
			out = append(out, DisplayType(typ))
			continue
		}
		out = append(out, arg)
	}
	return out
}
