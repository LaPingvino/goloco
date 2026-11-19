package cursor

import (
	"image/color"
	"testing"
)

// TestTable: initialiseCursors' split (pointer and handPointer the system's,
// every other id a bitmap) and Cursor.cpp's encoding, checked on upArrow
// (hot spot 16,0; its tip a black pixel, white inside below it).
func TestTable(t *testing.T) {
	for id := Id(0); id < Count; id++ {
		sys := id == Pointer || id == HandPointer
		if (id.System() != "") != sys {
			t.Errorf("cursor %d: system %q", id, id.System())
		}
		if (id.Image() == nil) != sys {
			t.Errorf("cursor %d: image %v", id, id.Image() != nil)
		}
	}
	if x, y := UpArrow.Hot(); x != 16 || y != 0 {
		t.Errorf("upArrow hot spot %d,%d; want 16,0", x, y)
	}
	img := UpArrow.Image()
	if c := img.RGBAAt(16, 0); c != (color.RGBA{0, 0, 0, 255}) {
		t.Errorf("upArrow 16,0 = %v; want black", c)
	}
	if c := img.RGBAAt(16, 1); c != (color.RGBA{255, 255, 255, 255}) {
		t.Errorf("upArrow 16,1 = %v; want white", c)
	}
	if c := img.RGBAAt(0, 0); c.A != 0 {
		t.Errorf("upArrow 0,0 = %v; want transparent", c)
	}
	if LandTool != 18 || PlaceStation != 44 || Count != 45 {
		t.Errorf("enum: landTool %d, placeStation %d, count %d", LandTool, PlaceStation, Count)
	}
}
