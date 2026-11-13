# Driving the real OpenLoco for reference shots

OpenLoco (AUR openloco-git, /usr/bin/openloco — 26.09.r22.g09d74ee39) runs headless under Xvfb
and is driven through X11's XTest extension (ctypes on libXtst — no xdotool).
No game data and no screenshots here: OpenLoco reads the player's own
Locomotion folder, shots go to a scratch directory.

    export OLREF_HOME=$SCRATCH/olcfg       # OpenLoco's HOME — never the player's
    ol.sh start [SAVE.SV5]                 # Xvfb :97 + openloco
    ol.sh do click 784 14                  # one input command
    ol.sh run tours/windows-openloco.txt   # or a script of them
    ol.sh shot out.png                     # the 800x600 game frame
    ol.sh stop

Xvfb needs the Claude Code sandbox off (it blocks the X socket). Two harnesses on
one display draw into each other's game and take each other's clicks: a second
one running at the same time needs its own OLREF_DISPLAY (and OLREF_HOME).

| file | what |
|------|------|
| ol.sh | launcher: scratch config, Xvfb, openloco, shot, stop |
| xt.py | input: move, click, rightclick, press, release, drag, wheel, key, type, wait, shot, state, run |
| untitle.py | Data/title.dat → a save the real OpenLoco loads |
| pair.py | goloco crop, OpenLoco crop, and their differing pixels, side by side |
| match.py | two map views of one save: find the offset, then diff |
| setview/ | `go run ./scripts/olref/setview -tile X,Y [-zoom Z] [-rotation R] in.SV5 out.SV5`: a copy of a save that opens on that view |
| tours/ | the input scripts of past comparisons |

## Coordinates and scripts

Every coordinate is a game-window pixel, and xt.py's commands are the ones of
goloco's GOLOCO_SCRIPT (cmd/goloco/script.go): a script written for one game
mostly runs on the other — goloco's window stack opens, stacks and resizes
windows as OpenLoco's WindowManager does (tours/windowstack.txt; windows-goloco.txt clicks where
windows-openloco.txt does) — though `seed`, `scenario` and `camera` exist
in goloco only.

OpenLoco refuses a window that leaves less than 160 px of desktop around it
(Ui.cpp getSaneWindowedResolution), so the screen is 1024x768 and the window
sits at 112,84 (OLREF_ORIGIN); xt.py adds that and `shot` crops it away.

A window's tabs are at window x+3+31·n, y+15…41; the window moves and grows
when a tab changes its size, so take a shot before aiming at the next tab.

## What ol.sh sets in the scratch openloco.yml

- window 800x600, scale 1;
- `toolbarAutoMenu: false`. 26.08 opens a top-toolbar menu as soon as the
  pointer rests on its button (ToolbarTop.cpp onMouseHover; a config switch in
  26.08, unconditional in the older source we port) — it looks like a stuck
  mouse button and is not: `xt.py state` shows none held. With it off a move
  only hovers and a click opens the default entry; a menu entry is `press` on
  the button, `move` to the entry, `release`. OLREF_AUTOMENU=true keeps the
  release's default. The toolbar's buttons have no tooltips either way; other
  widgets show theirs about two seconds after the pointer stops;
- `edgeScrolling: false`, so a pointer left at the frame's edge can't scroll
  the map between two shots.

Pause (click 688 592) before a series of shots: numbers and graphs stay put.

Two things that stop OpenLoco from starting, both silent on the screen (a
black frame; the reason is in $OLREF_HOME/.olref/openloco.log):

- "Another OpenLoco session has been found running": its single-instance lock
  is the file OpenLoco.lock in the *current directory*. Two harnesses (two
  agents) must start it from different directories — `env -C $SCRATCH/cwd
  ol.sh start …` — and on different displays (OLREF_DISPLAY).
- "OpenLoco data path could not be found": from another directory it doesn't
  find its own language / objects folder; OPENLOCO_DATA_DIR=/usr/share/openloco
  names it.

A process that failed this way keeps running (it waits on its error box) and
`ol.sh stop`'s SIGTERM doesn't end it: `pkill -9 -f "openloco .*$SCRATCH"`.

## The same world in both games

