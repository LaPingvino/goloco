package main

import (
	"log"
	"strconv"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
)

// options.go — the Options window's wiring to the game and the config file,
// and the currency Options and scenario loads choose.

// openOptionsWindow opens an Options window with proper 3D toggle buttons for
// sound and music, styled after OpenLoco's option panels.
func (g *Game) openOptionsWindow() {
	// The ported Options window, Audio tab (pkg/ui/options.go).
	var tc, oc uint8 = 0, 1
	if sk := g.skin(); sk != nil {
		tc, oc = sk.WindowTitlebarColour, sk.WindowOptionsColour
	}
	w := ui.NewOptions(g.sw, g.sh, tc, oc)
	w.TitleMusic = g.musicEnabled
	w.MusicVolume = 1
	if g.audioMgr != nil {
		w.MusicVolume = g.audioMgr.MusicVolume()
	}
	w.OpenDropdown = func(d *ui.DropdownMenu) { g.dropdown = d }
	// Screen mode: Ebitengine's fullscreen stands for both fullscreen modes
	// (it has no exclusive mode); windowed leaves it.
	if ebiten.IsFullscreen() {
		w.ScreenMode = 2
	}
	w.OnScreenMode = func(m int) { ebiten.SetFullscreen(m != 0) }
	// getFullscreenResolutions: SDL lists the display's modes; Ebitengine's
	// fullscreen always uses the monitor's own mode, so that is the one
	// resolution it can offer (PLATFORM: no mode switching).
	if g.fsResolution == [2]int{} {
		g.fsResolution = [2]int{1920, 1080} // Config default
	}
	w.Resolution = g.fsResolution
	if mw, mh := ebiten.Monitor().Size(); mw > 0 && mh > 0 {
		w.Resolutions = [][2]int{{mw, mh}}
	}
	w.OnResolution = func(i int) {
		g.fsResolution = w.Resolutions[i]
		ebiten.SetFullscreen(true)
	}
	// Frame rate limit: vertical sync = Ebitengine's vsync; unrestricted =
	// vsync off; internal speed draws with vsync too (upstream draws only at
	// its update rate — goloco draws every frame; noted in LOOP).
	w.FrameLimit = g.frameLimit
	w.OnFrameLimit = func(m int) {
		g.frameLimit = m
		ebiten.SetVsyncEnabled(m != 2)
	}
	// Regional tab: the scenario's currency object, the preferred one
	// (Config default CURRDOLL), and title mode.
	w.TitleMode = g.inTitleScreen
	if g.r.ObjMgr != nil {
		if c := g.r.ObjMgr.GetCurrencyObject(g.r.ObjMgr.CurrencySlot); c != nil {
			w.CurrencyName = c.Name
		}
		if g.preferredCurrency == "" {
			g.preferredCurrency = "CURRDOLL" // Config default
		}
		if c := g.r.ObjMgr.GetCurrencyObject(g.preferredCurrency); c != nil {
			w.PreferredCurrencyName = c.Name
		}
		w.Currencies = g.r.ObjMgr.AvailableCurrencies()
		langs := i18n.Available()
		for i, l := range langs {
			w.Languages = append(w.Languages, l.NativeName)
			if l.Locale == i18n.Locale() {
				w.Language = i
			}
		}
		w.OnLanguage = func(i int) { // languageDropdown: the pack, then Config::write
			if err := i18n.Load(langs[i].Locale); err != nil {
				log.Printf("[Options] language: %v", err)
				return
			}
			if err := setConfigValue("language", langs[i].Locale); err != nil {
				log.Printf("[Options] saving the language: %v", err)
			}
		}
		w.OnPreferredCurrency = func(dat string) { // preferredCurrencyDropdown
			g.preferredCurrency = dat
			g.loadPreferredCurrencyAlways()
			if c := g.r.ObjMgr.GetCurrencyObject(g.r.ObjMgr.CurrencySlot); c != nil {
				w.CurrencyName = c.Name
			}
		}
		w.PreferredCurrency = g.preferredCurrencyFlags
		w.OnPreferredToggle = func(i int, v bool) { // PARTIAL: Config::write
			g.preferredCurrencyFlags[i] = v
			if i == 1 {
				g.loadPreferredCurrencyAlways()
				if c := g.r.ObjMgr.GetCurrencyObject(g.r.ObjMgr.CurrencySlot); c != nil {
					w.CurrencyName = c.Name
				}
			}
		}
		w.OnCurrency = func(dat string) { useCurrency(g.r.ObjMgr, dat) } // currencyDropdown
	}
	// Controls tab: Config::edgeScrolling (default on). zoomToCursor and
	// invertRightMouseViewPan are kept in the window only: goloco has neither yet.
	w.Controls[0] = !g.noEdgeScrolling
	w.OnControl = func(i int, v bool) {
		if i == 0 {
			g.noEdgeScrolling = !v
		}
	}
	w.Scale = g.scale
	w.OnScale = func(s float64) { // setWindowScaling: Config::write, then the resize (Layout)
		g.scale = s
		setConfigValue("scaleFactor", strconv.FormatFloat(s, 'f', 2, 64))
	}
	w.ShowFPS = g.showFPS
	w.OnShowFPS = func(on bool) {
		g.showFPS = on
		v := "false"
		if on {
			v = "true"
		}
		setConfigValue("showFPS", v)
	}
	w.OnTitleMusic = func(on bool) {
		g.musicEnabled = on
		if !on && g.audioMgr != nil {
			g.audioMgr.StopMusic()
		}
	}
	w.OnMusicVol = func(v float64) {
		if g.audioMgr != nil {
			g.audioMgr.SetMusicVolume(v)
		}
	}
	g.showCentred(w)
}

// useCurrency makes dat the loaded currency object: its slot, then
// Gfx::loadCurrency and formatCurrency's object (scenario load, and Options'
// currencyDropdown).
func useCurrency(om *objects.ObjectManager, dat string) {
	om.CurrencySlot = dat
	if c := om.GetCurrencyObject(dat); c != nil {
		ui.LoadCurrency(c.ObjectIcon)
		uilayout.SetCurrency(c.Prefix, c.Suffix, c.Separator, c.Factor)
	}
}

// loadPreferredCurrency is Scenario.cpp loadPreferredCurrency: the preferred
// currency becomes the loaded one, unless it already is or it can't be
// loaded (then the old one stays).
func (g *Game) loadPreferredCurrency() {
	if g.r.ObjMgr == nil || g.preferredCurrency == "" || g.preferredCurrency == g.r.ObjMgr.CurrencySlot {
		return
	}
	if g.r.ObjMgr.GetCurrencyObject(g.preferredCurrency) == nil {
		return
	}
	useCurrency(g.r.ObjMgr, g.preferredCurrency)
}

// loadPreferredCurrencyAlways: only with usePreferredCurrencyAlways.
func (g *Game) loadPreferredCurrencyAlways() {
	if g.preferredCurrencyFlags[1] {
		g.loadPreferredCurrency()
	}
}

// loadPreferredCurrencyNewGame: with usePreferredCurrencyForNewGames, else
// as loadPreferredCurrencyAlways.
func (g *Game) loadPreferredCurrencyNewGame() {
	if !g.preferredCurrencyFlags[0] {
		g.loadPreferredCurrencyAlways()
		return
	}
	g.loadPreferredCurrency()
}
