package ui

import (
	"image/color"

	"github.com/LaPingvino/goloco/pkg/graphics"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/render"
	"github.com/hajimehoshi/ebiten/v2"
)

// buttonImageIDs maps button tooltips to InterfaceSkin image IDs
var buttonImageIDs = map[string][2]uint32{
	// [normal, hover] pairs
	"Load/Save":      {objects.ISToolbarLoadsave, objects.ISToolbarLoadsaveHover},
	"Audio":          {objects.ISToolbarAudioActive, objects.ISToolbarAudioActiveHover},
	"Zoom":           {objects.ISToolbarZoom, objects.ISToolbarZoomHover},
	"Rotate":         {objects.ISToolbarRotate, objects.ISToolbarRotateHover},
	"View":           {objects.ISToolbarView, objects.ISToolbarViewHover},
	"Terraform":      {objects.ISToolbarTerraform, objects.ISToolbarTerraformHover},
	"Railroad":       {objects.ISToolbarEmptyOpaque, objects.ISToolbarEmptyOpaqueHover},
	"Road":           {objects.ISToolbarEmptyOpaque, objects.ISToolbarEmptyOpaqueHover},
	"Port/Airport":   {objects.ISToolbarAirports, objects.ISToolbarAirportsHover},
	"Build Vehicles": {objects.ISToolbarBuildVehicleTrain, objects.ISToolbarBuildVehicleTrainHover},
	"Vehicles":       {objects.ISToolbarEmptyOpaque, objects.ISToolbarEmptyOpaqueHover},
	"Stations":       {objects.ISToolbarStations, objects.ISToolbarStationsHover},
	"Towns":          {objects.ISToolbarTowns, objects.ISToolbarTownsHover},
}

