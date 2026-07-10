package assets

import (
	"encoding/binary"
	"errors"
	"image"
	"image/color"
)

// ParseSimpleG1 parses a minimal G1-like structure from data at the provided
// offset. This is a first-pass, testable parser used to iterate on G1
// extraction. The expected format (synthetic for tests) is:
// [0..3] 4-byte magic: 'G','1',0,0
// [4..7] uint32 LE: elementCount
// [8..]   elementCount * uint32 LE offsets (relative to header start)
// each element at (headerStart + offset):
// [0..1] uint16 LE width
// [2..3] uint16 LE height
// [4..]  width*height bytes: palette indices (0-255)
// The parser maps palette index -> grayscale color (R=G=B=index, A=255).

func ParseSimpleG1(data []byte, headerOffset int) ([]*image.RGBA, error) {
	if headerOffset+8 > len(data) {
		return nil, errors.New("data too small for G1 header")
	}
	// check magic
	if !(data[headerOffset] == 'G' && data[headerOffset+1] == '1' && data[headerOffset+2] == 0 && data[headerOffset+3] == 0) {
		return nil, errors.New("invalid G1 magic")
	}
	count := int(binary.LittleEndian.Uint32(data[headerOffset+4 : headerOffset+8]))
	if count <= 0 {
		return nil, errors.New("invalid element count")
	}
	// read offsets
	offTableStart := headerOffset + 8
	if offTableStart+4*count > len(data) {
		return nil, errors.New("data too small for offsets")
	}
	offsets := make([]int, count)
	for i := 0; i < count; i++ {
		off := int(binary.LittleEndian.Uint32(data[offTableStart+i*4 : offTableStart+i*4+4]))
		// off is relative to headerStart
		if off < 0 {
			return nil, errors.New("negative offset")
		}
		offsets[i] = off
	}
	images := make([]*image.RGBA, 0, count)
	for i, rel := range offsets {
		elemStart := headerOffset + rel
		var elemEnd int
		if i+1 < len(offsets) {
			elemEnd = headerOffset + offsets[i+1]
		} else {
			elemEnd = len(data)
		}
		if elemStart+4 > len(data) {
			return nil, errors.New("element truncated")
		}
		if elemEnd > len(data) || elemEnd < elemStart+4 {
			return nil, errors.New("element offsets out of range")
		}
		w := int(binary.LittleEndian.Uint16(data[elemStart : elemStart+2]))
		h := int(binary.LittleEndian.Uint16(data[elemStart+2 : elemStart+4]))
		if w <= 0 || h <= 0 {
			return nil, errors.New("invalid element dims")
		}
		// element payload follows at elemStart+4 up to elemEnd
		payload := data[elemStart+4 : elemEnd]
		// prefer RLE decode if payload looks like it contains line offsets
		var pix []byte
		if len(payload) >= h*2 {
			// check that the uint16 offsets point inside payload
			valid := true
			for y := 0; y < h; y++ {
				off := int(payload[y*2]) | int(payload[y*2+1])<<8
				if off < 0 || off >= len(payload) {
					valid = false
					break
				}
			}
			if valid {
				if decoded, err := decodeRLEElement(payload, w, h); err == nil {
					pix = decoded
				}
			}
		}
		// fallback: if pix still nil and payload is at least w*h, treat as raw pixels
		if pix == nil && len(payload) >= w*h {
			pix = payload[:w*h]
		}
		if pix == nil {
			return nil, errors.New("could not decode element pixels")
		}
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		// map palette index to grayscale
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				idx := pix[y*w+x]
				c := color.RGBA{idx, idx, idx, 0xFF}
				img.SetRGBA(x, y, c)
			}
		}
		images = append(images, img)
	}
	return images, nil
}
