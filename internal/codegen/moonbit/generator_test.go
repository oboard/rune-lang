package moonbitcodegen_test

import (
	"strings"
	"testing"

	moonbitcodegen "github.com/oboard/rune-lang/internal/codegen/moonbit"
	"github.com/oboard/rune-lang/internal/compiler"
)

func TestGenerateMainPrintln(t *testing.T) {
	src := `main() => {
  @io.println("Hello")
}`
	got := generateSource(t, src)
	if !strings.Contains(got, "fn main {\n") {
		t.Fatalf("generated main =\n%s\nwant MoonBit main without parameter list", got)
	}
	if !strings.Contains(got, `println("Hello")`) {
		t.Fatalf("generated source =\n%s\nwant println call", got)
	}
}

func TestGeneratePatternBlockFib(t *testing.T) {
	src := `fib(n: Int) -> Int => {
  0 => 0
  1 => 1
  _ => fib(n - 1) + fib(n - 2)
}`
	got := generateSource(t, src)
	for _, want := range []string{
		"(n : Int) -> Int {",
		"match n {",
		"0 => 0",
		"_ =>",
		"(n - 1) +",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated source =\n%s\nmissing %q", got, want)
		}
	}
}

func TestGenerateRejectsGoFFI(t *testing.T) {
	src := `main() => {
  @go.stmt("println")
}`
	prog, diags := compiler.AnalyzeSource("test.rn", src)
	if len(diags) > 0 {
		t.Fatalf("AnalyzeSource() diagnostics = %#v", diags)
	}
	_, err := moonbitcodegen.GenerateIR(prog.IR)
	if err == nil || !strings.Contains(err.Error(), "MoonBit backend does not support @go FFI") {
		t.Fatalf("GenerateIR() error = %v, want @go FFI diagnostic", err)
	}
}

func TestGenerateMapGetOr(t *testing.T) {
	src := `main() => {
  values := @map.new("", "")
  @io.println(values.getOr("name", "Rune"))
}`
	got := generateSource(t, src)
	if !strings.Contains(got, `.get_or_default("name", "Rune")`) {
		t.Fatalf("generated source =\n%s\nwant Map::get_or_default", got)
	}
}

func TestGenerateMapIndexFallback(t *testing.T) {
	src := `main() => {
  values := {
    "name": "Rune",
  }
  @io.println(values["missing"] ?? "fallback")
}`
	got := generateSource(t, src)
	if !strings.Contains(got, `.get("missing").unwrap_or("fallback")`) {
		t.Fatalf("generated source =\n%s\nwant Map::get with Option unwrap_or", got)
	}
	if strings.Contains(got, "??") {
		t.Fatalf("generated source still contains Rune fallback operator:\n%s", got)
	}
}

func TestGenerateBytesIntrinsics(t *testing.T) {
	src := `main() => {
  bytes := @bytes.fromInts([1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0])
  @io.println(bytes.length())
  @io.println(bytes.toInts()[1])
  @io.println(@int8.toInt(bytes.setInt8(2, @int8.fromInt(0 - 2))))
  @io.println(@int16.toInt(bytes.setInt16(3, @int16.fromInt(0 - 1234), true)))
  @io.println(bytes.getInt(3, true))
  @io.println(@float.toDouble(bytes.setFloat32(7, @float.fromDouble(1.5), true)))
  @io.println(@float.toDouble(bytes.getFloat(7, true)))
}`
	got := generateSource(t, src)
	for _, want := range []string{
		"let bytes = [1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0].copy()",
		"println((bytes.length()).to_string())",
		"(bytes.copy())[1]",
		"bytes[2] = ({ let __n = (0 - 2) & 255;",
		"Float::from_double(1.5)",
		"Float::reinterpret_from_uint",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated source =\n%s\nmissing %q", got, want)
		}
	}
	if strings.Contains(got, "MoonBit intrinsic bytes.") {
		t.Fatalf("generated source still traps bytes intrinsics:\n%s", got)
	}
}

