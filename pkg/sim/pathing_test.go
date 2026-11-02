package sim

import "testing"

// trackLongestPathingCalculate along a line sums getTrackMiscData weights
// (straight: 32); at a junction trackLongestPathing takes the branch that
// leads further; trackAimlessWanderPathing finds no signals on plain track.
func TestTrackLongestPathing(t *testing.T) {
	s := flatState(40, 64)
	s.ConsHead = ConstructionHead{Active: true, X: 30, Y: 20, BaseZ: 64, Rotation: 0}
	for range 5 {
		if !s.PlaceTrackAtHead(0, 0) {
			t.Fatal("straight can't be built")
		}
	}
	p := Pathing{}
	if got := s.TrackLongestPathingCalculate(30*32, 20*32, 64*4, 0, p); got != 5*32 {
		t.Errorf("line of 5 straights: %d, want %d", got, 5*32)
	}
	// A junction at (25,20): the line continues 4 more straights west; a
	// small left curve (track 4) branches off there and ends at once.
	for range 4 {
		s.PlaceTrackAtHead(0, 0)
	}
	s.ConsHead = ConstructionHead{Active: true, X: 25, Y: 20, BaseZ: 64, Rotation: 0}
	if !s.PlaceTrackAtHead(4, 0) {
		t.Fatal("curve can't be built")
	}
	e, _ := TrackConnectionEnd(26*32, 20*32, 64*4, 0) // the end of the straight at (26,20)
	tc := s.TrackConnections(e.X, e.Y, e.Z, e.NextRotation, 0, 0, 0, 0)
	if len(tc.Connections) != 2 {
		t.Fatalf("junction connections %v, want 2", tc.Connections)
	}
	if got := s.TrackLongestPathing(e.X, e.Y, e.Z, tc.Connections, p); got&TaDBasicMask != 0 {
		t.Errorf("junction: chose %d, want the straight (0)", got&TaDBasicMask)
	}
	if f := s.TrackAimlessWanderPathing(30*32, 20*32, 64*4, 0, p); f != 0 {
		t.Errorf("wander on plain track: flags %#x", f)
	}
}

// trackTargetedPathing at the same junction: aiming at the straight line's
// far piece, the straight branch reaches it (distance 0, no signals) and the
// curve does not; isTrackRoutingResultBetter prefers the straight.
func TestTrackTargetedPathing(t *testing.T) {
	s := flatState(40, 64)
	s.ConsHead = ConstructionHead{Active: true, X: 30, Y: 20, BaseZ: 64, Rotation: 0}
	for range 9 {
		s.PlaceTrackAtHead(0, 0)
	}
	s.ConsHead = ConstructionHead{Active: true, X: 25, Y: 20, BaseZ: 64, Rotation: 0}
	s.PlaceTrackAtHead(4, 0)
	e, _ := TrackConnectionEnd(26*32, 20*32, 64*4, 0)
	tc := s.TrackConnections(e.X, e.Y, e.Z, e.NextRotation, 0, 0, 0, 0)
	target := &RouteTarget{StationID: -1, X: 22 * 32, Y: 20 * 32, Z: 64 * 4, TaD: 0, RX: -1, RY: -1}
	var straight, curve RoutingResult
	for _, c := range tc.Connections {
		r := s.TrackTargetedPathing(e.X, e.Y, e.Z, c&(TaDBasicMask|TaDHasSignal), Pathing{}, target)
		if c&TaDBasicMask == 0 {
			straight = r
		} else {
			curve = r
		}
	}
	if straight.BestDistToTarget != 0 || straight.SignalState != RouteNoSignals {
		t.Errorf("straight branch: %+v, want the target reached with no signals", straight)
	}
	if curve.BestDistToTarget == 0 {
		t.Errorf("curve branch reached the target: %+v", curve)
	}
	if !IsTrackRoutingResultBetter(curve, straight) || IsTrackRoutingResultBetter(straight, curve) {
		t.Error("isTrackRoutingResultBetter does not prefer the reached target")
	}
}
