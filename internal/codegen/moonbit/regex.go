package moonbitcodegen

func (g *generator) regexRuntime() {
	g.line("struct RuneRegex {")
	g.indent++
	g.line("source : String")
	g.line("flags : String")
	g.line("regex : @string.Regex")
	g.line("mut last_index : Int")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_regex_new(source : String, flags : String) -> RuneRegex {")
	g.indent++
	g.line("let pattern = rune_regex_pattern(source, flags)")
	g.line("{ source, flags, regex: @string.Regex::unsafe_from_string(pattern), last_index: 0 }")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_regex_pattern(source : String, flags : String) -> String {")
	g.indent++
	g.line("let base = source.replace_all(old=\"\\\\d\", new=\"[[:digit:]]\").replace_all(old=\"\\\\D\", new=\"[^[:digit:]]\").replace_all(old=\"\\\\w\", new=\"[[:word:]]\").replace_all(old=\"\\\\W\", new=\"[^[:word:]]\").replace_all(old=\"\\\\s\", new=\"[[:space:]]\").replace_all(old=\"\\\\S\", new=\"[^[:space:]]\")")
	g.line("if flags.contains(\"i\") { \"(?i:\" + base + \")\" } else { base }")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_regex_exec(regex : RuneRegex, input : String) -> Array[String] {")
	g.indent++
	g.line("let start = if regex.flags.contains(\"g\") { regex.last_index } else { 0 }")
	g.line("match regex.regex.execute(input, last_index=start) {")
	g.indent++
	g.line("None => { regex.last_index = 0; [] }")
	g.line("Some(m) => {")
	g.indent++
	g.line("if regex.flags.contains(\"g\") { regex.last_index = m.before().length() + m.content().length() }")
	g.line("rune_regex_match_groups(m)")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_regex_match_groups(m : @string.MatchResult) -> Array[String] {")
	g.indent++
	g.line("let groups : Array[String] = []")
	g.line("let mut index = 0")
	g.line("while true {")
	g.indent++
	g.line("match m.group(index) {")
	g.indent++
	g.line("None => break")
	g.line("Some(value) => { groups.push(value.to_owned()); index = index + 1 }")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("groups")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_regex_match(regex : RuneRegex, input : String) -> Array[String] {")
	g.indent++
	g.line("if regex.flags.contains(\"g\") { regex.regex.find(input).map(fn(m) { m.content().to_owned() }).to_array() } else { rune_regex_exec(regex, input) }")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_regex_match_all(regex : RuneRegex, input : String) -> Array[Array[String]] {")
	g.indent++
	g.line("regex.regex.find(input).map(fn(m) { rune_regex_match_groups(m) }).to_array()")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_regex_test(regex : RuneRegex, input : String) -> Bool {")
	g.indent++
	g.line("rune_regex_exec(regex, input).length() > 0")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_regex_replace(regex : RuneRegex, input : String, template : String, all : Bool) -> String {")
	g.indent++
	g.line("if all || regex.flags.contains(\"g\") {")
	g.indent++
	g.line("regex.regex.replace_by(input, fn(m) { rune_regex_replacement(template, m) }).to_owned()")
	g.indent--
	g.line("} else {")
	g.indent++
	g.line("regex.regex.replace_by(input, fn(m) { rune_regex_replacement(template, m) }, limit=1).to_owned()")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_regex_replacement(template : String, m : @string.MatchResult) -> String {")
	g.indent++
	g.line("let builder = StringBuilder()")
	g.line("let mut index = 0")
	g.line("while index < template.length() {")
	g.indent++
	g.line("if (template[index].to_int() == 36) && ((index + 1) < template.length()) {")
	g.indent++
	g.line("let code = template[index + 1].to_int()")
	g.line("if code >= 48 && code <= 57 {")
	g.indent++
	g.line("match m.group(code - 48) {")
	g.indent++
	g.line("Some(value) => builder.write_stringview(value)")
	g.line("None => ()")
	g.indent--
	g.line("}")
	g.line("index = index + 2")
	g.indent--
	g.line("} else {")
	g.indent++
	g.line("builder.write_stringview(template[index:index + 1])")
	g.line("index = index + 1")
	g.indent--
	g.line("}")
	g.indent--
	g.line("} else {")
	g.indent++
	g.line("builder.write_stringview(template[index:index + 1])")
	g.line("index = index + 1")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("builder.to_string()")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_regex_search(regex : RuneRegex, input : String) -> Int {")
	g.indent++
	g.line("match regex.regex.execute(input) {")
	g.indent++
	g.line("None => -1")
	g.line("Some(m) => m.before().length()")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_regex_split(regex : RuneRegex, input : String) -> Array[String] {")
	g.indent++
	g.line("regex.regex.split(input).map(fn(part) { part.to_owned() }).to_array()")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_regex_has_flag(regex : RuneRegex, flag : String) -> Bool {")
	g.indent++
	g.line("regex.flags.contains(flag)")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_regex_escape(value : String) -> String {")
	g.indent++
	g.line("let builder = StringBuilder()")
	g.line("let mut index = 0")
	g.line("while index < value.length() {")
	g.indent++
	g.line("let ch = value[index].to_int()")
	g.line("if ((((((((((((ch == 92) || (ch == 46)) || (ch == 43)) || (ch == 42)) || (ch == 63)) || (ch == 94)) || (ch == 36)) || (ch == 40)) || (ch == 41)) || (ch == 91)) || (ch == 93)) || (ch == 123)) || (ch == 125) { builder.write_string(\"\\\\\") }")
	g.line("builder.write_stringview(value[index:index + 1])")
	g.line("index = index + 1")
	g.indent--
	g.line("}")
	g.line("builder.to_string()")
	g.indent--
	g.line("}")
}
