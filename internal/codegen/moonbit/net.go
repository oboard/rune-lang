package moonbitcodegen

func (g *generator) netRuntime() {
	g.line("struct RuneTCPConnection {")
	g.indent++
	g.line("tcp : @socket.Tcp")
	g.indent--
	g.line("}")
	g.line("")
	g.line("struct RuneTCPListener {")
	g.indent++
	g.line("server : @socket.TcpServer")
	g.line("address : String")
	g.indent--
	g.line("}")
	g.line("")
	g.line("async fn rune_net_connect(address : String) -> Result[RuneTCPConnection, String] {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("let tcp = @socket.Tcp::connect(@socket.Addr::parse(address))")
	g.line("Ok(RuneTCPConnection::{ tcp })")
	g.indent--
	g.line("} catch {")
	g.indent++
	g.line("err => Err(err.to_string())")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("async fn rune_net_listen(address : String) -> Result[RuneTCPListener, String] {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("let server = @socket.TcpServer(@socket.Addr::parse(address))")
	g.line("Ok(RuneTCPListener::{ server, address })")
	g.indent--
	g.line("} catch {")
	g.indent++
	g.line("err => Err(err.to_string())")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("async fn rune_net_connection_read(connection : RuneTCPConnection, length : Int) -> Result[Array[Int], String] {")
	g.indent++
	g.line("if length < 0 { return Err(\"net read length out of range\") }")
	g.line("if length == 0 { return Ok([]) }")
	g.line("try {")
	g.indent++
	g.line("match connection.tcp.read_some(max_len=length) {")
	g.indent++
	g.line("Some(data) => Ok(rune_bytes_to_ints(data))")
	g.line("None => Ok([])")
	g.indent--
	g.line("}")
	g.indent--
	g.line("} catch {")
	g.indent++
	g.line("err => Err(err.to_string())")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("async fn rune_net_connection_write(connection : RuneTCPConnection, data : Array[Int]) -> Result[Int, String] {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("let bytes = rune_ints_to_bytes(data)")
	g.line("let written = connection.tcp.write_once(bytes, offset=0, len=bytes.length())")
	g.line("Ok(written)")
	g.indent--
	g.line("} catch {")
	g.indent++
	g.line("err => Err(err.to_string())")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_net_connection_close(connection : RuneTCPConnection) -> Result[Unit, String] {")
	g.indent++
	g.line("connection.tcp.close()")
	g.line("Ok(())")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_net_listener_address(listener : RuneTCPListener) -> String {")
	g.indent++
	g.line("listener.address")
	g.indent--
	g.line("}")
	g.line("")
	g.line("async fn rune_net_listener_accept(listener : RuneTCPListener) -> Result[RuneTCPConnection, String] {")
	g.indent++
	g.line("try {")
	g.indent++
	g.line("let (tcp, _) = listener.server.accept()")
	g.line("Ok(RuneTCPConnection::{ tcp })")
	g.indent--
	g.line("} catch {")
	g.indent++
	g.line("err => Err(err.to_string())")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_net_listener_close(listener : RuneTCPListener) -> Result[Unit, String] {")
	g.indent++
	g.line("listener.server.close()")
	g.line("Ok(())")
	g.indent--
	g.line("}")
	g.line("")
}
