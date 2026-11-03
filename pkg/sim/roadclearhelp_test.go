package sim

import (
	"fmt"
	"os"
	"strings"
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

// createRoad's start: an extra that doesn't fit the piece is dropped, a height
// off the 16 grid is refused, and the cost is the road's (and a valid extra's)
// inflation-adjusted price times the piece's cost factor / 256.
func TestCreateRoadCost(t *testing.T) {
	hdr := func(name string) *objects.ObjectHeader {
		h := &objects.ObjectHeader{}
		copy(h.Name[:], name)
		return h
	}
	road := &objects.RoadObject{Header: hdr("ROADTOWN"), BuildCostFactor: 100, CostIndex: 0, ModNames: []string{"PAVEMENT"}}
	om := &objects.ObjectManager{RoadObjects: []*objects.RoadObject{road}}
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	m := &Map{}
	base := m.InflationAdjustedCost(100, 0, 10)
	mods, cost, ok := m.createRoadCost(c, roadPlacementArgs{Z: 32, RoadID: 0, RoadObjectID: 0, Mods: 1})
	if !ok || mods != 0 { // no PAVEMENT object loaded: the extra is dropped
		t.Errorf("mods %b ok %v", mods, ok)
	}
	if want := base * int32(RoadCostFactor[0]) / 256; cost != want {
		t.Errorf("straight road cost %d, want %d", cost, want)
	}
	if _, _, ok := m.createRoadCost(c, roadPlacementArgs{Z: 36, RoadObjectID: 0}); ok {
		t.Error("a height off the 16 grid must fail")
	}
	_, curve, _ := m.createRoadCost(c, roadPlacementArgs{Z: 32, RoadID: 3, RoadObjectID: 0})
	if want := base * int32(RoadCostFactor[3]) / 256; curve != want {
		t.Errorf("piece 3 cost %d, want %d", curve, want)
	}
}

// createRoad's checks on a flat map: a straight road on the ground passes (one
// piece, its heights, the base cost); raised it needs a bridge; below 16 it is
// too low.
func TestCreateRoadCheck(t *testing.T) {
	hdr := func(name string) *objects.ObjectHeader {
		h := &objects.ObjectHeader{}
		copy(h.Name[:], name)
		return h
	}
	road := &objects.RoadObject{Header: hdr("ROADTOWN"), BuildCostFactor: 100}
	c := &Cmd{Company: CompanyNeutral, Objects: &objects.ObjectManager{RoadObjects: []*objects.RoadObject{road}}}
	m := &Map{Width: 8, Height: 8, Tiles: make([][]Tile, 8)}
	for y := range m.Tiles {
		m.Tiles[y] = make([]Tile, 8)
		for x := range m.Tiles[y] {
			m.Tiles[y][x].BaseZ = 4 // the ground at z 16
		}
	}
	a := roadPlacementArgs{X: 3 * 32, Y: 3 * 32, Z: 16, RoadID: 0, RoadObjectID: 0, Bridge: 0xFF}
	r, ok := m.createRoadCheck(c, a, 0)
	_, base, _ := m.createRoadCost(c, a)
	if !ok || len(r.Pieces) != 1 || r.Pieces[0].BaseZ != 4 || r.Pieces[0].ClearZ != 12 || r.Cost != base {
		t.Fatalf("flat straight road: ok %v, %+v", ok, r)
	}
	a.Z = 32
	if _, ok := m.createRoadCheck(c, a, 0); ok {
		t.Error("raised without a bridge: must fail")
	}
	a.Z = 8
	if _, ok := m.createRoadCheck(c, a, 0); ok {
		t.Error("below 16: too low")
	}
}

// createRoad's apply step: the road element is placed (its heights, owner,
// last flag, its ref's quarters), a wall on the tile removed, snow cleared;
// the same road again is refused (the new element collides).
func TestCreateRoadApply(t *testing.T) {
	hdr := &objects.ObjectHeader{}
	copy(hdr.Name[:], "ROADTOWN")
	road := &objects.RoadObject{Header: hdr, BuildCostFactor: 100}
	c := &Cmd{Company: CompanyNeutral, Objects: &objects.ObjectManager{RoadObjects: []*objects.RoadObject{road}}}
	m := &Map{Width: 8, Height: 8, Tiles: make([][]Tile, 8)}
	for y := range m.Tiles {
		m.Tiles[y] = make([]Tile, 8)
		for x := range m.Tiles[y] {
			m.Tiles[y][x].BaseZ = 4
			m.Tiles[y][x].Order = []scenario.ElemRef{{Kind: scenario.KindSurface}}
		}
	}
	tile := m.TileAt(3, 3)
	tile.Slope = 2 << 5 // snow
	tile.Walls = []scenario.WallElement{{BaseZ: 4, ClearZ: 8}}
	tile.insertRef(scenario.KindWall, 0, 4)
	a := roadPlacementArgs{X: 3 * 32, Y: 3 * 32, Z: 16, RoadID: 0, RoadObjectID: 0, Bridge: 0xFF}
	r, ok := m.createRoadCheck(c, a, CmdApply)
	if !ok || !m.createRoadApply(c, a, CmdApply, &r) {
		t.Fatal("placing a road on flat ground failed")
	}
	if len(tile.Roads) != 1 || len(tile.Walls) != 0 || tile.Slope&0xE0 != 0 {
		t.Fatalf("after: %d roads, %d walls, slope %#x", len(tile.Roads), len(tile.Walls), tile.Slope)
	}
	el := tile.Roads[0]
	if el.BaseZ != 4 || el.ClearZ != 12 || el.Owner != CompanyNeutral || !el.IsLast || el.HasBridge {
		t.Errorf("element %+v", el)
	}
	var ref scenario.ElemRef
	for _, x := range tile.Order {
		if x.Kind == scenario.KindRoad {
			ref = x
		}
	}
	if ref.Flags&0xF != rotateQuarterTile(RoadPieces[0][0].SubTileClearance, 0)&0xF || ref.Flags&0xF == 0 {
		t.Errorf("the road's quarters %#x", ref.Flags)
	}
	if _, ok := m.createRoadCheck(c, a, 0); ok {
		t.Error("the same road again must collide")
	}
}

// createRoad's junction pass against upstream: every combination of the
// straight / left / right junction bits (roadIdUnk[0..2], 4 rotations each),
// the pieces placed (rotation:roadId, in order) and the bits after, as
// upstream's own loop (CreateRoad.cpp 0x47656B, built by itself with the
// placement recorded) gives them — testdata/roadjunctions.txt.
func TestCreateRoadJunctionsUpstream(t *testing.T) {
	data, err := os.ReadFile("testdata/roadjunctions.txt")
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Split(strings.TrimSpace(string(data)), "\n")
	c := &Cmd{Company: CompanyNeutral}
	for in := 0; in < 4096; in++ {
		m := &Map{Width: 4, Height: 4, Tiles: make([][]Tile, 4)}
		for y := range m.Tiles {
			m.Tiles[y] = make([]Tile, 4)
		}
		r := &createRoadResult{}
		r.RoadIDUnk[0], r.RoadIDUnk[1], r.RoadIDUnk[2] = uint8(in&0xF)|0x80, uint8(in>>4&0xF), uint8(in>>8&0xF)
		m.createRoadJunctions(c, roadPlacementArgs{X: 32, Y: 32, Z: 64, Bridge: 0xFF}, CmdApply, r)
		var b strings.Builder
		fmt.Fprintf(&b, "%d", in)
		for _, e := range m.TileAt(1, 1).Roads {
			fmt.Fprintf(&b, " %d:%d", e.Rotation, e.RoadID)
		}
		fmt.Fprintf(&b, " | %d %d %d", r.RoadIDUnk[0], r.RoadIDUnk[1], r.RoadIDUnk[2])
		if b.String() != want[in] {
			t.Fatalf("got  %s\nwant %s", b.String(), want[in])
		}
	}
}
