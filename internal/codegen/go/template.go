package gocodegen

func (g *generator) templateRuntime() {
	g.line("func runeTemplateString(value any) string {")
	g.indent++
	g.line("switch v := value.(type) {")
	g.line("case nil:")
	g.indent++
	g.line(`return "null"`)
	g.indent--
	g.line("case rune:")
	g.indent++
	g.line("return string(v)")
	g.indent--
	g.line("default:")
	g.indent++
	g.line("return fmt.Sprint(v)")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
}
