package gocodegen

func (g *generator) binaryRuntime() {
	g.line("func runeInt4(value int) int8 {")
	g.indent++
	g.line("n := value & 0xf")
	g.line("if n >= 8 {")
	g.indent++
	g.line("return int8(n - 16)")
	g.indent--
	g.line("}")
	g.line("return int8(n)")
	g.indent--
	g.line("}")
	g.line("")
	g.line("type runeBinary struct {")
	g.indent++
	g.line("data []byte")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func newRuneBinary(length int) *runeBinary {")
	g.indent++
	g.line("if length < 0 {")
	g.indent++
	g.line("panic(\"binary length out of range\")")
	g.indent--
	g.line("}")
	g.line("return &runeBinary{data: make([]byte, length)}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeBinaryFromInts(values []int) *runeBinary {")
	g.indent++
	g.line("out := make([]byte, len(values))")
	g.line("for i, value := range values {")
	g.indent++
	g.line("out[i] = byte(value)")
	g.indent--
	g.line("}")
	g.line("return &runeBinary{data: out}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeBinaryOrder(littleEndian bool) binary.ByteOrder {")
	g.indent++
	g.line("if littleEndian {")
	g.indent++
	g.line("return binary.LittleEndian")
	g.indent--
	g.line("}")
	g.line("return binary.BigEndian")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func (b *runeBinary) check(offset int, size int) {")
	g.indent++
	g.line("if offset < 0 || size < 0 || offset+size > len(b.data) {")
	g.indent++
	g.line("panic(\"binary offset out of range\")")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func (b *runeBinary) ByteLength() int { return len(b.data) }")
	g.line("")
	g.line("func (b *runeBinary) Clone() *runeBinary {")
	g.indent++
	g.line("return &runeBinary{data: append([]byte(nil), b.data...)}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func (b *runeBinary) Slice(start int, end int) *runeBinary {")
	g.indent++
	g.line("b.check(start, end-start)")
	g.line("return &runeBinary{data: append([]byte(nil), b.data[start:end]...)}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func (b *runeBinary) ToInts() []int {")
	g.indent++
	g.line("out := make([]int, len(b.data))")
	g.line("for i, value := range b.data {")
	g.indent++
	g.line("out[i] = int(value)")
	g.indent--
	g.line("}")
	g.line("return out")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func (b *runeBinary) GetInt4(index int) int8 {")
	g.indent++
	g.line("if index < 0 { panic(\"binary offset out of range\") }")
	g.line("byteIndex := index / 2")
	g.line("b.check(byteIndex, 1)")
	g.line("value := b.data[byteIndex]")
	g.line("var nibble byte")
	g.line("if index%2 == 0 {")
	g.indent++
	g.line("nibble = value >> 4")
	g.indent--
	g.line("} else {")
	g.indent++
	g.line("nibble = value & 0x0f")
	g.indent--
	g.line("}")
	g.line("if nibble >= 8 {")
	g.indent++
	g.line("return int8(nibble) - 16")
	g.indent--
	g.line("}")
	g.line("return int8(nibble)")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func (b *runeBinary) SetInt4(index int, value int8) int8 {")
	g.indent++
	g.line("if index < 0 { panic(\"binary offset out of range\") }")
	g.line("byteIndex := index / 2")
	g.line("b.check(byteIndex, 1)")
	g.line("nibble := byte(value) & 0x0f")
	g.line("if index%2 == 0 {")
	g.indent++
	g.line("b.data[byteIndex] = (b.data[byteIndex] & 0x0f) | (nibble << 4)")
	g.indent--
	g.line("} else {")
	g.indent++
	g.line("b.data[byteIndex] = (b.data[byteIndex] & 0xf0) | nibble")
	g.indent--
	g.line("}")
	g.line("return value")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func (b *runeBinary) GetInt8(offset int) int8 { b.check(offset, 1); return int8(b.data[offset]) }")
	g.line("func (b *runeBinary) SetInt8(offset int, value int8) int8 { b.check(offset, 1); b.data[offset] = byte(value); return value }")
	g.line("func (b *runeBinary) GetUInt8(offset int) uint8 { b.check(offset, 1); return b.data[offset] }")
	g.line("func (b *runeBinary) SetUInt8(offset int, value uint8) uint8 { b.check(offset, 1); b.data[offset] = value; return value }")
	g.line("")
	g.line("func (b *runeBinary) GetInt16(offset int, littleEndian bool) int16 { b.check(offset, 2); return int16(runeBinaryOrder(littleEndian).Uint16(b.data[offset:])) }")
	g.line("func (b *runeBinary) SetInt16(offset int, value int16, littleEndian bool) int16 { b.check(offset, 2); runeBinaryOrder(littleEndian).PutUint16(b.data[offset:], uint16(value)); return value }")
	g.line("func (b *runeBinary) GetUInt16(offset int, littleEndian bool) uint16 { b.check(offset, 2); return runeBinaryOrder(littleEndian).Uint16(b.data[offset:]) }")
	g.line("func (b *runeBinary) SetUInt16(offset int, value uint16, littleEndian bool) uint16 { b.check(offset, 2); runeBinaryOrder(littleEndian).PutUint16(b.data[offset:], value); return value }")
	g.line("")
	g.line("func (b *runeBinary) GetInt(offset int, littleEndian bool) int { b.check(offset, 4); return int(int32(runeBinaryOrder(littleEndian).Uint32(b.data[offset:]))) }")
	g.line("func (b *runeBinary) SetInt(offset int, value int, littleEndian bool) int { b.check(offset, 4); runeBinaryOrder(littleEndian).PutUint32(b.data[offset:], uint32(int32(value))); return value }")
	g.line("func (b *runeBinary) GetUInt(offset int, littleEndian bool) uint { b.check(offset, 4); return uint(runeBinaryOrder(littleEndian).Uint32(b.data[offset:])) }")
	g.line("func (b *runeBinary) SetUInt(offset int, value uint, littleEndian bool) uint { b.check(offset, 4); runeBinaryOrder(littleEndian).PutUint32(b.data[offset:], uint32(value)); return value }")
	g.line("")
	g.line("func (b *runeBinary) GetInt64(offset int, littleEndian bool) int64 { b.check(offset, 8); return int64(runeBinaryOrder(littleEndian).Uint64(b.data[offset:])) }")
	g.line("func (b *runeBinary) SetInt64(offset int, value int64, littleEndian bool) int64 { b.check(offset, 8); runeBinaryOrder(littleEndian).PutUint64(b.data[offset:], uint64(value)); return value }")
	g.line("func (b *runeBinary) GetUInt64(offset int, littleEndian bool) uint64 { b.check(offset, 8); return runeBinaryOrder(littleEndian).Uint64(b.data[offset:]) }")
	g.line("func (b *runeBinary) SetUInt64(offset int, value uint64, littleEndian bool) uint64 { b.check(offset, 8); runeBinaryOrder(littleEndian).PutUint64(b.data[offset:], value); return value }")
	g.line("")
	g.line("func (b *runeBinary) GetFloat(offset int, littleEndian bool) float32 { b.check(offset, 4); return math.Float32frombits(runeBinaryOrder(littleEndian).Uint32(b.data[offset:])) }")
	g.line("func (b *runeBinary) SetFloat(offset int, value float32, littleEndian bool) float32 { b.check(offset, 4); runeBinaryOrder(littleEndian).PutUint32(b.data[offset:], math.Float32bits(value)); return value }")
	g.line("func (b *runeBinary) GetDouble(offset int, littleEndian bool) float64 { b.check(offset, 8); return math.Float64frombits(runeBinaryOrder(littleEndian).Uint64(b.data[offset:])) }")
	g.line("func (b *runeBinary) SetDouble(offset int, value float64, littleEndian bool) float64 { b.check(offset, 8); runeBinaryOrder(littleEndian).PutUint64(b.data[offset:], math.Float64bits(value)); return value }")
}
