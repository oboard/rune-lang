package checker

import (
	"fmt"

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
	Name    string
	Fields  []FieldInfo
	ByName  map[string]FieldInfo
	Methods map[string]*FuncInfo
	Node    *ast.StructType
}

type Info struct {
	Functions map[string]*FuncInfo
	Types     map[string]*StructInfo
	Stdlib    *stdlib.Registry
}

func Check(file *ast.File) (*Info, []Diagnostic) {
	reg, err := stdlib.LoadDefault()
	c := &checker{
		info: &Info{
			Functions: map[string]*FuncInfo{},
			Types:     map[string]*StructInfo{},
			Stdlib:    reg,
		},
	}
	if err != nil {
		c.errorf(lexer.Position{}, "%s", err.Error())
	}
	c.checkGoImports(file)
	c.collect(file)
	for _, typ := range file.Types {
		c.inferMethods(typ)
	}
	for _, fn := range file.Functions {
		c.inferFunction(fn)
	}
	return c.info, c.diags
}

type checker struct {
	info  *Info
	diags []Diagnostic
}

func (c *checker) checkGoImports(file *ast.File) {
	for _, imp := range file.GoImports {
		fn, ok := c.info.Stdlib.Function("go", "import")
		if !ok || fn.Intrinsic != "go.import" || !fn.TopLevelOnly {
			c.errorf(imp.Pos, "@go.import is not declared in core/go")
		}
	}
}

func (c *checker) collect(file *ast.File) {
	for _, typ := range file.Types {
		if _, exists := c.info.Types[typ.Name]; exists {
			c.errorf(typ.NamePos, "duplicate type %q", typ.Name)
			continue
		}
		c.info.Types[typ.Name] = &StructInfo{
			Name:    typ.Name,
			ByName:  map[string]FieldInfo{},
			Methods: map[string]*FuncInfo{},
			Node:    typ,
		}
	}
	for _, typ := range file.Types {
		info := c.info.Types[typ.Name]
		if info == nil {
			continue
		}
		for _, field := range typ.Fields {
			if _, exists := info.ByName[field.Name]; exists {
				c.errorf(field.Pos, "duplicate field %q", field.Name)
				continue
			}
			fieldType := c.resolveType(field.Type)
			if fieldType == Unknown {
				c.errorf(field.Pos, "unknown type %q", field.Type)
			}
			fieldInfo := FieldInfo{Name: field.Name, Type: fieldType}
			info.Fields = append(info.Fields, fieldInfo)
			info.ByName[field.Name] = fieldInfo
		}
		for _, method := range typ.Methods {
			if _, exists := info.Methods[method.Name]; exists {
				c.errorf(method.NamePos, "duplicate method %s.%s", typ.Name, method.Name)
				continue
			}
			methodInfo := c.collectFunction(method)
			methodInfo.ReceiverType = Type(typ.Name)
			info.Methods[method.Name] = methodInfo
		}
	}
	for _, fn := range file.Functions {
		if _, exists := c.info.Functions[fn.Name]; exists {
			c.errorf(fn.NamePos, "duplicate function %q", fn.Name)
			continue
		}
		c.info.Functions[fn.Name] = c.collectFunction(fn)
	}
}

func (c *checker) collectFunction(fn *ast.Function) *FuncInfo {
	info := &FuncInfo{Name: fn.Name, Node: fn, Return: Unknown}
	seenParams := map[string]bool{}
	for _, param := range fn.Params {
		if seenParams[param.Name] {
			c.errorf(param.Pos, "duplicate parameter %q", param.Name)
		}
		seenParams[param.Name] = true
		typ := c.resolveType(param.Type)
		if typ == Unknown {
			c.errorf(param.Pos, "unknown type %q", param.Type)
		}
		info.Params = append(info.Params, ParamInfo{Name: param.Name, Type: typ})
	}
	if fn.ReturnType != "" {
		typ := c.resolveType(fn.ReturnType)
		if typ == Unknown {
			c.errorf(fn.NamePos, "unknown return type %q", fn.ReturnType)
		}
		info.Return = typ
		info.ReturnDeclared = true
	}
	return info
}

func (c *checker) inferMethods(typ *ast.StructType) {
	structInfo := c.info.Types[typ.Name]
	if structInfo == nil {
		return
	}
	for _, method := range typ.Methods {
		info := structInfo.Methods[method.Name]
		if info == nil {
			continue
		}
		env := map[string]Type{"this": Type(typ.Name)}
		for _, param := range info.Params {
			env[param.Name] = param.Type
		}
		ret := c.inferExpr(method.Body, env)
		c.finishFunctionReturn(info, ret, method)
	}
}

