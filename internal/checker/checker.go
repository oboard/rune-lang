package checker

import (
	"fmt"
	"hash/fnv"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/lexer"
	"github.com/oboard/rune-lang/internal/stdlib"
)

type Type string

const (
	Unknown       Type = "Unknown"
	Int           Type = "Int"
	Int4          Type = "Int4"
	Int8          Type = "Int8"
	Int16         Type = "Int16"
	Int64         Type = "Int64"
	Double        Type = "Double"
	Float         Type = "Float"
	BigInt        Type = "BigInt"
	UInt          Type = "UInt"
	UInt8         Type = "UInt8"
	UInt16        Type = "UInt16"
	UInt64        Type = "UInt64"
	String        Type = "String"
	Char          Type = "Char"
	Bool          Type = "Bool"
	Null          Type = "Null"
	Object        Type = "Object"
	Bytes         Type = "Bytes"
	Buffer        Type = "Buffer"
	Reader        Type = "Reader"
	Writer        Type = "Writer"
	StringBuffer  Type = "StringBuffer"
	FileStat      Type = "FileStat"
	TCPConnection Type = "TCPConnection"
	TCPListener   Type = "TCPListener"
	Data          Type = "Data"
	Error         Type = "Error"
	Never         Type = "Never"
	Symbol        Type = "Symbol"
	Regex         Type = "Regex"
	Void          Type = "Void"
	HTMLElement   Type = "HTMLElement"
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
	LinkName       string
	Private        bool
	SourcePath     string
	Routine        bool
	Generics       []string
	ReceiverType   Type
	Params         []ParamInfo
	Return         Type
	ReturnDeclared bool
	Node           *ast.Function
}

type FieldInfo struct {
	Name       string
	Private    bool
	SourcePath string
	Type       Type
}

type StructInfo struct {
	Name       string
	Private    bool
	SourcePath string
	Generics   []string
	Fields     []FieldInfo
	ByName     map[string]FieldInfo
	Methods    map[string]*FuncInfo
	Node       *ast.StructType
}

type EnumMemberInfo struct {
	Name       string
	Private    bool
	SourcePath string
	Value      int
}

type EnumInfo struct {
	Name       string
	Private    bool
	SourcePath string
	Members    []EnumMemberInfo
	ByName     map[string]EnumMemberInfo
	Node       *ast.EnumType
}

type Info struct {
	Functions         map[string]*FuncInfo
	FunctionDecls     map[*ast.Function]*FuncInfo
	ResolvedFunctions map[*ast.Identifier]*FuncInfo
	Types             map[string]*StructInfo
	Enums             map[string]*EnumInfo
	Stdlib            *stdlib.Registry
	ExprTypes         map[ast.Expr]Type
	AsyncCalls        map[*ast.CallExpr]bool
	AwaitCalls        map[*ast.CallExpr]bool

	functionsByName map[string][]*FuncInfo
}

func privateLinkName(sourcePath string, name string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(normalizeSourcePath(sourcePath)))
	return fmt.Sprintf("__rune_private_%08x_%s", h.Sum32(), name)
}
