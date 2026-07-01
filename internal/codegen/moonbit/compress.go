package moonbitcodegen

func (g *generator) compressRuntime() {
	g.line("async fn rune_compress_gzip(data : Array[Int]) -> Result[Array[Int], String] {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("let reader = @io.MemoryReader() <| writer => {")
	g.indent++
	g.line("@gzip.Encoder(writer)..write(rune_ints_to_bytes(data)).end()")
	g.indent--
	g.line("}")
	g.line("let compressed = reader.read_all().binary()")
	g.line("reader.close()")
	g.line("Ok(rune_bytes_to_ints(compressed))")
	g.indent--
	g.line("} catch {")
	g.indent++
	g.line("err => Err(err.to_string())")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("async fn rune_compress_gunzip(data : Array[Int]) -> Result[Array[Int], String] {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("let reader = @io.MemoryReader() <| writer => {")
	g.indent++
	g.line("writer.write(rune_ints_to_bytes(data))")
	g.indent--
	g.line("}")
	g.line("let decoded = @gzip.Decoder(reader).read_all().binary()")
	g.line("reader.close()")
	g.line("Ok(rune_bytes_to_ints(decoded))")
	g.indent--
	g.line("} catch {")
	g.indent++
	g.line("err => Err(err.to_string())")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.compressBytesFunction("deflate", "@flate.compress")
	g.line("")
	g.compressBytesFunction("inflate", "@flate.decompress")
	g.line("")
	g.compressBytesFunction("brotli", "@brotli.compress")
	g.line("")
	g.compressBytesFunction("unbrotli", "@brotli.decompress")
	g.line("")
	g.compressBytesFunction("zstd", "@zstd.compress")
	g.line("")
	g.compressBytesFunction("unzstd", "@zstd.decompress")
	g.line("")
	g.line("async fn rune_compress_gzip_text(value : String) -> Result[Array[Int], String] {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("let reader = @io.MemoryReader() <| writer => {")
	g.indent++
	g.line("@gzip.Encoder(writer)..write(value).end()")
	g.indent--
	g.line("}")
	g.line("let compressed = reader.read_all().binary()")
	g.line("reader.close()")
	g.line("Ok(rune_bytes_to_ints(compressed))")
	g.indent--
	g.line("} catch {")
	g.indent++
	g.line("err => Err(err.to_string())")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("async fn rune_compress_gunzip_text(data : Array[Int]) -> Result[String, String] {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("let reader = @io.MemoryReader() <| writer => {")
	g.indent++
	g.line("writer.write(rune_ints_to_bytes(data))")
	g.indent--
	g.line("}")
	g.line("let decoded = @gzip.Decoder(reader).read_all().text()")
	g.line("reader.close()")
	g.line("Ok(decoded)")
	g.indent--
	g.line("} catch {")
	g.indent++
	g.line("err => Err(err.to_string())")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.compressTextEncodeFunction("brotli")
	g.line("")
	g.compressTextDecodeFunction("unbrotli")
	g.line("")
	g.compressTextEncodeFunction("zstd")
	g.line("")
	g.compressTextDecodeFunction("unzstd")
}

func (g *generator) compressBytesFunction(name string, callee string) {
	g.linef("async fn rune_compress_%s(data : Array[Int]) -> Result[Array[Int], String] {", name)
	g.indent++
	g.line("try {")
	g.indent++
	g.linef("let out = %s(rune_ints_to_bytes(data))", callee)
	g.line("Ok(rune_bytes_to_ints(out))")
	g.indent--
	g.line("} catch {")
	g.indent++
	g.line("err => Err(err.to_string())")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
}

func (g *generator) compressTextEncodeFunction(name string) {
	g.linef("async fn rune_compress_%s_text(value : String) -> Result[Array[Int], String] {", name)
	g.indent++
	g.line("try {")
	g.indent++
	g.linef("let out = @%s.compress(@utf8.encode(value))", name)
	g.line("Ok(rune_bytes_to_ints(out))")
	g.indent--
	g.line("} catch {")
	g.indent++
	g.line("err => Err(err.to_string())")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
}

func (g *generator) compressTextDecodeFunction(name string) {
	codec := name
	switch name {
	case "unbrotli":
		codec = "brotli"
	case "unzstd":
		codec = "zstd"
	}
	g.linef("async fn rune_compress_%s_text(data : Array[Int]) -> Result[String, String] {", name)
	g.indent++
	g.line("try {")
	g.indent++
	g.linef("let out = @%s.decompress(rune_ints_to_bytes(data))", codec)
	g.line("Ok(@utf8.decode(out))")
	g.indent--
	g.line("} catch {")
	g.indent++
	g.line("err => Err(err.to_string())")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
}
