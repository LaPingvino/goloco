package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// clearanceMap is a 6 × 6 flat map at baseZ 4 with only surfaces.
func clearanceMap() *Map {
	m := &Map{Width: 6, Height: 6}
	m.Tiles = make([][]Tile, 6)
	for y := range m.Tiles {
		m.Tiles[y] = make([]Tile, 6)
		for x := range m.Tiles[y] {
			m.Tiles[y][x] = Tile{BaseZ: 4, ClearZ: 4, Order: []scenario.ElemRef{{Kind: scenario.KindSurface}}}
		}
	}
	return m
}

// canConstructAtWithClear's own error texts (TileClearance.cpp 0x00462937):
// off the map's drawable edge; partly above and below water (with no clear
// function, or one that answers a plain collision — a clear function's own
// error stands); what is in the way (setCollisionErrorMessage), set even over
// an earlier text; another company about to build (aiCompanyAboutToBuildCheck).
func TestClearanceErrorTexts(t *testing.T) {
	m := clearanceMap()
	c := &Cmd{Company: 0}
	if ok, _ := m.canConstructAtWithClear(c, 5, 64, 4, 8, 0xF, 0, false, nil); ok || c.ErrorText != i18n.OffEdgeOfMap {
		t.Errorf("off the edge: %v %d", ok, c.ErrorText)
	}

	w := m.TileAt(2, 2)
	w.WaterLevel = 2 // water at baseZ 8
	c = &Cmd{Company: 0}
	if ok, _ := m.canConstructAtWithClear(c, 64, 64, 4, 12, 0xF, 0xF, false, nil); ok || c.ErrorText != i18n.CannotBuildPartlyAboveBelowWater {
		t.Errorf("partly under water: %v %d", ok, c.ErrorText)
	}
	c = &Cmd{Company: 0}
	own := func(*Tile, int) ClearResult { c.ErrorText = i18n.ObjectInTheWay; return ClearCollisionErrorSet }
	if ok, _ := m.canConstructAtWithClear(c, 64, 64, 4, 12, 0xF, 0xF, false, own); ok || c.ErrorText != i18n.ObjectInTheWay {
		t.Errorf("the clear function's own text: %v %d", ok, c.ErrorText)
	}
	plain := func(*Tile, int) ClearResult { return ClearCollision }
	if ok, _ := m.canConstructAtWithClear(c, 64, 64, 4, 12, 0xF, 0xF, false, plain); ok || c.ErrorText != i18n.CannotBuildPartlyAboveBelowWater {
		t.Errorf("a plain collision under water: %v %d", ok, c.ErrorText)
	}

	h := &objects.ObjectHeader{}
	copy(h.Name[:], "TRACKST ")
	tr := m.TileAt(3, 3)
	tr.Tracks = []scenario.TrackElement{{BaseZ: 4, ClearZ: 8, Owner: 1}}
	tr.Order = append(tr.Order, scenario.ElemRef{Kind: scenario.KindTrack, Flags: 0xF})
	c = &Cmd{Company: 0, Objects: &objects.ObjectManager{TrackObjects: []*objects.TrackObject{{Header: h, Name: "Standard track"}}}}
	c.ErrorText = i18n.JunctionsNotPossible
	if ok, _ := m.canConstructAtWithClear(c, 96, 96, 4, 8, 0xF, 0, false, nil); ok || c.ErrorText != i18n.StringIdInTheWay || c.ErrorArg != "Standard track" {
		t.Errorf("the track in the way: %v %d %q", ok, c.ErrorText, c.ErrorArg)
	}

	// The same track allocated by company 1, an AI: the neutral company (a
	// town) may not build there; a player company may, and tells the AI (unk1)
	// — unless the AI is converting it already (unk2).
	tr.Order[1].Flags |= elemFlagAiAllocated
	m.CompanyFaces = make([]scenario.CompanyFace, 2)
	players := func(id uint8) bool { return id == 0 }
	c = &Cmd{Company: CompanyNeutral, Players: players}
	if ok, _ := m.canConstructAtWithClear(c, 96, 96, 4, 8, 0xF, 0, false, nil); ok || c.ErrorText != i18n.AnotherCompanyIsAboutToBuildHere {
		t.Errorf("a town over an AI's allocation: %v %d", ok, c.ErrorText)
	}
	c = &Cmd{Company: 0, Players: players}
	if ok, _ := m.canConstructAtWithClear(c, 96, 96, 4, 8, 0xF, 0, false, nil); !ok || m.CompanyFaces[1].ChallengeFlags&companyFlagUnk1 == 0 {
		t.Errorf("a player over an AI's allocation: %v flags %b", ok, m.CompanyFaces[1].ChallengeFlags)
	}
	m.CompanyFaces[1].ChallengeFlags |= companyFlagUnk2
	c = &Cmd{Company: 0, Players: players}
	if ok, _ := m.canConstructAtWithClear(c, 96, 96, 4, 8, 0xF, 0, false, nil); ok || c.ErrorText != i18n.AnotherCompanyIsAboutToBuildHere {
		t.Errorf("a player over a converting AI's allocation: %v %d", ok, c.ErrorText)
	}
}
