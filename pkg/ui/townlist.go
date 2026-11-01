package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
)

// TownList — a port of Ui/Windows/TownList.cpp, the town list tab: 600×197 at
// (screen width - 600, 30), skin windowTitlebarColour / windowColour; caption,
// close button, panel, four tabs, the sortable column headers (Name, Type,
// Population, Stations), a ScrollView of 10 px rows and the "N towns" status.
// PARTIAL: only the list tab (build-town/buildings tabs not ported), tab
// icons, header sorting (rows keep the scenario order), window resizing and
// the resize handle.

const (
	tlW, tlH    = 600, 197
	tlRowHeight = 10
)

// TownRow is one town.
type TownRow struct {
	Name, Type           string
	Population, Stations int
}

// TableColumn is one column of a table list window: its header widget
// (x, width, label) and where/how wide the row text is drawn.
type TableColumn struct {
	HeaderX, HeaderW int
	Label            string
	TextX, TextW     int
}

// TownListWindow is a 600×197 table list window (TownList.cpp, StationList.cpp
// share the layout): caption, close, panel, tabs, column headers, a scroll
// view of 10 px rows and a status line.
type TownListWindow struct {
	X, Y        int
	TitleColour uint8
	Colour      uint8
	Caption     string
	Tabs        int
	Columns     []TableColumn
	Rows        [][]string
	Status      string
	Hover       int
	Scroll      int
	content     [4]int
	OnClose     func()

	// Geometry: window size, row height, the scroll view rect, the status
	// label's y and the text's y offset in a row (VehicleList: +2).
	W, H, RowHeight, StatusY, TextDY int
	List                             [4]int
	// Outline draws row text outlined (VehicleList: AdvancedColour(black).outline()
	// with {COLOUR WINDOW_2}); HoverSolid highlights the hovered row with a solid
	// shade-0 rect instead of the translucent one.
	Outline, HoverSolid bool
	// CellColour, if set, gives a cell's text colour (-1 = the default).
	CellColour func(row, col int) int
	// CurrentTab is the selected tab; OnTab is called when a tab is clicked.
	CurrentTab int
	OnTab      func(tab int)
}

// setTableGeometry fills the 600×197 town/station list geometry.
func (w *TownListWindow) setTableGeometry() *TownListWindow {
	w.W, w.H, w.RowHeight, w.StatusY, w.List = tlW, tlH, tlRowHeight, tlH-17, tlList
	return w
}

// NewTownList opens the town list at (sw - 600, 30), as TownList::open.
func NewTownList(sw int, towns []TownRow, titleColour, colour uint8) *TownListWindow {
	rows := make([][]string, len(towns))
	for i, t := range towns {
		rows[i] = []string{t.Name, t.Type, itoa3(t.Population), itoa3(t.Stations)}
	}
	status := itoa3(len(rows)) + " towns"
	if len(rows) == 1 {
		status = "1 town"
	}
	return (&TownListWindow{X: sw - tlW, Y: 30, TitleColour: titleColour, Colour: colour, Caption: "Towns", Tabs: 4,
		Columns: []TableColumn{{4, 200, "Name", 0, 198}, {204, 80, "Type", 200, 78}, {284, 70, "Population", 280, 68}, {354, 70, "Stations", 350, 68}},
		Rows:    rows, Status: status, Hover: -1}).setTableGeometry()
}

// StationRow is one station of the station list.
type StationRow struct {
	Name, Status, Accepts string
	Waiting               int
}

// NewStationList opens the station list (StationList.cpp): columns Name,
// Status, Total waiting, Accepts; five tabs. PARTIAL: the transport-mode icons after
// the name, the per-mode tabs' filtering, the company select button.
func NewStationList(sw int, stations []StationRow, titleColour, colour uint8) *TownListWindow {
	rows := make([][]string, len(stations))
	for i, s := range stations {
		rows[i] = []string{s.Name, s.Status, itoa3(s.Waiting) + " units", s.Accepts}
	}
	status := itoa3(len(rows)) + " stations"
	if len(rows) == 1 {
		status = "1 station"
	}
	return (&TownListWindow{X: sw - tlW, Y: 30, TitleColour: titleColour, Colour: colour, Caption: "All Stations", Tabs: 5,
		Columns: []TableColumn{{4, 200, "Name", 0, 198}, {204, 200, "Status", 200, 198}, {404, 90, "Total waiting", 400, 88}, {494, 120, "Accepts", 490, 118}},
		Rows:    rows, Status: status, Hover: -1}).setTableGeometry()
}

