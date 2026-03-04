// diagnostics reports how complete the GoLoco reimplementation is relative to OpenLoco.
//
// Run from any directory inside the goloco module:
//
//	go run ./cmd/diagnostics
//
// Output:
//   - Per-category completion bar (what fraction is implemented)
//   - Per-feature confidence bar (how accurate/complete the implementation likely is)
//   - Overall weighted score
//
// Confidence is estimated from static signals:
//   - Test coverage (0.35):  a test file contains a test for this feature
//   - OpenLoco refs  (0.25): the source file has "// OpenLoco reference:" comments
//   - TODO-free      (0.15): no TODO/FIXME/HACK in the primary file
//   - Bonus          (0.25): a specific correctness indicator is present
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// ─── detection primitives ────────────────────────────────────────────────────

// fileExists returns true if root/rel exists.
func fileExists(root, rel string) bool {
	_, err := os.Stat(filepath.Join(root, rel))
	return err == nil
}

// hasPattern returns true if any line in root/rel matches the regexp.
func hasPattern(root, rel, pattern string) bool {
	f, err := os.Open(filepath.Join(root, rel))
	if err != nil {
		return false
	}
	defer f.Close()
	re := regexp.MustCompile(pattern)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if re.MatchString(sc.Text()) {
			return true
		}
	}
	return false
}

// countMatches counts lines matching pattern across *.go files under root/dir,
// skipping the _conversion_generated subtree (reference artifacts, not game code).
func countMatches(root, dir, pattern string) int {
	re := regexp.MustCompile(pattern)
	n := 0
	_ = filepath.Walk(filepath.Join(root, dir), func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		// Skip reference artifacts — they are not compiled into the game.
		if fi.IsDir() && fi.Name() == "_conversion_generated" {
			return filepath.SkipDir
		}
		if fi.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			if re.MatchString(sc.Text()) {
				n++
			}
		}
		return nil
	})
	return n
}

// ─── confidence signal building blocks ───────────────────────────────────────
//
// Each signal returns its contribution (0 or its weight) based on a source
// code check.  They are composed with conf() which sums and clamps to [0, 1].

// sigTests returns 0.35 if testFile contains a line matching pattern.
func sigTests(testFile, pattern string) func(root string) float64 {
	return func(root string) float64 {
		if hasPattern(root, testFile, pattern) {
			return 0.35
		}
		return 0
	}
}

// sigOLRef returns 0.25 if file has an "OpenLoco reference:" comment.
func sigOLRef(file string) func(root string) float64 {
	return func(root string) float64 {
		if hasPattern(root, file, `OpenLoco reference:`) {
			return 0.25
		}
		return 0
	}
}

// sigNoTodo returns 0.15 if file has no TODO/FIXME/HACK lines.
func sigNoTodo(file string) func(root string) float64 {
	return func(root string) float64 {
		if !hasPattern(root, file, `TODO|FIXME|HACK`) {
			return 0.15
		}
		return 0
	}
}

// sigBonus returns 0.25 if file has a line matching pattern (specific
// correctness indicator: a known-good constant, algorithm, or fix).
func sigBonus(file, pattern string) func(root string) float64 {
	return func(root string) float64 {
		if hasPattern(root, file, pattern) {
			return 0.25
		}
		return 0
	}
}

// conf combines signal functions, clamping the result to [0, 1].
func conf(sigs ...func(string) float64) func(string) float64 {
	return func(root string) float64 {
		total := 0.0
		for _, s := range sigs {
			total += s(root)
		}
		if total > 1.0 {
			return 1.0
		}
		return total
	}
}

// ─── feature definition ───────────────────────────────────────────────────────

// Feature is a single checkable capability.
type Feature struct {
	Name       string
	Detect     func(root string) bool          // true if the feature is implemented
	Confidence func(root string) float64       // nil = unknown; else 0-1 accuracy estimate
	Note       string                          // printed alongside the feature name
}

