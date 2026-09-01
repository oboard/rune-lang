package stdlibhelpers

import (
	"sort"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
	"github.com/oboard/rune-lang/internal/stdlib"
)

// HelperName returns the helper link name for a Rune core function. Backends
// apply their regular private linkage when they emit this IR symbol.
func HelperName(moduleName string, functionName string) string {
	return moduleName + "_" + functionName
}

// Closure is the compilable portion of core required by a lowered user file.
// It contains declarations as well as functions: core functions routinely use
// records and enum constructors that cannot be supplied by a function-only
// helper list.
type Closure struct {
	Types     []*ir.StructType
	Enums     []*ir.EnumType
	Constants []*ir.ConstDecl
	Functions []*ir.Function
}

func (c Closure) Empty() bool {
	return len(c.Types) == 0 && len(c.Enums) == 0 && len(c.Constants) == 0 && len(c.Functions) == 0
}

// With merges the closure ahead of file. Core declarations are private and its
// functions are uniquely prefixed, so they cannot be exported accidentally.
func (c Closure) With(file *ir.File) *ir.File {
	if file == nil || c.Empty() {
		return file
	}
	out := *file
	out.Types = append(append([]*ir.StructType{}, c.Types...), file.Types...)
	out.Enums = append(append([]*ir.EnumType{}, c.Enums...), file.Enums...)
	out.Constants = append(append([]*ir.ConstDecl{}, c.Constants...), file.Constants...)
	out.Functions = append(append([]*ir.Function{}, c.Functions...), file.Functions...)
	return &out
}

// BodyHelpers remains for callers which only need helper functions.
func BodyHelpers(file *ir.File) []*ir.Function { return Collect(file).Functions }

// Collect discovers the transitive body-backed core dependency closure. It
// lowers a synthetic core file so declarations and function bodies share the
// same checker view, then rewrites selected @module.function calls in the user
// IR to the private emitted helper symbols.
func Collect(file *ir.File) Closure {
	if file == nil || file.Stdlib == nil {
		return Closure{}
	}
	requested := map[string]map[string]*stdlib.Function{}
	var request func(string, *stdlib.Function, bool)
	request = func(moduleName string, fn *stdlib.Function, allowIntrinsic bool) {
		if fn == nil || fn.Receiver != "" || fn.Macro || isBackendStreamIntrinsic(fn.Intrinsic) || (fn.Body == nil && (!allowIntrinsic || fn.Intrinsic == "")) {
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
			if sel, ok := call.Callee.(*ast.SelectorExpr); ok {
				if at, ok := sel.Receiver.(*ast.AtExpr); ok && at.Name != "" {
					if next, ok := file.Stdlib.Function(at.Name, sel.Name); ok {
						request(at.Name, next, false)
					}
				}
			}
		})
	}
	collectExpr := func(expr ir.Expr) {
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
			if fn, ok := file.Stdlib.Function(moduleName, sel.Name); ok {
				request(moduleName, fn, false)
			}
		})
	}
	for _, constant := range file.Constants {
		collectExpr(constant.Value)
	}
	for _, fn := range file.Functions {
		collectExpr(fn.Body)
	}
	for _, test := range file.Tests {
		collectExpr(test.Body)
	}
	for _, typ := range file.Types {
		for _, method := range typ.Methods {
			collectExpr(method.Body)
		}
	}
	for _, enum := range file.Enums {
		for _, method := range enum.Methods {
			collectExpr(method.Body)
		}
	}
	if len(requested) == 0 {
		return Closure{}
	}

	modules := make([]string, 0, len(requested))
	for moduleName := range requested {
		modules = append(modules, moduleName)
	}
	sort.Strings(modules)
	helperFile := &ast.File{}
	for _, moduleName := range modules {
		mod := file.Stdlib.Modules[moduleName]
		if mod == nil {
			continue
		}
		// Types are module-scoped in core. Include the module's declarations so
		// every selected function body has its record/constructor definitions.
		for i := range mod.Types {
			typ := &mod.Types[i]
			if len(typ.Constructors) > 0 {
				helperFile.Enums = append(helperFile.Enums, astEnumFromStdlib(typ))
			} else {
				helperFile.Types = append(helperFile.Types, astTypeFromStdlib(typ))
			}
		}
		names := make([]string, 0, len(requested[moduleName]))
		for name := range requested[moduleName] {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			helperFile.Functions = append(helperFile.Functions, astFunctionFromStdlib(moduleName, requested[moduleName][name]))
		}
	}
	info, diags := checker.CheckWithStdlib(helperFile, file.Stdlib)
	if len(diags) > 0 {
		return Closure{}
	}
	lowered := ir.LowerFile(helperFile, info)
	for _, typ := range lowered.Types {
		typ.Private = true
		for _, method := range typ.Methods {
			method.Private = true
		}
	}
	for _, enum := range lowered.Enums {
		enum.Private = true
		for _, method := range enum.Methods {
			method.Private = true
		}
	}
	renameHelpers(lowered.Functions, requested)
	rewriteSelectedCalls(file, requested)
	return Closure{Types: lowered.Types, Enums: lowered.Enums, Constants: lowered.Constants, Functions: lowered.Functions}
}

