package main

import (
	"strconv"

	"github.com/LaPingvino/goloco/pkg/companyai" // registers goloco's AIs
	"github.com/LaPingvino/goloco/pkg/sim"
)

// companyai.go — the competitors' settings (AI-INTERFACE.md G13): which of
// goloco's AIs the competitors run, how many may start and how soon, in the
// style of OpenTTD's AI settings. goloco's own, so they live in goloco.yml
// (golocoConfigPath), never in openloco.yml:
//
//	companyAI:
//	  ai: buslink          # a registered AI, or none
//	  competitors: scenario  # or 0–14: maxCompetingCompanies for a new game
//	  startDelay: scenario   # or 0–240 months: competitorStartDelay for a new game
//
// Default: buslink, and the scenario's own count and delay — competitors as
// OpenLoco has them by default, run by goloco's AI. The count and the delay
// are a scenario's (S5 Options / GameState maxCompetingCompanies,
// competitorStartDelay), so a choice other than "scenario" replaces them when
// a new game starts (Scenario::start), as OpenTTD's settings apply to the
// games started after; a loaded game keeps its own. The AI applies to every
// game loaded from then on: the competitors produceCompanies starts run it
// ("none": none start, as Config::companyAIDisabled does for upstream's).
// The title screen's game never gets competitors. Config::companyAIDisabled
// (the Miscellaneous tab's checkbox) still stops them all.

// aiSettings are the companyAI section; -1: the scenario's value.
type aiSettings struct {
	AI                     string // "" none
	Competitors, StartDelay int
}

// Scenario.h kMaxCompetingCompanies, kMaxCompetitorStartDelay.
const (
	maxCompetingCompanies    = 14
	maxCompetitorStartDelay  = 240
	aiSettingScenario        = -1
	aiSettingNone            = "none"
	aiSettingScenarioKeyword = "scenario"
)

func loadAISettings() aiSettings {
	s := aiSettings{AI: companyai.BusLinkName, Competitors: aiSettingScenario, StartDelay: aiSettingScenario}
	switch v := golocoConfigValue("companyAI.ai"); {
	case v == aiSettingNone:
		s.AI = ""
	case v != "":
		for _, n := range sim.AINames() {
			if n == v {
				s.AI = v
			}
		}
	}
	num := func(key string, limit int) int {
		n, err := strconv.Atoi(golocoConfigValue(key))
		if err != nil {
			return aiSettingScenario
		}
		return min(max(n, 0), limit)
	}
	s.Competitors = num("companyAI.competitors", maxCompetingCompanies)
	s.StartDelay = num("companyAI.startDelay", maxCompetitorStartDelay)
	return s
}

func (s aiSettings) save() error {
	ai := s.AI
	if ai == "" {
		ai = aiSettingNone
	}
	num := func(n int) string {
		if n == aiSettingScenario {
			return aiSettingScenarioKeyword
		}
		return strconv.Itoa(n)
	}
	return setGolocoConfigValues([][2]string{{"companyAI.ai", ai}, {"companyAI.competitors", num(s.Competitors)},
		{"companyAI.startDelay", num(s.StartDelay)}})
}

// applyNewGame is the settings at Scenario::start: the count and the delay
// that aren't the scenario's replace the game state's.
func (s aiSettings) applyNewGame(m *sim.Map) {
	if s.Competitors != aiSettingScenario {
		m.MaxCompetingCompanies = uint8(s.Competitors)
	}
	if s.StartDelay != aiSettingScenario {
		m.CompetitorStartDelay = uint8(s.StartDelay)
	}
}

// applyGame gives a loaded or new game its competitors' AI.
func (s aiSettings) applyGame(m *sim.Map) { m.Competitors = sim.Competitors{AI: s.AI} }
