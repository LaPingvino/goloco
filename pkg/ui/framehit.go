package ui

// Drag by the frame for the windows without a full widget list: each one's
// widget list as its tab's prepareDraw / onResize leaves it, for
// Window::findWidgetAt (Ui/Window.cpp: the last widget, neither empty nor
// hidden, whose rectangle holds the point). MouseInput.cpp stateNormalLeft
// moves the window when that widget is a frame, panel or newsPanel (caption:
// WindowStack.Press's onCaption) and resizes it on those in the grip corner;
// any other widget — a label, a group box, a Wt3, a viewport, a scroll view,
// a button — takes the press. Text and graphs drawn straight on the panel
// have no widget: a press there moves the window.
//
// Every list is in makeWidgets order, as inclusive left, top, right, bottom
// (Widget::left … bottom): fhW for a widget declared as origin and size,
// BVWidget{…} for an edge prepareDraw sets.

// fhW is a widget declared at origin (x, y) with size (w, h): right = x + w
// − 1, bottom = y + h − 1.
func fhW(x, y, w, h int) BVWidget { return bvW(x, y, w, h) }

// fhHidden is a widget prepareDraw hides (widget.hidden).
func fhHidden(b BVWidget, hidden bool) BVWidget { b.Hidden = hidden; return b }

// fhCommon is the frame, caption, close button and panel as the windows'
// common prepareDraw resizes them: frame to width − 1, height − 1; caption
// (1, 1) to width − 2, 13; close_button width − 15 … width − 3 at (2, 13 ×
// 13); panel from (0, panelTop) to width − 1, height − 1.
func fhCommon(w, h, panelTop int) []BVWidget {
	return []BVWidget{
		{L: 0, T: 0, R: w - 1, B: h - 1},
		{L: 1, T: 1, R: w - 2, B: 13},
		{L: w - 15, T: 2, R: w - 3, B: 14},
		{L: 0, T: panelTop, R: w - 1, B: h - 1},
	}
}

// fhTabs appends the 31 × 27 tabs at y 15 from a tabLayout: x per tab, −1
// for a tab leftAlignTabs hides (Widget::leftAlignTabs: right = left + 30).
func fhTabs(lw []BVWidget, xs []int) []BVWidget {
	for _, x := range xs {
		lw = append(lw, fhHidden(fhW(max(x, 0), 15, 31, 27), x < 0))
	}
	return lw
}

// fhDropdown is Widgets::dropdownWidgets: the combo box, then its 11 × 10
// button at (x + w − 12, y + 1).
func fhDropdown(x, y, w, h int) []BVWidget {
	return []BVWidget{fhW(x, y, w, h), fhW(x+w-12, y+1, 11, 10)}
}

// fhStepper is Widgets::stepperWidgets: the text box, then − (x + w − 26,
// y + 1, 13 × h − 2) and + (x + w − 13, y + 1, 12 × h − 2).
func fhStepper(x, y, w, h int) []BVWidget {
	return []BVWidget{fhW(x, y, w, h), fhW(x+w-26, y+1, 13, h-2), fhW(x+w-13, y+1, 12, h-2)}
}

// fhWidgetAt is findWidgetAt over a list placed at (wx, wy): the index of
// the last widget not hidden whose inclusive rectangle holds (x, y), −1 none.
func fhWidgetAt(lw []BVWidget, wx, wy, x, y int) int {
	at := -1
	for i, c := range lw {
		if !c.Hidden && x >= wx+c.L && x <= wx+c.R && y >= wy+c.T && y <= wy+c.B {
			at = i
		}
	}
	return at
}

// fhBare reports the widget at the point one of the bare ones (the frame,
// the panel or the newsPanel: indices in the list).
func fhBare(lw []BVWidget, wx, wy, x, y int, bare ...int) bool {
	at := fhWidgetAt(lw, wx, wy, x, y)
	for _, i := range bare {
		if at == i {
			return at >= 0
		}
	}
	return false
}

// The common lists' frame and panel indices.
const (
	fhFrame = 0
	fhPanel = 3
)

// --- Station (StationWindow.cpp) ---

// OnFrame: findWidgetAt's frame or panel at the point.
func (w *StationWindow) OnFrame(x, y int) bool {
	return fhBare(w.frameWidgets(), w.X, w.Y, x, y, fhFrame, fhPanel)
}

