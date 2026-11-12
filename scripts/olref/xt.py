"""Mouse and keyboard for the real OpenLoco, through X11's XTest extension.

    xt.py move X Y | click X Y | rightclick X Y | press X Y | release X Y
    xt.py drag X1 Y1 X2 Y2 | wheel DY | key NAME | type TEXT | wait SECONDS
    xt.py shot FILE.png | state
    xt.py run SCRIPT.txt

The commands are goloco's GOLOCO_SCRIPT ones (cmd/goloco/script.go), so one
script can drive both games; `run` plays a file of them ('#' comments). seed,
scenario and camera reach into goloco and have no counterpart here: `run`
stops at them. `state` prints the pointer and the buttons the X server holds
down (0 = none) — the first thing to look at when input seems stuck.

Coordinates are game-window pixels: OLREF_ORIGIN (default 112,84 — an 800x600
window centred in ol.sh's 1024x768 screen) is added on the way to the server
and cropped away by shot.

ctypes on libX11/libXtst, nothing to install. DISPLAY picks the server (ol.sh
sets it). Every command leaves all buttons released.
"""
import ctypes
import os
import subprocess
import sys
import time

x11 = ctypes.CDLL("libX11.so.6")
xtst = ctypes.CDLL("libXtst.so.6")
x11.XOpenDisplay.restype = ctypes.c_void_p
x11.XDefaultRootWindow.restype = ctypes.c_ulong
x11.XDefaultRootWindow.argtypes = [ctypes.c_void_p]
x11.XStringToKeysym.restype = ctypes.c_ulong
x11.XKeysymToKeycode.argtypes = [ctypes.c_void_p, ctypes.c_ulong]
x11.XKeysymToKeycode.restype = ctypes.c_ubyte
x11.XSync.argtypes = [ctypes.c_void_p, ctypes.c_int]
xtst.XTestFakeMotionEvent.argtypes = [ctypes.c_void_p, ctypes.c_int, ctypes.c_int, ctypes.c_int, ctypes.c_ulong]
xtst.XTestFakeButtonEvent.argtypes = [ctypes.c_void_p, ctypes.c_uint, ctypes.c_int, ctypes.c_ulong]
xtst.XTestFakeKeyEvent.argtypes = [ctypes.c_void_p, ctypes.c_uint, ctypes.c_int, ctypes.c_ulong]

dpy = x11.XOpenDisplay(None)
if not dpy:
    sys.exit("xt.py: cannot open the X display (DISPLAY unset, or run from inside a sandbox?)")

# OpenLoco reads its input once per tick (25 ms); a press and its release in
# one tick are both seen, but a move needs a tick before the press that
# follows it lands on the widget under it.
TICK = 0.1
OX, OY = (int(v) for v in os.environ.get("OLREF_ORIGIN", "112,84").split(","))

KEYS = {"ESC": "Escape", "ESCAPE": "Escape", "ENTER": "Return", "SPACE": "space",
        "BACKSPACE": "BackSpace", "TAB": "Tab", "UP": "Up", "DOWN": "Down",
        "LEFT": "Left", "RIGHT": "Right"}
CHARS = {" ": "space", ".": "period", ",": "comma", "-": "minus", "'": "apostrophe"}


def sync(pause=0.0):
    x11.XSync(dpy, 0)
    if pause:
        time.sleep(pause)


def move(x, y):
    xtst.XTestFakeMotionEvent(dpy, -1, x + OX, y + OY, 0)
    sync(TICK)


def button(b, down):
    xtst.XTestFakeButtonEvent(dpy, b, 1 if down else 0, 0)
    sync(TICK)


def tap(keysym_name, shift=False):
    code = x11.XKeysymToKeycode(dpy, x11.XStringToKeysym(keysym_name.encode()))
    if not code:
        sys.exit(f"xt.py: no key {keysym_name!r}")
    shift_code = x11.XKeysymToKeycode(dpy, x11.XStringToKeysym(b"Shift_L"))
    if shift:
        xtst.XTestFakeKeyEvent(dpy, shift_code, 1, 0)
    xtst.XTestFakeKeyEvent(dpy, code, 1, 0)
    sync(0.05)
    xtst.XTestFakeKeyEvent(dpy, code, 0, 0)
    if shift:
        xtst.XTestFakeKeyEvent(dpy, shift_code, 0, 0)
    sync(0.05)


def state():
    root = ctypes.c_ulong(x11.XDefaultRootWindow(dpy))
    w1, w2 = ctypes.c_ulong(), ctypes.c_ulong()
    rx, ry, wx, wy = ctypes.c_int(), ctypes.c_int(), ctypes.c_int(), ctypes.c_int()
    mask = ctypes.c_uint()
    x11.XQueryPointer(ctypes.c_void_p(dpy), root, ctypes.byref(w1), ctypes.byref(w2), ctypes.byref(rx),
                      ctypes.byref(ry), ctypes.byref(wx), ctypes.byref(wy), ctypes.byref(mask))
    held = [b for b in range(1, 6) if mask.value & (0x80 << b)]
    return rx.value, ry.value, held


def do(cmd, args):
    n = lambda i: int(args[i])
    if cmd == "wait":
        time.sleep(float(args[0]))
    elif cmd == "move":
        move(n(0), n(1))
    elif cmd in ("click", "rightclick"):
        b = 1 if cmd == "click" else 3
        move(n(0), n(1))
        button(b, True)
        button(b, False)
    elif cmd == "press":
        move(n(0), n(1))
        button(1, True)
    elif cmd == "release":
        move(n(0), n(1))
        button(1, False)
    elif cmd == "drag":
        x1, y1, x2, y2 = n(0), n(1), n(2), n(3)
        move(x1, y1)
        button(1, True)
        for i in range(1, 6):
            move(x1 + (x2 - x1) * i // 5, y1 + (y2 - y1) * i // 5)
        button(1, False)
    elif cmd == "wheel":  # goloco's sign: negative = towards the player = X button 5
        dy = float(args[0])
        for _ in range(int(abs(dy))):
            b = 5 if dy < 0 else 4
            button(b, True)
            button(b, False)
    elif cmd == "key":
        name = args[0]
        tap(KEYS.get(name.upper(), name))
    elif cmd == "type":
        for ch in " ".join(args):
            tap(CHARS.get(ch, ch), shift=ch.isupper())
    elif cmd == "shot":
        subprocess.run(["import", "-window", "root", "-crop", f"800x600+{OX}+{OY}", "+repage", args[0]], check=True)
    elif cmd == "state":
        x, y, held = state()
        print(f"pointer {x - OX},{y - OY} buttons held: {held or 0}")
    elif cmd == "quit":
        return False
    else:
        sys.exit(f"xt.py: {cmd!r} is not something the real OpenLoco can be told")
    return True


def main(argv):
    if not argv:
        sys.exit(__doc__)
    try:
        if argv[0] == "run":
            for line in open(argv[1]):
                f = line.split("#")[0].split()
                if f and not do(f[0].lower(), f[1:]):
                    break
        else:
            do(argv[0].lower(), argv[1:])
    finally:  # never leave a button down behind us (a lone "press" means to)
        _, _, held = state()
        for b in [] if argv[0] == "press" else held:
            xtst.XTestFakeButtonEvent(dpy, b, 0, 0)
        sync()


main(sys.argv[1:])
