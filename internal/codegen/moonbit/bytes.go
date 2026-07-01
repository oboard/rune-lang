package moonbitcodegen

func (g *generator) bytesRuntime() {
	g.line("fn rune_bytes_to_ints(data : Bytes) -> Array[Int] {")
	g.indent++
	g.line("let out : Array[Int] = []")
	g.line("let mut index = 0")
	g.line("while index < data.length() {")
	g.indent++
	g.line("out.push(data[index].to_int())")
	g.line("index = index + 1")
	g.indent--
	g.line("}")
	g.line("out")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_ints_to_bytes(data : Array[Int]) -> Bytes {")
	g.indent++
	g.line("let bytes : Array[Byte] = []")
	g.line("let mut index = 0")
	g.line("while index < data.length() {")
	g.indent++
	g.line("bytes.push((data[index] & 255).to_byte())")
	g.line("index = index + 1")
	g.indent--
	g.line("}")
	g.line("Bytes::from_array(bytes)")
	g.indent--
	g.line("}")
}
