package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/hajimehoshi/ebiten/v2"
)

// MusicJukebox — a port of Ui/Windows/MusicJukebox.cpp: 366×109, centred,
// the skin's windowTitlebarColour and windowOptionsColour; "Currently
// playing:" with the playlist's tracks in a dropdown, the stop / play / next
// buttons, the playlist dropdown (current era, all, custom selection) and
// "Edit selection..." (the Music Selection window), enabled for the custom
// playlist.
//
// clickable(MusicJukebox.cpp): close_button music_controls_stop music_controls_play
// clickable(MusicJukebox.cpp): music_controls_next edit_selection
// clickable(MusicJukebox.cpp): music_playlist_btn currently_playing_btn
//
// The game fills State before each draw (prepareDraw) and takes the
// widgets' events through the On… callbacks.

const (
	mjW, mjH             = 366, 109
	imgMusicControlsStop = 2383 // ImageIds::music_controls_stop
	imgMusicControlsPlay = 2384
	imgMusicControlsNext = 2385
)

// The widgets' rectangles (x, y, w, h), as MusicJukebox.cpp's _widgets.
var (
	mjLabel       = [4]int{10, 27, 145, 12}
	mjPlaying     = [4]int{160, 27, 196, 12}
	mjControls    = [3][4]int{{10, 42, 24, 24}, {34, 42, 24, 24}, {58, 42, 24, 24}} // stop, play, next
	mjPlaylist    = [4]int{10, 69, 346, 12}
	mjEdit        = [4]int{183, 86, 173, 12}
	mjControlImgs = [3]int{imgMusicControlsStop, imgMusicControlsPlay, imgMusicControlsNext}
)

// MusicJukeboxState is what prepareDraw reads: the scene, the jukebox and
// the config.
type MusicJukeboxState struct {
	// SongName is the track shown ("[None]", the title music's name on the
	// title screen with title music on).
	SongName string
	// PlayMode is SceneManager::isPlayMode: outside it the track dropdown
	// and the three buttons are disabled.
	PlayMode bool
	// Playing is Jukebox::isMusicPlaying: play appears pressed, else stop.
	Playing bool
	// Playlist is Config audio.playlist (0 current era, 1 all, 2 custom).
	Playlist int
}

// MusicJukeboxWindow is the window.
type MusicJukeboxWindow struct {
	WinBase
	TitleColour, Colour uint8
	State               MusicJukeboxState
	// OnControl is a press of stop (0), play (1) or next (2).
	OnControl func(i int)
	// Tracks are the selected playlist's track titles and the index of the
	// one playing (-1 none) — currentlyPlayingMouseDown; OnTrack is the one
	// chosen there (currentlyPlayingDropdown).
	Tracks  func() (titles []string, current int)
	OnTrack func(i int)
	// OnPlaylist is musicPlaylistDropdown; OnEditSelection the button.
	OnPlaylist      func(i int)
	OnEditSelection func()
	OpenDropdown    func(*DropdownMenu)
	// ScreenW, ScreenH keep the dropdowns on the screen.
	ScreenW, ScreenH int
	hover            int // the control button under the mouse (-1 none)
}

// NewMusicJukebox is MusicJukebox::open's window (createWindowCentred).
func NewMusicJukebox(sw, sh int, titleColour, colour uint8) *MusicJukeboxWindow {
	return &MusicJukeboxWindow{WinBase: WinBase{WinType: WindowTypeMusicJukebox, X: (sw - mjW) / 2, Y: (sh - mjH) / 2, W: mjW, H: mjH},
		TitleColour: titleColour, Colour: colour, ScreenW: sw, ScreenH: sh, hover: -1}
}

func (w *MusicJukeboxWindow) in(r [4]int, px, py int) bool {
	return px >= w.X+r[0] && px < w.X+r[0]+r[2] && py >= w.Y+r[1] && py < w.Y+r[1]+r[3]
}

// musicPlaylistNames are playlist_string_ids.
func musicPlaylistNames() [3]string {
	return [3]string{i18n.Label(i18n.PlayOnlyMusicFromCurrentEra), i18n.Label(i18n.PlayAllMusic), i18n.Label(i18n.PlayCustomMusicSelection)}
}

