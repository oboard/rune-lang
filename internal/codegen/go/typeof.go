package gocodegen

import (
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/ir"
)

func fileUsesTypeofRuntime(file *ir.File) bool {
	found := false
	visit := func(expr ir.Expr) {
		switch expr.(type) {
		case *ir.TypeOfExpr, *ir.NarrowExpr:
			found = true
		}
	}
	for _, fn := range file.Functions {
		ir.WalkExpr(fn.Body, visit)
	}
	for _, typ := range file.Types {
		for _, method := range typ.Methods {
			ir.WalkExpr(method.Body, visit)
		}
	}
	for _, enum := range file.Enums {
		for _, method := range enum.Methods {
			ir.WalkExpr(method.Body, visit)
		}
	}
	return found
}

// typeofRuntime emits the helpers used by `typeof` and union narrowing. They
// are tiny and unused helpers are legal Go, so they are emitted once per file
// instead of threading a usage flag through lowering.
func (g *generator) typeofRuntime() {
	g.line("func runeTypeof(value any) string {")
	g.indent++
	g.line("switch value.(type) {")
	g.line("case runeUndefinedValue:")
	g.indent++
	g.line(`return "undefined"`)
	g.indent--
	g.line("case nil:")
	g.indent++
	g.line(`return "object"`)
	g.indent--
	g.line("case string:")
	g.indent++
	g.line(`return "string"`)
	g.indent--
	g.line("case bool:")
	g.indent++
	g.line(`return "boolean"`)
	g.indent--
	g.line("case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:")
	g.indent++
	g.line(`return "number"`)
	g.indent--
	g.line("case func():")
	g.indent++
	g.line(`return "function"`)
	g.indent--
	g.line("default:")
	g.indent++
	g.line(`return "object"`)
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeAsString(value any) string {")
	g.indent++
	g.line("text, _ := value.(string)")
	g.line("return text")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeAsBool(value any) bool {")
	g.indent++
	g.line("flag, _ := value.(bool)")
	g.line("return flag")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeAsDouble(value any) float64 {")
	g.indent++
	g.line("switch number := value.(type) {")
	g.line("case float64:")
	g.indent++
	g.line("return number")
	g.indent--
	g.line("case float32:")
	g.indent++
	g.line("return float64(number)")
	g.indent--
	g.line("case int:")
	g.indent++
	g.line("return float64(number)")
	g.indent--
	g.line("case int8:")
	g.indent++
	g.line("return float64(number)")
	g.indent--
	g.line("case int16:")
	g.indent++
	g.line("return float64(number)")
	g.indent--
	g.line("case int32:")
	g.indent++
	g.line("return float64(number)")
	g.indent--
	g.line("case int64:")
	g.indent++
	g.line("return float64(number)")
	g.indent--
	g.line("case uint:")
	g.indent++
	g.line("return float64(number)")
	g.indent--
	g.line("case uint8:")
	g.indent++
	g.line("return float64(number)")
	g.indent--
	g.line("case uint16:")
	g.indent++
	g.line("return float64(number)")
	g.indent--
	g.line("case uint64:")
	g.indent++
	g.line("return float64(number)")
	g.indent--
	g.line("default:")
	g.indent++
	g.line("return 0")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeAsInt64(value any) int64 {")
	g.indent++
	g.line("switch number := value.(type) {")
	g.line("case int:")
	g.indent++
	g.line("return int64(number)")
	g.indent--
	g.line("case int8:")
	g.indent++
	g.line("return int64(number)")
	g.indent--
	g.line("case int16:")
	g.indent++
	g.line("return int64(number)")
	g.indent--
	g.line("case int32:")
	g.indent++
	g.line("return int64(number)")
	g.indent--
	g.line("case int64:")
	g.indent++
	g.line("return number")
	g.indent--
	g.line("case uint:")
	g.indent++
	g.line("return int64(number)")
	g.indent--
	g.line("case uint8:")
	g.indent++
	g.line("return int64(number)")
	g.indent--
	g.line("case uint16:")
	g.indent++
	g.line("return int64(number)")
	g.indent--
	g.line("case uint64:")
	g.indent++
	g.line("return int64(number)")
	g.indent--
	g.line("case float64:")
	g.indent++
	g.line("return int64(number)")
	g.indent--
	g.line("default:")
	g.indent++
	g.line("return 0")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
}

func (g *generator) narrowExpr(e *ir.NarrowExpr) string {
	value := g.exprPrec(e.Expr, 0)
	switch e.AssertType {
	case checker.String:
		return "runeAsString(" + value + ")"
	case checker.Bool:
		return "runeAsBool(" + value + ")"
	case checker.Double:
		return "runeAsDouble(" + value + ")"
	case checker.Float:
		return "float32(runeAsDouble(" + value + "))"
	case checker.Int32, checker.Int4, checker.Int8, checker.Int16:
		return "int(runeAsInt64(" + value + "))"
	case checker.Int64:
		return "runeAsInt64(" + value + ")"
	case checker.UInt, checker.UInt8, checker.UInt16:
		return "uint(runeAsInt64(" + value + "))"
	case checker.UInt64:
		return "uint64(runeAsInt64(" + value + "))"
	case checker.Char:
		return "rune(runeAsInt64(" + value + "))"
	case checker.Null:
		return "any(nil)"
	case checker.Undefined:
		return "any(runeUndefined)"
	default:
		return "(" + value + ".(" + goType(e.AssertType) + "))"
	}
}
