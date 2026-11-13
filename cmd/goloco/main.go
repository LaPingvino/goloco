package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/LaPingvino/goloco/pkg/audio"
	"github.com/LaPingvino/goloco/pkg/sim"
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
		useVirtualClock()
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
	if game.script != nil && game.script.failed {
		os.Exit(1) // a playthrough's expect failed
	}
}

// openTrackTest is GOLOCO_OPEN=track, the headless construction test: the
// window open and a line of straights laid from the map centre (sim.LayTrack:
// no game command), a station on its first and last piece.
func (g *Game) openTrackTest() {
	g.buildMode = buildModeTrack
	g.wins.Show(g.newPortedConstructionWindow())
	mw, mh := g.w.GetMapSize()
	t := g.w.TileAt(mw/2, mh/2)
	if t == nil {
		return
	}
	end := sim.ConnectionEnd{X: int32(mw/2) * 32, Y: int32(mh/2) * 32, Z: int32(t.BaseZ) * 4}
	var firstX, firstY, lastX, lastY int
	for i := 0; i < 5; i++ {
		lastX, lastY = int(end.X/32), int(end.Y/32)
		if i == 0 {
			firstX, firstY = lastX, lastY
		}
		next, ok := g.w.LayTrack(end.X, end.Y, end.Z, end.NextRotation, 0, 0)
		if !ok {
			log.Printf("[ConsTest] no straight at %+v", end)
			break
		}
		end = next
	}
	log.Printf("[ConsTest] the line ends at %+v", end)
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
// straight pieces and very small curves chained (sim.LayRoad).
func (g *Game) openRoadTest() {
	g.buildMode = buildModeRoad
	g.wins.Show(g.newPortedConstructionWindow())
	mw, mh := g.w.GetMapSize()
	t := g.w.TileAt(mw/2, mh/2)
	if t == nil {
		return
	}
	end := sim.ConnectionEnd{X: int32(mw/2) * 32, Y: int32(mh/2) * 32, Z: int32(t.BaseZ) * 4}
	for _, id := range []int{0, 2, 0, 1, 0} {
		next, ok := g.w.LayRoad(end.X, end.Y, end.Z, end.NextRotation, id, 0)
		if !ok {
			log.Printf("[RoadConsTest] no piece %d at %+v", id, end)
			break
		}
		end = next
	}
	log.Printf("[RoadConsTest] the line ends at %+v", end)
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
