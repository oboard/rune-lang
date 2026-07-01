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
	case "Int4":
		return Int4
	case "Int8":
		return Int8
	case "Int16":
		return Int16
	case "Int64":
		return Int64
	case "Double":
		return Double
	case "Float":
		return Float
	case "BigInt":
		return BigInt
	case "UInt":
		return UInt
	case "UInt8":
		return UInt8
	case "UInt16":
		return UInt16
	case "UInt64":
		return UInt64
	case "String":
		return String
	case "Char":
		return Char
	case "Bool":
		return Bool
	case "Null":
		return Null
	case "Object":
		return Object
	case "Bytes":
		return Bytes
	case "Buffer":
		return Buffer
	case "Reader":
		return Reader
	case "Writer":
		return Writer
	case "StringBuffer":
		return StringBuffer
	case "FileStat":
		return FileStat
	case "TCPConnection":
		return TCPConnection
	case "TCPListener":
		return TCPListener
	case "Never":
		return Never
	case "Symbol":
		return Symbol
	case "Regex":
		return Regex
	case "Void":
		return Void
	case "HTMLElement":
		return HTMLElement
	case "WebComponent":
		return WebComponent
	default:
		if strings.HasPrefix(name, "&") {
			traitName := strings.TrimPrefix(name, "&")
			if c.info.Traits[traitName] != nil {
				return Type(name)
			}
			return Unknown
		}
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
		if base, args, ok := parseGenericType(name); ok && (isBuiltinGenericType(base) || c.coreTypeExists(base) || c.coreEnumExists(base)) {
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
		if info, ok := c.info.Types[name]; ok {
			if !c.canAccessPrivate(info.Private, info.SourcePath) {
				return Unknown
			}
			return Type(name)
		}
		if info, ok := c.info.Enums[name]; ok {
			if !c.canAccessPrivate(info.Private, info.SourcePath) {
				return Unknown
			}
			return Type(name)
		}
		return Unknown
	}
}

func (c *checker) coreTypeExists(name string) bool {
	return c != nil && c.info != nil && c.info.Types[name] != nil
}

func (c *checker) typesCompatible(expected Type, actual Type, generics []string) bool {
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
			return c.typesCompatible(Type(expectedInner), Type(actualInner), generics)
		}
		return c.typesCompatible(Type(expectedInner), actual, generics)
	}
	if _, ok := parseNullableType(string(actual)); ok {
		return false
	}
	if expected == Object {
		return isObjectLike(actual)
	}
	genericTypes := genericSet(generics...)
	if genericTypes[string(expected)] || genericTypes[string(actual)] {
		return true
	}
	if expectedElem, ok := parseArrayType(string(expected)); ok {
		actualElem, actualArray := parseArrayType(string(actual))
		if !actualArray {
			return false
		}
		return c.typesCompatible(Type(expectedElem), Type(actualElem), generics)
	}
	if expectedBase, expectedArgs, ok := parseGenericType(string(expected)); ok && isBuiltinGenericType(expectedBase) {
		actualBase, actualArgs, actualGeneric := parseGenericType(string(actual))
		if !actualGeneric || expectedBase != actualBase || len(expectedArgs) != len(actualArgs) {
			return false
		}
		for i, expectedArg := range expectedArgs {
			if !c.typesCompatible(Type(expectedArg), Type(actualArgs[i]), generics) {
				return false
			}
		}
		return true
	}
	if expectedParams, expectedRet, ok := parseCallableType(string(expected)); ok {
		actualParams, actualRet, actualFunc := parseCallableType(string(actual))
		if !actualFunc || len(expectedParams) != len(actualParams) {
			return false
		}
		for i, expectedParam := range expectedParams {
			if !c.typesCompatible(Type(expectedParam), Type(actualParams[i]), generics) {
				return false
			}
		}
		return c.typesCompatible(Type(expectedRet), Type(actualRet), generics)
	}
	if strings.HasPrefix(string(expected), "&") {
		return c.typeImplementsTrait(actual, strings.TrimPrefix(string(expected), "&"))
	}
	if isObjectType(expected) && isObjectType(actual) {
		return objectTypesCompatible(expected, actual)
	}
	return false
}

