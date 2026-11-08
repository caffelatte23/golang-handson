package pic

import "testing"

func TestPic(t *testing.T) {
	const dx, dy = 7, 5
	p := Pic(dx, dy)
	if len(p) != dy {
		t.Fatalf("len(Pic) = %d, want %d", len(p), dy)
	}
	for y, row := range p {
		if len(row) != dx {
			t.Fatalf("len(Pic[%d]) = %d, want %d", y, len(row), dx)
		}
		for x, v := range row {
			if v != uint8(x^y) {
				t.Errorf("Pic[%d][%d] = %d, want %d", y, x, v, x^y)
			}
		}
	}
}
