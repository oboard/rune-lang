package selfhostrunner

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oboard/rune-lang/internal/ir"
	"github.com/oboard/rune-lang/internal/lexer"
)

const (
	exprUnknown           = 0
	exprIdentifier        = 1
	exprAt                = 2
	exprThis              = 3
	exprInt               = 4
	exprDouble            = 5
	exprBigInt            = 6
	exprString            = 7
	exprTemplate          = 8
	exprChar              = 9
	exprRegex             = 10
	exprBool              = 12
	exprNull              = 13
	exprUnary             = 14
	exprPostfix           = 15
	exprUnwrap            = 16
	exprBinary            = 18
	exprTernary           = 19
	exprAssign            = 20
	exprCall              = 21
	exprLambda            = 23
	exprSelector          = 24
	exprIndex             = 25
	exprArray             = 26
	exprTuple             = 27
	exprMap               = 28
	exprEntry             = 29
	exprSpread            = 30
	exprStruct            = 32
	exprObject            = 33
	exprField             = 34
	exprPrivateField      = 35
	exprBlock             = 38
	exprPatternBlock      = 39
	exprMatch             = 40
	exprBranch            = 41
	exprPattern           = 42
	exprLet               = 43
	exprObjectDestructure = 44
	exprError             = 45
	exprWatch             = 46
)

type selfhostIRFile struct {
	Imports   []selfhostIRImport `json:"imports"`
	Structs   []selfhostIRStruct `json:"structs"`
	Enums     []selfhostIREnum   `json:"enums"`
	Constants []selfhostIRConst  `json:"constants"`
	Functions []selfhostIRFunc   `json:"functions"`
	Tests     []selfhostIRTest   `json:"tests"`
	Errors    []selfhostParseErr `json:"errors"`
}

type selfhostIRImport struct {
	Path   string `json:"path"`
	Go     bool   `json:"go"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

type selfhostIRParam struct {
	Name     string `json:"name"`
	TypeName string `json:"typeName"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
}

type selfhostIRExpr struct {
	Kind     int               `json:"kind"`
	Text     string            `json:"text"`
	Name     string            `json:"name"`
	Value    string            `json:"value"`
	Op       string            `json:"op"`
	Params   []selfhostIRParam `json:"params"`
	Children []selfhostIRExpr  `json:"children"`
	Line     int               `json:"line"`
	Column   int               `json:"column"`
}

type selfhostIRField struct {
	Name       string `json:"name"`
	Private    bool   `json:"private"`
	TypeName   string `json:"typeName"`
	JSONName   string `json:"jsonName"`
	JSONIgnore bool   `json:"jsonIgnore"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
}

type selfhostIREnumMember struct {
	Name    string            `json:"name"`
	Private bool              `json:"private"`
	Value   string            `json:"value"`
	Params  []selfhostIRParam `json:"params"`
	Line    int               `json:"line"`
	Column  int               `json:"column"`
}

type selfhostIRFunc struct {
	Name         string            `json:"name"`
	Private      bool              `json:"private"`
	Routine      bool              `json:"routine"`
	ReceiverType string            `json:"receiverType"`
	Generics     []string          `json:"generics"`
	Params       []selfhostIRParam `json:"params"`
	ReturnType   string            `json:"returnType"`
	Body         selfhostIRExpr    `json:"body"`
	Line         int               `json:"line"`
	Column       int               `json:"column"`
}

type selfhostIRStruct struct {
	Name     string            `json:"name"`
	Private  bool              `json:"private"`
	Generics []string          `json:"generics"`
	Fields   []selfhostIRField `json:"fields"`
	Methods  []selfhostIRFunc  `json:"methods"`
	Line     int               `json:"line"`
	Column   int               `json:"column"`
}

type selfhostIRConst struct {
	Name     string         `json:"name"`
	Private  bool           `json:"private"`
	TypeName string         `json:"typeName"`
	Value    selfhostIRExpr `json:"value"`
	Line     int            `json:"line"`
	Column   int            `json:"column"`
}

type selfhostIREnum struct {
	Name     string                 `json:"name"`
	Private  bool                   `json:"private"`
	Generics []string               `json:"generics"`
	Members  []selfhostIREnumMember `json:"members"`
	Methods  []selfhostIRFunc       `json:"methods"`
	Line     int                    `json:"line"`
	Column   int                    `json:"column"`
}

type selfhostIRTest struct {
	Name   string         `json:"name"`
	Body   selfhostIRExpr `json:"body"`
	Line   int            `json:"line"`
	Column int            `json:"column"`
}

type selfhostParseErr struct {
	Message string `json:"message"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
}

