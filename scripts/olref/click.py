import ctypes, sys, time
x11 = ctypes.CDLL("libX11.so.6"); xt = ctypes.CDLL("libXtst.so.6")
x11.XOpenDisplay.restype = ctypes.c_void_p
d = ctypes.c_void_p(x11.XOpenDisplay(None))
x, y = int(sys.argv[1]), int(sys.argv[2])
xt.XTestFakeMotionEvent(d, -1, x, y, 0); x11.XFlush(d); time.sleep(0.4)
xt.XTestFakeButtonEvent(d, 1, 1, 0); x11.XFlush(d); time.sleep(0.1)
xt.XTestFakeButtonEvent(d, 1, 0, 0); x11.XFlush(d); time.sleep(0.3)
