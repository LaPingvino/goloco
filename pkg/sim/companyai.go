package sim

import (
	"fmt"
	"log"
	"sort"

	"github.com/LaPingvino/goloco/pkg/core"
	"github.com/LaPingvino/goloco/pkg/objects"
)

// The Company AI interface (PORT-PLANS F step 2, docs/AI-INTERFACE.md §6).
//
// goloco's own design, not a port: OpenLoco's CompanyAi.cpp was only studied
// for the end points an AI needs. An AI sees what a player could see (the
// map, towns, industries, stations, its own company and vehicles; never other
// companies' plans) and does only what a player can do: game commands with
// explicit arguments, each run through DoCommandFor (the query pass, the
// funding check, the applying pass, the payment), with no flags of its own —
// no noPayment, aiAllocated, ghost or flag_7. Each AI has its own PRNG, never
// Map.Rng1, and a decision log.
//
// Hooked in at CompanyManager::tick's place (CompanyManager.cpp 0x00430319,
// after EffectsManager, before AnimationManager: Scenes/GameScene.cpp). With
// no AI attached to any company the hook does nothing at all — the AI-off
// game is the one the economy comparison with OpenLoco (simcmp) checks.
//
// Competitors appear as upstream's produceCompanies starts them (every 192
// ticks, after the start delay, up to maxCompetingCompanies), each running
// the AI Map.Competitors names; with none named, no competitor starts and
// the timeout doesn't run, as with Config::companyAIDisabled. An AI's state
// is saved in goloco's own chunk of the save (EncodeAIChunk), which OpenLoco
// doesn't read.
//
// Still open (AI-INTERFACE.md §7): removing a company (G5), taking over a
// loaded OpenLoco AI company (G10: refused here, so such companies keep
// standing still), events (G12), saving the decision log (G11: the log
// starts empty after a load).

// CompanyAI is one AI, run for one company, chosen when the company is
// created (CreateAICompany).
type CompanyAI interface {
	// Name is the AI's registered name.
	Name() string
	// Think is one bounded step, every 16 ticks for its company's slot
	// (scenarioTicks & 0x0F == id, as CompanyManager::tick). Long work spreads
	// over thinks in the AI's own state.
	Think(v *AIView)
	// Save is the AI's own state, kept in the saved game (EncodeAIChunk);
	// Load restores it into a new AI of the same name. The state must be
	// enough to go on as if the game had never been saved.
	Save() []byte
	Load(b []byte) error
}

// Competitors is the goloco AI that produceCompanies gives the competitors
// it starts, and its switches; AI "" starts none.
type Competitors struct {
	AI   string
	Opts AIOptions
}

var aiRegistry = map[string]func() CompanyAI{}

// RegisterAI makes an AI available by name (from its package's init).
func RegisterAI(name string, newAI func() CompanyAI) { aiRegistry[name] = newAI }

