package jukebox

import (
	"reflect"
	"testing"

	"github.com/LaPingvino/goloco/pkg/i18n"
)

func newTest(year int) (*Jukebox, *Config, *int) {
	cfg := DefaultConfig()
	stops := 0
	j := New(&cfg)
	j.Year = func() int { return year }
	j.StopMusic = func() { stops++ }
	j.Rand = func(n int) int { return n - 1 } // the playlist's last track
	return j, &cfg, &stops
}

// The table is upstream's: thirty tracks, the title music last and in no era.
func TestMusicInfo(t *testing.T) {
	if got := Info(0); got.File != "20s1.DAT" || got.TitleID != i18n.MusicChugginAlong || got.StartYear != 1925 || got.EndYear != 1933 {
		t.Errorf("track 0: %+v", got)
	}
	if got := Info(29); got.File != "CSS5.DAT" || got.Key != "locomotionTitle" || got.StartYear != EndOfTime || got.EndYear != BeginningOfTime {
		t.Errorf("track 29: %+v", got)
	}
	seen := map[string]bool{}
	for i := range NumMusicTracks {
		mi := Info(MusicID(i))
		if seen[mi.File] || seen[mi.Key] || mi.File == "" || mi.Key == "" {
			t.Errorf("track %d: file or key missing or twice: %+v", i, mi)
		}
		seen[mi.File], seen[mi.Key] = true, true
	}
}

// makeSelectedPlaylist for each playlist type (years checked against
// kMusicInfo by hand: 1950 has A Traveller's Serenade 1940–48 no, Latino
// Trip 1943–51, A Good Head of Steam 1950–58, Hop to the Bop 1946–54).
func TestSelectedPlaylist(t *testing.T) {
	j, cfg, _ := newTest(1950)
	if got, want := j.SelectedPlaylist(), []MusicID{18, 19, 20}; !reflect.DeepEqual(got, want) {
		t.Errorf("current era 1950: %v, want %v", got, want)
	}
	cfg.Playlist = PlaylistAll
	if got := j.SelectedPlaylist(); len(got) != 30 || got[0] != 0 || got[29] != 29 {
		t.Errorf("all: %v", got)
	}
	cfg.Playlist = PlaylistCustom
	if got := j.SelectedPlaylist(); len(got) != 29 || got[28] != 28 {
		t.Errorf("custom default (all but the title music): %v", got)
	}
	cfg.CustomJukebox = [NumMusicTracks]bool{3: true, 7: true}
	if got, want := j.SelectedPlaylist(), []MusicID{3, 7}; !reflect.DeepEqual(got, want) {
		t.Errorf("custom: %v, want %v", got, want)
	}
}

// chooseNextMusicTrack: no track twice in a row, the fallbacks of an empty
// playlist, and a requested track first.
func TestChangeTrack(t *testing.T) {
	j, cfg, stops := newTest(1950)
	if mi := j.ChangeTrack(); j.CurrentTrack() != 20 || mi.File != "40s3.DAT" {
		t.Fatalf("first: track %d %+v", j.CurrentTrack(), mi)
	}
	j.ChangeTrack() // 20 is taken out: the last of {18, 19}
	if j.CurrentTrack() != 19 {
		t.Errorf("second: %d, want 19", j.CurrentTrack())
	}
	if !j.IsMusicPlaying() || j.SelectedTrackTitleID() != i18n.MusicAGoodHeadOfSteam {
		t.Errorf("playing %v, title %d", j.IsMusicPlaying(), j.SelectedTrackTitleID())
	}
	// An empty custom selection plays the current era's.
	cfg.Playlist, cfg.CustomJukebox = PlaylistCustom, [NumMusicTracks]bool{}
	j.Reset()
	if j.ChangeTrack(); j.CurrentTrack() != 20 {
		t.Errorf("empty custom: %d, want 20", j.CurrentTrack())
	}
	// A single track plays again.
	cfg.CustomJukebox[5] = true
	j.ChangeTrack()
	if j.ChangeTrack(); j.CurrentTrack() != 5 {
		t.Errorf("single: %d, want 5", j.CurrentTrack())
	}
	// A request stops the music and is next; the current one is no request.
	if j.RequestTrack(5) || *stops != 0 {
		t.Errorf("requesting the current track: stops %d", *stops)
	}
	if !j.RequestTrack(12) || *stops != 1 || j.SelectedTrackTitleID() != i18n.MusicChrysanthemum {
		t.Errorf("request: stops %d, title %d", *stops, j.SelectedTrackTitleID())
	}
	if j.ChangeTrack(); j.CurrentTrack() != 12 {
		t.Errorf("requested: %d, want 12", j.CurrentTrack())
	}
}

