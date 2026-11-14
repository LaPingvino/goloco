package scenario

import (
	"path/filepath"
	"testing"

	"github.com/LaPingvino/goloco/pkg/assets"
)

// TestEncodeGameState_identity: a loaded game's state written back unchanged
// is the file's own, byte for byte, but for upstream's save normalisations
// (normalised here on the file's bytes too).
func TestEncodeGameState_identity(t *testing.T) {
	n := 0
	for _, f := range gameFiles(t) {
		sc, err := LoadScenarioData(f)
		if err != nil || sc.Raw == nil || sc.Header.Flags&(assets.HeaderFlagIsRaw|assets.HeaderFlagIsDump) != 0 {
			continue
		}
		n++
		for _, d := range sc.DiffFromFile(sc.EncodeGameState(), nil, 40, false) {
			t.Errorf("%s: %s", filepath.Base(f), d)
		}
	}
	if n == 0 {
		t.Skip("no game data")
	}
}