// AINames lists the registered AIs, sorted.
func AINames() []string {
	var out []string
	for n := range aiRegistry {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// AIOptions are a company AI's switches. Every advantage over a player is a
// separate switch, off by default (PORT-PLANS F, "cheats: configurable, off
// by default").
type AIOptions struct {
	// AutoBorrow keeps OpenLoco's rules for non-player companies
	// (AI-INTERFACE.md H3, H4, G6): ensureCompanyFunding takes out a loan for
	// a command it can't pay, and updateMonthly1 pays the loan back from the
	// cash each month. Off, the company is treated as a player: refused
	// without the cash, and its loan is its own business (ChangeLoan).
	AutoBorrow bool
}

// AIDecision is one line of an AI's decision log: a command it ran (Do)
// with its result, or a note it wrote. Queries are only counted
// (AIQueries).
type AIDecision struct {
	Tick uint32 // ScenarioTicks
	Day  uint32 // CurrentDay
	Text string
	// For a command: its result and cost (negative: a refund) and the
	// refusal's StringId.
	Command bool
	OK      bool
	Cost    int32
	Err     int
}

// aiLogMax bounds a company's kept log (the oldest lines go first).
const aiLogMax = 4096

// companyAI is one company's AI and what goloco keeps for it.
type companyAI struct {
	ai   CompanyAI
	name string
	opts AIOptions
	rng  core.Prng
	log  []AIDecision
	// queries counts the dry runs (Query), which the log leaves out.
	queries uint64
}

// aiSeed is a company AI's PRNG seed: the game's rng state when the AI is
// attached (read, not drawn) mixed with the company id, so two companies of
// one game, and one company of two games, draw differently.
func (m *Map) aiSeed(id int) (uint32, uint32) {
	k := uint32(id + 1)
	return m.Rng1.Srand0() ^ k*0x9E3779B9, m.Rng1.Srand1() ^ k*0x85EBCA6B
}

// SetCompanyAI runs the registered AI name for company id from now on. Only
// a non-player company created in this game may get one (AI-INTERFACE.md
// G10: an OpenLoco AI company's thoughts would dangle).
func (m *Map) SetCompanyAI(id int, name string, opts AIOptions) error {
	newAI, ok := aiRegistry[name]
	if !ok {
		return fmt.Errorf("no AI named %q", name)
	}
	if id < 0 || id >= companyMaxCount || m.CompanyEmpty(id) || m.IsPlayerCompany(id) {
		return fmt.Errorf("company %d can't have an AI", id)
	}
	for k := range m.Companies[id].ThoughtType {
		if m.Companies[id].ThoughtType[k] != 0xFF {
			return fmt.Errorf("company %d has OpenLoco AI thoughts (taking over is G10)", id)
		}
	}
	s0, s1 := m.aiSeed(id)
	m.companyAIs[id] = &companyAI{ai: newAI(), name: name, opts: opts, rng: *core.NewPrngSeed(s0, s1)}
	return nil
}

// CreateAICompany is createAiCompany (CompanyManager.cpp 0x0042F9AC's
// selectNewCompetitor, then createCompany as a competitor) with the AI
// choice: the new company runs the AI name. Returns its id, 0xFF if none
// was made (no free face or slot, the competitor branch refused, or no
// such AI).
func (m *Map) CreateAICompany(name string, opts AIOptions, om *objects.ObjectManager) uint8 {
	if _, ok := aiRegistry[name]; !ok {
		return companyIDNull
	}
	competitor := m.selectNewCompetitor(om)
	if competitor == 0xFF {
		return companyIDNull
	}
	id := m.createCompany(competitor, false, om)
	if id == companyIDNull {
		return companyIDNull
	}
	if err := m.SetCompanyAI(int(id), name, opts); err != nil {
		return companyIDNull
	}
	return id
}

// CompanyAIName is the AI a company runs ("" none).
func (m *Map) CompanyAIName(id int) string {
	if id < 0 || id >= companyMaxCount || m.companyAIs[id] == nil {
		return ""
	}
	return m.companyAIs[id].name
}

// AIDecisions is a company AI's log, oldest first (a copy).
func (m *Map) AIDecisions(id int) []AIDecision {
	if id < 0 || id >= companyMaxCount || m.companyAIs[id] == nil {
		return nil
	}
	return append([]AIDecision(nil), m.companyAIs[id].log...)
}

// AIQueries is how many dry runs a company AI has made.
func (m *Map) AIQueries(id int) uint64 {
	if id < 0 || id >= companyMaxCount || m.companyAIs[id] == nil {
		return 0
	}
	return m.companyAIs[id].queries
}

// aiPaysAsPlayer is ensureCompanyFunding's switch (G6): a company with one
// of goloco's AIs and AutoBorrow off is refused as a player is.
func (m *Map) aiPaysAsPlayer(id int) bool {
	return id >= 0 && id < companyMaxCount && m.companyAIs[id] != nil && !m.companyAIs[id].opts.AutoBorrow
}

// companiesTick is CompanyManager::tick (0x00430319) for goloco's AIs: not
// with Config::companyAIDisabled, the company scenarioTicks & 0x0F thinks if
// it isn't a player's, isn't empty and has an AI; then, every 192 ticks,
// produceCompanies. The timeout runs only while Map.Competitors names an AI
// (goloco's "no competitors" is upstream's companyAIDisabled for it), so a
// game without one is tick for tick the AI-off game. PARTIAL: the editor
// check (the sim has no editor mode flag).
func (m *Map) companiesTick(om *objects.ObjectManager) {
	if Config.CompanyAIDisabled {
		return
	}
	id := int(m.ScenarioTicks & 0x0F)
	if st := m.aiOf(id); st != nil && !m.IsPlayerCompany(id) && !m.CompanyEmpty(id) {
		st.ai.Think(&AIView{m: m, om: om, id: uint8(id), st: st})
	}
	if m.Competitors.AI == "" {
		return
	}
	m.ProduceAICompanyTimeout++
	if m.ProduceAICompanyTimeout >= 192 {
		m.ProduceAICompanyTimeout = 0
		m.produceCompanies(om)
	}
}

func (m *Map) aiOf(id int) *companyAI {
	if id < 0 || id >= companyMaxCount {
		return nil
	}
	return m.companyAIs[id]
}

// produceCompanies is CompanyManager.cpp produceCompanies (0x004306D1):
// once the start delay is over and competitors are allowed, a one in 16
// chance, then randNext(maxCompetingCompanies) + 1 above the companies that
// aren't a player's starts one — createAiCompany, as CreateAICompany with
// Map.Competitors' AI instead of OpenLoco's. Both draws are the game's
// gPrng1, as upstream (and selectNewCompetitor / createCompany draw it too),
// so a game with competitors parts from the AI-off game here.
func (m *Map) produceCompanies(om *objects.ObjectManager) {
	if m.CompetitorStartDelay != 0 || m.MaxCompetingCompanies == 0 {
		return
	}
	active := 0
	for id := range m.CompanyNames {
		if !m.CompanyEmpty(id) && !m.IsPlayerCompany(id) {
			active++
		}
	}
	if m.Rng1.RandNextBound(16) != 0 {
		return
	}
	if int(m.Rng1.RandNextBound(int32(m.MaxCompetingCompanies)))+1 > active {
		if id := m.CreateAICompany(m.Competitors.AI, m.Competitors.Opts, om); id != companyIDNull {
			log.Printf("[AI] competitor %d started: %s, running %s", id, m.CompanyName(int(id)), m.Competitors.AI)
		}
	}
}

// competitorsUpdateMonthly1 is CompanyManager::updateMonthly1's first line:
// the competitor start delay counts down a month (Math::Bound::sub, never
// below 0), with the AI on or off.
func (m *Map) competitorsUpdateMonthly1() {
	if m.CompetitorStartDelay > 0 {
		m.CompetitorStartDelay--
	}
}

func (st *companyAI) record(d AIDecision) {
	if len(st.log) >= aiLogMax {
		st.log = append(st.log[:0], st.log[len(st.log)-aiLogMax+1:]...)
	}
	st.log = append(st.log, d)
}
