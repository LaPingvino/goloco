package ui

import "github.com/LaPingvino/goloco/pkg/gfx8"

// extColourUnk33 is ExtColour::unk33, the translucent palette map laid over
// inactive tabs.
const extColourUnk33 = 0x33

// tabNoIcon is Widget::kContentUnk as a tab's content: no icon, but the
// inactive overlay and line.
const tabNoIcon = -2

// drawTab is Tab::draw with Widget::drawTab's content at (x, y) of rt:
//   - background: ImageIds::tab (+1 when active) in the window's secondary
//     colour (sec);
//   - content (icon >= 0): active, the icon at (x, y); inactive, the icon one
//     pixel lower, the tab image drawn translucent (ExtColour::unk33) over it,
//     and a 31×1 line in shade 7 at y + 26.
//
// iconMap is the icon's palette map (nil: none). With icon < 0 only the
// background is drawn (a tab with ImageIds::tab as its image); with tabNoIcon
// the inactive overlay is drawn without an icon (kContentUnk).
func drawTab(rt *gfx8.RenderTarget, x, y int, active bool, sec []uint8, icon int, iconMap []uint8) {
	g1 := globalG1Font.g1
	bg := imgTab
	if active {
		bg++
	}
	if sp, err := gfx8.Decode(&g1.Elements[bg]); err == nil {
		gfx8.DrawSpriteNoise(rt, sp, x, y, gfx8.OpTransparent|gfx8.OpSrc, sec, nil)
	}
	if icon < 0 && icon != tabNoIcon {
		return
	}
	op := func(sp *gfx8.Sprite) gfx8.BlendOp { return gfx8.OpFor(sp, iconMap != nil, false) }
	if active {
		if icon < 0 {
			return
		}
		if sp, err := gfx8.Decode(&g1.Elements[icon]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, x, y, op(sp), iconMap, nil)
		}
		return
	}
	if icon >= 0 {
		if sp, err := gfx8.Decode(&g1.Elements[icon]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, x, y+1, op(sp), iconMap, nil)
		}
	}
	if sp, err := gfx8.Decode(&g1.Elements[imgTab]); err == nil {
		if m, err := g1.GetPaletteMap(extColourUnk33); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, x, y, gfx8.OpFor(sp, false, true), m, nil)
		}
	}
	gfx8.FillRect(rt, x, y+26, x+31-1, y+26, gfx8.Shade(sec, 7))
}

// drawTabDisabled is drawTabBackground for a disabled tab (opaque colour):
// the tab's silhouette solid in shade 6 one pixel down-right, then in shade
// 4 over it (drawImageSolid).
func drawTabDisabled(rt *gfx8.RenderTarget, x, y int, sec []uint8) {
	sp, err := gfx8.Decode(&globalG1Font.g1.Elements[imgTab])
	if err != nil {
		return
	}
	solid := func(c uint8) []uint8 {
		m := make([]uint8, 256)
		for i := range m {
			m[i] = c
		}
		return m
	}
	gfx8.DrawSprite(rt, sp, x+1, y+1, gfx8.OpTransparent|gfx8.OpSrc, solid(gfx8.Shade(sec, 6)))
	gfx8.DrawSprite(rt, sp, x, y, gfx8.OpTransparent|gfx8.OpSrc, solid(gfx8.Shade(sec, 4)))
}
