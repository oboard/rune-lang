package moonbitcodegen

func (g *generator) stringRuntime() {
	g.line("fn rune_string_compare(left : String, right : String) -> Int {")
	g.indent++
	g.line("let left_chars : Array[Char] = left.iter().collect()")
	g.line("let right_chars : Array[Char] = right.iter().collect()")
	g.line("let limit = if left_chars.length() < right_chars.length() { left_chars.length() } else { right_chars.length() }")
	g.line("let mut index = 0")
	g.line("while index < limit {")
	g.indent++
	g.line("let left_code = left_chars[index].to_int()")
	g.line("let right_code = right_chars[index].to_int()")
	g.line("if left_code < right_code { return -1 }")
	g.line("if left_code > right_code { return 1 }")
	g.line("index = index + 1")
	g.indent--
	g.line("}")
	g.line("if left_chars.length() < right_chars.length() { return -1 }")
	g.line("if left_chars.length() > right_chars.length() { return 1 }")
	g.line("0")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_string_length(value : String) -> Int {")
	g.indent++
	g.line("value.iter().count()")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_string_at(value : String, index : Int) -> Char {")
	g.indent++
	g.line("let mut i = 0")
	g.line("let mut out = '\\u{0}'")
	g.line("value.iter().each(fn(ch) {")
	g.indent++
	g.line("if i == index { out = ch }")
	g.line("i = i + 1")
	g.indent--
	g.line("})")
	g.line("out")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_string_slice(value : String, start : Int, end : Int) -> String {")
	g.indent++
	g.line("let builder = StringBuilder()")
	g.line("let mut i = 0")
	g.line("value.iter().each(fn(ch) {")
	g.indent++
	g.line("if i >= start && i < end { builder.write_char(ch) }")
	g.line("i = i + 1")
	g.indent--
	g.line("})")
	g.line("builder.to_string()")
	g.indent--
	g.line("}")
}
