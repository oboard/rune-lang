package compiler

import (
	"fmt"
	"math"
	"math/big"
	"sort"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/interpreter"
	"github.com/oboard/rune-lang/internal/ir"
	"github.com/oboard/rune-lang/internal/lexer"
)

func expandCompileTimeExprs(file *ast.File, info *checker.Info) (bool, map[*ast.Function]bool, []checker.Diagnostic) {
	if !hasCompileTimeExpr(file) {
		return false, nil, nil
	}
	runtime := interpreter.New(ir.LowerFile(file, info), interpreter.WithCompileTime())
	rewriter := compileTimeRewriter{
		info:                 info,
		runtime:              runtime,
		compileTimeFunctions: map[*ast.Function]bool{},
	}
	rewriter.file(file)
	return rewriter.changed, rewriter.compileTimeFunctions, rewriter.diags
}

type compileTimeRewriter struct {
	info                 *checker.Info
	runtime              *interpreter.Interpreter
	changed              bool
	compileTimeFunctions map[*ast.Function]bool
	diags                []checker.Diagnostic
}

func (r *compileTimeRewriter) file(file *ast.File) {
	for idx := range file.Traits {
		for _, method := range file.Traits[idx].Methods {
			r.function(method)
		}
	}
	for idx := range file.Types {
		r.annotations(file.Types[idx].Annotations)
		for fieldIdx := range file.Types[idx].Fields {
			r.annotations(file.Types[idx].Fields[fieldIdx].Annotations)
		}
		for _, method := range file.Types[idx].Methods {
			r.function(method)
		}
	}
	for idx := range file.Enums {
		r.annotations(file.Enums[idx].Annotations)
		for memberIdx := range file.Enums[idx].Members {
			r.annotations(file.Enums[idx].Members[memberIdx].Annotations)
		}
	}
	for _, fn := range file.Functions {
		r.function(fn)
	}
	for _, test := range file.Tests {
		test.Body = r.expr(test.Body)
	}
}

func (r *compileTimeRewriter) annotations(annotations []ast.Annotation) {
	for idx := range annotations {
		for argIdx := range annotations[idx].Args {
			annotations[idx].Args[argIdx] = r.expr(annotations[idx].Args[argIdx])
		}
	}
}

func (r *compileTimeRewriter) function(fn *ast.Function) {
	if fn == nil {
		return
	}
	r.annotations(fn.Annotations)
	fn.Body = r.expr(fn.Body)
}

func (r *compileTimeRewriter) stmt(stmt ast.Stmt) ast.Stmt {
	switch s := stmt.(type) {
	case *ast.LetStmt:
		s.Value = r.expr(s.Value)
	case *ast.ObjectDestructureStmt:
		s.Value = r.expr(s.Value)
	case *ast.AssignStmt:
		s.Value = r.expr(s.Value)
	case *ast.ExprStmt:
		s.Expr = r.expr(s.Expr)
	}
	return stmt
}

