# Company AI: the interface, from a study of the original

Written 2026-11-19 for PORT-PLANS F step 1. Joop's decision: goloco gets its own Company AI on a clean
interface, not a port of OpenLoco's. OpenLoco's AI (~/Projects/OpenLoco at 09d74ee) is studied here only to find
the end points such an interface needs: what an AI reads (observations), what it does (actions), what the
original does that a player cannot (hidden powers), and what it keeps in the company record (state that goloco
must carry so OpenLoco saves still round-trip).

Sources read: `src/OpenLoco/src/World/CompanyAi/CompanyAi.cpp` (7889 lines), `CompanyAiPathfinding.cpp` (2665),
`CompanyAiPlaceVehicle.cpp` (659), `include/OpenLoco/World/CompanyAi/*.h`, the three AI-only game commands in
`src/GameCommands/CompanyAi/`, and the engine code the AI relies on (CompanyManager, GameCommands,
TileClearance, S5Company). Paths below are relative to `~/Projects/OpenLoco/src/OpenLoco/`; `CompanyAi.cpp`,
`Pathfinding.cpp` and `PlaceVehicle.cpp` without a directory are the three files in `src/World/CompanyAi/`.
goloco paths are relative to the repo root.

Marks: **(unsure)** means not verified to the end. **unkN** names are upstream's; where one appears, it is
described by what the code does with it, not by a guessed meaning.

## 0. How the original AI works, in one page

- **Cadence.** `CompanyManager::tick` (src/World/CompanyManager.cpp:223-247, called from
  src/Scenes/GameScene.cpp:237 after EffectsManager, before AnimationManager) runs `aiThink` for the company
  whose id is `scenarioTicks & 0x0F`. Each AI company thinks once every 16 ticks. Each think runs one step
  of a state machine, then one step of the vehicle-placement machine (CompanyAi.cpp:7694-7774). Separately,
  `Company::updateDailyLogic` calls `setAiObservation` once a day (src/World/Company.cpp:147), and
  `Company::updateQuarterly` updates each thought's income, age and running cost (Company.cpp:497-517).
- **Thoughts.** A company holds up to 60 `AiThought`s (CompanyAi.h:111-153). A thought is one service:
  one of 20 types (CompanyAi.cpp:123-198; rail, tram, road, air or water; town-only, town-to-town,
  industry-to-industry or industry-to-town), with 2 or 4 stations, a cargo, a track/road object, a vehicle
  consist and up to 8 vehicles.
- **States** (`AiThinkState`, CompanyAi.h:10-50; dispatch table CompanyAi.cpp:7679-7691):
  - 0: count to 672 thinks, then review (CompanyAi.cpp:567-613).
  - 1: walk the existing thoughts, sell bad ones, re-equip ones whose engines are obsolete or unreliable
    (1753-1791).
  - 2: plan a new thought in 14 sub-steps (4132-4154): pick type and destinations, reject duplicates, apply
    the playstyle, site stations, choose track, vehicles, station and signal objects, estimate costs and
    revenue, and check the money.
  - 3: build the plan as **aiAllocated** (invisible, free) stations, track and signals; pathfind between
    them; decide whether to go ahead (5702-5717).
  - 4: convert the allocation to real assets (paying), buy the vehicles, queue them for placement
    (6361-6375).
  - 6: undo a failed plan (7338-7350).
  - 7: sell a thought's land assets (7372-7382).
  - 8: refresh a thought's vehicles (7561-7574).
  - 10: end the company (7668-7676).
- **Randomness** comes from `gPrng1`, the game's main stream (see O22), so the original AI changes the whole
  game's random sequence.

## 1. Actions: every game command the AI issues

38 entries. Column "goloco" names the `pkg/sim` equivalent, or **missing**. Flags: `apply`, `aiAllocated`
(AA), `noPayment` (NP), `flag_7` (F7) are GameCommands::Flags (include/OpenLoco/GameCommands/GameCommands.h:37-49).
"Direct" means the command function is called without `doCommand`, so no payment and no nest-level
handling.

### Building track and road

| # | Upstream command | Where the AI calls it | goloco |
|---|---|---|---|
| A1 | `createTrack` (TrackPlacementArgs) | **Query, direct** in the pathfinder's look-ahead, AA\|NP without apply (Pathfinding.cpp:395-412). Build of the plan, AA\|NP\|apply, first without a bridge, then with one (Pathfinding.cpp:876-903). | `Map.createTrack` (pkg/sim/createtrack.go:260) is unexported. Its only caller is `State.ConstructTrack` (pkg/sim/construction.go:205), which takes the construction window's head, not explicit args. **Needs an exported command.** |
| A2 | `createRoad` (RoadPlacementArgs) | **Query, direct**, AA\|NP, with a retry on the alternate road object or bridge from the legacy return state (Pathfinding.cpp:552-578). Build of the plan, AA\|NP\|apply, in three tries: no bridge, bridge, then the alternate object or bridge (Pathfinding.cpp:1123-1161). Conversion of AA road to real road, apply\|F7, retried with NP (CompanyAi.cpp:5891-5905). | `Map.createRoad` (pkg/sim/createroad.go:340) is unexported. Its callers are `State.ConstructRoad` (construction.go:236) and town growth. **Needs an exported command.** |
| A3 | `removeTrack` | Undo of a plan piece, AA\|NP\|apply (Pathfinding.cpp:742-748). Sell-off walk, apply (CompanyAi.cpp:6916-6922). Undo of stray AA track, AA\|NP (7078-7084, 7249-7255), and NP alone (7266-7272). End of company, NP plus AA where the piece is AA (7601-7608). **Direct**, apply, station tiles (6689-6698, `removeTrainStationAndTrack`). | `Map.RemoveTrack` (pkg/sim/removetrack.go:84) |
| A4 | `removeRoad` | Undo of a plan piece (Pathfinding.cpp:1011-1017). Conversion step that removes the AA road before rebuilding it real (CompanyAi.cpp:5878-5884). Sell-off (6672-6678, 6798-6804). Undo (7054-7060, 7196-7202). End of company (7619-7626). | `Map.RemoveRoad` (pkg/sim/removeroad.go:89) |
| A5 | `createTrackMod` | Query for the cost (flags none), then apply, then apply again with NP if that failed (CompanyAi.cpp:7406-7419). Called from 7445 (cost estimate) and 7484-7497 (upgrade). | `Map.CreateTrackMod` (pkg/sim/trackmods.go:244) |
| A6 | `createRoadMod` | Same pattern (CompanyAi.cpp:7389-7402). | `Map.CreateRoadMod` (pkg/sim/roadmods.go:202) |

### Stations, signals and AI-only composites

