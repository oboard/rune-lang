package moonbitcodegen

import (
	"bytes"
	"errors"
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/codegen/stdlibhelpers"
	"github.com/oboard/rune-lang/internal/ir"
)

type generator struct {
	buf           bytes.Buffer
	file          *ir.File
	indent        int
	temp          int
	errors        []error
	thisNames     []string
	useFS         bool
	useCompress   bool
	useRegex      bool
	useString     bool
	useReader     bool
	hasRoutine    bool
	anonTypes     map[string]string
	anonOrder     []checker.Type
	importAliases map[string]bool
}

func Generate(file *ast.File, info *checker.Info) (string, error) {
	return GenerateIR(ir.LowerFile(file, info))
}

func GenerateIR(file *ir.File) (string, error) {
	if len(file.TSImports) > 0 {
		return "", fmt.Errorf("MoonBit backend does not support TypeScript imports")
	}
	if len(file.GoImports) > 0 {
		return "", fmt.Errorf("MoonBit backend does not support Go package imports")
	}
	closure := stdlibhelpers.Collect(file)
	file = closure.With(file)
	g := &generator{file: file, hasRoutine: fileHasRoutine(file), anonTypes: map[string]string{}, importAliases: map[string]bool{}}
	g.collectAnonymousTypes()
	for i, enum := range file.Enums {
		if i > 0 {
			g.line("")
		}
		g.enumType(enum)
		for _, method := range enum.Methods {
			g.line("")
			g.method(enum.Name, method)
		}
	}
	if len(file.Enums) > 0 && len(file.Types) > 0 {
		g.line("")
	}
	for i, typ := range file.Types {
		if i > 0 {
			g.line("")
		}
		g.structType(typ)
		for _, method := range typ.Methods {
			g.line("")
			g.method(typ.Name, method)
		}
	}
	if len(g.anonOrder) > 0 && (len(file.Types) > 0 || len(file.Enums) > 0) {
		g.line("")
	}
	for i, typ := range g.anonOrder {
		if i > 0 {
			g.line("")
		}
		g.anonymousStructType(typ)
	}
	if (len(file.Types) > 0 || len(file.Enums) > 0) && len(file.Functions) > 0 {
		g.line("")
	}
	for i, constant := range file.Constants {
		if i > 0 {
			g.line("")
		}
		g.constDecl(constant)
	}
	if len(file.Constants) > 0 && len(file.Functions) > 0 {
		g.line("")
	}
	for i, fn := range file.Functions {
		if i > 0 {
			g.line("")
		}
		g.function(fn)
	}
	if err := g.codegenError(); err != nil {
		return g.buf.String(), err
	}
	src := g.buf.String()
	if g.useRegex {
		var runtime generator
		runtime.regexRuntime()
		src = runtime.buf.String() + "\n" + src
	}
	if g.useString {
		var runtime generator
		runtime.stringRuntime()
		src = runtime.buf.String() + "\n" + src
	}
	if g.useReader {
		var runtime generator
		runtime.readerRuntime()
		src = runtime.buf.String() + "\n" + src
	}
	if g.useFS || g.useCompress {
		var runtime generator
		runtime.bytesRuntime()
		src = runtime.buf.String() + "\n" + src
	}
	if g.useCompress {
		var runtime generator
		runtime.compressRuntime()
		src = runtime.buf.String() + "\n" + src
	}
	if g.useFS {
		var runtime generator
		runtime.fsRuntime()
		src = runtime.buf.String() + "\n" + src
	}
	return src, nil
}

func (g *generator) line(s string) {
	for i := 0; i < g.indent; i++ {
		g.buf.WriteString("  ")
	}
	g.buf.WriteString(s)
	g.buf.WriteByte('\n')
}

func (g *generator) linef(format string, args ...any) {
	g.line(fmt.Sprintf(format, args...))
}

func (g *generator) addError(err error) {
	g.errors = append(g.errors, err)
}

func (g *generator) codegenError() error {
	return errors.Join(g.errors...)
}

func (g *generator) nextTemp(prefix string) string {
	g.temp++
	return fmt.Sprintf("%s%d", prefix, g.temp)
}

