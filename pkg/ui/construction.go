package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/hajimehoshi/ebiten/v2"
)

// ConstructionFrame is the frame of the construction window
// (Ui/Windows/Construction/Common.cpp makeCommonWidgets, createConstructionWindow):
// 138×276, lighter frame in the owner company's colour (WindowManager draws an
// owned window's primary colour as the company's), a colourText caption with
// the track's name, the close button, the panel in the skin's
// windowConstructionColour, and the four tabs (construction, station, signal,
// overhead) drawn by drawTrackTabs.
// PARTIAL: the station, signal and overhead tabs' previews (zoomed station
// image, animated signal); tabs hidden by setDisabledWidgets/repositionTabs;
// road, airport and dock tab sets; the tab contents (ConstructionTab etc.).
type ConstructionFrame struct {
	X, Y          int
	W             int // the tab's width (0: ConstructionW)
	Title         string
	CompanyColour uint8
	SkinColour    uint8 // InterfaceSkin windowConstructionColour
	TrackImage    int   // the track object's image (TrackObject::image)
	Tab           int   // 0 construction, 1 station, 2 signal, 3 overhead
	FrameNo       int   // window frameNo (counts updates; animates the tab)
	// TabFrames is how many preview frames the construction tab cycles through
	// (drawTrackTabs: 16, drawRoadTabs: 32).
	TabFrames int
	// Disabled tabs (setDisabledWidgets) are hidden and the others move left
	// (repositionTabs): index 0 construction, 1 station, 2 signal, 3 overhead.
	Disabled [4]bool
	// H is the window height for the current tab (0: ConstructionH).
	H int
	// StationTabIcon is the station tab's image (drawDockTabs / drawAirportTabs:
	// the skin's toolbar_menu_ship_port / _airport; 0: none, as drawTrackTabs).
	StationTabIcon int
	// Track holds drawTrackTabs' content for the station, signal and overhead
	// tabs (nil: road, dock and airport windows).
	Track *TrackTabIcons
}

// TrackTabIcons is what drawTrackTabs draws on the track window's tabs: the
// selected station type's preview (image, windows translucent in the company
// colour when recolourable, else ExtColour::unk2E), the selected signal
// (image, numFrames, animationSpeed) and the available extras' tab images
// (var_0E).
type TrackTabIcons struct {
	StationImage        int
	StationRecolourable bool
	SignalImage         int
	SignalNumFrames     int
	SignalAnimSpeed     int
	ExtraImages         []int
}

// signalFrames is TrainSignalObject.h's signalFrames (2-, 3- and 4-state).
var signalFrames = [3][]uint8{
	{1, 2, 3, 3, 3, 3, 3, 3, 2, 1, 0, 0, 0, 0, 0},
	{1, 2, 3, 3, 3, 3, 3, 3, 3, 4, 5, 6, 6, 6, 6, 6, 6, 6, 6, 5, 4, 3, 2, 1, 0, 0, 0, 0, 0, 0, 0},
	{1, 2, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 4, 5, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 7, 8, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

// Width is the window's width (the signal tab's is wider).
func (w *ConstructionFrame) Width() int {
	if w.W > 0 {
		return w.W
	}
	return ConstructionW
}

// Height is the window's height (each tab sets its own size).
func (w *ConstructionFrame) Height() int {
	if w.H > 0 {
		return w.H
	}
	return ConstructionH
}

// ConstructionW/H is Construction::kWindowSize.
const ConstructionW, ConstructionH = 138, 276

// Draw draws the frame and tabs.
func (w *ConstructionFrame) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	prim, _ := g1.GetPaletteMap(int(w.CompanyColour))
	sec, err := g1.GetPaletteMap(int(w.SkinColour))
	if err != nil {
		return
	}
	h, wd := w.Height(), w.Width()
	DrawWindowFrame(screen, w.X, w.Y, wd, h, w.CompanyColour, true)
	draw8(screen, w.X, w.Y, wd, h, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 41, wd-1, h-1, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, wd-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		// Construction tab: the track image, animated through kTrackPreviewImages
		// (+0..15, frameNo / 4) while it is the current tab; recoloured to the
		// company colour.
		icon := w.TrackImage
		if w.Tab == 0 {
			icon += (w.FrameNo / 4) % max(w.TabFrames, 1)
		}
		for i, x := range w.TabX() {
			if x < 0 {
				continue
			}
			if i == 0 {
				drawTab(rt, x, 15, w.Tab == 0, sec, icon, prim)
			} else if w.Track != nil {
				w.drawTrackTab(rt, i, x, 15, sec)
			} else if i == 1 && w.StationTabIcon > 0 {
				drawTab(rt, x, 15, w.Tab == i, sec, w.StationTabIcon, nil)
			} else {
				drawTab(rt, x, 15, w.Tab == i, sec, tabNoIcon, nil)
			}
		}
	})
}

