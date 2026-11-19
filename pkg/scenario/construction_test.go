package scenario

import "testing"

// Scenario::Construction sits at GameState 0x15A: seven lists of eight in
// the S5GameState order (signals, bridges, trainStations, trackMods,
// var_17A, roadStations, roadMods), 0x38 bytes, nothing else touched.
func TestConstructionLayout(t *testing.T) {
	gs := make([]byte, GameStateSize)
	for i := range gs {
		gs[i] = 0xAA
	}
	for i := range 0x38 {
		gs[ConstructionOffset+i] = byte(i)
	}
	var c Construction
	c.Decode(gs)
	if c.Signals[0] != 0 || c.Bridges[0] != 8 || c.TrainStations[7] != 23 || c.TrackMods[0] != 24 ||
		c.Var17A[0] != 32 || c.RoadStations[0] != 40 || c.RoadMods[7] != 55 {
		t.Fatalf("decoded %+v", c)
	}
	out := make([]byte, GameStateSize)
	for i := range out {
		out[i] = 0xAA
	}
	c.Encode(out)
	for i := range gs {
		if gs[i] != out[i] {
			t.Fatalf("byte 0x%X: %d, want %d", i, out[i], gs[i])
		}
	}
	c.ResetTrackObjects()
	if c.Signals[3] != 0xFF || c.Bridges[3] != 0xFF || c.TrainStations[3] != 0xFF || c.TrackMods[3] != 0xFF || c.RoadStations[3] != 43 {
		t.Fatalf("resetTrackObjects: %+v", c)
	}
	c.ResetRoadObjects()
	if c.Var17A[0] != 0xFF || c.RoadStations[3] != 0xFF || c.RoadMods[7] != 0xFF {
		t.Fatalf("resetRoadObjects: %+v", c)
	}
}
