package ui

import (
	"fmt"
	"image/color"
	"math/bits"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
)

// The Build Vehicle window's details panel (Ui/Windows/BuildVehicle.cpp draw
// 0x4C2F23, from `if (self.rowHover == -1)` on): buffer_1250 built up from
// the stats strings for the hovered vehicle and drawn wrapped below the
// preview.

// length2DP is {INT32_2DP} of StringManager::internalLengthToComma1DP
// (Formatting.cpp: length × 100 / 4 / 32): a vehicle's length in tiles, two
// decimals.
func length2DP(length int) string {
	n := length * 100 / 4 / 32
	return fmt.Sprintf("%d.%02d", n/100, n%100)
}

// VehicleStatsArgs are what draw reads beside the vehicle object: its
// inflation-adjusted cost (getInflationAdjustedCost(costFactor, costIndex, 6))
// and running cost (runCostFactor, runCostIndex, 10), and whether the
// controlling company has it unlocked (Company::isVehicleIndexUnlocked).
type VehicleStatsArgs struct {
	Cost, RunningCost int64
	Unlocked          bool
}

// VehicleStats is draw's buffer_1250 for a vehicle object, the strings
// formatted (FillArgs) in upstream's order: stats_cost, stats_running_cost,
// stats_designed (stats_proposed_design when locked) for a design year,
// stats_obsolete unless 0 or 0xFFFF; on rail and road stats_requires and the
// track's (road's; StringIds::road for any road) name, " + " and each
// required extra's name, stats_string_steep_slope with the rack rail;
// stats_power (rail and road, when it has power), stats_length, stats_weight,
// stats_max_speed, stats_velocity_on_string with the rack rail; then
// VehicleObject::getCargoString.
func VehicleStats(v *objects.VehicleObject, om *objects.ObjectManager, a VehicleStatsArgs) string {
	var b strings.Builder
	b.WriteString(FillArgs(i18n.Str(i18n.StatsCost), currency32(a.Cost)))
	b.WriteString(FillArgs(i18n.Str(i18n.StatsRunningCost), currency32(a.RunningCost)))
	if v.Designed != 0 {
		id := i18n.StatsProposedDesign
		if a.Unlocked {
			id = i18n.StatsDesigned
		}
		b.WriteString(FillArgs(i18n.Str(id), strconv.Itoa(int(v.Designed))))
	}
	if v.Obsolete != 0 && v.Obsolete != 0xFFFF {
		b.WriteString(FillArgs(i18n.Str(i18n.StatsObsolete), strconv.Itoa(int(v.Obsolete))))
	}
	landed := v.Mode == objects.TransportModeRail || v.Mode == objects.TransportModeRoad
	rackName := ""
	if v.Flags&objects.VehicleFlagRackRail != 0 && om != nil {
		rackName = om.ExtraName(v.RackRailName) // ObjectManager::get<TrackExtraObject>(rackRailType)->name
	}
	if landed {
		b.WriteString(i18n.Str(i18n.StatsRequires))
		trackName := i18n.Str(i18n.Road)
		if om != nil {
			if t := om.TrackTypeFor(v); v.Mode == objects.TransportModeRoad {
				if t != 0xFF {
					if r := om.GetRoadObjectByIndex(int(t)); r != nil {
						trackName = r.Name
					}
				}
			} else if tr := om.GetTrackObjectByIndex(int(t)); tr != nil {
				trackName = tr.Name
			}
		}
		b.WriteString(trackName)
		for i := 0; i < int(v.NumTrackExtras) && i < len(v.ExtraNames); i++ {
			b.WriteString(" + ")
			if om != nil {
				b.WriteString(om.ExtraName(v.ExtraNames[i]))
			}
		}
		if v.Flags&objects.VehicleFlagRackRail != 0 {
			b.WriteString(FillArgs(i18n.Str(i18n.StatsStringSteepSlope), rackName))
		}
	}
	if v.Power != 0 && landed {
		b.WriteString(FillArgs(i18n.Str(i18n.StatsPower), L.Power(int(v.Power))))
	}
	b.WriteString(FillArgs(i18n.Str(i18n.StatsLength), length2DP(v.Length())))
	b.WriteString(FillArgs(i18n.Str(i18n.StatsWeight), strconv.Itoa(int(v.Weight))))
	b.WriteString(FillArgs(i18n.Str(i18n.StatsMaxSpeed), L.Velocity(int(v.Speed))))
	if v.Flags&objects.VehicleFlagRackRail != 0 {
		b.WriteString(FillArgs(i18n.Str(i18n.StatsVelocityOnString), L.Velocity(int(v.RackSpeed)), rackName))
	}
	if om != nil {
		b.WriteString(VehicleCargoString(v, om))
	}
	return b.String()
}

