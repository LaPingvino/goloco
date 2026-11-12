package ui

import (
	"github.com/LaPingvino/goloco/pkg/core"
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
	"sort"
	"strings"
)

// TownList — a port of Ui/Windows/TownList.cpp, the town list tab: 600×197 at
// (screen width - 600, 30), skin windowTitlebarColour / windowColour; caption,
// close button, panel, four tabs (the three build tabs disabled outside the
// editor, so hidden), the sortable column headers (Name, Type, Population,
// Stations), a ScrollView of 10 px rows and the "N towns" status. The tabs'
// images come from TabIcon (drawTabs).
// PARTIAL: the editor's build tabs, window resizing and the resize handle.

const (
	tlW, tlH    = 600, 197
	tlRowHeight = 10
)

// TownRow is one town.
type TownRow struct {
	ID                   int // the town's index (Town::id)
	Name, Type           string
	Size                 int // Town::size (orderByType)
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
	// RowIDs are the rows' object ids in their current order; OnRow is a
	// row clicked (onScrollMouseDown: Town::open for the town list).
	RowIDs  []int
	OnRow   func(id int)
	Status  string
	Hover   int
	Scroll  int
	hold    L.VScrollHold // a press on the scrollbar held (stateScrollLeft)
	content [4]int
	OnClose func()

	// Geometry: window size, row height, the scroll view rect, the status
	// label's y and the text's y offset in a row (VehicleList: +2).
	W, H, RowHeight, StatusY, TextDY int
	// RowDraw draws extra content into row i whose top is at y in the scroll
	// view's own coordinates (VehicleList's drawTrainInline); clipped to it.
	RowDraw func(rt *gfx8.RenderTarget, i, y int)
	// CellSplit, if it returns at ≥ 0, draws a cell's text from byte at on
	// in another colour (a mid-string {COLOUR …}, as VehicleList's red
	// statuses).
	CellSplit func(r, c int) (at, colour int)
	// Window-specific widgets: Extra8 draws into the 8-bit frame, ExtraText
	// draws text over it, ExtraClick takes a click first (true = handled);
	// OpenDropdown shows a dropdown for the window.
	Extra8     func(rt *gfx8.RenderTarget, sec []uint8)
	ExtraText  func(screen *ebiten.Image)
	ExtraClick func(px, py int) bool
	// ExtraMouseOver replaces MouseOver's row hover inside the list.
	ExtraMouseOver func(px, py int)
	// Tooltip is the window's tooltip at a screen point ("" none) — a
	// content tooltip, before the widgets' (WidgetTooltip).
	Tooltip func(px, py int) string
	// TabTooltips are the tabs' tooltips (the Tab widgets' tooltip strings).
	TabTooltips []string
	// PlacementCost is the Fund New Industries tool's ghost cost (ok false:
	// none, or kFailure — the object's own cost shows).
	PlacementCost func() (int64, bool)
	OpenDropdown  func(*DropdownMenu)
	// Face, if > 0, is the company owner's image drawn on the company select
	// button (ImageButton at width − 28, 14; image at +1, +1) recoloured to
	// FaceColour (the company's primary colour).
	Face       int
	FaceColour uint8
	// Companies feed the company select dropdown (populateCompanySelect);
	// OnCompany switches the window to the chosen company.
	Companies []CompanyChoice
	OwnerID   int
	OnCompany func(id int)
	// ScreenW, ScreenH keep dropdowns on the screen (Ui::width/height).
	ScreenW, ScreenH int
	List             [4]int
	// Outline draws row text outlined (VehicleList: AdvancedColour(black).outline()
	// with {COLOUR WINDOW_2}); HoverSolid highlights the hovered row with a solid
	// shade-0 rect instead of the translucent one.
	Outline, HoverSolid bool
	// CellColour, if set, gives a cell's text colour (-1 = the default).
	CellColour func(row, col int) int
	// CurrentTab is the selected tab; OnTab is called when a tab is clicked.
	CurrentTab int
	OnTab      func(tab int)
	// TabIcon is drawTabs' image for tab i (a g1 element id; < 0 none) and
	// the colour it is recoloured to (Gfx::recolour; < 0 none).
	TabIcon func(i int) (image, colour int)
	// TabDisabled reports tab i disabled (disabledWidgets): hidden, the tabs
	// after it moved left (leftAlignTabs).
	TabDisabled func(i int) bool
	// ColourCaption: the caption is Style::colourText (StationList,
	// VehicleList) rather than whiteText (TownList).
	ColourCaption bool
	// Sort, if set, re-sorts Rows by a column (the header clicked); SortMode
	// is the current one, its header shows "▼" (table_header_*_desc).
	Sort     func(mode int)
	SortMode int
	// Reopen rebuilds the window from the game's current data (the lists'
	// onUpdate repopulating); nil: kept as opened.
	Reopen func()
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
	status := countLabel(len(rows), i18n.StatusTownsSingular, i18n.StatusTownsPlural) // TownList.cpp: status_towns_*
	if len(rows) == 1 {
		status = "1 town"
	}
	w := (&TownListWindow{X: sw - tlW, Y: 30, TitleColour: titleColour, Colour: colour, Caption: i18n.Label(i18n.TitleTowns), Tabs: 4,
		Columns: []TableColumn{{4, 200, i18n.Label(i18n.TableHeaderName), 0, 198}, {204, 80, i18n.Label(i18n.TableHeaderTownType), 200, 78}, {284, 70, i18n.Label(i18n.TableHeaderPopulation), 280, 68}, {354, 70, i18n.Label(i18n.TableHeaderStations), 350, 68}},
		Rows:    rows, Status: status, Hover: -1}).setTableGeometry()
	// sortTownList (run every update): a stable sort by the header clicked,
	// name by default.
	w.Sort = func(mode int) {
		order := append([]TownRow(nil), towns...)
		sort.SliceStable(order, func(i, j int) bool { return townOrder(mode, order[i], order[j]) })
		w.RowIDs = w.RowIDs[:0]
		for i, t := range order {
			w.Rows[i] = []string{t.Name, t.Type, itoa3(t.Population), itoa3(t.Stations)}
			w.RowIDs = append(w.RowIDs, t.ID)
		}
	}
	w.Sort(0)
	return w
}