// Category groups related features and carries a weight for the overall score.
type Category struct {
	Name     string
	Weight   float64
	Features []Feature
}

// ─── feature catalogue ────────────────────────────────────────────────────────

func catalogue() []Category {
	return []Category{

		// ── 1. Object parsers ─────────────────────────────────────────────────
		{
			Name:   "Object Parsers",
			Weight: 0.25,
			Features: []Feature{
				{
					Name:   "Land (terrain surface)",
					Detect: func(r string) bool { return hasPattern(r, "pkg/objects/land.go", `ParseLandObject`) },
					Confidence: conf(
						sigTests("pkg/world/world_test.go", "func Test"),
						sigOLRef("pkg/objects/land.go"),
						sigNoTodo("pkg/objects/land.go"),
						sigBonus("pkg/world/world.go", "growthStage"),
					),
				},
				{
					Name:   "Water",
					Detect: func(r string) bool { return hasPattern(r, "pkg/objects/water.go", `ParseWaterObject`) },
					Confidence: conf(
						sigOLRef("pkg/objects/water.go"),
						sigNoTodo("pkg/objects/water.go"),
						sigBonus("pkg/world/world.go", "KSlopeToWater|kSlopeToWater"),
					),
				},
				{
					Name:   "CliffEdge",
					Detect: func(r string) bool { return hasPattern(r, "pkg/objects/cliffedge.go", `ParseCliffEdge`) },
					Confidence: conf(
						sigTests("pkg/world/measurements_test.go", "func Test"),
						sigOLRef("pkg/objects/cliffedge.go"),
						sigNoTodo("pkg/objects/cliffedge.go"),
						sigBonus("pkg/world/world.go", "paintCliffEdges"),
					),
				},
				{
					Name:   "Tree",
					Detect: func(r string) bool { return hasPattern(r, "pkg/objects/tree.go", `ParseTreeObject`) },
					Confidence: conf(
						sigOLRef("pkg/objects/tree.go"),
						sigNoTodo("pkg/objects/tree.go"),
						sigBonus("pkg/world/world.go", "quadrant|Quadrant"),
					),
				},
				{
					Name:   "Building",
					Detect: func(r string) bool { return hasPattern(r, "pkg/objects/building.go", `ParseBuildingObject`) },
					Confidence: conf(
						sigOLRef("pkg/objects/building.go"),
						sigBonus("pkg/objects/building.go", "NumImagesPerVariation|numImagesPerVariation"),
					),
					Note: "non-constructed rendering incomplete",
				},
				{
					Name:   "Wall",
					Detect: func(r string) bool { return hasPattern(r, "pkg/objects/wall.go", `ParseWallObject`) },
					Confidence: conf(
						sigOLRef("pkg/objects/wall.go"),
						sigNoTodo("pkg/objects/wall.go"),
					),
				},
				{
					Name:   "Track",
					Detect: func(r string) bool { return hasPattern(r, "pkg/objects/track.go", `ParseTrackObject`) },
					Confidence: conf(
						sigOLRef("pkg/objects/track.go"),
						sigNoTodo("pkg/objects/track.go"),
						sigBonus("pkg/world/track_sprites.go", "rotateTrackPP"),
					),
				},
				{
					Name:   "Road",
					Detect: func(r string) bool { return hasPattern(r, "pkg/objects/road.go", `ParseRoadObject`) },
					Confidence: conf(
						sigTests("pkg/world/road_sprites_test.go", "func Test"),
						sigOLRef("pkg/objects/road.go"),
						sigNoTodo("pkg/world/road_sprites.go"),
					),
				},
				{
					Name:   "TrainStation",
					Detect: func(r string) bool { return hasPattern(r, "pkg/objects/trainstationobject.go", `ParseTrainStation`) },
					Confidence: conf(
						sigOLRef("pkg/objects/trainstationobject.go"),
						sigNoTodo("pkg/objects/trainstationobject.go"),
						sigBonus("pkg/world/world.go", "paintStation"),
					),
				},
				{
					Name:   "Vehicle",
					Detect: func(r string) bool { return hasPattern(r, "pkg/objects/vehicle.go", `ParseVehicleObject`) },
					Confidence: conf(
						sigOLRef("pkg/objects/vehicle.go"),
						sigNoTodo("pkg/objects/vehicle.go"),
					),
					Note: "sprites loaded, world rendering partial",
				},
				{
					Name:   "InterfaceSkin",
					Detect: func(r string) bool { return hasPattern(r, "pkg/objects/interfaceskin.go", `ParseInterface`) },
					Confidence: conf(
						sigOLRef("pkg/objects/interfaceskin.go"),
						sigNoTodo("pkg/objects/interfaceskin.go"),
					),
				},
				{Name: "Bridge", Detect: func(r string) bool { return fileExists(r, "pkg/objects/bridge.go") }, Note: "needed for elevated roads/tracks"},
				{Name: "Tunnel", Detect: func(r string) bool { return fileExists(r, "pkg/objects/tunnel.go") }},
				{Name: "TrainSignal", Detect: func(r string) bool { return fileExists(r, "pkg/objects/trainsignal.go") }},
				{Name: "LevelCrossing", Detect: func(r string) bool { return fileExists(r, "pkg/objects/levelcrossing.go") }},
				{Name: "StreetLight", Detect: func(r string) bool { return fileExists(r, "pkg/objects/streetlight.go") }},
				{Name: "Industry", Detect: func(r string) bool { return fileExists(r, "pkg/objects/industry.go") }},
				{Name: "Airport", Detect: func(r string) bool { return fileExists(r, "pkg/objects/airport.go") }},
				{Name: "Dock", Detect: func(r string) bool { return fileExists(r, "pkg/objects/dock.go") }},
				{Name: "RoadStation", Detect: func(r string) bool { return fileExists(r, "pkg/objects/roadstation.go") }},
				{Name: "TrackExtra (mods)", Detect: func(r string) bool { return fileExists(r, "pkg/objects/trackextra.go") }},
				{Name: "RoadExtra (mods)", Detect: func(r string) bool { return fileExists(r, "pkg/objects/roadextra.go") }},
				{Name: "Snow / Climate / HillShapes", Detect: func(r string) bool {
					return fileExists(r, "pkg/objects/snow.go") || fileExists(r, "pkg/objects/climate.go")
				}},
			},
		},

		// ── 2. Tile element rendering ─────────────────────────────────────────
		{
			Name:   "Tile Element Rendering",
			Weight: 0.30,
			Features: []Feature{
				{
					Name:   "Surface (terrain + slopes)",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/world.go", `paintCliffEdges|drawY.*baseZ`) },
					Confidence: conf(
						sigTests("pkg/world/world_test.go", "func Test"),
						sigOLRef("pkg/world/world.go"),
						sigBonus("pkg/world/world.go", "NumImagesPerGrowthStage|numImagesPerGrowthStage"),
					),
				},
				{
					Name:   "Cliff edges",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/world.go", `func.*paintCliffEdges`) },
					Confidence: conf(
						sigTests("pkg/world/measurements_test.go", "func Test"),
						sigBonus("pkg/world/world.go", "drawYFlat"),
					),
				},
				{
					Name:   "Water overlay",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/world.go", `func.*paintWater`) },
					Confidence: conf(
						sigOLRef("pkg/world/world.go"),
						sigBonus("pkg/world/world.go", "KSlopeToWater|kSlopeToWater|waterLevel.*16"),
					),
				},
				{
					Name:   "Trees",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/world.go", `func.*paintTrees`) },
					Confidence: conf(
						sigOLRef("pkg/world/world.go"),
						sigBonus("pkg/world/world.go", "quadrant|Quadrant"),
					),
				},
				{
					Name:   "Buildings",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/world.go", `func.*paintBuildings`) },
					Confidence: conf(
						sigOLRef("pkg/world/world.go"),
						sigBonus("pkg/world/world.go", "variation|seqIdx"),
					),
					Note: "non-constructed rendering skipped",
				},
				{
					Name:   "Walls/fences",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/world.go", `func.*paintWalls`) },
					Confidence: conf(
						sigOLRef("pkg/world/world.go"),
					),
				},
				{
					Name:   "Track (railway)",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/world.go", `func.*paintTracks`) },
					Confidence: conf(
						sigOLRef("pkg/world/world.go"),
						sigBonus("pkg/world/track_sprites.go", "rotateTrackPP"),
					),
				},
				{
					Name:   "Road",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/world.go", `func.*paintRoads`) },
					Confidence: conf(
						sigTests("pkg/world/road_sprites_test.go", "Style0|Style2"),
						sigOLRef("pkg/world/world.go"),
					),
				},
				{
					Name:   "Road junction/merge system",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/road_sprites.go", `kExitsToMergeId`) },
					Confidence: conf(
						sigTests("pkg/world/road_sprites_test.go", "Merge|Junction|Crossroads"),
						sigBonus("pkg/world/road_sprites.go", "kRoadMergeExits"),
					),
				},
				{
					Name:   "Water cliff edges",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/world.go", `func.*paintWaterCliffEdges`) },
					Confidence: conf(
						sigOLRef("pkg/world/world.go"),
						sigBonus("pkg/world/world.go", "paintSurfaceWaterCliffEdge|waterMicroZ"),
					),
				},
				{
					Name:   "Train station",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/world.go", `func.*paintStation`) },
					Confidence: conf(
						sigOLRef("pkg/world/world.go"),
						sigBonus("pkg/world/world.go", "SequenceImageOffsets"),
					),
				},
				{
					Name:   "Vehicles (entities)",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/world.go", `func.*paintVehicle`) },
					Confidence: conf(
						sigOLRef("pkg/world/world.go"),
						sigBonus("pkg/world/world.go", "VehicleHead|vehicleHead"),
					),
					Note: "entity sprites rendered, movement/animation not yet",
				},
				{Name: "Signals", Detect: func(r string) bool { return hasPattern(r, "pkg/world/world.go", `func.*paintSignal`) }, Note: "PaintSignal.cpp unimplemented"},
				{Name: "Industries", Detect: func(r string) bool { return hasPattern(r, "pkg/world/world.go", `func.*paintIndustr`) }, Note: "PaintIndustry.cpp unimplemented"},
				{Name: "Bridge rendering", Detect: func(r string) bool { return hasPattern(r, "pkg/world/world.go", `func.*paintBridge`) }, Note: "PaintBridge.cpp unimplemented"},
				{Name: "Road stations", Detect: func(r string) bool { return hasPattern(r, "pkg/world/world.go", `func.*paintRoadStation`) }, Note: "PaintRoadStation.cpp unimplemented"},
				{Name: "Airport / Dock", Detect: func(r string) bool {
					return hasPattern(r, "pkg/world/world.go", `func.*paintAirport|func.*paintDock`)
				}, Note: "PaintAirport/Docks.cpp unimplemented"},
			},
		},

		// ── 3. Road sprite system ─────────────────────────────────────────────
		{
			Name:   "Road Sprite System",
			Weight: 0.10,
			Features: []Feature{
				{
					Name:   "Style0 table (10 road IDs)",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/road_sprites.go", `kRoadPartsStyle0`) },
					Confidence: conf(
						sigTests("pkg/world/road_sprites_test.go", "Style0"),
						sigOLRef("pkg/world/road_sprites.go"),
						sigNoTodo("pkg/world/road_sprites.go"),
					),
				},
				{
					Name:   "Style1 table (ballast/sleeper/rail)",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/road_sprites.go", `kRoadPartsStyle1`) },
					Confidence: conf(
						sigOLRef("pkg/world/road_sprites.go"),
					),
				},
				{
					Name:   "Style2 table (dual carriageway)",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/road_sprites.go", `kRoadPartsStyle2`) },
					Confidence: conf(
						sigTests("pkg/world/road_sprites_test.go", "Style2"),
						sigOLRef("pkg/world/road_sprites.go"),
					),
					Note: "NE/SE may differ from SW/NW in dual carriageway",
				},
				{
					Name:   "Junction / T-junction / crossroads merge",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/road_sprites.go", `kExitsToMergeId`) },
					Confidence: conf(
						sigTests("pkg/world/road_sprites_test.go", "Merge|Junction|Crossroads"),
						sigBonus("pkg/world/road_sprites.go", "kMergeBaseOffset"),
					),
				},
				{
					Name:   "Slope pieces (SlopeUp/Down, SteepSlope)",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/road_sprites.go", `SlopeUp|SlopeDown`) },
					Confidence: conf(
						sigOLRef("pkg/world/road_sprites.go"),
					),
				},
				{Name: "Bridge fallback rendering", Detect: func(r string) bool { return hasPattern(r, "pkg/world/world.go", `hasBridge.*road|roadBridge`) }, Note: "elevated roads on bridges"},
			},
		},

		// ── 4. Track sprite system ────────────────────────────────────────────
		{
			Name:   "Track Sprite System",
			Weight: 0.10,
			Features: []Feature{
				{
					Name:   "Straight pieces",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/track_sprites.go", `Straight`) },
					Confidence: conf(sigOLRef("pkg/world/track_sprites.go"), sigBonus("pkg/world/track_sprites.go", "rotateTrackPP")),
				},
				{
					Name:   "Small curves (4 pieces each)",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/track_sprites.go", `CurveSmall`) },
					Confidence: conf(sigOLRef("pkg/world/track_sprites.go"), sigBonus("pkg/world/track_sprites.go", "rotateTrackPP")),
				},
				{
					Name:   "Large curves (4 pieces each)",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/track_sprites.go", `CurveLarge|LargeCurve`) },
					Confidence: conf(sigOLRef("pkg/world/track_sprites.go"), sigBonus("pkg/world/track_sprites.go", "rotateTrackPP")),
				},
				{
					Name:   "Slope pieces (SlopeUp/Down)",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/track_sprites.go", `SlopeUp`) },
					Confidence: conf(sigOLRef("pkg/world/track_sprites.go"), sigBonus("pkg/world/track_sprites.go", "rotateTrackPP")),
				},
				{
					Name:   "Steep slope pieces",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/track_sprites.go", `SteepSlope`) },
					Confidence: conf(sigOLRef("pkg/world/track_sprites.go"), sigBonus("pkg/world/track_sprites.go", "rotateTrackPP")),
				},
				{
					Name:   "Slope curve pieces (18 variants)",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/track_sprites.go", `SlopeCurve|slopeCurve|LeftCurveSmallSlopeUp`) },
					Confidence: conf(sigOLRef("pkg/world/track_sprites.go"), sigBonus("pkg/world/track_sprites.go", "rotateTrackPP")),
				},
				{
					Name:   "S-bends",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/track_sprites.go", `SBend|sBend`) },
					Confidence: conf(sigOLRef("pkg/world/track_sprites.go"), sigBonus("pkg/world/track_sprites.go", "rotateTrackPP")),
				},
				{Name: "Bridge fallback rendering", Detect: func(r string) bool { return hasPattern(r, "pkg/world/world.go", `hasBridge.*track|trackBridge`) }, Note: "elevated tracks on bridges"},
			},
		},

		// ── 5. Scenario / S5 loading ──────────────────────────────────────────
		{
			Name:   "S5 / Scenario Loading",
			Weight: 0.10,
			Features: []Feature{
				{
					Name:   "S5 file header + chunk dispatch",
					Detect: func(r string) bool { return hasPattern(r, "pkg/scenario/loader.go", `LoadScenarioData`) },
					Confidence: conf(
						sigTests("pkg/scenario/loader_test.go", "func Test"),
						sigOLRef("pkg/scenario/loader.go"),
						sigNoTodo("pkg/scenario/loader.go"),
					),
				},
				{
					Name:   "All 9 tile element types parsed",
					Detect: func(r string) bool { return hasPattern(r, "pkg/scenario/loader.go", `IndustryElement|industryElement`) },
					Confidence: conf(
						sigTests("pkg/scenario/coverage_test.go", "func Test"),
						sigOLRef("pkg/scenario/loader.go"),
					),
				},
				{
					Name:   "RequiredObjects slot map",
					Detect: func(r string) bool { return hasPattern(r, "pkg/scenario/loader.go", `parseObjectOrder|RequiredObjects`) },
					Confidence: conf(
						sigOLRef("pkg/scenario/loader.go"),
						sigBonus("pkg/scenario/loader.go", "trainStationSlotStart|roadStationSlotStart"),
					),
				},
				{
					Name:   "Element type mask fix (0x3C)",
					Detect: func(r string) bool { return hasPattern(r, "pkg/scenario/loader.go", `0x3C|elementTypeMask`) },
					Confidence: conf(
						sigOLRef("pkg/scenario/loader.go"),
						sigBonus("pkg/scenario/loader.go", `elementTypeMask\s*=\s*0x3C`),
					),
				},
				{
					Name:   "ScenarioOptions chunk (name/desc/preview)",
					Detect: func(r string) bool { return hasPattern(r, "pkg/scenario/loader.go", `parseOptions|ScenarioOptions`) },
					Confidence: conf(sigOLRef("pkg/scenario/loader.go")),
				},
				{
					Name:   "GameState chunk (SV5 entities)",
					Detect: func(r string) bool { return hasPattern(r, "pkg/scenario/loader.go", `GameState|parseEntityChunk|VehicleHead`) },
					Confidence: conf(
						sigOLRef("pkg/scenario/loader.go"),
						sigBonus("pkg/scenario/loader.go", "VehicleHead|entityTable|EntityType"),
					),
					Note: "companies/towns/stations not yet used",
				},
				{
					Name:   "PackedObjects chunk",
					Detect: func(r string) bool { return hasPattern(r, "pkg/scenario/loader.go", `PackedObject`) },
					Confidence: conf(sigOLRef("pkg/scenario/loader.go")),
				},
				{
					Name:   "GrowthStage terrain formula",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/world.go", `GrowthStage|growthStage`) },
					Confidence: conf(
						sigBonus("pkg/world/world.go", "NumImagesPerGrowthStage|numImagesPerGrowthStage"),
						sigBonus("pkg/objects/land.go", "NumImagesPerGrowthStage"),
					),
				},
			},
		},

		// ── 6. Game / UI systems ──────────────────────────────────────────────
		{
			Name:   "Game & UI Systems",
			Weight: 0.15,
			Features: []Feature{
				{
					Name:   "Title screen with camera animation",
					Detect: func(r string) bool { return hasPattern(r, "pkg/title/title.go", `func.*Start`) },
					Confidence: conf(
						sigTests("pkg/title/title_test.go", "func Test"),
						sigOLRef("pkg/title/title.go"),
						sigNoTodo("pkg/title/title.go"),
					),
				},
				{
					Name:   "Camera pan, zoom, edge-scroll",
					Detect: func(r string) bool { return hasPattern(r, "pkg/world/world.go", `func.*SetZoom|func.*SetCamera`) },
					Confidence: conf(
						sigTests("pkg/world/world_test.go", "Camera|Zoom|zoom"),
						sigOLRef("pkg/world/world.go"),
					),
				},
				{
					Name:   "Toolbar with dropdown menus",
					Detect: func(r string) bool { return fileExists(r, "pkg/ui/dropdown.go") && hasPattern(r, "pkg/ui/dropdown.go", `DropdownItem`) },
					Confidence: conf(
						sigOLRef("pkg/ui/dropdown.go"),
						sigNoTodo("pkg/ui/toolbar.go"),
					),
				},
				{
					Name:   "Draggable window system",
					Detect: func(r string) bool { return hasPattern(r, "pkg/ui/simplewindow.go", `HandleDrag`) },
					Confidence: conf(sigNoTodo("pkg/ui/simplewindow.go")),
				},
				{
					Name:   "Load game file browser",
					Detect: func(r string) bool { return hasPattern(r, "cmd/goloco/main.go", `openLoadGameWindow|openFileWindow`) },
					Confidence: conf(sigBonus("cmd/goloco/main.go", "openLoadGameWindow")),
				},
				{
					Name:   "Save game (.goloco JSON)",
					Detect: func(r string) bool { return hasPattern(r, "cmd/goloco/main.go", `saveGame|GoLocoSave`) },
					Confidence: conf(sigBonus("cmd/goloco/main.go", "GoLocoSave")),
				},
				{
					Name:   "Options dialog (sound/music)",
					Detect: func(r string) bool { return hasPattern(r, "cmd/goloco/main.go", `openOptionsWindow`) },
					Confidence: conf(sigBonus("cmd/goloco/main.go", "soundEnabled|musicEnabled")),
				},
				{
					Name:   "Music playback",
					Detect: func(r string) bool { return hasPattern(r, "pkg/audio/audio.go", `LoadAndPlayMusic|PlayMusic`) },
					Confidence: conf(sigOLRef("pkg/audio/audio.go")),
				},
				{
					Name:   "Quit to Menu restores title world",
					Detect: func(r string) bool { return hasPattern(r, "cmd/goloco/main.go", `quitToMenu|titleWorld`) },
					Confidence: conf(
						sigTests("pkg/title/title_test.go", "Restart"),
						sigBonus("cmd/goloco/main.go", "titleWorld"),
					),
				},
				{Name: "Sound effects", Detect: func(r string) bool { return hasPattern(r, "pkg/audio/audio.go", `PlaySound|SoundEffect`) }, Note: "CSSx.DAT files not loaded"},
				{Name: "Game clock (date/time)", Detect: func(r string) bool { return countMatches(r, "pkg/game", `GameDate|GameClock|Year.*Month.*Day`) > 0 }, Note: "year/month/day tracking"},
				{Name: "Economy (income/expenses)", Detect: func(r string) bool { return countMatches(r, "pkg/game", `PayLoan|income|profit|balance`) > 2 }, Note: "cargo delivery, company finances"},
				{Name: "Pathfinding / route calculation", Detect: func(r string) bool { return fileExists(r, "pkg/pathfinding") || fileExists(r, "pkg/ai/pathfinding.go") }, Note: "required for vehicle movement"},
			},
		},
	}
}

