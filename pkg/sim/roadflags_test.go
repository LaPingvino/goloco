package sim

import "testing"

// sub_4748D4's refreshers on the title objects: the any-road-type flags match
// the objects, the last track type option is a two-way town road, the
// default land is the first "everywhere" land.
func TestScenarioObjectOptions(t *testing.T) {
	_, om := loadDataMap(t, "title.dat")
	m := &Map{}
	m.UpdateRoadObjectIdFlags(om)
	m.UpdateScenarioObjectOptions(om)
	for i := range om.RoadObjects {
		r := om.GetRoadObjectByIndex(i)
		if r == nil {
			continue
		}
		if got, want := m.RoadAnyTypeCompatible&(1<<i) != 0, r.Flags&roadFlagAnyRoadTypeCompatible != 0; got != want {
			t.Errorf("road %d %q: any-type bit %v, flag %v", i, r.Name, got, want)
		}
	}
	if r := om.GetRoadObjectByIndex(int(m.LastTrackTypeOption)); r == nil || r.Flags&roadFlagIsOneWay != 0 || r.Flags&roadFlagAnyRoadTypeCompatible == 0 {
		t.Errorf("last track type option %d", m.LastTrackTypeOption)
	} else {
		t.Logf("last track type option: %d %q", m.LastTrackTypeOption, r.Name)
	}
	if l := om.GetLandObjectByIndex(int(m.LastLandOption)); l == nil || l.DistributionPattern != 0 {
		t.Errorf("default land %d", m.LastLandOption)
	}
}