- **Locomotion's title game** — a game with years of history, stations,
  vehicles and ratings. goloco opens Data/title.dat as it is; OpenLoco needs
  `untitle.py title.dat $SCRATCH/title.SV5` first (clears the header's
  isTitleSequence flag). The player there is a new company with nothing of its
  own. tours/windows-*.txt walk both through the Industry, Fund New Industries
  and Town windows on it.
- **OpenLoco's own autosave** — start a scenario in OpenLoco, zoom with the
  wheel, wait for the monthly autosave ($OLREF_HOME/.config/OpenLoco/save/
  autosave, about a minute); `ol.sh start that.SV5` and `goloco that.SV5`
  show the same world from the same place: both open a save on its saved view
  (place, zoom, rotation — goloco since 2026-11-13), so the two frames line up
  pixel for pixel (`match.py goloco.png openloco.png diff.png` finds an offset
  when they don't). `setview` writes a copy of a save that opens anywhere, at
  any zoom and rotation; the rotate key (Enter) then turns both about the
  view's centre — how the four rotations were compared (2026-11-13: the title
  save at thirteen places, zoom 0–2: every rotation lines up at offset 0,0 and
  differs from OpenLoco no more than rotation 0 does — water, smoke, vehicles
  and trees by the seconds the two ran, and some farm-field tiles, which
  differ at rotation 0 too). 2026-11-05: zoom 1 and 2 matched pixel for pixel apart from
  the town labels and things that change with time. OpenLoco can't read every
  .SV5 (DEBUG_AUTO_DUMP.SV5: "Unsupported S5 format").

tours/construction-rail.txt and construction-road.txt are the construction
window's state on such an autosave of Sandbox Settler (flat desert): hover,
rotate, first click, build, bulldozer back to the start and once more, the red
cross, a refused click. One script runs on both games; goloco doesn't take the
save's view, so the two build at different tiles of the same flat ground.

tours/music.txt (the Audio menu, the Jukebox, Music Selection, Options' Audio
tab, on the title game) and music-title.txt (the Jukebox on the title screen)
start both games from the same audio settings — OpenLoco picks its track with
std::rand, so the scratch openloco.yml's audio section names one: `playlist:
custom` and a customJukebox with one track true. goloco reads the same section
from $XDG_CONFIG_HOME/goloco/openloco.yml.

tours/news.txt (the Messages window from the time panel's date, its rows
opened as news, a subject's window, the settings tab) runs on both games on
the title game: it posts the same two speed records in its first days in
both. tours/news-ticker.txt is the ticker — both configs' `news:` section
set to `ticker` for every kind (majorCompany … advice; OpenLoco's
openloco.yml has the section already: change it, a second one is ignored).
Its Town window is the placement check: with the news window up
OpenLoco opens it at (0, 439) — bringToFront swaps a pressed window with the
list's last one (the time panel, or the news window), so the panels come
before the other windows in the list and "below the player panel" is tried
first (ui.WindowStack's order; pkg/ui/news_test.go).
tours/news-rotated.txt is the news window's viewports at rotation 1 and —
turned with the window up — 2.
A message shows for 384 interface updates, some 10 s in OpenLoco, paused or
not: shoot soon after the click. tours/news-goloco.txt is goloco alone — its
script command `news TYPE [subjects] [date DAY]` posts a message — for the
advice panel, a new vehicle and the pre-1945 newspaper, which the title game
doesn't bring up.

For one window: `pair.py out.png goloco.png@X,Y,W,H openloco.png@X,Y,W,H 2`.

## Which source decides

Since 2026-11-13 the binary (openloco-git 09d74ee39) and the source goloco
ports (~/Projects/OpenLoco, moved to 09d74ee) are the same commit: where they
differ from goloco, goloco is behind. goloco was ported from b93e074
(2026-07-06), 161 upstream commits earlier — some of what the binary shows (the
Town window's fourth tab, "Cargo awaiting …" in an industry's statistics, white
station captions, toolbarAutoMenu) is that gap, to port, not a bug. `git -C
~/Projects/OpenLoco log b93e074..09d74ee` lists it. OLREF_BIN runs another
build. Before, the binary was the 26.08 release — older than b93e074 in parts.
