package tests

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// fixture_stubs.go
//
// Test stubs that check for the presence of minimal test fixtures. Each
// test will fail with a clear message if the corresponding fixture(s)
// are missing. This makes running `go test` a reliable reminder that the
// fixture work needs to be implemented and expanded.
//
// These tests are intentionally small and test-only; they do not modify
// production code. They also include precise messages and suggested
// fixture names so the next engineer (or Claude) can add the files quickly.
//
// How paths are resolved:
// - Tests compute the path relative to this source file so they are
//   robust to working-directory differences when `go test` runs.
//
// Expected fixture layout (examples):
// - pkg/testdata/g1/g1_rle.bin
// - pkg/testdata/objects/interdef_small.dat
// - pkg/testdata/s5/rotate.bin
// - pkg/testdata/golden/title_golden.png

// testdataPath composes a path to pkg/testdata from this file's directory.
func testdataPath(parts ...string) string {
	_, fn, _, _ := runtime.Caller(0)
	dir := filepath.Dir(fn)
	all := []string{dir, "..", "testdata"}
	all = append(all, parts...)
	return filepath.Clean(filepath.Join(all...))
}

func TestG1FixturesPresence(t *testing.T) {
	// Expect at least one representative G1 fixture file to exist.
	// Suggested filename: pkg/testdata/g1/g1_rle.bin
	path := testdataPath("g1", "g1_rle.bin")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("G1 fixture missing: %s\nAdd a small synthetic G1 fixture at that path (or adjust the test). See pkg/assets/g1_parser.go and pkg/assets/g1_rle.go for parser expectations.", path)
	} else if err != nil {
		t.Fatalf("error stat'ing G1 fixture %s: %v", path, err)
	}
}

func TestInterfaceSkinFixturesPresence(t *testing.T) {
	// Expect a minimal InterfaceSkin/ObjData fixture.
	// Suggested filename: pkg/testdata/objects/interdef_small.dat
	path := testdataPath("objects", "interdef_small.dat")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("InterfaceSkin fixture missing: %s\nAdd a small INTERDEF-like fixture (minimal headers + string table + image table) to pkg/testdata/objects/ and re-run tests. See pkg/objects/interfaceskin.go for parser layout.", path)
	} else if err != nil {
		t.Fatalf("error stat'ing InterfaceSkin fixture %s: %v", path, err)
	}
}

func TestS5ExpandedFixturesPresence(t *testing.T) {
	// Expect example binary fixtures for S5 encodings. Suggested names:
	// pkg/testdata/s5/rotate.bin and pkg/testdata/s5/game_state.bin
	rotatePath := testdataPath("s5", "rotate.bin")
	gameStatePath := testdataPath("s5", "game_state.bin")

	missing := []string{}
	if _, err := os.Stat(rotatePath); os.IsNotExist(err) {
		missing = append(missing, rotatePath)
	} else if err != nil {
		t.Fatalf("error stat'ing S5 rotate fixture %s: %v", rotatePath, err)
	}
	if _, err := os.Stat(gameStatePath); os.IsNotExist(err) {
		missing = append(missing, gameStatePath)
	} else if err != nil {
		t.Fatalf("error stat'ing S5 game_state fixture %s: %v", gameStatePath, err)
	}

	if len(missing) > 0 {
		t.Fatalf("S5 expanded fixtures missing: %v\nAdd representative chunk fixtures for rotate and GameState variants under pkg/testdata/s5/ so decodeSawyerChunk and S5ChunkReader can be exercised.", missing)
	}
}

func TestVisualGoldenPresence(t *testing.T) {
	// Expect at least one golden image for visual regression checks.
	// Suggested filename: pkg/testdata/golden/title_golden.png
	path := testdataPath("golden", "title_golden.png")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("visual golden image missing: %s\nAdd a small, deterministic golden PNG for the title screen under pkg/testdata/golden/ so visual regression tests can compare outputs.", path)
	} else if err != nil {
		t.Fatalf("error stat'ing golden image %s: %v", path, err)
	}
}
