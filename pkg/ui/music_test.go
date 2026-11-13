package ui

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/jukebox"
)

// A click on a row toggles its track (onScrollMouseDown); the last selected
// track stays; the headers cycle the sort order.
func TestMusicSelection(t *testing.T) {
	cfg := jukebox.Config{}
	cfg.CustomJukebox[0], cfg.CustomJukebox[2] = true, true
	changes := 0
	w := NewMusicSelection(&cfg, 0, 1)
	w.X, w.Y = 100, 50
	w.OnChange = func() { changes++ }
	row := func(i int) (int, int) { return w.X + 60, w.Y + 30 + 1 + i*msRowHeight + 5 }

	w.HandleClick(row(1))
	if !cfg.CustomJukebox[1] || changes != 1 || w.selected() != 3 {
		t.Fatalf("row 1 on: %v, changes %d", cfg.CustomJukebox[:3], changes)
	}
	for _, i := range []int{0, 1, 2} {
		w.HandleClick(row(i))
	}
	if !cfg.CustomJukebox[2] || w.selected() != 1 {
		t.Errorf("the last track must stay selected: %v", cfg.CustomJukebox[:3])
	}

	// Years: descending first, then ascending, then the original order.
	sy := w.sortYears()
	w.HandleClick(w.X+sy[0]+5, w.Y+sy[1]+5)
	if w.sortMode != jukebox.SortYearsDescending || w.playlist[0] != 29 {
		t.Errorf("years, first click: mode %d, first %d", w.sortMode, w.playlist[0])
	}
	w.HandleClick(w.X+sy[0]+5, w.Y+sy[1]+5)
	w.HandleClick(row(0)) // Eugenia (track 13) is first by years ascending
	if w.sortMode != jukebox.SortYearsAscending || !cfg.CustomJukebox[13] {
		t.Errorf("years ascending: mode %d, eugenia %v", w.sortMode, cfg.CustomJukebox[13])
	}
	w.HandleClick(w.X+sy[0]+5, w.Y+sy[1]+5)
	st := w.sortTitle()
	w.HandleClick(w.X+st[0]+5, w.Y+st[1]+5)
	if w.sortMode != jukebox.SortTitleAscending {
		t.Errorf("title: mode %d", w.sortMode)
	}

	// onResize: the years header keeps to the right edge, the list to the frame.
	w.W, w.H = 500, 400
	if sy, l := w.sortYears(), w.list(); sy[0] != 500-1-4-75+1 || l != [4]int{4, 30, 492, 400 - 13 - 30} {
		t.Errorf("resized: years %v, list %v", sy, l)
	}
}

// drawScroll's years: a range, an open end, an open start (right-aligned),
// and the title music, which never plays in an era.
func TestMusicYears(t *testing.T) {
	for _, c := range []struct {
		track jukebox.MusicID
		start bool
	}{{0, true}, {8, true}, {13, false}, {29, false}} {
		text, right := musicYears(jukebox.Info(c.track))
		mi := jukebox.Info(c.track)
		if right != (c.track == 13) {
			t.Errorf("track %d (%d–%d): %q right %v", c.track, mi.StartYear, mi.EndYear, text, right)
		}
	}
}

// The jukebox window's widgets: the controls work in play mode only, Edit
// selection with the custom playlist only.
func TestMusicJukeboxClicks(t *testing.T) {
	w := NewMusicJukebox(800, 600, 0, 1)
	var controls []int
	edits := 0
	w.OnControl = func(i int) { controls = append(controls, i) }
	w.OnEditSelection = func() { edits++ }
	click := func(r [4]int) { w.HandleClick(w.X+r[0]+r[2]/2, w.Y+r[1]+r[3]/2) }
	click(mjControls[0])
	click(mjEdit)
	if len(controls) != 0 || edits != 0 {
		t.Fatalf("title mode, current-era playlist: controls %v, edits %d", controls, edits)
	}
	w.State = MusicJukeboxState{PlayMode: true, Playlist: 2}
	for _, r := range mjControls {
		click(r)
	}
	click(mjEdit)
	if len(controls) != 3 || controls[0] != 0 || controls[1] != 1 || controls[2] != 2 || edits != 1 {
		t.Errorf("play mode: controls %v, edits %d", controls, edits)
	}
	// The dropdowns open at their comboboxes, as wide as them less 4.
	var d *DropdownMenu
	w.OpenDropdown = func(m *DropdownMenu) { d = m }
	chosen := -1
	w.OnPlaylist = func(i int) { chosen = i }
	click(mjPlaylist)
	if d == nil || len(d.Items) != 3 || !d.Items[2].Selected || d.ItemWidth != 346-4 || d.Y != w.Y+69+12 {
		t.Fatalf("playlist dropdown: %+v", d)
	}
	d.OnSelect(1)
	if chosen != 1 {
		t.Errorf("playlist chosen: %d", chosen)
	}
}

// The Audio menu's ticked items are 10 px wider than their text
// (dropdown_without_checkmark's MOVE_X 10).
func TestDropdownCheckmarkWidth(t *testing.T) {
	plain := NewDropdownMenu(0, 0, []DropdownItem{{Text: "Mute"}})
	ticked := NewDropdownMenu(0, 0, []DropdownItem{{Text: "Mute", Checkmark: true}})
	if ticked.ItemWidth != plain.ItemWidth+10 {
		t.Errorf("widths %d and %d", plain.ItemWidth, ticked.ItemWidth)
	}
}
