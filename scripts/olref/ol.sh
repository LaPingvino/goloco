#!/bin/bash
# The real OpenLoco, headless, on its own X display and its own scratch config.
#
#   OLREF_HOME=<scratch dir> ol.sh start [SAVE.SV5 | SCENARIO.SC5]
#   ol.sh do move 592 14        # xt.py commands, in game-window pixels
#   ol.sh run script.txt        # a goloco-style input script (xt.py)
#   ol.sh shot out.png          # the 800x600 game frame
#   ol.sh stop
#
# OLREF_HOME    scratch HOME for OpenLoco (config in $OLREF_HOME/.config/OpenLoco,
#               autosaves in …/save/autosave) — required, never the player's own
# OLREF_DISPLAY X display of the harness's Xvfb (default :97)
# OLREF_DATA    the Locomotion folder (default ~/Projects/locomotion)
# OLREF_BIN     the OpenLoco binary (default: openloco on PATH)
#
# All coordinates are game-window pixels — the ones goloco's GOLOCO_SCRIPT
# files use. OpenLoco refuses a window less than 160 px smaller than the
# desktop (Ui.cpp getSaneWindowedResolution), so the 800x600 window sits
# centred in a 1024x768 screen, at OLREF_ORIGIN (112,84); xt.py adds it and
# shot crops by it.
set -eu
here=$(dirname "$(readlink -f "$0")")
disp=${OLREF_DISPLAY:-:97}
data=${OLREF_DATA:-$HOME/Projects/locomotion}
export DISPLAY=$disp OLREF_ORIGIN=${OLREF_ORIGIN:-112,84}

need_home() {
    : "${OLREF_HOME:?set OLREF_HOME to a scratch directory}"
    cfgdir=$OLREF_HOME/.config/OpenLoco
    run=$OLREF_HOME/.olref
    mkdir -p "$cfgdir" "$run"
}

# set_key KEY VALUE — a top-level "key: value" of openloco.yml, added if absent.
set_key() {
    if grep -q "^$1:" "$cfgdir/openloco.yml"; then
        sed -i "s/^$1:.*/$1: $2/" "$cfgdir/openloco.yml"
    else
        echo "$1: $2" >>"$cfgdir/openloco.yml"
    fi
}

configure() {
    local yml=$cfgdir/openloco.yml
    if ! grep -q '^display:' "$yml" 2>/dev/null; then
        printf 'display:\n  mode: window\n  window_resolution:\n    width: 800\n    height: 600\n' >"$yml"
    fi
    # the first width/height pair is display.window_resolution
    sed -i '/^  window_resolution:/,/^  [a-z_]*resolution:\|^[a-z]/{s/^    width:.*/    width: 800/;s/^    height:.*/    height: 600/}' "$yml"
    set_key scale_factor 1
    # 26.08 opens a top-toolbar menu when the pointer merely rests on its
    # button (ToolbarTop.cpp onMouseHover, on by default). Off: a move is a
    # hover, a click opens the menu — as in goloco and in older OpenLoco shots.
    # OLREF_AUTOMENU=true keeps the release's default.
    set_key toolbarAutoMenu "${OLREF_AUTOMENU:-false}"
    # A pointer left at the frame's edge would scroll the map between shots.
    set_key edgeScrolling false
    set_key autosave_frequency 1
}

case ${1:-} in
start)
    shift
    need_home
    if [ -e "$run/openloco.pid" ] && kill -0 "$(cat "$run/openloco.pid")" 2>/dev/null; then
        echo "already running (ol.sh stop first)" >&2
        exit 1
    fi
    configure
    # -noreset: the server keeps its state when the last client leaves.
    setsid Xvfb "$disp" -screen 0 1024x768x24 -nolisten tcp -noreset >"$run/xvfb.log" 2>&1 &
    echo $! >"$run/xvfb.pid"
    for _ in $(seq 50); do
        python3 -I "$here/xt.py" state >/dev/null 2>&1 && break
        sleep 0.1
    done
    HOME=$OLREF_HOME XDG_CONFIG_HOME=$OLREF_HOME/.config SDL_AUDIODRIVER=dummy \
        setsid "${OLREF_BIN:-openloco}" --locomotion_path "$data" "$@" >"$run/openloco.log" 2>&1 &
    echo $! >"$run/openloco.pid"
    echo "openloco $! on $disp (log: $run/openloco.log)"
    ;;
stop)
    need_home
    for p in openloco xvfb; do
        [ -e "$run/$p.pid" ] && kill "$(cat "$run/$p.pid")" 2>/dev/null || true
        rm -f "$run/$p.pid"
    done
    ;;
shot)
    mkdir -p "$(dirname "$2")"
    import -window root -crop "800x600+${OLREF_ORIGIN/,/+}" +repage "$2"
    ;;
do)
    shift
    exec python3 -I "$here/xt.py" "$@"
    ;;
run)
    exec python3 -I "$here/xt.py" run "$2"
    ;;
*)
    sed -n '2,20p' "$0" | sed 's/^# \?//'
    exit 2
    ;;
esac
