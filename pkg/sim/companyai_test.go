package sim

import (
	"fmt"
	"testing"

	"github.com/LaPingvino/goloco/pkg/core"
	"github.com/LaPingvino/goloco/pkg/i18n"
)

// testAI records its thinks and draws from its own PRNG each time.
type testAI struct {
	thinks []uint32
	draws  []uint32
}

func (a *testAI) Name() string { return "test" }

// Save / Load: the thinks count, as text.
func (a *testAI) Save() []byte { return []byte(fmt.Sprint(len(a.thinks))) }
func (a *testAI) Load(b []byte) error {
	var n int
	if _, err := fmt.Sscan(string(b), &n); err != nil {
		return err
	}
	a.thinks = make([]uint32, n)
	return nil
}
func (a *testAI) Think(v *AIView) {
	a.thinks = append(a.thinks, v.Ticks())
	a.draws = append(a.draws, v.Rng().RandNext())
	v.Logf("think %d", len(a.thinks))
}

var lastTestAI *testAI

func init() {
	RegisterAI("test", func() CompanyAI { lastTestAI = &testAI{}; return lastTestAI })
}

// aiTestMap is a map without game data with company 0 the player's and
// company 3 a competitor's (non-empty, no thoughts).
func aiTestMap() *Map {
	m := &Map{}
	m.ResetCompanies()
	m.CompanyNames[0] = stringIDNewCompany
	m.CompanyNames[3] = stringIDNewCompany
	for k := range m.Companies[3].ThoughtType {
		m.Companies[3].ThoughtType[k] = 0xFF
	}
	m.Rng1 = *core.NewPrngSeed(0x1234, 0x5678)
	m.CurrencyMultiplicationFactor[0] = 1024 // the loan cap: 1000 × 1024 >> 8 = 4000
	m.MaxLoanSize = 1000
	return m
}

// With no AI attached the company tick does nothing: the rng, the cash and
// the companies stay as they were (the AI-off game is OpenLoco's with
// companyAIDisabled).
func TestCompaniesTickWithoutAI(t *testing.T) {
	m := aiTestMap()
	m.Companies[3].Cash = 500
	before := m.Rng1
	for range 64 {
		m.ScenarioTicks++
		m.companiesTick(nil)
	}
	if m.Rng1 != before || m.Companies[3].Cash != 500 {
		t.Errorf("the tick changed the game: rng %v → %v, cash %d", before, m.Rng1, m.Companies[3].Cash)
	}
}

// An AI thinks every 16 ticks, on its company's slot (scenarioTicks & 0x0F),
// draws only from its own PRNG (the game's stays put), logs; the same game
// state gives the same draws; Config.CompanyAIDisabled stops it.
func TestCompanyAIThinkCadenceAndRng(t *testing.T) {
	run := func() ([]uint32, []uint32, []AIDecision) {
		m := aiTestMap()
		if err := m.SetCompanyAI(3, "test", AIOptions{}); err != nil {
			t.Fatal(err)
		}
		ai := lastTestAI
		rng0 := m.Rng1
		for range 64 {
			m.ScenarioTicks++
			m.companiesTick(nil)
		}
		if m.Rng1 != rng0 {
			t.Error("the AI drew from the game's rng")
		}
		return ai.thinks, ai.draws, m.AIDecisions(3)
	}
	thinks, draws, log := run()
	if len(thinks) != 4 {
		t.Fatalf("thinks at %v, want 4", thinks)
	}
	for _, tk := range thinks {
		if tk&0x0F != 3 {
			t.Errorf("thought on tick %d, not company 3's slot", tk)
		}
	}
	if len(log) != 4 || log[0].Text != "think 1" || log[0].Tick != thinks[0] {
		t.Errorf("log %+v", log)
	}
	_, draws2, _ := run()
	for i := range draws {
		if draws[i] != draws2[i] {
			t.Fatalf("same game, different draws: %v vs %v", draws, draws2)
		}
	}
	if draws[0] == draws[1] {
		t.Error("the PRNG doesn't advance")
	}

	m := aiTestMap()
	_ = m.SetCompanyAI(3, "test", AIOptions{})
	Config.CompanyAIDisabled = true
	defer func() { Config.CompanyAIDisabled = false }()
	for range 32 {
		m.ScenarioTicks++
		m.companiesTick(nil)
	}
	if len(lastTestAI.thinks) != 0 {
		t.Error("companyAIDisabled didn't stop the AI")
	}
}

