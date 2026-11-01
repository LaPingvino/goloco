// Command render8 renders a map region through the ported software renderer
// (pkg/gfx8 + pkg/paint) to a PNG, with no graphics library — the CPU
// reference the Ebiten renderer is compared against (LOOP item 3i, step 3).
// So far: land surfaces, cliff edges, water (no waves), trees, buildings and roads.
//
//	go run ./cmd/render8 -scenario ../locomotion/Data/title.dat -tile 196,190 -out surf.png
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/LaPingvino/goloco/pkg/assets"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/scene"
	"github.com/LaPingvino/goloco/pkg/sim"
)

func main() {
	data := flag.String("data", "../locomotion", "Locomotion data directory")
	scen := flag.String("scenario", "../locomotion/Data/title.dat", "scenario or save to render")
	tile := flag.String("tile", "196,190", "centre tile x,y")
	w := flag.Int("w", 480, "width")
	h := flag.Int("h", 360, "height")
	out := flag.String("out", "render8.png", "output PNG")
	window := flag.Int("window", 0, "only tiles within this many of the centre (0 = all that reach the view)")
	flag.Parse()

	var tx, ty int
	if _, err := fmt.Sscanf(*tile, "%d,%d", &tx, &ty); err != nil {
		log.Fatalf("render8: -tile: %v", err)
	}
	g1, err := assets.LoadG1(filepath.Join(*data, "Data", "g1.DAT"))
	must(err)
	om := objects.NewObjectManager(filepath.Join(*data, "ObjData"))
	om.SetBaseSpriteIndex(uint32(g1.GetSpriteCount()))
	om.G1File = g1
	must(om.LoadAllObjects())
	sc, err := scenario.LoadScenarioData(*scen)
	must(err)
	for _, po := range sc.PackedObjectsRaw {
		_, _ = om.LoadObjectFromHeaderAndData(po.HeaderBytes, po.Data)
	}
	if len(sc.LandObjectOrder) > 0 {
		om.ReorderLandObjects(sc.LandObjectOrder)
	}
	if len(sc.TreeObjectOrder) > 0 {
		om.ReorderTreeObjects(sc.TreeObjectOrder)
	}
	if len(sc.BuildingObjectOrder) > 0 {
		om.ReorderBuildingObjects(sc.BuildingObjectOrder)
	}
	if len(sc.RoadObjectOrder) > 0 {
		om.ReorderRoadObjects(sc.RoadObjectOrder)
	}
	if len(sc.WallObjectOrder) > 0 {
		om.ReorderWallObjects(sc.WallObjectOrder)
	}
	if len(sc.TrackObjectOrder) > 0 {
		om.ReorderTrackObjects(sc.TrackObjectOrder)
	}
	if len(sc.TrainSignalObjectOrder) > 0 {
		om.ReorderTrainSignalObjects(sc.TrainSignalObjectOrder)
	}
	if len(sc.TrainStationObjectOrder) > 0 {
		om.ReorderTrainStationObjects(sc.TrainStationObjectOrder)
	}
	if len(sc.IndustryObjectOrder) > 0 {
		om.ReorderIndustryObjects(sc.IndustryObjectOrder)
	}
	om.ReorderRoadStationObjects(sc.RoadStationObjectOrder)
	om.ReorderAirportObjects(sc.AirportObjectOrder)
	om.ReorderDockObjects(sc.DockObjectOrder)
	om.ResolveCliffEdges()

	var m sim.Map
	m.LoadScenario(sc)
	cz := 0
	if ct := m.TileAt(tx, ty); ct != nil {
		cz = int(ct.BaseZ) * 4
	}
	cx, cy := paint.GameToScreen(paint.Pos3{X: tx*32 + 16, Y: ty*32 + 16, Z: cz}, 0)
	r := scene.New(g1, om)
	view := scene.View{X: cx - *w/2, Y: cy - *h/2, W: *w, H: *h}
	if *window > 0 {
		view.Tiles = image.Rect(tx-*window, ty-*window, tx+*window, ty+*window)
	}
	img := r.ToRGBA(r.Render(&m, view))
	f, err := os.Create(*out)
	must(err)
	must(png.Encode(f, img))
	must(f.Close())
	log.Printf("render8: %d sprites decoded → %s", r.SpriteCount(), *out)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