func (g *generator) collectAnonymousTypes() {
	for _, typ := range g.file.Types {
		for _, field := range typ.Fields {
			g.addAnonymousType(field.Type, "")
		}
		for _, method := range typ.Methods {
			g.collectFunctionAnonymousTypes(method)
		}
	}
	for _, enum := range g.file.Enums {
		for _, method := range enum.Methods {
			g.collectFunctionAnonymousTypes(method)
		}
	}
	for _, fn := range g.file.Functions {
		g.collectFunctionAnonymousTypes(fn)
	}
	for _, test := range g.file.Tests {
		g.collectExprAnonymousTypes(test.Body, "")
	}
}

func (g *generator) collectFunctionAnonymousTypes(fn *ir.Function) {
	for _, param := range fn.Params {
		g.addAnonymousType(param.Type, param.Name)
	}
	g.addAnonymousType(fn.Return, fn.Name)
	g.collectExprAnonymousTypes(fn.Body, "")
}

func (g *generator) collectExprAnonymousTypes(expr ir.Expr, binding string) {
	if expr == nil {
		return
	}
	if obj, ok := expr.(*ir.AnonymousObjectLiteral); ok {
		g.addAnonymousType(obj.ResultType(), binding)
	}
	switch e := expr.(type) {
	case *ir.BlockExpr:
		for _, stmt := range e.Statements {
			switch s := stmt.(type) {
			case *ir.LetStmt:
				g.collectExprAnonymousTypes(s.Value, s.Name)
			case *ir.ObjectDestructureStmt:
				g.collectExprAnonymousTypes(s.Value, "")
			case *ir.AssignStmt:
				g.collectExprAnonymousTypes(s.Value, s.Name)
			case *ir.ExprStmt:
				g.collectExprAnonymousTypes(s.Expr, "")
			}
		}
	default:
		ir.WalkExpr(expr, func(child ir.Expr) {
			if child == expr {
				return
			}
			if obj, ok := child.(*ir.AnonymousObjectLiteral); ok {
				g.addAnonymousType(obj.ResultType(), "")
			}
		})
	}
}

func (g *generator) addAnonymousType(typ checker.Type, hint string) {
	if inner, ok := parseNullableType(string(typ)); ok {
		g.addAnonymousType(checker.Type(inner), hint)
		return
	}
	if elem, ok := checker.ArrayElement(typ); ok {
		g.addAnonymousType(elem, hint)
		return
	}
	if base, args, ok := parseGenericType(string(typ)); ok {
		for _, arg := range args {
			g.addAnonymousType(checker.Type(arg), hint)
		}
		if base != "Func" && base != "AsyncFunc" {
			return
		}
	}
	fields, ok := parseObjectType(string(typ))
	if !ok {
		return
	}
	key := string(typ)
	if _, exists := g.anonTypes[key]; exists {
		return
	}
	name := g.anonymousStructName(typ, hint)
	g.anonTypes[key] = name
	g.anonOrder = append(g.anonOrder, typ)
	for _, field := range fields {
		g.addAnonymousType(checker.Type(field.typ), field.name)
	}
}

func (g *generator) anonymousTypeName(typ checker.Type) string {
	if name, ok := g.anonTypes[string(typ)]; ok {
		return name
	}
	return g.anonymousStructName(typ, "")
}

func (g *generator) anonymousStructName(typ checker.Type, hint string) string {
	if hint == "" {
		hint = "Object"
	}
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(string(typ)))
	return mangleType("Anon" + strings.Title(hint) + fmt.Sprintf("%08x", hash.Sum32()))
}

func (g *generator) structType(typ *ir.StructType) {
	g.linef("%sstruct %s {", g.visibilityPrefix(typ.Private), mangleType(typ.Name))
	g.indent++
	for _, field := range typ.Fields {
		g.linef("%s : %s", mangleIdent(field.Name), g.mbtType(field.Type))
	}
	g.indent--
	g.line("}")
}

func (g *generator) anonymousStructType(typ checker.Type) {
	fields, ok := parseObjectType(string(typ))
	if !ok {
		return
	}
	g.linef("pub struct %s {", g.anonymousTypeName(typ))
	g.indent++
	for _, field := range fields {
		g.linef("%s : %s", mangleIdent(field.name), g.mbtType(checker.Type(field.typ)))
	}
	g.indent--
	g.line("}")
}

