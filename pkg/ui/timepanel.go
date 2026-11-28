package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/paint"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
)

// ImageIds::speed_* (the _active variant is +1) and the interface skin's
// toolbar_menu_map_north (InterfaceSkin::ImageIds; west, south and east
// follow: map_sprites_by_rotation).
const (
	imgSpeedPause     = 2460
	imgSpeedNormal    = 2462
	imgSpeedFast      = 2464
	imgSpeedExtraFast = 2466
	skinImgMapNorth   = 466
	extColourUnk34    = 0x34
)

// TimePanelState is what TimePanel.cpp draws from.
type TimePanelState struct {
	Date       string // formatDateDMYFull of the current day
	Speed      int    // 0 paused, 1 normal, 2 fast forward, 3 extra fast forward
	Hover      int    // the hovered widget (uilayout.TimePanelWidx*; −1 none): its draw_3 frame, the date in white
	SkinColour uint8  // InterfaceSkinObject::timeToolbarColour
	SkinImage  int    // InterfaceSkinObject image base (map globe); <=0: none
	Rotation   uint8  // WindowManager::getCurrentRotation: which way the globe's arrow points
}

// DrawTimePanel is TimePanel.cpp draw for the window at the bottom-right: the
// unk34 dim over the outer frame, the two Wt3 panels and the speed buttons
// (translucent window colour), the inset border, the centred date and the map
// globe. The hovered image button (Input::isHovering → widgetState.hovered)
// is drawn by ImageButton's draw_3: the translucent inset frame, then its
// image. PARTIAL: a pressed button's borderInset (activated) is not drawn.
func DrawTimePanel(screen *ebiten.Image, s TimePanelState) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	wx, wy := L.TimePanelPos(sw, sh)
	cmap, err := g1.GetPaletteMap(int(s.SkinColour))
	if err != nil {
		return
	}
	var trans [3][]uint8
	for k := range trans {
		trans[k], _ = g1.GetPaletteMap(int(paintTranslucent(s.SkinColour, k)))
	}
	tc := gfx8.InsetColour{Map: cmap, Translucent: true, Trans: trans}
	m34, _ := g1.GetPaletteMap(extColourUnk34)
	images := [4]int{imgSpeedPause, imgSpeedNormal, imgSpeedFast, imgSpeedExtraFast}
	draw8(screen, wx, wy, L.TimePanelW, L.TimePanelH, func(rt *gfx8.RenderTarget) {
		o := L.TimePanelOuter
		gfx8.FillRectTranslucent(rt, o[0], o[1], o[0]+o[2]-1, o[1]+o[3]-1, m34)
		// the outer_frame Wt3 (prepareDraw hides inner_frame)
		gfx8.FillRectInset(rt, o[0], o[1], o[0]+o[2]-1, o[1]+o[3]-1, tc, 0)
		hoverFrame := func(r [4]int) { // ImageButton draw_3: fillRectInset(size, colour.translucent(), flags)
			gfx8.FillRectInset(rt, r[0], r[1], r[0]+r[2]-1, r[1]+r[3]-1, tc, 0)
		}
		switch s.Hover { // map_chat_menu and date_btn have no image (kContentNull)
		case L.TimePanelWidxMapChat:
			hoverFrame(L.TimePanelMapChat)
		case L.TimePanelWidxDate:
			hoverFrame(L.TimePanelDate)
		}
		for i, r := range L.TimePanelSpeed {
			if s.Hover == L.TimePanelWidxPause+i {
				hoverFrame(r)
			}
			img := images[i]
			if i == s.Speed {
				img++ // the active variant
			}
			if sp, err := gfx8.Decode(&g1.Elements[img]); err == nil {
				// ImageButton::drawImage: recolour(image, window colour).
				gfx8.DrawSpriteNoise(rt, sp, r[0], r[1], gfx8.OpTransparent|gfx8.OpSrc, cmap, nil)
			}
		}
		// drawRectInset(frame + 1, size - 2, secondary colour, borderInset | fillNone)
		gfx8.FillRectInset(rt, o[0]+1, o[1]+1, o[0]+1+o[2]-2-1, o[1]+1+o[3]-2-1, tc, gfx8.InsetBorderInset|gfx8.InsetFillNone)
		if s.SkinImage > 0 {
			mc := L.TimePanelMapChat
			if sp, err := gfx8.Decode(&g1.Elements[s.SkinImage+skinImgMapNorth+int(s.Rotation&3)]); err == nil {
				gfx8.DrawSpriteNoise(rt, sp, mc[0]-2, mc[1]-1, gfx8.OpFor(sp, false, false), nil, nil)
			}
		}
	})
	pal := g1.GetPalette()
	c := pal[gfx8.Shade(cmap, 9)]
	if s.Hover == L.TimePanelWidxDate {
		if wm, err := g1.GetPaletteMap(colourWhite); err == nil {
			c = pal[gfx8.Shade(wm, 9)]
		}
	}
	d := L.TimePanelDate
	tw, _ := MeasureTextBold(s.Date)
	DrawTextBold(screen, s.Date, wx+d[0]+d[2]/2-tw/2, wy+d[1]+1, c)
}

func paintTranslucent(c uint8, shade int) uint8 {
	if int(c) >= len(paint.TranslucentColour) {
		return 0
	}
	return paint.TranslucentColour[c][shade]
}
