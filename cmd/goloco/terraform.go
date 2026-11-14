package main

import (
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scene"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/hajimehoshi/ebiten/v2"
	"log"
)

// The Terraform window's tools (Ui/Windows/TerraForm.cpp): the toolbar's
// terraform menu opens the window on a tab; each tab sets its tool
// (tabReset's toolSet), which follows the cursor (onToolUpdate: the map
// selection, the ghost, the cost), acts on a press (onToolDown), while the
// button is held (toolDrag) and on its release (toolUp), through the game
// commands of pkg/sim. The window itself: pkg/ui terraformwindow.go.
// PARTIAL: the construct sounds; the tool cursors; the editor (no cost,
// buildImmediately); defaultTreeObjectId / defaultWallObjectId are kept for
// the session, not in the save.

// terraformTool is the window's statics.
type terraformTool struct {
	on  bool
	win *ui.TerraformWindow
	// _adjustToolSize and the tabs' own (_adjustLandToolSize,
	// _clearAreaToolSize, _adjustWaterToolSize; Gui::init: 1, 2, 1).
	toolSize                       int
	landSize, clearSize, waterSize int
	lastLand                       uint8 // _lastSelectedLand
	// _raiseLandCost / _lowerLandCost (the water tab: _raiseWaterCost /
	// _lowerWaterCost), ok false: kFailure.
	raiseCost, lowerCost int32
	raiseOK, lowerOK     bool
	mountain, paint      bool // AdjustLand's isMountainMode, isPaintMode
	// arrow: the tool cursor is CursorId::upDownArrow (the button went down
	// on a selection: it stays, and dragging raises / lowers).
	arrow     bool
	held      bool // Input::State::viewportLeft with the tool: toolDrag runs
	dragLastY int  // Input::getDragLastLocation().y
	// The ghost (_terraformGhostPlacedFlags, …Pos, …BaseZ, …Type, …Quadrant,
	// …Rotation, …TreeBuildImmediate).
	ghostTree, ghostWall bool
	ghostX, ghostY       int32
	ghostBaseZ           uint8
	ghostType            uint8
	ghostQuadrant        uint8
	ghostRotation        uint8
	ghostImmediate       bool
	// _treeRotation (2 at start), _treeColour, _treeClusterType,
	// _lastTreeCost (ok false: kFailure).
	treeRotation, treeColour uint8
	cluster                  int
	treeCost                 int32
	treeCostOK               bool
	// GameState defaultTreeObjectId / defaultWallObjectId (−1: 0xFF, none).
	defaultTree, defaultWall int
	inited                   bool
}

// terraformMenu is ToolbarTop.cpp terraformMenuMouseDown (0x0043A3C3): clear
// area, adjust land, adjust water, plant trees, build walls — the skin's
// images, the default land's and the water's —, the first highlighted.
// Choosing one is terraformMenuDropdown: Terraform::openClearArea ….
func (g *Game) terraformMenu(btn *ui.ToolbarButton) {
	if g.skin() == nil || g.r.ObjMgr == nil {
		return
	}
	skin := int(g.skin().ImageOffset)
	land, water := g.terraformTabImages()
	images := [5]int{skin + 180, land, water, skin + 179, skin + 187} // toolbar_menu_bulldozer, _plant_trees, _build_walls
	labels := [5]int{i18n.MenuClearArea, i18n.MenuAdjustLand, i18n.MenuAdjustWater, i18n.MenuPlantTrees, i18n.MenuBuildWalls}
	var items []ui.DropdownItem
	for i := range 5 {
		tab := i
		items = append(items, ui.DropdownItem{MenuSprite: true, Sprite: images[i], Text: i18n.Label(labels[i]), Action: func() { g.openTerraform(tab) }})
	}
	d := g.toolbarImageMenu(btn, items)
	d.Highlighted = 0
	g.dropdown = d
}

// terraformTabImages are drawTabs' land and water images: the default land
// object's toolbar_terraform_land (mapPixelImage + 3), the water object's
// kToolbarTerraformWater (image + 42).
func (g *Game) terraformTabImages() (land, water int) {
	om := g.r.ObjMgr
	if om == nil {
		return 0, 0
	}
	if i := int(g.w.Map.LastLandOption); i < len(om.LandObjects) && om.LandObjects[i] != nil {
		land = int(om.LandObjects[i].MapPixelImage) + 3
	}
	if om.WaterObj != nil {
		water = int(om.WaterObj.ImageOffset) + 42
	}
	return land, water
}

func (g *Game) terraformWin() *ui.TerraformWindow {
	return winOf[*ui.TerraformWindow](g, ui.WindowTypeTerraform, ui.AnyNumber)
}

