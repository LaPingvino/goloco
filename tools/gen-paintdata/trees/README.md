gen_trees.cpp checks pkg/sim's tree generation against upstream: generateTrees on goloco's generated Aerophobia
map, three cases, every tree hashed, plus the first town's site (CreateTown) right after the trees.
Output: pkg/scene/testdata/trees_hashes.txt, read by pkg/scene TestTreesMatchUpstream.

Regenerate (whenever the map before the trees changes, or upstream's tree code does):

1. Dump the map (game data: never commit it; keep it under /tmp/claude-1000/). A throwaway test in pkg/scene,
   deleted afterwards, does what TestTreesMatchUpstream does up to the trees — saveView of
   Scenarios/Aerophobia.SC5, the same TerrainOptions, m.GenerateTerrainOnly — and writes treein.txt:
   - one line: numberOfForests minForestRadius maxForestRadius minForestDensity maxForestDensity
     numberRandomTrees minAltitudeForTrees maxAltitudeForTrees (sc.Options) m.CurrentSnowLine m.CurrentSeason
     m.Width*m.Height (entriesEnd) seaLevel
   - 64 lines, tree object slots: ok flags initialHeight height growth currentSeason (0 0 0 0 0 0 if empty)
   - 32 lines, land object slots: ok flags
   - 384*384 lines, row-major (y outer): TerrainIndex BaseZ Slope WaterLevel
2. Run `tools/gen-paintdata/trees/extract.sh treein.txt [~/Projects/OpenLoco] [workdir]
   > pkg/scene/testdata/trees_hashes.txt` (no display needed; workdir defaults to a new dir under $TMPDIR).
   It copies the upstream code verbatim into workdir/tt_*.inc (each piece from its signature line to the closing
   brace at that indentation; a renamed signature stops the script), compiles gen_trees.cpp with upstream's
   Prng.cpp and Trigonometry.cpp, and runs it:
   - tt_height.inc: Map/TileManager.cpp kTileSize = 31, getOneCornerUpLandHeight … getValleyLandHeight,
     getHeight, countSurroundingWaterTiles
   - tt_water.inc: TileManager::countNearbyWaterTiles
   - tt_clear.inc: Map/TileClearance.cpp isCollisionResult, callClearFunction, canConstructAtCheckSurfaceElement
     … canConstructAt (aiCompanyAboutToBuildCheck and BuildingCollisionType are stubbed in gen_trees.cpp)
   - tt_create.inc: GameCommands/Terraform/CreateTree.cpp _quadrantCentreOffsets + createTree
   - tt_quad.inc: Map/MapSelection.cpp getQuadrantFromPos
   - tt_tree.inc: Map/Tree.cpp getRandomTreeTypeFromSurface, placeRandomTree, placeTreeCluster
   - tt_gen.inc: Map/MapGenerator/MapGenerator.cpp updateTreeSeasons + generateTrees
   - tt_town.inc: GameCommands/Town/CreateTown.cpp checkSurroundings + createTown (initialiseTown stubbed to
     record the chosen position)
   The .inc files are upstream code: never keep them in the repo.
3. `go test -exec "$PWD/scripts/xvfb-exec.sh" -run TestTreesMatchUpstream ./pkg/scene`. If goloco misses the new
   hashes, find the divergence (TTLOG=1 makes the harness print its first inserts); never fit the hashes.

A check that the harness itself is sound: run it on a dump of the last map the committed hashes were made from
(a git worktree at that commit); it must print them unchanged.