// A year no track belongs to (the 1990s have three, 1922 has Sandy Track
// Blues and Get Me to Gladstone Bay): the era's list is never empty in the
// game's years, but year 0 has three ragtime tracks from the beginning of time.
func TestCurrentEraEarly(t *testing.T) {
	j, _, _ := newTest(0)
	if got, want := j.SelectedPlaylist(), []MusicID{12, 13, 15}; !reflect.DeepEqual(got, want) {
		t.Errorf("year 0: %v, want %v", got, want)
	}
}

func TestEnableDisableSkip(t *testing.T) {
	j, cfg, stops := newTest(1950)
	writes := 0
	j.Write = func() { writes++ }
	j.ChangeTrack()
	if j.EnableMusic() || writes != 0 {
		t.Error("enabling music that is on")
	}
	if !j.SkipCurrentTrack() || *stops != 1 {
		t.Errorf("skip: stops %d", *stops)
	}
	if !j.DisableMusic() || cfg.PlayJukeboxMusic || writes != 1 || *stops != 2 || j.CurrentTrack() != NoSong || j.IsMusicPlaying() {
		t.Errorf("disable: cfg %v writes %d stops %d track %d", cfg.PlayJukeboxMusic, writes, *stops, j.CurrentTrack())
	}
	if j.SelectedTrackTitleID() != i18n.MusicNone {
		t.Errorf("title after disabling: %d", j.SelectedTrackTitleID())
	}
	if j.DisableMusic() || j.SkipCurrentTrack() || *stops != 2 {
		t.Error("disable or skip with the music off")
	}
	if !j.EnableMusic() || !cfg.PlayJukeboxMusic || writes != 2 {
		t.Errorf("enable: cfg %v writes %d", cfg.PlayJukeboxMusic, writes)
	}
}

// Audio::revalidateCurrentTrack's test.
func TestTrackStillApplies(t *testing.T) {
	year := 1950
	j, cfg, _ := newTest(0)
	j.Year = func() int { return year }
	if !j.TrackStillApplies() {
		t.Error("no track")
	}
	j.ChangeTrack() // 20: Hop to the Bop, 1946–1954
	year = 1955
	if j.TrackStillApplies() {
		t.Error("1955 is past Hop to the Bop")
	}
	cfg.Playlist = PlaylistAll
	if !j.TrackStillApplies() {
		t.Error("all music")
	}
	cfg.Playlist = PlaylistCustom
	if !j.TrackStillApplies() {
		t.Error("custom, selected")
	}
	cfg.CustomJukebox[20] = false
	if j.TrackStillApplies() {
		t.Error("custom, deselected")
	}
}

// makeAllMusicPlaylist's orders: by years the start year, then the end year.
func TestSortYears(t *testing.T) {
	asc := AllMusicPlaylist(SortYearsAscending)
	if want := []MusicID{13, 12, 15}; !reflect.DeepEqual(asc[:3], want) || asc[29] != 29 {
		t.Errorf("years ascending: %v", asc)
	}
	desc := AllMusicPlaylist(SortYearsDescending)
	if want := []MusicID{29, 10, 9, 8}; !reflect.DeepEqual(desc[:4], want) || desc[29] != 13 {
		t.Errorf("years descending: %v", desc)
	}
}
