package world

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"time"

	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/render"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/hajimehoshi/ebiten/v2"
)

// slopeToDisplaySlope maps raw slope values (0-31) to display slope indices (0-18)
// From OpenLoco PaintSurface.cpp kSlopeToDisplaySlope
// Slopes that can't be displayed are mapped to 0 (flat)
var slopeToDisplaySlope = [32]uint8{
	0, 2, 1, 3, 8, 10, 9, 11,
	4, 6, 5, 7, 12, 14, 13, 15,
	0, 0, 0, 0, 0, 0, 0, 17,
	0, 0, 0, 16, 0, 18, 15, 0,
}

// TileType represents different terrain types
type TileType int

const (
	TileGrass TileType = iota
	TileDirt
	TileWater
)

// tile stores per-tile render state
type tile struct {
	tileType     TileType
	terrainIndex uint8 // raw LandObject slot index from the scenario
	baseZ        uint8 // height in SmallZ units (OpenLoco: 1 unit = 4px)
	slope        uint8 // slope corners and flags from byte 4 of surface element
	waterLevel   uint8 // water level from surface element byte 5 bits [4:0]; 0 = no water
	growthStage  uint8 // terrain growth/season stage 0-7 (byte 6 bits [7:5]); selects sprite set
	variation    uint8 // per-tile flat-grass variation (byte 7); 0 = plain flat image
	isIndustrial bool  // surface owned by an industry; terrainIndex is not a land slot
	trees        []scenario.TreeElement
	buildings    []scenario.BuildingElement
	walls        []scenario.WallElement
	tracks       []scenario.TrackElement
	roads        []scenario.RoadElement
	stations     []scenario.StationElement
	signals      []scenario.SignalElement
	industries   []scenario.IndustryElement
}

// DemoTrain is a simple animated train entity for the title sequence and early gameplay.
// It follows the tile grid, moving one tile at a time in direction (dx,dy).
type DemoTrain struct {
	tileX, tileY int     // current tile (leading car position)
	dx, dy       int     // movement direction in tile space (one axis only)
	progress     float64 // 0.0 = at tile centre, 1.0 = at next tile
	vehicleIdx   int     // index into renderer.ObjMgr.Vehicles
}

// World holds a tile grid for an isometric game world
type World struct {
	renderer *render.Renderer
	width    int
	height   int
	tiles    [][]tile
	trains   []*DemoTrain             // demo trains (title screen / scenarios without saved entities)
	entities []scenario.VehicleEntity // real vehicle entities from .SV5 saved-game files
	// isometric tile dimensions (standard 2:1 ratio)
	tileW int // tile width in pixels (64)
	tileH int // tile height in pixels (32)
	// cache colored diamond images per tile type (fallback)
	waterImage *ebiten.Image // solid water base diamond (cached, fallback only)
	// camera offset in pixels (in world-space before zoom is applied)
	camX float64
	camY float64
	// last tile CenterOn snapped to (set by CenterOn, read by caller for diagTileX/Y)
	LastCenterTileX, LastCenterTileY int
	// when true, Draw() does not recentre the camera —
	// an external caller (e.g. title sequence) is driving it via SetCamera.
	externalCamera bool
	// zoom level 0–3 matching OpenLoco convention:
	//   0 = full size (1×), 1 = half (2×), 2 = quarter (4×), 3 = eighth (8×)
	// The pixel scale factor is 1 << zoom applied in reverse:
	// world coordinates are divided by (1 << zoom) for screen placement.
	zoom     int
	animTick int // incremented every frame; drives water wave animation

	// 2-pass blend rendering (water, glass, etc.)
	worldBuf     *ebiten.Image      // pass-1 target: all opaque sprites
	blendBuf     *ebiten.Image      // pass-2 target: blend-mode overlay
	blendOps     []worldBlendOp     // queued during pass 1
	highlightOps []worldHighlightOp // ripple/wave sprites drawn after blend
	waterOps     []worldHighlightOp // water surface ops drawn after all terrain (pass 1.5)

	// RenderIssues tracks tiles where rendering failed silently during the last Draw().
	// Accumulated per-frame; exposed via WorstRenderIssue() for --diag targeting.
	renderIssues []RenderIssue

	// step-based full-map PNG save (one chunk per frame)
	mapSave *mapSaveState
}

// mapSaveState holds in-progress full-map PNG render state.
type mapSaveState struct {
	filename       string
	final          *image.RGBA
	imgW, imgH     int
	minVpX, minVpY float64
	chunkX, chunkY int
	total, done    int
	chunkPix       []byte
	// saved world state restored when saving is complete
	savedCamX, savedCamY float64
	savedZoom            int
	savedExternal        bool
	savedWorldBuf        *ebiten.Image
	savedBlendBuf        *ebiten.Image
}

// RenderIssue records a tile where rendering failed silently during Draw.
// Severity is the number of missing/failed sprites on that tile.
type RenderIssue struct {
	TileX, TileY int
	Severity     int    // higher = worse
	Desc         string // human-readable description of the problem
}

// worldBlendOp records a pending destination-blend draw for pass 2.
type worldBlendOp struct {
	maskImg  *ebiten.Image   // white-mask from Renderer.GetSpriteMask
	tileRect image.Rectangle // screen-space pixel region for the mask
	tint     color.RGBA
	strength float32
}

// worldHighlightOp records a post-blend sprite draw (ripple, wave, etc.).
type worldHighlightOp struct {
	img  *ebiten.Image
	opts *ebiten.DrawImageOptions
}

func NewWorld(r *render.Renderer) *World {
	w := &World{
		renderer: r,
		width:    20,
		height:   15,
		tileW:    64, // standard isometric tile width
		tileH:    32, // standard isometric tile height (64×32 = 2:1 ratio)
		tiles:    make([][]tile, 15),
		zoom:     0, // start at full zoom (1× scale)
	}

	// Initialize tiles with some variety
	for y := 0; y < w.height; y++ {
		w.tiles[y] = make([]tile, w.width)
		for x := 0; x < w.width; x++ {
			tt := TileGrass
			if x < 2 || y < 2 || x >= w.width-2 || y >= w.height-2 {
				tt = TileWater
			} else if (x+y)%7 == 0 {
				tt = TileDirt
			}
			w.tiles[y][x] = tile{tileType: tt}
		}
	}

	// Center camera on the map
	cx, cy := w.tileToScreen(w.width/2, w.height/2)
	w.camX = cx - 400 // half screen width
	w.camY = cy - 300 // half screen height
	return w
}

// tileToScreen converts tile coordinates to viewport pixel coordinates.
// Matches OpenLoco Map/Tile.cpp::gameToScreen (rotation 0):
//
//	vpX = (tileY - tileX) * kTileSize        where kTileSize = 32
//	vpY = (tileY + tileX) * (kTileSize / 2)  = (tileY + tileX) * 16
//
// Height is NOT included here; it is subtracted separately in Draw().
// Camera offset and zoom are also applied in Draw().
func (w *World) tileToScreen(tileX, tileY int) (float64, float64) {
	vpX := float64((tileY - tileX) * 32)
	vpY := float64((tileY + tileX) * 16)
	return vpX, vpY
}

func (w *World) LoadFromScenario(sc *scenario.Scenario) {
	if sc == nil || sc.Tiles == nil {
		return
	}
	w.width = sc.MapWidth
	w.height = sc.MapHeight
	w.tiles = make([][]tile, w.height)
	totalBuildings := 0
	for y := 0; y < w.height; y++ {
		w.tiles[y] = make([]tile, w.width)
		for x := 0; x < w.width; x++ {
			st := sc.GetTile(x, y)
			if st == nil {
				w.tiles[y][x] = tile{tileType: TileGrass}
				continue
			}
			totalBuildings += len(st.Buildings)

			tt := TileGrass
			switch st.Surface {
			case scenario.SurfaceWater:
				tt = TileWater
			case scenario.SurfaceDirt, scenario.SurfaceSand, scenario.SurfaceRock:
				tt = TileDirt
			}
			w.tiles[y][x] = tile{
				tileType:     tt,
				terrainIndex: st.TerrainIndex,
				baseZ:        st.Height,
				slope:        st.Slope,
				waterLevel:   st.Water,
				growthStage:  st.GrowthStage,
				variation:    st.Variation,
				isIndustrial: st.IsIndustrial,
				trees:        st.Trees,
				buildings:    st.Buildings,
				walls:        st.Walls,
				tracks:       st.Tracks,
				roads:        st.Roads,
				stations:     st.Stations,
				signals:      st.Signals,
				industries:   st.Industries,
			}
		}
	}

	log.Printf("[World] LoadFromScenario: %dx%d tiles, totalBuildings=%d", w.width, w.height, totalBuildings)

	// Center camera on the map
	cx, cy := w.tileToScreen(w.width/2, w.height/2)
	w.camX = cx - 400
	w.camY = cy - 300

	// Store real vehicle entities from saved games.  When entities are present
	// they are rendered by paintVehicleEntities; otherwise fall back to demo trains.
	w.entities = sc.Entities

	if len(w.entities) == 0 {
		// No real entities (scenario file or empty save) — spawn scripted demo trains
		// so the title screen and freshly-loaded scenarios show some movement.
		w.spawnDemoTrains()
	} else {
		w.trains = nil
	}
}

// spawnDemoTrains places a handful of trains on track tiles found in the map.
func (w *World) spawnDemoTrains() {
	w.trains = nil
	if w.renderer == nil || w.renderer.ObjMgr == nil || len(w.renderer.ObjMgr.Vehicles) == 0 {
		return
	}

	// Find tiles with at least one track element. Spread spawns evenly.
	type trackPos struct{ x, y int }
	var trackTiles []trackPos
	step := 1
	if w.width*w.height > 1000 {
		step = 8 // sample every 8th tile on large maps
	}
	for y := 0; y < w.height; y += step {
		for x := 0; x < w.width; x += step {
			if len(w.tiles[y][x].tracks) > 0 {
				trackTiles = append(trackTiles, trackPos{x, y})
			}
		}
	}
	if len(trackTiles) == 0 {
		return
	}

	// Spawn up to 5 trains, evenly spaced in the found list.
	maxTrains := 5
	if len(trackTiles) < maxTrains {
		maxTrains = len(trackTiles)
	}
	for i := 0; i < maxTrains; i++ {
		idx := i * len(trackTiles) / maxTrains
		tp := trackTiles[idx]

		// Prefer a direction that has a neighbouring track tile; fall back to +Y.
		dx, dy := 0, 1
		for _, cand := range [][2]int{{0, 1}, {1, 0}, {0, -1}, {-1, 0}} {
			nx, ny := tp.x+cand[0], tp.y+cand[1]
			if nx >= 0 && nx < w.width && ny >= 0 && ny < w.height &&
				len(w.tiles[ny][nx].tracks) > 0 {
				dx, dy = cand[0], cand[1]
				break
			}
		}

		vIdx := i % len(w.renderer.ObjMgr.Vehicles)
		w.trains = append(w.trains, &DemoTrain{
			tileX:      tp.x,
			tileY:      tp.y,
			dx:         dx,
			dy:         dy,
			progress:   float64(i) / float64(maxTrains), // stagger start positions
			vehicleIdx: vIdx,
		})
	}
}

// SetCamera overrides the camera position directly (used by the title
// sequence to pan across the map). Uses the same projection as tileToScreen.
func (w *World) SetCamera(tileX, tileY float64) {
	w.externalCamera = true
	// Mirror of tileToScreen: vpX = (tileY-tileX)*32, vpY = (tileY+tileX)*16
	worldX := (tileY - tileX) * 32.0
	worldY := (tileY + tileX) * 16.0
	halfW := 400.0 * float64(int(1)<<w.zoom)
	halfH := 300.0 * float64(int(1)<<w.zoom)
	w.camX = worldX - halfW
	w.camY = worldY - halfH
}

// ZoomIn decreases the zoom level (more detail, fewer tiles visible).
// Zoom 0 is full size; cannot go below 0.
//
// OpenLoco reference: src/OpenLoco/src/Ui/Window.cpp
//
//	Window::viewportZoomIn(bool toCursor)
func (w *World) ZoomIn() {
	if w.zoom > 0 {
		w.zoom--
	}
}

// ZoomOut increases the zoom level (less detail, more tiles visible).
// Zoom 3 is the minimum size (1/8); cannot go above 3.
//
// OpenLoco reference: src/OpenLoco/src/Ui/Window.cpp
//
//	Window::viewportZoomOut(bool toCursor)
func (w *World) ZoomOut() {
	if w.zoom < 3 {
		w.zoom++
	}
}

// ZoomInAt / ZoomOutAt change zoom while keeping the world point under the
// screen position (sx, sy) fixed — wheel zoom targets the cursor, keyboard
// zoom the viewport centre, like OpenLoco's viewportZoomIn(toCursor).
func (w *World) ZoomInAt(sx, sy int) {
	if w.zoom > 0 {
		w.zoomAt(w.zoom-1, sx, sy)
	}
}

func (w *World) ZoomOutAt(sx, sy int) {
	if w.zoom < 3 {
		w.zoomAt(w.zoom+1, sx, sy)
	}
}

func (w *World) zoomAt(newZoom, sx, sy int) {
	oldScale := 1.0 / float64(int(1)<<w.zoom)
	newScale := 1.0 / float64(int(1)<<newZoom)
	vpX := float64(sx)/oldScale + w.camX
	vpY := float64(sy)/oldScale + w.camY
	w.zoom = newZoom
	w.camX = vpX - float64(sx)/newScale
	w.camY = vpY - float64(sy)/newScale
}

// SetZoom sets the zoom level directly (0=full, 1=half, 2=quarter, 3=eighth)
func (w *World) SetZoom(level int) {
	if level >= 0 && level <= 3 {
		w.zoom = level
	}
}

// GetMapSize returns the map dimensions in tiles
func (w *World) GetMapSize() (width, height int) {
	return w.width, w.height
}

// GetZoom returns the current zoom level (0=full, 1=half, 2=quarter, 3=eighth).
// The screen scale factor is 1/(1<<zoom).
func (w *World) GetZoom() int {
	return w.zoom
}

