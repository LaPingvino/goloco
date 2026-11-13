package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/LaPingvino/goloco/pkg/audio"
	"github.com/hajimehoshi/ebiten/v2"
)

// main.go — the command line (--diag, --shots), the GOLOCO_OPEN / GOLOCO_SCRIPT
// start-up switches, and the Ebitengine run loop.

func main() {
	audio.ProbeChild() // the sound-device probe (item AU): exits when this process is it
	game := NewGame()

	// Parse CLI args:
	//   ./goloco [file]                    — load scenario normally
	//   ./goloco [file] --diag TX,TY       — save diag.png centred on tile TX,TY, quit
	//   ./goloco [file] --diag worst       — jump to tile with most render issues (dynamic)
	//   ./goloco [file] --diag water       — scan for nearest water tile, quit
	//   ./goloco [file] --diag building    — same for building / track / road / tree / station /
	//                                        industry / wall / train
	//   ./goloco [file] --diag             — save diag.png centred on map middle, quit
	//   ./goloco [file] --shots S,N       — save shot_01..NN.png every S seconds, quit
	args := os.Args[1:]
	var scenarioPath string
	diagTileX, diagTileY := -1, -1
	diagKind := ""
	diagMode := false
	for i := 0; i < len(args); i++ {
		if args[i] == "--shots" && i+1 < len(args) {
			var secs, count int
			if n, _ := fmt.Sscanf(args[i+1], "%d,%d", &secs, &count); n >= 1 && secs > 0 {
				if count <= 0 {
					count = 5
				}
				game.shotsEvery = time.Duration(secs) * time.Second
				game.shotsLeft = count
				game.shotsLast = time.Now()
			}
			i++
			continue
		}
		if args[i] != "--diag" {
			scenarioPath = args[i]
			continue
		}
		diagMode = true
		if i+1 >= len(args) {
			break
		}
		next := args[i+1]
		n, _ := fmt.Sscanf(next, "%d,%d", &diagTileX, &diagTileY)
		if n == 2 {
			i++ // consumed as TX,TY
		} else if len(next) > 0 && next[0] != '-' {
			diagKind = next // consumed as kind keyword (water, worst, building, etc.)
			i++
		}
	}

	if scenarioPath != "" {
		if err := game.loadScenario(scenarioPath); err != nil {
			log.Printf("Failed to load %s: %v", scenarioPath, err)
		}
	}

	if diagMode {
		if diagTileX < 0 {
			if diagKind == "" {
				diagKind = "worst"
			}
			game.w.CenterOn(diagKind) // handles "worst", "water", "building", etc.
			// Store the tile CenterOn snapped to so Draw() re-applies camera each
			// frame, overriding the title sequence tour camera.
			game.diagTileX = game.w.LastCenterTileX
			game.diagTileY = game.w.LastCenterTileY
			game.diagPending = true
			game.diagQuit = true
		} else {
			// Move camera to centre on the requested tile so it is
			// fully rendered when saveDiagCrop runs after Draw().
			game.w.SetCamera(float64(diagTileX), float64(diagTileY))
			game.diagPending = true
			game.diagQuit = true
			game.diagTileX = diagTileX
			game.diagTileY = diagTileY
		}
	}

	game.progress = loadProgress()
	if p := os.Getenv("GOLOCO_SCRIPT"); p != "" {
		game.script = loadInputScript(p)
	}
	if os.Getenv("GOLOCO_OPEN") == "newgame" {
		game.openNewGameWindow()
	}
	if os.Getenv("GOLOCO_OPEN") == "track" {
		game.openTrackTest()
	}
	if os.Getenv("GOLOCO_OPEN") == "road" {
		game.openRoadTest()
	}

	ebiten.SetWindowSize(defaultWidth, defaultHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("GoLoco - Locomotion Reimplementation")
	defer startCPUProfile()() // GOLOCO_CPUPROFILE=file (LOOP item 0)
	if err := ebiten.RunGame(game); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}
}

// openTrackTest is GOLOCO_OPEN=track, the headless construction test: the
// window open and a piece sequence chained from the map centre, so captures
// verify connections.
func (g *Game) openTrackTest() {
	g.buildMode = buildModeTrack
	g.wins.Show(g.newPortedConstructionWindow())
	mw, mh := g.w.GetMapSize()
	g.w.SetConstructionHead(mw/2, mh/2)
	// Straight line so the piece origins fall on a predictable tile row and
	// the first/last pieces can carry stations for the board→deliver loop.
	seq := []int{0, 0, 0, 0, 0}
	var firstX, firstY, lastX, lastY int
	for i, id := range seq {
		h := g.w.ConstructionHeadState()
		if i == 0 {
			firstX, firstY = h.X, h.Y
		}
		if i == len(seq)-1 {
			lastX, lastY = h.X, h.Y
		}
		if !g.w.PlaceTrackAtHead(id, 0) {
			log.Printf("[ConsTest] piece %d not placeable at head %+v", id, g.w.ConstructionHeadState())
		}
	}
	log.Printf("[ConsTest] head after sequence: %+v", g.w.ConstructionHeadState())
	// Stations on the first and last pieces of the line.
	stationSlot := g.firstTrainStationSlot()
	if g.w.PlaceStationOnTile(firstX, firstY, stationSlot) {
		log.Printf("[ConsTest] station A placed at (%d,%d)", firstX, firstY)
	}
	if g.w.PlaceStationOnTile(lastX, lastY, stationSlot) {
		log.Printf("[ConsTest] station B placed at (%d,%d)", lastX, lastY)
	}
}

// openRoadTest is GOLOCO_OPEN=road, the headless road construction test:
// straight pieces and very small curves chained.
func (g *Game) openRoadTest() {
	g.buildMode = buildModeRoad
	g.wins.Show(g.newPortedConstructionWindow())
	mw, mh := g.w.GetMapSize()
	g.w.SetConstructionHead(mw/2, mh/2)
	for _, id := range []int{0, 2, 0, 1, 0} {
		if !g.w.PlaceRoadAtHead(id, 0) {
			log.Printf("[RoadConsTest] piece %d not placeable at head %+v", id, g.w.ConstructionHeadState())
		}
	}
	log.Printf("[RoadConsTest] head after sequence: %+v", g.w.ConstructionHeadState())
}

// firstTrainStationSlot returns the first populated train-station object slot,
// or 0 when none are loaded. Used to pick a station type for placement.
func (g *Game) firstTrainStationSlot() uint8 {
	if g.r.ObjMgr == nil {
		return 0
	}
	for i := 0; i < 16; i++ {
		if g.r.ObjMgr.GetTrainStationObjectByIndex(i) != nil {
			return uint8(i)
		}
	}
	return 0
}
