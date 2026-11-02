package ui

import (
	"github.com/LaPingvino/goloco/pkg/scenario"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/hajimehoshi/ebiten/v2"
)

// PromptBrowse — a port of Ui/Windows/PromptBrowseWindow.cpp in load mode:
// 500×380, centred, black/mutedSeaGreen; caption, close button, panel, parent
// folder button, "Folder:" line, the file list (a ScrollView, 11 px rows,
// folders first with icon_folder) and the selected file's name on the right.
// Save mode (NewPromptBrowseSave): the "File name:" text box (TextBox at
// (88, h-31)…(w-4, h-18), text from (box+1, box+2), a 1×9 vertical caret in
// shade 9 blinking every 16 updates), the OK button at (w-86, h-15)…(w-16, h-4),
// and the list ending at h-34. Enter = OK, Escape = close.
// A file row picks on one click, as onScrollMouseDown; the mouse-over row is
// the selected one (onScrollMouseOver).
// PARTIAL: the text box edits only at its end (no caret movement, no scrolling
// offset); overwrite confirmation; right-click delete; the saved-game / landscape
// preview pane (drawSavePreview), root-path drives, the folder line's "..."
// shortening, window resizing.

const (
	pbW, pbH        = 500, 380
	pbRowHeight     = 11
	imgParentFolder = 2310
	imgFolder       = 2311
)

var (
	pbCaption = [4]int{1, 1, 498, 13}
	pbClose   = [4]int{485, 2, 13, 13}
	pbPanel   = [4]int{0, 15, 500, 365}
	pbParent  = [4]int{473, 18, 24, 24}
	// scrollview: left 3, top 45; right = width - 259 (saved games), bottom = height - 4.
	pbList = [4]int{3, 45, (pbW - 259) - 3 + 1, (pbH - 4) - 45 + 1}
)

// PromptBrowseWindow is the window state.
type PromptBrowseWindow struct {
	X, Y     int
	Title    string
	Dir      string
	Exts     []string // lower-case extensions listed, e.g. ".sv5"
	entries  []pbEntry
	Selected int
	Scroll   int
	dragging bool // the scrollbar's thumb is being dragged
	dragY    int
	content  [4]int
	OnPick   func(path string)
	OnClose  func()

	// Save mode: the filename being typed, the extension added on OK, the
	// caret blink counter and the handler.
	Save    bool
	Input   []rune
	SaveExt string
	frame   int
	OnSave  func(path string)
	// Preview, if set, is shown for any selected file (diagnostics);
	// otherwise a saved game's own details are read (loadFileDetails).
	Preview     *scenario.SaveDetails
	detailsPath string
	detailsData *scenario.SaveDetails
	// Confirm asks before replacing an existing file (PromptOkCancel:
	// caption, text, OK label, then the save); nil saves straight away.
	Confirm func(caption, text, ok string, then func())
}

// NewPromptBrowseSave opens the window in save mode with name in the text box.
func NewPromptBrowseSave(title, dir, name, ext string, sw, sh int) *PromptBrowseWindow {
	w := NewPromptBrowse(title, dir, []string{ext}, sw, sh)
	w.Save, w.Input, w.SaveExt = true, []rune(name), ext
	return w
}

// list is the scroll view rect: bottom at h-4, or h-34 in save mode.
func (w *PromptBrowseWindow) list() [4]int {
	l := pbList
	if w.Save {
		l[3] = (pbH - 34) - l[1] + 1
	}
	return l
}

var (
	pbTextBox = [4]int{88, pbH - 31, (pbW - 4) - 88 + 1, (pbH - 18) - (pbH - 31) + 1}
	pbOK      = [4]int{pbW - 86, pbH - 15, (pbW - 16) - (pbW - 86) + 1, (pbH - 4) - (pbH - 15) + 1}
)

// HandleHover selects the row under the mouse (onScrollMouseOver).
func (w *PromptBrowseWindow) HandleHover(px, py int) {
	l := w.list()
	if !w.in(l, px, py) {
		return
	}
	if idx := (py - (w.Y + l[1] + 1) + w.Scroll) / pbRowHeight; idx >= 0 && idx < len(w.entries) {
		w.Selected = idx
	}
}

// HandleText types characters into the filename (save mode); backspace,
// enter and escape as keyUp.
func (w *PromptBrowseWindow) HandleText(chars []rune, backspace, enter, escape bool) {
	w.frame++
	switch {
	case escape:
		if w.OnClose != nil {
			w.OnClose()
		}
		return
	case enter && w.Save:
		w.ok()
		return
	}
	if !w.Save {
		return
	}
	if backspace && len(w.Input) > 0 {
		w.Input = w.Input[:len(w.Input)-1]
		w.frame = 0
	}
	for _, r := range chars {
		if r >= 32 && len(w.Input) < 200 { // InputSession(…, 200)
			w.Input = append(w.Input, r)
			w.frame = 0
		}
	}
}