func (c *checker) inferFunction(fn *ast.Function) {
	info := c.info.Functions[fn.Name]
	if info == nil {
		return
	}
	env := map[string]Type{}
	for _, param := range info.Params {
		env[param.Name] = param.Type
	}
	ret := c.inferExpr(fn.Body, env)
	if fn.Name == "main" {
		if info.ReturnDeclared && info.Return != Void {
			c.errorf(fn.NamePos, "main must return Void, got %s", info.Return)
		}
		info.Return = Void
		return
	}
	c.finishFunctionReturn(info, ret, fn)
}

func (c *checker) finishFunctionReturn(info *FuncInfo, ret Type, fn *ast.Function) {
	if info.ReturnDeclared {
		if ret != Unknown && ret != info.Return {
			c.errorf(fn.Body.Position(), "function %q returns %s, expected %s", fn.Name, ret, info.Return)
		}
		return
	}
	if ret == Unknown {
		ret = Void
	}
	info.Return = ret
}

func (c *checker) inferExpr(expr ast.Expr, env map[string]Type) Type {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return Int
	case *ast.StringLiteral:
		return String
	case *ast.BoolLiteral:
		return Bool
	case *ast.Identifier:
		if typ, ok := env[e.Name]; ok {
			return typ
		}
		if _, ok := c.info.Functions[e.Name]; ok {
			return Unknown
		}
		if e.Name != "<error>" {
			c.errorf(e.Pos, "undefined name %q", e.Name)
		}
		return Unknown
	case *ast.AtExpr:
		return Unknown
	case *ast.ThisExpr:
		typ, ok := env["this"]
		if !ok {
			c.errorf(e.Pos, "implicit this selector can only be used inside a method")
			return Unknown
		}
		return typ
	case *ast.SelectorExpr:
		receiver := c.inferExpr(e.Receiver, env)
		structInfo := c.info.Types[string(receiver)]
		if structInfo == nil {
			if receiver != Unknown {
				c.errorf(e.Pos, "type %s has no fields", receiver)
			}
			return Unknown
		}
		field, ok := structInfo.ByName[e.Name]
		if !ok {
			c.errorf(e.Pos, "type %s has no field %q", receiver, e.Name)
			return Unknown
		}
		return field.Type
	case *ast.StructLiteral:
		return c.inferStructLiteral(e, env)
	case *ast.UnaryExpr:
		typ := c.inferExpr(e.Expr, env)
		switch e.Op {
		case lexer.Minus:
			if typ != Int && typ != Unknown {
				c.errorf(e.Pos, "operator '-' expects Int, got %s", typ)
			}
			return Int
		case lexer.Bang:
			if typ != Bool && typ != Unknown {
				c.errorf(e.Pos, "operator '!' expects Bool, got %s", typ)
			}
			return Bool
		default:
			return Unknown
		}
	case *ast.BinaryExpr:
		left := c.inferExpr(e.Left, env)
		right := c.inferExpr(e.Right, env)
		switch e.Op {
		case lexer.Plus, lexer.Minus, lexer.Star, lexer.Slash, lexer.Percent:
			if left != Int && left != Unknown {
				c.errorf(e.Left.Position(), "arithmetic expects Int, got %s", left)
			}
			if right != Int && right != Unknown {
				c.errorf(e.Right.Position(), "arithmetic expects Int, got %s", right)
			}
			return Int
		case lexer.EqualEqual, lexer.BangEqual, lexer.Less, lexer.LessEqual, lexer.Greater, lexer.GreaterEqual:
			return Bool
		default:
			return Unknown
		}
	case *ast.CallExpr:
		return c.inferCall(e, env)
	case *ast.BlockExpr:
		return c.inferBlock(e, env)
	case *ast.PatternBlock:
		return c.inferPatternBlock(e, env)
	default:
		return Unknown
	}
}

