package objects

import (
	"encoding/binary"
	"fmt"
	"io"
)

// DecodeSawyerChunk reads and decompresses a Sawyer-encoded chunk
func DecodeSawyerChunk(r io.Reader) ([]byte, error) {
	// Read encoding type
	var encoding SawyerEncoding
	if err := binary.Read(r, binary.LittleEndian, &encoding); err != nil {
		return nil, fmt.Errorf("reading encoding: %w", err)
	}

	// Read compressed length
	var length uint32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return nil, fmt.Errorf("reading length: %w", err)
	}

	// Read compressed data
	compressed := make([]byte, length)
	if _, err := io.ReadFull(r, compressed); err != nil {
		return nil, fmt.Errorf("reading compressed data: %w", err)
	}

	// Decode based on encoding type
	return decodeSawyer(encoding, compressed)
}

func decodeSawyer(encoding SawyerEncoding, data []byte) ([]byte, error) {
	switch encoding {
	case SawyerEncodingUncompressed:
		return data, nil
	case SawyerEncodingRunLengthSingle:
		return decodeRunLengthSingle(data)
	case SawyerEncodingRunLengthMulti:
		// First pass: single RLE
		pass1, err := decodeRunLengthSingle(data)
		if err != nil {
			return nil, err
		}
		// Second pass: multi RLE
		return decodeRunLengthMulti(pass1)
	case SawyerEncodingRotate:
		return decodeRotate(data)
	default:
		return nil, fmt.Errorf("unknown encoding: %d", encoding)
	}
}

func decodeRunLengthSingle(data []byte) ([]byte, error) {
	result := make([]byte, 0, len(data)*2)

	for i := 0; i < len(data); {
		rleCode := data[i]
		i++

		if rleCode&0x80 != 0 {
			// Repeat next byte (257 - rleCode) times
			if i >= len(data) {
				return nil, fmt.Errorf("invalid RLE: unexpected end")
			}
			copyLen := 257 - int(rleCode)
			copyByte := data[i]
			i++
			for j := 0; j < copyLen; j++ {
				result = append(result, copyByte)
			}
		} else {
			// Copy next (rleCode + 1) bytes directly
			copyLen := int(rleCode) + 1
			if i+copyLen > len(data) {
				return nil, fmt.Errorf("invalid RLE: not enough data")
			}
			result = append(result, data[i:i+copyLen]...)
			i += copyLen
		}
	}

	return result, nil
}

func decodeRunLengthMulti(data []byte) ([]byte, error) {
	result := make([]byte, 0, len(data)*2)

	for i := 0; i < len(data); {
		rleCode := data[i]
		i++

		if rleCode&0x80 != 0 {
			// Repeat pattern
			if i >= len(data) {
				return nil, fmt.Errorf("invalid multi-RLE: unexpected end")
			}

			copyLen := 257 - int(rleCode)
			if rleCode == 0xFF {
				// Special case: copy from previous output
				if i >= len(data) {
					return nil, fmt.Errorf("invalid multi-RLE: no offset")
				}
				offset := int(data[i])
				i++
				srcStart := len(result) - offset - 1
				if srcStart < 0 {
					srcStart = 0
				}
				for j := 0; j < copyLen && srcStart+j < len(result); j++ {
					result = append(result, result[srcStart+j])
				}
			} else {
				copyByte := data[i]
				i++
				for j := 0; j < copyLen; j++ {
					result = append(result, copyByte)
				}
			}
		} else {
			// Copy next (rleCode + 1) bytes directly
			copyLen := int(rleCode) + 1
			if i+copyLen > len(data) {
				return nil, fmt.Errorf("invalid multi-RLE: not enough data")
			}
			result = append(result, data[i:i+copyLen]...)
			i += copyLen
		}
	}

	return result, nil
}

func decodeRotate(data []byte) ([]byte, error) {
	result := make([]byte, len(data))
	var srcBit byte = 1

	for i := 0; i < len(data); i++ {
		result[i] = rotateRight(data[i], srcBit)
		srcBit = (srcBit + 2) & 7
	}

	return result, nil
}

func rotateRight(b byte, count byte) byte {
	count &= 7
	return (b >> count) | (b << (8 - count))
}
