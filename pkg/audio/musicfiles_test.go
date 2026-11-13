package audio

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaPingvino/goloco/pkg/jukebox"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

// Every track of the jukebox is a file of the Locomotion Data folder, under
// the name OpenLoco asks for (whatever its case on disk), and decodes as the
// WAV LoadAndPlayMusic plays. Runs with the game's data (GOLOCO_DATA, or the
// locomotion folder beside the repository).
func TestJukeboxTracksDecode(t *testing.T) {
	dir := ""
	for _, d := range []string{filepath.Join(os.Getenv("GOLOCO_DATA"), "Data"), os.Getenv("GOLOCO_DATA"),
		filepath.Join("..", "..", "..", "locomotion", "Data"), filepath.Join(os.Getenv("HOME"), "Projects", "locomotion", "Data")} {
		if _, err := FindFile(filepath.Join(d, "CSS5.DAT")); err == nil {
			dir = d
			break
		}
	}
	if dir == "" {
		t.Skip("no Locomotion Data folder")
	}
	for i := range jukebox.NumMusicTracks {
		mi := jukebox.Info(jukebox.MusicID(i))
		p, err := FindFile(filepath.Join(dir, mi.File))
		if err != nil {
			t.Errorf("track %d: %v", i, err)
			continue
		}
		f, err := os.Open(p)
		if err != nil {
			t.Errorf("track %d: %v", i, err)
			continue
		}
		s, err := wav.DecodeWithSampleRate(SampleRate, f)
		if err != nil || s.Length() <= 0 {
			t.Errorf("track %d (%s): decode %v", i, p, err)
		}
		f.Close()
	}
}
