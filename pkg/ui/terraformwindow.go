package ui

import (
	"strings"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
)

// The Terraform window — a port of Ui/Windows/TerraForm.cpp's widgets and
// drawing: the common frame (caption, close button, panel, five tabs) and
// the tabs Clear Area, Adjust Land, Adjust Water, Plant Trees and Build
// Walls. The tools (hover, ghosts, drag, the game commands) are cmd/goloco's
// terraform.go, which fills the window's state before each draw and takes
// its clicks through the On… callbacks.

// The tabs, in widget order (tab_clear_area …).
const (
	TerraformTabClear = iota
	TerraformTabLand
	TerraformTabWater
	TerraformTabTrees
	TerraformTabWalls
)

// tabInformationByTabOffset's window sizes (Adjust Land: 161 × 115 in the
// editor, its onResize).
var terraformTabSize = [5][2]int{{161, 105}, {161, 140}, {161, 115}, {634, 162}, {418, 108}}

// The tree and wall grids (PlantTrees / BuildWalls kRowHeight, kColumnWidth,
// kColsPerRow).
var terraformGrid = [5]struct{ rowH, colW, cols int }{3: {102, 66, 9}, 4: {48, 40, 10}}

const (
	imgToolArea         = 2447 // ImageIds::tool_area (+ the size, 1–10)
	imgDecreaseToolArea = 2443
	imgIncreaseToolArea = 2445
	imgSlopeUp          = 2351 // construction_slope_up
	imgPaintbrush       = 2367
	imgPlantClusterSel  = 2380 // plant_cluster_selected_tree
	imgPlantClusterRand = 2381
)

// TerraformCell is one tree or wall of a grid: its object id (rowInfo), name
// and image (a tree's by drawTreeThumb for the window's rotation; a wall's
// sprite), and a tree's colour options (TreeObject::colours; 0 none).
type TerraformCell struct {
	ID      int
	Name    string
	Image   int
	Colours uint32
	Cost    int64 // a tree's build cost (inflation adjusted)
	// Rotatable: more than one rotation (the rotate button shows).
	Rotatable bool
}

// TerraformWindow is the window. The game sets the state fields (the tool
// sizes, costs, modes, the grids' cells) before each draw.
type TerraformWindow struct {
	WinBase
	Tab         int
	TitleColour uint8 // the owner company's colour
	Colour      uint8 // InterfaceSkin windowTerraFormColour
	FrameNo     int
	Editor      bool // SceneManager::isEditorMode
	// drawTabs' images: the skin's (bulldozer, plant trees, build walls),
	// the default land's toolbar_terraform_land, the water's
	// kToolbarTerraformWater.
	SkinImage, LandTabImage, WaterTabImage int

	// The tool-size tabs: _adjustToolSize, the costs (ok false: kFailure).
	ToolSize             int
	RaiseCost, LowerCost int64
	RaiseOK, LowerOK     bool
	// Adjust Land: isMountainMode, isPaintMode, the chosen land's icon
	// (mapPixelImage + landscape_generator_tile_icon) and its map pixel
	// colour (the paintbrush's secondary).
	MountainMode, PaintMode bool
	LandIcon                int
	LandPixelColour         uint8

	// The grids: the cells, the chosen one's id (rowHover) and the one under
	// the mouse (var_846; −1 none).
	Cells    []TerraformCell
	Selected int
	Hover    int
	Scroll   int
	hold     L.VScrollHold
	// expandContentCounter: updates the mouse has rested on the list.
	expand int
	// Plant Trees: _treeRotation, _treeColour, _treeClusterType (0 none,
	// 1 selected, 2 random), _lastTreeCost (ok false: kFailure).
	TreeRotation, TreeColour uint8
	Cluster                  int
	TreeCost                 int64
	TreeCostOK               bool

	held   int // the tool-size button held: −1 decrease, 1 increase
	heldAt int // frames it has been held
	mx, my int // the mouse over the window (window coordinates; far away: off it)

	// OnTab is Common::switchTab; OnToolSize the decrease / increase buttons
	// (delta ∓1); OnMountain, OnPaint the mode buttons; OnLandMaterial the
	// land button (the widget's screen rectangle, for the dropdown);
	// OnSelect a cell clicked (scrollMouseDown); OnRotate, OnColour (the
	// widget's rectangle), OnCluster (1 selected, 2 random) the tree buttons.
	OnTab          func(tab int)
	OnToolSize     func(delta int)
	OnMountain     func()
	OnPaint        func()
	OnLandMaterial func(x, y, w, h int)
	OnSelect       func(id int)
	OnRotate       func()
	OnColour       func(x, y, w, h int)
	OnCluster      func(kind int)
}

