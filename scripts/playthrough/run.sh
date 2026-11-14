#!/bin/sh
# Runs the playthroughs (all, or the named ones) and says which pass: each on
# an Xvfb of its own, from the repository root, its log in $TMPDIR. A script
# fails on a non-zero exit — a panic, or one of its "expect" checks. Under a
# script the game clock is virtual (cmd/goloco clock.go): every run of a
# script plays the same. Each runs on a scratch config folder
# (XDG_CONFIG_HOME) holding a copy of the player's openloco.yml files: what a
# playthrough changes in the options, its saves and its autosaves (save.go)
# stay out of the player's own.
#
#   scripts/playthrough/run.sh [a-newgame h2-paying-route ...]
here=$(dirname "$(readlink -f "$0")")
root=$here/../..
bin=$(mktemp "${TMPDIR:-/tmp}/goloco-playthrough-bin.XXXXXX") # its own: runs in parallel (other agents) mustn't share it
trap 'rm -f "$bin"' EXIT
go -C "$root" build -o "$bin" ./cmd/goloco || exit 1
names=${*:-$(ls "$here"/*.txt | xargs -n1 basename | sed 's/\.txt$//')}
status=0
for p in $names; do
	log=${TMPDIR:-/tmp}/goloco-playthrough-$p.log
	cfg=$(mktemp -d "${TMPDIR:-/tmp}/goloco-playthrough-cfg.XXXXXX")
	for app in goloco OpenLoco; do
		real=${XDG_CONFIG_HOME:-$HOME/.config}/$app/openloco.yml
		mkdir -p "$cfg/$app"
		[ -f "$real" ] && cp "$real" "$cfg/$app/"
	done
	start=$(date +%s)
	if env -C "$root" XDG_CONFIG_HOME="$cfg" GOLOCO_SCRIPT="scripts/playthrough/$p.txt" timeout 1800 "$here/../xvfb-exec.sh" "$bin" >"$log" 2>&1; then
		echo "ok    $p ($(( $(date +%s) - start ))s)"
	else
		status=1
		echo "FAIL  $p ($(( $(date +%s) - start ))s): $(grep -m1 -E 'EXPECT FAILED|panic' "$log" || echo "exit code, see $log")"
	fi
	rm -rf "$cfg"
done
exit $status
