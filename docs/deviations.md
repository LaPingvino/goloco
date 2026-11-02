# Where goloco deviates from OpenLoco

Inventory for the correction loop (LOOP.md item 3), started 2026-11-01. Source: every marker in cmd/ and
pkg/ (approximat|simplif|fudge|hack|stub|TODO|FIXME|placeholder|not yet implemented|heuristic|for now —
103 lines, cmd/diagnostics excluded) plus the invented heuristics found while reviewing. Upstream paths
are under `~/Projects/OpenLoco/src/OpenLoco/src/` unless noted.

Severity: **H** = gameplay or visible output is wrong; **M** = visible but local; **L** = cosmetic or
internal. Size: S (one tick), M (2–4 ticks), L (a series).

## A. Gameplay — approximations standing in for the real rules

| Where | What goloco does | What upstream does | Sev | Size |
|---|---|---|---|---|
| ~~pkg/world/vehicle_sim.go (whole file)~~ RESOLVED tick 243 | (was: straight-segment SimVehicles) | the ported sim (pkg/sim) runs every vehicle; SimVehicles retired | — | — |
| ~~pkg/game flat fare~~ RESOLVED tick 247 | (was: flat fare per unit) | ported: calculateDeliveredCargoPayment (pkg/sim/economy.go); pkg/game deleted | — | — |
| (not ported) vehicle running costs | Since pkg/game was retired (tick 247) no running cost is charged at all | Vehicles: VehicleHead::updateMonthly / the running-cost expenditure, per vehicle object, inflated | H | M |
| cmd/goloco/main.go:2448 | Company value = player money | Company value from assets (vehicles, infrastructure) — Company.cpp | M | M |
| pkg/world/station.go (PlaceStationOnTile) | Places only the station element; no Station record, name, catchment or cargo (the invented passenger supply was retired tick 243) | GameCommands createTrainStation, StationManager, Station.cpp | H | M |
| pkg/scenario/generate.go:9 | "Simplified approximation" of landscape generation | Map/MapGenerator/ with the game's PRNG (pkg/core/prng.go is already ported, unused) | M | L |
| ~~pkg/game placeholders~~ RESOLVED tick 247 | (was: placeholder movement, dirt tiles) | pkg/game deleted | — | — |

| cmd/goloco/main.go (Update) | One world tick per frame (60/s) | Game tick every 25 ms (40/s), independent of frame rate (OpenLoco.cpp tick loop) | M | S |

## B. Rendering

| Where | What goloco does | What upstream does | Sev | Size |
|---|---|---|---|---|
| pkg/world/world.go:2767 | Vehicles drawn in a post-tile pass ("depth approximation") | Paint/ — vehicles go through the same paint session and sort as everything else (addToPlotList, quadrant sort) | H | M |
| pkg/world/world.go:1590 | Buildings: first variation's parts stacked | Paint/PaintBuilding.cpp — the tile's variation and part heights | M | S |
| pkg/world/world.go:1611 | Buildings under construction skipped | PaintBuilding draws construction stages | M | S |
| pkg/world/world.go:2615 | Industry fields not drawn | Paint/PaintIndustry.cpp via IndustryObject | M | S |
| pkg/world/world.go:1541 | Viewport rotation fixed at 0 | Rotation 0–3 throughout the paint code | M | M |
| pkg/world/track_sprites.go:20–26 | Large curves, small-curve slopes and pieces 20–25 TODO | Paint/PaintTrack.cpp — every track piece | H | M |
| pkg/objects/vehicle.go:433 | Second half of rotation frames not mirrored | Vehicle sprite selection in VehicleDraw.cpp | M | S |
| pkg/ui/title_menu.go:199–204 | Globe drawn as a square "for now" | The animated globe sprites in the title menu window | L | S |
| pkg/ui/toolbar.go:94, 219 | Placeholder positions and colours | Toolbar layout and colours from the InterfaceSkin object | L | S |
| pkg/graphics/drawing_context.go:199 | Palette "heuristic" | Palette maps (Gfx/PaletteMap) | M | S |

| pkg/world/world.go paintWater + pass-2 blend | FIXED 2026-11-01 (pkg/palblend, zoom 0; zoom > 0 still the old opaque remap). Was: water +35 remapped to an opaque flat colour; blend pass lerps the terrain toward one tint | Blended image without primary = mask: each covered destination pixel is mapped through the palette map (water: image +41) — Graphics/DrawSprite.cpp getDrawBlendOp | H | S |

## C. Text and UI

| pkg/world/lists.go TownList | FIXED 2026-11-01: town names from the GameState user strings (StringIds 0x8000+). Was: placeholders "Town N" | Town::name is a user StringId (the town name generator allocates user strings) | M | M |

| Where | What goloco does | What upstream does | Sev | Size |
|---|---|---|---|---|
| pkg/ui/g1font.go:45–52 | NOT A DEVIATION (checked 2026-11-01): the {-1,-1,-1,1} "fudge" is upstream's own widthFudge table (Graphics/Gfx.cpp initialiseCharacterWidths) | — | — | — |
| pkg/graphics/drawing_context.go:282–297 | DrawStringCentered measures len×7 px | String width from the font (getStringWidth) | M | S |
| pkg/ui/button_widget.go:25 | Text height "approximate" | Widget text placement as in Ui/Widget.cpp | L | S |
| pkg/ui/input.go:166–176 | No text input | Text input windows (TextInputWindow.cpp) | M | M |
| pkg/ui/window_manager.go:42–174, pkg/ui/types.go:248 | Stub functions (ProcessEvents, UpdateAnimations, ViewportsUpdatePosition, GfxRender, Window.Draw; UiWidth/Height fixed 800×600) — none is called | — dead (see D) | L | S |

## D. Dead scaffolding (imported by nothing in cmd/ or live packages)

- pkg/loco/objects, pkg/loco/vehicles, pkg/loco/s5, pkg/loco/economy, pkg/loco/worldmap (only imported by the
  dead pkg/loco/vehicles) — type stubs from the pruned conversion pipeline; the live code is pkg/objects,
  pkg/scenario, pkg/world. Their "stub" markers warn about code nothing runs.
- pkg/tests (stub_integration_tests.go).
- The uncalled stub functions in pkg/ui/window_manager.go and pkg/ui/types.go.
- Keep: pkg/core (Prng, golden vectors from upstream) and pkg/olmath — real ports, needed by A.

Removing dead code removes what its markers warn about, so it is allowed under the loop rule; do it in its own
commit, after `go build ./... && go vet ./...` confirm nothing used it.

## E. Not implemented at all

Tutorial, Scenario Editor, Terraform, Ports/Airports, Build Vehicles (port queued as LOOP item 5b), per-frame
audio (pkg/audio/audio.go:158–174), music while the data lacks it.

## Fix order (feeds LOOP.md)

1. D — remove dead scaffolding (S).
2. A vehicle movement (LOOP item 5) — the base for economy, stations and the Build Vehicle window.
3. B vehicle depth (draw vehicles in the paint sort) — with the movement port.
4. B track pieces 8–25; C font metrics (g1font fudge, len×7) — small, very visible.
5. A economy (payment, running costs, company value), station cargo.
6. The rest of B, then E.