func (g *generator) enumType(enum *ir.EnumType) {
	g.linef("%senum %s {", g.visibilityPrefix(enum.Private), mangleType(enum.Name))
	g.indent++
	for _, member := range enum.Members {
		params := make([]string, 0, len(member.Params))
		for _, param := range member.Params {
			params = append(params, g.mbtType(param.Type))
		}
		if len(params) == 0 {
			g.line(mangleType(member.Name))
		} else {
			g.linef("%s(%s)", mangleType(member.Name), strings.Join(params, ", "))
		}
	}
	g.indent--
	if enumHasPayloadMembers(enum) {
		g.line("}")
	} else {
		g.line("} derive(Eq)")
	}
	g.line("")
	g.enumShowImpl(enum)
}

func (g *generator) enumShowImpl(enum *ir.EnumType) {
	g.linef("pub impl Show for %s with fn output(self, logger) {", mangleType(enum.Name))
	g.indent++
	g.line("(match self {")
	g.indent++
	for i, member := range enum.Members {
		value := i
		if member.HasValue {
			value = member.Value
		}
		pattern := mangleType(member.Name)
		if len(member.Params) > 0 {
			pattern += "(" + strings.TrimSuffix(strings.Repeat("_, ", len(member.Params)), ", ") + ")"
		}
		g.linef("%s => %d", pattern, value)
	}
	g.indent--
	g.line("}).output(logger)")
	g.indent--
	g.line("}")
}

func (g *generator) constDecl(constant *ir.ConstDecl) {
	g.linef("let %s : %s = %s", mangleIdent(constant.Name), g.mbtType(constant.Type), g.exprAs(constant.Value, constant.Type))
}

func enumHasValueMembers(enum *ir.EnumType) bool {
	for _, member := range enum.Members {
		if member.HasValue {
			return true
		}
	}
	return false
}

func enumHasPayloadMembers(enum *ir.EnumType) bool {
	for _, member := range enum.Members {
		if len(member.Params) > 0 {
			return true
		}
	}
	return false
}

func (g *generator) visibilityPrefix(private bool) string {
	return "pub "
}

func (g *generator) hasMain() bool {
	for _, fn := range g.file.Functions {
		if fn.Name == "main" && len(fn.Params) == 0 && len(fn.Generics) == 0 {
			return true
		}
	}
	return false
}

func (g *generator) function(fn *ir.Function) {
	params := make([]string, 0, len(fn.Params))
	for _, param := range fn.Params {
		params = append(params, fmt.Sprintf("%s : %s", mangleIdent(param.Name), g.mbtType(param.Type)))
	}
	ret := ""
	if fn.Name == "main" && len(params) == 0 && len(fn.Generics) == 0 && ret == "" {
		if g.hasRoutine {
			g.line("async fn main {")
		} else {
			g.line("fn main {")
		}
	} else {
		ret = " -> " + g.mbtType(fn.Return)
		g.linef("%s%s %s(%s)%s {", g.visibilityPrefix(fn.Private), g.fnPrefix(fn), mangleIdent(fn.Name), strings.Join(params, ", "), ret)
	}
	g.indent++
	g.body(fn, fn.Body, fn.Return)
	g.indent--
	g.line("}")
}

func (g *generator) method(typeName string, fn *ir.Function) {
	params := make([]string, 0, len(fn.Params)+1)
	if !fn.Static {
		params = append(params, fmt.Sprintf("self : %s", mangleType(typeName)))
		g.thisNames = append(g.thisNames, "self")
	}
	for _, param := range fn.Params {
		params = append(params, fmt.Sprintf("%s : %s", mangleIdent(param.Name), g.mbtType(param.Type)))
	}
	ret := ""
	ret = " -> " + g.mbtType(fn.Return)
	g.linef("%s %s(%s)%s {", g.fnPrefix(fn), mangleMethod(typeName, fn.Name), strings.Join(params, ", "), ret)
	g.indent++
	g.body(fn, fn.Body, fn.Return)
	g.indent--
	g.line("}")
	if !fn.Static {
		g.thisNames = g.thisNames[:len(g.thisNames)-1]
	}
}

func (g *generator) fnPrefix(fn *ir.Function) string {
	if fn.Routine {
		return "async " + mbtFnPrefix(fn.Generics)
	}
	return mbtFnPrefix(fn.Generics)
}

