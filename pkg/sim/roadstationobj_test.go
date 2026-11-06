package sim

import "testing"

// The title save's road stops load their fixed fields: a build cost and a
// (negative: refunded) sell cost; their height is 0 and pieces 0 (straight
// only: RoadCompatibleFlags[0] is 0) in the original data.
func TestRoadStationObjectFields(t *testing.T) {
	_, om := loadDataMap(t, "title.dat")
	n := 0
	for i, s := range om.RoadStationObjects {
		if s == nil {
			continue
		}
		n++
		t.Logf("road station %d %q: height %d pieces %#x build %d sell %d costIndex %d flags %#x", i, s.Name, s.Height, s.RoadPieces, s.BuildCostFactor, s.SellCostFactor, s.CostIndex, s.Flags)
		if s.BuildCostFactor <= 0 || s.SellCostFactor >= 0 || -s.SellCostFactor > s.BuildCostFactor || s.RoadPieces&^0x3F != 0 {
			t.Errorf("road station %d %q: implausible fields", i, s.Name)
		}
	}
	if n == 0 {
		t.Fatal("no road stations")
	}
}
