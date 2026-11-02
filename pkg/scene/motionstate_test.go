package scene

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// The saved vehicles' motion state agrees with the ported tables: every rail
// and road component that moves along the track (head, vehicle 1 and 2,
// bogies, tail — not bodies, which are placed from their bogies) has its
// current routing equal to its trackAndDirection, and stands exactly at
// tile + getTrackSubPositon / getRoadSubPositon[subPosition]. This checks the
// entity offsets, the routing table's GameState offset and the generated
// SubpositionData together.
func TestSavedVehicleMotionState(t *testing.T) {
	_, m, _ := titleView(t)
	n := 0
	for _, e := range m.Entities {
		if e.EntityType > scenario.VehicleTypeTail || e.EntityType == scenario.VehicleTypeBodyStart || e.EntityType == scenario.VehicleTypeBodyContinued || e.Mode > 1 {
			continue
		}
		n++
		if r := m.Routing(sim.RoutingHandle(e.RoutingHandle)); r&sim.TaDBasicMask != e.TrackAndDirection&sim.TaDBasicMask {
			t.Errorf("entity %d (type %d): routing %#x, trackAndDirection %#x", e.ID, e.EntityType, r, e.TrackAndDirection)
		}
		var p []sim.MoveInfo
		if e.Mode == 0 {
			p = sim.TrackSubPosition(e.TrackAndDirection & sim.TaDBasicMask)
		} else {
			p = sim.RoadSubPosition(e.TrackAndDirection, m.TrafficHandedness != 0)
		}
		if int(e.SubPosition) >= len(p) {
			t.Errorf("entity %d: sub-position %d of %d", e.ID, e.SubPosition, len(p))
			continue
		}
		mi := p[e.SubPosition]
		if int32(mi.X)+int32(e.TileX) != int32(e.PosX) || int32(mi.Y)+int32(e.TileY) != int32(e.PosY) || int32(mi.Z)+int32(e.TileBaseZ)*4 != int32(e.PosZ) {
			t.Errorf("entity %d (type %d): at (%d,%d,%d), the path says %+v from tile (%d,%d,%d)", e.ID, e.EntityType, e.PosX, e.PosY, e.PosZ, mi, e.TileX, e.TileY, e.TileBaseZ)
		}
	}
	if n == 0 {
		t.Fatal("no vehicles")
	}
}

// Driving each saved rail component forward with updateTrackMotion and
// updateTrackMotionNewTrackPiece follows its reserved routings over real
// track connections: no noRouteFound, it stops (unk_m00) where its
// reservation ends, and it ends on its routing, exactly on its path.
func TestSavedTrainsDriveTheirRoutings(t *testing.T) {
	_, m, _ := titleView(t)
	mods := map[uint16]uint8{}
	for _, e := range m.Entities {
		if e.EntityType == scenario.VehicleTypeHead {
			mods[e.ID] = e.HeadVar53
		}
	}
	driven, pieces := 0, 0
	for _, e := range m.Entities {
		if e.EntityType > scenario.VehicleTypeTail || e.EntityType == scenario.VehicleTypeBodyStart || e.EntityType == scenario.VehicleTypeBodyContinued || e.Mode != 0 {
			continue
		}
		c := sim.MotionComponent{
			TrackAndDirection: e.TrackAndDirection & sim.TaDBasicMask, SubPosition: e.SubPosition,
			TileX: int32(e.TileX), TileY: int32(e.TileY), TileBaseZ: e.TileBaseZ,
			RemainingDistance: e.RemainingDistance, X: int32(e.PosX), Y: int32(e.PosY), Z: int32(e.PosZ),
			RoutingHandle: sim.RoutingHandle(e.RoutingHandle), Owner: e.Owner, TrackType: e.TrackType,
		}
		start := c.RoutingHandle
		r := c.UpdateMotion(0x7FFFFFF, func(c *sim.MotionComponent, f *uint32) bool {
			return m.TrackNewPiece(c, f, mods[e.HeadID], false, 0)
		})
		driven++
		pieces += int(c.RoutingHandle.Index()-start.Index()) & 63
		if r.Flags&sim.MotionNoRouteFound != 0 || !r.Blocked() {
			t.Errorf("entity %d: flags %#x", e.ID, r.Flags)
		}
		if got := m.Routing(c.RoutingHandle) & sim.TaDBasicMask; got != c.TrackAndDirection {
			t.Errorf("entity %d: on %d but its routing is %d", e.ID, c.TrackAndDirection, got)
		}
		p := sim.TrackSubPosition(c.TrackAndDirection)
		if mi := p[c.SubPosition]; int32(mi.X)+c.TileX != c.X || int32(mi.Y)+c.TileY != c.Y || int32(mi.Z)+int32(c.TileBaseZ)*4 != c.Z {
			t.Errorf("entity %d: off its path", e.ID)
		}
	}
	t.Logf("%d rail components drove %d reserved pieces in all", driven, pieces)
	if pieces == 0 {
		t.Error("no component moved onto another piece")
	}
}

