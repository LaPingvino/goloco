package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// aiElementCounts counts the company's elements on the map: AI-allocated
// and real, of track, road, station and signal.
type aiElementCounts struct{ ai, real [4]int }

func countCompanyElements(m *Map, id uint8) aiElementCounts {
	var n aiElementCounts
	add := func(k int, ai bool) {
		if ai {
			n.ai[k]++
		} else {
			n.real[k]++
		}
	}
	for y := range m.Tiles {
		for x := range m.Tiles[y] {
			t := &m.Tiles[y][x]
			for _, e := range t.Tracks {
				if e.Owner == id && !e.Ghost {
					add(0, e.AiAllocated)
				}
			}
			for _, e := range t.Roads {
				if e.Owner == id && !e.Ghost {
					add(1, e.AiAllocated)
				}
			}
			for _, e := range t.Stations {
				if e.Owner == id && !e.Ghost {
					add(2, e.AiAllocated)
				}
			}
			for _, r := range t.Order {
				if r.Kind != scenario.KindSignal {
					continue
				}
				// a signal has no owner: the track it stands on has
				for _, tr := range t.Tracks {
					if tr.Owner == id && tr.BaseZ == t.Signals[r.Index].BaseZ && !t.Signals[r.Index].Ghost {
						add(3, t.Signals[r.Index].AiAllocated)
						break
					}
				}
			}
		}
	}
	return n
}

// G10 on the title save: OpenLoco's AI companies hold planned (AI-allocated)
// track, road and stations. Taking one over removes those its thoughts plan,
// keeps everything it really owns and its cash, leaves the other companies
// alone, and nulls its thoughts and its vehicles' aiThoughtId. Left, as
// upstream's undo leaves them: AI-allocated road stops no thought names
// (companies 2 and 4: one and two), which removeAiAllocatedCompanyTracksRoadsOnTile
// skips (an AI-allocated station on the road) and removeAiAllocatedStations
// never reaches.
func TestTakeOverCompanyTitle(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	if err := m.TakeOverCompanyAI(1, "none-registered", AIOptions{}, om); err == nil {
		t.Fatal("TakeOverCompanyAI took an unregistered AI")
	}
	taken := 0
	for id := range m.Companies {
		if m.CompanyEmpty(id) || m.IsPlayerCompany(id) {
			continue
		}
		before := make([]aiElementCounts, len(m.Companies))
		for i := range m.Companies {
			before[i] = countCompanyElements(m, uint8(i))
		}
		if before[id].ai == [4]int{} {
			continue
		}
		taken++
		cash := m.Companies[id].Cash
		planned := map[[3]int]bool{} // the thoughts' AI-allocated stations: tile and baseZ
		for k, typ := range m.Companies[id].ThoughtType {
			if typ == 0xFF {
				continue
			}
			for _, st := range m.Companies[id].ThoughtStations[k][:min(4, int(m.Companies[id].ThoughtNumStations[k]))] {
				if st.Flags&aiStationFlagAiAllocated != 0 {
					planned[[3]int{int(st.X) >> 5, int(st.Y) >> 5, int(st.BaseZ)}] = true
				}
			}
		}
		m.TakeOverCompany(uint8(id), om)
		after := countCompanyElements(m, uint8(id))
		t.Logf("company %d: ai %v → %v, real %v", id, before[id].ai, after.ai, after.real)
		if after.ai[0] != 0 || after.ai[3] != 0 || after.ai[1] != after.ai[2] {
			t.Errorf("company %d: AI-allocated elements left: %v", id, after.ai)
		}
		for y := range m.Tiles {
			for x := range m.Tiles[y] {
				for _, e := range m.Tiles[y][x].Stations {
					if int(e.Owner) == id && e.AiAllocated && (e.StationType != 1 || planned[[3]int{x, y, int(e.BaseZ)}]) {
						t.Errorf("company %d: planned station left at %d,%d: %+v", id, x, y, e)
					}
				}
			}
		}
		if after.real != before[id].real {
			t.Errorf("company %d: real elements %v, want %v", id, after.real, before[id].real)
		}
		if m.Companies[id].Cash != cash {
			t.Errorf("company %d: cash %d, want %d", id, m.Companies[id].Cash, cash)
		}
		for i := range m.Companies {
			if i != id && countCompanyElements(m, uint8(i)) != before[i] {
				t.Errorf("company %d's takeover changed company %d", id, i)
			}
		}
		c := m.Companies[id]
		for k, typ := range c.ThoughtType {
			if typ != 0xFF {
				t.Errorf("company %d: thought %d type %d", id, k, typ)
			}
		}
		if c.AiThinkState != 0 {
			t.Errorf("company %d: think state %d", id, c.AiThinkState)
		}
		for _, e := range m.Entities {
			if e.EntityType == scenario.VehicleTypeHead && int(e.Owner) == id && len(e.Raw) > 0x60 && e.Raw[0x60] != 0xFF {
				t.Errorf("company %d: head %d aiThoughtId %d", id, e.ID, e.Raw[0x60])
			}
		}
	}
	if taken == 0 {
		t.Fatal("no company with AI-allocated elements on the title save")
	}
	// Each taken-over company's thoughts are null now: SetCompanyAI no
	// longer refuses it (no AI registered in this package: only the check).
	for id := range m.Companies {
		if !m.CompanyEmpty(id) && !m.IsPlayerCompany(id) {
			if err := m.SetCompanyAI(id, "none-registered", AIOptions{}); err == nil {
				t.Fatal("SetCompanyAI took an unregistered AI")
			}
		}
	}
}

