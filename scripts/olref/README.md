# Driving the real OpenLoco for reference shots

OpenLoco (the AUR release, /usr/bin/openloco) runs headless under Xvfb; these
scripts press its buttons through X11's XTest extension (ctypes on libXtst — no
xdotool needed). No game data here: OpenLoco reads the player's own Locomotion
folder.

    export XDG_CONFIG_HOME=$TMP/olcfg HOME=$TMP/olcfg SDL_AUDIODRIVER=dummy   # never the user's own config
    openloco --locomotion_path ~/Projects/locomotion &
    python3 click.py X Y        # a left click at screen X, Y
    python3 wheel.py X Y N [B]  # N wheel clicks at X, Y (button 5 = zoom out, 4 = in)
    import -window root shot.png

In a 1024x768 Xvfb screen with an 800x600 game window (set window_resolution in
the scratch openloco.yml) the window sits at (112, 84). Title → new game:
the first globe at (402, 624); the scenario list's first entry of the open tab at
(422, 237) starts it; the "Name Owner" box's OK at (637, 418); the New Company
window's close box at (374, 121); then the map takes wheel zooms. The title view
itself ignores the wheel.

The binary is the 26.08 release, older than the source we port: where it and
goloco differ, the source decides (LOOP.md).
