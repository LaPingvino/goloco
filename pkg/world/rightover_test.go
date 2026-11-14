package world

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/ui"
)

// The map tooltip strings as formatString fills them: the name first, then
// the second line in {COLOUR WINDOW_3}.
func TestTipFormat(t *testing.T) {
	if i18n.Str(i18n.StringidRightClickToModify) == "" {
		t.Skip("no language pack")
	}
	c3 := string(ui.MapTipColour3)
	for _, tc := range []struct{ got, want string }{
		{tipFormat(i18n.StringidRightClickToModify, "Road"), "Road" + tipNewline + c3 + "Right-Click to Modify"},
		{tipFormat(i18n.StringOwnedByString, "Railway Track", "Acme"), "Railway Track" + tipNewline + c3 + "Owned by Acme"},
		{tipFormat(i18n.StringidRightClickToRemove, tipFormat(i18n.StringStationPlatform, "Dallas")), "Dallas station platform" + tipNewline + c3 + "Right-Click to Remove"},
		{tipFormat(i18n.Accepts) + tipFormat(i18n.QuantityEigth+1) + "Passengers", tipNewline + c3 + "Accepts 1/4 Passengers"},
	} {
		if tc.got != tc.want {
			t.Errorf("got %q, want %q", tc.got, tc.want)
		}
	}
}
