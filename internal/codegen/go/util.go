package gocodegen

import (
	"fmt"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
)

func (g *generator) line(s string) {
	for i := 0; i < g.indent; i++ {
		g.buf.WriteByte('\t')
	}
	g.buf.WriteString(s)
	g.buf.WriteByte('\n')
}

func (g *generator) linef(format string, args ...any) {
	g.line(fmt.Sprintf(format, args...))
}

func goType(typ checker.Type) string {
	if elem, ok := checker.ArrayElement(typ); ok {
		return "[]" + goType(elem)
	}
	switch typ {
	case checker.Int:
		return "int"
	case checker.String:
		return "string"
	case checker.Bool:
		return "bool"
	case checker.Unknown:
		return "any"
	default:
		return mangleIdent(string(typ))
	}
}

func zeroValue(typ checker.Type) string {
	if _, ok := checker.ArrayElement(typ); ok {
		return "nil"
	}
	switch typ {
	case checker.Int:
		return "0"
	case checker.String:
		return `""`
	case checker.Bool:
		return "false"
	default:
		return fmt.Sprintf("%s{}", goType(typ))
	}
}

func hasMain(file *ir.File) bool {
	for _, fn := range file.Functions {
		if fn.Name == "main" {
			return true
		}
	}
	return false
}

func mangleIdent(name string) string {
	return "__" + name
}
