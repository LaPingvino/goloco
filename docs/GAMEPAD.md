# Gamepad play for mouse-driven sim games

Written 2026-11-15. goloco is the first place this is built, but the first half is meant to be general: a reference
for making OpenTTD, OpenRCT2 and other mouse-heavy management games playable with a controller on SteamOS (Steam
Deck), on consoles, in the browser, and for players who can't use a mouse well. The second half is goloco's plan
(plan E, next to PORT-PLANS.md).

## Part 1: the general design

### Why not just emulate the mouse

Steam Input can already map a stick or trackpad to a mouse pointer. That makes these games *possible* on a Deck,
but not pleasant: steering a pointer to a 10-pixel button is slow and tiring, and a game made for a mouse has dozens
of small targets per window. The better model has four layers. Each one works on its own, and every layer can fall
back to the one below it:

1. **Actions, not buttons.** The game reads abstract actions (`confirm`, `cancel`, `nextWindow`, `rotate`,
   `zoomIn` …), and the keyboard, mouse and pad all feed the same actions. The binding table lives in one place, so
   players can rebind it, and it gives the button prompts (see below) their source.
2. **Focus navigation.** Every window has a focus order over its controls. The shoulder buttons cycle the focused
   window (like Ctrl-Tab), the triggers cycle the tabs inside it, the d-pad moves between controls, A activates, and
   B closes or backs out. The same focus works with Tab, Shift-Tab and the arrows on a keyboard, and it gives screen
   readers and switch-access devices something to hook into. This is the accessibility gain, and it is worth having
   even without a pad.
3. **Context tools.** When a tool is active (build track, build road, place a station, bulldoze, terraform), the
   buttons change meaning to suit it. A places, B cancels the tool, X/Y step through the tool's main choices (a
   curve or slope, a station size), and the left stick moves the tool's ghost **one tile per step**, with the view
   following. A tile cursor that snaps to the grid beats a free pointer for almost every construction tool.
4. **Free cursor.** The left stick moves a pointer (with acceleration and a precision modifier) and A/B are the
   left and right clicks. Nothing is ever out of reach, even in windows that aren't focus-aware yet.

### Stick modes: pointer, scroll, tile

The left stick has three modes, and one button (say, clicking the left stick) cycles through them, with the current
mode shown by the cursor's shape:
- **Pointer:** layer 4's free cursor.
- **Scroll:** the stick pans the view (for a pad without a usable right stick, or one-handed play).
- **Tile:** a cursor that sits on one map tile and steps one tile per push, along the map's axes for the current
  rotation (so "up" on the stick stays "up the screen" in every view). The view scrolls to keep it in sight. A and B
  are a left and right click **on that tile**: the game gets a click at the tile's screen position, exactly as if the
  mouse had been there.

The tile mode is the big one. In OpenTTD, OpenRCT2 and Locomotion nearly everything on the map is done by clicking the
right tile: build a piece, place a station or a signal, bulldoze, raise land, open a vehicle or an industry. A tile
cursor makes all of that precise without any per-tool code, because it only produces the clicks the tools already
take. Drag tools (OpenTTD's rail drag, area select, station drag) work the same way: holding A is holding the button,
and each step of the stick extends the drag by one tile. Details that make it pleasant:
- holding the stick repeats the step after a short delay, faster the longer it's held;
- a modifier (a held shoulder or trigger) steps several tiles at once;
- the tile mode is the default while a map tool is active, and a tool can still move its own ghost (layer 3) where it
  does more than take a click;
- the tile under the cursor is highlighted the way the game highlights the mouse's tile, so no new drawing is needed;
- a click in the tile mode can target the tile's centre or one edge or corner (a signal side, a quarter tile): a
  modifier plus the d-pad nudges within the tile.

On the main view, without a tool: the right stick pans, the triggers zoom (or a modifier plus the d-pad does), A
opens what is under the centre or the cursor (a vehicle, station, town or industry window), and Y opens a radial or
list menu of the toolbar's tools. Console builders (Cities: Skylines, Planet Coaster console editions) settled on
much the same shape: the camera on the sticks, tools in a radial menu, a snapping cursor. That is worth a look before
inventing new things.

### Button prompts

Show the button next to the control it triggers ("Ⓐ Place  Ⓑ Cancel  ⓧ Curve") whenever the last input came from the
pad, and hide the prompts as soon as the mouse moves. The glyphs follow the pad's layout (Xbox, PlayStation, Nintendo,
the Deck). The prompt text comes from layer 1's binding table, never from hard-coded strings.

### SteamOS and Steam Deck notes

- **The screen is 1280×800 at 7 inches.** UI scale matters more than input. A 2× scale of a 640-pixel-wide UI fits
  exactly, which is goloco's zoom plan (C) and its UI scale.