// townOrder is TownList.cpp getOrder: name (strcmp, ascending), type (size
// descending, then population), population and stations (descending).
func townOrder(mode int, a, b TownRow) bool {
	switch mode {
	case 0:
		return a.Name < b.Name
	case 1:
		if a.Size != b.Size {
			return b.Size < a.Size
		}
		return b.Population < a.Population
	case 2:
		return b.Population < a.Population
	case 3:
		return b.Stations < a.Stations
	}
	return false
}

// StationRow is one station of the station list.
type StationRow struct {
	ID                    int // the station's id (the row opens its window)
	Name, Status, Accepts string
	Waiting               int
	AcceptsKey            string // orderByAccepts: the accepted cargo names run together
}

// NewStationList opens the station list (StationList.cpp): columns Name,
// Status, Total waiting, Accepts; five tabs (the game filters by tab:
// cmd/goloco openStationList). PARTIAL: the transport-mode icons after the
// name, the company select button.
func NewStationList(sw int, stations []StationRow, titleColour, colour uint8) *TownListWindow {
	rows := make([][]string, len(stations))
	for i, s := range stations {
		rows[i] = []string{s.Name, s.Status, itoa3(s.Waiting) + " units", s.Accepts}
	}
	status := countLabel(len(rows), i18n.StatusNumStationsSingular, i18n.StatusNumStationsPlural) // StationList.cpp: status_num_stations_*
	if len(rows) == 1 {
		status = "1 station"
	}
	w := (&TownListWindow{X: sw - tlW, Y: 30, TitleColour: titleColour, Colour: colour, Caption: i18n.Label(i18n.AllStations), Tabs: 5, ColourCaption: true,
		Columns: []TableColumn{{4, 200, i18n.Label(i18n.TableHeaderName), 0, 198}, {204, 200, i18n.Label(i18n.TableHeaderStatus), 200, 198}, {404, 90, i18n.Label(i18n.TableHeaderTotalWaiting), 400, 88}, {494, 120, i18n.Label(i18n.TableHeaderAccepts), 490, 118}},
		Rows:    rows, Status: status, Hover: -1}).setTableGeometry()
	// sortStationList: a stable sort by the header clicked, name by default.
	w.Sort = func(mode int) {
		order := append([]StationRow(nil), stations...)
		sort.SliceStable(order, func(i, j int) bool { return stationOrder(mode, order[i], order[j]) })
		w.RowIDs = w.RowIDs[:0]
		for i, s := range order {
			w.Rows[i] = []string{s.Name, s.Status, itoa3(s.Waiting) + " units", s.Accepts}
			w.RowIDs = append(w.RowIDs, s.ID)
		}
	}
	w.Sort(0)
	return w
}

