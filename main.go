package main

import (
	"image/color"
	"log"
	"math"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 1024
	screenHeight = 768
)

// Game holds game state
type Game struct {
	cameraX, cameraY float64
	zoom             float64
	frame            int
}

func NewGame() *Game {
	return &Game{zoom: 1.0}
}

func (g *Game) Update() error {
	// Input
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
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
	g.frame++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0x80, 0xa0, 0xff, 0xff})

	// Draw a simple tilemap
	tileSize := 32.0
	// adjust for zoom
	s := g.zoom
	// compute viewport bounds in world coords
	left := g.cameraX - float64(screenWidth)/(2*s)
	right := g.cameraX + float64(screenWidth)/(2*s)
	top := g.cameraY - float64(screenHeight)/(2*s)
	bottom := g.cameraY + float64(screenHeight)/(2*s)

	minTileX := int(math.Floor(left/tileSize)) - 1
	maxTileX := int(math.Ceil(right/tileSize)) + 1
	minTileY := int(math.Floor(top/tileSize)) - 1
	maxTileY := int(math.Ceil(bottom/tileSize)) + 1

	for ty := minTileY; ty <= maxTileY; ty++ {
		for tx := minTileX; tx <= maxTileX; tx++ {
			// world pos
			x := float64(tx)*tileSize + tileSize/2
			y := float64(ty)*tileSize + tileSize/2
			// screen pos
			sx := (x-g.cameraX)*s + float64(screenWidth)/2
			sy := (y-g.cameraY)*s + float64(screenHeight)/2

			// tile color by simple pattern
			shade := uint8((tx+ty)&1)*40 + 140
			c := color.RGBA{shade, shade + 10, shade + 30, 0xff}

			r := tileSize * s * 0.5
			ebitenutil.DrawRect(screen, sx-r, sy-r, r*2, r*2, c)
			// draw grid lines
			ebitenutil.DrawLine(screen, sx-r, sy-r, sx+r, sy-r, color.RGBA{0, 0, 0, 0x40})
			ebitenutil.DrawLine(screen, sx-r, sy-r, sx-r, sy+r, color.RGBA{0, 0, 0, 0x40})
		}
	}

	// HUD
	ebitenutil.DebugPrintAt(screen, "goloco - minimal prototype", 8, 8)
	ebitenutil.DebugPrintAt(screen,
		"WASD / arrows: pan  Q/E: zoom  Esc: quit\nZoom: "+
			floatToStr(g.zoom)+"  Cam: ("+floatToStr(g.cameraX)+","+floatToStr(g.cameraY)+")",
		8, 24)
}

func floatToStr(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("goloco")
	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
