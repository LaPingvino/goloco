package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
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
	"Build Vehicles": {objects.ISToolbarEmptyOpaque, objects.ISToolbarEmptyOpaqueHover}, // its face: the last type (Draw)
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
	// What the Railroad and Road buttons show (ToolbarTop.cpp draw): the
	// game state's last options (bit 7: a road object; 0xFF: none) in the
	// player's colour. The game sets them each frame.
	LastRailroadOption, LastRoadOption uint8
	CompanyColour                      uint8
	// The Vehicles and Build Vehicles buttons' faces: the last VehicleType
	// each menu chose (lastVehicleType, lastBuildVehiclesOption).
	LastVehicleType, LastBuildVehiclesOption uint8
	// AudioMuted is !Audio::isAudioEnabled: the Audio button shows its
	// inactive image (prepareDraw).
	AudioMuted bool
	// Centred is Config::toolbarButtonsCentred (Options → Controls): Layout's mode.
	Centred bool
	// Colours are the window's four colours (ToolbarTop open: the interface
	// skin's topToolbarPrimary … QuaternaryColour), by ColorGroup.
	Colours [4]uint8
}

// vehicleButtonFaces is ToolbarTop.cpp draw's button_face_image_ids and
// kBuildVehicleImages, by VehicleType (train, bus, truck, tram, aircraft,
// ship): interface-skin image indices.
var vehicleButtonFaces = [6]int{objects.ISVehicleTrainFrame0, objects.ISVehicleBusesFrame0, objects.ISVehicleTrucksFrame0,
	objects.ISVehicleTramsFrame0, objects.ISVehicleAircraftFrame0, objects.ISVehicleShipsFrame0}
var buildVehicleButtonFaces = [6]int{objects.ISToolbarBuildVehicleTrain, objects.ISToolbarBuildVehicleBus, objects.ISToolbarBuildVehicleTruck,
	objects.ISToolbarBuildVehicleTram, 39, 41} // toolbar_build_vehicle_airplane, _boat

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
		// prepareDraw's leftAlignButtons: the hidden cheats slot takes no
		// room — the group starts 11 px after the audio button (justifyToolbar).
		{"Zoom", 71, 1, false},
		{"Rotate", 101, 1, false},
		{"View", 131, 1, false},

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

// Layout is ToolbarTop's prepareDraw placement (ToolbarTopCommon.cpp
// justifyToolbar / centreToolbar, 2da9f6f + 7ba7099): justified — the
// load/save, audio (and cheats) and the zoom, rotate, view groups from the
// left, 11 apart; the construction and the list groups from the right
// (max(640, width) − 1), 11 apart; the toolbar the screen's width. Centred
// (Config::toolbarButtonsCentred) — every group side by side from the left,
// 11 apart, and the toolbar shrunk to them (n × 30 + 3 × 11) in the middle of
// the screen. Hidden buttons take no room.
func (t *Toolbar) Layout(screenWidth int, centred bool) {
	group := func(g int) []int {
		var out []int
		for i := range t.Buttons {
			if t.Buttons[i].ColorGroup == g && !t.Buttons[i].Hidden {
				out = append(out, i)
			}
		}
		return out
	}
	left := func(x int, g int) int { // leftAlignButtons: left = x, right = x + 29, x += 30
		for _, i := range group(g) {
			t.Buttons[i].X = t.X + x
			x += 30
		}
		return x
	}
	if centred {
		n := 0
		for i := range t.Buttons {
			if !t.Buttons[i].Hidden {
				n++
			}
		}
		total := n*30 + 3*11
		t.X, t.Width = max(0, (screenWidth-total)/2), total
		x := left(0, 0) + 11
		x = left(x, 1) + 11
		x = left(x, 2) + 11
		left(x, 3)
		return
	}
	t.X, t.Width = 0, screenWidth
	left(left(0, 0)+11, 1)
	right := func(x int, g int) int { // rightAlignButtons: right = x, left = x − 29, x −= 30 (the group's last first)
		ids := group(g)
		for k := len(ids) - 1; k >= 0; k-- {
			t.Buttons[ids[k]].X = x - 29
			x -= 30
		}
		return x
	}
	x := max(640, screenWidth) - 1
	x = right(x, 3) - 11
	right(x, 2)
}

