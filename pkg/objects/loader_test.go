package objects

import "testing"

func TestGetLandObjectByIndex_valid(t *testing.T) {
	mgr := &ObjectManager{
		LandObjects: []*LandObject{
			{Name: "GRASS1"},
			{Name: "SAND1"},
			{Name: "ROCK1"},
		},
	}

	cases := []struct {
		index int
		want  string
	}{
		{0, "GRASS1"},
		{1, "SAND1"},
		{2, "ROCK1"},
	}
	for _, c := range cases {
		land := mgr.GetLandObjectByIndex(c.index)
		if land == nil {
			t.Fatalf("GetLandObjectByIndex(%d) = nil, want %s", c.index, c.want)
		}
		if land.Name != c.want {
			t.Errorf("GetLandObjectByIndex(%d).Name = %q, want %q", c.index, land.Name, c.want)
		}
	}
}

func TestGetLandObjectByIndex_out_of_range(t *testing.T) {
	mgr := &ObjectManager{
		LandObjects: []*LandObject{
			{Name: "GRASS1"},
		},
	}

	cases := []int{-1, 1, 100}
	for _, idx := range cases {
		if got := mgr.GetLandObjectByIndex(idx); got != nil {
			t.Errorf("GetLandObjectByIndex(%d) = %v, want nil", idx, got)
		}
	}
}

func TestGetLandObjectByIndex_empty(t *testing.T) {
	mgr := &ObjectManager{LandObjects: nil}
	if got := mgr.GetLandObjectByIndex(0); got != nil {
		t.Errorf("GetLandObjectByIndex(0) on empty = %v, want nil", got)
	}
}