// ─── rendering ────────────────────────────────────────────────────────────────

const (
	colReset  = "\033[0m"
	colGreen  = "\033[32m"
	colRed    = "\033[31m"
	colYellow = "\033[33m"
	colCyan   = "\033[36m"
	colBold   = "\033[1m"
)

func pct(done, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(done) / float64(total) * 100
}

// completionBar returns a wide 20-block bar for the category completion ratio.
func completionBar(done, total, width int) string {
	if total == 0 {
		return strings.Repeat("░", width)
	}
	filled := done * width / total
	if filled > width {
		filled = width
	}
	return colGreen + strings.Repeat("█", filled) + colReset + strings.Repeat("░", width-filled)
}

// confidenceColor returns the appropriate color for a confidence value.
func confidenceColor(score float64) string {
	if score >= 0.75 {
		return colGreen
	}
	if score >= 0.45 {
		return colYellow
	}
	return colRed
}

// miniBar returns a compact 5-block confidence bar with color.
func miniBar(score float64) string {
	filled := int(score*5 + 0.5)
	if filled > 5 {
		filled = 5
	}
	col := confidenceColor(score)
	return col + strings.Repeat("▓", filled) + colReset + strings.Repeat("░", 5-filled)
}

func trafficLight(ratio float64) string {
	switch {
	case ratio >= 0.8:
		return colGreen + "●" + colReset
	case ratio >= 0.4:
		return colYellow + "●" + colReset
	default:
		return colRed + "●" + colReset
	}
}

