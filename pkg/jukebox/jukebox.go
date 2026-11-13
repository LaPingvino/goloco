// Package jukebox is the in-game music's track choice — a port of
// OpenLoco's Jukebox.cpp: the thirty tracks with their eras, the three
// playlists (the current era's, all, the player's own selection), the next
// track, and the "Play Music" switch.
//
// OpenLoco reference: src/OpenLoco/src/Jukebox.cpp,
// src/OpenLoco/include/OpenLoco/Jukebox.h
package jukebox

import (
	"math/rand"
	"sort"

	"github.com/LaPingvino/goloco/pkg/i18n"
)

// MusicID is Jukebox::MusicId: a track's index in kMusicInfo.
type MusicID = uint8

const (
	NumMusicTracks          = 30   // kNumMusicTracks
	NoSong          MusicID = 0xFF // kNoSong
	BeginningOfTime         = 0    // kBeginningOfTime
	EndOfTime               = 9999 // kEndOfTime
)

// MusicInfo is Jukebox::MusicInfo, with the track's file name under the
// Locomotion Data folder (upstream: its PathId) and its config key.
type MusicInfo struct {
	File               string
	Key                string
	TitleID            int
	StartYear, EndYear int
}

// SortMode is Jukebox::MusicSortMode.
type SortMode uint8

const (
	SortOriginal SortMode = iota // internal id order, as vanilla
	SortTitleAscending
	SortTitleDescending
	SortYearsAscending
	SortYearsDescending
)

// PlaylistType is Config::MusicPlaylistType.
type PlaylistType uint8

const (
	PlaylistCurrentEra PlaylistType = iota
	PlaylistAll
	PlaylistCustom
)

// Config is the part of Config::Audio the jukebox reads and sets.
type Config struct {
	PlayJukeboxMusic bool
	Playlist         PlaylistType
	CustomJukebox    [NumMusicTracks]bool
}

// DefaultConfig is Config.cpp's defaults: music on, the current era's
// playlist, every track but the title music in the custom selection
// (resetPlaylistConfig).
func DefaultConfig() Config {
	c := Config{PlayJukeboxMusic: true}
	c.ResetPlaylist()
	return c
}

// ResetPlaylist is Config.cpp resetPlaylistConfig.
func (c *Config) ResetPlaylist() {
	for i := range c.CustomJukebox {
		c.CustomJukebox[i] = true
	}
	c.CustomJukebox[NumMusicTracks-1] = false // PlaylistItem::locomotionTitle
}

// Jukebox is Jukebox.cpp's state and what it reaches outside itself: the
// config, Config::write, the current year (getCurrentYear), Audio::stopMusic
// and std::rand. A nil Write or StopMusic does nothing; a nil Rand is
// math/rand — never the game's own generator, which the simulation owns.
type Jukebox struct {
	Cfg       *Config
	Write     func()
	Year      func() int
	StopMusic func()
	Rand      func(n int) int

	current, previous, requestedNext MusicID
}

// New is a jukebox on a config, playing nothing.
func New(cfg *Config) *Jukebox {
	j := &Jukebox{Cfg: cfg}
	j.Reset()
	return j
}

// Info is getMusicInfo.
func Info(track MusicID) MusicInfo { return musicInfo[track] }

// IsMusicPlaying is isMusicPlaying.
func (j *Jukebox) IsMusicPlaying() bool {
	return j.current != NoSong && j.Cfg.PlayJukeboxMusic
}

// CurrentTrack is getCurrentTrack.
func (j *Jukebox) CurrentTrack() MusicID { return j.current }

// SelectedTrackTitleID is getSelectedTrackTitleId: the requested track's
// title, else the current one's, else "[None]".
func (j *Jukebox) SelectedTrackTitleID() int {
	if j.requestedNext != NoSong {
		return musicInfo[j.requestedNext].TitleID
	}
	if j.current != NoSong {
		return musicInfo[j.current].TitleID
	}
	return i18n.MusicNone
}

// AllMusicPlaylist is makeAllMusicPlaylist: every track, in an order.
// PLATFORM: the titles compare as byte strings where upstream uses the C
// locale's strcoll.
func AllMusicPlaylist(ordering SortMode) []MusicID {
	p := make([]MusicID, NumMusicTracks)
	for i := range p {
		p[i] = MusicID(i)
	}
	years := func(a, b MusicID) (bool, bool) { // a before b ascending; equal
		x, y := musicInfo[a], musicInfo[b]
		if x.StartYear != y.StartYear {
			return x.StartYear < y.StartYear, false
		}
		return x.EndYear < y.EndYear, x.EndYear == y.EndYear
	}
	switch ordering {
	case SortTitleAscending, SortTitleDescending:
		titles := make([]string, NumMusicTracks)
		for i := range titles {
			titles[i] = i18n.Label(musicInfo[i].TitleID)
		}
		sort.SliceStable(p, func(a, b int) bool {
			if ordering == SortTitleDescending {
				return titles[p[a]] > titles[p[b]]
			}
			return titles[p[a]] < titles[p[b]]
		})
	case SortYearsAscending:
		sort.SliceStable(p, func(a, b int) bool { less, _ := years(p[a], p[b]); return less })
	case SortYearsDescending:
		sort.SliceStable(p, func(a, b int) bool { less, eq := years(p[a], p[b]); return !less && !eq })
	}
	return p
}