// ok is processFileForLoadSave from the OK button: the typed name in the
// current folder, with the file type's extension.
func (w *PromptBrowseWindow) ok() {
	name := strings.TrimSpace(string(w.Input))
	if name == "" || strings.ContainsAny(name, `/\:*?"<>|`) { // filenameContainsInvalidChars
		return
	}
	w.save(filepath.Join(w.Dir, name+w.SaveExt))
}

// save is processFileForLoadSave in save mode: an existing file is replaced
// only after "Replace existing file: “name”?" is confirmed.
func (w *PromptBrowseWindow) save(path string) {
	if w.OnSave == nil {
		return
	}
	if _, err := os.Stat(path); err == nil && w.Confirm != nil {
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		w.Confirm(w.Title, "Replace existing file: “"+name+"”?", "Replace", func() { w.OnSave(path) })
		return
	}
	w.OnSave(path)
}

type pbEntry struct {
	path string
	dir  bool
}

// NewPromptBrowse opens the window (createWindowCentred) on dir.
func NewPromptBrowse(title, dir string, exts []string, sw, sh int) *PromptBrowseWindow {
	w := &PromptBrowseWindow{X: (sw - pbW) / 2, Y: (sh - pbH) / 2, Title: title, Exts: exts, Selected: -1}
	w.chdir(dir)
	return w
}

