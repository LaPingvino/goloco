package ui

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/jukebox"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// frameCase is a window-relative point and whether findWidgetAt finds the
// frame or the panel there (MouseInput stateNormalLeft: a move).
type frameCase struct {
	name string
	x, y int
	bare bool
}

func checkFrame(t *testing.T, win string, w Win, cases []frameCase) {
	t.Helper()
	b := w.Base()
	f, ok := w.(FrameHitter)
	if !ok {
		t.Fatalf("%s: no OnFrame", win)
	}
	for _, c := range cases {
		x, y := c.x, c.y
		if x < 0 {
			x += b.W
		}
		if y < 0 {
			y += b.H
		}
		if got := f.OnFrame(b.X+x, b.Y+y); got != c.bare {
			t.Errorf("%s: %s (%d, %d): bare %v, want %v", win, c.name, x, y, got, c.bare)
		}
	}
}

// dragBy presses at a window-relative point, drags by (dx, dy) and reports
// how far the window moved.
func dragBy(s *WindowStack, w Win, x, y, dx, dy int) (mx, my int) {
	b := w.Base()
	ox, oy := b.X, b.Y
	s.Press(b.X+x, b.Y+y)
	s.Held(ox+x+dx, oy+y+dy)
	s.Release(ox+x+dx, oy+y+dy)
	return b.X - ox, b.Y - oy
}

func TestFrameHitStation(t *testing.T) {
	w := NewStationWindow(800, 600, "Station", "", 0, 1)
	w.X, w.Y = 100, 100
	checkFrame(t, "station", w, []frameCase{
		{"the viewport", 50, 60, false},
		{"the status bar", 10, -6, false},
		{"the centre button", -10, -20, false},
		{"the panel under the tabs", 150, 42, true},
		{"the panel right of the viewport", -3, 60, true},
		{"past the viewport's last row", 50, -13, true},
		{"the grip corner", -2, -2, true},
		{"a hidden vehicle tab's slot", 100, 20, true},
		{"the caption", 50, 5, false},
	})
	w.VehicleTypes[0] = true // trains call: their tab at 96
	checkFrame(t, "station", w, []frameCase{{"the trains tab", 100, 20, false}})
	w.Tab = 1
	w.PrepareDraw()
	checkFrame(t, "station cargo", w, []frameCase{
		{"station_catchment", -10, 50, false},
		{"between the list and the catchment button", -26 + 1, 100, true},
	})
	w.Tab = 3
	w.PrepareDraw()
	checkFrame(t, "station trains", w, []frameCase{
		{"the status bar to 402", -5, -5, false},
		{"under the status bar", -5, -2, true},
		{"the list", 100, 100, false},
	})

	s := &WindowStack{ScreenW: 1024, ScreenH: 768}
	st := NewStationWindow(1024, 768, "Station", "", 0, 1)
	st.X, st.Y = 100, 100
	s.Open(st)
	if dx, dy := dragBy(s, st, 150, 42, 30, 20); dx != 30 || dy != 20 {
		t.Errorf("dragged by the panel: moved %d,%d", dx, dy)
	}
	if dx, dy := dragBy(s, st, 10, st.H-6, 30, 20); dx != 0 || dy != 0 {
		t.Errorf("a press on the status bar moved the window %d,%d", dx, dy)
	}
	if dx, dy := dragBy(s, st, st.W-2, st.H-2, 40, 30); dx != 0 || dy != 0 || st.W != 263 || st.H != 166 {
		t.Errorf("the grip: moved %d,%d, size %dx%d", dx, dy, st.W, st.H)
	}
}

func TestFrameHitLists(t *testing.T) {
	town := NewTownList(800, nil, 0, 1)
	town.WinType = WindowTypeTownList
	town.TabDisabled = func(i int) bool { return i > 0 }
	town.Resizable(192, 100, 600, 900)
	checkFrame(t, "town list", town, []frameCase{
		{"a header", 10, 48, false},
		{"right of the headers", 500, 48, true},
		{"the list", 50, 100, false},
		{"the list's last row (height − 14)", 50, -14, false},
		{"the status bar", 10, -6, false},
		{"the grip corner", -2, -2, true},
		{"a hidden build tab's slot", 40, 20, true},
	})

	st := NewStationList(800, nil, 0, 1)
	st.WinType = WindowTypeStationList
	checkFrame(t, "station list", st, []frameCase{
		{"company_select", -10, 20, false},
		{"the last tab", 3 + 4*31 + 5, 20, false},
		{"past the tabs", 3 + 5*31 + 5, 20, true},
	})

	ind := NewIndustryList(800, nil, 0, 1)
	ind.WinType = WindowTypeIndustryList
	checkFrame(t, "industry list", ind, []frameCase{
		{"above the headers (y 44)", 10, 43, true},
		{"a header", 10, 44, false},
		{"the second tab", 40, 20, false},
	})
	grid := NewIndustryGrid(800, nil, -1, 0, 1, nil)
	grid.WinType = WindowTypeIndustryList
	checkFrame(t, "new industries", grid, []frameCase{
		{"the grid", 10, 50, false},
		{"the cost line", 10, -10, true},
	})

	vl := NewVehicleList(800, 0, nil, 0, 1, nil)
	vl.WinType = WindowTypeVehicleList
	checkFrame(t, "vehicle list", vl, []frameCase{
		{"the cargo dropdown's last row", -50, -3, false},
		{"the filter dropdown", -200, -8, false},
		{"the grip corner", -2, -2, true},
		{"the list", 100, 100, false},
	})
	vl.W = 300 // no room for the filters: hidden
	checkFrame(t, "narrow vehicle list", vl, []frameCase{{"where the cargo dropdown was", -50, -3, true}})

	other := NewTownList(800, nil, 0, 1) // no list type: the caption only
	if other.OnFrame(other.X+500, other.Y+48) {
		t.Error("a list of no known type reports its frame")
	}
}

