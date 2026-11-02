package sim

import "testing"

// Along straight track (piece 0): each axis step costs 0x220C, so a budget
// of n·0x220C (plus the 0x368A threshold margin) advances the expected number
// of sub-positions; at the piece's end with no next piece the component stops
// with remainingDistance 0x3689 and reports the rest.
func TestUpdateMotionStraight(t *testing.T) {
	p := TrackSubPosition(0)
	c := MotionComponent{TileX: 32 * 10, TileY: 32 * 20, TileBaseZ: 4}
	c.X, c.Y, c.Z = int32(p[0].X)+c.TileX, int32(p[0].Y)+c.TileY, int32(p[0].Z)+16
	r := c.UpdateMotion(0x368A, nil) // exactly one step's threshold
	if r.Blocked || c.SubPosition != 1 {
		t.Fatalf("one step: sub %d blocked %v", c.SubPosition, r.Blocked)
	}
	if want := int32(0x368A - 0x220C); c.RemainingDistance != want {
		t.Errorf("remaining %#x, want %#x", c.RemainingDistance, want)
	}
	if c.X != int32(p[1].X)+c.TileX || c.Y != int32(p[1].Y)+c.TileY {
		t.Errorf("position (%d,%d), want sub-position 1", c.X, c.Y)
	}
	r = c.UpdateMotion(0x7FFFFFF, nil)
	if !r.Blocked || int(c.SubPosition) != len(p)-1 || c.RemainingDistance != 0x3689 {
		t.Errorf("end: blocked %v sub %d/%d remaining %#x", r.Blocked, c.SubPosition, len(p)-1, c.RemainingDistance)
	}
	// The next-piece hook continues onto sub-position 0 of the new piece (the
	// stop left 0x3689, so one more unit is exactly one step).
	calls := 0
	c.UpdateMotion(1, func(c *MotionComponent) bool { calls++; c.TileX += 32; return true })
	if calls != 1 || c.SubPosition != 0 {
		t.Errorf("next piece: %d calls, sub %d", calls, c.SubPosition)
	}
}