func selfhostFile(file *ir.File) selfhostIRFile {
	out := selfhostIRFile{
		Imports:   []selfhostIRImport{},
		Structs:   make([]selfhostIRStruct, 0, len(file.Types)),
		Enums:     make([]selfhostIREnum, 0, len(file.Enums)),
		Constants: make([]selfhostIRConst, 0, len(file.Constants)),
		Functions: make([]selfhostIRFunc, 0, len(file.Functions)),
		Tests:     make([]selfhostIRTest, 0, len(file.Tests)),
		Errors:    []selfhostParseErr{},
	}
	for _, typ := range file.Types {
		out.Structs = append(out.Structs, selfhostStruct(typ))
	}
	for _, enum := range file.Enums {
		out.Enums = append(out.Enums, selfhostEnum(enum))
	}
	for _, constant := range file.Constants {
		out.Constants = append(out.Constants, selfhostConst(constant))
	}
	for _, fn := range file.Functions {
		out.Functions = append(out.Functions, selfhostFunc(fn))
	}
	for _, test := range file.Tests {
		out.Tests = append(out.Tests, selfhostIRTest{
			Name:   test.Name,
			Body:   selfhostExpr(test.Body),
			Line:   test.Pos.Line,
			Column: test.Pos.Column,
		})
	}
	return out
}

func selfhostStruct(typ *ir.StructType) selfhostIRStruct {
	out := selfhostIRStruct{
		Name:     typ.Name,
		Private:  typ.Private,
		Generics: append([]string(nil), typ.Generics...),
		Fields:   make([]selfhostIRField, 0, len(typ.Fields)),
		Methods:  make([]selfhostIRFunc, 0, len(typ.Methods)),
		Line:     typ.Pos.Line,
		Column:   typ.Pos.Column,
	}
	for _, field := range typ.Fields {
		out.Fields = append(out.Fields, selfhostIRField{
			Name:       field.Name,
			Private:    field.Private,
			TypeName:   string(field.Type),
			JSONName:   field.JSONName,
			JSONIgnore: field.JSONIgnore,
			Line:       field.Pos.Line,
			Column:     field.Pos.Column,
		})
	}
	for _, method := range typ.Methods {
		out.Methods = append(out.Methods, selfhostFunc(method))
	}
	return out
}

func selfhostConst(constant *ir.ConstDecl) selfhostIRConst {
	return selfhostIRConst{
		Name:     constant.Name,
		Private:  constant.Private,
		TypeName: string(constant.Type),
		Value:    selfhostExpr(constant.Value),
		Line:     constant.Pos.Line,
		Column:   constant.Pos.Column,
	}
}

func selfhostEnum(enum *ir.EnumType) selfhostIREnum {
	out := selfhostIREnum{
		Name:     enum.Name,
		Private:  enum.Private,
		Generics: append([]string(nil), enum.Generics...),
		Members:  make([]selfhostIREnumMember, 0, len(enum.Members)),
		Methods:  make([]selfhostIRFunc, 0, len(enum.Methods)),
		Line:     enum.Pos.Line,
		Column:   enum.Pos.Column,
	}
	for _, member := range enum.Members {
		value := ""
		if member.HasValue {
			value = strconv.Itoa(member.Value)
		}
		out.Members = append(out.Members, selfhostIREnumMember{
			Name:    member.Name,
			Private: member.Private,
			Value:   value,
			Params:  selfhostParams(member.Params),
			Line:    member.Pos.Line,
			Column:  member.Pos.Column,
		})
	}
	for _, method := range enum.Methods {
		out.Methods = append(out.Methods, selfhostFunc(method))
	}
	return out
}