// stationOrder is StationList.cpp getOrder: name (strcmp), status and total
// waiting both by the units waiting (descending), accepts (strcmp of the
// accepted cargo names run together).
func stationOrder(mode int, a, b StationRow) bool {
	switch mode {
	case 0:
		return a.Name < b.Name
	case 1, 2:
		return b.Waiting < a.Waiting
	case 3:
		return a.AcceptsKey < b.AcceptsKey
	}
	return false
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
	if w.ExtraClick != nil && w.ExtraClick(px, py) {
		return
	}
	if len(w.Companies) > 0 && w.OpenDropdown != nil && w.in([4]int{w.W - 28, 14, 26, 26}, px, py) {
		w.OpenDropdown(w.companySelect())
		return
	}
	if i := tabAt(tabLayout(w.Tabs, w.TabDisabled), px-w.X, py-w.Y); i >= 0 {
		if w.OnTab != nil && i != w.CurrentTab {
			w.OnTab(i)
		}
		return
	}
	// A column header (TableHeader, 12 px at y 43): sort by it, once per change.
	if w.Sort != nil {
		for i, c := range w.Columns {
			if w.in([4]int{c.HeaderX, 43, c.HeaderW, 12}, px, py) {
				if i != w.SortMode {
					w.SortMode, w.Hover = i, -1
					w.Sort(i)
				}
				return
			}
		}
	}
	// The scrollbar (when the rows don't fit): buttons and track move the
	// list (ScrollView's vButton/vTrack), the thumb starts a drag.
	if l := w.List; w.in(l, px, py) && len(w.Rows)*w.RowHeight > l[3] &&
		px >= w.X+l[0]+l[2]-L.ScrollbarSize-L.ScrollbarMargin {
		w.Scroll = w.hold.Begin(py-(w.Y+l[1]), py, l[3], len(w.Rows)*w.RowHeight, w.Scroll)
		return
	}
	if w.in(w.List, px, py) {
		if i := (py - (w.Y + w.List[1] + 1) + w.Scroll) / w.RowHeight; i >= 0 && i < len(w.Rows) {
			w.Hover = i
			if w.OnRow != nil && i < len(w.RowIDs) {
				w.OnRow(w.RowIDs[i])
			}
		}
	}
}