// openTerraform is Terraform::open and the tab's callOnMouseUp
// (openClearArea …): the window created at the screen's right under the
// toolbar, the windows there moved down, the gridlines shown; then the tab.
func (g *Game) openTerraform(tab int) {
	t := &g.tf
	if !t.inited { // Gui::init's tool sizes, resetDefaultObjectIds
		t.inited, t.landSize, t.clearSize, t.waterSize, t.treeRotation, t.defaultTree, t.defaultWall = true, 1, 2, 1, 2, -1, -1
	}
	win := g.terraformWin()
	if win == nil {
		g.cancelOtherTools()
		g.syncScreen()
		win = ui.NewTerraformWindow(g.sw)
		win.WinFlags = ui.WindowFlagNoAutoClose
		win.OnClose = g.terraformEnd
		win.OnTab = g.terraformSwitchTab
		win.OnToolSize = g.terraformToolSize
		win.OnMountain = func() { t.mountain, t.paint = !t.mountain, false; g.terraformTabReset() }
		win.OnPaint = func() { t.mountain, t.paint = false, !t.paint; g.terraformTabReset() }
		win.OnLandMaterial = g.terraformLandDropdown
		win.OnSelect = g.terraformSelect
		win.OnRotate = func() { t.treeRotation = (t.treeRotation + 1) & 3 }
		win.OnColour = g.terraformColourDropdown
		win.OnCluster = func(kind int) {
			if t.cluster == kind {
				kind = 0
			}
			t.cluster = kind
		}
		t.win, t.on = win, true
		t.ghostTree, t.ghostWall, t.treeCostOK, t.cluster = false, false, false, 0
		g.wins.Show(win)
		g.wins.MoveOtherWindowsDown(win)
		g.w.SetGridlines(true)
	} else {
		g.wins.BringToFront(win)
	}
	g.terraformSwitchTab(tab)
}

// cancelOtherTools is ToolManager::toolSet's toolCancel for the tools of
// other windows: the industry, pickup and route tools end; a construction
// window with its tool on the map (hovering, or on another tab) closes.
func (g *Game) cancelOtherTools() {
	if g.indTool.on {
		g.industryToolAbort()
	}
	if g.pickupOn {
		head := g.pickupHead
		g.pickupToolAbort()
		g.pickupOn = false
		g.refreshVehicleWindow(head, 1)
	}
	g.routeTool = false
	if g.consToolOn() {
		g.wins.CloseType(ui.WindowTypeConstruction, ui.AnyNumber)
	}
}

// consToolOn reports the construction window's tool on the map: any state
// but building on from a position on the Construction tab (which has no
// tool — toolCancel).
func (g *Game) consToolOn() bool {
	if g.buildMode == buildModeNone {
		return false
	}
	return !(g.consHasHead() && g.consTab == 0 && !g.w.Cons.Hover)
}

// closeTerraform closes the window (another window's toolSet took the tool:
// Common::onUpdate closes it).
func (g *Game) closeTerraform() {
	if g.tf.on {
		g.wins.CloseType(ui.WindowTypeTerraform, ui.AnyNumber)
	}
}

// terraformEnd is the tabs' onClose and the tool's end (toolCancel →
// onToolAbort): the ghosts removed, the selection and the gridlines off.
func (g *Game) terraformEnd() {
	t := &g.tf
	if !t.on {
		return
	}
	g.removeTreeGhost()
	g.removeWallGhost()
	g.w.SetMapSelection(false, 0, 0, 0, 0)
	g.w.SetGridlines(false) // Main::hideGridlines (PARTIAL: Config gridlinesOnLandscape keeps them)
	t.on, t.arrow, t.held, t.win = false, false, false, nil
}

// terraformSwitchTab is Common::switchTab: the tool cancelled, the tab's
// size and widgets, then its tabReset.
func (g *Game) terraformSwitchTab(tab int) {
	t := &g.tf
	if t.win == nil {
		return
	}
	g.removeTreeGhost() // toolCancel → onToolAbort
	g.removeWallGhost()
	g.w.SetMapSelection(false, 0, 0, 0, 0)
	t.arrow, t.held = false, false
	t.win.Editor = false
	t.win.SwitchTab(tab)
	g.terraformTabReset()
	g.prepareTerraformWindow(t.win)
	if tab == ui.TerraformTabTrees || tab == ui.TerraformTabWalls {
		t.win.FocusSelected() // updateActiveThumb
	}
}

// terraformTabReset is the tab's tabReset: the costs forgotten, the tab's
// tool size, the ghost flags; the grids' lists with their chosen object
// (refreshTreeList / refreshWallList: the default one, else the first).
func (g *Game) terraformTabReset() {
	t, om := &g.tf, g.r.ObjMgr
	if t.win == nil || om == nil {
		return
	}
	t.arrow = false
	g.w.SetMapSelection(false, 0, 0, 0, 0)
	switch t.win.Tab {
	case ui.TerraformTabClear:
		t.raiseOK, t.toolSize = false, t.clearSize
	case ui.TerraformTabLand:
		for i, lo := range om.LandObjects { // the first land object
			if i < 32 && lo != nil {
				t.lastLand = uint8(i)
				t.raiseOK, t.lowerOK, t.toolSize = false, false, t.landSize
				break
			}
		}
	case ui.TerraformTabWater:
		t.raiseOK, t.lowerOK, t.toolSize = false, false, t.waterSize
	case ui.TerraformTabTrees:
		t.ghostTree, t.ghostWall, t.treeCostOK = false, false, false
		t.win.Selected = firstObject(t.defaultTree, len(om.TreeObjects), func(i int) bool { return om.TreeObjects[i] != nil })
		g.updateTreeColours()
	case ui.TerraformTabWalls:
		t.ghostTree, t.ghostWall = false, false
		t.win.Selected = firstObject(t.defaultWall, len(om.WallObjects), func(i int) bool { return om.WallObjects[i] != nil })
	}
}