// NewTerraformWindow is Terraform::open's createWindow: at (screen width −
// 634, 30), lighter frame, on the Plant Trees tab's size; SwitchTab then
// gives it its tab.
func NewTerraformWindow(sw int) *TerraformWindow {
	s := terraformTabSize[TerraformTabTrees]
	return &TerraformWindow{WinBase: WinBase{WinType: WindowTypeTerraform, X: sw - s[0], Y: 30, W: s[0], H: s[1],
		MinW: s[0], MinH: s[1], MaxW: s[0], MaxH: s[1]}, Tab: TerraformTabTrees, Selected: -1, Hover: -1, mx: -1 << 20, my: -1 << 20}
}

// SwitchTab is Common::switchTab's window part: the tab, frameNo 0, the
// tab's fixed size, the scroll and hover reset, moved inside the screen.
func (w *TerraformWindow) SwitchTab(tab int) {
	w.Tab, w.FrameNo, w.Scroll, w.Hover, w.expand, w.held = tab, 0, 0, -1, 0, 0
	s := w.tabSize()
	w.SetFixedSize(s[0], s[1])
	w.W, w.H = s[0], s[1]
	w.MoveInsideScreenEdges()
}

func (w *TerraformWindow) tabSize() [2]int {
	s := terraformTabSize[w.Tab]
	if w.Tab == TerraformTabLand && w.Editor {
		s[1] = 115
	}
	return s
}

func (w *TerraformWindow) isGrid() bool {
	return w.Tab == TerraformTabTrees || w.Tab == TerraformTabWalls
}

// list is the scroll view's rectangle (prepareDraw: trees {3, 45} to width
// − 26, height − 14; walls {2, 45} to width − 4, height − 14).
func (w *TerraformWindow) list() [4]int {
	if w.Tab == TerraformTabTrees {
		return [4]int{3, 45, w.W - 26 - 3 + 1, w.H - 14 - 45 + 1}
	}
	return [4]int{2, 45, w.W - 4 - 2 + 1, w.H - 14 - 45 + 1}
}

// contentHeight is getScrollSize: the rows (at least one) × the row height.
func (w *TerraformWindow) contentHeight() int {
	g := terraformGrid[w.Tab]
	if g.cols == 0 {
		return 0
	}
	return max((len(w.Cells)+g.cols-1)/g.cols, 1) * g.rowH
}

// FocusSelected is updateActiveThumb: the list scrolled to the chosen
// cell's row.
func (w *TerraformWindow) FocusSelected() {
	g := terraformGrid[w.Tab]
	if g.cols == 0 {
		return
	}
	w.Scroll = 0
	for i, c := range w.Cells {
		if c.ID == w.Selected {
			w.Scroll = i / g.cols * g.rowH
		}
	}
	w.clampScroll()
}

func (w *TerraformWindow) clampScroll() {
	w.Scroll = min(max(w.Scroll, 0), max(w.contentHeight()-w.list()[3], 0))
}

// The widgets of the tool-size tabs (window coordinates: x, y, w, h).
var (
	tfToolArea     = [4]int{49, 45, 64, 44}
	tfDecrease     = [4]int{50, 46, 16, 16}
	tfIncrease     = [4]int{96, 72, 16, 16}
	tfMountain     = [4]int{57, 92, 24, 24}
	tfPaint        = [4]int{83, 92, 24, 24}
	tfLandMaterial = [4]int{112, 94, 20, 20}
)

// treeButton is the i-th of the Plant Trees tab's buttons (rotate, colour,
// cluster selected, cluster random): width − 25 … width − 2, 24 high from
// y 46.
func (w *TerraformWindow) treeButton(i int) [4]int { return [4]int{w.W - 25, 46 + 24*i, 24, 24} }