func (r *compileTimeRewriter) expr(expr ast.Expr) ast.Expr {
	switch e := expr.(type) {
	case nil:
		return nil
	case *ast.TemplateLiteral:
		for idx := range e.Parts {
			e.Parts[idx].Expr = r.expr(e.Parts[idx].Expr)
		}
	case *ast.UnaryExpr:
		e.Expr = r.expr(e.Expr)
	case *ast.PostfixExpr:
		e.Expr = r.expr(e.Expr)
	case *ast.ResultUnwrapExpr:
		e.Expr = r.expr(e.Expr)
	case *ast.CompileTimeExpr:
		r.markCompileTimeExpr(e.Expr)
		e.Expr = r.expr(e.Expr)
		return r.evaluate(e)
	case *ast.BinaryExpr:
		e.Left = r.expr(e.Left)
		e.Right = r.expr(e.Right)
	case *ast.TernaryExpr:
		e.Condition = r.expr(e.Condition)
		e.Consequence = r.expr(e.Consequence)
		e.Alternative = r.expr(e.Alternative)
	case *ast.AssignExpr:
		e.Target = r.expr(e.Target)
		e.Value = r.expr(e.Value)
	case *ast.CallExpr:
		e.Callee = r.expr(e.Callee)
		for idx := range e.Args {
			e.Args[idx] = r.expr(e.Args[idx])
		}
	case *ast.LambdaExpr:
		e.Body = r.expr(e.Body)
	case *ast.SelectorExpr:
		e.Receiver = r.expr(e.Receiver)
	case *ast.IndexExpr:
		e.Receiver = r.expr(e.Receiver)
		e.Index = r.expr(e.Index)
	case *ast.ArrayLiteral:
		for idx := range e.Elements {
			e.Elements[idx] = r.expr(e.Elements[idx])
		}
	case *ast.TupleLiteral:
		for idx := range e.Elements {
			e.Elements[idx] = r.expr(e.Elements[idx])
		}
	case *ast.SpreadExpr:
		e.Expr = r.expr(e.Expr)
	case *ast.ReactiveLiteral:
		e.Value = r.expr(e.Value)
	case *ast.MapLiteral:
		for idx := range e.Entries {
			e.Entries[idx].Key = r.expr(e.Entries[idx].Key)
			e.Entries[idx].Value = r.expr(e.Entries[idx].Value)
		}
	case *ast.StructLiteral:
		for idx := range e.Fields {
			e.Fields[idx].Value = r.expr(e.Fields[idx].Value)
		}
	case *ast.AnonymousObjectLiteral:
		for idx := range e.Fields {
			e.Fields[idx].Value = r.expr(e.Fields[idx].Value)
		}
	case *ast.XMLElement:
		for idx := range e.Attrs {
			e.Attrs[idx].Value = r.expr(e.Attrs[idx].Value)
		}
		for idx := range e.Children {
			e.Children[idx].Expr = r.expr(e.Children[idx].Expr)
		}
	case *ast.BlockExpr:
		for idx := range e.Statements {
			e.Statements[idx] = r.stmt(e.Statements[idx])
		}
	case *ast.PatternBlock:
		for idx := range e.Branches {
			r.pattern(e.Branches[idx].Pattern)
			e.Branches[idx].Expr = r.expr(e.Branches[idx].Expr)
		}
	case *ast.MatchExpr:
		e.Subject = r.expr(e.Subject)
		for idx := range e.Branches {
			r.pattern(e.Branches[idx].Pattern)
			e.Branches[idx].Expr = r.expr(e.Branches[idx].Expr)
		}
	case *ast.WatchExpr:
		e.Target = r.expr(e.Target)
		e.Handler = r.expr(e.Handler)
	}
	return expr
}

func (r *compileTimeRewriter) markCompileTimeExpr(expr ast.Expr) {
	r.markResolvedFunctions(expr, r.markCompileTimeFunction)
}

func (r *compileTimeRewriter) markCompileTimeFunction(fn *ast.Function) {
	if fn == nil || r.compileTimeFunctions[fn] {
		return
	}
	r.compileTimeFunctions[fn] = true
	r.markResolvedFunctions(fn.Body, r.markCompileTimeFunction)
}

func (r *compileTimeRewriter) markResolvedFunctions(expr ast.Expr, mark func(*ast.Function)) {
	if expr == nil || r.info == nil {
		return
	}
	ast.WalkExpr(expr, func(expr ast.Expr) {
		switch e := expr.(type) {
		case *ast.Identifier:
			if fn := r.info.ResolvedFunctions[e]; fn != nil {
				mark(fn.Node)
			}
		case *ast.SelectorExpr:
			if fn := r.info.ResolvedSelectorFunctions[e]; fn != nil {
				mark(fn.Node)
			}
		}
	})
}