// CollectSpriteIDs returns all G1 sprite indices that this world will need
// during rendering.  Used by the loading screen to pre-warm the sprite atlas.
//
// Collected categories:
//   - Progress bar / UI sprites (G1 2326-2334)
//   - Cliff-edge mask sprites (G1 3726-3735)
//   - Each LandObject's terrain sprites and cliff-edge sprites
//   - WaterObject sprites (surface shapes + wave frames + palette map)
//   - All other DAT object sprites referenced from tiles
func (w *World) CollectSpriteIDs() []int {
	if w.renderer == nil || w.renderer.ObjMgr == nil {
		return nil
	}
	seen := make(map[int]struct{})
	// Preload bounds are deliberately generous over-estimates per object type,
	// so clamp to the pool: ids past the end are guaranteed decode failures.
	poolSize := -1
	if w.renderer.G1 != nil {
		poolSize = int(w.renderer.G1.GetTotalSprites())
	}
	add := func(id int) {
		if id >= 0 && (poolSize < 0 || id < poolSize) {
			seen[id] = struct{}{}
		}
	}

	// Progress bar train sprites (OpenLoco G1 2326-2334 + track 2330)
	for id := 2326; id <= 2334; id++ {
		add(id)
	}

	// Cliff-edge mask sprites (5 masks × 2 directions: SW/NE use 3726, SE/NW use 3731)
	for id := 3726; id <= 3735; id++ {
		add(id)
	}

	objMgr := w.renderer.ObjMgr

	// Land objects — terrain sprites + cliff-edge sprites
	for _, land := range objMgr.LandObjects {
		if land == nil || land.ImageOffset == 0 {
			continue
		}
		// GetObjectSprite calls GetSprite(int(land.ImageOffset) + localIndex),
		// so atlas keys are absolute G1 indices = ImageOffset + localIndex.
		imgBase := int(land.ImageOffset)
		flatLocalBase := int(land.GetFlatTerrainSpriteIndex())
		count := int(land.NumImagesPerGrowthStage) * int(land.NumGrowthStages)
		if count <= 0 {
			count = 19 * 8 // safe upper bound
		}
		for i := 0; i < count; i++ {
			add(imgBase + flatLocalBase + i)
		}
		// Cliff-edge sprites: CliffEdgeImage is already an absolute G1 index.
		if land.CliffEdgeImage > 0 {
			cliffBase := int(land.CliffEdgeImage)
			// 4 factor offsets (0,16,32,48) × up to 16 height steps + spare
			for f := 0; f < 4; f++ {
				for h := 0; h < 16; h++ {
					add(cliffBase + f*16 + h)
					add(cliffBase + 16 + f*16 + h) // +16 for SE/NW direction
				}
			}
		}
	}

	// Water object sprites
	if wo := objMgr.WaterObj; wo != nil && wo.ImageOffset > 0 {
		base := int(wo.ImageOffset)
		for i := 0; i <= 75; i++ { // 0-29 misc, 30-39 surface shapes, 40-59 misc, 60-75 waves
			add(base + i)
		}
	}

	// Track objects
	for _, tr := range objMgr.TrackObjects {
		if tr == nil || tr.Image == 0 {
			continue
		}
		base := int(tr.Image)
		for i := 0; i < 800; i++ { // generous bound covers all track piece variants
			add(base + i)
		}
	}

	// Road objects
	for _, ro := range objMgr.RoadObjects {
		if ro == nil || ro.Image == 0 {
			continue
		}
		base := int(ro.Image)
		for i := 0; i < 400; i++ {
			add(base + i)
		}
	}

	// Tree objects — scan tiles for used tree object IDs
	usedTrees := make(map[uint8]struct{})
	for _, row := range w.tiles {
		for _, t := range row {
			for _, te := range t.trees {
				usedTrees[te.TreeObjectID] = struct{}{}
			}
		}
	}
	for treeID := range usedTrees {
		treeObj := objMgr.GetTreeObjectByIndex(int(treeID))
		if treeObj == nil || treeObj.ImageOffset == 0 {
			continue
		}
		base := int(treeObj.ImageOffset)
		for i := 0; i < 300; i++ {
			add(base + i)
		}
	}

	// Building objects — scan tiles
	usedBuildings := make(map[uint8]struct{})
	for _, row := range w.tiles {
		for _, t := range row {
			for _, be := range t.buildings {
				usedBuildings[be.ObjectID] = struct{}{}
			}
		}
	}
	for bldID := range usedBuildings {
		bldObj := objMgr.GetBuildingObjectByIndex(int(bldID))
		if bldObj == nil || bldObj.ImageOffset == 0 {
			continue
		}
		base := int(bldObj.ImageOffset)
		for i := 0; i < 600; i++ {
			add(base + i)
		}
	}

	ids := make([]int, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	return ids
}

func (w *World) Update() {
	w.animTick++
	// Advance demo trains at ~2 tiles/second (60fps → 1/30 per frame).
	const speed = 1.0 / 30.0
	for _, tr := range w.trains {
		tr.progress += speed
		if tr.progress >= 1.0 {
			tr.progress -= 1.0
			// Move to the next tile
			nx := tr.tileX + tr.dx
			ny := tr.tileY + tr.dy
			// Bounce if out-of-bounds or no track at destination
			if nx < 0 || nx >= w.width || ny < 0 || ny >= w.height ||
				len(w.tiles[ny][nx].tracks) == 0 {
				tr.dx = -tr.dx
				tr.dy = -tr.dy
			} else {
				tr.tileX = nx
				tr.tileY = ny
			}
		}
	}
}

// mapSaveChunkSize is the pixel size of each rendering tile during full-map save.
// Small enough that the CPU blend pass completes within a single frame.
const mapSaveChunkSize = 512

// StartMapSave begins a step-based full-map PNG render. Call StepMapSave each
// frame; when it returns true, call FinishMapSave to write the PNG.
func (w *World) StartMapSave(filename string) {
	const chunkMax = mapSaveChunkSize
	minVpX := -(float64(w.width) - 1) * 32.0
	maxVpX := (float64(w.height) - 1) * 32.0
	minVpY := -512.0
	maxVpY := (float64(w.width)+float64(w.height)-2)*16.0 + 64.0
	imgW := int(maxVpX-minVpX) + 128
	imgH := int(maxVpY-minVpY) + 64

	total := 0
	for cy := 0; cy < imgH; cy += chunkMax {
		for cx := 0; cx < imgW; cx += chunkMax {
			_ = cx
			total++
		}
	}

	final := image.NewRGBA(image.Rect(0, 0, imgW, imgH))
	for i := 0; i < len(final.Pix); i += 4 {
		final.Pix[i], final.Pix[i+1], final.Pix[i+2], final.Pix[i+3] = 135, 206, 235, 255
	}

	w.mapSave = &mapSaveState{
		filename: filename,
		final:    final,
		imgW:     imgW, imgH: imgH,
		minVpX: minVpX, minVpY: minVpY,
		total:    total,
		chunkPix: make([]byte, chunkMax*chunkMax*4),
		// save world state
		savedCamX: w.camX, savedCamY: w.camY,
		savedZoom:     w.zoom,
		savedExternal: w.externalCamera,
		savedWorldBuf: w.worldBuf,
		savedBlendBuf: w.blendBuf,
	}
	w.zoom = 0
	w.externalCamera = true // prevent Draw() from overriding our chunk camera position
	log.Printf("[World] StartMapSave: %dx%d px, %d chunks", imgW, imgH, total)
}

// StepMapSave renders the next chunk. Returns true when all chunks are done.
// Call FinishMapSave after it returns true.
func (w *World) StepMapSave() bool {
	const chunkMax = mapSaveChunkSize
	s := w.mapSave
	if s == nil {
		return true
	}
	chunkW := min(chunkMax, s.imgW-s.chunkX)
	chunkH := min(chunkMax, s.imgH-s.chunkY)

	w.camX = s.minVpX + float64(s.chunkX)
	w.camY = s.minVpY + float64(s.chunkY)
	w.worldBuf = nil
	w.blendBuf = nil

	log.Printf("[MapSave] chunk %d/%d at (%d,%d) size %dx%d", s.done+1, s.total, s.chunkX, s.chunkY, chunkW, chunkH)
	t0 := time.Now()
	chunk := ebiten.NewImage(chunkW, chunkH)
	chunk.Fill(color.RGBA{135, 206, 235, 255})
	w.Draw(chunk)
	log.Printf("[MapSave] chunk %d/%d rendered in %v", s.done+1, s.total, time.Since(t0))

	pix := s.chunkPix[:chunkW*chunkH*4]
	chunk.ReadPixels(pix)
	for row := 0; row < chunkH; row++ {
		srcOff := row * chunkW * 4
		dstOff := (s.chunkY+row)*s.imgW*4 + s.chunkX*4
		copy(s.final.Pix[dstOff:dstOff+chunkW*4], pix[srcOff:srcOff+chunkW*4])
	}
	s.done++

	// Advance to next chunk.
	s.chunkX += chunkMax
	if s.chunkX >= s.imgW {
		s.chunkX = 0
		s.chunkY += chunkMax
	}
	if s.chunkY >= s.imgH {
		return true
	}
	// DEBUG: bail after first chunk to verify output — remove when confirmed correct.
	return s.done >= 1
}

// MapSaveProgress returns (done, total) chunks for the in-progress save.
// total is 0 if no save is in progress.
func (w *World) MapSaveProgress() (done, total int) {
	if w.mapSave == nil {
		return 0, 0
	}
	return w.mapSave.done, w.mapSave.total
}

// FinishMapSave writes the PNG, restores world state, and clears the save state.
func (w *World) FinishMapSave() error {
	s := w.mapSave
	if s == nil {
		return nil
	}
	// Restore world state.
	w.camX, w.camY = s.savedCamX, s.savedCamY
	w.zoom = s.savedZoom
	w.externalCamera = s.savedExternal
	w.worldBuf = s.savedWorldBuf
	w.blendBuf = s.savedBlendBuf
	w.mapSave = nil

	f, err := os.Create(s.filename)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := png.Encode(f, s.final); err != nil {
		return err
	}
	log.Printf("[World] Full map saved to %s (%dx%d)", s.filename, s.imgW, s.imgH)
	return nil
}

// CenterOn moves the camera to the nearest tile (from map centre outward) that
// matches kind. Recognised kinds:
//
//	"water"    — tile with a water surface
//	"tree"     — tile with at least one tree element
//	"building" — tile with at least one building element
//	"track"    — tile with at least one track element
//	"road"     — tile with at least one road element
//	"station"  — tile with at least one station element
//	"industry" — tile with at least one industry element
//	"wall"     — tile with at least one wall element
//	"train"    — first demo-train or vehicle entity position
//	""         — no-op (leaves camera where it is)
//
// Returns true if a matching position was found.
func (w *World) CenterOn(kind string) bool {
	if kind == "" {
		return false
	}

	// Trains live outside the tile grid; handle separately.
	if kind == "train" || kind == "trains" || kind == "vehicle" || kind == "vehicles" {
		// Find first placed Body entity with valid tile coordinates.
		// Skip entities at tile (0,0) — those are unplaced/invalid slots.
		for i := range w.entities {
			e := &w.entities[i]
			if e.EntityType != scenario.VehicleTypeBody {
				continue
			}
			tx, ty := e.PosX.Tile(), e.PosY.Tile()
			if tx <= 0 || ty <= 0 || tx.Int() >= w.width || ty.Int() >= w.height {
				continue
			}
			log.Printf("[World] CenterOn(%q): entity at (%d,%d)", kind, tx, ty)
			w.LastCenterTileX, w.LastCenterTileY = tx.Int(), ty.Int()
			w.SetCamera(float64(tx), float64(ty))
			w.externalCamera = false
			return true
		}
		if len(w.trains) > 0 {
			tr := w.trains[0]
			log.Printf("[World] CenterOn(%q): demo train at (%d,%d)", kind, tr.tileX, tr.tileY)
			w.LastCenterTileX, w.LastCenterTileY = tr.tileX, tr.tileY
			w.SetCamera(float64(tr.tileX), float64(tr.tileY))
			w.externalCamera = false
			return true
		}
		log.Printf("[World] CenterOn(%q): no trains found", kind)
		return false
	}

	// "worst" scans all tiles, scores each by expected rendering problems,
	// and jumps to the tile with the highest score.
	if kind == "worst" {
		bestScore := -1
		bestX, bestY := w.width/2, w.height/2
		for y := 0; y < w.height; y++ {
			for x := 0; x < w.width; x++ {
				t := &w.tiles[y][x]
				score := 0
				// Water tile: known rendering issues
				if t.waterLevel > 0 && uint16(t.waterLevel)*16 > uint16(t.baseZ)*4 {
					score += 5
					if w.renderer == nil || w.renderer.ObjMgr == nil || w.renderer.ObjMgr.WaterObj == nil {
						score += 20
					}
				}
				// Buildings with missing objects
				for _, be := range t.buildings {
					if w.renderer == nil || w.renderer.ObjMgr == nil ||
						w.renderer.ObjMgr.GetBuildingObjectByIndex(int(be.ObjectID)) == nil {
						score += 10
					}
				}
				// Terrain with missing land object
				if w.renderer != nil && w.renderer.ObjMgr != nil &&
					w.renderer.ObjMgr.GetLandObjectByIndex(int(t.terrainIndex)) == nil {
					score += 8
				}
				if score > bestScore {
					bestScore = score
					bestX, bestY = x, y
				}
			}
		}
		if bestScore <= 0 {
			log.Printf("[World] CenterOn(%q): no render issues detected in tile data", kind)
			return false
		}
		log.Printf("[World] CenterOn(%q): worst tile at (%d,%d) score=%d", kind, bestX, bestY, bestScore)
		w.LastCenterTileX, w.LastCenterTileY = bestX, bestY
		w.SetCamera(float64(bestX), float64(bestY))
		w.externalCamera = false
		return true
	}

	match := func(t *tile) bool {
		switch kind {
		case "water":
			return t.waterLevel > 0 && uint16(t.waterLevel)*16 > uint16(t.baseZ)*4
		case "tree", "trees":
			return len(t.trees) > 0
		case "building", "buildings":
			return len(t.buildings) > 0
		case "track", "tracks":
			return len(t.tracks) > 0
		case "road", "roads":
			return len(t.roads) > 0
		case "station", "stations":
			return len(t.stations) > 0
		case "industry", "industries":
			return len(t.industries) > 0
		case "wall", "walls":
			return len(t.walls) > 0
		}
		return false
	}

	cx, cy := w.width/2, w.height/2
	for r := 0; r < w.width/2+w.height/2; r++ {
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				if abs(dx) != r && abs(dy) != r {
					continue
				}
				x, y := cx+dx, cy+dy
				if x < 0 || y < 0 || x >= w.width || y >= w.height {
					continue
				}
				if match(&w.tiles[y][x]) {
					log.Printf("[World] CenterOn(%q): found at (%d,%d)", kind, x, y)
					w.LastCenterTileX, w.LastCenterTileY = x, y
					w.SetCamera(float64(x), float64(y))
					w.externalCamera = false
					return true
				}
			}
		}
	}
	log.Printf("[World] CenterOn(%q): nothing found", kind)
	return false
}