// Taking over every OpenLoco AI company of the title save (the loaded-games
// setting's "take over"): each non-player company then runs the AI, the
// player's stays without one, and a second call finds none left.
func TestTakeOverOpenLocoAIsTitle(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	before := m.OpenLocoAICompanies()
	if len(before) == 0 {
		t.Fatal("no OpenLoco AI company on the title save")
	}
	taken := m.TakeOverOpenLocoAIs("test", AIOptions{}, om)
	if len(taken) != len(before) {
		t.Errorf("taken over %v, want %v", taken, before)
	}
	for id := range m.Companies {
		switch {
		case m.CompanyEmpty(id):
		case m.IsPlayerCompany(id):
			if m.CompanyAIName(id) != "" {
				t.Errorf("the player's company %d got an AI", id)
			}
		case m.CompanyAIName(id) != "test":
			t.Errorf("company %d runs %q", id, m.CompanyAIName(id))
		}
	}
	if again := m.TakeOverOpenLocoAIs("test", AIOptions{}, om); len(again) != 0 || len(m.OpenLocoAICompanies()) != 0 {
		t.Errorf("a second takeover took %v", again)
	}
}

// The game's KeepOpenLocoAI goes into the AI chunk (version 3) and comes
// back; a game with neither AIs nor the switch writes no chunk.
func TestAIChunkKeepOpenLocoAI(t *testing.T) {
	m := aiTestMap()
	if m.EncodeAIChunk() != nil {
		t.Fatal("a chunk without AIs or switches")
	}
	m.KeepOpenLocoAI = true
	data := m.EncodeAIChunk()
	if data == nil {
		t.Fatal("no chunk for the switch")
	}
	back := aiTestMap()
	if err := back.DecodeAIChunk(data); err != nil || !back.KeepOpenLocoAI || len(back.OpenLocoAICompanies()) != 1 {
		t.Errorf("decoded: %v, keep %v, OpenLoco companies %v", err, back.KeepOpenLocoAI, back.OpenLocoAICompanies())
	}
	if err := m.SetCompanyAI(3, "test", AIOptions{}); err != nil {
		t.Fatal(err)
	}
	m.KeepOpenLocoAI = false
	back = aiTestMap()
	back.KeepOpenLocoAI = true
	if err := back.DecodeAIChunk(m.EncodeAIChunk()); err != nil || back.KeepOpenLocoAI || back.CompanyAIName(3) != "test" {
		t.Errorf("decoded: %v, keep %v, company 3 %q", err, back.KeepOpenLocoAI, back.CompanyAIName(3))
	}
}
