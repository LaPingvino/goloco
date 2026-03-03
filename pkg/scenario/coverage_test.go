package scenario

// coverage_test.go — iterates every .SC5 file in the Locomotion Scenarios
// directory, parses it, and scores how well the current implementation handles
// each scenario.  The test produces a sorted report (worst-first) with an
// overall "implementation coverage" percentage.
//
// Run with:
//
//	go test ./pkg/scenario/ -v -run TestScenarioCoverage
//
// The test skips gracefully when no Locomotion data directory is found.

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// findLocoScenariosDir searches common install locations for the Locomotion
// Scenarios directory and returns the first one found.  It also checks the
// LOCO_SCENARIOS_DIR environment variable.
func findLocoScenariosDir() string {
	if env := os.Getenv("LOCO_SCENARIOS_DIR"); env != "" {
		return env
	}
	home := os.Getenv("HOME")
	candidates := []string{
		filepath.Join(home, "locomotion/Scenarios"),
		filepath.Join(home, ".local/share/Steam/steamapps/common/Locomotion/Scenarios"),
		filepath.Join(home, ".steam/steam/steamapps/common/Locomotion/Scenarios"),
		"/usr/share/games/locomotion/Scenarios",
		"../../../locomotion/Scenarios",
		"../../../../locomotion/Scenarios",
	}
	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && fi.IsDir() {
			return c
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// Scoring
// ---------------------------------------------------------------------------

// maxScore is the total points a perfect scenario would achieve.
const maxScoreTotal = 100

// scenarioReport holds one scenario's score and diagnostic notes.
type scenarioReport struct {
	name    string
	score   int
	issues  []string // ⚠ something wrong / unimplemented
	notes   []string // · informational
}

func (r *scenarioReport) pct() float64 {
	return float64(r.score) / float64(maxScoreTotal) * 100
}

func (r *scenarioReport) add(pts int, msg string) {
	r.score += pts
	if msg != "" {
		r.notes = append(r.notes, fmt.Sprintf("+%d %s", pts, msg))
	}
}

func (r *scenarioReport) warn(msg string) {
	r.issues = append(r.issues, msg)
}

// ---------------------------------------------------------------------------
// Per-scenario evaluation
// ---------------------------------------------------------------------------

func evaluateScenario(filePath string) scenarioReport {
	name := filepath.Base(filePath)
	r := scenarioReport{name: name}

	// --- 1. Parse (20 pts) ---
	sc, err := LoadScenarioData(filePath)
	if err != nil {
		r.warn(fmt.Sprintf("PARSE ERROR: %v", err))
		return r
	}
	r.add(20, "parsed without error")

	// --- 2. Map dimensions (5 pts) ---
	if sc.MapWidth > 0 && sc.MapHeight > 0 {
		r.add(5, fmt.Sprintf("map %d×%d", sc.MapWidth, sc.MapHeight))
	} else {
		r.warn(fmt.Sprintf("zero map dimensions (%d×%d)", sc.MapWidth, sc.MapHeight))
	}

	if sc.Tiles == nil {
		r.warn("no tiles parsed")
		return r
	}

	// Collect per-tile statistics.
	var (
		totalTiles   int
		heightFreq   [256]int
		slopeFreq    [32]int
		terrainFreq  [32]int
		growthFreq   [8]int
		variationFreq [256]int
		waterTiles   int
		treeTiles    int
		buildTiles   int
		trackTiles   int
		roadTiles    int
	)
	for _, row := range sc.Tiles {
		for _, t := range row {
			totalTiles++
			heightFreq[t.Height]++
			slopeFreq[t.Slope&0x1F]++
			if t.TerrainIndex < 32 {
				terrainFreq[t.TerrainIndex]++
			}
			growthFreq[t.GrowthStage&0x07]++
			variationFreq[t.Variation]++
			if t.Water > 0 {
				waterTiles++
			}
			if len(t.Trees) > 0 {
				treeTiles++
			}
			if len(t.Buildings) > 0 {
				buildTiles++
			}
			if len(t.Tracks) > 0 {
				trackTiles++
			}
			if len(t.Roads) > 0 {
				roadTiles++
			}
		}
	}

	// --- 3. Terrain diversity (20 pts) ---
	terrainEntropy := shannonEntropy(terrainFreq[:], totalTiles)
	distinctTerrain := countNonZero(terrainFreq[:])

	if distinctTerrain >= 2 {
		r.add(10, fmt.Sprintf("terrain diversity: %d distinct indices", distinctTerrain))
	} else {
		r.warn(fmt.Sprintf("only %d distinct terrain index — all tiles look the same", distinctTerrain))
	}
	if terrainEntropy >= 1.0 {
		r.add(10, fmt.Sprintf("terrain entropy %.2f bits", terrainEntropy))
	} else if terrainEntropy >= 0.1 {
		r.add(5, fmt.Sprintf("terrain entropy %.2f bits (low — single type dominates)", terrainEntropy))
		r.warn("low terrain entropy: one terrain type covers most tiles")
	} else {
		r.warn(fmt.Sprintf("degenerate terrain entropy %.2f bits — repeating pattern bug", terrainEntropy))
	}

	// --- 4. Height variation (10 pts) ---
	minH, maxH := indexRange(heightFreq[:])
	heightRange := maxH - minH
	if heightRange >= 20 {
		r.add(10, fmt.Sprintf("height range %d–%d (good relief)", minH, maxH))
	} else if heightRange >= 4 {
		r.add(5, fmt.Sprintf("height range %d–%d (modest relief)", minH, maxH))
	} else {
		r.warn(fmt.Sprintf("height range only %d (flat or height not parsed)", heightRange))
	}

	// --- 5. Slope variety (5 pts) ---
	distinctSlopes := countNonZero(slopeFreq[:])
	if distinctSlopes >= 5 {
		r.add(5, fmt.Sprintf("slope variety: %d distinct slopes", distinctSlopes))
	} else {
		r.warn(fmt.Sprintf("only %d distinct slope(s)", distinctSlopes))
	}

	// --- 6. Elements (15 pts: 3 pts each for trees, buildings, tracks/roads, water, land objects) ---
	if treeTiles > 0 {
		r.add(3, fmt.Sprintf("trees on %d tiles", treeTiles))
	} else {
		r.warn("no trees found (tree parsing may be broken, or bare scenario)")
	}
	if buildTiles > 0 {
		r.add(3, fmt.Sprintf("buildings on %d tiles", buildTiles))
	} else {
		r.warn("no buildings found")
	}
	if trackTiles > 0 || roadTiles > 0 {
		r.add(3, fmt.Sprintf("tracks: %d tiles  roads: %d tiles", trackTiles, roadTiles))
	} else {
		r.warn("no tracks or roads found")
	}
	if waterTiles > 0 {
		r.add(3, fmt.Sprintf("water on %d tiles (%.1f%%)", waterTiles, float64(waterTiles)/float64(totalTiles)*100))
	}

	// --- 7. Land object order (5 pts) ---
	landUsed := 0
	for _, s := range sc.LandObjectOrder {
		if s != "" {
			landUsed++
		}
	}
	if landUsed >= 2 {
		r.add(5, fmt.Sprintf("land objects: %d slots occupied in RequiredObjects", landUsed))
	} else if landUsed == 1 {
		r.add(2, "only 1 land object slot occupied — limited terrain variety")
	} else {
		r.warn("no land objects in RequiredObjects")
	}

	// --- 8. GrowthStage usage (5 pts) ---
	distinctGrowth := countNonZero(growthFreq[:])
	nonZeroGrowth := totalTiles - growthFreq[0]
	if distinctGrowth >= 2 {
		r.add(5, fmt.Sprintf("growth stages: %d distinct (%.1f%% of tiles non-zero)", distinctGrowth, float64(nonZeroGrowth)/float64(totalTiles)*100))
	} else {
		// Only stage 0 — normal for most temperate scenarios.
		r.notes = append(r.notes, "· growth stage: all tiles at stage 0 (normal for non-seasonal scenarios)")
	}

	// --- 9. Variation usage (5 pts: diagnostic — not yet rendered) ---
	nonZeroVariation := totalTiles - variationFreq[0]
	distinctVariation := countNonZero(variationFreq[:])
	if nonZeroVariation > 0 {
		// Non-zero variation is present but our renderer ignores byte 7 currently.
		r.warn(fmt.Sprintf("variation (byte 7) non-zero on %d tiles (%.1f%%) — %d distinct values — NOT YET RENDERED",
			nonZeroVariation, float64(nonZeroVariation)/float64(totalTiles)*100, distinctVariation))
	} else {
		r.add(5, "variation: all tiles use default style (byte 7 = 0)")
	}

	// --- Clamp score ---
	if r.score > maxScoreTotal {
		r.score = maxScoreTotal
	}

	return r
}

// ---------------------------------------------------------------------------
// Statistics helpers
// ---------------------------------------------------------------------------

// shannonEntropy computes H = -Σ p_i log2(p_i) for a frequency table.
func shannonEntropy(freq []int, total int) float64 {
	if total == 0 {
		return 0
	}
	var h float64
	for _, f := range freq {
		if f <= 0 {
			continue
		}
		p := float64(f) / float64(total)
		h -= p * math.Log2(p)
	}
	return h
}

// countNonZero returns how many entries in freq are > 0.
func countNonZero(freq []int) int {
	n := 0
	for _, f := range freq {
		if f > 0 {
			n++
		}
	}
	return n
}

// indexRange returns the lowest and highest indices that have freq > 0.
func indexRange(freq []int) (minIdx, maxIdx int) {
	minIdx, maxIdx = -1, -1
	for i, f := range freq {
		if f > 0 {
			if minIdx < 0 {
				minIdx = i
			}
			maxIdx = i
		}
	}
	if minIdx < 0 {
		minIdx = 0
	}
	if maxIdx < 0 {
		maxIdx = 0
	}
	return
}

// progressBar renders a Unicode progress bar of the given width.
func progressBar(pct float64, width int) string {
	filled := int(math.Round(pct / 100 * float64(width)))
	if filled > width {
		filled = width
	}
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
}

// ---------------------------------------------------------------------------
// Test entry point
// ---------------------------------------------------------------------------

func TestScenarioCoverage(t *testing.T) {
	scenDir := findLocoScenariosDir()
	if scenDir == "" {
		t.Skip("No Locomotion Scenarios directory found; set LOCO_SCENARIOS_DIR or install Locomotion at ~/locomotion/")
	}

	entries, err := os.ReadDir(scenDir)
	if err != nil {
		t.Fatalf("ReadDir %s: %v", scenDir, err)
	}

	var reports []scenarioReport
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.EqualFold(filepath.Ext(entry.Name()), ".sc5") {
			continue
		}
		path := filepath.Join(scenDir, entry.Name())
		reports = append(reports, evaluateScenario(path))
	}

	if len(reports) == 0 {
		t.Fatal("No .SC5 files found in " + scenDir)
	}

	// Sort worst-first so the most broken scenarios appear at the top.
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].pct() < reports[j].pct()
	})

	// Compute overall coverage score.
	var totalScore, totalMax int
	for range reports {
		totalMax += maxScoreTotal
	}
	for _, r := range reports {
		totalScore += r.score
	}
	overallPct := float64(totalScore) / float64(totalMax) * 100

	// Print report.
	sep := strings.Repeat("─", 72)
	t.Logf("")
	t.Logf("╔%s╗", strings.Repeat("═", 72))
	t.Logf("║  GoLoco Scenario Coverage Report — %d scenarios%s║",
		len(reports), strings.Repeat(" ", 72-37-len(fmt.Sprintf("%d", len(reports)))))
	t.Logf("║  Overall implementation coverage: %5.1f%%%s║",
		overallPct, strings.Repeat(" ", 72-37))
	t.Logf("╚%s╝", strings.Repeat("═", 72))
	t.Log(sep)

	for _, r := range reports {
		bar := progressBar(r.pct(), 24)
		t.Logf("%-42s %s %5.1f%%", r.name, bar, r.pct())
		for _, issue := range r.issues {
			t.Logf("    ⚠ %s", issue)
		}
		if testing.Verbose() {
			for _, note := range r.notes {
				t.Logf("    · %s", note)
			}
		}
	}

	t.Log(sep)
	t.Logf("Overall: %.1f%% (%d/%d pts across %d scenarios)",
		overallPct, totalScore, totalMax, len(reports))
	t.Logf("")

	// The test itself doesn't fail on low scores — it's a diagnostic report.
	// Add a soft floor: if the overall drops below 30%, something is very wrong.
	if overallPct < 30 {
		t.Errorf("Overall coverage %.1f%% < 30%% — major parsing regression", overallPct)
	}
}
