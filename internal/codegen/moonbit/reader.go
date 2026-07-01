package moonbitcodegen

func (g *generator) readerRuntime() {
	g.line("struct RuneReader {")
	g.indent++
	g.line("data : Array[Int]")
	g.line("mut position : Int")
	g.line("mut nibble : Int")
	g.indent--
	g.line("}")
}