// currentEraPlaylist is makeCurrentEraPlaylist.
func (j *Jukebox) currentEraPlaylist() []MusicID {
	var p []MusicID
	year := 0
	if j.Year != nil {
		year = j.Year()
	}
	for i, mi := range musicInfo {
		if year >= mi.StartYear && year <= mi.EndYear {
			p = append(p, MusicID(i))
		}
	}
	return p
}

// customSelectionPlaylist is makeCustomSelectionPlaylist.
func (j *Jukebox) customSelectionPlaylist() []MusicID {
	var p []MusicID
	for i, on := range j.Cfg.CustomJukebox {
		if on {
			p = append(p, MusicID(i))
		}
	}
	return p
}

// SelectedPlaylist is makeSelectedPlaylist: the configured playlist's tracks.
func (j *Jukebox) SelectedPlaylist() []MusicID {
	switch j.Cfg.Playlist {
	case PlaylistAll:
		return AllMusicPlaylist(SortOriginal)
	case PlaylistCustom:
		return j.customSelectionPlaylist()
	}
	return j.currentEraPlaylist()
}

// chooseNextMusicTrack: a random track of the selected playlist (an empty
// one falls back to the current era's, then to all), never the last one
// again unless it is the only one.
func (j *Jukebox) chooseNextMusicTrack(last MusicID) MusicID {
	p := j.SelectedPlaylist()
	if len(p) == 0 && j.Cfg.Playlist != PlaylistCurrentEra {
		p = j.currentEraPlaylist()
	}
	if len(p) == 0 {
		p = AllMusicPlaylist(SortOriginal)
	}
	if len(p) > 1 {
		for i, t := range p {
			if t == last {
				p = append(p[:i:i], p[i+1:]...)
				break
			}
		}
	}
	if j.Rand != nil {
		return p[j.Rand(len(p))]
	}
	return p[rand.Intn(len(p))]
}

// ChangeTrack is changeTrack: the requested track, or a chosen one, becomes
// the current track; its info is what the caller plays.
func (j *Jukebox) ChangeTrack() MusicInfo {
	j.previous = j.current
	if j.requestedNext != NoSong {
		j.current, j.requestedNext = j.requestedNext, NoSong
	} else {
		j.current = j.chooseNextMusicTrack(j.previous)
	}
	return musicInfo[j.current]
}

func (j *Jukebox) stopMusic() {
	if j.StopMusic != nil {
		j.StopMusic()
	}
}

func (j *Jukebox) write() {
	if j.Write != nil {
		j.Write()
	}
}

// RequestTrack is requestTrack: the player picks a track from the window's
// drop-down; the current one stops so it plays next. False: it is playing.
func (j *Jukebox) RequestTrack(track MusicID) bool {
	if track == j.current {
		return false
	}
	j.requestedNext = track
	j.stopMusic()
	return true
}

// SkipCurrentTrack is skipCurrentTrack: the music stops, so the next tick
// plays another track. False with the music switched off.
func (j *Jukebox) SkipCurrentTrack() bool {
	if !j.Cfg.PlayJukeboxMusic {
		return false
	}
	j.stopMusic()
	return true
}

// DisableMusic is disableMusic ("Play Music" unticked, or the stop button).
func (j *Jukebox) DisableMusic() bool {
	if !j.Cfg.PlayJukeboxMusic {
		return false
	}
	j.Cfg.PlayJukeboxMusic = false
	j.write()
	j.stopMusic()
	j.current = NoSong
	return true
}

// EnableMusic is enableMusic.
func (j *Jukebox) EnableMusic() bool {
	if j.Cfg.PlayJukeboxMusic {
		return false
	}
	j.Cfg.PlayJukeboxMusic = true
	j.write()
	return true
}

// Reset is resetJukebox.
func (j *Jukebox) Reset() {
	j.current, j.previous, j.requestedNext = NoSong, NoSong, NoSong
}

// TrackStillApplies is the test of Audio::revalidateCurrentTrack (Music.cpp):
// the current track belongs to the configured playlist — true with no track.
func (j *Jukebox) TrackStillApplies() bool {
	if j.current == NoSong {
		return true
	}
	switch j.Cfg.Playlist {
	case PlaylistCurrentEra:
		year, mi := 0, musicInfo[j.current]
		if j.Year != nil {
			year = j.Year()
		}
		return year >= mi.StartYear && year <= mi.EndYear
	case PlaylistCustom:
		return j.Cfg.CustomJukebox[j.current]
	}
	return true
}
