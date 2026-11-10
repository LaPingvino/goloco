package sim

import (
	"github.com/LaPingvino/goloco/pkg/core"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Town growth's roads (World/Town.cpp: appendToRoadEnd and its helpers),
// item 5.9f3, built with createRoad.

// maxTownBridgeLength is Town.cpp's kMaxTownBridgeLength.
const maxTownBridgeLength = 15

// placeRoadBridge is Town.cpp 0x00498801: from pos in the rotation's
// direction, the bridge types (the road's, available this year) that every
// tile so far takes, tile by tile until the ground comes up to the road
// again (at most 15 tiles); a bridge only over water or at least 8 small
// steps up. One of the types drawn with the town's own prng, the pieces
// built. Whether a bridge was built.
func (m *Map) placeRoadBridge(c *Cmd, town int, x, y, z int32, rotation, roadObjectID uint8) bool {
	if c.Objects == nil {
		return false
	}
	var valid []uint8
	for _, b := range c.Objects.AvailableCompatibleBridges(int(roadObjectID), true, int(m.CurrentYear)) {
		valid = append(valid, uint8(b))
	}
	args := func(bridge uint8, bx, by int32) roadPlacementArgs {
		return roadPlacementArgs{X: bx, Y: by, Z: z, Rotation: rotation, RoadID: 0, Mods: 0, Bridge: bridge, RoadObjectID: roadObjectID}
	}
	off := RotationOffset[rotation&3]
	bx, by := x, y
	length := 1
	overWater, high := false, false
	for ; length <= maxTownBridgeLength; length++ {
		var ok []uint8
		for _, b := range valid {
			if _, fine := m.createRoad(c, args(b, bx, by), 0); fine {
				ok = append(ok, b)
			}
		}
		if len(ok) == 0 {
			return false
		}
		valid = ok
		t := m.TileAt(int(bx/tileSize), int(by/tileSize))
		if t.WaterLevel != 0 {
			overWater = true
		}
		if z-int32(t.BaseZ)*smallZStep >= 8*smallZStep {
			high = true
		}
		bx, by = bx+int32(off[0]), by+int32(off[1])
		if !m.validCoords(bx, by) {
			return false
		}
		if next := m.TileAt(int(bx/tileSize), int(by/tileSize)); z <= int32(next.BaseZ)*smallZStep {
			break
		}
	}
	if !high && !overWater || length > maxTownBridgeLength {
		return false
	}
	tw := &m.Towns[town]
	p := core.NewPrngSeed(tw.Prng[0], tw.Prng[1])
	bridge := valid[int(p.RandNext()&0xFF)*len(valid)/256]
	tw.Prng = [2]uint32{p.Srand0(), p.Srand1()}
	bx, by = x, y
	for range length {
		doCommand(CmdApply, func(f uint8) (int32, bool) { return m.createRoad(c, args(bridge, bx, by), f) })
		bx, by = bx+int32(off[0]), by+int32(off[1])
	}
	return true
}

// buggedRotationOffset is Town.cpp's kBuggedRotationOffset: appendToRoadEnd
// looks at its sides through index 1 + (rotation ± 1) — a CS mistake, kept
// (the intended offsets are kRotationOffset[(rotation ± 1) & 3]).
var buggedRotationOffset = [6][2]int16{{-25, 0}, {-32, 0}, {0, 32}, {32, 0}, {0, -32}, {-32, 0}}

// rotr4bit is Numerics::rotr4bit.
func rotr4bit(v, r uint8) uint8 {
	r &= 3
	v &= 0xF
	return (v>>r | v<<(4-r)) & 0xF
}

// parallelTownRoadAt is Town.cpp's sub_498D21: a straight town road (an
// any-road-type-compatible object, not ghost or AI-allocated) at exactly
// height z on the tile, running along rotation (either way).
func (m *Map) parallelTownRoadAt(x, y, z int32, rotation uint8) bool {
	if !m.validCoords(x, y) {
		return false
	}
	t := m.TileAt(int(x/tileSize), int(y/tileSize))
	for _, r := range t.Order {
		if r.Kind != scenario.KindRoad {
			continue
		}
		e := t.Roads[r.Index]
		if int32(e.BaseZ)*smallZStep != z || m.RoadAnyTypeCompatible&(1<<e.RoadObjectID) == 0 || e.RoadID != 0 || e.Ghost || e.AiAllocated {
			continue
		}
		if rotation == e.Rotation || rotation^2 == e.Rotation {
			return true
		}
	}
	return false
}

// appendToRoadEnd is Town.cpp 0x004986EA: a town road's end grows by one
// piece. Past the second try (not on a bridge) only where at least two
// buildings stand within two tiles; not beside a parallel town road; a bridge
// where the road is above the ground (placeRoadBridge); never on water; the
// ground raised or lowered where the road is a little above or below it, or a
// steep-slope piece where the slope fits; else a straight piece (1% a left
// curve, 1% a right, by the town's prng) — turned to fit a road already there.
func (m *Map) appendToRoadEnd(c *Cmd, town int, x, y, z int32, rotation, roadObjectID uint8, iteration int, onBridge bool) {
	if !m.validCoords(x, y) {
		return
	}
	if iteration >= 2 && !onBridge {
		tx, ty := int(x/tileSize), int(y/tileSize)
		n := 0
		for yy := max(ty-2, 0); yy <= min(ty+2, m.Height-1); yy++ {
			for xx := max(tx-2, 0); xx <= min(tx+2, m.Width-1); xx++ {
				n += len(m.Tiles[yy][xx].Buildings)
			}
		}
		if n < 2 {
			return
		}
	}
	for _, j := range []int{-1, 1} {
		off := buggedRotationOffset[1+int(rotation)+j]
		if m.parallelTownRoadAt(x+int32(off[0]), y+int32(off[1]), z, rotation) {
			return
		}
	}
	t := m.TileAt(int(x/tileSize), int(y/tileSize))
	if z > int32(t.BaseZ)*smallZStep && m.placeRoadBridge(c, town, x, y, z, rotation, roadObjectID) {
		return
	}
	if t.WaterLevel != 0 {
		return
	}
	place := func(roadID uint8) {
		a := roadPlacementArgs{X: x, Y: y, Z: z, Rotation: rotation, RoadID: roadID, Bridge: 0xFF, RoadObjectID: roadObjectID}
		doCommand(CmdApply, func(f uint8) (int32, bool) { return m.createRoad(c, a, f) })
	}
	double := t.Slope&slopeDoubleHeight != 0
	corners := rotr4bit(t.Slope&0xF, rotation)
	heightDiff := int(z/smallZStep) - int(t.BaseZ)
	if heightDiff > 0 {
		if heightDiff == 4 && !double {
			if corners == slopeCornerNorth|slopeCornerEast { // SideUp::northeast
				place(8) // straightSteepSlopeDown
				return
			}
			if corners > 8 { // upstream: 'a mistake by CS', kept
				return
			}
		}
		if heightDiff > 8 {
			return
		}
		m.raiseLandAt(c, x, y)
		return
	}
	if heightDiff < 0 && heightDiff >= -4 {
		m.lowerLandAt(c, x, y)
		return
	}
	if t.Slope&0x1F != 0 && heightDiff == 0 {
		if !double && corners == slopeCornerSouth|slopeCornerWest { // SideUp::southwest
			place(7) // straightSteepSlopeUp
			return
		}
		m.lowerLandAt(c, x, y)
		return
	}
	tw := &m.Towns[town]
	p := core.NewPrngSeed(tw.Prng[0], tw.Prng[1])
	randVar := p.RandNext() & 0xFFFF
	tw.Prng = [2]uint32{p.Srand0(), p.Srand1()}
	var roadID uint8 // straight
	if randVar < 0x028F {
		roadID = 1 // left curve
	} else if randVar < 0x051E {
		roadID = 2 // right curve
	}
	for _, r := range t.Order { // the first road at this height
		if r.Kind != scenario.KindRoad || int32(t.Roads[r.Index].BaseZ)*smallZStep != z {
			continue
		}
		e := t.Roads[r.Index]
		switch roadID {
		case 1:
			if e.RoadID == 1 && e.Rotation^2 == rotation || e.RoadID == 2 && (e.Rotation+1)&3 == rotation {
				roadID = 0
			}
		case 2:
			if e.RoadID == 1 && (e.Rotation-1)&3 == rotation || e.RoadID == 2 && e.Rotation^2 == rotation {
				roadID = 0
			}
		default:
			if e.RoadID == 0 && e.Rotation^2 != rotation {
				roadID = 1
			}
		}
		break
	}
	place(roadID)
}

// raiseLandAt / lowerLandAt are the RaiseLand / LowerLand game commands on
// one whole tile (pointA = pointB = centre, MapSelectionType::full), applied
// (terraform.go).
func (m *Map) raiseLandAt(c *Cmd, x, y int32) {
	tx, ty := int(x/tileSize), int(y/tileSize)
	doCommand(CmdApply, func(f uint8) (int32, bool) { return m.raiseLand(c, tx, ty, tx, ty, mapSelectionFull, f) })
}

func (m *Map) lowerLandAt(c *Cmd, x, y int32) {
	tx, ty := int(x/tileSize), int(y/tileSize)
	doCommand(CmdApply, func(f uint8) (int32, bool) { return m.lowerLand(c, tx, ty, tx, ty, mapSelectionFull, f) })
}

// roadInformation is Town.cpp's RoadInformation (0x0047AC3E's result).
type roadInformation struct {
	Owner            uint8 // CompanyId::null (0xFF) when no matching road was found
	RoadObjectID     uint8
	IsStationRoadEnd bool
	HasStreetLight   bool  // streetLightStyle has a value (no level crossing)
	StreetLightStyle uint8 // RoadElement::streetLightStyle (byte 5 bits 2–3, shared with the level crossing id)
}

// getRoadInformation is Town.cpp 0x0047AC3E: for the road piece at (x, y,
// z) in tad (id << 3 | reversed << 2 | direction) — from its far end when
// reversed —, the matching town road's (any-road-type-compatible, first
// sequence, same piece / rotation / height) owner, road object, street light
// style (when it has no level crossing) and whether a road station there is a
// road end (RoadStationFlags::roadEnd). Owner 0xFF when there is none.
func (m *Map) getRoadInformation(c *Cmd, x, y, z int32, tad uint16) roadInformation {
	if tad&4 != 0 { // reversed: from the piece's other end
		rc := RoadCoordinates[tad&0x7F]
		x, y, z = x+int32(rc.DX), y+int32(rc.DY), z+int32(rc.DZ)
		if rc.RotEnd < 12 {
			x -= int32(RotationOffset[rc.RotEnd][0])
			y -= int32(RotationOffset[rc.RotEnd][1])
		}
	}
	dir, id := uint8(tad&3), uint8(tad>>3&0xF)
	p0 := RoadPieces[id][0]
	ox, oy := rotOffset(p0.X, p0.Y, dir)
	x, y, z = x+int32(ox), y+int32(oy), z+int32(p0.Z)
	none := roadInformation{Owner: 0xFF}
	if !m.validCoords(x, y) {
		return none
	}
	t := m.TileAt(int(x/tileSize), int(y/tileSize))
	for _, r := range t.Order {
		if r.Kind != scenario.KindRoad {
			continue
		}
		e := t.Roads[r.Index]
		if int32(e.BaseZ)*smallZStep != z || e.Rotation != dir || e.SequenceIndex != 0 || e.RoadID != id {
			continue
		}
		if ro := c.roadObject(e.RoadObjectID); ro == nil || ro.Flags&roadFlagAnyRoadTypeCompatible == 0 {
			continue
		}
		res := roadInformation{Owner: e.Owner, RoadObjectID: e.RoadObjectID}
		if !e.HasLevelCrossing {
			res.HasStreetLight, res.StreetLightStyle = true, e.LevelCrossingObjectID
		}
		if e.HasStationElement {
			if st := t.roadStation(id, dir, uint8(z/smallZStep)); st != nil && c.Objects != nil && int(st.ObjectID) < len(c.Objects.RoadStationObjects) {
				if so := c.Objects.RoadStationObjects[st.ObjectID]; so != nil && so.Flags&(1<<3) != 0 { // RoadStationFlags::roadEnd
					res.IsStationRoadEnd = true
				}
			}
		}
		return res
	}
	return none
}

// TownGrowFlags (World/Town.h).
const (
	TownGrowBuildInitialRoad    = 1 << 0
	TownGrowRoadUpdate          = 1 << 1
	TownGrowNeutralRoadTakeover = 1 << 2
	TownGrowAllowRoadExpansion  = 1 << 3
	TownGrowAllowRoadBranching  = 1 << 4
	TownGrowConstructBuildings  = 1 << 5
	TownGrowBuildImmediately    = 1 << 6
	TownGrowAlwaysUpdate        = 1 << 7
)

// buildInitialRoad is Town.cpp 0x00498101: over the square around the town,
// on the first tile where it works, a straight town road (the ideal type) on
// top of the surface, in a direction the town's prng draws.
func (m *Map) buildInitialRoad(c *Cmd, town int) {
	tw := &m.Towns[town]
	m.squareSearch(int32(tw.X), int32(tw.Y), func(x, y int32) bool {
		t := m.TileAt(int(x/tileSize), int(y/tileSize))
		h := int32(t.BaseZ) * smallZStep
		if t.Slope&0x1F != 0 {
			h += 16
			if t.Slope&slopeDoubleHeight != 0 {
				h += 16
			}
		}
		roadObj, ok := idealTownRoadID(tw.Size, c.Objects)
		if !ok {
			return true
		}
		p := core.NewPrngSeed(tw.Prng[0], tw.Prng[1])
		rot := uint8(p.RandNextBound(3))
		tw.Prng = [2]uint32{p.Srand0(), p.Srand1()}
		a := roadPlacementArgs{X: x, Y: y, Z: h, Rotation: rot, RoadID: 0, Bridge: 0xFF, RoadObjectID: roadObj}
		_, built := doCommand(CmdApply, func(f uint8) (int32, bool) { return m.createRoad(c, a, f) })
		return !built
	})
}

// GrowTown is Town::grow (Town.cpp 0x00498116) as the neutral company: from
// a random end of one of the town's roads (findRoadExtent), up to 75 steps
// along the road while the closest town is still this one; at a dead end
// (not a station's road end) the road is extended (appendToRoadEnd); where
// it branches, a branch is drawn by the town's prng. No road yet: the first
// one (buildInitialRoad), when growFlags allow.
// constructBuilding runs on each piece off a bridge (constructBuildings);
// after each step addRoadJunction (allowRoadBranching), the walk ending when
// it placed a road. Before that, on each piece, a company's road is taken
// over / a road updated to the town's type / its street lights renewed
// (neutralRoadTakeover, roadUpdate: townRoadUpdate), the walk stopping after
// a takeover or a type update.
func (m *Map) GrowTown(om *objects.ObjectManager, town int, growFlags uint8) {
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	ext, ok := m.findRoadExtent(town, om)
	if !ok {
		if growFlags&TownGrowBuildInitialRoad != 0 {
			m.buildInitialRoad(c, town)
		}
		return
	}
	tw := &m.Towns[town]
	p := core.NewPrngSeed(tw.Prng[0], tw.Prng[1])
	save := func() { tw.Prng = [2]uint32{p.Srand0(), p.Srand1()} }
	x, y, z, tad := ext.X, ext.Y, ext.Z, ext.TAD
	if p.RandBool() { // from the piece's other end
		rc := RoadCoordinates[tad&0x7F]
		x, y, z = x+int32(rc.DX), y+int32(rc.DY), z+int32(rc.DZ)
		if rc.RotEnd < 12 {
			x -= int32(RotationOffset[rc.RotEnd][0])
			y -= int32(RotationOffset[rc.RotEnd][1])
		}
		tad ^= 4
	}
	save()
	ideal, ok := idealTownRoadID(tw.Size, om)
	if !ok {
		return
	}
	onBridge := ext.IsBridge
	for i := 0; i < 75; i++ {
		closest, density, ok := m.ClosestTownAndDensity(x, y)
		if !ok || closest != town {
			break
		}
		info := m.getRoadInformation(c, x, y, z, tad)
		if m.townRoadUpdate(om, info, x, y, z, tad, onBridge, ideal, density, growFlags) {
			break
		}
		if growFlags&TownGrowConstructBuildings != 0 && !onBridge {
			p = core.NewPrngSeed(tw.Prng[0], tw.Prng[1])
			m.constructBuilding(c, town, p, x, y, z, tad, growFlags)
			save()
		}
		end, _ := RoadConnectionEnd(x, y, z, tad)
		rc := m.RoadConnections(end.X, end.Y, end.Z, end.NextRotation, CompanyNeutral, 0xFF, 0, 0)
		if len(rc.Connections) == 0 {
			if growFlags&TownGrowAllowRoadExpansion != 0 && !info.IsStationRoadEnd {
				m.appendToRoadEnd(c, town, end.X, end.Y, end.Z, end.NextRotation, ideal, i, onBridge)
			}
			break
		}
		conn := rc.Connections[0]
		if len(rc.Connections) > 1 { // several branches: one drawn
			p = core.NewPrngSeed(tw.Prng[0], tw.Prng[1])
			conn = rc.Connections[int(p.RandNext()&0xFFFF)*len(rc.Connections)/65536]
			save()
		}
		x, y, z = end.X, end.Y, end.Z
		onBridge = conn&TaDHasBridge != 0
		tad = conn & TaDBasicMask
		if growFlags&TownGrowAllowRoadBranching == 0 {
			continue
		}
		p = core.NewPrngSeed(tw.Prng[0], tw.Prng[1])
		added := m.addRoadJunction(c, p, x, y, z, tad, onBridge, ideal, rc.Connections)
		save()
		if added {
			break
		}
	}
}

// buildSpeedToGrowthPerTick is Town::update's kBuildSpeedToGrowthPerTick.
var buildSpeedToGrowthPerTick = [12]uint8{0, 1, 3, 5, 7, 9, 12, 16, 22, 0, 0, 0}

// UpdateTowns is TownManager::update (0x00496B6D), from the game tick after
// dateTick as upstream's tickLogic: every 8 scenario ticks the town
// (ticks / 8) % 0x7F, when it exists, runs Town::update (updateTown).
// PARTIAL: Config townGrowthDisabled (no such option in goloco yet: growth
// always on); the editor never ticks here.
func (m *Map) UpdateTowns(om *objects.ObjectManager) {
	if om == nil || m.ScenarioTicks%8 != 0 {
		return
	}
	id := int(m.ScenarioTicks / 8 % 0x7F)
	if id >= len(m.Towns) || m.Towns[id].Name == 0xFFFF {
		return
	}
	m.updateTown(om, id)
}

// updateTown is Town::update (0x0049742F): the size recalculated, then the
// town grows by its build speed — none or the slowest (and then 7 in 8 by
// gPrng1): one grow that only builds a first road and updates roads; faster:
// that many full grows (roads extended and branched, buildings built).
func (m *Map) updateTown(om *objects.ObjectManager, id int) {
	m.RecalculateTownSize(id)
	growth := buildSpeedToGrowthPerTick[m.Towns[id].BuildSpeed%12]
	if growth == 0 || growth == 1 && m.Rng1.RandNext()&7 != 0 {
		m.GrowTown(om, id, TownGrowBuildInitialRoad|TownGrowRoadUpdate|TownGrowNeutralRoadTakeover)
		return
	}
	for range growth {
		m.GrowTown(om, id, TownGrowBuildInitialRoad|TownGrowRoadUpdate|TownGrowNeutralRoadTakeover|TownGrowAllowRoadExpansion|TownGrowAllowRoadBranching|TownGrowConstructBuildings)
	}
}
