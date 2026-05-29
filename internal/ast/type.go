package ast

import "strings"

type TypeKind int

const (
	TypeInvalid TypeKind = iota
	TypeName
	TypeTuple
	TypeFunction
	TypeGrouped
	TypeRaw
)

type Type struct {
	Kind        TypeKind
	Name        string
	Module      string
	Nullable    bool
	Args        []Type
	Params      []TypeParam
	Return      *Type
	Elem        *Type
	Raw         string
	DisplayText string
}

type TypeParam struct {
	Name     string
	Optional bool
	Type     Type
}

func NamedType(name string) Type {
	return Type{Kind: TypeName, Name: name}
}

func QualifiedType(module string, name string) Type {
	return Type{Kind: TypeName, Module: module, Name: name}
}

func RawType(canonical string) Type {
	return Type{Kind: TypeRaw, Raw: canonical}
}

func GroupedType(elem Type) Type {
	return Type{Kind: TypeGrouped, Elem: typePtr(elem)}
}

func TupleType(elements []TypeParam) Type {
	return Type{Kind: TypeTuple, Params: append([]TypeParam(nil), elements...)}
}

func FunctionType(params []TypeParam, ret Type) Type {
	return Type{Kind: TypeFunction, Params: append([]TypeParam(nil), params...), Return: typePtr(ret)}
}

func (t Type) WithArgs(args []Type) Type {
	t.Args = append([]Type(nil), args...)
	return t
}

func (t Type) WithNullable() Type {
	t.Nullable = true
	return t
}

func (t Type) IsZero() bool {
	return t.Canonical() == ""
}

func (t Type) Canonical() string {
	switch t.Kind {
	case TypeName:
		if t.Name == "" {
			return ""
		}
		out := t.Name
		if len(t.Args) > 0 {
			out += "[" + joinTypeCanonical(t.Args) + "]"
		}
		if t.Nullable {
			out += "?"
		}
		return out
	case TypeTuple:
		parts := make([]string, 0, len(t.Params))
		for _, param := range t.Params {
			parts = append(parts, param.Type.Canonical())
		}
		return "Tuple[" + strings.Join(parts, ",") + "]"
	case TypeFunction:
		parts := make([]string, 0, len(t.Params)+1)
		for _, param := range t.Params {
			parts = append(parts, param.Type.Canonical())
		}
		if t.Return != nil {
			parts = append(parts, t.Return.Canonical())
		}
		return "Func[" + strings.Join(parts, ",") + "]"
	case TypeGrouped:
		if t.Elem == nil {
			return ""
		}
		return t.Elem.Canonical()
	case TypeRaw:
		return t.Raw
	default:
		return ""
	}
}

func (t Type) Display() string {
	switch t.Kind {
	case TypeName:
		if t.Name == "" {
			return ""
		}
		out := t.Name
		if t.Module != "" {
			out = "@" + t.Module + "." + out
		}
		if len(t.Args) > 0 {
			out += "[" + joinTypeDisplay(t.Args) + "]"
		}
		if t.Nullable {
			out += "?"
		}
		return out
	case TypeTuple:
		parts := make([]string, 0, len(t.Params))
		for _, param := range t.Params {
			parts = append(parts, param.Display())
		}
		return "(" + strings.Join(parts, ", ") + ")"
	case TypeFunction:
		parts := make([]string, 0, len(t.Params))
		for _, param := range t.Params {
			parts = append(parts, param.Display())
		}
		ret := ""
		if t.Return != nil {
			ret = t.Return.Display()
		}
		return "(" + strings.Join(parts, ", ") + ") -> " + ret
	case TypeGrouped:
		if t.Elem == nil {
			return ""
		}
		return "(" + t.Elem.Display() + ")"
	case TypeRaw:
		if t.DisplayText != "" {
			return t.DisplayText
		}
		return t.Raw
	default:
		return ""
	}
}

func (t Type) String() string {
	return t.Canonical()
}

func (p TypeParam) Display() string {
	if p.Name == "" {
		return p.Type.Display()
	}
	suffix := ": "
	if p.Optional {
		suffix = "?: "
	}
	return p.Name + suffix + p.Type.Display()
}

func typePtr(t Type) *Type {
	return &t
}

func joinTypeCanonical(types []Type) string {
	parts := make([]string, 0, len(types))
	for _, typ := range types {
		parts = append(parts, typ.Canonical())
	}
	return strings.Join(parts, ",")
}

func joinTypeDisplay(types []Type) string {
	parts := make([]string, 0, len(types))
	for _, typ := range types {
		parts = append(parts, typ.Display())
	}
	return strings.Join(parts, ", ")
}
