#!/usr/bin/env bash
# simcmp.sh — the same save played N ticks by OpenLoco and by goloco, compared
# by OpenLoco's own GameSaveCompare: `openloco simulate IN N CMP` plays IN for N
# ticks headless and lists every divergence from CMP (here goloco's gosave of IN
# after N ticks). LHS is OpenLoco's state, RHS goloco's.
#
#   OLREF_HOME=<scratch> scripts/olref/simcmp.sh in.SV5 N [-a]
#
# OLREF_HOME keeps OpenLoco's config away from the real ~/.config/OpenLoco.
# OpenLoco writes OpenLoco.lock in the working directory: a scratch one is used.
set -euo pipefail
: "${OLREF_HOME:?set OLREF_HOME to a scratch directory}"
in=$(realpath "$1") n=$2 all=${3:-}
repo=$(cd "$(dirname "$0")/../.." && pwd)
work=$OLREF_HOME/simcmp
mkdir -p "$work" "$OLREF_HOME/.config"
cp "$in" "$work/in.SV5"
go -C "$repo" run ./scripts/olref/gosave -ticks "$n" "$work/in.SV5" "$work/go_$n.SV5" >/dev/null 2>&1
env -C "$work" HOME="$OLREF_HOME" XDG_CONFIG_HOME="$OLREF_HOME/.config" OPENLOCO_DATA_DIR=${OPENLOCO_DATA_DIR:-/usr/share/openloco} \
	openloco simulate ${all:+--all} --locomotion_path "${OLREF_DATA:-$HOME/Projects/locomotion}" in.SV5 "$n" "go_$n.SV5" 2>&1 |
	grep -v "Loading\|objects\|Searching\|Platform\|Warning" || true