func (c *checker) typeImplementsTrait(actual Type, traitName string) bool {
	if actual == Unknown {
		return true
	}
	trait := c.info.Traits[traitName]
	if trait == nil {
		return false
	}
	if actual == Type("&"+traitName) {
		return true
	}
	object := c.info.Types[baseTypeName(actual)]
	if object == nil {
		return false
	}
	if traitName == "FromJson" && object.Node != nil &&
		hasJSONAnnotation(object.Node.Annotations, "object") &&
		object.Methods["fromJson"] == nil &&
		object.StaticMethods["fromJson"] == nil {
		return true
	}
	for _, required := range trait.Fields {
		field, ok := object.ByName[required.Name]
		if !ok || field.Private {
			return false
		}
		expected := substituteSelfType(required.Type, actual)
		if !c.typesCompatible(expected, structFieldType(object, actual, field), nil) {
			return false
		}
	}
	for name, required := range trait.Methods {
		actualMethod := object.Methods[name]
		if actualMethod == nil {
			if field, ok := object.ByName[name]; ok && !field.Private {
				expected := traitMethodType(required, actual)
				if !c.typesCompatible(expected, structFieldType(object, actual, field), nil) {
					return false
				}
				continue
			}
			return false
		}
		if actualMethod.Private || actualMethod.Routine != required.Routine || len(actualMethod.Params) != len(required.Params) {
			return false
		}
		for idx, param := range required.Params {
			if !c.typesCompatible(substituteSelfType(param.Type, actual), actualMethod.Params[idx].Type, nil) {
				return false
			}
		}
		if !c.typesCompatible(substituteSelfType(required.Return, actual), actualMethod.Return, nil) {
			return false
		}
	}
	for name, required := range trait.StaticMethods {
		actualMethod := object.StaticMethods[name]
		if actualMethod == nil || actualMethod.Private || actualMethod.Routine != required.Routine || len(actualMethod.Params) != len(required.Params) {
			return false
		}
		for idx, param := range required.Params {
			if !c.typesCompatible(substituteSelfType(param.Type, actual), actualMethod.Params[idx].Type, nil) {
				return false
			}
		}
		if !c.typesCompatible(substituteSelfType(required.Return, actual), actualMethod.Return, nil) {
			return false
		}
	}
	return true
}

func genericConstraintName(typ ast.Type) string {
	name := typ.Canonical()
	name = strings.TrimPrefix(name, "&")
	return name
}

func (c *checker) genericConstraintExists(name string) bool {
	if name == "" {
		return false
	}
	return c.info != nil && c.info.Traits[name] != nil
}

func (c *checker) genericConstraintSatisfied(actual Type, constraint string) bool {
	if actual == Unknown || constraint == "" {
		return true
	}
	return c.typeImplementsTrait(actual, constraint)
}

func hasJSONAnnotation(annotations []ast.Annotation, name string) bool {
	for _, annotation := range annotations {
		if annotation.Module == "json" && annotation.Name == name {
			return true
		}
	}
	return false
}

func traitMethodType(method *FuncInfo, self Type) Type {
	params := make([]Type, 0, len(method.Params))
	for _, param := range method.Params {
		params = append(params, substituteSelfType(param.Type, self))
	}
	return FuncOfTypes(params, substituteSelfType(method.Return, self))
}

func substituteSelfType(typ Type, self Type) Type {
	return substituteTypeParams(typ, map[string]Type{"Self": self})
}