// VehicleCargoString is VehicleObject::getCargoString: for each cargo slot
// (the first, then the second when it carries two at once) its capacity
// with the first cargo's unit name (singular for 1) — stats_capacity, then
// stats_plus_string —, the slot's other cargoes " (or A  or B)" (each
// stats_or_string and a space, a space before every one but the first, the
// last space turned into ')'), and after the first slot stats_refittable for
// a refittable vehicle.
func VehicleCargoString(v *objects.VehicleObject, om *objects.ObjectManager) string {
	c := om.CargoFor(v)
	if c.NumSimultaneousCargoTypes == 0 {
		return ""
	}
	var b strings.Builder
	slot := func(i int, capacityID int) {
		cats := c.CompatibleCargoCategories[i]
		if cats == 0 {
			return
		}
		first := bits.TrailingZeros32(cats)
		others := cats &^ (1 << first)
		if first < len(om.CargoObjects) && om.CargoObjects[first] != nil {
			co := om.CargoObjects[first]
			unit := co.UnitNamePlural
			if c.MaxCargo[i] == 1 {
				unit = co.UnitNameSingular
			}
			b.WriteString(FillArgs(i18n.Str(capacityID), ObjectString(unit), Comma32(int64(c.MaxCargo[i]))))
		}
		if others == 0 {
			return
		}
		s := " ("
		for others != 0 {
			ct := bits.TrailingZeros32(others)
			others &^= 1 << ct
			if !strings.HasSuffix(s, "(") {
				s += " "
			}
			name := ""
			if ct < len(om.CargoObjects) && om.CargoObjects[ct] != nil {
				name = om.CargoObjects[ct].Name
			}
			s += FillArgs(i18n.Str(i18n.StatsOrString), name) + " "
		}
		b.WriteString(s[:len(s)-1] + ")") // buffer[-1] = ')'
	}
	slot(0, i18n.StatsCapacity)
	if v.Flags&objects.VehicleFlagRefittable != 0 {
		b.WriteString(i18n.Str(i18n.StatsRefittable))
	}
	if c.NumSimultaneousCargoTypes > 1 {
		slot(1, i18n.StatsPlusString)
	}
	return b.String()
}

// Runes standing for the colour codes in a coded string laid out by
// wrapLines (below 32: no width, and not the newline).
const (
	codeWindow2 = 1 // {COLOUR WINDOW_2} (and any other colour but black)
	codeBlack   = 2 // {COLOUR BLACK}
)

// codedRunes turns a game string's codes into wrapLines' runes: {NEWLINE}
// the newline (5), the colours codeWindow2 / codeBlack; other codes dropped.
func codedRunes(s string) string {
	var b strings.Builder
	for len(s) > 0 {
		i := strings.IndexByte(s, '{')
		if i < 0 {
			b.WriteString(s)
			break
		}
		b.WriteString(s[:i])
		end := strings.IndexByte(s[i:], '}')
		if end < 0 {
			b.WriteString(s[i:])
			break
		}
		switch code := s[i+1 : i+end]; {
		case code == "NEWLINE":
			b.WriteRune(controlNewline)
		case code == "COLOUR BLACK":
			b.WriteRune(codeBlack)
		case strings.HasPrefix(code, "COLOUR"):
			b.WriteRune(codeWindow2)
		}
		s = s[i+end+1:]
	}
	return b.String()
}

