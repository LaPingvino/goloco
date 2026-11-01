package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/hajimehoshi/ebiten/v2"
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
	return &TownListWindow{X: sw - tlW, Y: 30, TitleColour: titleColour, Colour: colour, Caption: "Towns", Tabs: 4,
		Columns: []TableColumn{{4, 200, "Name", 0, 198}, {204, 80, "Type", 200, 78}, {284, 70, "Population", 280, 68}, {354, 70, "Stations", 350, 68}},
		Rows:    rows, Status: status, Hover: -1}
}

// StationRow is one station of the station list.
type StationRow struct {
	Name, Status, Accepts string
	Waiting               int
}

// NewStationList opens the station list (StationList.cpp): columns Name,
// Status, Total waiting, Accepts; five tabs. PARTIAL: the caption lacks the
// company name ("{STRINGID} - All Stations"), the transport-mode icons after
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
	return &TownListWindow{X: sw - tlW, Y: 30, TitleColour: titleColour, Colour: colour, Caption: "All Stations", Tabs: 5,
		Columns: []TableColumn{{4, 200, "Name", 0, 198}, {204, 200, "Status", 200, 198}, {404, 90, "Total waiting", 400, 88}, {494, 120, "Accepts", 490, 118}},
		Rows:    rows, Status: status, Hover: -1}
}

var tlList = [4]int{3, 56, 594, 126}

func (w *TownListWindow) in(r [4]int, px, py int) bool {
	return px >= w.X+r[0] && px < w.X+r[0]+r[2] && py >= w.Y+r[1] && py < w.Y+r[1]+r[3]
}

// Contains reports whether the point is on the window.
func (w *TownListWindow) Contains(px, py int) bool { return w.in([4]int{0, 0, tlW, tlH}, px, py) }

// HandleClick: close, or a row (selects it, as rowHover).
func (w *TownListWindow) HandleClick(px, py int) {
	if w.in([4]int{tlW - 15, 2, 13, 13}, px, py) {
		if w.OnClose != nil {
			w.OnClose()
		}
		return
	}
	if w.in(tlList, px, py) {
		if i := (py - (w.Y + tlList[1] + 1) + w.Scroll) / tlRowHeight; i >= 0 && i < len(w.Rows) {
			w.Hover = i
		}
	}
}

// HandleWheel scrolls the list.
func (w *TownListWindow) HandleWheel(dy float64) {
	maxScroll := max(len(w.Rows)*tlRowHeight-tlList[3], 0)
	w.Scroll = min(max(w.Scroll-int(dy*3*tlRowHeight), 0), maxScroll)
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
	DrawWindowFrame(screen, w.X, w.Y, tlW, tlH, w.TitleColour, false)
	draw8(screen, w.X, w.Y, tlW, tlH, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 41, tlW-1, tlH-1, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, tlW-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		for i := range w.Tabs {
			img := imgTab
			if i == 0 {
				img++
			}
			if sp, err := gfx8.Decode(&g1.Elements[img]); err == nil {
				gfx8.DrawSpriteNoise(rt, sp, 3+31*i, 15, gfx8.OpTransparent|gfx8.OpSrc, sec, nil)
			}
		}
		for _, c := range w.Columns {
			hw := min(c.HeaderW, tlW-4-c.HeaderX) // headers are clipped to the window (prepareDraw)
			gfx8.FillRectInset(rt, c.HeaderX, 43, c.HeaderX+hw-1, 43+12-1, gfx8.InsetColour{Map: sec}, 0)
		}
		l := tlList
		cx, cy, cw, ch, _ := drawVScrollView(rt, l[0], l[1], l[2], l[3], gfx8.InsetColour{Map: sec}, len(w.Rows)*tlRowHeight, w.Scroll)
		w.content = [4]int{cx, cy, cw, ch}
		gfx8.FillRect(rt, cx, cy, cx+cw-1, cy+ch-1, gfx8.Shade(sec, 3)) // drawScroll: clearSingle(shade 3)
		m30, _ := g1.GetPaletteMap(0x30)
		if w.Hover >= 0 {
			y := cy + w.Hover*tlRowHeight - w.Scroll
			if y+tlRowHeight > cy && y < cy+ch {
				gfx8.FillRectTranslucent(rt, cx, max(y, cy), cx+cw-1, min(y+tlRowHeight-1, cy+ch-1), m30)
			}
		}
	})
	DrawWindowCaption(screen, w.X+1, w.Y+1, tlW-2, w.Caption)
	for _, c := range w.Columns {
		DrawTextBold(screen, c.Label, w.X+c.HeaderX, w.Y+43, black)
	}
	ct := w.content
	if len(w.Rows)*tlRowHeight > tlList[3] {
		drawVScrollArrows(screen, w.X+tlList[0], w.Y+tlList[1], tlList[2], tlList[3])
	}
	for i, r := range w.Rows {
		y := w.Y + ct[1] + i*tlRowHeight - w.Scroll
		if y < w.Y+ct[1] || y+10 > w.Y+ct[1]+ct[3] {
			continue
		}
		c := black
		if i == w.Hover {
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
			DrawTextBold(screen, t, x+cl.TextX, y, c)
		}
	}
	DrawTextBold(screen, w.Status, w.X+4, w.Y+tlH-17, black)
}
