package ir

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
)

func LowerPackage(name string, modules []*Module) *Package {
	return &Package{Name: name, Modules: modules}
}

func LowerModule(name string, files []*File) *Module {
	return &Module{Name: name, Files: files}
}

func LowerFile(file *ast.File, info *checker.Info) *File {
	l := lowerer{info: info}
	out := &File{Stdlib: info.Stdlib}
	for _, imp := range file.GoImports {
		out.GoImports = append(out.GoImports, GoImport{Path: imp.Path, Pos: imp.Pos})
	}
	for _, typ := range file.Types {
		out.Types = append(out.Types, l.structType(typ))
	}
	for _, fn := range file.Functions {
		out.Functions = append(out.Functions, l.function(fn, ""))
	}
	return out
}

func LowerExpr(expr ast.Expr, info *checker.Info) Expr {
	return lowerer{info: info}.expr(expr)
}

func LowerStmt(stmt ast.Stmt, info *checker.Info) Stmt {
	return lowerer{info: info}.stmt(stmt)
}

type lowerer struct {
	info *checker.Info
}

func (l lowerer) exprType(expr ast.Expr) checker.Type {
	if l.info == nil || expr == nil {
		return checker.Unknown
	}
	if typ, ok := l.info.ExprTypes[expr]; ok {
		return typ
	}
	return checker.Unknown
}

func (l lowerer) base(expr ast.Expr) ExprBase {
	return ExprBase{Pos: expr.Position(), Type: l.exprType(expr)}
}

func (l lowerer) structType(typ *ast.StructType) *StructType {
	out := &StructType{Name: typ.Name, Pos: typ.Pos, NamePos: typ.NamePos}
	if l.info != nil {
		if info := l.info.Types[typ.Name]; info != nil {
			for _, field := range typ.Fields {
				fieldType := checker.Unknown
				if fieldInfo, ok := info.ByName[field.Name]; ok {
					fieldType = fieldInfo.Type
				}
				out.Fields = append(out.Fields, Field{Name: field.Name, Type: fieldType, Pos: field.Pos})
			}
			for _, method := range typ.Methods {
				out.Methods = append(out.Methods, l.function(method, typ.Name))
			}
			return out
		}
	}
	for _, field := range typ.Fields {
		out.Fields = append(out.Fields, Field{Name: field.Name, Type: checker.Type(field.Type), Pos: field.Pos})
	}
	for _, method := range typ.Methods {
		out.Methods = append(out.Methods, l.function(method, typ.Name))
	}
	return out
}

func (l lowerer) function(fn *ast.Function, receiver string) *Function {
	out := &Function{
		Name:         fn.Name,
		ReceiverType: checker.Type(receiver),
		Return:       checker.Unknown,
		Body:         l.expr(fn.Body),
		Pos:          fn.Pos,
		NamePos:      fn.NamePos,
	}
	if l.info != nil {
		if receiver != "" {
			if typ := l.info.Types[receiver]; typ != nil {
				if info := typ.Methods[fn.Name]; info != nil {
					l.fillFunctionInfo(out, info)
					return out
				}
			}
		} else if info := l.info.Functions[fn.Name]; info != nil {
			l.fillFunctionInfo(out, info)
			return out
		}
	}
	for _, param := range fn.Params {
		out.Params = append(out.Params, Param{Name: param.Name, Type: checker.Type(param.Type), Pos: param.Pos})
	}
	if fn.ReturnType != "" {
		out.Return = checker.Type(fn.ReturnType)
	}
	return out
}

func (l lowerer) fillFunctionInfo(fn *Function, info *checker.FuncInfo) {
	fn.ReceiverType = info.ReceiverType
	fn.Return = info.Return
	for _, param := range info.Params {
		fn.Params = append(fn.Params, Param{Name: param.Name, Type: param.Type})
	}
}

