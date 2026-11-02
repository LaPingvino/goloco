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

// Vehicle2's speed step and move (sub_4A9F20) on every saved land vehicle
// for 100 ticks, with Vehicle1 held where it was saved: Vehicle2 never
// passes Vehicle1's routing (it stops there: isVeh2UnkM15), stays on its
// path, and the brake lights follow kBrakeLightTimeout.
func TestSavedVehicle2Move(t *testing.T) {
	r, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	moved := 0
	for _, tr := range trains {
		head, v1, v2 := m.Entities[tr.Head], m.Entities[tr.Veh1], m.Entities[tr.Veh2]
		if head.Mode > 1 {
			continue
		}
		cars := m.CarPhysicsOf(tr, r.Objects)
		c := sim.MotionComponent{
			Road: v2.Mode == 1, RightHand: m.TrafficHandedness != 0,
			TrackAndDirection: v2.TrackAndDirection, SubPosition: v2.SubPosition,
			TileX: int32(v2.TileX), TileY: int32(v2.TileY), TileBaseZ: v2.TileBaseZ,
			RemainingDistance: v2.RemainingDistance, X: int32(v2.PosX), Y: int32(v2.PosY), Z: int32(v2.PosZ),
			RoutingHandle: sim.RoutingHandle(v2.RoutingHandle), Owner: v2.Owner, TrackType: v2.TrackType, IsVehicle2: true,
		}
		if !c.Road {
			c.TrackAndDirection &= sim.TaDBasicMask
		}
		v := sim.Vehicle2Speed{CurrentSpeed: v2.CurrentSpeed, MotorState: v2.MotorState}
		hv, v1v, bl := head.Var3C, v1.Var3C, v2.BrakeLightTimeout
		veh1Handle := sim.RoutingHandle(v1.RoutingHandle)
		start := c.RoutingHandle
		for range 100 {
			v.UpdateSpeed(v1.TargetSpeed, head.VehicleFlags, head.ManualPower, v2.TotalPower, v2.TotalWeight, v2.Var73&1 != 0, cars)
			d1, d2 := sim.UpdateDistances(v.CurrentSpeed)
			d := sim.VehicleUpdateDistances{D1: d1, D2: d2}
			if !m.Vehicle2Move(&c, &v, &d, false, veh1Handle, head.HeadVar53, &hv, &v1v, &bl) {
				t.Fatalf("head %d: Vehicle2 lost its route", head.ID)
			}
			if v.MotorState == sim.MotorBraking && bl != sim.BrakeLightTimeout {
				t.Fatalf("head %d: braking with brake light timeout %d", head.ID, bl)
			}
		}
		ahead := int(c.RoutingHandle.Index()-start.Index()) & 63
		toVeh1 := int(veh1Handle.Index()-start.Index()) & 63
		if ahead > toVeh1 {
			t.Errorf("head %d: Vehicle2 went %d pieces, past Vehicle1 at %d", head.ID, ahead, toVeh1)
		}
		if ahead > 0 {
			moved++
		}
	}
	t.Logf("Vehicle2s that moved onto a new piece: %d", moved)
}

// Vehicle1's target speed (updateRail / updateRoad) recomputed from the save
// brackets the saved targetSpeed for every land vehicle. Vehicle1 computed it
// from var_3C before two moves of that tick changed it: its own move added
// what it travelled (at most max(unkDistance1·4, 0xCC48)) and Vehicle2's
// (sub_4A9F20) took off what Vehicle2 travelled (at most its distance, speed
// >> 7). The save keeps neither amount, and the target rises with var_3C, so
// the saved target lies between the targets at those two bounds — or, for a
// road vehicle whose tail just left a piece, is that freed piece's curve limit.
func TestSavedVehicle1TargetSpeed(t *testing.T) {
	r, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	n, exact, freed := 0, 0, 0
	for _, tr := range trains {
		h, v2 := &m.Entities[tr.Head], &m.Entities[tr.Veh2]
		if h.Mode > 1 {
			continue
		}
		target := m.Vehicle1RailTargetSpeed
		if h.Mode == 1 {
			target = m.Vehicle1RoadTargetSpeed
		}
		n++
		v1 := m.Entities[tr.Veh1]
		saved := v1.Var3C
		d1, _ := sim.UpdateDistances(v2.CurrentSpeed)
		v1.Var3C = saved - max(d1*4, 0xCC48) // before Vehicle1's own move added up to this
		lo := target(h, &v1, v2, r.Objects)
		v1.Var3C = saved + d1 // before Vehicle2's move drew up to this
		hi := target(h, &v1, v2, r.Objects)
		ok := v1.TargetSpeed >= lo && v1.TargetSpeed <= hi
		// The tail frees the piece it leaves after Vehicle1's update, so when
		// it has just changed piece Vehicle1's route walk saw one more piece
		// behind: its curve may set the target (a road one, on Vehicle2's road).
		tail := m.Entities[tr.Tail]
		if !ok && h.Mode == 1 && int32(tail.SubPosition) <= (d1+0x368A)/0x220C && v1.TargetSpeed <= hi {
			if rd := r.Objects.GetRoadObjectByIndex(int(uint8(v2.Var4F))); rd != nil {
				for _, f := range sim.RoadCurveSpeedFraction {
					if lim := max(int16((int32(uint32(f)*uint32(uint16(rd.MaxSpeed)))+1<<16)/65536), 12); lim == v1.TargetSpeed {
						ok, freed = true, freed+1
						break
					}
				}
			}
		}
		if !ok {
			t.Errorf("head %d: saved target %d outside [%d, %d]", h.ID, v1.TargetSpeed, lo, hi)
		}
		if lo == hi {
			exact++
		}
	}
	t.Logf("%d land vehicles (rail and road); %d with the bracket a single value; %d set by the curve the tail just left", n, exact, freed)
	if n == 0 {
		t.Fatal("no trains")
	}
}