| # | Upstream command | Where the AI calls it | goloco |
|---|---|---|---|
| A7 | `aiCreateTrackAndStation` (AI only; src/GameCommands/CompanyAi/AiCreateTrackAndStation.cpp:170) | Station plus its stub track in one command, AA\|NP\|apply (CompanyAi.cpp:4799-4825). | **missing** (and not wanted; see §6) |
| A8 | `aiCreateRoadAndStation` (AI only; AiCreateRoadAndStation.cpp:122) | Road station with its road, AA\|NP\|apply (CompanyAi.cpp:4648-4724). | **missing** (not wanted) |
| A9 | `createRoadStation` | Station on an existing road, AA\|NP\|apply (CompanyAi.cpp:4738-4745). Conversion to real, apply\|F7, retried with NP (5915-5925). | `Map.CreateRoadStation` (pkg/sim/createroadstation.go:49). PARTIAL: the AI's placement without a road. |
| A10 | `removeRoadStation` | Conversion, AA\|NP (CompanyAi.cpp:5866-5872). **Direct**, apply, both directions, in the sell-off (6648-6663). Undo, AA\|NP (7150-7162). | `Map.RemoveRoadStation` (pkg/sim/removeroadstation.go:51) |
| A11 | `createTrainStation` | Only inside A7 and A14. The AI never issues it on its own (CompanyAi.cpp includes RemoveTrainStation.h but calls neither). | `Map.CreateTrainStation` (pkg/sim/createtrainstation.go:51). PARTIAL: the AI's query without a track. |
| A12 | `createSignal` | One-sided signals spaced along the planned line, AA\|NP\|apply (CompanyAi.cpp:5380-5390, from `placeAiAllocatedSignalsEvenlySpaced` 5333-5396). | `Map.CreateSignal` (pkg/sim/createsignal.go:144) |
| A13 | `createAirport` | Plan, AA\|NP\|apply (CompanyAi.cpp:4361-4365). Conversion: apply, retried with NP; the new station id is read from the legacy return state (5987-6000). | `Map.CreateAirport` (pkg/sim/airportcmd.go:210). PARTIAL: the AA surface flag. |
| A14 | `aiTrackReplacement` (AI only; AiTrackReplacement.cpp:428) | Converts AA track (and its AA station and signal elements) to real, paying the replacement cost. Apply, retried with NP: station tiles (CompanyAi.cpp:6049-6060) and line (6294-6304). **Direct** with flags none, as a cost query in `measureTrackRoute` (Pathfinding.cpp:2215-2223). | **missing** (not wanted) |
| A15 | `removeAirport` | Review: removes own airports no live thought uses, apply (CompanyAi.cpp:513-516). Conversion, AA\|NP (5982-5984). Undo (7296-7298). | `Map.RemoveAirport` (pkg/sim/airportcmd.go:387) |
| A16 | `createPort` | Plan (CompanyAi.cpp:4627-4631). Conversion (6013-6025). | `Map.CreatePort` (pkg/sim/port.go:239). PARTIAL: the AA surface flag. |
| A17 | `removePort` | As A15 (CompanyAi.cpp:556-559, 6008-6010, 7302-7304). | `Map.RemovePort` (pkg/sim/port.go:335) |

### Vehicles

| # | Upstream command | Where the AI calls it | goloco |
|---|---|---|---|
| A18 | `createVehicle` (new head, then each further car with `vehicleId = head`) | apply; on failure again with **NP** ("try create a free vehicle!", CompanyAi.cpp:259-270). | `Map.CreateVehicle` (pkg/sim/createvehiclecmd.go:26) plus `Map.AddCarToVehicle` (:200). PARTIAL: payment, so it needs `DoCommandFor`. |
| A19 | `vehicleRefit` | When no car carries the thought's cargo and the cargo allows refit (CompanyAi.cpp:330-333). | `Map.VehicleRefit` (pkg/sim/refit.go:11) |
| A20 | `vehicleOrderInsert` | Stop-at for each station, plus a wait-for-cargo order after the first station on industry thoughts (CompanyAi.cpp:348-379). | `Map.VehicleOrderInsert` (pkg/sim/ordercmds.go:67) |
| A21 | `vehicleOrderSkip` | Twice, on multi-vehicle industry rail thoughts (CompanyAi.cpp:392-395). | `Map.VehicleOrderSkip` (pkg/sim/ordercmds.go:136) |
| A22 | `vehiclePlace` (rail/road) | Run as the **neutral company** (PlaceVehicle.cpp:329-337; placement args 92-140, 214-260, 520-560). | `Map.VehiclePlace` (pkg/sim/vehicleplace.go:22) |
| A23 | `vehiclePlaceWater` | As neutral (PlaceVehicle.cpp:388-398). | `Map.VehiclePlaceWater` (pkg/sim/water.go:42) |
| A24 | `vehiclePlaceAir` | As neutral, at the first terminal node whose clear-edges are free (PlaceVehicle.cpp:445-490). | `Map.VehiclePlaceAir` (pkg/sim/airplace.go:55) |
| A25 | `vehicleChangeRunningMode` | Start, as neutral, after a placement (PlaceVehicle.cpp:76-90). Stop, before a pickup or sale (CompanyAi.cpp:6399-6402). | `Map.VehicleChangeRunningMode` (pkg/sim/runningmode.go:27) |
| A26 | `vehiclePickup` | Before selling a rail or road vehicle (CompanyAi.cpp:6412-6414). | `Map.VehiclePickup` (pkg/sim/liftup.go:109) |
| A27 | `vehiclePickupAir` | Before selling an aircraft (CompanyAi.cpp:6419-6421). | `Map.VehiclePickupAir` (pkg/sim/airplace.go:140) |
| A28 | `vehiclePickupWater` | Before selling a ship (CompanyAi.cpp:6426-6428). | `Map.VehiclePickupWater` (pkg/sim/water.go:126) |
| A29 | `vehicleSell` | Selling a thought's vehicles one per think (CompanyAi.cpp:6439-6447). A failed placement, run as **neutral** (PlaceVehicle.cpp:43-73). | `Map.VehicleSell` (pkg/sim/vehiclesell.go:18) |
| A30 | `vehicleSell`, issued by the **engine**, not the AI | An AI vehicle crashed or stuck past `kAiSellCrashedVehicleTimeout` is sold as its owner, and the thought is fixed up (src/Vehicles/VehicleHead.cpp:342-361). | goloco counts `CrashedTimeout` (pkg/sim/date.go:191) but does not sell. Keep this as an engine rule or drop it (see H16). |

### Company