// CenterOnWater is a convenience wrapper around CenterOn("water").
func (w *World) CenterOnWater() bool {
	return w.CenterOn("water")
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// PanCamera moves the camera by (dx, dy) pixels in world-space.
// Called each frame when WASD/arrow keys are held.
func (w *World) PanCamera(dx, dy float64) {
	w.externalCamera = false // user is now driving the camera
	w.camX += dx
	w.camY += dy
}

// getTile returns a pointer to the tile at (x, y), or nil if out of bounds
func (w *World) getTile(x, y int) *tile {
	if x < 0 || x >= w.width || y < 0 || y >= w.height {
		return nil
	}
	return &w.tiles[y][x]
}

// TileToScreenPx returns the screen pixel (sx, sy) for the centre of tile (tileX, tileY)
// given the current camera and zoom. Inverse of ScreenToTile.
func (w *World) TileToScreenPx(tileX, tileY int) (int, int) {
	scale := 1.0 / float64(int(1)<<w.zoom)
	vpX, vpY := w.tileToScreen(tileX, tileY)
	if t := w.getTile(tileX, tileY); t != nil {
		vpY -= float64(t.baseZ) * 4.0 // same height shift Draw() applies
	}
	sx := int((vpX - w.camX) * scale)
	sy := int((vpY - w.camY) * scale)
	return sx, sy
}

// ScreenToTile converts a screen pixel (sx, sy) to tile coordinates.
// Returns (-1, -1) if the position maps outside the map bounds.
// Terrain height is compensated for by iterative refinement (see below).
//
// Inverse of tileToScreen:
//
//	vpX = (tileY - tileX) * 32  →  tileY - tileX = vpX/32
//	vpY = (tileY + tileX) * 16  →  tileY + tileX = vpY/16
//	tileY = (vpX/32 + vpY/16) / 2
//	tileX = (vpY/16 - vpX/32) / 2
func (w *World) ScreenToTile(sx, sy int) (int, int) {
	scale := 1.0 / float64(int(1)<<w.zoom)
	// un-zoom and un-camera to get world-space viewport coords
	vpX := float64(sx)/scale + w.camX
	vpY := float64(sy)/scale + w.camY
	// Draw() shifts each tile up by baseZ*4 px, so picking must shift the
	// cursor down by the height of the tile actually under it. That tile is
	// only known once picked — iterate a few times to converge on hills.
	tileX, tileY := -1, -1
	adjY := vpY
	for i := 0; i < 4; i++ {
		tx := int(math.Floor((adjY/16.0 - vpX/32.0) / 2.0))
		ty := int(math.Floor((vpX/32.0 + adjY/16.0) / 2.0))
		if tx == tileX && ty == tileY {
			break
		}
		tileX, tileY = tx, ty
		t := w.getTile(tx, ty)
		if t == nil {
			break
		}
		adjY = vpY + float64(t.baseZ)*4.0
	}
	if tileX < 0 || tileX >= w.width || tileY < 0 || tileY >= w.height {
		return -1, -1
	}
	return tileX, tileY
}

// ── Surface edge smoothing ────────────────────────────────────────────────────
// OpenLoco reference: src/OpenLoco/src/Paint/PaintSurface.cpp
// paintSurfaceSmoothenEdge — where two tiles meet coplanar along an edge, the
// neighbour's blend image (locals 19-24 of its growth block) is drawn through
// a per-edge/slope mask sprite so terrain types fade into each other.

// surfaceSmoothMaskBase[edge] = G1 index of surfaceSmooth{edge}Slope0.
var surfaceSmoothMaskBase = [4]int{3784, 3765, 3803, 3746}

// Per-corner tint tables indexed by display slope (k4FDA5E/71/84/97).
var (
	tintBottomLeft  = [19]uint8{2, 5, 1, 4, 2, 5, 1, 2, 2, 4, 1, 2, 1, 3, 0, 3, 1, 5, 0}
	tintTopLeft     = [19]uint8{2, 5, 2, 4, 2, 5, 1, 1, 3, 4, 3, 2, 1, 2, 0, 3, 1, 5, 0}
	tintTopRight    = [19]uint8{2, 2, 2, 4, 0, 0, 1, 1, 3, 4, 3, 5, 1, 2, 2, 3, 1, 5, 0}
	tintBottomRight = [19]uint8{2, 2, 1, 4, 0, 0, 1, 2, 2, 4, 1, 5, 1, 3, 2, 3, 1, 5, 0}
)

// cornerHeightsTable[slope] = corner raises {top, right, bottom, left} in MicroZ.
var cornerHeightsTable = [32][4]uint8{
	{0, 0, 0, 0}, {0, 0, 1, 0}, {0, 0, 0, 1}, {0, 0, 1, 1},
	{1, 0, 0, 0}, {1, 0, 1, 0}, {1, 0, 0, 1}, {1, 0, 1, 1},
	{0, 1, 0, 0}, {0, 1, 1, 0}, {0, 1, 0, 1}, {0, 1, 1, 1},
	{1, 1, 0, 0}, {1, 1, 1, 0}, {1, 1, 0, 1}, {1, 1, 1, 1},
	{0, 0, 0, 0}, {0, 0, 1, 0}, {0, 0, 0, 1}, {0, 0, 1, 1},
	{1, 0, 0, 0}, {1, 0, 1, 0}, {1, 0, 0, 1}, {1, 0, 1, 2},
	{0, 1, 0, 0}, {0, 1, 1, 0}, {0, 1, 0, 1}, {0, 1, 2, 1},
	{1, 1, 0, 0}, {1, 2, 1, 0}, {2, 1, 0, 1}, {1, 1, 1, 1},
}

// smoothNeighbourOffsets[edge] = tile delta at rotation 0: SW, SE, NW, NE.
var smoothNeighbourOffsets = [4][2]int{{1, 0}, {0, 1}, {0, -1}, {-1, 0}}

func (w *World) paintSurfaceSmoothing(dst *ebiten.Image, x, y int, t *tile, land *objects.LandObject, drawX, drawY, scale float64) {
	if os.Getenv("GOLOCO_NO_SMOOTH") == "1" {
		return
	}
	if t.isIndustrial {
		return // industry fields have no land blend images
	}
	selfSnow := t.slope >> 5
	if selfSnow >= 3 { // snow blending not supported yet
		return
	}
	selfSlope := int(t.slope & 0x1F)
	dsSelf := int(slopeToDisplaySlope[selfSlope])
	selfMicro := int(t.baseZ) / 4
	sc := cornerHeightsTable[selfSlope]

	for edge := 0; edge < 4; edge++ {
		nb := w.getTile(x+smoothNeighbourOffsets[edge][0], y+smoothNeighbourOffsets[edge][1])
		if nb == nil || nb.isIndustrial || nb.slope>>5 >= 4 {
			continue
		}
		nbSlope := int(nb.slope & 0x1F)
		nbMicro := int(nb.baseZ) / 4
		nc := cornerHeightsTable[nbSlope]

		// Only smooth when the shared edge is coplanar (matching corner heights).
		var s0, n0, s1, n1 int
		switch edge {
		case 0: // SW: self L/B vs neighbour T/R
			s0, n0 = selfMicro+int(sc[3]), nbMicro+int(nc[0])
			s1, n1 = selfMicro+int(sc[2]), nbMicro+int(nc[1])
		case 1: // SE: self R/B vs neighbour T/L
			s0, n0 = selfMicro+int(sc[1]), nbMicro+int(nc[0])
			s1, n1 = selfMicro+int(sc[2]), nbMicro+int(nc[3])
		case 2: // NW: self T/L vs neighbour R/B
			s0, n0 = selfMicro+int(sc[0]), nbMicro+int(nc[1])
			s1, n1 = selfMicro+int(sc[3]), nbMicro+int(nc[2])
		case 3: // NE: self T/R vs neighbour L/B
			s0, n0 = selfMicro+int(sc[0]), nbMicro+int(nc[3])
			s1, n1 = selfMicro+int(sc[1]), nbMicro+int(nc[2])
		}
		if s0 != n0 || s1 != n1 {
			continue
		}

		dsNb := int(slopeToDisplaySlope[nbSlope])
		var dh, cl uint8
		switch edge {
		case 0:
			dh, cl = tintBottomLeft[dsSelf], tintTopRight[dsNb]
		case 1:
			dh, cl = tintBottomRight[dsSelf], tintTopLeft[dsNb]
		case 2:
			dh, cl = tintTopLeft[dsSelf], tintBottomRight[dsNb]
		case 3:
			dh, cl = tintTopRight[dsSelf], tintBottomLeft[dsNb]
		}

		nbLand := w.renderer.ObjMgr.GetLandObjectByIndex(int(nb.terrainIndex))
		if nbLand == nil || nbLand.ImageOffset == 0 {
			continue
		}

		if t.terrainIndex == nb.terrainIndex && t.growthStage == nb.growthStage {
			if cl == dh {
				continue // same tint — nothing to blend
			}
			if land.HasFlags(objects.LandFlagHasSharpSlopeTransition) {
				continue
			}
		} else {
			if land.HasFlags(objects.LandFlagDisableSmoothTileTransition) ||
				nbLand.HasFlags(objects.LandFlagDisableSmoothTileTransition) {
				continue
			}
		}

		maskIdx := surfaceSmoothMaskBase[edge] + dsSelf
		colourLocal := nbLand.GetFlatTerrainSpriteIndex() +
			int(nb.growthStage)*int(nbLand.NumImagesPerGrowthStage) + 19 + int(cl)
		img := w.renderer.GetMaskedSpriteAligned(int(nbLand.ImageOffset)+colourLocal, maskIdx)
		if img == nil {
			continue
		}
		_, _, mXO, mYO, ok := w.renderer.GetSpriteInfo(maskIdx)
		if !ok {
			continue
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(drawX+float64(mXO)*scale, drawY+float64(mYO)*scale)
		dst.DrawImage(img, op)
	}
}

// GetTileBaseZ returns the baseZ of tile (x, y), or -1 if out of bounds.
func (w *World) GetTileBaseZ(x, y int) int {
	t := w.getTile(x, y)
	if t == nil {
		return -1
	}
	return int(t.baseZ)
}

// AddTrack appends a track element to tile (x, y).
// Returns false if the tile is out of bounds or an identical element already exists.
func (w *World) AddTrack(x, y int, te scenario.TrackElement) bool {
	t := w.getTile(x, y)
	if t == nil {
		return false
	}
	for _, ex := range t.tracks {
		if ex.TrackObjectID == te.TrackObjectID &&
			ex.TrackID == te.TrackID &&
			ex.Rotation == te.Rotation &&
			ex.BaseZ == te.BaseZ {
			return false
		}
	}
	t.tracks = append(t.tracks, te)
	return true
}

// AddRoad appends a road element to tile (x, y).
// Returns false if the tile is out of bounds or an identical element already exists.
func (w *World) AddRoad(x, y int, re scenario.RoadElement) bool {
	t := w.getTile(x, y)
	if t == nil {
		return false
	}
	for _, ex := range t.roads {
		if ex.RoadObjectID == re.RoadObjectID &&
			ex.RoadID == re.RoadID &&
			ex.Rotation == re.Rotation &&
			ex.BaseZ == re.BaseZ {
			return false
		}
	}
	t.roads = append(t.roads, re)
	return true
}

// RemoveLastTrack removes the most recently placed track element from tile (x, y).
func (w *World) RemoveLastTrack(x, y int) bool {
	t := w.getTile(x, y)
	if t == nil || len(t.tracks) == 0 {
		return false
	}
	t.tracks = t.tracks[:len(t.tracks)-1]
	return true
}

// RemoveLastRoad removes the most recently placed road element from tile (x, y).
func (w *World) RemoveLastRoad(x, y int) bool {
	t := w.getTile(x, y)
	if t == nil || len(t.roads) == 0 {
		return false
	}
	t.roads = t.roads[:len(t.roads)-1]
	return true
}

// Edge directions for neighbor lookup
const (
	EdgeSW = 0 // South-West (bottom-left)
	EdgeSE = 1 // South-East (bottom-right)
	EdgeNW = 2 // North-West (top-left)
	EdgeNE = 3 // North-East (top-right)
)

// getNeighborTile returns the neighbor tile in the given edge direction.
// Offsets match OpenLoco PaintSurface.cpp::kNeighbourOffsets[rotation=0]:
//
//	SW: Pos2{+32,0}  → tileX+1
//	SE: Pos2{0,+32}  → tileY+1
//	NW: Pos2{0,-32}  → tileY-1
//	NE: Pos2{-32,0}  → tileX-1
func (w *World) getNeighborTile(x, y int, edge int) *tile {
	var dx, dy int
	switch edge {
	case EdgeSW: // world +x
		dx, dy = 1, 0
	case EdgeSE: // world +y
		dx, dy = 0, 1
	case EdgeNW: // world -y
		dx, dy = 0, -1
	case EdgeNE: // world -x
		dx, dy = -1, 0
	default:
		return nil
	}
	return w.getTile(x+dx, y+dy)
}

// CornerHeight represents the height of the 4 corners of a tile
// From OpenLoco CornerHeight struct
type CornerHeight struct {
	Top    uint8 // North corner
	Right  uint8 // East corner
	Bottom uint8 // South corner
	Left   uint8 // West corner
}

// cornerHeights lookup table from OpenLoco kCornerHeights
// Maps slope value (0-31) to corner heights (0, 1, or 2)
var cornerHeights = [32]CornerHeight{
	{0, 0, 0, 0}, {0, 0, 1, 0}, {0, 0, 0, 1}, {0, 0, 1, 1},
	{1, 0, 0, 0}, {1, 0, 1, 0}, {1, 0, 0, 1}, {1, 0, 1, 1},
	{0, 1, 0, 0}, {0, 1, 1, 0}, {0, 1, 0, 1}, {0, 1, 1, 1},
	{1, 1, 0, 0}, {1, 1, 1, 0}, {1, 1, 0, 1}, {1, 1, 1, 1},
	{0, 0, 0, 0}, {0, 0, 1, 0}, {0, 0, 0, 1}, {0, 0, 1, 1},
	{1, 0, 0, 0}, {1, 0, 1, 0}, {1, 0, 0, 1}, {1, 0, 1, 2},
	{0, 1, 0, 0}, {0, 1, 1, 0}, {0, 1, 0, 1}, {0, 1, 2, 1},
	{1, 1, 0, 0}, {1, 2, 1, 0}, {2, 1, 0, 1}, {1, 1, 1, 1},
}

// getCornerHeights returns the absolute corner heights for a tile in MicroZ units.
// OpenLoco PaintSurface.cpp: microZ = baseZ / kMicroToSmallZStep (= baseZ / 4).
// 1 MicroZ = 16 viewport pixels at zoom-0.
func (w *World) getCornerHeights(t *tile) CornerHeight {
	if t == nil {
		return CornerHeight{0, 0, 0, 0}
	}
	// SmallZ → MicroZ: divide by kMicroToSmallZStep = 4
	microZ := uint8(t.baseZ) / 4
	slope := t.slope & 0x1F
	rel := cornerHeights[slope]
	return CornerHeight{
		Top:    microZ + rel.Top,
		Right:  microZ + rel.Right,
		Bottom: microZ + rel.Bottom,
		Left:   microZ + rel.Left,
	}
}

// EdgeHeight represents the heights of the two corners on an edge
type EdgeHeight struct {
	Self0     uint8 // First corner on self side
	Neighbor0 uint8 // First corner on neighbor side
	Self1     uint8 // Second corner on self side
	Neighbor1 uint8 // Second corner on neighbor side
}

// getEdgeHeights calculates the corner heights for a specific edge
func (w *World) getEdgeHeights(x, y int, edge int, selfCorners CornerHeight) EdgeHeight {
	neighbor := w.getNeighborTile(x, y, edge)
	if neighbor == nil {
		return EdgeHeight{0, 0, 0, 0}
	}

	// If the neighbor is submerged, its terrain is drawn at waterLevel height.
	// Use waterLevel as effective corner height so cliff edges terminate at the
	// water surface rather than at the (lower) terrain base.
	var neighborCorners CornerHeight
	if neighbor.waterLevel > 0 && uint16(neighbor.waterLevel)*16 > uint16(neighbor.baseZ)*4 {
		wl := uint8(neighbor.waterLevel)
		neighborCorners = CornerHeight{wl, wl, wl, wl}
	} else {
		neighborCorners = w.getCornerHeights(neighbor)
	}

	// Map edge direction to corner pairs
	// From OpenLoco paintSurface edge height calculation
	var eh EdgeHeight
	switch edge {
	case EdgeSW: // 0
		eh.Self0 = selfCorners.Left
		eh.Neighbor0 = neighborCorners.Top
		eh.Self1 = selfCorners.Bottom
		eh.Neighbor1 = neighborCorners.Right
	case EdgeSE: // 1
		eh.Self0 = selfCorners.Right
		eh.Neighbor0 = neighborCorners.Top
		eh.Self1 = selfCorners.Bottom
		eh.Neighbor1 = neighborCorners.Left
	case EdgeNW: // 2
		eh.Self0 = selfCorners.Top
		eh.Neighbor0 = neighborCorners.Right
		eh.Self1 = selfCorners.Left
		eh.Neighbor1 = neighborCorners.Bottom
	case EdgeNE: // 3
		eh.Self0 = selfCorners.Top
		eh.Neighbor0 = neighborCorners.Left
		eh.Self1 = selfCorners.Right
		eh.Neighbor1 = neighborCorners.Bottom
	}

	return eh
}

// Edge factor offsets for cliff sprite variation per edge direction.
// OpenLoco reference: PaintSurface.cpp kEdgeFactorOffset
var kEdgeFactorOffset = [4]uint32{0, 16, 16, 0} // SW, SE, NW, NE

// kCliffEdgeMaskBase[edge] is the G1 index of the FIRST of 5 consecutive mask sprites
// for that edge direction.
//
// OpenLoco reference: PaintSurface.cpp kEdgeMaskImageFromSlope, ImageIds.h 3726-3735
//
// G1 indices:
//
//	SW / NE edges: cliffEdge0MaskSlope0..4 = G1 3726–3730
//	SE / NW edges: cliffEdge1MaskSlope0..4 = G1 3731–3735
//
// Mask slot semantics (paintSurfaceCliffEdgeImpl, lines 1230–1319):
//
//	[0] body  — rectangular, used for all interior height steps
//	[1] top-A — topmost step when uHeight < self0 (one corner still higher)
//	[2] top-B — topmost step when uHeight >= self0 (other corner is the remaining top)
//	[3] bot-A — bottom slope when neigh1 < neigh0 (neigh1 is the lower start)
//	[4] bot-B — bottom slope when neigh1 >= neigh0 (snap uHeight down to neigh0)
//
// CRITICAL: when self0 == self1 (flat-topped cliff OR water cliff edge), the top
// section is NEVER drawn — the algorithm exits early. Using mask[1] or [2] for
// the last body iteration in that case produces triangular "chevron" artifacts.
var kCliffEdgeMaskBase = [4]int{3726, 3731, 3731, 3726} // indexed by EdgeSW/SE/NW/NE

// paintSurfaceCliffEdgeImpl is the Go port of OpenLoco's paintSurfaceCliffEdgeImpl.
//
// OpenLoco reference: PaintSurface.cpp paintSurfaceCliffEdgeImpl (lines 1211–1319)
//
// Parameters (all heights in MicroZ units; 1 MicroZ = 16 screen px at zoom-0):
//
//	self0, self1   — the two corner heights of THIS tile's edge (from getEdgeHeights)
//	neigh0, neigh1 — the corresponding corner heights of the NEIGHBOR tile's edge
//	drawYFlat      — screen Y at this tile with baseZ contribution removed
//	                 (= drawY + baseZ*4*scale)
//	emit           — callback to place the finished DrawImageOptions; callers use
//	                 this to write to screen (terrain) or to highlightOps (water).
//
// Algorithm in three steps:
//
//	Step 1 — Bottom slope section (only when neigh0 != neigh1):
//	  Start uHeight = neigh1.
//	  If neigh1 >= neigh0: maskSlot=4, snap uHeight=neigh0 (lowest neighbor corner).
//	  Else:                maskSlot=3.
//	  Draw that single section (unless uHeight is already at self0 or self1), uHeight++.
//
//	Step 2 — Body sections:
//	  while uHeight < self0 && uHeight < self1: draw maskSlot=0 (rectangular), uHeight++.
//
//	Step 3 — Top section (one final sloped section where the two self-corners diverge):
//	  maskSlot = 1
//	  if uHeight >= self0:  maskSlot = 2
//	  if uHeight >= self1:  return  ← cliff complete, no top section needed
//	  draw maskSlot
//
// When self0 == self1 (flat top OR water cliff), step 3 always hits the early return
// because after step 2 uHeight == self0 == self1. No top-slope section is drawn.
func (w *World) paintSurfaceCliffEdgeImpl(
	cliffEdgeImageBase uint32,
	factor uint32, maskBase int, edge int,
	self0, self1, neigh0, neigh1 uint8,
	drawX, drawYFlat float64, scale float64,
	emit func(*ebiten.Image, *ebiten.DrawImageOptions),
) {
	// Guard: this tile's edge is not higher than the neighbor's — nothing to draw.
	if self0 <= neigh0 && self1 <= neigh1 {
		return
	}

	drawSection := func(uHeight uint8, maskSlot int) {
		spriteID := int(cliffEdgeImageBase) + int(factor) + int(uHeight&0xF)
		img := w.renderer.GetMaskedSprite(spriteID, maskBase+maskSlot)
		if img == nil {
			return
		}
		_, _, xOff, yOff, ok := w.renderer.GetSpriteInfo(spriteID)
		if !ok {
			return
		}
		edgeOffsetX, edgeOffsetY := w.getCliffEdgeOffset(edge, uHeight)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(
			drawX+float64(xOff+edgeOffsetX)*scale,
			drawYFlat+float64(yOff+edgeOffsetY)*scale)
		emit(img, op)
	}

	// Step 1: bottom slope — only when the two neighbor corners differ in height.
	uHeight := neigh1
	if uHeight != neigh0 {
		maskSlot := 3 // neigh1 < neigh0: start at neigh1 (lower), use bot-A mask
		if uHeight >= neigh0 {
			maskSlot = 4 // neigh1 > neigh0: snap down to neigh0, use bot-B mask
			uHeight = neigh0
		}
		// Only draw this section if we are not already sitting on one of the self corners.
		if uHeight != self0 && uHeight != self1 {
			drawSection(uHeight, maskSlot)
			uHeight++
		}
	}

	// Step 2: body sections — rectangular mask[0] for every interior step.
	for uHeight < self0 && uHeight < self1 {
		drawSection(uHeight, 0)
		uHeight++
	}

	// Step 3: top section — one sloped section at the boundary between the two
	// self-corner heights.
	//
	// When self0 == self1 (flat-topped cliff or water), the loop above exits with
	// uHeight == self0 == self1, so both conditions below are true → return.
	// NO top-slope section is drawn in that case. (Previous bug: incorrectly
	// drew mask[1] for the last body iteration, creating triangular chevrons.)
	maskSlot := 1
	if uHeight >= self0 {
		maskSlot = 2
		if uHeight >= self1 {
			return // cliff complete — both self-corner heights have been passed
		}
	}
	drawSection(uHeight, maskSlot)
}

// paintCliffEdges renders cliff-edge sprites for terrain height transitions.
//
// OpenLoco reference: PaintSurface.cpp paintSurfaceCliffEdge() (line 1322)
//
//	which calls paintSurfaceCliffEdgeImpl() for the actual section loop.
func (w *World) paintCliffEdges(screen *ebiten.Image, x, y int, t *tile, land *objects.LandObject, drawX, drawY, scale float64) {
	if w.renderer == nil || w.renderer.G1 == nil || land == nil {
		return
	}

	// drawY already has baseZ*4 subtracted (the tile's height offset).
	// getCliffEdgeOffset returns offsets relative to the FLAT (zero-height) tile
	// position, so restore drawYFlat = drawY + baseZ*4*scale.
	drawYFlat := drawY + float64(t.baseZ)*4.0*scale

	selfCorners := w.getCornerHeights(t)

	// Checkerboard factor: OpenLoco uses ((worldX ^ worldY) & 0x20) with worldX=tileX*32.
	// Equivalent in tile coords: ((tileX ^ tileY) & 1) * 32 — alternates every tile.
	checkerboard := uint32((x^y)&1) * 32

	for _, edge := range []int{EdgeNW, EdgeNE, EdgeSW, EdgeSE} {
		if w.getNeighborTile(x, y, edge) == nil {
			continue
		}
		eh := w.getEdgeHeights(x, y, edge, selfCorners)
		factor := checkerboard + kEdgeFactorOffset[edge]
		w.paintSurfaceCliffEdgeImpl(
			land.CliffEdgeImage,
			factor, kCliffEdgeMaskBase[edge], edge,
			eh.Self0, eh.Self1, eh.Neighbor0, eh.Neighbor1,
			drawX, drawYFlat, scale,
			func(img *ebiten.Image, op *ebiten.DrawImageOptions) {
				screen.DrawImage(img, op)
			},
		)
	}
}

// Tree quadrant offsets within a tile (in sub-tile coordinates).
// OpenLoco reference: PaintTree.cpp kTreeQuadrantOffset
var kTreeQuadrantOffset = [4][2]int16{
	{7, 7},   // quadrant 0
	{7, 23},  // quadrant 1
	{23, 23}, // quadrant 2
	{23, 7},  // quadrant 3
}

// paintWaterCliffEdges queues water-shoreline cliff-edge sprites into highlightOps
// so they composite after terrain (pass 1) and water (pass 1.5) are both drawn.
//
// OpenLoco reference: PaintSurface.cpp paintSurfaceWaterCliffEdge() (line 1332)
//
//	which calls paintSurfaceCliffEdgeImpl with self0 = self1 = waterHeight/kMicroZStep.
//
// Because self0 == self1 == waterMicroZ always, paintSurfaceCliffEdgeImpl's step 3
// (top-slope section) is ALWAYS skipped.  Every section uses mask[0] (rectangular).
// The previous code drew mask[1] for the topmost iteration, producing chevrons.
//
// Neighbor corners used per edge (OpenLoco TileDescriptor assembly, line 1408+):
//
//	EdgeSW(0): neigh0=neighbor.Top,   neigh1=neighbor.Right
//	EdgeSE(1): neigh0=neighbor.Top,   neigh1=neighbor.Left
//	EdgeNW(2): neigh0=neighbor.Right, neigh1=neighbor.Bottom
//	EdgeNE(3): neigh0=neighbor.Left,  neigh1=neighbor.Bottom
func (w *World) paintWaterCliffEdges(x, y int, t *tile, land *objects.LandObject, drawX, drawY, scale float64) {
	if w.renderer == nil || w.renderer.G1 == nil || land == nil || land.CliffEdgeImage == 0 {
		return
	}
	if t.waterLevel == 0 {
		return
	}

	// drawYFlat restores the flat-tile Y reference (drawY has baseZ*4 subtracted).
	drawYFlat := drawY + float64(t.baseZ)*4.0*scale

	// waterMicroZ: water surface height in MicroZ (1 MicroZ = 16 screen px at zoom-0).
	// Stored as-is in tile.waterLevel; OpenLoco: waterHeight / kMicroZStep.
	waterMicroZ := t.waterLevel

	checkerboard := uint32((x^y)&1) * 32

	for _, edge := range []int{EdgeNW, EdgeNE, EdgeSW, EdgeSE} {
		neighbor := w.getNeighborTile(x, y, edge)
		if neighbor == nil || neighbor.waterLevel == t.waterLevel {
			continue
		}

		// Neighbor terrain corner heights in MicroZ (baseZ/4 + corner offset).
		nc := w.getCornerHeights(neighbor)
		var neigh0, neigh1 uint8
		switch edge {
		case EdgeSW:
			neigh0, neigh1 = nc.Top, nc.Right
		case EdgeSE:
			neigh0, neigh1 = nc.Top, nc.Left
		case EdgeNW:
			neigh0, neigh1 = nc.Right, nc.Bottom
		case EdgeNE:
			neigh0, neigh1 = nc.Left, nc.Bottom
		}

		factor := checkerboard + kEdgeFactorOffset[edge]
		// self0 == self1 == waterMicroZ: paintSurfaceCliffEdgeImpl will use mask[0]
		// for every section and skip the top-slope section entirely (step 3 returns).
		w.paintSurfaceCliffEdgeImpl(
			land.CliffEdgeImage,
			factor, kCliffEdgeMaskBase[edge], edge,
			waterMicroZ, waterMicroZ, neigh0, neigh1,
			drawX, drawYFlat, scale,
			func(img *ebiten.Image, op *ebiten.DrawImageOptions) {
				w.highlightOps = append(w.highlightOps, worldHighlightOp{img: img, opts: op})
			},
		)
	}
}

// paintTrees renders tree sprites for a tile.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintTree.cpp  paintTree()
func (w *World) paintTrees(screen *ebiten.Image, t *tile, drawX, drawY, scale float64) {
	if w.renderer == nil || w.renderer.ObjMgr == nil || len(t.trees) == 0 {
		return
	}

	for _, te := range t.trees {
		treeObj := w.renderer.ObjMgr.GetTreeObjectByIndex(int(te.TreeObjectID))
		if treeObj == nil {
			continue
		}

		// Calculate frame number: rotation + growth * numRotations
		// OpenLoco: (viewableRotation % numRotations) + growth * numRotations
		// We use rotation=0 for now (no viewport rotation support yet)
		viewRotation := uint32(te.Rotation) % uint32(treeObj.NumRotations)
		treeFrameNum := viewRotation + uint32(te.Growth)*uint32(treeObj.NumRotations)

		// Select season sprite base
		season := te.Season
		if season >= 6 {
			season = 0
		}
		var spriteBase uint32
		if te.HasSnow {
			spriteBase = treeObj.SnowSprites[season]
		} else {
			spriteBase = treeObj.Sprites[season]
		}

		spriteID := int(spriteBase + treeFrameNum)

		img := w.renderer.GetSprite(spriteID)
		if img == nil {
			continue
		}

		_, _, xOff, yOff, ok := w.renderer.GetSpriteInfo(spriteID)
		if !ok {
			continue
		}

		// Position within tile based on quadrant
		quadrant := te.Quadrant % 4
		qx := float64(kTreeQuadrantOffset[quadrant][0]) - 16 // center around tile
		qy := float64(kTreeQuadrantOffset[quadrant][1]) - 16

		// Height offset: drawY already includes the surface baseZ, so we
		// only need the extra height if the tree sits above the surface.
		extraHeight := float64(int(te.BaseZ)-int(t.baseZ)) * 4.0

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(
			drawX+(float64(xOff)+qx)*scale,
			drawY+(float64(yOff)-extraHeight+qy)*scale)
		screen.DrawImage(img, op)
	}
}

// paintBuildings renders building sprites for a tile.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintBuilding.cpp  paintBuilding()
// Simplified: renders the first variation's parts stacked vertically.
var buildingDrawCount int
var buildingSkipNilImg int
var buildingSkipSpriteNil int
var buildingSkipEarlyReturn int

func (w *World) paintBuildings(screen *ebiten.Image, t *tile, drawX, drawY, scale float64) {
	if w.renderer == nil || w.renderer.ObjMgr == nil || len(t.buildings) == 0 {
		if len(t.buildings) > 0 {
			buildingSkipEarlyReturn++ // renderer/ObjMgr nil but tile has buildings
		}
		return
	}

	for _, be := range t.buildings {
		bldgObj := w.renderer.ObjMgr.GetBuildingObjectByIndex(int(be.ObjectID))
		if bldgObj == nil || bldgObj.ImageOffset == 0 {
			buildingSkipNilImg++
			continue
		}

		// Only render constructed buildings (skip under-construction for now)
		if !be.IsConstructed {
			continue
		}

		rotation := be.Rotation & 0x03
		variation := be.Variation
		if int(variation) >= int(bldgObj.NumVariations) {
			variation = 0
		}

		// For multi-tile (2x2) buildings only the "front" tile renders the sprites.
		// OpenLoco: (sequenceIndex ^ 2) == ((-viewportRotation) & 3).
		// With fixed viewport rotation=0: render seqIdx==2, skip all others.
		if bldgObj.HasFlags(objects.BuildingFlagLargeTile) && be.SequenceIndex != 2 {
			continue
		}

		parts := bldgObj.GetBuildingParts(variation)
		if len(parts) == 0 {
			continue
		}

		// Height offset for building's own baseZ vs surface baseZ
		extraHeight := float64(int(be.BaseZ)-int(t.baseZ)) * 4.0

		// Sub-tile image offset: OpenLoco kImageOffsetBase1x1={16,16,0}, kImageOffsetBase2x2={0,0,0}.
		// Projected to screen: screenX=wy-wx, screenY=(wx+wy)/2.
		// 1x1: {16,16,0} → screenDelta=(0,16); 2x2: {0,0,0} → screenDelta=(0,0).
		offsetX := float64(0)
		offsetY := float64(16)
		if bldgObj.HasFlags(objects.BuildingFlagLargeTile) {
			offsetX = 0
			offsetY = 0
		}

		// Stack parts vertically
		partZ := float64(0)
		for _, part := range parts {
			// Each part has 4 rotation variants: part * 4 + rotation
			spriteID := int(bldgObj.ImageOffset) + int(part)*4 + int(rotation)

			img := w.renderer.GetSprite(spriteID)
			if img == nil {
				buildingSkipSpriteNil++
				continue
			}

			_, _, xOff, yOff, ok := w.renderer.GetSpriteInfo(spriteID)
			if !ok {
				continue
			}

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(
				drawX+(float64(xOff)+offsetX)*scale,
				drawY+(float64(yOff)-extraHeight-partZ+offsetY)*scale)
			screen.DrawImage(img, op)
			buildingDrawCount++

			// Stack next part on top
			if int(part) < len(bldgObj.PartHeights) {
				partZ += float64(bldgObj.PartHeights[part])
			}
		}
	}
}

// noteRenderIssue records a rendering failure on a tile.
// desc should be a short label like "water:no-mask" or "building:nil-img".
func (w *World) noteRenderIssue(tileX, tileY int, desc string) {
	for i := range w.renderIssues {
		if w.renderIssues[i].TileX == tileX && w.renderIssues[i].TileY == tileY {
			w.renderIssues[i].Severity++
			return
		}
	}
	w.renderIssues = append(w.renderIssues, RenderIssue{
		TileX: tileX, TileY: tileY, Severity: 1, Desc: desc,
	})
}

// WorstRenderIssue returns the tile with the most rendering failures from the
// last Draw() call. Returns ok=false if no issues were recorded.
func (w *World) WorstRenderIssue() (tileX, tileY int, desc string, ok bool) {
	if len(w.renderIssues) == 0 {
		return 0, 0, "", false
	}
	best := w.renderIssues[0]
	for _, ri := range w.renderIssues[1:] {
		if ri.Severity > best.Severity {
			best = ri
		}
	}
	return best.TileX, best.TileY, best.Desc, true
}

// LogBuildingStats logs counts of building draw attempts/skips for diagnostics.
func LogBuildingStats() {
	log.Printf("[Buildings] drawn=%d skipNilImg=%d skipSpriteNil=%d skipEarlyReturn=%d", buildingDrawCount, buildingSkipNilImg, buildingSkipSpriteNil, buildingSkipEarlyReturn)
}

// kWallScreenOffsets: screen-pixel (deltaX, deltaY) per rotation.
// OpenLoco source: PaintWall.cpp kOffsets stores world sub-tile Pos3 values:
//
//	rot0={0,0,0}, rot1={1,31,0}, rot2={31,0,0}, rot3={2,1,0}
//
// Projection to screen pixels (matching tileToScreen formula):
//
//	screenX = wy - wx
//	screenY = (wy + wx) / 2
//
// rot0: (0,0)   → (0,   0 )
// rot1: (1,31)  → (30,  16)
// rot2: (31,0)  → (-31, 15)
// rot3: (2,1)   → (-1,  1 )
var kWallScreenOffsets = [4][2]int16{
	{0, 0},    // rotation 0 (SE edge)
	{30, 16},  // rotation 1 (SW edge)
	{-31, 15}, // rotation 2 (NW edge)
	{-1, 1},   // rotation 3 (NE edge)
}

// kWallImageOffsets[rotation][slopeIndex] → WallObj sprite offset.
// OpenLoco source: PaintWall.cpp kImageOffsets[not-twoSided][rotation][slope]
//
//	kFlatSE=0, kFlatNE=1, kSlopedSE=2, kSlopedNE=3, kSlopedNW=4, kSlopedSW=5
//
// slopeIndex: 0=downwards, 1=upwards, 2=flat  (from slopeFlagsToIndex)
var kWallImageOffsets = [4][3]uint32{
	{3, 5, 1}, // rot 0: kSlopedNE, kSlopedSW, kFlatNE
	{2, 4, 0}, // rot 1: kSlopedSE, kSlopedNW, kFlatSE
	{5, 3, 1}, // rot 2: kSlopedSW, kSlopedNE, kFlatNE
	{4, 2, 0}, // rot 3: kSlopedNW, kSlopedSE, kFlatSE
}

// paintBridgeDeck renders the bridge deck for a bridged track or road element.
// It draws the flat deck sprite (BridgeDeckNoSupport) at the clearZ height.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintBridge.cpp paintBridge()
func (w *World) paintBridgeDeck(screen *ebiten.Image, bridgeID uint8, tileBaseZ, elemBaseZ uint8, drawX, drawY, scale float64) {
	bridgeObj := w.renderer.ObjMgr.GetBridgeObjectByIndex(int(bridgeID))
	if bridgeObj == nil || bridgeObj.ImageOffset == 0 {
		return
	}
	spriteID := int(bridgeObj.ImageOffset) + objects.BridgeDeckNoSupport
	// The deck sits directly under the track/road, i.e. at the ELEMENT's base
	// height (clearZ is the clearance top — using it floated every deck a few
	// height units up-left onto the neighbouring terrain).
	extraHeight := float64(int(elemBaseZ)-int(tileBaseZ)) * 4.0
	w.drawTrackRoadSprite(screen, spriteID, drawX, drawY, extraHeight, scale)
}

// paintLevelCrossing renders level crossing gate sprites for a road element.
// Draws 4 sprites at the tile corners (gate positions) for both NE and SE orientations.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintRoad.cpp paintLevelCrossing()
func (w *World) paintLevelCrossing(screen *ebiten.Image, re *scenario.RoadElement, baseZ uint8, drawX, drawY, scale float64) {
	lcObj := w.renderer.ObjMgr.GetLevelCrossingObjectByIndex(int(re.LevelCrossingObjectID))
	if lcObj == nil || lcObj.ImageOffset == 0 {
		return
	}
	frame := int(re.AnimFrame)
	rotation := int(re.Rotation)
	imageIndex0 := int(lcObj.ImageOffset) + ((rotation & 1) * 4) + (frame * 8)

	// 4 gate sprites at sub-tile positions {tileX, tileY}:
	//   +0: {2,2}, +1: {2,30}, +2: {30,2}, +3: {30,30}
	subTilePos := [4][2]float64{{2, 2}, {2, 30}, {30, 2}, {30, 30}}
	for i, pos := range subTilePos {
		tileX, tileY := pos[0], pos[1]
		vpDX := (tileY - tileX) * scale
		vpDY := (tileY + tileX) / 2.0 * scale
		w.drawTrackRoadSprite(screen, imageIndex0+i, drawX+vpDX, drawY+vpDY, 0, scale)
	}
}

// paintTracks renders railway track sprites for a tile.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintTrack.cpp paintTrack()
//
// Sprite layout: kTrackParts[trackID][seqIdx] → image offsets per rotation and layer.
// Mergeable pieces: 3 layers (Ballast=0, Sleeper=1, Rail=2).
// Non-mergeable (slope) pieces: only layer 0.
//
// When multiple track elements share a tile (e.g. a Y-junction or level crossing),
// all Ballast layers must be drawn before all Sleeper layers, and all Sleeper layers
// before all Rail layers.  This matches OpenLoco's addToPlotListTrackRoad() which
// places layers into priority buckets 0/1/3 and flushes them in order.
// Rendering element-by-element (ballast+sleeper+rail for elem1, then elem2) would
// let the second element's ballast overdraw the first element's sleepers/rails.
func (w *World) paintTracks(screen *ebiten.Image, t *tile, drawX, drawY, scale float64) {
	if os.Getenv("GOLOCO_NO_TRACK") == "1" {
		return
	}
	if w.renderer == nil || w.renderer.ObjMgr == nil || len(t.tracks) == 0 {
		return
	}

	// Resolve all track elements into (piece, rotation, extraHeight, baseImage) tuples.
	type resolved struct {
		base        uint32
		piece       trackPiece
		rotation    int
		extraHeight float64
	}
	var elems [8]resolved // fixed array avoids allocation; 8 track elements per tile is more than enough
	n := 0
	for _, te := range t.tracks {
		if n >= len(elems) {
			break
		}
		trackObj := w.renderer.ObjMgr.GetTrackObjectByIndex(int(te.TrackObjectID))
		if trackObj == nil || trackObj.Image == 0 {
			continue
		}
		trackID := int(te.TrackID)
		if trackID >= len(kTrackParts) || kTrackParts[trackID] == nil {
			continue
		}
		parts := kTrackParts[trackID]
		seqIdx := int(te.SequenceIndex)
		if seqIdx >= len(parts) {
			continue
		}
		elems[n] = resolved{
			base:        trackObj.Image,
			piece:       parts[seqIdx],
			rotation:    int(te.Rotation & 0x03),
			extraHeight: float64(int(te.BaseZ)-int(t.baseZ)) * 4.0,
		}
		n++
	}

	// Bridge decks go under the track: draw before ballast/sleepers/rails.
	for _, te := range t.tracks {
		if te.HasBridge {
			w.paintBridgeDeck(screen, te.BridgeID, t.baseZ, te.BaseZ, drawX, drawY, scale)
		}
	}

	// Render layer-by-layer: all Ballast (0), then all Sleeper (1), then all Rail (2).
	// Non-mergeable (slope) pieces only have layer 0 and are skipped in layers 1 and 2.
	for layer := 0; layer < 3; layer++ {
		for i := 0; i < n; i++ {
			el := &elems[i]
			if layer > 0 && el.piece.nonMergeable {
				continue
			}
			offset := el.piece.img[el.rotation][layer]
			if offset == 0 {
				continue
			}
			w.drawTrackRoadSprite(screen, int(el.base)+int(offset), drawX, drawY, el.extraHeight, scale)
		}
	}

}

// paintRoads renders road sprites for a tile.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintRoad.cpp paintRoad()
//
// Road IDs 0–2 (straight, left/right very-small curve) use the multi-tile merge system:
// each element on the tile contributes bits to an exits bitmask, and one composite sprite
// is rendered via kExitsToMergeId + kMergeBaseOffset (like OpenLoco's finalisePaintRoad).
// Road IDs 3–9 render directly from the piece tables.
func (w *World) paintRoads(screen *ebiten.Image, t *tile, drawX, drawY, scale float64) {
	if w.renderer == nil || w.renderer.ObjMgr == nil || len(t.roads) == 0 {
		return
	}

	// Accumulate exits for the merge system, keyed by (roadObjectID, baseZ).
	type mergeKey struct{ objID, baseZ uint8 }
	type mergeState struct {
		roadObj     *objects.RoadObject
		exits       uint8 // Style0 / Style2-left exits
		exitsRight  uint8 // Style2-right exits (dual carriageway only)
		extraHeight float64
		hasLeft     bool
		hasRight    bool
	}
	var merges [8]mergeState // fixed-size to avoid allocation on the hot path
	mergeIdx := map[mergeKey]int{}
	nMerges := 0

	for _, re := range t.roads {
		roadObj := w.renderer.ObjMgr.GetRoadObjectByIndex(int(re.RoadObjectID))
		if roadObj == nil || roadObj.Image == 0 {
			continue
		}

		roadID := int(re.RoadID)
		seqIdx := int(re.SequenceIndex)
		rotation := int(re.Rotation & 0x03)
		extraHeight := float64(int(re.BaseZ)-int(t.baseZ)) * 4.0

		// Road IDs 0–2 use the multi-tile merge (junction) system.
		if roadID <= 2 {
			key := mergeKey{re.RoadObjectID, re.BaseZ}
			idx, ok := mergeIdx[key]
			if !ok {
				if nMerges >= len(merges) {
					continue // safety cap
				}
				idx = nMerges
				nMerges++
				merges[idx] = mergeState{roadObj: roadObj, extraHeight: extraHeight}
				mergeIdx[key] = idx
			}
			exits := kRoadMergeExits[roadID][rotation]
			if roadObj.PaintStyle == 2 {
				// Style2 dual carriageway: straight rot 0/1 = left lane, rot 2/3 = right.
				// Very-small-right-curve → right lane; left-curve → left lane.
				isRight := (roadID == 0 && rotation >= 2) || roadID == 2
				if isRight {
					merges[idx].exitsRight |= exits
					merges[idx].hasRight = true
				} else {
					merges[idx].exits |= exits
					merges[idx].hasLeft = true
				}
			} else {
				merges[idx].exits |= exits
				merges[idx].hasLeft = true
			}
			continue
		}

		// Road IDs 3–9: render directly from the piece tables.
		switch roadObj.PaintStyle {
		case 0:
			if roadID >= len(kRoadPartsStyle0) {
				continue
			}
			pieces := kRoadPartsStyle0[roadID]
			if seqIdx >= len(pieces) {
				continue
			}
			offset := pieces[seqIdx].img[rotation]
			if offset == 0 {
				continue
			}
			w.drawTrackRoadSprite(screen, int(roadObj.Image)+int(offset), drawX, drawY, extraHeight, scale)
		case 2:
			// OpenLoco reference: RoadObject.h Style2 namespace
			if roadID >= len(kRoadPartsStyle2) {
				continue
			}
			pieces := kRoadPartsStyle2[roadID]
			if seqIdx >= len(pieces) {
				continue
			}
			offset := pieces[seqIdx].img[rotation]
			if offset == 0 {
				continue
			}
			w.drawTrackRoadSprite(screen, int(roadObj.Image)+int(offset), drawX, drawY, extraHeight, scale)
		case 1:
			// Style1: 3 layers (Ballast, Sleeper, Rail) — no merge system.
			if roadID >= len(kRoadPartsStyle1) {
				continue
			}
			pieces := kRoadPartsStyle1[roadID]
			if seqIdx >= len(pieces) {
				continue
			}
			piece := pieces[seqIdx]
			for layer := 0; layer < 3; layer++ {
				offset := piece.img[rotation][layer]
				if offset == 0 {
					continue
				}
				w.drawTrackRoadSprite(screen, int(roadObj.Image)+int(offset), drawX, drawY, extraHeight, scale)
			}
		}
	}

	// Render the accumulated merge sprites (straight, curve, junction, or crossroads).
	// OpenLoco reference: PaintRoad.cpp finalisePaintRoad()
	for i := 0; i < nMerges; i++ {
		ms := &merges[i]
		if ms.hasLeft {
			mergeID := kExitsToMergeId[ms.exits&0x0F]
			spriteID := int(ms.roadObj.Image) + int(kMergeBaseOffset) + int(mergeID)
			w.drawTrackRoadSprite(screen, spriteID, drawX, drawY, ms.extraHeight, scale)
		}
		if ms.hasRight {
			mergeID := kExitsToMergeId[ms.exitsRight&0x0F]
			spriteID := int(ms.roadObj.Image) + int(kMergeBaseOffsetStyle2Right) + int(mergeID)
			w.drawTrackRoadSprite(screen, spriteID, drawX, drawY, ms.extraHeight, scale)
		}
	}

	// Draw bridge decks for bridged road elements.
	for _, re := range t.roads {
		if re.HasBridge {
			w.paintBridgeDeck(screen, re.BridgeID, t.baseZ, re.BaseZ, drawX, drawY, scale)
		}
	}

	// Draw level crossings.
	for i := range t.roads {
		re := &t.roads[i]
		if re.HasLevelCrossing {
			w.paintLevelCrossing(screen, re, t.baseZ, drawX, drawY, scale)
		}
	}
}

// drawTrackRoadSprite fetches a sprite and draws it at the correct screen position.
func (w *World) drawTrackRoadSprite(screen *ebiten.Image, spriteID int, drawX, drawY, extraHeight, scale float64) {
	img := w.renderer.GetSprite(spriteID)
	if img == nil {
		return
	}
	_, _, xOff, yOff, ok := w.renderer.GetSpriteInfo(spriteID)
	if !ok {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(
		drawX+float64(xOff)*scale,
		drawY+(float64(yOff)-extraHeight)*scale)
	screen.DrawImage(img, op)
}

// DrawTrackGhost renders a translucent preview of a track piece at tile (tileX, tileY).
// Used to show the player where a piece will be placed before clicking.
// trackID=0 straight, 2=LeftCurveVerySmall, 3=RightCurveVerySmall, etc.
func (w *World) DrawTrackGhost(screen *ebiten.Image, tileX, tileY, trackID, rotation, objID int) {
	if w.renderer == nil || w.renderer.ObjMgr == nil {
		return
	}
	t := w.getTile(tileX, tileY)
	if t == nil {
		return
	}
	trackObj := w.renderer.ObjMgr.GetTrackObjectByIndex(objID)
	if trackObj == nil || trackObj.Image == 0 {
		return
	}
	if trackID >= len(kTrackParts) || kTrackParts[trackID] == nil || len(kTrackParts[trackID]) == 0 {
		return
	}
	piece := kTrackParts[trackID][0]
	rot := rotation & 0x03
	scale := 1.0 / float64(int(1)<<w.zoom)
	vpX, vpY := w.tileToScreen(tileX, tileY)
	vpY -= float64(t.baseZ) * 4.0
	drawX := math.Round((vpX - w.camX) * scale)
	drawY := math.Round((vpY - w.camY) * scale)

	for layer := 0; layer < 3; layer++ {
		offset := piece.img[rot][layer]
		if piece.nonMergeable && layer > 0 {
			break
		}
		if offset == 0 {
			continue
		}
		spriteID := int(trackObj.Image) + int(offset)
		img := w.renderer.GetSprite(spriteID)
		if img == nil {
			continue
		}
		_, _, xOff, yOff, ok := w.renderer.GetSpriteInfo(spriteID)
		if !ok {
			continue
		}
		op := &ebiten.DrawImageOptions{}
		op.ColorScale.ScaleAlpha(0.5)
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(drawX+float64(xOff)*scale, drawY+float64(yOff)*scale)
		screen.DrawImage(img, op)
	}
}

// DrawRoadGhost renders a translucent preview of a road piece at tile (tileX, tileY).
func (w *World) DrawRoadGhost(screen *ebiten.Image, tileX, tileY, roadID, rotation, objID int) {
	if w.renderer == nil || w.renderer.ObjMgr == nil {
		return
	}
	t := w.getTile(tileX, tileY)
	if t == nil {
		return
	}
	roadObj := w.renderer.ObjMgr.GetRoadObjectByIndex(objID)
	if roadObj == nil || roadObj.Image == 0 {
		return
	}
	if roadID >= len(kRoadPartsStyle0) || kRoadPartsStyle0[roadID] == nil || len(kRoadPartsStyle0[roadID]) == 0 {
		return
	}
	rot := rotation & 0x03
	scale := 1.0 / float64(int(1)<<w.zoom)
	vpX, vpY := w.tileToScreen(tileX, tileY)
	vpY -= float64(t.baseZ) * 4.0
	drawX := math.Round((vpX - w.camX) * scale)
	drawY := math.Round((vpY - w.camY) * scale)

	// Road ghost: use the merge sprite for just this piece's exits
	// Straight (roadID=0): exits 0101 (rot0/2) or 1010 (rot1/3)
	var exits uint8
	switch rot {
	case 0, 2:
		exits = 0b0101
	case 1, 3:
		exits = 0b1010
	}
	mergeID := kExitsToMergeId[exits&0x0F]
	spriteID := int(roadObj.Image) + int(kMergeBaseOffset) + int(mergeID)
	img := w.renderer.GetSprite(spriteID)
	if img == nil {
		return
	}
	_, _, xOff, yOff, ok := w.renderer.GetSpriteInfo(spriteID)
	if !ok {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleAlpha(0.5)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(drawX+float64(xOff)*scale, drawY+float64(yOff)*scale)
	screen.DrawImage(img, op)
}

// paintStation renders train station sprites for a tile.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintTrainStation.cpp
//
// Sprite layout (Style0, relative to SequenceImageOffsets[seqIdx]):
//
//	NE orientation (rotation 0/2): +0=BackNE, +1=FrontNE, +2=CanopyNE
//	SE orientation (rotation 1/3): +4=BackSE, +5=FrontSE, +6=CanopySE
func (w *World) paintStation(screen *ebiten.Image, t *tile, drawX, drawY, scale float64) {
	if os.Getenv("GOLOCO_NO_STATION") == "1" {
		return
	}
	if w.renderer == nil || w.renderer.ObjMgr == nil || len(t.stations) == 0 {
		return
	}

	for _, se := range t.stations {
		switch se.StationType {
		case 0: // train station
			w.paintTrainStation(screen, t, &se, drawX, drawY, scale)
		case 1: // road station (bus stop)
			w.paintRoadStation(screen, t, &se, drawX, drawY, scale)
		}
	}
}

// paintSignals renders railway signal sprites for a tile.
//
// Each SignalElement may have a left signal, a right signal, or both.
// The signal post position is taken from SignalPosLeft / SignalPosRight tables
// indexed by trackRotation (0-15).  For straight tracks (trackId=0), the track
// rotation matches the element's Rotation field directly (0-3).
//
// Image offset formula:
//
//	imageRotationOffset = (rotation & 0x3) << 1  (0,2,4,6 for rotations 0-3)
//	spriteID = signalObj.ImageOffset + imageRotationOffset + (frame << 3)
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintSignal.cpp paintSignalSide()
func (w *World) paintSignals(screen *ebiten.Image, t *tile, drawX, drawY, scale float64) {
	if w.renderer == nil || w.renderer.ObjMgr == nil || len(t.signals) == 0 {
		return
	}

	for _, se := range t.signals {
		rot := int(se.Rotation & 0x03)

		// imageRotationOffset: ((rot & 0x3) << 1) — 0,2,4,6
		imageRotOffset := (rot & 0x3) << 1

		extraHeight := float64(int(se.BaseZ)-int(t.baseZ)) * 4.0

		if se.HasLeftSignal {
			sigObj := w.renderer.ObjMgr.GetTrainSignalObjectByIndex(int(se.SignalObjectID))
			if sigObj != nil && sigObj.ImageOffset > 0 {
				frame := int(se.LeftFrame)
				spriteID := int(sigObj.ImageOffset) + imageRotOffset + (frame << 3)
				pos := objects.SignalPosLeft[rot]
				// Convert tile-space offset to viewport-space offset.
				// Tile-space: (sigX, sigY) in 0-31; viewport: dx = sigY-sigX, dy = (sigY+sigX)/2
				sigX, sigY := float64(pos[0]), float64(pos[1])
				vpDX := (sigY - sigX) * scale
				vpDY := (sigY + sigX) / 2.0 * scale
				w.drawTrackRoadSprite(screen, spriteID, drawX+vpDX, drawY+vpDY, extraHeight, scale)
			}
		}

		if se.HasRightSignal {
			sigObj := w.renderer.ObjMgr.GetTrainSignalObjectByIndex(int(se.SignalObjectID))
			if sigObj != nil && sigObj.ImageOffset > 0 {
				frame := int(se.RightFrame)
				spriteID := int(sigObj.ImageOffset) + imageRotOffset + 1 + (frame << 3)
				pos := objects.SignalPosRight[rot]
				sigX, sigY := float64(pos[0]), float64(pos[1])
				vpDX := (sigY - sigX) * scale
				vpDY := (sigY + sigX) / 2.0 * scale
				w.drawTrackRoadSprite(screen, spriteID, drawX+vpDX, drawY+vpDY, extraHeight, scale)
			}
		}
	}
}

// paintTrainStation renders a train station platform sprite for a single StationElement.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintTrainStation.cpp
func (w *World) paintTrainStation(screen *ebiten.Image, t *tile, se *scenario.StationElement, drawX, drawY, scale float64) {
	stationObj := w.renderer.ObjMgr.GetTrainStationObjectByIndex(int(se.ObjectID))
	if stationObj == nil || stationObj.ImageOffset == 0 {
		return
	}

	seqIdx := int(se.SequenceIndex)
	if seqIdx >= len(stationObj.SequenceImageOffsets) {
		seqIdx = 0
	}
	seqBase := int(stationObj.SequenceImageOffsets[seqIdx])

	// Select NE or SE sprite set based on rotation.
	var backOff, frontOff, canopyOff int
	if se.Rotation&1 == 0 { // NE-facing (rotations 0, 2)
		backOff = objects.StStraightBackNE
		frontOff = objects.StStraightFrontNE
		canopyOff = objects.StStraightCanopyNE
	} else { // SE-facing (rotations 1, 3)
		backOff = objects.StStraightBackSE
		frontOff = objects.StStraightFrontSE
		canopyOff = objects.StStraightCanopySE
	}

	extraHeight := float64(int(se.BaseZ)-int(t.baseZ)) * 4.0

	for _, sprOffset := range []int{backOff, frontOff, canopyOff} {
		spriteID := seqBase + sprOffset
		w.drawTrackRoadSprite(screen, spriteID, drawX, drawY, extraHeight, scale)
	}
}

// paintRoadStation renders a road station (bus stop) sprite for a single StationElement.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintRoadStation.cpp
func (w *World) paintRoadStation(screen *ebiten.Image, t *tile, se *scenario.StationElement, drawX, drawY, scale float64) {
	stationObj := w.renderer.ObjMgr.GetRoadStationObjectByIndex(int(se.ObjectID))
	if stationObj == nil || stationObj.ImageOffset == 0 {
		return
	}

	seqIdx := int(se.SequenceIndex)
	if seqIdx >= len(stationObj.SequenceImageOffsets) {
		seqIdx = 0
	}
	seqBase := int(stationObj.SequenceImageOffsets[seqIdx])

	// Select NE or SE sprite set based on rotation.
	var backOff, frontOff, canopyOff int
	if se.Rotation&1 == 0 { // NE-facing (rotations 0, 2)
		backOff = objects.RsStraightBackNE
		frontOff = objects.RsStraightFrontNE
		canopyOff = objects.RsStraightCanopyNE
	} else { // SE-facing (rotations 1, 3)
		backOff = objects.RsStraightBackSE
		frontOff = objects.RsStraightFrontSE
		canopyOff = objects.RsStraightCanopySE
	}

	extraHeight := float64(int(se.BaseZ)-int(t.baseZ)) * 4.0

	for _, sprOffset := range []int{backOff, frontOff, canopyOff} {
		spriteID := seqBase + sprOffset
		w.drawTrackRoadSprite(screen, spriteID, drawX, drawY, extraHeight, scale)
	}
}

// paintWalls renders wall sprites for a tile.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintWall.cpp paintWall()
func (w *World) paintWalls(screen *ebiten.Image, t *tile, drawX, drawY, scale float64) {
	if w.renderer == nil || w.renderer.ObjMgr == nil || len(t.walls) == 0 {
		return
	}

	for _, we := range t.walls {
		wallObj := w.renderer.ObjMgr.GetWallObjectByIndex(int(we.WallObjectID))
		if wallObj == nil || wallObj.Sprite == 0 {
			continue
		}

		rot := we.Rotation & 0x03

		// Slope index from EdgeSlope flags (OpenLoco: slopeFlagsToIndex).
		// EdgeSlope::downwards=bit1, EdgeSlope::upwards=bit0.
		slopeIdx := 2 // flat
		if we.EdgeSlope&0x02 != 0 {
			slopeIdx = 0 // downwards
		} else if we.EdgeSlope&0x01 != 0 {
			slopeIdx = 1 // upwards
		}

		spriteOffset := kWallImageOffsets[rot][slopeIdx]
		spriteID := int(wallObj.Sprite + spriteOffset)

		img := w.renderer.GetSprite(spriteID)
		if img == nil {
			continue
		}

		_, _, xOff, yOff, ok := w.renderer.GetSpriteInfo(spriteID)
		if !ok {
			continue
		}

		// Height offset for wall's baseZ vs surface baseZ
		extraHeight := float64(int(we.BaseZ)-int(t.baseZ)) * 4.0

		// Screen-space edge offset (world sub-tile Pos3 projected to pixels).
		edgeX := float64(kWallScreenOffsets[rot][0])
		edgeY := float64(kWallScreenOffsets[rot][1])

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(
			drawX+(float64(xOff)+edgeX)*scale,
			drawY+(float64(yOff)-extraHeight+edgeY)*scale)
		screen.DrawImage(img, op)
	}
}

// getWaterImage returns a cached solid-blue diamond for the water tile base.
// Uses the actual water colour from the palette at full opacity so that sprite
// cutouts (holes in +30 and transparent +35) never expose the terrain below.
func (w *World) getWaterImage() *ebiten.Image {
	if w.waterImage != nil {
		return w.waterImage
	}
	img := ebiten.NewImage(w.tileW, w.tileH)
	centerX := w.tileW / 2
	centerY := w.tileH / 2
	wc := w.waterColourFromPalette()
	wc.A = 255 // fully opaque base — no see-through gaps
	for y := 0; y < w.tileH; y++ {
		var halfWidth int
		if y < centerY {
			halfWidth = (y * centerX) / centerY
		} else {
			halfWidth = ((w.tileH - 1 - y) * centerX) / centerY
		}
		for x := centerX - halfWidth; x <= centerX+halfWidth; x++ {
			if x >= 0 && x < w.tileW {
				img.Set(x, y, wc)
			}
		}
	}
	w.waterImage = img
	return img
}

// paintWater renders a water surface overlay on a tile.
//
// Approach (2-layer):
//
//	Layer 0 (pass 1, worldBuf): +35 blend sprite with water colour forced, at its
//	  native xOff=-32, yOff=0. The +35 sprite is the game-canonical water surface shape
//	  (a bottom-half-of-diamond, 64×31 pixels). Adjacent water tiles drawn earlier in
//	  depth order cover the "top 15 rows" of each tile's screen area, so tessellation
//	  is seamless across flat ocean. Water cliff edges cover exposed terrain at boundaries.
//	Layer 1 (pass 4, screen): +30 ripple detail with real teal palette colours.
//	Layer 2 (pass 4, screen, zoom 0 only): wave animation frame (+60+frame).
//
// Fallback: when no WaterObj is loaded, uses getWaterImage() hand-drawn diamond.
//
// OpenLoco reference: PaintSurface.cpp:1720-1759 (paintSurfaceWater)
//
//	sprite+35 .withBlend(ExtColour::water) → dst-based palette remap
//	sprite+30 attached → ripple detail on top
//	sprite+60+frame attached → wave animation at zoom 0
func (w *World) paintWater(target *ebiten.Image, tileX, tileY int, t *tile, drawX, drawY, scale float64) {
	// waterLevel is MicroZ (×16 px/unit); baseZ is SmallZ (×4 px/unit).
	waterHeightDiff := float64(t.waterLevel)*16.0 - float64(t.baseZ)*4.0

	if w.renderer == nil || w.renderer.ObjMgr == nil || w.renderer.ObjMgr.WaterObj == nil {
		// Fallback: no water object — use hand-drawn diamond
		baseImg := w.getWaterImage()
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(
			drawX+float64(-32)*scale,
			drawY+float64(-15)*scale-waterHeightDiff*scale)
		target.DrawImage(baseImg, op)
		return
	}

	waterObj := w.renderer.ObjMgr.WaterObj
	// OpenLoco: kSlopeToWaterShape only applied when waterHeight <= baseHeight + kMicroZStep.
	// For deeper water the surface is always flat (shape 0) — avoids triangle artefacts.
	// OpenLoco reference: PaintSurface.cpp paintSurfaceWater() line 1730-1733
	shape := uint8(0)
	if uint16(t.waterLevel)*16 <= uint16(t.baseZ)*4+16 {
		shape = objects.KSlopeToWaterShape[t.slope&0x0F]
	}
	flatID := int(waterObj.GetWaterSpriteIndex(shape, false))
	blendID := int(waterObj.GetWaterSpriteIndex(shape, true))

	// Get the water palette remap table (WaterObject sprite +41).
	// The blend sprite (+35, 64×31) uses palette indices 1-4 which map to
	// water-blue shades via this remap. Without it the blend sprite is transparent.
	// OpenLoco reference: PaintSurface.cpp paintSurfaceWater — withBlend(ExtColour::water)
	waterPalMap, _ := w.renderer.G1.GetPaletteMapAt(int(waterObj.ImageOffset) + 41)

	// queueWater queues a sprite into waterOps (drawn after all terrain in pass 1.5).
	queueWater := func(spriteID int, palMap []byte) bool {
		var img *ebiten.Image
		if palMap != nil {
			img = w.renderer.GetSpriteWithPaletteMap(spriteID, palMap)
		} else {
			img = w.renderer.GetSprite(spriteID)
		}
		if img == nil {
			return false
		}
		_, _, xOff, yOff, ok := w.renderer.GetSpriteInfo(spriteID)
		if !ok {
			xOff, yOff = -32, -15
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(
			drawX+float64(xOff)*scale,
			drawY+float64(yOff)*scale-waterHeightDiff*scale)
		w.waterOps = append(w.waterOps, worldHighlightOp{img: img, opts: op})
		return true
	}

	// Blend sprite (+35, full 64×31 tile) with water remap applied — opaque blue base.
	// Flat ripple sprite (+30, 45×27) on top for surface detail.
	// OpenLoco reference: PaintSurface.cpp paintSurfaceWater()
	if !queueWater(blendID, waterPalMap) {
		w.noteRenderIssue(tileX, tileY, "water:no-blend-sprite")
	}
	queueWater(flatID, nil)

	// Wave animation frames at zoom 0 (+60+frame).
	if w.zoom == 0 {
		const waveFrames = 16
		waveFrame := (w.animTick/4 + tileX + tileY) & (waveFrames - 1)
		queueWater(int(waterObj.ImageOffset)+60+waveFrame, nil)
	}
	_ = target
}

// waterColourFromPalette samples the water palette map (WaterObject sprite +41)
// to derive the actual water blue.  The map converts any terrain index into a
// water-tinted shade; we average several mid-range samples to get a representative
// water colour.
func (w *World) waterColourFromPalette() color.RGBA {
	const defR, defG, defB uint8 = 25, 120, 205 // brighter blue fallback
	if w.renderer == nil || w.renderer.G1 == nil || w.renderer.ObjMgr == nil || w.renderer.ObjMgr.WaterObj == nil {
		return color.RGBA{defR, defG, defB, 255}
	}
	palMap, err := w.renderer.G1.GetPaletteMapAt(int(w.renderer.ObjMgr.WaterObj.ImageOffset) + 41)
	if err != nil {
		return color.RGBA{defR, defG, defB, 255}
	}
	samples := []byte{10, 30, 60, 90, 120, 150, 180, 210}
	var rSum, gSum, bSum uint32
	for _, idx := range samples {
		c := w.renderer.G1.Palette[palMap[idx]]
		rSum += uint32(c.R)
		gSum += uint32(c.G)
		bSum += uint32(c.B)
	}
	n := uint32(len(samples))
	wc := color.RGBA{R: uint8(rSum / n), G: uint8(gSum / n), B: uint8(bSum / n), A: 255}
	// Sanity check: water should be bluer than red; use fallback if not.
	if int(wc.R) > int(wc.B) {
		return color.RGBA{defR, defG, defB, 255}
	}
	return wc
}

// getCliffEdgeOffset converts kEdgeImageOffset[edge] + Pos3{0,0,h*kMicroZStep}
// to viewport (vpX_off, vpY_off) using gameToScreen:
//
//	vpX_off = posY - posX
//	vpY_off = (posY+posX)/2 - (posZ_base + h*16)
//
// h is in MicroZ units; 1 MicroZ = 16 viewport pixels (kMicroZStep=16).
// See MEASUREMENTS.md "Cliff edge Pos3 → screen offset" table.
func (w *World) getCliffEdgeOffset(edge int, h uint8) (int16, int16) {
	heightPx := int16(h) * 16
	switch edge {
	case EdgeSW: // Pos3{30,0,0}: vpX=0-30=-30, vpY=(0+30)/2-0=15
		return -30, 15 - heightPx
	case EdgeSE: // Pos3{0,30,0}: vpX=30-0=30, vpY=(30+0)/2-0=15
		return 30, 15 - heightPx
	case EdgeNW: // Pos3{0,-2,1}: vpX=-2-0=-2, vpY=(-2+0)/2-1=-2
		return -2, -2 - heightPx
	case EdgeNE: // Pos3{-2,0,1}: vpX=0-(-2)=2, vpY=(0-2)/2-1=-2
		return 2, -2 - heightPx
	default:
		return 0, -heightPx
	}
}

func (w *World) Draw(screen *ebiten.Image) {
	sw := screen.Bounds().Dx()
	sh := screen.Bounds().Dy()
	scale := 1.0 / float64(int(1)<<w.zoom) // zoom 0→1.0, 1→0.5, 2→0.25, 3→0.125

	// ── Allocate / reuse off-screen buffers ──
	if w.worldBuf == nil || w.worldBuf.Bounds().Dx() != sw || w.worldBuf.Bounds().Dy() != sh {
		w.worldBuf = ebiten.NewImage(sw, sh)
		w.blendBuf = ebiten.NewImage(sw, sh)
	}
	w.worldBuf.Clear()
	w.blendBuf.Clear()
	w.blendOps = w.blendOps[:0]
	w.highlightOps = w.highlightOps[:0]
	w.waterOps = w.waterOps[:0]
	w.renderIssues = w.renderIssues[:0]

	// ── Pass 1: render all opaque sprites to worldBuf ──
	// Camera coords are in world-space (unscaled). They are set by
	// PanCamera (gameplay) or SetCamera (title sequence).

	// Precompute per-frame constants for visible y-range within each depth slice.
	//
	// For depth d (x+y=d): screenX(y) = (2y−d)·32, screenY(d) = d·16.
	// drawX(y) = (screenX(y) − camX)·scale; visible when drawX ∈ [−128, sw+64].
	// Solving: y ∈ [depth/2 + camX/64 − 128·invScale/64,
	//               depth/2 + camX/64 + (sw+64)·invScale/64]
	// The constant term (camX/64 ± margin) is computed once here.
	invScale := float64(int(1) << w.zoom) // 2^zoom = 1/scale
	xBase := w.camX / 64.0
	xLeftMargin := -128.0 * invScale / 64.0
	xRightMargin := float64(sw+64) * invScale / 64.0

	// Y-cull for depth slice: screenY = depth·16; height offset can shift a tile
	// up by at most ~336 px (baseZ max=84, ×4). Use 512 for extra safety.
	const maxHeightOffsetPx = 512.0

	// Draw tiles in depth order (back to front)
	for depth := 0; depth < w.width+w.height; depth++ {
		// Y-cull entire depth slice before iterating tiles within it.
		drawYBase := (float64(depth)*16.0 - w.camY) * scale
		if drawYBase < -(float64(sh)+maxHeightOffsetPx) || drawYBase > float64(sh)+maxHeightOffsetPx {
			continue
		}

		// Restrict y to the range whose tiles land within the horizontal screen bounds.
		depthHalf := float64(depth) * 0.5
		yMinF := depthHalf + xBase + xLeftMargin
		yMaxF := depthHalf + xBase + xRightMargin
		yMin := int(math.Floor(yMinF))
		yMax := int(math.Ceil(yMaxF))
		if yMin < 0 {
			yMin = 0
		}
		if yMax >= w.height {
			yMax = w.height - 1
		}

		for y := yMin; y <= yMax; y++ {
			x := depth - y
			if x < 0 || x >= w.width {
				continue
			}

			t := w.tiles[y][x]
			screenX, screenY := w.tileToScreen(x, y)

			// OpenLoco: loc.z = baseZ * kSmallZStep = baseZ * 4 viewport pixels.
			// Cliff edge sprites fill the gaps at height transitions.
			// OpenLoco ref: TileElementBase.h baseHeight() = _baseZ * kSmallZStep
			heightOffset := float64(t.baseZ) * 4.0
			screenY -= heightOffset

			// Convert world position to screen position: subtract camera,
			// then apply zoom scale. Round to nearest integer so adjacent tile
			// anchors snap consistently — sub-pixel sprite offsets then tend
			// toward overlap rather than leaving gaps.
			drawX := math.Round((screenX - w.camX) * scale)
			drawY := math.Round((screenY - w.camY) * scale)

			// Skip tiles that are entirely off-screen (with margin for sprite offsets)
			if drawX < -128 || drawX > float64(sw)+64 || drawY < -128 || drawY > float64(sh)+64 {
				continue
			}

			// Draw terrain using LandObject embedded sprites
			// OpenLoco reference: Paint/PaintSurface.cpp paintSurface()
			//   imageIndex = landObj->image + variation + displaySlope
			if w.renderer != nil && w.renderer.ObjMgr != nil {
				land := w.renderer.ObjMgr.GetLandObjectByIndex(int(t.terrainIndex))
				if t.isIndustrial {
					// TODO: industry fields draw via IndustryObject
					// fieldImageIds (PaintSurface.cpp isIndustrial branch).
					// Until industry objects load, draw bare land (slot 0,
					// growth 0) instead of leaving a hole.
					land = w.renderer.ObjMgr.GetLandObjectByIndex(0)
				}
				if land != nil {
					// Map slope byte (0-31) to display slope (0-18)
					rawSlope := t.slope & 0x1F // Lower 5 bits
					displaySlope := int(0)
					if rawSlope < uint8(len(slopeToDisplaySlope)) {
						displaySlope = int(slopeToDisplaySlope[rawSlope])
					}

					// OpenLoco PaintSurface.cpp ~line 676-688: when water is above the
					// terrain base, the terrain is drawn as flat (slope=0) AT THE WATER
					// HEIGHT so it is hidden under the water sprite on top.
					// OpenLoco ref: `slope = 0; height = elSurface.waterHeight();`
					waterHeightDiff := float64(t.waterLevel)*16.0 - float64(t.baseZ)*4.0
					isSubmerged := t.waterLevel > 0 && waterHeightDiff > 0
					if isSubmerged {
						displaySlope = 0
					}

					// Zoom-0 flat terrain sprite: skip all zoom3/zoom2/zoom1 images.
					// Growth stage selects the season/maturity variant within the zoom-0 block.
					// OpenLoco reference: src/OpenLoco/src/Objects/LandObject.cpp getTerrainImage()
					//   variation = numImagesPerGrowthStage * growthStage
					//   imageIndex = image + variation + displaySlope
					effGrowth := int(t.growthStage)
					if t.isIndustrial || effGrowth >= int(land.NumGrowthStages) {
						effGrowth = 0 // industrial growth stages index industry fields, not land sets
					}
					spriteIdx := land.GetFlatTerrainSpriteIndex() +
						effGrowth*int(land.NumImagesPerGrowthStage) + displaySlope
					// Fully-grown flat tiles with a non-zero per-tile variation
					// draw one of the dedicated variation images instead of the
					// plain flat image, breaking up texture repetition.
					// OpenLoco reference: Paint/PaintSurface.cpp ~0x00465E92
					//   image = mapPixelImage + 3 + variation
					if !t.isIndustrial && t.variation != 0 && displaySlope == 0 && t.waterLevel == 0 &&
						land.NumGrowthStages > 0 && t.growthStage == land.NumGrowthStages-1 {
						spriteIdx = int(land.MapPixelImage-land.ImageOffset) + 3 + int(t.variation)
					}

					if img := w.renderer.GetObjectSprite(land, spriteIdx); img != nil {
						_, _, xOff, yOff, ok := w.renderer.GetObjectSpriteInfo(land, spriteIdx)
						if ok {
							// For submerged tiles: draw terrain sprite at water height (not
							// terrain height) so it sits flush under the water surface.
							terrainDrawY := drawY
							if isSubmerged {
								terrainDrawY = drawY - waterHeightDiff*scale
							}
							op := &ebiten.DrawImageOptions{}
							op.GeoM.Scale(scale, scale)
							op.GeoM.Translate(
								drawX+float64(xOff)*scale,
								terrainDrawY+float64(yOff)*scale)
							w.worldBuf.DrawImage(img, op)

							// Blend neighbouring terrain types along coplanar edges
							w.paintSurfaceSmoothing(w.worldBuf, x, y, &t, land, drawX, terrainDrawY, scale)

							// Draw cliff edges for height transitions
							if land.CliffEdgeImage > 0 {
								w.paintCliffEdges(w.worldBuf, x, y, &t, land, drawX, drawY, scale)
							}

							// Queue water blend ops + draw highlights (pass 2/4)
							if t.waterLevel > 0 && uint16(t.waterLevel)*16 > uint16(t.baseZ)*4 {
								w.paintWater(w.worldBuf, x, y, &t, drawX, drawY, scale)
								w.paintWaterCliffEdges(x, y, &t, land, drawX, drawY, scale)
							}

							// Draw infrastructure and scenery on this tile
							w.paintTracks(w.worldBuf, &t, drawX, drawY, scale)
							w.paintRoads(w.worldBuf, &t, drawX, drawY, scale)
							w.paintStation(w.worldBuf, &t, drawX, drawY, scale)
							w.paintSignals(w.worldBuf, &t, drawX, drawY, scale)
							w.paintTrees(w.worldBuf, &t, drawX, drawY, scale)
							w.paintBuildings(w.worldBuf, &t, drawX, drawY, scale)
							w.paintWalls(w.worldBuf, &t, drawX, drawY, scale)

							continue
						}
					}
				}
			}

			// Terrain sprite missing or land obj nil — record issue but don't draw.
			// Previously drew a debug diamond here; suppress in normal rendering.
			w.noteRenderIssue(x, y, "terrain:fallback-diamond")
		}
	}

	// ── Pass 1.5: flush water surface sprites after ALL terrain ──
	// Water must be drawn after all opaque terrain so it appears on top
	// rather than being overwritten by terrain tiles drawn "in front".
	for _, hi := range w.waterOps {
		w.worldBuf.DrawImage(hi.img, hi.opts)
	}

	// ── Pass 2: apply blend ops (CPU terrain tint) → blendBuf ──
	// Sample the original terrain from worldBuf, tint it toward the water colour
	// where the mask is opaque, then draw into blendBuf.
	// worldBuf is never modified, so every terrain sample is pristine.
	for _, op := range w.blendOps {
		if op.tileRect.Empty() || op.maskImg == nil {
			continue
		}
		mw, mh := op.maskImg.Bounds().Dx(), op.maskImg.Bounds().Dy()
		if mw <= 0 || mh <= 0 {
			continue
		}
		terrainImg := ebiten.NewImage(mw, mh)
		copyOpts := &ebiten.DrawImageOptions{}
		copyOpts.GeoM.Translate(-float64(op.tileRect.Min.X), -float64(op.tileRect.Min.Y))
		terrainImg.DrawImage(w.worldBuf, copyOpts)
		maskPx := make([]byte, mw*mh*4)
		op.maskImg.ReadPixels(maskPx)
		terrainPx := make([]byte, mw*mh*4)
		terrainImg.ReadPixels(terrainPx)
		tR := float32(op.tint.R)
		tG := float32(op.tint.G)
		tB := float32(op.tint.B)
		s := op.strength
		for i := 0; i < len(maskPx); i += 4 {
			if maskPx[i+3] == 0 {
				terrainPx[i+3] = 0 // transparent outside mask — worldBuf shows through
				continue
			}
			terrainPx[i] = byte(float32(terrainPx[i])*(1-s) + tR*s)
			terrainPx[i+1] = byte(float32(terrainPx[i+1])*(1-s) + tG*s)
			terrainPx[i+2] = byte(float32(terrainPx[i+2])*(1-s) + tB*s)
			terrainPx[i+3] = 255
		}
		terrainImg.WritePixels(terrainPx)
		drawOpts := &ebiten.DrawImageOptions{}
		drawOpts.GeoM.Translate(float64(op.tileRect.Min.X), float64(op.tileRect.Min.Y))
		w.blendBuf.DrawImage(terrainImg, drawOpts)
	}

	// ── Pass 3: composite worldBuf + blendBuf → screen ──
	screen.DrawImage(w.worldBuf, nil)
	screen.DrawImage(w.blendBuf, nil)

	// ── Pass 4: highlight/wave sprites on top of everything ──
	for _, hi := range w.highlightOps {
		screen.DrawImage(hi.img, hi.opts)
	}

	// ── Vehicle rendering — post-tile pass (depth approximation) ──
	// Use real parsed entities when available (loaded .SV5), otherwise demo trains.
	if len(w.entities) > 0 {
		w.paintVehicleEntities(screen, scale)
	} else {
		w.paintDemoTrains(screen, scale)
	}
}

// paintDemoTrains renders all active demo trains after the tile loop.
// OpenLoco reference: src/OpenLoco/src/Paint/PaintVehicle.cpp
func (w *World) paintDemoTrains(screen *ebiten.Image, scale float64) {
	if w.renderer == nil || w.renderer.ObjMgr == nil {
		return
	}
	vehicles := w.renderer.ObjMgr.Vehicles
	if len(vehicles) == 0 {
		return
	}

	for _, tr := range w.trains {
		if tr.tileX < 0 || tr.tileX >= w.width || tr.tileY < 0 || tr.tileY >= w.height {
			continue
		}

		// Fractional world position (smooth between tiles)
		fracX := float64(tr.tileX) + float64(tr.dx)*tr.progress
		fracY := float64(tr.tileY) + float64(tr.dy)*tr.progress
		vpX := (fracY - fracX) * 32.0
		vpY := (fracY + fracX) * 16.0

		// Height: interpolate between current and next tile
		baseZ := float64(w.tiles[tr.tileY][tr.tileX].baseZ)
		nx := tr.tileX + tr.dx
		ny := tr.tileY + tr.dy
		if nx >= 0 && nx < w.width && ny >= 0 && ny < w.height {
			nextZ := float64(w.tiles[ny][nx].baseZ)
			baseZ = baseZ*(1.0-tr.progress) + nextZ*tr.progress
		}
		vpY -= baseZ * 4.0

		drawX := math.Round((vpX-w.camX)*scale) - 16 // centre horizontally
		drawY := math.Round((vpY-w.camY)*scale) - 16 // centre vertically

		// Direction → sprite rotation (32-direction wheel; 0=SE, 8=SW, 16=NW, 24=NE)
		dir32 := 0
		switch {
		case tr.dy > 0 && tr.dx == 0:
			dir32 = 0
		case tr.dx > 0 && tr.dy == 0:
			dir32 = 8
		case tr.dy < 0 && tr.dx == 0:
			dir32 = 16
		case tr.dx < 0 && tr.dy == 0:
			dir32 = 24
		}

		v := vehicles[tr.vehicleIdx%len(vehicles)]
		if v.ImageOffset == 0 {
			continue // sprites not loaded
		}

		// Use the body sprite for car component 0
		bodyIdx := int(v.CarComponents[0].BodySpriteIdx)
		spriteID, ok := v.GetBodySpriteID(bodyIdx, dir32)
		if !ok {
			continue
		}

		// Apply company-colour palette remap (default ColourBlue=7).
		// OpenLoco reference: PaintVehicle.cpp — ImageId::withPrimary(companyColour)
		const defaultCompanyColour = 7
		img := w.renderer.GetSpriteColoured(int(spriteID), defaultCompanyColour)
		if img == nil {
			img = w.renderer.GetSprite(int(spriteID))
		}
		if img == nil {
			continue
		}
		_, _, xOff, yOff, hasInfo := w.renderer.GetSpriteInfo(int(spriteID))
		if hasInfo {
			drawX += float64(xOff) * scale
			drawY += float64(yOff) * scale
		}

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(drawX, drawY)
		screen.DrawImage(img, op)
	}
}

// paintVehicleEntities renders VehicleBody entities parsed from an .SV5 save file.
// Each body entity carries its own tile position (tileX/tileY) and sprite info
// (objectId, objectSpriteType, spriteYaw) — no interpolation needed.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintVehicle.cpp
func (w *World) paintVehicleEntities(screen *ebiten.Image, scale float64) {
	if w.renderer == nil || w.renderer.ObjMgr == nil {
		return
	}
	vehicles := w.renderer.ObjMgr.Vehicles
	if len(vehicles) == 0 {
		return
	}

	for i := range w.entities {
		ent := &w.entities[i]
		if ent.EntityType != scenario.VehicleTypeBody {
			continue // only Body entities have renderable car sprites
		}

		// Derive tile indices from PosX/PosY using the WorldUnits type methods.
		// Unplaced vehicles have PosX=PosY=0 (map corner) — skip tile ≤ 0.
		// OpenLoco reference: VehicleHead::Status::unplaced → position is zeroed.
		tx := ent.PosX.Tile()
		ty := ent.PosY.Tile()
		if tx <= 0 || ty <= 0 || tx.Int() >= w.width || ty.Int() >= w.height {
			continue
		}

		// Convert tile → viewport (same formula as tileToScreen).
		// Use sub-tile offset within the tile for fractional precision.
		fracX := float64(tx) + float64(ent.PosX.SubOffset())/32.0
		fracY := float64(ty) + float64(ent.PosY.SubOffset())/32.0

		vpX := (fracY - fracX) * 32.0
		vpY := (fracY + fracX) * 16.0

		// Height: use entity's PosZ (world SmallZ) for pixel offset.
		// OpenLoco formula: heightPx = posZ * 4 (kSmallZStep = 4 px/unit).
		// We apply the current global compromise (×2 instead of ×4).
		vpY -= float64(ent.PosZ) * 2.0

		drawX := math.Round((vpX-w.camX)*scale) - 16
		drawY := math.Round((vpY-w.camY)*scale) - 16

		// Vehicle object lookup.  objectId is 0-based index into loaded vehicles.
		if int(ent.ObjectID) >= len(vehicles) {
			continue
		}
		v := vehicles[ent.ObjectID]
		if v == nil || v.ImageOffset == 0 {
			continue
		}

		// Direction: spriteYaw is stored 0-63 (64 steps per full rotation).
		// Halve to get 0-31 for 32-direction sprite tables.
		dir32 := int(ent.SpriteYaw) >> 1

		bodyIdx := int(ent.ObjectSpriteType)
		spriteID, ok := v.GetBodySpriteID(bodyIdx, dir32)
		if !ok {
			continue
		}
		// Guard against out-of-range sprite IDs (corrupt entity data).
		if w.renderer.G1 == nil || int(spriteID) >= len(w.renderer.G1.Elements) {
			continue
		}

		// Apply company-colour palette remap (default ColourBlue=7).
		// OpenLoco reference: PaintVehicle.cpp — ImageId::withPrimary(companyColour)
		const defaultCompanyColour = 7
		img := w.renderer.GetSpriteColoured(int(spriteID), defaultCompanyColour)
		if img == nil {
			img = w.renderer.GetSprite(int(spriteID))
		}
		if img == nil {
			continue
		}
		_, _, xOff, yOff, hasInfo := w.renderer.GetSpriteInfo(int(spriteID))
		if hasInfo {
			drawX += float64(xOff) * scale
			drawY += float64(yOff) * scale
		}

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(drawX, drawY)
		screen.DrawImage(img, op)
	}
}
