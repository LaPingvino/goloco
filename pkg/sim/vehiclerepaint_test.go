package sim

import "testing"

func TestVehicleRepaint(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	var tr Train
	found := false
	for _, x := range trains {
		if len(x.Cars) >= 2 && len(x.Cars[0]) >= 1 {
			tr, found = x, true
			break
		}
	}
	if !found {
		t.Skip("no train with two cars")
	}
	head := m.Entities[tr.Head]
	c := &Cmd{Company: head.Owner, Objects: om}
	body := tr.Cars[0][0].Body
	front := tr.Cars[0][0].Front
	other := tr.Cars[1][0].Body
	red := ColourScheme{Primary: 7, Secondary: 9}
	a := VehicleRepaint{Vehicle: m.Entities[body].ID, Colours: [4]ColourScheme{red, red, red}, Flags: RepaintBody}
	before := m.Entities[body].ColourPrimary

	if _, ok := m.VehicleRepaint(c, 0, a); !ok || m.Entities[body].ColourPrimary != before {
		t.Fatal("the dry run painted")
	}
	// Only the body of that component.
	m.VehicleRepaint(c, CmdApply, a)
	if m.Entities[body].ColourPrimary != 7 || m.Entities[body].ColourSecondary != 9 {
		t.Errorf("body %d/%d, want 7/9", m.Entities[body].ColourPrimary, m.Entities[body].ColourSecondary)
	}
	if m.Entities[front].ColourPrimary == 7 && before != 7 {
		t.Error("the front bogie was painted with only the body flag")
	}
	if m.Entities[other].ColourPrimary == 7 && before != 7 {
		t.Error("another car was painted")
	}
	// The entire train, all parts.
	blue := ColourScheme{Primary: 3, Secondary: 4}
	m.VehicleRepaint(c, CmdApply, VehicleRepaint{Vehicle: a.Vehicle, Colours: [4]ColourScheme{blue, blue, blue}, Flags: RepaintBody | RepaintFrontBogie | RepaintBackBogie | RepaintEntireTrain})
	for _, car := range tr.Cars {
		for _, comp := range car {
			for _, i := range []int{comp.Front, comp.Back, comp.Body} {
				if m.Entities[i].ColourPrimary != 3 {
					t.Fatalf("entity %d not painted with the train", m.Entities[i].ID)
				}
			}
		}
	}
	// Another company's vehicle is refused.
	if _, ok := m.VehicleRepaint(&Cmd{Company: (head.Owner + 1) % 8}, CmdApply, a); ok && head.Owner != CompanyNeutral {
		t.Error("another company repainted it")
	}
}