// DrawCaption draws the title. It goes after the tab's content: the tabs'
// draw8 covers the whole window and writes back the 8-bit screen copy, which
// holds the caption's fill but not its text (drawn on the screen directly) —
// drawn before, the title was wiped whenever the copy was usable.
func (w *ConstructionFrame) DrawCaption(screen *ebiten.Image) {
	DrawWindowCaptionColour(screen, w.X+1, w.Y+1, w.Width()-2, w.Title, int(w.CompanyColour))
}

// TabX is repositionTabs: each tab's x, -1 for a hidden (disabled) one; the
// enabled tabs sit side by side from x 3, 31 apart.
func (w *ConstructionFrame) TabX() [4]int {
	var xs [4]int
	x := 3
	for i := range xs {
		if w.Disabled[i] {
			xs[i] = -1
			continue
		}
		xs[i] = x
		x += 31
	}
	return xs
}

// TabAt returns the tab under (rx, ry) relative to the window, or -1.
func (w *ConstructionFrame) TabAt(rx, ry int) int {
	if ry < 15 || ry >= 15+27 {
		return -1
	}
	for i, x := range w.TabX() {
		if x >= 0 && rx >= x && rx < x+31 {
			return i
		}
	}
	return -1
}

// drawTrackTab is drawTrackTabs' station, signal and overhead tabs: the tab,
// its content, then the inactive overlay (Widget::drawTab with kContentUnk).
// Station: clipped to (x + 1, y + 1, 29 × 25, one more high when the station
// tab is current), drawn at zoom 1, the preview at (−4, −9) recoloured to the
// company colour and its windows translucent. Signal: the same clip (its
// height compared with the station tab, as upstream does), the signal at
// (15, 31), animated through signalFrames while current. Overhead: each
// available extra's tab image at (x + 2, y + 2), + (frameNo / 2) % 8 while
// current.
func (w *ConstructionFrame) drawTrackTab(rt *gfx8.RenderTarget, i, x, y int, sec []uint8) {
	g1 := globalG1Font.g1
	t := w.Track
	active := w.Tab == i
	drawTab(rt, x, y, active, sec, -1, nil)
	clip := func(draw func(c *gfx8.RenderTarget)) {
		h := 25
		if w.Tab == 1 {
			h++
		}
		c := gfx8.NewRenderTarget(29, h)
		draw(c)
		for yy := 0; yy < h; yy++ {
			for xx := 0; xx < 29; xx++ {
				if v := c.Bits[yy*29+xx]; v != 0 {
					px, py := x+1+xx, y+1+yy
					if px >= 0 && py >= 0 && px < rt.Width && py < rt.Height {
						rt.Bits[py*rt.Width+px] = v
					}
				}
			}
		}
	}
	sprite := func(id int) *gfx8.Sprite {
		if id <= 0 || id >= len(g1.Elements) {
			return nil
		}
		sp, err := gfx8.Decode(&g1.Elements[id])
		if err != nil {
			return nil
		}
		return sp
	}
	switch i {
	case 1:
		clip(func(c *gfx8.RenderTarget) {
			prim, _ := g1.GetPaletteMap(int(w.CompanyColour))
			if sp := sprite(t.StationImage); sp != nil {
				gfx8.DrawSpriteZoom(c, sp, -4, -9, 1, gfx8.OpFor(sp, prim != nil, false), prim, nil)
			}
			win := 0x2E // ExtColour::unk2E
			if t.StationRecolourable {
				win = int(paintTranslucent(w.CompanyColour, 1))
			}
			if m, err := g1.GetPaletteMap(win); err == nil {
				if sp := sprite(t.StationImage + 1); sp != nil {
					gfx8.DrawSpriteZoom(c, sp, -4, -9, 1, gfx8.OpTransparent|gfx8.OpDst, m, nil)
				}
			}
		})
	case 2:
		clip(func(c *gfx8.RenderTarget) {
			img := t.SignalImage
			if active && t.SignalNumFrames > 0 {
				if k := (t.SignalNumFrames+2)/3 - 2; k >= 0 && k < len(signalFrames) {
					frames := signalFrames[k]
					n := (len(frames) - 1) & (w.FrameNo >> t.SignalAnimSpeed)
					img += int(frames[n]) << 3
				}
			}
			if sp := sprite(img); sp != nil {
				gfx8.DrawSpriteNoise(c, sp, 15, 31, gfx8.OpFor(sp, false, false), nil, nil)
			}
		})
	case 3:
		for _, img := range t.ExtraImages {
			if active {
				img += (w.FrameNo / 2) % 8
			}
			if sp := sprite(img); sp != nil {
				gfx8.DrawSpriteNoise(rt, sp, x+2, y+2, gfx8.OpFor(sp, false, false), nil, nil)
			}
		}
	}
	if !active { // Widget::drawTab(kContentUnk): the inactive overlay and line
		if sp := sprite(imgTab); sp != nil {
			if m, err := g1.GetPaletteMap(extColourUnk33); err == nil {
				gfx8.DrawSpriteNoise(rt, sp, x, y, gfx8.OpFor(sp, false, true), m, nil)
			}
		}
		gfx8.FillRect(rt, x, y+26, x+31-1, y+26, gfx8.Shade(sec, 7))
	}
}
