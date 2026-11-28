package main

import "testing"

// TestTooltipOwner steps the tooltip's input handling (MouseInput.cpp
// stateNormalHover / stateNormalLeft / the widget release) with a fake clock.
func TestTooltipOwner(t *testing.T) {
	brush := tipKey{widget: 5}
	swatch := tipKey{widget: 6}
	const brushTip, swatchTip = "Repaint", "Select main colour"
	var s tooltipState
	s.notShown = 20000 // long without a tooltip (and not about to wrap: uint16)
	rest := func(ms int, k tipKey, text string, x, y int) {
		for i := 0; i < ms/20; i++ {
			s.step(20, k, text, x, y, tipInput{})
		}
	}
	rest(1900, brush, brushTip, 10, 10)
	if s.showing {
		t.Fatal("open before 2 s at rest")
	}
	rest(200, brush, brushTip, 10, 10)
	if !s.showing || s.x != 10 || s.y != 10 {
		t.Fatalf("not open at 10,10 after 2 s: %+v", s)
	}
	rest(100, brush, brushTip, 13, 12) // moved inside the widget: stays, where it opened
	if !s.showing || s.x != 10 || s.y != 10 {
		t.Fatalf("moved inside the widget: %+v", s)
	}
	// a click on it: closed, and not again while the mouse stays on it
	s.step(20, brush, brushTip, 13, 12, tipInput{press: true, held: true})
	s.step(20, brush, brushTip, 13, 12, tipInput{held: true})
	s.step(20, brush, brushTip, 13, 12, tipInput{release: true})
	rest(5000, brush, brushTip, 13, 12)
	rest(3000, brush, brushTip, 11, 12)
	if s.showing {
		t.Fatal("the clicked widget's tooltip came back")
	}
	// another widget: closeAndReset, then (the last tooltip closed 8 s ago)
	// after 2 s at rest
	rest(1000, swatch, swatchTip, 30, 12)
	if s.showing {
		t.Fatal("the swatch's tooltip before 2 s")
	}
	rest(1100, swatch, swatchTip, 30, 12)
	if !s.showing || s.text != swatchTip {
		t.Fatalf("the swatch's tooltip: %+v", s)
	}
	// back on the brush within the second: closeAndReset, then at once
	s.step(20, brush, brushTip, 10, 10, tipInput{})
	s.step(20, brush, brushTip, 10, 10, tipInput{})
	if !s.showing || s.text != brushTip {
		t.Fatalf("quick open on the brush: %+v", s)
	}
	// 8 s: closed, and not again on the same widget
	rest(8100, brush, brushTip, 10, 10)
	if s.showing {
		t.Fatal("open past 8 s")
	}
	rest(4000, brush, brushTip, 10, 10)
	if s.showing {
		t.Fatal("reopened on the same widget after its 8 s")
	}
}

// TestTooltipCarRows: the same text on two car rows is two widgets (the
// hovered car resets the tooltip, Vehicle.cpp scrollMouseOver).
func TestTooltipCarRows(t *testing.T) {
	var s tooltipState
	s.notShown = 20000
	row1 := tipKey{widget: vehicleRowKey(1)}
	row2 := tipKey{widget: vehicleRowKey(2)}
	for i := 0; i < 110; i++ {
		s.step(20, row1, "Wagon", 5, 40, tipInput{})
	}
	if !s.showing || s.y != 40 {
		t.Fatalf("row 1's tooltip: %+v", s)
	}
	s.step(20, row2, "Wagon", 5, 50, tipInput{})
	s.step(20, row2, "Wagon", 5, 50, tipInput{})
	if !s.showing || s.y != 50 {
		t.Fatalf("row 2's tooltip (reopened at once, at row 2): %+v", s)
	}
	// a press on a row (scrollLeft) leaves no owner: its tooltip comes back
	s.step(20, row2, "Wagon", 5, 50, tipInput{press: true, held: true, resetPress: true})
	s.step(20, row2, "Wagon", 5, 50, tipInput{release: true})
	for i := 0; i < 110; i++ {
		s.step(20, row2, "Wagon", 5, 50, tipInput{})
	}
	if !s.showing {
		t.Fatal("no tooltip after a row press and 2 s at rest")
	}
}

func vehicleRowKey(row int) int { return 1000 + row }