// firstObject is refreshTreeList's rowHover: the default object when it is
// loaded, else the first loaded one, −1 none.
func firstObject(def, n int, loaded func(int) bool) int {
	if def >= 0 && def < n && loaded(def) {
		return def
	}
	for i := range n {
		if loaded(i) {
			return i
		}
	}
	return -1
}

// updateTreeColours (0x004BB6B2): the chosen tree's last colour option
// becomes the tree colour.
func (g *Game) updateTreeColours() {
	t := &g.tf
	if to := g.treeObject(t.win.Selected); to != nil && to.Colours != 0 {
		for i := range 32 {
			if to.Colours&(1<<i) != 0 {
				t.treeColour = uint8(i)
			}
		}
	}
}

func (g *Game) treeObject(id int) *objects.TreeObject {
	if om := g.r.ObjMgr; om != nil && id >= 0 && id < len(om.TreeObjects) {
		return om.TreeObjects[id]
	}
	return nil
}

// terraformToolSize is the decrease / increase buttons: the size within
// 1–64, kept as the tab's own.
func (g *Game) terraformToolSize(delta int) {
	t := &g.tf
	t.toolSize = min(max(t.toolSize+delta, 1), 64)
	switch t.win.Tab {
	case ui.TerraformTabClear:
		t.clearSize = t.toolSize
	case ui.TerraformTabLand:
		t.landSize = t.toolSize
	case ui.TerraformTabWater:
		t.waterSize = t.toolSize
	}
}

// terraformSelect is scrollMouseDown: the cell's object chosen and made the
// default; a tree's colour and cost follow.
func (g *Game) terraformSelect(id int) {
	t := &g.tf
	if t.win.Tab == ui.TerraformTabTrees {
		t.defaultTree, t.treeCostOK = id, false
		g.updateTreeColours()
	} else {
		t.defaultWall = id
	}
}

// terraformLandDropdown is AdjustLand's showDropdown (0x004BCB47): the land
// objects' icons (20 × 20), the chosen one highlighted; onDropdown takes it.
func (g *Game) terraformLandDropdown(x, y, w, h int) {
	t, om := &g.tf, g.r.ObjMgr
	var images []int
	var ids []uint8
	highlighted := -1
	for i, lo := range om.LandObjects {
		if i >= 32 || lo == nil {
			continue
		}
		if uint8(i) == t.lastLand {
			highlighted = len(ids)
		}
		images = append(images, int(lo.MapPixelImage)+1) // landscape_generator_tile_icon
		ids = append(ids, uint8(i))
	}
	d := ui.NewImageDropdown(x, y, h, 20, images, ids, highlighted, t.win.Colour, g.sw, g.sh)
	d.OnSelect = func(id uint8) { t.lastLand = id }
	g.colourDropdown = d
}

// terraformColourDropdown is PlantTrees' onMouseDown: Dropdown::showColour
// with the tree's colour options.
func (g *Game) terraformColourDropdown(x, y, w, h int) {
	t := &g.tf
	to := g.treeObject(t.win.Selected)
	if to == nil {
		return
	}
	d := ui.NewColourDropdown(x, y, h, to.Colours, t.treeColour, t.win.Colour, g.sw, g.sh)
	d.OnSelect = func(c uint8) { t.treeColour = c }
	g.colourDropdown = d
}

// treeGrowth is TreeObject.cpp's treeGrowth (0x00500775):
// getTreeGrowthDisplayOffset by the object's growth.
var treeGrowth = [11]uint8{1, 0, 1, 2, 2, 3, 4, 5, 6, 0, 0}

