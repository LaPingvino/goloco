package objects

import "testing"

func TestAvailableCompatibleSignals(t *testing.T) {
	trk := &TrackObject{Header: hdr("TRACKA"), SignalNames: []string{"SIGOWN"}}
	own := &TrainSignalObject{Header: hdr("SIGOWN"), DesignedYear: 1900, ObsoleteYear: 2100}
	other := &TrainSignalObject{Header: hdr("SIGOTH"), DesignedYear: 1950, ObsoleteYear: 2100, CompatibleNames: []string{"tracka"}}
	old := &TrainSignalObject{Header: hdr("SIGOLD"), DesignedYear: 1850, ObsoleteYear: 1899, CompatibleNames: []string{"TRACKA"}}
	unrelated := &TrainSignalObject{Header: hdr("SIGNONE"), DesignedYear: 1900, ObsoleteYear: 2100}
	m := &ObjectManager{TrackObjects: []*TrackObject{trk}, TrainSignalObjects: []*TrainSignalObject{other, old, nil, own, unrelated}}
	got := m.AvailableCompatibleSignals(0, 1960)
	if len(got) != 2 || got[0] != 0 || got[1] != 3 {
		t.Errorf("1960: %v, want [0 3] (the compatible one and the track's own, sorted)", got)
	}
	if got := m.AvailableCompatibleSignals(0, 1920); len(got) != 1 || got[0] != 3 {
		t.Errorf("1920: %v, want [3] (SIGOTH not designed yet)", got)
	}
}

func hdr(name string) *ObjectHeader {
	h := &ObjectHeader{}
	copy(h.Name[:], name)
	return h
}

func TestAvailableCompatibleStations(t *testing.T) {
	trk := &TrackObject{Header: hdr("TRACKA"), StationNames: []string{"STAOWN"}}
	own := &TrainStationObject{Header: hdr("STAOWN"), DesignedYear: 1900, ObsoleteYear: 2100}
	lists := &TrainStationObject{Header: hdr("STALIST"), DesignedYear: 1900, ObsoleteYear: 1940, CompatibleNames: []string{"TRACKA"}}
	m := &ObjectManager{TrackObjects: []*TrackObject{trk}, TrainStationObjects: []*TrainStationObject{own, nil, lists}}
	if got := m.AvailableCompatibleStations(0, false, 1930); len(got) != 2 || got[0] != 0 || got[1] != 2 {
		t.Errorf("1930: %v, want [0 2]", got)
	}
	if got := m.AvailableCompatibleStations(0, false, 1950); len(got) != 1 || got[0] != 0 {
		t.Errorf("1950: %v, want [0] (STALIST obsolete)", got)
	}
}

func TestAvailableCompatibleMods(t *testing.T) {
	trk := &TrackObject{Header: hdr("TRACKA"), ModNames: []string{"CATENARY", "THIRDRL"}}
	electric := &VehicleObject{Mode: TransportModeRail, TrackName: "TRACKA", ExtraNames: []string{"THIRDRL"}}
	steam := &VehicleObject{Mode: TransportModeRail, TrackName: "TRACKA"}
	tram := &VehicleObject{Mode: TransportModeRail, TrackName: "OTHER", ExtraNames: []string{"CATENARY"}}
	m := &ObjectManager{TrackObjects: []*TrackObject{trk}, Vehicles: []*VehicleObject{electric, steam, tram},
		TrackExtraOrder: []string{"THIRDRL", "CATENARY"}}
	all := func(int) bool { return true }
	if got := m.AvailableCompatibleMods(0, all); got != [4]uint8{0xFF, 0, 0xFF, 0xFF} {
		t.Errorf("all unlocked: %v, want third rail (slot 0) in the track's second mod place", got)
	}
	none := func(int) bool { return false }
	if got := m.AvailableCompatibleMods(0, none); got != [4]uint8{0xFF, 0xFF, 0xFF, 0xFF} {
		t.Errorf("nothing unlocked: %v", got)
	}
}

// Road mods: a tram on this road that needs the overhead wire makes it
// available (in its road-extra slot); a bus needs nothing; locked vehicles
// don't count.
func TestAvailableCompatibleRoadMods(t *testing.T) {
	m := &ObjectManager{
		RoadObjects:    []*RoadObject{{ModNames: []string{"ROADEXT1"}}},
		RoadExtraOrder: []string{"", "ROADEXT1"},
		Vehicles: []*VehicleObject{
			{Mode: TransportModeRoad, ExtraNames: []string{"ROADEXT1"}},
			{Mode: TransportModeRoad},
		},
	}
	m.RoadObjects[0].Header = &ObjectHeader{Name: [8]byte{'R', 'O', 'A', 'D', '1'}}
	m.Vehicles[0].TrackName, m.Vehicles[1].TrackName = "ROAD1", "ROAD1"
	all := func(int) bool { return true }
	if got := m.AvailableCompatibleRoadMods(0, all); got != [4]uint8{1, 0xFF, 0xFF, 0xFF} {
		t.Errorf("with the tram: %v", got)
	}
	busOnly := func(v int) bool { return v == 1 }
	if got := m.AvailableCompatibleRoadMods(0, busOnly); got != [4]uint8{0xFF, 0xFF, 0xFF, 0xFF} {
		t.Errorf("bus only: %v", got)
	}
}

// Bridges: the track's own order, only those designed by this year.
func TestAvailableCompatibleBridges(t *testing.T) {
	hdr := func(n string) *ObjectHeader {
		h := &ObjectHeader{}
		copy(h.Name[:], n)
		return h
	}
	m := &ObjectManager{
		TrackObjects:  []*TrackObject{{BridgeNames: []string{"BRDGSTL", "BRDGWOOD"}}},
		BridgeObjects: []*BridgeObject{{Header: hdr("BRDGWOOD"), DesignedYear: 1900}, {Header: hdr("BRDGSTL"), DesignedYear: 1950}},
	}
	if got := m.AvailableCompatibleBridges(0, false, 1930); len(got) != 1 || got[0] != 0 {
		t.Errorf("1930: %v", got)
	}
	if got := m.AvailableCompatibleBridges(0, false, 1960); len(got) != 2 || got[0] != 1 || got[1] != 0 {
		t.Errorf("1960: %v", got)
	}
}
