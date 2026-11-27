package scenario

import (
	"bytes"
	"testing"
)

// ResetSurfaceClearance sets a flat surface's clearZ to its baseZ and leaves
// sloped surfaces and other elements alone (TileManager::resetSurfaceClearance):
// a generated map's edge surface, initialise's default (baseZ 4, clearZ 0),
// comes back from a load with clearZ 4.
func TestResetSurfaceClearance(t *testing.T) {
	const track = 1 << elementTypeShift
	in := []byte{
		KindSurface, 0, 4, 0, 0x00, 0, 0, 0, // flat edge surface: clearZ 0 → 4
		KindSurface, 0, 8, 10, 0x03, 0, 0, 0, // sloped: kept
		track, 0, 6, 0, 0x00, 0, 0, 0, // not a surface: kept
	}
	want := append([]byte(nil), in...)
	want[3] = 4
	if got := ResetSurfaceClearance(in); !bytes.Equal(got, want) {
		t.Errorf("got % x, want % x", got, want)
	}
}
