package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// clearRoad's helpers: the junction trait, compatibility either way,
// single-tile pieces only, company compatibility, aiUnkCheck.
func TestClearRoadHelpers(t *testing.T) {
	hdr := func(name string) *objects.ObjectHeader {
		h := &objects.ObjectHeader{}
		copy(h.Name[:], name)
		return h
	}
	town := &objects.RoadObject{Header: hdr("ROADTOWN"), RoadPieces: roadTraitJunction}
	tram := &objects.RoadObject{Header: hdr("TRAMWAY"), CompatibleNames: []string{"ROADTOWN"}}
	plain := &objects.RoadObject{Header: hdr("DIRT")}
	om := &objects.ObjectManager{RoadObjects: []*objects.RoadObject{town, tram, plain}}
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	el := func(obj uint8) *scenario.RoadElement { return &scenario.RoadElement{RoadObjectID: obj, RoadID: 0} }
	args := func(obj, roadID uint8) roadClearArgs { return roadClearArgs{RoadObjectID: obj, RoadID: roadID} }

	if !c.roadTraitAndCompatibilityCheck(el(0), args(0, 0), town, town) {
		t.Error("the same road with the junction trait: a junction")
	}
	if c.roadTraitAndCompatibilityCheck(el(2), args(2, 0), plain, plain) {
		t.Error("the same road without the trait: no junction")
	}
	if !c.roadTraitAndCompatibilityCheck(el(0), args(1, 0), town, tram) {
		t.Error("tram compatible with the town road: a crossing")
	}
	if c.roadTraitAndCompatibilityCheck(el(0), args(2, 0), town, plain) {
		t.Error("dirt and the town road are not compatible")
	}
	if c.finalJunctionChecks(el(0), args(0, 3), town, town) { // road id 3: a multi-tile piece
		if len(RoadPieces[3]) > 1 {
			t.Error("a multi-tile piece: no junction")
		}
	}
	co := &Cmd{Company: 2}
	if !co.companyCompatible(2) || !co.companyCompatible(CompanyNeutral) || co.companyCompatible(3) {
		t.Error("companyCompatible")
	}
	if !c.companyCompatible(3) {
		t.Error("a town (neutral) may build over any company's road")
	}
	ai := &Cmd{Company: 2}
	e := &scenario.RoadElement{Owner: 2}
	if ai.aiUnkCheck(e, roadClearArgs{Flags: CmdAiAllocated, UnkFlags: 1 << 5}) != true || ai.aiUnkCheck(e, roadClearArgs{UnkFlags: 1 << 5}) {
		t.Error("aiUnkCheck")
	}
}

// clearRoad: a junction where the town road may have one, "already built
// here" for the same piece, an overlay of a compatible road type, and the
// refusals — another height, a signal, a road without the junction trait.
func TestClearRoad(t *testing.T) {
	hdr := func(name string) *objects.ObjectHeader {
		h := &objects.ObjectHeader{}
		copy(h.Name[:], name)
		return h
	}
	town := &objects.RoadObject{Header: hdr("ROADTOWN"), RoadPieces: roadTraitJunction}
	tram := &objects.RoadObject{Header: hdr("TRAMWAY"), CompatibleNames: []string{"ROADTOWN"}}
	dirt := &objects.RoadObject{Header: hdr("DIRT")}
	om := &objects.ObjectManager{RoadObjects: []*objects.RoadObject{town, tram, dirt}}
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	m := &Map{Width: 4, Height: 4, Tiles: make([][]Tile, 4)}
	for y := range m.Tiles {
		m.Tiles[y] = make([]Tile, 4)
	}
	road := func(obj, rot uint8) *scenario.RoadElement {
		return &scenario.RoadElement{RoadObjectID: obj, RoadID: 0, Rotation: rot, BaseZ: 8, Owner: CompanyNeutral}
	}
	args := func(obj, rot uint8, z int32) roadClearArgs {
		return roadClearArgs{X: 64, Y: 64, Z: z, RoadObjectID: obj, RoadID: 0, Rotation: rot, IsLastIndex: true}
	}
	run := func(e *scenario.RoadElement, a roadClearArgs) (roadClearResult, roadCmdState) {
		var st roadCmdState
		return m.clearRoad(c, &st, e, a), st
	}
	if r, _ := run(road(0, 0), args(0, 1, 32)); r.Res != ClearNoCollision || r.RoadIDUnk[0]&(1|1<<7) != 1|1<<7 {
		t.Errorf("town road crossing town road: %+v", r)
	}
	if r, _ := run(road(0, 0), args(0, 0, 32)); r.Res != ClearCollisionErrorSet {
		t.Errorf("the same piece again: %v, want already built here", r.Res)
	}
	if r, st := run(road(0, 0), args(1, 0, 32)); r.Res != ClearNoCollision || st.Flags73&(1<<5) == 0 {
		t.Errorf("tram over the town road: %v, flags %#x", r.Res, st.Flags73)
	}
	if r, _ := run(road(0, 0), args(0, 1, 40)); r.Res != ClearCollisionErrorSet {
		t.Errorf("another height: %v", r.Res)
	}
	sig := road(0, 0)
	sig.HasSignalElement = true
	if r, _ := run(sig, args(0, 1, 32)); r.Res != ClearCollisionErrorSet {
		t.Errorf("a signal: %v", r.Res)
	}
	if r, _ := run(road(2, 0), args(2, 1, 32)); r.Res != ClearCollisionErrorSet {
		t.Errorf("dirt crossing dirt (no junction trait): %v", r.Res)
	}
}
