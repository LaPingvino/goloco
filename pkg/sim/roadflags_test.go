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

// staticSet iterates as the compiled sfl set: newest first, repeats kept in place.
func TestStaticSetOrder(t *testing.T) {
	var s staticSet
	for _, v := range []uint8{0x83, 0x81, 0x80, 0x85, 0x82, 0x84, 0x81} {
		s.insert(v)
	}
	want := []uint8{0x84, 0x82, 0x85, 0x80, 0x81, 0x83} // the probe's output
	for i := range want {
		if i >= len(s) || s[i] != want[i] {
			t.Fatalf("order % x, want % x", []byte(s), want)
		}
	}
}

// After a scenario start (sub_4748D4 with the player company), the toolbar's
// last road is a road type and the last railroad an available rail track.
func TestPlayerInfrastructureOptions(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	m.CompanyRecords.Speed[0], m.CompanyRecords.Company[0], m.CompanyRecords.Date[0] = 120, 3, 9999
	m.StartScenario(1950, om)
	if r := m.CompanyRecords; r.Speed != [3]uint16{} || r.Company != [3]uint8{0xFF, 0xFF, 0xFF} || r.Date != [3]uint32{} {
		t.Errorf("records after the start: %+v", r)
	}
	pc := m.PlayerCompanies[0]
	if r := om.GetRoadObjectByIndex(int(m.LastRoadOption &^ (1 << 7))); m.LastRoadOption != 0xFF && (m.LastRoadOption&(1<<7) == 0 || r == nil) {
		t.Errorf("last road option %#x", m.LastRoadOption)
	}
	rails := m.companyAvailableRailTracks(pc, om)
	t.Logf("last road %#x, last railroad %#x, rail tracks % x", m.LastRoadOption, m.LastRailroadOption, rails)
	if len(rails) > 0 && m.LastRailroadOption != rails[0] {
		t.Errorf("last railroad %#x, want %#x", m.LastRailroadOption, rails[0])
	}
}

// companyGetAvailableRoads: roads (bit 7) that aren't trams, then road-type
// rail tracks (no bit 7); every any-road-type non-tram road is in it.
func TestCompanyAvailableRoads(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	m.StartScenario(1950, om)
	got := m.companyAvailableRoads(m.PlayerCompanies[0], om)
	t.Logf("available roads % x", got)
	seenTrack := false
	for _, id := range got {
		if id&(1<<7) != 0 {
			if seenTrack {
				t.Error("a road after the road-type tracks")
			}
			if r := om.GetRoadObjectByIndex(int(id &^ (1 << 7))); r == nil || r.Flags&roadObjFlagIsRail != 0 {
				t.Errorf("entry %#x isn't a non-tram road", id)
			}
		} else {
			seenTrack = true
			if tr := om.GetTrackObjectByIndex(int(id)); tr == nil || tr.Flags&trackObjFlagIsRoad == 0 {
				t.Errorf("entry %#x isn't a road-type track", id)
			}
		}
	}
	for i := range om.RoadObjects {
		r := om.GetRoadObjectByIndex(i)
		if r == nil || r.Flags&roadFlagAnyRoadTypeCompatible == 0 || r.Flags&roadObjFlagIsRail != 0 {
			continue
		}
		found := false
		for _, id := range got {
			found = found || id == uint8(i)|1<<7
		}
		if !found {
			t.Errorf("road %d %q missing", i, r.Name)
		}
	}
}