// prepareTerraformWindow is the tabs' prepareDraw: what the window shows of
// the tool's state, and the grids' cells (drawTreeThumb's image for the
// tree rotation; the wall's sprite).
func (g *Game) prepareTerraformWindow(w *ui.TerraformWindow) {
	t, om, m := &g.tf, g.r.ObjMgr, &g.w.Map
	if om == nil || t.win != w {
		return
	}
	company, _ := g.w.ControllingCompany()
	if company >= 0 && company < len(g.w.CompanyColours) {
		w.TitleColour = g.w.CompanyColours[company]
	}
	if sk := g.skin(); sk != nil {
		w.Colour, w.SkinImage = sk.WindowTerraFormColour, int(sk.ImageOffset)
	}
	w.LandTabImage, w.WaterTabImage = g.terraformTabImages()
	w.ToolSize = t.toolSize
	w.RaiseCost, w.LowerCost, w.RaiseOK, w.LowerOK = int64(t.raiseCost), int64(t.lowerCost), t.raiseOK, t.lowerOK
	w.MountainMode, w.PaintMode = t.mountain, t.paint
	if i := int(t.lastLand); i < len(om.LandObjects) && om.LandObjects[i] != nil {
		lo := om.LandObjects[i]
		w.LandIcon = int(lo.MapPixelImage) + 1
		if g1 := g.r.G1; g1 != nil && int(lo.MapPixelImage) < len(g1.Elements) && len(g1.Elements[lo.MapPixelImage].Data) > 0 {
			w.LandPixelColour = g1.Elements[lo.MapPixelImage].Data[0] // getG1Element(mapPixelImage)->offset[0]
		}
	}
	w.TreeRotation, w.TreeColour, w.Cluster = t.treeRotation, t.treeColour, t.cluster
	w.TreeCost, w.TreeCostOK = int64(t.treeCost), t.treeCostOK
	w.Cells = w.Cells[:0]
	switch w.Tab {
	case ui.TerraformTabTrees:
		for i, to := range om.TreeObjects {
			if to == nil {
				continue
			}
			var growth uint8
			if int(to.Growth) < len(treeGrowth) {
				growth = treeGrowth[to.Growth]
			}
			img := int(growth)*int(to.NumRotations) + int((to.NumRotations-1)&t.treeRotation)
			if int(to.SeasonState) < len(to.Sprites) {
				img += int(to.Sprites[to.SeasonState])
			}
			w.Cells = append(w.Cells, ui.TerraformCell{ID: i, Name: to.Name, Image: img, Colours: to.Colours,
				Cost: int64(m.InflationAdjustedCost(to.BuildCostFactor, to.CostIndex, 12)), Rotatable: to.NumRotations != 1})
		}
	case ui.TerraformTabWalls:
		for i, wo := range om.WallObjects {
			if wo != nil {
				w.Cells = append(w.Cells, ui.TerraformCell{ID: i, Name: wo.Name, Image: int(wo.Sprite)})
			}
		}
	}
}

// updateTerraform is the window's and the tool's share of a frame: the
// window closes when another tool took over (Common::onUpdate); the tool
// follows the cursor (processMouseTool → onToolUpdate), drags while the
// button that went down on the map is held (stateViewportLeft → toolDrag)
// and ends the drag on its release (toolUp).
func (g *Game) updateTerraform() {
	t := &g.tf
	if !t.on || g.inTitleScreen {
		return
	}
	if t.win == nil || t.win.Stack() == nil {
		g.terraformEnd()
		return
	}
	if g.consToolOn() || g.indTool.on || g.pickupOn || g.routeTool {
		g.closeTerraform()
		return
	}
	mx, my := g.mouseX, g.mouseY
	onMap := !g.overWindow(mx, my) && !(g.dropdown != nil && g.dropdown.Visible)
	g.terraformToolUpdate(mx, my, onMap)
	if t.held {
		if g.inMousePressed(ebiten.MouseButtonLeft) {
			g.terraformToolDrag(mx, my, onMap)
		} else {
			t.held = false
			g.terraformToolUp()
		}
	}
}

// selectTiles is setMapSelectionTiles: toolSize × toolSize tiles around the
// map position, of a selection type.
func (g *Game) selectTiles(px, py int32, typ uint8, size int) {
	size = max(size, 1)
	ax, ay := (int(px)-(size*32/2-16))>>5, (int(py)-(size*32/2-16))>>5
	g.w.SetMapSelectionType(true, typ, ax, ay, ax+size-1, ay+size-1)
}

// selectionArgs are the commands' arguments from the selection: the tiles
// a..b and the centre ((pointA + pointB) / 2, world units).
func (g *Game) selectionArgs() (a [4]int, cx, cy int32, typ uint8) {
	_, typ, a = g.w.MapSelection()
	return a, int32(a[0]+a[2]) * 32 / 2, int32(a[1]+a[3]) * 32 / 2, typ
}

