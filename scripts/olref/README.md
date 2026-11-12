# Driving the real OpenLoco for reference shots

OpenLoco (the AUR release, /usr/bin/openloco — 26.08) runs headless under Xvfb
and is driven through X11's XTest extension (ctypes on libXtst — no xdotool).
No game data and no screenshots here: OpenLoco reads the player's own
Locomotion folder, shots go to a scratch directory.

    export OLREF_HOME=$SCRATCH/olcfg       # OpenLoco's HOME — never the player's
    ol.sh start [SAVE.SV5]                 # Xvfb :97 + openloco
    ol.sh do click 784 14                  # one input command
    ol.sh run tours/windows-openloco.txt   # or a script of them
    ol.sh shot out.png                     # the 800x600 game frame
    ol.sh stop

Xvfb needs the Claude Code sandbox off (it blocks the X socket).

| file | what |
|------|------|
| ol.sh | launcher: scratch config, Xvfb, openloco, shot, stop |
| xt.py | input: move, click, rightclick, press, release, drag, wheel, key, type, wait, shot, state, run |
| untitle.py | Data/title.dat → a save the real OpenLoco loads |
| pair.py | goloco crop, OpenLoco crop, and their differing pixels, side by side |
| match.py | two map views of one save: find the offset, then diff |
| tours/ | the input scripts of past comparisons |

## Coordinates and scripts

Every coordinate is a game-window pixel, and xt.py's commands are the ones of
goloco's GOLOCO_SCRIPT (cmd/goloco/script.go): a script written for one game
mostly runs on the other. Not quite always — the two games open their windows
in different places (see the tours), and `seed`, `scenario` and `camera` exist
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
  (GOLOCO_ZOOM=N; goloco doesn't take the saved zoom) show the same world from
  nearly the same place and `match.py goloco.png openloco.png diff.png` finds
  the exact offset. 2026-11-05: zoom 1 and 2 matched pixel for pixel apart from
  the town labels and things that change with time. OpenLoco can't read every
  .SV5 (DEBUG_AUTO_DUMP.SV5: "Unsupported S5 format").

For one window: `pair.py out.png goloco.png@X,Y,W,H openloco.png@X,Y,W,H 2`.

## Which source decides

The binary is the 26.08 release (2026-08-13); the source goloco ports,
~/Projects/OpenLoco, is at b93e074 (2026-07-06) — older. Where the binary and
goloco differ, read that source before calling it a goloco defect (LOOP.md):
some of what the binary shows (the Town window's fourth tab, "Cargo awaiting
…" in an industry's statistics, white station captions) arrived after it. The
release's own source unpacks from ~/.cache/yay/openloco/openloco-26.08.tar.gz.