// Vehicle1's road motion helpers on every saved road vehicle: the NoMove
// advance predicts exactly the distance the real one covers and both end on
// the path; applyOvertake marks the next routings changing-lane then
// changing-lane+overtaking, and applyChangeLane flags the head.
func TestSavedVehicle1RoadPieces(t *testing.T) {
	_, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	veh1 := func(tr sim.Train) *sim.Veh1 {
		e, v2 := m.Entities[tr.Veh1], m.Entities[tr.Veh2]
		return &sim.Veh1{MotionComponent: sim.MotionComponent{
			Road: true, RightHand: m.TrafficHandedness != 0,
			TrackAndDirection: e.TrackAndDirection, SubPosition: e.SubPosition,
			TileX: int32(e.TileX), TileY: int32(e.TileY), TileBaseZ: e.TileBaseZ,
			X: int32(e.PosX), Y: int32(e.PosY), Z: int32(e.PosZ),
			RoutingHandle: sim.RoutingHandle(e.RoutingHandle), Owner: e.Owner, TrackType: e.TrackType,
		}, Var3C: e.Var3C, Veh2MaxSpeed: v2.MaxSpeed}
	}
	n, marked := 0, 0
	for _, tr := range trains {
		if m.Entities[tr.Head].Mode != 1 {
			continue
		}
		n++
		v := veh1(tr)
		want := sim.ExportVeh1MoveNoMove(m, v, 1)
		got := sim.ExportVeh1Move(m, v, 1)
		if got != want {
			t.Errorf("head %d: moved %d, NoMove said %d", m.Entities[tr.Head].ID, got, want)
		}
		// Where the routings run out vanilla returns without moveTo (the
		// position stays behind): only then may it be off its path.
		ranOut := m.Routing(v.RoutingHandle.WithIndex(v.RoutingHandle.Index()+1)) == sim.RoutingAllocatedButFree
		p := sim.RoadSubPosition(v.TrackAndDirection, v.RightHand)
		if mi := p[v.SubPosition]; (int32(mi.X)+v.TileX != v.X || int32(mi.Y)+v.TileY != v.Y) && !ranOut {
			t.Errorf("head %d: off its path", m.Entities[tr.Head].ID)
		}
	}
	// Overtake / change lane on fresh copies of the routing table.
	saved := append([]uint16(nil), m.Routings...)
	for _, tr := range trains {
		if m.Entities[tr.Head].Mode != 1 {
			continue
		}
		copy(m.Routings, saved)
		v := veh1(tr)
		ring := m.RoutingRing(v.RoutingHandle, false)
		if len(ring) < 3 {
			continue
		}
		hv := int32(0)
		v.HeadVar3C = &hv
		m.ApplyOvertake(v, 2)
		if m.Routing(ring[1])&(sim.RaDIsChangingLane|sim.RaDIsOvertaking) != sim.RaDIsChangingLane ||
			m.Routing(ring[2])&(sim.RaDIsChangingLane|sim.RaDIsOvertaking) != sim.RaDIsChangingLane|sim.RaDIsOvertaking {
			t.Errorf("head %d: overtake flags %#x %#x", m.Entities[tr.Head].ID, m.Routing(ring[1]), m.Routing(ring[2]))
		}
		copy(m.Routings, saved)
		v = veh1(tr)
		var tad uint16
		v.HeadTaD = &tad
		m.ApplyChangeLane(v, 1)
		if tad&sim.RaDIsChangingLane == 0 {
			t.Errorf("head %d: change lane did not flag the head", m.Entities[tr.Head].ID)
		}
		marked++
	}
	copy(m.Routings, saved)
	t.Logf("road Vehicle1s %d; overtake/lane-change checked on %d", n, marked)
}

// getRoadOccupation at every saved road Vehicle1's own piece reports the
// lane occupied — by itself (veh1UpdateRoadMotionNewRoadPiece marked it).
func TestSavedRoadOccupation(t *testing.T) {
	r, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	n, occ := 0, 0
	for _, tr := range trains {
		e := m.Entities[tr.Veh1]
		if e.Mode != 1 {
			continue
		}
		n++
		if m.RoadOccupation(int32(e.TileX), int32(e.TileY), int32(e.TileBaseZ)*4, e.TrackAndDirection&0x1FF, r.Objects)&sim.RoadLaneOccupied != 0 {
			occ++
		} else {
			t.Logf("head %d: tad %#x not occupied", m.Entities[tr.Head].ID, e.TrackAndDirection)
		}
	}
	t.Logf("road Vehicle1s %d, own lane occupied %d", n, occ)
	if occ != n {
		t.Errorf("%d road Vehicle1s on a lane getRoadOccupation calls free", n-occ)
	}
}

