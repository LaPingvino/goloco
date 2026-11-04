# XTest via ctypes: move the pointer to (x, y) and turn the wheel n clicks (button 5 = down = zoom out)
import ctypes, sys, time
x11 = ctypes.CDLL("libX11.so.6"); xt = ctypes.CDLL("libXtst.so.6")
x11.XOpenDisplay.restype = ctypes.c_void_p
d = x11.XOpenDisplay(None)
x, y, n, btn = int(sys.argv[1]), int(sys.argv[2]), int(sys.argv[3]), int(sys.argv[4]) if len(sys.argv) > 4 else 5
xt.XTestFakeMotionEvent(ctypes.c_void_p(d), -1, x, y, 0); x11.XFlush(ctypes.c_void_p(d)); time.sleep(0.3)
for _ in range(n):
    xt.XTestFakeButtonEvent(ctypes.c_void_p(d), btn, 1, 0); x11.XFlush(ctypes.c_void_p(d)); time.sleep(0.05)
    xt.XTestFakeButtonEvent(ctypes.c_void_p(d), btn, 0, 0); x11.XFlush(ctypes.c_void_p(d)); time.sleep(0.4)
