package assets

import (
	"encoding/binary"
	"errors"
	"io"
)

// S5File represents a Sawyer-compressed file (S5 format used by Locomotion)
// Header format:
// [0:1]   uint16 LE: signature (0x2003 for S5)
// [2:4]   uint16 LE: chunk type (usually 0x0000)
// [4:8]   uint32 LE: uncompressed length
// [8:12]  uint32 LE: compressed length
// [12:]   compressed data (Run-Length Encoded with a specific format)

type S5Header struct {
	Signature       uint16
	ChunkType       uint16
	UncompressedLen uint32
	CompressedLen   uint32
}

// ParseS5Header reads the S5 file header from data
func ParseS5Header(data []byte) (S5Header, error) {
	if len(data) < 12 {
		return S5Header{}, errors.New("data too small for S5 header")
	}
	h := S5Header{
		Signature:       binary.LittleEndian.Uint16(data[0:2]),
		ChunkType:       binary.LittleEndian.Uint16(data[2:4]),
		UncompressedLen: binary.LittleEndian.Uint32(data[4:8]),
		CompressedLen:   binary.LittleEndian.Uint32(data[8:12]),
	}
	if h.Signature != 0x2003 {
		return h, errors.New("invalid S5 signature")
	}
	return h, nil
}

// DecompressS5 decompresses Sawyer-compressed data (S5 format)
// The compression uses a simple RLE scheme with optional copy operations
func DecompressS5(compressedData []byte, expectedLen uint32) ([]byte, error) {
	out := make([]byte, 0, expectedLen)
	r := 0

	for r < len(compressedData) && uint32(len(out)) < expectedLen {
		code := compressedData[r]
		r++

		// Code byte format in Sawyer compression:
		// 0x00-0x7F: Run of (code+1) zeros
		// 0x80-0xFF: Copy (256 - code) bytes literally, then a zero

		if code < 0x80 {
			// code 0-127: repeat zeros (code+1) times
			count := int(code) + 1
			for i := 0; i < count && uint32(len(out)) < expectedLen; i++ {
				out = append(out, 0)
			}
		} else {
			// code 128-255: copy (256-code) literal bytes, then add a zero
			count := 256 - int(code)
			for i := 0; i < count && r < len(compressedData) && uint32(len(out)) < expectedLen; i++ {
				out = append(out, compressedData[r])
				r++
			}
			// After literal bytes, add a terminating zero
			if uint32(len(out)) < expectedLen {
				out = append(out, 0)
			}
		}
	}

	if uint32(len(out)) < expectedLen {
		return out, errors.New("decompression ended before expected length")
	}
	return out[:expectedLen], nil
}

// LoadAndDecompressS5 loads an S5 file and returns the decompressed data
func LoadAndDecompressS5(filePath string) ([]byte, error) {
	data, err := LoadDatFile(filePath)
	if err != nil {
		return nil, err
	}
	return DecompressS5Data(data)
}

// DecompressS5Data decompresses S5-formatted data in memory
func DecompressS5Data(data []byte) ([]byte, error) {
	if len(data) < 12 {
		return nil, errors.New("data too small for S5 header")
	}

	header, err := ParseS5Header(data)
	if err != nil {
		return nil, err
	}

	compressedPayload := data[12:]
	if len(compressedPayload) < int(header.CompressedLen) {
		return nil, errors.New("compressed data truncated")
	}

	return DecompressS5(compressedPayload[:header.CompressedLen], header.UncompressedLen)
}

// S5ChunkIterator allows iterating over chunks in an S5 file
type S5ChunkIterator struct {
	data   []byte
	offset int
}

// NewS5ChunkIterator creates an iterator for S5 chunks
func NewS5ChunkIterator(data []byte) *S5ChunkIterator {
	return &S5ChunkIterator{data: data, offset: 0}
}

// Next returns the next decompressed chunk or io.EOF
func (it *S5ChunkIterator) Next() ([]byte, error) {
	if it.offset >= len(it.data) {
		return nil, io.EOF
	}

	if it.offset+12 > len(it.data) {
		return nil, errors.New("truncated S5 chunk header")
	}

	header, err := ParseS5Header(it.data[it.offset:])
	if err != nil {
		return nil, err
	}

	chunkEnd := it.offset + 12 + int(header.CompressedLen)
	if chunkEnd > len(it.data) {
		return nil, errors.New("compressed chunk data truncated")
	}

	compressedPayload := it.data[it.offset+12 : chunkEnd]
	decompressed, err := DecompressS5(compressedPayload, header.UncompressedLen)
	if err != nil {
		return nil, err
	}

	it.offset = chunkEnd
	return decompressed, nil
}

// Offset returns the current byte offset in the file
func (it *S5ChunkIterator) Offset() int {
	return it.offset
}
