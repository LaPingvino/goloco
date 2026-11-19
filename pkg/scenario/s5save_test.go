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

// TestDefaultTreeWall_roundTrip: GameState defaultTreeObjectId (0x199) and
// defaultWallObjectId (0x1B2) are written and read back (GameState.h).
func TestDefaultTreeWall_roundTrip(t *testing.T) {
	sc := &Scenario{DefaultTreeObjectId: 7, DefaultWallObjectId: 3}
	gs := sc.EncodeGameState()
	if gs[0x199] != 7 || gs[0x1B2] != 3 {
		t.Fatalf("written %d, %d at 0x199, 0x1B2; want 7, 3", gs[0x199], gs[0x1B2])
	}
	var back Scenario
	parseGeneralState(&back, gs)
	if back.DefaultTreeObjectId != 7 || back.DefaultWallObjectId != 3 {
		t.Fatalf("read back %d, %d; want 7, 3", back.DefaultTreeObjectId, back.DefaultWallObjectId)
	}
}
