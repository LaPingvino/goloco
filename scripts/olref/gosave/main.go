// Command gosave writes goloco's own saved game of a game, headless: the file
// loaded into goloco's map (a .SC5 started as a new game, as the New Game
// list does), played -ticks game ticks, and saved through goloco's S5 writer
// (pkg/scenario S5File, sim.State.ExportScenario) — for loading goloco's
// saves in the real OpenLoco (scripts/olref) and in goloco.
//
//	go run ./scripts/olref/gosave [-ticks N] [-seed S] [-ai NAME [-ais N] [-competitors]] ../locomotion/Data/title.dat $SCRATCH/title-goloco.SV5
//
// No preview image (the save details' 250×200 picture is blank: there is no
// renderer here). The save is game
// data: keep it out of git.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/LaPingvino/goloco/pkg/assets"
	_ "github.com/LaPingvino/goloco/pkg/companyai" // registers goloco's AIs
	"github.com/LaPingvino/goloco/pkg/core"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
)

func main() {
	ticks := flag.Int("ticks", 0, "game ticks to play before saving")
	monthly := flag.Bool("monthly", false, "also save at each month's start, as OpenLoco's monthly autosave (out-YYYY-MM.SV5 beside out)")
	seed := flag.Uint("seed", 1950, "a .SC5's new-game rng seed (the script's \"seed N\")")
	data := flag.String("data", "../locomotion", "the Locomotion folder")
	otherTrace := flag.Bool("rngtrace-other", false, "with -rngtrace: also the draws of the other generators (a town's, an industry's: copies the code makes as it goes)")
	rngTrace := flag.Int("rngtrace", 0, "print each main-rng (gPrng1) draw of this tick of the run (1-based) with its callers, as scripts/olref/rngtrace.patch does in OpenLoco")
	aiName := flag.String("ai", "", "start AI companies running this goloco AI ("+strings.Join(sim.AINames(), ", ")+") before playing; their decision logs go to stderr at the end")
	aiCount := flag.Int("ais", 1, "with -ai: how many AI companies to start at once (0: none)")
	produce := flag.Bool("competitors", false, "with -ai: also let produceCompanies start competitors running it on schedule, as a game with the AI chosen in the options (draws the game's rng: not for simcmp)")
	flag.Parse()
	if flag.NArg() != 2 {
		log.Fatal("usage: gosave [-ticks N] [-seed S] [-data DIR] in.(SV5|SC5|dat) out.SV5")
	}
	in, out := flag.Arg(0), flag.Arg(1)
	g1, err := assets.LoadG1(filepath.Join(*data, "Data", "g1.DAT"))
	must(err)
	om := objects.NewObjectManager(filepath.Join(*data, "ObjData"))
	om.SetBaseSpriteIndex(uint32(g1.GetSpriteCount()))
	om.G1File = g1
	must(om.LoadAllObjects())
	sc, err := scenario.LoadScenarioData(in)
	must(err)
	for _, po := range sc.PackedObjectsRaw {
		_, _ = om.LoadObjectFromHeaderAndData(po.HeaderBytes, po.Data)
	}
	sim.ReorderObjectsForScenario(om, sc)
	s := &sim.State{}
	s.LoadScenario(sc)
	s.CreateAllMapAnimations(om) // importGameState: the industries' animations
	s.LoadWaves(sc.Waves)
	if strings.EqualFold(filepath.Ext(in), ".sc5") {
		s.Rng1 = *core.NewPrngSeed(uint32(*seed), uint32(*seed)^0x5EED)
		var gen *sim.TerrainOptions
		if sc.LandscapeToGenerate {
			o := sc.Options
			gen = &sim.TerrainOptions{MinLandHeight: o.MinLandHeight, TopographyStyle: o.TopographyStyle, HillDensity: o.HillDensity,
				HillsEdgeOfMap: o.Flags&2 != 0, SeaLevel: sc.SeaLevel, NumRiverbeds: o.NumRiverbeds, MinRiverWidth: o.MinRiverWidth,
				MaxRiverWidth: o.MaxRiverWidth, RiverbankWidth: o.RiverbankWidth, RiverMeanderRate: o.RiverMeanderRate,
				LandDistributionPatterns: o.LandDistributionPatterns, NumberOfForests: o.NumberOfForests, NumberRandomTrees: o.NumberRandomTrees,
				MinForestRadius: o.MinForestRadius, MaxForestRadius: o.MaxForestRadius, MinForestDensity: o.MinForestDensity,
				MaxForestDensity: o.MaxForestDensity, MinAltitudeForTrees: o.MinAltitudeForTrees, MaxAltitudeForTrees: o.MaxAltitudeForTrees,
				NumberOfTowns: o.NumberOfTowns, MaxTownSize: o.MaxTownSize, NumberOfIndustries: o.OptNumberOfIndustries}
		}
		var raw sim.G1Raw
		if r, ok := om.G1File.(sim.G1Raw); ok {
			raw = r
		}
		s.StartScenarioGenerating(sc.Options.StartYear, gen, raw, om)
		s.ScenarioFileName, s.ScenarioName, s.ScenarioDetails = in, sc.Options.Name, sc.Options.Description
	}
	if *monthly { // GameScene's autosaveCheck: inside the tick, after the monthly updates — where OpenLoco autosaves
		s.AutosaveCheck = func() {
			name := strings.TrimSuffix(out, filepath.Ext(out)) + fmt.Sprintf("-%04d-%02d.SV5", s.CurrentYear, int(s.CurrentMonth)+1)
			writeSave(s, sc, om, name, *ticks)
		}
	}
	if *produce && *aiName != "" {
		s.Competitors = sim.Competitors{AI: *aiName}
	}
	for i := 0; *aiName != "" && i < *aiCount; i++ {
		if s.CreateAICompany(*aiName, sim.AIOptions{}, om) == 0xFF {
			log.Fatalf("gosave: no AI company made for %q", *aiName)
		}
	}
	tracing := false
	core.TraceHook = func(p *core.Prng) {
		switch {
		case !tracing:
		case p == &s.Rng1:
			traceDraw("RNG")
		case *otherTrace && p != &s.Rng2:
			traceDraw("OTHER")
		}
	}
	for i := 1; i <= *ticks; i++ {
		tracing = i == *rngTrace
		s.Tick(offScreen, om)
	}
	core.TraceHook = nil
	s.AutosaveCheck = nil
	for id := range 15 {
		if s.CompanyAIName(id) == "" {
			continue
		}
		for _, d := range s.AIDecisions(id) {
			fmt.Fprintf(os.Stderr, "ai %d tick %d: %s (ok %v cost %d err %d)\n", id, d.Tick, d.Text, d.OK, d.Cost, d.Err)
		}
	}
	writeSave(s, sc, om, out, *ticks)
}

