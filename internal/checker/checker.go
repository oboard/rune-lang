package checker

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
	"github.com/oboard/rune-lang/internal/stdlib"
)

type Type string

const (
	Unknown Type = "Unknown"
	Int     Type = "Int"
	String  Type = "String"
	Bool    Type = "Bool"
	Void    Type = "Void"
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

type Info struct {
	Functions map[string]*FuncInfo
	Types     map[string]*StructInfo
	Stdlib    *stdlib.Registry
	ExprTypes map[ast.Expr]Type
}
