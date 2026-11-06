package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// Boats' pathfinding (AW.a): VehicleHead.cpp's waterPathfind (0x00427FC9)
// and its helpers findNearbyTilesWithBoats (0x00427F1C),
// getDockTargetFromStation (0x00428379) and waterPathfindToTarget
// (0x00428237) — a depth-limited search (7 steps) over the tiles of the same
// water level, scored by the distance to the target tile, choosing which of
// the four neighbouring tiles to head for next.

// WaterPathingResult is the next point for the boat's head, and — when it is
// a dock's boat position — that dock's station and tile.
type WaterPathingResult struct {
	HeadX, HeadY int32
	StationID    uint16 // stationIDNull: no dock; the station fields unused
	StationX     int32
	StationY     int32
	StationZ     int32
}

// nearbyBoats is NearbyBoats: which of the 16×16 tiles from startTile hold
// a boat's first body.
type nearbyBoats struct {
	found            [16][16]bool
	startTX, startTY int
}

// findNearbyTilesWithBoats marks the tiles within 7 of (px, py)'s tile (the
// range clamped to the map) holding a boat's first body (by its sprite
// position — EntityTileList).
func (m *Map) findNearbyTilesWithBoats(px, py int32) nearbyBoats {
	tx, ty := int(px)>>5, int(py)>>5
	res := nearbyBoats{startTX: tx - 7, startTY: ty - 7}
	x0, y0 := max(tx-7, 0), max(ty-7, 0)
	x1, y1 := min(tx+7, m.Width-1), min(ty+7, m.Height-1)
	for i := range m.Entities {
		e := &m.Entities[i]
		if e.EntityType != scenario.VehicleTypeBodyStart || e.Mode != 3 || e.PosX == locationNull {
			continue
		}
		ex, ey := int(e.PosX)>>5, int(e.PosY)>>5
		if ex < x0 || ex > x1 || ey < y0 || ey > y1 {
			continue
		}
		res.found[ex-res.startTX][ey-res.startTY] = true
	}
	return res
}

// dockTarget is getDockTargetFromStation: the station's first free dock
// (not ghost or AI, flag_6 clear) at its recorded height — its boat position
// (+ 32, 32) and the dock tile; ok false: none.
func (m *Map) dockTarget(stationID uint16, docks func(obj uint8) (int16, int16, bool)) (WaterPathingResult, bool) {
	if int(stationID) >= len(m.StationNames) {
		return WaterPathingResult{}, false
	}
	st := &m.StationNames[stationID]
	for i := 0; i < int(st.TileSize) && i < len(st.Tiles); i++ {
		pos := st.Tiles[i]
		x, y, z := int32(pos[0]), int32(pos[1]), int32(pos[2])
		t := m.TileAt(int(x)>>5, int(y)>>5)
		if t == nil {
			continue
		}
		var el *scenario.StationElement
		for j := range t.Stations {
			s := &t.Stations[j]
			if s.Ghost || s.AiAllocated || int32(s.BaseZ) != z/4 {
				continue
			}
			el = s
			break
		}
		if el == nil || el.StationType != 3 || el.Flag6 {
			continue
		}
		bx, by, ok := docks(el.ObjectID)
		if !ok {
			continue
		}
		ox, oy := rotOffset(bx, by, el.Rotation)
		return WaterPathingResult{HeadX: int32(ox) + x + 32, HeadY: int32(oy) + y + 32, StationID: stationID,
			StationX: x, StationY: y, StationZ: z &^ 3}, true // Numerics::floor2(z, 4)
	}
	return WaterPathingResult{}, false
}

// pathScore is PathFindingResult: lower is better, compared score first.
type pathScore struct {
	score uint16
	cost  uint8
}

func (a pathScore) less(b pathScore) bool {
	return a.score < b.score || a.score == b.score && a.cost < b.cost
}