| # | Upstream command | Where the AI calls it | goloco |
|---|---|---|---|
| A31 | `buildCompanyHeadquarters` | Once it has started (`aiHasStarted`) and has no HQ, every think: a random tile within ±16 tiles of the newest thought's first destination (CompanyAi.cpp:7724-7773). | `Map.BuildCompanyHeadquarters` (pkg/sim/headquarters.go:71) |
| A32 | `removeCompanyHeadquarters` | End of company, NP (CompanyAi.cpp:7656-7663). | `Map.RemoveCompanyHeadquarters` (pkg/sim/headquarters.go:34) |

### What the AI does without a command (for completeness)

| # | What | Where | goloco |
|---|---|---|---|
| A33 | **Loans: no `changeLoan` command.** Borrowing happens inside `ensureCompanyFunding`, which raises a non-player company's loan in 1000 steps up to the inflation-adjusted maximum. It is called from `doCommand` for every paid command (src/GameCommands/GameCommands.cpp:358-369) and directly by the AI's money checks (CompanyAi.cpp:4100, 5644). Repayment is the engine's monthly auto-repay for non-players (src/World/Company.cpp:710-720). | src/World/CompanyManager.cpp:1149-1203 | Both ported: pkg/sim/companyvalue.go:97-127 (funding) and :182-187 (repay). `ChangeLoan` exists (pkg/sim/changeloan.go:13) for an AI that borrows as a player does. |
| A34 | **Names: no rename commands.** Company and owner names come from the competitor object in `createCompany` (src/World/CompanyManager.cpp:546+). | | `RenameCompany` / `RenameCompanyOwner` exist (pkg/sim/companyrename.go:14, 48). The competitor branch of createCompany is PARTIAL (pkg/sim/newgame.go:13). |
| A35 | Company removal, `CompanyManager::aiDestroy`, when it never started within 42 days (CompanyAi.cpp:1790) or after the end-of-company sweep (7674). | src/World/CompanyManager.cpp:1436 | **missing** |
| A36 | Company creation, `produceCompanies` every 192 ticks (src/World/CompanyManager.cpp:240-245). | | **missing** |
| A37 | News: `MessageManager::post(newCompany)` when the first thought converts (CompanyAi.cpp:5696). | | the news system exists (pkg/sim/news.go); nothing posts this |
| A38 | Owner emotions (`companyEmotionEvent`: disgusted 1772, happy 6354, thinking 6371) and observations (`companySetObservation` 5826, 7795, 7833; also AiTrackReplacement.cpp:197). | | `companyEmotionEvent`, `setAiObservation` (pkg/sim/companydaily.go:39, 110) |

Not issued at all: terraforming, clearing land, trees, walls, building removal (only implicitly, through the
clearance inside A1-A16), order delete or reorder, colour, rename, vehicle clone, rearrange, reverse, speed,
local/express.

## 2. Observations: everything the AI reads

23 entries.

