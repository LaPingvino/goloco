#!/usr/bin/env bash
# extract.sh — build and run the tree harness (gen_trees.cpp) against an OpenLoco checkout.
#   tools/gen-paintdata/trees/extract.sh <treein.txt> [~/Projects/OpenLoco] [workdir]
# Copies the upstream code gen_trees.cpp includes into <workdir>/tt_*.inc verbatim (each piece from its
# signature line to the closing brace at the signature's indentation), compiles the harness with the real
# Prng.cpp and Trigonometry.cpp, and runs it on the dump: its stdout is pkg/scene/testdata/trees_hashes.txt.
# The .inc files and the dump are upstream code / game data: keep them out of the repo (workdir defaults to a
# fresh directory under $TMPDIR).
set -euo pipefail
in=$(realpath "$1")
OL=${2:-$HOME/Projects/OpenLoco}
here=$(cd "$(dirname "$0")" && pwd)
work=${3:-$(mktemp -d "${TMPDIR:-/tmp}/trees.XXXXXX")}
mkdir -p "$work"
S=$OL/src/OpenLoco/src

# piece FILE SIGNATURE: the first line whose text (indentation stripped) starts with SIGNATURE, through the
# first later line that is that indentation + "}" or "};" (a one-line declaration ending in ";" stands alone).
piece() {
  awk -v sig="$2" '
    !on { t = $0; sub(/^[ \t]+/, "", t)
          if (index(t, sig) == 1) { on = 1; ind = substr($0, 1, length($0) - length(t)); print
                                    if (t ~ /;[ \t]*$/ && t !~ /\{/) exit; next } }
    on  { print; if ($0 == ind "}" || $0 == ind "};") exit }' "$1"
  # the signature must exist: a renamed upstream function should fail loudly, not vanish
  grep -qF -- "$2" "$1" || { echo "extract.sh: '$2' not found in $1" >&2; exit 1; }
}

{ piece "$S/Map/TileManager.cpp" 'constexpr uint8_t kTileSize = 31;'
  piece "$S/Map/TileManager.cpp" 'static int16_t getOneCornerUpLandHeight('
  piece "$S/Map/TileManager.cpp" 'static int16_t getOneSideUpLandHeight('
  piece "$S/Map/TileManager.cpp" 'static int16_t getOneCornerDownLandHeight('
  piece "$S/Map/TileManager.cpp" 'static int16_t getValleyLandHeight('
  piece "$S/Map/TileManager.cpp" 'TileHeight getHeight(const Pos2& pos)'
  piece "$S/Map/TileManager.cpp" 'uint16_t countSurroundingWaterTiles(const Pos2& pos)'
} > "$work/tt_height.inc"
piece "$S/Map/TileManager.cpp" 'uint16_t countNearbyWaterTiles(Pos2 pos)' > "$work/tt_water.inc"
# TileClearance: isCollisionResult, callClearFunction, canConstructAtCheckSurfaceElement … canConstructAt.
# aiCompanyAboutToBuildCheck (company AI) and the BuildingCollisionType enum are the harness's own.
{ piece "$S/Map/TileClearance.cpp" 'constexpr bool isCollisionResult('
  piece "$S/Map/TileClearance.cpp" 'static ClearFuncResult callClearFunction('
  piece "$S/Map/TileClearance.cpp" 'static ClearFuncResult canConstructAtCheckSurfaceElement('
  piece "$S/Map/TileClearance.cpp" 'static ClearFuncResult canConstructAtCheckNonSurfaceElement('
  piece "$S/Map/TileClearance.cpp" 'static bool canConstructAtWithClear('
  piece "$S/Map/TileClearance.cpp" 'bool applyClearAtAllHeights('
  piece "$S/Map/TileClearance.cpp" 'bool applyClearAtStandardHeight('
  piece "$S/Map/TileClearance.cpp" 'bool canConstructAt('
} > "$work/tt_clear.inc"
{ piece "$S/GameCommands/Terraform/CreateTree.cpp" 'constexpr std::array<World::Pos2, 4> _quadrantCentreOffsets'
  piece "$S/GameCommands/Terraform/CreateTree.cpp" 'static uint32_t createTree(const TreePlacementArgs& args, const Flags flags)'
} > "$work/tt_create.inc"
piece "$S/Map/MapSelection.cpp" 'MapSelectionType getQuadrantFromPos(const Pos2& loc)' > "$work/tt_quad.inc"
{ piece "$S/Map/Tree.cpp" 'std::optional<uint8_t> getRandomTreeTypeFromSurface('
  piece "$S/Map/Tree.cpp" 'bool placeRandomTree('
  piece "$S/Map/Tree.cpp" 'bool placeTreeCluster('
} > "$work/tt_tree.inc"
{ piece "$S/Map/MapGenerator/MapGenerator.cpp" 'static void updateTreeSeasons()'
  piece "$S/Map/MapGenerator/MapGenerator.cpp" 'static void generateTrees()'
} > "$work/tt_gen.inc"
{ piece "$S/GameCommands/Town/CreateTown.cpp" 'static bool checkSurroundings('
  piece "$S/GameCommands/Town/CreateTown.cpp" 'static uint32_t createTown(const TownPlacementArgs& args, const Flags flags)'
} > "$work/tt_town.inc"
wc -l "$work"/tt_*.inc >&2

inc=(-I"$work" -I"$OL/src/Core/include" -I"$OL/src/Math/include" -I"$OL/src/Core/include/OpenLoco/Core" -I"$OL/src/Math/include/OpenLoco/Math")
g++ -std=c++20 -fno-char8_t -DDEBUG=0 -O2 "${inc[@]}" -o "$work/tt" "$here/gen_trees.cpp" \
  "$OL/src/Core/src/Prng.cpp" "$OL/src/Math/src/Trigonometry.cpp"
echo "OpenLoco $(git -C "$OL" rev-parse --short HEAD 2>/dev/null || echo '?'); work dir $work" >&2
"$work/tt" "$in"
