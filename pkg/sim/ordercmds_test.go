package sim

import "testing"

// OrderRouteWaypointRaw writes what the getters read (Orders.cpp): a tile past
// 255 on both axes, a height, a direction and a track id past 31.
func TestOrderRouteWaypointRaw(t *testing.T) {
	m := &Map{}
	m.Orders = OrderRouteWaypointRaw(300, 257, 0x2A, 5, 40)
	o := Order{m, 0}
	x, y, z := o.Waypoint()
	if o.Type() != OrderRouteWaypoint || x != 300*32 || y != 257*32 || z != 0x2A*8 || o.Direction() != 5 || o.TrackID() != 40 {
		t.Errorf("type %d, %d,%d,%d, dir %d, id %d", o.Type(), x, y, z, o.Direction(), o.TrackID())
	}
}
