package ir

import (
	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/lexer"
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
	for _, imp := range file.TSImports {
		tsImport := TSImport{Path: imp.Path, Specifier: imp.Specifier, Pos: imp.Pos}
		for _, fn := range imp.Functions {
			tsImport.Functions = append(tsImport.Functions, TSFunction{Name: fn.Name})
		}
		for _, value := range imp.Values {
			tsImport.Values = append(tsImport.Values, TSValue{Name: value.Name})
		}
		out.TSImports = append(out.TSImports, tsImport)
	}
	for _, typ := range file.Types {
		out.Types = append(out.Types, l.structType(typ))
	}
	for _, enum := range file.Enums {
		out.Enums = append(out.Enums, l.enumType(enum))
	}
	for _, fn := range file.Functions {
		if fn.Macro {
			continue
		}
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
	jsonObject := hasAnnotation(typ.Annotations, "json", "object")
	out := &StructType{Name: typ.Name, Private: typ.Private, Generics: append([]string(nil), typ.Generics...), JSONObject: jsonObject, Pos: typ.Pos, NamePos: typ.NamePos}
	if l.info != nil {
		if info := l.info.Types[typ.Name]; info != nil {
			for _, field := range typ.Fields {
				fieldType := checker.Unknown
				if fieldInfo, ok := info.ByName[field.Name]; ok {
					fieldType = fieldInfo.Type
				}
				out.Fields = append(out.Fields, lowerField(field, fieldType, jsonObject))
			}
			for _, method := range typ.Methods {
				out.Methods = append(out.Methods, l.function(method, typ.Name))
			}
			return out
		}
	}
	for _, field := range typ.Fields {
		out.Fields = append(out.Fields, lowerField(field, checker.Type(field.Type.Canonical()), jsonObject))
	}
	for _, method := range typ.Methods {
		out.Methods = append(out.Methods, l.function(method, typ.Name))
	}
	return out
}

func lowerField(field ast.Field, typ checker.Type, jsonObject bool) Field {
	out := Field{Name: field.Name, Private: field.Private, Type: typ, JSONName: field.Name, Pos: field.Pos}
	if !jsonObject {
		return out
	}
	out.JSONIgnore = hasAnnotation(field.Annotations, "json", "ignore")
	if annotation := findAnnotation(field.Annotations, "json", "name"); annotation != nil && len(annotation.Args) > 0 {
		if value, ok := annotation.Args[0].(*ast.StringLiteral); ok {
			out.JSONName = value.Value
		}
	}
	return out
}

func hasAnnotation(annotations []ast.Annotation, module string, name string) bool {
	return findAnnotation(annotations, module, name) != nil
}

func findAnnotation(annotations []ast.Annotation, module string, name string) *ast.Annotation {
	for idx := range annotations {
		annotation := &annotations[idx]
		if annotation.Module == module && annotation.Name == name {
			return annotation
		}
	}
	return nil
}

func (l lowerer) enumType(enum *ast.EnumType) *EnumType {
	out := &EnumType{Name: enum.Name, Private: enum.Private, Generics: append([]string(nil), enum.Generics...), Pos: enum.Pos, NamePos: enum.NamePos}
	for _, member := range enum.Members {
		lowered := EnumMember{Name: member.Name, Private: member.Private, Value: member.Value, HasValue: member.HasValue, Pos: member.Pos}
		if l.info != nil {
			if enumInfo := l.info.Enums[enum.Name]; enumInfo != nil {
				if memberInfo, ok := enumInfo.ByName[member.Name]; ok {
					lowered.Params = append(lowered.Params, paramsFromInfo(memberInfo.Params, member.Params)...)
				}
			}
		}
		if len(lowered.Params) == 0 {
			for _, param := range member.Params {
				lowered.Params = append(lowered.Params, Param{Name: param.Name, Type: checker.Type(param.Type.Canonical()), Pos: param.Pos})
			}
		}
		out.Members = append(out.Members, lowered)
	}
	return out
}

func paramsFromInfo(infos []checker.ParamInfo, params []ast.Param) []Param {
	out := make([]Param, 0, len(infos))
	for idx, info := range infos {
		pos := lexer.Position{}
		if idx < len(params) {
			pos = params[idx].Pos
		}
		out = append(out, Param{Name: info.Name, Type: info.Type, Pos: pos})
	}
	return out
}

func (l lowerer) function(fn *ast.Function, receiver string) *Function {
	out := &Function{
		Name:         fn.Name,
		SourceName:   fn.Name,
		Private:      fn.Private,
		Static:       fn.Static,
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
		out.Params = append(out.Params, Param{Name: param.Name, Type: checker.Type(param.Type.Canonical()), Pos: param.Pos})
	}
	if returnName := fn.ReturnType.Canonical(); returnName != "" {
		out.Return = checker.Type(returnName)
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
	fn.Generics = append([]string(nil), info.Generics...)
	if len(info.GenericConstraints) > 0 {
		fn.GenericConstraints = map[string]string{}
		for name, constraint := range info.GenericConstraints {
			fn.GenericConstraints[name] = constraint
		}
	}
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
			} else if value := l.info.ResolvedValues[e]; value != nil && value.LinkName != "" {
				name = value.LinkName
			}
		}
		return &Identifier{ExprBase: l.base(e), Name: name}
	case *ast.AtExpr:
		return &AtExpr{ExprBase: l.base(e), Name: e.Name, Path: e.Path, SourcePath: e.SourcePath}
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
	case *ast.TemplateLiteral:
		out := &TemplateLiteral{ExprBase: l.base(e)}
		for _, part := range e.Parts {
			out.Parts = append(out.Parts, TemplatePart{Text: part.Text, Expr: l.expr(part.Expr), Pos: part.Pos})
		}
		return out
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
		var alternative Expr
		if e.Alternative != nil {
			alternative = l.expr(e.Alternative)
		}
		return &TernaryExpr{ExprBase: l.base(e), Condition: l.expr(e.Condition), Consequence: l.expr(e.Consequence), Alternative: alternative}
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
		resolved := ""
		if l.info != nil {
			if fn := l.info.ResolvedSelectorFunctions[e]; fn != nil && fn.LinkName != "" {
				resolved = fn.LinkName
			} else if value := l.info.ResolvedSelectorValues[e]; value != nil && value.LinkName != "" {
				resolved = value.LinkName
			}
		}
		return &SelectorExpr{ExprBase: l.base(e), Receiver: l.expr(e.Receiver), Name: e.Name, Static: e.Static, ResolvedName: resolved}
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
		for _, field := range l.structLiteralFields(e) {
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

func (l lowerer) structLiteralFields(lit *ast.StructLiteral) []ast.FieldValue {
	hasSpread := false
	for _, field := range lit.Fields {
		if field.Spread {
			hasSpread = true
			break
		}
	}
	if !hasSpread || l.info == nil {
		return lit.Fields
	}
	structInfo := l.info.Types[lit.TypeName]
	if structInfo == nil {
		return lit.Fields
	}
	fields := make([]ast.FieldValue, 0, len(lit.Fields)+len(structInfo.Fields))
	for _, field := range lit.Fields {
		if !field.Spread {
			fields = append(fields, field)
			continue
		}
		for _, spreadField := range structInfo.Fields {
			fields = append(fields, ast.FieldValue{
				Name:    spreadField.Name,
				Private: spreadField.Private,
				Value: &ast.SelectorExpr{
					Receiver: field.Value,
					Name:     spreadField.Name,
					Pos:      field.Value.Position(),
					NamePos:  field.Pos,
				},
				Pos: field.Pos,
			})
		}
	}
	return dedupeStructLiteralFields(fields)
}

func dedupeStructLiteralFields(fields []ast.FieldValue) []ast.FieldValue {
	keep := make([]bool, len(fields))
	seen := map[string]bool{}
	for idx := len(fields) - 1; idx >= 0; idx-- {
		name := fields[idx].Name
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		keep[idx] = true
	}
	out := make([]ast.FieldValue, 0, len(seen))
	for idx, field := range fields {
		if keep[idx] {
			out = append(out, field)
		}
	}
	return out
}

func (l lowerer) stmt(stmt ast.Stmt) Stmt {
	switch s := stmt.(type) {
	case *ast.LetStmt:
		typ := checker.Unknown
		if !s.Type.IsZero() {
			typ = checker.Type(s.Type.Canonical())
		}
		return &LetStmt{Name: s.Name, Mutable: s.Mutable, Signal: s.Signal, Value: l.expr(s.Value), Type: typ, Pos: s.Pos}
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
	case *ast.BindingPattern:
		return &BindingPattern{Name: p.Name, Type: checker.Type(p.Type), Pos: p.Pos}
	case *ast.LiteralPattern:
		return &LiteralPattern{Value: l.expr(p.Value), Pos: p.Pos}
	case *ast.ComparePattern:
		return &ComparePattern{Op: p.Op, Value: l.expr(p.Value), Pos: p.Pos}
	case *ast.RangePattern:
		return &RangePattern{Start: l.expr(p.Start), End: l.expr(p.End), Pos: p.Pos}
	case *ast.OrPattern:
		out := &OrPattern{Pos: p.Pos}
		for _, alternative := range p.Alternatives {
			out.Alternatives = append(out.Alternatives, l.pattern(alternative))
		}
		return out
	case *ast.TuplePattern:
		out := &TuplePattern{Pos: p.Pos}
		for _, elem := range p.Elements {
			out.Elements = append(out.Elements, l.pattern(elem))
		}
		return out
	case *ast.ConstructorPattern:
		return &ConstructorPattern{Name: p.Name, Binding: p.Binding, BindingPos: p.BindingPos, Pos: p.Pos}
	case *ast.MapPattern:
		out := &MapPattern{Rest: p.Rest, Pos: p.Pos}
		for _, entry := range p.Entries {
			out.Entries = append(out.Entries, MapPatternEntry{
				Key:      l.expr(entry.Key),
				Pattern:  l.pattern(entry.Pattern),
				Optional: entry.Optional,
				Pos:      entry.Pos,
			})
		}
		return out
	case *ast.ObjectPattern:
		out := &ObjectPattern{Rest: p.Rest, Pos: p.Pos}
		for _, field := range p.Fields {
			out.Fields = append(out.Fields, ObjectPatternField{
				Name:     field.Name,
				Pattern:  l.pattern(field.Pattern),
				Optional: field.Optional,
				Exists:   field.Exists,
				Type:     checker.Type(field.Type),
				Pos:      field.Pos,
			})
		}
		return out
	default:
		return nil
	}
}