func (c *checker) coreEnumExists(name string) bool {
	return c != nil && c.info != nil && c.info.Enums[name] != nil
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

func MapOf(key Type, value Type) Type {
	return genericTypeOf("Map", []Type{key, value})
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
	if _, ok := parseNullableType(string(elem)); ok {
		return elem
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

func AsyncFuncOfTypes(params []Type, ret Type) Type {
	parts := make([]string, 0, len(params)+1)
	for _, param := range params {
		parts = append(parts, string(param))
	}
	parts = append(parts, string(ret))
	return Type("AsyncFunc[" + strings.Join(parts, ",") + "]")
}

func ResultOf(ok Type, err Type) Type {
	return genericTypeOf("Result", []Type{ok, err})
}

func TaskOf(result Type) Type {
	return genericTypeOf("Task", []Type{result})
}

func ModuleNamespaceOf(name string) Type {
	return Type("@module:" + name)
}

func ModuleNamespaceName(typ Type) (string, bool) {
	name := strings.TrimPrefix(string(typ), "@module:")
	return name, name != string(typ) && name != ""
}

func ImportNamespaceOf(path string) Type {
	return Type("@import:" + path)
}

func ImportNamespacePath(typ Type) (string, bool) {
	path := strings.TrimPrefix(string(typ), "@import:")
	return path, path != string(typ) && path != ""
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
	if params, ret, ok := parseAsyncFuncType(name); ok {
		for i, param := range params {
			params[i] = displayTypeName(param)
		}
		return "async " + displayFuncType(params, displayTypeName(ret))
	}
	if base, args, ok := parseGenericType(name); ok {
		for i, arg := range args {
			args[i] = displayTypeName(arg)
		}
		return base + "[" + strings.Join(args, ", ") + "]"
	}
	if name == string(Data) {
		return "@io.Data"
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

func MapKeyValue(typ Type) (Type, Type, bool) {
	base, args, ok := parseGenericType(string(typ))
	if !ok || base != "Map" || len(args) != 2 {
		return Unknown, Unknown, false
	}
	return Type(args[0]), Type(args[1]), true
}

func TupleElements(typ Type) ([]Type, bool) {
	base, args, ok := parseGenericType(string(typ))
	if !ok || (base != "Tuple" && base != "ReadonlyTuple") || len(args) == 0 {
		return nil, false
	}
	out := make([]Type, 0, len(args))
	for _, arg := range args {
		out = append(out, Type(arg))
	}
	return out, true
}

func IterValue(typ Type) (Type, bool) {
	base, args, ok := parseGenericType(string(typ))
	if !ok || base != "Iter" || len(args) != 1 {
		return Unknown, false
	}
	return Type(args[0]), true
}

func receiverType(typ *ast.StructType) Type {
	if len(typ.Generics) == 0 {
		return Type(typ.Name)
	}
	return Type(typ.Name + "[" + strings.Join(typ.Generics, ",") + "]")
}

func enumReceiverType(enum *ast.EnumType) Type {
	if len(enum.Generics) == 0 {
		return Type(enum.Name)
	}
	return Type(enum.Name + "[" + strings.Join(enum.Generics, ",") + "]")
}

func structFieldType(info *StructInfo, receiver Type, field FieldInfo) Type {
	return substituteTypeParams(field.Type, typeParamBindingsForStruct(info, receiver))
}

func FieldType(info *Info, receiver Type, name string) (Type, bool) {
	if info == nil {
		return Unknown, false
	}
	structInfo := info.Types[baseTypeName(receiver)]
	if structInfo == nil {
		return Unknown, false
	}
	field, ok := structInfo.ByName[name]
	if !ok {
		return Unknown, false
	}
	return structFieldType(structInfo, receiver, field), true
}

func typeParamBindingsForStruct(info *StructInfo, receiver Type) map[string]Type {
	if info == nil || len(info.Generics) == 0 {
		return nil
	}
	base, args, ok := parseGenericType(string(receiver))
	if !ok || base != info.Name {
		return nil
	}
	bindings := make(map[string]Type, len(info.Generics))
	for idx, name := range info.Generics {
		if idx < len(args) {
			bindings[name] = Type(args[idx])
		}
	}
	return bindings
}

func typeParamBindingsForEnum(info *EnumInfo, receiver Type) map[string]Type {
	if info == nil || len(info.Generics) == 0 {
		return nil
	}
	base, args, ok := parseGenericType(string(receiver))
	if !ok || base != info.Name {
		return nil
	}
	bindings := make(map[string]Type, len(info.Generics))
	for idx, name := range info.Generics {
		if idx < len(args) {
			bindings[name] = Type(args[idx])
		}
	}
	return bindings
}

func substituteTypeParams(typ Type, bindings map[string]Type) Type {
	if len(bindings) == 0 {
		return typ
	}
	if bound, ok := bindings[string(typ)]; ok {
		return bound
	}
	if inner, ok := parseNullableType(string(typ)); ok {
		return NullableOf(substituteTypeParams(Type(inner), bindings))
	}
	if elem, ok := parseArrayType(string(typ)); ok {
		return ArrayOf(substituteTypeParams(Type(elem), bindings))
	}
	if base, args, ok := parseGenericType(string(typ)); ok {
		resolved := make([]Type, 0, len(args))
		for _, arg := range args {
			resolved = append(resolved, substituteTypeParams(Type(arg), bindings))
		}
		return genericTypeOf(base, resolved)
	}
	if params, ret, ok := parseFuncType(string(typ)); ok {
		resolved := make([]Type, 0, len(params))
		for _, param := range params {
			resolved = append(resolved, substituteTypeParams(Type(param), bindings))
		}
		return FuncOfTypes(resolved, substituteTypeParams(Type(ret), bindings))
	}
	if params, ret, ok := parseAsyncFuncType(string(typ)); ok {
		resolved := make([]Type, 0, len(params))
		for _, param := range params {
			resolved = append(resolved, substituteTypeParams(Type(param), bindings))
		}
		return AsyncFuncOfTypes(resolved, substituteTypeParams(Type(ret), bindings))
	}
	return typ
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
	return name == "Dynamic"
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
	if expectedParams, expectedRet, ok := parseCallableType(string(expected)); ok {
		actualParams, actualRet, actualFunc := parseCallableType(string(actual))
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
	case string(Int), string(Int4), string(Int8), string(Int16), string(Int64),
		string(Double), string(Float), string(BigInt), string(UInt), string(UInt8),
		string(UInt16), string(UInt64), string(String), string(Char), string(Bool), string(Null),
		string(Void), string(Symbol), string(Regex), string(Bytes), string(Buffer),
		string(Reader), string(Writer), string(Data):
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
	if leftParams, leftRet, ok := parseCallableType(string(left)); ok {
		rightParams, rightRet, rightFunc := parseCallableType(string(right))
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
