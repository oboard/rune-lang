package moonbitcodegen

func (g *generator) fsRuntime() {
	g.line("struct RuneFileStat {")
	g.indent++
	g.line("size : Int")
	g.line("isFile : Bool")
	g.line("isDirectory : Bool")
	g.indent--
	g.line("}")
	g.line("")
	g.line("async fn rune_fs_read_file(path : String) -> Result[Array[Int], String] {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("Ok(rune_bytes_to_ints(@fs.read_file(path).binary()))")
	g.indent--
	g.line("} catch {")
	g.indent++
	g.line("err => Err(err.to_string())")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("async fn rune_fs_read_file_text(path : String) -> Result[String, String] {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("Ok(@fs.read_file(path).text())")
	g.indent--
	g.line("} catch {")
	g.indent++
	g.line("err => Err(err.to_string())")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("async fn rune_fs_write_file(path : String, data : Array[Int]) -> Result[Unit, String] {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("@fs.write_file(path, rune_ints_to_bytes(data), create_mode=CreateOrTruncate)")
	g.line("Ok(())")
	g.indent--
	g.line("} catch {")
	g.indent++
	g.line("err => Err(err.to_string())")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("async fn rune_fs_write_file_text(path : String, data : String) -> Result[Unit, String] {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("@fs.write_file(path, data, create_mode=CreateOrTruncate)")
	g.line("Ok(())")
	g.indent--
	g.line("} catch {")
	g.indent++
	g.line("err => Err(err.to_string())")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("async fn rune_fs_exists(path : String) -> Result[Bool, String] {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("Ok(@fs.exists(path))")
	g.indent--
	g.line("} catch {")
	g.indent++
	g.line("err => Err(err.to_string())")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("async fn rune_fs_readdir(path : String) -> Result[Array[String], String] {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("Ok(@fs.readdir(path))")
	g.indent--
	g.line("} catch {")
	g.indent++
	g.line("err => Err(err.to_string())")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("async fn rune_fs_mkdir(path : String) -> Result[Unit, String] {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("@fs.mkdir(path, recursive=true)")
	g.line("Ok(())")
	g.indent--
	g.line("} catch {")
	g.indent++
	g.line("err => Err(err.to_string())")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("async fn rune_fs_remove(path : String) -> Result[Unit, String] {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("@fs.remove(path)")
	g.line("Ok(())")
	g.indent--
	g.line("} catch {")
	g.indent++
	g.line("err => Err(err.to_string())")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("async fn rune_fs_stat(path : String) -> Result[RuneFileStat, String] {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("let kind = @fs.kind(path)")
	g.line("let size = match kind { Regular => @fs.open(path, mode=ReadOnly).size().to_int(); _ => 0 }")
	g.line("Ok(RuneFileStat::{ size, isFile: kind == Regular, isDirectory: kind == Directory })")
	g.indent--
	g.line("} catch {")
	g.indent++
	g.line("err => Err(err.to_string())")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
}
