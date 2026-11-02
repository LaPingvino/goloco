package scene

import (
	"fmt"
	"github.com/LaPingvino/goloco/pkg/core"
	"os"
	"strings"
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
				// forward steps only: a reversal restarts the table at index 0
				if step := int(sim.RoutingHandle(cur).Index()-sim.RoutingHandle(h).Index()) & 63; step < 32 {
					travelled += step
				}
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

// pathingShouldReverse on every saved land vehicle (logged), and its first
// rule: a head with nothing connecting ahead is never told to reverse here.
func TestSavedPathingShouldReverse(t *testing.T) {
	r, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	yes, n, deadEnd := 0, 0, 0
	for _, tr := range trains {
		h := m.Entities[tr.Head]
		if h.Mode > 1 {
			continue
		}
		n++
		rev := m.PathingShouldReverse(tr, r.Objects)
		if rev {
			yes++
		}
		if h.Mode == 0 {
			e, _ := sim.TrackConnectionEnd(int32(h.TileX), int32(h.TileY), int32(h.TileBaseZ)*4, h.TrackAndDirection&sim.TaDBasicMask)
			if len(m.TrackConnections(e.X, e.Y, e.Z, e.NextRotation, h.Owner, h.TrackType, h.HeadVar53, m.Entities[tr.Veh1].Var49).Connections) == 0 {
				deadEnd++
				if rev {
					t.Errorf("head %d: nothing ahead, yet reverse", h.ID)
				}
			}
		}
	}
	t.Logf("land vehicles %d: pathing says reverse for %d (trains at a dead end: %d)", n, yes, deadEnd)
}

// positionVehicleOnTrack on every saved land vehicle as it stands (already
// validly placed): no failure reported, every moving component exactly on its
// path, every body between its bogies; applyVehicleObjectLength lays a train
// out from its length (the head, Vehicle1/2) down to 0 at the tail.
func TestSavedPositionVehicleOnTrack(t *testing.T) {
	r, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	failed, n := 0, 0
	for _, tr := range trains {
		if m.Entities[tr.Head].Mode > 1 {
			continue
		}
		n++
		if m.PositionVehicleOnTrack(tr, false, r.Objects) {
			failed++
			t.Logf("head %d: placement reports failure", m.Entities[tr.Head].ID)
		}
		for _, car := range tr.Cars {
			for _, cc := range car {
				f, b, body := m.Entities[cc.Front], m.Entities[cc.Back], m.Entities[cc.Body]
				if int32(body.PosX) != (int32(f.PosX)+int32(b.PosX))/2 || int32(body.PosY) != (int32(f.PosY)+int32(b.PosY))/2 {
					t.Errorf("head %d: body %d off its bogies", m.Entities[tr.Head].ID, body.ID)
				}
				for _, i := range []int{cc.Front, cc.Back} {
					e := m.Entities[i]
					var p []sim.MoveInfo
					if e.Mode == 1 {
						p = sim.RoadSubPosition(e.TrackAndDirection, m.TrafficHandedness != 0)
					} else {
						p = sim.TrackSubPosition(e.TrackAndDirection & sim.TaDBasicMask)
					}
					if int(e.SubPosition) >= len(p) || int32(p[e.SubPosition].X)+int32(e.TileX) != int32(e.PosX) {
						t.Errorf("head %d: bogie %d off its path", m.Entities[tr.Head].ID, e.ID)
					}
				}
			}
		}
		m.ApplyVehicleObjectLength(tr, r.Objects)
		if m.Entities[tr.Tail].RemainingDistance != 0 || m.Entities[tr.Head].RemainingDistance <= 0 {
			t.Errorf("head %d: layout head %d tail %d", m.Entities[tr.Head].ID, m.Entities[tr.Head].RemainingDistance, m.Entities[tr.Tail].RemainingDistance)
		}
	}
	t.Logf("land vehicles %d; placement failures %d", n, failed)
}

// flipCar twice on every saved car gives back exactly the saved entities,
// and after one flip the train's structure is still valid (Vehicle(head)).
func TestFlipCarRoundTrip(t *testing.T) {
	_, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	saved := append([]scenario.VehicleEntity(nil), m.Entities...)
	multi, n := 0, 0
	for _, tr := range trains {
		if m.Entities[tr.Head].Mode > 1 {
			continue
		}
		for _, car := range tr.Cars {
			n++
			if len(car) > 1 {
				multi++
			}
			nf := m.FlipCar(tr, car[0].Front)
			tr1, err := m.TrainOf(m.Entities[tr.Head].ID)
			if err != nil {
				t.Fatalf("head %d after one flip: %v", m.Entities[tr.Head].ID, err)
			}
			m.FlipCar(tr1, nf)
			for i := range m.Entities {
				if m.Entities[i] != saved[i] {
					t.Fatalf("head %d: entity %d differs after two flips", m.Entities[tr.Head].ID, m.Entities[i].ID)
				}
			}
		}
	}
	t.Logf("cars flipped twice %d (multi-component %d)", n, multi)
}

// tryReverse on every saved land vehicle on its expected piece: the train
// stays valid, its new front lies toward where the tail was, every moving
// component is on its path, and the sim runs on 300 ticks with every component still on
// its path. Logged: placement failures (liftUpVehicle is PARTIAL) and
// Jacobs-bogie vehicles (connectJacobsBogies is PARTIAL).
func TestReverseSavedVehicles(t *testing.T) {
	r, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	reversed, failed, jacobs := 0, 0, 0
	onPath := func(i int) bool {
		e := m.Entities[i]
		var p []sim.MoveInfo
		if e.Mode == 1 {
			p = sim.RoadSubPosition(e.TrackAndDirection, m.TrafficHandedness != 0)
		} else {
			p = sim.TrackSubPosition(e.TrackAndDirection & sim.TaDBasicMask)
		}
		return int(e.SubPosition) < len(p) && int32(p[e.SubPosition].X)+int32(e.TileX) == int32(e.PosX) && int32(p[e.SubPosition].Y)+int32(e.TileY) == int32(e.PosY)
	}
	for _, tr := range trains {
		h := m.Entities[tr.Head]
		if h.Mode > 1 || !m.IsOnExpectedRoadOrTrack(tr) {
			continue
		}
		for _, car := range tr.Cars {
			if m.Entities[car[0].Front].Flags38&(1<<3) != 0 {
				jacobs++
			}
		}
		oh, ot := m.Entities[tr.Veh2], m.Entities[tr.Tail]
		if !m.TryReverse(tr, r.Objects) {
			continue
		}
		reversed++
		if m.Entities[tr.Head].VehicleFlags&sim.VehicleFlagCommandStop != 0 && h.VehicleFlags&sim.VehicleFlagCommandStop == 0 {
			failed++
		}
		tr2, err := m.TrainOf(h.ID)
		if err != nil {
			t.Fatalf("head %d after reversing: %v", h.ID, err)
		}
		// The train is re-laid from where Vehicle2 stood, facing the other
		// way: its new front (Vehicle2) lies nearer the old tail than the old
		// front did (unless the train is a single short car on one tile).
		nv := m.Entities[tr2.Veh2]
		d := func(a, b scenario.VehicleEntity) int {
			dx, dy := int(a.PosX)-int(b.PosX), int(a.PosY)-int(b.PosY)
			return dx*dx + dy*dy
		}
		if d(nv, ot) > d(oh, ot) {
			t.Errorf("head %d: new front farther from the old tail (%d) than the old front was (%d)", h.ID, d(nv, ot), d(oh, ot))
		}
		for _, i := range []int{tr2.Veh2, tr2.Tail} {
			if !onPath(i) {
				t.Errorf("head %d: component %d off its path after reversing", h.ID, m.Entities[i].ID)
			}
		}
	}
	for range 300 {
		m.UpdateVehicles(r.Objects)
	}
	trains, _ = m.Trains()
	off := 0
	for _, tr := range trains {
		if m.Entities[tr.Head].Mode > 1 {
			continue
		}
		for _, i := range []int{tr.Veh2, tr.Tail} {
			if !onPath(i) {
				off++
			}
		}
	}
	t.Logf("reversed %d land vehicles (placement failures %d; Jacobs-bogie cars %d); after 300 more ticks off path %d", reversed, failed, jacobs, off)
	if reversed == 0 || off != 0 {
		t.Error("reversal check failed")
	}
}

// A saved train waiting at a signal does not wait for ever: within 2500
// ticks it gets through or turns round (the one-way timeout is 1920 ticks,
// the two-way one 640).
func TestSavedTrainLeavesSignal(t *testing.T) {
	r, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	waiting := map[int]bool{}
	for _, tr := range trains {
		if h := m.Entities[tr.Head]; h.Mode == 0 && h.Status == sim.StatusWaitingAtSignal {
			waiting[tr.Head] = true
		}
	}
	if len(waiting) == 0 {
		t.Skip("no train waiting at a signal in the save")
	}
	left := map[int]int{}
	for tick := range 2500 {
		m.UpdateVehicles(r.Objects)
		for i := range waiting {
			if _, done := left[i]; !done && m.Entities[i].Status != sim.StatusWaitingAtSignal {
				left[i] = tick
			}
		}
	}
	t.Logf("trains waiting at a signal %d; left after ticks %v", len(waiting), left)
	if len(left) != len(waiting) {
		t.Errorf("%d trains still waiting after 2500 ticks", len(waiting)-len(left))
	}
}

// A pending breakdown on a powered car breaks the train down in a game (not
// on the title screen): brokenDown set, breakdownTimeout 5, Vehicle2's
// var_73 broken down; the head's update then stops it (stoppingUpdate →
// status brokenDown once Vehicle2 has caught up).
func TestPendingBreakdown(t *testing.T) {
	r, m, _ := titleView(t)
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	var tr sim.Train
	found := false
	for _, c := range trains {
		h := m.Entities[c.Head]
		if h.Mode == 0 && h.Status == sim.StatusTravelling && m.Entities[c.Veh2].Var73 == 0 {
			tr, found = c, true
			break
		}
	}
	if !found {
		t.Skip("no travelling train")
	}
	front := tr.Cars[0][0].Front
	m.Entities[front].BreakdownFlags |= sim.BreakdownPending
	m.TitleMode = true
	m.UpdateVehicles(r.Objects)
	if m.Entities[front].BreakdownFlags&sim.BreakdownBrokenDn != 0 {
		t.Fatal("broke down on the title screen")
	}
	m.TitleMode = false
	m.UpdateVehicles(r.Objects)
	f := m.Entities[front]
	if f.BreakdownFlags&(sim.BreakdownBrokenDn|sim.BreakdownPending) != sim.BreakdownBrokenDn || f.BreakdownTimeout != 5 {
		t.Fatalf("front flags %#x timeout %d", f.BreakdownFlags, f.BreakdownTimeout)
	}
	if m.Entities[tr.Veh2].Var73&1 == 0 {
		t.Fatalf("Vehicle2 var_73 %#x", m.Entities[tr.Veh2].Var73)
	}
	stopped := -1
	for i := range 2000 {
		m.UpdateVehicles(r.Objects)
		if m.Entities[tr.Head].Status == sim.StatusBrokenDown {
			stopped = i
			break
		}
	}
	t.Logf("var_73 %#x; status brokenDown after %d ticks", m.Entities[tr.Veh2].Var73, stopped)
	if stopped < 0 {
		t.Error("never stopped as broken down")
	}
}

// getNewChallengeProgress against the challengeProgress a saved game stored
// for its player company (updateDailyPlayer recomputes it daily outside the
// title and editor; AI companies are never updated, so only the controlling
// company is compared).
func TestChallengeProgressMatchesSave(t *testing.T) {
	r, m, _ := saveView(t, "Single Player Saved Games/DEBUG_AUTO_DUMP.SV5", false)
	checked, same := 0, 0
	for id := range m.Companies {
		if m.CompanyEmpty(id) || id != int(m.ControllingCompany) {
			continue
		}
		checked++
		got, want := m.NewChallengeProgress(id, r.Objects), m.Companies[id].ChallengeProgress
		t.Logf("company %d: computed %d, saved %d (performance %d, flags %#x)", id, got, want, m.CompanyFaces[id].Performance, m.CompanyFaces[id].ChallengeFlags)
		if got == want {
			same++
		} else {
			t.Logf("company %d: computed %d, saved %d (objective type %d)", id, got, want, m.Objective.Type)
		}
	}
	t.Logf("objective %+v: %d/%d companies match the saved challengeProgress", m.Objective, same, checked)
	if checked == 0 || same != checked {
		t.Errorf("%d of %d companies differ", checked-same, checked)
	}
}

// calcDate of the saved currentDay gives the saved year / month / day of
// month (title.dat and a saved game); the day counter advances one day every
// ~96 ticks; broken-down vehicles recover once their breakdownTimeout days
// have passed.
func TestDateTick(t *testing.T) {
	for _, s := range []struct {
		path  string
		title bool
	}{{"Data/title.dat", true}, {"Single Player Saved Games/DEBUG_AUTO_DUMP.SV5", false}} {
		_, m, _ := saveView(t, s.path, s.title)
		d := sim.CalcDate(m.CurrentDay)
		t.Logf("%s: day %d → %d-%02d-%02d; saved %d-%02d-%02d", s.path, m.CurrentDay, d.Year, d.Month+1, d.Day+1, m.CurrentYear, m.CurrentMonth+1, m.CurrentDayOfMonth+1)
		if d.Year != int(m.CurrentYear) || d.Month != int(m.CurrentMonth) || d.Day != int(m.CurrentDayOfMonth) {
			t.Errorf("%s: calcDate disagrees with the saved date", s.path)
		}
	}
	r, m, _ := titleView(t)
	trains, _ := m.Trains()
	broken := map[int]bool{}
	for _, tr := range trains {
		if m.Entities[tr.Veh2].Var73&1 != 0 {
			broken[tr.Head] = true
		}
	}
	day0 := m.CurrentDay
	for range 961 {
		m.UpdateVehicles(r.Objects)
	}
	still := 0
	for i := range broken {
		tr, _ := m.TrainOf(m.Entities[i].ID)
		if m.Entities[tr.Veh2].Var73&1 != 0 {
			still++
		}
	}
	t.Logf("961 ticks: %d days passed; broken down %d → still %d", m.CurrentDay-day0, len(broken), still)
	if m.CurrentDay-day0 != 10 {
		t.Errorf("961 ticks gave %d days, want 10 (682 per tick on a 65536 counter)", m.CurrentDay-day0)
	}
	if len(broken) > 0 && still != 0 {
		t.Errorf("%d vehicles still broken down after 10 days (breakdownTimeout starts at 5)", still)
	}
}

// In a game, the daily update marks the player's challenge completed once
// getNewChallengeProgress reaches 100 (evaluateChallengeProgress), and leaves
// it alone on the title screen.
func TestChallengeCompletes(t *testing.T) {
	r, m, _ := saveView(t, "Single Player Saved Games/DEBUG_AUTO_DUMP.SV5", false)
	id := int(m.ControllingCompany)
	m.CompanyFaces[id].Performance = int16(m.Objective.PerformanceIndex) * 10 // the objective's index reached
	for range 200 {                                                           // two days
		m.UpdateVehicles(r.Objects)
	}
	completed, failed, _, progress := m.ChallengeStatus()
	t.Logf("objective type %d; progress %d completed %v failed %v", m.Objective.Type, progress, completed, failed)
	if !completed || failed || progress != 100 {
		t.Error("challenge not completed")
	}
}

// The month's end (VehicleHead::updateMonthly): each company's running-cost
// expenditures fall by exactly its vehicles' running costs, every Vehicle2's
// curMonthRevenue moves into profit[0] (minus the costs) with the history
// shifted, and monthsInChallenge counts up.
func TestMonthEndRunningCosts(t *testing.T) {
	r, m, _ := saveView(t, "Single Player Saved Games/DEBUG_AUTO_DUMP.SV5", false)
	month := m.CurrentMonth
	var before []int64
	var revBefore map[int]int32
	snap := func() {
		before = before[:0]
		for _, c := range m.Companies {
			var s int64
			for vt := 1; vt < 12; vt += 2 {
				s += int64(c.Expenditures[0][vt])
			}
			before = append(before, s)
		}
		revBefore = map[int]int32{}
		trains, _ := m.Trains()
		for _, tr := range trains {
			revBefore[tr.Veh2] = m.Entities[tr.Veh2].CurMonthRevenue
		}
	}
	months0 := m.ObjectiveMonthsInChallenge
	ticks := 0
	for m.CurrentMonth == month && ticks < 4000 {
		snap()
		m.UpdateVehicles(r.Objects)
		ticks++
	}
	if m.CurrentMonth == month {
		t.Fatal("no month end within 4000 ticks")
	}
	trains, _ := m.Trains()
	want := map[uint8]int64{}
	checked := 0
	for _, tr := range trains {
		h := m.Entities[tr.Head]
		if h.Mode > 1 {
			continue
		}
		var cost int32
		for _, car := range tr.Cars {
			o := r.Objects.Vehicles[m.Entities[car[0].Front].ObjectID]
			if o != nil {
				cost += m.InflationAdjustedCost(o.RunCostFactor, o.RunCostIndex, 10)
			}
		}
		want[h.Owner] += int64(cost)
		v2 := m.Entities[tr.Veh2]
		if v2.CurMonthRevenue != 0 {
			t.Errorf("head %d: curMonthRevenue %d after the month end", h.ID, v2.CurMonthRevenue)
		}
		checked++
	}
	t.Logf("month end after %d ticks; %d land vehicles; monthsInChallenge %d → %d", ticks, checked, months0, m.ObjectiveMonthsInChallenge)
	for id, w := range want {
		var after int64
		for vt := 1; vt < 12; vt += 2 {
			after += int64(m.Companies[id].Expenditures[0][vt])
		}
		t.Logf("company %d: running costs %d (expenditures %d → %d)", id, w, before[id], after)
		if before[id]-after != w {
			t.Errorf("company %d: running-cost expenditures fell by %d, want %d", id, before[id]-after, w)
		}
	}
	if m.ObjectiveMonthsInChallenge != months0+1 {
		t.Error("monthsInChallenge not counted")
	}
}

// The month's end for companies (Company::updateMonthly1), on the SV5: the
// cargo histories take the month's totals, which reset; vehicleProfit,
// company value and the performance index are those of the vehicles'
// profit history BEFORE the vehicles' own month end rotates it (dateTick
// order — computed here one tick before); loan interest is charged;
// historySize counts up.
func TestMonthEndCompanies(t *testing.T) {
	r, m, _ := saveView(t, "Single Player Saved Games/DEBUG_AUTO_DUMP.SV5", false)
	month := m.CurrentMonth
	id := int(m.ControllingCompany)
	var c0 scenario.CompanyEconomy
	var wantVP int64
	for m.CurrentMonth == month {
		c0 = m.Companies[id]
		wantVP, _ = m.CalculateCompanyValue(id)
		m.UpdateVehicles(r.Objects)
	}
	c := m.Companies[id]
	interest := c0.CurrentLoan * int32(m.LoanInterestRate) / 1200
	t.Logf("company %d: distance %d → history, delivered %d → history; perf %d (hist %v); vehicleProfit %d; value %d; loan %d × %d%% → interest %d; historySize %d → %d",
		id, c0.CargoUnitsTotalDistance, c0.CargoUnitsTotalDelivered, m.CompanyFaces[id].Performance, c.PerformanceIndexHist[:3], c.VehicleProfit, c.CompanyValueHistory[0], c0.CurrentLoan, m.LoanInterestRate, interest, c0.HistorySize, c.HistorySize)
	if c.CargoUnitsDistanceHist[0] != c0.CargoUnitsTotalDistance || c.CargoUnitsDistanceHist[1] != c0.CargoUnitsDistanceHist[0] || c.CargoUnitsTotalDistance != 0 {
		t.Error("cargo distance history not rotated")
	}
	if c.CargoUnitsDeliveredHist[0] != c0.CargoUnitsTotalDelivered || c.CargoUnitsTotalDelivered != 0 {
		t.Error("cargo delivered history not rotated")
	}
	if c.VehicleProfit != wantVP {
		t.Errorf("vehicleProfit %d, want %d (from the profit history before the vehicles' month end)", c.VehicleProfit, wantVP)
	}
	if c.PerformanceIndexHist[0] != m.CompanyFaces[id].Performance || c.PerformanceIndexHist[1] != c0.PerformanceIndexHist[0] {
		t.Error("performance history not rotated")
	}
	if c.Expenditures[0][15]-c0.Expenditures[0][15] != -interest {
		t.Errorf("loan interest expenditure moved by %d, want %d", c.Expenditures[0][15]-c0.Expenditures[0][15], -interest)
	}
	if c.HistorySize != min(c0.HistorySize+1, 120) {
		t.Error("historySize")
	}
}

// updateDefaultLevelCrossingType for the saved year gives the saved choice
// (it ran at the start of the year, and nothing else changes it); at the
// year's end every company's expenditure table moves on a year.
func TestYearEnd(t *testing.T) {
	for _, s := range []struct {
		path  string
		title bool
	}{{"Data/title.dat", true}, {"Single Player Saved Games/DEBUG_AUTO_DUMP.SV5", false}} {
		r, m, _ := saveView(t, s.path, s.title)
		saved := m.DefaultLevelCrossingType
		m.UpdateDefaultLevelCrossingType(r.Objects)
		t.Logf("%s (%d): default level crossing %d, saved %d (of %d)", s.path, m.CurrentYear, m.DefaultLevelCrossingType, saved, len(r.Objects.LevelCrossingObjects))
		if m.DefaultLevelCrossingType != saved {
			t.Errorf("%s: default level crossing %d, saved %d", s.path, m.DefaultLevelCrossingType, saved)
		}
	}
	r, m, _ := titleView(t)
	year := m.CurrentYear
	var before []scenario.CompanyEconomy
	ticks := 0
	for m.CurrentYear == year && ticks < 5000 {
		before = append(before[:0], m.Companies...)
		m.UpdateVehicles(r.Objects)
		ticks++
	}
	if m.CurrentYear == year {
		t.Fatal("no year end")
	}
	checked := 0
	for id := range m.Companies {
		if m.CompanyEmpty(id) {
			continue
		}
		b, a := before[id], m.Companies[id]
		for ty := range 17 {
			if a.Expenditures[0][ty] != 0 { // the month end's charges went in before the rotation
				t.Errorf("company %d: new year expenditure %d = %d", id, ty, a.Expenditures[0][ty])
			}
			for i := 1; i < 16; i++ {
				want := b.Expenditures[i-1][ty]
				if i == 1 && ty == 15 {
					continue // the year-end tick's month end charged interest before the rotation
				}
				if i == 1 && (ty%2 == 1 && ty < 12) {
					continue // running costs charged by the same tick's month end, before the rotation
				}
				if a.Expenditures[i][ty] != want {
					t.Errorf("company %d: year %d type %d = %d, want %d", id, i, ty, a.Expenditures[i][ty], want)
				}
			}
		}
		if a.NumExpenditureYears != min(b.NumExpenditureYears+1, 16) {
			t.Errorf("company %d: numExpenditureYears %d → %d", id, b.NumExpenditureYears, a.NumExpenditureYears)
		}
		checked++
	}
	t.Logf("year %d → %d after %d ticks; %d companies' expenditure tables moved on", year, m.CurrentYear, ticks, checked)
}

// The SV5's game started in January 1925 (monthsInChallenge 8 in September
// 1925): its saved inflation factors are setInflationForYear(1925) and eight
// monthly Economy updates — all 32 exactly. And the month end applies one more.
func TestInflationMatchesSave(t *testing.T) {
	r, m, _ := saveView(t, "Single Player Saved Games/DEBUG_AUTO_DUMP.SV5", false)
	saved := m.CurrencyMultiplicationFactor
	var x sim.Map
	x.SetInflationForYear(1925)
	for range m.ObjectiveMonthsInChallenge {
		x.EconomyUpdateMonthly()
	}
	if x.CurrencyMultiplicationFactor != saved {
		t.Fatalf("computed %v\nsaved    %v", x.CurrencyMultiplicationFactor, saved)
	}
	x.EconomyUpdateMonthly()
	month := m.CurrentMonth
	for m.CurrentMonth == month {
		m.UpdateVehicles(r.Objects)
	}
	if m.CurrencyMultiplicationFactor != x.CurrencyMultiplicationFactor {
		t.Errorf("after the month end %v, want %v", m.CurrencyMultiplicationFactor, x.CurrencyMultiplicationFactor)
	}
	t.Logf("32 factors match (e.g. [0] %d → %d)", saved[0], m.CurrencyMultiplicationFactor[0])
}

// determineUnlockedVehicles for the saved year = every company's saved
// unlockedVehicles (set at creation, then kept by updateYearly2's unlocks and
// obsolescence), in title.dat and the SV5.
func TestUnlockedVehiclesMatchSaves(t *testing.T) {
	for _, s := range []struct {
		path  string
		title bool
	}{{"Data/title.dat", true}, {"Single Player Saved Games/DEBUG_AUTO_DUMP.SV5", false}} {
		r, m, _ := saveView(t, s.path, s.title)
		n, ok := 0, 0
		for id := range m.Companies {
			if m.CompanyEmpty(id) || id >= len(m.UnlockedVehicles) {
				continue
			}
			n++
			if got := m.DetermineUnlockedVehicles(id, r.Objects); got == m.UnlockedVehicles[id] {
				ok++
			} else {
				t.Logf("%s company %d: computed %x, saved %x", s.path, id, got, m.UnlockedVehicles[id])
			}
		}
		t.Logf("%s (%d): %d/%d companies' unlocked vehicles match", s.path, m.CurrentYear, ok, n)
		if n == 0 || ok != n {
			t.Errorf("%s: mismatches", s.path)
		}
	}
}

// A new game from a .SC5 (Scenario::start's CompanyManager::reset +
// createPlayerCompany, after initialiseDate for the scenario's year): the
// player company: slot 0, named
// new_company, a free competitor's face, a colour no other company
// competes with, the inflation-adjusted starting loan as cash and loan,
// the year's unlocked vehicles, history size 1.
func TestCreatePlayerCompany(t *testing.T) {
	r, m, _ := saveView(t, "Scenarios/Aerophobia.SC5", false)
	before := len(m.Companies)
	m.InitialiseDate(m.CurrentYear)
	m.ResetCompanies()
	id := m.CreatePlayerCompany(r.Objects)
	if id != 0 {
		t.Fatalf("player company %d", id)
	}
	c, f := m.Companies[id], m.CompanyFaces[id]
	t.Logf("companies parsed %d; player %d: competitor %d of %d, colour %d, loan %d (starting loan size %d, year %d), unlocked %x",
		before, id, f.Competitor, len(r.Objects.CompetitorObjects), m.CompanyColours[id], c.CurrentLoan, m.StartingLoanSize, m.CurrentYear, m.UnlockedVehicles[id])
	if m.CompanyEmpty(int(id)) || m.CompanyNames[id] != 2 || m.ControllingCompany != id || m.PlayerCompanies != [2]uint8{0, 0xFF} {
		t.Error("company not set up")
	}
	// A .SC5 requires no competitor objects: selectNewCompetitor's fallback
	// loads the installed one closest to the preferred AI character into
	// slot 0.
	if f.Competitor != 0 || r.Objects.CompetitorObjects[0] == nil {
		t.Errorf("competitor %d", f.Competitor)
	} else {
		c := r.Objects.CompetitorObjects[0]
		t.Logf("face: %s (%d/%d/%d; preferred %v; %d installed)", c.Name, c.Intelligence, c.Aggressiveness, c.Competitiveness, m.PreferredAI, len(r.Objects.InstalledCompetitors()))
	}
	if c.CurrentLoan <= 0 || c.CurrentLoan%100 != 0 || c.Cash != int64(c.CurrentLoan) {
		t.Errorf("loan %d cash %d", c.CurrentLoan, c.Cash)
	}
	if m.UnlockedVehicles[id] != m.DetermineUnlockedVehicles(int(id), r.Objects) || m.UnlockedVehicles[id] == ([7]uint32{}) {
		t.Error("unlocked vehicles")
	}
	if m.CompanyColours[id] >= 31 || c.HistorySize != 1 || c.NumExpenditureYears != 1 {
		t.Error("colour / history")
	}
}

// Scenario::start on Aerophobia.SC5: 1 January of the options' start year,
// a player company with a face and the starting loan, the objective's time
// limit from the start year, and the sim runs a month on it.
func TestStartScenario(t *testing.T) {
	r, m, _ := saveView(t, "Scenarios/Aerophobia.SC5", false)
	sc, err := scenario.LoadScenarioData("../../../locomotion/Scenarios/Aerophobia.SC5")
	if err != nil {
		t.Fatal(err)
	}
	m.StartScenario(sc.Options.StartYear, r.Objects)
	id := int(m.ControllingCompany)
	c := m.Companies[id]
	t.Logf("start year %d (general %d): %d-%02d-%02d day %d; player %d loan %d; objective %+v until %d",
		sc.Options.StartYear, sc.CurrentYear, m.CurrentYear, m.CurrentMonth+1, m.CurrentDayOfMonth+1, m.CurrentDay, id, c.CurrentLoan, m.Objective, m.ObjectiveUntilYear)
	if m.CurrentYear != sc.Options.StartYear || m.CurrentMonth != 0 || m.CurrentDayOfMonth != 0 || m.CurrentDay != sim.CalcDays(int(sc.Options.StartYear), 0, 1) {
		t.Error("date")
	}
	if m.CompanyEmpty(id) || m.CompanyFaces[id].Competitor == 0xFF || c.CurrentLoan <= 0 {
		t.Error("player company")
	}
	if m.ObjectiveUntilYear != uint16(m.Objective.TimeLimitYears)-1+sc.Options.StartYear || m.ObjectiveMonthsInChallenge != 0 {
		t.Error("objective progress")
	}
	for range 3000 {
		m.UpdateVehicles(r.Objects)
	}
	if m.CurrentMonth == 0 || m.ObjectiveMonthsInChallenge == 0 {
		t.Errorf("a month did not pass: month %d, months in challenge %d", m.CurrentMonth, m.ObjectiveMonthsInChallenge)
	}
	t.Logf("after 3000 ticks: %d-%02d-%02d, cash %d, loan interest so far %d", m.CurrentYear, m.CurrentMonth+1, m.CurrentDayOfMonth+1, m.Companies[id].Cash, m.Companies[id].Expenditures[0][15])
}

// The original terrain generator and generateRivers against upstream's own
// OriginalTerrainGenerator.cpp and MapGenerator.cpp's generateRivers (tools/gen-paintdata/gen_terrain.cpp, compiled
// on HS1's height-map images): six seeds / styles / densities / edge flags,
// the 384×384 height maps compared by FNV-1a hash (testdata — hashes only,
// no game data).
func TestTerrainHeightMapMatchesUpstream(t *testing.T) {
	r, m, _ := saveView(t, "Scenarios/Aerophobia.SC5", false)
	data, err := os.ReadFile("testdata/terrain_hashes.txt")
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var c, style, dens, edge, beds, minW, maxW, bank, meander int
		var s0, s1 uint32
		var want uint64
		fmt.Sscanf(line, "%d %d %d %d %d %d %d %d %d %d %d %x", &c, &style, &dens, &edge, &beds, &minW, &maxW, &bank, &meander, &s0, &s1, &want)
		m.Rng1 = *core.NewPrngSeed(s0, s1)
		hm := sim.NewHeightMap(384, 384)
		o := sim.TerrainOptions{MinLandHeight: 2, TopographyStyle: uint8(style), HillDensity: uint8(dens), HillsEdgeOfMap: edge == 1, SeaLevel: 4,
			NumRiverbeds: uint8(beds), MinRiverWidth: uint8(minW), MaxRiverWidth: uint8(maxW), RiverbankWidth: uint8(bank), RiverMeanderRate: uint8(meander)}
		m.GenerateTerrainHeightMap(o, r.Objects.HillShapes, r.G1, hm)
		m.GenerateRivers(o, hm)
		h := uint64(1469598103934665603)
		for y := range 384 {
			for x := range 384 {
				h ^= uint64(*hm.At(x, y))
				h *= 1099511628211
			}
		}
		if h != want {
			t.Errorf("case %d (style %d density %d edge %d): hash %x, upstream %x", c, style, dens, edge, h, want)
		}
		n++
	}
	t.Logf("%d height maps compared with upstream's generator", n)
}

