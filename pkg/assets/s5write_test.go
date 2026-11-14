package assets

import (
	"bytes"
	"math/rand"
	"testing"
)

// TestEncodeSawyer_roundtrip: every encoder's output decodes back to its
// input (SawyerStreamWriter::encode ∘ SawyerStreamReader::decode = id).
func TestEncodeSawyer_roundtrip(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	inputs := [][]byte{{}, {7}, {1, 1}, {1, 2}, bytes.Repeat([]byte{0}, 1000), bytes.Repeat([]byte{1, 2, 3}, 400)}
	for n := 0; n < 50; n++ {
		b := make([]byte, rng.Intn(3000))
		for i := range b {
			if rng.Intn(3) == 0 {
				b[i] = byte(rng.Intn(256))
			} else if i > 0 {
				b[i] = b[i-1]
			}
		}
		inputs = append(inputs, b)
	}
	for _, in := range inputs {
		for _, enc := range []SawyerEncoding{EncodingUncompressed, EncodingRunLengthSingle, EncodingRunLengthMulti, EncodingRotate} {
			got, err := decodeSawyerChunk(enc, EncodeSawyerChunk(enc, in))
			if err != nil || !bytes.Equal(got, in) {
				t.Fatalf("encoding %d, %d bytes: round trip failed (%v)", enc, len(in), err)
			}
		}
	}
}

// TestS5ChunkWriter_stream: a written stream reads back chunk by chunk and its
// checksum validates.
func TestS5ChunkWriter_stream(t *testing.T) {
	var buf bytes.Buffer
	w := NewS5ChunkWriter(&buf)
	w.WriteChunk(EncodingRotate, []byte("header"))
	w.Write([]byte{1, 2, 3})
	w.WriteChunk(EncodingRunLengthMulti, bytes.Repeat([]byte{9, 8}, 100))
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, ok := S5Checksum(buf.Bytes()); !ok {
		t.Fatal("checksum does not validate")
	}
	r := NewS5ChunkReader(buf.Bytes())
	if c, _ := r.ReadChunk(); string(c) != "header" {
		t.Fatalf("chunk 1 = %q", c)
	}
	r.AdvanceRaw(3)
	if c, _ := r.ReadChunk(); !bytes.Equal(c, bytes.Repeat([]byte{9, 8}, 100)) {
		t.Fatal("chunk 2 differs")
	}
}
