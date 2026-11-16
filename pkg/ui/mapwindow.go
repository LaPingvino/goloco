package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/hajimehoshi/ebiten/v2"
)

// MapWindow — a port of Ui/Windows/MapWindow.cpp (PORT-PLANS A, step 1a):
// 350×272, resizable 229×176 … 888×828, the skin's windowTitlebarColour and
// windowMapColour; the caption, the close button, the panel and the scroll
// view ({3, 44} to the window's width − 108 and height − 14) showing the map
// buffer (sim.MapBuffer: the map turned 45°) at the scroll offset.
//
// clickable(MapWindow.cpp): close
//
// PARTIAL (PORT-PLANS A 1b–d): the scroll bars, the remembered size, the centring on the main view and a click to centre
// it, the tabs, the graph key, the status bar, the overlays.

const (
	mapWinW, mapWinH       = 350, 272
	mapWinMinW, mapWinMinH = 229, 176
	mapWinMaxW, mapWinMaxH = sim.MapRenderedWidth + 120, sim.MapRenderedWidth + 60
)

// MapWindow is the window.
type MapWindow struct {
	WinBase
	TitleColour, Colour uint8
	// Buf is the map's pixels (filled by the game); ScrollX, ScrollY the
	// scroll view's content offset.
	Buf              *sim.MapBuffer
	ScrollX, ScrollY int
	// Rotation is var_846: the view's rotation the buffer was filled for.
	Rotation uint8
}

// NewMapWindow is MapWindow::open's window (createWindow: the stack places
// it), its view on the middle of the map.
func NewMapWindow(buf *sim.MapBuffer, titleColour, colour uint8) *MapWindow {
	w := &MapWindow{WinBase: WinBase{WinType: WindowTypeMap, W: mapWinW, H: mapWinH, MinW: mapWinMinW, MinH: mapWinMinH,
		MaxW: mapWinMaxW, MaxH: mapWinMaxH, WinFlags: WindowFlagResizable}, TitleColour: titleColour, Colour: colour, Buf: buf}
	_, _, cw, ch := w.scrollView()
	w.ScrollX, w.ScrollY = (sim.MapRenderedWidth-cw)/2, (sim.MapRenderedWidth-ch)/2
	return w
}

// scrollView is the scroll view widget as prepareDraw lays it out (x, y,
// w, h, window-relative).
func (w *MapWindow) scrollView() (x, y, ww, hh int) {
	return 3, 44, w.W - 108 - 3 + 1, w.H - 14 - 44 + 1
}

// PrepareDraw keeps the offset within the map (onResize).
func (w *MapWindow) PrepareDraw() {
	_, _, cw, ch := w.scrollView()
	w.ScrollX = min(max(w.ScrollX, 0), max(sim.MapRenderedWidth-cw, 0))
	w.ScrollY = min(max(w.ScrollY, 0), max(sim.MapRenderedWidth-ch, 0))
}

// HandleClick: the close button.
func (w *MapWindow) HandleClick(px, py int) {
	x, y := px-w.X, py-w.Y
	if x >= w.W-15 && x < w.W-2 && y >= 2 && y < 15 {
		w.Close()
	}
}

// WidgetTooltip is the tooltip of the widget under a screen point.
func (w *MapWindow) WidgetTooltip(px, py int) string {
	x, y := px-w.X, py-w.Y
	if x >= w.W-15 && x < w.W-2 && y >= 2 && y < 15 {
		return i18n.Str(i18n.TooltipCloseWindow)
	}
	return ""
}

// Draw draws the window: frame, caption, panel, close button, and the
// scroll view's drawScroll — the buffer from the offset.
func (w *MapWindow) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	sx, sy, sw, sh := w.scrollView()
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, false)
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 41, w.W-1, w.H-1, gfx8.InsetColour{Map: sec}, 0) // Panel
		drawResizeHandle(rt, &w.WinBase, sec)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, w.W-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		gfx8.FillRectInset(rt, sx, sy, sx+sw-1, sy+sh-1, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset) // the scroll view's border
		cx, cy, cw, ch := sx+1, sy+1, sw-2, sh-2
		if w.Buf == nil || cw <= 0 || ch <= 0 {
			return
		}
		for yy := range ch {
			my := w.ScrollY + yy
			if my < 0 || my >= sim.MapRenderedWidth || cy+yy >= rt.Height {
				continue
			}
			for xx := range cw {
				mx := w.ScrollX + xx
				if mx < 0 || mx >= sim.MapRenderedWidth || cx+xx >= rt.Width {
					continue
				}
				rt.Bits[(cy+yy)*rt.Width+cx+xx] = w.Buf.Pixels[my*sim.MapRenderedWidth+mx]
			}
		}
	})
	DrawWindowCaption(screen, w.X+1, w.Y+1, w.W-2, i18n.Label(i18n.TitleMap))
}