func TestGenerateBufferReaderWriterIntrinsics(t *testing.T) {
	src := `main() => {
  buffer := @buffer.new()
  buffer.append(@uint8.fromInt(1)).appendInt(2).appendBytes(@bytes.fromInts([3, 4]))
  writer := buffer.writer()
  writer.writeUInt8(@uint8.fromInt(5))
  writer.writeFloat(@float.fromDouble(1.5), true)
  writer.writeDouble(2.25, false)
  reader := buffer.reader()
  @io.println(buffer.length())
  @io.println(writer.toInts()[4])
  @io.println(reader.readUInt8())
  @io.println(@float.toDouble(reader.readFloat32(true)))
  @io.println(reader.readFloat64(false))
}`
	got := generateSource(t, src)
	for _, want := range []string{
		"pub struct Buffer {",
		"values : Array[Int]",
		"pub fn buffer_new() -> Array[Int]",
		"Buffer::{ values: [] }",
		"let buffer = buffer_new()",
		"(writer.toInts())[4]",
		"Float::from_double(1.5)",
		"reader.readFloat64(false)",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated source =\n%s\nmissing %q", got, want)
		}
	}
	for _, trap := range []string{
		"MoonBit intrinsic buffer.",
		"MoonBit intrinsic reader.",
		"MoonBit intrinsic writer.",
	} {
		if strings.Contains(got, trap) {
			t.Fatalf("generated source still traps stream intrinsics:\n%s", got)
		}
	}
}

func TestGenerateRegexIntrinsics(t *testing.T) {
	src := `main() => {
  re := /(\w+)-(\d+)/g
  hit := re.exec("id-42")
  @io.println(hit[0])
  @io.println(re.match("id-42 next-7")[1])
  @io.println(re.matchAll("id-42 next-7")[1][2])
  @io.println(re.test("id-42"))
  @io.println(re.replaceAll("id-42 next-7", "[$1]"))
  @io.println(re.source())
  @io.println(re.flags())
  @io.println(re.global())
  @io.println(re.lastIndex())
  @io.println(re.setLastIndex(0))
  @io.println(@regex.new("\\d+", "i").search("abc123"))
  @io.println(@regex.escape("a+b?"))
}`
	got := generateSource(t, src)
	for _, want := range []string{
		"struct RuneRegex {",
		"@string.Regex::unsafe_from_string(pattern)",
		"rune_regex_new(\"(\\\\w+)-(\\\\d+)\", \"g\")",
		"rune_regex_exec(re, \"id-42\")",
		"rune_regex_match(re, \"id-42 next-7\")",
		"rune_regex_match_all(re, \"id-42 next-7\")",
		"rune_regex_test(re, \"id-42\")",
		"rune_regex_replace(re, \"id-42 next-7\", \"[$1]\", true)",
		"rune_regex_has_flag(re, \"g\")",
		"{ re.last_index = 0; 0 }",
		"rune_regex_search(rune_regex_new(\"\\\\d+\", \"i\"), \"abc123\")",
		"rune_regex_escape(\"a+b?\")",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated source =\n%s\nmissing %q", got, want)
		}
	}
	for _, trap := range []string{
		"MoonBit backend does not support regex literals",
		"MoonBit intrinsic regex.",
	} {
		if strings.Contains(got, trap) {
			t.Fatalf("generated source still traps regex support:\n%s", got)
		}
	}
}

func TestGenerateCLIOptionDefaultValue(t *testing.T) {
	src := `main() => {
  cmd := @cli.command("ship", "Ship artifacts")
  @cli.withOption(cmd, @cli.option("output", "o", "FILE", "write output", true, null))
  @cli.withOption(cmd, @cli.option("mode", "m", "MODE", "mode", false, "check"))
}`
	got := generateSource(t, src)
	for _, want := range []string{
		`pub struct CliCommand`,
		`pub struct CliOption`,
		`pub fn cli_option(name : String, short : String, valueName : String, help : String, required : Bool, defaultValue : String?) -> CliOption`,
		`ignore(cli_withOption(cmd, cli_option("output", "o", "FILE", "write output", true, None)))`,
		`ignore(cli_withOption(cmd, cli_option("mode", "m", "MODE", "mode", false, Some("check"))))`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated source =\n%s\nmissing %q", got, want)
		}
	}
	if strings.Contains(got, "Some(None)") {
		t.Fatalf("generated source contains double optional default:\n%s", got)
	}
}

func TestGenerateResultUnwrap(t *testing.T) {
	src := `read(flag: Bool) -> Result[String, Error] => flag {
  true => Ok("Rune")
  _ => Err(Error { code: 1, message: "bad", cause: null })
}

~ load() => {
  value := read(true)?
  value
}`
	got := generateSource(t, src)
	for _, want := range []string{
		"load() -> Result[String, String] {",
		"let __result",
		"guard __result",
		"is Ok(value) else {",
		`return Err(match __result`,
		`Ok(_) => abort("unreachable result unwrap")`,
		"Ok(value)",
		`true => Ok("Rune")`,
		`_ => Err("bad")`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated source =\n%s\nmissing %q", got, want)
		}
	}
	if strings.Contains(got, "does not support result unwrap") {
		t.Fatalf("generated source still rejects result unwrap:\n%s", got)
	}
}