func (c *checker) inferCall(call *ast.CallExpr, env map[string]Type) Type {
	argTypes := make([]Type, 0, len(call.Args))
	for _, arg := range call.Args {
		argTypes = append(argTypes, c.inferExpr(arg, env))
	}
	if sel, ok := call.Callee.(*ast.SelectorExpr); ok {
		if at, ok := sel.Receiver.(*ast.AtExpr); ok {
			if fn, ok := c.info.Stdlib.Function(at.Name, sel.Name); ok {
				return c.inferStdlibCall(at.Name, sel, call, argTypes, fn)
			}
			c.errorf(sel.Pos, "unknown module function @%s.%s", at.Name, sel.Name)
			return Unknown
		}
		receiver := c.inferExpr(sel.Receiver, env)
		structInfo := c.info.Types[string(receiver)]
		if structInfo == nil {
			if receiver != Unknown {
				c.errorf(sel.Pos, "type %s has no methods", receiver)
			}
			return Unknown
		}
		method := structInfo.Methods[sel.Name]
		if method == nil {
			c.errorf(sel.Pos, "type %s has no method %q", receiver, sel.Name)
			return Unknown
		}
		c.checkArgs(sel.Name, method.Params, call.Args, argTypes, sel.Pos)
		return method.Return
	}
	if ident, ok := call.Callee.(*ast.Identifier); ok {
		fn := c.info.Functions[ident.Name]
		if fn == nil {
			c.errorf(ident.Pos, "undefined function %q", ident.Name)
			return Unknown
		}
		c.checkArgs(ident.Name, fn.Params, call.Args, argTypes, ident.Pos)
		return fn.Return
	}
	c.inferExpr(call.Callee, env)
	return Unknown
}

func (c *checker) inferStdlibCall(moduleName string, sel *ast.SelectorExpr, call *ast.CallExpr, argTypes []Type, fn *stdlib.Function) Type {
	if fn.TopLevelOnly {
		c.errorf(sel.Pos, "@%s.%s must be a top-level declaration", moduleName, sel.Name)
		return Void
	}

	switch fn.Intrinsic {
	case "go.stmt":
		c.checkStdlibArgs(moduleName, sel.Name, fn, call.Args, argTypes, sel.Pos)
		if len(call.Args) == 1 {
			if _, ok := call.Args[0].(*ast.StringLiteral); !ok {
				c.errorf(call.Args[0].Position(), "@go.stmt argument must be a string literal")
			}
		}
		return Void
	case "go.expr":
		c.checkStdlibArgs(moduleName, sel.Name, fn, call.Args, argTypes, sel.Pos)
		if len(call.Args) != 1 {
			return Unknown
		}
		if _, ok := call.Args[0].(*ast.StringLiteral); !ok {
			c.errorf(call.Args[0].Position(), "@go.expr body must be a string literal")
		}
		return c.resolveDeclaredReturn(fn.Return)
	}

	c.checkStdlibArgs(moduleName, sel.Name, fn, call.Args, argTypes, sel.Pos)
	return c.resolveDeclaredReturn(fn.Return)
}

func (c *checker) checkStdlibArgs(moduleName string, functionName string, fn *stdlib.Function, args []ast.Expr, argTypes []Type, pos lexer.Position) {
	if fn.Variadic {
		minArgs := len(fn.Params)
		if minArgs > 0 {
			minArgs--
		}
		if len(args) < minArgs {
			c.errorf(pos, "@%s.%s expects at least %d args, got %d", moduleName, functionName, minArgs, len(args))
		}
	} else if len(fn.Params) != len(args) {
		c.errorf(pos, "@%s.%s expects %d args, got %d", moduleName, functionName, len(fn.Params), len(args))
	}

	for i := 0; i < len(args) && i < len(fn.Params); i++ {
		expected := fn.Params[i]
		if fn.Variadic && i >= len(fn.Params)-1 {
			expected = fn.Params[len(fn.Params)-1]
		}
		c.checkDeclaredArg(moduleName, functionName, i, expected, args[i], argTypes[i])
	}
}

func (c *checker) checkDeclaredArg(moduleName string, functionName string, index int, expected string, arg ast.Expr, actual Type) {
	if expected == "Any" || expected == "Dynamic" {
		return
	}
	expectedType := c.resolveType(expected)
	if expectedType == Unknown {
		return
	}
	if actual != Unknown && actual != expectedType {
		c.errorf(arg.Position(), "argument %d to @%s.%s has type %s, expected %s", index+1, moduleName, functionName, actual, expectedType)
	}
}