var tlList = [4]int{3, 56, 594, 126}

func (w *TownListWindow) in(r [4]int, px, py int) bool {
	return px >= w.X+r[0] && px < w.X+r[0]+r[2] && py >= w.Y+r[1] && py < w.Y+r[1]+r[3]
}

// Contains reports whether the point is on the window.
func (w *TownListWindow) Contains(px, py int) bool { return w.in([4]int{0, 0, w.W, w.H}, px, py) }

// HandleClick: close, or a row (selects it, as rowHover).
func (w *TownListWindow) HandleClick(px, py int) {
	if w.in([4]int{w.W - 15, 2, 13, 13}, px, py) {
		if w.OnClose != nil {
			w.OnClose()
		}
		return
	}
	for i := range w.Tabs {
		if w.in([4]int{3 + 31*i, 15, 31, 27}, px, py) {
			if w.OnTab != nil && i != w.CurrentTab {
				w.OnTab(i)
			}
			return
		}
	}
	if w.in(w.List, px, py) {
		if i := (py - (w.Y + w.List[1] + 1) + w.Scroll) / w.RowHeight; i >= 0 && i < len(w.Rows) {
			w.Hover = i
		}
	}
}

// HandleWheel scrolls the list.
func (w *TownListWindow) HandleWheel(dy float64) {
	maxScroll := max(len(w.Rows)*w.RowHeight-w.List[3], 0)
	w.Scroll = min(max(w.Scroll-int(dy*3*float64(w.RowHeight)), 0), maxScroll)
}

// Draw draws the window.
func (w *TownListWindow) Draw(screen *ebiten.Image) {
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
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, false)
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 41, w.W-1, w.H-1, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, w.W-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		for i := range w.Tabs {
			img := imgTab
			if i == w.CurrentTab {
				img++
			}
			if sp, err := gfx8.Decode(&g1.Elements[img]); err == nil {
				gfx8.DrawSpriteNoise(rt, sp, 3+31*i, 15, gfx8.OpTransparent|gfx8.OpSrc, sec, nil)
			}
		}
		for _, c := range w.Columns {
			hw := min(c.HeaderW, w.W-4-c.HeaderX) // headers are clipped to the window (prepareDraw)
			gfx8.FillRectInset(rt, c.HeaderX, 43, c.HeaderX+hw-1, 43+12-1, gfx8.InsetColour{Map: sec}, 0)
		}
		l := w.List
		cx, cy, cw, ch, _ := drawVScrollView(rt, l[0], l[1], l[2], l[3], gfx8.InsetColour{Map: sec}, len(w.Rows)*w.RowHeight, w.Scroll)
		w.content = [4]int{cx, cy, cw, ch}
		gfx8.FillRect(rt, cx, cy, cx+cw-1, cy+ch-1, gfx8.Shade(sec, 3)) // drawScroll: clearSingle(shade 3)
		m30, _ := g1.GetPaletteMap(0x30)
		if w.Hover >= 0 && w.Hover < len(w.Rows) {
			y := cy + w.Hover*w.RowHeight - w.Scroll
			if y+w.RowHeight > cy && y < cy+ch {
				if w.HoverSolid { // drawRect(0, y, width, rowHeight, getShade(secondary, 0))
					gfx8.FillRect(rt, cx, max(y, cy), cx+cw-1, min(y+w.RowHeight-1, cy+ch-1), gfx8.Shade(sec, 0))
				} else {
					gfx8.FillRectTranslucent(rt, cx, max(y, cy), cx+cw-1, min(y+w.RowHeight-1, cy+ch-1), m30)
				}
			}
		}
	})
	DrawWindowCaption(screen, w.X+1, w.Y+1, w.W-2, w.Caption)
	for _, c := range w.Columns {
		DrawTextBold(screen, c.Label, w.X+c.HeaderX, w.Y+43, black)
	}
	ct := w.content
	if len(w.Rows)*w.RowHeight > w.List[3] {
		drawVScrollArrows(screen, w.X+w.List[0], w.Y+w.List[1], w.List[2], w.List[3])
	}
	for i, r := range w.Rows {
		y := w.Y + ct[1] + i*w.RowHeight - w.Scroll
		if y < w.Y+ct[1] || y+w.TextDY+10 > w.Y+ct[1]+ct[3] {
			continue
		}
		y += w.TextDY
		c := black
		if i == w.Hover && !w.Outline {
			c = win2
		}
		x := w.X + ct[0]
		for k, cl := range w.Columns {
			if k >= len(r) {
				break
			}
			t := r[k]
			for len(t) > 0 {
				if tw, _ := MeasureTextBold(t); tw <= cl.TextW {
					break
				}
				t = t[:len(t)-1]
			}
			switch {
			case w.Outline:
				var body color.Color = win2
				if w.CellColour != nil {
					if ci := w.CellColour(i, k); ci >= 0 {
						body, _ = TextColours(ci)
					}
				}
				DrawTextOutlined(screen, t, x+cl.TextX, y, body, black)
			default:
				DrawTextBold(screen, t, x+cl.TextX, y, c)
			}
		}
	}
	DrawTextBold(screen, w.Status, w.X+4, w.Y+w.StatusY, black)
}

