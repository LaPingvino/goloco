package main

import (
	"log"
	"path/filepath"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/jukebox"
	"github.com/LaPingvino/goloco/pkg/ui"
)

// music.go — the game's music: a port of Audio/Music.cpp (the background
// music's state machine over the jukebox), Audio.cpp's toggleSound /
// pauseSound for the music channel, Title.cpp's title music, and the audio
// section of the config (Config.cpp) the jukebox reads.
//
// OpenLoco reference: src/OpenLoco/src/Audio/Music.cpp, Audio/Audio.cpp,
// Title.cpp, Config.cpp
//
// PLATFORM: without a sound device (or with GOLOCO_MUTE, or a track's file
// missing — the web page without its music pack) a track "plays" silently
// until something stops it: the jukebox keeps one current track, as it does
// with sound, where upstream would find no track playing and choose a new one
// every tick.

// titleMusicFile is PathId::css5 under the Data folder.
const titleMusicFile = "CSS5.DAT"

var playlistNames = [...]string{"currentEra", "all", "custom"} // ConfigConvert.hpp kMusicPlaylistTypes

// loadAudioConfig is Config.cpp's read of the audio section's music keys
// (their defaults: both musics on, the current era's playlist, every track
// but the title music in the custom selection).
func loadAudioConfig() (cfg jukebox.Config, playTitleMusic bool) {
	cfg = jukebox.DefaultConfig()
	cfg.PlayJukeboxMusic = configValue("audio.playJukeboxMusic") != "false"
	playTitleMusic = configValue("audio.play_title_music") != "false"
	for i, n := range playlistNames {
		if configValue("audio.playlist") == n {
			cfg.Playlist = jukebox.PlaylistType(i)
		}
	}
	for i := range jukebox.NumMusicTracks {
		switch configValue("audio.customJukebox." + jukebox.Info(jukebox.MusicID(i)).Key) {
		case "true":
			cfg.CustomJukebox[i] = true
		case "false":
			cfg.CustomJukebox[i] = false
		}
	}
	return cfg, playTitleMusic
}

// writeAudioConfig is Config::write for the audio section's music keys.
func (g *Game) writeAudioConfig() {
	b := func(v bool) string {
		if v {
			return "true"
		}
		return "false"
	}
	kvs := [][2]string{
		{"audio.playJukeboxMusic", b(g.audioCfg.PlayJukeboxMusic)},
		{"audio.play_title_music", b(g.playTitleMusic)},
		{"audio.playlist", playlistNames[g.audioCfg.Playlist]},
	}
	for i, on := range g.audioCfg.CustomJukebox {
		kvs = append(kvs, [2]string{"audio.customJukebox." + jukebox.Info(jukebox.MusicID(i)).Key, b(on)})
	}
	if err := setConfigValues(kvs); err != nil {
		log.Printf("[Audio] saving the audio settings: %v", err)
	}
}

// initMusic gives the game its jukebox on the saved audio settings.
func (g *Game) initMusic() {
	g.audioCfg, g.playTitleMusic = loadAudioConfig()
	g.jukebox = jukebox.New(&g.audioCfg)
	g.jukebox.Write = g.writeAudioConfig
	g.jukebox.Year = func() int { return int(g.w.Map.CurrentYear) }
	g.jukebox.StopMusic = g.stopMusic
}

// isAudioEnabled is Audio::isAudioEnabled (the Audio menu's "Mute" off).
func (g *Game) isAudioEnabled() bool { return !g.audioMuted }

// toggleSound is Audio::toggleSound, for the music channel (goloco has no
// vehicle or ambient noise yet): muting stops the music; the next tick with
// sound starts a track again.
func (g *Game) toggleSound() {
	g.audioMuted = !g.audioMuted
	if !g.audioMuted {
		return
	}
	g.stopMusic()
}

// playMusic is Audio::playMusic: a file of the Data folder on the music
// channel, in place of what played.
func (g *Game) playMusic(file string, loop bool) {
	if !g.isAudioEnabled() {
		return
	}
	g.stopMusic()
	if g.audioMgr == nil || g.dataDir == "" {
		g.silentTrack = true
		return
	}
	log.Printf("[Audio] music: %s", file)
	if err := g.audioMgr.LoadAndPlayMusic(filepath.Join(g.dataDir, file), loop); err != nil {
		log.Printf("[Audio] %s: %v (the track passes in silence)", file, err)
		g.silentTrack = true
	}
}

// stopMusic is Audio::stopMusic.
func (g *Game) stopMusic() {
	g.silentTrack, g.musicPaused = false, false
	g.audioMgr.StopMusic()
}

// musicPlaying is playBackgroundMusic's "musicPlaying": the music channel
// has a track that has not ended.
func (g *Game) musicPlaying() bool { return g.silentTrack || g.audioMgr.IsMusicPlaying() }

// resetMusic is Audio::resetMusic (a scenario starting, a game loaded).
func (g *Game) resetMusic() {
	g.stopMusic()
	g.jukebox.Reset()
}

// playBackgroundMusic is Audio::playBackgroundMusic, once a tick in a game:
// when no track plays (none yet, the last one over, or stopped to skip it),
// the jukebox's next one starts. PARTIAL: no editor mode to exclude.
func (g *Game) playBackgroundMusic() {
	if !g.audioCfg.PlayJukeboxMusic || !g.isAudioEnabled() || g.inTitleScreen || g.isPaused {
		return
	}
	if !g.musicPlaying() {
		g.playMusic(g.jukebox.ChangeTrack().File, false)
	}
}