// RightAlignButtons is Layout justified (the screen's width changed).
func (t *Toolbar) RightAlignButtons(screenWidth int) { t.Layout(screenWidth, t.Centred) }

// Draw is ToolbarTop's window draw (transparent, no background: the map
// shows between the buttons) — each button's ImageButtonAlt::draw: its skin
// image recoloured to its group's colour (Colours: topToolbarPrimary …
// QuaternaryColour), the next image while its menu is open (activated) —
// then drawTabs (0x00439DE4): the Railroad and Road buttons' face (the last
// option's track or road image in the player's colour, a pixel higher unless
// the menu is open) under the empty transparent frame, the Vehicles button's
// face (the last vehicle type's), the Build Vehicles button's face (the last
// type's, the next image while open).
func (t *Toolbar) Draw(screen *ebiten.Image, renderer *render.Renderer) {
	if !t.Visible || globalG1Font == nil || renderer == nil || renderer.ObjMgr == nil || renderer.ObjMgr.InterfaceSkin == nil {
		return
	}
	g1 := globalG1Font.g1
	om := renderer.ObjMgr
	skin := int(om.InterfaceSkin.ImageOffset)
	draw8(screen, t.X, t.Y, t.Width, t.Height, func(rt *gfx8.RenderTarget) {
		img := func(id, x, y int, colour uint8) {
			if id < 0 || id >= len(g1.Elements) {
				return
			}
			sp, err := gfx8.Decode(&g1.Elements[id])
			if err != nil {
				return
			}
			m, _ := g1.GetPaletteMap(int(colour))
			gfx8.DrawSpriteNoise(rt, sp, x-t.X, y-t.Y, gfx8.OpFor(sp, m != nil, false), m, nil)
		}
		for i := range t.Buttons {
			b := &t.Buttons[i]
			ids, ok := buttonImageIDs[b.Tooltip]
			if b.Hidden || !ok || b.ColorGroup < 0 || b.ColorGroup > 3 {
				continue
			}
			if b.Tooltip == "Audio" && t.AudioMuted { // prepareDraw: toolbar_audio_inactive without sound
				ids = [2]uint32{objects.ISToolbarAudioInactive, objects.ISToolbarAudioInactiveHover}
			}
			id := ids[0]
			if b.Pressed {
				id = ids[1] // withIndexOffset(1)
			}
			img(skin+int(id), b.X, b.Y, t.Colours[b.ColorGroup])
		}
		for i := range t.Buttons {
			b := &t.Buttons[i]
			if b.Hidden {
				continue
			}
			faceUnderFrame := func(face int) {
				y := b.Y - 1
				frame := skin + objects.ISToolbarEmptyTransparent
				if b.Pressed {
					y, frame = y+1, frame+1
				}
				img(face, b.X, y, t.CompanyColour)
				img(frame, b.X, b.Y, t.Colours[b.ColorGroup])
			}
			switch b.Tooltip {
			case "Railroad", "Road":
				opt := t.LastRailroadOption
				if b.Tooltip == "Road" {
					opt = t.LastRoadOption
				}
				if opt == 0xFF {
					continue
				}
				face := -1
				if opt&(1<<7) != 0 {
					if ro := om.GetRoadObjectByIndex(int(opt &^ (1 << 7))); ro != nil {
						face = int(ro.Image)
					}
				} else if tr := om.GetTrackObjectByIndex(int(opt)); tr != nil {
					face = int(tr.Image) // + kUiPreviewImage0 (0)
				}
				faceUnderFrame(face)
			case "Vehicles":
				if int(t.LastVehicleType) < len(vehicleButtonFaces) {
					faceUnderFrame(skin + vehicleButtonFaces[t.LastVehicleType])
				}
			case "Build Vehicles":
				if int(t.LastBuildVehiclesOption) < len(buildVehicleButtonFaces) {
					face := skin + buildVehicleButtonFaces[t.LastBuildVehiclesOption]
					if b.Pressed {
						face++
					}
					img(face, b.X, b.Y, t.CompanyColour)
				}
			}
		}
	})
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