func TestFrameHitCompany(t *testing.T) {
	w := &CompanyFinancesWindow{WinBase: WinBase{WinType: WindowTypeCompany, X: 100, Y: 100}}
	w.switchTab(CompanyTabStatus)
	checkFrame(t, "company status", w, []frameCase{
		{"the viewport", 20, 60, false},
		{"the face", -60, 80, false},
		{"another company's owner name (hidden)", -60, 135, true},
		{"the status label", 10, -6, false},
		{"the grip corner", -2, -2, true},
		{"the challenge tab of another company (hidden)", 3 + 5*31 + 5, 20, true},
		{"company_select", -10, 20, false},
	})
	w.Status.Own, w.Data.Own = true, true
	checkFrame(t, "own company status", w, []frameCase{
		{"change_owner_name", -60, 135, false},
		{"the challenge tab", 3 + 5*31 + 5, 20, false},
	})
	w.switchTab(CompanyTabChallenge)
	checkFrame(t, "challenge", w, []frameCase{
		{"where company_select was", -10, 20, true},
		{"the text", 50, 100, true},
	})
	w.switchTab(CompanyTabFinances)
	checkFrame(t, "finances", w, []frameCase{
		{"the table", 200, 100, false},
		{"the labels left of it", 50, 100, true},
		{"the loan +", 87 + 100 - 5, 268, false},
		{"autopay", 330, 268, false},
	})
	w.Data.Own = false
	checkFrame(t, "another's finances", w, []frameCase{{"the hidden loan box", 90, 268, true}})
	w.switchTab(CompanyTabColourScheme)
	checkFrame(t, "colour scheme", w, []frameCase{
		{"the main swatch", 225, 52, false},
		{"a checkbox", 20, 85, false},
		{"an unset type's swatch (hidden)", 225, 82, true},
	})
	w.Colours.Custom = 1 << 1
	checkFrame(t, "colour scheme", w, []frameCase{{"the steam locomotives' swatch", 225, 82, false}})
	w.switchTab(CompanyTabDetails)
	checkFrame(t, "details", w, []frameCase{
		{"the viewport", 250, 100, false},
		{"another company's build_hq (hidden)", 320, 100, true},
		{"the text", 20, 100, true},
	})
}

func TestFrameHitCompanyList(t *testing.T) {
	s := &WindowStack{ScreenW: 1024, ScreenH: 768}
	w := NewCompanyList(0, 1, nil)
	w.X, w.Y = 100, 100
	s.Open(w)
	checkFrame(t, "company list", w, []frameCase{
		{"the list", 300, 100, false},
		{"a header", 300, 48, false},
	})
	w.switchTab(1)
	checkFrame(t, "performance", w, []frameCase{{"the graph", 300, 150, true}})
	if dx, dy := dragBy(s, w, 300, 150, -20, 10); dx != -20 || dy != 10 {
		t.Errorf("dragged by the graph: moved %d,%d", dx, dy)
	}
}

func TestFrameHitMessagesOptions(t *testing.T) {
	m := NewMessagesWindow(800, 0, 1)
	checkFrame(t, "messages", m, []frameCase{
		{"the list", 100, 100, false},
		{"the grip corner", -2, -2, true},
	})
	m.Tab = 1
	m.W, m.H = messagesW, messageSettingsH
	checkFrame(t, "message settings", m, []frameCase{
		{"a label", 10, 50, false},
		{"between the label and the dropdown", 234, 50, true},
		{"a dropdown's button", 350, 50, false},
		{"the sound checkbox", 10, 140, false},
	})

	o := NewOptions(800, 600, 0, 1)
	for tab := range 7 {
		o.Tab = tab
		o.PrepareDraw()
		checkFrame(t, "options", o, []frameCase{{"the panel's left edge", 1, 60, true}, {"a tab", 3 + 31*tab + 5, 20, false}})
	}
	o.Tab = OptTabDisplay
	o.PrepareDraw()
	checkFrame(t, "options display", o, []frameCase{
		{"inside the hardware group box", 200, 140, false},
		{"under the group box", 200, 147, true},
	})
	o.Tab = OptTabControls
	o.PrepareDraw()
	checkFrame(t, "options controls", o, []frameCase{
		{"customize keys", 30, 128, false},
		{"right of it", 250, 128, true},
	})
}

