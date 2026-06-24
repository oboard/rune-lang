package moonbitcodegen

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
)

type generator struct {
	buf       bytes.Buffer
	file      *ir.File
	indent    int
	temp      int
	errors    []error
	thisNames []string
	useCLI    bool
}

func Generate(file *ast.File, info *checker.Info) (string, error) {
	return GenerateIR(ir.LowerFile(file, info))
}

func GenerateIR(file *ir.File) (string, error) {
	if len(file.TSImports) > 0 {
		return "", fmt.Errorf("MoonBit backend does not support TypeScript imports")
	}
	if len(file.GoImports) > 0 {
		return "", fmt.Errorf("MoonBit backend does not support @go.import")
	}
	g := &generator{file: file}
	for i, enum := range file.Enums {
		if i > 0 {
			g.line("")
		}
		g.enumType(enum)
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
			g.method(typ, method)
		}
	}
	if (len(file.Types) > 0 || len(file.Enums) > 0) && len(file.Functions) > 0 {
		g.line("")
	}
	if g.keepaliveEnabled() {
		g.keepaliveFunction()
		if len(file.Functions) > 0 {
			g.line("")
		}
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
	if g.useCLI {
		var runtime generator
		runtime.cliRuntime()
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

func (g *generator) structType(typ *ir.StructType) {
	g.linef("struct %s {", mangleType(typ.Name))
	g.indent++
	for _, field := range typ.Fields {
		g.linef("%s : %s", mangleIdent(field.Name), mbtType(field.Type))
	}
	g.indent--
	g.line("}")
}

func (g *generator) enumType(enum *ir.EnumType) {
	if !enumHasValueMembers(enum) {
		g.linef("enum %s {", mangleType(enum.Name))
		g.indent++
		for _, member := range enum.Members {
			g.line(mangleType(member.Name))
		}
		g.indent--
		g.line("} derive(Eq)")
		g.line("")
		g.enumShowImpl(enum)
		return
	}
	g.linef("type %s Int", mangleType(enum.Name))
	for _, member := range enum.Members {
		if !member.HasValue {
			continue
		}
		g.linef("let %s : %s = %d", mangleIdent(enum.Name+"_"+member.Name), mangleType(enum.Name), member.Value)
	}
}

func (g *generator) enumShowImpl(enum *ir.EnumType) {
	g.linef("impl Show for %s with fn output(self, logger) {", mangleType(enum.Name))
	g.indent++
	g.line("(match self {")
	g.indent++
	for i, member := range enum.Members {
		value := i
		if member.HasValue {
			value = member.Value
		}
		g.linef("%s => %d", mangleType(member.Name), value)
	}
	g.indent--
	g.line("}).output(logger)")
	g.indent--
	g.line("}")
}

func enumHasValueMembers(enum *ir.EnumType) bool {
	for _, member := range enum.Members {
		if member.HasValue {
			return true
		}
	}
	return false
}

func (g *generator) hasKeepalive() bool {
	return len(g.keepaliveLines()) > 0
}

func (g *generator) keepaliveEnabled() bool {
	if !g.hasMain() {
		return false
	}
	return g.hasKeepalive()
}

func (g *generator) hasMain() bool {
	for _, fn := range g.file.Functions {
		if fn.Name == "main" && len(fn.Params) == 0 && len(fn.Generics) == 0 {
			return true
		}
	}
	return false
}

func (g *generator) keepaliveFunction() {
	lines := g.keepaliveLines()
	if len(lines) == 0 {
		return
	}
	g.line("fn __rune_keepalive() -> Unit {")
	g.indent++
	for _, line := range lines {
		g.line(line)
	}
	g.indent--
	g.line("}")
}

func (g *generator) keepaliveLines() []string {
	var lines []string
	for _, enum := range g.file.Enums {
		if enumHasValueMembers(enum) {
			continue
		}
		typeName := mangleType(enum.Name)
		for _, member := range enum.Members {
			if len(member.Params) > 0 {
				continue
			}
			memberName := mangleType(member.Name)
			lines = append(lines, fmt.Sprintf("match %s::%s { %s => (); _ => () }", typeName, memberName, memberName))
		}
	}
	for _, typ := range g.file.Types {
		literal, ok := g.keepaliveStructLiteral(typ)
		if !ok {
			continue
		}
		tmp := "__rune_keep_" + mangleIdent(typ.Name)
		lines = append(lines, fmt.Sprintf("let %s = %s", tmp, literal))
		for _, field := range typ.Fields {
			lines = append(lines, fmt.Sprintf("ignore(%s.%s)", tmp, mangleIdent(field.Name)))
		}
		for _, method := range typ.Methods {
			if len(method.Generics) == 0 {
				lines = append(lines, fmt.Sprintf("ignore(%s)", mangleMethod(typ.Name, method.Name)))
			}
		}
	}
	for _, fn := range g.file.Functions {
		if fn.Name == "main" || len(fn.Generics) > 0 {
			continue
		}
		lines = append(lines, fmt.Sprintf("ignore(%s)", mangleIdent(fn.Name)))
	}
	return lines
}

func (g *generator) keepaliveStructLiteral(typ *ir.StructType) (string, bool) {
	if len(typ.Generics) > 0 {
		return "", false
	}
	fields := make([]string, 0, len(typ.Fields))
	for _, field := range typ.Fields {
		value, ok := g.keepaliveZeroValue(field.Type)
		if !ok {
			return "", false
		}
		fields = append(fields, fmt.Sprintf("%s: %s", mangleIdent(field.Name), value))
	}
	return fmt.Sprintf("%s::{ %s }", mangleType(typ.Name), strings.Join(fields, ", ")), true
}

func (g *generator) keepaliveZeroValue(typ checker.Type) (string, bool) {
	if enum := g.enumFromType(typ); enum != nil && !enumHasValueMembers(enum) {
		for _, member := range enum.Members {
			if len(member.Params) == 0 {
				return mangleType(enum.Name) + "::" + mangleType(member.Name), true
			}
		}
		return "", false
	}
	value := zeroValue(typ)
	if value == "()" && typ != checker.Void && typ != checker.Null && typ != checker.Never && typ != checker.Unknown {
		return "", false
	}
	return value, true
}

func (g *generator) function(fn *ir.Function) {
	params := make([]string, 0, len(fn.Params))
	for _, param := range fn.Params {
		params = append(params, fmt.Sprintf("%s : %s", mangleIdent(param.Name), mbtType(param.Type)))
	}
	ret := ""
	if fn.Name == "main" && len(params) == 0 && len(fn.Generics) == 0 && ret == "" {
		g.line("fn main {")
	} else {
		ret = " -> " + mbtType(fn.Return)
		g.linef("%s %s(%s)%s {", mbtFnPrefix(fn.Generics), mangleIdent(fn.Name), strings.Join(params, ", "), ret)
	}
	g.indent++
	if fn.Name == "main" && len(params) == 0 && len(fn.Generics) == 0 && g.keepaliveEnabled() {
		g.line("__rune_keepalive()")
	}
	g.body(fn, fn.Body, fn.Return)
	g.indent--
	g.line("}")
}

func (g *generator) method(typ *ir.StructType, fn *ir.Function) {
	params := make([]string, 0, len(fn.Params)+1)
	if !fn.Static {
		params = append(params, fmt.Sprintf("self : %s", mangleType(typ.Name)))
		g.thisNames = append(g.thisNames, "self")
	}
	for _, param := range fn.Params {
		params = append(params, fmt.Sprintf("%s : %s", mangleIdent(param.Name), mbtType(param.Type)))
	}
	ret := ""
	ret = " -> " + mbtType(fn.Return)
	g.linef("%s %s(%s)%s {", mbtFnPrefix(fn.Generics), mangleMethod(typ.Name, fn.Name), strings.Join(params, ", "), ret)
	g.indent++
	g.body(fn, fn.Body, fn.Return)
	g.indent--
	g.line("}")
	if !fn.Static {
		g.thisNames = g.thisNames[:len(g.thisNames)-1]
	}
}

func (g *generator) body(fn *ir.Function, expr ir.Expr, ret checker.Type) {
	switch e := expr.(type) {
	case *ir.PatternBlock:
		g.patternBlock(fn, e, ret)
	case *ir.BlockExpr:
		g.block(e, ret)
	default:
		if ret == checker.Void {
			if expr := g.expr(expr); expr != "" {
				g.line(expr)
			}
			return
		}
		g.line(g.expr(expr))
	}
}

func (g *generator) block(block *ir.BlockExpr, ret checker.Type) {
	for i, stmt := range block.Statements {
		last := i == len(block.Statements)-1
		switch s := stmt.(type) {
		case *ir.LetStmt:
			mut := ""
			if s.Mutable {
				mut = "mut "
			}
			g.linef("let %s%s = %s", mut, mangleIdent(s.Name), g.expr(s.Value))
		case *ir.ObjectDestructureStmt:
			tmp := g.nextTemp("__object")
			g.linef("let %s = %s", tmp, g.expr(s.Value))
			for _, field := range s.Fields {
				g.linef("let %s = %s.%s", mangleIdent(field.Name), tmp, mangleIdent(field.Field))
			}
		case *ir.AssignStmt:
			g.linef("%s = %s", mangleIdent(s.Name), g.expr(s.Value))
		case *ir.ExprStmt:
			expr := g.expr(s.Expr)
			if expr == "" {
				continue
			}
			if last && ret != checker.Void {
				g.line(expr)
			} else {
				g.line(g.discardExpr(s.Expr, expr))
			}
		}
	}
	if ret != checker.Void && len(block.Statements) == 0 {
		g.line(zeroValue(ret))
	}
}

func (g *generator) discardExpr(expr ir.Expr, rendered string) string {
	if expr.ResultType() == checker.Void {
		return rendered
	}
	return "ignore(" + rendered + ")"
}
