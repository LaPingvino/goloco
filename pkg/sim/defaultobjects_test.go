package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// TestLoadResetsDefaultTreeWall: every load resets the Terraform window's
// default tree and wall to none, whatever the file holds (S5.cpp
// importSaveToGameState → Terraform::resetDefaultObjectIds); the save writes
// the game's current ones (s5export).
func TestLoadResetsDefaultTreeWall(t *testing.T) {
	sc := &scenario.Scenario{MapWidth: 4, MapHeight: 4, DefaultTreeObjectId: 5, DefaultWallObjectId: 2}
	m := &Map{}
	m.LoadScenario(sc)
	if m.DefaultTreeObjectId != 0xFF || m.DefaultWallObjectId != 0xFF {
		t.Fatalf("after load %d, %d; want 0xFF, 0xFF", m.DefaultTreeObjectId, m.DefaultWallObjectId)
	}
}
