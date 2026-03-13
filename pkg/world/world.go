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
	trains   []*DemoTrain            // demo trains (title screen / scenarios without saved entities)
	entities []scenario.VehicleEntity // real vehicle entities from .SV5 saved-game files
	// isometric tile dimensions (standard 2:1 ratio)
	tileW int // tile width in pixels (64)
	tileH int // tile height in pixels (32)
	// cache colored diamond images per tile type (fallback)
	tileCache  map[TileType]*ebiten.Image
	waterImage          *ebiten.Image          // solid water base diamond (cached)
	waterBaseCache      map[int]*ebiten.Image  // +35 blend sprite with water colour forced (per shape)
	waterHighlightCache map[int]*ebiten.Image  // +30 highlight sprites (real teal colours)
	waterWaveCache      map[int]*ebiten.Image  // wave frame sprites (real palette colours)
	// camera offset in pixels (in world-space before zoom is applied)
	camX float64
	camY float64
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
	worldBuf      *ebiten.Image          // pass-1 target: all opaque sprites
	blendBuf      *ebiten.Image          // pass-2 target: blend-mode overlay
	blendOps     []worldBlendOp     // queued during pass 1
	highlightOps []worldHighlightOp // ripple/wave sprites drawn after blend

	// step-based full-map PNG save (one chunk per frame)
	mapSave *mapSaveState
}

