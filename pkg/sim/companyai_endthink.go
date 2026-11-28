package sim

import (
	"log"

	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// aiThinkStateEndCompany is AiThinkState::endCompany (CompanyAi.h): the
// state the acquire-assets cheat leaves a company in (cheats.go).
const aiThinkStateEndCompany = 10

// The one OpenLoco AI think state goloco runs for a company no goloco AI
// runs: aiThinkEndCompany (CompanyAi.cpp, _funcs_430786[endCompany]). Every
// other state of OpenLoco's AI (CompanyAi.cpp ~6000 lines) stays unported,
// so a company in one of them does nothing (companyai.go); the endCompany
// state is reached only through a takeover by the cheat or a save made in
// it, and leads nowhere but aiDestroy. PARTIAL: aiThink's
// processVehiclePlaceStateMachine after the state's think.
// OpenLoco reference: src/OpenLoco/src/World/CompanyAi/CompanyAi.cpp.

// aiThinkEndCompany is aiThinkEndCompany: once
// removeAllCompanyAssetsOnMapByChunk has gone over the whole map, the
// company is removed (aiDestroy).
func (m *Map) aiThinkEndCompany(id uint8, om *objects.ObjectManager) {
	if !m.removeAllCompanyAssetsOnMapByChunk(id, om) {
		return
	}
	m.AIDestroy(id, om)
}

// removeAllCompanyAssetsOnMapByChunk is CompanyAi.cpp 0x004884E7: from the
// tile at aiPathfindTargetPos to the map's far corner (a TilePosRangeView:
// every row starts again at the first tile's x), each tile's track and road
// of the company removed — 1500 tiles a think, a row's last tile counting
// twice (vanilla's, kept upstream) —, the tile it stopped at kept in
// aiPathfindTargetPos. False until the end is reached; then the company's
// headquarters goes (apply | noPayment) and it returns true.
func (m *Map) removeAllCompanyAssetsOnMapByChunk(id uint8, om *objects.ObjectManager) bool {
	c := &m.Companies[id]
	cmd := &Cmd{Company: id, Objects: om, Players: func(o uint8) bool { return m.IsPlayerCompany(int(o)) }}
	var g1 G1Raw
	if om != nil {
		g1, _ = om.G1File.(G1Raw)
	}
	x0, y0 := int(c.AiPathfindTargetX)/32, int(c.AiPathfindTargetY)/32 // toTileSpace
	count := 1500
	for y := y0; y <= mapColumns-1 && y < len(m.Tiles); y++ {
		for x := x0; x <= mapColumns-1 && x < len(m.Tiles[y]); x++ {
			m.removeCompanyTracksRoadsOnTile(cmd, g1, x, y)
			count--
			if x == mapColumns-1 {
				count--
			}
			if count == 0 {
				c.AiPathfindTargetX, c.AiPathfindTargetY = int16(x*32), int16(y*32) // toWorldSpace
				return false
			}
		}
	}
	if c.HeadquartersX != -1 {
		hx, hy, hz := int32(c.HeadquartersX), int32(c.HeadquartersY), int32(c.HeadquartersZ)*smallZStep
		m.DoCommandFor(cmd, CmdApply|CmdNoPayment, ExpenditureConstruction, func(f uint8) (int32, bool) {
			return m.RemoveCompanyHeadquarters(cmd, f, hx, hy, hz)
		})
	}
	return true
}

// removeCompanyTracksRoadsOnTile is CompanyAi.cpp 0x00488563: the tile's
// first track or road of the company is removed (apply | noPayment, and
// aiAllocated when the element is), and the tile is searched again until
// none of the company's is left. A removal that is refused ends the search
// (upstream's would retry the same element for ever).
func (m *Map) removeCompanyTracksRoadsOnTile(c *Cmd, g1 G1Raw, x, y int) {
	t := &m.Tiles[y][x]
	wx, wy := int32(x)*32, int32(y)*32
	for retry := true; retry; {
		retry = false
		for _, r := range t.Order {
			var ok bool
			switch r.Kind {
			case scenario.KindTrack:
				el := t.Tracks[r.Index]
				if el.Owner != c.Company {
					continue
				}
				flags := uint8(CmdApply | CmdNoPayment)
				if el.AiAllocated || r.Flags&elemFlagAiAllocated != 0 {
					flags |= CmdAiAllocated
				}
				a := TrackRemoval{X: wx, Y: wy, Z: int32(el.BaseZ) * smallZStep, Rotation: el.Rotation,
					TrackID: el.TrackID, Index: el.SequenceIndex, TrackObjectID: el.TrackObjectID}
				_, ok = m.DoCommandFor(c, flags, ExpenditureConstruction, func(f uint8) (int32, bool) {
					return m.RemoveTrack(c, f, a, g1)
				})
			case scenario.KindRoad:
				el := t.Roads[r.Index]
				if el.Owner != c.Company {
					continue
				}
				flags := uint8(CmdApply | CmdNoPayment)
				if el.AiAllocated || r.Flags&elemFlagAiAllocated != 0 {
					flags |= CmdAiAllocated
				}
				a := RoadRemoval{X: wx, Y: wy, Z: int32(el.BaseZ) * smallZStep, Rotation: el.Rotation,
					RoadID: el.RoadID, Index: el.SequenceIndex, RoadObjectID: el.RoadObjectID}
				_, ok = m.DoCommandFor(c, flags, ExpenditureConstruction, func(f uint8) (int32, bool) {
					return m.RemoveRoad(c, f, a, g1)
				})
			default:
				continue
			}
			if !ok {
				log.Printf("[AI] company %d: its piece on tile %d,%d wasn't removed (aiThinkEndCompany)", c.Company, x, y)
				return
			}
			retry = true
			break
		}
	}
}
