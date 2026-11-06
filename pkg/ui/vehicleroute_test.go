package ui

import "testing"

// The Route tab's buttons as Route::onMouseUp: each order button and
// Local / Express give their action; another company's vehicle has the order
// buttons hidden and Local / Express disabled (prepareDraw), so those clicks
// do nothing; Local / Express split the list's width.
func TestVehicleRouteClicks(t *testing.T) {
	for _, own := range []bool{true, false} {
		var got []string
		w := &VehicleWindow{X: 10, Y: 20, W: 265, H: VehicleRouteMinHeight, Tab: 4, Main: VehicleMainState{Own: own}, RouteSel: -1,
			OnRoute: func(a string) { got = append(got, a) }}
		lw := VehicleRouteLayout(w.W, w.H, 0, !own)
		click := func(i int) {
			c := lw[i]
			w.HandleClick(w.X+(c.L+c.R)/2, w.Y+(c.T+c.B)/2)
		}
		for _, i := range []int{vrDelete, vrUp, vrDown, vrSkip, vrReverse, vrLocalMode, vrExpressMode} {
			click(i)
		}
		want := []string{"delete", "up", "down", "skip", "reverse", "local", "express"}
		if !own {
			want = nil
		}
		if len(got) != len(want) {
			t.Fatalf("own %v: actions %q, want %q", own, got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("own %v: action %d %q, want %q", own, i, got[i], want[i])
			}
		}
		list := lw[vrRouteList]
		if lw[vrExpressMode].R != list.R || lw[vrLocalMode].L != 3 || lw[vrLocalMode].R+1 != lw[vrExpressMode].L {
			t.Errorf("own %v: local %v express %v list right %d", own, lw[vrLocalMode], lw[vrExpressMode], list.R)
		}
	}
}