func selfhostFunc(fn *ir.Function) selfhostIRFunc {
	return selfhostIRFunc{
		Name:         fn.Name,
		Private:      fn.Private,
		Routine:      fn.Routine,
		ReceiverType: string(fn.ReceiverType),
		Generics:     append([]string(nil), fn.Generics...),
		Params:       selfhostParams(fn.Params),
		ReturnType:   string(fn.Return),
		Body:         selfhostExpr(fn.Body),
		Line:         fn.Pos.Line,
		Column:       fn.Pos.Column,
	}
}

func selfhostParams(params []ir.Param) []selfhostIRParam {
	out := make([]selfhostIRParam, 0, len(params))
	for _, param := range params {
		out = append(out, selfhostIRParam{
			Name:     param.Name,
			TypeName: string(param.Type),
			Line:     param.Pos.Line,
			Column:   param.Pos.Column,
		})
	}
	return out
}

func selfhostExpr(expr ir.Expr) selfhostIRExpr {
	if expr == nil {
		return selfhostEmptyExpr()
	}
	base := selfhostExprBase(expr)
	switch e := expr.(type) {
	case *ir.Identifier:
		base.Kind = exprIdentifier
		base.Name = e.Name
	case *ir.AtExpr:
		base.Kind = exprAt
		base.Name = e.Name
		base.Value = strconv.Quote(e.Path)
	case *ir.ThisExpr:
		base.Kind = exprThis
	case *ir.IntegerLiteral:
		base.Kind = exprInt
		base.Value = strconv.Itoa(e.Value)
	case *ir.DoubleLiteral:
		base.Kind = exprDouble
		base.Value = e.Raw
		if base.Value == "" {
			base.Value = fmt.Sprintf("%g", e.Value)
		}
	case *ir.BigIntLiteral:
		base.Kind = exprBigInt
		base.Value = e.Value
	case *ir.StringLiteral:
		base.Kind = exprString
		base.Value = strconv.Quote(e.Value)
	case *ir.TemplateLiteral:
		base.Kind = exprTemplate
		base.Value = strconv.Quote(templateTextParts(e.Parts))
		base.Children = templateExprs(e.Parts)
	case *ir.CharLiteral:
		base.Kind = exprChar
		base.Value = strconv.QuoteRune(e.Value)
	case *ir.RegexLiteral:
		base.Kind = exprRegex
		base.Value = e.Raw
	case *ir.BoolLiteral:
		base.Kind = exprBool
		base.Value = strconv.FormatBool(e.Value)
	case *ir.NullLiteral:
		base.Kind = exprNull
	case *ir.UnaryExpr:
		base.Kind = exprUnary
		base.Op = e.Op.String()
		base.Children = []selfhostIRExpr{selfhostExpr(e.Expr)}
	case *ir.PostfixExpr:
		base.Kind = exprPostfix
		base.Op = e.Op.String()
		base.Children = []selfhostIRExpr{selfhostExpr(e.Expr)}
	case *ir.ResultUnwrapExpr:
		base.Kind = exprUnwrap
		base.Children = []selfhostIRExpr{selfhostExpr(e.Expr)}
	case *ir.BinaryExpr:
		base.Kind = exprBinary
		base.Op = e.Op.String()
		base.Children = []selfhostIRExpr{selfhostExpr(e.Left), selfhostExpr(e.Right)}
	case *ir.TernaryExpr:
		base.Kind = exprTernary
		base.Children = []selfhostIRExpr{selfhostExpr(e.Condition), selfhostExpr(e.Consequence)}
		if e.Alternative != nil {
			base.Children = append(base.Children, selfhostExpr(e.Alternative))
		}
	case *ir.AssignExpr:
		base.Kind = exprAssign
		base.Name = e.Name
		if e.Target != nil {
			base.Children = []selfhostIRExpr{selfhostExpr(e.Target), selfhostExpr(e.Value)}
		} else {
			base.Children = []selfhostIRExpr{selfhostExpr(e.Value)}
		}
	case *ir.CallExpr:
		base.Kind = exprCall
		base.Children = append([]selfhostIRExpr{selfhostExpr(e.Callee)}, selfhostExprs(e.Args)...)
	case *ir.LambdaExpr:
		base.Kind = exprLambda
		base.Params = lambdaParams(e.Params)
		base.Children = []selfhostIRExpr{selfhostExpr(e.Body)}
	case *ir.SelectorExpr:
		base.Kind = exprSelector
		base.Name = e.Name
		base.Op = "."
		if e.Static {
			base.Op = "::"
		}
		base.Children = []selfhostIRExpr{selfhostExpr(e.Receiver)}
	case *ir.IndexExpr:
		base.Kind = exprIndex
		base.Children = []selfhostIRExpr{selfhostExpr(e.Receiver), selfhostExpr(e.Index)}
	case *ir.ArrayLiteral:
		base.Kind = exprArray
		base.Children = selfhostExprs(e.Elements)
	case *ir.TupleLiteral:
		base.Kind = exprTuple
		base.Children = selfhostExprs(e.Elements)
	case *ir.MapLiteral:
		base.Kind = exprMap
		base.Children = make([]selfhostIRExpr, 0, len(e.Entries))
		for _, entry := range e.Entries {
			base.Children = append(base.Children, selfhostEntry(entry.Key, entry.Value, entry.Pos))
		}
	case *ir.SpreadExpr:
		base.Kind = exprSpread
		base.Children = []selfhostIRExpr{selfhostExpr(e.Expr)}
	case *ir.StructLiteral:
		base.Kind = exprStruct
		base.Name = e.TypeName
		base.Children = selfhostFields(e.Fields)
	case *ir.AnonymousObjectLiteral:
		base.Kind = exprObject
		base.Children = selfhostFields(e.Fields)
	case *ir.BlockExpr:
		base.Kind = exprBlock
		base.Children = selfhostStmts(e.Statements)
	case *ir.PatternBlock:
		base.Kind = exprPatternBlock
		base.Children = selfhostBranches(e.Branches)
	case *ir.MatchExpr:
		base.Kind = exprMatch
		base.Children = append([]selfhostIRExpr{selfhostExpr(e.Subject)}, selfhostBranches(e.Branches)...)
	case *ir.WatchExpr:
		base.Kind = exprWatch
		base.Children = []selfhostIRExpr{selfhostExpr(e.Target), selfhostExpr(e.Handler)}
	default:
		base.Kind = exprError
		base.Text = fmt.Sprintf("unsupported Go IR expression %T", expr)
	}
	return base
}

