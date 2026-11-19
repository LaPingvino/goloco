package ui

import "testing"

// TextInputWindow.cpp: the rename box is an ordinary window of the stack
// (stickToFront), one at a time, tied to the window that opened it.
func TestTextInputOnStack(t *testing.T) {
	s := &WindowStack{ScreenW: 800, ScreenH: 600}
	caller := &SimpleWindow{WinBase: WinBase{WinType: WindowTypeStation, Number: 7, W: 100, H: 50}}
	s.Open(caller)
	var got []string
	ti := OpenTextInput(s, WindowTypeStation, 7, "T", "M", "Dallas", 31, 0, 0, true, func(text string) { got = append(got, text) })
	if s.Find(WindowTypeTextInput, AnyNumber) != ti || !ti.HasFlags(WindowFlagStickToFront) {
		t.Fatal("the text input should be a stickToFront window of the stack")
	}
	if ti.X != 235 || ti.Y != 255 { // createWindowCentred, as OpenLoco's on an 800×600 screen
		t.Errorf("at %d,%d, want 235,255", ti.X, ti.Y)
	}
	// Focused at open: keys are the box's.
	if !ti.KeyUp(false, []rune("x"), nil) || string(ti.Session.Buffer) != "Dallasx" {
		t.Errorf("focused: %q", string(ti.Session.Buffer))
	}
	// Unfocused: the key goes on (false), Return still is OK.
	ti.Focused = false
	if ti.KeyUp(false, []rune("y"), nil) || string(ti.Session.Buffer) != "Dallasx" {
		t.Error("an unfocused text input took a key")
	}
	if !ti.KeyUp(true, nil, nil) || len(got) != 1 || got[0] != "Dallasx" || s.Find(WindowTypeTextInput, AnyNumber) != nil {
		t.Errorf("Return: %v, open %v", got, s.Find(WindowTypeTextInput, AnyNumber) != nil)
	}

	// A new one replaces the open one (openTextInput: cancel()).
	a := OpenTextInput(s, WindowTypeStation, 7, "T", "M", "a", 31, 0, 0, true, nil)
	b := OpenTextInput(s, WindowTypeStation, 7, "T", "M", "b", 31, 0, 0, true, nil)
	if !a.Closed || s.Find(WindowTypeTextInput, AnyNumber) != b {
		t.Error("a second text input should close the first")
	}
	// The caller closing takes it along (sub_4CE6C9).
	s.Close(caller)
	if s.Find(WindowTypeTextInput, AnyNumber) != nil {
		t.Error("the text input outlived its caller's close")
	}

	// OK with the caller gone (closed without its onClose reaching the box):
	// no textInput call; sub_4CE6FF cancels such a box.
	got = nil
	c := OpenTextInput(s, WindowTypeTown, 3, "T", "M", "x", 31, 0, 0, true, func(text string) { got = append(got, text) })
	c.ok()
	if len(got) != 0 {
		t.Error("OK reached a caller that is not open")
	}
	OpenTextInput(s, WindowTypeTown, 3, "T", "M", "x", 31, 0, 0, true, nil)
	CloseTextInputWithoutCaller(s)
	if s.Find(WindowTypeTextInput, AnyNumber) != nil {
		t.Error("sub_4CE6FF should cancel a text input without its caller")
	}
}