func (r *compileTimeRewriter) pattern(pattern ast.Pattern) {
	switch p := pattern.(type) {
	case *ast.LiteralPattern:
		p.Value = r.expr(p.Value)
	case *ast.ComparePattern:
		p.Value = r.expr(p.Value)
	case *ast.RangePattern:
		p.Start = r.expr(p.Start)
		p.End = r.expr(p.End)
	case *ast.OrPattern:
		for _, elem := range p.Alternatives {
			r.pattern(elem)
		}
	case *ast.TuplePattern:
		for _, elem := range p.Elements {
			r.pattern(elem)
		}
	case *ast.MapPattern:
		for idx := range p.Entries {
			p.Entries[idx].Key = r.expr(p.Entries[idx].Key)
			r.pattern(p.Entries[idx].Pattern)
		}
	case *ast.ObjectPattern:
		for idx := range p.Fields {
			r.pattern(p.Fields[idx].Pattern)
		}
	}
}

func (r *compileTimeRewriter) evaluate(expr *ast.CompileTimeExpr) ast.Expr {
	value, err := r.runtime.Eval(ir.LowerExpr(expr.Expr, r.info))
	if err != nil {
		r.diags = append(r.diags, checker.Diagnostic{
			Message: fmt.Sprintf("compile-time expression failed: %v", err),
			Pos:     expr.MarkPos,
		})
		return expr
	}
	literal, err := compileTimeValueExpr(value, expr.Pos)
	if err != nil {
		r.diags = append(r.diags, checker.Diagnostic{
			Message: fmt.Sprintf("compile-time expression returned unsupported value: %v", err),
			Pos:     expr.MarkPos,
		})
		return expr
	}
	r.changed = true
	return literal
}

func compileTimeValueExpr(value interpreter.Value, pos lexer.Position) (ast.Expr, error) {
	switch v := value.(type) {
	case nil:
		return &ast.NullLiteral{Pos: pos}, nil
	case int:
		return &ast.IntegerLiteral{Value: v, Pos: pos}, nil
	case int8:
		return &ast.IntegerLiteral{Value: int(v), Pos: pos}, nil
	case int16:
		return &ast.IntegerLiteral{Value: int(v), Pos: pos}, nil
	case int32:
		return &ast.IntegerLiteral{Value: int(v), Pos: pos}, nil
	case int64:
		if v < math.MinInt || v > math.MaxInt {
			return nil, fmt.Errorf("integer %d is outside Int range", v)
		}
		return &ast.IntegerLiteral{Value: int(v), Pos: pos}, nil
	case uint:
		if uint64(v) > uint64(math.MaxInt) {
			return nil, fmt.Errorf("integer %d is outside Int range", v)
		}
		return &ast.IntegerLiteral{Value: int(v), Pos: pos}, nil
	case uint8:
		return &ast.IntegerLiteral{Value: int(v), Pos: pos}, nil
	case uint16:
		return &ast.IntegerLiteral{Value: int(v), Pos: pos}, nil
	case uint32:
		if uint64(v) > uint64(math.MaxInt) {
			return nil, fmt.Errorf("integer %d is outside Int range", v)
		}
		return &ast.IntegerLiteral{Value: int(v), Pos: pos}, nil
	case uint64:
		if v > uint64(math.MaxInt) {
			return nil, fmt.Errorf("integer %d is outside Int range", v)
		}
		return &ast.IntegerLiteral{Value: int(v), Pos: pos}, nil
	case float32:
		return &ast.DoubleLiteral{Value: float64(v), Pos: pos}, nil
	case float64:
		return &ast.DoubleLiteral{Value: v, Pos: pos}, nil
	case *big.Int:
		return &ast.BigIntLiteral{Value: v.String(), Pos: pos}, nil
	case string:
		return &ast.StringLiteral{Value: v, Pos: pos}, nil
	case interpreter.Char:
		return &ast.CharLiteral{Value: rune(v), Pos: pos}, nil
	case bool:
		return &ast.BoolLiteral{Value: v, Pos: pos}, nil
	case *interpreter.Regex:
		return &ast.RegexLiteral{Pattern: v.Source, Flags: v.Flags, Raw: "/" + v.Source + "/" + v.Flags, Pos: pos}, nil
	case *interpreter.Array:
		out := &ast.ArrayLiteral{Pos: pos}
		for _, elem := range v.Elements {
			expr, err := compileTimeValueExpr(elem, pos)
			if err != nil {
				return nil, err
			}
			out.Elements = append(out.Elements, expr)
		}
		return out, nil
	case *interpreter.Tuple:
		out := &ast.TupleLiteral{Pos: pos}
		for _, elem := range v.Elements {
			expr, err := compileTimeValueExpr(elem, pos)
			if err != nil {
				return nil, err
			}
			out.Elements = append(out.Elements, expr)
		}
		return out, nil
	case *interpreter.Map:
		out := &ast.MapLiteral{Pos: pos}
		keys := make([]string, 0, len(v.Entries))
		for key := range v.Entries {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			entry := v.Entries[key]
			keyExpr, err := compileTimeValueExpr(entry.Key, pos)
			if err != nil {
				return nil, err
			}
			valueExpr, err := compileTimeValueExpr(entry.Value, pos)
			if err != nil {
				return nil, err
			}
			out.Entries = append(out.Entries, ast.MapEntry{Key: keyExpr, Value: valueExpr, Pos: pos})
		}
		return out, nil
	case *interpreter.Struct:
		if v == interpreter.NullValue {
			return &ast.NullLiteral{Pos: pos}, nil
		}
		out := &ast.StructLiteral{TypeName: v.TypeName, Pos: pos}
		names := append([]string(nil), v.Order...)
		if len(names) == 0 {
			for name := range v.Fields {
				names = append(names, name)
			}
			sort.Strings(names)
		}
		for _, name := range names {
			fieldValue, ok := v.Fields[name]
			if !ok {
				continue
			}
			expr, err := compileTimeValueExpr(fieldValue, pos)
			if err != nil {
				return nil, err
			}
			out.Fields = append(out.Fields, ast.FieldValue{Name: name, Value: expr, Pos: pos})
		}
		return out, nil
	default:
		if value == interpreter.NullValue {
			return &ast.NullLiteral{Pos: pos}, nil
		}
		return nil, fmt.Errorf("%T", value)
	}
}

