package ui

import (
	"sort"
	"strconv"
	"strings"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
)

// The industry list (Ui/Windows/IndustryList.cpp): its list tab — Name,
// Status, Production transported, Production last month — on the shared
// table window, 759 wide (kWindowSize). PARTIAL: the second tab (fund new
// industries) and the resizing.

// IndustryListRow is one industry: IndustryList.cpp drawScroll's cells.
type IndustryListRow struct {
	ID          int
	Name        string
	Status      string
	Produces    bool   // canProduceCargo: the two production cells are drawn
	Transported uint8  // percent (getAverageTransportedCargo)
	LastMonth   uint32 // getProductionLastMonth's amount
	Unit        string // its cargo's unit name ({COMMA32} …)
}

func (r IndustryListRow) cells() []string {
	c := []string{r.Name, r.Status, "", ""}
	if r.Produces {
		c[2] = FillArgs(i18n.Label(i18n.ProductionTransportedPercent), strconv.Itoa(int(r.Transported)))
		c[3] = FillArgs(ObjectString(r.Unit), Comma32(int64(r.LastMonth)))
	}
	return c
}

// industryOrder is IndustryList.cpp getOrder: name and status by string,
// production transported and last month descending.
func industryOrder(mode int, a, b IndustryListRow) bool {
	switch mode {
	case 0:
		return a.Name < b.Name
	case 1:
		return a.Status < b.Status
	case 2:
		return b.Transported < a.Transported
	case 3:
		return b.LastMonth < a.LastMonth
	}
	return false
}

// NewIndustryList opens the industry list at (sw - 759, 30), as
// IndustryList::open; rows open their industry's window via OnRow.
func NewIndustryList(sw int, industries []IndustryListRow, titleColour, colour uint8) *TownListWindow {
	const w759 = 759
	status := countLabel(len(industries), i18n.StatusNumIndustriesSingular, i18n.StatusNumIndustriesPlural)
	w := &TownListWindow{X: sw - w759, Y: 30, TitleColour: titleColour, Colour: colour, Caption: i18n.Label(i18n.TitleIndustries), Tabs: 2,
		Columns: []TableColumn{{4, 200, i18n.Label(i18n.IndustryTableHeader), 0, 198}, {204, 240, i18n.Label(i18n.IndustryTableHeaderStatus), 200, 238},
			{444, 160, i18n.Label(i18n.IndustryTableHeaderProduction), 440, 138}, {604, 151, i18n.Label(i18n.IndustryTableHeaderProductionLastMonth), 600, 138}},
		Status: status, Hover: -1}
	w.setTableGeometry()
	w.W, w.List = w759, [4]int{3, 56, w759 - 6, 126} // onResize: the scroll view to the frame
	w.Rows = make([][]string, len(industries))
	w.Sort = func(mode int) {
		order := append([]IndustryListRow(nil), industries...)
		sort.SliceStable(order, func(i, j int) bool { return industryOrder(mode, order[i], order[j]) })
		w.RowIDs = w.RowIDs[:0]
		for i, r := range order {
			w.Rows[i] = r.cells()
			w.RowIDs = append(w.RowIDs, r.ID)
		}
	}
	w.Sort(0)
	return w
}

// NewIndustryChoice is an industry the player can fund (IndustryList.cpp
// updateBuildableIndustries): its object slot, name, cost (inflation-adjusted
// costFactor) and what drawIndustry draws — the building images in its first
// available colour, the first variation's parts stacked by their heights.
type NewIndustryChoice struct {
	ID          int
	Name        string
	Cost        int64
	Image       uint32 // buildingImageIds
	Colour      uint8
	Parts       []uint8 // getBuildingParts(0)
	PartHeights []uint8
}

// The new-industries tab (NewIndustries, 578×172): a scroll view of 112 px
// cells, five to a row.
const newIndustryCell = 112