// VehicleRow is one vehicle of the vehicle list.
type VehicleRow struct {
	Name        string
	Profit      int64 // totalRecentProfit / 4
	AgeYears    int
	Reliability int
}

var vehicleListCaptions = [6]string{"Trains", "Buses", "Trucks", "Trams", "Aircraft", "Ships"}

// vehicleListRowHeights is VehicleList.cpp row_heights per tab.
var vehicleListRowHeights = [6]int{28, 28, 28, 28, 48, 36}

// NewVehicleList opens the vehicle list (VehicleList.cpp) on a tab: 550×213,
// six tabs, columns Name / Monthly Profit / Age / Reliability, rows
// row_heights[tab] tall, text outlined in the window's secondary colour (red
// for a loss). PARTIAL: the vehicle drawn inline at the bottom of each row
// (drawTrainInline), the status after the name (VehicleHead::getStatus), the
// filter dropdowns, sorting, the company select button.
func NewVehicleList(sw int, tab int, vehicles []VehicleRow, titleColour, colour uint8) *TownListWindow {
	rows := make([][]string, len(vehicles))
	for i, v := range vehicles {
		profit := v.Profit
		if profit < 0 {
			profit = -profit
		}
		age := itoa3(v.AgeYears) + " years"
		if v.AgeYears == 1 {
			age = "1 year"
		}
		p := currency32(profit) // {COLOUR RED}-{CURRENCY32} for a loss: '-' then the amount
		if v.Profit < 0 {
			p = "-" + p
		}
		rows[i] = []string{v.Name, p, age, itoa3(v.Reliability) + "%"}
	}
	status := itoa3(len(rows)) + " vehicles"
	if len(rows) == 1 {
		status = "1 vehicle"
	}
	w := &TownListWindow{X: sw - 550, Y: 30, TitleColour: titleColour, Colour: colour, Caption: vehicleListCaptions[tab], Tabs: 6,
		Columns: []TableColumn{{4, 310, "Name", 4 + 1 - 4, 310 - 2}, {314, 100, "Monthly Profit", 314 + 1 - 4, 100 - 2}, {414, 65, "Age", 414 + 1 - 4, 65 - 2}, {479, 67, "Reliability", 479 + 1 - 4, 67 - 2}},
		Rows:    rows, Status: status, Hover: -1, CurrentTab: tab,
		W: 550, H: 213, RowHeight: vehicleListRowHeights[tab], StatusY: 213 - 13, TextDY: 2, List: [4]int{3, 56, 544, 138},
		Outline: true, HoverSolid: true}
	w.CellColour = func(r, c int) int {
		if c == 1 && vehicles[r].Profit < 0 {
			return 27 // Colour::red ({COLOUR RED})
		}
		return -1
	}
	return w
}

// currency32 is {CURRENCY32}: formatCurrency, as uilayout.Currency48.
func currency32(v int64) string { return L.Currency48(v) }
