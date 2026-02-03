package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx"
	"github.com/LaPingvino/goloco/pkg/graphics"
)

// Title menu constants
const (
	TitleBtnSize     = 74 // Button width/height in pixels
	TitleBtnCount    = 4  // Number of main buttons
	TitleMenuWidth   = TitleBtnSize * TitleBtnCount
	TitleMenuHeight  = TitleBtnSize
	TitleMenuPadding = 25 // Distance from bottom of screen
)

// Title menu widget indices
const (
	TitleWidxScenarioList = iota
	TitleWidxLoadGame
	TitleWidxTutorial
	TitleWidxScenarioEditor
)

// TitleMenuState holds animation state for the title menu
type TitleMenuState struct {
	GlobeFrame    int // Current globe animation frame (0-31)
	HoveredButton int // Currently hovered button (-1 for none)
}

var titleMenuState = TitleMenuState{
	HoveredButton: -1,
}

// CreateTitleMenu creates the title menu window positioned at the bottom center
func CreateTitleMenu(screenWidth, screenHeight int) *Window {
	// Position: centered horizontally, padding from bottom
	x := int16((screenWidth - TitleMenuWidth) / 2)
	y := int16(screenHeight - TitleMenuHeight - TitleMenuPadding)

	// Create frameless, transparent window
	w := CreateWindow(
		WindowTypeTitleMenu,
		x, y,
		TitleMenuWidth, TitleMenuHeight,
		WindowFlagStickToFront|WindowFlagNoBackground,
	)

	// Set window colors (used for button rendering)
	w.Colours[0] = gfx.AdvancedColor{Color: gfx.Color(140)} // Button background
	w.Colours[1] = gfx.AdvancedColor{Color: gfx.Color(180)} // Button highlight
	w.Colours[2] = gfx.AdvancedColor{Color: gfx.Color(100)} // Button shadow

	// Create the 4 main buttons
	w.Widgets = []Widget{
		// Scenario List button
		{
			Type:    WidgetTypeImageButton,
			Index:   TitleWidxScenarioList,
			X:       0,
			Y:       0,
			Width:   TitleBtnSize,
			Height:  TitleBtnSize,
			Colour:  w.Colours[0],
			Text:    "Scenarios",
			Tooltip: "Browse and start scenarios",
			Visible: true,
			Flags:   WidgetFlagEnabled,
		},
		// Load Game button
		{
			Type:    WidgetTypeImageButton,
			Index:   TitleWidxLoadGame,
			X:       TitleBtnSize,
			Y:       0,
			Width:   TitleBtnSize,
			Height:  TitleBtnSize,
			Colour:  w.Colours[0],
			Text:    "Load",
			Tooltip: "Load a saved game",
			Visible: true,
			Flags:   WidgetFlagEnabled,
		},
		// Tutorial button
		{
			Type:    WidgetTypeImageButton,
			Index:   TitleWidxTutorial,
			X:       TitleBtnSize * 2,
			Y:       0,
			Width:   TitleBtnSize,
			Height:  TitleBtnSize,
			Colour:  w.Colours[0],
			Text:    "Tutorial",
			Tooltip: "Learn how to play",
			Visible: true,
			Flags:   WidgetFlagEnabled,
		},
		// Scenario Editor button
		{
			Type:    WidgetTypeImageButton,
			Index:   TitleWidxScenarioEditor,
			X:       TitleBtnSize * 3,
			Y:       0,
			Width:   TitleBtnSize,
			Height:  TitleBtnSize,
			Colour:  w.Colours[0],
			Text:    "Editor",
			Tooltip: "Create and edit scenarios",
			Visible: true,
			Flags:   WidgetFlagEnabled,
		},
	}

	// Set custom draw function
	w.eventHandlers = &WindowEventHandlers{
		Draw: drawTitleMenu,
	}

	return w
}

