package sim

import "github.com/LaPingvino/goloco/pkg/core"

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