// Only a competitor's company created in this game takes an AI: not the
// player's, not an empty slot, not one with OpenLoco AI thoughts (G10).
func TestSetCompanyAIRefusals(t *testing.T) {
	m := aiTestMap()
	if m.SetCompanyAI(0, "test", AIOptions{}) == nil {
		t.Error("the player's company took an AI")
	}
	if m.SetCompanyAI(5, "test", AIOptions{}) == nil {
		t.Error("an empty slot took an AI")
	}
	if m.SetCompanyAI(3, "nope", AIOptions{}) == nil {
		t.Error("an unknown AI was attached")
	}
	m.Companies[3].ThoughtType[7] = 2
	if m.SetCompanyAI(3, "test", AIOptions{}) == nil {
		t.Error("a company with OpenLoco thoughts was taken over")
	}
}

// ensureCompanyFunding (G6): an OpenLoco AI company (no goloco AI) still
// borrows by itself; one of goloco's AIs pays as a player unless AutoBorrow.
func TestAIFundingSwitch(t *testing.T) {
	m := aiTestMap()
	m.Companies[3].Cash = 100
	if !m.ensureCompanyFunding(nil, 3, 1500) || m.Companies[3].CurrentLoan != 2000 {
		t.Fatalf("OpenLoco's rule: loan %d", m.Companies[3].CurrentLoan)
	}
	m = aiTestMap()
	m.Companies[3].Cash = 100
	_ = m.SetCompanyAI(3, "test", AIOptions{})
	c := &Cmd{}
	if m.ensureCompanyFunding(c, 3, 1500) || m.Companies[3].CurrentLoan != 0 || c.ErrorText != i18n.NotEnoughCashRequiresCurrency32 {
		t.Errorf("goloco AI without AutoBorrow: loan %d, error %d", m.Companies[3].CurrentLoan, c.ErrorText)
	}
	m = aiTestMap()
	m.Companies[3].Cash = 100
	_ = m.SetCompanyAI(3, "test", AIOptions{AutoBorrow: true})
	if !m.ensureCompanyFunding(nil, 3, 1500) || m.Companies[3].CurrentLoan != 2000 {
		t.Errorf("AutoBorrow: loan %d", m.Companies[3].CurrentLoan)
	}
}

// An AI company's name, formatted as StringManager::formatString does the
// company-name strings: a prefix ("Ebony Transport"), the owner's last name
// ({STRINGID}: "Cross Lines"), a town ({POP16}{STRINGID}: "Duluth Freight").
func TestFormatAICompanyName(t *testing.T) {
	for _, c := range []struct {
		style, prefix int
		args          []any
		want          string
	}{
		{0, 0, []any{"Cross"}, "Ebony Transport"},
		{2, 11, []any{"Cross"}, "Cross Lines"},
		{11, 12, []any{"Dumblebum", "Duluth"}, "Duluth Freight"},
	} {
		got := formatNameString(companyAiPlaystyleString[c.style], append([]any{companyAiNamePrefixes[c.prefix]}, c.args...))
		if got != c.want {
			t.Errorf("style %d prefix %d: %q, want %q", c.style, c.prefix, got, c.want)
		}
	}
}

// updateMonthly1's automatic repayment (H4) goes with AutoBorrow: without it
// a goloco AI company's cash isn't taken for its loan (it would then be in
// the red after every month's interest, and bankrupt in nine); OpenLoco's AI
// companies still repay.
func TestAIMonthlyRepaySwitch(t *testing.T) {
	for _, c := range []struct {
		ai, autoBorrow bool
		wantLoan       int32
	}{{false, false, 0}, {true, false, 5000}, {true, true, 0}} {
		m := aiTestMap()
		m.Companies[3].Cash, m.Companies[3].CurrentLoan = 8000, 5000
		if c.ai {
			_ = m.SetCompanyAI(3, "test", AIOptions{AutoBorrow: c.autoBorrow})
		}
		m.companiesUpdateMonthly1(nil)
		if m.Companies[3].CurrentLoan != c.wantLoan {
			t.Errorf("AI %v AutoBorrow %v: loan %d, want %d", c.ai, c.autoBorrow, m.Companies[3].CurrentLoan, c.wantLoan)
		}
	}
}