// frameWidgets is the tab's widgets after prepareDraw: Common::
// makeCommonWidgets (the nine tabs, leftAlignTabs hiding the vehicle types
// without a vehicle calling) and Common::prepareDraw's resize, then
//   - Station (0x0048E352): viewport (3, 44) to width − 4, height − 14; the
//     status_bar label (3, height − 12) to width − 14, height − 3;
//     centre_on_viewport the viewport's bottom-right 24 × 24 inside it.
//   - Cargo (0x0048E7C0): the scroll view to width − 26, height − 14; the
//     status bar as Station's; station_catchment width − 25 … width − 2,
//     44 … 67.
//   - Cargo ratings (0x0048EC3B): the scroll view to width − 4, height − 14;
//     the status bar as Station's.
//   - VehiclesStopping: the scroll view to width − 4, height − 14; the status
//     bar at height − 13 … height − 3, its right never set (3 + 400 − 1).
func (w *StationWindow) frameWidgets() []BVWidget {
	lw := fhTabs(fhCommon(w.W, w.H, 41), w.stationTabs())
	status := BVWidget{L: 3, T: w.H - 12, R: w.W - 14, B: w.H - 3}
	switch {
	case w.Tab == 0:
		vr, vb := w.W-4, w.H-14
		return append(lw, BVWidget{L: 3, T: 44, R: vr, B: vb}, status, BVWidget{L: vr - 24, T: vb - 24, R: vr - 1, B: vb - 1})
	case w.Tab == 1:
		return append(lw, BVWidget{L: 3, T: 44, R: w.W - 26, B: w.H - 14}, status, BVWidget{L: w.W - 25, T: 44, R: w.W - 2, B: 44 + 24 - 1})
	case w.Tab == 2:
		return append(lw, BVWidget{L: 3, T: 44, R: w.W - 4, B: w.H - 14}, status)
	}
	return append(lw, BVWidget{L: 3, T: 44, R: w.W - 4, B: w.H - 14}, BVWidget{L: 3, T: w.H - 13, R: 3 + stationVehiclesW - 1, B: w.H - 3})
}

// --- The lists (TownList.cpp, StationList.cpp, IndustryList.cpp,
// VehicleList.cpp) ---

// OnFrame: findWidgetAt's frame or panel at the point.
func (w *TownListWindow) OnFrame(x, y int) bool {
	return fhBare(w.frameWidgets(), w.X, w.Y, x, y, fhFrame, fhPanel)
}