// WrapCoded is wrapString on a coded string: its lines for a width in a
// font, each still holding its colour runes.
func WrapCoded(s string, font, width int) []string {
	return WrapString(codedRunes(s), font, width)
}

// drawCodedWrapped is drawStringLeftWrapped for a string with colour codes
// (in medium bold, starting in black): wrapString's lines, a line height
// apart, the colour carried from one line into the next as the text
// renderer keeps it. Returns the y below the last line.
func drawCodedWrapped(screen *ebiten.Image, s string, x, y, width int, win2, black color.Color) int {
	const font = G1FontMediumBold
	col := black
	for _, line := range WrapCoded(s, font, width) {
		px := x
		seg := []rune{}
		flush := func() {
			if len(seg) > 0 {
				px += DrawTextFont(screen, string(seg), font, px, y, col)
				seg = seg[:0]
			}
		}
		for _, r := range line {
			switch r {
			case codeWindow2:
				flush()
				col = win2
			case codeBlack:
				flush()
				col = black
			default:
				if r >= 32 {
					seg = append(seg, r)
				}
			}
		}
		flush()
		y += g1FontHeight(font)
	}
	return y
}

// drawInlineCentredClipped is drawStringCentredClipped (TextRenderer.cpp) in
// medium bold for a text followed by inline sprites (a space and the image
// for each: ControlCodes::inlineSpriteStr, the pen moving by the image's
// width): clipString's cut ("..." after the longest prefix narrower than
// width, when it doesn't fit), centred on x by the clipped width.
func drawInlineCentredClipped(screen *ebiten.Image, text string, sprites []int, x, y, width int, clr color.Color) {
	const font = G1FontMediumBold
	g1 := globalG1Font.g1
	type tok struct {
		r      rune
		sprite int // -1: a character
	}
	var toks []tok
	for _, r := range text {
		toks = append(toks, tok{r, -1})
	}
	for _, s := range sprites {
		if s > 0 && s < len(g1.Elements) {
			toks = append(toks, tok{' ', -1}, tok{0, s})
		}
	}
	tw := func(t tok) int {
		if t.sprite >= 0 {
			return int(g1.Elements[t.sprite].Width)
		}
		return MeasureTextFont(string(t.r), font)
	}
	total := 0
	for _, t := range toks {
		total += tw(t)
	}
	dots := ""
	if total > width {
		if width < 6 {
			return
		}
		dw := MeasureTextFont("...", font)
		n, w := 0, 0
		for n < len(toks) && w+tw(toks[n])+dw < width {
			w += tw(toks[n])
			n++
		}
		toks, dots, total = toks[:n], "...", w+dw
	}
	px := x - total/2
	var seg []rune
	flush := func() {
		if len(seg) > 0 {
			px += DrawTextFont(screen, string(seg), font, px, y, clr)
			seg = seg[:0]
		}
	}
	for _, t := range toks {
		if t.sprite < 0 {
			seg = append(seg, t.r)
			continue
		}
		flush()
		if sp, err := gfx8.Decode(&g1.Elements[t.sprite]); err == nil {
			draw8(screen, px+sp.XOffset, y+sp.YOffset, sp.Width, sp.Height, func(rt *gfx8.RenderTarget) {
				gfx8.DrawSpriteNoise(rt, sp, -sp.XOffset, -sp.YOffset, gfx8.OpFor(sp, false, false), nil, nil)
			})
		}
		px += tw(t)
	}
	flush()
	DrawTextFont(screen, dots, font, px, y, clr)
}
