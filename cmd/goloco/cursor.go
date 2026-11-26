package main

import (
	"github.com/LaPingvino/goloco/pkg/cursor"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/hajimehoshi/ebiten/v2"
)

// The mouse cursor (Input/MouseInput.cpp processMouseOver's cursorId, Ui.cpp
// setCursor; the cursors themselves: pkg/cursor). The system's own cursors
// (pointer, handPointer) are set as they are; a bitmap cursor hides the
// system's and is drawn over the frame at the mouse, less its hot spot —
// after the frame's shots, as SDL's cursor is never in OpenLoco's screenshots.
// PARTIAL: drawn at the game's scale 1 (SDL's cursor isn't scaled); the
// windows' own cursor events are missing (updateCursor, LOOP.md N47).

type cursorState struct {
	id    cursor.Id // _lastCursorId / Ui::getCursor
	set   bool      // setCursor has run once
	image *ebiten.Image
}

// updateCursor is processMouseOver's choice of cursor, each update: the
// pointer; on a resizable window's grip (a panel or frame there, the
// window's size not fixed both ways) diagonalArrows; over a viewport widget
// (the main view's or a window's; no open dropdown) while a tool is on, the
// tool's (ToolManager::getToolCursor), without a tool handPointer when
// getItemLeft finds a vehicle, a label, an industry or a headquarters (the
// map tooltip's pass sets that: mapTipMouseOver); diagonalArrows while a
// window is being resized (Input::State::resizing).
// PARTIAL: the windows' cursor events (callCursor: a list row's
// handPointer, the Vehicle window's openHand / brush / inwardArrows) and the
// tool windows' callToolCursor are not ported (LOOP.md N47).
func (g *Game) updateCursor() {
	id := cursor.Pointer
	mx, my := g.mouseX, g.mouseY
	dropdown := (g.dropdown != nil && g.dropdown.Visible) || (g.colourDropdown != nil && g.colourDropdown.Visible)
	switch {
	case dropdown:
	case g.wins.ResizeGripAt(mx, my):
		id = cursor.DiagonalArrows
	case !g.inTitleScreen && g.viewportWidgetAt(mx, my):
		if g.toolActive() {
			if tool, ok := g.toolCursor(); ok {
				id = tool
			}
		} else if g.mapTip.hand {
			id = cursor.HandPointer
		}
	}
	if g.wins.Resizing() {
		id = cursor.DiagonalArrows
	}
	if !g.cursor.set || id != g.cursor.id { // setCursor only on a change
		g.cursor.id, g.cursor.set, g.cursor.image = id, true, nil
		if img := id.Image(); img != nil {
			g.cursor.image = ebiten.NewImageFromImage(img)
		}
		g.applyCursor()
	}
}

// pickupToolCursor is Vehicle.cpp's kTypeToToolCursor (0x0050029C): by the
// vehicle type, the second when placing the other way (pickupDirection).
var pickupToolCursor = [6][2]cursor.Id{
	{cursor.PlaceTrain, cursor.PlaceTrainAlt},
	{cursor.PlaceBus, cursor.PlaceBusAlt},
	{cursor.PlaceTruck, cursor.PlaceTruckAlt},
	{cursor.PlaceTram, cursor.PlaceTramAlt},
	{cursor.PlacePlane, cursor.PlacePlane},
	{cursor.PlaceShip, cursor.PlaceShip},
}

// toolCursor is ToolManager::getToolCursor: the cursor the tool's window
// gave toolSet (or setToolCursor since).
func (g *Game) toolCursor() (cursor.Id, bool) {
	switch {
	case g.tf.on && g.tf.win != nil:
		return g.terraformCursor()
	case g.consToolOn():
		switch g.consTab {
		case 0: // ConstructionTab.cpp onMouseDown construct / copy / paste, and the copy tool's onToolUp: toolSet(…, crosshair)
			return cursor.Crosshair, true
		case 1: // StationTab.cpp onMouseDown image: placeStation
			return cursor.PlaceStation, true
		case 2: // SignalTab.cpp onMouseDown both / single direction: placeSignal
			return cursor.PlaceSignal, true
		case 3: // OverheadTab.cpp onMouseDown image: crosshair
			return cursor.Crosshair, true
		}
	case g.industryToolOn(): // IndustryList.cpp NewIndustries switchTab: placeFactory
		return cursor.PlaceFactory, true
	case g.pickupOn: // Vehicle.cpp onPickup's toolSet, pickupToolUpdate's setToolCursor
		m := &g.w.Map
		tr, err := m.TrainOf(g.pickupHead)
		if err != nil {
			return 0, false
		}
		dir := 0
		if m.PickupDirection != 0 {
			dir = 1
		}
		if t := int(m.Entities[tr.Head].VehicleType); t < len(pickupToolCursor) {
			return pickupToolCursor[t][dir], true
		}
	case g.routeTool && g.vehicleWin(g.routeHead) != nil: // Vehicle.cpp Route tab's toolSet(…, widx::tool, crosshair)
		return cursor.Crosshair, true
	}
	return 0, false
}

// terraformCursor is the Terraform window's tools' cursors.
func (g *Game) terraformCursor() (cursor.Id, bool) {
	t := &g.tf
	switch t.win.Tab {
	case ui.TerraformTabClear: // ClearArea tabReset: toolSet(…, bulldozerTool)
		return cursor.BulldozerTool, true
	case ui.TerraformTabLand, ui.TerraformTabWater: // landTool / waterTool; setToolCursor(upDownArrow) while the button holds a selection
		if t.arrow {
			return cursor.UpDownArrow, true
		}
		if t.win.Tab == ui.TerraformTabWater {
			return cursor.WaterTool, true
		}
		return cursor.LandTool, true
	case ui.TerraformTabTrees: // PlantTrees: plantTree
		return cursor.PlantTree, true
	case ui.TerraformTabWalls: // BuildWalls: placeFence
		return cursor.PlaceFence, true
	}
	return 0, false
}

// applyCursor is SDL_SetCursor for the chosen cursor: a system one shown, a
// bitmap one hidden from the system and drawn (drawCursor). Not while the
// right drag holds the cursor hidden (Ui::hideCursor): its release shows the
// chosen one again.
func (g *Game) applyCursor() {
	if g.rightDragging() {
		return
	}
	switch g.cursor.id.System() {
	case "":
		ebiten.SetCursorMode(ebiten.CursorModeHidden)
		return
	case "SDL_SYSTEM_CURSOR_POINTER":
		ebiten.SetCursorShape(ebiten.CursorShapePointer)
	default:
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
	}
	ebiten.SetCursorMode(ebiten.CursorModeVisible)
}

// drawCursor draws a bitmap cursor at the mouse, its hot spot on it.
func (g *Game) drawCursor(screen *ebiten.Image) {
	if g.cursor.image == nil || g.rightDragging() {
		return
	}
	hx, hy := g.cursor.id.Hot()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(g.mouseX-hx), float64(g.mouseY-hy))
	screen.DrawImage(g.cursor.image, op)
}
