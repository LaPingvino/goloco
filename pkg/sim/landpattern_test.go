package sim

import "testing"

// The land patterns' marking after fd6db41: near water marks within 25 of
// the tiles at or below sea level (it used to mark around the dry land), and
// around cliffs marks within 6 of a height step of 4 to the next row or
// column (it used to compare a tile with itself and mark nothing).
func TestLandPatternMarks(t *testing.T) {
	const sea = 10
	hm := NewHeightMap(100, 100)
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			h := uint8(20) // dry land …
			if x < 5 {
				h = sea // … a strip of sea along the west edge
			}
			if x >= 70 {
				h = 30 // … and a cliff from 20 to 30 at x 69 → 70
			}
			*hm.At(x, y) = h
		}
	}
	markWhere(hm, 25, func(h uint8) bool { return h <= sea })
	if !hm.IsMarkerSet(29, 50) || hm.IsMarkerSet(30, 50) || hm.IsMarkerSet(60, 50) {
		t.Errorf("near water: marks at x 29 %v, 30 %v, 60 %v; want the 25 tiles beyond the sea only",
			hm.IsMarkerSet(29, 50), hm.IsMarkerSet(30, 50), hm.IsMarkerSet(60, 50))
	}
	markAroundCliffs(hm)
	if !hm.IsMarkerSet(63, 50) || !hm.IsMarkerSet(75, 50) || hm.IsMarkerSet(62, 50) || hm.IsMarkerSet(76, 50) || hm.IsMarkerSet(30, 50) {
		t.Errorf("around cliffs: x 62 %v, 63 %v, 75 %v, 76 %v, 30 %v; want 63–75 marked", hm.IsMarkerSet(62, 50),
			hm.IsMarkerSet(63, 50), hm.IsMarkerSet(75, 50), hm.IsMarkerSet(76, 50), hm.IsMarkerSet(30, 50))
	}
}