func selfhostExprBase(expr ir.Expr) selfhostIRExpr {
	pos := expr.Position()
	return selfhostIRExpr{
		Kind:     exprUnknown,
		Params:   []selfhostIRParam{},
		Children: []selfhostIRExpr{},
		Line:     pos.Line,
		Column:   pos.Column,
	}
}

func selfhostEmptyExpr() selfhostIRExpr {
	return selfhostIRExpr{
		Kind:     exprUnknown,
		Params:   []selfhostIRParam{},
		Children: []selfhostIRExpr{},
	}
}

func selfhostExprs(exprs []ir.Expr) []selfhostIRExpr {
	out := make([]selfhostIRExpr, 0, len(exprs))
	for _, expr := range exprs {
		out = append(out, selfhostExpr(expr))
	}
	return out
}

func selfhostStmts(stmts []ir.Stmt) []selfhostIRExpr {
	out := make([]selfhostIRExpr, 0, len(stmts))
	for _, stmt := range stmts {
		out = append(out, selfhostStmt(stmt))
	}
	return out
}

func selfhostStmt(stmt ir.Stmt) selfhostIRExpr {
	if stmt == nil {
		return selfhostEmptyExpr()
	}
	pos := stmt.Position()
	base := selfhostIRExpr{
		Params:   []selfhostIRParam{},
		Children: []selfhostIRExpr{},
		Line:     pos.Line,
		Column:   pos.Column,
	}
	switch s := stmt.(type) {
	case *ir.LetStmt:
		base.Kind = exprLet
		base.Name = s.Name
		base.Text = string(s.Type)
		base.Children = []selfhostIRExpr{selfhostExpr(s.Value)}
	case *ir.ObjectDestructureStmt:
		base.Kind = exprObjectDestructure
		base.Params = make([]selfhostIRParam, 0, len(s.Fields))
		for _, field := range s.Fields {
			base.Params = append(base.Params, selfhostIRParam{Name: field.Name, TypeName: field.Field})
		}
		base.Children = []selfhostIRExpr{selfhostExpr(s.Value)}
	case *ir.AssignStmt:
		base.Kind = exprAssign
		base.Name = s.Name
		base.Children = []selfhostIRExpr{selfhostExpr(s.Value)}
	case *ir.ExprStmt:
		return selfhostExpr(s.Expr)
	default:
		base.Kind = exprError
		base.Text = fmt.Sprintf("unsupported Go IR statement %T", stmt)
	}
	return base
}

