package sim

import "testing"

// trackLongestPathingCalculate along a line sums getTrackMiscData weights
// (straight: 32); at a junction trackLongestPathing takes the branch that
// leads further; trackAimlessWanderPathing finds no signals on plain track.
func TestTrackLongestPathing(t *testing.T) {
	s := flatState(40, 64)
	h := headAt(s, 30, 20, 64, 0)
	for range 5 {
		if !h.track(0) {
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
		h.track(0)
	}
	h = headAt(s, 25, 20, 64, 0)
	if !h.track(4) {
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
	h := headAt(s, 30, 20, 64, 0)
	for range 9 {
		h.track(0)
	}
	h = headAt(s, 25, 20, 64, 0)
	h.track(4)
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

// trackPathing with a RouteWaypoint order (encoded as upstream's
// setWaypoint/setDirection/setTrackId do) at the straight line's far piece
// picks the straight connection; aimed at the junction's own straight
// connection it is a direct hit (noSignals, distance 0).
func TestTrackPathingWaypoint(t *testing.T) {
	s := flatState(40, 64)
	h := headAt(s, 30, 20, 64, 0)
	for range 9 {
		h.track(0)
	}
	h = headAt(s, 25, 20, 64, 0)
	h.track(4)
	e, _ := TrackConnectionEnd(26*32, 20*32, 64*4, 0)
	tc := s.TrackConnections(e.X, e.Y, e.Z, e.NextRotation, 0, 0, 0, 0)
	waypoint := func(tx, ty int, z uint8, dir, track uint8) []byte {
		o := make([]byte, 7) // RouteWaypoint (6 bytes) then End
		o[0] = OrderRouteWaypoint | uint8((tx&0x100)>>1)
		o[1] = uint8(tx)
		o[2] = z/2 | uint8((ty&0x100)>>1) // baseZ is stored as z/8 world units: (z*4)/8
		o[3] = uint8(ty)
		o[4] = dir&7 | (track&0x1F)<<3
		o[5] = (track >> 5) & 1
		return o
	}
	s.Orders = waypoint(22, 20, 64, 0, 0)
	var st PathingState
	if got := s.TrackPathing(0, 0, e.X, e.Y, e.Z, tc.Connections, Pathing{}, false, &st); got&TaDBasicMask != 0 || !st.HadNewResult {
		t.Errorf("waypoint far down the straight: chose %d (new result %v)", got&TaDBasicMask, st.HadNewResult)
	}
	s.Orders = waypoint(int(e.X/32), int(e.Y/32), uint8(e.Z/4), 0, 0)
	st = PathingState{}
	if got := s.TrackPathing(0, 0, e.X, e.Y, e.Z, tc.Connections, Pathing{}, false, &st); got&TaDBasicMask != 0 || st.Result.SignalState != RouteNoSignals || st.Result.BestDistToTarget != 0 {
		t.Errorf("waypoint at the junction: chose %d, result %+v", got&TaDBasicMask, st.Result)
	}
}

// bringTrackElementToFront: at the junction tile (straight and curve both
// start there), the chosen piece moves to the front of the tile's track run.
func TestBringTrackElementToFront(t *testing.T) {
	s := flatState(40, 64)
	h := headAt(s, 25, 20, 64, 0)
	h.track(0)
	h = headAt(s, 25, 20, 64, 0)
	h.track(4)
	tl := s.TileAt(25, 20)
	first := func() uint8 {
		for _, r := range tl.Order {
			if r.Kind == 1 {
				return tl.Tracks[r.Index].TrackID
			}
		}
		return 0xFF
	}
	if first() != 0 {
		t.Fatalf("before: first track %d, want the straight (built first)", first())
	}
	s.BringTrackElementToFront(25*32, 20*32, 64*4, 0, 4<<3)
	if first() != 4 {
		t.Errorf("after bringing the curve forward: first track %d, want 4", first())
	}
	s.BringTrackElementToFront(25*32, 20*32, 64*4, 0, 0)
	if first() != 0 {
		t.Errorf("after bringing the straight back: first track %d, want 0", first())
	}
}
