package checker

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
	"github.com/oboard/rune-lang/internal/stdlib"
)

type Type string

const (
	Unknown     Type = "Unknown"
	Int         Type = "Int"
	Int4        Type = "Int4"
	Int8        Type = "Int8"
	Int16       Type = "Int16"
	Int64       Type = "Int64"
	Double      Type = "Double"
	Float       Type = "Float"
	BigInt      Type = "BigInt"
	UInt        Type = "UInt"
	UInt8       Type = "UInt8"
	UInt16      Type = "UInt16"
	UInt64      Type = "UInt64"
	String      Type = "String"
	Bool        Type = "Bool"
	Null        Type = "Null"
	Object      Type = "Object"
	Binary      Type = "Binary"
	Buffer      Type = "Buffer"
	Reader      Type = "Reader"
	Writer      Type = "Writer"
	Never       Type = "Never"
	Symbol      Type = "Symbol"
	Regex       Type = "Regex"
	Void        Type = "Void"
	HTMLElement Type = "HTMLElement"
)

type Diagnostic struct {
	Message string
	Pos     lexer.Position
}

type ParamInfo struct {
	Name string
	Type Type
}

type FuncInfo struct {
	Name           string
	Generics       []string
	ReceiverType   Type
	Params         []ParamInfo
	Return         Type
	ReturnDeclared bool
	Node           *ast.Function
}

type FieldInfo struct {
	Name string
	Type Type
}

type StructInfo struct {
	Name     string
	Generics []string
	Fields   []FieldInfo
	ByName   map[string]FieldInfo
	Methods  map[string]*FuncInfo
	Node     *ast.StructType
}

type EnumMemberInfo struct {
	Name  string
	Value int
}

type EnumInfo struct {
	Name    string
	Members []EnumMemberInfo
	ByName  map[string]EnumMemberInfo
	Node    *ast.EnumType
}

type Info struct {
	Functions map[string]*FuncInfo
	Types     map[string]*StructInfo
	Enums     map[string]*EnumInfo
	Stdlib    *stdlib.Registry
	ExprTypes map[ast.Expr]Type
}
