package ui

import "testing"

// Details::prepareDraw at the default 265×177: the route tab moves next to
// finances (leftAlignTabs), the right bar's buttons, the list above three
// text lines; another company's vehicle: no buttons, the list to the edge;
// the pickup button stays for an own placed vehicle (upstream un-hides it).
func TestVehicleDetailsLayout(t *testing.T) {
	w, _ := VehicleDetailsLayout(265, 177, VehicleDetailsState{Own: true, Placed: true})
	want := map[int]BVWidget{
		vwTabRoute:   {127, 15, 157, 41, false},
		vwClose:      {250, 2, 262, 14, false},
		vdCarList:    {3, 44, 239, 139, false},
		vdBuildNew:   {240, 44, 263, 67, false},
		vdPickup:     {240, 68, 263, 91, false},
		vdPaintBrush: {240, 115, 263, 139, false},
	}
	for i, x := range want {
		if w[i] != x {
			t.Errorf("widget %d: %+v, want %+v", i, w[i], x)
		}
	}
	if !w[vdPaintPrimary].Hidden {
		t.Error("paint colours shown without the paint tool")
	}
	o, _ := VehicleDetailsLayout(300, 200, VehicleDetailsState{})
	if !o[vdBuildNew].Hidden || !o[vdRemove].Hidden || !o[vdPaintBrush].Hidden || o[vdCarList].R != 296 || o[vdCarList].B != 162 {
		t.Errorf("another company's: list %+v, buttons hidden %v %v %v", o[vdCarList], o[vdBuildNew].Hidden, o[vdRemove].Hidden, o[vdPaintBrush].Hidden)
	}
	p, _ := VehicleDetailsLayout(265, 182, VehicleDetailsState{Own: true, PaintTool: true})
	if p[vdPaintPrimary] != (BVWidget{226, 149, 242, 165, false}) || p[vdPaintSecondary] != (BVWidget{244, 149, 260, 165, false}) {
		t.Errorf("paint colours %+v %+v", p[vdPaintPrimary], p[vdPaintSecondary])
	}
}