func (w *TerraformWindow) in(r [4]int, px, py int) bool {
	return px >= w.X+r[0] && px < w.X+r[0]+r[2] && py >= w.Y+r[1] && py < w.Y+r[1]+r[3]
}

func (w *TerraformWindow) selectedCell() *TerraformCell {
	for i := range w.Cells {
		if w.Cells[i].ID == w.Selected {
			return &w.Cells[i]
		}
	}
	return nil
}

// cellAt is getRowIndex for a screen point on the list's content: the
// cell's index, −1 past the last or off the content.
func (w *TerraformWindow) cellAt(px, py int) int {
	g, l := terraformGrid[w.Tab], w.list()
	if g.cols == 0 || !w.in(l, px, py) {
		return -1
	}
	cx, cy := px-(w.X+l[0]+1), py-(w.Y+l[1]+1)+w.Scroll
	if cx < 0 || cy < 0 || cx >= g.cols*g.colW || w.onScrollbar(px) {
		return -1
	}
	if i := cx/g.colW + cy/g.rowH*g.cols; i < len(w.Cells) {
		return i
	}
	return -1
}

func (w *TerraformWindow) onScrollbar(px int) bool {
	l := w.list()
	return w.contentHeight() > l[3] && px >= w.X+l[0]+l[2]-L.ScrollbarSize-L.ScrollbarMargin
}

// HandleClick is the window's onMouseDown / onMouseUp / scrollMouseDown: the
// close button, a tab, then the tab's widgets.
// clickable(TerraForm.cpp): close_button tab_clear_area tab_adjust_land tab_adjust_water tab_plant_trees tab_build_walls decrease_area increase_area mountain_mode paint_mode land_material scrollview rotate_object object_colour plant_cluster_selected plant_cluster_random
func (w *TerraformWindow) HandleClick(px, py int) {
	if w.in([4]int{w.W - 15, 2, 13, 13}, px, py) {
		w.Close()
		return
	}
	if i := tabAt(tabLayout(5, nil), px-w.X, py-w.Y); i >= 0 {
		if w.OnTab != nil {
			w.OnTab(i)
		}
		return
	}
	call := func(f func()) {
		if f != nil {
			f()
		}
	}
	switch w.Tab {
	case TerraformTabClear, TerraformTabLand, TerraformTabWater:
		switch {
		case w.in(tfDecrease, px, py):
			w.held, w.heldAt = -1, 0
			if w.OnToolSize != nil {
				w.OnToolSize(-1)
			}
		case w.in(tfIncrease, px, py):
			w.held, w.heldAt = 1, 0
			if w.OnToolSize != nil {
				w.OnToolSize(1)
			}
		case w.Tab == TerraformTabLand && w.in(tfMountain, px, py):
			call(w.OnMountain)
		case w.Tab == TerraformTabLand && w.in(tfPaint, px, py):
			call(w.OnPaint)
		case w.Tab == TerraformTabLand && w.PaintMode && w.in(tfLandMaterial, px, py):
			if r := tfLandMaterial; w.OnLandMaterial != nil {
				w.OnLandMaterial(w.X+r[0], w.Y+r[1], r[2], r[3])
			}
		}
	case TerraformTabTrees, TerraformTabWalls:
		l := w.list()
		if w.in(l, px, py) && w.onScrollbar(px) { // the scrollbar (stateScrollLeft)
			w.Scroll = w.hold.Begin(py-(w.Y+l[1]), py, l[3], w.contentHeight(), w.Scroll)
			return
		}
		if i := w.cellAt(px, py); i >= 0 { // scrollMouseDown
			w.Selected, w.expand = w.Cells[i].ID, -16
			if w.OnSelect != nil {
				w.OnSelect(w.Selected)
			}
			return
		}
		if w.Tab != TerraformTabTrees {
			return
		}
		sel := w.selectedCell()
		switch {
		case sel != nil && sel.Rotatable && w.in(w.treeButton(0), px, py):
			call(w.OnRotate)
		case sel != nil && sel.Colours != 0 && w.in(w.treeButton(1), px, py):
			if r := w.treeButton(1); w.OnColour != nil {
				w.OnColour(w.X+r[0], w.Y+r[1], r[2], r[3])
			}
		case w.in(w.treeButton(2), px, py):
			if w.OnCluster != nil {
				w.OnCluster(1)
			}
		case w.in(w.treeButton(3), px, py):
			if w.OnCluster != nil {
				w.OnCluster(2)
			}
		}
	}
}

