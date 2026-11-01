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
// PARTIAL: only the Audio tab exists (the other tabs show it too); tab icon
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
}

// NewOptions opens the window centred (createWindowCentred).
func NewOptions(sw, sh int, titleColour, colour uint8) *OptionsWindow {
	return &OptionsWindow{X: (sw - optW) / 2, Y: (sh - optH) / 2, TitleColour: titleColour, Colour: colour}
}

func (w *OptionsWindow) in(r [4]int, px, py int) bool {
	return px >= w.X+r[0] && px < w.X+r[0]+r[2] && py >= w.Y+r[1] && py < w.Y+r[1]+r[3]
}

// Contains reports whether the point is on the window.
func (w *OptionsWindow) Contains(px, py int) bool { return w.in([4]int{0, 0, optW, optH}, px, py) }

func optSliderRow(i int) (lx, sx, y int) {
	y = optVolOff + 16 + 26*(i/2)
	if i%2 == 0 {
		return 12, 82, y
	}
	return 190, 260, y
}

// HandleClick: close, the title-music checkbox, the music slider.
func (w *OptionsWindow) HandleClick(px, py int) {
	if w.in([4]int{optW - 15, 2, 13, 13}, px, py) {
		if w.OnClose != nil {
			w.OnClose()
		}
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
	DrawWindowFrame(screen, w.X, w.Y, optW, optH, w.TitleColour, false)
	vols := [6]float64{1, w.MusicVolume, 1, 1, 1, 1}
	draw8(screen, w.X, w.Y, optW, optH, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 41, optW-1, optH-1, gfx8.InsetColour{Map: sec}, 0) // panel (to the window bottom)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, optW-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		for i, img := range optTabImages {
			x := 3 + 31*i
			bg := imgTab
			if i == 2 { // the Audio tab is the current one
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
	DrawWindowCaption(screen, w.X+1, w.Y+1, optW-2, "Options - Audio")
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
