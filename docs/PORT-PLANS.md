# Port plans for the big items

Written 2026-11-14 so the agents doing these ports start with the map instead of exploring. Each plan names
upstream's files (~/Projects/OpenLoco, 09d74ee), what goloco already has, an order of work in steps that each end
green and committed, and how to verify. Re-check a plan's facts briefly before starting: the code moves.

General rules (CLAUDE.md has the rest):
- Port, don't approximate. Each step leaves `scripts/test-all.sh` and `scripts/playthrough/run.sh` green
  (private TMPDIR, sandbox off), with a commit.
- Compare against the real OpenLoco with scripts/olref, using your own OLREF_DISPLAY and OLREF_HOME.
- Never touch the user's openloco.yml; point XDG_CONFIG_HOME at a scratch copy.
- Budget: these are week-sized jobs. Stop at a green, committed step, and write what is left in LOOP.md.

## A. The map window (LOOP N34 showMap) — Ui/Windows/MapWindow.cpp, 2504 lines

goloco has none. The toolbar's map entries and the `showMap` shortcut (M) do nothing, and nothing in pkg/ or cmd/
draws a minimap.

Upstream's structure:
- **Pixel buffer:** `_mapPixels` (two 768×768 frames, flipped each update) is filled a few rows per tick:
  - `setMapPixelsOverall` (l. 258): surface colours by height and land type, water, track, road, buildings, trees;
  - `setMapPixelsVehicles` (l. 447), `setMapPixelsIndustries` (l. 549), `setMapPixelsRoutes` (l. 671);
  - the ownership tab (companies' track, road and stations).
- **Tabs:** overall, vehicles, industries, routes, ownership. Each has a graph key on the right (`drawGraphKey*`,
  l. 1337–1526) whose rows filter what is drawn.
- **drawScroll (l. 2169):** the buffer rotated 45° for the view's rotation, then vehicles (`drawVehiclesOnMap`),
  routes, town names (`drawTownNames`) and the main view's rectangle (`drawViewportPosition`).
- **Input:** a click centres the main view; scroll and resize; the window opens centred on the main view
  (`centerOnViewPoint`, also called by ToolbarTop's rotate).

Status (2026-11-16): step 1's pixels done — sim.MapPixelsOverall (pkg/sim/mappixels.go) is setMapPixelsOverall
per tile (colours and flash colours; tested on the title map, mappixels_test.go), and step 2's buffer: sim.MapBuffer
(_mapPixels / _mapAltPixels, cleared black0) filled a row per update by Map.FillMapRowOverall (setMapPixels'
stepping for the four rotations; layout tested). Next: the window shell, in tick-sized parts:
   a. DONE (2026-11-16) but the scroll bars: the zoom menu's map entry and the time panel's map button open it too.
      pkg/ui MapWindow (WindowType
      map), cmd/goloco openMap (M; the buffer filled whole at open). Was:
      pkg/ui MapWindow (WindowType map): frame, caption, close, the scroll view (content 768 × 768, both bars),
      size 350 × 272 (resizable, kMinWindowSize .. kMaxWindowSize); drawScroll = the buffer (Pixels, or Alt on
      flashing frames) blitted at the scroll offset. Opened by shortcuts.ShowMap (shortcut_actions.go's stub),
      ToolbarTop's map entry (l. 628) and TimePanel (l. 268/277); bringToFront when open.
   b. DONE but the remembered size (2026-11-16): onUpdate fills 80 rows a frame (cmd/goloco updateMap), a new
      rotation clears the buffer. Was: onUpdate: Map.FillMapRowOverall a row per update (the open's clearMap and var_846 rotation; a rotation
      change clears and restarts); getLastMapWindowAttributes (size, maximised) kept on close.
   c. DONE (2026-11-19): centerOnViewPoint (open; the rotate key, the right-held wheel and ToolbarTop's rotate
      menu: cmd/goloco Game.rotateView), scrollMouseDown / scrollMouseDrag centring the main view
      (World.CentreOnMapView = moveMainViewToMapView), the conversions mapWindowPosToLocation /
      locationToMapWindowPos (sim mapwindowpos.go, four rotations tested), and step 3's view rectangle
      (drawViewportPosition). Checked against OpenLoco (scripts/olref/tours/mapcentre.txt), playthrough t-map.
      Left: the scroll bars (step a), the main view's glide (Window.cpp viewportsUpdatePosition's
      scrollingToLocation: goloco jumps, for every centring), _flashingItems (with the graph key, d).
   d. DONE (2026-11-19) but the editor's disabled tabs: the five tabs (drawTabs' images, the vehicles' animated and
      recoloured, the routes' animated; onMouseUp's tab change: frameNo and var_854 reset, the buffer kept; the
      caption per tab), the fills setMapPixelsVehicles / Industries / Routes / Ownership (sim maptabs.go
      MapPixelsTab, FillMapRow by tab), assignIndustryColours / assignRouteColours (sim.MapColours, kept on the
      Game between opens as upstream's statics; IndustryObject mapColour 0xE3 parsed), the graph keys
      (drawGraphKeyOverall … Companies) with setHoverItemTab / setHoverItem (var_854 → _flashingItems at the
      next fill: the rows' items flash on mapFrameNumber's bit 2, the row white, the swatch blinking, the view
      frame's corners off), draw's minHeight and status bar (formatVehicleString / formatIndustryString),
      drawScroll's drawVehiclesOnMap (countVehiclesOnMap; getVehicleColour; the routes tab's aircraft and ship
      routes, drawLine ported with its unordered-rectangle early-out) and drawTownNames. Checked against OpenLoco
      (scripts/olref/tours/maptabs.txt: the title save, each tab with and without a hovered key row): equal but
      the scroll bars, the vehicles' positions by the seconds the games ran, and the animation / flash phase.
      Playthrough t-map (expect maptab / maphover). Left: the scroll bars (a), the remembered size (b), the
      glide (c), and the editor's disabled vehicles / routes / ownership tabs (goloco has no scenario editor).
   Skin colours: InterfaceSkinObject windowTitlebarColour / windowMapColour.

Order of work:
1. Window shell: open/close from the toolbar's Map entry and the M key, frame, the five tabs, a resizable scroll
   view, the overall tab's pixels filled all at once (correct first, incremental later). Test the pixel colour of a
   few known tiles against upstream's table.
2. Upstream's incremental fill: rows per tick, two buffers.
3. Overlays: the view rectangle, vehicles, town names; a click and drag centre the main view.
4. The other tabs' fills and their graph keys with filtering.
5. Compare with OpenLoco's map window on the title game: tabs, colours, overlays; the rotation in all four views.

Reuse:
- goloco's scroll views (`drawVScrollViewPressed`, ui.MusicSelectionWindow as a pattern);
- surface data in sim.Map.Tiles;
- vehicle positions in sim's entity lists;
- the land objects' map colours: check LandObject for `mapPixelColour`; it may need parsing.

## B. Sounds (LOOP N9) — Audio/Audio.cpp 576, VehicleAudio.cpp 254, AmbientAudio.cpp 178, Audio/src/AudioEngine.cpp 598

Steps 1–4 are done (2026-11-15 … 11-18); what is left is listed at the end.
- Step 1: pkg/objects/sound.go `ParseSoundObject` (the 42 SND*.DAT objects) and pkg/audio/css.go `ReadCSS` (css1.dat).
- Step 2: pkg/audio/effects.go `Effects` (channel volumes, kSoundVolumeTable, percentToDb, calculatePan, one-shots,
  a `Record`/`Played` test hook), cmd/goloco sfx.go; the world-position form (world.SoundAt) and the sim's sound
  queue (pkg/sim/sounds.go → playSimSounds). Every effect call site of the commands, the sim, the news and the tools.
  The widget clicks (2026-11-18): pkg/ui clicks.go — a window names the widget its HandleClick hit (WinBase.Button /
  PressWidget) and the stack sounds MouseInput's clickDown at window x + Widget::midX, clickUp on the release over it
  (or over the main viewport), clickPress on an unmoved caption / frame release; every window on the stack names its
  widgets (close, tabs, buttons, checkboxes, steppers, dropdowns), plus the toolbar, time panel and company panel.
- Step 3 (2026-11-18): the sim keeps VehicleSound (pkg/sim drivingsound.go: updateDrivingSounds' friction, simple and
  gearbox motors; updateTrainProperties' sounding objects and the turn-round swap), saved at Vehicle2 / tail 0x44–0x4B;
  sound objects in the save's slot order (ReorderSoundObjects, slots 1–128) — title.dat's 101 saved driving sounds
  match. The chuffs (exhaust.go chuffSound, tunnel chuffs underground and in stations) and the crossing whistle are
  queued; object one-shots play at the object's volume (Effects.PlayObjectAt). The mixer has lasting handles
  (voice.go: create / setVolume / setPan / setPitch / destroy on a looping stream), and VehicleAudio.cpp runs in
  pkg/audio vehicleaudio.go (updateSingleVehicleSound, getVehicleAudioAttributes) from cmd/goloco vehicleaudio.go.
  NB the plan's old premise "a fixed set of channels chosen by distance" was wrong: upstream keeps one handle per
  Vehicle2 / tail sound while a view shows it; AudioEngine's only cap is OpenAL's source count.
- Step 4 (2026-11-18): AmbientAudio.cpp — world AmbientScan (the 12 × 12 tiles about the main view's centre), the
  sea / wilderness / forest choice and its fades (pkg/audio ambient.go, css2–4.dat via ReadMusicSample). Audio::tick
  runs once a frame and once per game tick, as upstream.

Left:
- Clicks: widgets goloco's windows don't act on stay silent (labels, BuildVehicle's filters and search, Options'
  Company tab, the audio device dropdown and the five other volume sliders, the paint brush, the bars' wt3 frames);
  the release off the pressed widget over another window's widget; a dropdown's release path
  (dropdownRegisterSelection's clickUp); the modal prompts outside the stack (text input, save, OK / Cancel); the
  cheat words' click, the town tools' construct sounds (tools not ported).
- World sounds and vehicles are heard from the main view only (findBestViewportForSound / ForEntity's other
  windows' viewports; the main view following a train).
- The tunnel reverb (setReverb: OpenAL EFX; Ebitengine has none).
- With ALSA's null device the mixer is read as fast as it swallows (CPU as before vehicle audio: measured 144 s
  before, 128 s after for 60 s of title).

Upstream's structure:
- **Samples:** `css1.dat` (Environment::PathId::css1) holds the UI samples, read by `loadSoundsFromCSS`
  (Audio.cpp:54). The vehicle and ambient sounds are SoundObjects (`getSoundObject`, Audio.cpp:348) with their own
  sample data.
- **`playSound(SoundId, pos | pan)`:** a sound at a world position is given volume and pan from the main view
  (distance, zoom). 79 call sites; the most are in:
  - ViewportInteraction (demolish);
  - Vehicle and VehicleHead (horns, crashes, splash);
  - TerraForm, MouseInput (clicks), Keyboard, the construction tabs (construct).
- **Vehicle audio:** one handle per Vehicle2 / tail driving sound while a view shows the vehicle
  (VehicleAudio.cpp).
- **Ambient:** sea, forest, town and so on, from the tiles around the view (AmbientAudio.cpp).

Order of work:
1. A SoundObject loader (Object::load layout byte for byte, its PCM data), and the css1.dat reader. Unit-test the
   sample counts and lengths against upstream's.
2. An effects mixer on the existing audio.Context: `playSound` with volume and pan, and the sfx volume option.
   Wire the UI sounds (clicks, the construct and demolish sounds) at upstream's call sites; grep
   `Audio::playSound(` and port each. Every "PARTIAL: the … sound" note in goloco marks a call site
   (grep "sound" PARTIAL).
3. Vehicle audio channels.
4. Ambient audio.

Testing without a sound card: ALSA_CONFIG_PATH null device (CLAUDE.md). Assert in tests which sounds were
queued, through a test hook on the mixer.

## C. 2× / 4× zoom (LOOP N30) — upstream 9850f2d, 124 files

The core is ZoomLevel.hpp (negative levels: −1 = 2×, −2 = 4×), RenderTarget, the software drawing context, and
DrawSpriteBMP/RLE (scaling sprites up). The rest is mostly mechanical `zoom` → `ZoomLevel` call-site changes.
goloco's views clamp zoom to 0…3 (pkg/world/world.go ZoomIn/ZoomInAt); its renderer is pkg/scene with gfx8
sprites. Zoom-dependent things to find:
- label frames (Town/Station LabelFrame per zoom);
- vehiclesMinScale / stationNamesMinScale;
- the paint session's zoom checks (`s.Zoom <= …`);
- the money effects (zoom ≤ 1);
- map tooltips and hit tests (`<< zoom`).

Order of work:
1. gfx8: draw a sprite scaled ×2 / ×4, nearest neighbour as upstream's DrawSprite*.hpp do; golden-test against
   their pixel logic.
2. world/scene: negative zoom in the view maths. Replace the `<< zoom` shifts with a helper that handles negative
   levels, because shifts by a negative count panic or misbehave in Go.
3. Labels, hit tests, tooltips, the money effects at the new levels.
4. Compare with OpenLoco at both new levels on the title game.

## E. Gamepad and focus navigation — docs/GAMEPAD.md

The design (general, meant as a reference for other sim games too) and goloco's six steps are in docs/GAMEPAD.md.
Design agreed in outline by Joop 2026-11-15 (shoulders tab through windows, buttons per context); the details are open.

## F. Company AI — configurable, not loyal (Joop 2026-11-16)

Joop: "interesting to make configurable like with OpenTTD, because the Locomotion one is famously not that great...
That definitely doesn't have priority to be loyal to the original, rather it should be easy to make accurate,
configurable and good."

DECIDED (Joop 2026-11-19): the interface first and our own AI, not a port of the original — "the original AI is
not super great and if you want that, you can just play the original". The original is studied, not ported: "you
could start a to figure out which end points we need".

Direction:
- **Step 1 — study the original for its end points:** World/CompanyAi/CompanyAi.cpp (7889 lines),
  CompanyAiPathfinding.cpp (2665), CompanyAiPlaceVehicle.cpp (659). List everything it reads (observations) and
  every game command it issues (actions), and where its state lives in the company record. The result is the
  interface's first draft (docs/AI-INTERFACE.md), with goloco's existing equivalent or the gap for each.
- **An AI interface in pkg/sim:** what a company's AI may see (the map, towns, industries, its own finances) and do
  (game commands only, as a player does: no shortcuts). One AI per company, chosen when it's created.
- **Cheats: configurable, off by default, and designed so they don't break the AI** (Joop 2026-11-19: "make that
  configurable. There are stories about how the cheating actually breaks AIs badly, maybe read up on that, and
  it's probably good to still allow them to be challenging and for that allow some kind of cheating if turned on,
  but in a way that actually works and doesn't destroy them"). Default: the AI sees and does what a player can
  (own loans, private planning, its own rng, no peeking at competitors' plans). Advantages are separate switches
  per AI or difficulty level. To do before building them: read up on how cheats break game AIs (OpenLoco's own
  free-build fallback — retrying for free hides bad plans instead of teaching the AI to plan within its means;
  OpenTTD / TTD, Civ, RTS experiences) and pick advantages that keep the AI's decisions sound, e.g. a cost or
  income multiplier the AI plans WITH (it knows its real prices) rather than silent free retries it can't see.
- **The original is not ported.** Consequence: the economy comparison with OpenLoco (simcmp) stays an AI-off check;
  the AI companies' records (aiThoughts, the AI fields of the company) are kept as loaded, so OpenLoco saves still
  load and save.
- **Better AIs beside it:** written against the same interface; picked per scenario or game in the options, like
  OpenTTD's AI settings (which AI, how many competitors, how fast they start, how hard they play).
- **Easy to make accurate:** each AI's decisions are logged and replayable, so a bad choice can be traced to its
  inputs; the comparison tools (simcmp, rngtrace) work with the original AI on.

## D. Smaller follow-ups ready to pick up

- **N35 save leftovers:**
  - ~~write last_save_path~~ (done 2026-11-15: the load and save browsers write it);
  - ~~"Save this before loading?" when loading during play~~ (done 2026-11-15: PromptSaveWindow, for loading,
    quitting to the title and quitting);
  - ~~the Options autosave widgets read the config~~ (done 2026-11-14);
  - re-pack every custom object, not only the loaded file's.
- **N7:** a refused road's reason line, the bridge and alternate-road retries, shift's downward search.
- **N6:** (bars on the stack, flash/sound, resizing, vehicle list columns done 2026-11-18) drag by frame for the
  windows without a widget list; the Company window's headquarters viewport.
- **N32:** see-through checks still not compared with OpenLoco; the unexplained dark face at 186,143 underground.
