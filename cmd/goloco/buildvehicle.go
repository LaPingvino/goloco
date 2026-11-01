package main

import (
	"fmt"
	"image/color"

	"github.com/LaPingvino/goloco/pkg/render"
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/LaPingvino/goloco/pkg/ui"
)

// buildvehicle.go — the "Build Vehicles" toolbar window.  It lists every loaded
// vehicle object (ObjMgr.Vehicles) with its resolved name and stats, lets the
// player pick one, and places a running vehicle onto the track they built
// (construction-head tile if active, otherwise the first track tile found).
//
// This is the UI counterpart of World.SpawnVehicleAt / SpawnTestVehicle — the
// same placement the GOLOCO_OPEN=win/track debug hooks used, now reachable
// through real clicks.
//
// OpenLoco reference: src/OpenLoco/src/Ui/Windows/BuildVehicle.cpp
// (scrollable vehicle list + build; goloco's slice is single-object for now).

// buildVehicleRow is one selectable vehicle in the Build Vehicles window.
type buildVehicleRow struct {
	objID int
	name  string
	mode  string
	speed int // km/h
	power int
	cost  int
}

// buildVehicleRows resolves ObjMgr.Vehicles into display rows. The vehicle DAT
// name is a StringId in the object's own string table; loader.go already
// extracts it into VehicleObject.DisplayName (extractFirstString), so we reuse
// that, falling back to the header name then a synthesized "Vehicle N (mode)".
func (g *Game) buildVehicleRows() []buildVehicleRow {
	if g.objMgr == nil {
		return nil
	}
	rows := make([]buildVehicleRow, 0, len(g.objMgr.Vehicles))
	for i, v := range g.objMgr.Vehicles {
		if v == nil {
			continue
		}
		name := v.DisplayName
		if name == "" {
			name = v.Header.GetName()
		}
		if name == "" {
			name = fmt.Sprintf("Vehicle %d (%s)", i, v.Mode)
		}
		rows = append(rows, buildVehicleRow{
			objID: i,
			name:  name,
			mode:  v.Mode.String(),
			speed: int(v.GetSpeedKmh() + 0.5),
			power: int(v.Power),
			cost:  int(v.CostFactor),
		})
	}
	return rows
}

// Build Vehicles window layout.
const (
	bvWinW    = 360
	bvWinH    = 340
	bvRowH    = 15
	bvHeaderH = 18 // column-title strip
	bvFooterH = 30 // Build button + status
	bvPad     = 6
	bvBuildW  = 96
	bvBuildH  = 20
	bvNameX   = 6 // column x-offsets within content
	bvModeX   = 168
	bvSpeedX  = 224
	bvPowerX  = 272
	bvCostX   = 320
)

