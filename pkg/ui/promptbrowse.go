package ui

import (
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
// PARTIAL: save mode (filename text box + OK button), the saved-game / landscape
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
	case w.in(pbList, px, py):
		idx := (py - (w.Y + pbList[1] + 1) + w.Scroll) / pbRowHeight
		if idx < 0 || idx >= len(w.entries) {
			return
		}
		if idx != w.Selected {
			w.Selected = idx
			return
		}
		e := w.entries[idx]
		if e.dir {
			w.chdir(e.path)
		} else if w.OnPick != nil {
			w.OnPick(e.path)
		}
	}
}

// HandleWheel scrolls the list.
func (w *PromptBrowseWindow) HandleWheel(dy float64) {
	maxScroll := max(len(w.entries)*pbRowHeight-pbList[3], 0)
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
		l := pbList
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
	})
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
		right := pbList[0] + pbList[2] - 1
		width := pbW - right - 8
		x := w.X + right + 3
		name := stem(w.entries[w.Selected].path)
		tw, _ := MeasureTextBold(name)
		DrawTextBold(screen, name, x+width/2-tw/2, w.Y+45, win2)
	}
}
