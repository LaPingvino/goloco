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

// Under every saved road vehicle's tail the lanes its piece uses are marked
// occupied (getRoadOccupationMask); then every tail, driven 100 ticks at its
// train's saved speed (VehicleTail::update), never loses its route and
// frees each routing it leaves.
func TestSavedTailsUpdate(t *testing.T) {
	_, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	laneOK, laneN := 0, 0
	for _, tr := range trains {
		e := m.Entities[tr.Tail]
		if e.Mode != 1 {
			continue
		}
		laneN++
		lanes := sim.RoadOccupationMasks[(e.TrackAndDirection&0x1FF)>>2] >> 4
		for _, el := range m.TileAt(int(e.TileX)>>5, int(e.TileY)>>5).Roads {
			if uint16(el.RoadID) == e.TrackAndDirection>>3&0xF && uint16(el.Rotation) == e.TrackAndDirection&3 && el.LaneOccupation&lanes == lanes {
				laneOK++
				break
			}
		}
	}
	t.Logf("road tails on occupied lanes: %d of %d", laneOK, laneN)
	if laneOK != laneN {
		t.Errorf("%d road tails stand on lanes not marked occupied", laneN-laneOK)
	}
	freed := 0
	for _, tr := range trains {
		head, tail := m.Entities[tr.Head], m.Entities[tr.Tail]
		if tail.Mode > 1 {
			continue
		}
		c := sim.MotionComponent{
			Road: tail.Mode == 1, RightHand: m.TrafficHandedness != 0,
			TrackAndDirection: tail.TrackAndDirection, SubPosition: tail.SubPosition,
			TileX: int32(tail.TileX), TileY: int32(tail.TileY), TileBaseZ: tail.TileBaseZ,
			RemainingDistance: tail.RemainingDistance, X: int32(tail.PosX), Y: int32(tail.PosY), Z: int32(tail.PosZ),
			RoutingHandle: sim.RoutingHandle(tail.RoutingHandle), Owner: tail.Owner, TrackType: tail.TrackType,
		}
		if !c.Road {
			c.TrackAndDirection &= sim.TaDBasicMask
		}
		d1, _ := sim.UpdateDistances(m.Entities[tr.Veh2].CurrentSpeed)
		for range 100 {
			before := c.RoutingHandle
			if !m.TailUpdate(&c, d1, head.HeadVar53) {
				t.Fatalf("head %d: tail lost its route", head.ID)
			}
			if c.RoutingHandle != before {
				freed++
				if m.Routing(before) != sim.RoutingAllocatedButFree {
					t.Errorf("head %d: routing left behind not freed", head.ID)
				}
			}
		}
	}
	t.Logf("tails freed %d routings in 100 ticks", freed)
	if freed == 0 {
		t.Error("no tail left a piece")
	}
}

// Every saved bogie's wheel frame is updateRoll's (numAnimationFrames-1) &
// (var_44/4096) — the var_44 offset and the formula against the save; then
// the bogies of every train, driven 50 ticks at its saved speed with
// VehicleBogie::update, keep their routes and stay on their paths.
func TestSavedBogies(t *testing.T) {
	r, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	roll := func(e *scenario.VehicleEntity) sim.BogieRoll {
		b := sim.BogieRoll{Var44: e.Var44, WheelSlipping: e.WheelSlipping, Reversed: e.Flags38&2 != 0, AnimationIndex: e.Animation}
		if vs := r.Objects.Vehicles; int(e.ObjectID) < len(vs) && vs[e.ObjectID] != nil && e.ObjectSpriteType != 0xFF {
			b.NumAnimationFrames = vs[e.ObjectID].BogieSprites[e.ObjectSpriteType&1].NumAnimationFrames
		}
		return b
	}
	match, n := 0, 0
	for _, tr := range trains {
		for _, car := range tr.Cars {
			for _, cc := range car {
				for _, i := range []int{cc.Front, cc.Back} {
					b := roll(&m.Entities[i])
					if b.NumAnimationFrames == 0 {
						continue
					}
					n++
					if (b.NumAnimationFrames-1)&uint8(b.Var44/4096) == b.AnimationIndex {
						match++
					}
				}
			}
		}
	}
	t.Logf("animated bogies %d: frame matches var_44 in %d", n, match)
	if n == 0 || match != n {
		t.Errorf("%d of %d bogie frames differ from updateRoll", n-match, n)
	}
	moved := 0
	for _, tr := range trains {
		head := m.Entities[tr.Head]
		if head.Mode > 1 {
			continue
		}
		d1, d2 := sim.UpdateDistances(m.Entities[tr.Veh2].CurrentSpeed)
		for _, car := range tr.Cars {
			for _, cc := range car {
				for _, i := range []int{cc.Front, cc.Back} {
					e := &m.Entities[i]
					c := sim.MotionComponent{
						Road: e.Mode == 1, RightHand: m.TrafficHandedness != 0,
						TrackAndDirection: e.TrackAndDirection, SubPosition: e.SubPosition,
						TileX: int32(e.TileX), TileY: int32(e.TileY), TileBaseZ: e.TileBaseZ,
						RemainingDistance: e.RemainingDistance, X: int32(e.PosX), Y: int32(e.PosY), Z: int32(e.PosZ),
						RoutingHandle: sim.RoutingHandle(e.RoutingHandle), Owner: e.Owner, TrackType: e.TrackType,
					}
					if !c.Road {
						c.TrackAndDirection &= sim.TaDBasicMask
					}
					b := roll(e)
					x0, y0 := c.X, c.Y
					for range 50 {
						if !m.BogieUpdate(&c, &b, d1, d2, head.HeadVar53) {
							t.Fatalf("head %d: bogie lost its route", head.ID)
						}
					}
					if c.X != x0 || c.Y != y0 {
						moved++
					}
				}
			}
		}
	}
	t.Logf("bogies moved in 50 ticks: %d", moved)
	if moved == 0 {
		t.Error("no bogie moved")
	}
}

// Vehicle2's speed step on every saved land vehicle, 200 ticks from its saved
// state with its real per-car inputs (isOnRackRail, shouldStartWheelSlipping
// drawing from the saved gPrng1): speeds stay within [0, target] once not
// braking, and settle within 1.5 mph of the target or brake toward it.
func TestSavedVehicle2Speeds(t *testing.T) {
	r, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	n, settled := 0, 0
	for _, tr := range trains {
		head, v1, v2 := m.Entities[tr.Head], m.Entities[tr.Veh1], m.Entities[tr.Veh2]
		if head.Mode > 1 {
			continue
		}
		n++
		cars := m.CarPhysicsOf(tr, r.Objects)
		v := sim.Vehicle2Speed{CurrentSpeed: v2.CurrentSpeed, MotorState: v2.MotorState}
		target := int32(v1.TargetSpeed) << 16
		for range 200 {
			v.UpdateSpeed(v1.TargetSpeed, head.VehicleFlags, head.ManualPower, v2.TotalPower, v2.TotalWeight, v2.Var73&1 != 0, cars)
			if v.CurrentSpeed < 0 || (v.MotorState != sim.MotorBraking && v.CurrentSpeed > target) {
				t.Fatalf("head %d: speed %#x target %#x state %d", head.ID, v.CurrentSpeed, target, v.MotorState)
			}
		}
		if target-v.CurrentSpeed <= 0x18000 || v.MotorState == sim.MotorBraking {
			settled++
		}
	}
	t.Logf("%d land vehicles; %d at their target (or braking to it) after 200 ticks", n, settled)
}
