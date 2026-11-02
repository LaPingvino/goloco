package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/objects"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
)

// Options — a port of Ui/Windows/Options.cpp: the common frame (caption,
// close button, panel, seven 31 px tabs, each tab with upstream's size and
// caption) and all seven tabs' widgets. Colours: the skin's
// windowTitlebarColour and windowOptionsColour.
// Working: the tabs, Audio's title-music checkbox and music volume, Controls'
// edge scrolling. Display only (at Config's defaults, goloco has no such
// settings yet): Display, Rendering, Regional (dropdowns don't open), Company,
// Misc, and Controls' other two checkboxes.
// PARTIAL: tab icon animation; Customize Keys opens nothing; the audio device
// dropdown (shows "Default"); goloco has only a music channel, so the
// master/effects/vehicles/interface/ambient sliders are drawn at 100% and
// inert; slider dragging (a click sets the music volume).

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

	// Regional tab: the current game's currency name and the preferred one;
	// TitleMode disables the current-currency dropdown (prepareDraw).
	CurrencyName, PreferredCurrencyName string
	// Currencies are the installed currency objects (sorted by name);
	// OnPreferredCurrency is told the DAT name chosen as preferred.
	Currencies          []objects.CurrencyChoice
	OnPreferredCurrency func(dat string)
	// Languages are the available packs (native names, sorted) and Language
	// the current one's index; OnLanguage loads the chosen one.
	// OnCurrency switches the current currency object (currencyDropdown).
	OnCurrency func(dat string)
	Languages  []string
	Language   int
	OnLanguage func(i int)
	// OpenDropdown shows a dropdown for the window (the game's single
	// dropdown); nil: the Regional dropdowns don't open.
	OpenDropdown func(*DropdownMenu)
	// ScreenMode is Config display.mode (0 windowed, 1 fullscreen, 2
	// fullscreen window); OnScreenMode is told a new one (setDisplayMode).
	ScreenMode   int
	OnScreenMode func(int)
	// FrameLimit is 0 internal speed, 1 vertical sync, 2 unrestricted
	// (uncapFPS / display.vsync); OnFrameLimit is told a new one.
	FrameLimit   int
	OnFrameLimit func(int)
	// The Rendering tab's Config (vehiclesMinScale, stationNamesMinScale:
	// 0 full … 3 eighth, default 2; constructionMarker: 0 white, 1
	// translucent). PARTIAL: goloco's renderer doesn't read them yet.
	VehiclesMinScale, StationNamesMinScale, ConstructionMarker int
	TitleMode                                                  bool
}

// Regional tab rows (Regional::_widgets): label at x 10, dropdown at x 183,
// both 173×12.
// optScreenModes are options_mode_windowed / _fullscreen / _fullscreen_window.
var optScreenModes = []string{"Windowed", "Fullscreen", "Fullscreen (borderless window)"}

// optFrameLimits are frameRateLimitInternal / Vsync / Unrestricted.
var optFrameLimits = []string{"Internal speed (default)", "Vertical sync", "Unrestricted"}

// optScales are full_scale … eighth_scale; optMarkers white / translucent.
var (
	optScales  = []string{"Full-scale", "Half-scale", "Quarter-scale", "Eighth-scale"}
	optMarkers = []string{"White", "Translucent"}
	// optFrameStyles are windowFrameStyleGradient / Solid / Translucent.
	optFrameStyles = []string{"Gradient (default)", "Solid colour", "Translucent colour"}
)

var optRegionalRows = [5]struct {
	y     int
	label string
}{{49, "Language:"}, {69, "Distance and Speed:"}, {84, "Heights:"}, {104, "Current currency:"}, {119, "Preferred currency:"}}

