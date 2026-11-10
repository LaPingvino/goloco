package world

import "testing"

// updateConstructionArrow: the first update after the head is placed shows the
// arrow (frame 0 → 0xFF → 5, toggled on), aimed at the head; it toggles every
// 6th update after; head false switches it off and resets the frame.
func TestConstructionArrowBlinks(t *testing.T) {
	wld := newTestWorld(20, 20, 10)
	wld.SetConstructionHead(10, 10) // rotation 0, baseZ 10
	var seen []bool
	for i := 0; i < 13; i++ {
		wld.UpdateConstructionArrow(true)
		seen = append(seen, wld.consArrow.visible)
	}
	want := []bool{true, true, true, true, true, true, false, false, false, false, false, false, true}
	for i := range want {
		if seen[i] != want[i] {
			t.Fatalf("update %d: visible %v, want %v (all %v)", i, seen[i], want[i], seen)
		}
	}
	if a := wld.consArrow; a.x != 320 || a.y != 320 || a.z != 40 || a.dir != 0 {
		t.Fatalf("arrow at (%d, %d, %d) dir %d, want the head (320, 320, 40) dir 0", a.x, a.y, a.z, a.dir)
	}
	wld.UpdateConstructionArrow(false)
	if wld.consArrow.visible || wld.consArrow.frame != 0 {
		t.Fatalf("off: %+v", wld.consArrow)
	}
}
