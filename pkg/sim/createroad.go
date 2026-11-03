package sim

// createRoad (GameCommands/Road/CreateRoad.cpp 594–1105), item 5.9f3 — the
// game command towns grow their roads with. Ported in parts; this is its
// start: the arguments, the road extras that fit the piece, and the cost.

// roadPlacementArgs is RoadPlacementArgs.
type roadPlacementArgs struct {
	X, Y, Z      int32
	Rotation     uint8 // 0–3
	RoadID       uint8 // 0–15
	Mods         uint8 // the road extras asked for (bits 0, 1)
	Bridge       uint8 // a bridge object slot, 0xFF none
	RoadObjectID uint8
	UnkFlags     uint8
}

// createRoadCost is createRoad's start: the road extras asked for that fit
// this piece (the piece's compatibleFlags all in the extra's roadPieces),
// whether the height is valid (a multiple of 16), and the cost — the road's
// and each valid extra's inflation-adjusted build cost (divisor 10) times the
// piece's cost factor / 256.
func (m *Map) createRoadCost(c *Cmd, a roadPlacementArgs) (validMods uint8, cost int32, ok bool) {
	road := c.roadObject(a.RoadObjectID)
	if road == nil || int(a.RoadID) >= len(RoadCompatibleFlags) {
		return 0, 0, false
	}
	compat := RoadCompatibleFlags[a.RoadID]
	validMods = a.Mods
	extra := func(i int) (pieces uint16, buildCost int16, costIndex uint8, found bool) {
		if i >= len(road.ModNames) || c.Objects == nil {
			return 0, 0, 0, false
		}
		if ex := c.Objects.Extra(road.ModNames[i]); ex != nil {
			return ex.Pieces, ex.BuildCostFactor, ex.CostIndex, true
		}
		return 0, 0, 0, false
	}
	for i := range 2 {
		if a.Mods&(1<<i) == 0 {
			continue
		}
		if pieces, _, _, found := extra(i); !found || pieces&compat != compat {
			validMods &^= 1 << i
		}
	}
	if a.Z&0xF != 0 {
		return validMods, 0, false
	}
	factor := int32(RoadCostFactor[a.RoadID])
	cost = m.InflationAdjustedCost(road.BuildCostFactor, road.CostIndex, 10) * factor / 256
	for i := range 2 {
		if validMods&(1<<i) != 0 {
			_, bc, ci, _ := extra(i)
			cost += m.InflationAdjustedCost(bc, ci, 10) * factor / 256
		}
	}
	return validMods, cost, true
}
