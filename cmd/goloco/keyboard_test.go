package main

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/shortcuts"
	"github.com/hajimehoshi/ebiten/v2"
)

// Every shortcut of Shortcuts.cpp's table has its action.
func TestShortcutActions(t *testing.T) {
	for _, def := range shortcuts.Defs {
		if shortcutActions[def.ID] == nil {
			t.Errorf("%s has no action", def.ConfigName)
		}
	}
}

// Every default binding can be pressed: its key is one of Ebitengine's.
func TestDefaultBindingsHaveKeys(t *testing.T) {
	b := shortcuts.Load(nil)
	for _, def := range shortcuts.Defs {
		k := b.Get(def.ID)
		if k == shortcuts.Unbound {
			continue
		}
		if _, ok := keyCodeEbiten[k.KeyCode]; !ok {
			t.Errorf("%s: no Ebitengine key for %q", def.ConfigName, def.DefaultBinding)
		}
	}
	for key, code := range map[ebiten.Key]uint32{ebiten.KeyZ: 'z', ebiten.KeyDigit0: '0', ebiten.KeyDigit1: '1', ebiten.KeyEnter: shortcuts.KeyReturn,
		ebiten.KeyNumpadEnter: shortcuts.KeyKPEnter, ebiten.KeyArrowUp: shortcuts.KeyUp, ebiten.KeyInsert: shortcuts.KeyInsert,
		ebiten.KeyControlLeft: shortcuts.KeyLCtrl, ebiten.KeyAltRight: shortcuts.KeyRAlt, ebiten.KeySpace: shortcuts.KeySpace} {
		if ebitenKeyCodes[key] != code {
			t.Errorf("%v is key %d, want %d", key, ebitenKeyCodes[key], code)
		}
	}
}

func TestScriptKey(t *testing.T) {
	for _, c := range []struct {
		name string
		ev   keyEvent
		held []uint32
	}{
		{"Z", keyEvent{'z', 'z'}, nil},
		{"1", keyEvent{'1', '1'}, nil},
		{"ESC", keyEvent{shortcuts.KeyEscape, 0}, nil},
		{"escape", keyEvent{shortcuts.KeyEscape, 0}, nil},
		{"ENTER", keyEvent{shortcuts.KeyReturn, '\r'}, nil},
		{"Backspace", keyEvent{shortcuts.KeyBackspace, 8}, nil},
		{"SPACE", keyEvent{shortcuts.KeySpace, ' '}, nil},
		{"PageUp", keyEvent{shortcuts.KeyFromName("PageUp"), 0}, nil},
		{"Left Ctrl+S", keyEvent{'s', 0}, []uint32{shortcuts.KeyLCtrl}},
		{"Left Shift+Backspace", keyEvent{shortcuts.KeyBackspace, 8}, []uint32{shortcuts.KeyLShift}},
		{"Left Shift+a", keyEvent{'a', 'A'}, []uint32{shortcuts.KeyLShift}},
		{"+", keyEvent{'+', '+'}, nil},
	} {
		ev, held, ok := scriptKey(c.name)
		if !ok || ev != c.ev || len(held) != len(c.held) || len(held) > 0 && held[0] != c.held[0] {
			t.Errorf("scriptKey(%q) = %+v %v %v, want %+v %v", c.name, ev, held, ok, c.ev, c.held)
		}
	}
	if _, _, ok := scriptKey("No such key"); ok {
		t.Error("an unknown key was taken")
	}
}

// A typed character goes to the key press it came with; a character without
// one is its key pressed again (the key repeat).
func TestEnqueueText(t *testing.T) {
	g := &Game{}
	g.enqueueKey(shortcuts.KeyLShift)
	g.enqueueKey('s')
	g.enqueueKey('a')
	g.enqueueKey(shortcuts.KeyReturn)
	g.enqueueText(0, 'A')
	g.enqueueText(0, 's')
	g.enqueueText(0, 'x')
	want := []keyEvent{{shortcuts.KeyLShift, 0}, {'s', 's'}, {'a', 'A'}, {shortcuts.KeyReturn, '\r'}, {'x', 'x'}}
	if len(g.kb.queue) != len(want) {
		t.Fatalf("queue %+v", g.kb.queue)
	}
	for i, e := range want {
		if g.kb.queue[i] != e {
			t.Errorf("queue[%d] = %+v, want %+v", i, g.kb.queue[i], e)
		}
	}
}