// mapSaveState holds in-progress full-map PNG render state.
type mapSaveState struct {
	filename        string
	final           *image.RGBA
	imgW, imgH      int
	minVpX, minVpY  float64
	chunkX, chunkY  int
	total, done     int
	chunkPix        []byte
	// saved world state restored when saving is complete
	savedCamX, savedCamY float64
	savedZoom             int
	savedExternal         bool
	savedWorldBuf         *ebiten.Image
	savedBlendBuf         *ebiten.Image
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
		renderer:  r,
		width:     20,
		height:    15,
		tileW:     64, // standard isometric tile width
		tileH:     32, // standard isometric tile height (64×32 = 2:1 ratio)
		tiles:     make([][]tile, 15),
		tileCache: make(map[TileType]*ebiten.Image),
		zoom:      0, // start at full zoom (1× scale)
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
		minVpX:   minVpX, minVpY: minVpY,
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
		if len(w.entities) > 0 {
			e := w.entities[0]
			tx, ty := int(e.TileX), int(e.TileY)
			log.Printf("[World] CenterOn(%q): entity at (%d,%d)", kind, tx, ty)
			w.SetCamera(float64(tx), float64(ty))
			w.externalCamera = false
			return true
		}
		if len(w.trains) > 0 {
			tr := w.trains[0]
			log.Printf("[World] CenterOn(%q): demo train at (%d,%d)", kind, tr.tileX, tr.tileY)
			w.SetCamera(float64(tr.tileX), float64(tr.tileY))
			w.externalCamera = false
			return true
		}
		log.Printf("[World] CenterOn(%q): no trains found", kind)
		return false
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

// getFallbackImage returns a cached colored diamond for a tile type.
// Used when no LandObject sprite is available.
func (w *World) getFallbackImage(tt TileType) *ebiten.Image {
	if img, ok := w.tileCache[tt]; ok {
		return img
	}
	img := w.createDiamondTile(tt)
	w.tileCache[tt] = img
	return img
}

// createDiamondTile creates a simple diamond-shaped tile image
func (w *World) createDiamondTile(tt TileType) *ebiten.Image {
	img := ebiten.NewImage(w.tileW, w.tileH)

	var mainColor, darkColor, lightColor color.RGBA
	switch tt {
	case TileGrass:
		mainColor = color.RGBA{80, 160, 80, 255}
		darkColor = color.RGBA{60, 140, 60, 255}
		lightColor = color.RGBA{100, 180, 100, 255}
	case TileDirt:
		mainColor = color.RGBA{139, 90, 43, 255}
		darkColor = color.RGBA{100, 65, 30, 255}
		lightColor = color.RGBA{170, 120, 70, 255}
	case TileWater:
		mainColor = color.RGBA{64, 120, 192, 255}
		darkColor = color.RGBA{40, 90, 160, 255}
		lightColor = color.RGBA{100, 160, 220, 255}
	}

	centerX := w.tileW / 2
	centerY := w.tileH / 2

	for y := 0; y < w.tileH; y++ {
		var halfWidth int
		if y < centerY {
			halfWidth = (y * centerX) / centerY
		} else {
			halfWidth = ((w.tileH - 1 - y) * centerX) / centerY
		}

		for x := centerX - halfWidth; x <= centerX+halfWidth; x++ {
			if x < 0 || x >= w.tileW {
				continue
			}

			var c color.RGBA
			if y < centerY {
				if x < centerX {
					c = lightColor
				} else {
					c = mainColor
				}
			} else {
				if x < centerX {
					c = mainColor
				} else {
					c = darkColor
				}
			}

			img.Set(x, y, c)
		}
	}

	return img
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
	sx := int((vpX-w.camX)*scale)
	sy := int((vpY-w.camY)*scale)
	return sx, sy
}

// ScreenToTile converts a screen pixel (sx, sy) to tile coordinates.
// Returns (-1, -1) if the position maps outside the map bounds.
// Height is ignored (tiles are treated as flat) — small error on steep terrain is acceptable.
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
	tileX := int(math.Floor((vpY/16.0 - vpX/32.0) / 2.0))
	tileY := int(math.Floor((vpX/32.0 + vpY/16.0) / 2.0))
	if tileX < 0 || tileX >= w.width || tileY < 0 || tileY >= w.height {
		return -1, -1
	}
	return tileX, tileY
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
//   SW: Pos2{+32,0}  → tileX+1
//   SE: Pos2{0,+32}  → tileY+1
//   NW: Pos2{0,-32}  → tileY-1
//   NE: Pos2{-32,0}  → tileX-1
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
	Self0      uint8 // First corner on self side
	Neighbor0  uint8 // First corner on neighbor side
	Self1      uint8 // Second corner on self side
	Neighbor1  uint8 // Second corner on neighbor side
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

// kCliffEdgeMaskBase[edge] is the G1 index of the first mask sprite for that edge.
// OpenLoco reference: PaintSurface.cpp kEdgeMaskImageFromSlope, ImageIds.h 3726-3735
// Masks: cliffEdge0MaskSlope0..4 = 3726-3730 (SW, NE)
//        cliffEdge1MaskSlope0..4 = 3731-3735 (SE, NW)
// [0]=body, [1]=top-equalSelf, [2]=top-other, [3]=bottom-n1<n0, [4]=bottom-n1>n0
var kCliffEdgeMaskBase = [4]int{3726, 3731, 3731, 3726} // indexed by edge (SW,SE,NW,NE)

// paintCliffEdges renders cliff edge sprites for height transitions
// OpenLoco reference: PaintSurface.cpp paintSurfaceCliffEdge(), paintEdgeSection()
func (w *World) paintCliffEdges(screen *ebiten.Image, x, y int, t *tile, land *objects.LandObject, drawX, drawY, scale float64) {
	if w.renderer == nil || w.renderer.G1 == nil || land == nil {
		return
	}

	// drawY was computed with the tile height already subtracted (heightOffset = baseZ*4).
	// getCliffEdgeOffset returns offsets relative to the tile's FLAT (zero-height) viewport
	// position, so we need drawY without the height contribution.
	// drawYFlat = drawY + baseZ*4*scale restores the flat reference.
	drawYFlat := drawY + float64(t.baseZ)*4.0*scale

	// Get corner heights for this tile
	selfCorners := w.getCornerHeights(t)

	// Check all 4 edges
	edges := []int{EdgeSW, EdgeSE, EdgeNW, EdgeNE}

	for _, edge := range edges {
		neighbor := w.getNeighborTile(x, y, edge)
		if neighbor == nil {
			continue
		}

		edgeHeights := w.getEdgeHeights(x, y, edge, selfCorners)

		if edgeHeights.Self0 <= edgeHeights.Neighbor0 && edgeHeights.Self1 <= edgeHeights.Neighbor1 {
			continue
		}

		cliffEdgeImageBase := land.CliffEdgeImage

		minHeight := edgeHeights.Neighbor0
		if edgeHeights.Neighbor1 < minHeight {
			minHeight = edgeHeights.Neighbor1
		}
		maxHeight := edgeHeights.Self0
		if edgeHeights.Self1 > maxHeight {
			maxHeight = edgeHeights.Self1
		}

		// Checkerboard: OpenLoco uses world coords (tileX*32)^(tileY*32) & 0x20.
		// Equivalent with tile coords: ((tileX^tileY) & 1) * 32 — alternates every tile.
		checkerboard := uint32((x^y)&1) * 32
		factor := checkerboard + kEdgeFactorOffset[edge]

		maskBase := kCliffEdgeMaskBase[edge]
		for h := minHeight; h < maxHeight; h++ {
			spriteID := int(cliffEdgeImageBase) + int(factor) + int(h&0xF)

			// Select mask slope variant:
			//   body sections use slope0 (maskBase+0)
			//   topmost section uses slope1 when both self corners equal h+1,
			//   slope2 otherwise.
			// OpenLoco reference: PaintSurface.cpp paintSurfaceCliffEdgeImpl maskArr[]
			var slopeVariant int
			if h == maxHeight-1 {
				if edgeHeights.Self0 == edgeHeights.Self1 {
					slopeVariant = 1
				} else {
					slopeVariant = 2
				}
			}
			maskID := maskBase + slopeVariant

			if img := w.renderer.GetMaskedSprite(spriteID, maskID); img != nil {
				_, _, xOff, yOff, ok := w.renderer.GetSpriteInfo(spriteID)
				if ok {
					edgeOffsetX, edgeOffsetY := w.getCliffEdgeOffset(edge, h)

					op := &ebiten.DrawImageOptions{}
					op.GeoM.Scale(scale, scale)
					op.GeoM.Translate(
						drawX+float64(xOff+edgeOffsetX)*scale,
						drawYFlat+float64(yOff+edgeOffsetY)*scale)
					screen.DrawImage(img, op)
				}
			}
		}
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

// paintWaterCliffEdges queues cliff-edge sprites at the water-surface level into
// highlightOps so they render after the blend pass (pass 4) and are not covered
// by the water tint.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintSurface.cpp
//   paintSurfaceWaterCliffEdge() — called for all 4 edges after paintWater
func (w *World) paintWaterCliffEdges(x, y int, t *tile, land *objects.LandObject, drawX, drawY, scale float64) {
	if w.renderer == nil || w.renderer.G1 == nil || land == nil || land.CliffEdgeImage == 0 {
		return
	}
	if t.waterLevel == 0 {
		return
	}

	// drawYFlat is the Y position of this tile at terrain base (height zero reference).
	drawYFlat := drawY + float64(t.baseZ)*4.0*scale

	checkerboard := uint32((x^y)&1) * 32
	cliffEdgeImageBase := land.CliffEdgeImage
	waterMicroZ := t.waterLevel // waterLevel is stored in MicroZ (1 MicroZ = 16 px)

	edges := []int{EdgeSW, EdgeSE, EdgeNW, EdgeNE}
	for _, edge := range edges {
		neighbor := w.getNeighborTile(x, y, edge)
		if neighbor == nil || neighbor.waterLevel == t.waterLevel {
			continue
		}

		// Neighbor terrain corner heights in MicroZ.
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

		// Only draw where our water is above the neighbor terrain.
		if waterMicroZ <= neigh0 && waterMicroZ <= neigh1 {
			continue
		}
		minH := neigh0
		if neigh1 < minH {
			minH = neigh1
		}

		factor := checkerboard + kEdgeFactorOffset[edge]
		maskBase := kCliffEdgeMaskBase[edge]
		for h := minH; h < waterMicroZ; h++ {
			spriteID := int(cliffEdgeImageBase) + int(factor) + int(h&0xF)
			var slopeVariant int
			if h == waterMicroZ-1 {
				slopeVariant = 1 // topmost water cliff section
			}
			maskID := maskBase + slopeVariant
			img := w.renderer.GetMaskedSprite(spriteID, maskID)
			if img == nil {
				continue
			}
			_, _, xOff, yOff, ok := w.renderer.GetSpriteInfo(spriteID)
			if !ok {
				continue
			}
			edgeOffsetX, edgeOffsetY := w.getCliffEdgeOffset(edge, h)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(
				drawX+float64(xOff+edgeOffsetX)*scale,
				drawYFlat+float64(yOff+edgeOffsetY)*scale)
			w.highlightOps = append(w.highlightOps, worldHighlightOp{img: img, opts: op})
		}
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
	drawX := math.Round((vpX-w.camX)*scale)
	drawY := math.Round((vpY-w.camY)*scale)

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
	drawX := math.Round((vpX-w.camX)*scale)
	drawY := math.Round((vpY-w.camY)*scale)

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
//   Layer 0 (pass 1, worldBuf): +35 blend sprite with water colour forced, at its
//     native xOff=-32, yOff=0. The +35 sprite is the game-canonical water surface shape
//     (a bottom-half-of-diamond, 64×31 pixels). Adjacent water tiles drawn earlier in
//     depth order cover the "top 15 rows" of each tile's screen area, so tessellation
//     is seamless across flat ocean. Water cliff edges cover exposed terrain at boundaries.
//   Layer 1 (pass 4, screen): +30 ripple detail with real teal palette colours.
//   Layer 2 (pass 4, screen, zoom 0 only): wave animation frame (+60+frame).
//
// Fallback: when no WaterObj is loaded, uses getWaterImage() hand-drawn diamond.
//
// OpenLoco reference: PaintSurface.cpp:1720-1759 (paintSurfaceWater)
//   sprite+35 .withBlend(ExtColour::water) → dst-based palette remap
//   sprite+30 attached → ripple detail on top
//   sprite+60+frame attached → wave animation at zoom 0
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
	// OpenLoco PaintSurface.cpp: sloped water shape only when waterHeight <= baseHeight+kMicroZStep.
	// For deeply submerged tiles use flat shape (0).
	shape := uint8(0) // always flat for now; sloped shapes cause triangle artifacts
	blendID := int(waterObj.GetWaterSpriteIndex(shape, true))
	flatID := int(waterObj.GetWaterSpriteIndex(shape, false))

	// ── layer 0: +35 sprite with water colour forced (correct tessellating diamond) ──
	// The +35 mask (from decodeRLEPresence) marks exactly which pixels are the water
	// surface. We force all those pixels to the water colour for a clean solid blue.
	// Drawn at blendID's native xOff=-32, yOff=0 (the water surface anchor).
	if w.waterBaseCache == nil {
		w.waterBaseCache = make(map[int]*ebiten.Image)
	}
	// For non-zero shapes the blend sprite may not decode; fall back to shape 0 mask.
	maskID := blendID
	maskImg := w.renderer.GetSpriteMask(maskID)
	if maskImg == nil && shape != 0 {
		maskID = int(waterObj.GetWaterSpriteIndex(0, true))
		maskImg = w.renderer.GetSpriteMask(maskID)
	}
	if maskImg != nil {
		wc := w.waterColourFromPalette()
		waterBaseImg := buildColoredImage(w.waterBaseCache, int(shape), maskImg, wc)
		if waterBaseImg != nil {
			_, _, xOff, yOff, ok := w.renderer.GetSpriteInfo(blendID)
			if !ok {
				xOff, yOff = -32, 0
			}
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(
				drawX+float64(xOff)*scale,
				drawY+float64(yOff)*scale-waterHeightDiff*scale)
			target.DrawImage(waterBaseImg, op)
		}
	}

	// queueSprite draws a sprite with its own palette colours to the highlight list.
	queueSprite := func(cache map[int]*ebiten.Image, cacheKey, spriteID int) {
		src := w.renderer.GetSprite(spriteID)
		if src == nil {
			return
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
		w.highlightOps = append(w.highlightOps, worldHighlightOp{img: src, opts: op})
		_ = cache
		_ = cacheKey
	}

	// ── layer 1 (highlight pass): +30 ripple detail with real teal colours ──
	if w.waterHighlightCache == nil {
		w.waterHighlightCache = make(map[int]*ebiten.Image)
	}
	queueSprite(w.waterHighlightCache, int(shape), flatID)

	// ── layer 2 (highlight pass): wave animation frames at zoom 0 ──
	// OpenLoco reference: PaintSurface.cpp WaveManager, 16 frames × 4 ticks each
	if w.zoom == 0 {
		const waveFrames = 16
		waveFrame := (w.animTick/4 + tileX + tileY) & (waveFrames - 1)
		waveID := int(waterObj.ImageOffset) + 60 + waveFrame
		if w.waterWaveCache == nil {
			w.waterWaveCache = make(map[int]*ebiten.Image)
		}
		queueSprite(w.waterWaveCache, waveID, waveID)
	}
}

// buildColoredImage returns a version of srcImg where every non-transparent pixel
// is replaced with c.  Results are cached by cacheKey in the supplied map.
// alpha controls the output alpha (255 = fully opaque).
func buildColoredImage(cache map[int]*ebiten.Image, cacheKey int, srcImg *ebiten.Image, c color.RGBA) *ebiten.Image {
	if img, ok := cache[cacheKey]; ok {
		return img
	}
	bw, bh := srcImg.Bounds().Dx(), srcImg.Bounds().Dy()
	pixels := make([]byte, bw*bh*4)
	srcImg.ReadPixels(pixels)
	for i := 0; i < len(pixels); i += 4 {
		if pixels[i+3] > 0 {
			pixels[i+0] = c.R
			pixels[i+1] = c.G
			pixels[i+2] = c.B
			pixels[i+3] = c.A
		}
	}
	result := ebiten.NewImage(bw, bh)
	result.WritePixels(pixels)
	cache[cacheKey] = result
	return result
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

	// ── Pass 1: render all opaque sprites to worldBuf ──
	// Camera coords are in world-space (unscaled). They are set by
	// PanCamera (gameplay) or SetCamera (title sequence).

	// Draw tiles in depth order (back to front)
	for depth := 0; depth < w.width+w.height; depth++ {
		for y := 0; y < w.height; y++ {
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

					// Zoom-0 flat terrain sprite: skip all zoom3/zoom2/zoom1 images
					// OpenLoco reference: src/OpenLoco/src/Objects/LandObject.cpp getTerrainImage()
					spriteIdx := land.GetFlatTerrainSpriteIndex() + displaySlope

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
							w.paintTrees(w.worldBuf, &t, drawX, drawY, scale)
							w.paintBuildings(w.worldBuf, &t, drawX, drawY, scale)
							w.paintWalls(w.worldBuf, &t, drawX, drawY, scale)

							continue
						}
					}
				}
			}

			// Fall back to colored diamond
			img := w.getFallbackImage(t.tileType)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(drawX, drawY)
			w.worldBuf.DrawImage(img, op)
		}
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

		// Bounds check: tile coordinates from the save must be on the loaded map.
		tx := int(ent.TileX)
		ty := int(ent.TileY)
		if tx < 0 || tx >= w.width || ty < 0 || ty >= w.height {
			continue
		}

		// Convert tile → viewport (same formula as tileToScreen).
		// For sub-tile precision we shift by the sub-tile offset within the tile:
		//   PosX = tileX*32 + subtileX,  so subtileX = PosX - tileX*32
		subtileX := float64(ent.PosX) - float64(ent.TileX)*32.0
		subtileY := float64(ent.PosY) - float64(ent.TileY)*32.0
		fracX := float64(tx) + subtileX/32.0
		fracY := float64(ty) + subtileY/32.0

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