| # | What | Upstream reads it at | goloco exposes it? |
|---|---|---|---|
| O1 | **Towns:** list, `x,y`, `populationCapacity`. Thresholds 450-1200, distance windows. | CompanyAi.cpp:1805-1830, 1831-1876, 1878-1954 (plus ≥20 water tiles nearby), 2411-2472. Home town of the playstyle: 2963-2990. | yes: `Map.Towns`, `TownByID` (pkg/sim/stationcargo.go:83) |
| O2 | **What a town's buildings consume:** constructed, non-misc buildings within ±5 tiles; each building object's `consumedCargoType/Qty`; a cargo counts at a score of 32 or more. | `getTownRequiredCargoTypes` CompanyAi.cpp:2156-2216 (for mixed thoughts) | the data is there (tiles, building objects); no helper |
| O3 | **Industries:** list, `x,y`, the `closingDown` flag, `producedCargoQuantityPreviousMonth[2]` (≥150) or `dailyProduction[2]` (≥8), the object's produced and required cargo, `town`. | CompanyAi.cpp:1956-2041, 2043-2154, 2218-2303, 2305-2409, 2411-2472; 5693 (town) | yes: `Map.Industries` (`ProducedPrevMonth`, `DailyProduction`, `Flags`, `Town`); `industryByID` unexported (pkg/sim/unload.go:80) |
| O4 | **Most common building cargo** (the passenger cargo, in practice). | `IndustryManager::getMostCommonBuildingCargoType` at CompanyAi.cpp:2220, 2307, 2431, 2529 and others | `mostCommonBuildingCargoType` unexported (pkg/sim/industrynew.go:25) |
| O5 | **Water near a point:** a count of water tiles. | `TileManager::countNearbyWaterTiles` CompanyAi.cpp:1887, 1918, 2071, 2120, 2323, 2383 | `countNearbyWaterTiles` unexported (pkg/sim/towngen.go:19) |
| O6 | **Surface:** base height, corner height, slope, water level. | Track choice by the height difference between destinations (3493-3545). Airport footprint max height (4300-4320). Station footprint corner height and water (4880-4893, `getSurfaceCornerHeight`). Water midpoint (around 1920). HQ placement (7750-7762). | yes: `Map.Tiles`; `surfaceCornerHeight` unexported (pkg/sim/terraform.go:22) |
| O7 | **Tile elements:** buildings (count; object clear cost; HQ or indestructible), trees (clear cost), track and road (owner, aiAllocated, ghost, mods, bridge, station). | `isSuitableForStation` (fewer than 4 buildings in 5×5) 3139-3157. `estimateStationClearageCosts` 3906-3948. Existing mods at station 0 (1590-1650, 1700-1735). `nearbyTrackElementsStats` 5495-5544. Removal queries 6475-6644. Placement checks PlaceVehicle.cpp:140-325. | yes: `Map.Tiles` elements |
| O8 | **Catchment of a rectangle:** cargo accepted and produced. | `calcAcceptedCargoAi` (src/World/Station.cpp:437) at CompanyAi.cpp:4325, 4591, 4922 | partly: the ghost variants for train stations, airports and docks (pkg/sim/stationaccept.go:122, 177, 190). No arbitrary-rectangle form. |
| O9 | **Stations:** own stations, flags (air or water mode), `x,y`, `stationTiles`, `cargoStats[c].quantity` and `isAccepted()`, free station slots (fewer than 8 means no). | 470-560, 1362-1394 (`getUntransportedQuantity`), 4208-4260, 4410-4460, 5649 | yes: station records with CargoStats; slot count derivable |
| O10 | **Airport movement edges occupied.** | PlaceVehicle.cpp:445-470 | yes: `AirportOccupiedEdges` (pkg/sim/airplace.go:95) |
| O11 | **Availability:** `company.unlockedVehicles`; the company's rail tracks and roads; airports and docks by year; stations, signals and bridges compatible with a track or road. | unlockedVehicles 677, 736, 826, 923, 999, 1090, 1197. `companyGetAvailableRailTracks/Roads` 3510, 3550, 3592. `getAvailableAirports/Docks` 1423, 1448. `getAvailableCompatibleStations` 1471, 1523. Signals 1542. Bridges 4169-4173. | yes: `UnlockedVehicles`; `Map.CompanyAvailableRoads/RailTracks` (pkg/sim/roadflags.go:193, 199); `ObjectManager.AvailableAirports/Docks/CompatibleStations/Signals/Bridges` (pkg/objects/loader.go:1402-1664) |
| O12 | **Object data:** vehicle (mode, trackType, speed, power, designed, obsolete, maxCargo and categories, flags refittable / rackRail / topAndTail / centre / mustHavePair / helicopter, length, cost and run-cost factors, required track extras); track and road (piece traits, curveSpeed or maxSpeed, flags, mods, build and tunnel cost); station, airport, dock and signal (designedYear, costs, flags); bridge (maxHeight, maxSpeed, disabledTrackCfg); cargo (refit flag, transfer time); competitor (intelligence, aggressiveness, competitiveness); building and tree clear costs. | throughout: 640-1222, 1396-1560, 3493-3646, 3695-3891, 4158-4200, 3057, 4092, 6088 | yes: pkg/objects (competitor.go:21 has all three traits) |
| O13 | **Can two vehicle objects couple?** | `Vehicles::canVehiclesCouple` 902, 975, 1072, 1164, 1317 | `canVehiclesCouple` unexported (pkg/sim/createvehiclecmd.go:274) |
| O14 | **Economy formulas:** inflation-adjusted cost; delivered-cargo payment; units after a refit; TrackData cost factors and weightings. | `getInflationAdjustedCost` throughout; `calculateDeliveredCargoPayment` 4047; `getNumUnitsForCargo` 4039 | yes: `Map.InflationAdjustedCost`, `Map.DeliveredCargoPayment` (pkg/sim/economy.go:11, 21), `NumUnitsForCargo` (pkg/sim/refit.go:56), generated track data tables |
| O15 | **Date:** current year (station length, obsolescence), current day (start grace of 42 days, vehicle age). | 1578, 3102-3124; 1784; PlaceVehicle.cpp:50, 515 | yes |
| O16 | **Own company:** `challengeFlags` (bankrupt, aiHasStarted, and the two flags below), `competitorId`, `aiPlaystyleFlags/TownId`, `headquartersX`, `transportTypeCount`, `startedDate`. **Cash is never read directly.** Affordability is always `ensureCompanyFunding`, which borrows as it answers (4100, 5644). | 246, 619, 1784, 2901-3062, 7716, 7804 | mostly: `Companies[i]`, `CompanyFaces[i].ChallengeFlags`; the playstyle is in the raw record only |
| O17 | **Own vehicles:** head on or off the map (`tileX`), mode, trackType, breakdown flags, `aiThoughtId`, `aiPlacement*`; cars' objectId, `reliability` (< 0x1900), `creationDay`, `refundCost`. | 401-443, 1570-1580, 6384-6450, 7424-7442; PlaceVehicle.cpp:43-60, 495-560 | yes: `Map.Entities`; the AI fields in the raw head (pkg/scenario/s5entities.go:757) |
| O18 | **A service's results:** the thought's `income` (credited per delivery by the engine, src/Vehicles/VehicleHead.cpp:3303-3306), `previousIncome`, `totalRunningCost` (quarterly). It sells when the previous quarter's income is below 3⅜ × the running cost after 3 quarters. **It does not read vehicle profit.** | 616-628 | yes: pkg/sim/unload.go:225-229, pkg/sim/companydaily.go:195-230 |
| O19 | **Other companies' thoughts:** cargo, destinations, `previousIncome` against `totalRunningCost`, used to avoid competing routes. | `getSimilarThoughtsInAllCompanies` 2786-2866 (also 3057-3060) | partly kept (`ThoughtType`, `ThoughtDestinationA`, incomes); a hidden power, see H11 |
| O20 | **Game-state flags:** `roadObjectIdIsAnyRoadTypeCompatible`, `currentDefaultLevelCrossingType`, max loan (inside the funding check). | 1621-1622; PlaceVehicle.cpp:117, 196; Pathfinding.cpp:1386 | yes: `Map.RoadAnyTypeCompatible` (pkg/sim/map.go:57) |
| O21 | **Track and road connectivity,** including AA pieces: `getTrackConnectionsAi`, `getRoadConnectionsAiAllocated`, `getTrackConnectionEnd`, `getRoadConnections`. | 5339, 5453, 6210, 6263, 6779-6963; Pathfinding.cpp:758, 1031, 2247, 2338 | `Map.TrackConnections`, `Map.RoadConnections`, `TrackConnectionEnd`, `RoadConnectionEnd` (pkg/sim/connection.go:24-200) skip AA elements. No AA variant (not needed without reservations). |
| O22 | **Randomness:** `gPrng1`. Thought type (2491), vehicle-score jitter (821, 909, 983), station directions and offsets (3176, 3298, 4295, 4503, 4840), HQ tile (7745), observed vehicle (7806), backup placements (PlaceVehicle.cpp:94, 216). | as listed | `Map.Rng1` (pkg/sim/map.go:59). Our AI must **not** draw from it (see §6). |
| O23 | **Track and road geometry tables:** `getTrackMiscData` (weighting, cost factor, flags), `getUnkTrack/getUnkRoad` (piece end offsets and rotations), `getTrackPiece/getRoadPiece`. | Pathfinding.cpp throughout; CompanyAi.cpp:5333-5396 | yes: the generated tables (pkg/sim/track_data_gen.go, road_data_gen.go, `TrackCoordinates`) |

## 3. Hidden powers: what the original does that a player can't

18 entries. Each one should be **ruled out** for our AIs or **made an explicit, configurable rule** that
applies to every AI the same way.