// MouseOver is onScrollMouseOver: over the list's rows (not its scrollbar)
// the row under the mouse becomes the hover row (rowHover; past the last row
// none). ExtraMouseOver takes it first when set (the Fund New Industries
// grid's cells).
func (w *TownListWindow) MouseOver(px, py int) {
	l := w.List
	if !w.in(l, px, py) || px >= w.X+l[0]+l[2]-L.ScrollbarSize-L.ScrollbarMargin {
		if w.ExtraMouseOver != nil { // the grid's var_846 cleared off it (event_08); a list keeps rowHover
			w.ExtraMouseOver(-1<<20, -1<<20)
		}
		return
	}
	if w.ExtraMouseOver != nil {
		w.ExtraMouseOver(px, py)
		return
	}
	if w.RowHeight <= 0 {
		return
	}
	i := (py - (w.Y + l[1] + 1) + w.Scroll) / w.RowHeight
	if i < 0 || i >= len(w.Rows) {
		i = -1
	}
	w.Hover = i
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
		xs := tabLayout(w.Tabs, w.TabDisabled)
		for i := range w.Tabs { // drawTabs: Widget::drawTab with each tab's image
			if xs[i] < 0 { // disabled: hidden (leftAlignTabs)
				continue
			}
			icon, colour := tabNoIcon, -1
			if w.TabIcon != nil {
				if img, c := w.TabIcon(i); img >= 0 && img < len(g1.Elements) {
					icon, colour = img, c
				}
			}
			var m []uint8
			if colour >= 0 {
				m, _ = g1.GetPaletteMap(colour)
			}
			drawTab(rt, xs[i], 15, i == w.CurrentTab, sec, icon, m)
		}
		for _, c := range w.Columns {
			hw := min(c.HeaderW, w.W-4-c.HeaderX) // headers are clipped to the window (prepareDraw)
			gfx8.FillRectInset(rt, c.HeaderX, 43, c.HeaderX+hw-1, 43+12-1, gfx8.InsetColour{Map: sec}, 0)
		}
		l := w.List
		cx, cy, cw, ch, _ := drawVScrollViewPressed(rt, l[0], l[1], l[2], l[3], gfx8.InsetColour{Map: sec}, len(w.Rows)*w.RowHeight, w.Scroll, w.hold.PressedPart())
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
		if w.Extra8 != nil {
			w.Extra8(rt, sec)
		}
		if w.Face > 0 && w.Face < len(g1.Elements) { // draw: the company owner's image
			if sp, err := gfx8.Decode(&g1.Elements[w.Face]); err == nil {
				m, _ := g1.GetPaletteMap(int(w.FaceColour))
				gfx8.DrawSpriteNoise(rt, sp, w.W-28+1, 14+1, gfx8.OpFor(sp, m != nil, false), m, nil)
			}
		}
		if w.RowDraw != nil && cw > 0 && ch > 0 { // drawn into the scroll view's own target (its clip)
			sub := gfx8.NewRenderTarget(cw, ch)
			for yy := range ch {
				copy(sub.Bits[yy*cw:(yy+1)*cw], rt.Bits[(cy+yy)*rt.Width+cx:(cy+yy)*rt.Width+cx+cw])
			}
			for i := range w.Rows {
				if y := i*w.RowHeight - w.Scroll; y+w.RowHeight > 0 && y < ch {
					w.RowDraw(sub, i, y)
				}
			}
			for yy := range ch {
				copy(rt.Bits[(cy+yy)*rt.Width+cx:(cy+yy)*rt.Width+cx+cw], sub.Bits[yy*cw:(yy+1)*cw])
			}
		}
	})
	if w.ColourCaption { // Caption::Style::colourText: the window's primary colour
		DrawWindowCaptionColour(screen, w.X+1, w.Y+1, w.W-2, w.Caption, int(w.TitleColour))
	} else {
		DrawWindowCaption(screen, w.X+1, w.Y+1, w.W-2, w.Caption)
	}
	for i, c := range w.Columns {
		label := c.Label
		if w.Sort != nil && i == w.SortMode {
			label += " " + string(rune(170)) // LocoChar ▼
		}
		DrawTextBold(screen, label, w.X+c.HeaderX, w.Y+43, black)
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
				if w.CellSplit != nil {
					if at, ci := w.CellSplit(i, k); at >= 0 && at < len(t) {
						lw := DrawTextOutlined(screen, t[:at], x+cl.TextX, y, body, black)
						c2, _ := TextColours(ci)
						DrawTextOutlined(screen, t[at:], x+cl.TextX+lw, y, c2, black)
						break
					}
				}
				DrawTextOutlined(screen, t, x+cl.TextX, y, body, black)
			default:
				DrawTextBold(screen, t, x+cl.TextX, y, c)
			}
		}
	}
	DrawTextBold(screen, w.Status, w.X+4, w.Y+w.StatusY, black)
	if w.ExtraText != nil {
		w.ExtraText(screen)
	}
}