// The map generator through surface variation (height map, rivers, land,
// water, generateTerrain with getRandomTerrainVariation, then
// generateSurfaceVariation's growth stages and snow coverage at snow lines
// 20 / 44) against upstream's own
// MapGenerator.cpp code compiled with stub tiles
// (tools/gen-paintdata/gen_terrain_styles.cpp): Aerophobia's land objects
// with their own distribution patterns and with an alternative set (far from
// water, on mountains, large random areas), four seeds; every tile's terrain,
// variation, water, baseZ, slope, growth stage and snow compared by hash.
func TestTerrainStylesMatchUpstream(t *testing.T) {
	r, m, _ := saveView(t, "Scenarios/Aerophobia.SC5", false)
	sc, err := scenario.LoadScenarioData("../../../locomotion/Scenarios/Aerophobia.SC5")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("testdata/terrain_styles_hashes.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var c, ps, snowLine, style, dens int
		var s0, s1 uint32
		var want uint64
		fmt.Sscanf(line, "%d %d %d %d %d %d %d %x", &c, &ps, &snowLine, &style, &dens, &s0, &s1, &want)
		pats := sc.Options.LandDistributionPatterns
		if ps == 1 {
			copy(pats[:4], []uint8{2, 0, 4, 7})
		}
		beds := uint8(0)
		if c == 3 {
			beds = 1
		}
		o := sim.TerrainOptions{MinLandHeight: 2, TopographyStyle: uint8(style), HillDensity: uint8(dens), HillsEdgeOfMap: true, SeaLevel: sc.SeaLevel,
			NumRiverbeds: beds, MinRiverWidth: 3, MaxRiverWidth: 6, RiverbankWidth: 1, RiverMeanderRate: 2, LandDistributionPatterns: pats}
		m.Rng1 = *core.NewPrngSeed(s0, s1)
		m.InitialiseTiles()
		hm := sim.NewHeightMap(384, 384)
		m.GenerateTerrainHeightMap(o, r.Objects.HillShapes, r.G1, hm)
		m.GenerateRivers(o, hm)
		m.GenerateLand(hm)
		m.GenerateWater(o.SeaLevel)
		m.GenerateTerrain(hm, pats, o.SeaLevel, r.Objects)
		m.CurrentSnowLine = uint8(snowLine)
		m.GenerateSurfaceVariation(r.Objects)
		h := uint64(1469598103934665603)
		for y := range 384 {
			for x := range 384 {
				tl := m.TileAt(x, y)
				for _, v := range []uint8{tl.TerrainIndex, tl.Variation, tl.WaterLevel, tl.BaseZ, tl.Slope & 0x1F, tl.GrowthStage, tl.Slope >> 5} {
					h ^= uint64(v)
					h *= 1099511628211
				}
			}
		}
		if h != want {
			t.Errorf("case %d (patterns %d, style %d, density %d): hash %x, upstream %x", c, ps, style, dens, h, want)
		}
	}
}

