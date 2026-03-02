package world

// Static structural checks: verify world.go conforms to MEASUREMENTS.md.
// These run as part of `go test ./pkg/world/` so regressions are caught
// by the normal test suite without needing to remember to run the shell script.

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// worldGoSource reads world.go relative to this test file.
func worldGoSource(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	src, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "world.go"))
	if err != nil {
		t.Fatalf("reading world.go: %v", err)
	}
	return string(src)
}

// assertAbsent fails if pattern is found in src.
func assertAbsent(t *testing.T, src, pattern, msg string) {
	t.Helper()
	re := regexp.MustCompile(pattern)
	if m := re.FindString(src); m != "" {
		t.Errorf("BAD PATTERN found — %s\n  matched: %q", msg, m)
	}
}

// assertPresent fails if pattern is NOT found in src.
func assertPresent(t *testing.T, src, pattern, msg string) {
	t.Helper()
	re := regexp.MustCompile(pattern)
	if !re.MatchString(src) {
		t.Errorf("REQUIRED PATTERN absent — %s\n  pattern: %s", msg, pattern)
	}
}

// assertCaseFollowedBy checks that a switch case label is immediately followed
// (within the next few lines) by an assignment matching wantPat.
func assertCaseFollowedBy(t *testing.T, src, caseLabel, wantPat, msg string) {
	t.Helper()
	lines := strings.Split(src, "\n")
	for i, line := range lines {
		if strings.Contains(line, caseLabel) {
			// Check the next 3 lines for the assignment
			for j := i + 1; j < i+4 && j < len(lines); j++ {
				if regexp.MustCompile(wantPat).MatchString(lines[j]) {
					return // found
				}
			}
			t.Errorf("CASE MISMATCH — %s\n  case at line %d not followed by %q", msg, i+1, wantPat)
			return
		}
	}
	t.Errorf("CASE NOT FOUND — %s: could not find case label %q", msg, caseLabel)
}

func TestMeasurementChecks(t *testing.T) {
	src := worldGoSource(t)

	// --- tileToScreen ---
	t.Run("tileToScreen_X_not_mirrored", func(t *testing.T) {
		assertAbsent(t, src,
			`tileX-tileY|tileX - tileY`,
			"tileToScreen X must be (tileY-tileX)*32, not (tileX-tileY) — map would be mirrored")
	})
	t.Run("tileToScreen_no_bogus_X_offset", func(t *testing.T) {
		assertAbsent(t, src,
			`width\*tileW/4|width \* tileW / 4`,
			"Remove bogus centering offset; camera handles it")
	})
	t.Run("tileToScreen_no_plus50", func(t *testing.T) {
		assertAbsent(t, src,
			`\+ 50\b`,
			"Remove hardcoded +50 top margin; camera handles centering")
	})

	// --- Height factor ---
	t.Run("height_factor_not_times2", func(t *testing.T) {
		assertAbsent(t, src,
			`baseZ\) \* 2\.0`,
			"Height factor must be *4.0 (kSmallZStep=4), not *2.0")
	})
	t.Run("extraHeight_not_times2", func(t *testing.T) {
		assertAbsent(t, src,
			`extraHeight.*\* 2\.0`,
			"extraHeight for trees/buildings/walls must use *4.0, not *2.0")
	})

	// --- getCornerHeights ---
	t.Run("cornerHeights_divide_not_multiply", func(t *testing.T) {
		assertAbsent(t, src,
			`uint8\(t\.baseZ\) \* \d`,
			"getCornerHeights must use baseZ/4, not baseZ*N — MicroZ = SmallZ / kMicroToSmallZStep(4)")
	})
	t.Run("cornerHeights_divides_by_4", func(t *testing.T) {
		assertPresent(t, src,
			`uint8\(t\.baseZ\) / 4`,
			"getCornerHeights must have: uint8(t.baseZ) / 4")
	})

	// --- Neighbour offsets ---
	t.Run("EdgeSW_dx_plus1", func(t *testing.T) {
		assertCaseFollowedBy(t, src,
			"case EdgeSW:",
			`dx, dy = 1, 0`,
			"EdgeSW must set dx=+1,dy=0 (OpenLoco kNeighbourOffsets[0] = Pos2{+32,0} = tileX+1)")
	})
	t.Run("EdgeNE_dx_minus1", func(t *testing.T) {
		assertCaseFollowedBy(t, src,
			"case EdgeNE:",
			`dx, dy = -1, 0`,
			"EdgeNE must set dx=-1,dy=0 (OpenLoco kNeighbourOffsets[3] = Pos2{-32,0} = tileX-1)")
	})

	// --- getCliffEdgeOffset ---
	t.Run("cliffEdge_no_bare_int16_height", func(t *testing.T) {
		assertAbsent(t, src,
			`heightOffset := int16\(height\)`,
			"Cliff edge height must be h*16px (kMicroZStep=16), not plain int16(height)")
	})
	t.Run("cliffEdge_SW_offset_correct", func(t *testing.T) {
		assertPresent(t, src,
			`-30, 15 - heightPx`,
			"SW cliff edge must return (-30, 15-heightPx) — derived from Pos3{30,0,0} via gameToScreen")
	})

	// --- Water height ---
	t.Run("water_not_times2", func(t *testing.T) {
		assertAbsent(t, src,
			`waterHeightDiff.*\* 2`,
			"Water field is MicroZ (×16px), not ×2; use waterLevel*16.0 - baseZ*4.0")
	})
	t.Run("water_times16", func(t *testing.T) {
		assertPresent(t, src,
			`waterLevel\)?\s*\*\s*16\.0`,
			"Water height must multiply by 16.0 (kMicroZStep=16)")
	})

	// --- Checkerboard ---
	t.Run("checkerboard_not_0x20", func(t *testing.T) {
		assertAbsent(t, src,
			`\(x \^ y\) & 0x20|\(x\^y\) & 0x20`,
			"Checkerboard must be ((x^y)&1)*32, not &0x20 — &0x20 alternates every 32 tiles, not every tile")
	})
	t.Run("checkerboard_correct", func(t *testing.T) {
		assertPresent(t, src,
			`\(x\^y\)&1\) \* 32`,
			"Checkerboard must have ((x^y)&1)*32")
	})
}
