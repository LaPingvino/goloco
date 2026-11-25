package sim

import (
	"fmt"
	"log"

	"github.com/LaPingvino/goloco/pkg/objects"
)

// Removing a company (AI-INTERFACE.md G5): an AI that gives up, or never
// gets going, ends its company. Upstream's AI does so in two places
// (CompanyAi.cpp): aiThinkState1 when no thought has started the company
// (aiHasStarted unset) more than 42 days after its start, and
// aiThinkEndCompany once removeAllCompanyAssetsOnMapByChunk has taken
// everything off the map; both call CompanyManager::aiDestroy. goloco's AI
// does only what a player can, so it sells and removes its own assets with
// the ordinary commands first; EndCompany refuses while it owns any.

// EndCompany is the AI giving up: once this think is over the company is
// removed (aiDestroy). Refused while it still owns a vehicle, a station, or
// a piece of track or road (upstream's AI removes them all first:
// removeAllCompanyAssetsOnMapByChunk); its headquarters goes with the
// company, as in aiDestroy.
func (v *AIView) EndCompany() error {
	if what := v.m.companyOwnsAssets(v.id); what != "" {
		return fmt.Errorf("company %d still owns %s", v.id, what)
	}
	v.st.ending = true
	v.Logf("ending the company")
	return nil
}

// companyOwnsAssets names the first kind of asset company id still owns
// ("" none): a vehicle, a station, a track or road piece.
func (m *Map) companyOwnsAssets(id uint8) string {
	if trains, _ := m.Trains(); trains != nil {
		for _, tr := range trains {
			if m.Entities[tr.Head].Owner == id {
				return "a vehicle"
			}
		}
	}
	for _, s := range m.StationNames {
		if s.Name != 0 && s.Owner == id {
			return "a station"
		}
	}
	for y := range m.Tiles {
		for x := range m.Tiles[y] {
			t := &m.Tiles[y][x]
			for _, e := range t.Tracks {
				if e.Owner == id && !e.Ghost {
					return "track"
				}
			}
			for _, e := range t.Roads {
				if e.Owner == id && !e.Ghost {
					return "road"
				}
			}
		}
	}
	return ""
}

// AIDestroy is CompanyManager::aiDestroy (0x00435AEF, OpenLoco reference:
// src/OpenLoco/src/World/CompanyManager.cpp): the company's headquarters
// removed (doCommand apply, under the company), its windows closed and the
// company select dropdown with them (CompanyRemoved: the game's), every
// message's reference to it cleared (removeAllSubjectRefs), its speed
// records' holder cleared (removeCompaniesRecords), its name and owner name
// emptied — the slot is empty again —, and its competitor object unloaded
// (ObjectManager::unload + reloadAll: the slot freed, so selectNewCompetitor
// may load another). goloco's AI for it goes too. PARTIAL: CompanyList's
// removeCompany (goloco's list is built from the non-empty slots each time).
func (m *Map) AIDestroy(id uint8, om *objects.ObjectManager) {
	if int(id) >= len(m.Companies) || m.CompanyEmpty(int(id)) {
		return
	}
	if c := m.Companies[id]; c.HeadquartersX != -1 {
		cmd := &Cmd{Company: id, Objects: om, Players: func(o uint8) bool { return m.IsPlayerCompany(int(o)) }}
		x, y, z := int32(c.HeadquartersX), int32(c.HeadquartersY), int32(c.HeadquartersZ)*smallZStep
		m.DoCommandFor(cmd, CmdApply, ExpenditureConstruction, func(f uint8) (int32, bool) {
			return m.RemoveCompanyHeadquarters(cmd, f, x, y, z)
		})
	}
	if m.CompanyRemoved != nil {
		m.CompanyRemoved(id)
	}
	m.RemoveAllSubjectRefs(uint16(id), MsgArgCompany)
	m.removeCompaniesRecords(id)
	m.emptyUserString(m.CompanyNames[id])
	m.CompanyNames[id] = 0 // StringIds::empty
	face := &m.CompanyFaces[id]
	m.emptyUserString(face.OwnerName)
	face.OwnerName = 0
	if om != nil && int(face.Competitor) < len(om.CompetitorObjects) {
		om.CompetitorObjects[face.Competitor] = nil
	}
	m.companyAIs[id] = nil
	log.Printf("[AI] company %d removed", id)
}

// removeCompaniesRecords (CompanyManager.cpp): each speed record set
// (speed not kSpeedZero) held by the company has no holder any more.
func (m *Map) removeCompaniesRecords(id uint8) {
	r := &m.CompanyRecords
	for i := range 3 {
		if r.Speed[i] != 0 && r.Company[i] == id {
			r.Company[i] = companyIDNull
		}
	}
}