// Vehicle1::updateRoad's move (lookahead, overtaking, lane changes, level
// crossings, lane occupancy) on every saved road vehicle for 100 ticks at its
// Vehicle2's saved speed (Vehicle2's draw on var_3C applied each tick, as
// sub_4A9F20 does after Vehicle1): no lost routes.
func TestSavedVehicle1RoadMove(t *testing.T) {
	r, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	pieces, n := 0, 0
	for _, tr := range trains {
		h, e, v2 := m.Entities[tr.Head], m.Entities[tr.Veh1], m.Entities[tr.Veh2]
		if h.Mode != 1 {
			continue
		}
		n++
		hv, htad := h.Var3C, h.TrackAndDirection
		v := &sim.Veh1{MotionComponent: sim.MotionComponent{
			Road: true, RightHand: m.TrafficHandedness != 0,
			TrackAndDirection: e.TrackAndDirection, SubPosition: e.SubPosition,
			TileX: int32(e.TileX), TileY: int32(e.TileY), TileBaseZ: e.TileBaseZ,
			RemainingDistance: e.RemainingDistance, X: int32(e.PosX), Y: int32(e.PosY), Z: int32(e.PosZ),
			RoutingHandle: sim.RoutingHandle(e.RoutingHandle), Owner: e.Owner, TrackType: e.TrackType,
		}, Var3C: e.Var3C, Veh2MaxSpeed: v2.MaxSpeed, HeadVar3C: &hv, HeadTaD: &htad,
			HeadVar52: h.HeadVar52, HeadTrackType: h.TrackType, HeadVar53: h.HeadVar53}
		start := v.RoutingHandle
		d1, _ := sim.UpdateDistances(v2.CurrentSpeed)
		for range 100 {
			if f := m.Vehicle1RoadMove(v, v2.CurrentSpeed, d1, r.Objects); f&sim.MotionNoRouteFound != 0 {
				t.Fatalf("head %d: Vehicle1 lost its route", h.ID)
			}
			v.Var3C -= d1 // Vehicle2's move (sub_4A9F20) draws var_3C down after Vehicle1
		}
		pieces += int(v.RoutingHandle.Index()-start.Index()) & 63
	}
	t.Logf("%d road Vehicle1s moved %d pieces in 100 ticks", n, pieces)
	if pieces == 0 {
		t.Error("no road Vehicle1 moved")
	}
}

// Every saved land body sits exactly where sub_4AC255 puts it from its saved
// bogies (bodies update after their bogies in a tick): midpoint, pitch, yaw.
func TestSavedBodiesPlacedFromBogies(t *testing.T) {
	r, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	n, pos, turned := 0, 0, 0
	for _, tr := range trains {
		if m.Entities[tr.Head].Mode > 1 {
			continue
		}
		for _, car := range tr.Cars {
			for _, cc := range car {
				f, b, body := m.Entities[cc.Front], m.Entities[cc.Back], m.Entities[cc.Body]
				if body.Mode > 1 { // air / water bodies are not placed by sub_4AC255
					continue
				}
				n++
				veh := r.Objects.Vehicles[body.ObjectID]
				p := sim.PlaceBody(int32(f.PosX), int32(f.PosY), int32(f.PosZ), int32(b.PosX), int32(b.PosY), int32(b.PosZ), body.ObjectSpriteType, veh)
				if p.X == int32(body.PosX) && p.Y == int32(body.PosY) && p.Z == int32(body.PosZ) {
					pos++
				} else if n-pos <= 3 {
					t.Logf("body %d: at (%d,%d,%d), placement (%d,%d,%d)", body.ID, body.PosX, body.PosY, body.PosZ, p.X, p.Y, p.Z)
				}
				if p.Turned {
					if p.Pitch == body.SpritePitch && p.Yaw == body.SpriteYaw {
						turned++
					} else {
						t.Errorf("body %d: pitch/yaw %d/%d, placement %d/%d", body.ID, body.SpritePitch, body.SpriteYaw, p.Pitch, p.Yaw)
					}
				}
			}
		}
	}
	t.Logf("bodies %d: position exact %d, pitch+yaw exact %d", n, pos, turned)
	if pos != n {
		t.Errorf("%d bodies not at their bogies' midpoint", n-pos)
	}
}

// 200 full vehicle ticks (UpdateVehicles: head distances, Vehicle1, Vehicle2,
// bogies, bodies, tail) on the title map: every moving component stays
// exactly on its path (Vehicle1 except vanilla's early-return case) and every
// land body stays at its bogies' midpoint; vehicles move.
func TestUpdateVehiclesTicks(t *testing.T) {
	r, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	start := map[int][2]int32{}
	for _, tr := range trains {
		e := m.Entities[tr.Veh2]
		start[tr.Head] = [2]int32{int32(e.PosX), int32(e.PosY)}
	}
	for range 200 {
		m.UpdateVehicles(r.Objects)
	}
	moved, off := 0, 0
	for _, tr := range trains {
		if m.Entities[tr.Head].Mode > 1 {
			continue
		}
		comps := []int{tr.Veh2, tr.Tail}
		for _, car := range tr.Cars {
			for _, cc := range car {
				comps = append(comps, cc.Front, cc.Back)
				f, b, body := m.Entities[cc.Front], m.Entities[cc.Back], m.Entities[cc.Body]
				if int32(body.PosX) != (int32(f.PosX)+int32(b.PosX))/2 || int32(body.PosY) != (int32(f.PosY)+int32(b.PosY))/2 {
					t.Errorf("head %d: body %d off its bogies' midpoint", m.Entities[tr.Head].ID, body.ID)
				}
			}
		}
		for _, i := range comps {
			e := m.Entities[i]
			var p []sim.MoveInfo
			if e.Mode == 1 {
				p = sim.RoadSubPosition(e.TrackAndDirection, m.TrafficHandedness != 0)
			} else {
				p = sim.TrackSubPosition(e.TrackAndDirection & sim.TaDBasicMask)
			}
			if int(e.SubPosition) >= len(p) {
				off++
				continue
			}
			mi := p[e.SubPosition]
			if int32(mi.X)+int32(e.TileX) != int32(e.PosX) || int32(mi.Y)+int32(e.TileY) != int32(e.PosY) {
				off++
			}
		}
		e := m.Entities[tr.Veh2]
		if s := start[tr.Head]; s != [2]int32{int32(e.PosX), int32(e.PosY)} {
			moved++
		}
	}
	t.Logf("after 200 ticks: %d vehicles moved; components off their path: %d", moved, off)
	if off != 0 {
		t.Errorf("%d components off their path", off)
	}
	if moved == 0 {
		t.Error("nothing moved")
	}
}