// optDropdown draws dropdownWidgets: the ComboBox (inset, fillDarker) and the
// 11×10 Button at (x + w - 12, y + 1).
func optDropdown(rt *gfx8.RenderTarget, x, y, w int, sec []uint8) {
	gfx8.FillRectInset(rt, x, y, x+w-1, y+12-1, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
	bx := x + w - 12
	gfx8.FillRectInset(rt, bx, y+1, bx+11-1, y+1+10-1, gfx8.InsetColour{Map: sec}, 0)
}

func (w *OptionsWindow) size() (int, int) { s := optTabSize[w.Tab]; return s[0], s[1] }

// NewOptions opens the window centred (createWindowCentred).
func NewOptions(sw, sh int, titleColour, colour uint8) *OptionsWindow {
	return &OptionsWindow{X: (sw - optW) / 2, Y: (sh - optH) / 2, TitleColour: titleColour, Colour: colour, Tab: OptTabAudio,
		VehiclesMinScale: 2, StationNamesMinScale: 2}
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
	if w.Tab == OptTabRendering {
		// vehicleZoomMouseDown, stationNamesScaleMouseDown,
		// constructionMarkerMouseDown: Dropdown::show at the combobox, width() − 4.
		for _, c := range []struct {
			y     int
			items []string
			v     *int
		}{{63, optScales, &w.VehiclesMinScale}, {79, optScales, &w.StationNamesMinScale}, {95, optMarkers, &w.ConstructionMarker},
			{194, optFrameStyles, &FrameStyle}} { // windowFrameStyleMouseDown too
			if w.in([4]int{235, c.y, 154, 12}, px, py) && w.OpenDropdown != nil {
				v := c.v
				items := make([]DropdownItem, len(c.items))
				for i, t := range c.items {
					items[i] = DropdownItem{Text: t, Selected: i == *v, MoveX10: true}
				}
				d := NewDropdownShow(w.X+235, w.Y+c.y, 154-1-4, 11, w.Colour, items, 1<<20, 1<<20)
				d.OnSelect = func(i int) {
					if i >= 0 {
						*v = i // (PARTIAL: Config::write)
					}
				}
				w.OpenDropdown(d)
			}
		}
		return
	}
	if w.Tab == OptTabDisplay {
		// screenModeMouseDown: the three modes (dropdown_stringid), the current
		// one selected; screenModeDropdown sets it when it changes.
		if w.in([4]int{235, 63, 154, 12}, px, py) && w.OpenDropdown != nil {
			items := make([]DropdownItem, len(optScreenModes))
			for i, m := range optScreenModes {
				items[i] = DropdownItem{Text: m, Selected: i == w.ScreenMode, MoveX10: true}
			}
			d := NewDropdownShow(w.X+235, w.Y+63, 154-1-4, 11, w.Colour, items, 1<<20, 1<<20)
			d.OnSelect = func(i int) {
				if i >= 0 && i != w.ScreenMode {
					w.ScreenMode = i
					if w.OnScreenMode != nil {
						w.OnScreenMode(i)
					}
				}
			}
			w.OpenDropdown(d)
		}
		// frameLimitMouseDown: Dropdown::show at the combobox, width() (no −4).
		if w.in([4]int{235, 111, 154, 12}, px, py) && w.OpenDropdown != nil {
			items := make([]DropdownItem, len(optFrameLimits))
			for i, m := range optFrameLimits {
				items[i] = DropdownItem{Text: m, Selected: i == w.FrameLimit, MoveX10: true}
			}
			d := NewDropdownShow(w.X+235, w.Y+111, 154-1, 11, w.Colour, items, 1<<20, 1<<20)
			d.OnSelect = func(i int) {
				if i >= 0 {
					w.FrameLimit = i
					if w.OnFrameLimit != nil {
						w.OnFrameLimit(i)
					}
				}
			}
			w.OpenDropdown(d)
		}
		return
	}
	if w.Tab == OptTabRegional {
		// distanceSpeedMouseDown: Dropdown::show at the combobox (183, 69),
		// width() − 4, Imperial / Metric, the current one ticked.
		if r := optRegionalRows[1]; w.in([4]int{183, r.y, 173, 12}, px, py) && w.OpenDropdown != nil {
			d := NewDropdownShow(w.X+183, w.Y+r.y, 173-1-4, 11, w.Colour,
				[]DropdownItem{{Text: "Imperial", Selected: !L.Metric, MoveX10: true}, {Text: "Metric", Selected: L.Metric, MoveX10: true}}, 1<<20, 1<<20)
			d.OnSelect = func(i int) { L.Metric = i == 1 } // distanceSpeedDropdown (PARTIAL: Config::write)
			w.OpenDropdown(d)
		}
		// languageMouseDown: the available languages (dropdown_stringptr), the
		// current one selected.
		if r := optRegionalRows[0]; w.in([4]int{183, r.y, 173, 12}, px, py) && w.OpenDropdown != nil && len(w.Languages) > 0 {
			items := make([]DropdownItem, len(w.Languages))
			for i, n := range w.Languages {
				items[i] = DropdownItem{Text: n, Selected: i == w.Language, MoveX10: true}
			}
			d := NewDropdownShow(w.X+183, w.Y+r.y, 173-1-4, 11, w.Colour, items, 1<<20, 1<<20)
			d.OnSelect = func(i int) {
				if i >= 0 && i < len(w.Languages) {
					w.Language = i
					if w.OnLanguage != nil {
						w.OnLanguage(i)
					}
				}
			}
			w.OpenDropdown(d)
		}
		// currencyMouseDown: every installed currency, the loaded one selected;
		// the widget is disabled in title mode (prepareDraw).
		if r := optRegionalRows[3]; !w.TitleMode && w.in([4]int{183, r.y, 173, 12}, px, py) && w.OpenDropdown != nil && len(w.Currencies) > 0 {
			items := make([]DropdownItem, len(w.Currencies))
			for i, c := range w.Currencies {
				items[i] = DropdownItem{Text: c.Name, Selected: c.Name == w.CurrencyName, MoveX10: true}
			}
			d := NewDropdownShow(w.X+183, w.Y+r.y, 173-1-4, 11, w.Colour, items, 1<<20, 1<<20)
			d.OnSelect = func(i int) {
				if i >= 0 && i < len(w.Currencies) {
					w.CurrencyName = w.Currencies[i].Name
					if w.OnCurrency != nil {
						w.OnCurrency(w.Currencies[i].DAT)
					}
				}
			}
			w.OpenDropdown(d)
		}
		// preferredCurrencyMouseDown: every installed currency (dropdown_stringptr
		// "{MOVE_X 10}{STRING}"), the preferred one selected ("»").
		if r := optRegionalRows[4]; w.in([4]int{183, r.y, 173, 12}, px, py) && w.OpenDropdown != nil && len(w.Currencies) > 0 {
			items := make([]DropdownItem, len(w.Currencies))
			for i, c := range w.Currencies {
				items[i] = DropdownItem{Text: c.Name, Selected: c.Name == w.PreferredCurrencyName, MoveX10: true}
			}
			d := NewDropdownShow(w.X+183, w.Y+r.y, 173-1-4, 11, w.Colour, items, 1<<20, 1<<20)
			d.OnSelect = func(i int) { // preferredCurrencyDropdown (PARTIAL: Config::write, loadPreferredCurrencyAlways)
				if i >= 0 && i < len(w.Currencies) {
					w.PreferredCurrencyName = w.Currencies[i].Name
					if w.OnPreferredCurrency != nil {
						w.OnPreferredCurrency(w.Currencies[i].DAT)
					}
				}
			}
			w.OpenDropdown(d)
		}
		// heightsLabelsMouseDown: Units / Real Values, the current one selected.
		if r := optRegionalRows[2]; w.in([4]int{183, r.y, 173, 12}, px, py) && w.OpenDropdown != nil {
			d := NewDropdownShow(w.X+183, w.Y+r.y, 173-1-4, 11, w.Colour,
				[]DropdownItem{{Text: "Units", Selected: L.HeightUnits, MoveX10: true}, {Text: "Real Values", Selected: !L.HeightUnits, MoveX10: true}}, 1<<20, 1<<20)
			d.OnSelect = func(i int) { L.HeightUnits = i == 0 } // heightsLabelsDropdown (PARTIAL: Config::write)
			w.OpenDropdown(d)
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
			icon := img
			if img == imgTab { // Rendering/Company: skin icons drawn by drawTabs (PARTIAL: not yet)
				icon = tabNoIcon
			}
			drawTab(rt, 3+31*i, 15, i == w.Tab, sec, icon, nil)
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
		if w.Tab == OptTabMisc {
			groupBox(rt, 4, 49, 412, 62, sec, "Gameplay tweaks")
			groupBox(rt, 4, 115, 412, 77, sec, "Vehicle/track behaviour")
			groupBox(rt, 4, 196, 412, 65, sec, "Autosave preferences")
			for _, y := range []int{64, 79, 94, 130, 145, 160, 175} {
				gfx8.FillRectInset(rt, 10, y, 10+10-1, y+10-1, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
			}
			optDropdown(rt, 250, 211, 156, sec)
			gfx8.FillRectInset(rt, 250, 226, 250+156-1, 226+12-1, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
			gfx8.FillRectInset(rt, 250+156-26, 227, 250+156-26+13-1, 227+10-1, gfx8.InsetColour{Map: sec}, 0)
			gfx8.FillRectInset(rt, 250+156-13, 227, 250+156-13+12-1, 227+10-1, gfx8.InsetColour{Map: sec}, 0)
			return
		}
		if w.Tab == OptTabRendering {
			groupBox(rt, 4, 49, 392, 128, sec, "Map rendering")
			groupBox(rt, 4, 180, 392, 32, sec, "User interface")
			for _, y := range []int{63, 79, 95, 194} {
				optDropdown(rt, 235, y, 154, sec)
			}
			for _, y := range []int{110, 126, 142, 158} {
				gfx8.FillRectInset(rt, 10, y, 10+10-1, y+10-1, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
			}
			return
		}
		if w.Tab == OptTabCompany {
			groupBox(rt, 4, 50, 412, 80, sec, "Preferred owner")
			groupBox(rt, 4, 135, 412, 47, sec, "Preferred company")
			for _, y := range []int{64, 97, 149} { // Checkboxes
				gfx8.FillRectInset(rt, 10, y, 10+10-1, y+10-1, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
			}
			for _, y := range []int{79, 112, 164} { // "Change..." Buttons (265, y, 75×12)
				gfx8.FillRectInset(rt, 265, y, 265+75-1, y+12-1, gfx8.InsetColour{Map: sec}, 0)
			}
			// The face preview (ImageButton 345,59 66×66) has no image while
			// usePreferredOwnerFace is off, and then draws nothing.
			return
		}
		if w.Tab == OptTabDisplay {
			groupBox(rt, 4, 49, 392, 97, sec, "Hardware")
			for _, y := range []int{63, 79, 111} {
				optDropdown(rt, 235, y, 154, sec)
			}
			// stepperWidgets({235, 95}, {154, 12}): TextBox + "-" (x+w-26, 13 wide) and
			// "+" (x+w-13, 12 wide) Buttons, both y+1 and h-2 high.
			gfx8.FillRectInset(rt, 235, 95, 235+154-1, 95+12-1, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
			gfx8.FillRectInset(rt, 235+154-26, 96, 235+154-26+13-1, 96+10-1, gfx8.InsetColour{Map: sec}, 0)
			gfx8.FillRectInset(rt, 235+154-13, 96, 235+154-13+12-1, 96+10-1, gfx8.InsetColour{Map: sec}, 0)
			gfx8.FillRectInset(rt, 10, 127, 10+10-1, 127+10-1, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
			return
		}
		if w.Tab == OptTabRegional {
			for _, r := range optRegionalRows {
				optDropdown(rt, 183, r.y, 173, sec)
			}
			for i := range 2 { // Checkboxes at (10, 134) and (10, 148)
				y := 134 + 14*i
				gfx8.FillRectInset(rt, 10, y, 10+10-1, y+10-1, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
			}
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
	inset := func(t string, x, y int) { // colour.inset(): shade 4 at +1,+1, then shade 6
		DrawTextBold(screen, t, x+1, y+1, pal[gfx8.Shade(sec, 4)])
		DrawTextBold(screen, t, x, y, pal[gfx8.Shade(sec, 6)])
	}
	if w.Tab == OptTabMisc {
		// Config defaults: every tweak off, autosave every month, 12 autosaves.
		// "Export plug-in objects with saved games" is hidden: it shows only when
		// the object index holds custom objects.
		DrawTextBold(screen, "Gameplay tweaks", w.X+4+5, w.Y+49, text)
		DrawTextBold(screen, "Vehicle/track behaviour", w.X+4+5, w.Y+115, text)
		DrawTextBold(screen, "Autosave preferences", w.X+4+5, w.Y+196, text)
		for _, c := range []struct {
			y     int
			label string
		}{{64, "Enable cheats/debugging toolbar menu"}, {79, "Disable AI companies"}, {94, "Disable town expansion"},
			{130, "Disable vehicle breakdowns"}, {145, "Disable vehicle load penalty"},
			{160, "Disable station size limitations"}, {175, "Trains reverse at signals"}} {
			DrawTextBold(screen, c.label, w.X+10+10+4, w.Y+c.y, text)
		}
		DrawTextBold(screen, "Autosave frequency:", w.X+10, w.Y+211, text)
		DrawTextBold(screen, "Autosave amount:", w.X+10, w.Y+226, text)
		DrawTextBold(screen, "Every month", w.X+250+1, w.Y+211+1, black) // drawDropdownContent
		DrawTextBold(screen, "12", w.X+250+1, w.Y+226+1, black)
		arrow := string(rune(170))
		aw, _ := MeasureTextBold(arrow)
		DrawTextBold(screen, arrow, w.X+250+156-12+(11-1)/2-aw/2, w.Y+211+1, black)
		mw, _ := MeasureTextBold("-")
		pw, _ := MeasureTextBold("+")
		DrawTextBold(screen, "-", w.X+250+156-26+(13-1)/2-mw/2, w.Y+227, black)
		DrawTextBold(screen, "+", w.X+250+156-13+(12-1)/2-pw/2, w.Y+227, black)
		return
	}
	if w.Tab == OptTabRendering {
		// Config defaults: vehicles/station names min. scale 2 (Quarter-scale),
		// construction marker 0 (White), landscape smoothing and cash popups on,
		// gridlines and AI ghosts off, window frame style background (Gradient).
		DrawTextBold(screen, "Map rendering", w.X+4+5, w.Y+49, text)
		DrawTextBold(screen, "User interface", w.X+4+5, w.Y+180, text)
		arrow := string(rune(170))
		aw, _ := MeasureTextBold(arrow)
		for _, r := range []struct {
			ly, y        int
			label, value string
		}{{63, 63, "Vehicles min. display scale:", optScales[w.VehiclesMinScale]}, {79, 79, "Station names min. display scale:", optScales[w.StationNamesMinScale]},
			{95, 95, "Construction Marker:", optMarkers[w.ConstructionMarker]}, {195, 194, "Window frame style:", optFrameStyles[FrameStyle]}} {
			DrawTextBold(screen, r.label, w.X+10, w.Y+r.ly, text)
			DrawTextBold(screen, r.value, w.X+235+1, w.Y+r.y+1, black)
			DrawTextBold(screen, arrow, w.X+235+154-12+(11-1)/2-aw/2, w.Y+r.y+1, black)
		}
		tick := string(rune(0xAC))
		for _, c := range []struct {
			y     int
			label string
			on    bool
		}{{110, "Landscape Smoothing", true}, {126, "Gridlines on Landscape", false},
			{142, "Show cash popups", true}, {158, "Show planned AI track/road as ghosts", false}} {
			if c.on {
				DrawTextBold(screen, tick, w.X+10, w.Y+c.y, text)
			}
			DrawTextBold(screen, c.label, w.X+10+10+4, w.Y+c.y, text)
		}
		return
	}
	if w.Tab == OptTabCompany {
		// Config defaults: every "use preferred …" off, so the Change... buttons are
		// disabled (inset labels), the face label is empty and the names are "".
		DrawTextBold(screen, "Preferred owner", w.X+4+5, w.Y+50, text)
		DrawTextBold(screen, "Preferred company", w.X+4+5, w.Y+135, text)
		for _, c := range []struct {
			y     int
			label string
		}{{64, "Use preferred company face when starting a new game"}, {97, "Use preferred owner name when starting a new game"},
			{149, "Use preferred company name when starting a new game"}} {
			DrawTextBold(screen, c.label, w.X+10+10+4, w.Y+c.y, text)
		}
		cw, _ := MeasureTextBold("Change...")
		for _, y := range []int{79, 112, 164} {
			inset("Change...", w.X+265+(75-1)/2-cw/2, w.Y+y)
		}
		DrawTextBold(screen, "Preferred owner name: ", w.X+24, w.Y+112, text)
		DrawTextBold(screen, "Preferred company name: ", w.X+24, w.Y+164, text)
		return
	}
	if w.Tab == OptTabDisplay {
		// prepareDraw's values, at Config's defaults: window mode, fullscreen
		// resolution 1920×1080 (disabled outside fullscreen), scale 1.00 (its "-"
		// disabled at ScaleFactor::min), the internal frame rate, no fps counter.
		DrawTextBold(screen, "Hardware", w.X+4+5, w.Y+49, text)
		arrow := string(rune(170))
		aw, _ := MeasureTextBold(arrow)
		rows := []struct {
			y            int
			label, value string
			disabled     bool
		}{{63, "Screen mode:", optScreenModes[w.ScreenMode], false}, {79, "Display Resolution:", "1920 x 1080", w.ScreenMode != 1},
			{95, "Window scale factor:", "1.00", false}, {111, "Frame rate limit:", optFrameLimits[w.FrameLimit], false}}
		for _, r := range rows {
			DrawTextBold(screen, r.label, w.X+10, w.Y+r.y, text)
			if r.disabled {
				inset(r.value, w.X+235+1, w.Y+r.y+1)
				inset(arrow, w.X+235+154-12+(11-1)/2-aw/2, w.Y+r.y+1)
				continue
			}
			DrawTextBold(screen, r.value, w.X+235+1, w.Y+r.y+1, black)
			if r.y != 95 {
				DrawTextBold(screen, arrow, w.X+235+154-12+(11-1)/2-aw/2, w.Y+r.y+1, black)
			}
		}
		mw, _ := MeasureTextBold("-")
		pw, _ := MeasureTextBold("+")
		inset("-", w.X+235+154-26+(13-1)/2-mw/2, w.Y+96) // disabled at the minimum scale
		DrawTextBold(screen, "+", w.X+235+154-13+(12-1)/2-pw/2, w.Y+96, black)
		DrawTextBold(screen, "Display fps counter", w.X+10+10+4, w.Y+127, text)
		return
	}
	if w.Tab == OptTabRegional {
		// prepareDraw's values: Config defaults (language en-GB "English (UK)",
		// measurementFormat imperial, showHeightAsUnits false) and the currencies.
		measure := "Imperial"
		if L.Metric {
			measure = "Metric"
		}
		heights := "Real Values"
		if L.HeightUnits {
			heights = "Units"
		}
		lang := "English (UK)"
		if w.Language >= 0 && w.Language < len(w.Languages) {
			lang = w.Languages[w.Language]
		}
		values := [5]string{lang, measure, heights, w.CurrencyName, w.PreferredCurrencyName}
		arrow := string(rune(170)) // LocoChar::down, StringIds::dropdown "{COLOUR BLACK}▼"
		aw, _ := MeasureTextBold(arrow)
		for i, r := range optRegionalRows {
			DrawTextBold(screen, r.label, w.X+10, w.Y+r.y, text) // Label, left, {COLOUR WINDOW_2}
			v := values[i]
			for len(v) > 0 {
				if tw, _ := MeasureTextBold(v); tw <= 173-2 {
					break
				}
				v = v[:len(v)-1]
			}
			if i == 3 && w.TitleMode { // disabled: colour.inset()
				c6, c4 := pal[gfx8.Shade(sec, 6)], pal[gfx8.Shade(sec, 4)]
				DrawTextBold(screen, v, w.X+183+1+1, w.Y+r.y+1+1, c4)
				DrawTextBold(screen, v, w.X+183+1, w.Y+r.y+1, c6)
			} else {
				DrawTextBold(screen, v, w.X+183+1, w.Y+r.y+1, black) // ComboBox text, FD = colour 0
			}
			bx := 183 + 173 - 12
			DrawTextBold(screen, arrow, w.X+bx+(11-1)/2-aw/2, w.Y+r.y+1, black)
		}
		DrawTextBold(screen, "Use preferred currency when starting new game", w.X+10+10+4, w.Y+134, text)
		DrawTextBold(screen, "Use preferred currency always", w.X+10+10+4, w.Y+148, text)
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