// waterPathfindToTarget is the search: a tile of another water level (or
// with something standing within one micro step above that water), off the
// map, or with a boat (when inside the 16×16 search area — upstream's range
// check kept), ends the branch; else its score (the larger axis distance to
// the target plus a sixteenth of the smaller, in world units) is kept if
// better, and while not there and under 7 steps the four neighbours follow.
func (m *Map) waterPathfindToTarget(tx, ty int, waterMicroZ uint8, targetTX, targetTY int, nearby *nearbyBoats, cost uint8, best pathScore) pathScore {
	if tx < 0 || ty < 0 || tx >= m.Width || ty >= m.Height { // validCoords
		return best
	}
	t := m.TileAt(tx, ty)
	if t == nil || t.WaterLevel != waterMicroZ {
		return best
	}
	for i, r := range t.Order { // the element after the surface
		if r.Kind != scenario.KindSurface || i+1 >= len(t.Order) {
			continue
		}
		next := t.Order[i+1]
		if next.Flags&0x30 == 0 && int(t.baseZOf(next))/4-int(waterMicroZ) < 1 {
			return best
		}
		break
	}
	if nx, ny := tx-nearby.startTX, ty-nearby.startTY; nx >= 0 && nx < 16 && ny >= 0 && ny < 16 && nearby.found[nx][ny] {
		return best
	}
	dx, dy := abs32(int32(tx-targetTX)*32), abs32(int32(ty-targetTY)*32)
	score := uint16(max(dx, dy) + min(dx, dy)/16)
	if r := (pathScore{score, cost}); r.less(best) {
		best = r
	}
	if score != 0 {
		if cost >= 7 {
			return best
		}
		cost++
		for i := 0; i < 4; i++ {
			best = m.waterPathfindToTarget(tx+int(RotationOffset[i][0])/32, ty+int(RotationOffset[i][1])/32, waterMicroZ, targetTX, targetTY, nearby, cost, best)
		}
	}
	return best
}

// waterPathStep is waterPathfind's choice from the boat at (px, py, pz)
// facing yaw towards the target tile: the neighbour with the best search
// result (ties to the one it faces), its centre — or the dock's boat position
// when the dock is reached (a score and cost of 0); nowhere to go, the
// centre of its own tile.
func (m *Map) waterPathStep(px, py, pz int32, yaw uint8, targetTX, targetTY int, dock *WaterPathingResult) WaterPathingResult {
	nearby := m.findNearbyTilesWithBoats(px, py)
	curRotation := ((yaw + 7) >> 4) & 3
	tx, ty := int(px)>>5, int(py)>>5
	waterMicroZ := uint8(pz / 16) // kMicroZStep
	none := pathScore{0xFFFF, 0xFF}
	best, bestDir := none, -1
	for i := 0; i < 4; i++ {
		r := m.waterPathfindToTarget(tx+int(RotationOffset[i][0])/32, ty+int(RotationOffset[i][1])/32, waterMicroZ, targetTX, targetTY, &nearby, 0, none)
		if r != none && (r.less(best) || r == best && uint8(i) == curRotation) {
			best, bestDir = r, i
		}
	}
	if bestDir < 0 {
		return WaterPathingResult{HeadX: int32(tx)*32 + 16, HeadY: int32(ty)*32 + 16, StationID: stationIDNull}
	}
	if dock != nil && best == (pathScore{0, 0}) {
		return *dock
	}
	return WaterPathingResult{HeadX: int32(tx)*32 + int32(RotationOffset[bestDir][0]) + 16, HeadY: int32(ty)*32 + int32(RotationOffset[bestDir][1]) + 16, StationID: stationIDNull}
}

// WaterPathfind is waterPathfind for a boat: the target from its current
// order — a route waypoint's tile, a stop's free dock (its boat position's
// tile) or else the station's own tile — or, with neither, where it is; then
// waterPathStep from the head's position, Vehicle2's height and yaw.
func (m *Map) WaterPathfind(tr Train, docks func(obj uint8) (int16, int16, bool)) WaterPathingResult {
	head := &m.Entities[tr.Head]
	v2 := &m.Entities[tr.Veh2]
	targetTX, targetTY := int(head.PosX)>>5, int(head.PosY)>>5
	var dock *WaterPathingResult
	if orders := m.OrderRing(head.OrderTableOffset, head.CurrentOrder); len(orders) > 0 {
		switch o := orders[0]; o.Type() {
		case OrderRouteWaypoint:
			x, y, _ := o.Waypoint()
			targetTX, targetTY = int(x)>>5, int(y)>>5
		case OrderStopAt, OrderRouteThrough:
			sid := uint16(o.Station())
			if d, ok := m.dockTarget(sid, docks); ok {
				dock = &d
				targetTX, targetTY = int(d.HeadX)>>5, int(d.HeadY)>>5
			} else if int(sid) < len(m.StationNames) {
				targetTX, targetTY = int(m.StationNames[sid].X)>>5, int(m.StationNames[sid].Y)>>5
			}
		}
	}
	return m.waterPathStep(int32(head.PosX), int32(head.PosY), int32(v2.PosZ), v2.SpriteYaw, targetTX, targetTY, dock)
}

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}