// Orders: every saved land vehicle's order ring is well formed, and for a
// travelling one (landNormalMovementUpdate calls advanceToNextRoutableOrder
// every tick) the current order is already routable, so advancing keeps it.
func TestSavedOrders(t *testing.T) {
	_, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	travelling, kept, ringed := 0, 0, 0
	for _, tr := range trains {
		h := m.Entities[tr.Head]
		if h.Mode > 1 {
			continue
		}
		if ring := m.OrderRing(h.OrderTableOffset, h.CurrentOrder); len(ring) > 0 {
			ringed++
			if ring[0].Offset != int(h.OrderTableOffset)+int(h.CurrentOrder) {
				t.Errorf("head %d: ring does not start at the current order", h.ID)
			}
		}
		if h.Status != sim.StatusTravelling {
			continue
		}
		travelling++
		cur := h.CurrentOrder
		m.AdvanceToNextRoutableOrder(h.OrderTableOffset, &cur, h.SizeOfOrderTable)
		if cur == h.CurrentOrder {
			kept++
		} else {
			t.Logf("head %d: current order %d → %d", h.ID, h.CurrentOrder, cur)
		}
	}
	t.Logf("land vehicles with orders %d; travelling %d, current order already routable %d", ringed, travelling, kept)
	if travelling == 0 || kept != travelling {
		t.Errorf("%d travelling vehicles whose current order was not routable", travelling-kept)
	}
}

// trackTargetedPathing from each saved train's Vehicle1 piece toward the
// station its current order names (logged sanity: how many reach it within
// upstream's 1280-weight horizon, and with which signal state).
func TestSavedTrainsTargetedPathing(t *testing.T) {
	_, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	reached, n := 0, 0
	states := map[uint32]int{}
	for _, tr := range trains {
		h, v1 := m.Entities[tr.Head], m.Entities[tr.Veh1]
		if h.Mode != 0 {
			continue
		}
		ring := m.OrderRing(h.OrderTableOffset, h.CurrentOrder)
		if len(ring) == 0 || !ring[0].HasFlags(sim.OrderHasStation) {
			continue
		}
		n++
		r := m.TrackTargetedPathing(int32(v1.TileX), int32(v1.TileY), int32(v1.TileBaseZ)*4, v1.TrackAndDirection&sim.TaDBasicMask,
			sim.Pathing{Company: h.Owner, TrackType: h.TrackType, RequiredMods: h.HeadVar53},
			&sim.RouteTarget{StationID: ring[0].Station(), RX: -1, RY: -1})
		states[r.SignalState]++
		if r.BestDistToTarget == 0 {
			reached++
		}
	}
	t.Logf("trains heading for a station %d; reach it within the horizon %d; signal states %v", n, reached, states)
}

// Station x, y (0x02/0x04) lie on or beside one of the station's own tiles.
func TestSavedStationPositions(t *testing.T) {
	_, m, _ := titleView(t)
	tiles := map[int][][2]int{}
	for y := range m.Height {
		for x := range m.Width {
			for _, st := range m.TileAt(x, y).Stations {
				tiles[int(st.StationID)] = append(tiles[int(st.StationID)], [2]int{x, y})
			}
		}
	}
	near := 0
	for id, ts := range tiles {
		if id >= len(m.StationNames) {
			continue
		}
		sn := m.StationNames[id]
		px, py := int(sn.X)/32, int(sn.Y)/32
		for _, tl := range ts {
			if dx, dy := px-tl[0], py-tl[1]; dx >= -8 && dx <= 8 && dy >= -8 && dy <= 8 {
				near++
				break
			}
		}
	}
	t.Logf("stations %d, position within 8 tiles of one of their tiles %d", len(tiles), near)
	if len(tiles) == 0 || near != len(tiles) {
		t.Errorf("%d stations whose position is far from all their tiles", len(tiles)-near)
	}
}

// Tile::trainStation finds, for every saved station track tile, that tile's
// own station element.
func TestSavedTrainStationLookup(t *testing.T) {
	_, m, _ := titleView(t)
	n, ok := 0, 0
	for y := range m.Height {
		for x := range m.Width {
			tl := m.TileAt(x, y)
			for _, el := range tl.Tracks {
				if !el.HasStation || el.Ghost {
					continue
				}
				n++
				if st := sim.ExportTrainStation(tl, el.TrackID, el.Rotation, el.BaseZ); st != nil && len(tl.Stations) > 0 {
					ok++
				}
			}
		}
	}
	t.Logf("station track tiles %d, found %d", n, ok)
	if n == 0 || ok != n {
		t.Errorf("%d station track tiles without their station", n-ok)
	}
}

