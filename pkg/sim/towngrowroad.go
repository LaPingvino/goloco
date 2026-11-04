package sim

import (
	"github.com/LaPingvino/goloco/pkg/core"
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
		m.createRoad(c, args(bridge, bx, by), CmdApply)
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
		m.createRoad(c, roadPlacementArgs{X: x, Y: y, Z: z, Rotation: rotation, RoadID: roadID, Bridge: 0xFF, RoadObjectID: roadObjectID}, CmdApply)
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
	m.raiseLand(c, tx, ty, tx, ty, mapSelectionFull, CmdApply)
}

func (m *Map) lowerLandAt(c *Cmd, x, y int32) {
	tx, ty := int(x/tileSize), int(y/tileSize)
	m.lowerLand(c, tx, ty, tx, ty, mapSelectionFull, CmdApply)
}