func (w *PromptBrowseWindow) chdir(dir string) {
	w.Dir, w.Selected, w.Scroll, w.entries = dir, -1, 0, nil
	ents, _ := os.ReadDir(dir)
	var dirs, files []pbEntry
	for _, e := range ents {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		p := filepath.Join(dir, e.Name())
		if e.IsDir() {
			dirs = append(dirs, pbEntry{p, true})
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		for _, x := range w.Exts {
			if ext == x {
				files = append(files, pbEntry{p, false})
				break
			}
		}
	}
	by := func(s []pbEntry) {
		sort.Slice(s, func(i, j int) bool { return strings.ToLower(s[i].path) < strings.ToLower(s[j].path) })
	}
	by(dirs)
	by(files)
	w.entries = append(dirs, files...)
}

func (w *PromptBrowseWindow) in(r [4]int, px, py int) bool {
	return px >= w.X+r[0] && px < w.X+r[0]+r[2] && py >= w.Y+r[1] && py < w.Y+r[1]+r[3]
}

// Contains reports whether the point is on the window.
func (w *PromptBrowseWindow) Contains(px, py int) bool { return w.in([4]int{0, 0, pbW, pbH}, px, py) }

// HandleClick: close, parent folder, or a list row (select; a second click
// on the selected row opens the folder or picks the file).
func (w *PromptBrowseWindow) HandleClick(px, py int) {
	switch {
	case w.in(pbClose, px, py):
		if w.OnClose != nil {
			w.OnClose()
		}
	case w.in(pbParent, px, py):
		if parent := filepath.Dir(w.Dir); parent != w.Dir {
			w.chdir(parent)
		}
	case w.Save && w.in(pbOK, px, py):
		w.ok()
	case w.onScrollbar(px, py):
		// The scrollbar (ScrollView): buttons and track move the list, the
		// thumb starts a drag (Drag).
		l := w.list()
		switch part := L.VScrollPart(py-(w.Y+l[1]), l[3], len(w.entries)*pbRowHeight, w.Scroll); part {
		case "thumb":
			w.dragY, w.dragging = py, true
		default:
			w.Scroll = L.VScrollClick(part, l[3], len(w.entries)*pbRowHeight, w.Scroll)
		}
	case w.in(w.list(), px, py):
		l := w.list()
		idx := (py - (w.Y + l[1] + 1) + w.Scroll) / pbRowHeight
		if idx < 0 || idx >= len(w.entries) {
			return
		}
		e := w.entries[idx]
		switch {
		case e.dir:
			w.chdir(e.path)
		case w.Save:
			w.save(e.path) // asks before replacing it
		case w.OnPick != nil:
			w.OnPick(e.path)
		}
	}
}

// HandleWheel scrolls the list.
func (w *PromptBrowseWindow) HandleWheel(dy float64) {
	maxScroll := max(len(w.entries)*pbRowHeight-w.list()[3], 0)
	w.Scroll = min(max(w.Scroll-int(dy*3*pbRowHeight), 0), maxScroll)
}

func stem(p string) string { b := filepath.Base(p); return strings.TrimSuffix(b, filepath.Ext(b)) }

// Draw draws the window.
func (w *PromptBrowseWindow) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	pal := g1.GetPalette()
	prim, _ := g1.GetPaletteMap(0)
	sec, _ := g1.GetPaletteMap(colourMutedSeaGreen)
	black, win2 := pal[paletteBlack0], pal[gfx8.Shade(sec, 9)]
	DrawWindowFrame(screen, w.X, w.Y, pbW, pbH, 0, false)
	draw8(screen, w.X, w.Y, pbW, pbH, func(rt *gfx8.RenderTarget) {
		p := pbPanel
		gfx8.FillRectInset(rt, p[0], p[1], p[0]+p[2]-1, p[1]+p[3]-1, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, pbClose[0], pbClose[1], gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		if sp, err := gfx8.Decode(&g1.Elements[imgParentFolder]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, pbParent[0], pbParent[1], gfx8.OpTransparent|gfx8.OpSrc, sec, nil)
		}
		l := w.list()
		cx, cy, cw, ch, _ := drawVScrollView(rt, l[0], l[1], l[2], l[3], gfx8.InsetColour{Map: sec}, len(w.entries)*pbRowHeight, w.Scroll)
		w.content = [4]int{cx, cy, cw, ch}
		gfx8.FillRect(rt, cx, cy, cx+cw-1, cy+ch-1, gfx8.Shade(sec, 4))
		m30, _ := g1.GetPaletteMap(0x30)
		fold, ferr := gfx8.Decode(&g1.Elements[imgFolder])
		for i, e := range w.entries {
			y := cy + i*pbRowHeight - w.Scroll
			if y+pbRowHeight <= cy || y >= cy+ch {
				continue
			}
			if i == w.Selected {
				gfx8.FillRectTranslucent(rt, cx, max(y, cy), cx+cw-1, min(y+pbRowHeight-1, cy+ch-1), m30)
			}
			if e.dir && ferr == nil {
				gfx8.DrawSpriteNoise(rt, fold, cx+1, y, gfx8.OpFor(fold, false, false), nil, nil)
			}
		}
		if w.Save {
			b, o := pbTextBox, pbOK
			gfx8.FillRectInset(rt, b[0], b[1], b[0]+b[2]-1, b[1]+b[3]-1, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
			gfx8.FillRectInset(rt, o[0], o[1], o[0]+o[2]-1, o[1]+o[3]-1, gfx8.InsetColour{Map: sec}, 0)
		}
	})
	if w.Save {
		b, o := pbTextBox, pbOK
		DrawTextBold(screen, "File name:", w.X+3, w.Y+b[1]+2, win2) // {COLOUR WINDOW_2}
		// drawTextInput, clipped to the box's inside: text at (1, 1) of the clip.
		tx, ty := w.X+b[0]+1, w.Y+b[1]+1+1
		text := string(w.Input)
		for len(text) > 0 {
			if tw, _ := MeasureTextBold(text); tw <= b[2]-2 {
				break
			}
			text = text[1:]
		}
		DrawTextBold(screen, text, tx, ty, black)
		tw, _ := MeasureTextBold(text)
		if w.frame&0x10 == 0 && tx+tw < w.X+b[0]+b[2]-1 { // showCaret: (cursorFrame & 0x10) == 0
			if c, ok := screen.SubImage(image.Rect(tx+tw, ty, tx+tw+1, ty+9)).(*ebiten.Image); ok {
				c.Fill(win2) // drawRect(x, y, 1, 9, getShade(secondary, 9))
			}
		}
		lw, _ := MeasureTextBold("OK")
		DrawTextBold(screen, "OK", w.X+o[0]+(o[2]-1)/2-lw/2, w.Y+o[1], win2) // Button label, centred
	}
	DrawWindowCaption(screen, w.X+pbCaption[0], w.Y+pbCaption[1], pbCaption[2], w.Title)
	drawLabelValue(screen, w.X+3, w.Y+pbParent[1]+6, "Folder: ", w.Dir, win2, black)
	ct := w.content
	for i, e := range w.entries {
		y := w.Y + ct[1] + i*pbRowHeight - w.Scroll
		if y < w.Y+ct[1] || y+10 > w.Y+ct[1]+ct[3] {
			continue
		}
		x := 1
		if e.dir {
			x += 14
		}
		c := black
		if i == w.Selected {
			c = win2
		}
		DrawTextBold(screen, stem(e.path), w.X+ct[0]+x, y, c)
	}
	if w.Selected >= 0 && w.Selected < len(w.entries) && !w.entries[w.Selected].dir {
		right := w.list()[0] + w.list()[2] - 1
		width := pbW - right - 8
		x := w.X + right + 3
		name := stem(w.entries[w.Selected].path)
		tw, _ := MeasureTextBold(name)
		DrawTextBold(screen, name, x+width/2-tw/2, w.Y+45, win2)
		if d := w.details(w.entries[w.Selected].path); d != nil {
			w.drawSavePreview(screen, x, w.Y+45+12, width, 201, d, sec, win2, black)
		}
	}
}

// details is loadFileDetails for a saved game: its SaveDetails, read once
// per file (nil when it has none). Preview, if set, stands in (diagnostics).
func (w *PromptBrowseWindow) details(path string) *scenario.SaveDetails {
	if w.Preview != nil {
		return w.Preview
	}
	if !strings.EqualFold(filepath.Ext(path), ".sv5") {
		return nil
	}
	if w.detailsPath != path {
		w.detailsPath = path
		w.detailsData, _ = scenario.ReadSaveDetails(path)
	}
	return w.detailsData
}

// drawSavePreview is PromptBrowseWindow.cpp drawSavePreview: an inset frame
// (borderInset, fillNone) with the 250×200 image at (x+1, y+1), then at
// y + 207 the company, owner, date and the challenge line ("Company:"
// labels in the window colour, values black).
func (w *PromptBrowseWindow) drawSavePreview(screen *ebiten.Image, x, y, width, height int, d *scenario.SaveDetails, sec []uint8, win2, black color.Color) {
	rx, ry := x-w.X, y-w.Y
	draw8(screen, w.X, w.Y, pbW, pbH, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, rx, ry, rx+width, ry+height, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillNone)
		for iy := 0; iy < 200 && len(d.Image) >= 250*200; iy++ { // the image as a plain (opaque) G1 bitmap
			for ix := 0; ix < 250; ix++ {
				if px, py := rx+1+ix, ry+1+iy; px >= 0 && px < rt.Width && py >= 0 && py < rt.Height {
					rt.Bits[py*rt.Width+px] = d.Image[iy*250+ix]
				}
			}
		}
	})
	ty := y + 207
	line := func(label, value string) {
		lw, _ := MeasureTextBold(label)
		DrawTextBold(screen, label, x, ty, win2)
		DrawTextBold(screen, value, x+lw, ty, black)
		ty += 10
	}
	line("Company: ", d.Company)
	line("Owner: ", d.Owner)
	yy, m, day := L.CalcDate(d.Date)
	line("Date: ", L.DateDMY(day, monthNames[m], yy))
	const completed, failed, beatenByOpponent = 1 << 6, 1 << 7, 1 << 8 // CompanyFlags (Company.h)
	switch f := d.ChallengeFlags; {
	case f&beatenByOpponent != 0:
	case f&completed != 0:
		line("Challenge: ", "Completed")
	case f&failed != 0:
		line("Challenge: ", "Failed")
	default:
		line("Challenge progress: ", strconv.Itoa(int(d.ChallengeProgress))+"%")
	}
}