func (w *MusicJukeboxWindow) screen() (int, int) {
	if w.ScreenW > 0 && w.ScreenH > 0 {
		return w.ScreenW, w.ScreenH
	}
	return 1 << 20, 1 << 20
}

// HandleClick is onMouseUp and onMouseDown.
func (w *MusicJukeboxWindow) HandleClick(px, py int) {
	sw, sh := w.screen()
	switch {
	case w.in([4]int{w.W - 15, 2, 13, 13}, px, py):
		w.Close()
	case w.in(mjEdit, px, py):
		if w.State.Playlist == 2 && w.OnEditSelection != nil { // disabled for the other playlists
			w.OnEditSelection()
		}
	case w.in(mjPlaying, px, py):
		// currentlyPlayingMouseDown: Dropdown::show at the combobox, width() − 4
		// by height() (Widget.cpp: right − left + 1, bottom − top + 1).
		if !w.State.PlayMode || w.Tracks == nil || w.OpenDropdown == nil {
			return
		}
		titles, current := w.Tracks()
		if len(titles) == 0 {
			return
		}
		items := make([]DropdownItem, len(titles))
		for i, t := range titles {
			items[i] = DropdownItem{Text: t, Selected: i == current, MoveX10: true}
		}
		d := NewDropdownShow(w.X+mjPlaying[0], w.Y+mjPlaying[1], mjPlaying[2]-4, mjPlaying[3], w.Colour, items, sw, sh)
		d.OnSelect = func(i int) {
			if w.OnTrack != nil {
				w.OnTrack(i)
			}
		}
		w.OpenDropdown(d)
	case w.in(mjPlaylist, px, py):
		// musicPlaylistMouseDown
		if w.OpenDropdown == nil {
			return
		}
		var items []DropdownItem
		for i, t := range musicPlaylistNames() {
			items = append(items, DropdownItem{Text: t, Selected: i == w.State.Playlist, MoveX10: true})
		}
		d := NewDropdownShow(w.X+mjPlaylist[0], w.Y+mjPlaylist[1], mjPlaylist[2]-4, mjPlaylist[3], w.Colour, items, sw, sh)
		d.OnSelect = func(i int) {
			if w.OnPlaylist != nil {
				w.OnPlaylist(i)
			}
		}
		w.OpenDropdown(d)
	default:
		for i, r := range mjControls {
			if w.in(r, px, py) && w.State.PlayMode && w.OnControl != nil {
				w.OnControl(i)
			}
		}
	}
}

// MouseOver is the control button under the mouse (ImageButton's hovered
// state).
func (w *MusicJukeboxWindow) MouseOver(px, py int) {
	w.hover = -1
	for i, r := range mjControls {
		if w.in(r, px, py) {
			w.hover = i
		}
	}
}

// WidgetTooltip is the tooltip of the widget under a screen point.
func (w *MusicJukeboxWindow) WidgetTooltip(px, py int) string {
	switch {
	case w.in([4]int{w.W - 15, 2, 13, 13}, px, py):
		return i18n.Str(i18n.TooltipCloseWindow)
	case w.in(mjEdit, px, py) && w.State.Playlist == 2:
		return i18n.Str(i18n.EditMusicSelectionTip)
	}
	if w.State.PlayMode {
		for i, id := range [3]int{i18n.MusicControlsStopTip, i18n.MusicControlsPlayTip, i18n.MusicControlsNextTip} {
			if w.in(mjControls[i], px, py) {
				return i18n.Str(id)
			}
		}
	}
	return ""
}