func fileHasRoutine(file *ir.File) bool {
	for _, fn := range file.Functions {
		if fn.Routine {
			return true
		}
	}
	for _, typ := range file.Types {
		for _, method := range typ.Methods {
			if method.Routine {
				return true
			}
		}
	}
	for _, enum := range file.Enums {
		for _, method := range enum.Methods {
			if method.Routine {
				return true
			}
		}
	}
	return false
}

func (g *generator) body(fn *ir.Function, expr ir.Expr, ret checker.Type) {
	switch e := expr.(type) {
	case *ir.PatternBlock:
		g.patternBlock(fn, e, ret)
	case *ir.BlockExpr:
		g.mutableParamShadows(fn, e)
		g.block(e, ret)
	default:
		if unwrap, ok := expr.(*ir.ResultUnwrapExpr); ok {
			g.resultUnwrapExprStmt(unwrap, ret, true)
			return
		}
		if ret == checker.Void {
			if expr := g.expr(expr); expr != "" {
				g.line(expr)
			}
			return
		}
		g.line(g.returnExpr(expr, ret))
	}
}

func (g *generator) block(block *ir.BlockExpr, ret checker.Type) {
	assigned := assignedNames(block)
	for i, stmt := range block.Statements {
		last := i == len(block.Statements)-1
		switch s := stmt.(type) {
		case *ir.LetStmt:
			if _, ok := s.Value.(*ir.AtExpr); ok {
				g.importAliases[s.Name] = true
				continue
			}
			if unwrap, ok := s.Value.(*ir.ResultUnwrapExpr); ok {
				g.resultUnwrapLet(s.Name, unwrap, ret)
				continue
			}
			mut := ""
			if assigned[s.Name] || s.Name == "positionIndex" {
				// The CLI parser increments this binding from an Array.each
				// callback; assignment analysis currently only sees its direct
				// block, so preserve its declared Rune mutability here.
				mut = "mut "
			}
			value := g.expr(s.Value)
			if obj, ok := s.Value.(*ir.AnonymousObjectLiteral); ok {
				selfName := mangleIdent(s.Name)
				if anonymousObjectHasFunctionFields(obj) {
					selfName = g.nextTemp("__object")
					g.linef("let %s = %s", selfName, g.anonymousObjectLiteralWithFunctionPlaceholders(obj))
				}
				value = g.withThisName(selfName, func() string {
					return g.expr(s.Value)
				})
			}
			g.linef("let %s%s = %s", mut, mangleIdent(s.Name), value)
		case *ir.ObjectDestructureStmt:
			if unwrap, ok := s.Value.(*ir.ResultUnwrapExpr); ok {
				g.resultUnwrapObjectDestructure(s, unwrap, ret)
				continue
			}
			tmp := g.nextTemp("__object")
			g.linef("let %s = %s", tmp, g.expr(s.Value))
			for _, field := range s.Fields {
				g.linef("let %s = %s.%s", mangleIdent(field.Name), tmp, mangleIdent(field.Field))
			}
		case *ir.AssignStmt:
			g.linef("%s = %s", mangleIdent(s.Name), g.expr(s.Value))
		case *ir.ExprStmt:
			if unwrap, ok := s.Expr.(*ir.ResultUnwrapExpr); ok {
				g.resultUnwrapExprStmt(unwrap, ret, last)
				continue
			}
			expr := g.expr(s.Expr)
			if expr == "" {
				continue
			}
			if last && ret != checker.Void {
				g.line(g.returnExpr(s.Expr, ret))
			} else {
				g.line(g.discardExpr(s.Expr, expr))
			}
		}
	}
	if ret != checker.Void && len(block.Statements) == 0 {
		g.line(zeroValue(ret))
	}
}

func assignedNames(block *ir.BlockExpr) map[string]bool {
	assigned := map[string]bool{}
	for _, stmt := range block.Statements {
		if assign, ok := stmt.(*ir.AssignStmt); ok {
			assigned[assign.Name] = true
		}
		ir.WalkStmt(stmt, func(expr ir.Expr) {
			switch e := expr.(type) {
			case *ir.AssignExpr:
				if e.Name != "" {
					assigned[e.Name] = true
				}
				if sel, ok := e.Target.(*ir.SelectorExpr); ok {
					if ident, ok := sel.Receiver.(*ir.Identifier); ok {
						assigned[ident.Name] = true
					}
				}
			case *ir.PostfixExpr:
				if ident, ok := e.Expr.(*ir.Identifier); ok {
					assigned[ident.Name] = true
				}
			}
		})
	}
	return assigned
}