// VehicleRow is one vehicle of the vehicle list.
type VehicleRow struct {
	HeadID      uint16 // the vehicle (a row's click opens it)
	Name        string
	Profit      int64 // totalRecentProfit / 4
	AgeYears    int
	Reliability int
	// Sort keys: totalRecentProfit, dayCreated.
	RecentProfit int64
	DayCreated   uint32
	// Cars are the train drawn inline at the row's bottom.
	Cars []TrainCarDraw
	// Status is VehicleHead::getStatus as text; StatusRed its {COLOUR RED}.
	Status    string
	StatusRed bool
	// Cargos are the cargo ids of its UnloadAll / WaitFor orders
	// (vehicleIsTransportingCargo).
	Cargos []int
}

// vehicleOrder is VehicleList.cpp getOrder: name (strlogicalcmp, default), profit
// (descending), age (dayCreated ascending: the oldest first), reliability
// (descending).
func vehicleOrder(mode int, a, b VehicleRow) bool {
	switch mode {
	case 0: // orderByName: strlogicalcmp
		return core.StrLogicalCmp(a.Name, b.Name) < 0
	case 1:
		return b.RecentProfit-a.RecentProfit < 0
	case 2:
		return int32(a.DayCreated-b.DayCreated) < 0
	case 3:
		return b.Reliability < a.Reliability
	}
	return false
}

// vehicleListCaptions() are the vehicle list's captions per type (the active
// language's, read when used).
func vehicleListCaptions() [6]string {
	return [6]string{i18n.Label(i18n.TooltipTrains), i18n.Label(i18n.TooltipBuses), i18n.Label(i18n.TooltipTrucks),
		i18n.Label(i18n.TooltipTrams), i18n.Label(i18n.TooltipAircraft), i18n.Label(i18n.TooltipShips)}
}

// vehicleListRowHeights is VehicleList.cpp row_heights per tab.
var vehicleListRowHeights = [6]int{28, 28, 28, 28, 48, 36}