// writeSave is S5::exportGameStateToFile: the pre-save clean-up, then the
// file (header, save details, packed objects, object list, game state, tiles).
func writeSave(s *sim.State, sc *scenario.Scenario, om *objects.ObjectManager, out string, ticks int) {
	s.PrepareSave()
	ex := s.ExportScenario()
	ex.SavedViewX, ex.SavedViewY, ex.SavedViewZoom, ex.SavedViewRotation = sc.SavedViewX, sc.SavedViewY, sc.SavedViewZoom, sc.SavedViewRotation
	id := int(s.PlayerCompanies[0])
	d := scenario.SaveDetails{Date: s.CurrentDay, Image: make([]byte, 250*200)}
	var perf uint16
	if id < len(s.CompanyNames) {
		d.Company, d.Owner = s.CompanyName(id), s.OwnerName(id, om)
		perf, d.ChallengeFlags = uint16(s.CompanyFaces[id].Performance), s.CompanyFaces[id].ChallengeFlags
		d.ChallengeProgress = s.Companies[id].ChallengeProgress
	}
	f := &scenario.S5File{Header: scenario.SavedGameHeader(len(sc.PackedObjectsRaw)), SaveDetails: scenario.EncodeSaveDetails(d, perf, s.ScenarioName),
		PackedObjects: sc.PackedObjectsRaw, RequiredObjects: sim.RequiredObjects(sc.Raw.RequiredObjects, om),
		GameState: ex.EncodeGameState(), TileElements: scenario.EncodeTileElements(ex.Tiles), GolocoAI: ex.GolocoAI}
	var buf bytes.Buffer
	must(f.Write(&buf))
	must(os.WriteFile(out, buf.Bytes(), 0o644))
	log.Printf("gosave: %d ticks → %s (%d bytes)", ticks, out, buf.Len())
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// traceDraw prints a draw's callers (past RandNext and its bounded forms).
func traceDraw(kind string) {
	pc := make([]uintptr, 8)
	n := runtime.Callers(3, pc)
	frames := runtime.CallersFrames(pc[:n])
	out := kind
	for k := 0; k < 4; k++ {
		f, more := frames.Next()
		name := f.Function[strings.LastIndex(f.Function, "/")+1:]
		if strings.HasPrefix(name, "core.(*Prng).") {
			k--
		} else {
			out += " " + name
		}
		if !more {
			break
		}
	}
	fmt.Fprintln(os.Stderr, out)
}

// offScreen is what a headless game shows: nothing (OpenLoco's simulate has
// no window, so WaveManager::createWave's findWindowShowing finds none and
// no wave starts, nor is gPrng2 drawn for one).
func offScreen(int, int) bool { return false }
