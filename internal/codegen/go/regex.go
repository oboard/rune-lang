package gocodegen

func (g *generator) regexRuntime() {
	g.line("type runeRegex struct {")
	g.indent++
	g.line("source string")
	g.line("flags string")
	g.line("lastIndex int")
	g.line("expr *regexp.Regexp")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func newRuneRegex(source string, flags string) *runeRegex {")
	g.indent++
	g.line("canonical := canonicalRegexFlags(flags)")
	g.line("return &runeRegex{source: source, flags: canonical, expr: regexp.MustCompile(goRegexPattern(source, canonical))}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func goRegexPattern(source string, flags string) string {")
	g.indent++
	g.line("prefix := \"\"")
	g.line("if regexHasFlag(flags, 'i') {")
	g.indent++
	g.line("prefix += \"i\"")
	g.indent--
	g.line("}")
	g.line("if regexHasFlag(flags, 'm') {")
	g.indent++
	g.line("prefix += \"m\"")
	g.indent--
	g.line("}")
	g.line("if regexHasFlag(flags, 's') {")
	g.indent++
	g.line("prefix += \"s\"")
	g.indent--
	g.line("}")
	g.line("if prefix == \"\" {")
	g.indent++
	g.line("return source")
	g.indent--
	g.line("}")
	g.line("return \"(?\" + prefix + \")\" + source")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func canonicalRegexFlags(flags string) string {")
	g.indent++
	g.line("seen := map[rune]bool{}")
	g.line("for _, flag := range flags {")
	g.indent++
	g.line("switch flag {")
	g.line("case 'd', 'g', 'i', 'm', 's', 'u', 'v', 'y':")
	g.line("default:")
	g.indent++
	g.line("panic(\"invalid regex flag\")")
	g.indent--
	g.line("}")
	g.line("if seen[flag] {")
	g.indent++
	g.line("panic(\"duplicate regex flag\")")
	g.indent--
	g.line("}")
	g.line("seen[flag] = true")
	g.indent--
	g.line("}")
	g.line("if seen['u'] && seen['v'] {")
	g.indent++
	g.line("panic(\"regex flags u and v cannot be used together\")")
	g.indent--
	g.line("}")
	g.line("out := \"\"")
	g.line("for _, flag := range \"dgimsuvy\" {")
	g.indent++
	g.line("if seen[flag] {")
	g.indent++
	g.line("out += string(flag)")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("return out")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func regexHasFlag(flags string, flag rune) bool {")
	g.indent++
	g.line("for _, item := range flags {")
	g.indent++
	g.line("if item == flag {")
	g.indent++
	g.line("return true")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("return false")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func (r *runeRegex) exec(input string) []string {")
	g.indent++
	g.line("matches := r.execSubmatch(input)")
	g.line("if matches == nil {")
	g.indent++
	g.line("return []string{}")
	g.indent--
	g.line("}")
	g.line("return matches")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func (r *runeRegex) execSubmatch(input string) []string {")
	g.indent++
	g.line("start := 0")
	g.line("stateful := regexHasFlag(r.flags, 'g') || regexHasFlag(r.flags, 'y')")
	g.line("if stateful {")
	g.indent++
	g.line("start = r.lastIndex")
	g.line("if start < 0 || start > len(input) {")
	g.indent++
	g.line("r.lastIndex = 0")
	g.line("return nil")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("loc := r.expr.FindStringSubmatchIndex(input[start:])")
	g.line("if loc == nil || (regexHasFlag(r.flags, 'y') && loc[0] != 0) {")
	g.indent++
	g.line("if stateful {")
	g.indent++
	g.line("r.lastIndex = 0")
	g.indent--
	g.line("}")
	g.line("return nil")
	g.indent--
	g.line("}")
	g.line("if stateful {")
	g.indent++
	g.line("r.lastIndex = start + loc[1]")
	g.indent--
	g.line("}")
	g.line("return r.expr.FindStringSubmatch(input[start+loc[0]:start+loc[1]])")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func (r *runeRegex) match(input string) []string {")
	g.indent++
	g.line("if regexHasFlag(r.flags, 'g') {")
	g.indent++
	g.line("matches := r.expr.FindAllString(input, -1)")
	g.line("if matches == nil {")
	g.indent++
	g.line("return []string{}")
	g.indent--
	g.line("}")
	g.line("return matches")
	g.indent--
	g.line("}")
	g.line("return r.exec(input)")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func (r *runeRegex) matchAll(input string) [][]string {")
	g.indent++
	g.line("matches := r.expr.FindAllStringSubmatch(input, -1)")
	g.line("if matches == nil {")
	g.indent++
	g.line("return [][]string{}")
	g.indent--
	g.line("}")
	g.line("return matches")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func (r *runeRegex) test(input string) bool {")
	g.indent++
	g.line("return r.execSubmatch(input) != nil")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func (r *runeRegex) replace(input string, replacement string) string {")
	g.indent++
	g.line("if regexHasFlag(r.flags, 'g') {")
	g.indent++
	g.line("return r.replaceAll(input, replacement)")
	g.indent--
	g.line("}")
	g.line("loc := r.expr.FindStringSubmatchIndex(input)")
	g.line("if loc == nil {")
	g.indent++
	g.line("return input")
	g.indent--
	g.line("}")
	g.line("out := []byte(input[:loc[0]])")
	g.line("out = r.expr.ExpandString(out, replacement, input, loc)")
	g.line("out = append(out, input[loc[1]:]...)")
	g.line("return string(out)")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func (r *runeRegex) replaceAll(input string, replacement string) string {")
	g.indent++
	g.line("return r.expr.ReplaceAllString(input, replacement)")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func (r *runeRegex) search(input string) int {")
	g.indent++
	g.line("loc := r.expr.FindStringIndex(input)")
	g.line("if loc == nil {")
	g.indent++
	g.line("return -1")
	g.indent--
	g.line("}")
	g.line("return loc[0]")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func (r *runeRegex) split(input string) []string {")
	g.indent++
	g.line("return r.expr.Split(input, -1)")
	g.indent--
	g.line("}")
}