// terraformToolUpdate is the tab's onToolUpdate at a screen point (onMap
// false: over a window, where no surface is found).
func (g *Game) terraformToolUpdate(x, y int, onMap bool) {
	t := &g.tf
	surface := func() (int32, int32, bool) {
		if !onMap {
			return 0, 0, false
		}
		return g.w.SurfaceAt(x, y)
	}
	switch t.win.Tab {
	case ui.TerraformTabClear:
		g.clearToolUpdate(surface)
	case ui.TerraformTabLand:
		if t.paint {
			g.clearToolUpdate(surface) // onPaintToolUpdate: the same selection and (clear) cost
			return
		}
		// onAdjustLandToolUpdate
		if !t.arrow {
			px, py, ok := surface()
			if !ok {
				g.w.SetMapSelection(false, 0, 0, 0, 0)
				t.raiseOK, t.lowerOK = false, false
				return
			}
			if t.landSize == 1 && !t.mountain { // setMapSelectionSingleTile(loc, true): a corner, or the whole tile at its centre
				g.w.SetMapSelectionType(true, quadrantOrCentre(px, py), int(px)>>5, int(py)>>5, int(px)>>5, int(py)>>5)
			} else {
				g.selectTiles(px, py, scene.SelFull, t.toolSize)
			}
		} else if on, _, _ := g.w.MapSelection(); !on {
			return
		}
		t.lowerCost, t.lowerOK = g.lowerLand(sim.CmdAllowNegativeCashFlow)
		t.raiseCost, t.raiseOK = g.raiseLand(sim.CmdAllowNegativeCashFlow)
	case ui.TerraformTabWater:
		if !t.arrow {
			px, py, ok := int32(0), int32(0), false
			if onMap {
				px, py, ok = g.w.SurfaceOrWaterAt(x, y) // getMapCoordinatesFromPos(surface | water): the tile
			}
			if !ok {
				g.w.SetMapSelection(false, 0, 0, 0, 0)
				t.raiseOK, t.lowerOK = false, false
				return
			}
			g.selectTiles(px+16, py+16, scene.SelFullWater, t.toolSize)
		} else if on, _, _ := g.w.MapSelection(); !on {
			return
		}
		t.raiseCost, t.raiseOK = g.raiseWater(0)
		t.lowerCost, t.lowerOK = g.lowerWater(0)
	case ui.TerraformTabTrees:
		a, ok := g.treePlacementAt(surface)
		if !ok {
			g.w.SetMapSelection(false, 0, 0, 0, 0)
			g.removeTreeGhost()
			return
		}
		tx, ty := int(a.X)>>5, int(a.Y)>>5
		g.w.SetMapSelectionType(true, scene.SelQuarter0+(a.Quadrant^2), tx, ty, tx, ty)
		if t.ghostTree && t.ghostX == a.X && t.ghostY == a.Y && t.ghostQuadrant == a.Quadrant && t.ghostType == a.Type &&
			t.ghostRotation == a.Rotation && t.ghostImmediate == a.BuildImmediately {
			return
		}
		g.removeTreeGhost()
		t.ghostQuadrant, t.ghostRotation, t.ghostImmediate = a.Quadrant, a.Rotation, a.BuildImmediately
		t.treeCost, t.treeCostOK = g.placeTreeGhost(a)
	case ui.TerraformTabWalls:
		a, ok := g.wallPlacementAt(surface)
		if !ok {
			g.w.SetMapSelection(false, 0, 0, 0, 0)
			g.removeWallGhost()
			return
		}
		tx, ty := int(a.X)>>5, int(a.Y)>>5
		g.w.SetMapSelectionType(true, scene.SelEdge0+a.Rotation, tx, ty, tx, ty)
		if t.ghostWall && t.ghostX == a.X && t.ghostY == a.Y && t.ghostRotation == a.Rotation && t.ghostType == a.Type {
			return
		}
		g.removeWallGhost()
		t.ghostRotation = a.Rotation
		g.placeWallGhost(a)
	}
}

// clearToolUpdate is ClearArea's onToolUpdate (0x004BC677): the tool-size
// selection under the cursor and what clearing it would cost.
func (g *Game) clearToolUpdate(surface func() (int32, int32, bool)) {
	t := &g.tf
	px, py, ok := surface()
	if !ok {
		g.w.SetMapSelection(false, 0, 0, 0, 0)
		t.raiseOK = false
		return
	}
	g.selectTiles(px, py, scene.SelFull, t.toolSize)
	t.raiseCost, t.raiseOK = g.clearLand(sim.CmdAllowNegativeCashFlow | sim.CmdGhost)
}

// quadrantOrCentre is getQuadrantOrCentreFromPos (0x0045FD8E): the whole
// tile within 7 of its centre, else the corner the position is nearest.
func quadrantOrCentre(px, py int32) uint8 {
	dx, dy := px&^0x1F+16-px, py&^0x1F+16-py
	if max(max(dx, -dx), max(dy, -dy)) <= 7 {
		return scene.SelFull
	}
	return quadrantFromPos(px, py)
}

// quadrantFromPos is getQuadrantFromPos (0x0045FE05).
func quadrantFromPos(px, py int32) uint8 {
	xn, yn := px&0x1F, py&0x1F
	switch {
	case xn > 16 && yn >= 16:
		return 0
	case xn > 16:
		return 1
	case yn >= 16:
		return 3
	}
	return 2
}

// sideFromPos is getSideFromPos (0x0045FE4C): the tile edge the position is
// nearest.
func sideFromPos(px, py int32) uint8 {
	xn, yn := px&0x1F, py&0x1F
	switch {
	case xn < yn && xn+yn < 32:
		return 0
	case xn < yn:
		return 1
	case xn+yn < 32:
		return 3
	}
	return 2
}

