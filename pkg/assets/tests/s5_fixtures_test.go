package assets_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/LaPingvino/goloco/pkg/assets"
)

// Helper: build a framed Sawyer chunk: [1 byte encoding][4 bytes LE length][payload...]
func makeChunk(encoding byte, payload []byte) []byte {
	buf := make([]byte, 0, 5+len(payload))
	buf = append(buf, encoding)
	var lenBuf [4]byte
	binary.LittleEndian.PutUint32(lenBuf[:], uint32(len(payload)))
	buf = append(buf, lenBuf[:]...)
	buf = append(buf, payload...)
	return buf
}

// Test that an uncompressed chunk is passed through verbatim.
func TestS5ChunkReader_Uncompressed(t *testing.T) {
	payload := []byte{0xDE, 0xAD}
	stream := makeChunk(0x00, payload) // EncodingUncompressed == 0

	r := assets.NewS5ChunkReader(stream)
	out, err := r.ReadChunk()
	if err != nil {
		t.Fatalf("ReadChunk failed: %v", err)
	}
	if !bytes.Equal(out, payload) {
		t.Fatalf("uncompressed payload mismatch: got %v want %v", out, payload)
	}
}

// Test rotate encoding by supplying a small pre-encoded payload that should decode to "Hi".
// Nudges: if this fails, inspect decodeRotate() in:
//
//	goloco-project/goloco/pkg/assets/s5.go
func TestS5ChunkReader_Rotate(t *testing.T) {
	// Encoded form that decodeRotate should map back to "Hi".
	// Chosen bytes are consistent with the project's rotate scheme in tests elsewhere.
	encoded := []byte{0x90, 0x4B}      // encoded payload -> expected plain "Hi"
	stream := makeChunk(0x03, encoded) // EncodingRotate == 3

	r := assets.NewS5ChunkReader(stream)
	out, err := r.ReadChunk()
	if err != nil {
		t.Fatalf("rotate ReadChunk failed: %v", err)
	}
	if string(out) != "Hi" {
		t.Fatalf("rotate decoded wrong string: got %q want %q", string(out), "Hi")
	}
}

// Test runLengthSingle literal run decoding.
// Format reminder (see s5.go): code < 0x80 => copy (code+1) bytes literally.
func TestS5ChunkReader_RunLengthSingle_Literal(t *testing.T) {
	payload := []byte{0x02, 0xAA, 0xBB, 0xCC} // code=2 -> copy 3 bytes
	stream := makeChunk(0x01, payload)        // EncodingRunLengthSingle == 1

	r := assets.NewS5ChunkReader(stream)
	out, err := r.ReadChunk()
	if err != nil {
		t.Fatalf("RLS literal ReadChunk failed: %v", err)
	}
	expected := []byte{0xAA, 0xBB, 0xCC}
	if !bytes.Equal(out, expected) {
		t.Fatalf("RLS literal mismatch: got %v want %v", out, expected)
	}
}

// Test runLengthSingle repeat run decoding.
// Format reminder (see s5.go): code & 0x80 != 0 => repeat next byte (257 - code) times.
func TestS5ChunkReader_RunLengthSingle_Repeat(t *testing.T) {
	payload := []byte{0xFE, 0x42} // code=0xFE => repeat next byte 257-0xFE = 3 times
	stream := makeChunk(0x01, payload)

	r := assets.NewS5ChunkReader(stream)
	out, err := r.ReadChunk()
	if err != nil {
		t.Fatalf("RLS repeat ReadChunk failed: %v", err)
	}
	expected := []byte{0x42, 0x42, 0x42}
	if !bytes.Equal(out, expected) {
		t.Fatalf("RLS repeat mismatch: got %v want %v", out, expected)
	}
}

// Two-stage runLengthMulti pipeline test (runLengthSingle -> runLengthMulti).
//
// Stage-1 payload chosen so that decodeRunLengthSingle(stage1) -> stage2Input,
// and decodeRunLengthMulti(stage2Input) -> final expected output.
//
// This mirrors the two-stage pipeline implemented in decodeSawyerChunk().
// Nudges: if this fails, inspect decodeRunLengthSingle / decodeRunLengthMulti
// in goloco-project/goloco/pkg/assets/s5.go for pipeline ordering and back-reference logic.
func TestS5ChunkReader_RunLengthMulti_Pipeline(t *testing.T) {
	// Stage-1 payload: leading 4 means copy next 5 bytes literally:
	// stage1 payload = [4, 0xFF, 0xAA, 0xFF, 0xBB, 0xF0]
	// After stage1 decode => [0xFF, 0xAA, 0xFF, 0xBB, 0xF0]
	// Then decodeRunLengthMulti interprets:
	//   0xFF 0xAA -> literal 0xAA
	//   0xFF 0xBB -> literal 0xBB
	//   0xF0 -> back-ref byte that yields 0xAA as final byte
	stage1 := []byte{0x04, 0xFF, 0xAA, 0xFF, 0xBB, 0xF0}
	stream := makeChunk(0x02, stage1) // EncodingRunLengthMulti == 2

	r := assets.NewS5ChunkReader(stream)
	out, err := r.ReadChunk()
	if err != nil {
		t.Fatalf("RLM pipeline ReadChunk failed: %v", err)
	}
	expected := []byte{0xAA, 0xBB, 0xAA}
	if !bytes.Equal(out, expected) {
		t.Fatalf("RLM pipeline mismatch: got %v want %v", out, expected)
	}
}

// Test that S5ChunkReader reports exhaustion when stream has been fully read.
func TestS5ChunkReader_Exhaustion(t *testing.T) {
	stream := makeChunk(0x00, []byte{0x42}) // one uncompressed chunk
	r := assets.NewS5ChunkReader(stream)

	_, err := r.ReadChunk()
	if err != nil {
		t.Fatalf("first ReadChunk unexpected error: %v", err)
	}
	_, err = r.ReadChunk()
	if err == nil {
		t.Fatal("expected error on exhausted stream, got nil")
	}
}
