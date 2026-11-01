package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/hajimehoshi/ebiten/v2"
)

// Options — a port of Ui/Windows/Options.cpp: the common frame (caption,
// close button, panel, seven 31 px tabs) and the Audio tab (Sound group with
// the device row and "Play title screen music", Volume group with six
// sliders in two columns). Colours: the skin's windowTitlebarColour and
// windowOptionsColour.
// Tabs switch with upstream's size and caption per tab; Audio and Controls
// have their contents.
// PARTIAL: Display, Rendering, Regional, Company and Misc show only the common
// frame (their widgets aren't ported yet); Controls: only edge scrolling has a
// goloco counterpart, Customize Keys opens nothing; tab icon
// animation; the device dropdown (shows "Default"); goloco has only a music
// channel, so the master/effects/vehicles/interface/ambient sliders are drawn
// at 100% and inert; slider dragging (a click sets the music volume).

const (
	optW           = 366
	optSoundOff    = 49
	optSoundH      = 14 + (14 + 18) + 4 // kDeviceRowOffset + kCheckboxRowOffset + 4
	optVolOff      = optSoundOff + optSoundH + 4
	optVolH        = 16 + 3*26
	optH           = optVolOff + optVolH + 4
	imgTab         = 2387
	imgTabDisplay  = 2391
	imgTabControl  = 2392
	imgTabSound    = 2393
	imgTabMisc     = 2394
	imgTabGlobe    = 2395
	imgSliderTrack = 3543
	imgSliderThumb = 3544
	optSliderRange = 80
	optThumbHalf   = 10
)

var optTabImages = [7]int{imgTabDisplay, imgTab, imgTabSound, imgTabGlobe, imgTabControl, imgTab, imgTabMisc}

// Options tabs (Common::tab).
const (
	OptTabDisplay = iota
	OptTabRendering
	OptTabAudio
	OptTabRegional
	OptTabControls
	OptTabCompany
	OptTabMisc
)

// optTabSize is each tab namespace's kWindowSize; optTabCaption its caption
// (en-GB).
var (
	optTabSize    = [7][2]int{{400, 151}, {400, 218}, {optW, optH}, {366, 167}, {366, 114}, {420, 187}, {420, 266}}
	optTabCaption = [7]string{"Options - Display", "Options - Rendering", "Options - Audio", "Options - Regional",
		"Options - Controls", "Options - Company", "Options - Miscellaneous"}
	// Controls tab checkboxes (Controls::_widgets) and the Customize Keys button.
	optControlLabels = [3]string{"Scroll View when Pointer at Screen Edge", "Zoom to cursor position", "Invert right mouse dragging"}
)

// OptionsWindow is the window state.
type OptionsWindow struct {
	X, Y         int
	TitleColour  uint8
	Colour       uint8
	TitleMusic   bool
	MusicVolume  float64 // 0..1
	OnTitleMusic func(bool)
	OnMusicVol   func(float64)
	OnClose      func()

	// Tab is the current tab (Common::tab); the window takes that tab's size.
	Tab int
	// Controls holds edgeScrolling, zoomToCursor, invertRightMouseViewPan;
	// OnControl is called with the index and new value when one is toggled.
	Controls  [3]bool
	OnControl func(i int, v bool)
}

func (w *OptionsWindow) size() (int, int) { s := optTabSize[w.Tab]; return s[0], s[1] }

// NewOptions opens the window centred (createWindowCentred).
func NewOptions(sw, sh int, titleColour, colour uint8) *OptionsWindow {
	return &OptionsWindow{X: (sw - optW) / 2, Y: (sh - optH) / 2, TitleColour: titleColour, Colour: colour, Tab: OptTabAudio}
}

func (w *OptionsWindow) in(r [4]int, px, py int) bool {
	return px >= w.X+r[0] && px < w.X+r[0]+r[2] && py >= w.Y+r[1] && py < w.Y+r[1]+r[3]
}

