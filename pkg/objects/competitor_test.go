package objects

import "testing"

// load: present emotions take successive pairs of images, absent ones the base.
func TestCompetitorImages(t *testing.T) {
	c := &CompetitorObject{Emotions: 1<<0 | 1<<1 | 1<<4}
	c.setImages(100)
	want := [9]uint32{100, 102, 100, 100, 104, 100, 100, 100, 100}
	if c.Images != want {
		t.Errorf("images %v, want %v", c.Images, want)
	}
}
