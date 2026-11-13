package world

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/sim"
)

// The arrow the renderer gets is sim's while the construction tool is up,
// and none otherwise.
func TestSyncConstructionArrow(t *testing.T) {
	wld := newTestWorld(20, 20, 10)
	wld.ConsArrow = sim.ConstructionArrow{On: true, X: 320, Y: 320, Z: 40, Direction: 13}
	wld.SyncConstructionArrow(true)
	if wld.consArrow != wld.ConsArrow {
		t.Fatalf("arrow %+v, want %+v", wld.consArrow, wld.ConsArrow)
	}
	wld.SyncConstructionArrow(false)
	if wld.consArrow.On {
		t.Fatalf("off: %+v", wld.consArrow)
	}
}