| # | Power | Where |
|---|---|---|
| H1 | **Free, invisible reservations (aiAllocated).** All planning (stations, track, road, signals, airports, docks) is placed with AA\|NP: it costs nothing, is not drawn (unless `showAiPlanningAsGhosts`), does not remove trees or buildings, and **blocks players** with "another company is about to build here". OpenLoco's own change lets a player build over a plan that is not yet converting; this sets the AI's clearance flag (`challengeFlags` unk1), and the AI then undoes the plan. Once converting (unk2), the plan blocks. | GameCommands.h:46 ("ai can place down invisible ghosts ... that blocks players"); src/Map/TileClearance.cpp:150-175, 229; CompanyAi.cpp:4365, 4631, 4692, 4825, 5390; Pathfinding.cpp:899-903, 1142-1161 |
| H2 | **Build for free when broke.** Every paying step is retried with `noPayment` on failure: vehicles (CompanyAi.cpp:262-266), conversion of road, road station, airport, port and track (5902-5905, 5922-5925, 5992-5994, 6018-6020, 6058-6060, 6301-6304), and mods (7397-7402, 7414-7419). Upstream's own comment: "the AI is allowed to build for free if it can't afford it" (5937-5938). | as listed |
| H3 | **Automatic loans** inside every paid command: non-players borrow in 1000 steps instead of being refused (src/World/CompanyManager.cpp:1180-1200). The AI's money checks call it directly (CompanyAi.cpp:4100, 5644), so asking "can I afford it?" takes out the loan. **goloco has ported this** (pkg/sim/companyvalue.go:111-125), so any AI that runs commands through `DoCommandFor` gets it today. | |
| H4 | **Automatic monthly repayment** for non-players (src/World/Company.cpp:710-720; goloco pkg/sim/companyvalue.go:182-187). This is a direct cash and loan write, not a command. | |
| H5 | **Acting as the neutral company.** Placement, start and failed-placement sale run with `setUpdatingCompanyId(neutral)` ("// Why?", PlaceVehicle.cpp:68, 85, 333, 394, 480). `checkCompanyCompatibilityCommon` passes everything for neutral (src/GameCommands/GameCommands.cpp:515-528), so the ownership checks on track, road and vehicle are skipped. A side effect that cuts the other way: `applyPaymentToCompany(neutral)` finds no company (src/World/CompanyManager.cpp:168-176) and returns, so the **refund of a sale made as neutral is lost**. | as listed |
| H6 | **Commands called directly, outside `doCommand`:** no payment, no refund, no funding check. Station and track removal in the sell-off (CompanyAi.cpp:6656-6663, 6696-6698); the pathfinder's queries (Pathfinding.cpp:412, 566, 578, 2223). The queries are harmless (no apply); the removals skip the refund (**(unsure)** whether they skip a cost too). | |
| H7 | **AI-only composite commands:** A7, A8, A14. A14's conversion cost sets the bridge height factor to 1 (AiTrackReplacement.cpp:401, "Why 1???"). The pathfinder's road replacement cost uses 0 (Pathfinding.cpp:1404, "probably a bug"), so the estimate is too low. **(unsure)** whether the paid conversion in A14 also undercharges tall bridges compared with a player's createTrack. The code suggests so: the cost is recomputed rather than taken from createTrack. | |
| H8 | **Direct writes to vehicles:** `aiThoughtId` (CompanyAi.cpp:310); `aiPlacementPos/BaseZ/TaD` (417-437, PlaceVehicle.cpp:537); `breakdownFlags` breakdownPending set or cleared as a "place me" marker (439, 6396); `brokenDown` set after placement ("// Why?", PlaceVehicle.cpp:78). | |
| H9 | **Direct writes to the company:** the clearance and converting flags in `challengeFlags` (2762, 5683), `aiHasStarted` (5688), and the AI state. | |
| H10 | **Engine actions for the AI:** posting the "new company" news (5696); deleting itself (`aiDestroy`, 1790, 7674); emotions and observations (A38). | |
| H11 | **Reading other companies' private plans** (O19): every competitor's thoughts, including their income against running cost. A player sees only public figures (the company list and company windows). | |
| H12 | **AI-only command arguments** (`unkFlags`) that change the rules: track bit 2 refuses junctions (src/GameCommands/Track/CreateTrack.cpp:85-89); road bits 4 and 5 change how it treats existing roads under AA (src/GameCommands/Road/CreateRoad.cpp:238-253, 283). These restrict rather than empower, but they are rules a player's command never gets. | Pathfinding.cpp:361-389, 552-560 |
| H13 | **`flag_7`** on conversions (CompanyAi.cpp:5902, 5922). It is "ai only?" in the header (GameCommands.h:49). In building removal it skips the indestructible check (src/GameCommands/Buildings/RemoveBuilding.cpp:61). **(unsure)** of its full effect in createRoad and createRoadStation. | |
| H14 | **Engine branches that favour or limit the AI** (keep them as named rules, not hidden ones): at most 2 AI airports per town (src/GameCommands/Airports/CreateAirport.cpp:323-330); 500 AI vehicles in total (src/GameCommands/Vehicles/CreateVehicle.cpp:41, 48-70; goloco has it, pkg/sim/createvehiclecmd.go:17); separate forbidden-vehicle masks (src/Objects/ObjectManager.cpp:1050; goloco pkg/sim/yearly.go); unused town-type roads of non-players removed by the tile loop (src/Map/RoadElement.cpp:29-117; goloco pkg/sim/roadtick.go); stuck or crashed vehicles sold at once only for players (src/GameCommands/Vehicles/VehicleChangeRunningMode.cpp:208). | |
| H15 | **No emotion or "thinking" feedback for AI construction** (CreateTrack.cpp:518, CreateRoad.cpp:810, CreateRoadStation.cpp:591 are player-only). This is cosmetic. | |
| H16 | **Auto-sale of crashed AI vehicles** by the engine (A30). | |
| H17 | **Free undo:** plans removed with AA\|NP cost nothing, and stray AA track is swept by a whole-map scan (7024-7135). | |
| H18 | **Building over obstacles in the plan without paying for clearance.** Trees and buildings under AA pieces stay until conversion (src/Map/TileClearance.cpp:498-540; CreateTrack.cpp:525). The plan's cost estimate counts their clear cost (3906-3948), but the plan itself occupies the land for free (see H1). | |

## 4. AI state in the company record (kept as loaded)

From the packed S5 structs, `include/OpenLoco/S5/S5Company.h:18-163`. The record is 0x8FA8 bytes (0x8DC8 in the
older `CompanyType2`, the same layout up to 0x8C54). The `AiThought` is 0x8C bytes. The header comments in
CompanyAi.h mix two numberings (offset in the thought, and absolute); the packed struct is the authority. For
example `vehicleObjTypes` is at thought+0x46 (absolute 0x4EE for thought 0), not the 0x4EF a comment says.

18 entries.