// Draw draws the window.
func (w *MusicJukeboxWindow) Draw(screen *ebiten.Image) {
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
	st := w.State
	activated := 0 // stop appears pressed, or play
	if st.Playing {
		activated = 1
	}
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, false)
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 15, w.W-1, w.H-1, gfx8.InsetColour{Map: sec}, 0) // Panel
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, w.W-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		optDropdown(rt, mjPlaying[0], mjPlaying[1], mjPlaying[2], sec)
		optDropdown(rt, mjPlaylist[0], mjPlaylist[1], mjPlaylist[2], sec)
		for i, r := range mjControls {
			drawImageButton(rt, r, mjControlImgs[i], w.Colour, st.PlayMode && i == activated, st.PlayMode && i == w.hover, !st.PlayMode)
		}
		gfx8.FillRectInset(rt, mjEdit[0], mjEdit[1], mjEdit[0]+mjEdit[2]-1, mjEdit[1]+mjEdit[3]-1, gfx8.InsetColour{Map: sec}, 0) // Button
	})
	DrawWindowCaption(screen, w.X+1, w.Y+1, w.W-2, i18n.Label(i18n.JukeboxWindowTitle))
	text, black := pal[gfx8.Shade(sec, 9)], pal[paletteBlack0]
	inset := func(t string, x, y int) { DrawTextInset(screen, t, x, y, sec) } // a disabled widget: colour.inset()
	DrawTextBold(screen, i18n.Label(i18n.CurrentlyPlaying), w.X+mjLabel[0], w.Y+mjLabel[1], text)
	arrow := string(rune(170)) // StringIds::dropdown "▼"
	aw, _ := MeasureTextBold(arrow)
	for _, c := range []struct {
		r        [4]int
		value    string
		disabled bool
	}{{mjPlaying, st.SongName, !st.PlayMode}, {mjPlaylist, musicPlaylistNames()[min(max(st.Playlist, 0), 2)], false}} {
		v := ClipString(c.value, G1FontMediumBold, c.r[2]-15)
		ax := w.X + c.r[0] + c.r[2] - 12 + (11-1)/2 - aw/2
		if c.disabled {
			inset(v, w.X+c.r[0]+1, w.Y+c.r[1]+1)
			inset(arrow, ax, w.Y+c.r[1]+1)
			continue
		}
		DrawTextBold(screen, v, w.X+c.r[0]+1, w.Y+c.r[1]+1, black)
		DrawTextBold(screen, arrow, ax, w.Y+c.r[1]+1, black)
	}
	// Button's label, centred: x + (w − 1) / 2.
	label := i18n.Label(i18n.EditMusicSelection)
	lw, _ := MeasureTextBold(label)
	lx := w.X + mjEdit[0] + (mjEdit[2]-1)/2 - lw/2
	if st.Playlist != 2 {
		inset(label, lx, w.Y+mjEdit[1])
	} else {
		DrawTextBold(screen, label, lx, w.Y+mjEdit[1], text)
	}
}

// drawImageButton is ImageButton::draw for a widget r of an opaque window
// colour with an image recoloured to it: hovered (and not disabled) on a
// translucent inset; activated on an inset pressed in; else the image alone;
// disabled, the image solid in shade 6 one pixel down-right and shade 4 over
// it (drawImage).
func drawImageButton(rt *gfx8.RenderTarget, r [4]int, image int, colour uint8, activated, hovered, disabled bool) {
	g1 := globalG1Font.g1
	cmap, err := g1.GetPaletteMap(int(colour))
	if err != nil || image <= 0 || image >= len(g1.Elements) {
		return
	}
	flags := 0
	if activated {
		flags = gfx8.InsetBorderInset
	}
	switch {
	case hovered && !disabled:
		gfx8.FillRectInset(rt, r[0], r[1], r[0]+r[2]-1, r[1]+r[3]-1, translucentInset(colour), flags)
	case activated:
		gfx8.FillRectInset(rt, r[0], r[1], r[0]+r[2]-1, r[1]+r[3]-1, gfx8.InsetColour{Map: cmap}, flags)
	}
	sp, err := gfx8.Decode(&g1.Elements[image])
	if err != nil {
		return
	}
	if disabled {
		gfx8.DrawSprite(rt, sp, r[0]+1, r[1]+1, gfx8.OpTransparent|gfx8.OpSrc, solidMap(gfx8.Shade(cmap, 6)))
		gfx8.DrawSprite(rt, sp, r[0], r[1], gfx8.OpTransparent|gfx8.OpSrc, solidMap(gfx8.Shade(cmap, 4)))
		return
	}
	gfx8.DrawSpriteNoise(rt, sp, r[0], r[1], gfx8.OpTransparent|gfx8.OpSrc, cmap, nil)
}