// Rotate is Terraform::rotate: on the Plant Trees tab with the rotate
// button showing, a press of it; false otherwise.
func (w *TerraformWindow) Rotate() bool {
	if sel := w.selectedCell(); w.Tab == TerraformTabTrees && sel != nil && sel.Rotatable && w.OnRotate != nil {
		w.OnRotate()
		return true
	}
	return false
}

// Update is the window's share of a frame: the mouse over it (far away when
// it is on another window or none), the left button, the right one. It runs
// the held tool-size buttons (holdableWidgets: repeating while held), the
// list's scrollbar and hover (scrollMouseOver / onHandleInputBegin), and
// the grid tabs' onUpdate — the list growing to show every row once the
// mouse has rested on it for 8 updates (to at most 562 high; 358 on a
// screen under 600), and back to its size when the mouse leaves the window.
func (w *TerraformWindow) Update(px, py int, over, left, right bool, screenH int) {
	w.FrameNo++
	w.mx, w.my = -1<<20, -1<<20
	if over {
		w.mx, w.my = px-w.X, py-w.Y
	}
	if w.held != 0 {
		r := tfDecrease
		if w.held > 0 {
			r = tfIncrease
		}
		if !left || !w.in(r, px, py) {
			w.held = 0
		} else if w.heldAt++; w.heldAt >= 16 && w.heldAt%4 == 0 && w.OnToolSize != nil { // WindowManager: a held widget repeats
			w.OnToolSize(w.held)
		}
	}
	if !w.isGrid() {
		return
	}
	l := w.list()
	scrolling := w.hold.PressedPart() != ""
	w.Scroll = w.hold.Continue(left, over && w.in(l, px, py) && w.onScrollbar(px), py-(w.Y+l[1]), py, l[3], w.contentHeight(), w.Scroll)
	w.Hover = -1
	if i := w.cellAt(px, py); over && i >= 0 {
		w.Hover = w.Cells[i].ID
	}
	if right {
		return
	}
	size := terraformTabSize[w.Tab]
	if !over {
		w.expand = 0
		if !scrolling {
			w.setHeight(size[1])
		}
		return
	}
	rx, ry := px-w.X, py-w.Y
	onTab := tabAt(tabLayout(5, nil), rx, ry) >= 0
	onButton := false
	if w.Tab == TerraformTabTrees {
		for i := range 4 {
			onButton = onButton || w.in(w.treeButton(i), px, py)
		}
	}
	if (ry < 42 || rx+26 <= w.W) && (onTab || w.in(l, px, py) || onButton) { // a widget past the panel
		w.expand++
		if w.expand >= 8 {
			h := min(w.contentHeight()-1+60, 562)
			if screenH < 600 {
				h = min(h, 358)
			}
			w.setHeight(h)
		} else if !scrolling {
			w.setHeight(size[1])
		}
	}
}

func (w *TerraformWindow) setHeight(h int) {
	if w.H == h {
		return
	}
	w.SetFixedSize(w.W, h)
	w.H = h
	w.clampScroll()
}

// HandleWheel scrolls a grid tab's list by a row.
func (w *TerraformWindow) HandleWheel(dy float64) {
	if g := terraformGrid[w.Tab]; g.cols != 0 {
		if dy > 0 {
			w.Scroll -= g.rowH
		} else if dy < 0 {
			w.Scroll += g.rowH
		}
		w.clampScroll()
	}
}