| # | Offset | Field | goloco |
|---|---|---|---|
| S1 | 0x04 | `challengeFlags`: aiHasStarted (bit 0), clearance flag (bit 1), converting flag (bit 2), bankrupt (bit 9) | modelled (`CompanyFaces[i].ChallengeFlags`), written back |
| S2 | 0x18 | `competitorId` (selects intelligence, aggressiveness, competitiveness) | modelled |
| S3 | 0x34 | `unlockedVehicles` (224 bits) | modelled |
| S4 | 0x52, 0x56 | `aiPlaystyleFlags` (u32), `aiPlaystyleTownId` | raw only |
| S5 | 0x4A4-0x4A7 | `aiThinkState`, `aiThinkSubState`, `aiPlaceVehicleState`, pad | `AiThinkState` read (pkg/scenario/loader.go:1622) and written (s5save.go:292); the rest raw |
| S6 | 0x4A8 + 0x8C·k, k < 60 | `aiThoughts[60]`. Each thought: +0x00 type (0xFF null), +0x01/0x02 destinations A/B, +0x03 numStations, +0x04 stationLength, +0x06 4 × 0xE stations (id, flags, rotation, pos, baseZ, next/prev index, connection flags A/B), +0x3E trackObjId (bit 7 = road), +0x3F rackRailType, +0x40 mods, +0x42 cargoType, +0x43 numVehiclesTarget, +0x44 numVehicles, +0x45 numVehicleObjTypes, +0x46 vehicleObjTypes[16], +0x66 vehicles[8], +0x76 totalCost, +0x7C totalRunningCost, +0x80 income, +0x84 previousIncome, +0x88 thoughtAge, +0x89 stationObjId, +0x8A signalObjId, +0x8B purchaseFlags | type, destinationA, numVehicles, vehicles, runningCost, income, previousIncome and age are modelled and **mutated** by goloco's engine (unload.go:225-229 credits income; companydaily.go:206-228 quarterly). The rest is raw. |
| S7 | 0x2578 | `activeThoughtId` | modelled, written |
| S8 | 0x2579-0x257D | headquarters Z, X, Y | modelled |
| S9 | 0x257E | `activeThoughtRevenueEstimate`, a union with `thoughtState2AiStationIdx` | raw |
| S10 | 0x2582 | `aiThoughtCooldown`; then 0x2586-0x2595 pad | raw |
| S11 | 0x2596-0x259C | `aiBridgeSelectionCounter`, bridge types low, medium, high | raw |
| S12 | 0x259E | `aiPlaceVehicleIndex` (u32); then 0x25A2-0x25BD pad | raw |
| S13 | 0x25BE | `aiSavedThoughtType` | raw; set to 0xFF for a company created in play (pkg/scenario/s5save.go:203) |
| S14 | 0x25C0-0x85BF | `trackandRoadHashTable[0x1000]` (6 bytes each: x, y\|flag, z, id\|dir): the pathfinder's closed set | raw (24 KiB per company) |
| S15 | 0x85C0-0x85F7 | `hashTableLength`, `aiStationIndex`, `aiStationFlags`, pathfind target/start pos, baseZ, rot or tad ×2, weightings 1/2, `aiBestTrackRoadId`, `aiPathfindPhase`, `aiPathfindMaxWeighting`, undo counts 1/2, `aiPathfindIterCount`, `aiVehicleCost`, `aiThinkTimer` | raw |
| S16 | 0x8BB9-0x8BCA | observation fields (written by `setAiObservation` daily) | modelled (pkg/sim/companydaily.go) |
| S17 | vehicle head +0x60-0x67 | `aiThoughtId` (0xFF none), `aiPlacementPos`, `aiPlacementTaD`, `aiPlacementBaseZ` (include/OpenLoco/S5/S5Entity.h:168-171) | raw in the entity; set for created heads (s5entities.go:757-765); 0x60 read by unload.go |
| S18 | game state | competitor start delay, max competitors, preferred AI traits, forbidden-vehicle masks, `produceAICompanyTimeout` | `PreferredAI`, `ForbiddenVehicles*`, `MaxCompetitors` modelled (pkg/sim/map.go:105-107, pkg/scenario/types.go:665) |

Round-trip status: `EncodeGameState` starts from the loaded raw game state (pkg/scenario/s5save.go:98-115), so
every raw field above survives a load and save unchanged. Two open questions:

1. **Dangling references.** If a goloco AI is given an OpenLoco AI company with live thoughts, those thoughts
   name vehicles and stations that our AI may sell. goloco's engine would then credit income to a thought slot
   (unload.go uses head 0x60) and recount running cost from vehicle ids that are gone or reused. The original
   AI, loading that save back, would act on stale plans. Options: (a) our AI takes over a company only after
   clearing its thoughts to null, and the vehicles' `aiThoughtId` to 0xFF, as a fresh company has them; or
   (b) our AI only runs companies it created. Proposal: (a), done once at takeover and logged.
2. **Where our AI's own state lives.** It must not reuse OpenLoco's AI bytes, which the original would
   misread. Options: a goloco-only chunk in the save, or a sidecar file. The interface below keeps it opaque
   (`Save`/`Load` bytes); where it goes is a save-format decision for later.

## 5. The pathfinder's needs (CompanyAiPathfinding.cpp)

How it works: `aiPathfind` (2503-2665) runs one step per think. The upper bound is 384 steps per phase
(2531-2541), and the phases are 0, 1 and 2 (2400-2413). It grows track from **both** station stubs. At each
step, for every allowed piece id (`getPlacementVars`, 67-265: from the track or road object's traits, steep
slopes only with rack rail), it scores the piece by recursive look-ahead
(`queryTrackPlacementScoreRecurse`, 333-475). Each look-ahead piece is a **dry run of createTrack**. The
search goes down to a weighting limit of 138 or 224 (802). The score is
distance-to-target × 4 + bridge height weighting / 32 + buildings destroyed. The best piece is then built as
AA (876-940). Dead ends go into the company's hash table (closed set), and the search backtracks by
removing AA pieces (678-787). When the stubs meet (268-323), `measureRoute` (2169-2398) walks the result.
It rejects a route that is more than 1.75 × the straight distance or has too many bridge penalties
(2416-2500), and it adds the conversion cost (A14 queried directly) to the thought's cost. The road side
(526-676, 945-1203, 1426-1646, 1965-2168, 2272-2379) mirrors the track side; it was skimmed, not read line
by line **(unsure in detail)**.

13 entries.