func selfhostFields(fields []ir.FieldValue) []selfhostIRExpr {
	out := make([]selfhostIRExpr, 0, len(fields))
	for _, field := range fields {
		kind := exprField
		if field.Private {
			kind = exprPrivateField
		}
		out = append(out, selfhostIRExpr{
			Kind:     kind,
			Name:     field.Name,
			Params:   []selfhostIRParam{},
			Children: []selfhostIRExpr{selfhostExpr(field.Value)},
			Line:     field.Pos.Line,
			Column:   field.Pos.Column,
		})
	}
	return out
}

func selfhostEntry(key ir.Expr, value ir.Expr, pos lexer.Position) selfhostIRExpr {
	return selfhostIRExpr{
		Kind:     exprEntry,
		Params:   []selfhostIRParam{},
		Children: []selfhostIRExpr{selfhostExpr(key), selfhostExpr(value)},
		Line:     pos.Line,
		Column:   pos.Column,
	}
}

func selfhostBranches(branches []ir.PatternBranch) []selfhostIRExpr {
	out := make([]selfhostIRExpr, 0, len(branches))
	for _, branch := range branches {
		out = append(out, selfhostIRExpr{
			Kind:     exprBranch,
			Params:   []selfhostIRParam{},
			Children: []selfhostIRExpr{selfhostPattern(branch.Pattern), selfhostExpr(branch.Expr)},
			Line:     branch.Pos.Line,
			Column:   branch.Pos.Column,
		})
	}
	return out
}

func selfhostPattern(pattern ir.Pattern) selfhostIRExpr {
	if pattern == nil {
		return selfhostIRExpr{Kind: exprPattern, Text: "_", Params: []selfhostIRParam{}, Children: []selfhostIRExpr{}}
	}
	pos := pattern.Position()
	return selfhostIRExpr{
		Kind:     exprPattern,
		Text:     selfhostPatternText(pattern),
		Params:   []selfhostIRParam{},
		Children: []selfhostIRExpr{},
		Line:     pos.Line,
		Column:   pos.Column,
	}
}