// WidgetTooltip is the tooltip of the widget at a screen point ("" none).
func (w *TerraformWindow) WidgetTooltip(px, py int) string {
	if w.in([4]int{w.W - 15, 2, 13, 13}, px, py) {
		return i18n.Str(i18n.TooltipCloseWindow)
	}
	if i := tabAt(tabLayout(5, nil), px-w.X, py-w.Y); i >= 0 {
		return i18n.Str([5]int{i18n.TooltipClearLand, i18n.TooltipAdjustLand, i18n.TooltipAdjustWater, i18n.TooltipPlantTrees, i18n.TooltipBuildWalls}[i])
	}
	tips := map[int][3]int{
		TerraformTabClear: {i18n.TooltipDecreaseClearArea, i18n.TooltipIncreaseClearArea, i18n.TooltipClearArea},
		TerraformTabLand:  {i18n.TooltipDecreaseAdjustLandArea, i18n.TooltipIncreaseAdjustLandArea, i18n.TooltipAdjustLandTool},
		TerraformTabWater: {i18n.TooltipDecreaseAdjustWaterArea, i18n.TooltipIncreaseAdjustWaterArea, i18n.TooltipAdjustWaterTool},
	}
	if t, ok := tips[w.Tab]; ok {
		switch {
		case w.in(tfDecrease, px, py):
			return i18n.Str(t[0])
		case w.in(tfIncrease, px, py):
			return i18n.Str(t[1])
		case w.in(tfToolArea, px, py):
			return i18n.Str(t[2])
		case w.Tab == TerraformTabLand && w.in(tfMountain, px, py):
			return i18n.Str(i18n.MountainModeTooltip)
		case w.Tab == TerraformTabLand && w.in(tfPaint, px, py):
			return i18n.Str(i18n.TooltipPaintLandscapeTool)
		}
		return ""
	}
	if w.Tab == TerraformTabTrees {
		sel := w.selectedCell()
		switch {
		case sel != nil && sel.Rotatable && w.in(w.treeButton(0), px, py):
			return i18n.Str(i18n.RotateObject90)
		case sel != nil && sel.Colours != 0 && w.in(w.treeButton(1), px, py):
			return i18n.Str(i18n.TooltipObjectColour)
		case w.in(w.treeButton(2), px, py):
			return i18n.Str(i18n.PlantClusterSelectedTree)
		case w.in(w.treeButton(3), px, py):
			return i18n.Str(i18n.PlantClusterRandomTree)
		}
	}
	return ""
}

// GripBlocked: the window has no resize grip.
func (w *TerraformWindow) GripBlocked(px, py int) bool { return true }

