package assets

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
)

// SawyerEncoding identifies the compression applied to a single chunk.
// Values match OpenLoco's SawyerEncoding enum exactly.
type SawyerEncoding uint8

const (
	EncodingUncompressed    SawyerEncoding = 0
	EncodingRunLengthSingle SawyerEncoding = 1
	EncodingRunLengthMulti  SawyerEncoding = 2
	EncodingRotate          SawyerEncoding = 3
)

// S5Header is the 32-byte file header.
//
// Byte layout (matches OpenLoco S5::Header exactly):
//
//	[0]      uint8  type
//	[1]      uint8  flags
//	[2:4]    uint16 numPackedObjects  (little-endian)
//	[4:8]    uint32 version           (little-endian)
//	[8:12]   uint32 magic             (little-endian)
//	[12:32]  [20]byte padding
type S5Header struct {
	Type             uint8
	Flags            uint8
	NumPackedObjects uint16
	Version          uint32
	Magic            uint32
	Padding          [20]byte
}

// Header flag bits (OpenLoco HeaderFlags enum)
const (
	HeaderFlagIsRaw           uint8 = 1 << 0
	HeaderFlagIsDump          uint8 = 1 << 1
	HeaderFlagIsTitleSequence uint8 = 1 << 2
	HeaderFlagHasSaveDetails  uint8 = 1 << 3
)

// S5Type values
const (
	S5TypeSavedGame uint8 = 0
	S5TypeScenario  uint8 = 1
	S5TypeObjects   uint8 = 2
	S5TypeLandscape uint8 = 3
)

// ParseS5Header reads a 32-byte S5 header from data.
func ParseS5Header(data []byte) (S5Header, error) {
	if len(data) < 32 {
		return S5Header{}, errors.New("data too small for S5 header (need 32 bytes)")
	}
	h := S5Header{
		Type:             data[0],
		Flags:            data[1],
		NumPackedObjects: binary.LittleEndian.Uint16(data[2:4]),
		Version:          binary.LittleEndian.Uint32(data[4:8]),
		Magic:            binary.LittleEndian.Uint32(data[8:12]),
	}
	copy(h.Padding[:], data[12:32])
	return h, nil
}

// ---------------------------------------------------------------------------
// Sawyer chunk stream reader
// ---------------------------------------------------------------------------

// S5ChunkReader reads the sequence of Sawyer-encoded chunks that make up an
// S5 file.  Each chunk on disk is:
//
//	[1 byte]  SawyerEncoding
//	[4 bytes] uint32 LE compressed length
//	[N bytes] compressed payload
//
// The last 4 bytes of the file are a uint32 checksum (byte-sum of everything
// before it) which this reader does not validate — callers may do so
// separately if needed.
type S5ChunkReader struct {
	data   []byte
	offset int
}

// NewS5ChunkReader wraps raw file bytes (including the leading 32-byte header
// that was already parsed).  Pass the full file contents; the reader starts
// immediately after offset 0 (callers should advance past the header first
// by reading the first chunk, which IS the header chunk in rotate encoding).
func NewS5ChunkReader(data []byte) *S5ChunkReader {
	return &S5ChunkReader{data: data, offset: 0}
}

// ReadChunk reads the next chunk from the stream, decodes it according to its
// encoding byte, and returns the decompressed payload.  Returns an error when
// the stream is exhausted or the data is malformed.
func (r *S5ChunkReader) ReadChunk() ([]byte, error) {
	// Need at least 5 bytes for encoding + length
	if r.offset+5 > len(r.data) {
		return nil, fmt.Errorf("S5 stream exhausted at offset %d", r.offset)
	}

	encoding := SawyerEncoding(r.data[r.offset])
	compLen := binary.LittleEndian.Uint32(r.data[r.offset+1 : r.offset+5])
	r.offset += 5

	if r.offset+int(compLen) > len(r.data) {
		return nil, fmt.Errorf("S5 chunk at offset %d claims %d bytes but only %d remain",
			r.offset-5, compLen, len(r.data)-r.offset)
	}

	compressed := r.data[r.offset : r.offset+int(compLen)]
	r.offset += int(compLen)

	return decodeSawyerChunk(encoding, compressed)
}

// Offset returns the current byte position in the stream.
func (r *S5ChunkReader) Offset() int {
	return r.offset
}

// AdvanceRaw skips n raw bytes in the stream without any chunk decoding.
// Used to skip over structs that are written directly (not chunk-framed),
// such as the 16-byte ObjectHeader before each packed object's data chunk.
func (r *S5ChunkReader) AdvanceRaw(n int) {
	r.offset += n
}

// ---------------------------------------------------------------------------
// Sawyer decoding (matches OpenLoco SawyerStreamReader exactly)
// ---------------------------------------------------------------------------