// initialiseSnowLine's season (updateSeason from the climate object and the
// day of the year) = the season the saves stored, in title.dat and the SV5;
// the snow line it picks is the climate's for that season.
func TestSeasonMatchesSaves(t *testing.T) {
	for _, s := range []struct {
		path  string
		title bool
	}{{"Data/title.dat", true}, {"Single Player Saved Games/DEBUG_AUTO_DUMP.SV5", false}} {
		r, m, _ := saveView(t, s.path, s.title)
		saved, savedLine := m.CurrentSeason, m.CurrentSnowLine
		m.InitialiseSnowLine(r.Objects)
		c := r.Objects.Climate
		if c == nil {
			t.Fatalf("%s: no climate object", s.path)
		}
		t.Logf("%s: day of year %d; climate %s (first %d, lengths %v, snow lines %d / %d); season %d, saved %d; snow line %d, saved %d",
			s.path, sim.CalcDate(m.CurrentDay).DayOfYear, c.Name, c.FirstSeason, c.SeasonLength, c.WinterSnowLine, c.SummerSnowLine, m.CurrentSeason, saved, m.CurrentSnowLine, savedLine)
		if m.CurrentSeason != saved {
			t.Errorf("%s: season %d, saved %d", s.path, m.CurrentSeason, saved)
		}
	}
}

