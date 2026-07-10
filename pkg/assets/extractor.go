package assets

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
)

// ExportG1Images searches the data for G1 headers and attempts to parse them
// using ParseSimpleG1. For each successful parse, it writes PNG files into
// outDir/<basename>/img_<i>_j.png where i is the header index and j is the
// element index.
func ExportG1Images(data []byte, srcName string, outDir string) (int, error) {
	found := 0
	for off := 0; off < len(data)-4; off++ {
		if data[off] == 'G' && data[off+1] == '1' && data[off+2] == 0 && data[off+3] == 0 {
			imgs, err := ParseSimpleG1(data, off)
			if err != nil {
				// not a parsable simple G1 at this offset; continue
				continue
			}
			if len(imgs) == 0 {
				continue
			}
			base := filepath.Join(outDir, filepath.Base(srcName))
			err = os.MkdirAll(base, 0o755)
			if err != nil {
				return found, err
			}
			for i, img := range imgs {
				fname := filepath.Join(base, fmt.Sprintf("img_%d_%d.png", off, i))
				f, err := os.Create(fname)
				if err != nil {
					return found, err
				}
				encErr := png.Encode(f, img)
				if closeErr := f.Close(); encErr == nil {
					encErr = closeErr
				}
				if encErr != nil {
					return found, encErr
				}
				found++
			}
		}
	}
	return found, nil
}
