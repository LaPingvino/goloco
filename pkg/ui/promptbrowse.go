package ui

import (
	"image"
	"os"
	"path/filepath"
	"sort"
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
// shortening, scrollbar mouse interaction, window resizing.

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
	if w.OnSave != nil {
		w.OnSave(filepath.Join(w.Dir, name+w.SaveExt))
	}
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
			if w.OnSave != nil {
				w.OnSave(e.path) // PARTIAL: upstream asks before overwriting
			}
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
	}
}