func TestGenerateNumericAndProcessIntrinsics(t *testing.T) {
	src := `main() => {
  @assert.eq(7.toDouble(), 7.0)
  @assert.eq(42.toString(), "42")
  @assert.eq(5.toBigInt(), 5n)
  @assert.eq(6n.toString(), "6")
  @assert.eq(3.2.ceil(), 4)
  @assert.eq(@process.cwd().isEmpty(), false)
  @assert.eq(@process.env("__RUNE_TEST_MISSING_ENV__"), null)
}`
	got := generateSource(t, src)
	for _, want := range []string{
		"(7).to_double()",
		"(42).to_string()",
		"BigInt::from_int(5)",
		`(BigInt::from_string("6")).to_string()`,
		"(3.2).ceil()",
		`if ((7).to_double()) != (7.0)`,
		"@env.current_dir().unwrap_or(\"\")",
		`@env.get_env_var("__RUNE_TEST_MISSING_ENV__")`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated source =\n%s\nmissing %q", got, want)
		}
	}
	if strings.Contains(got, ".to_double().to_string()") {
		t.Fatalf("generated source mixes numeric conversions:\n%s", got)
	}
}

func TestGenerateCompressIntrinsics(t *testing.T) {
	src := `~ load(text: String) => {
  gz := @compress.gzipText(text)?
  @compress.gunzipText(gz)?
  deflated := @compress.deflate(gz)?
  inflated := @compress.inflate(deflated)?
  brotli := @compress.brotliText(text)?
  @compress.unbrotliText(brotli)?
  zstd := @compress.zstdText(text)?
  @compress.unzstdText(zstd)?
  @compress.gunzipText(inflated)?
}`
	got := generateSource(t, src)
	for _, want := range []string{
		"async fn rune_compress_gzip",
		"@gzip.Encoder",
		"@flate.compress",
		"@flate.decompress",
		"@brotli.compress",
		"@brotli.decompress",
		"@zstd.compress",
		"@zstd.decompress",
		"@utf8.encode(value)",
		"@utf8.decode(out)",
		"rune_compress_inflate",
		"rune_compress_unzstd_text",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated source =\n%s\nmissing %q", got, want)
		}
	}
	if strings.Contains(got, "MoonBit intrinsic compress.") {
		t.Fatalf("generated source still traps compress intrinsics:\n%s", got)
	}
}

func TestGenerateJSONIntrinsics(t *testing.T) {
	src := `#json.object
Account: {
  #json.name("display_name")
  name: String
  #json.ignore
  password: String
}

main() => {
  account := Account { name: "Ada", password: "secret" }
  @io.println(@json.stringify(account))
  value := { name: "Rune", nested: account, greet() => @io.println(.name) }
  @io.println(@json.stringify(value))
  parsed := @json.parse("{\"display_name\":\"Ada\"}") : Account
  @io.println(parsed.password)
}`
	got := generateSource(t, src)
	for _, want := range []string{
		`struct AnonValue`,
		`name : String`,
		`nested : Account`,
		`greet : () -> Unit`,
		`Json::object({ "display_name": Json::string(__json_value`,
		`.stringify()`,
		`greet: () => ()`,
		`let value = AnonValue`,
		`greet: () => println(__object`,
		`Json::object({ "name": Json::string(__json_value`,
		`"nested":`,
		`"display_name": Json::string(__json_value`,
		`Account::{ name: match __json_obj`,
		`password: Account::{ name: "", password: "" }.password`,
		`@json.parse("{\"display_name\":\"Ada\"}")`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated source =\n%s\nmissing %q", got, want)
		}
	}
	for _, unwanted := range []string{
		`"password": Json::string`,
		`get("password")`,
		`"greet":`,
	} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("generated source should ignore JSON field %q:\n%s", unwanted, got)
		}
	}
}

func TestGenerateParenthesizesTernaryInBinaryExpr(t *testing.T) {
	src := `main() => @io.println("hello " + (true ? "Rune" : "MoonBit"))`
	got := generateSource(t, src)
	if !strings.Contains(got, `"hello " + (if true { "Rune" } else { "MoonBit" })`) {
		t.Fatalf("generated source =\n%s\nwant parenthesized if expression in binary expr", got)
	}
}

func TestGenerateTernaryLambdaCallee(t *testing.T) {
	src := `fun(flag: Bool) => {
  (flag ? (x) => {
    k: x.a + 1
  } : (y) => {
    k: y.b + 1
  })({
    b: 2,
    z: false,
    a: 1
  }).k
}`
	got := generateSource(t, src)
	for _, want := range []string{
		"((if flag {",
		"(x : AnonObject",
		"(y : AnonObject",
		"})(AnonObject",
		"})).k",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated source =\n%s\nmissing %q", got, want)
		}
	}
}

