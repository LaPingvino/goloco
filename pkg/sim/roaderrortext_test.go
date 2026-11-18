package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// A refused road says why (CreateRoad.cpp's setErrorText sites, with
// FormatArguments::common(name) where upstream pushes one): clearRoad's,
// clearTrack's and createRoad's own.
func TestRoadErrorTexts(t *testing.T) {
	hdr := func(name string) *objects.ObjectHeader {
		h := &objects.ObjectHeader{}
		copy(h.Name[:], name)
		return h
	}
	town := &objects.RoadObject{Header: hdr("ROADTOWN"), Name: "Town road", RoadPieces: roadTraitJunction}
	tram := &objects.RoadObject{Header: hdr("TRAMWAY"), Name: "Tramway", CompatibleNames: []string{"ROADTOWN"}}
	dirt := &objects.RoadObject{Header: hdr("DIRT"), Name: "Dirt road"}
	rail := &objects.TrackObject{Header: hdr("TRACKST"), Name: "Railway"}
	om := &objects.ObjectManager{RoadObjects: []*objects.RoadObject{town, tram, dirt}, TrackObjects: []*objects.TrackObject{rail}}
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	m := &Map{Width: 8, Height: 8, Tiles: make([][]Tile, 8)}
	for y := range m.Tiles {
		m.Tiles[y] = make([]Tile, 8)
		for x := range m.Tiles[y] {
			m.Tiles[y][x].BaseZ = 4
			m.Tiles[y][x].Order = []scenario.ElemRef{{Kind: scenario.KindSurface}}
		}
	}
	road := func(obj, rot uint8) *scenario.RoadElement {
		return &scenario.RoadElement{RoadObjectID: obj, RoadID: 0, Rotation: rot, BaseZ: 8, Owner: CompanyNeutral}
	}
	args := func(obj, rot uint8, z int32) roadClearArgs {
		return roadClearArgs{X: 64, Y: 64, Z: z, RoadObjectID: obj, RoadID: 0, Rotation: rot, IsLastIndex: true}
	}
	check := func(what string, want int, wantArg string, run func()) {
		t.Helper()
		c.ErrorText, c.ErrorArg = 0, ""
		run()
		if c.ErrorText != want || c.ErrorArg != wantArg {
			t.Errorf("%s: error %d %q, want %d %q", what, c.ErrorText, c.ErrorArg, want, wantArg)
		}
	}
	var st roadCmdState
	check("the same piece again", i18n.AlreadyBuiltHere, "", func() { m.clearRoad(c, &st, road(0, 0), args(0, 0, 32)) })
	check("another height", i18n.StringIdInTheWayWrongHeightForJunction, "Town road", func() { m.clearRoad(c, &st, road(0, 0), args(0, 1, 40)) })
	sig := road(0, 0)
	sig.HasSignalElement = true
	check("a signal", i18n.SignalInTheWay, "", func() { m.clearRoad(c, &st, sig, args(0, 1, 32)) })
	check("dirt crossing dirt", i18n.JunctionsNotPossible, "", func() { m.clearRoad(c, &st, road(2, 0), args(2, 1, 32)) })
	check("dirt crossing the town road", i18n.UnableToCrossOrCreateJunctionWithString, "Town road", func() { m.clearRoad(c, &st, road(0, 0), args(2, 1, 32)) })
	if st.AltRoadObject != 0 {
		t.Errorf("alternateRoadObjectId: %d, want the town road met (0)", st.AltRoadObject)
	}
	track := &scenario.TrackElement{TrackObjectID: 0, BaseZ: 8}
	check("a road over incompatible track", i18n.UnableToCrossOrCreateJunctionWithString, "Railway", func() { m.clearTrack(c, &st, track, args(0, 1, 32)) })

	a := roadPlacementArgs{X: 3 * 32, Y: 3 * 32, Z: 32, RoadID: 0, RoadObjectID: 0, Bridge: 0xFF}
	check("raised without a bridge", i18n.BridgeNeeded, "", func() { m.createRoadCheck(c, a, 0) })
	a.Z = 0
	check("below 16", i18n.ErrorTooLow, "", func() { m.createRoadCheck(c, a, 0) })
}
