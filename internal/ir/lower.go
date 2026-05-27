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
	for _, enum := range file.Enums {
		out.Enums = append(out.Enums, l.enumType(enum))
	}
	for _, fn := range file.Functions {
		out.Functions = append(out.Functions, l.function(fn, ""))
	}
	for _, test := range file.Tests {
		out.Tests = append(out.Tests, l.test(test))
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
	out := &StructType{Name: typ.Name, Private: typ.Private, Generics: append([]string(nil), typ.Generics...), Pos: typ.Pos, NamePos: typ.NamePos}
	if l.info != nil {
		if info := l.info.Types[typ.Name]; info != nil {
			for _, field := range typ.Fields {
				fieldType := checker.Unknown
				if fieldInfo, ok := info.ByName[field.Name]; ok {
					fieldType = fieldInfo.Type
				}
				out.Fields = append(out.Fields, Field{Name: field.Name, Private: field.Private, Type: fieldType, Pos: field.Pos})
			}
			for _, method := range typ.Methods {
				out.Methods = append(out.Methods, l.function(method, typ.Name))
			}
			return out
		}
	}
	for _, field := range typ.Fields {
		out.Fields = append(out.Fields, Field{Name: field.Name, Private: field.Private, Type: checker.Type(field.Type), Pos: field.Pos})
	}
	for _, method := range typ.Methods {
		out.Methods = append(out.Methods, l.function(method, typ.Name))
	}
	return out
}

func (l lowerer) enumType(enum *ast.EnumType) *EnumType {
	out := &EnumType{Name: enum.Name, Private: enum.Private, Pos: enum.Pos, NamePos: enum.NamePos}
	for _, member := range enum.Members {
		out.Members = append(out.Members, EnumMember{Name: member.Name, Private: member.Private, Value: member.Value, Pos: member.Pos})
	}
	return out
}

