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
