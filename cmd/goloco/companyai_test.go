package main

import (
	"os"
	"strings"
	"testing"

	"github.com/LaPingvino/goloco/pkg/companyai"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// The competitors' settings: buslink and the scenario's count and delay by
// default; kept in goloco.yml, never in openloco.yml; a new game takes the
// count and delay that aren't the scenario's, every game the AI.
func TestAISettings(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	s := loadAISettings()
	if s != (aiSettings{AI: companyai.BusLinkName, Competitors: -1, StartDelay: -1}) {
		t.Fatalf("defaults %+v", s)
	}
	m := &sim.Map{MaxCompetingCompanies: 7, CompetitorStartDelay: 12}
	s.applyNewGame(m)
	s.applyGame(m)
	if m.MaxCompetingCompanies != 7 || m.CompetitorStartDelay != 12 || m.Competitors.AI != companyai.BusLinkName {
		t.Errorf("defaults applied: %d, %d, %q", m.MaxCompetingCompanies, m.CompetitorStartDelay, m.Competitors.AI)
	}
	if err := (aiSettings{AI: "", Competitors: 3, StartDelay: 0}).save(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(configPath()); err == nil {
		t.Error("the AI settings went into openloco.yml")
	}
	data, _ := os.ReadFile(golocoConfigPath())
	if !strings.Contains(string(data), "companyAI:\n  ai: none\n  competitors: 3\n  startDelay: 0\n") {
		t.Errorf("goloco.yml:\n%s", data)
	}
	s = loadAISettings()
	s.applyNewGame(m)
	s.applyGame(m)
	if m.MaxCompetingCompanies != 3 || m.CompetitorStartDelay != 0 || m.Competitors.AI != "" {
		t.Errorf("saved settings applied: %d, %d, %q", m.MaxCompetingCompanies, m.CompetitorStartDelay, m.Competitors.AI)
	}
	setGolocoConfigValues([][2]string{{"companyAI.ai", "nosuchai"}, {"companyAI.competitors", "99"}, {"companyAI.startDelay", "scenario"}})
	if s := loadAISettings(); s != (aiSettings{AI: companyai.BusLinkName, Competitors: 14, StartDelay: -1}) {
		t.Errorf("an unknown AI, too many competitors: %+v", s)
	}
}