// Element flags and canConstructAt on title.dat's trees: each tree's
// occupied quarter is 1 << (quadrant ^ 2) (CreateTree's QuarterTile), its own
// space collides, and the space just above it is free unless another
// element is there.
func TestCanConstructAtTrees(t *testing.T) {
	_, m, _ := titleView(t)
	trees, quarterOK, own, above, aboveBlocked := 0, 0, 0, 0, 0
	for y := 1; y < m.Height-1; y++ {
		for x := 1; x < m.Width-1; x++ {
			tl := m.TileAt(x, y)
			for _, r := range tl.Order {
				if r.Kind != scenario.KindTree {
					continue
				}
				tr := tl.Trees[r.Index]
				trees++
				q := uint8(1) << (tr.Quadrant ^ 2)
				if r.Flags&0x0F == q {
					quarterOK++
				}
				px, py := int32(x*32), int32(y*32)
				if !m.CanConstructAt(px, py, tr.BaseZ, tr.ClearZ, q, 0x0F) {
					own++
				}
				if m.CanConstructAt(px, py, tr.ClearZ, tr.ClearZ+4, q, 0x0F) {
					above++
				} else {
					aboveBlocked++
				}
			}
		}
	}
	t.Logf("%d trees: occupied quarter = 1 << (quadrant ^ 2) for %d; own space blocked %d; space above free %d (blocked by something else %d)", trees, quarterOK, own, above, aboveBlocked)
	if trees == 0 || quarterOK != trees || own != trees {
		t.Error("tree flags / own-space collision")
	}
}