func TestFrameHitMusicShortcuts(t *testing.T) {
	j := NewMusicJukebox(800, 600, 0, 1)
	checkFrame(t, "jukebox", j, []frameCase{
		{"the panel", 300, 50, true},
		{"edit selection", 200, 90, false},
		{"play", 40, 50, false},
		{"the frame between the caption and the panel", 5, 14, true},
	})
	ms := NewMusicSelection(&jukebox.Config{}, 0, 1)
	checkFrame(t, "music selection", ms, []frameCase{
		{"the list", 50, 100, false},
		{"the years header", -10, 20, false},
		{"the grip corner", -2, -2, true},
		{"the status bar", 10, -6, false},
	})
	s := &WindowStack{ScreenW: 1024, ScreenH: 768}
	ks := NewKeyboardShortcuts(1024, 768, 0, 1)
	s.Open(ks)
	checkFrame(t, "keyboard shortcuts", ks, []frameCase{
		{"the list", 100, 100, false},
		{"reset keys (151 wide, to width − 16)", -16, -8, false},
		{"right of reset keys", -15, -8, true},
	})
	if dx, dy := dragBy(s, ks, ks.W-2, ks.H-2, 50, 40); dx != 0 || dy != 0 || ks.W != 470 || ks.H != 278 {
		t.Errorf("keyboard shortcuts' grip: moved %d,%d, %dx%d", dx, dy, ks.W, ks.H)
	}
	e := NewEditShortcut(800, 600, 0, "x", 0, 1)
	checkFrame(t, "edit shortcut", e, []frameCase{{"a label", 10, 25, false}, {"under the labels", 10, 70, true}})
}

func TestFrameHitScenarioBrowse(t *testing.T) {
	sc := NewScenarioSelect([]*scenario.ScenarioMeta{{Category: 1}}, 800, 600)
	checkFrame(t, "scenario select", sc, []frameCase{
		{"the caption beside the tabs", 500, 25, true},
		{"the only category's tab, packed to x 3", 10, 30, false},
		{"where the second tab would be", 100, 30, true},
		{"the panel (a Wt3)", 500, 200, false},
		{"the list", 100, 200, false},
		{"the frame under the caption", 500, 40, true},
		{"the frame's left column (under the Wt3)", 0, 200, false},
	})
	pb := &PromptBrowseWindow{WinBase: WinBase{W: pbW, H: pbH}, Exts: []string{".sv5"}}
	checkFrame(t, "load game", pb, []frameCase{
		{"the preview right of the list", 300, 100, true},
		{"the list", 100, 100, false},
		{"the home button", -10, 30, false},
		{"the folder line", 50, 30, false},
		{"where save mode's text box is", 100, -25, false}, // the list, down to height − 4
	})
	pb.Exts, pb.Save = []string{".sc5"}, true
	checkFrame(t, "save landscape", pb, []frameCase{
		{"the wider list", 300, 100, false},
		{"the text box", 100, -25, false},
		{"left of the text box", 50, -25, true},
	})
}

func TestFrameHitTerraformMapNews(t *testing.T) {
	tf := NewTerraformWindow(800)
	tf.SwitchTab(TerraformTabClear)
	checkFrame(t, "clear area", tf, []frameCase{
		{"the tool area (a Wt3)", 70, 80, false},
		{"beside the tool area", 20, 80, true},
	})
	tf.SwitchTab(TerraformTabTrees)
	checkFrame(t, "plant trees", tf, []frameCase{
		{"rotate, no tree chosen (hidden)", -10, 50, true},
		{"cluster selected", -10, 100, false},
	})
	tf.Cells, tf.Selected = []TerraformCell{{ID: 4, Rotatable: true}}, 4
	checkFrame(t, "plant trees", tf, []frameCase{{"rotate", -10, 50, false}})

	m := NewMapWindow(nil, 0, 1)
	checkFrame(t, "map", m, []frameCase{
		{"the key right of the map", -50, 100, true},
		{"the map", 50, 100, false},
		{"the status bar", 50, -6, false},
		{"the tabs", 130, 20, false},
		{"the grip corner", -2, -2, true},
	})

	n := NewNewsWindow(true, true, true, 800, 600)
	checkFrame(t, "news", n, []frameCase{
		{"the paper", 100, 30, true},
		{"no subject", 100, -30, true},
		{"the close button", -10, 5, false},
	})
	n.Subjects[0] = NewsSubject{Kind: NewsSubjectView, Present: true}
	checkFrame(t, "news", n, []frameCase{{"a viewport without its button", 100, -30, false}})
	n.Subjects[0] = NewsSubject{Kind: NewsSubjectFace, Present: true, Selectable: true}
	checkFrame(t, "news", n, []frameCase{{"a face's button", 100, -30, false}})
}
