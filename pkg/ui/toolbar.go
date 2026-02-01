package ui

import (
	"image/color"

	"github.com/LaPingvino/goloco/pkg/graphics"
	"github.com/LaPingvino/goloco/pkg/render"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// ToolbarButton represents a button in the main toolbar
type ToolbarButton struct {
	Tooltip string
	X       int
	Y       int
	Width   int
	Height  int
	Pressed bool
	Hovered bool
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

	// Define toolbar buttons.  Real icons come from InterfaceSkin DAT objects
	// which are not yet loaded; abbreviated text labels are used until then.
	buttonDefs := []string{
		"Zoom In", "Zoom Out", "Rotate View", "Map",
		"", // separator
		"Build Tracks", "Build Roads", "Build Airports", "Build Ports",
		"", // separator
		"Vehicles", "Stations", "Towns", "Industries",
		"", // separator
		"Landscaping", "Trees", "Water", "Remove",
		"", // separator
		"Companies", "Finances", "Messages", "Screenshot",
	}

	x := 4
	for _, tooltip := range buttonDefs {
		if tooltip == "" {
			x += 10 // separator
			continue
		}
		t.Buttons = append(t.Buttons, ToolbarButton{
			Tooltip: tooltip,
			X:       x,
			Y:       3,
			Width:   26,
			Height:  26,
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

	// Use a drawing context so palette indices can be used consistently.
	dc := graphics.NewDrawingContext(screen)

	// Background: prefer palette-aware fill; fallback to gradient when no palette.
	if graphics.IsGlobalPaletteLoaded() {
		target := color.RGBA{70, 70, 70, 255}
		idx := graphics.MatchPaletteIndex(target)
		_ = dc.FillRect(int16(t.X), int16(t.Y), int16(t.Width), int16(t.Height), idx)

		// Border line at bottom
		borderIdx := graphics.MatchPaletteIndex(color.RGBA{40, 40, 40, 255})
		_ = dc.FillRect(int16(t.X), int16(t.Y+t.Height-1), int16(t.Width), 1, borderIdx)
	} else {
		// Fallback: slight gradient effect for development without real palette
		for y := t.Y; y < t.Y+t.Height; y++ {
			shade := uint8(70 - (y-t.Y)/2)
			bgColor := color.RGBA{shade, shade, shade, 245}
			for x := t.X; x < t.X+t.Width; x++ {
				screen.Set(x, y, bgColor)
			}
		}
		// fallback bottom border
		borderColor := color.RGBA{40, 40, 40, 255}
		for x := t.X; x < t.X+t.Width; x++ {
			screen.Set(x, t.Y+t.Height-1, borderColor)
		}
	}

	// Draw buttons
	var hoveredTooltip string
	var tooltipX int
	for i := range t.Buttons {
		t.drawButton(screen, &t.Buttons[i])
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

func (t *Toolbar) drawButton(screen *ebiten.Image, btn *ToolbarButton) {
	// Create a drawing context for palette-aware fills
	dc := graphics.NewDrawingContext(screen)

	// Button background palette targets
	normalTarget := color.RGBA{70, 70, 70, 255}
	hoverTarget := color.RGBA{100, 100, 100, 255}
	pressedTarget := color.RGBA{80, 80, 80, 255}

	// Map targets to palette indices (or use defaults via matching)
	normalIdx := graphics.MatchPaletteIndex(normalTarget)
	hoverIdx := graphics.MatchPaletteIndex(hoverTarget)
	pressedIdx := graphics.MatchPaletteIndex(pressedTarget)

	var bgIdx uint8
	if btn.Pressed {
		bgIdx = pressedIdx
	} else if btn.Hovered {
		bgIdx = hoverIdx
	} else {
		bgIdx = normalIdx
	}

	// Fill the button background using palette index
	_ = dc.FillRect(int16(btn.X), int16(btn.Y), int16(btn.Width), int16(btn.Height), bgIdx)

	// Draw button label text (centered).
	// Real toolbar icons come from InterfaceSkin DAT objects which are not
	// yet loaded; use abbreviated text labels until then.
	if btn.Tooltip != "" {
		label := btn.Tooltip
		if len(label) > 4 {
			label = label[:4]
		}
		// Center the short label in the button
		labelX := btn.X + (btn.Width-len(label)*6)/2
		labelY := btn.Y + btn.Height/2 - 4
		ebitenutil.DebugPrintAt(screen, label, labelX, labelY)
	}

	// Draw button border using palette indices for light/dark edges
	borderLightIdx := graphics.MatchPaletteIndex(color.RGBA{120, 120, 120, 255})
	borderDarkIdx := graphics.MatchPaletteIndex(color.RGBA{40, 40, 40, 255})

	// Top and left (light)
	_ = dc.FillRect(int16(btn.X), int16(btn.Y), int16(btn.Width), 1, borderLightIdx)
	_ = dc.FillRect(int16(btn.X), int16(btn.Y), 1, int16(btn.Height), borderLightIdx)

	// Bottom and right (dark)
	_ = dc.FillRect(int16(btn.X), int16(btn.Y+btn.Height-1), int16(btn.Width), 1, borderDarkIdx)
	_ = dc.FillRect(int16(btn.X+btn.Width-1), int16(btn.Y), 1, int16(btn.Height), borderDarkIdx)
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
