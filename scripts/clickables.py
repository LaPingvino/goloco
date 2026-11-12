#!/usr/bin/env python3
"""Completion measure: OpenLoco's clickables against goloco's.

For every window in OpenLoco (src/OpenLoco/src/Ui/Windows/**.cpp) it lists the
widgets the window's event handlers act on — each `case …widx::name:` label —
against goloco's markers (below).

goloco's side is counted from explicit markers in the ports —
"clickable(Window.cpp): name …" where a widget works as upstream's does,
"clickable-partial(Window.cpp): …" where it does something but not yet what
upstream does. A window without markers is "not annotated yet", not zero.

    python3 scripts/clickables.py [OPENLOCO_SRC] [--missing]
"""
import os, re, sys

UPSTREAM = os.path.expanduser("~/Projects/OpenLoco/src/OpenLoco/src/Ui/Windows")
GOLOCO = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

# OpenLoco window file → goloco files that port it.
PORTED = {
    "Options.cpp": ["pkg/ui/options.go", "cmd/goloco/options.go"],
    "TownList.cpp": ["pkg/ui/townlist.go", "cmd/goloco/listwindows.go"],
    "StationList.cpp": ["pkg/ui/townlist.go", "cmd/goloco/listwindows.go"],
    "VehicleList.cpp": ["pkg/ui/townlist.go", "pkg/ui/vehiclelistfilter.go", "cmd/goloco/listwindows.go"],
    "BuildVehicle.cpp": ["pkg/ui/buildvehicle.go", "cmd/goloco/buildvehicle.go"],
    "Vehicle.cpp": ["pkg/ui/vehiclewindow.go", "cmd/goloco/vehicle.go", "cmd/goloco/vehicle_pickup.go",
                    "cmd/goloco/vehicle_route.go"],
    "ToolbarTop.cpp": ["pkg/ui/toolbar.go", "cmd/goloco/toolbar.go"],
    "ToolbarTopCommon.cpp": ["pkg/ui/toolbar.go", "cmd/goloco/toolbar.go"],
    "TitleMenu.cpp": ["pkg/ui/title_menu.go", "cmd/goloco/title.go"],
    "TitleOptions.cpp": ["pkg/ui/title_menu.go", "cmd/goloco/title.go"],
    "TitleExit.cpp": ["pkg/ui/title_menu.go", "cmd/goloco/title.go"],
    "ScenarioSelect.cpp": ["pkg/ui/scenarioselect.go"],
    "PromptBrowseWindow.cpp": ["pkg/ui/promptbrowse.go"],
    "PromptOkCancelWindow.cpp": ["pkg/ui/promptokcancel.go"],
    "TimePanel.cpp": ["pkg/ui/timepanel.go", "cmd/goloco/draw.go", "cmd/goloco/update.go"],
    "PlayerInfoPanel.cpp": ["pkg/ui/playerpanel.go", "cmd/goloco/draw.go", "cmd/goloco/update.go"],
    "CompanyFaceSelection.cpp": ["pkg/ui/companyselect.go"],
    "ProgressBar.cpp": ["pkg/ui/progressbar.go"],
    "Construction": ["pkg/ui/construction.go", "pkg/ui/constructionoverhead.go", "pkg/ui/constructionsignal.go",
                     "pkg/ui/constructionstation.go", "pkg/ui/constructiontab.go", "cmd/goloco/construction.go",
                     "cmd/goloco/stationtool.go"],
}

LABEL = re.compile(r"case\s+(?:[A-Za-z_]+::)*[wW]idx::([A-Za-z_0-9]+)\s*:")
# goloco's markers: "clickable(Window.cpp): name name …" where the widget works as upstream's does,
# "clickable-partial(Window.cpp): …" where it does something but not yet what upstream does.
MARKER = re.compile(r"clickable(-partial)?\(([A-Za-z0-9_./]+)\):\s*([A-Za-z0-9_ ]+)")


def upstream_windows(root):
    """{window: set(widget names)} — a folder of files (Construction, News) is one window."""
    out = {}
    for name in sorted(os.listdir(root)):
        path = os.path.join(root, name)
        files = [path] if name.endswith(".cpp") else \
            [os.path.join(path, f) for f in sorted(os.listdir(path)) if f.endswith(".cpp")] if os.path.isdir(path) else []
        labels = set()
        for f in files:
            labels |= set(LABEL.findall(open(f, encoding="utf-8", errors="replace").read()))
        if labels:
            out[name] = labels
    return out


def goloco_markers(root):
    """{window: (set full, set partial)} from the markers in goloco's Go files."""
    out = {}
    for d in ("pkg", "cmd"):
        for dirpath, _, files in os.walk(os.path.join(root, d)):
            for f in files:
                if not f.endswith(".go"):
                    continue
                for m in MARKER.finditer(open(os.path.join(dirpath, f), encoding="utf-8").read()):
                    full, partial = out.setdefault(m.group(2), (set(), set()))
                    (partial if m.group(1) else full).update(m.group(3).split())
    return out


def main():
    args = [a for a in sys.argv[1:] if not a.startswith("--")]
    root = args[0] if args else UPSTREAM
    show_missing = "--missing" in sys.argv
    marks = goloco_markers(GOLOCO)
    total = have = part = 0
    rows = []
    for win, labels in upstream_windows(root).items():
        full, partial = marks.get(win, (set(), set()))
        full, partial = full & labels, (partial & labels) - full
        total += len(labels)
        have += len(full)
        part += len(partial)
        rows.append((win, len(full), len(partial), len(labels), win in marks, sorted(labels - full - partial)))
    rows.sort(key=lambda r: (-r[4], (r[1] + r[2]) / r[3], -r[3]))
    print(f"{'window':28} {'works':>5} {'part':>4} / {'OpenLoco':<8}")
    for win, n, p, m, annotated, missing in rows:
        note = "" if annotated else "  (not annotated yet)"
        print(f"{win:28} {n:5} {p:4} / {m:<8}{note}" + ("  missing: " + ", ".join(missing) if show_missing and annotated and missing else ""))
    print(f"\n{'all windows':28} {have:5} {part:4} / {total:<8} ({100 * have / total:.0f}% working as upstream, {100 * part / total:.0f}% partly)")
    print(f"annotated: {sum(1 for r in rows if r[4])} of {len(rows)} windows")


if __name__ == "__main__":
    main()