func TestGenerateStringTrimReturnsString(t *testing.T) {
	src := `trimmed() -> String => " Rune ".trim()`
	got := generateSource(t, src)
	if !strings.Contains(got, `" Rune ".trim().to_owned()`) {
		t.Fatalf("generated source =\n%s\nwant trim converted back to String", got)
	}
}

func TestGenerateStringSearchUsesCurrentMoonBitNames(t *testing.T) {
	src := `startsAndEnds(value: String) -> Bool => value.startsWith("Ru") && value.endsWith("ne")`
	got := generateSource(t, src)
	for _, want := range []string{
		`value.has_prefix("Ru")`,
		`value.has_suffix("ne")`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated source =\n%s\nmissing %q", got, want)
		}
	}
	for _, deprecated := range []string{"starts_with", "ends_with"} {
		if strings.Contains(got, deprecated) {
			t.Fatalf("generated source contains deprecated %q:\n%s", deprecated, got)
		}
	}
}

func TestGenerateNotParenthesizesAndExpr(t *testing.T) {
	src := `fun(a: Bool, b: Bool) -> Bool => {
  left := !(a && b)
  right := (a && b).not()
  left && right
}`
	got := generateSource(t, src)
	if gotCount := strings.Count(got, "!(a && b)"); gotCount != 2 {
		t.Fatalf("generated source =\n%s\nwant two !(a && b), got %d", got, gotCount)
	}
}

func TestGenerateEscapesReservedIdentifiers(t *testing.T) {
	src := `Token: {
  module: String
  static: Int
}

useReserved(method: String) -> String => {
  member := Token { module: method, static: 1 }
  member.module
}`
	got := generateSource(t, src)
	for _, want := range []string{
		"module_ : String",
		"static_ : Int",
		"method_ : String",
		"let member_ = Token::{ module_: method_, static_: 1 }",
		"member_.module_",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated source =\n%s\nmissing escaped identifier %q", got, want)
		}
	}
}

func TestGenerateEnumShowUsesOrdinalValues(t *testing.T) {
	src := `Kind: {
  A
  B
}

main() => @io.println(Kind.B)`
	got := generateSource(t, src)
	for _, want := range []string{
		"} derive(Eq)",
		"pub impl Show for Kind with fn output(self, logger) {",
		"B => 1",
		"}).output(logger)",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated source =\n%s\nmissing %q", got, want)
		}
	}
}

func TestGeneratePublicDeclarationsWithoutKeepalive(t *testing.T) {
	src := `Token: {
  kind: Kind
  offset: Int
}

Kind: {
  A
  B
}

unused(kind: Kind) -> String => kind {
  Kind.A => "a"
  Kind.B => "b"
  _ => "unknown"
}

main() => {
  token := Token { kind: Kind.A, offset: 1 }
  @io.println(token.kind)
}`
	got := generateSource(t, src)
	for _, want := range []string{
		"pub struct Token {",
		"pub enum Kind {",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated source =\n%s\nmissing %q", got, want)
		}
	}
	for _, bad := range []string{"__rune_keepalive", "__rune_keep_", "pub let", "pub(all)", "_ => \"unknown\""} {
		if strings.Contains(got, bad) {
			t.Fatalf("generated source contains %q:\n%s", bad, got)
		}
	}
}

func TestGenerateIterAndStringBuffer(t *testing.T) {
	src := `main() => {
  values := @iter.range(1, 3)
  first := values.next()
  @io.println(first[0])
  @iter.range(4, 6).map((value) => value * 2).each((value) => @io.println(value))
  buffer := @stringbuffer.from("Rune")
  buffer.append(" core")
  @io.println(buffer.toString())
}`
	got := generateSource(t, src)
	for _, want := range []string{
		"iter_new(() =>",
		"values.next()",
		"first.0",
		".map((value : Int) => value * 2).to_array()",
		"while __index",
		"StringBuilder()",
		`write_string(" core")`,
		"ignore({ buffer.write_string",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated source =\n%s\nmissing %q", got, want)
		}
	}
	if strings.Contains(got, "().range") || strings.Contains(got, "??") {
		t.Fatalf("generated source contains stale Rune/MoonBit lowering:\n%s", got)
	}
}

func generateSource(t *testing.T, src string) string {
	t.Helper()
	prog, diags := compiler.AnalyzeSource("test.rn", src)
	if len(diags) > 0 {
		t.Fatalf("AnalyzeSource() diagnostics = %#v", diags)
	}
	got, err := moonbitcodegen.GenerateIR(prog.IR)
	if err != nil {
		t.Fatalf("GenerateIR() error = %v\nsource:\n%s", err, got)
	}
	return got
}