func decodeSawyerChunk(encoding SawyerEncoding, data []byte) ([]byte, error) {
	switch encoding {
	case EncodingUncompressed:
		out := make([]byte, len(data))
		copy(out, data)
		return out, nil

	case EncodingRunLengthSingle:
		return decodeRunLengthSingle(data)

	case EncodingRunLengthMulti:
		// Two-stage: runLengthSingle first, then runLengthMulti on top
		stage1, err := decodeRunLengthSingle(data)
		if err != nil {
			return nil, fmt.Errorf("runLengthMulti stage1: %w", err)
		}
		return decodeRunLengthMulti(stage1)

	case EncodingRotate:
		return decodeRotate(data), nil

	default:
		return nil, fmt.Errorf("unknown Sawyer encoding: %d", encoding)
	}
}

// decodeRunLengthSingle implements OpenLoco's SawyerStreamReader::decodeRunLengthSingle.
//
// Format per the C++ source:
//
//	if byte & 0x80 (high bit set):
//	    copyLen = 257 − byte
//	    repeat next byte copyLen times
//	else:
//	    copyLen = byte + 1
//	    copy next copyLen bytes literally
func decodeRunLengthSingle(data []byte) ([]byte, error) {
	out := make([]byte, 0, len(data)*2)
	i := 0
	for i < len(data) {
		code := data[i]
		i++
		if code&0x80 != 0 {
			// Repeat run
			if i >= len(data) {
				return nil, errors.New("runLengthSingle: truncated repeat run")
			}
			count := 257 - int(code)
			val := data[i]
			i++
			for j := 0; j < count; j++ {
				out = append(out, val)
			}
		} else {
			// Literal run
			count := int(code) + 1
			if i+count > len(data) {
				return nil, errors.New("runLengthSingle: truncated literal run")
			}
			out = append(out, data[i:i+count]...)
			i += count
		}
	}
	return out, nil
}

// decodeRunLengthMulti implements OpenLoco's SawyerStreamReader::decodeRunLengthMulti.
//
// This is applied AFTER decodeRunLengthSingle (two-stage decode).
//
// Format:
//
//	if byte == 0xFF:
//	    emit next byte literally
//	else:
//	    offset = (byte >> 3) − 32   (always negative — back-reference)
//	    copyLen = (byte & 0x07) + 1  (1–8 bytes)
//	    copy copyLen bytes from output[current + offset]
func decodeRunLengthMulti(data []byte) ([]byte, error) {
	out := make([]byte, 0, len(data)*2)
	i := 0
	for i < len(data) {
		code := data[i]
		i++
		if code == 0xFF {
			if i >= len(data) {
				return nil, errors.New("runLengthMulti: truncated literal")
			}
			out = append(out, data[i])
			i++
		} else {
			offset := int(code>>3) - 32 // negative
			copyLen := int(code&0x07) + 1

			srcPos := len(out) + offset
			if srcPos < 0 {
				return nil, fmt.Errorf("runLengthMulti: back-reference offset %d exceeds output length %d", offset, len(out))
			}
			// Copy byte-by-byte (source and dest may overlap in theory,
			// but the C++ code copies to a temp buffer first so they don't)
			for j := 0; j < copyLen; j++ {
				out = append(out, out[srcPos+j])
			}
		}
	}
	return out, nil
}

// decodeRotate implements OpenLoco's SawyerStreamReader::decodeRotate.
//
// Each byte is rotated right by a cycling code: 1, 3, 5, 7, 1, 3, 5, 7, …
func decodeRotate(data []byte) []byte {
	out := make([]byte, len(data))
	code := uint(1)
	for i, b := range data {
		out[i] = rotateRight(b, code)
		code = (code + 2) & 7
	}
	return out
}

func rotateRight(v byte, n uint) byte {
	n &= 7
	return (v >> n) | (v << (8 - n))
}

// ---------------------------------------------------------------------------
// Legacy entry points (kept for compatibility with existing callers)
// ---------------------------------------------------------------------------

// DecompressS5 is the old single-blob decompressor.  It is now incorrect for
// real S5 files (which are chunk streams), but is kept so existing callers
// compile.  Use S5ChunkReader for correct multi-chunk decoding.
//
// Deprecated: use S5ChunkReader.ReadChunk() instead.
func DecompressS5(compressedData []byte, expectedLen uint32) ([]byte, error) {
	log.Println("[S5] DecompressS5: DEPRECATED — called with", len(compressedData), "bytes; use S5ChunkReader instead")
	// Treat the input as a single runLengthSingle chunk for best-effort compat
	out, err := decodeRunLengthSingle(compressedData)
	if err != nil {
		return nil, err
	}
	if uint32(len(out)) > expectedLen {
		out = out[:expectedLen]
	}
	return out, nil
}

// LoadAndDecompressS5 loads an S5 file and returns decompressed data.
//
// Deprecated: use S5ChunkReader for proper chunk-by-chunk access.
func LoadAndDecompressS5(filePath string) ([]byte, error) {
	data, err := LoadDatFile(filePath)
	if err != nil {
		return nil, err
	}
	return DecompressS5Data(data)
}

// DecompressS5Data is a legacy wrapper.
//
// Deprecated: use S5ChunkReader.
func DecompressS5Data(data []byte) ([]byte, error) {
	if len(data) < 32 {
		return nil, errors.New("data too small for S5 header")
	}
	_, err := ParseS5Header(data)
	if err != nil {
		return nil, err
	}
	compressedPayload := data[32:]
	return DecompressS5(compressedPayload, 1699535)
}