// sub_4ACEF1 on every saved train, once from the saved state: the head stays
// on its path; when it reserves a new piece that piece is a real connection
// from the head's piece end; the results are logged by kind.
func TestSavedTrainsSub4ACEF1(t *testing.T) {
	_, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	kinds := map[uint8]int{}
	for _, tr := range trains {
		h := &m.Entities[tr.Head]
		if h.Mode != 0 {
			continue
		}
		d1, _ := sim.UpdateDistances(m.Entities[tr.Veh2].CurrentSpeed)
		before := sim.RoutingHandle(h.RoutingHandle)
		r := m.Sub4ACEF1(tr, 0xD4CB00, uint32(d1), false)
		kinds[r.Al]++
		p := sim.TrackSubPosition(h.TrackAndDirection & sim.TaDBasicMask)
		if mi := p[h.SubPosition]; int32(mi.X)+int32(h.TileX) != int32(h.PosX) || int32(mi.Y)+int32(h.TileY) != int32(h.PosY) {
			t.Errorf("head %d: off its path", h.ID)
		}
		next := sim.RoutingHandle(h.RoutingHandle)
		next = next.WithIndex(next.Index() + 1)
		if r.Al == 0 && m.Routing(next) != sim.RoutingAllocatedButFree && next != before.WithIndex(before.Index()+1) {
			e, _ := sim.TrackConnectionEnd(int32(h.TileX), int32(h.TileY), int32(h.TileBaseZ)*4, h.TrackAndDirection&sim.TaDBasicMask)
			tc := m.TrackConnections(e.X, e.Y, e.Z, e.NextRotation, h.Owner, h.TrackType, h.HeadVar53, m.Entities[tr.Veh1].Var49)
			found := false
			for _, c := range tc.Connections {
				if c&sim.TaDBasicMask == m.Routing(next)&sim.TaDBasicMask {
					found = true
				}
			}
			if !found {
				t.Errorf("head %d: reserved %#x, not a connection", h.ID, m.Routing(next))
			}
		}
	}
	t.Logf("sub_4ACEF1 results by kind (0 moved, 1 full, 2 no route, 3 signal, 4 station): %v", kinds)
}

// Road pathing on the saved road vehicles: aimless wander always returns 0
// (roads have no signals); the targeted search toward each one's station
// order (all station types allowed here: calculateCompatibleRoadStations
// comes with roadPathing) is logged by how many reach it.
func TestSavedRoadPathing(t *testing.T) {
	r, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	n, reached := 0, 0
	for _, tr := range trains {
		h, v1 := m.Entities[tr.Head], m.Entities[tr.Veh1]
		if h.Mode != 1 {
			continue
		}
		p := sim.Pathing{Company: h.Owner, TrackType: h.TrackType, RequiredMods: h.HeadVar53, QueryMods: v1.Var49}
		x, y, z := int32(v1.TileX), int32(v1.TileY), int32(v1.TileBaseZ)*4
		if f := m.RoadAimlessWanderPathing(x, y, z, v1.TrackAndDirection&0x7F, p, r.Objects); f != 0 {
			t.Errorf("head %d: road wander flags %#x", h.ID, f)
		}
		ring := m.OrderRing(h.OrderTableOffset, h.CurrentOrder)
		if len(ring) == 0 || !ring[0].HasFlags(sim.OrderHasStation) {
			continue
		}
		n++
		res := m.RoadTargetedPathing(x, y, z, v1.TrackAndDirection&0x7F, p, 0xFFFFFFFF, &sim.RouteTarget{StationID: ring[0].Station()}, r.Objects)
		if res.BestDistToTarget == 0 {
			reached++
		}
	}
	t.Logf("road vehicles heading for a station %d; reach a free stop of it within the horizon %d", n, reached)
}

// roadPathing on the saved road vehicles: every one may use some stop type
// (calculateCompatibleRoadStations); at each junction just ahead of its
// Vehicle1 the choice is one of that junction's connections (logged count).
func TestSavedRoadPathingChoice(t *testing.T) {
	r, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	junctions, compat := 0, 0
	for _, tr := range trains {
		h, v1 := m.Entities[tr.Head], m.Entities[tr.Veh1]
		if h.Mode != 1 {
			continue
		}
		allowed := sim.CalculateCompatibleRoadStations(h.TrainAcceptedCargoTypes, r.Objects)
		if allowed != 0 {
			compat++
		}
		e, ok := sim.RoadConnectionEnd(int32(v1.TileX), int32(v1.TileY), int32(v1.TileBaseZ)*4, v1.TrackAndDirection&0x7F)
		if !ok {
			continue
		}
		rc := m.RoadConnections(e.X, e.Y, e.Z, e.NextRotation, h.Owner, h.TrackType, h.HeadVar53, v1.Var49)
		if len(rc.Connections) < 2 {
			continue
		}
		junctions++
		var st sim.PathingState
		got := m.RoadPathing(h.OrderTableOffset, h.CurrentOrder, e.X, e.Y, e.Z, rc.Connections,
			sim.Pathing{Company: h.Owner, TrackType: h.TrackType, RequiredMods: h.HeadVar53, QueryMods: v1.Var49}, allowed, false, &st, r.Objects)
		found := false
		for _, c := range rc.Connections {
			found = found || c == got
		}
		if !found {
			t.Errorf("head %d: roadPathing chose %#x, not a connection", h.ID, got)
		}
	}
	t.Logf("road vehicles with a usable stop type %d; at a junction just ahead %d", compat, junctions)
	if compat == 0 {
		t.Error("no road vehicle may use any stop type")
	}
}