// Contains reports whether the point is on the window.
func (w *OptionsWindow) Contains(px, py int) bool {
	ww, wh := w.size()
	return w.in([4]int{0, 0, ww, wh}, px, py)
}

func optSliderRow(i int) (lx, sx, y int) {
	y = optVolOff + 16 + 26*(i/2)
	if i%2 == 0 {
		return 12, 82, y
	}
	return 190, 260, y
}

// HandleClick: close, the title-music checkbox, the music slider.
func (w *OptionsWindow) HandleClick(px, py int) {
	ww, _ := w.size()
	if w.in([4]int{ww - 15, 2, 13, 13}, px, py) {
		if w.OnClose != nil {
			w.OnClose()
		}
		return
	}
	for i := range 7 {
		if w.in([4]int{3 + 31*i, 15, 31, 27}, px, py) {
			w.Tab = i // switchTab: the window keeps its position and takes the tab's size
			return
		}
	}
	if w.Tab == OptTabControls {
		for i := range w.Controls {
			if w.in([4]int{10, 49 + 15*i, 346, 12}, px, py) {
				w.Controls[i] = !w.Controls[i]
				if w.OnControl != nil {
					w.OnControl(i, w.Controls[i])
				}
				return
			}
		}
		return
	}
	if w.Tab != OptTabAudio {
		return
	}
	if w.in([4]int{10, optSoundOff + 14 + 18, 346, 12}, px, py) {
		w.TitleMusic = !w.TitleMusic
		if w.OnTitleMusic != nil {
			w.OnTitleMusic(w.TitleMusic)
		}
		return
	}
	_, sx, y := optSliderRow(1) // the music slider
	if w.in([4]int{sx, y - 3, 196, 18}, px, py) {
		v := float64(px-(w.X+sx)-optThumbHalf) / optSliderRange
		w.MusicVolume = min(max(v, 0), 1)
		if w.OnMusicVol != nil {
			w.OnMusicVol(w.MusicVolume)
		}
	}
}

