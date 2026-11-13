#!/bin/sh
# Runs the playthroughs (all, or the named ones) and says which pass: each on
# an Xvfb of its own, from the repository root, its log in $TMPDIR. A script
# fails on a non-zero exit — a panic, or one of its "expect" checks. Under a
# script the game clock is virtual (cmd/goloco clock.go): every run of a
# script plays the same.
#
#   scripts/playthrough/run.sh [a-newgame h2-paying-route ...]
here=$(dirname "$(readlink -f "$0")")
root=$here/../..
bin=${TMPDIR:-/tmp}/goloco-playthrough-bin
go -C "$root" build -o "$bin" ./cmd/goloco || exit 1
names=${*:-$(ls "$here"/*.txt | xargs -n1 basename | sed 's/\.txt$//')}
status=0
for p in $names; do
	log=${TMPDIR:-/tmp}/goloco-playthrough-$p.log
	start=$(date +%s)
	if env -C "$root" GOLOCO_SCRIPT="scripts/playthrough/$p.txt" timeout 1800 "$here/../xvfb-exec.sh" "$bin" >"$log" 2>&1; then
		echo "ok    $p ($(( $(date +%s) - start ))s)"
	else
		status=1
		echo "FAIL  $p ($(( $(date +%s) - start ))s): $(grep -m1 -E 'EXPECT FAILED|panic' "$log" || echo "exit code, see $log")"
	fi
done
exit $status
