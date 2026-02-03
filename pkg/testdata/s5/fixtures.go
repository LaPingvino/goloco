package s5testdata

// Small S5 chunk fixtures for tests.
// These are minimal, human-constructed byte slices that represent either:
// - a compressed payload (the raw bytes to pass to decodeSawyerChunk), or
// - a framed chunk stream (encoding byte + 4-byte LE length + payload) that
//   can be fed into S5ChunkReader as part of a stream.
//
// Tests can import this package and use either the Payload variants (to call
// decodeSawyerChunk directly) or the Chunk variants (to build a full stream
// for NewS5ChunkReader).
//
// Note: the numeric encoding values match the SawyerEncoding constants in the
// codebase:
//   EncodingUncompressed   = 0
//   EncodingRunLengthSingle = 1
//   EncodingRunLengthMulti  = 2
//   EncodingRotate          = 3
//
// These fixtures are intentionally tiny and illustrative.

var (
	// Uncompressed example
	// Payload bytes are just copied through by the decoder.
	UncompressedPayload = []byte{0xDE, 0xAD}
	// Chunk: [encoding=0][len=2 little-endian][payload...]
	UncompressedChunk = []byte{
		0x00,                   // EncodingUncompressed
		0x02, 0x00, 0x00, 0x00, // length = 2
		0xDE, 0xAD,
	}

	// RunLengthSingle (literal run) example
	// Payload: code=2 -> copy next (2+1)=3 bytes literally
	RunLengthSingleLiteralPayload = []byte{
		0x02, 0xAA, 0xBB, 0xCC,
	}
	RunLengthSingleLiteralChunk = []byte{
		0x01,                   // EncodingRunLengthSingle
		0x04, 0x00, 0x00, 0x00, // length = 4
		0x02, 0xAA, 0xBB, 0xCC,
	}

	// RunLengthSingle (repeat run) example
	// Payload: code=0xFE (high bit set) -> repeat next byte 257-0xFE = 3 times
	RunLengthSingleRepeatPayload = []byte{
		0xFE, 0x42,
	}
	RunLengthSingleRepeatChunk = []byte{
		0x01,                   // EncodingRunLengthSingle
		0x02, 0x00, 0x00, 0x00, // length = 2
		0xFE, 0x42,
	}

	// RunLengthMulti example (two-stage): this payload is intended to be fed
	// to decodeSawyerChunk with EncodingRunLengthMulti. The code path will
	// first run decodeRunLengthSingle on the payload, then runLengthMulti.
	//
	// The stage-1 payload here is:
	//   [4, 0xFF, 0xAA, 0xFF, 0xBB, 0xF0]
	// decodeRunLengthSingle interprets the leading 4 as "copy next 5 bytes
	// literally", producing [0xFF,0xAA,0xFF,0xBB,0xF0]. decodeRunLengthMulti
	// then interprets that sequence to produce the final output [0xAA,0xBB,0xAA].
	RunLengthMultiStage1Payload = []byte{
		0x04, 0xFF, 0xAA, 0xFF, 0xBB, 0xF0,
	}
	RunLengthMultiChunk = []byte{
		0x02,                   // EncodingRunLengthMulti
		0x06, 0x00, 0x00, 0x00, // length = 6
		0x04, 0xFF, 0xAA, 0xFF, 0xBB, 0xF0,
	}

	// Rotate encoding example
	// This payload is the encoded form of the ASCII string "Hi" using the
	// rotate-encoding scheme (rotate-left for encode, decodeRotate performs
	// the inverse rotation-right).
	// Encoded bytes computed for "Hi" => [0x90, 0x4B]
	RotatePayload = []byte{0x90, 0x4B}
	RotateChunk   = []byte{
		0x03,                   // EncodingRotate
		0x02, 0x00, 0x00, 0x00, // length = 2
		0x90, 0x4B,
	}
)
