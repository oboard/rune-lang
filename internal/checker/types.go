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
	case "String":
		return String
	case "Bool":
		return Bool
	case "Void":
		return Void
	default:
		if isDynamicTypeName(name) {
			return Unknown
		}
		if generics[name] {
			return Type(name)
		}
		if elemName, ok := parseArrayType(name); ok {
			elem := c.resolveTypeWithGenerics(elemName, generics)
			if elem == Unknown {
				return Unknown
			}
			return ArrayOf(elem)
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
	return false
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
	c.diags = append(c.diags, Diagnostic{Message: fmt.Sprintf(format, args...), Pos: pos})
}
