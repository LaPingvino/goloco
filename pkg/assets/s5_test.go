package assets

import (
	"encoding/binary"
	"testing"
)

// ---------------------------------------------------------------------------
// ParseS5Header
// ---------------------------------------------------------------------------

func TestParseS5Header_correct_layout(t *testing.T) {
	// Build a 32-byte header with known values
	var buf [32]byte
	buf[0] = 1                                                    // Type = scenario
	buf[1] = HeaderFlagHasSaveDetails | HeaderFlagIsTitleSequence // Flags
	binary.LittleEndian.PutUint16(buf[2:4], 3)                    // NumPackedObjects
	binary.LittleEndian.PutUint32(buf[4:8], 0x62262)              // Version
	binary.LittleEndian.PutUint32(buf[8:12], 0x62300)             // Magic

	h, err := ParseS5Header(buf[:])
	if err != nil {
		t.Fatalf("ParseS5Header: %v", err)
	}
	if h.Type != 1 {
		t.Errorf("Type = %d, want 1", h.Type)
	}
	if h.Flags != HeaderFlagHasSaveDetails|HeaderFlagIsTitleSequence {
		t.Errorf("Flags = 0x%02X, want 0x%02X", h.Flags, HeaderFlagHasSaveDetails|HeaderFlagIsTitleSequence)
	}
	if h.NumPackedObjects != 3 {
		t.Errorf("NumPackedObjects = %d, want 3", h.NumPackedObjects)
	}
	if h.Version != 0x62262 {
		t.Errorf("Version = 0x%X, want 0x62262", h.Version)
	}
	if h.Magic != 0x62300 {
		t.Errorf("Magic = 0x%X, want 0x62300", h.Magic)
	}
}

func TestParseS5Header_too_short(t *testing.T) {
	_, err := ParseS5Header([]byte{0, 1, 2})
	if err == nil {
		t.Fatal("expected error for short input, got nil")
	}
}

// ---------------------------------------------------------------------------
// decodeRotate
// ---------------------------------------------------------------------------

func TestDecodeRotate_roundtrip(t *testing.T) {
	// Encode then decode should be identity.
	// Encode: rotl by code; Decode: rotr by code.  Same cycling sequence.
	plain := []byte("Hello, Locomotion!")
	encoded := encodeRotate(plain)
	decoded := decodeRotate(encoded)
	if string(decoded) != string(plain) {
		t.Errorf("rotate roundtrip: got %q, want %q", decoded, plain)
	}
}

func TestDecodeRotate_known_vector(t *testing.T) {
	// Single byte: rotate right by 1.  Input 0x80 → rotr(0x80,1) = 0x40
	input := []byte{0x80}
	out := decodeRotate(input)
	if out[0] != 0x40 {
		t.Errorf("decodeRotate([0x80]) = 0x%02X, want 0x40", out[0])
	}
}

// encodeRotate is the inverse used only in tests (matches OpenLoco encodeRotate)
func encodeRotate(data []byte) []byte {
	out := make([]byte, len(data))
	code := uint(1)
	for i, b := range data {
		out[i] = (b << code) | (b >> (8 - code))
		code = (code + 2) & 7
	}
	return out
}

// ---------------------------------------------------------------------------
// decodeRunLengthSingle
// ---------------------------------------------------------------------------

