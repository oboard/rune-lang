package moonbitcodegen

func (g *generator) readerRuntime() {
	g.line("struct RuneReader {")
	g.indent++
	g.line("data : Array[Int32]")
	g.line("mut position : Int32")
	g.line("mut nibble : Int32")
	g.indent--
	g.line("}")
}