func (l lowerer) expr(expr ast.Expr) Expr {
	switch e := expr.(type) {
	case *ast.Identifier:
		return &Identifier{ExprBase: l.base(e), Name: e.Name}
	case *ast.AtExpr:
		return &AtExpr{ExprBase: l.base(e), Name: e.Name}
	case *ast.ThisExpr:
		return &ThisExpr{ExprBase: l.base(e)}
	case *ast.IntegerLiteral:
		return &IntegerLiteral{ExprBase: l.base(e), Value: e.Value}
	case *ast.StringLiteral:
		return &StringLiteral{ExprBase: l.base(e), Value: e.Value}
	case *ast.BoolLiteral:
		return &BoolLiteral{ExprBase: l.base(e), Value: e.Value}
	case *ast.UnaryExpr:
		return &UnaryExpr{ExprBase: l.base(e), Op: e.Op, Expr: l.expr(e.Expr)}
	case *ast.BinaryExpr:
		return &BinaryExpr{ExprBase: l.base(e), Left: l.expr(e.Left), Op: e.Op, Right: l.expr(e.Right)}
	case *ast.CallExpr:
		out := &CallExpr{ExprBase: l.base(e), Callee: l.expr(e.Callee)}
		for _, arg := range e.Args {
			out.Args = append(out.Args, l.expr(arg))
		}
		return out
	case *ast.LambdaExpr:
		return &LambdaExpr{ExprBase: l.base(e), Params: append([]string(nil), e.Params...), Body: l.expr(e.Body)}
	case *ast.SelectorExpr:
		return &SelectorExpr{ExprBase: l.base(e), Receiver: l.expr(e.Receiver), Name: e.Name}
	case *ast.IndexExpr:
		return &IndexExpr{ExprBase: l.base(e), Receiver: l.expr(e.Receiver), Index: l.expr(e.Index)}
	case *ast.ArrayLiteral:
		out := &ArrayLiteral{ExprBase: l.base(e)}
		for _, elem := range e.Elements {
			out.Elements = append(out.Elements, l.expr(elem))
		}
		return out
	case *ast.StructLiteral:
		out := &StructLiteral{ExprBase: l.base(e), TypeName: e.TypeName}
		for _, field := range e.Fields {
			out.Fields = append(out.Fields, FieldValue{Name: field.Name, Value: l.expr(field.Value), Pos: field.Pos})
		}
		return out
	case *ast.BlockExpr:
		out := &BlockExpr{ExprBase: l.base(e)}
		for _, stmt := range e.Statements {
			out.Statements = append(out.Statements, l.stmt(stmt))
		}
		return out
	case *ast.PatternBlock:
		out := &PatternBlock{ExprBase: l.base(e)}
		for _, branch := range e.Branches {
			out.Branches = append(out.Branches, PatternBranch{
				Pattern: l.pattern(branch.Pattern),
				Expr:    l.expr(branch.Expr),
				Pos:     branch.Pos,
			})
		}
		return out
	default:
		return nil
	}
}

func (l lowerer) stmt(stmt ast.Stmt) Stmt {
	switch s := stmt.(type) {
	case *ast.LetStmt:
		return &LetStmt{Name: s.Name, Mutable: s.Mutable, Value: l.expr(s.Value), Pos: s.Pos}
	case *ast.AssignStmt:
		return &AssignStmt{Name: s.Name, Value: l.expr(s.Value), Pos: s.Pos}
	case *ast.ExprStmt:
		return &ExprStmt{Expr: l.expr(s.Expr), Pos: s.Pos}
	default:
		return nil
	}
}

func (l lowerer) pattern(pattern ast.Pattern) Pattern {
	switch p := pattern.(type) {
	case *ast.WildcardPattern:
		return &WildcardPattern{Pos: p.Pos}
	case *ast.LiteralPattern:
		return &LiteralPattern{Value: l.expr(p.Value), Pos: p.Pos}
	case *ast.ComparePattern:
		return &ComparePattern{Op: p.Op, Value: l.expr(p.Value), Pos: p.Pos}
	case *ast.TuplePattern:
		out := &TuplePattern{Pos: p.Pos}
		for _, elem := range p.Elements {
			out.Elements = append(out.Elements, l.pattern(elem))
		}
		return out
	default:
		return nil
	}
}
