package main

import (
	"strings"

	"github.com/LaPingvino/goloco/pkg/shortcuts"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// keyboard.go — every key the game takes goes through here: the key queue,
// the modifiers, the arrow keys' and the screen edge's scrolling, the cheat
// words, and normalKey's order — the windows that take keys (front to back),
// then the shortcuts by their bindings (pkg/shortcuts; the actions are in
// shortcut_actions.go). Nothing else reads the keyboard for a key press.
//
// OpenLoco reference: src/OpenLoco/src/Input/Keyboard.cpp (handleKeyboard
// 0x004BE92A, processKeyboardInput / cheat 0x004BEC5B, normalKey 0x004BEDA0,
// edgeScroll, keyScroll, enqueueKey / handleKeyInput 0x00406FBA, enqueueText,
// readKeyboardState 0x0040477F), src/OpenLoco/src/Input.cpp (the SDL events
// that feed them).
//
// Keys are SDL keycodes, as upstream's (shortcuts.KeyFromName): Ebitengine's
// keys are places on the keyboard, read as SDL's default layout has them.
// PARTIAL: the tutorial's playback (loc_4BEFEF) and the intro's keys — goloco
// has neither; Alt+Return's full screen toggle (Input.cpp).

// keyEvent is Keyboard.cpp's Key: a key press and the character it typed
// (0: none).
type keyEvent struct {
	keyCode  uint32
	charCode rune
}

// keyboardState is Keyboard.cpp's state.
type keyboardState struct {
	modifier    shortcuts.KeyModifier // _keyModifier
	cheatBuffer string                // _cheatBuffer (0x0011364A5)
	queue       []keyEvent            // _keyQueue
	bindings    *shortcuts.Bindings   // Shortcuts.cpp's _bindings
	// _screenshotCountdown (Ui/Screenshot.cpp, 0x00508F16) and the shot it
	// asks the next Draw for.
	screenshotCountdown int
	screenshotNow       bool
}

// ebitenScancodes: each Ebitengine key's SDL scancode (the USB usage id of
// its place on the keyboard).
var ebitenScancodes = map[ebiten.Key]uint16{
	ebiten.KeyEnter: 40, ebiten.KeyEscape: 41, ebiten.KeyBackspace: 42, ebiten.KeyTab: 43, ebiten.KeySpace: 44,
	ebiten.KeyMinus: 45, ebiten.KeyEqual: 46, ebiten.KeyBracketLeft: 47, ebiten.KeyBracketRight: 48, ebiten.KeyBackslash: 49,
	ebiten.KeySemicolon: 51, ebiten.KeyQuote: 52, ebiten.KeyBackquote: 53, ebiten.KeyComma: 54, ebiten.KeyPeriod: 55, ebiten.KeySlash: 56,
	ebiten.KeyCapsLock: 57,
	ebiten.KeyF1:       58, ebiten.KeyF2: 59, ebiten.KeyF3: 60, ebiten.KeyF4: 61, ebiten.KeyF5: 62, ebiten.KeyF6: 63,
	ebiten.KeyF7: 64, ebiten.KeyF8: 65, ebiten.KeyF9: 66, ebiten.KeyF10: 67, ebiten.KeyF11: 68, ebiten.KeyF12: 69,
	ebiten.KeyPrintScreen: 70, ebiten.KeyScrollLock: 71, ebiten.KeyPause: 72, ebiten.KeyInsert: 73, ebiten.KeyHome: 74,
	ebiten.KeyPageUp: 75, ebiten.KeyDelete: 76, ebiten.KeyEnd: 77, ebiten.KeyPageDown: 78,
	ebiten.KeyArrowRight: 79, ebiten.KeyArrowLeft: 80, ebiten.KeyArrowDown: 81, ebiten.KeyArrowUp: 82,
	ebiten.KeyNumLock: 83, ebiten.KeyNumpadDivide: 84, ebiten.KeyNumpadMultiply: 85, ebiten.KeyNumpadSubtract: 86,
	ebiten.KeyNumpadAdd: 87, ebiten.KeyNumpadEnter: 88,
	ebiten.KeyNumpad1: 89, ebiten.KeyNumpad2: 90, ebiten.KeyNumpad3: 91, ebiten.KeyNumpad4: 92, ebiten.KeyNumpad5: 93,
	ebiten.KeyNumpad6: 94, ebiten.KeyNumpad7: 95, ebiten.KeyNumpad8: 96, ebiten.KeyNumpad9: 97, ebiten.KeyNumpad0: 98,
	ebiten.KeyNumpadDecimal: 99, ebiten.KeyContextMenu: 101, ebiten.KeyNumpadEqual: 103,
	ebiten.KeyF13: 104, ebiten.KeyF14: 105, ebiten.KeyF15: 106, ebiten.KeyF16: 107, ebiten.KeyF17: 108, ebiten.KeyF18: 109,
	ebiten.KeyF19: 110, ebiten.KeyF20: 111, ebiten.KeyF21: 112, ebiten.KeyF22: 113, ebiten.KeyF23: 114, ebiten.KeyF24: 115,
	ebiten.KeyControlLeft: 224, ebiten.KeyShiftLeft: 225, ebiten.KeyAltLeft: 226, ebiten.KeyMetaLeft: 227,
	ebiten.KeyControlRight: 228, ebiten.KeyShiftRight: 229, ebiten.KeyAltRight: 230, ebiten.KeyMetaRight: 231,
}

// ebitenKeyCodes / keyCodeEbiten: Ebitengine's keys as SDL keycodes, and back.
var ebitenKeyCodes, keyCodeEbiten = func() (map[ebiten.Key]uint32, map[uint32]ebiten.Key) {
	sc := map[ebiten.Key]uint16{}
	for k, v := range ebitenScancodes {
		sc[k] = v
	}
	for i := 0; i < 26; i++ {
		sc[ebiten.KeyA+ebiten.Key(i)] = uint16(4 + i)
	}
	for i := 1; i <= 9; i++ {
		sc[ebiten.KeyDigit0+ebiten.Key(i)] = uint16(29 + i)
	}
	sc[ebiten.KeyDigit0] = 39
	codes, keys := map[ebiten.Key]uint32{}, map[uint32]ebiten.Key{}
	for k, s := range sc {
		if c := shortcuts.KeyFromScancode(s); c != 0 {
			codes[k], keys[c] = c, k
		}
	}
	return codes, keys
}()

// isPrintableKey: a key that types a character (SDL sends its text with the
// key press, and the system's key repeat brings both again).
func isPrintableKey(code uint32) bool { return code >= 0x20 && code < 0x7F }

// isKeypadTextKey: a keypad key that types a character with NumLock on
// (Keypad / … Keypad ., Keypad =; not Keypad Enter).
func isKeypadTextKey(code uint32) bool {
	const kpDivide, kpEnter, kpPeriod, kpEquals = 1<<30 | 84, 1<<30 | 88, 1<<30 | 99, 1<<30 | 103
	return code >= kpDivide && code <= kpPeriod && code != kpEnter || code == kpEquals
}

func isModifierKey(code uint32) bool {
	switch code {
	case shortcuts.KeyLShift, shortcuts.KeyRShift, shortcuts.KeyLCtrl, shortcuts.KeyRCtrl, shortcuts.KeyLAlt, shortcuts.KeyRAlt:
		return true
	}
	return false
}

// keyHeld is _keyboardState (readKeyboardState): the key is down — on the
// keyboard, or held by a script.
func (g *Game) keyHeld(code uint32) bool {
	if g.script != nil && (g.script.keysHeld[code] || g.script.tapHeld[code]) {
		return true
	}
	k, ok := keyCodeEbiten[code]
	return ok && ebiten.IsKeyPressed(k)
}

// enqueueKey is handleKeyInput: the key on the queue, Return, Backspace and
// Delete with themselves as their character.
func (g *Game) enqueueKey(code uint32) {
	e := keyEvent{keyCode: code}
	switch code {
	case shortcuts.KeyReturn, shortcuts.KeyBackspace, shortcuts.KeyDelete:
		e.charCode = rune(code)
	}
	g.kb.queue = append(g.kb.queue, e)
}

// readKeys is Input.cpp's key events of this frame onto the queue: each key
// pressed (SDL_EVENT_KEY_DOWN → handleKeyInput) and each character typed
// (SDL_EVENT_TEXT_INPUT → enqueueText: the character of the key just
// queued), then a script's.
//
// SDL's key-down comes again while a key is held (the system's key repeat),
// and upstream takes every one. Ebitengine reports a press once, and the
// repeat only as typed characters: a repeated character is its key pressed
// again; the other keys repeat here, after half a second, twenty times a
// second.
func (g *Game) readKeys() {
	first := len(g.kb.queue)
	for _, k := range inpututil.AppendJustPressedKeys(nil) {
		if code, ok := ebitenKeyCodes[k]; ok {
			g.enqueueKey(code)
		}
	}
	for k, code := range ebitenKeyCodes {
		if d := inpututil.KeyPressDuration(k); !isPrintableKey(code) && !isKeypadTextKey(code) && !isModifierKey(code) && d > 30 && d%3 == 0 {
			g.enqueueKey(code)
		}
	}
	for _, r := range ebiten.AppendInputChars(nil) {
		g.enqueueText(first, r)
	}
	if g.script != nil {
		g.kb.queue = append(g.kb.queue, g.script.keyTaps...)
	}
}

// enqueueText gives a typed character to the key press it came with — a
// printable or keypad key of this frame still without one, the character's
// own first — or queues it as its key pressed again (the key repeat): a
// keypad key's while one is held (its character is not its key: a keypad 1
// is no "1" shortcut), else the character's own.
func (g *Game) enqueueText(first int, r rune) {
	own := keyOfChar(r)
	at := -1
	for i := first; i < len(g.kb.queue); i++ {
		if e := g.kb.queue[i]; e.charCode == 0 && (isPrintableKey(e.keyCode) || isKeypadTextKey(e.keyCode)) && (at < 0 || e.keyCode == own) {
			at = i
			if e.keyCode == own {
				break
			}
		}
	}
	if at >= 0 {
		g.kb.queue[at].charCode = r
		return
	}
	for k, code := range ebitenKeyCodes {
		if isKeypadTextKey(code) && ebiten.IsKeyPressed(k) {
			own = code
			break
		}
	}
	g.kb.queue = append(g.kb.queue, keyEvent{own, r})
}

// keyOfChar is the key that types a character on SDL's default layout, as
// far as shortcuts go: the character itself, a letter lower-case (0: none).
func keyOfChar(r rune) uint32 {
	switch {
	case r >= 'A' && r <= 'Z':
		return uint32(r - 'A' + 'a')
	case r >= 0x20 && r < 0x7F:
		return uint32(r)
	}
	return 0
}

// updateKeyboard is a frame's keyboard: the keys read, handleKeyboard, then
// processKeyboardInput.
func (g *Game) updateKeyboard() {
	g.readKeys()
	g.handleKeyboard()
	g.cheat()
	g.normalKey()
}

// handleKeyboard is Input::handleKeyboard: the screenshot's countdown, the
// edge scrolling, the modifiers as the keyboard has them, the arrow keys'
// scrolling.
func (g *Game) handleKeyboard() {
	g.handleScreenshotCountdown()
	g.edgeScroll()

	kb := &g.kb
	kb.modifier &^= shortcuts.ModShift | shortcuts.ModControl | shortcuts.ModAlt | shortcuts.ModUnknown
	for _, m := range [...]struct {
		key uint32
		mod shortcuts.KeyModifier
	}{{shortcuts.KeyLShift, shortcuts.ModShift}, {shortcuts.KeyRShift, shortcuts.ModShift},
		{shortcuts.KeyLCtrl, shortcuts.ModLeftControl}, {shortcuts.KeyRCtrl, shortcuts.ModRightControl},
		{shortcuts.KeyLAlt, shortcuts.ModLeftAlt}, {shortcuts.KeyRAlt, shortcuts.ModRightAlt}} {
		if g.keyHeld(m.key) {
			kb.modifier |= m.mod
		}
	}
	g.keyScroll()
}

// hasKeyModifier is Input::hasKeyModifier.
func (g *Game) hasKeyModifier(m shortcuts.KeyModifier) bool { return g.kb.modifier&m != 0 }

// edgeScroll is Keyboard.cpp's edgeScroll: the cursor on the screen's very
// edge (x == 0, x >= width − 1, likewise y) moves the main view by
// edgeScrollingSpeed (Config default 12) << zoom — with the window focused,
// only in the normal input state (not while dragging the view or moving a
// window), not with Shift or Ctrl held, not on the title screen.
// PARTIAL: Input::State::dropdownActive scrolls too upstream; the news
// window's viewportNoScrolling.
func (g *Game) edgeScroll() {
	if !ebiten.IsFocused() || g.noEdgeScrolling || g.vpRight.on || g.wins.Dragging() {
		return
	}
	if g.hasKeyModifier(shortcuts.ModShift) || g.hasKeyModifier(shortcuts.ModControl) {
		return
	}
	speed := float64(edgeScrollingSpeed)
	var dx, dy float64
	if g.mouseX == 0 {
		dx -= speed
	}
	if g.mouseX >= g.sw-1 {
		dx += speed
	}
	if g.mouseY == 0 {
		dy -= speed
	}
	if g.mouseY >= g.sh-1 {
		dy += speed
	}
	g.scrollMainView(dx, dy)
}

// keyScroll is Keyboard.cpp's keyScroll: the arrow keys held move the main
// view by edgeScrollingSpeed << zoom — not under a modal window or a text
// input, not on the title screen.
func (g *Game) keyScroll() {
	if g.prompt != nil || g.browseWin() != nil { // WindowManager::getCurrentModalType
		return
	}
	if g.textInput != nil {
		return
	}
	speed := float64(edgeScrollingSpeed)
	var dx, dy float64
	if g.keyHeld(shortcuts.KeyLeft) {
		dx -= speed
	}
	if g.keyHeld(shortcuts.KeyUp) {
		dy -= speed
	}
	if g.keyHeld(shortcuts.KeyDown) {
		dy += speed
	}
	if g.keyHeld(shortcuts.KeyRight) {
		dx += speed
	}
	g.scrollMainView(dx, dy)
}

// scrollMainView is the end of edgeScroll and keyScroll: the main view moved
// by the delta at its zoom — not while a window forbids it
// (WindowFlags::viewportNoScrolling: the news window), not on the title
// screen.
func (g *Game) scrollMainView(dx, dy float64) {
	if dx == 0 && dy == 0 || g.inTitleScreen {
		return
	}
	scale := float64(int(1) << g.w.GetZoom())
	g.w.PanCamera(dx*scale, dy*scale)
}

// cheat is Keyboard.cpp's cheat (0x004BEC5B): while Insert or an Alt key is
// held the letters typed are collected (normalKey); let go, the word is
// looked up — not on the title screen.
func (g *Game) cheat() {
	kb := &g.kb
	if g.keyHeld(shortcuts.KeyInsert) || g.keyHeld(shortcuts.KeyLAlt) || g.keyHeld(shortcuts.KeyRAlt) {
		if !g.hasKeyModifier(shortcuts.ModCheat) {
			kb.modifier |= shortcuts.ModCheat
			kb.cheatBuffer = ""
		}
		return
	}
	if !g.hasKeyModifier(shortcuts.ModCheat) {
		return
	}
	kb.modifier &^= shortcuts.ModCheat
	if g.inTitleScreen {
		return
	}
	switch kb.cheatBuffer { // kCheats
	case "DRIVER":
		g.cheatDriver()
	case "SHUNT":
		g.cheatShunt()
	case "FREECASH":
		g.cheatFreeCash()
	}
}

// normalKey is Keyboard.cpp's normalKey (0x004BEDA0): each queued key in
// turn — to the cheat word while one is being typed, else to the windows
// that take keys, front to back, else to the shortcut bound to it.
func (g *Game) normalKey() {
	kb := &g.kb
	queue := kb.queue
	kb.queue = nil
	for _, key := range queue {
		if isModifierKey(key.keyCode) {
			continue
		}
		if g.hasKeyModifier(shortcuts.ModCheat) {
			c := key.charCode
			if c >= 'a' && c <= 'z' {
				c += 'A' - 'a'
			}
			if c >= 'A' && c <= 'Z' {
				kb.cheatBuffer += string(c)
			}
			if kb.modifier&shortcuts.ModAlt == 0 {
				continue
			}
		}
		if g.callKeyUp(key) { // WindowManager::callKeyUpEventBackToFront
			continue
		}
		if g.tryShortcut(shortcuts.Screenshot, key) {
			continue
		}
		// (Tutorial::stop while one plays.)
		if g.tryShortcut(shortcuts.OpenDebugWindow, key) {
			continue
		}
		// (The intro's keys.)
		if !g.inTitleScreen {
			if id, ok := kb.bindings.Find(key.keyCode, kb.modifier); ok {
				g.executeShortcut(id)
			}
			continue
		}
		if g.tryShortcut(shortcuts.SendMessage, key) {
			continue
		}
		if g.tryShortcut(shortcuts.Screenshot, key) {
			continue
		}
		g.tryShortcut(shortcuts.ShowOptionsWindow, key)
	}
}

// tryShortcut is Keyboard.cpp's tryShortcut: the shortcut run when the key,
// with the modifiers held, is its binding.
func (g *Game) tryShortcut(id shortcuts.Shortcut, key keyEvent) bool {
	if g.kb.bindings.Match(id, key.keyCode, g.kb.modifier) {
		g.executeShortcut(id)
		return true
	}
	return false
}

// callKeyUp is WindowManager::callKeyUpEventBackToFront: the key offered to
// each window that takes keys, the front one first, until one says it was
// its own. The confirmation prompt and the text input are in front of the
// stack's windows (the prompt modal, the text input stickToFront).
// PARTIAL: the Build Vehicle window's search box, the object selection
// window's and the edit-keyboard-shortcut window (none of them ported).
func (g *Game) callKeyUp(key keyEvent) bool {
	if p := g.prompt; p != nil { // PromptOkCancelWindow.cpp keyUp
		switch key.keyCode {
		case shortcuts.KeyEscape:
			p.Key(true, false)
		case shortcuts.KeyReturn, shortcuts.KeyKPEnter:
			p.Key(false, true)
		default:
			return false
		}
		if p.Closed && g.prompt == p {
			g.prompt = nil
		}
		return true
	}
	if t := g.textInput; t != nil {
		g.textInputKeyUp(t, key)
		if t.Closed && g.textInput == t {
			g.textInput = nil
		}
		return true
	}
	for i := len(g.wins.Windows) - 1; i >= 0; i-- {
		if b, ok := g.wins.Windows[i].(*ui.PromptBrowseWindow); ok && g.browseKeyUp(b, key) {
			return true
		}
	}
	return false
}

// textInputKeyUp is TextInputWindow.cpp's keyUp (0x004CE910): Return is OK;
// else the key is the input session's (InputSession::handleInput: a typed
// character, Backspace, Delete, Home, End, Left, Right) — which takes every
// key, whatever it is: no shortcut runs under an open text input.
// Upstream closes on a *character* Escape, which SDL never sends: Escape
// does nothing there (checked on OpenLoco 26.09).
// PARTIAL: Input::isFocused — the text box loses the keys when another
// widget is clicked.
func (g *Game) textInputKeyUp(t *ui.TextInputWindow, key keyEvent) {
	if key.charCode == rune(shortcuts.KeyReturn) {
		t.Keys(nil, nil, true, false)
		return
	}
	var typed []rune
	var keys []int
	switch c := key.charCode; {
	case c >= 0x20 && c < 0x7F || c >= 159:
		typed = []rune{c}
	case c == rune(shortcuts.KeyBackspace):
		keys = []int{ui.KeyBackspace}
	default:
		if k, ok := map[uint32]int{shortcuts.KeyDelete: ui.KeyDelete, shortcuts.KeyHome: ui.KeyHome, shortcuts.KeyEnd: ui.KeyEnd,
			shortcuts.KeyLeft: ui.KeyLeft, shortcuts.KeyRight: ui.KeyRight}[key.keyCode]; ok {
			keys = []int{k}
		}
	}
	t.Keys(typed, keys, false, false)
}

// browseKeyUp is PromptBrowseWindow.cpp's keyUp: Return is OK, Escape the
// close button; with the filename box (save mode) every other key is the
// box's; without it (load mode) the key goes on.
func (g *Game) browseKeyUp(b *ui.PromptBrowseWindow, key keyEvent) bool {
	switch key.keyCode {
	case shortcuts.KeyReturn:
		b.HandleText(nil, false, true, false)
		return true
	case shortcuts.KeyEscape:
		b.HandleText(nil, false, false, true)
		return true
	}
	if !b.Save {
		return false
	}
	var typed []rune
	if c := key.charCode; c >= 0x20 && c < 0x7F || c >= 159 {
		typed = []rune{c}
	}
	b.HandleText(typed, key.charCode == rune(shortcuts.KeyBackspace), false, false)
	return true
}

// loadShortcutBindings is Shortcuts::loadBindings on the config file's
// shortcuts section (a shortcut it doesn't name is on its default; nothing
// is written).
func loadShortcutBindings() *shortcuts.Bindings {
	cfg := map[string]string{}
	for _, def := range shortcuts.Defs {
		if v, ok := configLookup("shortcuts." + def.ConfigName); ok {
			cfg[def.ConfigName] = v
		}
	}
	return shortcuts.Load(cfg)
}

// scriptKey reads a script's key: a binding as the config file writes them
// ("Z", "PageUp", "Left Ctrl+S", "Left Shift+Backspace" — SDL's key names,
// any case), with ESC and ENTER for Escape and Return. It gives the key's
// event and the modifier keys held with it (ok false: no such key).
func scriptKey(name string) (ev keyEvent, held []uint32, ok bool) {
	parts := strings.Split(name, "+")
	if name == "+" || strings.HasSuffix(name, "++") { // the + key itself
		parts = append(parts[:len(parts)-2], "+")
	}
	for i, p := range parts {
		p = strings.TrimSpace(p)
		switch strings.ToUpper(p) {
		case "ESC":
			p = "Escape"
		case "ENTER":
			p = "Return"
		}
		code := shortcuts.KeyFromName(p)
		if code == 0 {
			return keyEvent{}, nil, false
		}
		if i < len(parts)-1 {
			held = append(held, code)
			continue
		}
		ev.keyCode = code
	}
	ctrlOrAlt := false
	shift := false
	for _, h := range held {
		switch h {
		case shortcuts.KeyLShift, shortcuts.KeyRShift:
			shift = true
		default:
			ctrlOrAlt = true
		}
	}
	switch {
	case ev.keyCode == shortcuts.KeyReturn, ev.keyCode == shortcuts.KeyBackspace, ev.keyCode == shortcuts.KeyDelete:
		ev.charCode = rune(ev.keyCode) // handleKeyInput
	case isPrintableKey(ev.keyCode) && !ctrlOrAlt: // the text SDL sends with the key
		ev.charCode = rune(ev.keyCode)
		if shift && ev.charCode >= 'a' && ev.charCode <= 'z' {
			ev.charCode += 'A' - 'a'
		}
	}
	return ev, held, true
}
