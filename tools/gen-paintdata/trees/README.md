gen_trees.cpp checks pkg/sim's tree generation against upstream. The tt_*.inc files it includes are copied from
OpenLoco's sources (Map/TileManager.cpp getHeight + helpers + countSurroundingWaterTiles; Map/TileClearance.cpp
isCollisionResult … canConstructAt; GameCommands/Terraform/CreateTree.cpp; Map/MapSelection.cpp getQuadrantFromPos;
Map/Tree.cpp; Map/MapGenerator/MapGenerator.cpp updateTreeSeasons + generateTrees; TileManager::countNearbyWaterTiles and GameCommands/Town/CreateTown.cpp checkSurroundings + createTown, with initialiseTown stubbed to record the chosen position) by the extraction script in
LOOP.md tick 280 — they are not kept here; treein.txt is dumped from the game data by a throwaway test (never
committed). Output: pkg/scene/testdata/trees_hashes.txt.
