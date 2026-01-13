package assets

import (
	"bytes"
	"testing"
)

func TestFindG1Header_NotFound(t *testing.T) {
	data := []byte("this has no header")
	_, err := FindG1Header(data)
	if err == nil {
		t.Fatal("expected error when G1 header not found")
	}
}

func TestFindG1Header_Found(t *testing.T) {
	// synthetic data containing 'G1\0' in the middle
	buf := bytes.Repeat([]byte{0xAA}, 10)
	buf = append(buf, []byte{'G', '1', 0, 0x10, 0x20}...)
	buf = append(buf, bytes.Repeat([]byte{0xBB}, 20)...)
	off, err := FindG1Header(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if off != 10 {
		t.Fatalf("expected offset 10, got %d", off)
	}
}
