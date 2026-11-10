package sim

import (
	"github.com/LaPingvino/goloco/pkg/core"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Town::grow's building side (World/Town.cpp): constructBuilding and its
// helpers getMaxHeightOfNewBuilding, generateNewBuildingArgs and sub_42CEBF.

// townAmenityPer128Buildings is kTownAmenityPer128Buildings (0x004F92A6).
var townAmenityPer128Buildings = [8]uint8{1, 1, 2, 2, 1, 3, 1, 1}

// largeBuildingShift is k4FF704: a large building's start from the tile beside the road, by rotation.
var largeBuildingShift = [4][2]int32{{0, 0}, {0, -32}, {-32, -32}, {-32, 0}}

func variationHeight(obj *objects.BuildingObject, v uint8) int {
	h := 0
	for _, part := range obj.GetBuildingParts(v) {
		if int(part) < len(obj.PartHeights) {
			h += int(obj.PartHeights[part])
		}
	}
	return h
}

// townBuildingCandidates is sub_42CEBF: the (non-misc) building objects for
// this density in their years, of the tile size, an amenity only when wanted,
// with a variation no taller than the target height.
func townBuildingCandidates(om *objects.ObjectManager, year uint16, density uint8, large bool, amenities uint32, target int) []uint8 {
	var out []uint8
	for i, b := range om.BuildingObjects {
		if b == nil || b.Flags&objects.BuildingFlagMiscBuilding != 0 || b.GeneratorFunction&(1<<density) == 0 ||
			year < b.DesignedYear || year > b.ObsoleteYear || large != (b.Flags&objects.BuildingFlagLargeTile != 0) {
			continue
		}
		if b.TownAmenityCategory != 0xFF && amenities&(1<<b.TownAmenityCategory) == 0 {
			continue
		}
		for v := uint8(0); v < b.NumVariations; v++ {
			if variationHeight(b, v) <= target {
				out = append(out, uint8(i))
				break
			}
		}
	}
	return out
}

// buildingMaxHeightClear is getBuildingMaxHeightClearFunc (0x0042DB35):
// trees don't count; a building collides unless updates are allowed and it
// is a built non-HQ older than 30; anything else collides at or below the
// base, else narrows the room above.
func (m *Map) buildingMaxHeightClear(om *objects.ObjectManager, t *Tile, k int, baseZ uint8, minDiff *uint8, allowUpdate bool) ClearResult {
	r := t.Order[k]
	switch r.Kind {
	case scenario.KindTree:
		return ClearNoCollision
	case scenario.KindBuilding:
		if !allowUpdate {
			return ClearCollision
		}
		b := t.Buildings[r.Index]
		if !b.IsConstructed {
			return ClearCollision
		}
		if obj := om.GetBuildingObjectByIndex(int(b.ObjectID)); obj == nil || obj.Flags&objects.BuildingFlagIsHeadquarters != 0 {
			return ClearCollision
		}
		if b.Age > 30 {
			return ClearNoCollision
		}
		return ClearCollision
	}
	diff := int(t.baseZOf(r)) - int(baseZ)
	if diff <= 0 {
		return ClearCollision
	}
	*minDiff = min(*minDiff, uint8(diff))
	return ClearNoCollision
}

// maxHeightOfNewBuilding is getMaxHeightOfNewBuilding (0x0042D9FA): 0 when
// the footprint's surfaces are more than 8 steps off z or something in the
// way stays; else the room above (world units).
func (m *Map) maxHeightOfNewBuilding(om *objects.ObjectManager, x, y, z int32, large, allowUpdate bool) int {
	offs := buildingTileOffsets(large)
	maxDiff := int32(0)
	for _, o := range offs {
		t := m.TileAt(int(x+o[0])>>5, int(y+o[1])>>5)
		if t == nil || x+o[0] < 0 || y+o[1] < 0 {
			continue
		}
		maxDiff = max(maxDiff, abs32(int32(t.BaseZ)*smallZStep-z))
	}
	if maxDiff/smallZStep > 8 {
		return 0
	}
	minClear := uint8(0xFF)
	for _, o := range offs {
		px, py := x+o[0], y+o[1]
		t := m.TileAt(int(px)>>5, int(py)>>5)
		if t == nil || px < 0 || py < 0 {
			return 0
		}
		minZ := uint8(min(int32(t.BaseZ)*smallZStep, z) / smallZStep)
		clear := func(tt *Tile, k int) ClearResult {
			return m.buildingMaxHeightClear(om, tt, k, uint8(z/smallZStep), &minClear, allowUpdate)
		}
		if ok, _ := m.ApplyClearAtStandardHeight(px&^31, py&^31, minZ, 255, 0xF, 0xF, clear); !ok {
			return 0
		}
	}
	return int(minClear) * smallZStep
}

// newBuildingArgs is generateNewBuildingArgs (0x0042CF7C): the closest
// town's wanted amenities (none before 64 buildings), the candidates for its
// density (or one lower), one drawn with the town's prng, a variation that
// fits and a colour from the object's set; nothing when the closest town
// isn't the growing one.
func (m *Map) newBuildingArgs(om *objects.ObjectManager, town int, p *core.Prng, x, y, z int32, target int, rotation uint8, large, immediate bool) (BuildingPlacement, bool) {
	ti, density, ok := m.ClosestTownAndDensity(x, y)
	if !ok {
		return BuildingPlacement{}, false
	}
	tw := &m.Towns[ti]
	var amenities uint32
	factor := (int(tw.NumBuildings) + 64) / 128
	for i, per := range townAmenityPer128Buildings {
		if int(per)*factor > int(tw.AmenityCounts[i]) {
			amenities |= 1 << i
		}
	}
	cands := townBuildingCandidates(om, m.CurrentYear, density, large, amenities, target)
	if len(cands) == 0 {
		if density == 0 {
			return BuildingPlacement{}, false
		}
		if cands = townBuildingCandidates(om, m.CurrentYear, density-1, large, amenities, target); len(cands) == 0 {
			return BuildingPlacement{}, false
		}
	}
	// The draws use the growing town's prng (Town::grow's town; the closest
	// one is only for density and amenities, as upstream).
	id := cands[int(p.RandNext()&0xFFFF)*len(cands)/65536]
	obj := om.BuildingObjects[id]
	var vars []uint8
	for v := uint8(0); v < obj.NumVariations; v++ {
		if variationHeight(obj, v) <= target {
			vars = append(vars, v)
		}
	}
	variation := vars[int(p.RandNext()&0xFFFF)*len(vars)/65536]
	var colours []uint8
	for j := uint8(0); j < 31; j++ { // Colour::max
		if obj.Colours&(1<<j) != 0 {
			colours = append(colours, j)
		}
	}
	colour := uint8(0) // Colour::black
	if len(colours) > 0 {
		colour = colours[int(p.RandNext()&0xFFFF)*len(colours)/65536]
	}
	if ti != town {
		return BuildingPlacement{}, false
	}
	return BuildingPlacement{X: x, Y: y, Z: z, Rotation: rotation, Type: id, Variation: variation, Colour: colour, BuildImmediately: immediate}, true
}

// constructBuilding is Town::grow's constructBuilding (Town.cpp): on a road
// piece not below the surface, one of the piece's beside-positions
// (getRoadUnkNextTo) drawn, a large building one time in two (srand0 bit 31)
// shifted by k4FF704, building updates allowed by the flag or for one large
// in 32, the room there, a building drawn and placed with createBuilding
// (apply), built at once by the flag.
func (m *Map) constructBuilding(c *Cmd, town int, p *core.Prng, x, y, z int32, rad uint16, growFlags uint8) {
	t := m.TileAt(int(x)>>5, int(y)>>5)
	if t == nil || z < int32(t.BaseZ)*smallZStep {
		return
	}
	idx := int(rad&0x7F) / 4
	if idx >= len(RoadNextTo) || len(RoadNextTo[idx]) == 0 {
		return
	}
	next := RoadNextTo[idx]
	nt := next[len(next)*int(p.RandNext()&0xFF)/256]
	dir := uint8(rad & 3)
	ox, oy := rotOffset(nt.X, nt.Y, dir)
	bx, by, bz := x+int32(ox), y+int32(oy), z+int32(nt.Z)
	rot := (nt.Rotation + dir) & 3
	large := p.Srand0()&(1<<31) != 0
	if large {
		bx += largeBuildingShift[rot][0]
		by += largeBuildingShift[rot][1]
	}
	allowUpdate := growFlags&TownGrowAlwaysUpdate != 0 || (large && p.Srand0()&0x7C000000 == 0)
	h := m.maxHeightOfNewBuilding(c.Objects, bx, by, bz, large, allowUpdate)
	if h == 0 {
		return
	}
	a, ok := m.newBuildingArgs(c.Objects, town, p, bx, by, bz, h, rot, large, false)
	if !ok {
		return
	}
	if growFlags&TownGrowBuildImmediately != 0 {
		a.BuildImmediately = true
	}
	m.CreateBuilding(c, CmdApply, a)
}

// intersectionCheck is Town.cpp kIntersectionCheck: the tiles either side
// along an edge's axis (even edges: y, odd: x) that must not already have a
// road reaching that edge.
var intersectionCheck = [2][4][2]int32{
	{{0, -2}, {0, -1}, {0, 1}, {0, 2}},
	{{-2, 0}, {-1, 0}, {1, 0}, {2, 0}},
}

// roadReachesEdge is sub_498D9A (0x00498D9A): a (non-ghost, non-AI) road of
// an any-road-type-compatible object at this height on the tile whose piece
// connects to the edge.
func (m *Map) roadReachesEdge(x, y, z int32, edge uint8) bool {
	if x < 0 || y < 0 {
		return false
	}
	t := m.TileAt(int(x)>>5, int(y)>>5)
	if t == nil {
		return false
	}
	for _, r := range t.Order {
		if r.Kind != scenario.KindRoad || int(r.Index) >= len(t.Roads) {
			continue
		}
		el := t.Roads[r.Index]
		if int32(el.BaseZ)*smallZStep != z || m.RoadAnyTypeCompatible&(1<<el.RoadObjectID) == 0 || el.Ghost || el.AiAllocated {
			continue
		}
		if tad := int(el.RoadID)<<3 | int(el.Rotation&3); tad < len(RoadTadConnectionEdge) && RoadTadConnectionEdge[tad]&(1<<edge) != 0 {
			return true
		}
	}
	return false
}

// addRoadJunction is Town::grow's addRoadJunction (0x004984C5): on a straight
// or very small curve off a bridge, one time in about 40 (randNext & 0xFFFF ≤
// 0x666) and not below the surface, a free edge of the piece is drawn (from
// the same draw's bits 16–23); unless a road already reaches that edge two
// tiles either way, a straight (or a very small curve when the opposite /
// next edge is free) is placed through createRoad towards it. Whether it
// placed one (the walk then stops).
func (m *Map) addRoadJunction(c *Cmd, p *core.Prng, x, y, z int32, rad uint16, onBridge bool, roadObj uint8, conns []uint16) bool {
	if id := rad >> 3 & 0xF; id > 2 || onBridge {
		return false
	}
	rv := p.RandNext()
	if rv&0xFFFF > 0x666 {
		return false
	}
	t := m.TileAt(int(x)>>5, int(y)>>5)
	if t == nil || z < int32(t.BaseZ)*smallZStep {
		return false
	}
	var edges uint8
	for _, cn := range conns {
		if i := int(cn & TaDBasicMask); i < len(RoadTadConnectionEdge) {
			edges |= RoadTadConnectionEdge[i]
		}
	}
	edges ^= 0xF
	n := 0
	for i := range 4 {
		if edges&(1<<i) != 0 {
			n++
		}
	}
	if n == 0 {
		return false
	}
	num := int32((rv>>16)&0xFF) * int32(n) / 256
	i := uint8(0)
	for ; i < 4 && num >= 0; i++ {
		if edges&(1<<i) != 0 {
			num--
		}
	}
	edge := i - 1
	for _, o := range intersectionCheck[edge&1] {
		if m.roadReachesEdge(x+o[0]*tileSize, y+o[1]*tileSize, z, edge) {
			return false
		}
	}
	newRoad := uint8(0)
	if edges&(1<<(edge^2)) != 0 {
		if edges&(1<<((edge-1)&3)) != 0 {
			newRoad = 1
		} else {
			newRoad = 2
		}
	}
	m.createRoad(c, roadPlacementArgs{X: x, Y: y, Z: z, Rotation: edge ^ 2, RoadID: newRoad, Bridge: 0xFF, RoadObjectID: roadObj}, CmdApply)
	return true
}
