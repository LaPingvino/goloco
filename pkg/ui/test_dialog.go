package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx"
)

// CreateTestDialog creates a simple test dialog to verify UI rendering
func CreateTestDialog() *Window {
	w := CreateWindow(WindowTypeDialog, 200, 150, 400, 300, WindowFlagNone)

	// Set window colors
	w.Colours[0] = gfx.AdvancedColor{Color: gfx.Color(180), Translucent: false, Outline: false, Inset: false}

	// Create widgets
	w.Widgets = []Widget{
		// Frame (background)
		{
			Type:    WidgetTypeFrame,
			Index:   0,
			X:       0,
			Y:       0,
			Width:   400,
			Height:  300,
			Colour:  gfx.AdvancedColor{Color: gfx.Color(180)},
			Visible: true,
		},
		// Caption bar
		{
			Type:    WidgetTypeCaption,
			Index:   1,
			X:       0,
			Y:       0,
			Width:   400,
			Height:  20,
			Colour:  gfx.AdvancedColor{Color: gfx.Color(150)},
			Text:    "Test Dialog",
			Visible: true,
		},
		// Close button (X)
		{
			Type:      WidgetTypeButton,
			Index:     2,
			X:         375,
			Y:         2,
			Width:     20,
			Height:    16,
			Colour:    gfx.AdvancedColor{Color: gfx.Color(200)},
			Text:      "X",
			Visible:   true,
			Activated: false,
		},
		// OK Button
		{
			Type:      WidgetTypeButton,
			Index:     2,
			X:         100,
			Y:         250,
			Width:     80,
			Height:    30,
			Colour:    gfx.AdvancedColor{Color: gfx.Color(200)},
			Text:      "OK",
			Visible:   true,
			Activated: false,
		},
		// Cancel Button
		{
			Type:      WidgetTypeButton,
			Index:     3,
			X:         220,
			Y:         250,
			Width:     80,
			Height:    30,
			Colour:    gfx.AdvancedColor{Color: gfx.Color(200)},
			Text:      "Cancel",
			Visible:   true,
			Activated: false,
		},
		// Panel (content area)
		{
			Type:    WidgetTypePanel,
			Index:   4,
			X:       10,
			Y:       30,
			Width:   380,
			Height:  200,
			Colour:  gfx.AdvancedColor{Color: gfx.Color(160)},
			Visible: true,
		},
		// Checkbox
		{
			Type:      WidgetTypeCheckbox,
			Index:     5,
			X:         20,
			Y:         50,
			Width:     200,
			Height:    20,
			Colour:    gfx.AdvancedColor{Color: gfx.Color(220)},
			Text:      "Enable feature",
			Visible:   true,
			Activated: false,
		},
		// TextBox
		{
			Type:    WidgetTypeTextBox,
			Index:   6,
			X:       20,
			Y:       80,
			Width:   360,
			Height:  30,
			Colour:  gfx.AdvancedColor{Color: gfx.Color(255)},
			Text:    "Sample text input",
			Visible: true,
		},
	}

	return w
}

// CreateToolbarWindow creates a simple toolbar at the top
func CreateToolbarWindow() *Window {
	w := CreateWindow(WindowTypeToolbar, 0, 0, 800, 40, WindowFlagStickToFront)

	w.Colours[0] = gfx.AdvancedColor{Color: gfx.Color(190)}

	w.Widgets = []Widget{
		// Frame
		{
			Type:    WidgetTypeFrame,
			Index:   0,
			X:       0,
			Y:       0,
			Width:   800,
			Height:  40,
			Colour:  gfx.AdvancedColor{Color: gfx.Color(190)},
			Visible: true,
		},
		// File button
		{
			Type:    WidgetTypeButton,
			Index:   1,
			X:       5,
			Y:       5,
			Width:   60,
			Height:  30,
			Colour:  gfx.AdvancedColor{Color: gfx.Color(200)},
			Text:    "File",
			Visible: true,
		},
		// View button
		{
			Type:    WidgetTypeButton,
			Index:   2,
			X:       70,
			Y:       5,
			Width:   60,
			Height:  30,
			Colour:  gfx.AdvancedColor{Color: gfx.Color(200)},
			Text:    "View",
			Visible: true,
		},
		// Options button
		{
			Type:    WidgetTypeButton,
			Index:   3,
			X:       135,
			Y:       5,
			Width:   70,
			Height:  30,
			Colour:  gfx.AdvancedColor{Color: gfx.Color(200)},
			Text:    "Options",
			Visible: true,
		},
	}

	return w
}