// ToolbarButton represents a button in the main toolbar
type ToolbarButton struct {
	Tooltip    string
	X          int
	Y          int
	Width      int
	Height     int
	Pressed    bool
	Hovered    bool
	ColorGroup int // 0=primary, 1=secondary, 2=tertiary, 3=quaternary
	Hidden     bool
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

// NewToolbar creates a new toolbar matching OpenLoco's layout
//
// OpenLoco reference: src/OpenLoco/src/Ui/Windows/ToolbarTop.cpp
//
//	Game::open() creates the toolbar at (0, 0) with full screen width and height 28.
//	All buttons are 30 pixels wide × 28 pixels tall.
//
// Button layout (13 buttons organized in 4 color groups):
//
//	PRIMARY (leftmost): Load/Save, Audio, (Cheats placeholder)
//	SECONDARY: Zoom, Rotate, View
//	TERTIARY (middle): Terraform, Railroad, Road, Port/Airport, Build Vehicles
//	QUATERNARY (right-aligned): Vehicles, Stations, Towns/Industry
func NewToolbar(screenWidth int) *Toolbar {
	t := &Toolbar{
		X:       0,
		Y:       0,
		Width:   screenWidth,
		Height:  28, // Match OpenLoco exactly
		Visible: true,
	}

	// Button definitions matching OpenLoco's exact layout and order.
	// Fixed positions for left buttons, right-aligned buttons positioned later.
	buttonDefs := []struct {
		tooltip    string
		x          int
		colorGroup int
		hidden     bool
	}{
		// PRIMARY COLOR GROUP (WindowColour::primary)
		{"Load/Save", 0, 0, false},
		{"Audio", 30, 0, false},
		{"", 60, 0, true}, // Reserved slot (cheats in game mode)

		// SECONDARY COLOR GROUP (WindowColour::secondary)
		{"Zoom", 104, 1, false},
		{"Rotate", 134, 1, false},
		{"View", 164, 1, false},

		// TERTIARY COLOR GROUP (WindowColour::tertiary)
		// These will be right-aligned; X positions are placeholders
		{"Terraform", 267, 2, false},
		{"Railroad", 357, 2, false},
		{"Road", 387, 2, false},
		{"Port/Airport", 417, 2, false},
		{"Build Vehicles", 447, 2, false},

		// QUATERNARY COLOR GROUP (WindowColour::quaternary)
		// Right-aligned; X positions calculated dynamically
		{"Vehicles", 490, 3, false},
		{"Stations", 520, 3, false},
		{"Towns", 550, 3, false},
	}

	for _, def := range buttonDefs {
		t.Buttons = append(t.Buttons, ToolbarButton{
			Tooltip:    def.tooltip,
			X:          def.x,
			Y:          0,  // All buttons start at Y=0
			Width:      30, // Match OpenLoco: 30 pixels wide
			Height:     28, // Match OpenLoco: 28 pixels tall
			ColorGroup: def.colorGroup,
			Hidden:     def.hidden,
		})
	}

	// Right-align the tertiary and quaternary buttons
	t.RightAlignButtons(screenWidth)

	return t
}

// RightAlignButtons repositions tertiary and quaternary buttons to the right side
// of the toolbar, matching OpenLoco's dynamic positioning logic.
//
// OpenLoco reference: src/OpenLoco/src/Ui/Windows/ToolbarTop.cpp
//
//	prepareDraw() lines 977-997
// RightAlignButtons repositions the right-aligned toolbar buttons for the
// given screen width. Called on creation and again whenever the window resizes.
func (t *Toolbar) RightAlignButtons(screenWidth int) {
	// Start from the right edge (match OpenLoco: max(640, width) - 1)
	x := screenWidth - 1

	// Right-align quaternary buttons (Towns, Stations, Vehicles)
	// Work backwards so they appear right-to-left in reverse order
	for i := len(t.Buttons) - 1; i >= 0; i-- {
		if t.Buttons[i].ColorGroup == 3 && !t.Buttons[i].Hidden {
			x -= t.Buttons[i].Width
			t.Buttons[i].X = x
		}
	}

	// 11 pixel gap before build vehicles (OpenLoco: x -= 11)
	x -= 11

	// Right-align tertiary buttons (Build Vehicles, Port, Road, Railroad, Terraform)
	// Work backwards through tertiary buttons
	for i := len(t.Buttons) - 1; i >= 0; i-- {
		if t.Buttons[i].ColorGroup == 2 && !t.Buttons[i].Hidden {
			x -= t.Buttons[i].Width
			t.Buttons[i].X = x
		}
	}
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
		if !t.Buttons[i].Hidden {
			t.drawButton(screen, &t.Buttons[i], renderer)
			if t.Buttons[i].Hovered && t.Buttons[i].Tooltip != "" {
				hoveredTooltip = t.Buttons[i].Tooltip
				tooltipX = t.Buttons[i].X
			}
		}
	}

	// Draw tooltip if hovering
	if hoveredTooltip != "" {
		tooltipY := t.Y + t.Height + 12
		DrawText(screen, hoveredTooltip, tooltipX, tooltipY, color.White)
	}
}