| # | Need | Upstream | goloco |
|---|---|---|---|
| P1 | **Dry-run a track or road piece** at (pos, rotation, piece, object, bridge, mods): ok or refused, cost | createTrack/createRoad without apply (Pathfinding.cpp:412, 566) | the functions exist, unexported (see G1) |
| P2 | From that dry run: **a bridge is needed**, and the **bridge height** | legacy return state `flags_1136073` bit 0, `byte_1136074` (Pathfinding.cpp:419-427) | `Cmd.Flags73` is exposed; the height exists as `MaxBridgeDiff` in the internal result structs (pkg/sim/createtrackclear.go:27, roadclear.go:16), **not exposed** |
| P3 | From that dry run: **buildings would be destroyed** | `flags_1136073` bit 4 (Pathfinding.cpp:429-433; set at CreateTrack.cpp:300, CreateRoad.cpp:572) | yes: `Cmd.Flags73` bit 4 (set in pkg/sim/createtrackclear.go:203, roadclear.go:127) |
| P4 | Road: **overlays an existing road** (bit 5), **alternate road object**, **existing bridge type** | Pathfinding.cpp:566-590; `alternateRoadObjectId`, `byte_1136075` | `Cmd.AltRoadObject`, `Cmd.Bridge` exposed; bit 5 set in pkg/sim/roadclearhelp.go:140, 216 |
| P5 | **Piece geometry:** end position and rotation of a (piece, rotation), whether it is diagonal (rotation ≥ 12), start z offset | `getUnkTrack/getUnkRoad`, `getTrackPiece/getRoadPiece` | yes: `TrackCoordinates`, the road tables, and the piece tables |
| P6 | **Piece weighting** (length cost) and **traits** (slope, steep, curve) | `getTrackMiscData(id).unkWeighting/.flags` (Pathfinding.cpp:367, 418) | yes: the weighting is in the generated piece data (`TrackUnkWeighting`, pkg/sim/trackpieces_gen.go:98) |
| P7 | **Which piece ids a track or road object allows** | object traits (Pathfinding.cpp:186-262) | yes: object data |
| P8 | **Connections at a position**, own company, given object | `getTrackConnectionsAi` (758, 2247); `getRoadConnectionsAiAllocated` (1031, 2338) | `Map.TrackConnections`, `Map.RoadConnections`; no AA variant (unneeded without H1) |
| P9 | **Does a piece join an existing track or road** (reuse instead of building) | `connectsToExistingTrack/Road` 1870-2063 | data in tiles; no helper |
| P10 | **Bridge-penalty test** (does the piece span water or roads in a costly way) | `trackBridgeIncursPenalty`, `roadBridgeIncursPenalty` 1426-1869 | data in tiles; no helper |
| P11 | **Would a road destroy a building** | `willRoadDestroyABuilding` 2064-2168 | P3 covers it via a dry run |
| P12 | **Replacement or conversion cost** of a planned route | `aiTrackReplacement` query (2215-2223) and `RoadReplacePrice` (1213-1424) | without reservations: the sum of dry-run costs (P1) |
| P13 | **Search memory** (closed set, progress) that lasts across thinks | the hash table and pathfind fields in the company record (S14, S15) | the AI's own state (§6), not the company record |

Conclusion: the original pathfinder needs nothing from the world beyond **tiles, piece tables, connections and
a dry-run createTrack/createRoad that reports its side information**. Its look-ahead does hundreds of dry runs
per step, so the dry run must be cheap and side-effect free. goloco's split into `createTrackCheck` (no tile
writes) and `createTrackApply` already allows that. A better AI would search in its own memory (A* over the
tile grid with dry-run checks) and build the found route with real, paid commands. It would not lay invisible
track piece by piece.

## 6. A first draft of the Go interface

Principles, following PORT-PLANS F: an AI sees what a player could see and does only what a player can do, through
the same commands, at the same price.

- **No flags parameter.** The Actor never offers `noPayment`, `aiAllocated`, `ghost`, `flag_7`, or
  `allowNegativeCashFlow`. Every action runs the normal `DoCommandFor` query and apply pass under the
  company's id.
- **The dry run is first-class.** Every action can be queried for its cost and side information without
  applying it.
- **Its own randomness.** Each AI gets its own PRNG, seeded from the game seed and the company id, and saved
  with its state. It never draws from `Map.Rng1`. That keeps a goloco game with AI companies comparable,
  tick for tick, with the same game with the AI off.
- **Logged and replayable.** Each think records its observations' hash, its decisions and their results, so a
  bad choice can be traced to its inputs (PORT-PLANS F).

```go
package ai // pkg/ai: the interface; pkg/sim implements Observer and Actor over the live Map.

// Observer is a read-only view of the game for one company: what a player
// can see in the windows and on the map. Values are copies; nothing returned
// aliases sim state.
type Observer interface {
	Self() CompanyID
	Date() sim.Date           // year, month, day; days since start
	Ticks() uint32

	// The map
	MapSize() (w, h int)
	Surface(x, y int) Surface // base and corner heights, slope, water level, terrain
	Elements(x, y int) []Element // track, road, station, signal, building, tree, industry, wall:
	                              // kind, owner, object, piece id, rotation, height, mods, bridge
	                              // (ghosts and other companies' reservations are not shown)
	NearbyWater(x, y int) int     // countNearbyWaterTiles
	Catchment(r TileRect) (accepted, produced CargoSet) // calcAcceptedCargoAi

	// Places
	Towns() []Town              // id, name, position, population, capacity, ratings
	Industries() []Industry     // id, object, position, town, produced last month, daily production,
	                            // closing down, accepted cargo
	Stations() []Station        // all: owner, position, modes, tiles; cargo waiting and acceptance
	TownDemand(id TownID) CargoSet // what its buildings consume (getTownRequiredCargoTypes)

	// Companies: own in full, others as the company windows show them
	Company() CompanyView       // cash, loan, max loan, interest rate, bankrupt, HQ, value,
	                            // expenditures, performance, started date
	Competitors() []PublicCompany // name, value, performance, vehicle counts (no plans)

	// Vehicles: own in full; others' positions and owners as the map shows
	Vehicles() []Vehicle        // head id, mode, object ids per car, cargo, orders, status,
	                            // age, reliability, profit this and last year, running cost
	OtherVehicles() []VehicleSighting

	// What can be built and bought now
	Objects() Catalogue         // vehicles (unlocked for this company), tracks, roads, stations,
	                            // signals, bridges, airports, docks, mods; their stats and costs
	CanCouple(a, b VehicleObjectID) bool

	// Rules and formulas
	Cost(factor int16, index, divisor uint8) Money       // inflation adjusted
	Payment(cargo CargoID, units, distance int, days int) Money
	Connections(p TrackPos, obj ObjectID, road bool) []TrackPos // the own network from a piece end
	PieceEnd(p TrackPos, road bool) TrackPos                    // geometry tables
	Pieces(obj ObjectID, road bool) []PieceID                   // which pieces the object allows
}

// Actor issues game commands for one company: only commands a player has,
// each through DoCommandFor (query pass, funding check, apply pass, payment).
// Query runs the query pass alone and changes nothing.
type Actor interface {
	Query(c Command) Result
	Do(c Command) Result
}

// Command is one game command with explicit arguments, as a player's window
// would send it. The set, from §1:
//   BuildTrack{Pos, Rot, Piece, Track, Bridge, Mods}  RemoveTrack{...}
//   BuildRoad{...}  RemoveRoad{...}
//   BuildTrainStation{...} RemoveTrainStation{...}
//   BuildRoadStation{...}  RemoveRoadStation{...}
//   BuildAirport{...} RemoveAirport{...}  BuildPort{...} RemovePort{...}
//   BuildSignal{...} RemoveSignal{...}  BuildTrackMod{...} BuildRoadMod{...}
//   BuyVehicle{Object} AddCar{Head, Object}  Refit{Head, Cargo}
//   InsertOrder{Head, Offset, Order} DeleteOrder{...} SkipOrder{Head}
//   PlaceVehicle{Head, Pos, TaD} PlaceAircraft{Head, Station, Node} PlaceShip{Head, Pos}
//   Pickup{Head}  Start{Head} Stop{Head}  SellVehicle{Head}
//   BuildHQ{Pos, Rot} RemoveHQ{}  ChangeLoan{To}  RenameCompany{Name}
type Command interface{ command() }

type Result struct {
	OK     bool
	Cost   Money  // negative: a refund
	Err    string // the error text a player would see
	// the dry-run side information the pathfinder needs (§5 P2-P4)
	NeedsBridge      bool
	BridgeHeight     uint8
	RemovesBuildings bool
	OverlaysRoad     bool
	AltRoadObject    ObjectID // 0xFF none
	// what the command made
	Vehicle EntityID
	Station StationID
}

// CompanyAI is one AI, run for one company, chosen when the company is created.
type CompanyAI interface {
	Name() string
	// Start: a new or taken-over company; cfg carries the difficulty knobs
	// (OpenTTD-style: start delay, how hard, which transport modes).
	Start(o Observer, a Actor, cfg Config, rng *core.Prng)
	// Think: every 16 ticks for this company slot, as CompanyManager::tick
	// (scenarioTicks & 0x0F == id). One bounded step; long work (route search)
	// spreads over thinks.
	Think(o Observer, a Actor)
	// Events: a new day, month, quarter, year; a vehicle crashed, stuck or
	// broken down; a station's rating or an industry closing; the company bankrupt.
	OnEvent(o Observer, a Actor, e Event)
	// The AI's own state, saved with the game (not in OpenLoco's AI bytes, §4).
	Save() []byte
	Load(b []byte) error
}
```

