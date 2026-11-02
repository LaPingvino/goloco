package paint

import "testing"

// The generated addition tables cover every track piece's sequence, like the
// track pieces themselves (TrackParts).
func TestTrackAdditionTablesShape(t *testing.T) {
	if len(TrackAdditionParts0) != len(TrackParts) || len(TrackAdditionParts1) != len(TrackParts) {
		t.Fatalf("track ids: %d / %d / %d", len(TrackAdditionParts0), len(TrackAdditionParts1), len(TrackParts))
	}
	for id := range TrackParts {
		if len(TrackAdditionParts0[id]) != len(TrackParts[id]) || len(TrackAdditionParts1[id]) != len(TrackParts[id]) {
			t.Errorf("trackId %d: %d / %d additions for %d pieces", id, len(TrackAdditionParts0[id]), len(TrackAdditionParts1[id]), len(TrackParts[id]))
		}
	}
}