func isBackendStreamIntrinsic(intrinsic string) bool {
	return intrinsic == "buffer.new" || intrinsic == "buffer.fromBytes" ||
		intrinsic == "reader.new" || intrinsic == "writer.new" ||
		intrinsic == "writer.withCapacity"
}

func selectorModuleName(sel *ir.SelectorExpr) (string, bool) {
	if at, ok := sel.Receiver.(*ir.AtExpr); ok && at.Name != "" {
		return at.Name, true
	}
	return checker.ModuleNamespaceName(sel.Receiver.ResultType())
}

func astTypeFromStdlib(typ *stdlib.Type) *ast.StructType {
	out := &ast.StructType{Name: typ.Name, Private: true, Generics: append([]string(nil), typ.Generics...), Pos: typ.Pos, NamePos: typ.Pos, SourcePath: typ.SourcePath}
	if len(typ.GenericConstraints) > 0 {
		out.GenericConstraints = map[string]ast.Type{}
		for name, constraint := range typ.GenericConstraints {
			out.GenericConstraints[name] = ast.RawType(constraint)
		}
	}
	for _, field := range typ.Fields {
		out.Fields = append(out.Fields, ast.Field{Name: field.Name, Type: ast.RawType(field.Type), Pos: field.Pos})
	}
	return out
}

func astEnumFromStdlib(typ *stdlib.Type) *ast.EnumType {
	out := &ast.EnumType{Name: typ.Name, Private: true, Generics: append([]string(nil), typ.Generics...), Pos: typ.Pos, NamePos: typ.Pos, SourcePath: typ.SourcePath}
	if len(typ.GenericConstraints) > 0 {
		out.GenericConstraints = map[string]ast.Type{}
		for name, constraint := range typ.GenericConstraints {
			out.GenericConstraints[name] = ast.RawType(constraint)
		}
	}
	for _, constructor := range typ.Constructors {
		member := ast.EnumMember{Name: constructor.Name, Pos: constructor.Pos}
		for index, paramType := range constructor.Params {
			name := ""
			if index < len(constructor.ParamNames) {
				name = constructor.ParamNames[index]
			}
			member.Params = append(member.Params, ast.Param{Name: name, Type: ast.RawType(paramType), Pos: constructor.Pos})
		}
		out.Members = append(out.Members, member)
	}
	return out
}

func astFunctionFromStdlib(moduleName string, fn *stdlib.Function) *ast.Function {
	body := fn.Body
	if body == nil && fn.Intrinsic != "" {
		body = &ast.CallExpr{Callee: &ast.SelectorExpr{Receiver: &ast.AtExpr{Name: moduleName, Pos: fn.Pos}, Name: fn.Name, Pos: fn.Pos, NamePos: fn.Pos}, Pos: fn.Pos}
	}
	// Keep the synthetic declaration public while checking so its original name
	// survives lowering; it is made private after its helper link is installed.
	out := &ast.Function{Name: fn.Name, Routine: fn.Routine, Generics: append([]string(nil), fn.Generics...), ReturnType: ast.RawType(fn.Return), Body: body, Pos: fn.Pos, NamePos: fn.Pos, SourcePath: fn.SourcePath}
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

func renameHelpers(functions []*ir.Function, requested map[string]map[string]*stdlib.Function) {
	for _, fn := range functions {
		for moduleName, byName := range requested {
			if byName[fn.Name] != nil {
				fn.Name = HelperName(moduleName, fn.Name)
				fn.SourceName = fn.Name
				fn.Private = true
				break
			}
		}
		rewriteHelperCalls(fn.Body, requested)
	}
}

func rewriteSelectedCalls(file *ir.File, requested map[string]map[string]*stdlib.Function) {
	visit := func(expr ir.Expr) { rewriteCalls(expr, requested) }
	for _, constant := range file.Constants {
		visit(constant.Value)
	}
	for _, fn := range file.Functions {
		visit(fn.Body)
	}
	for _, test := range file.Tests {
		visit(test.Body)
	}
	for _, typ := range file.Types {
		for _, method := range typ.Methods {
			visit(method.Body)
		}
	}
	for _, enum := range file.Enums {
		for _, method := range enum.Methods {
			visit(method.Body)
		}
	}
}

func rewriteHelperCalls(expr ir.Expr, requested map[string]map[string]*stdlib.Function) {
	rewriteCalls(expr, requested)
}

func rewriteCalls(expr ir.Expr, requested map[string]map[string]*stdlib.Function) {
	ir.WalkExpr(expr, func(node ir.Expr) {
		call, ok := node.(*ir.CallExpr)
		if !ok {
			return
		}
		if ident, ok := call.Callee.(*ir.Identifier); ok {
			for moduleName, byName := range requested {
				if byName[ident.Name] != nil {
					ident.Name = HelperName(moduleName, ident.Name)
					return
				}
			}
			return
		}
		sel, ok := call.Callee.(*ir.SelectorExpr)
		if !ok {
			return
		}
		moduleName, ok := selectorModuleName(sel)
		if !ok || requested[moduleName][sel.Name] == nil {
			return
		}
		call.Callee = &ir.Identifier{ExprBase: ir.ExprBase{Pos: sel.Pos, Type: sel.ResultType()}, Name: HelperName(moduleName, sel.Name)}
	})
}