// NewIndustryGrid is the industry list on its second tab (NewIndustries):
// "Fund New Industries", a grid of the industries the player can fund, each
// drawn by drawIndustry; the selected one (rowHover) inset; at the bottom its
// name and build cost ("No industry available" when there are none). A click
// selects (onScrollMouseDown) and calls OnSelect; the cell under the mouse
// (onScrollMouseOver: var_846) is lit and named. PARTIAL: its tooltip. The
// placement tool:
// cmd/goloco industrytool.go.
func NewIndustryGrid(sw int, choices []NewIndustryChoice, selected int, titleColour, colour uint8, onSelect func(id int)) *TownListWindow {
	const w578, h172 = 578, 172
	w := &TownListWindow{X: sw - 759, Y: 30, W: w578, H: h172, TitleColour: titleColour, Colour: colour,
		Caption: i18n.Label(i18n.TitleFundNewIndustries), Tabs: 2, CurrentTab: 1, RowHeight: newIndustryCell, Hover: -1}
	w.List = [4]int{3, 45, w578 - 4 - 3 + 1, h172 - 14 - 45 + 1}
	w.Rows = make([][]string, (len(choices)+4)/5) // getScrollSize: (4 + rowCount) / 5 rows
	sel, hov := -1, -1                            // rowHover (chosen), var_846 (under the mouse)
	for i, c := range choices {
		if c.ID == selected {
			sel = i
		}
	}
	w.RowDraw = func(rt *gfx8.RenderTarget, row, y int) {
		g1 := globalG1Font.g1
		sec, err := g1.GetPaletteMap(int(w.Colour))
		if err != nil {
			return
		}
		gfx8.FillRect(rt, 0, max(y, 0), rt.Width-1, min(y+newIndustryCell-1, rt.Height-1), gfx8.Shade(sec, 4)) // clearSingle(shade 4)
		for k := range 5 {
			i := row*5 + k
			if i >= len(choices) {
				break
			}
			x := k * newIndustryCell
			if i == sel {
				gfx8.FillRectInset(rt, x, y, x+newIndustryCell-1, y+newIndustryCell-1, gfx8.InsetColour{Map: sec}, gfx8.InsetColourLight|gfx8.InsetBorderInset)
			} else if i == hov {
				gfx8.FillRectInset(rt, x, y, x+newIndustryCell-1, y+newIndustryCell-1, gfx8.InsetColour{Map: sec}, gfx8.InsetColourLight)
			}
			drawIndustryPreview(rt, x+1, y+1, choices[i])
		}
	}
	cellAt := func(px, py int) int { // getRowIndex: -1 past the last
		l := w.List
		cx, cy := px-(w.X+l[0]+1), py-(w.Y+l[1]+1)+w.Scroll
		if i := cx/newIndustryCell + cy/newIndustryCell*5; cx >= 0 && cy >= 0 && cx < 5*newIndustryCell && i < len(choices) {
			return i
		}
		return -1
	}
	w.ExtraMouseOver = func(px, py int) { hov = cellAt(px, py) } // onScrollMouseOver: var_846
	w.ExtraClick = func(px, py int) bool {
		l := w.List
		if !w.in(l, px, py) || px >= w.X+l[0]+l[2]-L.ScrollbarSize-L.ScrollbarMargin {
			return false
		}
		cx, cy := px-(w.X+l[0]+1), py-(w.Y+l[1]+1)+w.Scroll
		if i := cx/newIndustryCell + cy/newIndustryCell*5; cx >= 0 && cy >= 0 && i < len(choices) { // getRowIndex
			sel = i
			if onSelect != nil {
				onSelect(choices[i].ID)
			}
		}
		return true
	}
	w.ExtraText = func(screen *ebiten.Image) {
		g1 := globalG1Font.g1
		black := g1.GetPalette()[paletteBlack0]
		y := w.Y + h172 - 13
		if len(choices) == 0 {
			DrawTextBold(screen, i18n.Label(i18n.NoIndustryAvailable), w.X+3, y, black)
			return
		}
		shown := hov // draw: the industry under the mouse (var_846), else the chosen one
		if shown < 0 {
			shown = sel
		}
		if shown < 0 {
			return
		}
		c := choices[shown]
		if hov < 0 && w.PlacementCost != nil { // the chosen one's ghost cost, unless it failed
			if cost, ok := w.PlacementCost(); ok {
				c.Cost = cost
			}
		}
		// build_cost "{COLOUR WINDOW_2}Cost: {COLOUR BLACK}{CURRENCY32}", right-aligned at width − 19
		sec, _ := g1.GetPaletteMap(int(w.Colour))
		label, _, _ := strings.Cut(i18n.Label(i18n.BuildCost), "{COLOUR BLACK}")
		amount := currency32(c.Cost)
		lw, _ := MeasureTextBold(label)
		aw, _ := MeasureTextBold(amount)
		x := w.X + 3 + w578 - 19 - lw - aw
		DrawTextBold(screen, label, x, y, g1.GetPalette()[gfx8.Shade(sec, 9)])
		DrawTextBold(screen, amount, x+lw, y, black)
		DrawTextBold(screen, c.Name, w.X+3, y, black)
	}
	return w
}

// drawIndustryPreview is IndustryObject::drawIndustry clipped to the cell's
// 110 × 110 (clipRenderTarget): from (56, 96) each part's image (part · 4 + 1)
// in the industry's colour, the next part raised by this one's height.
func drawIndustryPreview(rt *gfx8.RenderTarget, x, y int, c NewIndustryChoice) {
	g1 := globalG1Font.g1
	const size = 110
	clip := gfx8.NewRenderTarget(size, size)
	for yy := range size {
		for xx := range size {
			if sx, sy := x+xx, y+yy; sx >= 0 && sy >= 0 && sx < rt.Width && sy < rt.Height {
				clip.Bits[yy*size+xx] = rt.Bits[sy*rt.Width+sx]
			}
		}
	}
	pm, _ := g1.GetPaletteMap(int(c.Colour))
	px, py := 56, 96
	for _, part := range c.Parts {
		id := int(c.Image) + int(part)*4 + 1
		if id < len(g1.Elements) {
			if sp, err := gfx8.Decode(&g1.Elements[id]); err == nil {
				gfx8.DrawSprite(clip, sp, px, py, gfx8.OpFor(sp, pm != nil, false), pm)
			}
		}
		if int(part) < len(c.PartHeights) {
			py -= int(c.PartHeights[part])
		}
	}
	for yy := range size {
		for xx := range size {
			if sx, sy := x+xx, y+yy; sx >= 0 && sy >= 0 && sx < rt.Width && sy < rt.Height {
				rt.Bits[sy*rt.Width+sx] = clip.Bits[yy*size+xx]
			}
		}
	}
}