// clickable(TerraForm.cpp): panel
//
// terraformToolDown is the tab's onToolDown: a press on the map with the
// tool. The drag begins (State::viewportLeft; _dragLast).
func (g *Game) terraformToolDown(x, y int) {
	t := &g.tf
	t.held, t.dragLastY = true, y
	surface := func() (int32, int32, bool) { return g.w.SurfaceAt(x, y) }
	if g.script != nil { // a playthrough's expect names the tile
		if px, py, ok := surface(); ok {
			tl := g.w.Map.TileAt(int(px)>>5, int(py)>>5)
			log.Printf("[Script] terraform tool down on tile %d,%d (height %d, slope %#x)", px>>5, py>>5, tl.BaseZ, tl.Slope&0x1F)
		}
	}
	switch t.win.Tab {
	case ui.TerraformTabClear:
		g.clearLandApply()
	case ui.TerraformTabLand:
		if t.paint {
			g.paintLand()
		} else if on, _, _ := g.w.MapSelection(); on {
			t.arrow = true // setToolCursor(upDownArrow)
		}
	case ui.TerraformTabWater:
		if on, _, _ := g.w.MapSelection(); on {
			t.arrow = true
		}
	case ui.TerraformTabTrees:
		g.removeTreeGhost()
		a, ok := g.treePlacementAt(surface)
		if !ok {
			return
		}
		c := g.playerCmd()
		g.unpauseForCommand(sim.CmdApply)
		switch t.cluster {
		case 0:
			if _, ok := g.w.Map.DoCommandFor(c, sim.CmdApply, sim.ExpenditureConstruction, func(f uint8) (int32, bool) {
				cost, _, ok := g.w.Map.CreateTreeCmd(c, f, a)
				return cost, ok
			}); !ok {
				g.openError(i18n.Label(i18n.CantPlantThisHere), cmdErrorText(c))
			}
		case 1: // placeTreeCluster(pos, 320, 3, the tree)
			if !g.w.Map.PlaceTreeClusterFor(c, int(a.X)>>5, int(a.Y)>>5, 320, 3, int(a.Type)) {
				g.openError(i18n.Label(i18n.CantPlantThisHere), "")
			}
		case 2: // placeTreeCluster(pos, 384, 4, any)
			if !g.w.Map.PlaceTreeClusterFor(c, int(a.X)>>5, int(a.Y)>>5, 384, 4, -1) {
				g.openError(i18n.Label(i18n.CantPlantThisHere), "")
			}
		}
	case ui.TerraformTabWalls:
		g.removeWallGhost()
		a, ok := g.wallPlacementAt(surface)
		if !ok {
			return
		}
		c := g.playerCmd()
		g.unpauseForCommand(sim.CmdApply)
		if _, ok := g.w.Map.DoCommandFor(c, sim.CmdApply, sim.ExpenditureConstruction, func(f uint8) (int32, bool) {
			return g.w.Map.CreateWall(c, f, a)
		}); !ok {
			g.openError(i18n.Label(i18n.ErrorCantBuildThisHere), cmdErrorText(c))
		}
	}
}

// terraformToolDrag is the tab's toolDrag, each frame the button is held:
// clearing and painting go on under the cursor (unless an error window is
// open); adjusting land or water, every 16 screen pixels (at zoom 0; fewer
// zoomed out) the cursor moved up from where the button went down raise the
// selection a step, as many down lower it.
func (g *Game) terraformToolDrag(x, y int, onMap bool) {
	t := &g.tf
	water := t.win.Tab == ui.TerraformTabWater
	switch {
	case t.win.Tab == ui.TerraformTabClear:
		if g.errorWin == nil {
			g.clearLandApply()
		}
		return
	case t.win.Tab == ui.TerraformTabLand && t.paint:
		if g.errorWin == nil {
			g.paintLand()
		}
		return
	case t.win.Tab == ui.TerraformTabLand || water:
	default:
		return
	}
	if !onMap { // findAt: not the viewport
		return
	}
	step := 16 >> g.w.GetZoom() // zoom.applyInversedTo(16)
	if step == 0 {
		step = 1
	}
	raise, lower := g.raiseLand, g.lowerLand
	if water {
		raise, lower = g.raiseWater, g.lowerWater
	}
	if delta := y - t.dragLastY; delta <= -step {
		t.dragLastY -= step
		raise(sim.CmdApply)
	} else if delta >= step {
		t.dragLastY += step
		lower(sim.CmdApply)
	} else {
		return
	}
	t.raiseOK, t.lowerOK = false, false
}

// terraformToolUp is the tab's toolUp: the selection goes, the cursor back
// to the tab's (the tree and wall tabs have none).
func (g *Game) terraformToolUp() {
	t := &g.tf
	if t.win == nil {
		return
	}
	switch t.win.Tab {
	case ui.TerraformTabClear, ui.TerraformTabLand, ui.TerraformTabWater:
		g.w.SetMapSelection(false, 0, 0, 0, 0)
		t.arrow = false
	}
}