// frameWidgets is the list's widgets after its prepareDraw / onResize, by
// the list it is (WinType; nil for a list goloco doesn't open: it moves by
// its caption only):
//   - TownList (0x00499F53): four tabs (leftAlignTabs: the build tabs
//     hidden outside the editor), the headers at y 43 … 54 each edge at
//     most width − 8, the scroll view (3, 56) to width − 4, height − 14, the
//     status bar (4, height − 12) to width − 19, height − 2.
//   - StationList prepareDraw: five tabs (leftAlignTabs), company_select
//     width − 28 … width − 3, 14 … 39, the headers to width − 4, the scroll
//     view and the status bar as TownList's.
//   - IndustryList (prepareDraw): two tabs (as declared), the
//     headers at y 44 … 54 to width − 4, the scroll view and the status bar
//     as TownList's; its NewIndustries tab (0x0045819F): the scroll view
//     (3, 45) to width − 4, height − 14.
//   - VehicleList onResize: six tabs (leftAlignTabs: the vehicle types the
//     company can't build hidden), company_select, the headers (the name's
//     right at max(310, width − 232) and each next one from there), the
//     scroll view (3, 56) to width − 4, height − 15, the status bar (3,
//     height − 13) to width − 19, height − 4, the filter and cargo
//     dropdowns from the bottom right (hidden at a width of 320 or less).
func (w *TownListWindow) frameWidgets() []BVWidget {
	W, H := w.W, w.H
	header := func(l, r, t, b, edge int) BVWidget {
		return BVWidget{L: min(l, W-edge), T: t, R: min(r, W-edge), B: b}
	}
	switch w.WinType {
	case WindowTypeTownList:
		if w.CurrentTab != 0 { // the editor's build tabs: not ported
			return nil
		}
		lw := fhTabs(fhCommon(W, H, 41), tabLayout(4, w.TabDisabled))
		return append(lw, BVWidget{L: 4, T: 43, R: min(203, W-8), B: 54}, header(204, 283, 43, 54, 8), header(284, 353, 43, 54, 8), header(354, 423, 43, 54, 8),
			BVWidget{L: 3, T: 56, R: W - 4, B: H - 14}, BVWidget{L: 4, T: H - 12, R: W - ResizeHandleSize - 1, B: H - 2})
	case WindowTypeStationList:
		lw := fhTabs(fhCommon(W, H, 41), tabLayout(5, w.TabDisabled))
		return append(lw, BVWidget{L: W - 28, T: 14, R: W - 3, B: 14 + 26 - 1},
			BVWidget{L: 4, T: 43, R: min(203, W-4), B: 54}, header(204, 403, 43, 54, 4), header(404, 493, 43, 54, 4), header(494, 613, 43, 54, 4),
			BVWidget{L: 3, T: 56, R: W - 4, B: H - 14}, BVWidget{L: 4, T: H - 12, R: W - ResizeHandleSize - 1, B: H - 2})
	case WindowTypeIndustryList:
		lw := fhTabs(fhCommon(W, H, 41), tabLayout(2, nil))
		if w.CurrentTab == 1 {
			return append(lw, BVWidget{L: 3, T: 45, R: W - 4, B: H - 14})
		}
		return append(lw, BVWidget{L: 4, T: 44, R: min(W-4, 203), B: 54}, header(204, 443, 44, 54, 4), header(444, 603, 44, 54, 4), header(603, 762, 44, 54, 4),
			BVWidget{L: 3, T: 56, R: W - 4, B: H - 14}, BVWidget{L: 4, T: H - 12, R: W - ResizeHandleSize - 1, B: H - 2})
	case WindowTypeVehicleList:
		lw := fhTabs(fhCommon(W, H, 41), tabLayout(6, w.TabDisabled))
		lw = append(lw, BVWidget{L: W - 28, T: 14, R: W - 3, B: 14 + 26 - 1})
		edge := W - 4
		r := min(max(310, W-100-65-67), edge) // sort_name: left 4 as declared
		lw = append(lw, BVWidget{L: 4, T: 43, R: r, B: 54})
		for _, hw := range [3]int{100, 65, 67} { // sort_profit, sort_age, sort_reliability
			l := min(r+1, edge)
			r = min(l+hw, edge)
			lw = append(lw, BVWidget{L: l, T: 43, R: r, B: 54})
		}
		lw = append(lw, BVWidget{L: 3, T: 56, R: W - 4, B: H - 15}, BVWidget{L: 3, T: H - 13, R: W - ResizeHandleSize - 1, B: H - 4})
		shown := W > 150+120+50 // enoughSpaceForCargoFilters
		cr := W - ResizeHandleSize - 2
		cl := cr - 150
		fr := cl - 2
		fl := fr - 120
		return append(lw,
			fhHidden(BVWidget{L: fl, T: H - 14, R: fr, B: H - 3}, !shown), fhHidden(BVWidget{L: fr - 10, T: H - 13, R: fr - 1, B: H - 4}, !shown),
			fhHidden(BVWidget{L: cl, T: H - 14, R: cr, B: H - 3}, !shown), fhHidden(BVWidget{L: cr - 10, T: H - 13, R: cr - 1, B: H - 4}, !shown))
	}
	return nil
}

// --- Company (CompanyWindow.cpp) ---

// OnFrame: findWidgetAt's frame or panel at the point.
func (w *CompanyFinancesWindow) OnFrame(x, y int) bool {
	return fhBare(w.frameWidgets(), w.X, w.Y, x, y, fhFrame, fhPanel)
}

