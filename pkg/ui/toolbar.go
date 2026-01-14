package ui

import (
	"image/color"

	"github.com/LaPingvino/goloco/pkg/render"
	"github.com/LaPingvino/goloco/pkg/sprites"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// ToolbarButton represents a button in the main toolbar
type ToolbarButton struct {
	SpriteID int
	Tooltip  string
	X        int
	Y        int
	Width    int
	Height   int
	Pressed  bool
	Hovered  bool
}

// Toolbar represents the main game toolbar
type Toolbar struct {
	X       int
	Y       int
	Width   int
	Height  int
	Buttons []ToolbarButton
	Visible bool
}

// NewToolbar creates a new toolbar with standard Locomotion buttons
func NewToolbar(screenWidth int) *Toolbar {
	t := &Toolbar{
		X:       0,
		Y:       0,
		Width:   screenWidth,
		Height:  32,
		Visible: true,
	}

	// Define toolbar buttons with their sprite IDs
	// Using construction/UI icons from G1.DAT
	buttonDefs := []struct {
		spriteID int
		tooltip  string
	}{
		{sprites.ToolbarZoomIn, "Zoom In"},
		{sprites.ToolbarZoomOut, "Zoom Out"},
		{sprites.ToolbarRotate, "Rotate View"},
		{sprites.ToolbarMap, "Map"},
		{0, ""}, // separator
		{sprites.ToolbarTracks, "Build Tracks"},
		{sprites.ToolbarRoads, "Build Roads"},
		{sprites.ToolbarAirport, "Build Airports"},
		{sprites.ToolbarPorts, "Build Ports"},
		{0, ""}, // separator
		{sprites.ToolbarVehicles, "Vehicles"},
		{sprites.ToolbarStations, "Stations"},
		{sprites.ToolbarTowns, "Towns"},
		{sprites.ToolbarIndustries, "Industries"},
		{0, ""}, // separator
		{sprites.ToolbarLandscaping, "Landscaping"},
		{sprites.ToolbarTrees, "Trees"},
		{sprites.ToolbarWater, "Water"},
		{sprites.ToolbarRemove, "Remove"},
		{0, ""}, // separator
		{sprites.ToolbarCompanies, "Companies"},
		{sprites.ToolbarFinances, "Finances"},
		{sprites.ToolbarNews, "Messages"},
		{sprites.ToolbarCamera, "Screenshot"},
	}

	x := 4
	for _, def := range buttonDefs {
		if def.spriteID == 0 {
			// Separator - just add spacing
			x += 10
			continue
		}

		t.Buttons = append(t.Buttons, ToolbarButton{
			SpriteID: def.spriteID,
			Tooltip:  def.tooltip,
			X:        x,
			Y:        3,
			Width:    26,
			Height:   26,
		})
		x += 28
	}

	return t
}

// Draw renders the toolbar
func (t *Toolbar) Draw(screen *ebiten.Image, renderer *render.Renderer) {
	if !t.Visible {
		return
	}

	// Draw toolbar background with slight gradient effect
	for y := t.Y; y < t.Y+t.Height; y++ {
		shade := uint8(70 - (y-t.Y)/2)
		bgColor := color.RGBA{shade, shade, shade, 245}
		for x := t.X; x < t.X+t.Width; x++ {
			screen.Set(x, y, bgColor)
		}
	}

	// Draw border
	borderColor := color.RGBA{40, 40, 40, 255}
	for x := t.X; x < t.X+t.Width; x++ {
		screen.Set(x, t.Y+t.Height-1, borderColor)
	}

	// Draw buttons
	var hoveredTooltip string
	var tooltipX int
	for i := range t.Buttons {
		t.drawButton(screen, renderer, &t.Buttons[i])
		if t.Buttons[i].Hovered && t.Buttons[i].Tooltip != "" {
			hoveredTooltip = t.Buttons[i].Tooltip
			tooltipX = t.Buttons[i].X
		}
	}

	// Draw tooltip if hovering
	if hoveredTooltip != "" {
		tooltipY := t.Y + t.Height + 2
		ebitenutil.DebugPrintAt(screen, hoveredTooltip, tooltipX, tooltipY)
	}
}

func (t *Toolbar) drawButton(screen *ebiten.Image, renderer *render.Renderer, btn *ToolbarButton) {
	// Button background
	var bgColor color.RGBA
	if btn.Pressed {
		bgColor = color.RGBA{80, 80, 80, 255}
	} else if btn.Hovered {
		bgColor = color.RGBA{100, 100, 100, 255}
	} else {
		bgColor = color.RGBA{70, 70, 70, 255}
	}

	// Draw button background
	for y := btn.Y; y < btn.Y+btn.Height; y++ {
		for x := btn.X; x < btn.X+btn.Width; x++ {
			screen.Set(x, y, bgColor)
		}
	}

	// Draw sprite if available
	if renderer != nil && renderer.G1 != nil {
		if img := renderer.GetSprite(btn.SpriteID); img != nil {
			w, h, xOff, yOff, ok := renderer.GetSpriteInfo(btn.SpriteID)
			if ok {
				// Center sprite in button
				drawX := float64(btn.X) + float64(btn.Width-int(w))/2 + float64(xOff)
				drawY := float64(btn.Y) + float64(btn.Height-int(h))/2 + float64(yOff)

				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(drawX, drawY)
				screen.DrawImage(img, op)
			}
		}
	}

	// Draw button border
	borderLight := color.RGBA{120, 120, 120, 255}
	borderDark := color.RGBA{40, 40, 40, 255}

	// Top and left (light)
	for x := btn.X; x < btn.X+btn.Width; x++ {
		screen.Set(x, btn.Y, borderLight)
	}
	for y := btn.Y; y < btn.Y+btn.Height; y++ {
		screen.Set(btn.X, y, borderLight)
	}
	// Bottom and right (dark)
	for x := btn.X; x < btn.X+btn.Width; x++ {
		screen.Set(x, btn.Y+btn.Height-1, borderDark)
	}
	for y := btn.Y; y < btn.Y+btn.Height; y++ {
		screen.Set(btn.X+btn.Width-1, y, borderDark)
	}
}

// HandleClick checks if a click hit any button and returns the index, or -1
func (t *Toolbar) HandleClick(x, y int) int {
	if !t.Visible {
		return -1
	}
	if y < t.Y || y >= t.Y+t.Height {
		return -1
	}

	for i := range t.Buttons {
		btn := &t.Buttons[i]
		if x >= btn.X && x < btn.X+btn.Width && y >= btn.Y && y < btn.Y+btn.Height {
			return i
		}
	}
	return -1
}

// UpdateHover updates hover state based on mouse position
func (t *Toolbar) UpdateHover(x, y int) {
	if !t.Visible {
		return
	}

	for i := range t.Buttons {
		btn := &t.Buttons[i]
		btn.Hovered = x >= btn.X && x < btn.X+btn.Width && y >= btn.Y && y < btn.Y+btn.Height
	}
}