// terraformCommand runs a terraform command of the selection as the player
// (doCommand): with CmdApply a refusal opens the error window under the
// title set before it (setErrorTitle).
func (g *Game) terraformCommand(flags uint8, title int, run func(c *sim.Cmd, f uint8) (int32, bool)) (int32, bool) {
	c := g.playerCmd()
	g.unpauseForCommand(flags)
	cost, ok := g.w.Map.DoCommandFor(c, flags, sim.ExpenditureConstruction, func(f uint8) (int32, bool) { return run(c, f) })
	if !ok && flags&sim.CmdApply != 0 {
		g.openError(i18n.Label(title), cmdErrorText(c))
	}
	return cost, ok
}

// unpauseForCommand is doCommandForReal's start for a command that unpauses
// the game (kGameCommandDefinitions' unpausesGame: every terraform command),
// applied and no ghost, by the controlling company: the player's pause ends
// and the speed returns to normal. PARTIAL: the other windows' commands do
// not do this yet.
func (g *Game) unpauseForCommand(flags uint8) {
	if flags&sim.CmdApply != 0 && flags&(sim.CmdGhost|sim.CmdAiAllocated) == 0 {
		g.setSpeed(1)
	}
}

// raiseLand / lowerLand are AdjustLand's (0x00468D1D, 0x00468DFD): applied,
// the construction window's ghosts go first (sub_4A69DD); the mountain
// command in mountain mode, else raiseLand / lowerLand on the selection
// with its corner.
func (g *Game) raiseLand(flags uint8) (int32, bool) { return g.adjustLand(flags, 1) }
func (g *Game) lowerLand(flags uint8) (int32, bool) { return g.adjustLand(flags, -1) }

func (g *Game) adjustLand(flags uint8, adjustment int16) (int32, bool) {
	if flags&sim.CmdApply != 0 {
		g.removeConstructionGhosts()
	}
	a, cx, cy, corner := g.selectionArgs()
	title := i18n.ErrorCantRaiseLandHere
	if adjustment < 0 {
		title = i18n.ErrorCantLowerLandHere
	}
	m := &g.w.Map
	return g.terraformCommand(flags, title, func(c *sim.Cmd, f uint8) (int32, bool) {
		if g.tf.mountain {
			return m.LowerRaiseLandMountain(c, f, sim.MountainArgs{CX: cx, CY: cy, AX: int32(a[0]) * 32, AY: int32(a[1]) * 32,
				BX: int32(a[2]) * 32, BY: int32(a[3]) * 32, Adjustment: adjustment})
		}
		args := sim.LandArgs{CX: cx, CY: cy, AX: a[0], AY: a[1], BX: a[2], BY: a[3], Corner: corner}
		if adjustment > 0 {
			return m.RaiseLand(c, f, args)
		}
		return m.LowerLand(c, f, args)
	})
}

// removeConstructionGhosts is sub_4A69DD: with a construction window open,
// its ghosts removed.
func (g *Game) removeConstructionGhosts() {
	if g.consWin() != nil && g.r.ObjMgr != nil {
		g.w.RemoveConstructionGhosts(g.playerCmd())
	}
}

// raiseWater / lowerWater are AdjustWater's: the water commands on the
// selection (both under error_cant_raise_water_here, as upstream).
func (g *Game) raiseWater(flags uint8) (int32, bool) {
	if flags&sim.CmdApply != 0 {
		g.removeConstructionGhosts()
	}
	a, _, _, _ := g.selectionArgs()
	return g.terraformCommand(flags, i18n.ErrorCantRaiseWaterHere, func(c *sim.Cmd, f uint8) (int32, bool) {
		return g.w.Map.RaiseWater(c, f, a[0], a[1], a[2], a[3])
	})
}

func (g *Game) lowerWater(flags uint8) (int32, bool) {
	if flags&sim.CmdApply != 0 {
		g.removeConstructionGhosts()
	}
	a, _, _, _ := g.selectionArgs()
	return g.terraformCommand(flags, i18n.ErrorCantRaiseWaterHere, func(c *sim.Cmd, f uint8) (int32, bool) {
		return g.w.Map.LowerWater(c, f, a[0], a[1], a[2], a[3])
	})
}

// clearLand is ClearArea's clearLand: the clearLand command on the selection.
func (g *Game) clearLand(flags uint8) (int32, bool) {
	a, cx, cy, _ := g.selectionArgs()
	return g.terraformCommand(flags, i18n.ErrorCantClearEntireArea, func(c *sim.Cmd, f uint8) (int32, bool) {
		return g.w.Map.ClearLand(c, f, cx, cy, a[0], a[1], a[2], a[3])
	})
}

func (g *Game) clearLandApply() {
	if on, _, _ := g.w.MapSelection(); on {
		g.clearLand(sim.CmdApply)
	}
}