// UpdateTitleMenu updates the globe animation and hover state
func UpdateTitleMenu(w *Window) {
	if w == nil || w.Type != WindowTypeTitleMenu {
		return
	}

	// Advance globe animation (32 frames, update every 2 ticks)
	titleMenuState.GlobeFrame = (titleMenuState.GlobeFrame + 1) % 64

	// Update hover state based on mouse position
	mx, my := GetMousePosition()
	relX := int16(mx) - w.X
	relY := int16(my) - w.Y

	titleMenuState.HoveredButton = -1
	if relX >= 0 && relX < TitleMenuWidth && relY >= 0 && relY < TitleMenuHeight {
		titleMenuState.HoveredButton = int(relX / TitleBtnSize)
	}
}

// drawTitleMenu renders the title menu with globe animations
func drawTitleMenu(w *Window, dc *graphics.DrawingContext) {
	if w == nil || dc == nil {
		return
	}

	// Draw each button
	for i := range w.Widgets {
		widget := &w.Widgets[i]
		if !widget.Visible {
			continue
		}

		// Calculate absolute position
		btnX := w.X + widget.X
		btnY := w.Y + widget.Y

		// Determine if this button is hovered
		isHovered := (i == titleMenuState.HoveredButton)

		// Draw button background (3D beveled look)
		drawTitleButton(dc, btnX, btnY, TitleBtnSize, TitleBtnSize, isHovered, i)

		// Draw button label (centered at bottom)
		labelX := btnX + TitleBtnSize/2
		labelY := btnY + TitleBtnSize - 12
		_ = dc.DrawStringCentered(labelX, labelY, widget.Text, 0) // Black text
	}
}

// drawTitleButton draws a single title menu button with 3D effect
func drawTitleButton(dc *graphics.DrawingContext, x, y, w, h int16, hovered bool, buttonIndex int) {
	// Button colors
	var bgColor, lightColor, darkColor uint8
	if hovered {
		bgColor = 160 // Lighter when hovered
		lightColor = 200
		darkColor = 120
	} else {
		bgColor = 140
		lightColor = 180
		darkColor = 100
	}

	// Draw main background
	_ = dc.FillRect(x, y, w, h, bgColor)

	// Draw 3D border (light on top/left, dark on bottom/right)
	// Top edge (light)
	_ = dc.FillRect(x, y, w, 2, lightColor)
	// Left edge (light)
	_ = dc.FillRect(x, y, 2, h, lightColor)
	// Bottom edge (dark)
	_ = dc.FillRect(x, y+h-2, w, 2, darkColor)
	// Right edge (dark)
	_ = dc.FillRect(x+w-2, y, 2, h, darkColor)

	// Draw a simple globe placeholder (circle) in the center
	// In the future, this will be replaced with actual G1 sprite rendering
	globeX := x + w/2
	globeY := y + h/2 - 8 // Shift up to leave room for label

	// Draw a simple filled circle approximation (square for now)
	globeSize := int16(40)
	globeColor := uint8(60 + (titleMenuState.GlobeFrame/2)%20) // Subtle animation
	if buttonIndex == TitleWidxScenarioEditor {
		// Editor button has construction globe (different color)
		globeColor = uint8(80 + (titleMenuState.GlobeFrame/2)%20)
	}
	_ = dc.FillRect(globeX-globeSize/2, globeY-globeSize/2, globeSize, globeSize, globeColor)
}

// GetTitleMenuClickedButton returns the button index that was clicked, or -1 if none
func GetTitleMenuClickedButton(w *Window) int {
	if w == nil || w.Type != WindowTypeTitleMenu {
		return -1
	}

	if !IsMousePressed() {
		return -1
	}

	mx, my := GetMousePosition()
	relX := int16(mx) - w.X
	relY := int16(my) - w.Y

	if relX >= 0 && relX < TitleMenuWidth && relY >= 0 && relY < TitleMenuHeight {
		return int(relX / TitleBtnSize)
	}

	return -1
}
