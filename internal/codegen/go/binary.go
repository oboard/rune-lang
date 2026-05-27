package gocodegen

func (g *generator) bytesRuntime() {
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
	g.line("type runeBytes struct {")
	g.indent++
	g.line("data []byte")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func newRuneBytes(length int) *runeBytes {")
	g.indent++
	g.line("if length < 0 {")
	g.indent++
	g.line("panic(\"bytes length out of range\")")
	g.indent--
	g.line("}")
	g.line("return &runeBytes{data: make([]byte, length)}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeBytesFromInts(values []int) *runeBytes {")
	g.indent++
	g.line("out := make([]byte, len(values))")
	g.line("for i, value := range values {")
	g.indent++
	g.line("out[i] = byte(value)")
	g.indent--
	g.line("}")
	g.line("return &runeBytes{data: out}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func runeBytesOrder(littleEndian bool) binary.ByteOrder {")
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
	g.line("func (b *runeBytes) check(offset int, size int) {")
	g.indent++
	g.line("if offset < 0 || size < 0 || offset+size > len(b.data) {")
	g.indent++
	g.line("panic(\"bytes offset out of range\")")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func (b *runeBytes) ByteLength() int { return len(b.data) }")
	g.line("")
	g.line("func (b *runeBytes) Clone() *runeBytes {")
	g.indent++
	g.line("return &runeBytes{data: append([]byte(nil), b.data...)}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func (b *runeBytes) Slice(start int, end int) *runeBytes {")
	g.indent++
	g.line("b.check(start, end-start)")
	g.line("return &runeBytes{data: append([]byte(nil), b.data[start:end]...)}")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func (b *runeBytes) ToInts() []int {")
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
	g.line("func (b *runeBytes) GetInt4(index int) int8 {")
	g.indent++
	g.line("if index < 0 { panic(\"bytes offset out of range\") }")
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
	g.line("func (b *runeBytes) SetInt4(index int, value int8) int8 {")
	g.indent++
	g.line("if index < 0 { panic(\"bytes offset out of range\") }")
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
	g.line("func (b *runeBytes) GetInt8(offset int) int8 { b.check(offset, 1); return int8(b.data[offset]) }")
	g.line("func (b *runeBytes) SetInt8(offset int, value int8) int8 { b.check(offset, 1); b.data[offset] = byte(value); return value }")
	g.line("func (b *runeBytes) GetUInt8(offset int) uint8 { b.check(offset, 1); return b.data[offset] }")
	g.line("func (b *runeBytes) SetUInt8(offset int, value uint8) uint8 { b.check(offset, 1); b.data[offset] = value; return value }")
	g.line("")
	g.line("func (b *runeBytes) GetInt16(offset int, littleEndian bool) int16 { b.check(offset, 2); return int16(runeBytesOrder(littleEndian).Uint16(b.data[offset:])) }")
	g.line("func (b *runeBytes) SetInt16(offset int, value int16, littleEndian bool) int16 { b.check(offset, 2); runeBytesOrder(littleEndian).PutUint16(b.data[offset:], uint16(value)); return value }")
	g.line("func (b *runeBytes) GetUInt16(offset int, littleEndian bool) uint16 { b.check(offset, 2); return runeBytesOrder(littleEndian).Uint16(b.data[offset:]) }")
	g.line("func (b *runeBytes) SetUInt16(offset int, value uint16, littleEndian bool) uint16 { b.check(offset, 2); runeBytesOrder(littleEndian).PutUint16(b.data[offset:], value); return value }")
	g.line("")
	g.line("func (b *runeBytes) GetInt(offset int, littleEndian bool) int { b.check(offset, 4); return int(int32(runeBytesOrder(littleEndian).Uint32(b.data[offset:]))) }")
	g.line("func (b *runeBytes) SetInt(offset int, value int, littleEndian bool) int { b.check(offset, 4); runeBytesOrder(littleEndian).PutUint32(b.data[offset:], uint32(int32(value))); return value }")
	g.line("func (b *runeBytes) GetUInt(offset int, littleEndian bool) uint { b.check(offset, 4); return uint(runeBytesOrder(littleEndian).Uint32(b.data[offset:])) }")
	g.line("func (b *runeBytes) SetUInt(offset int, value uint, littleEndian bool) uint { b.check(offset, 4); runeBytesOrder(littleEndian).PutUint32(b.data[offset:], uint32(value)); return value }")
	g.line("")
	g.line("func (b *runeBytes) GetInt64(offset int, littleEndian bool) int64 { b.check(offset, 8); return int64(runeBytesOrder(littleEndian).Uint64(b.data[offset:])) }")
	g.line("func (b *runeBytes) SetInt64(offset int, value int64, littleEndian bool) int64 { b.check(offset, 8); runeBytesOrder(littleEndian).PutUint64(b.data[offset:], uint64(value)); return value }")
	g.line("func (b *runeBytes) GetUInt64(offset int, littleEndian bool) uint64 { b.check(offset, 8); return runeBytesOrder(littleEndian).Uint64(b.data[offset:]) }")
	g.line("func (b *runeBytes) SetUInt64(offset int, value uint64, littleEndian bool) uint64 { b.check(offset, 8); runeBytesOrder(littleEndian).PutUint64(b.data[offset:], value); return value }")
	g.line("")
	g.line("func (b *runeBytes) GetFloat(offset int, littleEndian bool) float32 { b.check(offset, 4); return math.Float32frombits(runeBytesOrder(littleEndian).Uint32(b.data[offset:])) }")
	g.line("func (b *runeBytes) SetFloat(offset int, value float32, littleEndian bool) float32 { b.check(offset, 4); runeBytesOrder(littleEndian).PutUint32(b.data[offset:], math.Float32bits(value)); return value }")
	g.line("func (b *runeBytes) GetDouble(offset int, littleEndian bool) float64 { b.check(offset, 8); return math.Float64frombits(runeBytesOrder(littleEndian).Uint64(b.data[offset:])) }")
	g.line("func (b *runeBytes) SetDouble(offset int, value float64, littleEndian bool) float64 { b.check(offset, 8); runeBytesOrder(littleEndian).PutUint64(b.data[offset:], math.Float64bits(value)); return value }")
	g.line("")
	g.line("type runeBuffer struct {")
	g.indent++
	g.line("data []byte")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func newRuneBuffer() *runeBuffer { return &runeBuffer{} }")
	g.line("func newRuneBufferFromBytes(value *runeBytes) *runeBuffer { return &runeBuffer{data: append([]byte(nil), value.data...)} }")
	g.line("func (b *runeBuffer) ByteLength() int { return len(b.data) }")
	g.line("func (b *runeBuffer) Clear() { b.data = nil }")
	g.line("func (b *runeBuffer) Clone() *runeBuffer { return &runeBuffer{data: append([]byte(nil), b.data...)} }")
	g.line("func (b *runeBuffer) ToBytes() *runeBytes { return &runeBytes{data: append([]byte(nil), b.data...)} }")
	g.line("func (b *runeBuffer) ToInts() []int { return (&runeBytes{data: b.data}).ToInts() }")
	g.line("func (b *runeBuffer) Append(value uint8) *runeBuffer { b.data = append(b.data, value); return b }")
	g.line("func (b *runeBuffer) AppendInt(value int) *runeBuffer { b.data = append(b.data, byte(value)); return b }")
	g.line("func (b *runeBuffer) AppendBytes(value *runeBytes) *runeBuffer { b.data = append(b.data, value.data...); return b }")
	g.line("func (b *runeBuffer) Reader() *runeReader { return &runeReader{data: append([]byte(nil), b.data...)} }")
	g.line("func (b *runeBuffer) Writer() *runeWriter { return &runeWriter{data: append([]byte(nil), b.data...)} }")
	g.line("")
	g.line("type runeReader struct {")
	g.indent++
	g.line("data []byte")
	g.line("offset int")
	g.line("nibble int")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func newRuneReader(value *runeBytes) *runeReader { return &runeReader{data: append([]byte(nil), value.data...)} }")
	g.line("func (r *runeReader) ByteLength() int { return len(r.data) }")
	g.line("func (r *runeReader) Position() int { return r.offset }")
	g.line("func (r *runeReader) Remaining() int { return len(r.data) - r.offset }")
	g.line("func (r *runeReader) Seek(position int) int { if position < 0 || position > len(r.data) { panic(\"reader offset out of range\") }; r.offset = position; r.nibble = 0; return r.offset }")
	g.line("func (r *runeReader) Skip(count int) int { return r.Seek(r.offset + count) }")
	g.line("func (r *runeReader) align() { if r.nibble == 1 { r.offset++; r.nibble = 0 }; if r.offset > len(r.data) { panic(\"reader offset out of range\") } }")
	g.line("func (r *runeReader) take(size int) *runeBytes { r.align(); if size < 0 || r.offset+size > len(r.data) { panic(\"reader offset out of range\") }; start := r.offset; r.offset += size; return &runeBytes{data: r.data[start:r.offset]} }")
	g.line("func (r *runeReader) ReadBytes(length int) *runeBytes { return (&runeBytes{data: append([]byte(nil), r.take(length).data...)}) }")
	g.line("func (r *runeReader) ReadInt4() int8 { if r.offset < 0 || r.offset >= len(r.data) { panic(\"reader offset out of range\") }; value := (&runeBytes{data: r.data}).GetInt4(r.offset*2 + r.nibble); if r.nibble == 0 { r.nibble = 1 } else { r.nibble = 0; r.offset++ }; return value }")
	g.line("func (r *runeReader) ReadInt8() int8 { return r.take(1).GetInt8(0) }")
	g.line("func (r *runeReader) ReadUInt8() uint8 { return r.take(1).GetUInt8(0) }")
	g.line("func (r *runeReader) ReadInt16(littleEndian bool) int16 { return r.take(2).GetInt16(0, littleEndian) }")
	g.line("func (r *runeReader) ReadUInt16(littleEndian bool) uint16 { return r.take(2).GetUInt16(0, littleEndian) }")
	g.line("func (r *runeReader) ReadInt(littleEndian bool) int { return r.take(4).GetInt(0, littleEndian) }")
	g.line("func (r *runeReader) ReadUInt(littleEndian bool) uint { return r.take(4).GetUInt(0, littleEndian) }")
	g.line("func (r *runeReader) ReadInt64(littleEndian bool) int64 { return r.take(8).GetInt64(0, littleEndian) }")
	g.line("func (r *runeReader) ReadUInt64(littleEndian bool) uint64 { return r.take(8).GetUInt64(0, littleEndian) }")
	g.line("func (r *runeReader) ReadFloat(littleEndian bool) float32 { return r.take(4).GetFloat(0, littleEndian) }")
	g.line("func (r *runeReader) ReadDouble(littleEndian bool) float64 { return r.take(8).GetDouble(0, littleEndian) }")
	g.line("")
	g.line("type runeWriter struct {")
	g.indent++
	g.line("data []byte")
	g.line("nibble int")
	g.indent--
	g.line("}")
	g.line("")
	g.line("func newRuneWriter() *runeWriter { return &runeWriter{} }")
	g.line("func newRuneWriterWithCapacity(capacity int) *runeWriter { if capacity < 0 { panic(\"writer capacity out of range\") }; return &runeWriter{data: make([]byte, 0, capacity)} }")
	g.line("func (w *runeWriter) Position() int { return len(w.data) }")
	g.line("func (w *runeWriter) Clear() { w.data = nil; w.nibble = 0 }")
	g.line("func (w *runeWriter) ToBytes() *runeBytes { return &runeBytes{data: append([]byte(nil), w.data...)} }")
	g.line("func (w *runeWriter) ToInts() []int { return (&runeBytes{data: w.data}).ToInts() }")
	g.line("func (w *runeWriter) align() { w.nibble = 0 }")
	g.line("func (w *runeWriter) appendBytes(value *runeBytes) *runeWriter { w.align(); w.data = append(w.data, value.data...); return w }")
	g.line("func (w *runeWriter) WriteBytes(value *runeBytes) *runeWriter { return w.appendBytes(value) }")
	g.line("func (w *runeWriter) WriteInt4(value int8) *runeWriter { nibble := byte(value) & 0x0f; if w.nibble == 0 { w.data = append(w.data, nibble << 4); w.nibble = 1 } else { w.data[len(w.data)-1] = (w.data[len(w.data)-1] & 0xf0) | nibble; w.nibble = 0 }; return w }")
	g.line("func (w *runeWriter) WriteInt8(value int8) *runeWriter { b := newRuneBytes(1); b.SetInt8(0, value); return w.appendBytes(b) }")
	g.line("func (w *runeWriter) WriteUInt8(value uint8) *runeWriter { b := newRuneBytes(1); b.SetUInt8(0, value); return w.appendBytes(b) }")
	g.line("func (w *runeWriter) WriteInt16(value int16, littleEndian bool) *runeWriter { b := newRuneBytes(2); b.SetInt16(0, value, littleEndian); return w.appendBytes(b) }")
	g.line("func (w *runeWriter) WriteUInt16(value uint16, littleEndian bool) *runeWriter { b := newRuneBytes(2); b.SetUInt16(0, value, littleEndian); return w.appendBytes(b) }")
	g.line("func (w *runeWriter) WriteInt(value int, littleEndian bool) *runeWriter { b := newRuneBytes(4); b.SetInt(0, value, littleEndian); return w.appendBytes(b) }")
	g.line("func (w *runeWriter) WriteUInt(value uint, littleEndian bool) *runeWriter { b := newRuneBytes(4); b.SetUInt(0, value, littleEndian); return w.appendBytes(b) }")
	g.line("func (w *runeWriter) WriteInt64(value int64, littleEndian bool) *runeWriter { b := newRuneBytes(8); b.SetInt64(0, value, littleEndian); return w.appendBytes(b) }")
	g.line("func (w *runeWriter) WriteUInt64(value uint64, littleEndian bool) *runeWriter { b := newRuneBytes(8); b.SetUInt64(0, value, littleEndian); return w.appendBytes(b) }")
	g.line("func (w *runeWriter) WriteFloat(value float32, littleEndian bool) *runeWriter { b := newRuneBytes(4); b.SetFloat(0, value, littleEndian); return w.appendBytes(b) }")
	g.line("func (w *runeWriter) WriteDouble(value float64, littleEndian bool) *runeWriter { b := newRuneBytes(8); b.SetDouble(0, value, littleEndian); return w.appendBytes(b) }")
}