func (l lowerer) function(fn *ast.Function, receiver string) *Function {
	out := &Function{
		Name:         fn.Name,
		Private:      fn.Private,
		Routine:      fn.Routine,
		Generics:     append([]string(nil), fn.Generics...),
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
		} else if info := l.info.FunctionDecls[fn]; info != nil {
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

func (l lowerer) test(test *ast.Test) *Test {
	return &Test{
		Name: test.Name,
		Body: l.expr(test.Body),
		Pos:  test.Pos,
	}
}

func (l lowerer) fillFunctionInfo(fn *Function, info *checker.FuncInfo) {
	if info.LinkName != "" {
		fn.Name = info.LinkName
	}
	fn.ReceiverType = info.ReceiverType
	fn.Return = info.Return
	for _, param := range info.Params {
		fn.Params = append(fn.Params, Param{Name: param.Name, Type: param.Type})
	}
}

func (l lowerer) expr(expr ast.Expr) Expr {
	switch e := expr.(type) {
	case *ast.Identifier:
		name := e.Name
		if l.info != nil {
			if fn := l.info.ResolvedFunctions[e]; fn != nil && fn.LinkName != "" {
				name = fn.LinkName
			}
		}
		return &Identifier{ExprBase: l.base(e), Name: name}
	case *ast.AtExpr:
		return &AtExpr{ExprBase: l.base(e), Name: e.Name}
	case *ast.ThisExpr:
		return &ThisExpr{ExprBase: l.base(e)}
	case *ast.IntegerLiteral:
		return &IntegerLiteral{ExprBase: l.base(e), Value: e.Value}
	case *ast.DoubleLiteral:
		return &DoubleLiteral{ExprBase: l.base(e), Value: e.Value, Raw: e.Raw}
	case *ast.BigIntLiteral:
		return &BigIntLiteral{ExprBase: l.base(e), Value: e.Value}
	case *ast.StringLiteral:
		return &StringLiteral{ExprBase: l.base(e), Value: e.Value}
	case *ast.CharLiteral:
		return &CharLiteral{ExprBase: l.base(e), Value: e.Value}
	case *ast.RegexLiteral:
		return &RegexLiteral{ExprBase: l.base(e), Pattern: e.Pattern, Flags: e.Flags, Raw: e.Raw}
	case *ast.BoolLiteral:
		return &BoolLiteral{ExprBase: l.base(e), Value: e.Value}
	case *ast.NullLiteral:
		return &NullLiteral{ExprBase: l.base(e)}
	case *ast.UnaryExpr:
		return &UnaryExpr{ExprBase: l.base(e), Op: e.Op, Expr: l.expr(e.Expr)}
	case *ast.PostfixExpr:
		return &PostfixExpr{ExprBase: l.base(e), Op: e.Op, Expr: l.expr(e.Expr)}
	case *ast.ResultUnwrapExpr:
		return &ResultUnwrapExpr{ExprBase: l.base(e), Expr: l.expr(e.Expr)}
	case *ast.BinaryExpr:
		return &BinaryExpr{ExprBase: l.base(e), Left: l.expr(e.Left), Op: e.Op, Right: l.expr(e.Right)}
	case *ast.TernaryExpr:
		return &TernaryExpr{ExprBase: l.base(e), Condition: l.expr(e.Condition), Consequence: l.expr(e.Consequence), Alternative: l.expr(e.Alternative)}
	case *ast.AssignExpr:
		return &AssignExpr{ExprBase: l.base(e), Name: e.Name, Target: l.expr(e.Target), Value: l.expr(e.Value)}
	case *ast.CallExpr:
		out := &CallExpr{ExprBase: l.base(e), Callee: l.expr(e.Callee)}
		if l.info != nil {
			out.Async = l.info.AsyncCalls[e]
			out.Await = l.info.AwaitCalls[e]
		}
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
	case *ast.TupleLiteral:
		out := &TupleLiteral{ExprBase: l.base(e)}
		for _, elem := range e.Elements {
			out.Elements = append(out.Elements, l.expr(elem))
		}
		return out
	case *ast.MapLiteral:
		out := &MapLiteral{ExprBase: l.base(e)}
		for _, entry := range e.Entries {
			out.Entries = append(out.Entries, MapEntry{Key: l.expr(entry.Key), Value: l.expr(entry.Value), Pos: entry.Pos})
		}
		return out
	case *ast.SpreadExpr:
		return &SpreadExpr{ExprBase: l.base(e), Expr: l.expr(e.Expr)}
	case *ast.ReactiveLiteral:
		return &ReactiveLiteral{ExprBase: l.base(e), Value: l.expr(e.Value)}
	case *ast.StructLiteral:
		out := &StructLiteral{ExprBase: l.base(e), TypeName: e.TypeName}
		for _, field := range e.Fields {
			out.Fields = append(out.Fields, FieldValue{Name: field.Name, Private: field.Private, Value: l.expr(field.Value), Pos: field.Pos})
		}
		return out
	case *ast.AnonymousObjectLiteral:
		out := &AnonymousObjectLiteral{ExprBase: l.base(e)}
		for _, field := range e.Fields {
			out.Fields = append(out.Fields, FieldValue{Name: field.Name, Private: field.Private, Value: l.expr(field.Value), Pos: field.Pos})
		}
		return out
	case *ast.XMLElement:
		out := &XMLElement{ExprBase: l.base(e), Tag: e.Tag}
		for _, attr := range e.Attrs {
			out.Attrs = append(out.Attrs, XMLAttr{Name: attr.Name, Event: attr.Event, Value: l.expr(attr.Value), Pos: attr.Pos})
		}
		for _, child := range e.Children {
			out.Children = append(out.Children, XMLChild{Text: child.Text, Expr: l.expr(child.Expr), Pos: child.Pos})
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
	case *ast.MatchExpr:
		out := &MatchExpr{ExprBase: l.base(e), Subject: l.expr(e.Subject)}
		for _, branch := range e.Branches {
			out.Branches = append(out.Branches, PatternBranch{
				Pattern: l.pattern(branch.Pattern),
				Expr:    l.expr(branch.Expr),
				Pos:     branch.Pos,
			})
		}
		return out
	case *ast.WatchExpr:
		return &WatchExpr{ExprBase: l.base(e), Target: l.expr(e.Target), Handler: l.expr(e.Handler)}
	default:
		return nil
	}
}

func (l lowerer) stmt(stmt ast.Stmt) Stmt {
	switch s := stmt.(type) {
	case *ast.LetStmt:
		return &LetStmt{Name: s.Name, Mutable: s.Mutable, Signal: s.Signal, Value: l.expr(s.Value), Pos: s.Pos}
	case *ast.ObjectDestructureStmt:
		value := l.expr(s.Value)
		out := &ObjectDestructureStmt{Mutable: s.Mutable, Signal: s.Signal, Value: value, Pos: s.Pos}
		for _, field := range s.Fields {
			typ := checker.Unknown
			if value != nil {
				if fieldType, ok := checker.FieldType(l.info, value.ResultType(), field.Field); ok {
					typ = fieldType
				}
			}
			out.Fields = append(out.Fields, ObjectBindingField{
				Field:    field.Field,
				Name:     field.Name,
				Type:     typ,
				FieldPos: field.FieldPos,
				NamePos:  field.NamePos,
			})
		}
		return out
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
	case *ast.RangePattern:
		return &RangePattern{Start: l.expr(p.Start), End: l.expr(p.End), Pos: p.Pos}
	case *ast.TuplePattern:
		out := &TuplePattern{Pos: p.Pos}
		for _, elem := range p.Elements {
			out.Elements = append(out.Elements, l.pattern(elem))
		}
		return out
	case *ast.ConstructorPattern:
		return &ConstructorPattern{Name: p.Name, Binding: p.Binding, BindingPos: p.BindingPos, Pos: p.Pos}
	default:
		return nil
	}
}
