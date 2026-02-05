# GoLoco Object Types - Complete Reference

## Currently Loaded (4 types)

1. **InterfaceSkin** (ObjectType 0) - ✅ WORKING
   - File: `pkg/objects/interfaceskin.go`
   - Sprites: UI buttons, toolbar, windows
   - Status: Fully implemented with dynamic G1 loading

2. **Land** (ObjectType 6) - ✅ WORKING
   - File: `pkg/objects/land.go`
   - Sprites: Terrain tiles (grass, sand, rock, etc.)
   - Status: Fully implemented with dynamic G1 loading

3. **Cargo** (ObjectType 8) - ⚠️ PARTIAL
   - File: `pkg/loco/objects/cargo.go`
   - Sprites: Cargo icons
   - Status: Type definitions only, no loading

4. **Vehicle** (ObjectType 23) - ⚠️ PARTIAL
   - File: `pkg/objects/vehicle.go`
   - Sprites: Trains, buses, trucks, trams, aircraft, ships
   - Status: Parsing implemented, NO dynamic sprite loading

## Not Yet Loaded (30 types)

### High Priority (Core Gameplay)

5. **CliffEdge** (ObjectType 4) - 🔴 NEEDED FOR TERRAIN
   - OpenLoco: `Objects/CliffEdgeObject.h`
   - Sprites: Cliff edges for height transitions
   - Why: Referenced by LandObject, needed to fill terrain gaps

6. **Road** (ObjectType 20) - 🔴 CORE
   - OpenLoco: `Objects/RoadObject.h`
   - Sprites: Road tiles, connectors
   - Why: Primary transport infrastructure

7. **Track** (ObjectType 17) - 🔴 CORE
   - OpenLoco: `Objects/TrackObject.h`
   - Sprites: Railway tiles, connectors
   - Why: Primary transport infrastructure

8. **Building** (ObjectType 28) - 🔴 CORE
   - OpenLoco: `Objects/BuildingObject.h`
   - Sprites: Town buildings, houses
   - Why: Towns need buildings to render

9. **Tree** (ObjectType 24) - 🔴 VISUAL
   - OpenLoco: `Objects/TreeObject.h`
   - Sprites: Various tree types
   - Why: Visible on terrain, important for aesthetics

10. **Water** (ObjectType 5) - 🔴 TERRAIN
    - OpenLoco: `Objects/WaterObject.h`
    - Sprites: Water tiles, waves
    - Why: Water tiles need sprites

### Medium Priority (Stations & Infrastructure)

11. **TrainStation** (ObjectType 15)
    - OpenLoco: `Objects/TrainStationObject.h`
    - Sprites: Railway station buildings

12. **RoadStation** (ObjectType 18)
    - OpenLoco: `Objects/RoadStationObject.h`
    - Sprites: Bus/truck stops

13. **Airport** (ObjectType 21)
    - OpenLoco: `Objects/AirportObject.h`
    - Sprites: Airport buildings, runways

14. **Dock** (ObjectType 22)
    - OpenLoco: `Objects/DockObject.h`
    - Sprites: Ship docks, harbors

15. **Bridge** (ObjectType 14)
    - OpenLoco: `Objects/BridgeObject.h`
    - Sprites: Bridge sections

16. **Tunnel** (ObjectType 13)
    - OpenLoco: `Objects/TunnelObject.h`
    - Sprites: Tunnel entrances

17. **Wall** (ObjectType 9)
    - OpenLoco: `Objects/WallObject.h`
    - Sprites: Fences, walls

### Lower Priority (Supporting Elements)

18. **TrackSignal** (ObjectType 10)
    - OpenLoco: `Objects/TrainSignalObject.h`
    - Sprites: Railway signals

19. **TrackExtra** (ObjectType 16)
    - OpenLoco: `Objects/TrackExtraObject.h`
    - Sprites: Overhead wires, etc.

20. **RoadExtra** (ObjectType 19)
    - OpenLoco: `Objects/RoadExtraObject.h`
    - Sprites: Streetlights, tram wires

21. **LevelCrossing** (ObjectType 11)
    - OpenLoco: `Objects/LevelCrossingObject.h`
    - Sprites: Road/rail crossing gates

22. **StreetLight** (ObjectType 12)
    - OpenLoco: `Objects/StreetLightObject.h`
    - Sprites: Street lamps

23. **Industry** (ObjectType 30)
    - OpenLoco: `Objects/IndustryObject.h`
    - Sprites: Factories, mines, etc.

24. **Scaffolding** (ObjectType 29)
    - OpenLoco: `Objects/ScaffoldingObject.h`
    - Sprites: Building construction

### Metadata/Configuration (No Sprites)

25. **Sound** (ObjectType 1)
    - OpenLoco: `Objects/SoundObject.h`
    - Data: Sound effects (audio data, not sprites)

26. **Currency** (ObjectType 2)
    - OpenLoco: `Objects/CurrencyObject.h`
    - Data: Currency definitions (text, no sprites)

27. **Steam** (ObjectType 3)
    - OpenLoco: `Objects/SteamObject.h`
    - Sprites: Steam puff effects

28. **TownNames** (ObjectType 7)
    - OpenLoco: `Objects/TownNamesObject.h`
    - Data: Town name generation rules (text, no sprites)

29. **Snow** (ObjectType 25)
    - OpenLoco: `Objects/SnowObject.h`
    - Sprites: Snow coverage overlays

30. **Climate** (ObjectType 26)
    - OpenLoco: `Objects/ClimateObject.h`
    - Data: Climate definitions (config, no sprites)

31. **HillShapes** (ObjectType 27)
    - OpenLoco: `Objects/HillShapesObject.h`
    - Data: Map generation rules (config, no sprites)

32. **Region** (ObjectType 31)
    - OpenLoco: `Objects/RegionObject.h`
    - Data: Regional settings

33. **Competitor** (ObjectType 32)
    - OpenLoco: `Objects/CompetitorObject.h`
    - Data: AI competitor definitions

34. **ScenarioText** (ObjectType 33)
    - OpenLoco: `Objects/ScenarioTextObject.h`
    - Data: Scenario descriptions (text, no sprites)

## Implementation Plan

### Phase 1: Critical Terrain (Next)
- [ ] CliffEdgeObject - Fill terrain height gaps
- [ ] WaterObject - Water tiles
- [ ] TreeObject - Decorative elements

### Phase 2: Core Infrastructure
- [ ] RoadObject - Roads
- [ ] TrackObject - Railways
- [ ] BuildingObject - Towns

### Phase 3: Vehicles (Enable with existing VehicleObject)
- [ ] Update VehicleObject to use dynamic G1 loading

### Phase 4: Stations & Extended Infrastructure
- [ ] TrainStation, RoadStation, Airport, Dock
- [ ] Bridge, Tunnel, Wall

### Phase 5: Details
- [ ] Signals, extras, crossing, lights
- [ ] Industry, scaffolding
- [ ] Steam, snow effects

### Phase 6: Metadata
- [ ] Sound, Currency, TownNames
- [ ] Climate, HillShapes, Region
- [ ] Competitor, ScenarioText

## OpenLoco Reference Files

All object definitions: `/home/joop/goloco-project/OpenLoco/src/OpenLoco/src/Objects/`

Conversion references: `/home/joop/goloco-project/goloco/pkg/_conversion_clean/objects/`
