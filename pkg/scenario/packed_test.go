package scenario

import (
	"encoding/binary"
	"testing"
)

func packedHeader(flags uint32, name string) PackedObjectRaw {
	var po PackedObjectRaw
	binary.LittleEndian.PutUint32(po.HeaderBytes[0:], flags)
	copy(po.HeaderBytes[4:12], name)
	binary.LittleEndian.PutUint32(po.HeaderBytes[12:], 0x12345678)
	return po
}

// TestPackedObjectsToSave: upstream a7aa65e (#4033)'s filter as goloco
// applies it — vanilla objects (by name and type, whatever the source bits)
// and empty headers dropped, custom and OpenLoco-source objects kept.
func TestPackedObjectsToSave(t *testing.T) {
	custom := packedHeader(23, "MYLOCO  ")     // vehicle, source custom
	misflagged := packedHeader(23, "4F      ") // vanilla name and type, source custom
	og := packedHeader(17|3<<6, "OG_TRKST")    // track, source OpenLoco
	wrongType := packedHeader(17, "4F      ")  // a vanilla name with another type
	var empty PackedObjectRaw
	for i := range empty.HeaderBytes {
		empty.HeaderBytes[i] = 0xFF
	}
	got := PackedObjectsToSave([]PackedObjectRaw{custom, misflagged, og, wrongType, empty})
	want := []PackedObjectRaw{custom, og, wrongType}
	if len(got) != len(want) {
		t.Fatalf("%d packed, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].HeaderBytes != want[i].HeaderBytes {
			t.Errorf("packed %d: %q, want %q", i, got[i].HeaderBytes[4:12], want[i].HeaderBytes[4:12])
		}
	}
}