// paintLand is AdjustLand's paintLand: the selection's land changed to the
// chosen one (changeLandMaterial).
func (g *Game) paintLand() {
	t := &g.tf
	if on, _, _ := g.w.MapSelection(); !on {
		return
	}
	a, _, _, _ := g.selectionArgs()
	g.terraformCommand(sim.CmdApply, i18n.ErrorCantChangeLandType, func(c *sim.Cmd, f uint8) (int32, bool) {
		return g.w.Map.ChangeLandMaterial(c, f, a[0], a[1], a[2], a[3], t.lastLand)
	})
}

const ghostFlags = sim.CmdApply | sim.CmdNoErrorWindow | sim.CmdNoPayment | sim.CmdGhost

// treePlacementAt is getTreePlacementArgsFromCursor (0x004BD1D9): the
// surface under the cursor, the chosen tree; the quadrant under the cursor
// (its opposite: ^ 2), the tree colour, the tree rotation less the view's.
func (g *Game) treePlacementAt(surface func() (int32, int32, bool)) (sim.TreePlacement, bool) {
	t := &g.tf
	px, py, ok := surface()
	if !ok || t.win.Selected < 0 {
		return sim.TreePlacement{}, false
	}
	return sim.TreePlacement{X: px &^ 0x1F, Y: py &^ 0x1F, Type: uint8(t.win.Selected), Quadrant: quadrantFromPos(px, py) ^ 2,
		Colour: t.treeColour, Rotation: (t.treeRotation - g.w.Rotation()) & 3}, true
}

// removeTreeGhost is 0x004BD297's tree part.
func (g *Game) removeTreeGhost() {
	t := &g.tf
	if !t.ghostTree {
		return
	}
	t.ghostTree = false
	c := g.playerCmd()
	g.w.Map.DoCommandFor(c, ghostFlags, sim.ExpenditureConstruction, func(f uint8) (int32, bool) {
		return g.w.Map.RemoveTree(c, f, sim.TreeRemoval{X: t.ghostX, Y: t.ghostY, Z: int32(t.ghostBaseZ) * 4, Type: t.ghostType,
			Quadrant: t.ghostQuadrant, Rotation: t.ghostRotation})
	})
}

// placeTreeGhost is 0x004BD237: the ghost placed, remembered for its
// removal; the cost.
func (g *Game) placeTreeGhost(a sim.TreePlacement) (int32, bool) {
	t := &g.tf
	g.removeTreeGhost()
	c := g.playerCmd()
	var baseZ uint8
	cost, ok := g.w.Map.DoCommandFor(c, ghostFlags, sim.ExpenditureConstruction, func(f uint8) (int32, bool) {
		cost, z, ok := g.w.Map.CreateTreeCmd(c, f, a)
		baseZ = z
		return cost, ok
	})
	if ok {
		t.ghostTree, t.ghostX, t.ghostY, t.ghostRotation, t.ghostQuadrant, t.ghostType, t.ghostBaseZ, t.ghostImmediate =
			true, a.X, a.Y, a.Rotation, a.Quadrant, a.Type, baseZ, a.BuildImmediately
	}
	return cost, ok
}

// wallPlacementAt is getWallPlacementArgsFromCursor (0x004BD48E): the tile
// under the cursor and its edge nearest it, the chosen wall, black.
func (g *Game) wallPlacementAt(surface func() (int32, int32, bool)) (sim.WallPlacement, bool) {
	t := &g.tf
	px, py, ok := surface()
	if !ok || t.win.Selected < 0 {
		return sim.WallPlacement{}, false
	}
	return sim.WallPlacement{X: px &^ 0x1F, Y: py &^ 0x1F, Type: uint8(t.win.Selected), Rotation: sideFromPos(px, py)}, true
}

// removeWallGhost is 0x004BD297's wall part.
func (g *Game) removeWallGhost() {
	t := &g.tf
	if !t.ghostWall {
		return
	}
	t.ghostWall = false
	c := g.playerCmd()
	g.w.Map.DoCommandFor(c, ghostFlags, sim.ExpenditureConstruction, func(f uint8) (int32, bool) {
		return g.w.Map.RemoveWall(c, f, t.ghostX, t.ghostY, int32(t.ghostBaseZ)*4, t.ghostRotation)
	})
}

// placeWallGhost is 0x004BD4C8.
func (g *Game) placeWallGhost(a sim.WallPlacement) {
	t := &g.tf
	g.removeWallGhost()
	c := g.playerCmd()
	if _, ok := g.w.Map.DoCommandFor(c, ghostFlags, sim.ExpenditureConstruction, func(f uint8) (int32, bool) {
		return g.w.Map.CreateWall(c, f, a)
	}); ok {
		t.ghostWall, t.ghostX, t.ghostY, t.ghostRotation, t.ghostType, t.ghostBaseZ = true, a.X, a.Y, a.Rotation, a.Type, g.w.Map.LastPlacedWallBaseZ
	}
}