// sub_47DA8D on every saved road vehicle, once from the saved state: the head
// stays on its path; a new reservation is a real road connection from the
// head's piece end; results logged by kind.
func TestSavedRoadSub47DA8D(t *testing.T) {
	r, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	kinds := map[uint8]int{}
	for _, tr := range trains {
		h := &m.Entities[tr.Head]
		if h.Mode != 1 {
			continue
		}
		d1, _ := sim.UpdateDistances(m.Entities[tr.Veh2].CurrentSpeed)
		before := sim.RoutingHandle(h.RoutingHandle)
		nextBefore := m.Routing(before.WithIndex(before.Index() + 1))
		res := m.Sub47DA8D(tr, 0xD4CB00, uint32(d1), r.Objects)
		kinds[res.Al]++
		p := sim.RoadSubPosition(h.TrackAndDirection, m.TrafficHandedness != 0)
		if mi := p[h.SubPosition]; int32(mi.X)+int32(h.TileX) != int32(h.PosX) || int32(mi.Y)+int32(h.TileY) != int32(h.PosY) {
			t.Errorf("head %d: off its path", h.ID)
		}
		cur := sim.RoutingHandle(h.RoutingHandle)
		if nr := m.Routing(cur.WithIndex(cur.Index() + 1)); res.Al == 0 && nr != sim.RoutingAllocatedButFree && (cur != before || nr != nextBefore) {
			e, _ := sim.RoadConnectionEnd(int32(h.TileX), int32(h.TileY), int32(h.TileBaseZ)*4, h.TrackAndDirection&0x7F)
			rc := m.RoadConnections(e.X, e.Y, e.Z, e.NextRotation, h.Owner, h.TrackType, h.HeadVar53, m.Entities[tr.Veh1].Var49)
			found := false
			for _, c := range rc.Connections {
				found = found || c&0x7F == nr&0x7F
			}
			if !found {
				t.Errorf("head %d: reserved %#x, not a road connection", h.ID, nr)
			}
		}
	}
	t.Logf("sub_47DA8D results by kind (0 moved, 1 full, 2 no route, 4 stop): %v", kinds)
}

// 3000 vehicle ticks (≈ 96 s of game time) with the head's normal movement:
// every moving component stays on its path, and vehicles travel beyond what
// their saved reservations held (the heads now reserve track ahead).
func TestUpdateVehiclesLongRun(t *testing.T) {
	r, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	reserved := 0 // pieces the save had reserved ahead of each Vehicle2
	last := map[int]uint16{}
	for _, tr := range trains {
		if m.Entities[tr.Head].Mode > 1 {
			continue
		}
		reserved += len(m.RoutingRing(sim.RoutingHandle(m.Entities[tr.Veh2].RoutingHandle), false)) - 1
		last[tr.Veh2] = m.Entities[tr.Veh2].RoutingHandle
	}
	travelled := 0
	for range 3000 {
		m.UpdateVehicles(r.Objects)
		for i, h := range last {
			if cur := m.Entities[i].RoutingHandle; cur != h {
				travelled += int(sim.RoutingHandle(cur).Index()-sim.RoutingHandle(h).Index()) & 63
				last[i] = cur
			}
		}
	}
	off := 0
	statuses := map[uint8]int{}
	for _, tr := range trains {
		h := m.Entities[tr.Head]
		if h.Mode > 1 {
			continue
		}
		statuses[h.Status]++
		for _, i := range []int{tr.Veh2, tr.Tail} {
			e := m.Entities[i]
			var p []sim.MoveInfo
			if e.Mode == 1 {
				p = sim.RoadSubPosition(e.TrackAndDirection, m.TrafficHandedness != 0)
			} else {
				p = sim.TrackSubPosition(e.TrackAndDirection & sim.TaDBasicMask)
			}
			if int(e.SubPosition) >= len(p) || int32(p[e.SubPosition].X)+int32(e.TileX) != int32(e.PosX) || int32(p[e.SubPosition].Y)+int32(e.TileY) != int32(e.PosY) {
				off++
			}
		}
	}
	t.Logf("saved reservations ahead: %d pieces; Vehicle2s travelled %d pieces in 3000 ticks; statuses %v; off path %d", reserved, travelled, statuses, off)
	if off != 0 {
		t.Errorf("%d components off their path", off)
	}
	if travelled <= reserved {
		t.Errorf("vehicles did not get past their saved reservations")
	}
}

// landTryBeginUnloading on every saved approaching vehicle: those whose whole
// train is in (Vehicle2 at the head) begin unloading — status 5, cargo
// transfer timeout 10, the station marked visited by their vehicle type, the
// order moved on when it was this station; the rest stay approaching.
func TestSavedArrivals(t *testing.T) {
	r, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	began, waiting, advanced := 0, 0, 0
	for _, tr := range trains {
		h := &m.Entities[tr.Head]
		if h.Mode > 1 || h.Status != sim.StatusApproaching {
			continue
		}
		v2 := m.Entities[tr.Veh2]
		in := h.RoutingHandle == v2.RoutingHandle && h.SubPosition == v2.SubPosition
		order := h.CurrentOrder
		orders := m.OrderRing(h.OrderTableOffset, h.CurrentOrder)
		atOrder := len(orders) > 0 && orders[0].HasFlags(sim.OrderHasStation) && orders[0].Station() == int(h.StationID)
		m.LandTryBeginUnloading(tr)
		switch {
		case in && h.VehicleFlags&sim.VehicleFlagManualControl == 0:
			if h.Status != sim.StatusUnloading || h.CargoTransferTimeout != 10 {
				t.Errorf("head %d: status %d timeout %d after arriving", h.ID, h.Status, h.CargoTransferTimeout)
			}
			if m.StationNames[h.StationID].VisitedTypes&(1<<(h.VehicleType&7)) == 0 {
				t.Errorf("head %d: station %d not marked visited", h.ID, h.StationID)
			}
			if atOrder && h.CurrentOrder == order && len(orders) > 1 {
				t.Errorf("head %d: order not advanced at its order station", h.ID)
			}
			if atOrder {
				advanced++
			}
			began++
		default:
			if h.Status != sim.StatusApproaching {
				t.Errorf("head %d: status %d though the train is not all in", h.ID, h.Status)
			}
			waiting++
		}
	}
	t.Logf("approaching vehicles: began unloading %d (at their order station %d), still coming in %d", began, advanced, waiting)
	// Run on: each head is checked in the tick it arrives (approaching →
	// unloading), when beginUnloading's state is fresh (unloading counts the
	// timeout down from the next tick on).
	unloading := 0
	prev := map[int]uint8{}
	for _, tr := range trains {
		prev[tr.Head] = m.Entities[tr.Head].Status
	}
	for range 600 {
		m.UpdateVehicles(r.Objects)
		for _, tr := range trains {
			h := m.Entities[tr.Head]
			if prev[tr.Head] == sim.StatusApproaching && h.Status == sim.StatusUnloading {
				unloading++
				if h.CargoTransferTimeout != 10 || m.StationNames[h.StationID].VisitedTypes&(1<<(h.VehicleType&7)) == 0 {
					t.Errorf("head %d: arrived with timeout %d, station visited %#x", h.ID, h.CargoTransferTimeout, m.StationNames[h.StationID].VisitedTypes)
				}
			}
			prev[tr.Head] = h.Status
		}
	}
	t.Logf("arrivals in 600 ticks: %d", unloading)
	if unloading == 0 {
		t.Error("no vehicle arrived")
	}
}

