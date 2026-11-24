package companyai

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// step is one game tick as gosave plays it (no window: no waves).
func step(s *sim.State, om *objects.ObjectManager) { s.Tick(func(int, int) bool { return false }, om) }

// saveBytes is the saved game's file: header, game state, tiles and goloco's
// AI chunk (scenario.S5File).
func saveBytes(t *testing.T, s *sim.State) []byte {
	t.Helper()
	s.PrepareSave()
	sc := s.ExportScenario()
	var buf bytes.Buffer
	if err := sc.SavedGameFile(make([]byte, scenario.SaveDetailsSize), s.Raw.RequiredObjects).Write(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// reload is a load of a saved game's file into a fresh state.
func reload(t *testing.T, data []byte) *sim.State {
	t.Helper()
	sc, err := scenario.ParseScenario(data, "resume.SV5")
	if err != nil {
		t.Fatal(err)
	}
	s := &sim.State{}
	s.LoadScenario(sc)
	s.LoadWaves(sc.Waves)
	return s
}

// firstCompetitor plays the title save with buslink competitors on until
// produceCompanies starts one, and returns its id (0xFF: none in max ticks).
func firstCompetitor(s *sim.State, om *objects.ObjectManager, max int) (uint8, int) {
	for tick := 1; tick <= max; tick++ {
		step(s, om)
		for id := range 15 {
			if s.CompanyAIName(id) != "" {
				return uint8(id), tick
			}
		}
	}
	return 0xFF, max
}

// produceCompanies on the title save (14 competitors allowed, 10 there,
// no start delay): a competitor starts on schedule — only at a 192-tick
// timeout — running buslink, never OpenLoco's AI, and the companies that
// were there keep no goloco AI.
func TestProduceCompaniesStartsBuslink(t *testing.T) {
	s, om := loadDataMap(t, "title.dat")
	s.Competitors = sim.Competitors{AI: BusLinkName}
	timeout0 := s.ProduceAICompanyTimeout
	before := map[int]bool{}
	for id := range 15 {
		before[id] = !s.CompanyEmpty(id)
	}
	id, tick := firstCompetitor(s, om, 200000)
	if id == 0xFF {
		t.Fatal("no competitor started in 200000 ticks")
	}
	t.Logf("company %d %q started at tick %d", id, s.CompanyName(int(id)), tick)
	if before[int(id)] {
		t.Errorf("company %d was there before", id)
	}
	if (int(timeout0)+tick)%192 != 0 {
		t.Errorf("started at tick %d, not at a 192-tick timeout (timeout %d at the start)", tick, timeout0)
	}
	if s.CompanyAIName(int(id)) != BusLinkName {
		t.Errorf("runs %q", s.CompanyAIName(int(id)))
	}
	for k := range s.Companies[id].ThoughtType {
		if s.Companies[id].ThoughtType[k] != 0xFF {
			t.Fatalf("the new company has an OpenLoco thought %d", k)
		}
	}
}

// The start delay: with months to go, produceCompanies starts no one and
// the game's rng isn't drawn for it; and without a competitor AI the
// timeout doesn't run at all.
func TestProduceCompaniesWaitsForDelay(t *testing.T) {
	s, om := loadDataMap(t, "title.dat")
	s.Competitors = sim.Competitors{AI: BusLinkName}
	s.CompetitorStartDelay = 200
	for range 8000 { // the first competitor of TestProduceCompaniesStartsBuslink starts at tick 5140
		step(s, om)
	}
	for id := range 15 {
		if s.CompanyAIName(id) != "" {
			t.Fatalf("company %d started during the delay", id)
		}
	}
	if s.CompetitorStartDelay >= 200 {
		t.Errorf("the delay didn't count down: %d", s.CompetitorStartDelay)
	}

	off, om := loadDataMap(t, "title.dat")
	t0 := off.ProduceAICompanyTimeout
	for range 500 {
		step(off, om)
	}
	if off.ProduceAICompanyTimeout != t0 {
		t.Errorf("no competitor AI, yet the timeout ran: %d → %d", t0, off.ProduceAICompanyTimeout)
	}
}

// G11: a saved game with a goloco AI loads and goes on as if never saved.
// The title save is played until buslink is building its road (mid-plan),
// saved and loaded; the loaded game and the unbroken one are played on and
// their saves compared byte for byte. The same with no AI is the control.
func TestAISaveResumes(t *testing.T) {
	const more = 1500
	run := func(ai bool) (saved, unbroken, resumed []byte) {
		s, om := loadDataMap(t, "title.dat")
		var id uint8 = 0xFF
		if ai {
			s.Competitors = sim.Competitors{AI: BusLinkName}
			var tick int
			id, tick = firstCompetitor(s, om, 200000)
			if id == 0xFF {
				t.Fatal("no competitor started")
			}
			building := false
			for n := 0; n < 20000 && !building; n++ {
				step(s, om)
				for _, d := range s.AIDecisions(int(id)) {
					building = building || strings.HasPrefix(d.Text, "plan: join")
				}
			}
			if !building {
				t.Fatal("buslink planned no road to build")
			}
			for range 64 { // four thinks into the road
				step(s, om)
			}
			t.Logf("company %d from tick %d; saved after %d decisions", id, tick, len(s.AIDecisions(int(id))))
		} else {
			for range 5000 {
				step(s, om)
			}
		}
		saved = saveBytes(t, s)
		s2 := reload(t, saved)
		s2.Competitors = s.Competitors // the options' choice, not the save's
		if ai && s2.CompanyAIName(int(id)) != BusLinkName {
			t.Fatalf("the loaded company %d runs %q", id, s2.CompanyAIName(int(id)))
		}
		for range more {
			step(s, om)
			step(s2, om)
		}
		return saved, saveBytes(t, s), saveBytes(t, s2)
	}
	_, offA, offB := run(false)
	if d := saveDiff(t, offA, offB); d != "" {
		t.Fatalf("control (no AI): the resumed game differs from the unbroken one: %s", d)
	}
	_, a, b := run(true)
	if d := saveDiff(t, a, b); d != "" {
		t.Errorf("with buslink: the resumed game differs from the unbroken one: %s", d)
	}
}

// saveDiff compares two saved games' game state, tiles and AI chunk ("" the
// same). Left out: the news messages (GameState 0x13B6-0xB95A), whose text
// after the terminator a loaded game keeps from the file and a played one
// zeroes — a save round trip's difference with or without AIs.
func saveDiff(t *testing.T, a, b []byte) string {
	t.Helper()
	sa, err := scenario.ParseScenario(a, "a.SV5")
	if err != nil {
		t.Fatal(err)
	}
	sb, err := scenario.ParseScenario(b, "b.SV5")
	if err != nil {
		t.Fatal(err)
	}
	ga, gb := append([]byte(nil), sa.Raw.GameState...), append([]byte(nil), sb.Raw.GameState...)
	clear(ga[0x13B6:0xB95A])
	clear(gb[0x13B6:0xB95A])
	for i := range ga {
		if ga[i] != gb[i] {
			return fmt.Sprintf("game state from 0x%X", i)
		}
	}
	if !bytes.Equal(sa.Raw.TileElements, sb.Raw.TileElements) {
		return "tile elements"
	}
	if !bytes.Equal(sa.GolocoAI, sb.GolocoAI) {
		return fmt.Sprintf("AI chunk (%d vs %d bytes)", len(sa.GolocoAI), len(sb.GolocoAI))
	}
	return ""
}

// OpenLoco's own reading of the file stops after the tile elements: the
// chunk sits between them and the checksum, so a file with it is OpenLoco's
// file with one more chunk, and the AI-off save has none.
func TestAIChunkOnlyWithAIs(t *testing.T) {
	s, om := loadDataMap(t, "title.dat")
	if s.EncodeAIChunk() != nil {
		t.Fatal("a chunk without any goloco AI")
	}
	id := s.CreateAICompany(BusLinkName, sim.AIOptions{}, om)
	if id == 0xFF {
		t.Fatal("no AI company")
	}
	for range 32 {
		step(s, om)
	}
	sc, err := scenario.ParseScenario(saveBytes(t, s), "chunk.SV5")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(sc.GolocoAI, []byte(scenario.GolocoAIMagic)) {
		t.Fatalf("the loaded file has no goloco chunk (%d bytes)", len(sc.GolocoAI))
	}
}