// frameWidgets is the tab's widgets after prepareDraw: Common::
// makeCommonWidgets — the six tabs (Common::prepareDraw's leftAlignTabs:
// Challenge hidden for another company, disableChallengeTab), company_select
// width − 28 … width − 3, 14 … 39 (hidden on Challenge) — then
//   - Status (0x00431EBB): unk_11 (3, height − 12) to width − 14, height −
//     3; the viewport (3, 44) to width − 119, height − 14 and
//     centre_on_viewport in its corner; face width − 93 … width − 28, 57 …
//     122; change_owner_name width − 116 … width − 4, 124 … 145, hidden for
//     another company.
//   - Details (0x004327CF): the viewport (219, 54) to width − 26, height −
//     14; build_hq (315, 92, 24 × 24), hidden for another company; rotate_hq
//     below it, hidden while the headquarters tool isn't on (goloco has no
//     such tool: always); centre_on_viewport.
//   - ColourScheme (0x00432E0F): the ten checkboxes, the main scheme's two
//     swatches, each vehicle type's (hidden while its scheme isn't set).
//   - Finances (0x004332E4): the scroll view, the loan stepper and autopay
//     (hidden for another company).
//   - CargoDelivered, Challenge: the common widgets only.
func (w *CompanyFinancesWindow) frameWidgets() []BVWidget {
	W, H := w.W, w.H
	lw := fhTabs(fhCommon(W, H, 41), tabLayout(6, func(i int) bool { return i == CompanyTabChallenge && !w.Data.Own }))
	lw = append(lw, fhHidden(BVWidget{L: W - 28, T: 14, R: W - 3, B: 14 + 26 - 1}, w.Tab == CompanyTabChallenge))
	centre := func(vr, vb int) BVWidget { return BVWidget{L: vr - 24, T: vb - 24, R: vr - 1, B: vb - 1} }
	switch w.Tab {
	case CompanyTabStatus:
		vr, vb := W-119, H-14
		return append(lw, BVWidget{L: 3, T: H - 12, R: W - 14, B: H - 3}, BVWidget{L: 3, T: 44, R: vr, B: vb}, centre(vr, vb),
			BVWidget{L: W - 93, T: 57, R: W - 28, B: 57 + 66 - 1}, fhHidden(BVWidget{L: W - 116, T: 124, R: W - 4, B: 124 + 22 - 1}, !w.Status.Own))
	case CompanyTabDetails:
		vr, vb := W-26, H-14
		return append(lw, BVWidget{L: 219, T: 54, R: vr, B: vb}, fhHidden(fhW(315, 92, 24, 24), !w.Data.Own), fhHidden(fhW(315, 92+26, 24, 24), true), centre(vr, vb))
	case CompanyTabColourScheme:
		for i := range 10 {
			lw = append(lw, fhW(15, 81+17*i, 204, 12))
		}
		for _, x := range [2]int{221, 239} { // main_colour_*, then secondary_colour_*
			lw = append(lw, fhW(x, 48, 16, 16))
			for i := range 10 {
				lw = append(lw, fhHidden(fhW(x, 78+17*i, 16, 16), w.Colours.Custom&(1<<(i+1)) == 0))
			}
		}
		return lw
	case CompanyTabFinances:
		lw = append(lw, fhW(133, 45, 499, 215))
		for _, b := range append(fhStepper(87, 264, 100, 12), fhW(320, 264, 204, 12)) {
			lw = append(lw, fhHidden(b, !w.Data.Own))
		}
		return lw
	}
	return lw
}

// --- Company list (CompanyList.cpp) ---

// OnFrame: findWidgetAt's frame or panel at the point (the graph tabs have
// no widget but the common ones: a press on a graph moves the window).
func (w *CompanyListWindow) OnFrame(x, y int) bool {
	return fhBare(w.frameWidgets(), w.X, w.Y, x, y, fhFrame, fhPanel)
}

// frameWidgets is the tab's widgets: Common::makeCommonWidgets (seven tabs
// as declared) resized by Common::prepareDraw; on the list tab
// (CompanyList::prepareDraw) the headers at y 43 … 54, each
// edge at most width − 8, the scroll view (3, 56) to width − 4, height −
// 14, the status bar (3, height − 13) to width − 19, height − 3.
func (w *CompanyListWindow) frameWidgets() []BVWidget {
	W, H := w.W, w.H
	lw := fhTabs(fhCommon(W, H, 41), tabLayout(7, nil))
	if w.Tab != 0 {
		return lw
	}
	header := func(l, r int) BVWidget { return BVWidget{L: min(l, W-8), T: 43, R: min(r, W-8), B: 54} }
	return append(lw, BVWidget{L: 4, T: 43, R: min(178, W-8), B: 54}, header(179, 388), header(389, 533), header(534, 633),
		BVWidget{L: 3, T: 56, R: W - 4, B: H - 14}, BVWidget{L: 3, T: H - 13, R: W - ResizeHandleSize - 1, B: H - 3})
}

// --- Messages (MessageWindow.cpp) ---

// OnFrame: findWidgetAt's frame or panel at the point.
func (w *MessagesWindow) OnFrame(x, y int) bool {
	return fhBare(w.frameWidgets(), w.X, w.Y, x, y, fhFrame, fhPanel)
}