var monthNames = [12]string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}

// onScrollbar reports a point on the list's scrollbar (shown when the rows
// don't fit).
func (w *PromptBrowseWindow) onScrollbar(px, py int) bool {
	l := w.list()
	return w.in(l, px, py) && len(w.entries)*pbRowHeight > l[3] &&
		px >= w.X+l[0]+l[2]-L.ScrollbarSize-L.ScrollbarMargin
}

// Drag follows the mouse while the thumb is held (verticalDragFollow); held
// false ends the drag.
func (w *PromptBrowseWindow) Drag(py int, held bool) {
	if !w.dragging {
		return
	}
	if !held {
		w.dragging = false
		return
	}
	if dy := py - w.dragY; dy != 0 {
		w.Scroll = L.VScrollDrag(w.list()[3], len(w.entries)*pbRowHeight, w.Scroll, dy)
		w.dragY = py
	}
}

// HandleRightClick is a right click on the list (onScrollMouseDown with the
// right button): on a file it asks "Delete file: “name”?" (Delete) and then
// removes the file and refreshes the list (processFileForDelete). Folders
// are left alone.
func (w *PromptBrowseWindow) HandleRightClick(px, py int) {
	l := w.list()
	if !w.in(l, px, py) || w.onScrollbar(px, py) {
		return
	}
	idx := (py - (w.Y + l[1] + 1) + w.Scroll) / pbRowHeight
	if idx < 0 || idx >= len(w.entries) || w.entries[idx].dir {
		return
	}
	path := w.entries[idx].path
	remove := func() {
		_ = os.Remove(path)
		w.chdir(w.Dir) // refreshDirectoryList
	}
	if w.Confirm == nil {
		return // no prompt to ask with: never delete unasked
	}
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	w.Confirm(w.Title, "Delete file: “"+name+"”?", "Delete", remove)
}

// ChangeDir lists another folder (as a click on a folder row).
func (w *PromptBrowseWindow) ChangeDir(dir string) { w.chdir(dir) }

// SelectFirstFile selects the list's first file (not a folder), as a hover
// would (diagnostics).
func (w *PromptBrowseWindow) SelectFirstFile() {
	for i, e := range w.entries {
		if !e.dir {
			w.Selected = i
			return
		}
	}
}
