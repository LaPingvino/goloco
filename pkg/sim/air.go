package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Aircraft (AW.b): VehicleHead.cpp's airport movement — the airport
// object's movement graph (objects/airport.go) placed at the station's
// airport. PARTIAL: VehicleHead::updateAir and the rest of AW.b.

// airportStationElement is the station element of the airport at the
// station's airportStartPos tile and height (any station element there, as
// upstream takes the first at that baseZ).
func (m *Map) airportStationElement(stationID uint16) (*scenario.StationElement, [3]int16, bool) {
	if int(stationID) >= len(m.StationNames) {
		return nil, [3]int16{}, false
	}
	pos := m.StationNames[stationID].AirportStartPos
	t := m.TileAt(int(pos[0])>>5, int(pos[1])>>5)
	if t == nil {
		return nil, pos, false
	}
	for i := range t.Stations {
		if int16(t.Stations[i].BaseZ) == pos[2]/4 {
			return &t.Stations[i], pos, true
		}
	}
	return nil, pos, false
}

// AirportGetMovementEdgeTarget is airportGetMovementEdgeTarget (0x00426E26):
// the flags and world position of the node edge curEdge leads to, at the
// target station's airport — the node rotated about the tile's centre by
// the airport's rotation, from its start position; on the ground (taxiing)
// at the node's height above the airport, else at the airport's height
// + 255 in flight, or 960 (above everything) for the nodes neither taxiing
// nor in flight. ok false: no airport there (upstream asserts).
func (m *Map) AirportGetMovementEdgeTarget(om *objects.ObjectManager, stationID uint16, curEdge uint8) (flags uint16, x, y, z int32, ok bool) {
	el, pos, found := m.airportStationElement(stationID)
	if !found || om == nil || int(el.ObjectID) >= len(om.AirportObjects) || om.AirportObjects[el.ObjectID] == nil {
		return 0, 0, 0, 0, false
	}
	a := om.AirportObjects[el.ObjectID]
	if int(curEdge) >= len(a.MovementEdges) || int(a.MovementEdges[curEdge].NextNode) >= len(a.MovementNodes) {
		return 0, 0, 0, 0, false
	}
	node := a.MovementNodes[a.MovementEdges[curEdge].NextNode]
	rx, ry := rotOffset(node.X-16, node.Y-16, el.Rotation)
	x, y = int32(rx)+16+int32(pos[0]), int32(ry)+16+int32(pos[1])
	z = int32(node.Z) + int32(pos[2])
	if node.Flags&objects.AirportNodeTaxiing == 0 {
		z = int32(pos[2]) + 255
		if node.Flags&objects.AirportNodeInFlight == 0 {
			z = 960
		}
	}
	return node.Flags, x, y, z, true
}