// frameWidgets: Common::makeCommonWidgets (two tabs) resized by Common::
// prepareDraw; Messages (prepareDraw) the scroll view (3, 45)
// to width − 4, height − 14; Settings each news kind's label (4, y, 230 ×
// 12) and dropdown (236, y, 124 × 12) from y 47, 15 a row, and the sound
// checkbox (4, 137, 346 × 12).
func (w *MessagesWindow) frameWidgets() []BVWidget {
	lw := fhTabs(fhCommon(w.W, w.H, 41), tabLayout(2, nil))
	if w.Tab == 0 {
		return append(lw, BVWidget{L: 3, T: 45, R: w.W - 4, B: w.H - 14})
	}
	for i := range 6 {
		y := messageSettingsRowY + messageSettingsRowDistance*i
		lw = append(append(lw, fhW(4, y, 230, 12)), fhDropdown(236, y, 124, 12)...)
	}
	return append(lw, fhW(4, 137, 346, 12))
}

// --- Options (Options.cpp) ---

// OnFrame: findWidgetAt's frame or panel at the point (a group box is a
// widget of its own: a press inside one doesn't move the window).
func (w *OptionsWindow) OnFrame(x, y int) bool {
	return fhBare(w.frameWidgets(), w.X, w.Y, x, y, fhFrame, fhPanel)
}

// frameWidgets is the tab's _widgets: Common::makeCommonWidgets (seven tabs,
// disableTabsByCurrentScene's leftAlignTabs: none hidden outside the editor)
// resized by Common::prepareDraw, then the tab's own as declared — none of
// them moves (the only hidden one, Misc's export_plugin_objects, shows only
// with custom objects in the index: goloco has none).
func (w *OptionsWindow) frameWidgets() []BVWidget {
	lw := fhTabs(fhCommon(w.W, w.H, 41), tabLayout(7, nil))
	add := func(b ...BVWidget) { lw = append(lw, b...) }
	label := func(x, y, wd int) BVWidget { return fhW(x, y, wd, 12) }
	switch w.Tab {
	case OptTabDisplay:
		add(fhW(4, 49, 392, 97))
		for i, y := range [4]int{63, 79, 95, 111} {
			add(label(10, y, 215))
			if i == 2 {
				add(fhStepper(235, y, 154, 12)...) // display_scale
			} else {
				add(fhDropdown(235, y, 154, 12)...)
			}
		}
		add(fhW(10, 127, 380, 12)) // show_fps
	case OptTabRendering:
		add(fhW(4, 49, 392, 128))
		for _, y := range [3]int{63, 79, 95} {
			add(label(10, y, 215))
			add(fhDropdown(235, y, 154, 12)...)
		}
		for _, y := range [4]int{110, 126, 142, 158} {
			add(fhW(10, y, 346, 12))
		}
		add(fhW(4, 180, 392, 32), label(10, 195, 215))
		add(fhDropdown(235, 194, 154, 12)...)
	case OptTabAudio: // AudioTab's offsets: sound group 49, volume 85, music 183
		add(fhW(4, 49, optW-8, 32))
		add(fhDropdown(10, 63, 346, 12)...)
		add(fhW(4, 85, optW-8, 16+3*26))
		for i := range 6 {
			y := 85 + 16 + 26*(i/2) // getSliderRowY
			lx, sx := 12, 82
			if i%2 == 1 {
				lx, sx = 190, 260
			}
			add(fhW(lx, y+3, 70, 12), fhW(sx, y-3, 100, 24))
		}
		add(fhW(4, 183, optW-8, 46), fhW(10, 197, 346, 12), fhW(10, 211, 173, 12), fhW(236, 211, 120, 12))
	case OptTabRegional:
		for _, y := range [5]int{49, 69, 84, 104, 119} {
			add(fhW(10, y, 173, 12))
			add(fhDropdown(183, y, 173, 12)...)
		}
		add(fhW(10, 134, 346, 12), fhW(10, 148, 346, 12))
	case OptTabControls:
		for i := range 5 {
			add(fhW(10, 49+15*i, 346, 12))
		}
		add(fhW(26, 124, 160, 12)) // customize_keys
	case OptTabCompany:
		add(fhW(4, 50, 412, 80), fhW(10, 64, 400, 12), fhW(265, 79, 75, 12), fhW(24, 79, 240, 12),
			fhW(10, 97, 400, 12), fhW(265, 112, 75, 12), fhW(24, 112, 240, 12), fhW(345, 59, 66, 66),
			fhW(4, 135, 412, 47), fhW(10, 149, 400, 12), fhW(265, 164, 75, 12), fhW(24, 164, 240, 12))
	case OptTabMisc:
		add(fhW(4, 49, 412, 62), fhW(10, 64, 400, 12), fhW(10, 79, 400, 12), fhW(10, 94, 400, 12),
			fhW(4, 115, 412, 77), fhW(10, 130, 400, 12), fhW(10, 145, 400, 12), fhW(10, 160, 400, 12), fhW(10, 175, 400, 12),
			fhW(4, 196, 412, 65), label(10, 211, 200))
		add(fhDropdown(250, 211, 156, 12)...)
		add(label(10, 226, 200))
		add(fhStepper(250, 226, 156, 12)...)
		add(fhHidden(fhW(10, 241, 400, 12), true)) // export_plugin_objects
		// goloco's competing companies group (options.go optCompY)
		add(fhW(4, optCompY, 412, 65))
		for i, y := range optCompRows {
			add(label(10, y, 235))
			if i == 0 {
				add(fhDropdown(250, y, 156, 12)...)
			} else {
				add(fhStepper(250, y, 156, 12)...)
			}
		}
	}
	return lw
}

