package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// The road network walk (Vehicles/Routing.cpp findAllUsableRoadPieces,
// findAllUsableRoadInNetwork, findAllRoadsFilterTransform), the road twin of
// network.go's track walk.

// RaDBasicMask is AdditionalTaDFlags::basicRaDMask.
const RaDBasicMask uint16 = 0b0111_1111

// roadEndTile is the tile a road piece's far end sits on: start +
// getUnkRoad.pos, minus kRotationOffset[rotationEnd] for non-diagonal ends.
func roadEndTile(x, y, z int32, rad uint16) (int32, int32, int32) {
	c := RoadCoordinates[rad&RaDBasicMask]
	x, y, z = x+int32(c.DX), y+int32(c.DY), z+int32(c.DZ)
	if c.RotEnd < 12 {
		x -= int32(RotationOffset[c.RotEnd][0])
		y -= int32(RotationOffset[c.RotEnd][1])
	}
	return x, y, z
}

// findAllUsableRoadPieces (0x00479EFA): with SearchUnk2, every (non-ghost,
// non-AI) road sharing a tile edge with a tile of the interest's piece is
// added, forwards and reversed.
func (m *Map) findAllUsableRoadPieces(queue *[]LocationOfInterest, flags uint16, in LocationOfInterest, filter func(LocationOfInterest) bool, res *RoutingResults) {
	if flags&SearchUnk2 == 0 {
		return
	}
	rad := in.TaD()
	id, dir := rad>>3&0xF, uint8(rad&3)
	nx, ny, nz := in.X, in.Y, in.Z
	if rad&4 != 0 {
		nx, ny, nz = roadEndTile(nx, ny, nz, rad)
	}
	if int(id) >= len(RoadPieces) {
		return
	}
	for _, piece := range RoadPieces[id] {
		connect := piece.ConnectFlags[dir]
		ox, oy := rotOffset(piece.X, piece.Y, dir)
		px, py, pz := nx+int32(ox), ny+int32(oy), nz+int32(piece.Z)
		t := m.TileAt(int(px/tileSize), int(py/tileSize))
		if t == nil {
			continue
		}
		for _, r := range t.Order {
			if r.Kind != scenario.KindRoad || int(r.Index) >= len(t.Roads) {
				continue
			}
			el := t.Roads[r.Index]
			if int32(el.BaseZ) != pz/4 || el.AiAllocated || el.Ghost || int(el.RoadID) >= len(RoadPieces) || int(el.SequenceIndex) >= len(RoadPieces[el.RoadID]) {
				continue
			}
			target := RoadPieces[el.RoadID][el.SequenceIndex]
			if target.ConnectFlags[el.Rotation&3]&connect == 0 {
				continue
			}
			if el.Rotation == dir && el.SequenceIndex == piece.Index && el.RoadObjectID == in.TrackType && uint16(el.RoadID) == id {
				continue
			}
			tx, ty := rotOffset(target.X, target.Y, el.Rotation)
			sx, sy, sz := px-int32(tx), py-int32(ty), int32(el.BaseZ)*smallZStep-int32(target.Z)
			rad2 := uint16(el.RoadID)<<3 | uint16(el.Rotation&3)
			l := LocationOfInterest{sx, sy, sz, rad2, el.Owner, el.RoadObjectID}
			if res.Reachable.TryAdd(l) && !filter(l) {
				m.findAllUsableRoadPieces(queue, flags, l, filter, res)
				*queue = append(*queue, l)
			}
			ex, ey, ez := roadEndTile(sx, sy, sz, rad2)
			lr := LocationOfInterest{ex, ey, ez, rad2 ^ 4, el.Owner, el.RoadObjectID}
			if res.Reachable.TryAdd(lr) && !filter(lr) {
				m.findAllUsableRoadPieces(queue, flags, lr, filter, res)
				*queue = append(*queue, lr)
			}
		}
	}
}

// findAllUsableRoadInNetwork (0x00479DB1): the roads the piece's end leads
// onto (none: a dead end) and, unless excluded, those its start leads onto.
func (m *Map) findAllUsableRoadInNetwork(queue *[]LocationOfInterest, flags uint16, in LocationOfInterest, filter func(LocationOfInterest) bool, res *RoutingResults) {
	visit := func(x, y, z int32, conns []uint16) {
		for _, c := range conns {
			l := LocationOfInterest{x, y, z, c & (RaDBasicMask | TaDHasSignal), in.Company, in.TrackType}
			if res.Reachable.TryAdd(l) && !filter(l) {
				m.findAllUsableRoadPieces(queue, flags, l, filter, res)
				*queue = append(*queue, l)
			}
		}
	}
	e, ok := RoadConnectionEnd(in.X, in.Y, in.Z, in.TaD()&RaDBasicMask)
	if !ok {
		return
	}
	tc := m.RoadConnections(e.X, e.Y, e.Z, e.NextRotation, in.Company, in.TrackType, 0, 0)
	if len(tc.Connections) > 0 {
		visit(e.X, e.Y, e.Z, tc.Connections)
	} else {
		res.HasDeadEnd = true
	}
	if flags&SearchExcludeReverseDirection == 0 {
		x, y, z := roadEndTile(in.X, in.Y, in.Z, in.TaD())
		rot := ReverseRotation[RoadCoordinates[in.TaD()&RaDBasicMask].RotEnd]
		visit(x, y, z, m.RoadConnections(x, y, z, rot, in.Company, in.TrackType, 0, 0).Connections)
	}
}

// FindAllRoadsFilterTransform: the stack-driven walk from (loc, roadAndDirection).
func (m *Map) FindAllRoadsFilterTransform(res *RoutingResults, flags uint16, x, y, z int32, rad uint16, company, roadType uint8, filter func(LocationOfInterest) bool, transform func(*RoutingResults)) {
	stack := []LocationOfInterest{{x, y, z, rad, company, roadType}}
	for len(stack) > 0 {
		in := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		m.findAllUsableRoadInNetwork(&stack, flags, in, filter, res)
	}
	if transform != nil {
		transform(res)
	}
}