func (c *checker) checkArgs(name string, params []ParamInfo, args []ast.Expr, argTypes []Type, pos lexer.Position) {
	if len(params) != len(args) {
		c.errorf(pos, "function %q expects %d args, got %d", name, len(params), len(args))
	}
	limit := min(len(params), len(argTypes))
	for i := 0; i < limit; i++ {
		if argTypes[i] != Unknown && params[i].Type != Unknown && argTypes[i] != params[i].Type {
			c.errorf(args[i].Position(), "argument %d to %q has type %s, expected %s", i+1, name, argTypes[i], params[i].Type)
		}
	}
}

func (c *checker) inferStructLiteral(lit *ast.StructLiteral, env map[string]Type) Type {
	structInfo := c.info.Types[lit.TypeName]
	if structInfo == nil {
		c.errorf(lit.Pos, "unknown type %q", lit.TypeName)
		for _, field := range lit.Fields {
			c.inferExpr(field.Value, env)
		}
		return Unknown
	}

	seen := map[string]bool{}
	for _, field := range lit.Fields {
		fieldInfo, ok := structInfo.ByName[field.Name]
		if !ok {
			c.errorf(field.Pos, "type %s has no field %q", lit.TypeName, field.Name)
			c.inferExpr(field.Value, env)
			continue
		}
		if seen[field.Name] {
			c.errorf(field.Pos, "duplicate field value %q", field.Name)
		}
		seen[field.Name] = true
		valueType := c.inferExpr(field.Value, env)
		if valueType != Unknown && fieldInfo.Type != Unknown && valueType != fieldInfo.Type {
			c.errorf(field.Value.Position(), "field %s.%s has type %s, expected %s", lit.TypeName, field.Name, valueType, fieldInfo.Type)
		}
	}
	for _, field := range structInfo.Fields {
		if !seen[field.Name] {
			c.errorf(lit.Pos, "missing field %s.%s", lit.TypeName, field.Name)
		}
	}
	return Type(lit.TypeName)
}

func (c *checker) inferBlock(block *ast.BlockExpr, env map[string]Type) Type {
	local := cloneEnv(env)
	result := Void
	for _, stmt := range block.Statements {
		switch s := stmt.(type) {
		case *ast.LetStmt:
			if _, exists := local[s.Name]; exists {
				c.errorf(s.Pos, "name %q is already defined", s.Name)
			}
			local[s.Name] = c.inferExpr(s.Value, local)
			result = Void
		case *ast.AssignStmt:
			if _, exists := local[s.Name]; !exists {
				c.errorf(s.Pos, "cannot assign undefined name %q", s.Name)
			}
			c.inferExpr(s.Value, local)
			result = Void
		case *ast.ExprStmt:
			result = c.inferExpr(s.Expr, local)
		}
	}
	return result
}

func (c *checker) inferPatternBlock(block *ast.PatternBlock, env map[string]Type) Type {
	result := Unknown
	for _, branch := range block.Branches {
		c.checkPattern(branch.Pattern, env)
		typ := c.inferExpr(branch.Expr, env)
		if result == Unknown {
			result = typ
			continue
		}
		if typ != Unknown && result != typ {
			c.errorf(branch.Expr.Position(), "pattern branch returns %s, expected %s", typ, result)
		}
	}
	return result
}

func (c *checker) checkPattern(pattern ast.Pattern, env map[string]Type) {
	switch p := pattern.(type) {
	case *ast.LiteralPattern:
		c.inferExpr(p.Value, env)
	case *ast.ComparePattern:
		typ := c.inferExpr(p.Value, env)
		if typ != Int && typ != Unknown {
			c.errorf(p.Pos, "comparison pattern expects Int literal")
		}
	case *ast.TuplePattern:
		for _, elem := range p.Elements {
			c.checkPattern(elem, env)
		}
	}
}

func (c *checker) resolveType(name string) Type {
	switch name {
	case "Int":
		return Int
	case "String":
		return String
	case "Bool":
		return Bool
	case "Void":
		return Void
	default:
		if _, ok := c.info.Types[name]; ok {
			return Type(name)
		}
		return Unknown
	}
}

func (c *checker) resolveDeclaredReturn(name string) Type {
	if name == "Dynamic" || name == "Any" {
		return Unknown
	}
	return c.resolveType(name)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func cloneEnv(env map[string]Type) map[string]Type {
	out := make(map[string]Type, len(env))
	for k, v := range env {
		out[k] = v
	}
	return out
}

func (c *checker) errorf(pos lexer.Position, format string, args ...any) {
	c.diags = append(c.diags, Diagnostic{Message: fmt.Sprintf(format, args...), Pos: pos})
}