// --- Music (MusicJukebox.cpp, MusicSelection.cpp) ---

// OnFrame: findWidgetAt's frame or panel at the point.
func (w *MusicJukeboxWindow) OnFrame(x, y int) bool {
	return fhBare(w.frameWidgets(), w.X, w.Y, x, y, fhFrame, fhPanel)
}

// frameWidgets is _widgets after prepareDraw's resize: the panel from y 15,
// "Currently playing:" and its dropdown, stop / play / next, the playlist
// dropdown, edit_selection.
func (w *MusicJukeboxWindow) frameWidgets() []BVWidget {
	lw := append(fhCommon(w.W, w.H, 15), fhW(mjLabel[0], mjLabel[1], mjLabel[2], mjLabel[3]))
	lw = append(lw, fhDropdown(mjPlaying[0], mjPlaying[1], mjPlaying[2], mjPlaying[3])...)
	for _, c := range mjControls {
		lw = append(lw, fhW(c[0], c[1], c[2], c[3]))
	}
	lw = append(lw, fhDropdown(mjPlaylist[0], mjPlaylist[1], mjPlaylist[2], mjPlaylist[3])...)
	return append(lw, fhW(mjEdit[0], mjEdit[1], mjEdit[2], mjEdit[3]))
}

// OnFrame: findWidgetAt's frame or panel at the point.
func (w *MusicSelectionWindow) OnFrame(x, y int) bool {
	return fhBare(w.frameWidgets(), w.X, w.Y, x, y, fhFrame, fhPanel)
}

// frameWidgets is _widgets after onResize: the panel from y 15; sort_years
// width − 79 … width − 5 and sort_title from 5 to its left, both 17 … 28;
// the scroll view (4, 30) to width − 5, height − 14; the status bar (4,
// height − 12) to width − 22, height − 2.
func (w *MusicSelectionWindow) frameWidgets() []BVWidget {
	W, H := w.W, w.H
	yr := W - 1 - msPadding
	yl := yr - msColumnYearsWidth + 1
	return append(fhCommon(W, H, 15), BVWidget{L: msPadding + 1, T: 17, R: yl - 1, B: 28}, BVWidget{L: yl, T: 17, R: yr, B: 28},
		BVWidget{L: msPadding, T: 30, R: W - 1 - msPadding, B: H - 1 - msStatusClearance},
		BVWidget{L: msPadding, T: H - 12, R: W - ResizeHandleSize - msPadding, B: H - 2})
}

// --- Keyboard shortcuts (KeyboardShortcuts.cpp, EditKeyboardShortcut.cpp) ---

// OnFrame: findWidgetAt's frame or panel at the point.
func (w *KeyboardShortcutsWindow) OnFrame(x, y int) bool {
	return fhBare(w.frameWidgets(), w.X, w.Y, x, y, fhFrame, fhPanel)
}

// frameWidgets is _widgets after onResize: the panel from y 15, the list (4,
// 19) to width − 4, height − 20, reset_keys_btn width − 166 … width − 16,
// height − 16 … height − 4.
func (w *KeyboardShortcutsWindow) frameWidgets() []BVWidget {
	return append(fhCommon(w.W, w.H, 15), BVWidget{L: 4, T: 19, R: w.W - 4, B: w.H - 20},
		BVWidget{L: w.W - 150 - 4 - 12, T: w.H - 4 - 12, R: w.W - 4 - 12, B: w.H - 4})
}