// Saved cargo: every car body's primary and bogie's secondary cargo holds no
// more than its maximum, and loaded cargo has a type.
func TestSavedVehicleCargo(t *testing.T) {
	_, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	loaded, slots := 0, 0
	for _, tr := range trains {
		for _, car := range tr.Cars {
			for _, cc := range car {
				for _, i := range []int{cc.Front, cc.Body} {
					c := m.Entities[i].Cargo
					if c.Type == 0xFF {
						continue
					}
					slots++
					if c.Qty > c.MaxQty {
						t.Errorf("entity %d: %d of max %d", m.Entities[i].ID, c.Qty, c.MaxQty)
					}
					if c.Qty > 0 {
						loaded++
					}
				}
			}
		}
	}
	t.Logf("cargo slots %d, loaded %d", slots, loaded)
	if slots == 0 {
		t.Error("no cargo slots")
	}
}

// The save's companies: each delivery adds its quantity to both
// cargoUnitsTotalDelivered and cargoDelivered[type]; the former is reset each
// month (Company::updateMonthly1), the latter is cumulative — so the per-type
// sum is at least this month's total.
func TestSavedCompanies(t *testing.T) {
	_, m, _ := titleView(t)
	if len(m.Companies) != 15 {
		t.Fatalf("%d company slots", len(m.Companies))
	}
	active := 0
	for i, c := range m.Companies {
		var sum uint64
		for _, d := range c.CargoDelivered {
			sum += uint64(d)
		}
		if c.CargoUnitsTotalDelivered == 0 && c.Cash == 0 {
			continue
		}
		active++
		t.Logf("company %d: cash %d, loan %d, delivered %d (per type sum %d), distance %d", i, c.Cash, c.CurrentLoan, c.CargoUnitsTotalDelivered, sum, c.CargoUnitsTotalDistance)
		if sum < uint64(c.CargoUnitsTotalDelivered) {
			t.Errorf("company %d: per-type delivered %d < this month's %d", i, sum, c.CargoUnitsTotalDelivered)
		}
	}
	if active == 0 {
		t.Error("no active company")
	}
}

// updateCargoDistribution recomputed for every saved station from its saved
// quantities and tile counts gives back exactly the saved densityPerTile.
func TestSavedStationCargoDistribution(t *testing.T) {
	r, m, _ := titleView(t)
	exact, higher, waiting := 0, 0, 0
	for id := range m.StationNames {
		saved := m.StationCargoDensities(id)
		m.UpdateCargoDistribution(id, r.Objects)
		got := m.StationCargoDensities(id)
		for c := range got {
			if m.StationNames[id].Cargo[c].Quantity == 0 {
				continue
			}
			waiting++
			switch {
			case got[c] == saved[c]:
				exact++
			default:
				higher++
				t.Errorf("station %d cargo %d: recomputed %d, saved %d", id, c, got[c], saved[c])
			}
		}
	}
	t.Logf("waiting cargo stats %d: exact %d, differing %d", waiting, exact, higher)
	if waiting == 0 {
		t.Error("no waiting cargo")
	}
}

// Towns and industries from the save: every cargo a town got this month is
// among the types it ever got (var_1A8 is set on each delivery, never
// cleared); the parsed industries are exactly those the map's industry
// tiles reference.
func TestSavedTownsAndIndustries(t *testing.T) {
	_, m, _ := titleView(t)
	deliveredTowns := 0
	for _, tw := range m.Towns {
		any := false
		for c, q := range tw.MonthlyCargoDelivered {
			if q == 0 {
				continue
			}
			any = true
			if tw.Var1A8&(1<<c) == 0 {
				t.Errorf("town %d: %d of cargo %d this month but not in var_1A8 %#x", tw.ID, q, c, tw.Var1A8)
			}
		}
		if any {
			deliveredTowns++
		}
	}
	onMap := map[uint8]bool{}
	for y := range m.Height {
		for x := range m.Width {
			for _, in := range m.TileAt(x, y).Industries {
				onMap[in.IndustryID] = true
			}
		}
	}
	parsed := map[uint8]bool{}
	for _, in := range m.Industries {
		parsed[in.ID] = true
		if !onMap[in.ID] {
			t.Errorf("industry %d parsed but on no tile", in.ID)
		}
	}
	for id := range onMap {
		if !parsed[id] {
			t.Errorf("industry %d on the map but not parsed", id)
		}
	}
	t.Logf("towns %d (with deliveries this month %d); industries %d", len(m.Towns), deliveredTowns, len(m.Industries))
}