// Draw draws the window: the frame, the tabs (Common::drawTabs), the tab's
// widgets and its draw().
func (w *TerraformWindow) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	pal := g1.GetPalette()
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	black, win2 := pal[paletteBlack0], pal[gfx8.Shade(sec, 9)]
	sprite := func(id int) *gfx8.Sprite {
		if id <= 0 || id >= len(g1.Elements) {
			return nil
		}
		sp, err := gfx8.Decode(&g1.Elements[id])
		if err != nil {
			return nil
		}
		return sp
	}
	hovered := func(r [4]int) bool { return w.mx >= r[0] && w.mx < r[0]+r[2] && w.my >= r[1] && w.my < r[1]+r[3] }
	var list [4]int
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, true)
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 41, w.W-1, w.H-1, gfx8.InsetColour{Map: sec}, 0)
		if sp := sprite(imgCloseButton); sp != nil {
			gfx8.DrawSpriteNoise(rt, sp, w.W-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		// drawTabs: the bulldozer, the default land's and the water's toolbar
		// images (the water animated while its tab is current), plant trees,
		// build walls.
		icons := [5]int{w.SkinImage + skinToolbarMenuBulldozer, w.LandTabImage, w.WaterTabImage, w.SkinImage + skinToolbarMenuPlantTrees, w.SkinImage + skinToolbarMenuBuildWalls}
		if w.Tab == TerraformTabWater {
			icons[2] += (w.FrameNo / 2) % 16
		}
		for i, x := range tabLayout(5, nil) {
			icon := icons[i]
			if icon <= 0 || icon >= len(g1.Elements) {
				icon = tabNoIcon
			}
			drawTab(rt, x, 15, i == w.Tab, sec, icon, nil)
		}
		switch w.Tab {
		case TerraformTabClear, TerraformTabLand, TerraformTabWater:
			w.drawToolArea(rt, sec, sprite)
			if w.Tab == TerraformTabLand {
				drawImageButton(rt, tfMountain, imgSlopeUp, w.Colour, w.MountainMode, hovered(tfMountain), false)
				// the paintbrush: recolour2(paintbrush, white, the land's pixel colour)
				flags := 0
				if w.PaintMode {
					flags = gfx8.InsetBorderInset
				}
				if hovered(tfPaint) {
					gfx8.FillRectInset(rt, tfPaint[0], tfPaint[1], tfPaint[0]+tfPaint[2]-1, tfPaint[1]+tfPaint[3]-1, translucentInset(w.Colour), flags)
				} else if w.PaintMode {
					gfx8.FillRectInset(rt, tfPaint[0], tfPaint[1], tfPaint[0]+tfPaint[2]-1, tfPaint[1]+tfPaint[3]-1, gfx8.InsetColour{Map: sec}, flags)
				}
				if sp := sprite(imgPaintbrush); sp != nil {
					m, _ := g1.GetPaletteMap2(colourWhite, int(w.LandPixelColour))
					gfx8.DrawSpriteNoise(rt, sp, tfPaint[0], tfPaint[1], gfx8.OpFor(sp, m != nil, false), m, nil)
				}
				if w.PaintMode {
					if sp := sprite(w.LandIcon); sp != nil {
						gfx8.DrawSpriteNoise(rt, sp, tfLandMaterial[0], tfLandMaterial[1], gfx8.OpFor(sp, false, false), nil, nil)
					}
				}
			}
		case TerraformTabTrees, TerraformTabWalls:
			list = w.list()
			w.drawGrid(rt, sec, sprite)
			if w.Tab == TerraformTabTrees {
				sel := w.selectedCell()
				if sel != nil && sel.Rotatable {
					drawImageButton(rt, w.treeButton(0), imgRotateObject, w.Colour, false, hovered(w.treeButton(0)), false)
				}
				if sel != nil && sel.Colours != 0 { // ColourButton: the swatch in the tree's colour
					if sp := sprite(imgColourSwatchRecolourable); sp != nil {
						m, _ := g1.GetPaletteMap(int(w.TreeColour))
						r := w.treeButton(1)
						gfx8.DrawSpriteNoise(rt, sp, r[0], r[1], gfx8.OpFor(sp, m != nil, false), m, nil)
					}
				}
				drawImageButton(rt, w.treeButton(2), imgPlantClusterSel, w.Colour, w.Cluster == 1, hovered(w.treeButton(2)), false)
				drawImageButton(rt, w.treeButton(3), imgPlantClusterRand, w.Colour, w.Cluster == 2, hovered(w.treeButton(3)), false)
			}
		}
	})
	if w.isGrid() && w.contentHeight() > list[3] {
		drawVScrollArrows(screen, w.X+list[0], w.Y+list[1], list[2], list[3])
	}
	caption := [5]int{i18n.ClearArea, i18n.TitleAdjustLand, i18n.TitleAdjustWater, i18n.TitlePlantTrees, i18n.TitleBuildWalls}[w.Tab]
	DrawWindowCaptionColour(screen, w.X+1, w.Y+1, w.W-2, i18n.Label(caption), int(w.TitleColour))

	// two-colour strings centred: "{COLOUR WINDOW_2}label{COLOUR BLACK}amount"
	centred := func(label, amount string, cx, y int) {
		lw, _ := MeasureTextBold(label)
		aw, _ := MeasureTextBold(amount)
		x := cx - (lw+aw)/2
		DrawTextBold(screen, label, x, y, win2)
		DrawTextBold(screen, amount, x+lw, y, black)
	}
	costLabel := func(id int) string { // the string up to its {COLOUR BLACK}
		label, _, _ := strings.Cut(i18n.Label(id), "{COLOUR BLACK}")
		return label
	}
	midX, midY, bottom := tfToolArea[0]+(tfToolArea[2]-1)/2, tfToolArea[1]+(tfToolArea[3]-1)/2, tfToolArea[1]+tfToolArea[3]-1
	switch w.Tab {
	case TerraformTabClear, TerraformTabLand, TerraformTabWater:
		// the size as a number when no sprite fits (Adjust Land: from 10, the
		// others above 10): tile_inspector_coord, "{COLOUR WINDOW_2}{INT16}"
		if w.ToolSize > 10 || w.Tab == TerraformTabLand && w.ToolSize == 10 {
			n := Comma16(w.ToolSize)
			nw, _ := MeasureTextBold(n)
			DrawTextBold(screen, n, w.X+midX-nw/2, w.Y+midY-5, win2)
		}
		y := bottom + 5
		if w.Tab == TerraformTabLand {
			y = bottom + 28
		}
		if w.Tab == TerraformTabClear { // clear_land_cost
			if w.RaiseOK && w.RaiseCost != 0 {
				centred(costLabel(i18n.ClearLandCost), currency32(w.RaiseCost), w.X+midX, w.Y+y)
			}
			break
		}
		// increase_height_cost / decrease_height_cost: "▲  cost", "▼  cost"
		if w.RaiseOK && w.RaiseCost != 0 {
			centred(string(rune(160)), "  "+currency32(w.RaiseCost), w.X+midX, w.Y+y)
		}
		if w.LowerOK && w.LowerCost != 0 {
			centred(string(rune(170)), "  "+currency32(w.LowerCost), w.X+midX, w.Y+y+10)
		}
	case TerraformTabTrees, TerraformTabWalls:
		// the cell under the mouse (var_846), else the chosen one: its name at
		// (3, height − 13), clipped to width − 19; a tree's cost right-aligned
		// at width − 14 (build_cost) — the ghost's when it is the chosen tree.
		var shown *TerraformCell
		for i := range w.Cells {
			if w.Cells[i].ID == w.Hover {
				shown = &w.Cells[i]
			}
		}
		hover := shown != nil
		if shown == nil {
			shown = w.selectedCell()
		}
		if shown == nil {
			return
		}
		y := w.Y + w.H - 13
		if w.Tab == TerraformTabTrees && !w.Editor {
			cost := shown.Cost
			if !hover && w.TreeCostOK {
				cost = w.TreeCost
			}
			label, amount := costLabel(i18n.BuildCost), currency32(cost)
			lw, _ := MeasureTextBold(label)
			aw, _ := MeasureTextBold(amount)
			x := w.X + 3 + w.W - 17 - lw - aw
			DrawTextBold(screen, label, x, y, win2)
			DrawTextBold(screen, amount, x+lw, y, black)
		}
		DrawTextBold(screen, ClipString(shown.Name, 0, w.W-19), w.X+3, y, black)
	}
}