// revalidateCurrentTrack is Audio::revalidateCurrentTrack: a track the
// playlist no longer holds stops.
func (g *Game) revalidateCurrentTrack() {
	if !g.jukebox.TrackStillApplies() {
		g.resetMusic()
	}
}

// updateMusic is the music's share of a game frame: pauseSound /
// unpauseSound as the game pauses and runs again (Game.cpp, SceneManager.cpp
// — the title's music is not paused), then playBackgroundMusic
// (GameScene.cpp tick).
func (g *Game) updateMusic() {
	if paused := g.isPaused && !g.inTitleScreen; paused != g.musicPaused {
		g.musicPaused = paused
		if paused {
			g.audioMgr.PauseMusic()
		} else {
			g.audioMgr.UnpauseMusic()
		}
	}
	g.playBackgroundMusic()
}

// startTitleMusic is Title::start's music: the title tune, looping, when
// "Play title music" is on.
func (g *Game) startTitleMusic() {
	g.resetMusic()
	if g.playTitleMusic {
		g.playMusic(titleMusicFile, true)
	}
}

// setPlayTitleMusic is Options' playTitleMusicOnMouseUp.
func (g *Game) setPlayTitleMusic(on bool) {
	g.playTitleMusic = on
	g.writeAudioConfig()
	if !g.inTitleScreen {
		return
	}
	if on {
		g.playMusic(titleMusicFile, true)
	} else {
		g.stopMusic()
	}
}

// togglePlayMusic is the Audio menu's "Play Music" (ToolbarTop.cpp
// audioMenuDropdown) and Options' play_game_music in a game.
func (g *Game) togglePlayMusic() {
	if g.audioCfg.PlayJukeboxMusic {
		g.jukebox.DisableMusic()
	} else {
		g.jukebox.EnableMusic()
	}
}

// setPlaylist is MusicJukebox's musicPlaylistDropdown.
func (g *Game) setPlaylist(p jukebox.PlaylistType) {
	g.audioCfg.Playlist = p
	g.writeAudioConfig()
	if !g.inTitleScreen { // "Prevents title music from stopping"
		g.revalidateCurrentTrack()
	}
	g.wins.CloseType(ui.WindowTypeMusicSelection, ui.AnyNumber)
}

// optionsColours are the skin's windowTitlebarColour and windowOptionsColour
// (Options, the jukebox, the music selection).
func (g *Game) optionsColours() (tc, oc uint8) {
	if sk := g.skin(); sk != nil {
		return sk.WindowTitlebarColour, sk.WindowOptionsColour
	}
	return 0, 1
}

// musicJukeboxState is MusicJukebox.cpp prepareDraw's reading of the scene,
// the jukebox and the config.
func (g *Game) musicJukeboxState() ui.MusicJukeboxState {
	st := ui.MusicJukeboxState{SongName: i18n.Label(i18n.MusicNone), PlayMode: !g.inTitleScreen, Playlist: int(g.audioCfg.Playlist)}
	switch {
	case st.PlayMode:
		st.SongName = i18n.Label(g.jukebox.SelectedTrackTitleID())
		st.Playing = g.jukebox.IsMusicPlaying()
	case g.playTitleMusic:
		st.SongName = i18n.Label(i18n.MusicLocomotionTitle)
	}
	return st
}

// openMusicJukebox is MusicJukebox::open (the Audio menu's "Open jukebox",
// Options' jukebox button).
func (g *Game) openMusicJukebox() {
	if w := g.jukeboxWin(); w != nil {
		g.wins.BringToFront(w)
		return
	}
	tc, oc := g.optionsColours()
	w := ui.NewMusicJukebox(g.sw, g.sh, tc, oc)
	w.State = g.musicJukeboxState()
	w.OpenDropdown = func(d *ui.DropdownMenu) { g.dropdown = d }
	w.OnControl = func(i int) {
		switch i {
		case 0: // stopMusic
			g.jukebox.DisableMusic()
		case 1: // playMusic
			g.jukebox.EnableMusic()
		case 2: // playNextSong
			g.jukebox.SkipCurrentTrack()
		}
	}
	w.Tracks = func() (titles []string, current int) {
		current = -1
		for i, t := range g.jukebox.SelectedPlaylist() {
			titles = append(titles, i18n.Label(jukebox.Info(t).TitleID))
			if t == g.jukebox.CurrentTrack() {
				current = i
			}
		}
		return titles, current
	}
	w.OnTrack = func(i int) {
		if p := g.jukebox.SelectedPlaylist(); i >= 0 && i < len(p) {
			g.jukebox.RequestTrack(p[i])
		}
	}
	w.OnPlaylist = func(i int) { g.setPlaylist(jukebox.PlaylistType(i)) }
	w.OnEditSelection = g.openMusicSelection
	g.showCentred(w)
}

// openMusicSelection is MusicSelection::open.
func (g *Game) openMusicSelection() {
	if w := g.wins.Find(ui.WindowTypeMusicSelection, ui.AnyNumber); w != nil {
		g.wins.BringToFront(w)
		return
	}
	tc, oc := g.optionsColours()
	w := ui.NewMusicSelection(&g.audioCfg, tc, oc)
	w.OnChange = func() { // onScrollMouseDown
		g.writeAudioConfig()
		g.revalidateCurrentTrack()
	}
	g.showAuto(w, 0, 0)
}