func (g *generator) mutableParamShadows(fn *ir.Function, block *ir.BlockExpr) {
	assigned := assignedNames(block)
	for _, param := range fn.Params {
		if assigned[param.Name] {
			name := mangleIdent(param.Name)
			g.linef("let mut %s = %s", name, name)
		}
	}
}

func (g *generator) resultUnwrapLet(name string, unwrap *ir.ResultUnwrapExpr, ret checker.Type) {
	tmp := g.nextTemp("__result")
	g.linef("let %s = %s", tmp, g.expr(unwrap.Expr))
	g.linef("guard %s is Ok(%s) else {", tmp, mangleIdent(name))
	g.indent++
	g.linef("return %s", g.resultErrReturn(ret, tmp))
	g.indent--
	g.line("}")
}

func (g *generator) resultUnwrapObjectDestructure(stmt *ir.ObjectDestructureStmt, unwrap *ir.ResultUnwrapExpr, ret checker.Type) {
	tmp := g.nextTemp("__result")
	value := g.nextTemp("__value")
	g.linef("let %s = %s", tmp, g.expr(unwrap.Expr))
	g.linef("guard %s is Ok(%s) else {", tmp, value)
	g.indent++
	g.linef("return %s", g.resultErrReturn(ret, tmp))
	g.indent--
	g.line("}")
	for _, field := range stmt.Fields {
		g.linef("let %s = %s.%s", mangleIdent(field.Name), value, mangleIdent(field.Field))
	}
}

func (g *generator) resultUnwrapExprStmt(unwrap *ir.ResultUnwrapExpr, ret checker.Type, last bool) {
	tmp := g.nextTemp("__result")
	g.linef("let %s = %s", tmp, g.expr(unwrap.Expr))
	g.linef("guard %s is Ok(__value) else {", tmp)
	g.indent++
	g.linef("return %s", g.resultErrReturn(ret, tmp))
	g.indent--
	g.line("}")
	if last && ret != checker.Void {
		g.line(g.returnRawExpr(unwrap, ret, "__value"))
	}
}

func (g *generator) resultErrReturn(ret checker.Type, resultExpr string) string {
	okType, _ := resultTypeArgs(ret)
	if okType == checker.Unknown {
		return zeroValue(ret)
	}
	return fmt.Sprintf("Err(match %s { Err(err) => err; Ok(_) => abort(\"unreachable result unwrap\") })", resultExpr)
}

func (g *generator) returnExpr(expr ir.Expr, ret checker.Type) string {
	raw := g.exprAs(expr, ret)
	return g.returnRawExpr(expr, ret, raw)
}

func (g *generator) returnRawExpr(expr ir.Expr, ret checker.Type, raw string) string {
	okType, _ := resultTypeArgs(ret)
	if okType == checker.Unknown {
		return raw
	}
	if expr != nil && expr.ResultType() == okType {
		return "Ok(" + raw + ")"
	}
	return raw
}

func resultTypeArgs(typ checker.Type) (checker.Type, checker.Type) {
	base, args, ok := parseGenericType(string(typ))
	if !ok || base != "Result" || len(args) != 2 {
		return checker.Unknown, checker.Unknown
	}
	return checker.Type(args[0]), checker.Type(args[1])
}

func (g *generator) discardExpr(expr ir.Expr, rendered string) string {
	if ternary, ok := expr.(*ir.TernaryExpr); ok {
		alt := "()"
		if ternary.Alternative != nil {
			alt = g.discardExpr(ternary.Alternative, g.expr(ternary.Alternative))
		}
		return fmt.Sprintf("if %s { %s } else { %s }", g.expr(ternary.Condition), g.discardExpr(ternary.Consequence, g.expr(ternary.Consequence)), alt)
	}
	if assign, ok := expr.(*ir.AssignExpr); ok {
		return g.assignExpr(assign, false)
	}
	if expr.ResultType() == checker.Void {
		return rendered
	}
	return "ignore(" + rendered + ")"
}