Engine rules that stay (they apply to every AI equally, and each is named in the config):

- the 500-vehicle AI cap;
- 2 AI airports per town;
- the competitors' forbidden-vehicle mask;
- the monthly auto-repay (H4);
- the removal of unused town-type roads;
- **(decision for Joop)** auto-borrowing in the funding check (H3): keep it, as OpenLoco does for non-players,
  or treat our AI companies like players, so they must `ChangeLoan` themselves. Proposal: players' rule.
  Borrowing becomes a visible decision in the log, and the H3 code path stays only for OpenLoco's own AI
  companies when the original rules are in force (AI off, simcmp).

Ruled out: H1 (reservations), H2 (free building), H5 (acting as neutral), H6 (direct calls), H7 (AI-only
commands), H8 and H9 (direct writes; the engine keeps its own bookkeeping), H11 (others' plans), H12 and H13
(AI-only argument bits), H16 (engine sale on the AI's behalf; the AI gets the crash event and decides).

## 7. Gaps in goloco to fill before any AI can run

13 entries, roughly in order of need.

1. **G1 Exported track and road commands with explicit args.** `createTrack`/`createRoad` are reachable only
   through the construction window's state (pkg/sim/construction.go:205-260). Add `Map.CreateTrack(c, flags,
   TrackPlacement)` and `Map.CreateRoad(...)` like the other commands (CreateSignal, CreateTrainStation...).
2. **G2 The dry run's side information on `Cmd`:** bridge height (`MaxBridgeDiff`, now internal). `Flags73` bits 4
   (buildings removed) and 5 (road overlay) are already set as upstream (P3, P4).
3. **G3 A company tick.** `updateManagers` has no CompanyManager::update ("PARTIAL", pkg/sim/vehicletick.go:81),
   and `Config.CompanyAIDisabled` is a placeholder (pkg/sim/config.go:13). Add the hook: after the effects,
   before the animations, company `ScenarioTicks & 0x0F`, only non-player and non-empty companies.
4. **G4 Creating AI companies:** `produceCompanies` (every 192 ticks; start delay, max competitors) and the
   competitor branch of `createCompany` (name, playstyle, home town). Both are missing or PARTIAL
   (pkg/sim/newgame.go:9-13).
5. **G5 Removing a company** (`aiDestroy`, src/World/CompanyManager.cpp:1436): for an AI that gives up or never
   starts. Bankruptcy for AIs follows the engine's existing rules.
6. **G6 The funding decision (H3)** in `ensureCompanyFunding`/`DoCommandFor`: a per-company switch,
   "borrows automatically", default off for our AIs.
7. **G7 Observer helpers that exist but are unexported:** `countNearbyWaterTiles`, `mostCommonBuildingCargoType`,
   `surfaceCornerHeight`, `canVehiclesCouple`, `industryByID`; a catchment for an arbitrary rectangle (only
   ghost-station forms exist, pkg/sim/stationaccept.go:122-190); town demand (`getTownRequiredCargoTypes`).
8. **G8 Vehicle purchase as one action:** CreateVehicle and AddCarToVehicle exist but leave payment to the caller
   ("PARTIAL", createvehiclecmd.go:23). The Actor must run both through `DoCommandFor` under VehiclePurchases.
9. **G9 Air and water vehicles do not move yet** (pkg/sim/vehicletick.go:16-17: "Air and water vehicles are
   not updated"; airmotion.go exists, so check the current state). An AI should begin with rail and road.
10. **G10 Takeover of a loaded OpenLoco AI company:** clear its thoughts and the vehicles' `aiThoughtId` (§4
    question 1). Until then, run our AI only on companies it created.
11. **G11 Saving the AI's own state** (a goloco chunk or sidecar; §4 question 2), the per-AI PRNG, and the
    decision log.
12. **G12 Events** for `OnEvent`: crashes, breakdowns and closures are posted as news today. An event channel
    for the AI hook is needed (it could be the news messages themselves, filtered by company).
13. **G13 The options UI** (which AI, how many, start delay, difficulty) is later work, after G3 and G4.

## 8. Notes and surprises

- The original AI **never reads its cash**. Every "can I afford this" is `ensureCompanyFunding`, which borrows
  as a side effect. Combined with H2, the AI can never really run out of money until bankruptcy.
- Planning happens **on the map**, as invisible free track that blocks players. Undo happens there too. The
  "pathfinder" is mostly a loop of dry-run and AA-apply commands, with a closed set stored in the save
  (24 KiB per company).
- Placing, starting and failed-placement sales run **as the neutral company**, which skips ownership checks
  and throws away the sale refund.
- Aggressiveness has one use: it sets how many thinks pass between converted pieces (1-25, CompanyAi.cpp:6079-6090).
  Intelligence scales the revenue it demands (4069-4101). Competitiveness below 5 avoids any route another
  company already serves (3057-3060).
- The AI shares `gPrng1` with the game. With the AI on, goloco and OpenLoco can only match if the AI is a
  port. That is why the economy comparison stays an AI-off check (PORT-PLANS F).
