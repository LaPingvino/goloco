package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Town::grow and its helpers (World/Town.cpp), ported piece by piece (item
// 5.9f3).

// townPopulations is Town.cpp's _populations: the capacity each size starts at.
var townPopulations = [5]uint32{0, 150, 500, 2500, 8000}

// RecalculateTownSize is Town::recalculateSize: one size up when the capacity
// reaches the next size's, one down when it is 100 short of this one's.
func (m *Map) RecalculateTownSize(i int) {
	tw := &m.Towns[i]
	size := tw.Size
	if tw.Size < 4 && tw.PopulationCapacity >= townPopulations[tw.Size+1] {
		size++
	} else if tw.Size != 0 && tw.PopulationCapacity+100 < townPopulations[tw.Size] {
		size--
	}
	tw.Size = size
}

const (
	roadFlagAnyRoadTypeCompatible = 1 << 3
)

// idealTownRoadID is getIdealTownRoadId: of the town roads (any-road-type
// compatible, two-way), the one whose target size is the town's, else the
// nearest below (the largest), else the smallest above. ok false when none.
func idealTownRoadID(size uint8, om *objects.ObjectManager) (id uint8, ok bool) {
	var found uint8
	for rid := 0; rid < 8; rid++ {
		if rid >= len(om.RoadObjects) || om.RoadObjects[rid] == nil {
			continue
		}
		r := om.RoadObjects[rid]
		if r.Flags&roadFlagAnyRoadTypeCompatible == 0 || r.Flags&roadFlagIsOneWay != 0 {
			continue
		}
		if ok {
			set := false
			if r.TargetTownSize == size {
				set = true
			} else if r.TargetTownSize < size {
				set = found > size || r.TargetTownSize >= found
			} else {
				set = found > size && r.TargetTownSize < found
			}
			if !set {
				continue
			}
		}
		id, found, ok = uint8(rid), r.TargetTownSize, true
	}
	return id, ok
}

// squareSearchRange9 is kSquareSearchRange<9>: 100 tile offsets in squares of
// growing size around the centre (left, up, right, down along each ring).
var squareSearchRange9 = func() (r [100][2]int) {
	dirs := [4][2]int{{-1, 0}, {0, 1}, {1, 0}, {0, -1}}
	x, y, k := 0, 0, 0
	for i := 1; i <= 9; i += 2 {
		for d := range 4 {
			for j := i; j != 0; j-- {
				r[k] = [2]int{x, y}
				k++
				x += dirs[d][0]
				y += dirs[d][1]
			}
		}
		x, y = x+1, y-1
	}
	return
}()

// squareSearch is Town.cpp's squareSearch (size 9): pred on each world
// position of the range that is on the map, until it returns false.
func (m *Map) squareSearch(cx, cy int32, pred func(x, y int32) bool) {
	for _, o := range squareSearchRange9 {
		x, y := cx+int32(o[0])*tileSize, cy+int32(o[1])*tileSize
		if x < 0 || y < 0 || x >= int32(m.Width)*tileSize || y >= int32(m.Height)*tileSize {
			continue
		}
		if !pred(x, y) {
			return
		}
	}
}

// roadExtent is findRoadExtent's result: where a road piece starts (its base
// minus the piece's first z), its road id << 3 | rotation, and whether it is on
// a bridge.
type roadExtent struct {
	X, Y, Z  int32
	TAD      uint16
	IsBridge bool
}

// findRoadExtent is Town.cpp's findRoadExtent: over the square around the
// town, the first tile of a town-compatible road piece (not ghost, not AI
// allocated) above the surface on each tile; once one is found each later
// one replaces it on a 0 bit of the town's srand0, read and rotated per
// candidate.
func (m *Map) findRoadExtent(i int, om *objects.ObjectManager) (roadExtent, bool) {
	tw := &m.Towns[i]
	randVal := tw.Prng[0]
	var found *scenario.RoadElement
	var fx, fy int32
	m.squareSearch(int32(tw.X), int32(tw.Y), func(x, y int32) bool {
		t := m.TileAt(int(x/tileSize), int(y/tileSize))
		if t == nil {
			return true
		}
		passed := false
		for _, r := range t.Order {
			if r.Kind == scenario.KindSurface {
				passed = true
				continue
			}
			if !passed || r.Kind != scenario.KindRoad {
				continue
			}
			e := &t.Roads[r.Index]
			if e.Ghost || e.AiAllocated || e.SequenceIndex != 0 {
				continue
			}
			if int(e.RoadObjectID) >= len(om.RoadObjects) || om.RoadObjects[e.RoadObjectID] == nil || om.RoadObjects[e.RoadObjectID].Flags&roadFlagAnyRoadTypeCompatible == 0 {
				continue
			}
			if found != nil {
				bit := randVal & 1
				randVal = randVal>>1 | randVal<<31
				if bit != 0 {
					return true
				}
			}
			found, fx, fy = e, x, y
			return true
		}
		return true
	})
	if found == nil {
		return roadExtent{}, false
	}
	z := int32(found.BaseZ)*4 - int32(RoadPieceTiles[found.RoadID][0][2])
	return roadExtent{X: fx, Y: fy, Z: z, TAD: uint16(found.RoadID)<<3 | uint16(found.Rotation), IsBridge: found.HasBridge}, true
}