// OnFrame: findWidgetAt's frame or panel at the point.
func (w *EditShortcutWindow) OnFrame(x, y int) bool {
	return fhBare(w.frameWidgets(), w.X, w.Y, x, y, fhFrame, fhPanel)
}

// frameWidgets is _widgets (280 × 78, never resized): the panel from y 15,
// the three centred labels (4, 20 / 34 / 54, width − 8 × 12).
func (w *EditShortcutWindow) frameWidgets() []BVWidget {
	lw := append(fhCommon(w.W, w.H, 15)[:2], fhW(265, 2, 13, 13), fhW(0, 15, w.W, w.H-15))
	for _, y := range [3]int{20, 34, 54} {
		lw = append(lw, fhW(4, y, w.W-8, 12))
	}
	return lw
}

// --- Scenario select (ScenarioSelect.cpp) ---

// OnFrame: findWidgetAt's frame or caption at the point. The panel is a
// Wt3Widget, not a panel: a press on it doesn't move the window; the
// caption is 34 high, under and between the tabs.
func (w *ScenarioSelectWindow) OnFrame(x, y int) bool {
	return fhBare(w.frameWidgets(), w.X, w.Y, x, y, 0, 1)
}

// frameWidgets is _widgets (610 × 412, never resized) with initTabs: the
// categories without a scenario hidden, the others packed from x 3, 91 a
// tab.
func (w *ScenarioSelectWindow) frameWidgets() []BVWidget {
	lw := []BVWidget{fhW(0, 0, w.W, w.H), fhW(1, 1, 608, 34), fhW(595, 2, 13, 13), fhW(0, 48, 610, 364)}
	x := 3
	for i := range 5 {
		if len(w.byCat[i]) == 0 {
			lw = append(lw, fhHidden(fhW(ssTabs[i][0], 15, 91, 34), true))
			continue
		}
		lw = append(lw, fhW(x, 15, 91, 34))
		x += 91
	}
	return append(lw, fhW(ssList[0], ssList[1], ssList[2], ssList[3]))
}

// --- File browser (PromptBrowseWindow.cpp) ---

// OnFrame: findWidgetAt's frame or panel at the point.
func (w *PromptBrowseWindow) OnFrame(x, y int) bool {
	return fhBare(w.frameWidgets(), w.X, w.Y, x, y, fhFrame, fhPanel)
}

// frameWidgets is the widgets after prepareDraw (0x00445C8F): the panel from
// y 15; folder_path from (3, 18) to parent_button's left; parent_button and
// home_button 25 wide from width − 3 leftwards, 18 … 41; text_filename and
// ok_button in save mode only; the scroll view (3, 45) to width − 259
// (+ 122 when not browsing saved games), height − 34 (save) or − 4.
func (w *PromptBrowseWindow) frameWidgets() []BVWidget {
	W, H := w.W, w.H
	homeR := W - 3
	homeL := homeR - 24
	parentL := homeL - 24
	right := W - 259
	saved := false
	for _, e := range w.Exts {
		saved = saved || e == ".sv5" // BrowseFileType::savedGame
	}
	if !saved {
		right += 122
	}
	bottom := H - 4
	if w.Save {
		bottom = H - 34
	}
	return append(fhCommon(W, H, 15), BVWidget{L: 3, T: 18, R: parentL, B: 41}, BVWidget{L: parentL, T: 18, R: homeL, B: 41}, BVWidget{L: homeL, T: 18, R: homeR, B: 41},
		fhHidden(BVWidget{L: 88, T: H - 31, R: W - 4, B: H - 18}, !w.Save), fhHidden(BVWidget{L: W - 86, T: H - 15, R: W - 16, B: H - 4}, !w.Save),
		BVWidget{L: 3, T: 45, R: right, B: bottom})
}

// --- Terraform (TerraForm.cpp) ---

// OnFrame: findWidgetAt's frame or panel at the point (the tool area is a
// Wt3Widget: it takes the press).
func (w *TerraformWindow) OnFrame(x, y int) bool {
	return fhBare(w.frameWidgets(), w.X, w.Y, x, y, fhFrame, fhPanel)
}

