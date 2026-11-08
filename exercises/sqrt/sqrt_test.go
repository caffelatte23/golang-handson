package sqrt

import (
	"math"
	"testing"
)

func TestSqrt(t *testing.T) {
	for _, x := range []float64{0, 1, 2, 9, 1e-4, 123456.789} {
		got, want := Sqrt(x), math.Sqrt(x)
		if math.Abs(got-want) > 1e-9 {
			t.Errorf("Sqrt(%v) = %v, want %v", x, got, want)
		}
	}
}