// InterfaceSkin::ImageIds of drawTabs (toolbar_menu_plant_trees, _bulldozer,
// _build_walls).
const (
	skinToolbarMenuPlantTrees = 179
	skinToolbarMenuBulldozer  = 180
	skinToolbarMenuBuildWalls = 187
)

// drawToolArea is the tool-size tabs' tool_area (a Wt3Widget, activated: an
// inset frame) with the size's grid image — tool_area + size up to 10
// (Adjust Land: below 10; in mountain mode the plain grid under it, drawn
// twice 4 px either side for an even size, and no size image at 1) —, and
// the decrease / increase buttons (ImageButtonAlt: the image in the window's
// colour, + 1 while held).
func (w *TerraformWindow) drawToolArea(rt *gfx8.RenderTarget, sec []uint8, sprite func(int) *gfx8.Sprite) {
	a := tfToolArea
	gfx8.FillRectInset(rt, a[0], a[1], a[0]+a[2]-1, a[1]+a[3]-1, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset)
	img := func(id, dx int) {
		if sp := sprite(id); sp != nil {
			gfx8.DrawSpriteNoise(rt, sp, a[0]+dx, a[1], gfx8.OpFor(sp, true, false), sec, nil)
		}
	}
	switch {
	case w.Tab != TerraformTabLand:
		if w.ToolSize <= 10 {
			img(imgToolArea+w.ToolSize, 0)
		}
	case w.ToolSize < 10:
		if w.MountainMode {
			if w.ToolSize&1 == 0 {
				img(imgToolArea, -4)
				img(imgToolArea, 4)
			} else {
				img(imgToolArea, 0)
			}
		}
		if !w.MountainMode || w.ToolSize > 1 {
			img(imgToolArea+w.ToolSize, 0)
		}
	}
	for _, b := range []struct {
		r        [4]int
		image, d int
	}{{tfDecrease, imgDecreaseToolArea, -1}, {tfIncrease, imgIncreaseToolArea, 1}} {
		id := b.image
		if w.held == b.d {
			id++
		}
		if sp := sprite(id); sp != nil {
			gfx8.DrawSpriteNoise(rt, sp, b.r[0], b.r[1], gfx8.OpFor(sp, true, false), sec, nil)
		}
	}
}

