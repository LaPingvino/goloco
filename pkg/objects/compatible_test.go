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
