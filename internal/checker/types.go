package checker

import (
	"fmt"
	"strings"

	"github.com/oboard/rune-lang/internal/lexer"
)

func (c *checker) resolveType(name string) Type {
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
		if strings.HasPrefix(name, "Array[") && strings.HasSuffix(name, "]") {
			elem := c.resolveType(strings.TrimSuffix(strings.TrimPrefix(name, "Array["), "]"))
			if elem == Unknown {
				return Unknown
			}
			return ArrayOf(elem)
		}
		if strings.HasPrefix(name, "Func[") && strings.HasSuffix(name, "]") {
			return Type(name)
		}
		if _, ok := c.info.Types[name]; ok {
			return Type(name)
		}
		return Unknown
	}
}

func (c *checker) resolveDeclaredReturn(name string) Type {
	if name == "Dynamic" || name == "Any" {
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

func ArrayElement(typ Type) (Type, bool) {
	name := string(typ)
	if !strings.HasPrefix(name, "Array[") || !strings.HasSuffix(name, "]") {
		return Unknown, false
	}
	return Type(strings.TrimSuffix(strings.TrimPrefix(name, "Array["), "]")), true
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