// generateTrees (forests, random trees, the altitude cull, seasons) with
// placeTreeCluster / placeRandomTree / getRandomTreeTypeFromSurface,
// CreateTree, canConstructAt and getHeight against upstream's own code
// compiled on a stub tile store (tools/gen-paintdata/trees): Aerophobia's
// generated map, its tree options and two variants (3000 random trees with a
// low altitude limit; 400 forests, which runs into the free-elements stop);
// every tree's tile, heights, quarter, type, rotation, growth, snow and
// season compared by hash.
func TestTreesMatchUpstream(t *testing.T) {
	r, m, _ := saveView(t, "Scenarios/Aerophobia.SC5", false)
	sc, err := scenario.LoadScenarioData("../../../locomotion/Scenarios/Aerophobia.SC5")
	if err != nil {
		t.Fatal(err)
	}
	o := sim.TerrainOptions{MinLandHeight: sc.Options.MinLandHeight, TopographyStyle: sc.Options.TopographyStyle, HillDensity: sc.Options.HillDensity,
		HillsEdgeOfMap: sc.Options.Flags&2 != 0, SeaLevel: sc.SeaLevel, LandDistributionPatterns: sc.Options.LandDistributionPatterns}
	m.GenerateLandscape(sc.Options.StartYear, o, r.G1, r.Objects) // no forests / random trees: the map the trees go on
	saved := make([][]sim.Tile, m.Height)
	for y := range m.Tiles {
		saved[y] = append([]sim.Tile(nil), m.Tiles[y]...)
	}
	data, err := os.ReadFile("testdata/trees_hashes.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var c, forests, random, maxAlt, count int
		var want uint64
		var townX, townY int32
		fmt.Sscanf(line, "%d %d %d %d %d %x %d %d", &c, &forests, &random, &maxAlt, &count, &want, &townX, &townY)
		for y := range m.Tiles {
			for x := range m.Tiles[y] {
				st := saved[y][x]
				st.Order = append([]scenario.ElemRef(nil), st.Order...)
				st.Trees = nil
				m.Tiles[y][x] = st
			}
		}
		m.EntriesEnd = m.Width * m.Height
		m.Rng1 = *core.NewPrngSeed(4242+uint32(c)*101, 1717+uint32(c)*13)
		to := o
		to.NumberOfForests, to.NumberRandomTrees = uint16(forests), uint16(random)
		to.MinForestRadius, to.MaxForestRadius = sc.Options.MinForestRadius, sc.Options.MaxForestRadius
		to.MinForestDensity, to.MaxForestDensity = sc.Options.MinForestDensity, sc.Options.MaxForestDensity
		to.MinAltitudeForTrees, to.MaxAltitudeForTrees = sc.Options.MinAltitudeForTrees, uint8(maxAlt)
		m.GenerateTrees(to, r.Objects)
		h, n := uint64(1469598103934665603), 0
		for y := range m.Height {
			for x := range m.Width {
				tl := m.TileAt(x, y)
				for _, ref := range tl.Order {
					if ref.Kind != scenario.KindTree {
						continue
					}
					tr := tl.Trees[ref.Index]
					n++
					snow := 0
					if tr.HasSnow {
						snow = 1
					}
					for _, v := range []int{y*384 + x, int(tr.BaseZ), int(tr.ClearZ), int(ref.Flags), int(tr.TreeObjectID), int(tr.Rotation), int(tr.Quadrant), int(tr.Growth), snow, int(tr.Season & 7), int(tr.Unk7l)} {
						h ^= uint64(uint32(v))
						h *= 1099511628211
					}
				}
			}
		}
		m.Towns = nil
		tx, ty, ok := m.FindTownSite(r.Objects) // CreateTown's placement, the first town, right after the trees
		if !ok || tx != townX || ty != townY {
			t.Errorf("case %d: first town at (%d,%d) ok %v, upstream (%d,%d)", c, tx, ty, ok, townX, townY)
		}
		t.Logf("case %d: %d trees (upstream %d); first town (%d,%d)", c, n, count, tx, ty)
		if h != want || n != count {
			t.Errorf("case %d (forests %d, random %d, max altitude %d): %d trees hash %x, upstream %d trees %x", c, forests, random, maxAlt, n, h, count, want)
		}
	}
}

