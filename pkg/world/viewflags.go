package world

import (
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/scene"
)

// The main viewport's flags (WindowManager::getMainWindow()->viewports[0]->
// flags): what the View menu and its shortcuts toggle, what the construction
// and terraform windows show and hide. Every change redraws the view
// (mainWindow->invalidate()).

// mainView is the main viewport as a paint view without a size: its
// position, zoom and flags — what the hit tests (getMapCoordinatesFromPos)
// paint their one pixel with.
func (w *World) mainView() scene.View {
	return scene.View{X: int(w.camX), Y: int(w.camY), Zoom: w.zoom, Flags: w.viewFlags}
}

// ViewFlags are the main viewport's flags.
func (w *World) ViewFlags() paint.ViewFlags { return w.viewFlags }

// ToggleViewFlags is viewMenuDropdown's / Shortcuts.cpp's toggle:
// viewport->flags ^= flags, the view redrawn.
func (w *World) ToggleViewFlags(flags paint.ViewFlags) {
	w.viewFlags ^= flags
	w.InvalidateView()
}

// SetViewFlags sets (on) or clears flags; the view is redrawn only when that
// changed something (Main::showGridlines / hideGridlines, showDirectionArrows
// / hideDirectionArrows, WindowManager::viewportSetVisibility's cases).
func (w *World) SetViewFlags(flags paint.ViewFlags, on bool) {
	f := w.viewFlags &^ flags
	if on {
		f = w.viewFlags | flags
	}
	if f == w.viewFlags {
		return
	}
	w.viewFlags = f
	w.InvalidateView()
}

// ghostVisibility is placeTrackGhost's / placeRoadGhost's
// viewportSetVisibility calls (sim.State.GhostVisibility).
func (w *World) ghostVisibility(underground, sloped bool) {
	if underground {
		w.ViewportSetVisibility(VisibilityUndergroundView)
	} else {
		w.ViewportSetVisibility(VisibilityOvergroundView)
	}
	if sloped {
		w.ViewportSetVisibility(VisibilityHeightMarksOnTrack)
	}
}

// SetGridlines is Main::showGridlines / hideGridlines: the main view's
// gridlines_on_landscape flag.
func (w *World) SetGridlines(on bool) { w.SetViewFlags(paint.ViewGridlines, on) }

// ViewportVisibility is WindowManager::ViewportVisibility.
type ViewportVisibility uint8

const (
	VisibilityReset ViewportVisibility = iota
	VisibilityUndergroundView
	VisibilityHeightMarksOnTrack
	VisibilityOvergroundView
)

// ViewportSetVisibility is WindowManager::viewportSetVisibility (0x004CF456's
// neighbour): the construction window's hold on the main view — underground
// while the ghost is under the ground, height marks while it slopes, and on
// reset (the window closing) the underground view, every see-through mode and
// both height marks off, whoever set them.
func (w *World) ViewportSetVisibility(v ViewportVisibility) {
	switch v {
	case VisibilityUndergroundView:
		w.SetViewFlags(paint.ViewUnderground, true)
	case VisibilityHeightMarksOnTrack:
		w.SetViewFlags(paint.ViewHeightMarksOnTracks, true)
	case VisibilityOvergroundView:
		w.SetViewFlags(paint.ViewUnderground, false)
	default:
		w.SetViewFlags(paint.ViewUnderground|paint.ViewSeeThroughAny|paint.ViewHeightMarksOnLand|paint.ViewHeightMarksOnTracks, false)
	}
}
