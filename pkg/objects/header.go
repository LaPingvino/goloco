package objects

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

// ObjectHeader is the 16-byte header at the start of each DAT object
type ObjectHeader struct {
	Flags    uint32  // Lower 6 bits = type, bits 6-7 = source game
	Name     [8]byte // Object name (null-padded)
	Checksum uint32  // Object checksum
}

// HeaderSize is the size of ObjectHeader in bytes
const HeaderSize = 16

// ObjectEntry2 extends ObjectHeader with data size
type ObjectEntry2 struct {
	ObjectHeader
	DataSize uint32 // Size of object data after header
}

// Entry2Size is the size of ObjectEntry2 in bytes
const Entry2Size = 20

// ChecksumMagic is used for checksum validation
const ChecksumMagic = 0xF369A75B

// ReadHeader reads an ObjectHeader from a reader
func ReadHeader(r io.Reader) (*ObjectHeader, error) {
	h := &ObjectHeader{}
	if err := binary.Read(r, binary.LittleEndian, h); err != nil {
		return nil, fmt.Errorf("reading object header: %w", err)
	}
	return h, nil
}

// ReadEntry2 reads an ObjectEntry2 from a reader
func ReadEntry2(r io.Reader) (*ObjectEntry2, error) {
	e := &ObjectEntry2{}
	if err := binary.Read(r, binary.LittleEndian, e); err != nil {
		return nil, fmt.Errorf("reading object entry: %w", err)
	}
	return e, nil
}

// GetType returns the ObjectType from the flags
func (h *ObjectHeader) GetType() ObjectType {
	return ObjectType(h.Flags & 0x3F)
}

// GetSourceGame returns the source game from the flags
func (h *ObjectHeader) GetSourceGame() SourceGame {
	return SourceGame((h.Flags >> 6) & 0x3)
}

// GetName returns the object name as a string (trimmed)
func (h *ObjectHeader) GetName() string {
	// Find the null terminator or end of array
	end := len(h.Name)
	for i, b := range h.Name {
		if b == 0 {
			end = i
			break
		}
	}
	return strings.TrimSpace(string(h.Name[:end]))
}

// IsEmpty returns true if the header is empty/invalid
func (h *ObjectHeader) IsEmpty() bool {
	return h.Flags == 0xFFFFFFFF && h.Checksum == 0xFFFFFFFF
}

// IsCustom returns true if this is a custom (non-vanilla) object
func (h *ObjectHeader) IsCustom() bool {
	return h.GetSourceGame() == SourceGameCustom
}

// String returns a string representation of the header
func (h *ObjectHeader) String() string {
	return fmt.Sprintf("Object{Name: %q, Type: %s, Source: %d, Checksum: 0x%08X}",
		h.GetName(), h.GetType(), h.GetSourceGame(), h.Checksum)
}