func hasCompileTimeExpr(file *ast.File) bool {
	found := false
	visit := func(expr ast.Expr) {
		if _, ok := expr.(*ast.CompileTimeExpr); ok {
			found = true
		}
	}
	for _, trait := range file.Traits {
		for _, method := range trait.Methods {
			ast.WalkExpr(method.Body, visit)
		}
	}
	for _, typ := range file.Types {
		for idx := range typ.Annotations {
			for _, arg := range typ.Annotations[idx].Args {
				ast.WalkExpr(arg, visit)
			}
		}
		for idx := range typ.Fields {
			for annotationIdx := range typ.Fields[idx].Annotations {
				for _, arg := range typ.Fields[idx].Annotations[annotationIdx].Args {
					ast.WalkExpr(arg, visit)
				}
			}
		}
		for _, method := range typ.Methods {
			ast.WalkExpr(method.Body, visit)
		}
	}
	for _, enum := range file.Enums {
		for idx := range enum.Annotations {
			for _, arg := range enum.Annotations[idx].Args {
				ast.WalkExpr(arg, visit)
			}
		}
		for idx := range enum.Members {
			for annotationIdx := range enum.Members[idx].Annotations {
				for _, arg := range enum.Members[idx].Annotations[annotationIdx].Args {
					ast.WalkExpr(arg, visit)
				}
			}
		}
	}
	for _, fn := range file.Functions {
		for idx := range fn.Annotations {
			for _, arg := range fn.Annotations[idx].Args {
				ast.WalkExpr(arg, visit)
			}
		}
		ast.WalkExpr(fn.Body, visit)
	}
	for _, test := range file.Tests {
		ast.WalkExpr(test.Body, visit)
	}
	return found
}