// The same for the saved road vehicles: updateRoadMotion with
// updateRoadMotionNewRoadPiece follows each component's reserved routings
// over real road connections (getRoadConnections).
func TestSavedRoadVehiclesDriveTheirRoutings(t *testing.T) {
	_, m, _ := titleView(t)
	mods := map[uint16]uint8{}
	for _, e := range m.Entities {
		if e.EntityType == scenario.VehicleTypeHead {
			mods[e.ID] = e.HeadVar53
		}
	}
	driven, pieces := 0, 0
	for _, e := range m.Entities {
		if e.EntityType > scenario.VehicleTypeTail || e.EntityType == scenario.VehicleTypeBodyStart || e.EntityType == scenario.VehicleTypeBodyContinued || e.Mode != 1 {
			continue
		}
		c := sim.MotionComponent{
			Road: true, RightHand: m.TrafficHandedness != 0,
			TrackAndDirection: e.TrackAndDirection, SubPosition: e.SubPosition,
			TileX: int32(e.TileX), TileY: int32(e.TileY), TileBaseZ: e.TileBaseZ,
			RemainingDistance: e.RemainingDistance, X: int32(e.PosX), Y: int32(e.PosY), Z: int32(e.PosZ),
			RoutingHandle: sim.RoutingHandle(e.RoutingHandle), Owner: e.Owner, TrackType: e.TrackType,
			IsVehicle2: e.EntityType == scenario.VehicleTypeVeh2,
		}
		start := c.RoutingHandle
		r := c.UpdateMotion(0x7FFFFFF, func(c *sim.MotionComponent, f *uint32) bool {
			return m.RoadNewPiece(c, f, mods[e.HeadID], false, 0)
		})
		driven++
		pieces += int(c.RoutingHandle.Index()-start.Index()) & 63
		if r.Flags&sim.MotionNoRouteFound != 0 || !r.Blocked() {
			t.Errorf("entity %d: flags %#x", e.ID, r.Flags)
		}
		if got := m.Routing(c.RoutingHandle) & 0x1FF; got != c.TrackAndDirection&0x1FF {
			t.Errorf("entity %d: on %#x but its routing is %#x", e.ID, c.TrackAndDirection, got)
		}
		p := sim.RoadSubPosition(c.TrackAndDirection, c.RightHand)
		if mi := p[c.SubPosition]; int32(mi.X)+c.TileX != c.X || int32(mi.Y)+c.TileY != c.Y || int32(mi.Z)+int32(c.TileBaseZ)*4 != c.Z {
			t.Errorf("entity %d: off its path", e.ID)
		}
	}
	t.Logf("%d road components drove %d reserved pieces in all", driven, pieces)
	if pieces == 0 {
		t.Error("no component moved onto another piece")
	}
}

