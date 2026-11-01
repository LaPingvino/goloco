package main

import (
	"fmt"
	"image/color"

	"github.com/LaPingvino/goloco/pkg/render"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/hajimehoshi/ebiten/v2"
)

// listwindows.go — the three right-toolbar list windows (Vehicles, Stations,
// Towns).  Each is a scrollable list; clicking a row centres the camera on the
// item's tile.  They follow the New Game browser's list/scroll conventions
// (see openNewGameWindow) and SimpleWindow's themed drawing.
//
// OpenLoco reference: src/OpenLoco/src/Ui/Windows/VehicleList.cpp,
// StationList.cpp, TownList.cpp (scrollable rows, click to open/locate).

// listRow is one entry in a list window.
type listRow struct {
	label        string
	tileX, tileY int
	hasPos       bool // when false the row is informational (no camera jump)
}

// listWindow layout constants.
const (
	listItemH   = 14 // per-row height (10px glyphs + padding)
	listHeaderH = 18 // count line above the list
	listPad     = 6
)

// openListWindow builds and opens a generic scrollable list window.  Clicking a
// row with a position centres the camera on it.
func (g *Game) openListWindow(title string, x, y int, rows []listRow, emptyMsg string) {
	const winW, winH = 300, 320
	scroll := 0

	win := ui.NewSimpleWindow(title, x, y, winW, winH)

	visibleRows := func() int {
		contentH := win.Height - simpleWindowChrome
		return (contentH - listHeaderH) / listItemH
	}
	maxScroll := func() int {
		m := len(rows) - visibleRows()
		if m < 0 {
			return 0
		}
		return m
	}

	win.DrawContent = func(screen *ebiten.Image, cx, cy, cw, ch int, r *render.Renderer) {
		textCol := color.RGBA{20, 20, 20, 255}
		dimCol := color.RGBA{80, 80, 80, 255}

		// Count header.
		ui.DrawText(screen, fmt.Sprintf("%d total", len(rows)), cx+listPad, cy+3, dimCol)

		listY := cy + listHeaderH
		listH := ch - listHeaderH
		fillRect(screen, float64(cx), float64(listY), float64(cw), float64(listH), color.RGBA{210, 205, 185, 255})

		if len(rows) == 0 {
			ui.DrawText(screen, emptyMsg, cx+listPad, listY+8, dimCol)
			return
		}

		if scroll > maxScroll() {
			scroll = maxScroll()
		}
		if scroll < 0 {
			scroll = 0
		}

		visible := listH / listItemH
		for i := 0; i < visible; i++ {
			idx := scroll + i
			if idx >= len(rows) {
				break
			}
			iy := listY + i*listItemH
			// Alternate row shading for readability.
			if idx%2 == 1 {
				fillRect(screen, float64(cx), float64(iy), float64(cw), float64(listItemH), color.RGBA{201, 196, 176, 255})
			}
			ui.DrawText(screen, rows[idx].label, cx+listPad, iy+(listItemH-10)/2, textCol)
		}
	}

	win.OnContentClick = func(relX, relY int) {
		if relY < listHeaderH {
			return
		}
		row := (relY - listHeaderH) / listItemH
		idx := scroll + row
		if idx < 0 || idx >= len(rows) {
			return
		}
		if rows[idx].hasPos {
			g.w.SetCamera(float64(rows[idx].tileX), float64(rows[idx].tileY))
		}
	}

	win.OnScrollWheel = func(dy float64) {
		scroll = iclamp(scroll-int(dy), 0, maxScroll())
	}

	g.windowMgr.OpenWindow(win)
}

// simpleWindowChrome is the total non-content vertical space of a SimpleWindow
// (title bar + bottom border).  Mirrors simpleTitleBarHeight + simpleBorderWidth
// in pkg/ui/simplewindow.go.
const simpleWindowChrome = 22 + 3

// openVehiclesWindow lists save-loaded vehicle entities and runtime SimVehicles.
func (g *Game) openVehiclesWindow() { g.openVehicleList(0) }

// openVehicleList opens the ported VehicleList window on a tab (vehicle type).
func (g *Game) openVehicleList(tab int) {
	// VehicleList::open(getControllingId(), type): that company's vehicles.
	company, cname := g.w.ControllingCompany()
	var rows []ui.VehicleRow
	for _, v := range g.w.VehicleListRows() {
		if v.Type == tab && v.Owner == company {
			rows = append(rows, ui.VehicleRow{Name: v.Name, Profit: v.Profit, AgeYears: v.AgeYears, Reliability: v.Reliability})
		}
	}
	var tc, wc uint8 = 0, 1
	if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
		tc, wc = g.r.ObjMgr.InterfaceSkin.WindowTitlebarColour, g.r.ObjMgr.InterfaceSkin.WindowPlayerColor
	}
	w := ui.NewVehicleList(g.sw, tab, rows, tc, wc)
	if cname != "" {
		w.Caption = cname + " - " + w.Caption // stringid_trains etc.
	}
	w.OnClose = func() { g.townList = nil }
	w.OnTab = g.openVehicleList
	g.townList = w
}

// openStationsWindow lists every placed/loaded station on the map.
func (g *Game) openStationsWindow() {
	// The ported StationList window (pkg/ui/townlist.go, shared table layout).
	// StationList::open(getControllingId()): the controlling company's stations.
	company, cname := g.w.ControllingCompany()
	var rows []ui.StationRow
	for _, s := range g.w.StationList() {
		if s.Owner >= 0 && s.Owner != company {
			continue
		}
		rows = append(rows, ui.StationRow{Name: s.Name, Status: s.Status, Waiting: s.Waiting, Accepts: s.Accepts})
	}
	var tc, wc uint8 = 0, 1
	if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
		tc, wc = g.r.ObjMgr.InterfaceSkin.WindowTitlebarColour, g.r.ObjMgr.InterfaceSkin.WindowPlayerColor
	}
	w := ui.NewStationList(g.sw, rows, tc, wc)
	if cname != "" {
		w.Caption = cname + " - All Stations" // stringid_all_stations
	}
	w.OnClose = func() { g.townList = nil }
	g.townList = w
}

// openTownsWindow lists the scenario's towns with population and size.
func (g *Game) openTownsWindow() {
	// The ported TownList window (pkg/ui/townlist.go).
	infos := g.w.TownList()
	rows := make([]ui.TownRow, len(infos))
	for i, t := range infos {
		rows[i] = ui.TownRow{Name: t.Name, Type: t.Size, Population: int(t.Population), Stations: t.NumStations}
	}
	var tc, wc uint8 = 0, 1
	if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
		tc, wc = g.r.ObjMgr.InterfaceSkin.WindowTitlebarColour, g.r.ObjMgr.InterfaceSkin.WindowColour
	}
	w := ui.NewTownList(g.sw, rows, tc, wc)
	w.OnClose = func() { g.townList = nil }
	g.townList = w
}
