package main

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/cursor"
	"github.com/LaPingvino/goloco/pkg/world"
)

// The construction window's tabs' tool cursors (ConstructionTab.cpp,
// StationTab.cpp, SignalTab.cpp, OverheadTab.cpp onMouseDown toolSet).
func TestConstructionToolCursors(t *testing.T) {
	g := &Game{buildMode: buildModeTrack, w: &world.World{}}
	for tab, want := range []cursor.Id{cursor.Crosshair, cursor.PlaceStation, cursor.PlaceSignal, cursor.Crosshair} {
		g.consTab = tab
		if tab == 0 { // hovering (no head yet, or construct pressed): crosshair; copy / paste too
			g.w.Cons.Hover = true
			for _, tool := range []int{consToolConstruct, consToolCopy, consToolPaste} {
				g.consTool = tool
				if id, ok := g.toolCursor(); !ok || id != want {
					t.Errorf("tab 0 tool %d: %v %v", tool, id, ok)
				}
			}
			continue
		}
		if id, ok := g.toolCursor(); !ok || id != want {
			t.Errorf("tab %d: %v %v, want %v", tab, id, ok, want)
		}
	}
	g.consToolOff, g.consTab = true, 2 // Escape on the Signal tab: toolCancel
	if _, ok := g.toolCursor(); ok {
		t.Error("a cancelled tool still has its cursor")
	}
}
