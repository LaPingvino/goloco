package sim

import "testing"

// speedFromDistanceInATick16 is upstream e71abfb's toSpeed16(
// speedFromDistanceInATick(var_3C)) — a signed division of twice the
// distance — not vanilla's int16(uint32(var_3C) >> 15), which parts from it
// for a negative distance.
func TestSpeedFromDistanceInATick16(t *testing.T) {
	for _, c := range []struct {
		d    int32
		want int16
	}{{0x123456, 36}, {0, 0}, {-1, 0}, {-70000, -2}, {-0x10000, -2}} {
		if got := speedFromDistanceInATick16(c.d); got != c.want {
			t.Errorf("distance %d: %d, want %d (vanilla: %d)", c.d, got, c.want, int16(uint32(c.d)>>15))
		}
	}
}