// NewVehicleList opens the vehicle list (VehicleList.cpp) on a tab: 550×213,
// six tabs, columns Name / Monthly Profit / Age / Reliability, rows
// row_heights[tab] tall, text outlined in the window's secondary colour (red
// for a loss), each train drawn inline at the bottom of its row
// (drawTrainInline). PARTIAL: the status after the name
// (VehicleHead::getStatus), the filter dropdowns, the company select button.
func NewVehicleList(sw int, tab int, vehicles []VehicleRow, titleColour, colour uint8, cargos []VehicleListCargo) *TownListWindow {
	all := append([]VehicleRow(nil), vehicles...)
	vehicles = append([]VehicleRow(nil), vehicles...)
	rows := make([][]string, len(vehicles))
	cells := func(v VehicleRow) []string {
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
		name := v.Name // vehicle_list_status_2pos/3pos: "{name} - {status1}[ {status2}]"
		if v.Status != "" {
			name += " - " + v.Status
		}
		return []string{name, p, age, itoa3(v.Reliability) + "%"}
	}
	for i, v := range vehicles {
		rows[i] = cells(v)
	}
	status := vehicleListFooter(tab, len(rows))
	w := &TownListWindow{X: sw - 550, Y: 30, TitleColour: titleColour, Colour: colour, Caption: vehicleListCaptions()[tab], Tabs: 6,
		Columns: []TableColumn{{4, 310, i18n.Label(i18n.TableHeaderName), 4 + 1 - 4, 310 - 2}, {314, 100, i18n.Label(i18n.TableHeaderMonthlyProfit), 314 + 1 - 4, 100 - 2}, {414, 65, i18n.Label(i18n.TableHeaderAge), 414 + 1 - 4, 65 - 2}, {479, 67, i18n.Label(i18n.TableHeaderReliability), 479 + 1 - 4, 67 - 2}},
		Rows:    rows, Status: status, Hover: -1, CurrentTab: tab,
		W: 550, H: 213, RowHeight: vehicleListRowHeights[tab], StatusY: 213 - 13, TextDY: 2, List: [4]int{3, 56, 544, 138},
		Outline: true, HoverSolid: true, ColourCaption: true}
	// drawScroll: drawTrainInline at (0, y + (rowHeight − 28) / 2 + 6).
	w.RowDraw = func(rt *gfx8.RenderTarget, i, y int) {
		DrawTrainInline(rt, vehicles[i].Cars, 0, y+(w.RowHeight-28)/2+6)
	}
	w.CellSplit = func(r, c int) (int, int) {
		if c == 0 && vehicles[r].StatusRed && vehicles[r].Status != "" {
			return len(vehicles[r].Name) + len(" - "), 27 // Colour::red
		}
		return -1, 0
	}
	w.CellColour = func(r, c int) int {
		if c == 1 && vehicles[r].Profit < 0 {
			return 27 // Colour::red ({COLOUR RED})
		}
		return -1
	}
	// sortVehicleList: a stable sort by the header clicked, name by default
	// (the rows and their colours follow the sorted vehicles).
	w.Sort = func(mode int) {
		sort.SliceStable(vehicles, func(i, j int) bool { return vehicleOrder(mode, vehicles[i], vehicles[j]) })
		for i, v := range vehicles {
			w.Rows[i] = cells(v)
		}
	}
	w.Sort(0)
	addVehicleListFilters(w, tab, &vehicles, all, cells, cargos)
	return w
}

// currency32 is {CURRENCY32}: formatCurrency, as uilayout.Currency48.
func currency32(v int64) string { return L.Currency48(v) }

// Drag is the scrollbar's held press each tick (VScrollHold.Continue): the
// thumb follows the pointer, a held button repeats while on it.
func (w *TownListWindow) Drag(px, py int, held bool) {
	l := w.List
	onBar := w.in(l, px, py) && px >= w.X+l[0]+l[2]-L.ScrollbarSize-L.ScrollbarMargin
	w.Scroll = w.hold.Continue(held, onBar, py-(w.Y+l[1]), py, l[3], len(w.Rows)*w.RowHeight, w.Scroll)
}

// countLabel is a "{INT16} towns"-style status: the singular id for 1, else
// the plural, with the count put in.
func countLabel(n, singular, plural int) string {
	id := plural
	if n == 1 {
		id = singular
	}
	return strings.Replace(i18n.Label(id), "{INT16}", itoa3(n), 1)
}

// FormatCurrency is {CURRENCY32} (formatCurrency) for other packages.
func FormatCurrency(v int64) string { return currency32(v) }

// WidgetTooltip is the tooltip of the widget under a screen point: the
// close button's (tooltip_close_window) or a tab's.
func (w *TownListWindow) WidgetTooltip(px, py int) string {
	if w.in([4]int{w.W - 15, 2, 13, 13}, px, py) {
		return i18n.Str(i18n.TooltipCloseWindow)
	}
	if i := tabAt(tabLayout(w.Tabs, w.TabDisabled), px-w.X, py-w.Y); i >= 0 && i < len(w.TabTooltips) {
		return w.TabTooltips[i]
	}
	return ""
}