// Draw draws the window.
func (w *OptionsWindow) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	pal := g1.GetPalette()
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	text := pal[gfx8.Shade(sec, 9)] // opaque window colour → shade 9
	black := pal[paletteBlack0]
	ww, wh := w.size()
	DrawWindowFrame(screen, w.X, w.Y, ww, wh, w.TitleColour, false)
	vols := [6]float64{1, w.MusicVolume, 1, 1, 1, 1}
	draw8(screen, w.X, w.Y, ww, wh, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 41, ww-1, wh-1, gfx8.InsetColour{Map: sec}, 0) // panel (to the window bottom)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, ww-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		for i, img := range optTabImages {
			x := 3 + 31*i
			bg := imgTab
			if i == w.Tab {
				bg++
			}
			for _, id := range []int{bg, img} {
				if id == imgTab && img == imgTab && id != bg {
					continue
				}
				if sp, err := gfx8.Decode(&g1.Elements[id]); err == nil {
					gfx8.DrawSpriteNoise(rt, sp, x, 15, gfx8.OpTransparent|gfx8.OpSrc, sec, nil)
				}
			}
		}
		if w.Tab == OptTabControls {
			for i := range w.Controls { // Checkbox boxes at (10, 49 + 15i)
				y := 49 + 15*i
				gfx8.FillRectInset(rt, 10, y, 10+10-1, y+10-1, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
			}
			// Button({26, 94}, {160, 12}): fillRectInset with the window colour, raised.
			gfx8.FillRectInset(rt, 26, 94, 26+160-1, 94+12-1, gfx8.InsetColour{Map: sec}, 0)
			return
		}
		if w.Tab != OptTabAudio {
			return
		}
		groupBox(rt, 4, optSoundOff, optW-8, optSoundH, sec, "Sound")
		groupBox(rt, 4, optVolOff, optW-8, optVolH, sec, "Volume")
		// Device row (dropdown widget: inset field + button)
		gfx8.FillRectInset(rt, 10, optSoundOff+14, 10+346-1, optSoundOff+14+12-1, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
		// Checkbox box
		cy := optSoundOff + 14 + 18
		gfx8.FillRectInset(rt, 10, cy, 10+10-1, cy+10-1, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
		track, terr := gfx8.Decode(&g1.Elements[imgSliderTrack])
		thumb, herr := gfx8.Decode(&g1.Elements[imgSliderThumb])
		for i := range 6 {
			_, sx, y := optSliderRow(i)
			if terr == nil {
				gfx8.DrawSpriteNoise(rt, track, sx, y-3, gfx8.OpTransparent|gfx8.OpSrc, sec, nil)
			}
			if herr == nil {
				tx := optThumbHalf + int(vols[i]*100)*optSliderRange/100
				gfx8.DrawSpriteNoise(rt, thumb, sx+tx, y-3, gfx8.OpTransparent|gfx8.OpSrc, sec, nil)
			}
		}
	})
	DrawWindowCaption(screen, w.X+1, w.Y+1, ww-2, optTabCaption[w.Tab])
	if w.Tab == OptTabControls {
		for i, l := range optControlLabels {
			y := 49 + 15*i
			if w.Controls[i] {
				DrawTextBold(screen, string(rune(0xAC)), w.X+10, w.Y+y, text)
			}
			DrawTextBold(screen, l, w.X+10+10+4, w.Y+y, text)
		}
		// Label::draw centred: x + (w - 1) / 2, y + max(0, (12 - 10) / 2 - 1).
		const label = "Customize Keys..."
		lw, _ := MeasureTextBold(label)
		DrawTextBold(screen, label, w.X+26+(160-1)/2-lw/2, w.Y+94, text)
		return
	}
	if w.Tab != OptTabAudio {
		return
	}
	// GroupBox labels: drawn at (x+5, y) in the opaque window colour.
	DrawTextBold(screen, "Sound", w.X+4+5, w.Y+optSoundOff, text)
	DrawTextBold(screen, "Volume", w.X+4+5, w.Y+optVolOff, text)
	DrawTextBold(screen, "Default", w.X+10+2, w.Y+optSoundOff+14+1, black)
	cy := optSoundOff + 14 + 18
	if w.TitleMusic {
		DrawTextBold(screen, string(rune(0xAC)), w.X+10, w.Y+cy, text)
	}
	DrawTextBold(screen, "Play title screen music", w.X+10+10+4, w.Y+cy, text)
	labels := [6]string{"Master:", "Music:", "Effects:", "Vehicles:", "Interface:", "Ambient:"}
	for i, l := range labels {
		lx, _, y := optSliderRow(i)
		DrawTextBold(screen, l, w.X+lx, w.Y+y+3, text)
	}
}

// groupBox is GroupBox::draw's border (the label is drawn by the caller, so
// the gap is computed from the label width here).
func groupBox(rt *gfx8.RenderTarget, x, y, w, h int, cmap []uint8, label string) {
	s4, s7 := gfx8.Shade(cmap, 4), gfx8.Shade(cmap, 7)
	tw, _ := MeasureTextBold(label)
	textEnd := x + 5 + tw + 1
	by, bh := y+4, h-4
	f := func(l, t, ww, hh int, c uint8) { gfx8.FillRect(rt, l, t, l+ww-1, t+hh-1, c) }
	f(x, by, 5, 1, s4)
	f(x+1, by+1, 4, 1, s7)
	f(textEnd, by, w-(textEnd-x)-1, 1, s4)
	f(textEnd, by+1, w-(textEnd-x)-2, 1, s7)
	f(x+w-1, by+1, 1, bh-1, s4)
	f(x+w, by, 1, bh+1, s7)
	f(x, by+bh-1, w-1, 1, s4)
	f(x, by+bh, w, 1, s7)
	f(x, by+1, 1, bh-2, s4)
	f(x+1, by+2, 1, bh-3, s7)
}