// drawGrid is the scroll view and drawScroll: the content cleared to the
// window colour's shade 3, the chosen cell inset (light, border inset), the
// one under the mouse raised (light), each cell's image clipped to the cell
// less a pixel all round — a tree at (32, 96) in its colour (the tree
// colour when chosen or under the mouse, else its last colour option), a
// wall at (34, 28).
func (w *TerraformWindow) drawGrid(rt *gfx8.RenderTarget, sec []uint8, sprite func(int) *gfx8.Sprite) {
	g1 := globalG1Font.g1
	g, l := terraformGrid[w.Tab], w.list()
	cx, cy, cw, ch, _ := drawVScrollViewPressed(rt, l[0], l[1], l[2], l[3], gfx8.InsetColour{Map: sec}, w.contentHeight(), w.Scroll, w.hold.PressedPart())
	gfx8.FillRect(rt, cx, cy, cx+cw-1, cy+ch-1, gfx8.Shade(sec, 3))
	view := gfx8.NewRenderTarget(cw, ch) // the scroll view's own target: what it draws is clipped to it
	copyRect(view, rt, cx, cy)
	for i, c := range w.Cells {
		x, y := i%g.cols*g.colW, i/g.cols*g.rowH-w.Scroll
		if y+g.rowH <= 0 || y >= ch {
			continue
		}
		if c.ID == w.Selected {
			gfx8.FillRectInset(view, x, y, x+g.colW-1, y+g.rowH-1, gfx8.InsetColour{Map: sec}, gfx8.InsetColourLight|gfx8.InsetBorderInset)
		} else if c.ID == w.Hover {
			gfx8.FillRectInset(view, x, y, x+g.colW-1, y+g.rowH-1, gfx8.InsetColour{Map: sec}, gfx8.InsetColourLight)
		}
		sp := sprite(c.Image)
		if sp == nil {
			continue
		}
		clip := gfx8.NewRenderTarget(g.colW-2, g.rowH-2)
		copyRect(clip, view, x+1, y+1)
		if w.Tab == TerraformTabTrees {
			var m []uint8
			if c.Colours != 0 {
				colour := w.TreeColour
				if c.ID != w.Selected && c.ID != w.Hover {
					colour = uint8(bitScanReverse(c.Colours))
				}
				m, _ = g1.GetPaletteMap(int(colour))
			}
			gfx8.DrawSpriteNoise(clip, sp, 32, 96, gfx8.OpFor(sp, m != nil, false), m, nil)
		} else {
			gfx8.DrawSpriteNoise(clip, sp, g.colW-6, g.rowH-20, gfx8.OpFor(sp, false, false), nil, nil)
		}
		pasteRect(view, clip, x+1, y+1)
	}
	pasteRect(rt, view, cx, cy)
}

// bitScanReverse is Numerics::bitScanReverse (0 for no bit: Colour::black).
func bitScanReverse(v uint32) int {
	n := 0
	for i := range 32 {
		if v&(1<<i) != 0 {
			n = i
		}
	}
	return n
}

// copyRect fills dst from src's rectangle at (x, y); pasteRect writes it
// back. Pixels outside src are left.
func copyRect(dst, src *gfx8.RenderTarget, x, y int) {
	for yy := range dst.Height {
		for xx := range dst.Width {
			if sx, sy := x+xx, y+yy; sx >= 0 && sy >= 0 && sx < src.Width && sy < src.Height {
				dst.Bits[yy*dst.Width+xx] = src.Bits[sy*src.Width+sx]
			}
		}
	}
}

func pasteRect(dst, src *gfx8.RenderTarget, x, y int) {
	for yy := range src.Height {
		for xx := range src.Width {
			if dx, dy := x+xx, y+yy; dx >= 0 && dy >= 0 && dx < dst.Width && dy < dst.Height {
				dst.Bits[dy*dst.Width+dx] = src.Bits[yy*src.Width+xx]
			}
		}
	}
}