- **Steam Input sits in front of the game.** A game that reads the SDL GameController or Ebitengine's standard layout
  gets the right buttons. Ship a default Steam Input configuration that sends the pad through as a pad (not as
  keyboard and mouse), so layers 1–3 see it, and keep the trackpads as a mouse for layer 4 if wanted.
- **Text input** (company names, save names): open the on-screen keyboard. On Steam that means Steamworks
  `ShowGamepadTextInput` / `ShowFloatingGamepadTextInput`; without Steamworks, the player opens it with Steam+X. A
  game should at least never *need* typing: default names, and A to accept them.
- **Suspend and resume:** the Deck sleeps mid-game. Keep the game clock stable across a resume (cap the catch-up
  step), the way a long frame is capped already.
- **Hot-plug:** pads come and go. Pick up a new pad without a restart, and pause on disconnect.

### Browser (wasm)

The browser's Gamepad API gives the same standard layout; Ebitengine and Emscripten's SDL2 both use it. A page only
sees a pad after a button press while the page has focus, so say "press any button" on the title screen. Polling
happens once per frame, so there is nothing else to wire up.

### For OpenTTD and OpenRCT2

As far as I know (not checked recently), neither has built-in controller support beyond what Steam Input's mouse
emulation gives. Both are SDL2 games with their own widget systems, which suits this design: every window already
holds a widget list with rectangles, which is all that focus navigation (layer 2) needs. A good first contribution
in either would be layer 1 plus layer 2 behind a setting, with no change to the default mouse play. Check upstream's
issue trackers and any earlier attempts before proposing, and keep each step small enough to review.

## Part 2: goloco (plan E)

goloco runs on Ebitengine, so `ebiten.StandardGamepad*` covers Linux, Windows, macOS, the browser and the Deck with one
API. Port first, extend after: following OpenLoco is the starting point, not a limit. The faithful UI stays the default, and the pad
is a layer on top that doesn't change what the mouse does.

How goloco's code fits:
- **One input entry point.** cmd/goloco keyboard.go (Keyboard.cpp) and update.go read the input; the shortcut table
  (pkg/shortcuts) is the action list of layer 1 already. A pad binding column next to the key column is the cheapest
  start.
- **The window stack** (pkg/ui winstack.go) knows every window's rectangle and z-order, so cycling the focused window
  is a stack operation.
- **Windows take clicks by position** (`Win.HandleClick(x, y)`), not as widget events. So focus can be an optional
  interface, `FocusTargets() []image.Rectangle` in focus order, and activation is a click at the focused rectangle's
  centre. Every window works through the free cursor from day one, and each window gets real focus when it gets
  `FocusTargets`. Most windows already have a widget table to build that list from (e.g. buildvehicle.go's
  `BuildVehicleLayout`).
- **The construction tools** (track, road, signal, station, terraform) already move a ghost along the tile grid
  from the cursor; the pad moves that same tile position instead.

Order of work (each step green and committed, like the other plans):
1. Gamepad input: read the standard layout, hot-plug, the free cursor (layer 4) with A/B as clicks and a dead zone,
   the right stick panning and the triggers zooming. Then the stick modes: pointer, scroll and tile. The tile mode
   steps along the map axes for `sim.Map.CurrentRotation` and clicks at the tile's screen point (`sim.GameToScreen`),
   so every tool works with it at once. Under Xvfb, test with an injected pad state (the same pattern as
   the script player's input).
2. Actions: a pad column in the shortcut table (pause, rotate, the menus), shown in the keyboard shortcuts window, and
   saved in openloco.yml next to the keys.
3. Focus: the `FocusTargets` interface, a focus ring, the shoulders cycling windows and the triggers cycling tabs.
   Start with the Options, Build Vehicle and Station windows, then the rest. Tab and the arrows do the same.
4. Context tools: track and road construction first (A builds, B cancels, X/Y pick curve or slope, the stick steps the
   piece), then signals, stations, terraform and the bulldozer.
5. Button prompts, and a radial or list menu for the toolbar.
6. The Deck: a 2× UI scale (after plan C's zoom), a Steam Input template, and no typing needed for a new game.

For OpenTTD the tile mode alone would be the biggest single gain: its viewport code already turns a tile into a screen
position (and back for the mouse), and its tools all act on the tile under the cursor.

Other games in ~/Projects with gamepad work queued: Pits (SDL2 `sdl_compat.c`: map pad events onto `key_event`),
Pacmanio and Stelo (Ebitengine: an action layer over the existing key checks). These are layer 1 + 4 only: they have
no mouse-heavy windows.
