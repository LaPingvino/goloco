package assets

import (
	"testing"
)

// RLE format reminder (used by decodeRLEElement):
// - src begins with height*2 bytes: uint16 LE offsets (one per line) relative to src start
// - each offset points to a sequence of run chunks for that line
// - each run chunk:
//     [dataSize byte][firstX byte][dataSize & 0x7F bytes of pixel indices]
//     dataSize high bit (0x80) signals this is the final chunk for that line
//
// The tests below are intentionally defensive and include comments with
// small nudges for future debugging. Per project rules: do not modify
// non-test code from tests; instead, surface issues as test failures and
// actionable comments here so the next engineer (or Claude) can act safely.

// Helper to write a little-endian uint16 into a buffer at a given pos.
func putU16LE(b []byte, pos int, v uint16) {
	b[pos] = byte(v & 0xFF)
	b[pos+1] = byte(v >> 8)
}

// Test a single-line element with a single end-marked chunk that fills the width.
func TestDecodeRLEElement_SingleLineSimple(t *testing.T) {
	width := 3
	height := 1

	// Offsets table takes height*2 = 2 bytes. First chunk starts immediately after it at offset 2.
	// Build src: [offset0 lo][offset0 hi] [chunk...]
	src := []byte{0x02, 0x00} // offset = 2

	// Chunk: dataSize (0x80 | len=3), firstX=0, data=1,2,3
	src = append(src, 0x80|byte(3), 0x00, 1, 2, 3)

	out, err := decodeRLEElement(src, width, height)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != width*height {
		t.Fatalf("unexpected output length: got %d want %d", len(out), width*height)
	}
	if out[0] != 1 || out[1] != 2 || out[2] != 3 {
		t.Fatalf("unexpected pixels: %v", out[:3])
	}

	// Nudge for debuggers:
	// If this test fails, first check that the offsets are interpreted as little-endian
	// and relative to the start of `src` (not relative to some other header base).
}

// Test two lines, first line fills width, second line writes starting at firstX=1 with 2 pixels.
func TestDecodeRLEElement_MultiLine(t *testing.T) {
	width := 4
	height := 2

	// We'll construct:
	// offsets: [4, 10] (two uint16 LE values)
	// bytes layout:
	// 0-3: offsets (4,10)
	// 4-9: line0 chunk (dataSize=0x84, firstX=0, 4 bytes)
	// 10-?: line1 chunk (dataSize=0x82, firstX=1, 2 bytes)
	//
	// Build buffer with room
	src := make([]byte, 0, 32)
	src = append(src, 0, 0, 0, 0) // placeholder for offsets

	// first chunk at offset 4
	// dataSize 0x80|4 = 0x84, firstX=0, data 10,11,12,13
	src = append(src, 0x80|byte(4), 0x00, 10, 11, 12, 13)
	// second chunk starts here; record its index
	secondOffset := len(src)
	// dataSize 0x80|2 = 0x82, firstX=1, data 20,21
	src = append(src, 0x80|byte(2), 0x01, 20, 21)

	// now patch offsets at start as little-endian u16
	putU16LE(src, 0, uint16(4))
	putU16LE(src, 2, uint16(secondOffset))

	out, err := decodeRLEElement(src, width, height)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != width*height {
		t.Fatalf("unexpected output length: got %d want %d", len(out), width*height)
	}
	// first line should be [10,11,12,13]
	if out[0] != 10 || out[1] != 11 || out[2] != 12 || out[3] != 13 {
		t.Fatalf("line0 mismatch: %v", out[0:4])
	}
	// second line should be [0,20,21,0] (firstX=1 writes into positions 1 and 2)
	if out[4] != 0 || out[5] != 20 || out[6] != 21 || out[7] != 0 {
		t.Fatalf("line1 mismatch: %v", out[4:8])
	}

	// Nudge:
	// If this fails, examine whether decodeRLEElement handles multiple lines' offsets properly
	// and whether offsets pointing exactly after the offsets table are handled correctly.
}

// Truncated offsets should return an explicit error early.
func TestDecodeRLEElement_TruncatedOffsets(t *testing.T) {
	width := 4
	height := 2

	// Provide only 2 bytes when 4 are required for offsets.
	src := []byte{0x01, 0x00}

	_, err := decodeRLEElement(src, width, height)
	if err == nil {
		t.Fatal("expected error for truncated offsets, got nil")
	}
	// Nudge: decodeRLEElement should detect len(src) < height*2 and return quickly.
}

