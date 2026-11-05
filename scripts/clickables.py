#!/usr/bin/env python3
"""Completion measure: OpenLoco's clickables against goloco's.

For every window in OpenLoco (src/OpenLoco/src/Ui/Windows/**.cpp) it lists the
widgets the window's event handlers act on — each `case …widx::name:` label —
and counts one as present in goloco when that widget's name appears in the
goloco file(s) the window is ported in (goloco's ports cite upstream's widget
names: "Widx::show_fps", "display_scale_down_btn" …). A window with no goloco
file counts as none present.

It is a measure, not a proof: a name can be mentioned without being handled,
and a handled widget whose port doesn't name it is missed. Use it for the
trend and the per-window to-do lists; check a window by hand before calling
it done.

    python3 scripts/clickables.py [OPENLOCO_SRC] [--missing]
"""
import os, re, sys

UPSTREAM = os.path.expanduser("~/Projects/OpenLoco/src/OpenLoco/src/Ui/Windows")
GOLOCO = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

# OpenLoco window file → goloco files that port it.
PORTED = {
    "Options.cpp": ["pkg/ui/options.go"],
    "TownList.cpp": ["pkg/ui/townlist.go", "cmd/goloco/listwindows.go"],
    "StationList.cpp": ["pkg/ui/townlist.go", "cmd/goloco/listwindows.go"],
    "VehicleList.cpp": ["pkg/ui/townlist.go", "pkg/ui/vehiclelistfilter.go", "cmd/goloco/listwindows.go"],
    "BuildVehicle.cpp": ["pkg/ui/buildvehicle.go", "cmd/goloco/listwindows.go"],
    "Vehicle.cpp": ["pkg/ui/vehiclewindow.go", "cmd/goloco/listwindows.go"],
    "ToolbarTop.cpp": ["pkg/ui/toolbar.go", "cmd/goloco/main.go", "cmd/goloco/listwindows.go"],
    "ToolbarTopCommon.cpp": ["pkg/ui/toolbar.go", "cmd/goloco/main.go"],
    "TitleMenu.cpp": ["pkg/ui/title_menu.go", "cmd/goloco/main.go"],
    "TitleOptions.cpp": ["pkg/ui/title_menu.go", "cmd/goloco/main.go"],
    "TitleExit.cpp": ["pkg/ui/title_menu.go", "cmd/goloco/main.go"],
    "ScenarioSelect.cpp": ["pkg/ui/scenarioselect.go"],
    "PromptBrowseWindow.cpp": ["pkg/ui/promptbrowse.go"],
    "PromptOkCancelWindow.cpp": ["pkg/ui/promptokcancel.go"],
    "TimePanel.cpp": ["pkg/ui/timepanel.go", "cmd/goloco/main.go"],
    "PlayerInfoPanel.cpp": ["pkg/ui/playerpanel.go", "cmd/goloco/main.go"],
    "CompanyFaceSelection.cpp": ["pkg/ui/companyselect.go"],
    "ProgressBar.cpp": ["pkg/ui/progressbar.go"],
    "Construction": ["pkg/ui/construction.go", "pkg/ui/constructionoverhead.go", "pkg/ui/constructionsignal.go",
                     "pkg/ui/constructionstation.go", "pkg/ui/constructiontab.go", "cmd/goloco/construction.go"],
}

LABEL = re.compile(r"case\s+(?:[A-Za-z_]+::)*[wW]idx::([a-z_0-9]+)\s*:")


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


def main():
    args = [a for a in sys.argv[1:] if not a.startswith("--")]
    root = args[0] if args else UPSTREAM
    show_missing = "--missing" in sys.argv
    total = have = 0
    rows = []
    for win, labels in upstream_windows(root).items():
        text = ""
        for f in PORTED.get(win, []):
            p = os.path.join(GOLOCO, f)
            if os.path.exists(p):
                text += open(p, encoding="utf-8").read()
        present = {l for l in labels if text and re.search(r"\b" + re.escape(l) + r"\b", text)}
        total += len(labels)
        have += len(present)
        rows.append((win, len(present), len(labels), sorted(labels - present)))
    rows.sort(key=lambda r: (r[1] / r[2], -r[2]))
    print(f"{'window':28} {'goloco':>6} / {'OpenLoco':<8}")
    for win, n, m, missing in rows:
        print(f"{win:28} {n:6} / {m:<8}" + ("  missing: " + ", ".join(missing) if show_missing and missing else ""))
    print(f"\n{'all windows':28} {have:6} / {total:<8} ({100 * have / total:.0f}%)")


if __name__ == "__main__":
    main()
