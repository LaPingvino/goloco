package assets

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// Build a synthetic simple G1 blob with two 2x2 images.
func buildSyntheticG1() []byte {
	// header: magic + count + offset table
	buf := bytes.NewBuffer(nil)
	buf.Write([]byte{'G', '1', 0, 0})
	binary.Write(buf, binary.LittleEndian, uint32(2))
	// placeholder offsets (we'll fill later)
	off1Pos := buf.Len()
	binary.Write(buf, binary.LittleEndian, uint32(0))
	off2Pos := buf.Len()
	binary.Write(buf, binary.LittleEndian, uint32(0))
	// element 1 offset
	elem1Start := buf.Len()
	// elem1: w=2 h=2 pixels: 10,20,30,40
	binary.Write(buf, binary.LittleEndian, uint16(2))
	binary.Write(buf, binary.LittleEndian, uint16(2))
	buf.Write([]byte{10, 20, 30, 40})
	// element 2 offset
	elem2Start := buf.Len()
	binary.Write(buf, binary.LittleEndian, uint16(2))
	binary.Write(buf, binary.LittleEndian, uint16(2))
	buf.Write([]byte{100, 110, 120, 130})
	data := buf.Bytes()
	// patch offsets (relative to header start which is 0)
	binary.LittleEndian.PutUint32(data[off1Pos:off1Pos+4], uint32(elem1Start))
	binary.LittleEndian.PutUint32(data[off2Pos:off2Pos+4], uint32(elem2Start))
	return data
}

func TestParseSimpleG1(t *testing.T) {
	data := buildSyntheticG1()
	imgs, err := ParseSimpleG1(data, 0)
	if err != nil {
		t.Fatalf("ParseSimpleG1 error: %v", err)
	}
	if len(imgs) != 2 {
		t.Fatalf("expected 2 images, got %d", len(imgs))
	}
	// verify first image pixels
	r0 := imgs[0].Bounds()
	if r0.Dx() != 2 || r0.Dy() != 2 {
		t.Fatalf("unexpected dims: %v", r0)
	}
	c0 := imgs[0].At(0, 0)
	r, _, _, a := c0.RGBA()
	// RGBA returns 16-bit values (0..65535), original 8-bit value is high byte
	if byte(r>>8) != 10 || byte(a>>8) != 0xFF {
		t.Fatalf("expected pixel 10 alpha 255, got r=%d a=%d", byte(r>>8), byte(a>>8))
	}
	c1 := imgs[1].At(1, 1)
	r2, _, _, _ := c1.RGBA()
	if byte(r2>>8) != 130 {
		t.Fatalf("expected pixel 130, got %d", byte(r2>>8))
	}
}
