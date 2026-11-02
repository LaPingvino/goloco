package scene

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// The saved vehicles' motion state agrees with the ported tables: every rail
// and road component that moves along the track (head, vehicle 1 and 2,
// bogies, tail — not bodies, which are placed from their bogies) has its
// current routing equal to its trackAndDirection, and stands exactly at
// tile + getTrackSubPositon / getRoadSubPositon[subPosition]. This checks the
// entity offsets, the routing table's GameState offset and the generated
// SubpositionData together.
func TestSavedVehicleMotionState(t *testing.T) {
	_, m, _ := titleView(t)
	n := 0
	for _, e := range m.Entities {
		if e.EntityType > scenario.VehicleTypeTail || e.EntityType == scenario.VehicleTypeBodyStart || e.EntityType == scenario.VehicleTypeBodyContinued || e.Mode > 1 {
			continue
		}
		n++
		if r := m.Routing(sim.RoutingHandle(e.RoutingHandle)); r&sim.TaDBasicMask != e.TrackAndDirection&sim.TaDBasicMask {
			t.Errorf("entity %d (type %d): routing %#x, trackAndDirection %#x", e.ID, e.EntityType, r, e.TrackAndDirection)
		}
		var p []sim.MoveInfo
		if e.Mode == 0 {
			p = sim.TrackSubPosition(e.TrackAndDirection & sim.TaDBasicMask)
		} else {
			p = sim.RoadSubPosition(e.TrackAndDirection, m.TrafficHandedness != 0)
		}
		if int(e.SubPosition) >= len(p) {
			t.Errorf("entity %d: sub-position %d of %d", e.ID, e.SubPosition, len(p))
			continue
		}
		mi := p[e.SubPosition]
		if int32(mi.X)+int32(e.TileX) != int32(e.PosX) || int32(mi.Y)+int32(e.TileY) != int32(e.PosY) || int32(mi.Z)+int32(e.TileBaseZ)*4 != int32(e.PosZ) {
			t.Errorf("entity %d (type %d): at (%d,%d,%d), the path says %+v from tile (%d,%d,%d)", e.ID, e.EntityType, e.PosX, e.PosY, e.PosZ, mi, e.TileX, e.TileY, e.TileBaseZ)
		}
	}
	if n == 0 {
		t.Fatal("no vehicles")
	}
}
