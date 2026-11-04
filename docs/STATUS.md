# goloco — where it stands

A short overview for Joop (2026-11-04). The full queue and every tick are in LOOP.md; what still deviates
from OpenLoco is in docs/deviations.md. Updated when an item moves.

## Works, as OpenLoco does it

- **Loading** the original Locomotion data: objects, scenarios (.SC5) and saved games (.SV5), the title map.
- **Drawing**: OpenLoco's software renderer ported (8-bit, the paint sort, zoom 0–3, snow, bridges, tunnels,
  slopes, water), on the desktop and in the private web build (kiefte.eu/goloco/, redeployed by
  `scripts/redeploy-test-page.sh`).
- **Windows**: the title screen, toolbars, Options (with the language), Town / Station / Vehicle lists, the
  Construction window, New Game (scenario select), Load / Save, tooltips, dropdowns — ported, not improvised.
  The language switch works for those windows; the rest follows as windows are touched (0L).
- **The running game**: the date, the economy (inflation, payments for delivered cargo, running costs, company
  value, the scenario objective), saved vehicles moving faithfully (orders, rail and road pathing, signals,
  loading and unloading, breakdowns), towns growing their roads (bridges, slopes, raising / lowering land).
- **Game speed**: paced as OpenLoco does (40 ticks a second; slow web frames no longer slow the game).
- **A new game from a .SC5**: the landscape generated as OpenLoco does (hills, rivers, terrain, trees, the first
  towns), the player's company made.

## In progress

- **5P — playable vehicles** (top priority, Joop 2026-11-04). The vehicle simulation is done; the player's side is
  being built: the purchase command (CreateVehicle — its parts so far match saved trains exactly: entity slots,
  cars, cargo, colours, sprite sizes, the train's automatic layout), then placing a vehicle, start / stop, giving orders, an end-to-end test,
  and the Build Vehicle and Vehicle windows.
- **Towns** (5.9f3): roads grow; buildings along them, new junctions and road take-over are next.
- **Language switch** (0L): the remaining windows' texts.

## Missing, known

- **Water and lights don't animate** (W: OpenLoco's palette cycling is not ported).
- **Station building** in game is goloco's own simple version (no station record / catchment yet).
- **Industries and misc buildings** on a new map (5.9f4–f5).
- **Construction gap** (5.9g): track and road the player builds don't mark which part of a tile they use, so
  later building can overlap them.
- **CJK languages** need a fallback font (5c).

## Later

- **The open build on GitHub Pages** (O): goloco without any Locomotion data, on OpenGraphics (MIT) — approved,
  parked behind vehicles. Today without the original data it shows only an empty title.
- **The deviations list** (6), worst first.

## Waiting on Joop

- Nothing blocking. Optional: a screenshot from OpenLoco / Locomotion of the river bridge near tile 90,99 on the
  title map, to check the bridge drawing pixel for pixel.
