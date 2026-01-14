package main

import (
	"image/color"
	"log"
	"math"
	"strconv"

	"github.com/LaPingvino/goloco/pkg/game"
	"github.com/LaPingvino/goloco/pkg/graphics"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 1024
	screenHeight = 768
)

// Game holds game state
type Game struct {
	// Camera
	cameraX, cameraY float64
	zoom             float64
	frame            int

	// Core systems
	gameState *game.GameState
	mainWindow *ui.Window

	// Graphics
	renderTarget *graphics.RenderTarget
}

func NewGame() *Game {
	g := &Game{
		zoom:      2.0,
		gameState: game.NewGameState(64, 64),
	}

	// Initialize window manager
	ui.InitWindowManager()

	// Create main viewport window
	g.mainWindow = createMainWindow()

	// Create test UI windows
	toolbar := ui.CreateToolbarWindow()
	_ = toolbar // Toolbar is added to window manager automatically

	testDialog := ui.CreateTestDialog()
	_ = testDialog // Dialog is added to window manager automatically

	return g
}

func (g *Game) Update() error {
	// Input handling
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
	// Update mouse/keyboard input
	ui.UpdateInput()

		return ebiten.Termination
	}

	// Camera movement
	moveSpeed := 8.0 / g.zoom
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		g.cameraX -= moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		g.cameraX += moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		g.cameraY -= moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		g.cameraY += moveSpeed
	}

	// Zoom
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		g.zoom *= 1.02
	}
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		g.zoom /= 1.02
	}
	if g.zoom < 0.25 {
		g.zoom = 0.25
	}
	if g.zoom > 4 {
		g.zoom = 4
	}

	// Update game state
	g.gameState.Update()

	// Update UI windows
	ui.UpdateWindows()

	g.frame++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Clear screen
	screen.Fill(color.RGBA{0x80, 0xa0, 0xff, 0xff})

	// Create drawing context
	dc := graphics.NewDrawingContext(screen)

	// Draw world
	g.drawWorld(screen)

	// Draw UI windows
	ui.RenderWindows(dc, ui.Rect{X: 0, Y: 0, Width: screenWidth, Height: screenHeight})

	// Draw HUD overlay
	ebitenutil.DebugPrintAt(screen, "goloco - OpenLoco in Go", 8, 8)

	// Get mouse position
	mx, my := ui.GetMousePosition()
	mouseState := "up"
	if ui.IsMousePressed() {
		mouseState = "down"
	}

	ebitenutil.DebugPrintAt(screen,
		"WASD / arrows: pan  Q/E: zoom  Esc: quit\n"+
		"Zoom: "+floatToStr(g.zoom)+"  Cam: ("+floatToStr(g.cameraX)+","+floatToStr(g.cameraY)+")\n"+
		"Date: "+g.gameState.GameDate.Format("Jan 02, 1950")+"  Money: £"+strconv.FormatInt(g.gameState.PlayerMoney, 10)+"\n"+
		"Mouse: ("+strconv.Itoa(mx)+","+strconv.Itoa(my)+") "+mouseState+"\n"+
		"Click windows to drag | Click X/OK/Cancel to close | Click checkboxes to toggle",
		8, 24)

	// Draw mouse cursor crosshair
	ebitenutil.DrawLine(screen, float64(mx)-5, float64(my), float64(mx)+5, float64(my), color.RGBA{255, 255, 0, 255})
	ebitenutil.DrawLine(screen, float64(mx), float64(my)-5, float64(mx), float64(my)+5, color.RGBA{255, 255, 0, 255})
}

func (g *Game) drawWorld(screen *ebiten.Image) {
	// Draw a simple tilemap
	tileSize := 32.0
	s := g.zoom

	// Compute viewport bounds in world coords
	left := g.cameraX - float64(screenWidth)/(2*s)
	right := g.cameraX + float64(screenWidth)/(2*s)
	top := g.cameraY - float64(screenHeight)/(2*s)
	bottom := g.cameraY + float64(screenHeight)/(2*s)

	minTileX := int(math.Floor(left/tileSize)) - 1
	maxTileX := int(math.Ceil(right/tileSize)) + 1
	minTileY := int(math.Floor(top/tileSize)) - 1
	maxTileY := int(math.Ceil(bottom/tileSize)) + 1

	// Clamp to map bounds
	if minTileX < 0 {
		minTileX = 0
	}
	if minTileY < 0 {
		minTileY = 0
	}
	if maxTileX >= g.gameState.MapWidth {
		maxTileX = g.gameState.MapWidth - 1
	}
	if maxTileY >= g.gameState.MapHeight {
		maxTileY = g.gameState.MapHeight - 1
	}

	// Draw tiles
	for ty := minTileY; ty <= maxTileY; ty++ {
		for tx := minTileX; tx <= maxTileX; tx++ {
			// Get tile from game state
			tile := g.gameState.MapTiles[ty][tx]

			// World pos
			x := float64(tx)*tileSize + tileSize/2
			y := float64(ty)*tileSize + tileSize/2

			// Screen pos
			sx := (x-g.cameraX)*s + float64(screenWidth)/2
			sy := (y-g.cameraY)*s + float64(screenHeight)/2

			// Tile color based on type
			var c color.RGBA
			switch tile.Type {
			case game.TileGrass:
				c = color.RGBA{60, 140, 60, 255}
			case game.TileDirt:
				c = color.RGBA{140, 100, 60, 255}
			case game.TileWater:
				c = color.RGBA{40, 80, 160, 255}
			case game.TileRoad:
				c = color.RGBA{80, 80, 80, 255}
			case game.TileRail:
				c = color.RGBA{100, 100, 100, 255}
			default:
				c = color.RGBA{100, 100, 100, 255}
			}

			r := tileSize * s * 0.5
			ebitenutil.DrawRect(screen, sx-r, sy-r, r*2, r*2, c)

			// Draw grid lines
			ebitenutil.DrawLine(screen, sx-r, sy-r, sx+r, sy-r, color.RGBA{0, 0, 0, 0x40})
			ebitenutil.DrawLine(screen, sx-r, sy-r, sx-r, sy+r, color.RGBA{0, 0, 0, 0x40})
		}
	}

	// Draw vehicles
	for _, vehicle := range g.gameState.Vehicles {
		vx := vehicle.X
		vy := vehicle.Y

		sx := (vx-g.cameraX)*s + float64(screenWidth)/2
		sy := (vy-g.cameraY)*s + float64(screenHeight)/2

		ebitenutil.DrawRect(screen, sx-4, sy-4, 8, 8, color.RGBA{255, 100, 100, 255})
	}
}

func floatToStr(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// createMainWindow creates the main viewport window
func createMainWindow() *ui.Window {
	w := ui.CreateWindow(ui.WindowTypeMain, 0, 0, screenWidth, screenHeight, ui.WindowFlagNoBackground)

	// This window is transparent and just acts as the main container
	// We'll draw the world behind it

	return w
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("OpenLoco (Go)")

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