func (t *Toolbar) drawButton(screen *ebiten.Image, btn *ToolbarButton, renderer *render.Renderer) {
	// Create a drawing context for palette-aware fills
	dc := graphics.NewDrawingContext(screen)

	// Button background colors vary by color group
	// For now, use a simple scheme; real colors come from InterfaceSkinObject
	var normalTarget, hoverTarget, pressedTarget color.RGBA

	switch btn.ColorGroup {
	case 0: // PRIMARY
		normalTarget = color.RGBA{60, 60, 80, 255}
		hoverTarget = color.RGBA{80, 80, 110, 255}
		pressedTarget = color.RGBA{70, 70, 90, 255}
	case 1: // SECONDARY
		normalTarget = color.RGBA{70, 80, 70, 255}
		hoverTarget = color.RGBA{90, 110, 90, 255}
		pressedTarget = color.RGBA{80, 90, 80, 255}
	case 2: // TERTIARY
		normalTarget = color.RGBA{80, 70, 60, 255}
		hoverTarget = color.RGBA{110, 90, 80, 255}
		pressedTarget = color.RGBA{90, 80, 70, 255}
	case 3: // QUATERNARY
		normalTarget = color.RGBA{70, 70, 70, 255}
		hoverTarget = color.RGBA{100, 100, 100, 255}
		pressedTarget = color.RGBA{80, 80, 80, 255}
	default:
		normalTarget = color.RGBA{70, 70, 70, 255}
		hoverTarget = color.RGBA{100, 100, 100, 255}
		pressedTarget = color.RGBA{80, 80, 80, 255}
	}

	// Map targets to palette indices
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

	// Try to draw the InterfaceSkin sprite if available
	drewSprite := false
	if imageIDs, ok := buttonImageIDs[btn.Tooltip]; ok {
		imageID := imageIDs[0] // normal
		if btn.Hovered || btn.Pressed {
			imageID = imageIDs[1] // hover
		}

		sprite := renderer.GetInterfaceSkinSprite(imageID)
		if sprite != nil {
			// Draw sprite centered in button
			spriteW := sprite.Bounds().Dx()
			spriteH := sprite.Bounds().Dy()
			drawX := btn.X + (btn.Width-spriteW)/2
			drawY := btn.Y + (btn.Height-spriteH)/2

			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(float64(drawX), float64(drawY))
			screen.DrawImage(sprite, opts)
			drewSprite = true
		}
	}

	// Railroad / Road: draw the first available track/road preview image on top of the
	// empty-opaque background, mirroring OpenLoco ToolbarTop.cpp::draw().
	// OpenLoco reference: ToolbarTop.cpp:818-852 (railroad_menu draw)
	//   fg_image = recolour(obj->image + TrackObj::ImageIds::kUiPreviewImage0, companyColour)
	if !drewSprite || btn.Tooltip == "Railroad" || btn.Tooltip == "Road" {
		var previewSpriteID int = -1
		const defaultCompanyColour = 7
		if btn.Tooltip == "Railroad" && renderer.ObjMgr != nil {
			for _, tr := range renderer.ObjMgr.TrackObjects {
				if tr != nil {
					previewSpriteID = int(tr.Image) // kUiPreviewImage0 = offset 0
					break
				}
			}
		} else if btn.Tooltip == "Road" && renderer.ObjMgr != nil {
			for _, ro := range renderer.ObjMgr.RoadObjects {
				if ro != nil {
					previewSpriteID = int(ro.Image) // kUiPreviewImage0 = offset 0
					break
				}
			}
		}
		if previewSpriteID >= 0 {
			img := renderer.GetSpriteColoured(previewSpriteID, defaultCompanyColour)
			if img == nil {
				img = renderer.GetSprite(previewSpriteID)
			}
			if img != nil {
				spriteW := img.Bounds().Dx()
				spriteH := img.Bounds().Dy()
				drawX := btn.X + (btn.Width-spriteW)/2
				drawY := btn.Y + (btn.Height-spriteH)/2
				opts := &ebiten.DrawImageOptions{}
				opts.GeoM.Translate(float64(drawX), float64(drawY))
				screen.DrawImage(img, opts)
				drewSprite = true
			}
		}
	}

	// Fallback: abbreviated text label
	if !drewSprite && btn.Tooltip != "" {
		label := btn.Tooltip
		if len(label) > 4 {
			label = label[:4]
		}
		labelW, _ := MeasureText(label)
		labelX := btn.X + (btn.Width-labelW)/2
		labelY := btn.Y + btn.Height/2 + 4
		DrawText(screen, label, labelX, labelY, color.White)
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
		if btn.Hidden {
			continue
		}
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
		if btn.Hidden {
			btn.Hovered = false
			continue
		}
		btn.Hovered = x >= btn.X && x < btn.X+btn.Width && y >= btn.Y && y < btn.Y+btn.Height
	}
}
