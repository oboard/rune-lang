// Package tsdecl parses TypeScript declaration files (`.d.ts`) such as
// lib.dom.d.ts into a lightweight model that the Rune checker and LSP can use
// for type checking, completion, and hover information.
//
// The parser is deliberately tolerant: it understands the declaration-file
// grammar (interfaces, interface inheritance, type aliases, and global
// `declare function` statements) without attempting to parse JavaScript
// expressions. This makes it robust against the full TypeScript type system —
// constructs it cannot represent (conditional types, template literal types,
// path-mapped unions, etc.) degrade to the Rune Dynamic type instead of
// failing the overall parse.
package tsdecl

import "os"

// DOM is the parsed view of a TypeScript declaration library.
type DOM struct {
	Interfaces map[string]*Interface
	Aliases    map[string]*Alias
	Globals    map[string]*Value
	Functions  []*Function
	Order      []string
}

// Alias records a `type X = ...` declaration. The parsed node is retained
// when the shape is representable (callback types, simple refs); unsupported
// shapes keep Node nil and degrade to Dynamic downstream.
type Alias struct {
	Name string
	Node Type
	Pos  Position
}

// Value is a global `declare var/let/const` value (e.g. `document: Document`).
type Value struct {
	Name string
	Type Type
	Pos  Position
}

// Interface describes a parsed `interface` declaration with (possibly
// unresolved) base interface names.
type Interface struct {
	Name    string
	Bases   []string
	Members []Member
	Pos     Position
}

// Member is a property or method entry of an interface. Methods have Params
// populated; plain properties leave them nil.
type Member struct {
	Name      string
	Type      Type
	Params    []Param
	Optional  bool
	IsMethod  bool
	Pos       Position
	Signature string
}

// Param is a method/function parameter.
type Param struct {
	Name     string
	Type     Type
	Optional bool
	Rest     bool
}

// Function is a global `declare function` entry. Overloads with the same name
// appear as separate entries.
type Function struct {
	Name   string
	Params []Param
	Return Type
	Pos    Position
}

// Position tracks a byte offset within the parsed source so checker/LSP can
// map DOM declarations back to file locations if needed.
type Position struct {
	Offset int
}

// Type is the model of a TypeScript type expression. The model is deliberately
// narrow — TypeScript constructs the model cannot represent (template literal
// types, indexed access, path-mapped union values, etc.) fold into Any.
type Type interface{ isType() }

// Any corresponds to TypeScript any/unknown — anything may flow through.
type Any struct{}

func (Any) isType() {}

// Primitive maps TS scalar keywords (string, number, boolean, bigint,
// symbol, null, undefined, void, object) and literal string names.
type Primitive struct{ Name string }

func (Primitive) isType() {}

// Literal is a TS literal type ("a" | number-like); the value is preserved so
// unions folding to a set of string leaves can report the union in hover.
type Literal struct{ Value string }

func (Literal) isType() {}

// Ref names another declared type, optionally with generic arguments.
type Ref struct {
	Name string
	Args []Type
}

func (Ref) isType() {}

// Array is a homogeneous sequence type.
type Array struct{ Elem Type }

func (Array) isType() {}

// Union is a list of alternatives. Display picks the first non-null branch.
type Union struct{ Options []Type }

func (Union) isType() {}

// Intersect is a flattened `A & B` conjunction.
type Intersect struct{ Options []Type }

func (Intersect) isType() {}

// Callback is a function-typed member or alias, used to surface DOM callback
// parameter shapes (e.g. `(event: Event) => void`) to the Rune side.
type Callback struct {
	Params []Param
	Return Type
}

func (Callback) isType() {}

// LoadFile reads and parses a declaration file.
func LoadFile(path string) (*DOM, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(string(data)), nil
}