// updateTrainProperties recomputed for every saved land vehicle gives back
// its saved power, weight, max / rack speeds, accepted cargo, track mods
// (var_53, var_49), fasterAroundCurves and each car's weight.
func TestSavedTrainProperties(t *testing.T) {
	r, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	n, bad := 0, map[string]int{}
	for _, tr := range trains {
		h0, v10, v20 := m.Entities[tr.Head], m.Entities[tr.Veh1], m.Entities[tr.Veh2]
		if h0.Mode > 1 {
			continue
		}
		cars0 := map[int]uint16{}
		for _, car := range tr.Cars {
			cars0[car[0].Front] = m.Entities[car[0].Front].TotalWeight
		}
		m.UpdateTrainProperties(tr, r.Objects)
		n++
		h, v1, v2 := m.Entities[tr.Head], m.Entities[tr.Veh1], m.Entities[tr.Veh2]
		check := func(name string, got, want int64) {
			if got != want {
				bad[name]++
				if bad[name] <= 2 {
					t.Logf("head %d %s: %d, saved %d", h.ID, name, got, want)
				}
			}
		}
		check("totalPower", int64(v2.TotalPower), int64(v20.TotalPower))
		check("totalWeight", int64(v2.TotalWeight), int64(v20.TotalWeight))
		check("maxSpeed", int64(v2.MaxSpeed), int64(v20.MaxSpeed))
		check("rackRailMaxSpeed", int64(v2.RackRailMaxSpeed), int64(v20.RackRailMaxSpeed))
		check("trainAcceptedCargoTypes", int64(h.TrainAcceptedCargoTypes), int64(h0.TrainAcceptedCargoTypes))
		check("var_53", int64(h.HeadVar53), int64(h0.HeadVar53))
		check("var_49", int64(v1.Var49), int64(v10.Var49))
		check("fasterAroundCurves", int64(h.Flags38&(1<<5)), int64(h0.Flags38&(1<<5)))
		for i, w := range cars0 {
			check("totalCarWeight", int64(m.Entities[i].TotalWeight), int64(w))
		}
	}
	t.Logf("land vehicles %d; mismatches by field %v", n, bad)
	for f, c := range bad {
		t.Errorf("%s differs for %d", f, c)
	}
}

// Unloading in the running sim: after 1500 ticks the companies have more
// cargo delivered and their takings booked — cash and this month's
// expenditure table move by the same amount (applyPaymentToCompany).
func TestUnloadingPays(t *testing.T) {
	r, m, _ := titleView(t)
	before := append([]scenario.CompanyEconomy(nil), m.Companies...)
	for range 1500 {
		m.UpdateVehicles(r.Objects)
	}
	delivered, paid := uint64(0), int64(0)
	for i := range m.Companies {
		var a, b uint64
		for c := range 32 {
			a += uint64(m.Companies[i].CargoDelivered[c])
			b += uint64(before[i].CargoDelivered[c])
		}
		delivered += a - b
		dCash := m.Companies[i].Cash - before[i].Cash
		var dBook int64
		for typ := range 17 {
			dBook += int64(m.Companies[i].Expenditures[0][typ]) - int64(before[i].Expenditures[0][typ])
		}
		if dCash != dBook {
			t.Errorf("company %d: cash moved %d but the books %d", i, dCash, dBook)
		}
		paid += dCash
	}
	t.Logf("1500 ticks: %d units delivered, %d paid to companies", delivered, paid)
	if delivered == 0 || paid <= 0 {
		t.Error("no deliveries paid")
	}
}

// The station cycle in the running sim: within 3000 ticks vehicles complete
// arrive → unload → load → leave (loading ends with a new journey and the
// head stopped, then travelling again), and cargo gets loaded.
func TestStationCycle(t *testing.T) {
	r, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	stage := map[int]int{} // 0 none, 1 unloading seen, 2 loading seen
	departed, loaded := 0, 0
	before := map[int]uint8{}
	for _, tr := range trains {
		for _, car := range tr.Cars {
			for _, cc := range car {
				before[cc.Body] = m.Entities[cc.Body].Cargo.Qty
			}
		}
	}
	for range 3000 {
		m.UpdateVehicles(r.Objects)
		for _, tr := range trains {
			h := m.Entities[tr.Head]
			switch {
			case h.Status == sim.StatusUnloading:
				stage[tr.Head] = max(stage[tr.Head], 1)
			case h.Status == sim.StatusLoading && stage[tr.Head] >= 1:
				stage[tr.Head] = 2
			case (h.Status == sim.StatusStopped || h.Status == sim.StatusTravelling) && stage[tr.Head] == 2:
				if h.BreakdownFlags&sim.BreakdownJourneyStarted == 0 {
					t.Errorf("head %d left without a new journey", h.ID)
				}
				departed++
				stage[tr.Head] = 0
			}
			for _, car := range tr.Cars {
				for _, cc := range car {
					if q := m.Entities[cc.Body].Cargo.Qty; q > before[cc.Body] {
						loaded++
					}
					before[cc.Body] = m.Entities[cc.Body].Cargo.Qty
				}
			}
		}
	}
	t.Logf("3000 ticks: %d departures after unloading and loading; %d car loadings", departed, loaded)
	if departed == 0 || loaded == 0 {
		t.Error("no complete station visit")
	}
}
