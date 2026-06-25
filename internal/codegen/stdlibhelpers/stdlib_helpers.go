package stdlibhelpers

import (
	"sort"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func HelperName(moduleName string, functionName string) string {
	return moduleName + "_" + functionName
}

func BodyHelpers(file *ir.File) []*ir.Function {
	if file == nil || file.Stdlib == nil {
		return nil
	}
	requested := map[string]map[string]*stdlib.Function{}
	var request func(moduleName string, fn *stdlib.Function, allowIntrinsic bool)
	request = func(moduleName string, fn *stdlib.Function, allowIntrinsic bool) {
		if fn == nil || fn.Receiver != "" || moduleName == "cli" || moduleName == "iter" {
			return
		}
		if fn.Body == nil && (!allowIntrinsic || fn.Intrinsic == "") {
			return
		}
		if requested[moduleName] == nil {
			requested[moduleName] = map[string]*stdlib.Function{}
		}
		if requested[moduleName][fn.Name] != nil {
			return
		}
		requested[moduleName][fn.Name] = fn
		if fn.Body == nil {
			return
		}
		ast.WalkExpr(fn.Body, func(expr ast.Expr) {
			call, ok := expr.(*ast.CallExpr)
			if !ok {
				return
			}
			if ident, ok := call.Callee.(*ast.Identifier); ok {
				if next, ok := file.Stdlib.Function(moduleName, ident.Name); ok {
					request(moduleName, next, true)
				}
				return
			}
			sel, ok := call.Callee.(*ast.SelectorExpr)
			if !ok {
				return
			}
			if at, ok := sel.Receiver.(*ast.AtExpr); ok && at.Name != "" {
				if next, ok := file.Stdlib.Function(at.Name, sel.Name); ok {
					request(at.Name, next, false)
				}
			}
		})
	}
	collectIRExpr := func(expr ir.Expr) {
		ir.WalkExpr(expr, func(expr ir.Expr) {
			call, ok := expr.(*ir.CallExpr)
			if !ok {
				return
			}
			sel, ok := call.Callee.(*ir.SelectorExpr)
			if !ok {
				return
			}
			moduleName, ok := selectorModuleName(sel)
			if !ok {
				return
			}
			if fn, ok := file.Stdlib.Function(moduleName, sel.Name); ok && fn.Body != nil {
				request(moduleName, fn, false)
			}
		})
	}
	for _, fn := range file.Functions {
		collectIRExpr(fn.Body)
	}
	for _, test := range file.Tests {
		collectIRExpr(test.Body)
	}
	for _, typ := range file.Types {
		for _, method := range typ.Methods {
			collectIRExpr(method.Body)
		}
	}
	modules := make([]string, 0, len(requested))
	for moduleName := range requested {
		modules = append(modules, moduleName)
	}
	sort.Strings(modules)
	var out []*ir.Function
	for _, moduleName := range modules {
		names := make([]string, 0, len(requested[moduleName]))
		for name := range requested[moduleName] {
			names = append(names, name)
		}
		sort.Strings(names)
		helperFile := &ast.File{}
		for _, name := range names {
			fn := requested[moduleName][name]
			helperFile.Functions = append(helperFile.Functions, astFunctionFromStdlib(moduleName, fn))
		}
		info, diags := checker.CheckWithStdlib(helperFile, file.Stdlib)
		if len(diags) > 0 {
			continue
		}
		helpers := ir.LowerFile(helperFile, info).Functions
		renameModuleHelpers(moduleName, helpers, requested[moduleName])
		out = append(out, helpers...)
	}
	return out
}

func selectorModuleName(sel *ir.SelectorExpr) (string, bool) {
	if at, ok := sel.Receiver.(*ir.AtExpr); ok && at.Name != "" {
		return at.Name, true
	}
	return checker.ModuleNamespaceName(sel.Receiver.ResultType())
}

func astFunctionFromStdlib(moduleName string, fn *stdlib.Function) *ast.Function {
	body := fn.Body
	if body == nil && fn.Intrinsic != "" {
		body = &ast.CallExpr{
			Callee: &ast.SelectorExpr{
				Receiver: &ast.AtExpr{Name: moduleName, Pos: fn.Pos},
				Name:     fn.Name,
				Pos:      fn.Pos,
				NamePos:  fn.Pos,
			},
			Pos: fn.Pos,
		}
	}
	out := &ast.Function{
		Name:       fn.Name,
		Routine:    fn.Routine,
		Generics:   append([]string(nil), fn.Generics...),
		ReturnType: ast.RawType(fn.Return),
		Body:       body,
		Pos:        fn.Pos,
		NamePos:    fn.Pos,
		SourcePath: fn.SourcePath,
	}
	if len(fn.GenericConstraints) > 0 {
		out.GenericConstraints = map[string]ast.Type{}
		for name, constraint := range fn.GenericConstraints {
			out.GenericConstraints[name] = ast.RawType(constraint)
		}
	}
	for idx, paramType := range fn.Params {
		name := ""
		if idx < len(fn.ParamNames) {
			name = fn.ParamNames[idx]
		}
		out.Params = append(out.Params, ast.Param{Name: name, Type: ast.RawType(paramType), Pos: fn.Pos})
	}
	return out
}

func renameModuleHelpers(moduleName string, helpers []*ir.Function, requested map[string]*stdlib.Function) {
	for _, fn := range helpers {
		renameBodyHelperCalls(moduleName, fn.Body, requested)
		fn.Name = HelperName(moduleName, fn.Name)
		fn.SourceName = fn.Name
		fn.Private = true
	}
}

func renameBodyHelperCalls(moduleName string, expr ir.Expr, requested map[string]*stdlib.Function) {
	ir.WalkExpr(expr, func(expr ir.Expr) {
		call, ok := expr.(*ir.CallExpr)
		if !ok {
			return
		}
		if ident, ok := call.Callee.(*ir.Identifier); ok {
			if requested[ident.Name] != nil {
				ident.Name = HelperName(moduleName, ident.Name)
			}
			return
		}
		sel, ok := call.Callee.(*ir.SelectorExpr)
		if !ok {
			return
		}
		targetModule, ok := selectorModuleName(sel)
		if !ok || targetModule != moduleName {
			return
		}
		if fn := requested[sel.Name]; fn != nil && fn.Body != nil {
			call.Callee = &ir.Identifier{ExprBase: ir.ExprBase{Pos: sel.Pos, Type: sel.ResultType()}, Name: HelperName(moduleName, sel.Name)}
		}
	})
}