// If a chunk header (dataSize+firstX) does not fit in buffer, decoder should error.
func TestDecodeRLEElement_ChunkHeaderTruncated(t *testing.T) {
	width := 3
	height := 1

	// offsets area size = 2 bytes -> chunk header expected at offset 2
	src := []byte{0x02, 0x00}
	// but provide only one byte for header (truncated)
	src = append(src, 0x05) // dataSize only, missing firstX and data

	_, err := decodeRLEElement(src, width, height)
	if err == nil {
		t.Fatal("expected error for truncated chunk header, got nil")
	}
	// Nudge:
	// The implementation checks `if p+2 > len(src)` and should return an error here.
}

// If chunk data length claims more bytes than available, decoder should error.
func TestDecodeRLEElement_ChunkDataTruncated(t *testing.T) {
	width := 5
	height := 1

	src := []byte{0x02, 0x00} // offset=2
	// dataSize says 4 bytes (not end-marked), but only provide 2 bytes -> truncated
	src = append(src, 0x04, 0x00, 1, 2) // claims 4 data bytes but only 2 present

	_, err := decodeRLEElement(src, width, height)
	if err == nil {
		t.Fatal("expected error for truncated chunk data, got nil")
	}
	// Nudge:
	// The code checks `if p+dataSize > len(src)` and returns an error. This test ensures that behavior.
}

// If an offset value points outside the src buffer, we expect an explicit error.
func TestDecodeRLEElement_OffsetOutOfRange(t *testing.T) {
	width := 2
	height := 1

	// offsets area -> 2 bytes. Set offset to a value beyond len(src)
	src := make([]byte, 2)
	// Set offset to 0xFFFE, which will be > len(src)
	putU16LE(src, 0, 0xFFFE)

	_, err := decodeRLEElement(src, width, height)
	if err == nil {
		t.Fatal("expected error for out-of-range offset, got nil")
	}
	// Nudge:
	// If this test fails and returns nil or panic, consider adding bounds checks when reading offsets.
}

// Ensure decoder clamps writes to the width (doesn't panic or overflow)
func TestDecodeRLEElement_ClampsWritesToWidth(t *testing.T) {
	width := 3
	height := 1

	// offsets: 2
	src := []byte{0x02, 0x00}
	// Create a chunk that claims to start at firstX=1 but provides 5 bytes (exceeds width)
	// dataSize marked end-of-line: 0x80|5 = 0x85
	src = append(src, 0x80|byte(5), 0x01, 9, 10, 11, 12, 13)

	out, err := decodeRLEElement(src, width, height)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Expect that only positions 1 and 2 are written (width=3), index 0 remains zero.
	if out[0] != 0 {
		t.Fatalf("expected out[0]==0, got %d", out[0])
	}
	if out[1] != 9 || out[2] != 10 {
		t.Fatalf("unexpected values for clamped output: %v", out[:3])
	}

	// Nudge:
	// If behavior is wrong here (panic or overwrites out-of-range), add or adjust guards in the
	// inner write loop. Tests are deliberately written to surface that exact problem.
}

// Minor integration test to ensure decodeRLEElement can handle several small chunks for one line.
func TestDecodeRLEElement_MultipleChunksPerLine(t *testing.T) {
	width := 6
	height := 1

	// offsets size = 2 -> first chunk starts at 2
	// We'll create two chunks:
	//  - chunk A: literal 2 bytes at firstX=0 (dataSize=2, not end)
	//  - chunk B: literal 3 bytes at firstX=2 (dataSize=0x80|3 end)
	src := []byte{0x02, 0x00}

	// chunk A: dataSize=2 (copy 3 bytes? careful: value 2 means copy 3 bytes per Sawyer? NO.
	// Our decodeRLEElement expects dataSize to be stored as a single byte where high bit indicates end
	// and lower 7 bits is the count. In our implementation we simply use dataSize & 0x7F as his length.
	// So to copy 2 bytes without end: dataSize=2
	src = append(src, 2, 0x00, 1, 2) // copy 2 bytes at x=0
	// chunk B: end-marked with length 3
	src = append(src, 0x80|byte(3), 0x02, 3, 4, 5)

	out, err := decodeRLEElement(src, width, height)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Expect out[0..5] = [1,2,3,4,5,0]
	expected := []byte{1, 2, 3, 4, 5, 0}
	for i := 0; i < width; i++ {
		if out[i] != expected[i] {
			t.Fatalf("mismatch at %d: got %d want %d (full output %v)", i, out[i], expected[i], out[:width])
		}
	}

	// Nudge:
	// If this test starts failing after future changes, double-check the loop that advances `p`
	// and the `isEnd` handling (dataSize high bit). Off-by-one errors here are common.
}