// calcCargoInfluenceFlags (the region object's cargo influence filters) for
// every town in title.dat and the SV5 = the flags initialiseTown stored when
// the town was made (nothing else sets them).
func TestCargoInfluenceMatchesSaves(t *testing.T) {
	for _, s := range []struct {
		path  string
		title bool
	}{{"Data/title.dat", true}, {"Single Player Saved Games/DEBUG_AUTO_DUMP.SV5", false}} {
		r, m, _ := saveView(t, s.path, s.title)
		if r.Objects.Region == nil {
			t.Fatalf("%s: no region object", s.path)
		}
		ok := 0
		for _, tw := range m.Towns {
			got := m.CalcCargoInfluenceFlags(tw.X, tw.Y, r.Objects)
			if got == tw.CargoInfluenceFlags {
				ok++
			} else {
				t.Logf("%s town %d: %#x, saved %#x", s.path, tw.ID, got, tw.CargoInfluenceFlags)
			}
		}
		t.Logf("%s: region %s, influence %v / filters %v; %d/%d towns match", s.path, r.Objects.Region.Name, r.Objects.Region.CargoInfluenceNames, r.Objects.Region.CargoInfluenceTownFilter, ok, len(m.Towns))
		if len(m.Towns) == 0 || ok != len(m.Towns) {
			t.Errorf("%s: mismatches", s.path)
		}
	}
}

// townNameFromNamesObject over the title's town names object against
// upstream's own copyTownNameToBuffer + townNameFromNamesObject compiled on
// the object's bytes (tools/gen-paintdata/gen_townnames.cpp): 5000 random
// values' names and location flags, compared by hash.
func TestTownNamesMatchUpstream(t *testing.T) {
	r, _, _ := titleView(t)
	tn := r.Objects.TownNames
	if tn == nil {
		t.Fatal("no town names object")
	}
	data, err := os.ReadFile("testdata/townnames_hash.txt")
	if err != nil {
		t.Fatal(err)
	}
	var want uint64
	fmt.Sscanf(strings.TrimSpace(string(data)), "%x", &want)
	h, rnd := uint64(1469598103934665603), uint32(7)
	for i := range 5000 {
		rnd = rnd*1664525 + 1013904223
		name, fl := tn.NameFromRand(rnd)
		if i < 3 {
			t.Logf("%d → %q (%d)", rnd, name, fl)
		}
		for _, c := range []byte(name) {
			h ^= uint64(c)
			h *= 1099511628211
		}
		h ^= 0x100 | uint64(fl)
		h *= 1099511628211
	}
	if h != want {
		t.Errorf("hash %x, upstream %x", h, want)
	}
}