func TestDecodeRunLengthSingle_literal_run(t *testing.T) {
	// code=2 (no high bit) → copy next 3 bytes literally
	input := []byte{2, 0xAA, 0xBB, 0xCC}
	out, err := decodeRunLengthSingle(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []byte{0xAA, 0xBB, 0xCC}
	if string(out) != string(want) {
		t.Errorf("literal run: got %v, want %v", out, want)
	}
}

func TestDecodeRunLengthSingle_repeat_run(t *testing.T) {
	// code=0xFE (high bit set) → repeat next byte 257-0xFE = 3 times
	input := []byte{0xFE, 0x42}
	out, err := decodeRunLengthSingle(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []byte{0x42, 0x42, 0x42}
	if string(out) != string(want) {
		t.Errorf("repeat run: got %v, want %v", out, want)
	}
}

func TestDecodeRunLengthSingle_mixed(t *testing.T) {
	// 3 literal bytes, then repeat 0x7F twice (257-0xFF=2)
	input := []byte{
		2, 0x01, 0x02, 0x03, // literal: copy 3 bytes
		0xFF, 0x7F, // repeat: 0x7F × 2
	}
	out, err := decodeRunLengthSingle(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []byte{0x01, 0x02, 0x03, 0x7F, 0x7F}
	if string(out) != string(want) {
		t.Errorf("mixed: got %v, want %v", out, want)
	}
}

func TestDecodeRunLengthSingle_truncated_repeat(t *testing.T) {
	input := []byte{0xFE} // high bit set but no value byte
	_, err := decodeRunLengthSingle(input)
	if err == nil {
		t.Fatal("expected error for truncated repeat, got nil")
	}
}

func TestDecodeRunLengthSingle_truncated_literal(t *testing.T) {
	input := []byte{5, 0x01, 0x02} // claims 6 literal bytes, only 2 present
	_, err := decodeRunLengthSingle(input)
	if err == nil {
		t.Fatal("expected error for truncated literal, got nil")
	}
}

// ---------------------------------------------------------------------------
// decodeRunLengthMulti
// ---------------------------------------------------------------------------

func TestDecodeRunLengthMulti_literal(t *testing.T) {
	// 0xFF followed by a byte → emit that byte literally
	input := []byte{0xFF, 0xAB, 0xFF, 0xCD}
	out, err := decodeRunLengthMulti(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []byte{0xAB, 0xCD}
	if string(out) != string(want) {
		t.Errorf("literals: got %v, want %v", out, want)
	}
}

func TestDecodeRunLengthMulti_backreference(t *testing.T) {
	// Seed 3 bytes via literals, then back-reference the first 2.
	//
	// Back-ref encoding: byte = ((32 - offset) << 3) | (copyLen - 1)
	// We want offset = -3 (copy from 3 bytes back), copyLen = 2.
	// offset in the encoding = 32 + (-3) = 29
	// byte = (29 << 3) | (2-1) = 232 | 1 = 233 = 0xE9
	input := []byte{
		0xFF, 0xAA, // literal AA
		0xFF, 0xBB, // literal BB
		0xFF, 0xCC, // literal CC
		0xE9, // back-ref: offset=-3, len=2 → copy AA BB
	}
	out, err := decodeRunLengthMulti(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []byte{0xAA, 0xBB, 0xCC, 0xAA, 0xBB}
	if string(out) != string(want) {
		t.Errorf("backref: got %v, want %v", out, want)
	}
}

func TestDecodeRunLengthMulti_truncated_literal(t *testing.T) {
	input := []byte{0xFF} // 0xFF with no following byte
	_, err := decodeRunLengthMulti(input)
	if err == nil {
		t.Fatal("expected error for truncated literal, got nil")
	}
}

func TestDecodeRunLengthMulti_invalid_offset(t *testing.T) {
	// Back-ref with offset that exceeds output length.
	// Offset = (byte>>3) - 32.  byte=0x00 → offset = 0-32 = -32.
	// Output is empty so -32 is invalid.
	input := []byte{0x00}
	_, err := decodeRunLengthMulti(input)
	if err == nil {
		t.Fatal("expected error for invalid back-reference offset, got nil")
	}
}

// ---------------------------------------------------------------------------
// Two-stage runLengthMulti decode (as used for tile elements)
// ---------------------------------------------------------------------------

func TestDecodeSawyerChunk_runLengthMulti_pipeline(t *testing.T) {
	// Verify the two-stage pipeline (runLengthSingle → runLengthMulti)
	// produces the expected output for a known input.
	//
	// We want the final output to be [0xAA, 0xBB, 0xAA].
	// Stage 2 (runLengthMulti) input must be:
	//   0xFF 0xAA   — literal AA
	//   0xFF 0xBB   — literal BB
	//   0xF8        — back-ref: offset=(0xF8>>3)-32 = 31-32 = -1, len=(0xF8&7)+1 = 1
	//                 → copy 1 byte from output[-1] = BB  … that gives [AA BB BB]
	// Actually we want AA at the end.  offset=-2 len=1:
	//   offset = -2 → (32 + (-2)) = 30, so byte = (30<<3) | 0 = 240 = 0xF0
	//   That copies output[len-2] = AA.  Result: [AA BB AA]. Correct!
	//
	// Stage 2 input: [0xFF, 0xAA, 0xFF, 0xBB, 0xF0]
	// Stage 1 must produce exactly that.  Encode as one literal run of 5:
	//   code = 4 (means copy 5 bytes literally)
	stage1Input := []byte{4, 0xFF, 0xAA, 0xFF, 0xBB, 0xF0}

	out, err := decodeSawyerChunk(EncodingRunLengthMulti, stage1Input)
	if err != nil {
		t.Fatalf("two-stage decode: %v", err)
	}
	want := []byte{0xAA, 0xBB, 0xAA}
	if string(out) != string(want) {
		t.Errorf("two-stage pipeline: got %v, want %v", out, want)
	}
}

// ---------------------------------------------------------------------------
// S5ChunkReader
// ---------------------------------------------------------------------------

func TestS5ChunkReader_reads_sequence(t *testing.T) {
	// Build a stream with two chunks:
	//   Chunk 1: uncompressed, payload = [0xDE, 0xAD]
	//   Chunk 2: uncompressed, payload = [0xBE, 0xEF, 0x00]
	var stream []byte

	// Chunk 1
	stream = append(stream, byte(EncodingUncompressed))
	stream = appendU32LE(stream, 2)
	stream = append(stream, 0xDE, 0xAD)

	// Chunk 2
	stream = append(stream, byte(EncodingUncompressed))
	stream = appendU32LE(stream, 3)
	stream = append(stream, 0xBE, 0xEF, 0x00)

	r := NewS5ChunkReader(stream)

	c1, err := r.ReadChunk()
	if err != nil {
		t.Fatalf("chunk 1: %v", err)
	}
	if string(c1) != string([]byte{0xDE, 0xAD}) {
		t.Errorf("chunk 1: got %v, want [DE AD]", c1)
	}

	c2, err := r.ReadChunk()
	if err != nil {
		t.Fatalf("chunk 2: %v", err)
	}
	if string(c2) != string([]byte{0xBE, 0xEF, 0x00}) {
		t.Errorf("chunk 2: got %v, want [BE EF 00]", c2)
	}

	// Third read should fail (stream exhausted)
	_, err = r.ReadChunk()
	if err == nil {
		t.Fatal("expected error on exhausted stream, got nil")
	}
}

func TestS5ChunkReader_AdvanceRaw(t *testing.T) {
	// 5 garbage bytes, then one uncompressed chunk [0x42]
	var stream []byte
	stream = append(stream, 0x00, 0x00, 0x00, 0x00, 0x00) // 5 raw bytes to skip
	stream = append(stream, byte(EncodingUncompressed))
	stream = appendU32LE(stream, 1)
	stream = append(stream, 0x42)

	r := NewS5ChunkReader(stream)
	r.AdvanceRaw(5)

	chunk, err := r.ReadChunk()
	if err != nil {
		t.Fatalf("after AdvanceRaw: %v", err)
	}
	if len(chunk) != 1 || chunk[0] != 0x42 {
		t.Errorf("after AdvanceRaw: got %v, want [42]", chunk)
	}
}

func TestS5ChunkReader_rotate_chunk(t *testing.T) {
	// Encode "Hi" with rotate, wrap in a chunk, decode via reader
	plain := []byte("Hi")
	encoded := encodeRotate(plain) // test helper defined above

	var stream []byte
	stream = append(stream, byte(EncodingRotate))
	stream = appendU32LE(stream, uint32(len(encoded)))
	stream = append(stream, encoded...)

	r := NewS5ChunkReader(stream)
	out, err := r.ReadChunk()
	if err != nil {
		t.Fatalf("rotate chunk: %v", err)
	}
	if string(out) != "Hi" {
		t.Errorf("rotate chunk: got %q, want %q", out, "Hi")
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func appendU32LE(buf []byte, v uint32) []byte {
	var tmp [4]byte
	binary.LittleEndian.PutUint32(tmp[:], v)
	return append(buf, tmp[:]...)
}
