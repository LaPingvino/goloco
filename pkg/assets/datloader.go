package assets

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

// Minimal DAT/G1 loader skeleton.
// This file provides a tiny API to open a .dat file and search for a G1
// header signature to locate an image table. It does not implement full
// G1 parsing yet — just discovery and basic reading utilities used by tests.

// ErrG1NotFound is returned when no G1 header is found in the file.
var ErrG1NotFound = errors.New("G1 header not found")

// G1Header represents the minimal fields we care about from the G1 header.
// The full OpenLoco header has many fields; we only expose what's needed
// for initial discovery: a magic and an element count/offset.
// This struct mirrors the on-disk layout (little endian) where applicable.

// NOTE: This is a placeholder; future work will expand the fields.
type G1Header struct {
	Magic [4]byte
	// other fields omitted
}

// LoadDatFile opens the provided path and returns its contents as a reader.
// Future functions will parse the reader to extract G1 image tables.
func LoadDatFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	sz := fi.Size()
	if sz == 0 {
		return nil, io.EOF
	}

	buf := make([]byte, sz)
	_, err = io.ReadFull(f, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// FindG1Header searches the provided data for a G1 header signature.
// It returns the offset where the header starts or an error if not found.
func FindG1Header(data []byte) (int, error) {
	// In OpenLoco, G1 headers start with 'G1\0\0' or similar signature.
	// For now, search for ASCII "G1" followed by a zero byte.
	magic := []byte{'G', '1', 0}
	for i := 0; i < len(data)-len(magic); i++ {
		if data[i] == magic[0] && data[i+1] == magic[1] && data[i+2] == magic[2] {
			return i, nil
		}
	}
	return 0, ErrG1NotFound
}

// readUint32LE reads a uint32 from data at offset using little-endian.
func readUint32LE(data []byte, offset int) (uint32, error) {
	if offset+4 > len(data) {
		return 0, io.ErrUnexpectedEOF
	}
	return binary.LittleEndian.Uint32(data[offset : offset+4]), nil
}

// readUint16LE reads a uint16 from data at offset using little-endian.
func readUint16LE(data []byte, offset int) (uint16, error) {
	if offset+2 > len(data) {
		return 0, io.ErrUnexpectedEOF
	}
	return binary.LittleEndian.Uint16(data[offset : offset+2]), nil
}
