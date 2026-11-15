package audio

import (
	"path/filepath"
	"testing"
)

// css1.dat holds the 32 sounds of Audio.h SoundId (clickDown 0 … breakdown6 31).
func TestReadCSS1(t *testing.T) {
	p := filepath.Join("..", "..", "..", "locomotion", "Data", "css1.dat")
	s, err := ReadCSS(p)
	if err != nil {
		t.Skipf("no game data: %v", err)
	}
	if len(s) != 32 {
		t.Fatalf("%d sounds, want 32 (SoundId clickDown…breakdown6)", len(s))
	}
	for i, x := range s {
		w := x.Format
		if w.FormatTag != 1 || w.BlockAlign == 0 || len(x.PCM)%int(w.BlockAlign) != 0 {
			t.Errorf("sound %d: %+v, %d bytes", i, w, len(x.PCM))
		}
	}
	t.Logf("%d sounds; 0: %+v %d bytes", len(s), s[0].Format, len(s[0].PCM))
}
