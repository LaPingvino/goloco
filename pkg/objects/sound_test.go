package objects

import (
	"os"
	"path/filepath"
	"testing"
)

// Every sound object in the game data loads with a PCM sample whose length is
// whole frames of its WAVEFORMATEX (SoundObject::load).
func TestSoundObjectsLoad(t *testing.T) {
	dir := filepath.Join("..", "..", "..", "locomotion", "ObjData")
	files, _ := filepath.Glob(filepath.Join(dir, "SND*.DAT"))
	if len(files) == 0 {
		t.Skip("no game data")
	}
	m := NewObjectManager(dir)
	n := 0
	for _, f := range files {
		if _, err := os.Stat(f); err != nil {
			continue
		}
		lo, err := m.LoadObject(f)
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		if lo.Header.GetType() != ObjectTypeSound {
			continue
		}
		so, ok := lo.Object.(*SoundObject)
		if !ok {
			t.Fatalf("%s: not parsed as a sound", f)
		}
		w := so.Format
		if w.FormatTag != 1 || w.Channels == 0 || w.BlockAlign == 0 || len(so.PCM) == 0 || len(so.PCM)%int(w.BlockAlign) != 0 ||
			w.AvgBytesPerSec != w.SamplesPerSec*uint32(w.BlockAlign) {
			t.Errorf("%s (%s): bad sample %+v, %d bytes", filepath.Base(f), so.Name, w, len(so.PCM))
		}
		if n < 2 {
			t.Logf("%s %q loop=%v vol=%d %+v %d bytes", filepath.Base(f), so.Name, so.ShouldLoop, so.Volume, w, len(so.PCM))
		}
		n++
	}
	t.Logf("%d sound objects", n)
	if n == 0 {
		t.Fatal("no sound objects parsed")
	}
}
