package assets

import (
	"encoding/binary"
	"io"
)

// s5write.go — the Sawyer chunk stream writer: SawyerStreamWriter's three
// encoders, the chunk framing and the closing checksum (the byte sum of all
// that was written before it).
//
// OpenLoco reference: src/OpenLoco/src/S5/SawyerStream.cpp SawyerStreamWriter

// S5ChunkWriter writes an S5 file's chunk stream to w.
type S5ChunkWriter struct {
	w        io.Writer
	checksum uint32
	err      error
}

// NewS5ChunkWriter starts a chunk stream on w.
func NewS5ChunkWriter(w io.Writer) *S5ChunkWriter { return &S5ChunkWriter{w: w} }

// Write writes raw bytes into the stream (counted in the checksum), as
// SawyerStreamWriter::write — ObjectManager::writePackedObjects writes each
// packed object's header this way.
func (s *S5ChunkWriter) Write(p []byte) {
	if s.err != nil {
		return
	}
	_, s.err = s.w.Write(p)
	for _, b := range p {
		s.checksum += uint32(b)
	}
}

// WriteChunk is writeChunk: the encoding byte, the encoded length, the
// encoded data.
func (s *S5ChunkWriter) WriteChunk(enc SawyerEncoding, data []byte) {
	out := EncodeSawyerChunk(enc, data)
	var hdr [5]byte
	hdr[0] = byte(enc)
	binary.LittleEndian.PutUint32(hdr[1:], uint32(len(out)))
	s.Write(hdr[:])
	s.Write(out)
}

// Close is writeChecksum; it returns the first write error.
func (s *S5ChunkWriter) Close() error {
	if s.err != nil {
		return s.err
	}
	var c [4]byte
	binary.LittleEndian.PutUint32(c[:], s.checksum)
	_, s.err = s.w.Write(c[:])
	return s.err
}

// EncodeSawyerChunk is SawyerStreamWriter::encode.
func EncodeSawyerChunk(enc SawyerEncoding, data []byte) []byte {
	switch enc {
	case EncodingRunLengthSingle:
		return encodeRunLengthSingle(data)
	case EncodingRunLengthMulti:
		return encodeRunLengthSingle(encodeRunLengthMulti(data))
	case EncodingRotate:
		return encodeRotate(data)
	}
	return append([]byte(nil), data...)
}

// encodeRunLengthSingle is SawyerStreamWriter::encodeRunLengthSingle, step for
// step (its runs and literal blocks decide the file's bytes).
func encodeRunLengthSingle(data []byte) []byte {
	out := make([]byte, 0, len(data)/2+16)
	src, end, normStart := 0, len(data), 0
	count := 0
	for src < end-1 {
		if (count != 0 && data[src] == data[src+1]) || count > 125 {
			out = append(out, byte(count-1))
			out = append(out, data[normStart:normStart+count]...)
			normStart += count
			count = 0
		}
		if data[src] == data[src+1] {
			for ; count < 125 && src+count < end; count++ {
				if data[src] != data[src+count] {
					break
				}
			}
			out = append(out, byte(257-count), data[src])
			src += count
			normStart = src
			count = 0
		} else {
			count++
			src++
		}
	}
	if src == end-1 {
		count++
	}
	if count != 0 {
		out = append(out, byte(count-1))
		out = append(out, data[normStart:normStart+count]...)
	}
	return out
}

// encodeRunLengthMulti is SawyerStreamWriter::encodeRunLengthMulti: a
// literal (0xFF, byte) or a back-reference of 1–8 bytes within the last 32.
func encodeRunLengthMulti(data []byte) []byte {
	n := len(data)
	if n == 0 {
		return nil
	}
	out := make([]byte, 0, n)
	out = append(out, 0xFF, data[0])
	for i := 1; i < n; {
		searchIndex := 0
		if i >= 32 {
			searchIndex = i - 32
		}
		searchEnd := i - 1
		bestIndex, bestCount := 0, 0
		for r := searchIndex; r <= searchEnd; r++ {
			maxCount := min(7, searchEnd-r, n-i-1)
			count := 0
			for j := 0; j <= maxCount; j++ {
				if data[r+j] != data[i+j] {
					break
				}
				count++
			}
			if count > bestCount {
				bestIndex, bestCount = r, count
				if count == 8 {
					break
				}
			}
		}
		if bestCount == 0 {
			out = append(out, 0xFF, data[i])
			i++
		} else {
			out = append(out, byte((bestCount-1)|((32-(i-bestIndex))<<3)))
			i += bestCount
		}
	}
	return out
}

// encodeRotate is SawyerStreamWriter::encodeRotate: each byte rotated left by
// 1, 3, 5, 7, 1, …
func encodeRotate(data []byte) []byte {
	out := make([]byte, len(data))
	code := uint(1)
	for i, b := range data {
		out[i] = b<<code | b>>(8-code)
		code = (code + 2) & 7
	}
	return out
}

// S5Checksum is the byte sum SawyerStreamReader::validateChecksum compares
// with a file's last four bytes.
func S5Checksum(file []byte) (want, got uint32, ok bool) {
	if len(file) < 4 {
		return 0, 0, false
	}
	for _, b := range file[:len(file)-4] {
		got += uint32(b)
	}
	want = binary.LittleEndian.Uint32(file[len(file)-4:])
	return want, got, want == got
}

// Bytes is the header's 32 bytes as S5::Header lays them out.
func (h S5Header) Bytes() []byte {
	b := make([]byte, 32)
	b[0], b[1] = h.Type, h.Flags
	binary.LittleEndian.PutUint16(b[2:], h.NumPackedObjects)
	binary.LittleEndian.PutUint32(b[4:], h.Version)
	binary.LittleEndian.PutUint32(b[8:], h.Magic)
	copy(b[12:], h.Padding[:])
	return b
}