// Every saved head builds upstream's Vehicle structure (head, Vehicle1,
// Vehicle2, cars of bogie/bogie/body components, tail); every car begins
// with a body_start; the trains' Vehicle2 speeds include the 60 mph the
// Vehicles window shows for train 2.
func TestSavedTrainStructure(t *testing.T) {
	_, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	rail, road, cars, at60 := 0, 0, 0, 0
	for _, tr := range trains {
		switch m.Entities[tr.Head].Mode {
		case 0:
			rail++
		case 1:
			road++
		}
		for _, c := range tr.Cars {
			cars++
			if m.Entities[c[0].Body].EntityType != scenario.VehicleTypeBodyStart {
				t.Errorf("head %d: a car starts with body type %d", m.Entities[tr.Head].ID, m.Entities[c[0].Body].EntityType)
			}
		}
		if sp := m.Entities[tr.Veh2].CurrentSpeed; sp>>16 == 60 && m.Entities[tr.Head].Mode == 0 {
			at60++
		}
	}
	t.Logf("%d vehicles (%d rail, %d road, the rest air/water), %d cars; rail at 60 mph: %d", len(trains), rail, road, cars, at60)
	if rail != 14 || road != 83 {
		t.Errorf("want 14 rail and 83 road vehicles")
	}
	if at60 == 0 {
		t.Error("no train at 60 mph (Vehicle2 currentSpeed offset?)")
	}
}

// findAllTracksFilterTransform from each saved train's head piece reaches a
// set of real pieces (each forward result is a built piece origin at that
// height; reversed results end there), and searching again from any result
// reaches the same network: the walk is a closed connected component.
func TestTrackNetworkSearch(t *testing.T) {
	_, m, _ := titleView(t)
	search := func(l sim.LocationOfInterest, flags uint16) *sim.RoutingResults {
		res := &sim.RoutingResults{Reachable: sim.NewInterestSet(sim.TrackModHashSetSize)}
		m.FindAllTracksFilterTransform(res, flags, l.X, l.Y, l.Z, l.TrackAndDirection, l.Company, l.TrackType, func(sim.LocationOfInterest) bool { return false }, nil)
		return res
	}
	isPiece := func(l sim.LocationOfInterest) bool {
		tad := l.TaD()
		x, y, z := l.X, l.Y, l.Z
		if tad&4 != 0 { // reversed: the piece starts where a forward walk of it ends
			return true // checked through its forward twin below
		}
		tl := m.TileAt(int(x>>5), int(y>>5))
		if tl == nil {
			return false
		}
		for _, el := range tl.Tracks {
			if el.SequenceIndex == 0 && uint16(el.TrackID) == tad>>3 && uint16(el.Rotation) == tad&3 &&
				int32(el.BaseZ)*4-int32(sim.TrackPieces[el.TrackID][0].Z) == z {
				return true
			}
		}
		return false
	}
	for _, e := range m.Entities {
		if e.EntityType != scenario.VehicleTypeHead || e.Mode != 0 {
			continue
		}
		start := sim.LocationOfInterest{X: int32(e.TileX), Y: int32(e.TileY), Z: int32(e.TileBaseZ) * 4, TrackAndDirection: e.TrackAndDirection & 0x1FF, Company: e.Owner, TrackType: e.TrackType}
		for _, flags := range []uint16{0, sim.SearchUnk2} {
			all := search(start, flags).Reachable.All()
			bad := 0
			for _, l := range all {
				if !isPiece(l) {
					bad++
				}
			}
			again := search(all[len(all)/2], flags).Reachable.All()
			t.Logf("train %d flags %d: %d reachable, %d not real pieces; from a middle one %d", e.ID, flags, len(all), bad, len(again))
			if bad != 0 || len(again) != len(all) {
				t.Errorf("train %d flags %d: %d bad, %d vs %d", e.ID, flags, bad, len(again), len(all))
			}
		}
	}
}
