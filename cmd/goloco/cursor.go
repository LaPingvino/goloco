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
// PARTIAL: only the Terraform window's tools give their cursors (the other
// tools', the hand over a vehicle or a label, the resize corner's arrows:
// LOOP.md N47); drawn at the game's scale 1 (SDL's cursor isn't scaled).

type cursorState struct {
	id    cursor.Id // _lastCursorId / Ui::getCursor
	set   bool      // setCursor has run once
	image *ebiten.Image
}

// updateCursor is processMouseOver's choice of cursor, each update: the
// pointer, but over the map (no window, no open dropdown under the mouse)
// while a tool is on, the tool's (ToolManager::getToolCursor).
func (g *Game) updateCursor() {
	id := cursor.Pointer
	dropdown := (g.dropdown != nil && g.dropdown.Visible) || (g.colourDropdown != nil && g.colourDropdown.Visible)
	if !g.inTitleScreen && !dropdown && !g.overWindow(g.mouseX, g.mouseY) {
		if tool, ok := g.toolCursor(); ok {
			id = tool
		}
	}
	if !g.cursor.set || id != g.cursor.id { // setCursor only on a change
		g.cursor.id, g.cursor.set, g.cursor.image = id, true, nil
		if img := id.Image(); img != nil {
			g.cursor.image = ebiten.NewImageFromImage(img)
		}
		g.applyCursor()
	}
}

// toolCursor is ToolManager::getToolCursor for the window that has the tool.
func (g *Game) toolCursor() (cursor.Id, bool) {
	t := &g.tf
	if !t.on || t.win == nil {
		return 0, false
	}
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
	if g.vpRight.on {
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
	if g.cursor.image == nil || g.vpRight.on {
		return
	}
	hx, hy := g.cursor.id.Hot()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(g.mouseX-hx), float64(g.mouseY-hy))
	screen.DrawImage(g.cursor.image, op)
}
