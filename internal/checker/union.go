package checker

import (
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
)

// typeofGuard recognizes `typeof name == "tag"` in either operand order and
// with either equality operator. The booleans report the tag and whether the
// guard is an equality (true) or inequality (false) test.
func typeofGuard(cond ast.Expr) (name string, tag string, equal bool, ok bool) {
	binary, isBinary := cond.(*ast.BinaryExpr)
	if !isBinary {
		return "", "", false, false
	}
	switch binary.Op {
	case lexer.EqualEqual:
		equal = true
	case lexer.BangEqual:
		equal = false
	default:
		return "", "", false, false
	}
	if name, tag, ok = typeofOperand(binary.Left, binary.Right); ok {
		return name, tag, equal, true
	}
	if name, tag, ok = typeofOperand(binary.Right, binary.Left); ok {
		return name, tag, equal, true
	}
	return "", "", false, false
}

func typeofOperand(maybeTypeof, maybeLiteral ast.Expr) (string, string, bool) {
	typeOf, ok := maybeTypeof.(*ast.TypeOfExpr)
	if !ok {
		return "", "", false
	}
	ident, ok := typeOf.Argument.(*ast.Identifier)
	if !ok {
		return "", "", false
	}
	literal, ok := maybeLiteral.(*ast.StringLiteral)
	if !ok {
		return "", "", false
	}
	return ident.Name, literal.Value, true
}

// narrowEnv applies a `typeof` guard to the branch environment. Only union
// types change; non-union locals keep their declared type.
func (c *checker) narrowEnv(cond ast.Expr, env map[string]Type, positive bool) map[string]Type {
	name, tag, equal, ok := typeofGuard(cond)
	if !ok {
		return env
	}
	current, exists := env[name]
	if !exists {
		return env
	}
	match := positive == equal
	narrowed := narrowTypeForTypeof(current, tag, match)
	if narrowed == current {
		return env
	}
	out := cloneEnv(env)
	out[name] = narrowed
	return out
}

func narrowTypeForTypeof(current Type, tag string, wantMatch bool) Type {
	if !isUnionType(current) {
		return current
	}
	parts := splitUnionParts(string(current))
	kept := make([]string, 0, len(parts))
	for _, part := range parts {
		member := Type(part)
		if typeofMatches(member, tag) == wantMatch {
			kept = append(kept, part)
		}
	}
	if len(kept) == 0 || len(kept) == len(parts) {
		return current
	}
	if len(kept) == 1 {
		return Type(kept[0])
	}
	return Type(strings.Join(kept, "|"))
}

// IsUnionType reports whether typ is a canonical union (`A | B`).
func IsUnionType(typ Type) bool {
	return isUnionType(typ)
}

// UnionMembers returns the top-level members of a canonical union.
func UnionMembers(typ Type) ([]Type, bool) {
	parts, ok := parseUnionType(string(typ))
	if !ok {
		return nil, false
	}
	members := make([]Type, 0, len(parts))
	for _, part := range parts {
		members = append(members, Type(part))
	}
	return members, true
}

func typeofMatches(member Type, tag string) bool {
	if inner, ok := parseNullableType(string(member)); ok {
		if tag == "undefined" || tag == "object" {
			return true
		}
		return typeofMatches(Type(inner), tag)
	}
	switch tag {
	case "string":
		return member == String || member == Char
	case "number":
		return isNumericType(member)
	case "boolean":
		return member == Bool
	case "undefined":
		return member == Undefined
	case "object":
		return isObjectLike(member) || member == Null
	case "function":
		_, _, ok := parseCallableType(string(member))
		return ok
	default:
		return false
	}
}