func selfhostPatternText(pattern ir.Pattern) string {
	switch p := pattern.(type) {
	case *ir.WildcardPattern:
		return "_"
	case *ir.BindingPattern:
		if p.Constant {
			if p.Type != "" && p.LinkName == "" {
				return p.Name
			}
			return "=" + p.Name
		}
		return p.Name
	case *ir.LiteralPattern:
		return selfhostLiteralPatternText(p.Value)
	case *ir.RangePattern:
		op := "..<"
		if p.Inclusive {
			op = "..="
		}
		return selfhostRangePatternBound(p.Start) + op + selfhostRangePatternBound(p.End)
	case *ir.OrPattern:
		parts := make([]string, 0, len(p.Alternatives))
		for _, alt := range p.Alternatives {
			parts = append(parts, selfhostPatternText(alt))
		}
		return strings.Join(parts, " | ")
	case *ir.ConstructorPattern:
		args := make([]string, 0, len(p.Args))
		for _, arg := range p.Args {
			args = append(args, selfhostPatternText(arg))
		}
		if p.Rest {
			args = append(args, "..")
		}
		return p.Name + "(" + strings.Join(args, ",") + ")"
	case *ir.ArrayPattern:
		parts := make([]string, 0, len(p.Elements)+1)
		for idx, elem := range p.Elements {
			if p.RestIndex == idx {
				parts = append(parts, selfhostRestPatternText(p.RestBinding))
			}
			parts = append(parts, selfhostPatternText(elem))
		}
		if p.RestIndex == len(p.Elements) {
			parts = append(parts, selfhostRestPatternText(p.RestBinding))
		}
		return "[" + strings.Join(parts, ",") + "]"
	case *ir.SequenceSpreadPattern:
		return ".. " + selfhostLiteralPatternText(p.Value)
	case *ir.BitPattern:
		prefix := "u"
		if p.Signed {
			prefix = "i"
		}
		return fmt.Sprintf("%s%d%s(%s)", prefix, p.Width, p.Endian, selfhostPatternText(p.Value))
	case *ir.AliasPattern:
		return selfhostPatternText(p.Pattern) + "@" + p.Name
	case *ir.MapPattern:
		parts := make([]string, 0, len(p.Entries)+1)
		for _, entry := range p.Entries {
			key := selfhostMapKeyText(entry.Key)
			if entry.Optional {
				key += "?"
			}
			parts = append(parts, key+": "+selfhostPatternText(entry.Pattern))
		}
		if p.Rest {
			parts = append(parts, "..")
		}
		return "{" + strings.Join(parts, ",") + "}"
	default:
		return "_"
	}
}

func selfhostMapKeyText(expr ir.Expr) string {
	if ident, ok := expr.(*ir.Identifier); ok {
		return "=" + ident.Name
	}
	return selfhostLiteralPatternText(expr)
}

func selfhostRestPatternText(binding string) string {
	if binding == "" {
		return ".."
	}
	return ".." + binding
}

func selfhostRangePatternBound(expr ir.Expr) string {
	if expr == nil {
		return "_"
	}
	return selfhostLiteralPatternText(expr)
}

func selfhostLiteralPatternText(expr ir.Expr) string {
	switch e := expr.(type) {
	case *ir.IntegerLiteral:
		return strconv.Itoa(e.Value)
	case *ir.StringLiteral:
		return strconv.Quote(e.Value)
	case *ir.CharLiteral:
		return strconv.QuoteRune(e.Value)
	case *ir.BoolLiteral:
		return strconv.FormatBool(e.Value)
	case *ir.NullLiteral:
		return "null"
	case *ir.Identifier:
		return e.Name
	case *ir.SelectorExpr:
		if ident, ok := e.Receiver.(*ir.Identifier); ok {
			return ident.Name + "." + e.Name
		}
		return e.Name
	default:
		return "_"
	}
}

func lambdaParams(names []string) []selfhostIRParam {
	out := make([]selfhostIRParam, 0, len(names))
	for _, name := range names {
		out = append(out, selfhostIRParam{Name: name})
	}
	return out
}

func templateTextParts(parts []ir.TemplatePart) string {
	var out strings.Builder
	for idx, part := range parts {
		if idx > 0 {
			out.WriteString("<<<RUNE_TEMPLATE_PART>>>")
		}
		out.WriteString(part.Text)
	}
	return out.String()
}

func templateExprs(parts []ir.TemplatePart) []selfhostIRExpr {
	out := make([]selfhostIRExpr, 0, len(parts))
	for _, part := range parts {
		if part.Expr == nil {
			out = append(out, selfhostEmptyExpr())
		} else {
			out = append(out, selfhostExpr(part.Expr))
		}
	}
	return out
}
