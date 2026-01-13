package assets

import (
	"errors"
)

// decodeRLEElement decodes RLE-compressed element data into a width*height
// buffer of palette indices. The input `src` is the element data which for
// RLE contains a line offset table followed by run chunks as used by
// OpenLoco's drawRLESprite. We assume srcX/srcY are 0 for our use (no cropping).
func decodeRLEElement(src []byte, width, height int) ([]byte, error) {
	if len(src) < height*2 {
		return nil, errors.New("rle src too small for line offsets")
	}
	out := make([]byte, width*height)
	// For each line y, read the uint16 offset into src pointing to runs
	for y := 0; y < height; y++ {
		off := int(src[y*2]) | int(src[y*2+1])<<8
		if off < 0 || off >= len(src) {
			return nil, errors.New("line offset out of range")
		}
		p := off
		isEnd := false
		for !isEnd {
			if p+2 > len(src) {
				return nil, errors.New("rle chunk header truncated")
			}
			dataSize := int(src[p])
			firstX := int(src[p+1])
			p += 2
			isEnd = (dataSize & 0x80) != 0
			dataSize &= 0x7F
			// ensure chunk data present
			if p+dataSize > len(src) {
				return nil, errors.New("rle chunk data truncated")
			}
			// copy dataSize bytes starting at p to output at (firstX, y)
			dstX := firstX
			for i := 0; i < dataSize; i++ {
				if dstX+i >= width {
					break
				}
				out[y*width+dstX+i] = src[p+i]
			}
			p += dataSize
		}
	}
	return out, nil
}