// ─── main ─────────────────────────────────────────────────────────────────────

func findRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "."
	}
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	wd, _ := os.Getwd()
	return wd
}

func main() {
	root := findRoot()
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	cats := catalogue()

	// Header
	fmt.Printf("\n%s%s GoLoco Implementation Diagnostics%s\n", colBold, colCyan, colReset)
	fmt.Printf("Root: %s\n", root)
	fmt.Println(strings.Repeat("─", 70))

	overallScore := 0.0
	totalWeight := 0.0

	for _, cat := range cats {
		done := 0
		for _, f := range cat.Features {
			if f.Detect(root) {
				done++
			}
		}
		ratio := float64(done) / float64(len(cat.Features))
		overallScore += ratio * cat.Weight
		totalWeight += cat.Weight

		// Compute average confidence over implemented features.
		confSum := 0.0
		confCount := 0
		for _, f := range cat.Features {
			if f.Detect(root) && f.Confidence != nil {
				confSum += f.Confidence(root)
				confCount++
			}
		}

		// Category header
		fmt.Printf("\n%s%s%s  [%s]  %d/%d (%.0f%%)",
			colBold, cat.Name, colReset,
			completionBar(done, len(cat.Features), 20),
			done, len(cat.Features), pct(done, len(cat.Features)),
		)
		if confCount > 0 {
			avg := confSum / float64(confCount)
			col := confidenceColor(avg)
			fmt.Printf("  · avg accuracy %s%.0f%%%s", col, avg*100, colReset)
		}
		fmt.Println()

		// Features in definition order (not sorted), ✓ interleaved with ✗.
		for _, f := range cat.Features {
			if f.Detect(root) {
				note := ""
				if f.Note != "" {
					note = "  " + colYellow + "⚠ " + f.Note + colReset
				}
				if f.Confidence != nil {
					score := f.Confidence(root)
					col := confidenceColor(score)
					fmt.Printf("  %s✓%s %-40s [%s] %s%.0f%%%s%s\n",
						colGreen, colReset,
						f.Name,
						miniBar(score),
						col, score*100, colReset,
						note,
					)
				} else {
					fmt.Printf("  %s✓%s %s%s\n", colGreen, colReset, f.Name, note)
				}
			} else {
				note := ""
				if f.Note != "" {
					note = " (" + f.Note + ")"
				}
				fmt.Printf("  %s✗%s %s%s\n", colRed, colReset, f.Name, note)
			}
		}
	}

	// Overall score
	if totalWeight > 0 {
		overallScore /= totalWeight
	}
	overallPct := overallScore * 100
	fmt.Printf("\n%s%s Overall score: %.1f%%%s  %s\n",
		colBold, colCyan, overallPct, colReset,
		trafficLight(overallScore),
	)
	fmt.Printf("[%s]\n", completionBar(int(overallPct), 100, 40))
	fmt.Println()

	switch {
	case overallPct < 30:
		fmt.Println("Foundation stage — core parsing and basic rendering in progress.")
	case overallPct < 55:
		fmt.Println("Rendering stage — tile visuals mostly working, game systems next.")
	case overallPct < 75:
		fmt.Println("Systems stage — key gameplay loops need implementing.")
	default:
		fmt.Println("Polish stage — most systems present, filling gaps and fixing edge cases.")
	}
	fmt.Printf("\nConfidence key: %s[▓▓▓▓▓]%s high (≥75%%)  %s[▓▓▓░░]%s medium (45-75%%)  %s[▓▓░░░]%s low (<45%%)\n\n",
		colGreen, colReset, colYellow, colReset, colRed, colReset)
}