// openBuildVehicleWindow opens the Build Vehicles list. Selecting a row and
// clicking Build places that vehicle on the player's track.
func (g *Game) openBuildVehicleWindow() {
	rows := g.buildVehicleRows()
	selected := -1
	scroll := 0
	status := "Select a vehicle, then Build"

	win := ui.NewSimpleWindow("Build Vehicles", 70, 50, bvWinW, bvWinH)

	listArea := func(ch int) (top, height int) {
		top = bvHeaderH
		height = ch - bvHeaderH - bvFooterH
		return
	}
	visibleRows := func(ch int) int {
		_, h := listArea(ch)
		return h / bvRowH
	}
	maxScroll := func(ch int) int {
		m := len(rows) - visibleRows(ch)
		if m < 0 {
			return 0
		}
		return m
	}

	win.DrawContent = func(screen *ebiten.Image, cx, cy, cw, ch int, r *render.Renderer) {
		textCol := color.RGBA{20, 20, 20, 255}
		dimCol := color.RGBA{80, 80, 80, 255}

		// Column headers.
		ui.DrawText(screen, "Name", cx+bvNameX, cy+3, dimCol)
		ui.DrawText(screen, "Mode", cx+bvModeX, cy+3, dimCol)
		ui.DrawText(screen, "km/h", cx+bvSpeedX, cy+3, dimCol)
		ui.DrawText(screen, "Pwr", cx+bvPowerX, cy+3, dimCol)
		ui.DrawText(screen, "Cost", cx+bvCostX, cy+3, dimCol)

		listY, listH := listArea(ch)
		listTop := cy + listY
		fillRect(screen, float64(cx), float64(listTop), float64(cw), float64(listH), color.RGBA{210, 205, 185, 255})

		if len(rows) == 0 {
			ui.DrawText(screen, "No vehicle objects loaded", cx+bvPad, listTop+8, dimCol)
		} else {
			if scroll > maxScroll(ch) {
				scroll = maxScroll(ch)
			}
			if scroll < 0 {
				scroll = 0
			}
			visible := listH / bvRowH
			for i := 0; i < visible; i++ {
				idx := scroll + i
				if idx >= len(rows) {
					break
				}
				iy := listTop + i*bvRowH
				row := rows[idx]
				if idx == selected {
					fillRect(screen, float64(cx), float64(iy), float64(cw), float64(bvRowH), color.RGBA{120, 150, 200, 255})
				} else if idx%2 == 1 {
					fillRect(screen, float64(cx), float64(iy), float64(cw), float64(bvRowH), color.RGBA{201, 196, 176, 255})
				}
				ty := iy + (bvRowH-10)/2
				name := truncateToWidth(row.name, bvModeX-bvNameX-4)
				ui.DrawText(screen, name, cx+bvNameX, ty, textCol)
				ui.DrawText(screen, row.mode, cx+bvModeX, ty, textCol)
				ui.DrawText(screen, fmt.Sprintf("%d", row.speed), cx+bvSpeedX, ty, textCol)
				ui.DrawText(screen, fmt.Sprintf("%d", row.power), cx+bvPowerX, ty, textCol)
				ui.DrawText(screen, fmt.Sprintf("%d", row.cost), cx+bvCostX, ty, textCol)
			}
		}

		// Footer: Build button + status text.
		btnX := cx + bvPad
		btnY := cy + ch - bvFooterH + 5
		drawButton(screen, btnX, btnY, bvBuildW, bvBuildH, "Build", false, selected >= 0)
		ui.DrawText(screen, status, btnX+bvBuildW+10, btnY+(bvBuildH-10)/2, textCol)
	}

	win.OnContentClick = func(relX, relY int) {
		ch := win.Height - simpleWindowChrome
		// Build button.
		btnY := ch - bvFooterH + 5
		if relY >= btnY && relY < btnY+bvBuildH && relX >= bvPad && relX < bvPad+bvBuildW {
			if selected < 0 || selected >= len(rows) {
				return
			}
			objID := rows[selected].objID
			ok := false
			if h := g.w.ConstructionHeadState(); h.Active {
				ok = g.w.SpawnVehicleAt(objID, h.X, h.Y)
			}
			if !ok {
				ok = g.w.SpawnTestVehicle(objID)
			}
			if ok {
				status = fmt.Sprintf("Built %s", rows[selected].name)
			} else {
				status = "No track to place on"
			}
			return
		}
		// Row selection.
		listY, listH := listArea(ch)
		if relY >= listY && relY < listY+listH {
			row := (relY - listY) / bvRowH
			idx := scroll + row
			if idx >= 0 && idx < len(rows) {
				selected = idx
				status = fmt.Sprintf("Selected %s", rows[idx].name)
			}
		}
	}

	win.OnScrollWheel = func(dy float64) {
		ch := win.Height - simpleWindowChrome
		scroll = iclamp(scroll-int(dy), 0, maxScroll(ch))
	}

	g.windowMgr.OpenWindow(win)
}

// truncateToWidth clips a string to fit within maxW pixels (bitmap glyphs),
// appending nothing — used to keep the name column from bleeding into others.
func truncateToWidth(s string, maxW int) string {
	if w, _ := ui.MeasureText(s); w <= maxW {
		return s
	}
	for len(s) > 1 {
		s = s[:len(s)-1]
		if w, _ := ui.MeasureText(s + "."); w <= maxW {
			return s + "."
		}
	}
	return s
}