// frameWidgets is the tab's widgets after prepareDraw: Common::
// makeCommonWidgets (five tabs, leftAlignTabs: none disabled) resized by
// Common::prepareDraw, then
//   - Clear area, Adjust water: tool_area (49, 45, 64 × 44), decrease_area,
//     increase_area;
//   - Adjust land (prepareDraw): those, mountain_mode,
//     paint_mode, land_material (hidden outside paint mode);
//   - Plant trees (prepareDraw): the scroll view (3, 45) to
//     width − 26, height − 14; rotate_object (hidden unless the chosen tree
//     has rotations), object_colour (hidden unless it has colours), the two
//     cluster buttons, each width − 25 … width − 2;
//   - Build walls: the scroll view (2, 45) to width − 4, height − 14.
func (w *TerraformWindow) frameWidgets() []BVWidget {
	lw := fhTabs(fhCommon(w.W, w.H, 41), tabLayout(5, nil))
	rect := func(r [4]int) BVWidget { return fhW(r[0], r[1], r[2], r[3]) }
	switch w.Tab {
	case TerraformTabClear, TerraformTabWater:
		return append(lw, rect(tfToolArea), rect(tfDecrease), rect(tfIncrease))
	case TerraformTabLand:
		return append(lw, rect(tfToolArea), rect(tfDecrease), rect(tfIncrease), rect(tfMountain), rect(tfPaint), fhHidden(rect(tfLandMaterial), !w.PaintMode))
	case TerraformTabTrees:
		sel := w.selectedCell()
		lw = append(lw, BVWidget{L: 3, T: 45, R: w.W - 26, B: w.H - 14})
		for i := range 4 {
			hidden := i == 0 && (sel == nil || !sel.Rotatable) || i == 1 && (sel == nil || sel.Colours == 0)
			lw = append(lw, fhHidden(BVWidget{L: w.W - 25, T: 46 + 24*i, R: w.W - 2, B: 46 + 24*i + 23}, hidden))
		}
		return lw
	}
	return append(lw, BVWidget{L: 2, T: 45, R: w.W - 4, B: w.H - 14})
}

// --- Map (MapWindow.cpp) ---

// OnFrame: findWidgetAt's frame or panel at the point (the key drawn on the
// panel right of the map has no widget: a press there moves the window).
func (w *MapWindow) OnFrame(x, y int) bool {
	return fhBare(w.frameWidgets(), w.X, w.Y, x, y, fhFrame, fhPanel)
}

// frameWidgets is kWidgets after prepareDraw: the five tabs (leftAlignTabs:
// none disabled outside the editor), the scroll view (3, 44) to width −
// 108, height − 14, the status bar (3, height − 12) to width − 14, height −
// 3. goloco doesn't draw the tabs and the status bar yet (PARTIAL in
// mapwindow.go); a press there is theirs all the same.
func (w *MapWindow) frameWidgets() []BVWidget {
	lw := fhTabs(fhCommon(w.W, w.H, 41), tabLayout(5, nil))
	return append(lw, BVWidget{L: 3, T: 44, R: w.W - 108, B: w.H - 14}, BVWidget{L: 3, T: w.H - 12, R: w.W - 14, B: w.H - 3})
}

// --- News (News/News.cpp, News.h) ---

// OnFrame: findWidgetAt's frame — the Panel (News1) or NewsPanel (News2) —
// at the point.
func (w *NewsWindow) OnFrame(x, y int) bool {
	return fhBare(w.frameWidgets(), w.X, w.Y, x, y, fhFrame)
}

// frameWidgets is Common::makeCommonWidgets after initViewports: the frame,
// close_button, the two viewports (hidden without a view: no subject, a
// face, a vehicle image) and the two subject buttons (hidden when not
// selectable), each at its column (kViewportLayouts).
func (w *NewsWindow) frameWidgets() []BVWidget {
	lw := []BVWidget{fhW(0, 0, w.W, w.H), {L: w.W - 15, T: 2, R: w.W - 3, B: 14}}
	for i := range 2 {
		l, t, r, b := w.viewport(i)
		lw = append(lw, fhHidden(BVWidget{L: l, T: t, R: r, B: b}, w.Subjects[i].Kind != NewsSubjectView))
	}
	for i := range 2 {
		l, t, r, b := w.button(i)
		lw = append(lw, fhHidden(BVWidget{L: l, T: t, R: r, B: b}, !w.Subjects[i].Selectable))
	}
	return lw
}
